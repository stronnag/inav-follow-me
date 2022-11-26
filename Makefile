TARGET=pico
APP=inav-follow
SRC = main.go gpsreader.go msp.go geocalc.go prefs.go oled-ssd1306.go vbat.go

all : $(APP).elf

$(APP).elf: $(SRC) go.sum
	tinygo build -target $(TARGET) -size short -o $(APP).elf

go.sum: go.mod $(wildcard *.go)
	go mod tidy

flash: $(APP).elf
	tinygo flash -target $(TARGET)

uf2: $(APP).elf
	elf2uf2-rs $(APP).elf $(APP).uf2

clean:
	rm -f $(APP).elf $(APP).uf2 go.sum
