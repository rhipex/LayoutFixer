# LayoutFixer v1.0.1

This update improves mixed-language correction and makes the tray icon much easier to recognize at Windows system-tray size.

## Changes

- Mixed EN/RU selections are now converted by detected script/run instead of forcing the entire selection through one direction.
- `руддщ cdtn` now correctly converts to `hello свет` with one **Win + Shift** press.
- Original conversions remain intact, including `руддщ цщкдв` → `hello world` and `hello world` → `руддщ цщкдв`.
- Replaced the tray/app icon with a high-contrast split **A / Я** design optimized for 16–24 px display.
- Added multi-size icon resources for sharper Windows rendering.
