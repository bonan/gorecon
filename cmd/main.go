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

	"log"
	"net/http"

	"encoding/json"

	"github.com/bonan/gorecon"
	"github.com/bonan/gorecon/usbhid"
)

var reconDevices map[int]*gorecon.Device

func main() {

	reconDevices = map[int]*gorecon.Device{}

	rescan()
	fs := http.FileServer(http.Dir("ui"))

	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		dat, err := json.MarshalIndent(reconDevices, "", "  ")
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Write(dat)
	})
	http.Handle("/", fs)
	http.ListenAndServe(":8080", nil)
}

func rescan() {
	usbDevices := usbhid.Scan(0x0c45, 0x7100)

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
			delete(reconDevices, i)
		}(i, usb)
	}
}
