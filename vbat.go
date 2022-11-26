package main

import "machine"

const (
	VBAT_NONE = iota
	VBAT_PICO
	VBAT_PICO_W
)

var vin machine.ADC

func VBatInit() {
	if VBAT_MODE != VBAT_NONE {
		machine.InitADC()
		vin = machine.ADC{machine.ADC3}
		if VBAT_MODE == VBAT_PICO_W {
			gp25 := machine.GPIO25
			gp25.Configure(machine.PinConfig{Mode: machine.PinOutput})
			gp25.High()
		}
	}
}

func VBatRead() (uint16, uint16) {
	ivbat := uint16(0)
	inp := uint16(0)
	if VBAT_MODE != VBAT_NONE {
		inp = vin.Get()
		vbat := 9.9*float32(inp)/65535 + VBAT_OFFSET
		ivbat = uint16(10 * (vbat + 0.05))
	}
	return ivbat, inp
}
