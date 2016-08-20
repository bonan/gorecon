package gorecon

/**
 * This file contains the command bytes used to request info
 * and control the fan controller
 */

const (
	getCurrentChannel         = 0x10 /**< Get Current selected channel */
	readCurrentChannel        = 0x10 /**< Get Current selected channel */
	setDisplayChannelCh0      = 0x20 /**< OUT Set the Recon's display to channel */
	setDisplayChannelCh1      = 0x21 /**< OUT Set the Recon's display to channel */
	setDisplayChannelCh2      = 0x22 /**< OUT Set the Recon's display to channel */
	setDisplayChannelCh3      = 0x23 /**< OUT Set the Recon's display to channel */
	setDisplayChannelCh4      = 0x24 /**< OUT Set the Recon's display to channel */
	getTempAndSpeedChannel0   = 0x30 /**< OUT  Temp in F and current speed (RPM) */
	getTempAndSpeedChannel1   = 0x31 /**< OUT   in F and current speed (RPM) */
	getTempAndSpeedChannel2   = 0x32 /**< OUT   in F and current speed (RPM) */
	getTempAndSpeedChannel3   = 0x33 /**< OUT   in F and current speed (RPM) */
	getTempAndSpeedChannel4   = 0x34 /**< OUT   in F and current speed (RPM) */
	readTempAndSpeedChannel0  = 0x40 /**< IN  Temp in F and speed (RPM)*/
	readTempAndSpeedChannel1  = 0x41 /**< IN  Temp in F and speed (RPM)*/
	readTempAndSpeedChannel2  = 0x42 /**< IN  Temp in F and speed (RPM)*/
	readTempAndSpeedChannel3  = 0x43 /**< IN  Temp in F and speed (RPM)*/
	readTempAndSpeedChannel4  = 0x44 /**< IN  Temp in F and speed (RPM)*/
	getDeviceSettings         = 0x50 /**< OUT Device settings. */
	setDeviceSettings         = 0x60 /**< OUT Set Device settings. */
	readDeviceSettings        = 0x60 /**< IN  Device settings. */
	getAlarmAndSpeedChannel0  = 0x70 /**< OUT Alarm Temperature and Max RPM*/
	getAlarmAndSpeedChannel1  = 0x71 /**< OUT Alarm Temperature and Max RPM*/
	getAlarmAndSpeedChannel2  = 0x72 /**< OUT Alarm Temperature and Max RPM*/
	getAlarmAndSpeedChannel3  = 0x73 /**< OUT Alarm Temperature and Max RPM*/
	getAlarmAndSpeedChannel4  = 0x74 /**< OUT Alarm Temperature and Max RPM*/
	setAlarmAndSpeedChannel0  = 0x80 /**< OUT Alarm Temperature and Max RPM*/
	setAlarmAndSpeedChannel1  = 0x81 /**< OUT Alarm Temperature and Max RPM*/
	setAlarmAndSpeedChannel2  = 0x82 /**< OUT Alarm Temperature and Max RPM*/
	setAlarmAndSpeedChannel3  = 0x83 /**< OUT Alarm Temperature and Max RPM*/
	setAlarmAndSpeedChannel4  = 0x84 /**< OUT Alarm Temperature and Max RPM*/
	readAlarmAndSpeedChannel0 = 0x80 /**< IN  Alarm Temperature and Max RPM*/
	readAlarmAndSpeedChannel1 = 0x81 /**< IN  Alarm Temperature and Max RPM*/
	readAlarmAndSpeedChannel2 = 0x82 /**< IN  Alarm Temperature and Max RPM*/
	readAlarmAndSpeedChannel3 = 0x83 /**< IN  Alarm Temperature and Max RPM*/
	readAlarmAndSpeedChannel4 = 0x84 /**< IN  Alarm Temperature and Max RPM*/
	getDeviceStatus           = 0x90 /**< OUT Request Status: Ready/Not Ready */
	readDeviceStatus          = 0xA0 /**< IN  Status: Ready/Not Ready */
	readAck                   = 0xF0
	readNak                   = 0xFA
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
