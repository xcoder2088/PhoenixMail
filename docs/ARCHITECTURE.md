# PhoenixMail architecture

## Target

A cross-platform, local-first mail client. The application must remain useful on modest hardware and must not require a GPU.

## Layers

```text
UI / Tauri
   |
Application services
   |-- Mail service
   |-- Search service
   |-- Settings
   |-- AI service
   |
Adapters
   |-- IMAP / SMTP
   |-- Gmail OAuth/API
   |-- Microsoft Graph
   |-- llama.cpp
   |-- Ollama (optional)
   |
Local storage
   |-- SQLite
   |-- OS credential store
```

## AI abstraction

The UI must never know which model is running. It calls an AI interface with operations such as:

- correct
- rewrite
- translate
- summarize
- reply
- shorten
- extract actions

The engine selects an adapter and model according to local hardware capabilities.

## Hardware tiers

The target is not a minimum GPU. The target is graceful degradation:

- CPU-only machines: small quantized models
- 8 GB class systems: small/medium quantized models
- 16 GB+ systems: larger local models where practical
- GPU systems: optional acceleration

If local AI is unavailable, the mail client itself remains fully functional.
