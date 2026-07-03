package icon

import "encoding/binary"

func Data() []byte {
	const w, h = 16, 16
	andMask := ((w + 31) / 32 * 4) * h
	imgSize := 40 + w*h*4 + andMask
	b := make([]byte, 6+16+imgSize)
	binary.LittleEndian.PutUint16(b[2:], 1)
	binary.LittleEndian.PutUint16(b[4:], 1)
	b[6] = w
	b[7] = h
	b[8] = 0
	b[9] = 0
	binary.LittleEndian.PutUint16(b[10:], 1)
	binary.LittleEndian.PutUint16(b[12:], 32)
	binary.LittleEndian.PutUint32(b[14:], uint32(imgSize))
	binary.LittleEndian.PutUint32(b[18:], 22)
	o := 22
	binary.LittleEndian.PutUint32(b[o:], 40)
	binary.LittleEndian.PutUint32(b[o+4:], w)
	binary.LittleEndian.PutUint32(b[o+8:], h*2)
	binary.LittleEndian.PutUint16(b[o+12:], 1)
	binary.LittleEndian.PutUint16(b[o+14:], 32)
	binary.LittleEndian.PutUint32(b[o+20:], uint32(w*h*4))
	o += 40
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := o + (y*w+x)*4
			b[i+0] = 0x30
			b[i+1] = 0x90
			b[i+2] = 0xff
			b[i+3] = 0xff
		}
	}
	return b
}
