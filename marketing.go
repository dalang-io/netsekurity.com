package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// handleContact accepts the landing-page contact form (HTMX hx-post). Messages
// are stored in contact_messages and listed on /su; the handler used to answer
// "Message received!" and discard the body, so every lead sent through it was
// lost.
func handleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	msg := strings.TrimSpace(r.FormValue("message"))
	source := strings.TrimSpace(r.FormValue("source"))
	lang := strings.TrimSpace(r.FormValue("lang"))
	// Honeypot: bots fill every field, people never see this one.
	if r.FormValue("website") != "" {
		fmt.Fprint(w, contactOK(lang))
		return
	}
	if !strings.Contains(email, "@") || len(email) > 200 || msg == "" || len(msg) > 4000 || len(name) > 200 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, alertHTML("bad", contactText(lang, "invalid")))
		return
	}
	if len(source) > 100 {
		source = source[:100]
	}
	if _, err := db.Exec(`INSERT INTO contact_messages (name, email, message, source, ip) VALUES (?,?,?,?,?)`,
		name, email, msg, source, clientIP(r)); err != nil {
		log.Printf("contact: store failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, alertHTML("bad", contactText(lang, "failed")))
		return
	}
	log.Printf("contact: new message from %s (%s)", email, source)
	fmt.Fprint(w, contactOK(lang))
}

// contactText holds the form's answers per language. Anything but "id" is English.
func contactText(lang, key string) string {
	id := lang == "id"
	switch key {
	case "ok":
		if id {
			return "<strong>Pesan diterima.</strong> Kami membalas dari sales@dalang.io dalam 1 hari kerja."
		}
		return "<strong>Message received.</strong> We reply from sales@dalang.io within 1 business day."
	case "invalid":
		if id {
			return "Masukkan alamat email yang valid dan sebuah pesan."
		}
		return "Please enter a valid email and a message."
	}
	if id {
		return "Pesan Anda tidak dapat disimpan. Silakan email langsung ke sales@dalang.io."
	}
	return "Could not save your message. Please email sales@dalang.io directly."
}

func contactOK(lang string) string { return alertHTML("ok", contactText(lang, "ok")) }

// scansCompleted is the one number the site is allowed to quote as social
// proof: real completed assessments. Rendered only once it is large enough to
// mean something; a "3 scans run" badge would do the opposite of reassure.
func scansCompleted() int {
	var n int
	db.QueryRow(`SELECT COUNT(*) FROM pentests WHERE status='completed'`).Scan(&n)
	return n
}

const scansProofMin = 25

// renderScansProof returns the proof line for the landing page, or "" below the threshold.
func renderScansProof() string {
	n := scansCompleted()
	if n < scansProofMin {
		return ""
	}
	return fmt.Sprintf(`<p class="mt-3 text-sm text-gray-700"><strong class="font-extrabold text-gray-900">%d</strong> assessments delivered so far</p>`, n)
}

// handleProductScope renders PRODUCT_SCOPE.md as a real page inside the site
// layout: headings, lists and tables as semantic HTML that search engines and
// language models can read, instead of the preformatted text it used to be.
func handleProductScope(w http.ResponseWriter, r *http.Request) {
	md, err := staticFS.ReadFile("static/PRODUCT_SCOPE.md")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	body := template.HTML(`<article class="nsk-doc mx-auto max-w-4xl px-4 py-12 sm:px-6">` +
		string(renderMarkdown(string(md))) + `</article>`)
	serveRawPage(w, r, rawPage{
		Key: "scope", Path: "/product-scope", Crumb: "Product scope", Body: body,
		Title:       "Product Scope and Methodology: What Netsekurity Checks",
		Description: "Exactly what a Netsekurity external web security assessment checks, what it does not, how findings are scored, and the deliverables, with the whitebox escalation tier.",
	})
}

var _ = strings.TrimSpace
