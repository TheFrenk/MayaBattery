package device

import (
	"errors"
	"fmt"
	"time"

	"github.com/sstallion/go-hid"
)

const (
	vendorID  = 0x373E
	productID = 0x001E
	usagePage = 0xFFFF
	iface     = 2
)

var ErrNotFound = errors.New("мышь не найдена")

type Device struct{ h *hid.Device }

func findPath() (string, error) {
	var path string
	err := hid.Enumerate(vendorID, productID, func(info *hid.DeviceInfo) error {
		if info.InterfaceNbr == iface && info.UsagePage == usagePage {
			path = info.Path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", ErrNotFound
	}
	return path, nil
}

func Open() (*Device, error) {
	path, err := findPath()
	if err != nil {
		return nil, err
	}
	h, err := hid.OpenPath(path)
	if err != nil {
		return nil, fmt.Errorf("открытие устройства: %w", err)
	}
	return &Device{h: h}, nil
}

func (d *Device) Close() { _ = d.h.Close() }

func (d *Device) Read() (Battery, error) {
	if _, err := d.h.SendFeatureReport(batteryRequest()); err != nil {
		return Battery{}, fmt.Errorf("запрос батареи: %w", err)
	}
	time.Sleep(100 * time.Millisecond)

	resp := make([]byte, 65)
	n, err := d.h.GetFeatureReport(resp)
	if err != nil {
		return Battery{}, fmt.Errorf("чтение батареи: %w", err)
	}
	return parseResponse(resp[:n])
}

// Reader сам находит мышь и переподключается после ошибок.
type Reader struct{ dev *Device }

func (r *Reader) Read() (Battery, error) {
	if r.dev == nil {
		d, err := Open()
		if err != nil {
			return Battery{}, err
		}
		r.dev = d
	}
	b, err := r.dev.Read()
	if err != nil {
		r.dev.Close()
		r.dev = nil
		return Battery{}, err
	}
	return b, nil
}

func (r *Reader) Close() {
	if r.dev != nil {
		r.dev.Close()
		r.dev = nil
	}
}
