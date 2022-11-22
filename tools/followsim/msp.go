package main

import (
	"encoding/binary"
	"fmt"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
	"log"
	"math/rand"
	"os"
)

const (
	MSP_FC_VARIANT  uint16 = 2
	MSP_FC_VERSION  uint16 = 3
	MSP2_INAV_MIXER uint16 = 0x2010
	MSP_NAME        uint16 = 10
	MSP_RAW_GPS     uint16 = 106
	MSP_SET_WP      uint16 = 209
	MSP_NAV_STATUS  uint16 = 121
)

const (
	state_INIT = iota
	state_M
	state_X_HEADER2
	state_X_FLAGS
	state_X_ID1
	state_X_ID2
	state_X_LEN1
	state_X_LEN2
	state_X_DATA
	state_X_CHECKSUM
)

type SerDev interface {
	Read(buf []byte) (int, error)
	Write(buf []byte) (int, error)
	Close() error
}

type MSPSerial struct {
	SerDev
}

func crc8_dvb_s2(crc byte, a byte) byte {
	crc ^= a
	for i := 0; i < 8; i++ {
		if (crc & 0x80) != 0 {
			crc = (crc << 1) ^ 0xd5
		} else {
			crc = crc << 1
		}
	}
	return crc
}

func (m *MSPSerial) MSPCommand(cmd uint16, payload []byte) {
	rb := encode_msp2(cmd, payload)
	_, err := m.Write(rb)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *MSPSerial) Reader(c0 chan SChan) {
	inp := make([]byte, 128)
	var count = uint16(0)
	var crc = byte(0)
	var sc SChan

	n := state_INIT
	for {
		nb, err := p.Read(inp)
		if err == nil && nb > 0 {
			for i := 0; i < nb; i++ {
				switch n {
				case state_INIT:
					if inp[i] == '$' {
						n = state_M
						sc.ok = false
						sc.len = 0
						sc.cmd = 0
					}
				case state_M:
					if inp[i] == 'X' {
						n = state_X_HEADER2
					} else {
						n = state_INIT
					}
				case state_X_HEADER2:
					if inp[i] == '!' {
						n = state_X_FLAGS
					} else if inp[i] == '<' {
						n = state_X_FLAGS
						sc.ok = true
					} else {
						n = state_INIT
					}

				case state_X_FLAGS:
					crc = crc8_dvb_s2(0, inp[i])
					n = state_X_ID1

				case state_X_ID1:
					crc = crc8_dvb_s2(crc, inp[i])
					sc.cmd = uint16(inp[i])
					n = state_X_ID2

				case state_X_ID2:
					crc = crc8_dvb_s2(crc, inp[i])
					sc.cmd |= (uint16(inp[i]) << 8)
					n = state_X_LEN1

				case state_X_LEN1:
					crc = crc8_dvb_s2(crc, inp[i])
					sc.len = uint16(inp[i])
					n = state_X_LEN2

				case state_X_LEN2:
					crc = crc8_dvb_s2(crc, inp[i])
					sc.len |= (uint16(inp[i]) << 8)
					if sc.len > 0 {
						n = state_X_DATA
						count = 0
						sc.data = make([]byte, sc.len)
					} else {
						n = state_X_CHECKSUM
					}
				case state_X_DATA:
					crc = crc8_dvb_s2(crc, inp[i])
					sc.data[count] = inp[i]
					count++
					if count == sc.len {
						n = state_X_CHECKSUM
					}

				case state_X_CHECKSUM:
					ccrc := inp[i]
					if crc != ccrc {
						fmt.Fprintf(os.Stderr, "CRC error on %d\n", sc.cmd)
					} else {
						c0 <- sc
					}
					n = state_INIT
				}
			}
		} else {
			p.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
				sc.len = 0
				sc.ok = false
				sc.cmd = 0
				c0 <- sc
				return
			}
		}
	}
}

func encode_msp2(cmd uint16, payload []byte) []byte {
	paylen := uint16(0)
	if len(payload) > 0 {
		paylen = uint16(len(payload))
	}
	buf := make([]byte, 9+paylen)
	buf[0] = '$'
	buf[1] = 'X'
	buf[2] = '>'
	buf[3] = 0 // flags
	binary.LittleEndian.PutUint16(buf[4:6], uint16(cmd))
	binary.LittleEndian.PutUint16(buf[6:8], uint16(paylen))
	if paylen > 0 {
		copy(buf[8:], payload)
	}
	crc := byte(0)
	for _, b := range buf[3 : paylen+8] {
		crc = crc8_dvb_s2(crc, b)
	}
	buf[8+paylen] = crc
	return buf
}

func NewMSPSerial(name string) *MSPSerial {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	if name[2] == ':' && len(name) == 17 {
		bt := NewBT(name)
		return &MSPSerial{bt}
	} else {
		p, err := serial.Open(name, mode)
		if err != nil {
			log.Fatal(err)
		}
		return &MSPSerial{p}
	}
}

func (m *MSPSerial) MSPClose() {
	m.Close()
}

func Enumerate_ports() string {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	for _, port := range ports {
		if port.Name != "" {
			if port.IsUSB {
				return port.Name
			}
		}
	}
	return ""
}

func (m *MSPSerial) SendVariant() {
	buf := []byte("INAV")
	m.MSPCommand(MSP_FC_VARIANT, buf)
}

func (m *MSPSerial) SendVersion() {
	buf := make([]byte, 3)
	buf[0] = 6
	buf[1] = 6
	buf[2] = 6
	m.MSPCommand(MSP_FC_VERSION, buf)
}

func (m *MSPSerial) SendMixer() {
	buf := make([]byte, 5)
	buf[3] = 3
	m.MSPCommand(MSP2_INAV_MIXER, buf)
}

func (m *MSPSerial) SendName() {
	buf := []byte("Follower")
	m.MSPCommand(MSP_NAME, buf)
}

func (m *MSPSerial) SendGPS(armed bool) {
	buf := make([]byte, 18)

	hdop := uint16(999)
	if armed {
		buf[0] = 2
		buf[1] = byte(rand.Intn(20)) + 6
		hdop = uint16(496 - uint16(buf[1])*16)
	} else {
		buf[0] = 0
		buf[1] = byte(rand.Intn(5))
	}

	rnd := int32(rand.Intn(2000) - 1000)
	msplat := int32(BaseLat*1e7) + rnd
	rnd = int32(rand.Intn(2000) - 1000)
	msplon := int32(BaseLon*1e7) + rnd

	binary.LittleEndian.PutUint32(buf[2:6], uint32(msplat))
	binary.LittleEndian.PutUint32(buf[6:10], uint32(msplon))

	alt := uint16(39 + rand.Intn(6))

	binary.LittleEndian.PutUint16(buf[10:12], alt)  // alt
	binary.LittleEndian.PutUint16(buf[12:14], 0)    // cog
	binary.LittleEndian.PutUint16(buf[14:16], 0)    // spd
	binary.LittleEndian.PutUint16(buf[16:18], hdop) // hdop
	m.MSPCommand(MSP_RAW_GPS, buf)
}

func (m *MSPSerial) SendStatus(armed, poshold bool) {
	buf := make([]byte, 2)
	if armed && poshold {
		buf[0] = 1
	}
	m.MSPCommand(MSP_NAV_STATUS, buf)
}

func (m *MSPSerial) SendAckNak(cmd uint16, ack bool) {
	rb := encode_msp2(cmd, nil)
	if ack == false {
		rb[2] = '!'
	}
	m.Write(rb)
}

func (m *MSPSerial) deserialise_wp(b []byte) {
	var lat, lon float64
	var p1 int16
	var v, alt int32

	v = int32(binary.LittleEndian.Uint32(b[2:6]))
	lat = float64(v) / 1e7
	v = int32(binary.LittleEndian.Uint32(b[6:10]))
	lon = float64(v) / 1e7
	alt = int32(binary.LittleEndian.Uint32(b[10:14])) / 100
	p1 = int16(binary.LittleEndian.Uint16(b[14:16]))
	fmt.Printf("WP%d: %d %.6f %.6f %d %d %d\n", b[0], b[1], lat, lon, alt, p1, b[20])
	m.SendAckNak(MSP_SET_WP, true)
}
