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

const (
	OLED_ROW_TIME = iota
	OLED_ROW_GPS
	OLED_ROW_MODE
	OLED_ROW_INAV
	OLED_ROW_VSAT
	OLED_ROW_VPOS
	OLED_ROW_COUNT
)

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
	if y > OLED_ROW_GPS {
		o.d.YPos += 3
	}
}

func (o *OledDisplay) ClearTime(fail bool) {
	var str string
	o.setPos(0, OLED_ROW_TIME)
	o.cEOL()
	if fail {
		str = "??:??:??"
	} else {
		str = "--:--:--"
	}
	o.ShowTime(str)
}

func (o *OledDisplay) InitScreen() {
	o.ClearTime(false)

	o.setPos(0, OLED_ROW_GPS)
	o.d.PrintText("GPS :")

	o.drawSep()

	o.setPos(0, OLED_ROW_MODE)
	o.d.PrintText("Mode:")

	o.setPos(0, OLED_ROW_INAV)
	o.d.PrintText("INAV: -.-.-")

	o.setPos(0, OLED_ROW_VSAT)
	o.d.PrintText("VSat:")

	o.setPos(0, OLED_ROW_VPOS)
	o.d.PrintText("VPos:")
}

func (o *OledDisplay) ShowTime(t string) {
	xpos := (18 - len(t)) / 2
	o.setPos(xpos, OLED_ROW_TIME)
	o.d.PrintText(t)
	o.incX(len(t))
	o.cEOL()
}

func (o *OledDisplay) ShowGPS(nsat uint16, fix uint8) {
	o.setPos(6, OLED_ROW_GPS)
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
	o.setPos(6, OLED_ROW_INAV)
	o.d.PrintText(t)
}

func (o *OledDisplay) ShowMode(amode int16, imode int16) {
	o.setPos(6, OLED_ROW_MODE)
	var t string

	switch amode {
	case 0:
		t = "Starting"
	case 1:
		t = "Initialised"
	case 2:
		t = "Connecting"
	case 3:
		t = "Connected"
	default:
		t = "Failed"
	}
	o.d.PrintText(t)
	o.incX(len(t))
	o.cEOL()

	o.setPos(12, OLED_ROW_INAV)
	switch imode {
	case 0:
		t = "Idle"
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
	o.setPos(6, OLED_ROW_VSAT)

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
	o.setPos(6, OLED_ROW_VPOS)
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

func (o *OledDisplay) ClearRow(row int) {
	o.setPos(6, row)
	o.cEOL()
}

func (o *OledDisplay) INAVReset() {
	o.ShowINAVVers("?.?.?")
	for j := OLED_ROW_MODE; j < OLED_ROW_COUNT; j++ {
		o.ClearRow(j)
	}
}

func (o *OledDisplay) drawSep() {
	y := 1 + 2*FONT_H
	for x := int16(0); x < OLED_WIDTH; x++ {
		o.dev.SetPixel(x, y, color.RGBA{R: 1})
	}
}
