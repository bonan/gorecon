package gorecon

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"time"
)

// Device represents a Bitfinex Recon device
type Device struct {
	device         io.ReadWriter
	started        bool
	queue          chan *request
	stop           chan bool
	change         chan bool
	report         chan *Report
	init           byte
	curSetting     byte
	targetSetting  byte
	displayChannel uint8
	Channels       []*Channel
}

// DeviceExport ...
type DeviceExport struct {
	Audio          bool             `json:"audio"`
	Manual         bool             `json:"manual"`
	TempMode       string           `json:"temp_mode"`
	DisplayChannel int              `json:"display_channel"`
	Channels       []*ChannelExport `json:"channels"`
}

// Report ...
type Report struct {
}

// MarshalJSON ...
func (r *Device) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Export())
}

// Export ...
func (r *Device) Export() *DeviceExport {
	e := &DeviceExport{
		Audio:          r.IsAudioEnabled(),
		Manual:         r.IsManual(),
		TempMode:       "F",
		DisplayChannel: r.DisplayChannel(),
		Channels:       []*ChannelExport{},
	}
	if !r.IsTempF() {
		e.TempMode = "C"
	}
	for _, c := range r.Channels {
		e.Channels = append(e.Channels, c.Export())
	}
	return e
}

// NewDevice initializes a device, optional second argument is a channel for
//   reporting events.
func NewDevice(device io.ReadWriter, reportChan chan *Report) *Device {
	dev := &Device{
		device:  device,
		started: false,
		queue:   make(chan *request, 20),
		stop:    make(chan bool, 1),
		change:  make(chan bool, 1),
		report:  reportChan,
		init:    0x00,
	}
	return dev
}

// Stop will stop the polling of a device
func (r *Device) Stop() {
	r.stop <- true
}

// IsStarted returns true if device polling is running
func (r *Device) IsStarted() bool {
	return r.started
}

// Start will start polling a device
func (r *Device) Start() error {
	if r.started {
		return errors.New("Device polling is already started")
	}
	r.Channels = []*Channel{
		NewChannel(r.change),
		NewChannel(r.change),
		NewChannel(r.change),
		NewChannel(r.change),
		NewChannel(r.change),
	}
	r.started = true

	/**
	 * Launch event loop
	 */
	go r.loop()

	/**
	 * Send initial queries
	 */
	r.requestStatus()
	r.requestSpeedAndTemp()
	r.requestManualSettings()

	/**
	* Wait for response for initial queries
	 */
	<-time.After(16 * time.Second)
	if !r.IsInitialized() {
		r.started = false
		return errors.New("Device failed to initialize within timeout")
	}

	go func() {

		tickSpeedAndTemp := time.NewTicker(5 * time.Second)
		tickStatus := time.NewTicker(10 * time.Second)
		for r.IsStarted() {
			select {
			case <-tickSpeedAndTemp.C:
				r.requestSpeedAndTemp()
			case <-tickStatus.C:
				r.requestStatus()
				r.requestManualSettings()
			case <-r.change:
				r.checkForChanges()
			}
		}
		tickSpeedAndTemp.Stop()
		tickStatus.Stop()
	}()

	/**
	 * Wait for stop signal
	 */
	<-r.stop
	r.started = false
	return nil
}

// IsInitialized ...
func (r *Device) IsInitialized() bool {
	/*if r.init != 0x07 {
		return false
	}*/
	for _, c := range r.Channels {
		if !c.IsInitialized() {
			return false
		}
	}
	return true
}

// SetTempF ...
func (r *Device) SetTempF(b bool) {
	r.targetSetting = r.targetSetting ^ settingFarenheit
	if b {
		r.targetSetting = r.targetSetting | settingFarenheit
	}
	r.change <- true
}

// SetManual ...
func (r *Device) SetManual(b bool) {
	r.targetSetting = r.targetSetting ^ settingManual
	if b {
		r.targetSetting = r.targetSetting | settingManual
	}
	r.change <- true
}

// SetAudioEnabled ...
func (r *Device) SetAudioEnabled(b bool) {
	r.targetSetting = r.targetSetting ^ settingAudio
	if b {
		r.targetSetting = r.targetSetting | settingAudio
	}
	r.change <- true
}

// IsTempF ...
func (r *Device) IsTempF() bool {
	if r.curSetting&settingFarenheit == settingFarenheit {
		return true
	}
	return false
}

// IsManual ...
func (r *Device) IsManual() bool {
	if r.curSetting&settingManual == settingManual {
		return true
	}
	return false
}

// IsAudioEnabled ...
func (r *Device) IsAudioEnabled() bool {
	if r.curSetting&settingAudio == settingAudio {
		return true
	}
	return false
}

// DisplayChannel ...
func (r *Device) DisplayChannel() int {
	return int(r.displayChannel)
}

func (r *Device) requestStatus() {
	r.queue <- &request{ControlByte: reqDeviceStatus}
	r.queue <- &request{ControlByte: reqDeviceSettings}
	r.queue <- &request{ControlByte: reqDisplayChannel}
}

func (r *Device) requestSpeedAndTemp() {
	for i := range r.Channels {
		if i >= 0 && i < 5 {
			r.queue <- &request{ControlByte: reqTempAndSpeed + byte(i)}
		}
	}
}

func (r *Device) requestManualSettings() {
	for i := range r.Channels {
		if i >= 0 && i < 5 {
			r.queue <- &request{ControlByte: reqAlarmAndSpeed + byte(i)}
		}
	}
}

func (r *Device) loop() {
	for r.IsStarted() {

		buf := make([]byte, 8)
		if _, err := r.device.Read(buf); err == nil {
			req, err := parseRequest(buf)
			if err != nil {
				log.Println("Error while parsing data: ", err)
			} else {
				r.handleInput(req)
			}
		}
		select {
		case req := <-r.queue:
			r.device.Write(req.byteArray())
		default:
		}
	}
}

func (r *Device) checkForChanges() {

	for id, channel := range r.Channels {
		if channel.dirty {
			req := &request{
				Length:      3,
				ControlByte: setAlarmAndSpeed + byte(id),
				Data:        make([]byte, 3)}

			req.Data[0] = byte(channel.targetAlarm)
			req.Data[1], req.Data[2] = rpm2byte(channel.targetSpeed)
			r.queue <- req
			channel.dirty = false
			r.queue <- &request{ControlByte: reqAlarmAndSpeed + byte(id)}
		}
	}
	if r.curSetting != r.targetSetting {
		r.queue <- &request{
			ControlByte: setDeviceSettings,
			Data:        []byte{r.targetSetting},
		}
		r.curSetting = r.targetSetting
	}
}

func (r *Device) handleInput(req *request) {
	if req.ControlByte&0xf0 == recvTempAndSpeed {
		channel := int(req.ControlByte & 0x0f)
		if channel < len(r.Channels) {
			r.Channels[channel].report(req.Data)
		}
		return
	}
	if req.ControlByte&0xf0 == recvAlarmAndSpeed {
		channel := int(req.ControlByte & 0x0f)
		if channel < len(r.Channels) {
			r.Channels[channel].reportAlarm(req.Data)
		}
		return
	}
	if req.ControlByte&0xf0 == setDisplayChannel {
		if r.init&0x04 == 0 {
			r.init = r.init | 0x04
		}
		channel := int(req.ControlByte & 0x0f)
		r.displayChannel = uint8(channel)
		return
	}
	switch req.ControlByte {
	case recvDeviceStatus:
		if r.init&0x01 == 0 {
			r.init = r.init | 0x01
		}
	case recvDeviceSettings:
		if r.init&0x02 == 0 {
			r.init = r.init | 0x02
		}
		r.curSetting = req.Data[0]
	default:
		log.Println("Unknown data received", req)
	}

}
