package singleton

import (
	"errors"

	"golang.org/x/sys/windows"
)

var ErrAlreadyRunning = errors.New("программа уже запущена")

func Acquire(name string) (release func(), err error) {
	ptr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateMutex(nil, false, ptr)
	if err != nil {
		if h != 0 {
			_ = windows.CloseHandle(h)
		}
		if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
			return nil, ErrAlreadyRunning
		}
		return nil, err
	}
	return func() { _ = windows.CloseHandle(h) }, nil
}
