package main

import (
	"encoding/binary"
	"machine"
	"time"
)

const (
	MW_GPS_MODE_HOLD = 1
)

const (
	msp_INIT_NONE = 0 + iota
	msp_INIT_INIT
	msp_INIT_WIP
	msp_INIT_DONE
	msp_INIT_FAIL
)

func main() {
	uart0 := machine.UART0
	uart0.Configure(machine.UARTConfig{
		TX: machine.UART0_TX_PIN,
		RX: machine.UART0_RX_PIN,
	})
	uart0.SetBaudRate(GPSBAUD)
	fchan := make(chan Fix)

	uart1 := machine.UART1
	uart1.Configure(machine.UARTConfig{
		TX: machine.UART1_TX_PIN,
		RX: machine.UART1_RX_PIN,
	})
	uart1.SetBaudRate(MSPBAUD)
	mchan := make(chan MSPMsg)

	g := NewGPSUartReader(*uart0, fchan)
	go g.UartReader()

	m := NewMSPUartReader(*uart1, mchan)
	go m.UartReader()

	mspinit := msp_INIT_NONE
	mspmode := byte(0)
	msplat := float32(0)
	msplon := float32(0)
	itimer := time.NewTimer(2 * time.Second)
	mtimer := time.NewTimer(100 * time.Millisecond)
	mloop := 0

	for {
		select {
		case <-itimer.C:
			if mspinit == msp_INIT_NONE {
				println("Initialised")
			}
			mspinit = msp_INIT_INIT

		case <-mtimer.C:
			if mspinit != msp_INIT_NONE {
				m.MSPCommand(MSP_NAV_STATUS, nil)
				mtimer.Reset(100 * time.Millisecond)
			}

		case fix := <-fchan:
			if mspinit != msp_INIT_NONE {
				print(fix.Stamp.Format("15:04:05"))
				print(" [", mspinit, ":", mspmode, "]")
				println(" Qual: ", fix.Quality, " sats: ", fix.Sats, " lat: ", fix.Lat, " lon: ", fix.Lon)
				if fix.Quality > 0 && fix.Sats >= GPSMINSAT {
					if mspinit == msp_INIT_INIT {
						println("Starting MSP")
						mspinit = msp_INIT_WIP
						itimer.Reset(1 * time.Minute)
						m.MSPCommand(MSP_FC_VARIANT, nil)
					} else if mspinit == msp_INIT_DONE {
						if mspmode == MW_GPS_MODE_HOLD {
							c, d := m.Followme(fix.Lat, fix.Lon, msplat, msplon)
							println("Vehicle:", c, d)
						}
					}
				} else {
					mspinit = msp_INIT_INIT
				}
			}
		case v := <-mchan:
			if v.ok {
				switch v.cmd {
				case MSP_FC_VARIANT:
					vers := string(v.data[0:4])
					println("Firmware: ", vers)
					if vers == "INAV" {
						m.MSPCommand(MSP_FC_VERSION, nil)
					}
				case MSP_FC_VERSION:
					vbuf := make([]byte, 5)
					vbuf[0] = v.data[0] + 48
					vbuf[1] = '.'
					vbuf[2] = v.data[1] + 48
					vbuf[3] = '.'
					vbuf[4] = v.data[2] + 48
					println("Version: ", string(vbuf))
					m.MSPCommand(MSP_NAME, nil)

				case MSP_NAME:
					if v.len > 0 {
						println("Name: ", string(v.data))
					}
					m.MSPCommand(MSP2_INAV_MIXER, nil)

				case MSP2_INAV_MIXER:
					ptype := binary.LittleEndian.Uint16(v.data[3:5])
					println("Platform type: ", ptype)
					itimer.Stop()
					if ptype != DONT_FOLLOW_TYPE {
						mspinit = msp_INIT_DONE
						m.MSPCommand(MSP_NAV_STATUS, nil)
						mtimer.Reset(100 * time.Millisecond)
						mloop = 0
					} else {
						mspinit = msp_INIT_FAIL
					}
				case MSP_NAV_STATUS:
					if mloop%100 == 0 {
						println("nav status: ", mspmode)
					}
					mspmode = v.data[0]
					m.MSPCommand(MSP_RAW_GPS, nil)

				case MSP_RAW_GPS:
					mfix := v.data[0]
					msat := v.data[1]
					msplat = float32(int32(binary.LittleEndian.Uint32(v.data[2:6]))) / 1e7
					msplon = float32(int32(binary.LittleEndian.Uint32(v.data[6:10]))) / 1e7
					alt := int16(binary.LittleEndian.Uint16(v.data[10:12]))
					spd := float32(binary.LittleEndian.Uint16(v.data[12:14])) / 100.0
					cog := float32(binary.LittleEndian.Uint16(v.data[14:16])) / 10.0
					hdop := float32(99.9)
					if len(v.data) > 16 {
						hdop = float32(binary.LittleEndian.Uint16(v.data[16:18])) / 100.0
					}

					if mloop%100 == 0 {
						println("MSP: fix:", mfix, " sats:", msat, " lat:", msplat, " lon:", msplon,
							" alt:", alt, " spd", spd, " cog: ", cog, " hdop:", hdop)
					}
					if mspinit == msp_INIT_DONE {
						mloop += 1
					}

				case MSP_SET_WP:
					println("set wp")
				default:
					println("** msp cmd: ", v.cmd, " ***")
				}
			}
		}
	}
}
