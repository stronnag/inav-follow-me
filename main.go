package main

import (
	"encoding/binary"
	"machine"
	"time"
	"tinygo.org/x/drivers/ssd1306"
)

import (
	"geo"
	"gps"
	"msp"
	"oled"
	"vbat"
)

const (
	VERSION = "v1.2.0"
)

const (
	MW_GPS_MODE_HOLD = 1
	GPS_TIMEOUT      = 600 // (in 0.1 seconds)
	MSP_TIMEOUT      = 600 // (in 0.1 seconds)
	NAV_TIMEOUT      = 100 // (in 0.1 seconds)
	SPLASH_TIMEOUT   = 50  // (5 seconds)
)

const (
	msp_INIT_NONE = 0 + iota
	msp_INIT_INIT
	msp_INIT_WIP
	msp_INIT_DONE
	msp_INIT_FAIL
)

const (
	HOME_WP   = 0
	FOLLOW_WP = 255
)

func main() {
	uart0 := machine.UART0
	uart0.Configure(machine.UARTConfig{
		TX: machine.UART0_TX_PIN,
		RX: machine.UART0_RX_PIN,
	})
	uart0.SetBaudRate(GPSBAUD)
	fchan := make(chan gps.Fix)

	uart1 := machine.UART1
	uart1.Configure(machine.UARTConfig{
		TX: machine.UART1_TX_PIN,
		RX: machine.UART1_RX_PIN,
	})
	uart1.SetBaudRate(MSPBAUD)
	mchan := make(chan msp.MSPMsg)

	machine.I2C1.Configure(machine.I2CConfig{Frequency: 400000,
		SDA: machine.GP26, SCL: machine.GP27})
	dev := ssd1306.NewI2C(machine.I2C1)

	dev.Configure(ssd1306.Config{Width: oled.OLED_WIDTH, Height: oled.OLED_HEIGHT,
		Address: 0x3C, VccState: ssd1306.SWITCHCAPVCC})
	dev.ClearBuffer()

	vbat.VBatInit(USE_VBAT, VBAT_OFFSET)

	o := oled.NewOLED(dev)
	g := gps.NewGPSUartReader(*uart0, fchan, GPSBAUD)
	m := msp.NewMSPUartReader(*uart1, mchan, MSPBAUD)

	o.SplashScreen(VERSION)
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
				if ttick == SPLASH_TIMEOUT {
					println("Initialised")
					o.InitScreen()
					mspinit = msp_INIT_INIT
				}
			} else {
				if ttick%10 == 0 {
					vin, _ := vbat.VBatRead()
					o.ShowVBat(vin)
					o.ShowMode(int16(mspinit), int16(mspmode))
				}
			}

			if ttick-gtick > GPS_TIMEOUT {
				println("*** GPS timeout ***")
				gtick = ttick
				o.ClearTime(true)
				o.ShowGPS(0, 0)
				o.ClearRow(oled.OLED_ROW_VPOS, oled.OLED_EXTRA_SPACE)
			}

			if mspinit == msp_INIT_WIP && ttick-mtick > MSP_TIMEOUT {
				println("*** MSP INIT timeout ***")
				mtick = ttick
				mspinit = msp_INIT_INIT
			}

			if mspinit == msp_INIT_DONE {
				if ttick-mtick > NAV_TIMEOUT {
					println("*** MSP NAV TIMEOUT ***")
					o.INAVReset()
					mspinit = msp_INIT_INIT
					mspmode = 0
				} else {
					m.MSPCommand(msp.MSP_NAV_STATUS, nil)
				}
			}

		case fix := <-fchan:
			if mspinit != msp_INIT_NONE {
				gtick = ttick
				ts := fix.Stamp.Format(GPS_TIME_FORMAT)
				o.ShowTime(ts)
				o.ShowGPS(uint16(fix.Sats), fix.Quality)
				print(ts)
				print(" [", mspinit, ":", mspmode, "]")
				println(" Qual: ", fix.Quality, " sats: ", fix.Sats, " lat: ", fix.Lat, " lon: ", fix.Lon)
				if fix.Quality > 0 && fix.Sats >= GPSMINSAT {
					if mspinit == msp_INIT_INIT {
						println("Starting MSP")
						mspinit = msp_INIT_WIP
						m.MSPCommand(msp.MSP_FC_VARIANT, nil)
					} else if mspinit == msp_INIT_DONE {
						if mspmode == MW_GPS_MODE_HOLD && !(fix.Lat == 0.0 && fix.Lon == 0.0) {
							c, d := geo.Csedist(msplat, msplon, fix.Lat, fix.Lon)
							println("Follow (v->u)", msplat, msplon, fix.Lat, fix.Lon, " dist:", int(d), "m", "Brg: ", int(c))
							if d > MIN_FOLLOW_DIST {
								m.Update_WP(FOLLOW_WP, fix.Lat, fix.Lon, uint16(c))
								println("Vehicle:", c, d)
								o.ShowINAVPos(uint(d), uint16(c))
								if RESET_HOME {
									m.Update_WP(HOME_WP, fix.Lat, fix.Lon, uint16(c))
								}
							}
						}
					}
				} else {
					mspinit = msp_INIT_INIT
					mspmode = 0
					o.ClearRow(oled.OLED_ROW_VPOS, oled.OLED_EXTRA_SPACE)
					o.ClearRow(oled.OLED_ROW_VSAT, oled.OLED_EXTRA_SPACE)
				}
			}
		case v := <-mchan:
			if v.Ok {
				mtick = ttick
				switch v.Cmd {
				case msp.MSP_FC_VARIANT:
					vers := string(v.Data[0:4])
					println("Firmware: ", vers)
					if vers == "INAV" {
						m.MSPCommand(msp.MSP_FC_VERSION, nil)
					}
				case msp.MSP_FC_VERSION:
					vbuf := make([]byte, 5)
					vbuf[0] = v.Data[0] + 48
					vbuf[1] = '.'
					vbuf[2] = v.Data[1] + 48
					vbuf[3] = '.'
					vbuf[4] = v.Data[2] + 48
					println("Version: ", string(vbuf))
					o.ShowINAVVers(string(vbuf))
					m.MSPCommand(msp.MSP_NAME, nil)

				case msp.MSP_NAME:
					if v.Len > 0 {
						println("Name: ", string(v.Data))
					}
					m.MSPCommand(msp.MSP2_INAV_MIXER, nil)

				case msp.MSP2_INAV_MIXER:
					ptype := binary.LittleEndian.Uint16(v.Data[3:5])
					println("Platform type: ", ptype)
					if ptype != DONT_FOLLOW_TYPE {
						mspinit = msp_INIT_DONE
						mloop = 0
					} else {
						mspinit = msp_INIT_FAIL
					}
				case msp.MSP_NAV_STATUS:
					if mspmode != v.Data[0] {
						mspmode = v.Data[0]
						println("nav status: ", mspmode)
						if v.Data[0] == 0 {
							o.ClearRow(oled.OLED_ROW_VPOS, oled.OLED_EXTRA_SPACE)
						}
						o.ShowMode(int16(mspinit), int16(mspmode))
					}
					m.MSPCommand(msp.MSP_RAW_GPS, nil)

				case msp.MSP_RAW_GPS:
					mfix := v.Data[0]
					msat := v.Data[1]
					msplat = float32(int32(binary.LittleEndian.Uint32(v.Data[2:6]))) / 1e7
					msplon = float32(int32(binary.LittleEndian.Uint32(v.Data[6:10]))) / 1e7
					alt := int16(binary.LittleEndian.Uint16(v.Data[10:12]))
					spd := float32(binary.LittleEndian.Uint16(v.Data[12:14])) / 100.0
					cog := float32(binary.LittleEndian.Uint16(v.Data[14:16])) / 10.0
					hdop := uint16(999)
					if len(v.Data) > 16 {
						hdop = binary.LittleEndian.Uint16(v.Data[16:18])
					}

					if mloop%10 == 0 {
						o.ShowINAVSats(uint16(msat), hdop)
						if mloop%100 == 0 {
							println("MSP: fix:", mfix, " sats:", msat, " lat:", msplat, " lon:", msplon,
								" alt:", alt, " spd", spd, " cog: ", cog, " hdop:", hdop)
						}
					}
					if mspinit == msp_INIT_DONE {
						mloop += 1
					}

				case msp.MSP_SET_WP:
					println("set wp ACK")
				default:
					println("** msp cmd: ", v.Cmd, " ***")
				}
			}
		}
	}
}
