# LayoutFixer

LayoutFixer is a small Windows x64 tray utility that fixes text typed with the wrong English/Russian keyboard layout.

Select the mistyped text and press **Win + Shift** by itself. LayoutFixer converts the selection and leaves Windows on the layout that matches the final converted text.

## Examples

```text
руддщ цщкдв -> hello world
hello world -> руддщ цщкдв
руддщ cdtn -> hello свет
```

Mixed English/Russian selections are handled per detected text run, so one selection can contain text typed in both wrong layouts.

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

## v1.0.1 highlights

- Mixed-layout conversion: `руддщ cdtn` becomes `hello свет` in one operation.
- New high-contrast **A / Я** tray icon designed specifically for 16–24 px system-tray sizes.
- Multi-size icon resources for crisp Windows rendering.
- Preserves the original whole-selection EN↔RU conversion behavior.
- Keeps Windows on the layout matching the final converted text.

## Existing behavior

- Handles modifier-only **Win + Shift** separately from normal **Win + Shift + key** shortcuts.
- Prevents the internal `Ctrl+C` operation from becoming `Ctrl+Shift+C` and accidentally opening browser DevTools.
- Runs clipboard conversion away from the low-level keyboard-hook thread.
- Detects an already-running instance and shows a warning instead of silently exiting.
- Optional Windows startup entry from the tray menu.

## Build from source

Requires Go on Windows (or cross-compilation from another OS). For a basic build:

```powershell
go build -ldflags "-H=windowsgui -s -w" -o LayoutFixer.exe main.go
```

The release build additionally embeds `layoutfixer.ico` into the Windows executable resources so Explorer and application surfaces use the same icon as the tray app.

## Requirements

- Windows x64
- English and Russian keyboard layouts installed

## Security note

The executable is not code-signed, so Windows may show an **Unknown Publisher** or Microsoft Defender SmartScreen warning.

## SHA-256

See `SHA256SUMS.txt` for the checksum of the release executable.
