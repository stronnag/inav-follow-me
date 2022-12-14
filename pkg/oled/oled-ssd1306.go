package oled

import (
	font "github.com/Nondzu/ssd1306_font"
	"image/color"
	"strconv"
	"tinygo.org/x/drivers/ssd1306"
)

var FONT_H int16 = 10
var FONT_W int16 = 7

const OLED_WIDTH = 128
const OLED_HEIGHT = 64
const OLED_EXTRA_SPACE = 3

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
	d     font.Display
	dev   ssd1306.Device
	vmode bool
}

func NewOLED(dev ssd1306.Device) *OledDisplay {
	return &OledDisplay{d: font.NewDisplay(dev), dev: dev}
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

func (o *OledDisplay) setPos(x, y, offset int) {
	o.d.XPos = int16(x) * FONT_W
	o.d.YPos = int16(y) * FONT_H
	o.d.YPos += int16(offset)
}

func (o *OledDisplay) ClearTime(fail bool) {
	var str string
	o.setPos(0, OLED_ROW_TIME, 0)
	o.cEOL()
	if fail {
		str = "??:??:??"
	} else {
		str = "--:--:--"
	}
	o.ShowTime(str)
}

func (o *OledDisplay) SplashScreen(version string) {
	o.d.Configure(font.Config{FontType: font.FONT_11x18})
	FONT_W = 11
	FONT_H = 18
	o.CentreString("INAV", 0, 4)
	o.CentreString("Follow Me!", 1, 4)
	o.CentreString(version, 2, 4)
}

func (o *OledDisplay) InitScreen(vb bool) {
	o.dev.ClearDisplay()
	o.d.Configure(font.Config{FontType: font.FONT_7x10})
	FONT_W = 7
	FONT_H = 10
	o.vmode = vb

	o.ClearTime(false)

	o.setPos(0, OLED_ROW_GPS, 0)
	o.d.PrintText("GPS :")

	o.drawSep()

	o.setPos(0, OLED_ROW_MODE, OLED_EXTRA_SPACE)
	o.d.PrintText("Mode:")

	o.setPos(0, OLED_ROW_INAV, OLED_EXTRA_SPACE)
	o.d.PrintText("INAV: -.-.-")

	o.setPos(0, OLED_ROW_VSAT, OLED_EXTRA_SPACE)
	o.d.PrintText("VSat:")

	o.setPos(0, OLED_ROW_VPOS, OLED_EXTRA_SPACE)
	o.d.PrintText("VPos:")
}

func (o *OledDisplay) CentreString(t string, row int, offset int) {
	o.d.XPos = (OLED_WIDTH - FONT_W*int16(len(t))) / 2
	o.d.YPos = int16(row)*FONT_H + int16(offset)
	o.d.PrintText(t)
	o.incX(len(t))
	o.cEOL()
}

func (o *OledDisplay) ShowTime(t string) {
	if !o.vmode {
		o.CentreString(t, OLED_ROW_TIME, 0)
	} else {
		o.setPos(0, OLED_ROW_TIME, 0)
		o.d.PrintText(t)
	}
}

func (o *OledDisplay) ShowGPS(nsat uint16, fix uint8) {
	o.setPos(6, OLED_ROW_GPS, 0)
	t := strconv.FormatUint(uint64(nsat), 10)
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
	o.setPos(6, OLED_ROW_INAV, OLED_EXTRA_SPACE)
	o.d.PrintText(t)
}

func (o *OledDisplay) ShowMode(amode int16, imode int16) {
	o.setPos(6, OLED_ROW_MODE, OLED_EXTRA_SPACE)
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

	o.setPos(12, OLED_ROW_INAV, OLED_EXTRA_SPACE)
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
	o.setPos(6, OLED_ROW_VSAT, OLED_EXTRA_SPACE)

	t := strconv.FormatUint(uint64(nsat), 10)
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
	kv := float64(hdop) / 100
	t = strconv.FormatFloat(kv, 'f', 1, 32)
	o.d.PrintText(t)
	o.incX(3)
	o.cEOL()
}

func (o *OledDisplay) ShowINAVPos(dist uint, brg uint16) {
	o.setPos(6, OLED_ROW_VPOS, OLED_EXTRA_SPACE)
	if dist >= 100000 {
		o.d.PrintText(">100k")
		o.incX(5)
	} else if dist >= 10000 {
		kv := float64(dist) / 1000.0
		t := strconv.FormatFloat(kv, 'f', 1, 32)
		t = fill(t, 4, false)
		o.d.PrintText(t)
		o.incX(4)
		o.d.PrintChar('k')
		o.incX(1)
	} else {
		t := strconv.FormatUint(uint64(dist), 10)
		t = fill(t, 4, false)
		o.d.PrintText(t)
		o.incX(4)
		o.d.PrintChar('m')
		o.incX(1)
	}
	o.d.PrintChar(' ')
	o.incX(1)
	t := strconv.FormatUint(uint64(brg), 10)
	t = fill(t, 3, true)
	o.d.PrintText(t)
	o.incX(3)
	o.d.PrintChar('*')
}

func (o *OledDisplay) ClearRow(row, offset int) {
	o.setPos(6, row, offset)
	o.cEOL()
}

func (o *OledDisplay) INAVReset() {
	o.ShowINAVVers("?.?.?")
	for j := OLED_ROW_MODE; j < OLED_ROW_COUNT; j++ {
		o.ClearRow(j, 3)
	}
}

func (o *OledDisplay) drawSep() {
	y := 1 + 2*FONT_H
	for x := int16(0); x < OLED_WIDTH; x++ {
		o.dev.SetPixel(x, y, color.RGBA{R: 1})
	}
}

func (o *OledDisplay) ShowVBat(vin uint16) {
	vs := make([]byte, 4)
	vs[0] = '0' + byte(vin/10)
	vs[1] = '.'
	vs[2] = '0' + byte(vin%10)
	vs[3] = 'V'
	o.setPos(14, OLED_ROW_TIME, 0)
	t := string(vs)
	o.d.PrintText(t)
}
