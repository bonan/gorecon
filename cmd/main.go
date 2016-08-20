package main

/**
 * Add to /etc/udev/rules.d/80-bitfinex-recon.rules:
 *
 * # Bitfinex Recon Fan controller, change group to a group accessible by your user
 * SUBSYSTEM=="usb", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7100", GROUP="..."
 *
 */

import (
	//"encoding/hex"

	"fmt"
	"log"
	"net/http"

	"github.com/bonan/gorecon"
	"github.com/bonan/gorecon/usbhid"
)

var reconDevices map[int]*gorecon.Device

func main() {

	usbDevices := usbhid.Scan(0x0c45, 0x7100)

	reconDevices = map[int]*gorecon.Device{}

	for i, usb := range usbDevices {
		go func(i int, usb *usbhid.UsbHid) {
			if usb.Start() != nil {
				return
			}
			dev := gorecon.NewDevice(usb, nil)
			defer usb.Close()
			reconDevices[i] = dev
			if err := dev.Start(); err != nil {
				log.Println("Error while starting device:", err)
			}
			reconDevices[i] = nil
		}(i, usb)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := "<table><tr><th>Device</th><th>Ch</th><th>Temp</th><th>Speed</th><th>MaxSpeed</th><th>ManualSpeed</th><th>AlarmTemp</th></tr>"

		for deviceID, recon := range reconDevices {
			if !recon.IsStarted() || !recon.IsInitialized() {
				continue
			}
			for channelID, channel := range recon.Channels {
				html = html + "<tr>"
				html = html + fmt.Sprintf("<td>%d</td>", deviceID)
				html = html + fmt.Sprintf("<td>%d</td>", channelID)
				html = html + fmt.Sprintf("<td>%d</td>", channel.TempC())
				html = html + fmt.Sprintf("<td>%d</td>", channel.Speed())
				html = html + fmt.Sprintf("<td>%d</td>", channel.MaxSpeed())
				html = html + fmt.Sprintf("<td>%d</td>", channel.ManualSpeed())
				html = html + fmt.Sprintf("<td>%d</td>", channel.AlarmTempC())
				html = html + "</tr>"
			}
		}
		html = html + "</table>"
		w.Write([]byte(html))
	})

	http.ListenAndServe(":8080", nil)
}
