# inav-follow

## Overview

Simple 'follow me' application for INAV. The application runs on a RaspberryPi Pico (rp2040) and requires a NMEA GPS connected to the Pico.

The vehicle requires INAV firmware supporting MSPv2. This is an artificial requirement to simplify the code; the underlying INAV "follow me" functionality e.g. (`GCS_NAV`) has existed since 2016.

A bi-directional MSP capable transparent serial data link is required between the ground control station (GCS) and the vehicle. Examples of suitable data links include 3DR, HC-12 and LoRA based radio systems.

The RP Pico requires a GPS. An old NMEA capable Neo6M is more than adequate. It must provide `$GPGGA` and optionally `$GPRMC`.

In theory, "follow me" is available for all  types of INAV vehicle (`platform type`) that supports stationary `POSHOLD`, e.g. MultiRotor and possibly Rover and Boat. By default, fixed wing is excluded, but this can be changed by configuration.

Optionally, a SSD1306 OLED is supported.

## Pico Firmware Configuration

The configurable items are built into the application; it is necessary to rebuild the application to change them. Unless you have some (small) development skills and which to change other things, these are the only items you should change. See the source file `prefs.go`:

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
    	// GPS Time format, either integer seconds or 1 decimal
	GPS_TIME_FORMAT = "15:04:05"
	//GPS_TIME_FORMAT = "15:04:05.0"

	//  USE_VBAT boolean
	USE_VBAT = true
	// For Pico-W you need this; ignored for standard Pico
	VBAT_OFFSET = 0.8

	// if true, the HOME location will also be set to the follow me location
	RESET_HOME = false
)
/* End of user preferences */
```
If the configuration is changed, it is necessary to rebuild / reflash the firmware.

### Voltage Reporting

The method for reporting voltage differs between the Pico and Pico-W. This is now auto-detected if `USE_VBAT = true` . At least on the developer's Pico-W, an offset (`0.8`V) is also required to display the external (`VSYS` voltage). For standard Pico, an offset of `0.0` is applied.

In order to have voltage displayed, it is necessary to:

* Set `USE_VBAT` to `true` (default)
* Consider setting `VBAT_OFFSET` (even to `0.0`)
* Rebuild / reflash the firmware

## CLI

A number of preferences may be changed at runtime using a CLI. When a serial terminal program (`cu`, `minicom`, `picocom`, `tinygo monitor`, `cliterm -n`, `putty` etc.) is connected to the Pico device node (typically `/dev/ttyACM0`), informational data is displayed. This may be paused by pressing the hash key (`#`); a banner `INAV-followme! CLI` and prompt ` #` is then displayed and the user can issue commands. The commands, current value and ranges are shown by the `help` and `list` commands (which do the same thing).

```
$ cliterm -n
2022-12-07T09:09:51+0000 Registered serial device: /dev/ttyACM0 [2e8a:000a], Vendor: Raspberry_Pi, Model: Pico, Serial: (null), Driver: cdc_acm
open /dev/ttyACM0
09:09:55 [1:0] Qual:  0  sats:  0  lat:  0.000000  lon:  0.000000
09:09:56 [1:0] Qual:  0  sats:  0  lat:  0.000000  lon:  0.000000
09:09:57 [1:0] Qual:  0  sats:  0  lat:  0.000000  lon:  0.000000
09:09:58 [1:0] Qual:  0  sats:  0  lat:  0.000000  lon:  0.000000
09:09:59 [1:0] Qual:  0  sats:  0  lat:  0.000000  lon:  0.000000

INAV-followme! CLI

# list
gps_baud = 9600 [1200 - 115200]
msp_baud = 115200 [1200 - 115200]
vbat_offset = 0.8 [0.0 - 1.8]
reset_home = false [0/false - 1/true]
minsats = 6 [3 - 99]
help
list
#
09:10:05 [1:0] Qual:  0  sats:  0  lat:  0.000000  lon:  0.000000
```

Values are set as `key = value`, for example:

``` shell
reset_home = true
```

### CLI variables

| Key name | Usage |
| -------- | ----- |
| `gps_baud` | GPS baud rate, validated (1) |
| `msp_baud` | MSP baud rate, validated (1) |
| `vbat_offset` | VBAT voltage offset in the range 0.0 - 1.8V |
| `reset_home` | Defines whether a RESET HOME (WP#0) update is performed in addition to follow me (WP#255) (2) |
| `minsats` | The minimum satellite count for follow me / reset home to be asserted |

Note 1: Valid baud rates are 1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200.

Note 2: If true, `MSP_SET_WP` for WP#0 is only asserted when the vehicle is in POSHOLD (INAV does not require this, `GCS NAV` is sufficient).

### Control keys

* `#` : Opens CLI
* `Esc` : Escape key, closes CLI, informational message flow resumes.

### Caveat

Due to limitations of the `Tinygo` SDK, it is not possible to save values set via the CLI.

## Pico Hardware Connections

* The GPS is connected to UART0 (pins 1 & 2)
* The MSP serial link is connected to UART1 (pins 11 & 12)
* Optionally, the OLED is connected to I2C1 (SDA pin 31, SCL pin 32). These pins are used for ergonomics such that all the external connector are not on one side of the device.

These may be changed by updating the peripheral device configurations in `main.go`.

## Usage

* Power up the Pico.
* If the Pico is powered / connected via USB, then status information is provided over USB and may be viewed in any serial terminal.
* Status data will be displayed on the OLED.
  * When no valid data is available : "Initialised"
  * Once GPS time is available "HH:MM:SS"
	* GPS Quality (0/1/2), no fix, GPS fix, DGPS fix.
	* Number of satellites
* Once the required number of satellites is reached (`GPSMINSAT` above), then the vehicle is interrogated.
  * If the vehicle is of type `DONT_FOLLOW_TYPE` (typically FW), then follow me is not available.
  * Otherwise, navigation interrogation is started. If navigation mode `HOLD` is reported, and the distance between the vehicle and GCS is greater than `MIN_FOLLOW_DIST`, then follow me data (the required observer / GCS location) is sent to the vehicle.
  * The "follow me" status will be displayed on the OLED.
  * The vehicle will only react to this data if the user has also asserts `GCS NAV` mode. The user may switch between normal `POSHOLD` and "Follow me" by toggling a `GCS NAV` switch on the transmitter.

**Note** that as the vehicle has to be in `POSHOLD` for `GCS NAV` to work, if you experience any issues, disengaging the `GCS NAV` switch will revert to standard `POSHOLD`.

## Installation and Building

A `fl2` file may be provided (in the Release folder) with the default settings shown above. This may be dropped onto the Pico's boot loader mode pseudo-filesystem.

### Build requirements

* `tinygo` compiler (most Linux distros / FreeBSD provide packages or [Github Project releases](https://github.com/tinygo-org/tinygo/releases)) for others.
* Optionally, `make` to automate
* Internet access for required external packages (for first build).

### Make targets

* `make` : (default). Builds `.elf` file (`tinygo build -target pico -size short -o inav-follow.elf`)
* `make flash` : Builds `.elf`, flashes `.uf2` image to device. (`tinygo flash -target pico`)
* `make uf2` : Builds `.elf`, generates `.uf2` using `elf2uf2-rs`. `elf2uf2-rs inav-follow.elf inav-follow.uf2`
* `make clean` : Removes and `.elf` and `.uf2` files.

### Monitor over USB

`tinygo monitor [-port DEVICE_NODE]`

### OLED

It is possible to use a SSD1306 OLED to provide a clue as to what is happening.

The fields are as follows:

![Example](assets/oled.png)

* The 1st line shows the attached GPS time
* The 2nd line (**GPS**) shows the local GPS Status (satellites and fix type)
* The 3rd line (**Mode**) shows the INAV connection status
* When connected to INAV, the 4th line (**INAV**) shows the INAV Firmware version and and navigation mode.
* The 5th line (**VSat**) shows the vehicle's (INAV) satellite count and HDOP.
* The 6th Line (**VPos**) shows the distance and bearing from the vehicle to the user.

#### Status

* `Starting` : Application is starting
* `Initialised` : Application ready for GPS input and MSP connection
* `Connecting` : Connecting to FC / MSP (sufficient local satellites / fix)
* `Connected` : Connected to the FC
* `Failed` : FC did not return required information (in particular `FC_VARIANT` == `INAV` or excluded by `DONT_FOLLOW_TYPE`).

#### Navigation Modes

* `Idle` : Not in a navigation mode
* `PH` : Position Hold, application sends WP#255 location which will result in 'follow me' if the pilot also asserts `GCS NAV` mode
* `RTH` : Return to home
* `WP` : Waypoint mission

![IRL](assets/oled-fix.png)

Note: The image is from an earlier build with some UI elements rearranged.

## Caveat

This application has been bench tested; it has not tested in flight (by the author).

Running against a GPS replay and trivial MSP simulator, it appears to do the right thing.

Note that at the moment, copious debug output is written to any connected USB (USB serial console).

## Simulation Tools

A GPS replayer (`gpsrd`) and a MSP simulator (`followsim`, sufficient for this application only) may be found in the `tools` directory. These require a native `Go` compiler.

## Additional Infomation

Please see the wiki, in particular [pinout diagram and high level design](https://github.com/stronnag/inav-follow-me/wiki/Pinout-and-Design-reference) reference.

## Licence

(c) Jonathan Hudson 2022. 0-BSD.
