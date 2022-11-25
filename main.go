package main

import (
	"encoding/binary"
	"machine"
	"time"
	"tinygo.org/x/drivers/ssd1306"
)

const (
	MW_GPS_MODE_HOLD = 1
	GPS_TIMEOUT      = 600 // (in 0.1 seconds)
	MSP_TIMEOUT      = 600 // (in 0.1 seconds)
	NAV_TIMEOUT      = 100 // (in 0.1 seconds)
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

	machine.I2C1.Configure(machine.I2CConfig{Frequency: 400000,
		SDA: machine.GP26, SCL: machine.GP27})
	dev := ssd1306.NewI2C(machine.I2C1)

	dev.Configure(ssd1306.Config{Width: OLED_WIDTH, Height: OLED_HEIGHT,
		Address: 0x3C, VccState: ssd1306.SWITCHCAPVCC})
	dev.ClearBuffer()

	oled := NewOLED(dev)
	g := NewGPSUartReader(*uart0, fchan)
	m := NewMSPUartReader(*uart1, mchan)

	oled.SplashScreen()
	go g.UartReader()
	go m.UartReader()

	mspinit := msp_INIT_NONE
	mspmode := byte(0)
	msplat := float32(0)
	msplon := float32(0)
	ticker := time.NewTicker(100 * time.Millisecond)
	mloop := 0
	ttick := 0
	gtick := 0
	mtick := 0

	for {
		select {
		case <-ticker.C:
			ttick += 1

			if mspinit == msp_INIT_NONE {
				if ttick == 50 {
					println("Initialised")
					oled.InitScreen()
					mspinit = msp_INIT_INIT
				}
			} else {
				if ttick%10 == 0 {
					oled.ShowMode(int16(mspinit), int16(mspmode))
				}
			}

			if ttick-gtick > GPS_TIMEOUT {
				println("*** GPS timeout ***")
				gtick = ttick
				oled.ClearTime(true)
				oled.ShowGPS(0, 0)
				oled.ClearRow(OLED_ROW_VPOS, OLED_EXTRA_SPACE)
			}

			if mspinit == msp_INIT_WIP && ttick-mtick > MSP_TIMEOUT {
				println("*** MSP INIT timeout ***")
				mtick = ttick
				mspinit = msp_INIT_INIT
			}

			if mspinit == msp_INIT_DONE {
				if ttick-mtick > NAV_TIMEOUT {
					println("*** MSP NAV TIMEOUT ***")
					oled.INAVReset()
					mspinit = msp_INIT_INIT
					mspmode = 0
				} else {
					m.MSPCommand(MSP_NAV_STATUS, nil)
				}
			}

		case fix := <-fchan:
			if mspinit != msp_INIT_NONE {
				gtick = ttick
				ts := fix.Stamp.Format(GPS_TIME_FORMAT)
				oled.ShowTime(ts)
				oled.ShowGPS(uint16(fix.Sats), fix.Quality)
				print(ts)
				print(" [", mspinit, ":", mspmode, "]")
				println(" Qual: ", fix.Quality, " sats: ", fix.Sats, " lat: ", fix.Lat, " lon: ", fix.Lon)
				if fix.Quality > 0 && fix.Sats >= GPSMINSAT {
					if mspinit == msp_INIT_INIT {
						println("Starting MSP")
						mspinit = msp_INIT_WIP
						m.MSPCommand(MSP_FC_VARIANT, nil)
					} else if mspinit == msp_INIT_DONE {
						if mspmode == MW_GPS_MODE_HOLD && !(fix.Lat == 0.0 && fix.Lon == 0.0) {
							c, d := m.Followme(fix.Lat, fix.Lon, msplat, msplon)
							println("Vehicle:", c, d)
							oled.ShowINAVPos(uint(d), uint16(c))
						}
					}
				} else {
					mspinit = msp_INIT_INIT
					mspmode = 0
					oled.ClearRow(OLED_ROW_VPOS, OLED_EXTRA_SPACE)
					oled.ClearRow(OLED_ROW_VSAT, OLED_EXTRA_SPACE)
				}
			}
		case v := <-mchan:
			if v.ok {
				mtick = ttick
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
					oled.ShowINAVVers(string(vbuf))
					m.MSPCommand(MSP_NAME, nil)

				case MSP_NAME:
					if v.len > 0 {
						println("Name: ", string(v.data))
					}
					m.MSPCommand(MSP2_INAV_MIXER, nil)

				case MSP2_INAV_MIXER:
					ptype := binary.LittleEndian.Uint16(v.data[3:5])
					println("Platform type: ", ptype)
					if ptype != DONT_FOLLOW_TYPE {
						mspinit = msp_INIT_DONE
						mloop = 0
					} else {
						mspinit = msp_INIT_FAIL
					}
				case MSP_NAV_STATUS:
					if mspmode != v.data[0] {
						mspmode = v.data[0]
						println("nav status: ", mspmode)
						if v.data[0] == 0 {
							oled.ClearRow(OLED_ROW_VPOS, OLED_EXTRA_SPACE)
						}
						oled.ShowMode(int16(mspinit), int16(mspmode))
					}
					m.MSPCommand(MSP_RAW_GPS, nil)

				case MSP_RAW_GPS:
					mfix := v.data[0]
					msat := v.data[1]
					msplat = float32(int32(binary.LittleEndian.Uint32(v.data[2:6]))) / 1e7
					msplon = float32(int32(binary.LittleEndian.Uint32(v.data[6:10]))) / 1e7
					alt := int16(binary.LittleEndian.Uint16(v.data[10:12]))
					spd := float32(binary.LittleEndian.Uint16(v.data[12:14])) / 100.0
					cog := float32(binary.LittleEndian.Uint16(v.data[14:16])) / 10.0
					hdop := uint16(999)
					if len(v.data) > 16 {
						hdop = binary.LittleEndian.Uint16(v.data[16:18])
					}

					if mloop%10 == 0 {
						oled.ShowINAVSats(uint16(msat), hdop)
						if mloop%100 == 0 {
							println("MSP: fix:", mfix, " sats:", msat, " lat:", msplat, " lon:", msplon,
								" alt:", alt, " spd", spd, " cog: ", cog, " hdop:", hdop)
						}
					}
					if mspinit == msp_INIT_DONE {
						mloop += 1
					}

				case MSP_SET_WP:
					println("set wp ACK")
				default:
					println("** msp cmd: ", v.cmd, " ***")
				}
			}
		}
	}
}
