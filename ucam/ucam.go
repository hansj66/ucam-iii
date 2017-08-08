package ucam

import (
	"bytes"
	"log"
	"time"

	"errors"

	"fmt"

	"github.com/tarm/serial"
)

const maxRetries int = 60

type CameraResponseType byte

const (
	ACK     CameraResponseType = 0
	NAK     CameraResponseType = 1
	TIMEOUT CameraResponseType = 2
	EOF     CameraResponseType = 4
)

type FormatType byte

const (
	CrYCbY FormatType = 1
	RGB    FormatType = 2
)

type SnapshotType byte

const (
	RAW      SnapshotType = 0x01
	JPEG     SnapshotType = 0x02
	Snapshot SnapshotType = 0x05
)

type ImageFormatType byte

const (
	RAW8BitGrayScale     ImageFormatType = 0x03
	RAW16BitColourCrYCbY ImageFormatType = 0x08
	RAW16BitColourRGB    ImageFormatType = 0x06
	JPEG16Bit            ImageFormatType = 0x07
)

type RAWResolutionType byte

const (
	RAW80x60   RAWResolutionType = 0x01
	RAW160x120 RAWResolutionType = 0x03
	RAW128x128 RAWResolutionType = 0x09
	RAW128x96  RAWResolutionType = 0x0B
)

type JPEGResolutionType byte

const (
	JPEG160x128 JPEGResolutionType = 0x03
	JPEG320x240 JPEGResolutionType = 0x05
	JPEG640x480 JPEGResolutionType = 0x07
)

type ContrastType byte

const (
	CMin    ContrastType = 0x00
	CLow    ContrastType = 0x01
	CNormal ContrastType = 0x02
	CHigh   ContrastType = 0x03
	CMax    ContrastType = 0x04
)

type BrightnessType byte

const (
	BMin    BrightnessType = 0x00
	BLow    BrightnessType = 0x01
	BNormal BrightnessType = 0x02
	BHigh   BrightnessType = 0x03
	BMax    BrightnessType = 0x04
)

type ExposureType byte

const (
	EMinusTwo ExposureType = 0x00
	EMinusOne ExposureType = 0x00
	EZero     ExposureType = 0x00
	EPlusOne  ExposureType = 0x00
	EPlusTwo  ExposureType = 0x00
)

type Camera struct {
	port        *serial.Port
	ImageFormat ImageFormatType
	XSize       uint16
	YSize       uint16
	Bits        uint16
	logging     bool
	packageSize uint16
}

func NewCamera(serialPortName string) Camera {
	log.Println("Opening com port")
	config := &serial.Config{Name: serialPortName, Baud: 57600, ReadTimeout: time.Millisecond * 250}
	port, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}

	return Camera{
		port:    port,
		logging: false,
	}
}

func (c *Camera) Log(enable bool) {
	c.logging = enable
}

// uCam-III command templates
var syncCommand = []byte{0xAA, 0x0D, 0x00, 0x00, 0x00, 0x00}
var sleepCommand = []byte{0xAA, 0x15, 0x00, 0x00, 0x00, 0x00}
var initialCommand = []byte{0xAA, 0x01, 0x00, 0x00, 0x00, 0x00}
var exposureCommand = []byte{0xAA, 0x14, 0x00, 0x00, 0x00, 0x00}
var snapshotCommand = []byte{0xAA, 0x05, 0x00, 0x00, 0x00, 0x00}
var setPackageSizeCommand = []byte{0xAA, 0x06, 0x08, 0x00, 0x00, 0x00}
var getPictureCommand = []byte{0xAA, 0x04, 0x00, 0x00, 0x00, 0x00}

var timeout = 5 * time.Millisecond

var ack = []byte{0xAA, 0x0E, 0x00, 0x00, 0x00, 0x00}
var nak = []byte{0xAA, 0x0F, 0x00, 0x00, 0x00, 0x00}

func (c *Camera) command(command []byte) CameraResponseType {
	_, err := c.port.Write(command)
	if err != nil {
		log.Fatal(err)
	}
	if c.logging {
		log.Println("Sent    : ", fmt.Sprintf("%02X %02X %02X %02X %02X %02X", command[0], command[1], command[2], command[3], command[4], command[5]))
	}
	buf := make([]byte, 12)
	n, err2 := c.port.Read(buf)
	if n == 0 {
		return EOF
	} else if err2 != nil {
		log.Fatal(err)
	}

	if c.logging {
		log.Println("Received: ", fmt.Sprintf("%02X %02X %02X %02X %02X %02X %02X %02X", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7]))
	}
	if bytes.Equal(buf[:2], nak[:2]) {
		time.Sleep(time.Millisecond * 500)
		return NAK
	}
	if bytes.Equal(buf[:2], ack[:2]) {
		timeout = 0
		// This is not in the documentation, but it looks like communication is a couple
		// of orders of magnitude more stable if we introduce a delay after each command
		time.Sleep(time.Millisecond * 500)

		return ACK
	}
	time.Sleep(timeout)
	timeout++

	return TIMEOUT
}

// Connect sends sync command to camera. Camera should respond with ACK
// Camera will normally ack within 25 attempts
func (c *Camera) Connect() error {
	for attempt := 0; attempt < maxRetries; attempt++ {
		response := c.command(syncCommand)
		if response == ACK {
			// Acknowledge sync
			ack[2] = 0x0D
			_, errAck := c.port.Write(ack)
			if errAck != nil {
				log.Fatal(errAck)
			}

			log.Println("ACK     : ", fmt.Sprintf("%02X %02X %02X %02X %02X %02X", ack[0], ack[1], ack[2], ack[3], ack[4], ack[5]))

			if c.logging {
				log.Println("Established connection with camera")
			}
			return nil
		}
	}
	return errors.New("Unable to connect to camera (Sync command failed. Invalid camera response.)")
}

// DisableSleepTimeout disables sleep mode, by sending 0 as the sleep argument
func (c *Camera) DisableSleepTimeout() error {
	return c.SetSleepTimeout(0)
}

// SetSleepTimeout adjusts the sleep timeout from the default 15 seconds to the specified timeoutValue
func (c *Camera) SetSleepTimeout(timeoutValue byte) error {
	sleepCommand[2] = timeoutValue
	// The sleep command is a bit flaky. It may require multiple attempts before the camera responds with ACK
	for attempt := 0; attempt < maxRetries; attempt++ {
		response := c.command(sleepCommand)
		if response == ACK {
			if c.logging {
				log.Printf("Sleep timeout set to : %d\n", timeoutValue)
			}
			return nil
		}
	}
	return fmt.Errorf("Sleep command failed.")
}

func (c *Camera) SetImageFormats(format ImageFormatType, rawResolution RAWResolutionType, jpegResolution JPEGResolutionType) error {
	initialCommand[3] = byte(format)
	initialCommand[4] = byte(rawResolution)
	initialCommand[5] = byte(jpegResolution)
	for attempt := 0; attempt < maxRetries; attempt++ {
		response := c.command(initialCommand)
		if response == ACK {
			if c.logging {
				log.Printf("Image format and resolution set.\n")
			}
			return nil
		}
	}
	return fmt.Errorf("Initialize command failed")
}

func (c *Camera) SetExposure(contrast ContrastType, brightness BrightnessType, exposure ExposureType) error {
	exposureCommand[2] = byte(contrast)
	exposureCommand[3] = byte(brightness)
	exposureCommand[4] = byte(exposure)
	for attempt := 0; attempt < maxRetries; attempt++ {
		response := c.command(exposureCommand)
		if response == ACK {
			if c.logging {
				log.Printf("Exposure set.\n")
			}
			return nil
		}
	}
	return fmt.Errorf("Initialize command failed")
}

func (c *Camera) SetPackageSize(size uint16) error {
	for attempt := 0; attempt < maxRetries; attempt++ {
		setPackageSizeCommand[3] = byte(size & 0xFF)
		setPackageSizeCommand[4] = byte((size & 0xFF00) >> 8)
		response := c.command(setPackageSizeCommand)
		if response == ACK {
			if c.logging {
				log.Printf("Package size set.\n")
			}
			c.packageSize = size
			return nil
		}
	}
	return fmt.Errorf("Snapshot command failed")
}

func (c *Camera) Snapshot(pictureType SnapshotType) error {
	switch pictureType {
	case RAW:
		snapshotCommand[2] = 0x01
	case JPEG:
		snapshotCommand[2] = 0x00
	}
	for attempt := 0; attempt < maxRetries; attempt++ {
		response := c.command(snapshotCommand)
		if response == ACK {
			if c.logging {
				log.Printf("Snapshot held in buffer.\n")
			}
			return nil
		}
	}
	return fmt.Errorf("Snapshot command failed")
}

func (c *Camera) GetPicture(pictureType SnapshotType) ([]byte, error) {
	switch pictureType {
	case RAW:
		getPictureCommand[2] = 0x02
	case JPEG:
		getPictureCommand[2] = 0x05
	case Snapshot:
		getPictureCommand[2] = 0x01
	}
	for attempt := 0; attempt < maxRetries; attempt++ {

		response := c.command(getPictureCommand)

		if response == ACK {
			if c.logging {
				log.Printf("Fetching image data.\n")
			}

			var framecounter byte
			var frameRequestCommand = []byte{0xAA, 0x0E, 0x00, 0x00, 0x00, 0x00}
			var bytecounter int

			for {
				frameRequestCommand[2] = framecounter

				_, errRequest := c.port.Write(frameRequestCommand)
				if errRequest != nil {
					log.Fatal(errRequest)
				}

				frameReceiveBuffer := make([]byte, c.packageSize)
				n, err := c.port.Read(frameReceiveBuffer)

				bytecounter += n
				if err != nil {
					log.Fatal(err)
				}
				// TODO: Checksum

				log.Printf("Received %d bytes\n", n)

				_, errAck := c.port.Write(ack)
				if errAck != nil {
					log.Fatal(errAck)
				}
				// time.Sleep(time.Millisecond * 250)

				// TODO: Read out ID + size. Compose receive buffer
			}

		}
	}
	return nil, fmt.Errorf("Snapshot command failed")
}
