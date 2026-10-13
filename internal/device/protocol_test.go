package device

import "testing"

func resp(percent, charging byte) []byte {
	r := make([]byte, 65)
	r[1], r[6], r[7], r[8] = respHeader, cmdBattery, charging, percent
	return r
}

func TestParseResponse(t *testing.T) {
	b, err := parseResponse(resp(57, 1))
	if err != nil || b.Percent != 57 || !b.Charging {
		t.Fatalf("got %+v, %v", b, err)
	}
	if _, err := parseResponse(resp(101, 0)); err == nil {
		t.Error("ожидалась ошибка для заряда > 100")
	}
	if _, err := parseResponse(make([]byte, 5)); err == nil {
		t.Error("ожидалась ошибка для короткого ответа")
	}
}
