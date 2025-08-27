# Zephyre Multiseat

A Windows multiseat prototype in Golang for concurrent user access in educational/humanitarian settings.

## Setup
1. Install Golang.
2. `go mod tidy` to fetch deps.
3. Build: `go build -o zephyre.exe cmd/zephyre/main.go`.
4. Run as admin: `zephyre.exe` (for user creation/net commands).
5. For concurrency: Install RDP Wrapper (manual) for multi-RDP sessions.
6. Logs in zephyre.log.

## Features
- Scan hardware.
- Assign devices to seats via GUI.
- Create users for seats.
- Apply configs (stubs for RDP/VM).
- Internet sharing.

Extend for production use.