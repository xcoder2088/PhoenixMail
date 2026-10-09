# PhoenixMail roadmap

## V0.1 — Foundation
- [x] Branded UI
- [x] Local demo inbox
- [x] Compose window
- [x] AI action interface
- [x] AGPL project structure

## Current manual validation note (v0.3.4 candidate)
- [x] Account click selects its inbox instead of starting compose
- [x] A real Outlook INBOX fetch displayed 40 messages and a selected message body
- [ ] Send a controlled test email from outside PhoenixMail and confirm refresh adds it exactly once
- [ ] Confirm real SMTP send and remote read/delete/move behavior before calling Email Core complete

## V0.2 — Mail core
- [ ] SQLite schema / migrations
- [x] Initial IMAP INBOX fetch (last 40 messages per account)
- [x] Basic MIME reading (plain text, HTML fallback, attachment names)
- [x] SMTP send path
- [x] Multiple local account settings
- [x] Draft autosave
- [ ] Remote read/unread and flagged state updates
- [ ] Remote move/archive/delete operations
- [ ] Sent/Drafts/Archive mailbox discovery and synchronization
- [ ] Attachment download and opening
- [ ] Robust reconnect, UIDVALIDITY and incremental synchronization

## V0.3 — Local AI
- [x] GGUF model discovery / download manager (initial implementation)
- [x] Local llama.cpp start/stop adapter (initial implementation)
- [ ] Hardware detection and automatic model selection
- [ ] Context limits and streaming responses
- [ ] Broader local-AI integration tests

## V0.4 — Providers and identity
- [x] Microsoft OAuth2 authorization flow and token refresh
- [ ] Validate incremental new-message arrival and remote-state synchronization with a real Outlook account
- [ ] Gmail OAuth2 and Gmail IMAP authentication
- [ ] Provider-specific sync optimizations
- [ ] Nimiq Identity and PhoenixMail Verify

## V0.5 — Desktop
- [ ] Tauri shell
- [ ] Windows build
- [ ] Linux AppImage / Flatpak
- [ ] macOS build

## V1.0
- [ ] Stable release
- [ ] Full documentation
- [ ] Contributor guide
- [ ] Security audit checklist
