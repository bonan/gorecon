package gorecon

/**
 * This file contains the command bytes used to request info
 * and control the fan controller
 */

const (
	// Requests to get data from controller
	reqDisplayChannel = 0x10 /** Request current selected channel */
	reqTempAndSpeed   = 0x30 /** Request temperature and speed (0x30-34) */
	reqDeviceSettings = 0x50 /** Request Device settings. */
	reqAlarmAndSpeed  = 0x70 /** Request alarm temp and manual RPM (0x70-74) */
	reqDeviceStatus   = 0x90 /** Request device status */

	// Requests to set values
	setDisplayChannel = 0x20 /** Set display channel (0x20-24 depending on channel) */
	setDeviceSettings = 0x60 /** Set Device settings. */
	setAlarmAndSpeed  = 0x80 /** Set alarm temp and manual RPM (0x80-84) */

	// Responses from controller
	recvDisplayChannel = 0x20 /** Current display channel (0x20-24) */
	recvTempAndSpeed   = 0x40 /** Temp in F and speed (RPM) (0x40-44) */
	recvDeviceSettings = 0x60 /** Device settings. */
	recvAlarmAndSpeed  = 0x80 /** Alarm temp and manual RPM (0x80-84) */
	recvDeviceStatus   = 0xA0 /** Device status */
	recvAck            = 0xF0
	recvNak            = 0xFA

	// Bitwise settings map
	settingManual    = 0x01
	settingFarenheit = 0x02
	settingAudio     = 0x04
)

func byte2rpm(b, c byte) uint16 {
	return (uint16(c) << 8) + uint16(b)
}

func rpm2byte(b uint16) (byte, byte) {
	return byte(b % 256), byte(b / 256)
}

func f2c(b int) int {
	return int((float64(b) - 32) * 5 / 9)
}

func c2f(b int) int {
	return int(float64(b)*9/5 + 32)
}
