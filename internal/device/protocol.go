package device

import "fmt"

const (
	respHeader  = 0xA1
	cmdBattery  = 0x83
	minRespSize = 9
)

type Battery struct {
	Percent  byte
	Charging bool
}

func batteryRequest() []byte {
	r := make([]byte, 65)
	r[3] = 0x02
	r[4] = 0x02
	r[6] = cmdBattery
	return r
}

func parseResponse(resp []byte) (Battery, error) {
	if len(resp) < minRespSize {
		return Battery{}, fmt.Errorf("короткий ответ HID: %d байт", len(resp))
	}
	if resp[1] != respHeader || resp[6] != cmdBattery {
		return Battery{}, fmt.Errorf("неожиданный ответ HID")
	}
	if resp[8] > 100 {
		return Battery{}, fmt.Errorf("неверный заряд: %d", resp[8])
	}
	return Battery{Percent: resp[8], Charging: resp[7] == 1}, nil
}
