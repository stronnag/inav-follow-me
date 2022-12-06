package vbat

import "machine"

var (
	Vmode   bool
	voffset float32
	vin     machine.ADC
)

func VBatInit(v_mode bool, v_offset float32) {
	Vmode = v_mode
	voffset = v_offset
	if Vmode {
		machine.InitADC()
		vin = machine.ADC{machine.ADC3}
		inp := vin.Get()
		if inp < 1024 {
			gp25 := machine.GPIO25
			gp25.Configure(machine.PinConfig{Mode: machine.PinOutput})
			gp25.High()
		}
	}
}

func VBatRead() (uint16, uint16) {
	ivbat := uint16(0)
	inp := uint16(0)
	if Vmode {
		inp = vin.Get()
		vbat := 9.9*float32(inp)/65535 + voffset
		ivbat = uint16(10 * (vbat + 0.05))
	}
	return ivbat, inp
}
