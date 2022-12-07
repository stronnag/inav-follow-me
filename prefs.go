package main

/* user preferences */
const (
	// Baud rate for MSP
	MSPBAUD = 115200
	// Baud rate for GPS
	GPSBAUD = 9600
	// Minimum user sats for follow me
	GPSMINSAT = 6
	// Craft type for no follow (1 = FW)
	DONT_FOLLOW_TYPE = 1
	// Don't follow closer than this distance (m)
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
