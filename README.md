# MayaBattery

A tiny Windows tray app written in Go that shows the battery level of a
**LAMZU Maya X** mouse as a number right in the notification area, so you
don't have to open the web driver every time.

> **Disclaimer:** This is an unofficial, community project. It is not
> affiliated with or endorsed by LAMZU. It talks to the mouse over an
> undocumented vendor HID interface, so a firmware update may break it.

<!-- Add screenshots: tray on dark taskbar, tray on light taskbar, menu -->
<!-- ![Tray icon](docs/tray.png) -->

## Features

- Battery percentage drawn directly on the tray icon (numeric or battery-shaped).
- Color by level: green, yellow, red, and blue while charging.
- Crisp icons: rendered at 16, 20, 24, 32 and 48 px so Windows picks the right size at any DPI.
- Custom font: pick any `.ttf` or `.otf` file from the tray menu.
- Low battery notification with a switchable on/off option and a selectable threshold (10, 15, 20, 30%).
- Estimated remaining time, calculated from the recent discharge rate.
- Detailed tooltip: charge, state, remaining time, last update.
- Auto reconnect: finds the mouse by itself and recovers after sleep, unplugging or dongle changes.
- Start with Windows (toggle in the menu).
- Single instance: a second launch exits silently.
- Settings are saved between runs.

## Install

Download `MayaBattery.exe` from the
[Releases](https://github.com/TheFrenk/MayaBattery/releases) page and
run it. The icon appears in the notification area (it may be hidden under the
`^` arrow; drag it to the taskbar to keep it visible).

To start it with Windows, enable **Start with Windows** in the tray menu.
Do this from the final `.exe` location: the autostart entry stores the path of
the running executable.

## Usage

Right-click the tray icon to open the menu.

| Item | Description |
|---|---|
| Charge / Remaining | Current level, state and estimated time left |
| Numeric icon / Battery icon | Choose how the level is drawn |
| Choose font... / Reset font | Use any `.ttf` or `.otf` file, or go back to the built-in font |
| Notify on low battery | Turn the notification on or off |
| Notification threshold | Click to cycle through 10, 15, 20 and 30% |
| Start with Windows | Toggle autostart |
| Refresh now | Read the battery immediately |

The notification fires once when the level drops to the threshold while the
mouse is not charging. It is re-armed after charging or when the level rises
5% above the threshold.

## Configuration

Settings are stored in `%AppData%\MayaBattery\config.json`:

```json
{
  "icon_mode": "numeric",
  "font_path": "",
  "notify_low": true,
  "threshold": 20
}
```

`threshold` accepts any value from 1 to 99, even if it is not in the menu list.

## How it works

1. The app enumerates HID devices and selects the vendor-defined interface
   (interface `2`, usage page `0xFFFF`).
2. It sends a feature report with the battery command and reads the response.
3. The response is validated (header, command, value range 0-100) and parsed
   into a percentage and a charging flag.
4. The percentage is drawn into a multi-size ICO in memory and set as the tray icon.
5. A small tracker keeps recent samples and fits the discharge rate to estimate
   the remaining time. It resets while charging or after long gaps.

The battery value reported by the mouse is an integer percentage, so the
remaining time estimate is approximate, especially right after startup.

## Limitations

- Windows only.
- Tested with one device model (Maya X).
- Relies on a reverse-engineered protocol that may change with firmware updates.
