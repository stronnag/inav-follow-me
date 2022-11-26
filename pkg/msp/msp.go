package msp

import (
	"encoding/binary"
	"machine"
	"time"
)

type MSPMsg struct {
	Len  uint16
	Cmd  uint16
	Ok   bool
	Data []byte
}

type MSPReader struct {
	mchan chan MSPMsg
	uart  machine.UART
}

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
	state_MX
	state_HEADER2
	state_FLAGS
	state_ID1
	state_ID2
	state_LEN1
	state_LEN2
	state_DATA
	state_CHECKSUM
)

const (
	wp_WAYPOINT = 1
)

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

var (
	mspdelay time.Duration
)

func NewMSPUartReader(uart machine.UART, mchan chan MSPMsg, baud int) *MSPReader {
	mspdelay = time.Duration((10 * 1000000 / (2 * baud))) * time.Microsecond
	return &MSPReader{uart: uart, mchan: mchan}
}

func (m *MSPReader) UartReader() {
	count := uint16(0)
	crc := byte(0)
	msg := MSPMsg{}

	mstate := state_INIT
	for {
		if m.uart.Buffered() > 0 {
			c, err := m.uart.ReadByte()
			if err == nil {
				switch mstate {
				case state_INIT:
					if c == '$' {
						mstate = state_MX
						msg.Ok = false
						msg.Len = 0
						msg.Cmd = 0
					}

				case state_MX:
					if c == 'X' {
						mstate = state_HEADER2
					} else {
						mstate = state_INIT
					}

				case state_HEADER2:
					if c == '!' {
						mstate = state_FLAGS
					} else if c == '>' {
						mstate = state_FLAGS
						msg.Ok = true
					} else {
						mstate = state_INIT
					}

				case state_FLAGS:
					crc = crc8_dvb_s2(0, c)
					mstate = state_ID1

				case state_ID1:
					crc = crc8_dvb_s2(crc, c)
					msg.Cmd = uint16(c)
					mstate = state_ID2

				case state_ID2:
					crc = crc8_dvb_s2(crc, c)
					msg.Cmd |= uint16(uint16(c) << 8)
					mstate = state_LEN1

				case state_LEN1:
					crc = crc8_dvb_s2(crc, c)
					msg.Len = uint16(c)
					mstate = state_LEN2

				case state_LEN2:
					count = 0
					crc = crc8_dvb_s2(crc, c)
					msg.Len |= uint16(uint16(c) << 8)
					if msg.Len > 0 {
						mstate = state_DATA
						msg.Data = make([]byte, msg.Len)
					} else {
						mstate = state_CHECKSUM
					}

				case state_DATA:
					crc = crc8_dvb_s2(crc, c)
					msg.Data[count] = c
					count++
					if count == msg.Len {
						mstate = state_CHECKSUM
					}

				case state_CHECKSUM:
					ccrc := c
					if crc != ccrc {
						msg.Ok = false
					}
					m.mchan <- msg
					mstate = state_INIT
					msg = MSPMsg{}
				}
			}
		} else {
			time.Sleep(mspdelay)
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
	buf[2] = '<'
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

func (m *MSPReader) MSPCommand(cmd uint16, payload []byte) {
	rb := encode_msp2(cmd, payload)
	m.uart.Write(rb)
}

func (m *MSPReader) Update_WP255(lat, lon float32, brg uint16) {
	buf := make([]byte, 21)
	buf[0] = byte(255)
	buf[1] = wp_WAYPOINT
	v := int32(lat * 1e7)
	binary.LittleEndian.PutUint32(buf[2:6], uint32(v))
	v = int32(lon * 1e7)
	binary.LittleEndian.PutUint32(buf[6:10], uint32(v))
	binary.LittleEndian.PutUint32(buf[10:14], uint32(0)) // Alt (keep vehicle alt)
	binary.LittleEndian.PutUint16(buf[14:16], brg)
	binary.LittleEndian.PutUint16(buf[16:18], uint16(0))
	binary.LittleEndian.PutUint16(buf[18:20], uint16(0))
	buf[20] = byte(0xa5) // not checked, so 0 would do
	m.MSPCommand(MSP_SET_WP, buf)
}
