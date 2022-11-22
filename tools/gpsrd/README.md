# Simple GPS Reader / Replayer

`gpsrd` reads a file of NMEA GPS sentences and replays them at recorded speed over a serial interface with designated baud rate.

## Usage

```
$ gpsrd --help
Usage of gpsrd [options] file
 where "file" is a file containing NMEA sentences
  -baud int
    	Baud rate (default 9600)
  -device string
    	Serial device
```

## Installation

```
make
make install # -> ~/.local/bin/
# or
sudo make install prefix=/usr/local  # -> /usr/local/bin/
```

or copy the `gpsrd` binary to somewhere convenient
