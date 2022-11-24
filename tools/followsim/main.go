package main

import (
	"flag"
	"fmt"
	"github.com/eiannone/keyboard"
	"log"
	"math/rand"
	"os"
	"time"
)

type SChan struct {
	len  uint16
	cmd  uint16
	ok   bool
	data []byte
}

var BaseLat float64
var BaseLon float64

func main() {
	BaseLat = -90.0 + rand.Float64()*180.0
	BaseLon = -180.0 + rand.Float64()*360.0

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of followsim [options] device\n")
		flag.PrintDefaults()
	}

	flag.Float64Var(&BaseLat, "lat", 0, "Base latitude")
	flag.Float64Var(&BaseLon, "lon", 0, "Base longitude")
	flag.Parse()

	var sp *MSPSerial
	port := ""
	c0 := make(chan SChan)

	files := flag.Args()
	if len(files) > 0 {
		port = files[0]
	} else {
		port = Enumerate_ports()
	}
	if port != "" {
		fmt.Printf("Using %s\n", port)
		sp = NewMSPSerial(port)
		go sp.Reader(c0)
	} else {
		log.Fatalln("No serial device given or detected")
	}

	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	armed := false
	poshold := false

	for done := false; !done; {
		select {
		case ev := <-keysEvents:
			if ev.Err != nil {
				panic(ev.Err)
			}
			if ev.Key == 0 {
				switch ev.Rune {
				case 'a', 'A':
					armed = !armed
					fmt.Printf("Armed: %v\n", armed)
					poshold = false
				case 'p', 'P':
					poshold = !poshold
					fmt.Printf("Poshold: %v\n", poshold)
				case 'Q', 'q':
					done = true
				default:
				}
			} else {
				if ev.Key == keyboard.KeyCtrlC {
					done = true
				} else if ev.Key == keyboard.KeyEnter {
					fmt.Println()
				}
			}
		case v := <-c0:
			st := time.Now()
			fmt.Printf("%s ", st.Format("15:04:05"))
			switch v.cmd {
			case MSP_FC_VARIANT:
				fmt.Println("send varient")
				sp.SendVariant()
			case MSP_FC_VERSION:
				fmt.Println("send version")
				sp.SendVersion()
			case MSP2_INAV_MIXER:
				fmt.Println("send mixer")
				sp.SendMixer()
			case MSP_NAME:
				fmt.Println("send name")
				sp.SendName()
			case MSP_RAW_GPS:
				fmt.Println("send GPS")
				sp.SendGPS(armed)
			case MSP_SET_WP:
				fmt.Println("Set WP")
				sp.deserialise_wp(v.data)
			case MSP_NAV_STATUS:
				fmt.Printf("send nav status (arm %v, hold %v)\n", armed, poshold)
				sp.SendStatus(armed, poshold)
			case 0:
				done = true
			default:
				fmt.Printf("Unexpected MSP %d (0x%x)\n", v.cmd, v.cmd)
				sp.SendAckNak(v.cmd, false)
			}
		}
	}
}
