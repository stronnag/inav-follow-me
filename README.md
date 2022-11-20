# inav-follow

## Overview

Simple 'follow me' application for INAV. The application runs on a RaspberryPi Pico (rp2040) and requires a NMEA GPS connected to the Pico.

The vehicle requires INAV firmware supporting MSPv2. This is an artificial requirement to simplify the code; the underlying INAV "follow me" functionality e.g. (`GCS_NAV`) has existed since 2016.

A bi-directional MSP capable transparent serial data link is required between the ground control station (GCS) and the vehicle. Examples of suitable data links include 3DR, HC-12 and LoRA based radio systems.

The RP Pico requires a GPS. An old NMEA capable Neo6M is more than adequate. It must provide `$GPGGA` and optionally `$GPRMC`.

In theory, "follow me" is available for all  types of INAV vehicle (`platform type`) that supports stationary `POSHOLD`, e.g. MultiRotor and possibly Rover and Boat. By default, fixed wing is excluded, but this can be changed by configuration.

Optionally, a SSD1306 OLED will be supported (once the author's device is delivered).

## Pico Firmware Configuration

The configurable items are built into the application; it is necessary to rebuild the application to change them. Unless you have some (small) development skills, these are the only items you should change. See the source file `prefs.go`:

``` go
/* user preferences */
const (
	// Baud rate for MSP
	MSPBAUD = 115200
	// Baud rate for GPS
	GPSBAUD = 9600
	// Minimum user sats for follow me
	GPSMINSAT = 6
	// Craft type for no follow (1 = FW); 255 allows anything
	DONT_FOLLOW_TYPE = 1
	// Don't follow if closer than this distance (m), 0 disables this check
	MIN_FOLLOW_DIST float32 = 2.0
)
/* End of user preferences */
```
If the configuration is changed, it is necessary to rebuild the firmware.

## Pico Hardware Connections

* The GPS is connected to UART0 (pins 1 & 2)
* The MSP serial link is connected to UART1 (pins 11 & 12)
* Optionally, the OLED is connected to I2C0 (SDA pin 6, SCL pin 7)

These may be changed by updating the peripheral device configurations in `main.go`.

## Usage

* Power up the Pico.
* If the Pico is powered / connected via USB, then status information is provided over USB and may be viewed in any serial terminal.
* [Future] Status data will be displayed on the OLED.
  * When no valid data is available : "Waiting"
  * Once GPS time is available "HH:MM:SS"
	* GPS Quality (0/1/2), no fix, GPS fix, DGPS fix.
	* Number of satellites
* Once the required number of satellites is reached (`GPSMINSAT` above), then the vehicle is interrogated.
  * If the vehicle is of type `DONT_FOLLOW_TYPE` (typically FW), then follow me is not available.
  * Otherwise, navigation interrogation is started. If navigation mode `HOLD` is reported, and the distance between the vehicle and GCS is greater than `MIN_FOLLOW_DIST`, then follow me data (the required observer / GCS location) is sent to the vehicle.
  * [Future] The "follow me" status will be displayed on the OLED.
  * The vehicle will only react to this data if the user has also set `GCS NAV` mode. The user may switch between normal `POSHOLD` and "Follow me" by toggling a `GCS NAV` switch on the transmitter.

**Note** that as the vehicle has to be in `POSHOLD` for `GCS NAV` to work, if you experience any issues, disengaging the `GCS NAV` switch will revert to standard `POSHOLD`.


## Installation and Building

A `fl2` file may be provided (in the Release folder) with the default settings shown above. This may be dropped onto the Pico's boot loader mode pseudo-filesystem.

### Build requirements

* `tinygo` compiler (most Linux distros / FreeBSD provide packages or [Github Project releases](https://github.com/tinygo-org/tinygo/releases)).
* Optionally, `make` to automate
* Internet access for some external packages for first build.

### Make targets

* `make` : (default). Builds `.elf` file (`tinygo build -target pico -size short -o inav-follow.elf`)
* `make flash` : Builds `.elf`, flashes `.uf2` image to device. (`tinygo flash -target pico`)
* `make uf2` : Builds `.elf`, generates `.uf2` using `elf2uf2-rs`. `elf2uf2-rs inav-follow.elf inav-follow.uf2`
* `make clean` : Removes and `.elf` and `.uf2` files.

### Monitor over USB

`tinygo monitor [-port DEVICE_NODE]`

## Caveat

This application has been bench tested; it has not been flown.

Running against a GPS replay and trival MSP simulator, it appears to do the right thing.

Note that at the moment, there is no OLED support and copious debug output is written to any connected USB (USB serial console).

## Licence

(c) Jonathan Hudson 2022. 0-BSD.
