/*
 */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"go.bug.st/serial"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of gpsrd [options] file\n")
		fmt.Fprintf(os.Stderr, " where \"file\" is a file containing NMEA sentences\n")
		flag.PrintDefaults()
	}

	device := ""
	baud := 9600

	flag.StringVar(&device, "device", "", "Serial device")
	flag.IntVar(&baud, "baud", 9600, "Baud rate")
	flag.Parse()

	last := 0.0
	var port serial.Port
	files := flag.Args()
	if len(files) > 0 {
		file, err := os.Open(files[0])
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		if device != "" {
			mode := &serial.Mode{
				BaudRate: baud,
			}
			port, err = serial.Open(device, mode)
			if err != nil {
				panic(err)
			}
		}

		fh := bufio.NewReader(file)
		scanner := bufio.NewScanner(fh)
		if scanner != nil {
			for scanner.Scan() {
				l := scanner.Text()
				parts := strings.Split(l, ",")
				if len(parts) > 2 {
					if parts[0] == "$GPGGA" {
						now, _ := strconv.ParseFloat(parts[1], 32)
						if last != 0 {
							diff := (now - last) * 1000
							if diff > 0 {
								time.Sleep(time.Duration(diff) * time.Millisecond)
							}
						}
						last = now
					}
					fmt.Println(l)
					if device != "" {
						_, err := port.Write([]byte(l))
						if err == nil {
							_, err = port.Write([]byte("\r\n"))
						}
						if err != nil {
							log.Fatal(err)
						}
					}
				}
			}
		} else {
			log.Fatal(err)
		}
	} else {
		flag.Usage()
	}
}
