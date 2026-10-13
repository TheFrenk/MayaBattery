package icon

import (
	"bytes"
	"encoding/binary"
	"image"
)

func encodeICO(imgs []*image.NRGBA) []byte {
	blobs := make([][]byte, 0, len(imgs))
	for _, img := range imgs {
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		var b bytes.Buffer
		for _, v := range []any{
			uint32(40), int32(w), int32(h * 2), uint16(1), uint16(32),
			uint32(0), uint32(w * h * 4), int32(0), int32(0), uint32(0), uint32(0),
		} {
			binary.Write(&b, binary.LittleEndian, v)
		}
		for y := h - 1; y >= 0; y-- {
			for x := 0; x < w; x++ {
				p := img.NRGBAAt(x, y)
				b.Write([]byte{p.B, p.G, p.R, p.A})
			}
		}
		b.Write(make([]byte, ((w+31)/32)*4*h))
		blobs = append(blobs, b.Bytes())
	}

	var ico bytes.Buffer
	n := len(imgs)
	binary.Write(&ico, binary.LittleEndian, uint16(0))
	binary.Write(&ico, binary.LittleEndian, uint16(1))
	binary.Write(&ico, binary.LittleEndian, uint16(n))

	offset := 6 + 16*n
	for i, img := range imgs {
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		ico.Write([]byte{byte(w), byte(h), 0, 0})
		binary.Write(&ico, binary.LittleEndian, uint16(1))
		binary.Write(&ico, binary.LittleEndian, uint16(32))
		binary.Write(&ico, binary.LittleEndian, uint32(len(blobs[i])))
		binary.Write(&ico, binary.LittleEndian, uint32(offset))
		offset += len(blobs[i])
	}
	for _, b := range blobs {
		ico.Write(b)
	}
	return ico.Bytes()
}
