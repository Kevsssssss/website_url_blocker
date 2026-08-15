[Setup]
AppName=URL Blocker - Parental Control
AppVersion=1.0.1
DefaultDirName={pf}\URLBlocker
DefaultGroupName=URL Blocker
UninstallDisplayIcon={app}\urlblocker.exe
Compression=lzma2
SolidCompression=yes
OutputBaseFilename=URLBlocker_Setup_v1.0.1
ArchitecturesInstallIn64BitMode=x64

[Files]
Source: "urlblocker.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "blocklist.txt"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\URL Blocker Interactive Shell"; Filename: "{app}\urlblocker.exe"

[Run]
; Install and start the service silently after installation
Filename: "{app}\urlblocker.exe"; Parameters: "install"; Flags: runhidden runascurrentuser
Filename: "{app}\urlblocker.exe"; Parameters: "start"; Flags: runhidden runascurrentuser
; Launch the interactive shell for the user to set up their password and blocks
Filename: "{app}\urlblocker.exe"; Description: "Launch URL Blocker Interactive Shell"; Flags: postinstall runascurrentuser

[UninstallRun]
; Force kill any open interactive shells so uninstaller can delete the .exe
Filename: "taskkill"; Parameters: "/f /im urlblocker.exe"; Flags: runhidden
; Stop and uninstall the service before removing files
Filename: "{app}\urlblocker.exe"; Parameters: "stop"; Flags: runhidden runascurrentuser
Filename: "{app}\urlblocker.exe"; Parameters: "disable"; Flags: runhidden runascurrentuser
Filename: "{app}\urlblocker.exe"; Parameters: "uninstall"; Flags: runhidden runascurrentuser
