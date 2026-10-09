package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/http"
	netmail "net/mail"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	imapSyncLimit       = 40
	imapMaxMessageBytes = 16 << 20
)

type imapSyncResponse struct {
	line     string
	literal  []byte
	tooLarge bool
}

type imapSyncSession struct {
	conn net.Conn
	r    *bufio.Reader
	tag  uint64
}

func (s *imapSyncSession) nextTag() string {
	s.tag++
	return fmt.Sprintf("pm%04d", s.tag)
}

func (s *imapSyncSession) readLine() (string, error) {
	if err := s.conn.SetDeadline(time.Now().Add(20 * time.Second)); err != nil {
		return "", err
	}
	line, err := s.r.ReadString('\n')
	return strings.TrimRight(line, "\r\n"), err
}

func imapLiteralLength(line string) (int64, bool) {
	end := strings.LastIndex(line, "}")
	if end < 0 || end != len(line)-1 {
		return 0, false
	}
	start := strings.LastIndex(line[:end], "{")
	if start < 0 {
		return 0, false
	}
	n := strings.TrimSuffix(line[start+1:end], "+")
	v, err := strconv.ParseInt(n, 10, 64)
	return v, err == nil && v >= 0
}

// command reads tagged IMAP replies and consumes any RFC 3501 literals safely.
func (s *imapSyncSession) command(command string) ([]imapSyncResponse, error) {
	tag := s.nextTag()
	if err := s.conn.SetDeadline(time.Now().Add(20 * time.Second)); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(s.conn, "%s %s\r\n", tag, command); err != nil {
		return nil, err
	}
	var responses []imapSyncResponse
	for {
		line, err := s.readLine()
		if err != nil {
			return responses, err
		}
		if strings.HasPrefix(line, tag+" ") {
			upper := strings.ToUpper(line)
			if strings.Contains(upper, " NO ") || strings.Contains(upper, " BAD ") || strings.HasSuffix(upper, " NO") || strings.HasSuffix(upper, " BAD") {
				return responses, errors.New(line)
			}
			return responses, nil
		}
		if strings.HasPrefix(line, "+") {
			return responses, fmt.Errorf("unexpected IMAP continuation: %s", line)
		}
		resp := imapSyncResponse{line: line}
		if length, ok := imapLiteralLength(line); ok {
			if length > imapMaxMessageBytes {
				if _, err := io.CopyN(io.Discard, s.r, length); err != nil {
					return responses, err
				}
				resp.tooLarge = true
			} else {
				resp.literal = make([]byte, int(length))
				if _, err := io.ReadFull(s.r, resp.literal); err != nil {
					return responses, err
				}
			}
			suffix, err := s.r.ReadString('\n')
			if err != nil {
				return responses, err
			}
			resp.line += strings.TrimRight(suffix, "\r\n")
		}
		responses = append(responses, resp)
	}
}

func (s *imapSyncSession) authenticateXOAUTH2(username, token string) error {
	tag := s.nextTag()
	if err := s.conn.SetDeadline(time.Now().Add(20 * time.Second)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(s.conn, "%s AUTHENTICATE XOAUTH2\r\n", tag); err != nil {
		return err
	}
	response := base64.StdEncoding.EncodeToString([]byte("user=" + username + "\x01auth=Bearer " + token + "\x01\x01"))
	sentResponse := false
	sentEmpty := false
	for {
		line, err := s.readLine()
		if err != nil {
			return err
		}
		if strings.HasPrefix(line, tag+" ") {
			upper := strings.ToUpper(line)
			if strings.Contains(upper, " NO ") || strings.Contains(upper, " BAD ") || strings.HasSuffix(upper, " NO") || strings.HasSuffix(upper, " BAD") {
				return errors.New(line)
			}
			return nil
		}
		if strings.HasPrefix(line, "+") {
			if !sentResponse {
				if _, err := fmt.Fprintf(s.conn, "%s\r\n", response); err != nil {
					return err
				}
				sentResponse = true
				continue
			}
			if !sentEmpty {
				if _, err := fmt.Fprint(s.conn, "\r\n"); err != nil {
					return err
				}
				sentEmpty = true
				continue
			}
			return fmt.Errorf("unexpected IMAP XOAUTH2 continuation: %s", line)
		}
	}
}

func quoteIMAPString(value string) (string, error) {
	if strings.ContainsAny(value, "\r\n\x00") {
		return "", errors.New("valeur de connexion IMAP invalide")
	}
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return "\"" + value + "\"", nil
}

func dialIMAP(cfg IMAPConfig) (*imapSyncSession, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, errors.New("serveur IMAP non configuré pour ce compte")
	}
	port := cfg.Port
	if port == 0 {
		if strings.EqualFold(cfg.Security, "ssl") {
			port = 993
		} else {
			port = 143
		}
	}
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(port))
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("connexion IMAP impossible : %w", err)
	}
	if strings.EqualFold(cfg.Security, "ssl") || port == 993 {
		_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
		tc := tls.Client(conn, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
		if err := tc.Handshake(); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("négociation TLS IMAP impossible : %w", err)
		}
		conn = tc
	}
	s := &imapSyncSession{conn: conn, r: bufio.NewReader(conn)}
	greeting, err := s.readLine()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("aucune réponse du serveur IMAP : %w", err)
	}
	if !strings.HasPrefix(strings.ToUpper(greeting), "* OK") && !strings.HasPrefix(strings.ToUpper(greeting), "* PREAUTH") {
		_ = conn.Close()
		return nil, fmt.Errorf("salutation IMAP inattendue : %s", greeting)
	}
	if strings.EqualFold(cfg.Security, "starttls") {
		if _, err := s.command("STARTTLS"); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("STARTTLS IMAP impossible : %w", err)
		}
		tc := tls.Client(conn, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
		if err := tc.Handshake(); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("négociation STARTTLS IMAP impossible : %w", err)
		}
		s.conn = tc
		s.r = bufio.NewReader(tc)
	}
	return s, nil
}

func (a *App) refreshAccountInbox(accountID string) (map[string]any, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return nil, errors.New("sélectionne un compte à actualiser")
	}
	a.mu.RLock()
	acc, ok := a.accountByIDLocked(accountID)
	password := a.accountPasswords[accountID+":imap"]
	if password == "" {
		password = a.accountPasswords[accountID]
	}
	a.mu.RUnlock()
	if !ok {
		return nil, errors.New("compte introuvable")
	}

	imapCfg := acc.IMAP
	if strings.TrimSpace(imapCfg.Username) == "" {
		imapCfg.Username = acc.Email
	}
	s, err := dialIMAP(imapCfg)
	if err != nil {
		return nil, err
	}
	defer s.conn.Close()

	useOAuth := strings.EqualFold(strings.TrimSpace(imapCfg.Authentication), "OAuth2") || strings.EqualFold(strings.TrimSpace(acc.SMTP.Authentication), "OAuth2") || strings.Contains(strings.ToLower(acc.IMAP.Host), "outlook.office365.com") || strings.Contains(strings.ToLower(acc.SMTP.Host), "outlook") || strings.Contains(strings.ToLower(acc.SMTP.Host), "office365")
	if useOAuth {
		isMicrosoft := strings.Contains(strings.ToLower(imapCfg.Host), "outlook") || strings.Contains(strings.ToLower(imapCfg.Host), "office365") || strings.Contains(strings.ToLower(acc.SMTP.Host), "outlook") || strings.Contains(strings.ToLower(acc.SMTP.Host), "office365")
		if !isMicrosoft {
			return nil, errors.New("OAuth2 IMAP est configuré, mais ce fournisseur n’a pas encore de connecteur OAuth2 dans PhoenixMail")
		}
		token, err := a.oauthAccessToken(accountID)
		if err != nil {
			return nil, fmt.Errorf("connexion Microsoft requise : %w", err)
		}
		if err := s.authenticateXOAUTH2(imapCfg.Username, token); err != nil {
			return nil, fmt.Errorf("authentification IMAP Microsoft refusée : %w", err)
		}
	} else {
		if password == "" {
			return nil, errors.New("mot de passe IMAP absent de la session. Ouvre Paramètres → Comptes, saisis le mot de passe et enregistre le compte")
		}
		user, err := quoteIMAPString(imapCfg.Username)
		if err != nil {
			return nil, err
		}
		pass, err := quoteIMAPString(password)
		if err != nil {
			return nil, err
		}
		if _, err := s.command("LOGIN " + user + " " + pass); err != nil {
			return nil, fmt.Errorf("authentification IMAP refusée : %w", err)
		}
	}
	if _, err := s.command("SELECT INBOX"); err != nil {
		return nil, fmt.Errorf("ouverture de la boîte de réception impossible : %w", err)
	}
	search, err := s.command("UID SEARCH ALL")
	if err != nil {
		return nil, fmt.Errorf("recherche des messages IMAP impossible : %w", err)
	}
	var uids []uint32
	for _, resp := range search {
		if strings.HasPrefix(strings.ToUpper(resp.line), "* SEARCH") {
			for _, field := range strings.Fields(resp.line)[2:] {
				uid64, parseErr := strconv.ParseUint(field, 10, 32)
				if parseErr == nil && uid64 > 0 {
					uids = append(uids, uint32(uid64))
				}
			}
		}
	}
	if len(uids) > imapSyncLimit {
		uids = uids[len(uids)-imapSyncLimit:]
	}

	fetched := make([]Mail, 0, len(uids))
	for _, uid := range uids {
		responses, err := s.command(fmt.Sprintf("UID FETCH %d (UID FLAGS BODY.PEEK[])", uid))
		if err != nil {
			return nil, fmt.Errorf("lecture du message IMAP %d impossible : %w", uid, err)
		}
		for _, resp := range responses {
			if resp.tooLarge || len(resp.literal) == 0 || !strings.Contains(strings.ToUpper(resp.line), " FETCH ") {
				continue
			}
			if fetchUID(resp.line) != uid {
				continue
			}
			m, err := parseIMAPMail(acc, uid, resp.line, resp.literal)
			if err != nil {
				return nil, fmt.Errorf("analyse du message IMAP %d impossible : %w", uid, err)
			}
			fetched = append(fetched, m)
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	existing := make(map[string]int, len(a.data.Mails))
	for i := range a.data.Mails {
		existing[a.data.Mails[i].ID] = i
	}
	newCount := 0
	for _, m := range fetched {
		if idx, found := existing[m.ID]; found {
			old := a.data.Mails[idx]
			m.Important = old.Important
			if old.Folder != "inbox" {
				m.Folder = old.Folder
			}
			a.data.Mails[idx] = m
		} else {
			newCount++
			a.data.Mails = append(a.data.Mails, m)
		}
	}
	if err := a.saveLocked(); err != nil {
		return nil, fmt.Errorf("messages reçus, mais sauvegarde locale impossible : %w", err)
	}
	return map[string]any{"ok": true, "accountId": accountID, "account": acc.Email, "fetched": len(fetched), "newMessages": newCount}, nil
}

func fetchUID(line string) uint32 {
	fields := strings.Fields(line)
	for i := 0; i+1 < len(fields); i++ {
		if strings.EqualFold(strings.Trim(fields[i], "()"), "UID") {
			n, err := strconv.ParseUint(strings.Trim(fields[i+1], ")"), 10, 32)
			if err == nil {
				return uint32(n)
			}
		}
	}
	return 0
}

func decodeHeader(value string) string {
	decoded, err := new(mime.WordDecoder).DecodeHeader(value)
	if err != nil {
		return value
	}
	return decoded
}

var (
	htmlTagRE   = regexp.MustCompile(`<[^>]*>`)
	htmlSpaceRE = regexp.MustCompile(`[\t\x0b\f\r ]+`)
	scriptRE    = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	styleRE     = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
)

type mimeTextParts struct {
	plain       []string
	html        []string
	attachments []string
}

func decodeMIMETransfer(body []byte, encoding string) ([]byte, error) {
	var reader io.Reader = bytes.NewReader(body)
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "base64":
		reader = base64.NewDecoder(base64.StdEncoding, reader)
	case "quoted-printable":
		reader = quotedprintable.NewReader(reader)
	}
	return io.ReadAll(io.LimitReader(reader, 8<<20))
}

func collectMIMEPart(header textproto.MIMEHeader, body []byte, out *mimeTextParts) error {
	contentType := header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = "text/plain"
		params = map[string]string{}
	}
	disposition, dispParams, _ := mime.ParseMediaType(header.Get("Content-Disposition"))
	filename := dispParams["filename"]
	if filename == "" {
		filename = params["name"]
	}
	if filename != "" || strings.EqualFold(disposition, "attachment") {
		if filename != "" {
			out.attachments = append(out.attachments, decodeHeader(filename))
		}
		return nil
	}
	if strings.HasPrefix(strings.ToLower(mediaType), "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return nil
		}
		mr := multipart.NewReader(bytes.NewReader(body), boundary)
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			partBody, readErr := io.ReadAll(io.LimitReader(part, 8<<20))
			if readErr != nil {
				_ = part.Close()
				return readErr
			}
			if err := collectMIMEPart(part.Header, partBody, out); err != nil {
				_ = part.Close()
				return err
			}
			_ = part.Close()
		}
		return nil
	}
	decoded, err := decodeMIMETransfer(body, header.Get("Content-Transfer-Encoding"))
	if err != nil {
		return err
	}
	switch strings.ToLower(mediaType) {
	case "text/plain", "":
		out.plain = append(out.plain, string(decoded))
	case "text/html":
		out.html = append(out.html, string(decoded))
	}
	return nil
}

func htmlToText(input string) string {
	input = scriptRE.ReplaceAllString(input, " ")
	input = styleRE.ReplaceAllString(input, " ")
	input = regexp.MustCompile(`(?i)<br\s*/?>|</p>|</div>|</li>`).ReplaceAllString(input, "\n")
	input = htmlTagRE.ReplaceAllString(input, " ")
	input = html.UnescapeString(input)
	lines := strings.Split(input, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(htmlSpaceRE.ReplaceAllString(lines[i], " "))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func parseIMAPMail(account MailAccount, uid uint32, fetchLine string, raw []byte) (Mail, error) {
	message, err := netmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return Mail{}, err
	}
	headers := textproto.MIMEHeader(message.Header)
	from := decodeHeader(headers.Get("From"))
	fromEmail := ""
	if address, err := netmail.ParseAddress(from); err == nil {
		fromEmail = address.Address
		if strings.TrimSpace(address.Name) != "" {
			from = address.Name
		}
	} else {
		fromEmail = from
	}
	if from == "" {
		from = fromEmail
	}
	subject := decodeHeader(headers.Get("Subject"))
	if strings.TrimSpace(subject) == "" {
		subject = "(Sans objet)"
	}
	date := time.Now()
	if headerDate := headers.Get("Date"); headerDate != "" {
		if parsed, parseErr := netmail.ParseDate(headerDate); parseErr == nil {
			date = parsed.Local()
		}
	}
	bodyBytes, err := io.ReadAll(io.LimitReader(message.Body, imapMaxMessageBytes))
	if err != nil {
		return Mail{}, err
	}
	parts := &mimeTextParts{}
	if err := collectMIMEPart(headers, bodyBytes, parts); err != nil {
		return Mail{}, err
	}
	body := strings.TrimSpace(strings.Join(parts.plain, "\n\n"))
	if body == "" && len(parts.html) > 0 {
		for _, value := range parts.html {
			body = strings.TrimSpace(body + "\n\n" + htmlToText(value))
		}
	}
	if body == "" {
		body = "(Aucun contenu textuel disponible)"
	}
	id := fmt.Sprintf("imap-%s-%010d", account.ID, uid)
	to := decodeHeader(headers.Get("To"))
	cc := decodeHeader(headers.Get("Cc"))
	flags := strings.ToUpper(fetchLine)
	return Mail{ID: id, From: from, Email: fromEmail, AccountID: account.ID, To: to, Cc: cc, Subject: subject, Preview: trimPreview(firstLine(body)), Body: body, Time: date.Format("15:04"), Date: localizedMailDate(date), Folder: "inbox", Read: strings.Contains(flags, "\\SEEN"), Starred: strings.Contains(flags, "\\FLAGGED"), Attachments: parts.attachments}, nil
}

func trimPreview(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len([]rune(value)) > 180 {
		return string([]rune(value)[:177]) + "…"
	}
	return value
}

func localizedMailDate(date time.Time) string {
	now := time.Now()
	y, m, d := now.Date()
	y2, m2, d2 := date.Date()
	if y == y2 && m == m2 && d == d2 {
		return "Aujourd’hui"
	}
	if now.AddDate(0, 0, -1).Format("2006-01-02") == date.Format("2006-01-02") {
		return "Hier"
	}
	return date.Format("02/01/2006")
}

func (a *App) listForAccount(folder, query, accountID string) []Mail {
	mails := a.list(folder, query)
	if accountID == "" {
		return mails
	}
	out := make([]Mail, 0, len(mails))
	for _, m := range mails {
		if m.AccountID == accountID {
			out = append(out, m)
		}
	}
	return out
}

func (a *App) countsForAccount(accountID string) map[string]int {
	if accountID == "" {
		return a.counts()
	}
	counts := map[string]int{"inbox": 0, "important": 0, "drafts": 0, "sent": 0, "archive": 0, "trash": 0, "unread": 0}
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, m := range a.data.Mails {
		if m.AccountID != accountID {
			continue
		}
		if _, ok := counts[m.Folder]; ok {
			counts[m.Folder]++
		}
		if m.Important && m.Folder != "trash" {
			counts["important"]++
		}
		if !m.Read && m.Folder == "inbox" {
			counts["unread"]++
		}
	}
	return counts
}

func (a *App) accountRefreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	result, err := a.refreshAccountInbox(r.URL.Query().Get("id"))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}
