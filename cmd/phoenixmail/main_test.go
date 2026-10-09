package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDemoStoreHasConceptMail(t *testing.T) {
	s := demoStore()
	if len(s.Mails) < 9 {
		t.Fatalf("expected concept inbox messages, got %d", len(s.Mails))
	}
	if s.Mails[0].From != "Marie Dupont" || s.Mails[0].Subject != "Réunion de demain" {
		t.Fatalf("unexpected first demo mail: %+v", s.Mails[0])
	}
}

func TestStorePersistsAndCounts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	a := NewApp(path)
	got := a.list("inbox", "")
	if len(got) != 9 {
		t.Fatalf("expected 9 inbox messages, got %d", len(got))
	}
	if a.counts()["important"] != 2 {
		t.Fatalf("expected 2 important messages, got %d", a.counts()["important"])
	}
	if err := a.save(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("data file was not persisted: %v", err)
	}
}

func TestAccountCRUDAndDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	a := NewApp(path)
	body := `{"name":"Test","email":"test@example.com","displayName":"Test User","setDefault":true,"imap":{"host":"imap.example.com","port":993,"security":"ssl","username":"test@example.com"},"smtp":{"host":"smtp.example.com","port":587,"security":"starttls","username":"test@example.com","from":"test@example.com"}}`
	r := httptest.NewRequest("POST", "/api/accounts", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	a.accountsHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("account create failed: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Account MailAccount `json:"account"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Account.Email != "test@example.com" {
		t.Fatalf("unexpected account: %+v", created.Account)
	}

	w = httptest.NewRecorder()
	a.accountsHandler(w, httptest.NewRequest("GET", "/api/accounts", nil))
	if w.Code != 200 {
		t.Fatalf("account list failed: %d", w.Code)
	}
	var got struct {
		Accounts []MailAccount `json:"accounts"`
		Default  string        `json:"defaultAccount"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Accounts) != 4 || got.Default != created.Account.ID {
		t.Fatalf("unexpected account state: %+v default=%s", got.Accounts, got.Default)
	}
}

func TestAIModelsEndpointHasRecommendedModel(t *testing.T) {
	a := NewApp(filepath.Join(t.TempDir(), "data.json"))
	w := httptest.NewRecorder()
	a.aiModelsHandler(w, httptest.NewRequest("GET", "/api/ai/models", nil))
	if w.Code != 200 {
		t.Fatalf("ai models failed: %d %s", w.Code, w.Body.String())
	}
	var out struct {
		Recommended struct {
			Name    string `json:"name"`
			License string `json:"license"`
		} `json:"recommended"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Recommended.Name != "Qwen3-0.6B-Q4_0.gguf" || out.Recommended.License != "Apache-2.0" {
		t.Fatalf("unexpected recommendation: %+v", out.Recommended)
	}
}

func TestSendRejectsOAuth2WithoutProviderSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	a := NewApp(path)
	a.mu.Lock()
	a.data.Config.Accounts = []MailAccount{{ID: "oauth", Email: "user@outlook.com", DisplayName: "User", SMTP: SMTPConfig{Host: "smtp-mail.outlook.com", Port: 587, Security: "starttls", Username: "user@outlook.com", Authentication: "OAuth2", From: "user@outlook.com"}}}
	a.data.Config.DefaultAccount = "oauth"
	a.mu.Unlock()
	req := httptest.NewRequest("POST", "/api/send", strings.NewReader(`{"to":"test@example.com","subject":"Test","body":"Hello","accountId":"oauth"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	a.sendHandler(w, req)
	if w.Code != 409 {
		t.Fatalf("expected 409 for OAuth2 account without provider session, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "OAuth2") {
		t.Fatalf("expected OAuth2 guidance, got %s", w.Body.String())
	}
}

func TestOutlookHostIsTreatedAsOAuth2(t *testing.T) {
	if !isKnownOAuthSMTPHost("smtp-mail.outlook.com") {
		t.Fatal("expected Outlook SMTP host to require OAuth2")
	}
	if !isKnownOAuthSMTPHost("smtp.office365.com") {
		t.Fatal("expected Microsoft 365 SMTP host to require OAuth2")
	}
}

func TestSendRejectsKnownOAuthHostEvenWithoutSavedAuthFlag(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	a := NewApp(path)
	a.mu.Lock()
	a.data.Config.Accounts = []MailAccount{{ID: "outlook", Email: "user@outlook.com", DisplayName: "User", SMTP: SMTPConfig{Host: "smtp-mail.outlook.com", Port: 587, Security: "starttls", Username: "user@outlook.com", From: "user@outlook.com"}}}
	a.data.Config.DefaultAccount = "outlook"
	a.mu.Unlock()
	req := httptest.NewRequest("POST", "/api/send", strings.NewReader(`{"to":"test@example.com","subject":"Test","body":"Hello","accountId":"outlook"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	a.sendHandler(w, req)
	if w.Code != 409 {
		t.Fatalf("expected 409 for inferred OAuth2 Outlook host, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "OAuth2") {
		t.Fatalf("expected OAuth2 guidance, got %s", w.Body.String())
	}
}

func TestSMTPMessageIncludesPhoenixFooterWhenEnabled(t *testing.T) {
	msg, recipients, err := smtpMessageWithFooter("from@example.com", "Phoenix", "to@example.com", "", "", "Test", "Bonjour", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(recipients) != 1 || recipients[0] != "to@example.com" {
		t.Fatalf("unexpected recipients: %#v", recipients)
	}
	raw := string(msg)
	if !strings.Contains(raw, "Propulsé par PhoenixMail") {
		t.Fatalf("expected PhoenixMail footer in message: %s", raw)
	}
	if !strings.Contains(raw, "Content-Type: text/html") {
		t.Fatal("expected HTML alternative")
	}
}

func TestSMTPMessageOmitsPhoenixFooterWhenDisabled(t *testing.T) {
	msg, _, err := smtpMessageWithFooter("from@example.com", "Phoenix", "to@example.com", "", "", "Test", "Bonjour", false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(msg), "Propulsé par PhoenixMail") {
		t.Fatal("footer should be disabled")
	}
}

func TestOAuthRefreshTokenPersistsAcrossAppRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	a := NewApp(path)
	a.oauthMu.Lock()
	a.oauthTokens["acc"] = &oauthToken{RefreshToken: "refresh-token-test"}
	a.oauthMu.Unlock()
	if err := a.saveOAuthTokenCache(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(a.oauthCachePath())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected OAuth cache mode 0600, got %o", info.Mode().Perm())
	}
	b := NewApp(path)
	b.oauthMu.Lock()
	tok := b.oauthTokens["acc"]
	b.oauthMu.Unlock()
	if tok == nil || tok.RefreshToken != "refresh-token-test" {
		t.Fatalf("OAuth refresh token was not restored: %#v", tok)
	}
}

func TestParseIMAPMailDecodesMIMEHeadersAndBody(t *testing.T) {
	account := MailAccount{ID: "outlook-test", Email: "user@outlook.com"}
	raw := "From: =?UTF-8?Q?Ren=C3=A9?= <rene@example.com>\r\n" +
		"To: user@outlook.com\r\n" +
		"Subject: =?UTF-8?Q?Bonjour_caf=C3=A9?=\r\n" +
		"Date: Fri, 09 Oct 2026 09:15:00 -0400\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
		"Bonjour caf=C3=A9!\r\n"
	m, err := parseIMAPMail(account, 23, `* 1 FETCH (UID 23 FLAGS (\\Seen) BODY[] {1})`, []byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if m.From != "René" || m.Email != "rene@example.com" {
		t.Fatalf("unexpected sender parse: from=%q email=%q", m.From, m.Email)
	}
	if m.Subject != "Bonjour café" || m.Body != "Bonjour café!" {
		t.Fatalf("unexpected decoded message: subject=%q body=%q", m.Subject, m.Body)
	}
	if m.AccountID != account.ID || m.ID != "imap-outlook-test-0000000023" || m.Folder != "inbox" || !m.Read {
		t.Fatalf("unexpected message metadata: %+v", m)
	}
}

func TestAccountRefreshIMAPEndToEndWithLocalServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	message := "From: Test Sender <sender@example.com>\r\n" +
		"To: user@example.com\r\n" +
		"Subject: First real sync\r\n" +
		"Date: Fri, 09 Oct 2026 09:15:00 -0400\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		"Message fetched by IMAP.\r\n"
	serverErr := make(chan error, 2)
	go func() {
		for session := 0; session < 2; session++ {
			conn, err := listener.Accept()
			if err != nil {
				serverErr <- err
				return
			}
			func() {
				defer conn.Close()
				reader := bufio.NewReader(conn)
				if _, err := fmt.Fprint(conn, "* OK local test server ready\r\n"); err != nil {
					serverErr <- err
					return
				}
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if errors.Is(err, io.EOF) {
							serverErr <- nil
						} else {
							serverErr <- err
						}
						return
					}
					line = strings.TrimSpace(line)
					space := strings.IndexByte(line, ' ')
					if space < 0 {
						serverErr <- fmt.Errorf("invalid IMAP test command %q", line)
						return
					}
					tag, command := line[:space], line[space+1:]
					switch {
					case strings.HasPrefix(strings.ToUpper(command), "LOGIN "):
						_, err = fmt.Fprintf(conn, "%s OK logged in\r\n", tag)
					case strings.EqualFold(command, "SELECT INBOX"):
						_, err = fmt.Fprintf(conn, "* 1 EXISTS\r\n* OK [UIDVALIDITY 7] selected\r\n%s OK selected\r\n", tag)
					case strings.EqualFold(command, "UID SEARCH ALL"):
						_, err = fmt.Fprintf(conn, "* SEARCH 23\r\n%s OK search done\r\n", tag)
					case strings.HasPrefix(strings.ToUpper(command), "UID FETCH "):
						_, err = fmt.Fprintf(conn, "* 1 FETCH (UID 23 FLAGS (\\Seen) BODY[] {%d}\r\n", len(message))
						if err == nil {
							_, err = io.WriteString(conn, message)
						}
						if err == nil {
							_, err = fmt.Fprintf(conn, ")\r\n%s OK fetch done\r\n", tag)
						}
					default:
						err = fmt.Errorf("unexpected IMAP command %q", command)
					}
					if err != nil {
						serverErr <- err
						return
					}
				}
			}()
		}
	}()

	path := filepath.Join(t.TempDir(), "data.json")
	a := NewApp(path)
	a.mu.Lock()
	a.data.Config.Accounts = []MailAccount{{ID: "test-account", Email: "user@example.com", IMAP: IMAPConfig{Host: "127.0.0.1", Port: port, Security: "plain", Username: "user@example.com"}}}
	a.data.Config.DefaultAccount = "test-account"
	a.accountPasswords["test-account:imap"] = "test-password"
	a.mu.Unlock()
	result, err := a.refreshAccountInbox("test-account")
	if err != nil {
		t.Fatal(err)
	}
	if result["newMessages"] != 1 || result["fetched"] != 1 {
		t.Fatalf("unexpected refresh response: %#v", result)
	}
	mails := a.listForAccount("inbox", "", "test-account")
	if len(mails) != 1 || mails[0].Subject != "First real sync" || mails[0].Body != "Message fetched by IMAP." {
		t.Fatalf("unexpected synced mails: %#v", mails)
	}

	// Re-fetching the same IMAP UID must refresh the existing local row without
	// increasing the mailbox count or reporting a duplicate as a new message.
	second, err := a.refreshAccountInbox("test-account")
	if err != nil {
		t.Fatal(err)
	}
	if second["newMessages"] != 0 || second["fetched"] != 1 {
		t.Fatalf("repeated sync should report zero new and one fetched message, got %#v", second)
	}
	mails = a.listForAccount("inbox", "", "test-account")
	if len(mails) != 1 {
		t.Fatalf("repeated sync created a duplicate; mailbox contains %d messages", len(mails))
	}
	for i := 0; i < 2; i++ {
		if err := <-serverErr; err != nil && !strings.Contains(err.Error(), "closed network connection") {
			t.Fatalf("fake IMAP server failed: %v", err)
		}
	}
}

func TestListForAccountSeparatesMessagesAndCounts(t *testing.T) {
	a := NewApp(filepath.Join(t.TempDir(), "data.json"))
	a.mu.Lock()
	a.data.Mails = []Mail{
		{ID: "a", AccountID: "one", Folder: "inbox", Read: false},
		{ID: "b", AccountID: "two", Folder: "inbox", Read: true},
	}
	a.mu.Unlock()
	if got := a.listForAccount("inbox", "", "one"); len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("unexpected account-scoped list: %#v", got)
	}
	counts := a.countsForAccount("one")
	if counts["inbox"] != 1 || counts["unread"] != 1 {
		t.Fatalf("unexpected account-scoped counts: %#v", counts)
	}
}
