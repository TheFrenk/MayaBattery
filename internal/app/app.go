package app

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"

	"maya-battery/internal/autostart"
	"maya-battery/internal/config"
	"maya-battery/internal/device"
	"maya-battery/internal/estimate"
	"maya-battery/internal/icon"
	"maya-battery/internal/notify"
)

const (
	pollInterval   = 30 * time.Second
	retryInterval  = 5 * time.Second
	estimateWindow = 6 * time.Hour
)

type menu struct {
	numeric, battery               *systray.MenuItem
	fontLabel, fontPick, fontReset *systray.MenuItem
	notify, threshold, autostart   *systray.MenuItem
	refresh, quit                  *systray.MenuItem
}

type App struct {
	cfg     config.Config
	reader  *device.Reader
	icons   *icon.Renderer
	tracker *estimate.Tracker

	status, estimateItem *systray.MenuItem

	last        device.Battery
	lastAt      time.Time
	haveLast    bool
	connected   bool
	lowNotified bool
	lastErr     string
}

func New(cfg config.Config) *App {
	a := &App{
		cfg:     cfg,
		reader:  &device.Reader{},
		icons:   icon.NewRenderer(),
		tracker: estimate.New(estimateWindow),
	}
	if cfg.FontPath != "" {
		if err := a.icons.LoadFont(cfg.FontPath); err != nil {
			log.Println("шрифт из конфига не загружен:", err)
			a.cfg.FontPath = ""
		}
	}
	return a
}

func (a *App) Run() { systray.Run(a.onReady, a.onExit) }

func (a *App) onExit() { a.reader.Close() }

func (a *App) onReady() {
	systray.SetTitle("")
	systray.SetIcon(a.icons.Disconnected())
	systray.SetTooltip("MayaBattery — поиск мыши...")

	a.status = systray.AddMenuItem("Поиск мыши...", "")
	a.status.Disable()
	a.estimateItem = systray.AddMenuItem("Осталось: —", "")
	a.estimateItem.Disable()

	go a.loop(a.buildMenu())
}

func (a *App) buildMenu() *menu {
	m := &menu{}
	systray.AddSeparator()
	m.numeric = systray.AddMenuItemCheckbox("Числовая иконка", "", a.cfg.IconMode == icon.ModeNumeric)
	m.battery = systray.AddMenuItemCheckbox("Иконка батареи", "", a.cfg.IconMode == icon.ModeBattery)

	systray.AddSeparator()
	m.fontLabel = systray.AddMenuItem(a.fontText(), "")
	m.fontLabel.Disable()
	m.fontPick = systray.AddMenuItem("Выбрать шрифт…", "Файл .ttf или .otf")
	m.fontReset = systray.AddMenuItem("Сбросить шрифт", "")

	systray.AddSeparator()
	m.notify = systray.AddMenuItemCheckbox("Уведомлять о низком заряде", "", a.cfg.NotifyLow)
	m.threshold = systray.AddMenuItem(a.thresholdText(), "Нажмите для смены порога")
	m.autostart = systray.AddMenuItemCheckbox("Запускать вместе с Windows", "", autostart.Enabled())

	systray.AddSeparator()
	m.refresh = systray.AddMenuItem("Обновить сейчас", "")
	m.quit = systray.AddMenuItem("Выход", "")
	return m
}

func (a *App) loop(m *menu) {
	timer := time.NewTimer(0)
	defer timer.Stop()
	reschedule := func(d time.Duration) {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(d)
	}

	for {
		select {
		case <-timer.C:
			timer.Reset(a.update())

		case <-m.refresh.ClickedCh:
			reschedule(a.update())

		case <-m.numeric.ClickedCh:
			a.cfg.IconMode = icon.ModeNumeric
			a.save()
			setChecked(m.numeric, true)
			setChecked(m.battery, false)
			a.redraw()

		case <-m.battery.ClickedCh:
			a.cfg.IconMode = icon.ModeBattery
			a.save()
			setChecked(m.battery, true)
			setChecked(m.numeric, false)
			a.redraw()

		case <-m.fontPick.ClickedCh:
			file, err := dialog.File().
				Title("Выберите шрифт").
				Filter("Шрифты (*.ttf, *.otf)", "ttf", "otf").
				SetStartDir(`C:\Windows\Fonts`).
				Load()
			if err != nil {
				if !errors.Is(err, dialog.ErrCancelled) {
					log.Println("диалог:", err)
				}
				continue
			}
			if err := a.icons.LoadFont(file); err != nil {
				log.Println("шрифт:", err)
				continue
			}
			a.cfg.FontPath = file
			a.save()
			m.fontLabel.SetTitle(a.fontText())
			a.redraw()

		case <-m.fontReset.ClickedCh:
			a.icons.ResetFont()
			a.cfg.FontPath = ""
			a.save()
			m.fontLabel.SetTitle(a.fontText())
			a.redraw()

		case <-m.notify.ClickedCh:
			a.cfg.NotifyLow = !a.cfg.NotifyLow
			setChecked(m.notify, a.cfg.NotifyLow)
			a.lowNotified = false
			a.save()
			a.checkLow()

		case <-m.threshold.ClickedCh:
			a.cfg.Threshold = config.NextThreshold(a.cfg.Threshold)
			m.threshold.SetTitle(a.thresholdText())
			a.lowNotified = false
			a.save()
			a.checkLow()

		case <-m.autostart.ClickedCh:
			enable := !autostart.Enabled()
			if err := autostart.Set(enable); err != nil {
				log.Println("автозапуск:", err)
				continue
			}
			setChecked(m.autostart, enable)

		case <-m.quit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (a *App) update() time.Duration {
	b, err := a.reader.Read()
	if err != nil {
		a.connected = false
		if msg := err.Error(); msg != a.lastErr {
			log.Println(msg)
			a.lastErr = msg
		}
		a.redraw()
		a.refreshTexts()
		return retryInterval
	}

	now := time.Now()
	a.connected, a.haveLast = true, true
	a.last, a.lastAt, a.lastErr = b, now, ""
	a.tracker.Add(now, int(b.Percent), b.Charging)

	a.redraw()
	a.refreshTexts()
	a.checkLow()
	return pollInterval
}

func (a *App) redraw() {
	if !a.connected || !a.haveLast {
		systray.SetIcon(a.icons.Disconnected())
		return
	}
	systray.SetIcon(a.icons.Make(int(a.last.Percent), a.cfg.IconMode, a.last.Charging))
}

func (a *App) refreshTexts() {
	if !a.connected {
		a.status.SetTitle("Мышь не найдена — поиск...")
		a.estimateItem.SetTitle("Осталось: —")
		tip := "LAMZU Maya X\nМышь не найдена"
		if a.haveLast {
			tip += fmt.Sprintf("\nПоследний заряд: %d%% (%s)", a.last.Percent, a.lastAt.Format("15:04"))
		}
		setTooltip(tip)
		return
	}

	state, remaining := "от батареи", "считаем..."
	if a.last.Charging {
		state, remaining = "заряжается", "заряжается"
	} else if d, ok := a.tracker.Remaining(); ok {
		remaining = "~" + estimate.Format(d)
	}

	a.status.SetTitle(fmt.Sprintf("Заряд: %d%% — %s", a.last.Percent, state))
	a.estimateItem.SetTitle("Осталось: " + remaining)
	setTooltip(fmt.Sprintf("LAMZU Maya X\nЗаряд: %d%% (%s)\nОсталось: %s\nОбновлено: %s",
		a.last.Percent, state, remaining, a.lastAt.Format("15:04:05")))
}

func (a *App) checkLow() {
	if !a.haveLast || !a.connected {
		return
	}
	p := int(a.last.Percent)
	if a.cfg.NotifyLow && !a.last.Charging && p <= a.cfg.Threshold {
		if !a.lowNotified {
			a.lowNotified = true
			notify.Low(p)
		}
		return
	}
	if a.last.Charging || p > a.cfg.Threshold+5 {
		a.lowNotified = false
	}
}

func (a *App) save() {
	if err := config.Save(a.cfg); err != nil {
		log.Println("не удалось сохранить настройки:", err)
	}
}

func (a *App) fontText() string {
	if a.cfg.FontPath == "" {
		return "Шрифт: по умолчанию (Go Bold)"
	}
	return "Шрифт: " + filepath.Base(a.cfg.FontPath)
}

func (a *App) thresholdText() string {
	return fmt.Sprintf("Порог уведомления: %d%% (нажмите для смены)", a.cfg.Threshold)
}

func setChecked(item *systray.MenuItem, v bool) {
	if v {
		item.Check()
	} else {
		item.Uncheck()
	}
}

func setTooltip(s string) {
	if r := []rune(s); len(r) > 127 {
		s = string(r[:127])
	}
	systray.SetTooltip(s)
}
