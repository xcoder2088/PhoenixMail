package main

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const qwenModelName = "Qwen3-0.6B-Q4_0.gguf"
const qwenModelURL = "https://huggingface.co/ggml-org/Qwen3-0.6B-GGUF/resolve/main/Qwen3-0.6B-Q4_0.gguf"
const qwenModelSHA256 = "da2572f16c06133561ce56accaa822216f2391ef4d37fba427801cd6736417d4"

func (a *App) ensureAccountsLocked() {
	if a.data.Config.Accounts == nil {
		a.data.Config.Accounts = []MailAccount{}
	}
	if len(a.data.Config.Accounts) == 0 && a.data.Config.SMTP.Host != "" {
		a.data.Config.Accounts = append(a.data.Config.Accounts, MailAccount{ID: "default", Name: a.data.Config.SMTP.DisplayName, Email: a.data.Config.SMTP.From, DisplayName: a.data.Config.SMTP.DisplayName, SMTP: a.data.Config.SMTP})
	}
	if a.data.Config.DefaultAccount == "" && len(a.data.Config.Accounts) > 0 {
		a.data.Config.DefaultAccount = a.data.Config.Accounts[0].ID
	}
}

func (a *App) accountByIDLocked(id string) (MailAccount, bool) {
	for _, acc := range a.data.Config.Accounts {
		if acc.ID == id {
			return acc, true
		}
	}
	return MailAccount{}, false
}

func publicAccounts(accounts []MailAccount) []map[string]any {
	out := make([]map[string]any, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, map[string]any{
			"id": a.ID, "name": a.Name, "email": a.Email, "displayName": a.DisplayName,
			"imap": a.IMAP, "smtp": a.SMTP,
		})
	}
	return out
}

func (a *App) accountsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.mu.RLock()
		cfg := a.data.Config
		a.mu.RUnlock()
		writeJSON(w, http.StatusOK, map[string]any{"accounts": publicAccounts(cfg.Accounts), "defaultAccount": cfg.DefaultAccount})
		return
	case http.MethodPost:
		var req struct {
			ID          string     `json:"id"`
			Name        string     `json:"name"`
			Email       string     `json:"email"`
			DisplayName string     `json:"displayName"`
			IMAP        IMAPConfig `json:"imap"`
			SMTP        struct {
				Host           string `json:"host"`
				Port           int    `json:"port"`
				Security       string `json:"security"`
				Authentication string `json:"authentication"`
				Username       string `json:"username"`
				From           string `json:"from"`
				DisplayName    string `json:"displayName"`
				Password       string `json:"password"`
			} `json:"smtp"`
			IMAPPassword string `json:"imapPassword"`
			SetDefault   bool   `json:"setDefault"`
		}
		if err := decodeJSON(r, &req); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}
		id := strings.TrimSpace(req.ID)
		if id == "" {
			id = fmt.Sprintf("acc-%d", time.Now().UnixNano())
		}
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
		smtpCfg := SMTPConfig{Host: strings.TrimSpace(req.SMTP.Host), Port: port, Security: sec, Authentication: strings.TrimSpace(req.SMTP.Authentication), Username: strings.TrimSpace(req.SMTP.Username), From: strings.TrimSpace(req.SMTP.From), DisplayName: strings.TrimSpace(req.SMTP.DisplayName)}
		if smtpCfg.From == "" {
			smtpCfg.From = strings.TrimSpace(req.Email)
		}
		if smtpCfg.Authentication == "" && isKnownOAuthSMTPHost(smtpCfg.Host) {
			smtpCfg.Authentication = "OAuth2"
		}
		imap := req.IMAP
		imap.Security = strings.ToLower(strings.TrimSpace(imap.Security))
		if imap.Security == "" {
			imap.Security = "ssl"
		}
		if imap.Port == 0 {
			if imap.Security == "ssl" {
				imap.Port = 993
			} else {
				imap.Port = 143
			}
		}
		if imap.Username == "" {
			imap.Username = strings.TrimSpace(req.Email)
		}
		acc := MailAccount{ID: id, Name: strings.TrimSpace(req.Name), Email: strings.TrimSpace(req.Email), DisplayName: strings.TrimSpace(req.DisplayName), IMAP: imap, SMTP: smtpCfg}
		if acc.DisplayName == "" {
			acc.DisplayName = acc.Name
		}
		if acc.Email == "" {
			http.Error(w, "adresse courriel requise", 400)
			return
		}
		a.mu.Lock()
		a.ensureAccountsLocked()
		replaced := false
		for i := range a.data.Config.Accounts {
			if a.data.Config.Accounts[i].ID == id {
				a.data.Config.Accounts[i] = acc
				replaced = true
				break
			}
		}
		if !replaced {
			a.data.Config.Accounts = append(a.data.Config.Accounts, acc)
		}
		if req.SetDefault || a.data.Config.DefaultAccount == "" {
			a.data.Config.DefaultAccount = id
		}
		if req.SMTP.Password != "" {
			a.accountPasswords[id] = req.SMTP.Password
		}
		if req.IMAPPassword != "" {
			a.accountPasswords[id+":imap"] = req.IMAPPassword
		} else if req.SMTP.Password != "" {
			// The current account editor uses one password field for providers
			// that share credentials between IMAP and SMTP.
			a.accountPasswords[id+":imap"] = req.SMTP.Password
		}
		err := a.saveLocked()
		a.mu.Unlock()
		if err != nil {
			http.Error(w, "save failed", 500)
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "account": acc})
		return
	case http.MethodDelete:
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		if id == "" {
			http.Error(w, "id required", 400)
			return
		}
		a.mu.Lock()
		found := false
		for i := range a.data.Config.Accounts {
			if a.data.Config.Accounts[i].ID == id {
				a.data.Config.Accounts = append(a.data.Config.Accounts[:i], a.data.Config.Accounts[i+1:]...)
				found = true
				break
			}
		}
		delete(a.accountPasswords, id)
		delete(a.accountPasswords, id+":imap")
		if a.data.Config.DefaultAccount == id {
			a.data.Config.DefaultAccount = ""
			if len(a.data.Config.Accounts) > 0 {
				a.data.Config.DefaultAccount = a.data.Config.Accounts[0].ID
			}
		}
		err := a.saveLocked()
		a.mu.Unlock()
		if !found {
			http.Error(w, "account not found", 404)
			return
		}
		if err != nil {
			http.Error(w, "save failed", 500)
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})
		return
	default:
		http.Error(w, "method not allowed", 405)
	}
}

func imapReadLine(r *bufio.Reader) (string, error) {
	s, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(s, "\r\n"), nil
}
func imapCommand(conn net.Conn, r *bufio.Reader, tag, cmd string) error {
	if _, err := fmt.Fprintf(conn, "%s %s\r\n", tag, cmd); err != nil {
		return err
	}
	for {
		line, err := imapReadLine(r)
		if err != nil {
			return err
		}
		if strings.HasPrefix(line, tag+" ") {
			if strings.Contains(line, " NO ") || strings.Contains(line, " BAD ") {
				return errors.New(line)
			}
			return nil
		}
	}
}
func testIMAP(cfg IMAPConfig, password string) error {
	if cfg.Host == "" {
		return errors.New("serveur IMAP non configuré")
	}
	port := cfg.Port
	if port == 0 {
		if cfg.Security == "ssl" {
			port = 993
		} else {
			port = 143
		}
	}
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(port))
	var conn net.Conn
	var err error
	if cfg.Security == "ssl" || port == 993 {
		conn, err = tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
	} else {
		conn, err = net.DialTimeout("tcp", addr, 10*time.Second)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	r := bufio.NewReader(conn)
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
	if _, err := imapReadLine(r); err != nil {
		return err
	}
	if cfg.Security == "starttls" {
		if err := imapCommand(conn, r, "a001", "STARTTLS"); err != nil {
			return err
		}
		tc := tls.Client(conn, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
		if err := tc.Handshake(); err != nil {
			return err
		}
		conn = tc
		r = bufio.NewReader(conn)
	}
	user := strings.ReplaceAll(cfg.Username, "\\", "\\\\")
	user = strings.ReplaceAll(user, "\"", "\\\"")
	pass := strings.ReplaceAll(password, "\\", "\\\\")
	pass = strings.ReplaceAll(pass, "\"", "\\\"")
	if err := imapCommand(conn, r, "a002", fmt.Sprintf("LOGIN \"%s\" \"%s\"", user, pass)); err != nil {
		return err
	}
	return imapCommand(conn, r, "a003", "NOOP")
}

func testSMTPXOAUTH2(cfg SMTPConfig, token string) error {
	if cfg.Host == "" {
		return errors.New("serveur SMTP non configuré")
	}
	if token == "" {
		return errors.New("jeton OAuth2 manquant")
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
	var err error
	if cfg.Security == "ssl" || port == 465 {
		conn, e := tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
		if e != nil {
			return e
		}
		client, err = smtp.NewClient(conn, cfg.Host)
	} else {
		client, err = smtp.Dial(addr)
	}
	if err != nil {
		return err
	}
	defer client.Quit()
	if cfg.Security == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return errors.New("STARTTLS non disponible")
		}
		if err := client.StartTLS(&tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if ok, _ := client.Extension("AUTH"); !ok {
		return errors.New("SMTP AUTH indisponible")
	}
	if err := client.Auth(&xoauth2Auth{username: cfg.Username, token: token}); err != nil {
		return err
	}
	return client.Noop()
}

func testIMAPXOAUTH2(cfg IMAPConfig, username, token string) error {
	if cfg.Host == "" {
		return errors.New("serveur IMAP non configuré")
	}
	port := cfg.Port
	if port == 0 {
		if cfg.Security == "ssl" {
			port = 993
		} else {
			port = 143
		}
	}
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(port))
	var conn net.Conn
	var err error
	if cfg.Security == "ssl" || port == 993 {
		conn, err = tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
	} else {
		conn, err = net.DialTimeout("tcp", addr, 10*time.Second)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
	r := bufio.NewReader(conn)
	if _, err := imapReadLine(r); err != nil {
		return err
	}
	if cfg.Security == "starttls" {
		if err := imapCommand(conn, r, "a001", "STARTTLS"); err != nil {
			return err
		}
		tc := tls.Client(conn, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
		if err := tc.Handshake(); err != nil {
			return err
		}
		conn = tc
		r = bufio.NewReader(conn)
	}
	if username == "" {
		username = cfg.Username
	}
	if username == "" || token == "" {
		return errors.New("identifiants OAuth2 manquants")
	}
	ir := base64.StdEncoding.EncodeToString([]byte("user=" + username + "\x01auth=Bearer " + token + "\x01\x01"))
	if _, err := fmt.Fprintf(conn, "a002 AUTHENTICATE XOAUTH2 %s\r\n", ir); err != nil {
		return err
	}
	for {
		line, err := imapReadLine(r)
		if err != nil {
			return err
		}
		if strings.HasPrefix(line, "a002 ") {
			if strings.Contains(line, " NO ") || strings.Contains(line, " BAD ") {
				return errors.New(line)
			}
			break
		}
		if strings.HasPrefix(line, "+") {
			// XOAUTH2 normally succeeds with the SASL initial response. If the
			// server asks for another response, cancel the exchange rather than
			// accidentally sending credentials in clear text.
			if _, err := fmt.Fprint(conn, "*\r\n"); err != nil {
				return err
			}
			return errors.New("le serveur IMAP a demandé une réponse XOAUTH2 supplémentaire")
		}
	}
	return imapCommand(conn, r, "a003", "NOOP")
}

func (a *App) accountTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		AccountID    string `json:"accountId"`
		IMAP         bool   `json:"imap"`
		SMTP         bool   `json:"smtp"`
		Password     string `json:"password"`
		IMAPPassword string `json:"imapPassword"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}
	a.mu.RLock()
	acc, ok := a.accountByIDLocked(req.AccountID)
	pw := a.accountPasswords[req.AccountID]
	ipw := a.accountPasswords[req.AccountID+":imap"]
	a.mu.RUnlock()
	if !ok {
		http.Error(w, "account not found", 404)
		return
	}
	if req.Password != "" {
		pw = req.Password
	}
	if req.IMAPPassword != "" {
		ipw = req.IMAPPassword
	}
	result := map[string]any{"ok": true}
	if req.SMTP {
		if strings.EqualFold(strings.TrimSpace(acc.SMTP.Authentication), "OAuth2") || isKnownOAuthSMTPHost(acc.SMTP.Host) {
			token, err := a.oauthAccessToken(req.AccountID)
			if err != nil {
				result["smtp"] = "OAuth2: " + err.Error()
				result["ok"] = false
			} else if err := testSMTPXOAUTH2(acc.SMTP, token); err != nil {
				result["smtp"] = err.Error()
				result["ok"] = false
			} else {
				result["smtp"] = "Connexion SMTP OAuth2 réussie."
			}
		} else if err := sendSMTPTest(acc.SMTP, pw); err != nil {
			result["smtp"] = err.Error()
			result["ok"] = false
		} else {
			result["smtp"] = "Connexion SMTP réussie."
		}
	}
	if req.IMAP {
		if strings.EqualFold(strings.TrimSpace(acc.IMAP.Authentication), "OAuth2") || strings.EqualFold(strings.TrimSpace(acc.SMTP.Authentication), "OAuth2") || isKnownOAuthSMTPHost(acc.SMTP.Host) {
			token, err := a.oauthAccessToken(req.AccountID)
			if err != nil {
				result["imap"] = "OAuth2: " + err.Error()
				result["ok"] = false
			} else if err := testIMAPXOAUTH2(acc.IMAP, acc.IMAP.Username, token); err != nil {
				result["imap"] = err.Error()
				result["ok"] = false
			} else {
				result["imap"] = "Connexion IMAP OAuth2 réussie."
			}
		} else if err := testIMAP(acc.IMAP, ipw); err != nil {
			result["imap"] = err.Error()
			result["ok"] = false
		} else {
			result["imap"] = "Connexion IMAP réussie."
		}
	}
	writeJSON(w, 200, result)
}

func isKnownOAuthSMTPHost(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	return h == "smtp-mail.outlook.com" || h == "smtp.office365.com" || h == "smtp.gmail.com" || strings.HasSuffix(h, ".smtp.mail.yahoo.com") || h == "smtp.mail.yahoo.com"
}

func sendSMTPTest(cfg SMTPConfig, password string) error {
	if strings.TrimSpace(cfg.Authentication) == "" && isKnownOAuthSMTPHost(cfg.Host) {
		cfg.Authentication = "OAuth2"
	}
	if strings.EqualFold(strings.TrimSpace(cfg.Authentication), "OAuth2") {
		return errors.New("ce compte utilise OAuth2; une connexion au fournisseur est nécessaire")
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
	var err error
	if cfg.Security == "ssl" || port == 465 {
		conn, e := tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
		if e != nil {
			return e
		}
		client, err = smtp.NewClient(conn, cfg.Host)
	} else {
		client, err = smtp.Dial(addr)
		if err == nil && cfg.Security == "starttls" {
			if ok, _ := client.Extension("STARTTLS"); !ok {
				return errors.New("STARTTLS non disponible")
			}
			err = client.StartTLS(&tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
		}
	}
	if err != nil {
		return err
	}
	defer client.Quit()
	if cfg.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", cfg.Username, password, cfg.Host)); err != nil {
			return err
		}
	}
	return client.Noop()
}

func (a *App) modelDir() string { return filepath.Join(filepath.Dir(a.path), "models") }
func (a *App) aiModelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	_ = os.MkdirAll(a.modelDir(), 0o755)
	entries, _ := os.ReadDir(a.modelDir())
	models := []map[string]any{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".gguf") {
			continue
		}
		info, _ := e.Info()
		models = append(models, map[string]any{"name": e.Name(), "path": filepath.Join(a.modelDir(), e.Name()), "size": info.Size()})
	}
	a.mu.RLock()
	st := a.aiDownloadState
	er := a.aiDownloadError
	b := a.aiDownloadBytes
	t := a.aiDownloadTotal
	cfg := a.data.Config.AI
	a.mu.RUnlock()
	writeJSON(w, 200, map[string]any{"models": models, "recommended": map[string]any{"name": qwenModelName, "size": 429 * 1024 * 1024, "sha256": qwenModelSHA256, "license": "Apache-2.0", "source": qwenModelURL}, "download": map[string]any{"state": st, "error": er, "bytes": b, "total": t}, "ai": cfg})
}

func (a *App) aiDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	a.mu.Lock()
	if a.aiDownloadState == "downloading" {
		a.mu.Unlock()
		writeJSON(w, 409, map[string]any{"error": "download already in progress"})
		return
	}
	a.aiDownloadState = "downloading"
	a.aiDownloadError = ""
	a.aiDownloadBytes = 0
	a.aiDownloadTotal = 0
	a.mu.Unlock()
	go func() {
		_ = os.MkdirAll(a.modelDir(), 0o755)
		tmp := filepath.Join(a.modelDir(), qwenModelName+".part")
		dst := filepath.Join(a.modelDir(), qwenModelName)
		req, err := http.NewRequest(http.MethodGet, qwenModelURL, nil)
		if err == nil {
			client := &http.Client{Timeout: 30 * time.Minute}
			var resp *http.Response
			resp, err = client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					err = fmt.Errorf("téléchargement HTTP %s", resp.Status)
				} else {
					a.mu.Lock()
					a.aiDownloadTotal = resp.ContentLength
					a.mu.Unlock()
					f, e := os.Create(tmp)
					if e != nil {
						err = e
					} else {
						bw := &progressWriter{app: a}
						_, err = io.Copy(io.MultiWriter(f, bw), resp.Body)
						cerr := f.Close()
						if err == nil {
							err = cerr
						}
					}
				}
			}
		}
		if err == nil {
			f, e := os.Open(tmp)
			if e != nil {
				err = e
			} else {
				h := sha256.New()
				_, err = io.Copy(h, f)
				_ = f.Close()
				if err == nil {
					got := hex.EncodeToString(h.Sum(nil))
					if got != qwenModelSHA256 {
						err = fmt.Errorf("SHA-256 invalide: %s", got)
					}
				}
			}
		}
		if err == nil {
			err = os.Rename(tmp, dst)
		} else {
			_ = os.Remove(tmp)
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		if err != nil {
			a.aiDownloadState = "error"
			a.aiDownloadError = err.Error()
		} else {
			a.aiDownloadState = "ready"
			a.data.Config.AI.ModelPath = dst
			a.data.Config.AI.Model = qwenModelName
			_ = a.saveLocked()
		}
	}()
	writeJSON(w, 202, map[string]any{"ok": true})
}

type progressWriter struct{ app *App }

func (p *progressWriter) Write(b []byte) (int, error) {
	p.app.mu.Lock()
	p.app.aiDownloadBytes += int64(len(b))
	p.app.mu.Unlock()
	return len(b), nil
}

func (a *App) aiStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	a.mu.Lock()
	if a.aiCmd != nil && a.aiCmd.Process != nil {
		a.mu.Unlock()
		writeJSON(w, 409, map[string]any{"error": "moteur déjà démarré"})
		return
	}
	model := a.data.Config.AI.ModelPath
	a.mu.Unlock()
	if model == "" {
		http.Error(w, "Aucun modèle local sélectionné.", 409)
		return
	}
	if _, err := os.Stat(model); err != nil {
		http.Error(w, "Modèle introuvable: "+err.Error(), 409)
		return
	}
	bin, err := exec.LookPath("llama-server")
	if err != nil {
		http.Error(w, "llama-server n’est pas installé ou absent du PATH.", 409)
		return
	}
	endpointPort := "8080"
	cmd := exec.Command(bin, "-m", model, "--host", "127.0.0.1", "--port", endpointPort)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	a.mu.Lock()
	a.aiCmd = cmd
	a.data.Config.AI.Endpoint = "http://127.0.0.1:8080/v1/chat/completions"
	a.data.Config.AI.Model = qwenModelName
	_ = a.saveLocked()
	a.mu.Unlock()
	go func() { _ = cmd.Wait(); a.mu.Lock(); a.aiCmd = nil; a.mu.Unlock() }()
	writeJSON(w, 200, map[string]any{"ok": true, "endpoint": a.data.Config.AI.Endpoint})
}
func (a *App) aiStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	a.mu.Lock()
	cmd := a.aiCmd
	a.aiCmd = nil
	a.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		writeJSON(w, 409, map[string]any{"error": "moteur non démarré"})
		return
	}
	_ = cmd.Process.Kill()
	writeJSON(w, 200, map[string]any{"ok": true})
}
