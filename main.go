package main

import (
	"log"
	"time"

	"github.com/hansj66/ucam-iii/ucam"

	"github.com/tarm/serial"
)

func main() {

	log.Println("Opening com port")

	config := &serial.Config{Name: "COM18", Baud: 57600, ReadTimeout: time.Millisecond * 1000}
	port, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("com port opened ok")

	log.Print("Syncing...")
	err = ucam.Sync(port)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Sync complete")

	log.Println("Disabling sleep mode.")
	err = ucam.DisableSleepTimeout(port)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Sleep timeout disabled.")

}
