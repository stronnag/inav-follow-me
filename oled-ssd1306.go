package main

import (
	font "github.com/Nondzu/ssd1306_font"
	"image/color"
	"tinygo.org/x/drivers/ssd1306"
)

const FONT_H int16 = 10
const FONT_W int16 = 7
const OLED_WIDTH = 128
const OLED_HEIGHT = 64

type OledDisplay struct {
	d   font.Display
	dev ssd1306.Device
}

func NewOLED(dev ssd1306.Device) *OledDisplay {
	return &OledDisplay{d: font.NewDisplay(dev), dev: dev}
}

func Itoa(val int) string {
	if val < 0 {
		return "-" + Uitoa(uint(-val))
	}
	return Uitoa(uint(val))
}

func Uitoa(val uint) string {
	if val == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	for val >= 10 {
		q := val / 10
		buf[i] = byte('0' + val - q*10)
		i--
		val = q
	}
	// val < 10
	buf[i] = byte('0' + val)
	return string(buf[i:])
}

func fill(t string, sz int, zf bool) string {
	n := sz - len(t)
	if n > 0 {
		buf := make([]byte, sz)
		i := 0
		for ; i < n; i++ {
			if zf {
				buf[i] = '0'
			} else {
				buf[i] = ' '
			}
		}
		for _, c := range []byte(t) {
			buf[i] = c
			i += 1
		}
		return string(buf)
	}
	return t
}

func (o *OledDisplay) cEOL() {
	nspace := (OLED_WIDTH - o.d.XPos) / FONT_W
	spcs := make([]byte, nspace)
	for i := int16(0); i < nspace; i++ {
		spcs[i] = ' '
	}
	o.d.PrintText(string(spcs))
}

func (o *OledDisplay) incX(n int) {
	o.d.XPos += int16(n) * FONT_W
}

func (o *OledDisplay) setPos(x int, y int) {
	o.d.XPos = int16(x) * FONT_W
	o.d.YPos = int16(y) * FONT_H
	if y > 1 {
		o.d.YPos += 3
	}
}

func (o *OledDisplay) InitScreen() {
	o.ShowTime("--:--:--")

	o.setPos(0, 1)
	o.d.PrintText("GPS :")

	o.setPos(0, 2)
	o.d.PrintText("INAV:")

	o.drawSep()

	o.setPos(0, 3)
	o.d.PrintText("Mode:")

	o.setPos(0, 4)
	o.d.PrintText("VSat:")

	o.setPos(0, 5)
	o.d.PrintText("VPos:")
}

func (o *OledDisplay) ShowTime(t string) {
	xpos := (18 - len(t)) / 2
	o.setPos(xpos, 0)
	o.d.PrintText(t)
	o.incX(len(t))
	o.cEOL()
}

func (o *OledDisplay) ShowGPS(nsat uint16, fix uint8) {
	o.setPos(6, 1)
	t := Uitoa(uint(nsat))
	o.d.PrintText(t)
	o.incX(len(t))
	if nsat < 2 {
		t = " sat "
	} else {
		t = " sats "
	}
	o.d.PrintText(t)
	o.incX(len(t))

	switch fix {
	case 0:
		t = "NoFix"
	case 1:
		t = "Fix"
	case 2:
		t = "DFix"
	}
	o.d.PrintText(t)
	o.incX(len(t))
	o.cEOL()
}

func (o *OledDisplay) ShowINAVVers(t string) {
	o.setPos(6, 2)
	o.d.PrintText(t)
}

func (o *OledDisplay) ShowMode(amode int16, imode int16) {
	o.setPos(6, 3)

	var t string

	switch amode {
	case 0:
		t = "None"
	case 1:
		t = "Init"
	case 2:
		t = "Try"
	case 3:
		t = "Conn"
	default:
		t = "Fail"
	}
	o.d.PrintText(t)
	o.incX(len(t))

	o.d.PrintText(" / ")
	o.incX(3)

	switch imode {
	case 0:
		t = "None"
	case 1:
		t = "PH"
	case 2:
		t = "RTH"
	case 3:
		t = "WP"
	default:
		t = "---"
	}
	o.d.PrintText(t)
	o.incX(len(t))

	o.cEOL()
}

func (o *OledDisplay) ShowINAVSats(nsat uint16, hdop uint16) {
	o.setPos(6, 4)

	t := Uitoa(uint(nsat))
	t = fill(t, 2, false)
	o.d.PrintText(t)
	o.incX(len(t))
	if nsat < 2 {
		t = " sat "
	} else {
		t = " sats "
	}
	o.d.PrintText(t)
	o.incX(len(t))
	t = Uitoa(uint(hdop))
	t = fill(t, 3, true)
	o.d.PrintChar(t[0])
	o.incX(1)
	o.d.PrintChar('.')
	o.incX(1)
	o.d.PrintText(t[1:])
	o.incX(2)
	o.cEOL()
}

func (o *OledDisplay) ShowINAVPos(dist uint, brg uint16) {
	o.setPos(6, 5)
	if dist >= 100000 {
		o.d.PrintText("*****")
		o.incX(5)
	} else if dist >= 10000 {
		k := dist / 100
		t := Uitoa(uint(k))
		t = fill(t, 3, false)
		o.d.PrintText(t[:2])
		o.incX(2)
		o.d.PrintChar('.')
		o.incX(1)
		o.d.PrintText(t[2:])
		o.incX(1)
		o.d.PrintChar('k')
		o.incX(1)
	} else {
		t := Uitoa(uint(dist))
		t = fill(t, 4, false)
		o.d.PrintText(t)
		o.incX(4)
		o.d.PrintChar('m')
		o.incX(1)
	}
	o.d.PrintChar(' ')
	o.incX(1)
	t := Uitoa(uint(brg))
	t = fill(t, 3, true)
	o.d.PrintText(t)
	o.incX(3)
	o.d.PrintChar('*')
}

func (o *OledDisplay) INAVReset() {
	o.ShowINAVVers("?.?.?")
	for j := 3; j < 6; j++ {
		o.setPos(6, j)
		o.cEOL()
	}
}

func (o *OledDisplay) drawSep() {
	y := 1 + 2*FONT_H
	for x := int16(0); x < OLED_WIDTH; x++ {
		o.dev.SetPixel(x, y, color.RGBA{R: 1})
	}
}
