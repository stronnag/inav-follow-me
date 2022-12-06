package main

import (
	"errors"
	"machine"
	"strconv"
	"strings"
	"time"
)

const (
	I_GPSBAUD = iota
	I_MSPBAUD
	I_VOFFSET
	I_RESETHOME
	I_NSATS
	I_HELP
	I_NONE
)

type cmdfunc func(arg string) (int32, error)

type EditMsg struct {
	Id    byte
	Value int32
}

type CLIMsg struct {
	Id    byte
	Name  string
	cfunc cmdfunc
	vmin  string
	vmax  string
}

var Climsgs = []CLIMsg{
	{I_GPSBAUD, "gps_baud", cmdfunc(vbaud), "1200", "115200"},
	{I_MSPBAUD, "msp_baud", cmdfunc(vbaud), "1200", "115200"},
	{I_VOFFSET, "vbat_offset", cmdfunc(voffset), "0.0", "1.8"},
	{I_RESETHOME, "reset_home", cmdfunc(vbool), "0/false", "1/true"},
	{I_NSATS, "minsats", cmdfunc(vsats), "3", "99"},
	{I_HELP, "help", nil, "", ""},
	{I_HELP, "list", nil, "", ""},
}

func vbaud(s string) (int32, error) {
	bauds := [8]int32{1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200}
	iv, err := parseInt(s)
	if err == nil {
		for _, b := range bauds {
			if b == iv {
				return iv, nil
			}
		}
		iv = 0
		err = errors.New("Invalid baud rate")
	}
	return iv, err
}

func vsats(s string) (int32, error) {
	iv, err := parseInt(s)
	if err == nil {
		if iv < 3 || iv > 99 {
			return 0, errors.New("Invalid minsats")
		}
	}
	return iv, err
}

func voffset(s string) (int32, error) {
	iv, err := parseScaledFloat(s)
	if err == nil {
		if iv < 0 || iv > 1800 {
			return iv, errors.New("Invalid offset [0 - 1.8]")
		}
	}
	return iv, err
}

func vbool(s string) (int32, error) {
	return parseBool(s), nil
}

const consoleBufLen = 80

var (
	input   []byte
	console = machine.Serial
	cli     bool
)

func matchCLI(str string) byte {
	for _, v := range Climsgs {
		if v.Name == str {
			return v.Id
		}
	}
	return I_NONE
}

func parseInt(str string) (int32, error) {
	v, err := strconv.ParseInt(str, 10, 32)
	if err == nil {
		return int32(v), nil
	} else {
		return 0, err
	}
}

func parseBool(str string) int32 {
	if str[0] == '1' || str[0] == 't' || str[0] == 'y' || str[0] == 'T' || str[0] == 'Y' {
		return 1
	} else {
		return 0
	}
}

func parseScaledFloat(str string) (int32, error) {
	v, err := strconv.ParseFloat(str, 32)
	if err == nil {
		return int32(1000 * v), nil
	} else {
		return 0, err
	}
}

func process_input(mchan chan EditMsg, s string) {
	var err error = nil
	msg := EditMsg{}
	parts := strings.Split(s, "=")
	key := strings.TrimSpace(parts[0])
	iret := matchCLI(key)
	var val string
	if len(parts) == 2 {
		val = strings.TrimSpace(parts[1])
	}
	switch iret {
	case I_NONE:
		println("Unrecognised \"", key, "\"")
		return
	case I_HELP:
		for _, cl := range Climsgs {
			print(cl.Name)
			if cl.cfunc != nil {
				print(" = ")
				switch cl.Id {
				case I_GPSBAUD:
					print(GpsBaud)
				case I_MSPBAUD:
					print(MspBaud)
				case I_NSATS:
					print(MinSat)
				case I_VOFFSET:
					print(strconv.FormatFloat(float64(VBatOffset), 'f', -1, 32))
				case I_RESETHOME:
					print(ResetHome)
				}
				print(" [")
				print(cl.vmin)
				print(" - ")
				print(cl.vmax)
				println("]")
			} else {
				println()
			}
		}
		return
	default:
		msg.Id = iret
		msg.Value, err = Climsgs[iret].cfunc(val)
		break
	}
	if err != nil {
		println("Error", key, val, err)
	} else {
		println("OK:", key, "=", val)
		mchan <- msg
	}
}

func prompt() {
	console.Write([]byte("# "))
}

func Clireader(mchan chan EditMsg) {
	cli = false
	input = make([]byte, consoleBufLen)
	i := 0
	for {
		if console.Buffered() > 0 {
			data, _ := console.ReadByte()
			if data == '#' {
				if !cli {
					cli = true
					Debug = false
					println("\r\nINAV-followme! CLI\r\n")
					prompt()
					continue
				}
			}

			if data == 27 {
				cli = false
				Debug = true
				println()
				continue
			}

			if cli {
				switch data {
				case 0x8:
					fallthrough
				case 0x7f:
					if i > 0 {
						i -= 1
						console.Write([]byte{0x8, 0x20, 0x8})
					}
					break
				case 13:
					// return key
					console.Write([]byte{13, 10})
					process_input(mchan, string(input[:i]))
					prompt()
					i = 0
					break
				default:
					if data > 31 && data < 128 {
						if i < (consoleBufLen - 1) {
							console.WriteByte(data)
							input[i] = data
							i++
						}
					}
					break
				}
			}
		} else {
			time.Sleep(time.Millisecond * 50)
		}
	}
}
