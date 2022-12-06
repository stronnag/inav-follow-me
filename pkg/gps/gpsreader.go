package gps

import (
	"machine"
	"strconv"
	"strings"
	"time"
)

type Fix struct {
	Quality uint8
	Stamp   time.Time
	Lat     float32
	Lon     float32
	Alt     float32
	Sats    uint8
	Spd     float32
	Hdg     float32
}

type GPSReader struct {
	uart  machine.UART
	fchan chan Fix
	Fix   Fix
	idx   int
	line  []byte
}

var (
	gspdelay time.Duration
)

func NewGPSUartReader(uart machine.UART, fchan chan Fix) *GPSReader {
	line := make([]byte, 128)
	return &GPSReader{uart: uart, fchan: fchan, Fix: Fix{}, line: line}
}

func (g *GPSReader) SetBaud(baud uint32) {
	g.uart.SetBaudRate(baud)
	gspdelay = time.Duration((10 * 1000000 / (2 * baud))) * time.Microsecond
}

func parseLatLon(ll string, nsew string, width int) float32 {
	v := float32(0.0)
	if len(ll) > 4 {
		dd, err := strconv.ParseFloat(ll[0:width], 32)
		if err == nil {
			mm, err := strconv.ParseFloat(ll[width:], 32)
			if err == nil {
				v = float32(dd + (mm / 60))
				if nsew == "S" || nsew == "W" {
					v *= -1
				}
			}
		}
	}
	return v
}

func parseSats(str string) uint8 {
	v, err := strconv.ParseInt(str, 10, 32)
	if err == nil {
		return uint8(v)
	} else {
		return 0
	}
}

func parseF32(str string) float32 {
	v, err := strconv.ParseFloat(str, 32)
	if err == nil {
		return float32(v)
	} else {
		return float32(0.0)
	}
}

func parseTime(str string) time.Time {
	if len(str) < 6 {
		return time.Time{}
	}
	h, _ := strconv.ParseInt(str[0:2], 10, 8)
	m, _ := strconv.ParseInt(str[2:4], 10, 8)
	s, _ := strconv.ParseInt(str[4:6], 10, 8)
	ms := int64(0)
	if len(str) == 9 {
		ms, _ = strconv.ParseInt(str[7:9], 10, 8)
	}
	t := time.Date(0, 0, 0, int(h), int(m), int(s), int(ms), time.UTC)
	return t
}

func valid_nmea(str string) bool {
	if len(str) > 6 && str[0] == '$' && str[len(str)-3] == '*' {
		chk := byte(0)
		for i := 1; i < len(str)-3; i++ {
			chk ^= str[i]
		}
		cs, _ := strconv.ParseInt(str[len(str)-2:len(str)], 16, 8)
		return chk == byte(cs)
	} else {
		return false
	}
}

func (r *GPSReader) parse_nmea(nmea string) bool {
	if valid_nmea(nmea) {
		typ := nmea[3:6]
		last := r.Fix.Stamp
		switch typ {
		case "GGA":
			part := strings.Split(nmea, ",")
			if len(part) != 15 {
				return false
			}
			r.Fix.Stamp = parseTime(part[1])
			r.Fix.Lat = parseLatLon(part[2], part[3], 2)
			r.Fix.Lon = parseLatLon(part[4], part[5], 3)
			r.Fix.Alt = parseF32(part[9])
			r.Fix.Sats = parseSats(part[7])
			r.Fix.Quality = uint8(part[6][0] - 48)
			return r.Fix.Stamp != last
		case "RMC":
			part := strings.Split(nmea, ",")
			if len(part) != 13 {
				return false
			}
			r.Fix.Stamp = parseTime(part[1])
			r.Fix.Lat = parseLatLon(part[3], part[4], 2)
			r.Fix.Lon = parseLatLon(part[5], part[6], 3)
			r.Fix.Spd = parseF32(part[7])
			r.Fix.Hdg = parseF32(part[8])
			return r.Fix.Stamp != last
		default:
		}
	}
	return false
}

func (r *GPSReader) builder(c byte) {
	if c == '$' {
		r.idx = 0
	}
	if r.idx == 127 {
		r.idx = 0
	}
	if c == 0xd {
		return
	}
	if c == 0xa {
		if r.parse_nmea(string(r.line[:r.idx])) {
			r.fchan <- r.Fix
		}
		r.idx = 0
	} else {
		r.line[r.idx] = c
		r.idx += 1
	}
}

func (r *GPSReader) UartReader() {
	for {
		if r.uart.Buffered() > 0 {
			c, err := r.uart.ReadByte()
			if err == nil {
				r.builder(c)
			} else {
				println(err)
			}
		} else {
			time.Sleep(gspdelay)
		}
	}
}
