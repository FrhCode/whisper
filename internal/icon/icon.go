package icon

import "encoding/binary"

func Data() []byte {
	imgs := []image{{16, 16}, {32, 32}}
	head := make([]byte, 6+16*len(imgs))
	binary.LittleEndian.PutUint16(head[2:], 1)
	binary.LittleEndian.PutUint16(head[4:], uint16(len(imgs)))
	var data []byte
	off := len(head)
	for i, im := range imgs {
		png := dib(im.w, im.h)
		e := 6 + i*16
		head[e] = byte(im.w)
		head[e+1] = byte(im.h)
		binary.LittleEndian.PutUint16(head[e+4:], 1)
		binary.LittleEndian.PutUint16(head[e+6:], 32)
		binary.LittleEndian.PutUint32(head[e+8:], uint32(len(png)))
		binary.LittleEndian.PutUint32(head[e+12:], uint32(off))
		off += len(png)
		data = append(data, png...)
	}
	return append(head, data...)
}

type image struct{ w, h int }

func dib(w, h int) []byte {
	andMask := ((w + 31) / 32 * 4) * h
	b := make([]byte, 40+w*h*4+andMask)
	binary.LittleEndian.PutUint32(b[0:], 40)
	binary.LittleEndian.PutUint32(b[4:], uint32(w))
	binary.LittleEndian.PutUint32(b[8:], uint32(h*2))
	binary.LittleEndian.PutUint16(b[12:], 1)
	binary.LittleEndian.PutUint16(b[14:], 32)
	binary.LittleEndian.PutUint32(b[20:], uint32(w*h*4))
	pix := b[40:]
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bb, g, r, a := tile(x, y, w, h)
			set(pix, w, h, x, y, bb, g, r, a)
		}
	}
	drawMic(pix, w, h)
	return b
}

func tile(x, y, w, h int) (byte, byte, byte, byte) {
	m := w / 8
	if !rounded(x, y, w, h, m, w/5) {
		return 0, 0, 0, 0
	}
	t := float64(x+y) / float64(w+h-2)
	r := byte(0x60*t + 0x00*(1-t))
	g := byte(0xcd*t + 0x78*(1-t))
	bb := byte(0xff*t + 0xd4*(1-t))
	return bb, g, r, 255
}

func rounded(x, y, w, h, m, r int) bool {
	if x >= m && x < w-m || y >= m && y < h-m {
		return true
	}
	cx, cy := m, m
	if x >= w-m {
		cx = w - m - 1
	}
	if y >= h-m {
		cy = h - m - 1
	}
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= r*r
}

func drawMic(p []byte, w, h int) {
	s := w / 16
	if s < 1 {
		s = 1
	}
	cx := w / 2
	stroke := max(1, w/7)
	rx := max(2, w/8)
	top, bottom := h/4, h*9/16
	for y := top; y <= bottom; y++ {
		for x := cx - rx; x <= cx+rx; x++ {
			border := abs(x-(cx-rx)) < stroke || abs(x-(cx+rx)) < stroke || y < top+stroke || y > bottom-stroke
			if border && insideRoundRect(x, y, cx-rx, top, rx*2+1, bottom-top+1, rx) {
				set(p, w, h, x, y, 255, 255, 255, 245)
			}
		}
	}
	arcY := h * 9 / 16
	for x := w / 4; x <= w*3/4; x++ {
		dx := x - cx
		y := arcY + (dx*dx)/(w/3)
		for yy := y; yy < y+stroke; yy++ {
			set(p, w, h, x, yy, 255, 255, 255, 245)
		}
	}
	line(p, w, h, cx, h*11/16, cx, h*13/16, stroke)
	line(p, w, h, w*3/8, h*13/16, w*5/8, h*13/16, stroke)
}

func insideRoundRect(x, y, rx, ry, rw, rh, rr int) bool {
	cx, cy := x, y
	if cx < rx+rr {
		cx = rx + rr
	}
	if cx >= rx+rw-rr {
		cx = rx + rw - rr - 1
	}
	if cy < ry+rr {
		cy = ry + rr
	}
	if cy >= ry+rh-rr {
		cy = ry + rh - rr - 1
	}
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= rr*rr
}

func line(p []byte, w, h, x1, y1, x2, y2, stroke int) {
	if x1 == x2 {
		for y := min(y1, y2); y <= max(y1, y2); y++ {
			for x := x1 - stroke/2; x <= x1+stroke/2; x++ {
				set(p, w, h, x, y, 255, 255, 255, 245)
			}
		}
		return
	}
	for x := min(x1, x2); x <= max(x1, x2); x++ {
		for y := y1 - stroke/2; y <= y1+stroke/2; y++ {
			set(p, w, h, x, y, 255, 255, 255, 245)
		}
	}
}

func set(p []byte, w, h, x, y int, b, g, r, a byte) {
	if x < 0 || y < 0 || x >= w || y >= h {
		return
	}
	i := ((h-1-y)*w + x) * 4
	p[i], p[i+1], p[i+2], p[i+3] = b, g, r, a
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
