# followsim

Trivial MSP simulator for INAV follow me. The application simulates an INAV FC's MSP responses for the MSP required by the INAV follow me tool.

Note that the generated MSP is *just* that required by `inav-follow-me` and in particular, only the data attributes required by `inav-follow-me` are provided (which may also mean that the length value is not according to specification).

## Usage

```
Usage of followsim [options] device
  -lat float
    	Base latitude
  -lon float
    	Base longitude
```

If the latitude and longitude values are not provided, random values are used.

## Message catalogue

The following MSP messages are processed for input:

* `MSP_FC_VARIANT`
* `MSP_FC_VERSION`
* `MSP2_INAV_MIXER`
* `MSP_NAME`
* `MSP_RAW_GPS`
* `MSP_NAV_STATUS`

The following MSP messages are processed for output:

* `MSP_SET_WP` (meets MSP specification)

## Installation

Use the provided `make` file

```
make
make install # -> ~/.local/bin/
# or
sudo make install prefix=/usr/local  # -> /usr/local/bin/
```

or copy the `followsim` binary to somewhere convenient

## Example

``` sh
followsim -lat 35.761000 -lon 140.378945 /dev/ttyUSB0
```

Note that the vehicle is initially unarmed and not in `POSHOLD`. The arm and hold states may be toggled by pressing the following keys:

* `a`, `A` : Toggles arming (also sets `POSHOLD` off)
* `p`, `P` : Toggles `POSHOLD`
