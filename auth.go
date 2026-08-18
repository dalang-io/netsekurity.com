package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const authCookie = "nsk_session"

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ensureUser upserts a user by Google identity and returns the user row id.
func ensureUser(email, name, picture, sub string) (string, error) {
	var id string
	err := db.QueryRow(`SELECT id FROM users WHERE google_sub = ?`, sub).Scan(&id)
	if err == nil {
		db.Exec(`UPDATE users SET email=?, name=?, picture=? WHERE id=?`, email, name, picture, id)
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	// Existing user by email but not yet linked to google_sub.
	err = db.QueryRow(`SELECT id FROM users WHERE email = ?`, email).Scan(&id)
	if err == nil {
		db.Exec(`UPDATE users SET google_sub=?, name=?, picture=? WHERE id=?`, sub, name, picture, id)
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	id = "u_" + randomHex(12)
	if _, err := db.Exec(`INSERT INTO users (id, email, name, picture, google_sub) VALUES (?,?,?,?,?)`,
		id, email, name, picture, sub); err != nil {
		return "", err
	}
	// Ensure a credit balance row exists.
	db.Exec(`INSERT OR IGNORE INTO credit_balance (user_id, balance) VALUES (?,0)`, id)
	return id, nil
}

// handleGoogleLogin starts the Google OAuth authorization code flow.
func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	clientID := getenv("GOOGLE_CLIENT_ID", "")
	redirect := getenv("GOOGLE_REDIRECT_URL", "")
	if clientID == "" || redirect == "" {
		http.Error(w, "Google OAuth not configured", http.StatusInternalServerError)
		return
	}
	state := randomHex(16)
	http.SetCookie(w, &http.Cookie{Name: "nsk_oauth_state", Value: state, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, MaxAge: 600})
	u := url.URL{
		Scheme: "https", Host: "accounts.google.com", Path: "/o/oauth2/v2/auth",
		RawQuery: url.Values{
			"client_id":     {clientID},
			"redirect_uri":  {redirect},
			"response_type": {"code"},
			"scope":         {"openid email profile"},
			"access_type":   {"online"},
			"prompt":        {"select_account"}, // always show the Google account picker
			"state":         {state},
		}.Encode(),
	}
	http.Redirect(w, r, u.String(), http.StatusFound)
}

type googleUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// handleGoogleCallback exchanges the code, loads the profile, issues a JWT.
func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
	clientID := getenv("GOOGLE_CLIENT_ID", "")
	clientSecret := getenv("GOOGLE_CLIENT_SECRET", "")
	redirect := getenv("GOOGLE_REDIRECT_URL", "")
	form := url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirect},
		"grant_type":    {"authorization_code"},
	}
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		http.Error(w, "token exchange failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil || tok.AccessToken == "" {
		http.Error(w, "token exchange failed", http.StatusBadGateway)
		return
	}
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	uresp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "userinfo failed", http.StatusBadGateway)
		return
	}
	defer uresp.Body.Close()
	var gu googleUser
	if err := json.NewDecoder(uresp.Body).Decode(&gu); err != nil || gu.Email == "" {
		http.Error(w, "userinfo failed", http.StatusBadGateway)
		return
	}
	userID, err := ensureUser(gu.Email, gu.Name, "", gu.ID)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}
	token, err := issueJWT(userID, gu.Email)
	if err != nil {
		http.Error(w, "failed to issue token", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: authCookie, Value: token, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, MaxAge: int((24 * time.Hour).Seconds())})
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: authCookie, Value: "", Path: "/", HttpOnly: true, Secure: true, MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusFound)
}

// handleOneTap accepts a Google One Tap ID token, verifies it with Google, and
// creates a session — letting signed-in Chrome users log in with one tap.
func handleOneTap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Credential string `json:"credential"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Credential == "" {
		http.Error(w, "missing credential", http.StatusBadRequest)
		return
	}
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(body.Credential))
	if err != nil {
		http.Error(w, "verification failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	var info struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil || info.Sub == "" || info.Email == "" {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	if info.EmailVerified != "true" {
		http.Error(w, "email not verified", http.StatusUnauthorized)
		return
	}
	userID, err := ensureUser(info.Email, info.Name, "", info.Sub)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}
	token, err := issueJWT(userID, info.Email)
	if err != nil {
		http.Error(w, "failed to issue token", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: authCookie, Value: token, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, MaxAge: int((24 * time.Hour).Seconds())})
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

// currentUser extracts the JWT claims from the auth cookie.
func currentUser(r *http.Request) (*Claims, error) {
	c, err := r.Cookie(authCookie)
	if err != nil || c.Value == "" {
		return nil, fmt.Errorf("no session")
	}
	return verifyJWT(c.Value)
}

// requireAuth is middleware that redirects unauthenticated users to /login.
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := currentUser(r); err != nil {
			// HTMX requests get a redirect header so the client can navigate.
			w.Header().Set("HX-Redirect", "/login")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}

func readBody(r *http.Request) string {
	b, _ := io.ReadAll(r.Body)
	return strings.TrimSpace(string(b))
}
