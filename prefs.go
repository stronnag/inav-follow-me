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
)

/* End of user preferences */
