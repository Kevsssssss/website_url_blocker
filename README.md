# 🛡️ Website URL Blocker

A lightweight, background parental control tool for Windows built in Go. It blocks domains by modifying the Windows hosts file and runs as a Windows service.

## Features

- 🚫 **Blocks domains** system-wide via the Windows hosts file
- 🔄 **Hot-reload** — detects changes to `blocklist.txt` every 30 seconds without restarting
- 🔐 **Password-protected CLI** — prevents children from disabling the blocker
- ⚙️ **Runs as a Windows Service** — starts automatically with Windows
- 🪶 **Lightweight** — single binary, no background UI, minimal resource usage

## Requirements

- Windows 10 / 11
- Administrator privileges (for hosts file modification and service management)

## Installation

### 1. Build from source

```powershell
git clone https://github.com/Kevsssssss/website_url_blocker.git
cd website_url_blocker
go build -o urlblocker.exe .
```

### 2. Set a password (first run)

```powershell
.\urlblocker.exe setpassword
```

### 3. Add domains to block

Edit `blocklist.txt` directly, or use the CLI:

```powershell
.\urlblocker.exe add facebook.com
.\urlblocker.exe add tiktok.com
.\urlblocker.exe add instagram.com
```

### 4. Enable blocking immediately

```powershell
# Run as Administrator
.\urlblocker.exe enable
```

### 5. Install as a Windows service (auto-starts with Windows)

```powershell
# Run as Administrator
.\urlblocker.exe install
.\urlblocker.exe start
```

## CLI Reference

| Command | Description | Requires Password |
|---|---|---|
| `status` | Show service status and active blocks | No |
| `list` | List all domains in the blocklist | No |
| `add <domain>` | Add a domain to the blocklist | ✅ Yes |
| `remove <domain>` | Remove a domain from the blocklist | ✅ Yes |
| `enable` | Apply blocklist to hosts file now | ✅ Yes |
| `disable` | Remove all managed hosts file entries | ✅ Yes |
| `install` | Install as a Windows background service | ✅ Yes (+ Admin) |
| `uninstall` | Remove the Windows service | ✅ Yes (+ Admin) |
| `start` | Start the service | ✅ Yes (+ Admin) |
| `stop` | Stop the service | ✅ Yes (+ Admin) |
| `setpassword` | Set or change the CLI password | ✅ Old password |

## How It Works

1. You add domains to `blocklist.txt` (one per line)
2. The tool injects entries into `C:\Windows\System32\drivers\etc\hosts`:
   ```
   # --- urlblocker START ---
   127.0.0.1 facebook.com
   127.0.0.1 www.facebook.com
   # --- urlblocker END ---
   ```
3. When a browser tries to access `facebook.com`, it resolves to `127.0.0.1` (your own machine), causing a connection failure
4. The service polls `blocklist.txt` every 30 seconds and hot-reloads any changes

## Blocklist Format

`blocklist.txt` — one domain per line. Lines starting with `#` are comments.

```
# Social Media
facebook.com
instagram.com
tiktok.com

# Gaming
roblox.com
```

You don't need to include `www.` — it is automatically blocked alongside the base domain.

## Password Reset

If you forget the password, delete the hash file and set a new one:

```powershell
Remove-Item "$env:APPDATA\urlblocker\password.hash"
.\urlblocker.exe setpassword
```

> ⚠️ **Note**: Anyone with Administrator access can do this. For stronger protection, consider restricting access to the executable folder.

## Uninstalling

```powershell
# Run as Administrator
.\urlblocker.exe stop
.\urlblocker.exe disable
.\urlblocker.exe uninstall
```

## License

MIT
