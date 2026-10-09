# PhoenixMail

**Your Mail. Your AI. Your Control.**

PhoenixMail is an open-source, privacy-first desktop mail client with adaptive local AI assistance.

## Project principles

- **AGPL-3.0**
- No subscription
- No mandatory PhoenixMail account
- Local-first data model
- Local AI by default
- Works CPU-only; hardware acceleration is optional
- IMAP/SMTP first, provider APIs where they improve OAuth/integration
- No telemetry by default

## V0.2.8 status

This build keeps the approved PhoenixMail UI while adding real plumbing:

- PhoenixMail branded UI
- Persistent local mail store (atomic JSON)
- Inbox / Important / Drafts / Sent / Trash
- Search and message reader
- Read/unread, star, important, archive and trash actions
- Compose window with draft autosave
- Multiple local mail accounts with IMAP/SMTP settings
- SMTP + IMAP connection tests
- Default sending account selection
- Real SMTP send path
- Local AI model manager for Qwen3-0.6B Q4_0 GGUF
- Optional local llama.cpp server start/stop
- Approved UI concept stored under `reference/`
- Official standalone brand kit included

The AI model is downloaded separately; PhoenixMail does not bundle model weights.

## Run

Requirements: Go 1.23+

```bash
go run ./cmd/phoenixmail
```

Then open `http://localhost:8787`.

Optional port:

```bash
PHOENIXMAIL_PORT=9000 go run ./cmd/phoenixmail
```

## Roadmap

1. Core desktop shell (Tauri)
2. SQLite database and settings (after clean vendoring strategy is selected)
3. IMAP/SMTP account engine
4. OAuth providers (Gmail / Microsoft)
5. Real local AI via llama.cpp
6. Adaptive model selection based on CPU/RAM/GPU
7. Attachments and MIME handling
8. Search/indexing
9. Packaging: Windows, Linux, macOS
10. Community plugin/model architecture

## Branding

The official PhoenixMail brand kit is included under `assets/branding/`.

## License

PhoenixMail is intended to be released under **GNU Affero General Public License v3.0 (AGPL-3.0-only)**. See `LICENSE`.

## Microsoft OAuth2 (développement local)

PhoenixMail n'utilise pas le mot de passe pour les comptes Outlook/Microsoft 365 configurés en OAuth2. Le flux local utilise Authorization Code + PKCE et XOAUTH2 pour SMTP. La configuration de développement PhoenixMail utilise maintenant l'application Microsoft PhoenixMail enregistrée pour les tests. Aucun secret client n'est requis. Un autre Client ID peut être fourni par variable d'environnement si nécessaire :

```bash
export PHOENIXMAIL_MICROSOFT_CLIENT_ID="VOTRE-CLIENT-ID"
```

L'application Microsoft doit être configurée comme application desktop/public client avec une redirection `http://localhost` adaptée au port local utilisé par PhoenixMail. Les permissions OAuth nécessaires pour IMAP/SMTP sont `https://outlook.office.com/IMAP.AccessAsUser.All` et `https://outlook.office.com/SMTP.Send`, avec `offline_access` pour le renouvellement de session.
