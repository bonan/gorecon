package gorecon

import "errors"

// Channel has methods to read and set current speed and temperatures
type Channel struct {
	curSpeed    uint16    // Current speed
	maxSpeed    uint16    // Maximum speed
	manualSpeed uint16    // Manual speed as reported
	temp        uint8     // Current temperature in F
	curAlarm    uint8     // Current alarm temperature in F
	targetSpeed uint16    // Our speed setting
	targetAlarm uint8     // Our alarm temp setting (in F)
	dirty       bool      // Whether something has changed since last check
	init        byte      // Bitmap for channel initialization status
	change      chan bool // Channel for change updates
}

// NewChannel returns the a new channel with defaults set
func NewChannel(changeChan chan bool) *Channel {
	return &Channel{
		curSpeed:    0,
		maxSpeed:    0,
		manualSpeed: 0,
		temp:        0,
		curAlarm:    0,
		targetSpeed: 65535,
		targetAlarm: 255,
		dirty:       false,
		init:        0x00,
		change:      changeChan,
	}
}

// SetSpeed changes the target speed of the fan controller (RPM)
func (c *Channel) SetSpeed(speed int) error {
	if speed > int(c.maxSpeed) || speed >= 65535 {
		return errors.New("Speed exceeds maximum speed")
	}
	if speed < 0 {
		return errors.New("Speed cannot be negative")
	}
	c.targetSpeed = uint16(speed)
	c.dirty = true
	c.change <- true
	return nil
}

// SetAlarmTempC sets the alarm termperature in Celsius
func (c *Channel) SetAlarmTempC(temp int) error {
	return c.SetAlarmTempF(c2f(temp))
}

// SetAlarmTempF sets the alarm temperature in Farenheit
func (c *Channel) SetAlarmTempF(temp int) error {
	if temp < 0 || temp >= 255 {
		return errors.New("Temperature out of range")
	}
	c.targetAlarm = uint8(temp)
	c.dirty = true
	c.change <- true
	return nil
}

// Speed returns the current speed (RPM)
func (c *Channel) Speed() int {
	return int(c.curSpeed)
}

// MaxSpeed returns the maximum speed that the fan supports (RPM)
func (c *Channel) MaxSpeed() int {
	return int(c.maxSpeed)
}

// ManualSpeed returns the value of the controllers target speed (RPM)
func (c *Channel) ManualSpeed() int {
	return int(c.manualSpeed)
}

// TempC returns current temperature in Celsius
func (c *Channel) TempC() int {
	return f2c(int(c.temp))
}

// TempF returns current temperature in Farenheit
func (c *Channel) TempF() int {
	return int(c.temp)
}

// AlarmTempC returns the current alarm temperature in Celsius
func (c *Channel) AlarmTempC() int {
	return f2c(int(c.curAlarm))
}

// AlarmTempF returns the current alarm termperature in Farenheit
func (c *Channel) AlarmTempF() int {
	return int(c.curAlarm)
}

// IsInitialized checks if we've received initial data from fan controller
//    for this channel
func (c *Channel) IsInitialized() bool {
	return c.init == 0x03
}

func (c *Channel) report(data []byte) {
	if len(data) == 5 {
		if c.init&0x01 == 0 {
			c.init = c.init | 0x01
		}
		c.temp = uint8(data[0])
		c.curSpeed = byte2rpm(data[1], data[2])
		c.maxSpeed = byte2rpm(data[3], data[4])
	}
}

func (c *Channel) reportAlarm(data []byte) {
	if len(data) == 3 {
		if c.init&0x02 == 0 {
			c.init = c.init | 0x02
		}
		c.curAlarm = uint8(data[0])
		c.manualSpeed = byte2rpm(data[1], data[2])
		if c.targetSpeed == 65535 {
			c.targetSpeed = c.manualSpeed
		}
		if c.targetAlarm == 255 {
			c.targetAlarm = c.curAlarm
		}
	}
}
