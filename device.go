package gorecon

import (
	"errors"
	"io"
	"log"
	"time"
)

// Device represents a Bitfinex Recon device
type Device struct {
	device   io.ReadWriter
	started  bool
	queue    chan *request
	stop     chan bool
	change   chan bool
	report   chan *Report
	init     byte
	Channels map[int]*Channel
}

type Report struct {
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
	r.Channels = map[int]*Channel{
		0: NewChannel(r.change),
		1: NewChannel(r.change),
		2: NewChannel(r.change),
		3: NewChannel(r.change),
		4: NewChannel(r.change),
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

func (r *Device) requestStatus() {
	r.queue <- &request{ControlByte: getDeviceStatus}
	r.queue <- &request{ControlByte: getDeviceSettings}
	r.queue <- &request{ControlByte: getCurrentChannel}
}

func (r *Device) requestSpeedAndTemp() {
	for i := range r.Channels {
		if i >= 0 && i < 5 {
			r.queue <- &request{ControlByte: getTempAndSpeedChannel0 + byte(i)}
		}
	}
}

func (r *Device) requestManualSettings() {
	for i := range r.Channels {
		if i >= 0 && i < 5 {
			r.queue <- &request{ControlByte: getAlarmAndSpeedChannel0 + byte(i)}
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
				ControlByte: setAlarmAndSpeedChannel0 + byte(id),
				Data:        make([]byte, 3)}

			req.Data[0] = byte(channel.targetAlarm)
			req.Data[1], req.Data[2] = rpm2byte(channel.targetSpeed)
			r.queue <- req
			channel.dirty = false
			r.queue <- &request{ControlByte: getAlarmAndSpeedChannel0 + byte(id)}
		}
	}
}

func (r *Device) handleInput(req *request) {
	if req.ControlByte&0xf0 == readTempAndSpeedChannel0 {
		channel := int(req.ControlByte & 0x0f)
		if ch, ok := r.Channels[channel]; ok {
			ch.report(req.Data)
		}
		return
	}
	if req.ControlByte&0xf0 == readAlarmAndSpeedChannel0 {
		channel := int(req.ControlByte & 0x0f)
		if ch, ok := r.Channels[channel]; ok {
			ch.reportAlarm(req.Data)
		}
		return
	}
	switch req.ControlByte {
	case readDeviceStatus:
		if r.init&0x01 == 0 {
			r.init = r.init | 0x01
		}
		log.Println("Device status", req)
	case readDeviceSettings:
		if r.init&0x02 == 0 {
			r.init = r.init | 0x02
		}
		log.Println("Device settings", req)
	case readCurrentChannel:
		if r.init&0x04 == 0 {
			r.init = r.init | 0x04
		}
		log.Println("Current channel", req)
	default:
		log.Println("Unknown data received", req)
	}

}
