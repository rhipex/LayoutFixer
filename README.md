# LayoutFixer

LayoutFixer is a small Windows x64 tray utility that fixes text typed with the wrong English/Russian keyboard layout.

Select the mistyped text and press **Win + Shift** by itself. LayoutFixer converts the selected text and leaves Windows on the target keyboard layout.

> **v1.0.0** is the first public release of the internally tested v3 build.

## Example

```text
руддщ цщкдв -> hello world
hello world -> руддщ цщкдв
```

## Usage

1. Download `LayoutFixer.exe` from the latest release.
2. Exit any older LayoutFixer instance from its tray menu.
3. Run `LayoutFixer.exe`.
4. Select text typed using the wrong EN/RU layout.
5. Press **Win + Shift** by itself, then release both keys.
6. The selected text is converted and Windows is left on the target layout.

## Tray menu

- **Run at Windows startup** — toggles launch on sign-in.
- **Exit** — closes LayoutFixer.

## v1.0.0 highlights

- Converts selected text between English and Russian keyboard layouts.
- Keeps Windows on the target keyboard layout after conversion.
- Handles modifier-only **Win + Shift** separately from normal **Win + Shift + key** shortcuts.
- Prevents the internal `Ctrl+C` operation from becoming `Ctrl+Shift+C` and accidentally opening browser DevTools.
- Runs clipboard conversion away from the low-level keyboard-hook thread.
- Includes a proper **A↔Я** application/tray icon.
- Detects an already-running instance and shows a warning instead of silently exiting.
- Optional Windows startup entry from the tray menu.

## Build from source

Requires Go on Windows (or cross-compilation from another OS).

```powershell
go build -ldflags "-H=windowsgui -s -w" -o LayoutFixer.exe main.go
```

`main.go` embeds `layoutfixer.ico` for the tray icon. The distributed EXE also contains Windows icon resources for Explorer/application display.

## Requirements

- Windows x64
- English and Russian keyboard layouts installed

## Security note

The executable is not code-signed, so Windows may show an **Unknown Publisher** or Microsoft Defender SmartScreen warning.

## SHA-256

See `SHA256SUMS.txt` for the checksum of the included executable.
