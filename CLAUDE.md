# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**sshbox** is a CLI-based lightweight SSH connection management tool written in Go. It manages SSH connection configurations with tag-based grouping and quick connect features.

## Tech Stack

- **Language:** Go
- **CLI Framework:** `github.com/spf13/cobra`
- **SSH Library:** `golang.org/x/crypto/ssh`
- **Storage:** JSON file (`config.json`) stored alongside the executable

## Build & Development

```bash
# Build
go build -o sshbox .

# Run
go run . <command> [args] [flags]

# Test all
go test ./...

# Test single package
go test ./internal/service/

# Test single case
go test -run TestAddConnection ./internal/service/
```

## Architecture

```
cmd/           → Cobra command definitions (add, list, connect, edit, rm, show, export, import, tags)
internal/
  models/      → Connection, Config structs
  service/     → ConnectionManager (business logic, CRUD, import/export)
  storage/     → ConfigStore (JSON read/write, backup, file permissions)
  crypto/      → AES-256-CFB encrypt/decrypt, PBKDF2 key derivation
  keyring/     → System keyring integration for master key
  ssh/         → SSHClient (interactive session with PTY)
main.go        → Entry point
```

**Testing pattern:** cmd tests create a fresh `ConnectionManager` and fresh `cobra.Command` tree per test to avoid flag state leaking between tests.

**Config file location:** Same directory as the executable (`os.Executable()` → `filepath.Dir()` → `config.json`). This is intentional for portability — the entire directory can be copied/moved/synced.

## Key Commands (CLI interface)

| Command | Usage |
|---------|-------|
| `sshbox add <name>` | Add connection (`-H`/`--host`, `-P`/`--port`, `-u`/`--user`, `-p`/`--password`, `-t`/`--tag`, `-n`/`--notes`) |
| `sshbox list` | List connections (`--tag`, `--search`, `--format table/json`) |
| `sshbox connect <name>` | Open interactive SSH session |
| `sshbox password <name>` | Show connection password (plaintext) |
| `sshbox edit <name>` | Modify connection config |
| `sshbox rm <name>` | Delete connection (with confirmation) |
| `sshbox show <name>` | View connection details (password masked) |
| `sshbox export` | Export config (`--output`, `--tag`) |
| `sshbox import <file>` | Import config (`--merge` flag for merge mode) |
| `sshbox tags [tag]` | List all tags, or show connections under a specific tag |

## Security Design

- Passwords encrypted with AES-256-CFB, stored as base64 in config.json
- Master key stored in system keyring
- Key derivation via PBKDF2
- Config file permissions set to 600

## PRD References

- `PRD.md` — Full product requirements, data structures, error handling specs
- `技术可行性.md` — Technical feasibility, architecture design, code examples for encryption/SSH/cobra patterns
