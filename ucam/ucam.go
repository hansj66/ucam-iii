package ucam

import (
	"bytes"
	"log"
	"time"

	"errors"

	"fmt"

	"github.com/tarm/serial"
)

type CameraResponseType byte

const (
	ACK     CameraResponseType = 0
	NAK     CameraResponseType = 1
	TIMEOUT CameraResponseType = 2
)

var sync = []byte{0xAA, 0x0D, 0x00, 0x00, 0x00, 0x00}
var sleep = []byte{0xAA, 0x15, 0x00, 0x00, 0x00, 0x00}

var ack = []byte{0xAA, 0x0E}
var nak = []byte{0xAA, 0x0F}

func cameraCommand(port *serial.Port, command []byte) CameraResponseType {
	log.Println("Sent    : ", command)
	_, err := port.Write(command)
	if err != nil {
		log.Fatal(err)
	}

	//	time.Sleep(10 * time.Millisecond)

	buf := make([]byte, 128)
	n, err2 := port.Read(buf)
	log.Println("Received: ", buf[0:n])
	if err2 != nil {
		log.Fatal(err)
	}
	if bytes.Equal(buf[0:2], nak[0:2]) {
		log.Println("NAK")
		return NAK
	}
	if bytes.Equal(buf[0:2], ack[0:2]) {
		log.Println("ACK")
		return ACK
	}
	time.Sleep(50 * time.Millisecond)
	//	log.Println("TIMEOUT")

	return TIMEOUT
}

// Sync sends sync command to camera. Camera should respond with ACK
// Camera will normally ack within 25 attempts
func Sync(port *serial.Port) error {
	for attempt := 0; attempt < 60; attempt++ {
		// TODO: 5 ms delay + 1 ms pr retry
		response := cameraCommand(port, sync)
		if response == ACK {
			return nil
		}
	}

	return errors.New("Sync command failed. Invalid camera response")
}

// DisableSleepTimeout disables sleep mode, by sending 0 as the sleep argument
func DisableSleepTimeout(port *serial.Port) error {
	return SetSleepTimeout(port, 0)
}

// SetSleepTimeout adjusts the sleep timeout from the default 15 seconds to the specified timeoutValue
func SetSleepTimeout(port *serial.Port, timeoutValue byte) error {
	sleep[2] = timeoutValue
	response := cameraCommand(port, sleep)
	if response == ACK {
		log.Printf("Sleep timeout set to : %d\n", timeoutValue)
		return nil
	}
	return fmt.Errorf("Sleep command failed. Error: %d", response)
}
