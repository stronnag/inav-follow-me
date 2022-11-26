package vbat

import "machine"

const (
	VBAT_NONE = iota
	VBAT_PICO
	VBAT_PICO_W
)

var vin machine.ADC
var Vmode int
var voffset float32

func VBatInit(v_mode int, v_offset float32) {
	Vmode = v_mode
	voffset = v_offset
	if Vmode != VBAT_NONE {
		machine.InitADC()
		vin = machine.ADC{machine.ADC3}
		if Vmode == VBAT_PICO_W {
			gp25 := machine.GPIO25
			gp25.Configure(machine.PinConfig{Mode: machine.PinOutput})
			gp25.High()
		}
	}
}

func VBatRead() (uint16, uint16) {
	ivbat := uint16(0)
	inp := uint16(0)
	if Vmode != VBAT_NONE {
		inp = vin.Get()
		vbat := 9.9*float32(inp)/65535 + voffset
		ivbat = uint16(10 * (vbat + 0.05))
	}
	return ivbat, inp
}
