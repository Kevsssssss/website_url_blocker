# 🛡️ Website URL Blocker

A lightweight, background parental control tool for Windows built in Go. It blocks domains system-wide by modifying the Windows hosts file and can run as an auto-starting Windows service.

## Features

- 🖥️ **Interactive shell** — double-click `urlblocker.exe` to open a dedicated terminal with a live prompt
- 🚫 **Blocks domains system-wide** — via the Windows hosts file, affects all browsers and apps
- 🔄 **Hot-reload** — detects changes to `blocklist.txt` every 30 seconds without restarting
- 🔐 **Password-protected** — bcrypt-hashed password prevents children from disabling the blocker
- ⚙️ **Windows Service support** — install once, starts automatically with Windows
- 🛡️ **Runtime admin check** — commands that need Administrator show a clear error if not elevated
- 🪶 **Lightweight** — single binary, no background UI, minimal resource usage

---

## Requirements

- Windows 10 / 11
- [Go 1.18+](https://go.dev/dl/) (to build from source)
- Administrator privileges (for hosts file and service commands)

---

## Quick Start

### 1. Build from source

```powershell
git clone https://github.com/Kevsssssss/website_url_blocker.git
cd website_url_blocker
go build -o urlblocker.exe .
```

### 2. Double-click to launch

Double-click `urlblocker.exe` — a console window opens automatically with an interactive shell:

```
  ┌──────────────────────────────────────┐
  │   🛡️  URL Blocker - Parental Control  │
  │          Interactive Shell           │
  └──────────────────────────────────────┘

  Type 'help'        → see all commands
  Type 'exit'        → close this window

urlblocker>
```

> 💡 **Tip:** Right-click `urlblocker.exe` and select **Pin to taskbar** for one-click access.

### 3. Set a password (first run)

```
urlblocker> setpassword
```

### 4. Add domains to block

```
urlblocker> add facebook.com
urlblocker> add tiktok.com
urlblocker> add instagram.com
```

### 5. Enable blocking

Run your terminal **as Administrator**, then:

```
urlblocker> enable
```

### 6. Install as a Windows service (auto-starts with Windows)

Open an **Administrator** terminal and run:

```
urlblocker> install
urlblocker> start
```

---

## Usage Modes

### Interactive Shell (recommended)

Double-click `urlblocker.exe` to open the interactive shell. Type commands directly — no directory navigation, no `.\urlblocker.exe` prefix needed:

```
urlblocker> list
urlblocker> add youtube.com
urlblocker> remove youtube.com
urlblocker> status
urlblocker> exit
```

### Command-line Mode

You can also run single commands from any terminal:

```powershell
.\urlblocker.exe help
.\urlblocker.exe list
.\urlblocker.exe add facebook.com
```

---

## CLI Reference

| Command | Description | Admin | Password |
|---|---|---|---|
| `help` | Show all available commands | ❌ | ❌ |
| `status` | Show service status and active blocks | ❌ | ❌ |
| `list` | List all domains in the blocklist | ❌ | ❌ |
| `add <domain>` | Add a domain to the blocklist | ❌ | ✅ |
| `remove <domain>` | Remove a domain from the blocklist | ❌ | ✅ |
| `enable` | Apply the blocklist to the hosts file now | ✅ | ✅ |
| `disable` | Remove all managed hosts file entries | ✅ | ✅ |
| `install` | Install as a Windows background service | ✅ | ✅ |
| `uninstall` | Remove the Windows service | ✅ | ✅ |
| `start` | Start the background service | ✅ | ✅ |
| `stop` | Stop the background service | ✅ | ✅ |
| `setpassword` | Set or change the CLI password | ❌ | ✅ (old) |
| `exit` / `quit` | Close the interactive shell | ❌ | ❌ |

---

## How It Works

1. You add domains to `blocklist.txt` (one per line)
2. Running `enable` injects entries into `C:\Windows\System32\drivers\etc\hosts`:
   ```
   # --- urlblocker START ---
   127.0.0.1 facebook.com
   127.0.0.1 www.facebook.com
   # --- urlblocker END ---
   ```
3. Any browser or app trying to reach `facebook.com` gets redirected to `127.0.0.1` — connection refused
4. When running as a service, `blocklist.txt` is polled every 30 seconds and changes are applied automatically

---

## Blocklist Format

`blocklist.txt` — one domain per line. Lines starting with `#` are comments. The `www.` prefix is added automatically.

```
# Social Media
facebook.com
instagram.com
tiktok.com
twitter.com

# Adult Content
# pornhub.com

# Gaming
# roblox.com
```

---

## Project Structure

```
website_url_blocker/
├── main.go               # Entrypoint — interactive shell or Windows service
├── go.mod / go.sum       # Go module + dependency lock
├── blocklist.txt         # Your domain blocklist
├── app.manifest          # Windows application manifest
├── rsrc.syso             # Compiled manifest (embedded into binary)
├── README.md
├── .gitignore
├── config/
│   ├── config.go         # Paths, constants, app data dir
│   └── admin.go          # Runtime Windows admin privilege check
├── service/
│   ├── blocker.go        # Hosts file read/write/inject/strip logic
│   ├── service.go        # Windows service lifecycle + 30s hot-reload
│   └── helpers.go        # os.Stat wrapper
├── cli/
│   └── cli.go            # All CLI commands + password + admin gate
└── repl/
    └── repl.go           # Interactive shell (REPL) mode
```

---

## Password Reset

If you forget the password, delete the hash file and set a new one:

```powershell
Remove-Item "$env:APPDATA\urlblocker\password.hash"
.\urlblocker.exe setpassword
```

> ⚠️ **Note:** Anyone with Administrator access can reset the password. For stronger protection, restrict access to the folder containing `urlblocker.exe`.

---

## Uninstalling

Open an **Administrator** terminal and run:

```
urlblocker> stop
urlblocker> disable
urlblocker> uninstall
```

---

## License

MIT
