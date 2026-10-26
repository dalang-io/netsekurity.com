package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const verificationPath = "/.well-known/netsekurity-verification.txt"

// The dependency boundary lets tests exercise verification without public DNS.
var checkDomainControl = verifyDomainControl

func verificationMethod(raw string) string {
	switch raw {
	case "", "txt":
		return "txt"
	case "file", "meta":
		return raw
	}
	return ""
}

// Deny special-purpose networks as well as RFC1918. DNS answers are checked at
// connection time and the socket is opened to that exact IP (no second lookup).
var verificationDenied = func() []netip.Prefix {
	var out []netip.Prefix
	for _, s := range []string{"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24", "192.168.0.0/16", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "224.0.0.0/3", "::/96", "64:ff9b::/96", "64:ff9b:1::/48", "100::/64", "2001::/23", "2001:db8::/32", "2002::/16", "fc00::/7", "fe80::/10", "ff00::/8"} {
		out = append(out, netip.MustParsePrefix(s))
	}
	return out
}()

func publicVerificationIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	for _, p := range verificationDenied {
		if p.Contains(ip) {
			return false
		}
	}
	return true
}

func verificationDial(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil || (port != "443" && port != "53") {
		return nil, errors.New("invalid verification destination")
	}
	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil || len(ips) == 0 {
		return nil, errors.New("could not resolve verification destination")
	}
	for _, ip := range ips {
		if !publicVerificationIP(ip) {
			return nil, errors.New("verification requires a public address")
		}
	}
	var last error
	for _, ip := range ips {
		conn, err := (&net.Dialer{Timeout: 4 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		last = err
	}
	return nil, last
}

func verificationClient() *http.Client {
	return &http.Client{
		Timeout:   12 * time.Second,
		Transport: &http.Transport{Proxy: nil, DialContext: verificationDial, TLSHandshakeTimeout: 5 * time.Second, DisableKeepAlives: true, MaxResponseHeaderBytes: 32 << 10},
		// Verify the exact hostname. A redirect must never transfer proof from
		// another origin, a login page or an attacker-controlled upload service.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

func verifyDomainControl(ctx context.Context, domain, token, method string) error {
	if !isValidDomain(domain) || token == "" || verificationMethod(method) == "" {
		return errors.New("invalid verification request")
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if method == "txt" || method == "" {
		name := "_netsekurity." + domain
		records, _ := net.DefaultResolver.LookupTXT(ctx, name)
		for _, record := range records {
			if record == token {
				return nil
			}
		}
		// Ask authoritative servers too, with the same deadline and public-IP
		// restriction. Walk up to find the zone for a hostname without NS records.
		zone := domain
		for strings.Contains(zone, ".") {
			nss, err := net.DefaultResolver.LookupNS(ctx, zone)
			if err == nil && len(nss) > 0 {
				for i, ns := range nss {
					if i >= 4 {
						break
					}
					server := strings.TrimSuffix(ns.Host, ".")
					resolver := &net.Resolver{PreferGo: true, Dial: func(c context.Context, network, _ string) (net.Conn, error) {
						return verificationDial(c, network, net.JoinHostPort(server, "53"))
					}}
					records, _ = resolver.LookupTXT(ctx, name)
					for _, record := range records {
						if record == token {
							return nil
						}
					}
				}
				break
			}
			zone = zone[strings.IndexByte(zone, '.')+1:]
		}
		return errors.New("the DNS token could not be confirmed")
	}
	return verifyHTTPControl(ctx, verificationClient(), domain, token, method)
}

func verifyHTTPControl(ctx context.Context, client *http.Client, domain, token, method string) error {
	path := "/"
	limit := int64(1 << 20)
	if method == "file" {
		path, limit = verificationPath, 4096
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://"+domain+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "NetSekurity Verification")
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("could not retrieve verification proof over HTTPS")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("verification location must return HTTP 200 without redirecting")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil || int64(len(body)) > limit {
		return errors.New("verification response is unreadable or too large")
	}
	if method == "file" {
		if strings.TrimSpace(string(body)) == token {
			return nil
		}
	} else if method == "meta" && verificationMetaMatches(string(body), token) {
		return nil
	}
	return errors.New("the verification token was not found at the expected location")
}

func verificationMetaMatches(body, token string) bool {

	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return false
	}
	var head *html.Node
	for root := doc.FirstChild; root != nil; root = root.NextSibling {
		if root.Type != html.ElementNode || root.Data != "html" {
			continue
		}
		for child := root.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && child.Data == "head" {
				head = child
				break
			}
		}
	}
	if head == nil {
		return false
	}
	for node := head.FirstChild; node != nil; node = node.NextSibling {
		if node.Type != html.ElementNode || node.Data != "meta" {
			continue
		}
		var name, content string
		var names, contents int
		for _, a := range node.Attr {
			if a.Key == "name" {
				name = a.Val
				names++
			}
			if a.Key == "content" {
				content = a.Val
				contents++
			}
		}
		if names == 1 && contents == 1 && name == "netsekurity-verification" && content == token {
			return true
		}
	}
	return false
}

func recheckDomainControl(ctx context.Context, userID, domain string) error {
	var token, method, status string
	if err := db.QueryRow(`SELECT txt_verification_token, verified_by, status FROM domains WHERE user_id=? AND domain=?`, userID, domain).Scan(&token, &method, &status); err != nil || status != "verified" {
		return errors.New("verify this website first")
	}
	if verificationMethod(method) == "" {
		return errors.New("verify this website with DNS, a file or a meta tag before starting a self-service check")
	}
	if err := checkDomainControl(ctx, domain, token, method); err != nil {
		if _, updateErr := db.Exec(`UPDATE domains SET status='pending', verified_at=NULL WHERE user_id=? AND domain=?`, userID, domain); updateErr != nil {
			return updateErr
		}
		return fmt.Errorf("website control could not be confirmed; check your verification setup and try again: %w", err)
	}
	_, err := db.Exec(`UPDATE domains SET verified_at=? WHERE user_id=? AND domain=?`, now(), userID, domain)
	return err
}
