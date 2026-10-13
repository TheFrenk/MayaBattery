package autostart

import (
	"errors"
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	keyPath   = `Software\Microsoft\Windows\CurrentVersion\Run`
	valueName = "MayaBattery"
)

func Enabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(valueName)
	return err == nil
}

func Set(enable bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if !enable {
		err := k.DeleteValue(valueName)
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return k.SetStringValue(valueName, `"`+exe+`"`)
}
