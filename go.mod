module gpstest
go 1.19

require (
	github.com/Nondzu/ssd1306_font v1.0.1
	tinygo.org/x/drivers v0.23.0
	geo v1.0.0
	gps v1.0.0
	msp v1.0.0
	oled v1.0.0
	vbat v1.0.0
)

replace geo v1.0.0  => ./pkg/geo
replace gps v1.0.0 => ./pkg/gps
replace msp v1.0.0 => ./pkg/msp
replace oled v1.0.0 => ./pkg/oled
replace vbat v1.0.0 => ./pkg/vbat
