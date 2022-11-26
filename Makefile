TARGET=pico
APP=inav-follow
SRC = main.go prefs.go
PKGS = pkg/gps/gpsreader.go pkg/msp/msp.go pkg/geo/geocalc.go pkg/oled/oled-ssd1306.go pkg/vbat/vbat.go

all : $(APP).elf

$(APP).elf: $(SRC) $(PKGS) go.sum
	tinygo build -target $(TARGET) -size short -o $(APP).elf

go.sum: go.mod $(wildcard *.go)
	go mod tidy

flash: $(APP).elf
	tinygo flash -target $(TARGET)

uf2: $(APP).elf
	elf2uf2-rs $(APP).elf $(APP).uf2

clean:
	go clean
	rm -f $(APP).elf $(APP).uf2 go.sum
