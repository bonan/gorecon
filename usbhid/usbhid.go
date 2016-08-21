package usbhid

import (
	"io"
	"time"

	"github.com/zserge/hid"
)

// UsbHid ...
type UsbHid struct {
	device hid.Device
	open   bool
}

// Scan ...
func Scan(vendor, product uint16) []*UsbHid {

	devices := []*UsbHid{}
	hid.UsbWalk(func(device hid.Device) {
		info := device.Info()

		if info.Vendor == vendor && info.Product == product {
			dev := &UsbHid{
				device: device,
				open:   false,
			}
			devices = append(devices, dev)
		}
	})
	return devices
}

// Read ...
func (u *UsbHid) Read(p []byte) (n int, err error) {
	if !u.open {
		return 0, io.ErrClosedPipe
	}
	l := len(p)
	data, err := u.device.Read(l, 1*time.Second)
	if err != nil {
		return 0, err
	}
	for i, v := range data {
		if l-1 < i {
			break
		}
		p[i] = v
	}
	return len(data), nil
}

// Write ...
func (u *UsbHid) Write(p []byte) (n int, err error) {
	if !u.open {
		return 0, io.ErrClosedPipe
	}
	l, err := u.device.Write(p, 1*time.Second)
	if err != nil {
		return 0, err
	}
	return l, nil
}

// Start ...
func (u *UsbHid) Start() error {
	if err := u.device.Open(); err != nil {
		return err
	}
	if _, err := u.device.HIDReport(); err != nil {
		u.device.Close()
		return err
	}
	u.open = true
	return nil
}

// Close ...
func (u *UsbHid) Close() error {
	u.open = false
	u.device.Close()
	return nil
}
