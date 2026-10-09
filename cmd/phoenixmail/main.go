package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"embed"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed static/*
var webFS embed.FS

type Mail struct {
	ID          string   `json:"id"`
	From        string   `json:"from"`
	Email       string   `json:"email"`
	AccountID   string   `json:"accountId,omitempty"`
	To          string   `json:"to"`
	Cc          string   `json:"cc,omitempty"`
	Bcc         string   `json:"bcc,omitempty"`
	Subject     string   `json:"subject"`
	Preview     string   `json:"preview"`
	Body        string   `json:"body"`
	Time        string   `json:"time"`
	Date        string   `json:"date"`
	Folder      string   `json:"folder"`
	Read        bool     `json:"read"`
	Starred     bool     `json:"starred"`
	Important   bool     `json:"important"`
	Attachments []string `json:"attachments,omitempty"`
}

type SMTPConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Security       string `json:"security"`
	Username       string `json:"username"`
	Authentication string `json:"authentication,omitempty"`
	From           string `json:"from"`
	DisplayName    string `json:"displayName"`
}

type AIConfig struct {
	Endpoint  string `json:"endpoint"`
	Model     string `json:"model"`
	ModelPath string `json:"modelPath"`
	AutoStart bool   `json:"autoStart"`
}

type IMAPConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Security       string `json:"security"`
	Username       string `json:"username"`
	Authentication string `json:"authentication,omitempty"`
}

type MailAccount struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	DisplayName string     `json:"displayName"`
	IMAP        IMAPConfig `json:"imap"`
	SMTP        SMTPConfig `json:"smtp"`
}

type AppConfig struct {
	SMTP           SMTPConfig    `json:"smtp"`
	AI             AIConfig      `json:"ai"`
	BrandFooter    bool          `json:"brandFooter"`
	Accounts       []MailAccount `json:"accounts"`
	DefaultAccount string        `json:"defaultAccount"`
}

type Store struct {
	Version int       `json:"version"`
	Mails   []Mail    `json:"mails"`
	Config  AppConfig `json:"config"`
}

type oauthToken struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

type persistedOAuthToken struct {
	RefreshToken string `json:"refreshToken"`
}

type oauthPending struct {
	AccountID string
	Verifier  string
	Created   time.Time
}

// xoauth2Auth implements the XOAUTH2 SASL mechanism used by Outlook SMTP AUTH.
type xoauth2Auth struct {
	username string
	token    string
}

func (a *xoauth2Auth) Start(*smtp.ServerInfo) (string, []byte, error) {
	return "XOAUTH2", []byte("user=" + a.username + "\x01auth=Bearer " + a.token + "\x01\x01"), nil
}
func (a *xoauth2Auth) Next(challenge []byte, more bool) ([]byte, error) {
	if more {
		return nil, nil
	}
	return nil, nil
}

type App struct {
	mu               sync.RWMutex
	path             string
	data             Store
	smtpPassword     string
	accountPasswords map[string]string
	aiCmd            *exec.Cmd
	aiDownloadState  string
	aiDownloadError  string
	aiDownloadBytes  int64
	aiDownloadTotal  int64
	oauthMu          sync.Mutex
	oauthTokens      map[string]*oauthToken
	oauthPending     map[string]oauthPending
}

func demoStore() Store {
	return Store{Version: 4, Config: AppConfig{BrandFooter: true, Accounts: []MailAccount{
		{ID: "demo-personnel", Name: "Personnel", Email: "francois@gmail.com", DisplayName: "François", IMAP: IMAPConfig{}, SMTP: SMTPConfig{}},
		{ID: "demo-travail", Name: "Travail", Email: "travail@entreprise.com", DisplayName: "François – Travail", IMAP: IMAPConfig{}, SMTP: SMTPConfig{}},
		{ID: "demo-outlook", Name: "Outlook", Email: "autre@outlook.com", DisplayName: "François – Outlook", IMAP: IMAPConfig{}, SMTP: SMTPConfig{}},
	}, DefaultAccount: "demo-personnel"}, Mails: []Mail{
		{ID: "1", From: "Marie Dupont", Email: "marie.dupont@entreprise.com", To: "moi", Subject: "Réunion de demain", Preview: "Bonjour François, Est-ce que la réunion de demain est toujours confirmée ?", Body: "Bonjour François,\n\nEst-ce que la réunion de demain est toujours confirmée ?\n\nVoici l’ordre du jour :\n\n• Point sur le projet\n• Budget\n• Prochaines étapes\n\nN’hésite pas à me faire savoir si tu as des questions.\n\nMerci,\nMarie", Time: "09:15", Date: "Aujourd’hui", Folder: "inbox", Read: false, Starred: true, Important: true},
		{ID: "2", From: "Jean Tremblay", Email: "jean.tremblay@entreprise.com", To: "moi", Subject: "Facture #8492", Preview: "Voici la facture pour le mois de novembre. N’hésite pas à me contacter.", Body: "Bonjour François,\n\nVoici la facture pour le mois de novembre. N’hésite pas à me contacter si tu as des questions.\n\nJean", Time: "08:42", Date: "Aujourd’hui", Folder: "inbox", Read: false, Starred: true, Attachments: []string{"facture-8492.pdf"}},
		{ID: "3", From: "Service Client", Email: "service@entreprise.com", To: "moi", Subject: "Confirmation de commande", Preview: "Votre commande #4587 a été expédiée. Vous trouverez les détails ci-dessous.", Body: "Bonjour,\n\nVotre commande #4587 a été expédiée. Vous trouverez les détails dans la pièce jointe.\n\nMerci.", Time: "07:30", Date: "Aujourd’hui", Folder: "inbox", Read: true, Attachments: []string{"commande-4587.pdf"}},
		{ID: "4", From: "Luc Martin", Email: "luc.martin@entreprise.com", To: "moi", Subject: "Projet PhoenixMail", Preview: "Voici les derniers documents pour le projet PhoenixMail.", Body: "Bonjour François,\n\nVoici les derniers documents pour le projet PhoenixMail.\n\nLuc", Time: "Hier", Date: "Hier", Folder: "inbox", Read: true, Important: true, Attachments: []string{"phoenixmail-specs.pdf"}},
		{ID: "5", From: "Google", Email: "no-reply@google.com", To: "moi", Subject: "Alerte de sécurité", Preview: "Nouvelle connexion sur votre compte.", Body: "Une nouvelle connexion à votre compte a été détectée.", Time: "Hier", Date: "Hier", Folder: "inbox", Read: true},
		{ID: "6", From: "Amazon.ca", Email: "no-reply@amazon.ca", To: "moi", Subject: "Votre commande a été expédiée", Preview: "Votre colis est en route.", Body: "Votre colis est en route.", Time: "Hier", Date: "Hier", Folder: "inbox", Read: true},
		{ID: "7", From: "Sophie Roy", Email: "sophie.roy@example.com", To: "moi", Subject: "Photos du week-end", Preview: "Voici quelques photos de notre sortie…", Body: "Voici quelques photos de notre sortie de samedi.", Time: "Lun.", Date: "Lundi", Folder: "inbox", Read: true, Attachments: []string{"photos.zip"}},
		{ID: "8", From: "Banque Nationale", Email: "avis@bn.example.com", To: "moi", Subject: "Relevé mensuel", Preview: "Votre relevé du mois de novembre est maintenant disponible.", Body: "Votre relevé mensuel est maintenant disponible.", Time: "Lun.", Date: "Lundi", Folder: "inbox", Read: true},
		{ID: "9", From: "Pierre Gagnon", Email: "pierre.gagnon@example.com", To: "moi", Subject: "Souper vendredi ?", Preview: "Ça te tente un souper vendredi soir ?", Body: "Salut François,\n\nÇa te tente un souper vendredi soir ?", Time: "Lun.", Date: "Lundi", Folder: "inbox", Read: true},
		{ID: "10", From: "Moi", Email: "moi@phoenixmail.local", To: "jean.tremblay@entreprise.com", Subject: "Projet PhoenixMail", Preview: "Je voulais faire un suivi sur le projet PhoenixMail.", Body: "Bonjour Jean,\n\nJe voulais faire un suivi sur le projet PhoenixMail.\nPeux-tu me confirmer l’avancement et les prochaines étapes ?\n\nMerci !", Time: "10:02", Date: "Aujourd’hui", Folder: "drafts", Read: true},
	}}
}

func dataPath() string {
	if p := os.Getenv("PHOENIXMAIL_DATA"); p != "" {
		return p
	}
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "PhoenixMail", "data.json")
	}
	return filepath.Join("data", "phoenixmail.json")
}

func NewApp(path string) *App {
	a := &App{path: path, accountPasswords: map[string]string{}, oauthTokens: map[string]*oauthToken{}, oauthPending: map[string]oauthPending{}}
	_ = a.load()
	_ = a.loadOAuthTokenCache()
	return a
}

func (a *App) oauthCachePath() string { return a.path + ".oauth.json" }

func (a *App) loadOAuthTokenCache() error {
	b, err := os.ReadFile(a.oauthCachePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	var saved map[string]persistedOAuthToken
	if err := json.Unmarshal(b, &saved); err != nil {
		return err
	}
	a.oauthMu.Lock()
	defer a.oauthMu.Unlock()
	for id, tok := range saved {
		if strings.TrimSpace(tok.RefreshToken) != "" {
			a.oauthTokens[id] = &oauthToken{RefreshToken: tok.RefreshToken}
		}
	}
	return nil
}

func (a *App) saveOAuthTokenCache() error {
	a.oauthMu.Lock()
	saved := make(map[string]persistedOAuthToken, len(a.oauthTokens))
	for id, tok := range a.oauthTokens {
		if tok != nil && strings.TrimSpace(tok.RefreshToken) != "" {
			saved[id] = persistedOAuthToken{RefreshToken: tok.RefreshToken}
		}
	}
	a.oauthMu.Unlock()
	if err := os.MkdirAll(filepath.Dir(a.oauthCachePath()), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}
	tmp := a.oauthCachePath() + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, a.oauthCachePath())
}

func (a *App) load() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	b, err := os.ReadFile(a.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			a.data = demoStore()
			return a.saveLocked()
		}
		return err
	}
	if err := json.Unmarshal(b, &a.data); err != nil {
		return err
	}
	if a.data.Version < 4 {
		a.data.Version = 4
		a.data.Config.BrandFooter = true
	}
	a.ensureAccountsLocked()
	return nil
}

func (a *App) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(a.path), 0o755); err != nil {
		return err
	}
	tmp := a.path + ".tmp"
	b, err := json.MarshalIndent(a.data, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, a.path)
}

func (a *App) save() error { a.mu.Lock(); defer a.mu.Unlock(); return a.saveLocked() }

func (a *App) list(folder, q string) []Mail {
	a.mu.RLock()
	defer a.mu.RUnlock()
	q = strings.ToLower(strings.TrimSpace(q))
	out := make([]Mail, 0)
	for _, m := range a.data.Mails {
		if folder == "important" {
			if !m.Important || m.Folder == "trash" {
				continue
			}
		} else if folder == "all" {
			if m.Folder == "trash" {
				continue
			}
		} else if m.Folder != folder {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(strings.Join([]string{m.From, m.Email, m.Subject, m.Preview, m.Body}, " ")), q) {
			continue
		}
		out = append(out, m)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

func (a *App) find(id string) (*Mail, int) {
	for i := range a.data.Mails {
		if a.data.Mails[i].ID == id {
			return &a.data.Mails[i], i
		}
	}
	return nil, -1
}

func (a *App) counts() map[string]int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	c := map[string]int{"inbox": 0, "important": 0, "drafts": 0, "sent": 0, "archive": 0, "trash": 0, "unread": 0}
	for _, m := range a.data.Mails {
		if _, ok := c[m.Folder]; ok {
			c[m.Folder]++
		}
		if m.Important && m.Folder != "trash" {
			c["important"]++
		}
		if !m.Read && m.Folder == "inbox" {
			c["unread"]++
		}
	}
	return c
}

// Thunderbird-compatible autoconfiguration XML.
type autoConfigXML struct {
	XMLName  xml.Name `xml:"clientConfig"`
	Provider struct {
		ID               string   `xml:"id,attr"`
		Domain           []string `xml:"domain"`
		DisplayName      string   `xml:"displayName"`
		DisplayShortName string   `xml:"displayShortName"`
		Incoming         []struct {
			Type           string `xml:"type,attr"`
			Hostname       string `xml:"hostname"`
			Port           int    `xml:"port"`
			SocketType     string `xml:"socketType"`
			Authentication string `xml:"authentication"`
			Username       string `xml:"username"`
		} `xml:"incomingServer"`
		Outgoing []struct {
			Type           string `xml:"type,attr"`
			Hostname       string `xml:"hostname"`
			Port           int    `xml:"port"`
			SocketType     string `xml:"socketType"`
			Authentication string `xml:"authentication"`
			Username       string `xml:"username"`
		} `xml:"outgoingServer"`
	} `xml:"emailProvider"`
}

func normalizeSocketType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "ssl", "ssl/tls", "tls", "ssl_tls":
		return "ssl"
	case "starttls", "start_tls":
		return "starttls"
	default:
		return "plain"
	}
}

func resolveUsername(tpl, email string) string {
	t := strings.TrimSpace(tpl)
	if t == "" {
		return email
	}
	t = strings.ReplaceAll(t, "%EMAILADDRESS%", email)
	parts := strings.SplitN(email, "@", 2)
	local := parts[0]
	t = strings.ReplaceAll(t, "%EMAILLOCALPART%", local)
	return t
}

func (a *App) autoconfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	email := strings.TrimSpace(req.Email)
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		writeJSON(w, 400, map[string]any{"error": "Adresse courriel invalide."})
		return
	}
	domain := strings.ToLower(strings.TrimSpace(parts[1]))
	// High-confidence local profiles for common providers. These keep first-run
	// discovery useful even if the remote ISPDB is temporarily unreachable.
	if local, ok := localProviderAutoconfig(domain, email); ok {
		writeJSON(w, 200, local)
		return
	}
	client := &http.Client{Timeout: 8 * time.Second}
	urls := []string{
		"https://autoconfig.thunderbird.net/v1.1/" + domain,
		"https://autoconfig." + domain + "/mail/config-v1.1.xml?emailaddress=" + urlQueryEscape(email),
		"https://" + domain + "/.well-known/autoconfig/mail/config-v1.1.xml",
	}
	var lastErr error
	for _, u := range urls {
		resp, err := client.Get(u)
		if err != nil {
			lastErr = err
			continue
		}
		b, readErr := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}
		var cfg autoConfigXML
		if err := xml.Unmarshal(b, &cfg); err != nil {
			lastErr = err
			continue
		}
		var imap, smtp *struct {
			Host           string
			Port           int
			Security       string
			Authentication string
			Username       string
		}
		for _, in := range cfg.Provider.Incoming {
			if strings.EqualFold(in.Type, "imap") {
				v := struct {
					Host           string
					Port           int
					Security       string
					Authentication string
					Username       string
				}{in.Hostname, in.Port, normalizeSocketType(in.SocketType), strings.TrimSpace(in.Authentication), resolveUsername(in.Username, email)}
				imap = &v
				break
			}
		}
		for _, out := range cfg.Provider.Outgoing {
			if strings.EqualFold(out.Type, "smtp") {
				v := struct {
					Host           string
					Port           int
					Security       string
					Authentication string
					Username       string
				}{out.Hostname, out.Port, normalizeSocketType(out.SocketType), strings.TrimSpace(out.Authentication), resolveUsername(out.Username, email)}
				smtp = &v
				break
			}
		}
		if imap == nil || smtp == nil {
			continue
		}
		writeJSON(w, 200, map[string]any{
			"ok": true, "source": "thunderbird-autoconfig", "provider": cfg.Provider.DisplayName, "shortName": cfg.Provider.DisplayShortName,
			"imap": map[string]any{"host": imap.Host, "port": imap.Port, "security": imap.Security, "username": imap.Username, "authentication": imap.Authentication},
			"smtp": map[string]any{"host": smtp.Host, "port": smtp.Port, "security": smtp.Security, "username": smtp.Username, "authentication": smtp.Authentication, "from": email},
		})
		return
	}
	if lastErr == nil {
		lastErr = errors.New("configuration introuvable")
	}
	writeJSON(w, 404, map[string]any{"ok": false, "error": "Aucune configuration automatique trouvée pour ce domaine. La configuration manuelle reste disponible.", "detail": lastErr.Error()})
}

func localProviderAutoconfig(domain, email string) (map[string]any, bool) {
	d := strings.ToLower(domain)
	var imapHost, smtpHost, provider, shortName, imapAuth, smtpAuth string
	var imapPort, smtpPort int
	var imapSec, smtpSec string
	switch d {
	case "gmail.com", "googlemail.com":
		provider, shortName = "Google", "Gmail"
		imapHost, imapPort, imapSec, imapAuth = "imap.gmail.com", 993, "ssl", "OAuth2"
		smtpHost, smtpPort, smtpSec, smtpAuth = "smtp.gmail.com", 587, "starttls", "OAuth2"
	case "outlook.com", "hotmail.com", "live.com", "msn.com":
		provider, shortName = "Microsoft", "Outlook"
		imapHost, imapPort, imapSec, imapAuth = "outlook.office365.com", 993, "ssl", "OAuth2"
		smtpHost, smtpPort, smtpSec, smtpAuth = "smtp-mail.outlook.com", 587, "starttls", "OAuth2"
	case "yahoo.com", "ymail.com", "rocketmail.com":
		provider, shortName = "Yahoo", "Yahoo Mail"
		imapHost, imapPort, imapSec, imapAuth = "imap.mail.yahoo.com", 993, "ssl", "OAuth2"
		smtpHost, smtpPort, smtpSec, smtpAuth = "smtp.mail.yahoo.com", 465, "ssl", "OAuth2"
	case "icloud.com", "me.com", "mac.com":
		provider, shortName = "Apple", "iCloud Mail"
		imapHost, imapPort, imapSec, imapAuth = "imap.mail.me.com", 993, "ssl", "password-cleartext"
		smtpHost, smtpPort, smtpSec, smtpAuth = "smtp.mail.me.com", 587, "starttls", "password-cleartext"
	default:
		return nil, false
	}
	return map[string]any{"ok": true, "source": "phoenixmail-local-profile", "provider": provider, "shortName": shortName, "imap": map[string]any{"host": imapHost, "port": imapPort, "security": imapSec, "username": email, "authentication": imapAuth}, "smtp": map[string]any{"host": smtpHost, "port": smtpPort, "security": smtpSec, "username": email, "authentication": smtpAuth, "from": email}}, true
}

func urlQueryEscape(s string) string { return url.QueryEscape(s) }

func cleanCommandOutput(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"'")
	s = strings.TrimSpace(s)
	return s
}

func commandOutput(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	b, err := cmd.Output()
	if err != nil {
		return ""
	}
	return cleanCommandOutput(string(b))
}

func systemAppearance() map[string]string {
	a := map[string]string{"scheme": "dark", "accent": "blue"}
	switch runtime.GOOS {
	case "linux":
		if v := commandOutput("gsettings", "get", "org.gnome.desktop.interface", "color-scheme"); v != "" {
			switch v {
			case "prefer-light", "default":
				a["scheme"] = "light"
			case "prefer-dark":
				a["scheme"] = "dark"
			}
		}
		if v := commandOutput("gsettings", "get", "org.gnome.desktop.interface", "accent-color"); v != "" {
			a["accent"] = strings.ToLower(v)
		}
	case "darwin":
		if v := commandOutput("defaults", "read", "-g", "AppleInterfaceStyle"); strings.EqualFold(v, "Dark") {
			a["scheme"] = "dark"
		} else {
			a["scheme"] = "light"
		}
		if v := commandOutput("defaults", "read", "-g", "AppleAccentColor"); v != "" {
			colors := map[string]string{"0": "red", "1": "orange", "2": "yellow", "3": "green", "4": "blue", "5": "purple", "6": "pink"}
			if c, ok := colors[v]; ok {
				a["accent"] = c
			}
		}
	case "windows":
		// Browser media-query remains the fallback for Windows in this prototype.
	}
	return a
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func decodeJSON(r *http.Request, dst any) error { return json.NewDecoder(r.Body).Decode(dst) }

func (a *App) mailsHandler(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("folder")
	if folder == "" {
		folder = "inbox"
	}
	if folder == "all" {
		folder = "inbox"
	}
	if r.Method == http.MethodGet {
		accountID := strings.TrimSpace(r.URL.Query().Get("account"))
		writeJSON(w, 200, map[string]any{
			"mails":  a.listForAccount(folder, r.URL.Query().Get("q"), accountID),
			"counts": a.countsForAccount(accountID),
		})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		To, Cc, Bcc, Subject, Body, DraftID string
		Attachments                         []string `json:"attachments"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	now := time.Now()
	id := fmt.Sprintf("d-%d", now.UnixNano())
	m := Mail{ID: id, From: "Moi", Email: "moi@phoenixmail.local", To: strings.TrimSpace(req.To), Cc: strings.TrimSpace(req.Cc), Subject: req.Subject, Preview: firstLine(req.Body), Body: req.Body, Time: now.Format("15:04"), Date: "Aujourd’hui", Folder: "sent", Read: true, Attachments: req.Attachments}
	a.mu.Lock()
	if req.DraftID != "" {
		if _, idx := a.find(req.DraftID); idx >= 0 && a.data.Mails[idx].Folder == "drafts" {
			a.data.Mails = append(a.data.Mails[:idx], a.data.Mails[idx+1:]...)
		}
	}
	a.data.Mails = append(a.data.Mails, m)
	err := a.saveLocked()
	a.mu.Unlock()
	if err != nil {
		http.Error(w, "save failed", 500)
		return
	}
	writeJSON(w, 201, m)
}

func firstLine(s string) string {
	for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

func (a *App) mailAction(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/mails/")
	parts := strings.Split(strings.Trim(id, "/"), "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	id, action := parts[0], parts[1]
	a.mu.Lock()
	defer a.mu.Unlock()
	m, _ := a.find(id)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	switch action {
	case "read":
		m.Read = true
	case "unread":
		m.Read = false
	case "star":
		m.Starred = !m.Starred
	case "important":
		m.Important = !m.Important
	case "archive":
		m.Folder = "archive"
	case "trash":
		m.Folder = "trash"
	case "restore":
		m.Folder = "inbox"
	case "move":
		var req struct {
			Folder string `json:"folder"`
		}
		if err := decodeJSON(r, &req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		folder := strings.ToLower(strings.TrimSpace(req.Folder))
		switch folder {
		case "inbox", "important", "drafts", "sent", "archive", "trash":
			m.Folder = folder
		default:
			http.Error(w, "invalid folder", http.StatusBadRequest)
			return
		}
	default:
		http.NotFound(w, r)
		return
	}
	if err := a.saveLocked(); err != nil {
		http.Error(w, "save failed", 500)
		return
	}
	writeJSON(w, 200, m)
}

func (a *App) draftHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, 200, a.list("drafts", r.URL.Query().Get("q")))
		return
	}
	if r.Method == http.MethodDelete {
		id := strings.TrimPrefix(r.URL.Path, "/api/drafts/")
		if id == "" || id == r.URL.Path {
			http.Error(w, "draft id required", 400)
			return
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		if _, idx := a.find(id); idx >= 0 && a.data.Mails[idx].Folder == "drafts" {
			a.data.Mails = append(a.data.Mails[:idx], a.data.Mails[idx+1:]...)
			if err := a.saveLocked(); err != nil {
				http.Error(w, "save failed", 500)
				return
			}
		}
		writeJSON(w, 200, map[string]any{"ok": true})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct{ ID, To, Cc, Bcc, Subject, Body, AccountID string }
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	m, idx := a.find(req.ID)
	if m == nil {
		m = &Mail{ID: fmt.Sprintf("d-%d", time.Now().UnixNano()), From: "Moi", Email: "moi@phoenixmail.local", Folder: "drafts", Read: true}
		a.data.Mails = append(a.data.Mails, *m)
		idx = len(a.data.Mails) - 1
		m = &a.data.Mails[idx]
	}
	m.To = req.To
	m.Cc = req.Cc
	m.Bcc = req.Bcc
	m.AccountID = req.AccountID
	m.Subject = req.Subject
	m.Body = req.Body
	m.Preview = firstLine(req.Body)
	m.Time = time.Now().Format("15:04")
	m.Date = "Aujourd’hui"
	if err := a.saveLocked(); err != nil {
		http.Error(w, "save failed", 500)
		return
	}
	writeJSON(w, 200, m)
}

func (a *App) configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		a.mu.RLock()
		cfg := a.data.Config
		a.mu.RUnlock()
		writeJSON(w, 200, map[string]any{"smtp": cfg.SMTP, "smtpConfigured": cfg.SMTP.Host != "", "accounts": publicAccounts(cfg.Accounts), "defaultAccount": cfg.DefaultAccount, "ai": cfg.AI, "brandFooter": cfg.BrandFooter})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		SMTP *struct {
			Host           string `json:"host"`
			Port           int    `json:"port"`
			Security       string `json:"security"`
			Authentication string `json:"authentication"`
			Username       string `json:"username"`
			From           string `json:"from"`
			DisplayName    string `json:"displayName"`
			Password       string `json:"password"`
		} `json:"smtp"`
		AI          *AIConfig `json:"ai"`
		BrandFooter *bool     `json:"brandFooter"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if req.SMTP != nil {
		sec := strings.ToLower(strings.TrimSpace(req.SMTP.Security))
		if sec != "ssl" && sec != "starttls" && sec != "plain" {
			sec = "starttls"
		}
		port := req.SMTP.Port
		if port == 0 {
			if sec == "ssl" {
				port = 465
			} else {
				port = 587
			}
		}
		a.data.Config.SMTP = SMTPConfig{Host: strings.TrimSpace(req.SMTP.Host), Port: port, Security: sec, Username: strings.TrimSpace(req.SMTP.Username), Authentication: strings.TrimSpace(req.SMTP.Authentication), From: strings.TrimSpace(req.SMTP.From), DisplayName: strings.TrimSpace(req.SMTP.DisplayName)}
		if req.SMTP.Password != "" {
			a.smtpPassword = req.SMTP.Password
		}
	}
	if req.BrandFooter != nil {
		a.data.Config.BrandFooter = *req.BrandFooter
	}
	if req.AI != nil {
		a.data.Config.AI.Endpoint = strings.TrimSpace(req.AI.Endpoint)
		a.data.Config.AI.Model = strings.TrimSpace(req.AI.Model)
	}
	if err := a.saveLocked(); err != nil {
		http.Error(w, "save failed", 500)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func phoenixMailFooterHTML() string {
	return `<div style="margin-top:28px;padding-top:14px;border-top:1px solid #303745;font-family:Arial,sans-serif;font-size:11px;color:#8e96a5"><span style="display:inline-block;padding:5px 9px;border:1px solid #3b414e;border-radius:7px;background:#171b24;color:#aeb5c2">✦ Propulsé par <span style="color:#ff8a1f;font-weight:700">PhoenixMail</span></span></div>`
}

func htmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;").Replace(s)
}

func phoenixMailHTMLBody(body string, footer bool) string {
	parts := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	for i := range parts {
		parts[i] = htmlEscape(parts[i])
	}
	h := `<div style="font-family:Arial,Helvetica,sans-serif;font-size:15px;line-height:1.6;color:#20242b">` + strings.Join(parts, "<br>") + `</div>`
	if footer {
		h += phoenixMailFooterHTML()
	}
	return h
}

func smtpMessageWithFooter(from, fromName string, to, cc, bcc, subject, body string, footer bool) ([]byte, []string, error) {
	recipients := []string{}
	add := func(v string) {
		for _, x := range strings.Split(v, ",") {
			x = strings.TrimSpace(x)
			if x != "" {
				recipients = append(recipients, x)
			}
		}
	}
	add(to)
	add(cc)
	add(bcc)
	if len(recipients) == 0 {
		return nil, nil, errors.New("aucun destinataire")
	}
	if from == "" {
		return nil, nil, errors.New("adresse d’envoi absente")
	}
	fromHeader := from
	if fromName != "" {
		fromHeader = mime.QEncoding.Encode("UTF-8", fromName) + " <" + from + ">"
	}
	msg := "From: " + fromHeader + "\r\n"
	if to != "" {
		msg += "To: " + to + "\r\n"
	}
	if cc != "" {
		msg += "Cc: " + cc + "\r\n"
	}
	boundary := "=_PhoenixMail_" + fmt.Sprintf("%d", time.Now().UnixNano())
	plainBody := body
	if footer {
		plainBody += "\n\n— Propulsé par PhoenixMail —"
	}
	htmlBody := phoenixMailHTMLBody(body, footer)
	msg += "Subject: " + mime.QEncoding.Encode("UTF-8", subject) + "\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n\r\n"
	msg += "--" + boundary + "\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n" + plainBody + "\r\n\r\n"
	msg += "--" + boundary + "\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n" + htmlBody + "\r\n\r\n--" + boundary + "--\r\n"
	return []byte(msg), recipients, nil
}

func sendSMTP(cfg SMTPConfig, password, to, cc, bcc, subject, body string, footer bool) error {
	if strings.TrimSpace(cfg.Authentication) == "" && isKnownOAuthSMTPHost(cfg.Host) {
		cfg.Authentication = "OAuth2"
	}
	if strings.EqualFold(strings.TrimSpace(cfg.Authentication), "OAuth2") {
		return errors.New("ce compte utilise OAuth2; connecte le compte au fournisseur avant l’envoi")
	}
	msg, recipients, err := smtpMessageWithFooter(cfg.From, cfg.DisplayName, to, cc, bcc, subject, body, footer)
	if err != nil {
		return err
	}
	if cfg.Host == "" {
		return errors.New("serveur SMTP non configuré")
	}
	port := cfg.Port
	if port == 0 {
		if cfg.Security == "ssl" {
			port = 465
		} else {
			port = 587
		}
	}
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(port))
	var client *smtp.Client
	if cfg.Security == "ssl" || port == 465 {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
		if err != nil {
			return err
		}
		client, err = smtp.NewClient(conn, cfg.Host)
		if err != nil {
			_ = conn.Close()
			return err
		}
	} else {
		client, err = smtp.Dial(addr)
		if err != nil {
			return err
		}
		if cfg.Security == "starttls" {
			if ok, _ := client.Extension("STARTTLS"); !ok {
				_ = client.Quit()
				return errors.New("le serveur SMTP ne propose pas STARTTLS")
			}
			if err := client.StartTLS(&tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
				_ = client.Quit()
				return err
			}
		}
	}
	defer client.Quit()
	if cfg.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", cfg.Username, password, cfg.Host)); err != nil {
			return err
		}
	}
	if err := client.Mail(cfg.From); err != nil {
		return err
	}
	for _, r := range recipients {
		if err := client.Rcpt(r); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return nil
}

func sendSMTPXOAUTH2(cfg SMTPConfig, token, to, cc, bcc, subject, body string, footer bool) error {
	msg, recipients, err := smtpMessageWithFooter(cfg.From, cfg.DisplayName, to, cc, bcc, subject, body, footer)
	if err != nil {
		return err
	}
	if cfg.Host == "" {
		return errors.New("serveur SMTP non configuré")
	}
	port := cfg.Port
	if port == 0 {
		port = 587
	}
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(port))
	var client *smtp.Client
	if cfg.Security == "ssl" || port == 465 {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
		if err != nil {
			return err
		}
		client, err = smtp.NewClient(conn, cfg.Host)
		if err != nil {
			_ = conn.Close()
			return err
		}
	} else {
		client, err = smtp.Dial(addr)
		if err != nil {
			return err
		}
		if cfg.Security == "starttls" {
			ok, _ := client.Extension("STARTTLS")
			if !ok {
				_ = client.Quit()
				return errors.New("le serveur SMTP ne propose pas STARTTLS")
			}
			if err := client.StartTLS(&tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
				_ = client.Quit()
				return err
			}
		}
	}
	defer client.Quit()
	if ok, _ := client.Extension("AUTH"); !ok {
		return errors.New("SMTP AUTH indisponible")
	}
	if err := client.Auth(&xoauth2Auth{username: cfg.Username, token: token}); err != nil {
		return err
	}
	if err := client.Mail(cfg.From); err != nil {
		return err
	}
	for _, r := range recipients {
		if err := client.Rcpt(r); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func (a *App) smtpTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		SMTP struct {
			Host           string `json:"host"`
			Port           int    `json:"port"`
			Security       string `json:"security"`
			Authentication string `json:"authentication"`
			Username       string `json:"username"`
			From           string `json:"from"`
			DisplayName    string `json:"displayName"`
			Password       string `json:"password"`
		} `json:"smtp"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	cfg := SMTPConfig{Host: strings.TrimSpace(req.SMTP.Host), Port: req.SMTP.Port, Security: strings.TrimSpace(req.SMTP.Security), Username: strings.TrimSpace(req.SMTP.Username), Authentication: strings.TrimSpace(req.SMTP.Authentication), From: strings.TrimSpace(req.SMTP.From), DisplayName: strings.TrimSpace(req.SMTP.DisplayName)}
	if cfg.Authentication == "" && isKnownOAuthSMTPHost(cfg.Host) {
		cfg.Authentication = "OAuth2"
	}
	if strings.EqualFold(cfg.Authentication, "OAuth2") {
		writeJSON(w, 409, map[string]any{"code": "oauth2_required", "message": "OAuth2 requis : utilisez Connexion Microsoft/Google pour ce compte. PhoenixMail ne tentera pas une authentification par mot de passe."})
		return
	}
	if cfg.Security == "ssl" || cfg.Port == 465 {
		port := cfg.Port
		if port == 0 {
			port = 465
		}
		addr := net.JoinHostPort(cfg.Host, strconv.Itoa(port))
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
		if err != nil {
			writeJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		c, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			_ = conn.Close()
			writeJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		defer c.Quit()
		if cfg.Username != "" {
			if err := c.Auth(smtp.PlainAuth("", cfg.Username, req.SMTP.Password, cfg.Host)); err != nil {
				writeJSON(w, 400, map[string]any{"error": err.Error()})
				return
			}
		}
		_ = c.Noop()
		writeJSON(w, 200, map[string]any{"ok": true, "message": "Connexion SMTP réussie."})
		return
	}
	client, err := smtp.Dial(net.JoinHostPort(cfg.Host, strconv.Itoa(func() int {
		if cfg.Port == 0 {
			return 587
		}
		return cfg.Port
	}())))
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}
	defer client.Quit()
	if cfg.Security == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			writeJSON(w, 400, map[string]any{"error": "STARTTLS non disponible"})
			return
		}
		if err := client.StartTLS(&tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
			writeJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
	}
	if cfg.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", cfg.Username, req.SMTP.Password, cfg.Host)); err != nil {
			writeJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
	}
	_ = client.Noop()
	writeJSON(w, 200, map[string]any{"ok": true, "message": "Connexion SMTP réussie."})
}

const microsoftClientIDDefault = "59544476-fdef-431c-a3d4-f1f7b4b6ce83"

func microsoftClientID() string {
	if id := strings.TrimSpace(os.Getenv("PHOENIXMAIL_MICROSOFT_CLIENT_ID")); id != "" {
		return id
	}
	return microsoftClientIDDefault
}

func randomB64(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func pkceChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func (a *App) microsoftOAuthStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	accountID := strings.TrimSpace(r.URL.Query().Get("account"))
	if accountID == "" {
		http.Error(w, "account requis", 400)
		return
	}
	a.mu.RLock()
	_, ok := a.accountByIDLocked(accountID)
	a.mu.RUnlock()
	if !ok {
		http.Error(w, "compte introuvable", 404)
		return
	}
	clientID := microsoftClientID()
	if clientID == "" {
		writeJSON(w, 503, map[string]any{"code": "oauth_client_not_configured", "error": "PHOENIXMAIL_MICROSOFT_CLIENT_ID n'est pas configuré."})
		return
	}
	verifier, err := randomB64(32)
	if err != nil {
		http.Error(w, "génération PKCE impossible", 500)
		return
	}
	state, err := randomB64(24)
	if err != nil {
		http.Error(w, "génération state impossible", 500)
		return
	}
	a.oauthMu.Lock()
	a.oauthPending[state] = oauthPending{AccountID: accountID, Verifier: verifier, Created: time.Now()}
	a.oauthMu.Unlock()
	redirect := fmt.Sprintf("http://localhost:%s/oauth/microsoft/callback", getenvDefault("PHOENIXMAIL_PORT", "8787"))
	scopes := "openid profile email offline_access https://outlook.office.com/IMAP.AccessAsUser.All https://outlook.office.com/SMTP.Send"
	u := "https://login.microsoftonline.com/common/oauth2/v2.0/authorize?" + url.Values{
		"client_id": {clientID}, "response_type": {"code"}, "redirect_uri": {redirect}, "response_mode": {"query"},
		"scope": {scopes}, "state": {state}, "code_challenge": {pkceChallenge(verifier)}, "code_challenge_method": {"S256"},
	}.Encode()
	if err := openBrowser(u); err != nil {
		http.Error(w, "Impossible d'ouvrir le navigateur: "+err.Error(), 500)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "message": "Navigateur ouvert. Termine la connexion Microsoft puis reviens dans PhoenixMail."})
}

func (a *App) microsoftOAuthCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	a.oauthMu.Lock()
	pending, ok := a.oauthPending[state]
	delete(a.oauthPending, state)
	a.oauthMu.Unlock()
	if !ok || time.Since(pending.Created) > 10*time.Minute {
		http.Error(w, "session OAuth invalide ou expirée", 400)
		return
	}
	if e := r.URL.Query().Get("error"); e != "" {
		http.Error(w, "Microsoft OAuth: "+e+" "+r.URL.Query().Get("error_description"), 400)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code OAuth manquant", 400)
		return
	}
	clientID := microsoftClientID()
	redirect := fmt.Sprintf("http://localhost:%s/oauth/microsoft/callback", getenvDefault("PHOENIXMAIL_PORT", "8787"))
	form := url.Values{"client_id": {clientID}, "grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {redirect}, "code_verifier": {pending.Verifier}}
	resp, err := http.PostForm("https://login.microsoftonline.com/common/oauth2/v2.0/token", form)
	if err != nil {
		http.Error(w, "échange OAuth impossible: "+err.Error(), 502)
		return
	}
	defer resp.Body.Close()
	var tok struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		ExpiresIn        int    `json:"expires_in"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		http.Error(w, "réponse OAuth invalide", 502)
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || tok.AccessToken == "" {
		http.Error(w, "Microsoft OAuth: "+tok.Error+" "+tok.ErrorDescription, 400)
		return
	}
	a.oauthMu.Lock()
	a.oauthTokens[pending.AccountID] = &oauthToken{AccessToken: tok.AccessToken, RefreshToken: tok.RefreshToken, Expiry: time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)}
	a.oauthMu.Unlock()
	if err := a.saveOAuthTokenCache(); err != nil {
		log.Printf("PhoenixMail: impossible de sauvegarder le cache OAuth2: %v", err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, "<html><body style='font-family:system-ui;background:#10131a;color:#f5f6fa;padding:40px'><h2>PhoenixMail — connexion Microsoft réussie</h2><p>Tu peux fermer cette fenêtre et revenir à PhoenixMail.</p></body></html>")
}

func (a *App) microsoftOAuthStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("account"))
	if id == "" {
		writeJSON(w, 200, map[string]any{"connected": false, "clientConfigured": microsoftClientID() != ""})
		return
	}
	_, err := a.oauthAccessToken(id)
	connected := err == nil
	writeJSON(w, 200, map[string]any{"connected": connected, "clientConfigured": microsoftClientID() != "", "reauthRequired": !connected})
}

func openBrowser(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", u)
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		return errors.New("système non pris en charge")
	}
	return cmd.Start()
}
func getenvDefault(k, v string) string {
	if x := os.Getenv(k); x != "" {
		return x
	}
	return v
}

func (a *App) oauthAccessToken(accountID string) (string, error) {
	a.oauthMu.Lock()
	tok := a.oauthTokens[accountID]
	a.oauthMu.Unlock()
	if tok == nil {
		return "", errors.New("OAuth2: connexion Microsoft requise")
	}
	if time.Now().Before(tok.Expiry.Add(-60 * time.Second)) {
		return tok.AccessToken, nil
	}
	if tok.RefreshToken == "" {
		return "", errors.New("session Microsoft expirée; reconnecte le compte")
	}
	form := url.Values{"client_id": {microsoftClientID()}, "grant_type": {"refresh_token"}, "refresh_token": {tok.RefreshToken}, "scope": {"offline_access https://outlook.office.com/IMAP.AccessAsUser.All https://outlook.office.com/SMTP.Send"}}
	resp, err := http.PostForm("https://login.microsoftonline.com/common/oauth2/v2.0/token", form)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		ExpiresIn        int    `json:"expires_in"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || out.AccessToken == "" {
		return "", errors.New(out.ErrorDescription)
	}
	if out.RefreshToken == "" {
		out.RefreshToken = tok.RefreshToken
	}
	n := &oauthToken{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken, Expiry: time.Now().Add(time.Duration(out.ExpiresIn) * time.Second)}
	a.oauthMu.Lock()
	a.oauthTokens[accountID] = n
	a.oauthMu.Unlock()
	if err := a.saveOAuthTokenCache(); err != nil {
		log.Printf("PhoenixMail: impossible de sauvegarder le cache OAuth2 rafraîchi: %v", err)
	}
	return n.AccessToken, nil
}

func (a *App) sendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct{ To, Cc, Bcc, Subject, Body, DraftID, AccountID string }
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	a.mu.RLock()
	brandFooter := a.data.Config.BrandFooter
	cfg := a.data.Config.SMTP
	password := a.smtpPassword
	if req.AccountID != "" {
		if acc, ok := a.accountByIDLocked(req.AccountID); ok {
			cfg = acc.SMTP
			password = a.accountPasswords[acc.ID]
		}
	} else if a.data.Config.DefaultAccount != "" {
		if acc, ok := a.accountByIDLocked(a.data.Config.DefaultAccount); ok {
			cfg = acc.SMTP
			password = a.accountPasswords[acc.ID]
		}
	}
	a.mu.RUnlock()
	if cfg.Host == "" {
		writeJSON(w, 409, map[string]any{"error": "Aucun serveur SMTP n’est configuré. Ouvrez Paramètres → Comptes."})
		return
	}
	if strings.TrimSpace(cfg.Authentication) == "" && isKnownOAuthSMTPHost(cfg.Host) {
		cfg.Authentication = "OAuth2"
	}
	if strings.EqualFold(strings.TrimSpace(cfg.Authentication), "OAuth2") {
		token, err := a.oauthAccessToken(req.AccountID)
		if err != nil {
			writeJSON(w, 409, map[string]any{"code": "oauth2_required", "userMessage": "Connecte ce compte à Microsoft dans Paramètres → Comptes.", "error": "Échec de l’envoi SMTP: " + err.Error()})
			return
		}
		if err := sendSMTPXOAUTH2(cfg, token, req.To, req.Cc, req.Bcc, req.Subject, req.Body, brandFooter); err != nil {
			writeJSON(w, 502, map[string]any{"code": "smtp_oauth_error", "userMessage": "Échec de l’envoi SMTP OAuth2: " + err.Error(), "error": "Échec de l’envoi SMTP OAuth2: " + err.Error()})
			return
		}
		// Continue below only for local sent-copy persistence.
	} else {
		if cfg.Username != "" && password == "" {
			writeJSON(w, 409, map[string]any{"code": "password_required", "userMessage": "Le mot de passe du compte n’est pas disponible en session. Ouvre Paramètres → Comptes et saisis-le.", "error": "Authentification requise."})
			return
		}
		if err := sendSMTP(cfg, password, req.To, req.Cc, req.Bcc, req.Subject, req.Body, brandFooter); err != nil {
			writeJSON(w, 502, map[string]any{"code": "smtp_error", "userMessage": "Échec de l’envoi SMTP: " + err.Error(), "error": "Échec de l’envoi SMTP: " + err.Error()})
			return
		}
	}
	// Save a local sent copy only after successful SMTP delivery.
	now := time.Now()
	m := Mail{ID: fmt.Sprintf("s-%d", now.UnixNano()), From: "Moi", Email: cfg.From, AccountID: req.AccountID, To: strings.TrimSpace(req.To), Cc: strings.TrimSpace(req.Cc), Bcc: strings.TrimSpace(req.Bcc), Subject: req.Subject, Preview: firstLine(req.Body), Body: req.Body, Time: now.Format("15:04"), Date: "Aujourd’hui", Folder: "sent", Read: true}
	a.mu.Lock()
	if req.DraftID != "" {
		if _, idx := a.find(req.DraftID); idx >= 0 && a.data.Mails[idx].Folder == "drafts" {
			a.data.Mails = append(a.data.Mails[:idx], a.data.Mails[idx+1:]...)
		}
	}
	a.data.Mails = append(a.data.Mails, m)
	err := a.saveLocked()
	a.mu.Unlock()
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "Courriel envoyé, mais copie locale impossible."})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "mode": "smtp", "mail": m})
}

func (a *App) aiStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	a.mu.RLock()
	cfg := a.data.Config.AI
	a.mu.RUnlock()
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8080/v1/chat/completions"
	}
	client := &http.Client{Timeout: 500 * time.Millisecond}
	model := cfg.Model
	if model == "" {
		model = "local"
	}
	available := false
	if ep := strings.TrimRight(endpoint, "/"); strings.HasSuffix(ep, "/chat/completions") {
		probe := strings.TrimSuffix(ep, "/chat/completions") + "/models"
		if resp, err := client.Get(probe); err == nil {
			io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			available = resp.StatusCode < 500
		}
	}
	writeJSON(w, 200, map[string]any{"available": available, "engine": "llama.cpp", "endpoint": endpoint, "model": model})
}

func (a *App) aiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct{ Action, Text, Language string }
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	a.mu.RLock()
	cfg := a.data.Config.AI
	a.mu.RUnlock()
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8080/v1/chat/completions"
	}
	model := cfg.Model
	if model == "" {
		model = "local"
	}
	systemPrompt := map[string]string{"correct": "Corrige uniquement les fautes et la grammaire. Conserve le sens et le ton.", "rewrite": "Reformule de manière claire et professionnelle, sans inventer d’informations.", "translate": "Traduis le message dans la langue demandée, en conservant le sens.", "shorten": "Raccourcis le message en conservant toutes les informations essentielles.", "reply": "Rédige une réponse courte, naturelle et professionnelle au message."}[req.Action]
	if systemPrompt == "" {
		systemPrompt = "Améliore ce texte sans en inventer le contenu."
	}
	userText := req.Text
	if req.Action == "translate" && req.Language != "" {
		userText = "Langue cible: " + req.Language + "\n\n" + req.Text
	}
	payload := map[string]any{"model": model, "messages": []map[string]string{{"role": "system", "content": systemPrompt}, {"role": "user", "content": userText}}, "temperature": 0.2, "max_tokens": 700}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(b))
	if err != nil {
		writeJSON(w, 200, map[string]any{"action": req.Action, "result": "", "engine": "llama.cpp", "available": false, "note": "Moteur local non disponible. Installez un serveur llama.cpp compatible OpenAI sur 127.0.0.1:8080 ou configurez l’endpoint dans Paramètres."})
		return
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		writeJSON(w, 200, map[string]any{"action": req.Action, "result": "", "engine": "llama.cpp", "available": false, "note": "Le moteur local a refusé la requête."})
		return
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if json.Unmarshal(raw, &out) != nil || len(out.Choices) == 0 {
		writeJSON(w, 200, map[string]any{"action": req.Action, "result": "", "engine": "llama.cpp", "available": false, "note": "Réponse IA locale invalide."})
		return
	}
	writeJSON(w, 200, map[string]any{"action": req.Action, "result": strings.TrimSpace(out.Choices[0].Message.Content), "engine": "llama.cpp", "available": true})
}

func localDemoAI(action, text, language string) string {
	t := strings.TrimSpace(text)
	if t == "" {
		return ""
	}
	switch action {
	case "correct":
		return t
	case "rewrite":
		return "Bonjour,\n\n" + t + "\n\nMerci et bonne journée."
	case "translate":
		if strings.EqualFold(language, "en") {
			return "Hello,\n\n" + t + "\n\nBest regards."
		}
		return t
	case "shorten":
		r := []rune(t)
		if len(r) > 180 {
			return string(r[:180]) + "…"
		}
		return t
	case "reply":
		return "Bonjour,\n\nMerci pour votre message. Je vais vérifier cela et revenir vers vous rapidement.\n\nCordialement,"
	}
	return t
}

func main() {
	port := os.Getenv("PHOENIXMAIL_PORT")
	if port == "" {
		port = "8787"
	}
	app := NewApp(dataPath())
	mux := http.NewServeMux()
	mux.HandleFunc("/api/system-appearance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, systemAppearance())
	})
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "app": "PhoenixMail", "version": "0.3.5-responsive-layout"})
	})
	mux.HandleFunc("/api/mails", app.mailsHandler)
	mux.HandleFunc("/api/mails/", app.mailAction)
	mux.HandleFunc("/api/drafts", app.draftHandler)
	mux.HandleFunc("/api/drafts/", app.draftHandler)
	mux.HandleFunc("/api/config", app.configHandler)
	mux.HandleFunc("/api/smtp/test", app.smtpTestHandler)
	mux.HandleFunc("/api/send", app.sendHandler)
	mux.HandleFunc("/api/ai/status", app.aiStatusHandler)
	mux.HandleFunc("/api/ai", app.aiHandler)
	mux.HandleFunc("/api/ai/models", app.aiModelsHandler)
	mux.HandleFunc("/api/ai/download", app.aiDownloadHandler)
	mux.HandleFunc("/api/ai/start", app.aiStartHandler)
	mux.HandleFunc("/api/ai/stop", app.aiStopHandler)
	mux.HandleFunc("/api/autoconfig", app.autoconfigHandler)
	mux.HandleFunc("/api/accounts", app.accountsHandler)
	mux.HandleFunc("/api/accounts/test", app.accountTestHandler)
	mux.HandleFunc("/api/accounts/refresh", app.accountRefreshHandler)
	mux.HandleFunc("/oauth/microsoft/start", app.microsoftOAuthStart)
	mux.HandleFunc("/oauth/microsoft/callback", app.microsoftOAuthCallback)
	mux.HandleFunc("/api/oauth/microsoft/status", app.microsoftOAuthStatus)
	staticHandler := http.FileServer(http.FS(webFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			http.Redirect(w, r, "/static/", http.StatusTemporaryRedirect)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/static/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		staticHandler.ServeHTTP(w, r)
	})
	log.Printf("PhoenixMail v0.3.5 responsive inbox layout listening on http://localhost:%s", port)
	log.Printf("Data: %s", dataPath())
	log.Fatal(http.ListenAndServe(":"+port, logging(mux)))
}
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { next.ServeHTTP(w, r) })
}
