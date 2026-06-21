# agmsg-viewer

[日本語版はこちら](README-ja.md)

A tool to view agent-to-agent message history (SQLite database) recorded by [`fujibee/agmsg`](https://github.com/fujibee/agmsg) in a LINE-style chat interface via browser.

## 🚀 User Guide

### Features
- **LINE-style UI**: Displays sent/received messages in an intuitive bubble UI.
- **Agent identification**: Each sender agent gets a consistent color for easy recognition.
- **Auto refresh**: New messages automatically appear on screen via 15-second polling.
- **Team persistence**: Selected team is saved and auto-loaded on next access.
- **Portability**: Runs as a single Go binary with no external runtime dependencies.

### Installation & Launch
Obtain a prebuilt binary or build from source.

#### Launch Command
```bash
./agmsg-viewer -db /path/to/messages.db -port 8080
```

#### Options
| Flag | Default | Description |
|------|---------|-------------|
| `-db` | `messages.db` | Path to SQLite database file |
| `-port`| `8080` | HTTP server port |
| `-tail`| `40` | Initial number of recent messages (0 for all) |
| `-team`| (none) | Team to pre-select on load |

### Usage
1. Start the server and access `http://localhost:8080` in your browser.
2. Select a team from the "Team" dropdown at the top of the screen.
3. Hover over message timestamps to see detailed date/time.

---

## 🛠️ Development & Contributor Guide

### Tech Stack
- **Backend**: Go 1.26+ (Standard Library)
- **Database**: SQLite (via `modernc.org/sqlite` - CGO-free)
- **Frontend**: HTMX + Tailwind CSS (Vanilla JS)
- **Embedding**: Go `embed` package (HTML/CSS bundled into binary)

### Development Flow
Uses `go-task` for build/test management.

#### Basic Commands
- **Build**: `task build`
- **Run**: `task run`
- **Format**: `task fmt`
- **Lint**: `task lint`
- **Test**: `task test`
- **Clean**: `task clean`

#### Cross-platform Builds
- **Windows**: `task build:win`
- **Linux**: `task build:linux`
- **macOS**: `task build:mac`

### Database Schema
Assumes the `messages` table from `fujibee/agmsg`.

#### Table: `messages`
| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER | Message ID (PK) |
| `team` | TEXT | Team name |
| `from_agent` | TEXT | Sender agent name |
| `to_agent` | TEXT | Recipient agent name |
| `body` | TEXT | Message body |
| `created_at` | TEXT | Creation time (ISO 8601 UTC) |

---

## 📄 License

This project is released under the [MIT License](LICENSE).