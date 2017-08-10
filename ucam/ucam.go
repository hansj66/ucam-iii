package ucam

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

const maxRetries int = 60

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
	baudRate    int
	port        *serial.Port
	portName    string
	ImageFormat ImageFormatType
	XSize       uint16
	YSize       uint16
	Bits        uint16
	logging     bool
	packageSize uint16
}

// NewCamera ctreates a new camera instance
func NewCamera(portName string, baudrate int) Camera {
	log.Println("Opening com port")
	config := &serial.Config{Name: portName, Baud: baudrate, ReadTimeout: time.Millisecond * 500}
	port, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}

	return Camera{
		portName: portName,
		baudRate: 115200,
		port:     port,
		logging:  false,
	}
}

// Log enables/disables logging
func (c *Camera) Log(enable bool) {
	c.logging = enable
}

const defaultTimeout = 5 * time.Millisecond // [8.1 Synchronizing the uCam-III]
var timeout = defaultTimeout                // [8.1 Synchronizing the uCam-III]

var nak = []byte{0xAA, 0x0F, 0x00, 0x00, 0x00, 0x00}

func errorLookup(errorCode byte) string {
	switch errorCode {
	case 0x01:
		return "Picture Type Error"
	case 0x0B:
		return "Parameter Error"
	case 0x02:
		return "Picture Up Scale"
	case 0x0C:
		return "Send Register Timeout"
	case 0x03:
		return "Picture Scale Error"
	case 0x0D:
		return "Command ID Error"
	case 0x04:
		return "Unexpected Reply 04h"
	case 0x0F:
		return "Picture Not Ready 0Fh"
	case 0x05:
		return "Send Picture Timeout"
	case 0x10:
		return "Transfer Package Number Error"
	case 0x06:
		return "Unexpected Command"
	case 0x11:
		return "Set Transfer Package Size Wrong"
	case 0x07:
		return "SRAM JPEG Type Error"
	case 0xF0:
		return "Command Header Error"
	case 0x08:
		return "SRAM JPEG Size Error"
	case 0xF1:
		return "Command Length Error"
	case 0x09:
		return "Picture Format Error"
	case 0xF5:
		return "Send Picture Error"
	case 0x0A:
		return "Picture Size Error"
	case 0xFF:
		return "Send Command Error"
	}
	return "Doh! Undocumented error"
}

func commandLookup(commandID byte) string {
	switch commandID {
	case 0x01:
		return "INITIAL"
	case 0x04:
		return "GET PICTURE"
	case 0x05:
		return "SNAPSHOT"
	case 0x06:
		return "SET PACKAGE SIZE"
	case 0x07:
		return "SET BAUD RATE"
	case 0x08:
		return "RESET"
	case 0x0A:
		return "DATA"
	case 0x0D:
		return "SYNC"
	case 0x0E:
		return "ACK"
	case 0x0F:
		return "NAK"
	case 0x13:
		return "LIGHT"
	case 0x14:
		return "CONTRAST / BRIGHTNESS / EXPOSURE"
	case 0x15:
		return "SLEEP"
	}
	return fmt.Sprintf("Unknown command : %d", commandID)
}

// command method executes a uCam-III command [7. Command set]
func (c *Camera) command(command []byte) ([]byte, error) {
	var ack = []byte{0xAA, 0x0E, 0x00, 0x00, 0x00, 0x00}

	response := make([]byte, 128)
	for attempt := 0; attempt < maxRetries; attempt++ {
		_, errorWrite := c.port.Write(command)
		if errorWrite != nil {
			return nil, errorWrite
		}
		if c.logging {
			log.Println("Sent    : ", fmt.Sprintf("%02X %02X %02X %02X %02X %02X", command[0], command[1], command[2], command[3], command[4], command[5]))
		}

		n, errorRead := c.port.Read(response)
		// EOF is ok. The camera doesn't respond if it is asleep
		if errorRead != nil && n != 0 {
			return nil, errorRead
		}

		if c.logging {
			log.Println("Received: ", fmt.Sprintf("%02X %02X %02X %02X %02X %02X %02X %02X %02X %02X %02X %02X", response[0], response[1], response[2], response[3], response[4], response[5], response[6], response[7], response[8], response[9], response[10], response[11]))
		}

		if bytes.Equal(response[:2], ack[:2]) {
			timeout = defaultTimeout
			return response, nil
		} else if bytes.Equal(response[:2], nak[:2]) {
			log.Println(errorLookup(response[4]), " : ", fmt.Sprintf("%02X %02X %02X %02X %02X %02X %02X %02X %02X %02X %02X %02X", response[0], response[1], response[2], response[3], response[4], response[5], response[6], response[7], response[8], response[9], response[10], response[11]))
		}

		time.Sleep(timeout)
		timeout++
	}
	return nil, fmt.Errorf("No camera response. Command %02d failed (%s)", command[1], commandLookup(command[1]))
}

// Light : [7.11 LIGHT (AA13h)]
func (c *Camera) SetLightFrequency(frequency byte) error {
	if (frequency != 50) && (frequency != 60) {
		return fmt.Errorf("Invalid frequency(%d)", frequency)
	}
	var frequencyType byte
	switch frequency {
	case 50:
		frequencyType = 0
	case 60:
		frequencyType = 1
	}
	_, err := c.command([]byte{0xAA, 0x13, frequencyType, 0x00, 0x00, 0x00})
	if err == nil && c.logging {
		log.Printf("Light frequency set to %d Hz\n", frequency)
	}

	return err
}

// Light : [7.5 SET BAUD RATE (AA07h)]
// In case of NAK/COmmand Id errors, try lowering the baud rate.
func (c *Camera) SetBaudRate(baudrate int) error {
	var d1 byte
	var d2 byte

	switch baudrate {
	case 2400:
		d1 = 31
		d2 = 47
	case 4800:
		d1 = 31
		d2 = 23
	case 9600:
		d1 = 31
		d2 = 11
	case 19200:
		d1 = 31
		d2 = 5
	case 38400:
		d1 = 31
		d2 = 2
	case 57600:
		d1 = 31
		d2 = 1
	case 115200:
		d1 = 31
		d2 = 0
	case 153600:
		d1 = 7
		d2 = 2
	case 230400:
		d1 = 7
		d2 = 1
	case 460800:
		d1 = 7
		d2 = 0
	case 921600:
		d1 = 1
		d2 = 1
	case 1228800:
		d1 = 2
		d2 = 0
	case 1843200:
		d1 = 1
		d2 = 0
	case 3686400:
		d1 = 0
		d2 = 0
	default:
		return fmt.Errorf("Invalid baudrate (%d). Valid baudrates are: 2400, 4800, 9600, 19200, 38400, 57600, 115200, 153600, 230400, 460800, 921600, 1228800, 1843200, 3686400", baudrate)
	}

	_, err := c.command([]byte{0xAA, 0x07, d1, d2, 0x00, 0x00})
	if err == nil && c.logging {
		log.Printf("Baud rate changed to %d\n", baudrate)
	}

	c.port.Flush()
	c.port.Close()
	config := &serial.Config{Name: c.portName, Baud: baudrate, ReadTimeout: time.Millisecond * 500}
	port, _ := serial.OpenPort(config)
	c.port = port

	time.Sleep(time.Millisecond * 500)

	return err
}

// Connect sends sync command to camera. Camera should respond with ACK
// Camera will normally ack within 25-60
// [8.1 Synchronizing the uCam-III]
func (c *Camera) Connect() error {
	ack := []byte{0xAA, 0x0E, 0x00, 0x00, 0x00, 0x00}

	_, err := c.command([]byte{0xAA, 0x0D, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return err
	}
	// Acknowledge sync
	ack[2] = 0x0D
	_, errAck := c.port.Write(ack)
	if errAck != nil {
		log.Fatal(errAck)
	}

	if c.logging {
		log.Println("Established connection with camera")
	}
	c.port.Flush()
	time.Sleep(time.Millisecond * 2000) // [13. 'First Photo Delay']
	return nil
}

// DisableSleepTimeout disables sleep mode, by sending 0 as the sleep argument
func (c *Camera) DisableSleepTimeout() error {
	return c.SetSleepTimeout(0)
}

// SetSleepTimeout adjusts the sleep timeout from the default 15 seconds to the specified timeoutValue [7.13 SLEEP (AA15h)]
func (c *Camera) SetSleepTimeout(timeoutValue byte) error {
	var sleepCommand = []byte{0xAA, 0x15, timeoutValue, 0x00, 0x00, 0x00}

	_, err := c.command(sleepCommand)
	if err == nil && c.logging {
		log.Printf("Sleep timeout set to : %d\n", timeoutValue)
	}
	return err
}

// SetImageFormats : [7.1 INITIAL (AA01h)]
func (c *Camera) SetImageFormats(format ImageFormatType, rawResolution RAWResolutionType, jpegResolution JPEGResolutionType) error {
	_, err := c.command([]byte{0xAA, 0x01, 0x00, byte(format), byte(rawResolution), byte(jpegResolution)})
	if err == nil && c.logging {
		log.Printf("Image format and resolution set.\n")

	}
	return err
}

// SetExposure : [7.12 CONTRAST/BRIGHTNESS/EXPOSURE (AA14h)]
func (c *Camera) SetExposure(contrast ContrastType, brightness BrightnessType, exposure ExposureType) error {
	_, err := c.command([]byte{0xAA, 0x14, byte(contrast), byte(brightness), byte(exposure), 0x00})
	if err == nil && c.logging {
		log.Printf("Exposure set.\n")
	}
	return err
}

// SetPackageSize : [7.4 PACKAGE SIZE (AA06)]
func (c *Camera) SetPackageSize(size uint16) error {
	_, err := c.command([]byte{0xAA, 0x06, 0x08, byte(size & 0xFF), byte((size & 0xFF00) >> 8), 0x00})
	if err == nil && c.logging {
		log.Printf("Package size set.\n")
	}
	c.packageSize = size
	return err
}

// Snapshot : [7.3 SNAPSHOT (AA05h)]
func (c *Camera) Snapshot(pictureType SnapshotType) error {
	var snapshotCommand = []byte{0xAA, 0x05, 0x00, 0x00, 0x00, 0x00}
	switch pictureType {
	case RAW:
		snapshotCommand[2] = 0x01
	case JPEG:
		snapshotCommand[2] = 0x00
	}
	_, err := c.command(snapshotCommand)
	if err == nil && c.logging {
		log.Printf("Snapshot held in buffer.\n")
	}
	time.Sleep(time.Millisecond * 200) // [13. 'Shutter Delay']
	return err
}

// GetPicture : [7.2 GET PICTURE (AA04h)]
func (c *Camera) GetPicture(pictureType SnapshotType) ([]byte, error) {
	// TODO: Split into separate methods (PacketSize is only relevant for JPEG)
	var getPictureCommand = []byte{0xAA, 0x04, 0x00, 0x00, 0x00, 0x00}
	switch pictureType {
	case RAW:
		getPictureCommand[2] = 0x02
	case JPEG:
		getPictureCommand[2] = 0x05
	case Snapshot:
		getPictureCommand[2] = 0x01
	}

	response, err := c.command(getPictureCommand)
	if err != nil {
		return nil, err
	}

	imageSize := uint16(response[10])<<8 + uint16(response[9])
	if c.logging {
		log.Printf("Image size is: %d bytes", imageSize)
	}
	imageBuffer := make([]byte, imageSize)

	if c.logging {
		log.Printf("Fetching image data.\n")
	}

	frameAck := []byte{0xAA, 0x0E, 0x00, 0x00, 0x00, 0x00}

	_, errRequest := c.port.Write(frameAck)
	if errRequest != nil {
		log.Fatal(errRequest)
	}

	var imageBufferInsertIndex uint16
	var allFramesRead bool
	for {
		frameReceiveBuffer := make([]byte, c.packageSize)
		var bufferInsertIndex uint16
		var expectedDataSizeRead = false
		var frameSize uint16
		for bufferInsertIndex < c.packageSize {
			readBytes, _ := c.port.Read(frameReceiveBuffer[bufferInsertIndex:])
			bufferInsertIndex += uint16(readBytes)
			if readBytes >= 4 && !expectedDataSizeRead {
				frameSize = uint16(frameReceiveBuffer[2]) + uint16(frameReceiveBuffer[3])<<8
				log.Printf("Framesize is: %d\n", frameSize)
				expectedDataSizeRead = true
			}
			if readBytes == 0 && frameSize <= (c.packageSize+6) {
				allFramesRead = true
				break
			}
		}

		frameAck[4] = byte(frameReceiveBuffer[0])
		frameAck[5] = byte(frameReceiveBuffer[1])
		/*		log.Printf("ACK  %02X %02X %02X %02X %02X %02X\n",
				frameAck[0], frameAck[1], frameAck[2], frameAck[3], frameAck[4], frameAck[5])
		*/
		_, errAck := c.port.Write(frameAck)
		if errAck != nil {
			log.Fatal(errAck)
		}

		if expectedDataSizeRead {
			copy(imageBuffer[imageBufferInsertIndex:], frameReceiveBuffer[4:frameSize+4])
			imageBufferInsertIndex += frameSize
			log.Println("InsertIndex: ", imageBufferInsertIndex, "image size :", len(imageBuffer))
		}
		if allFramesRead {
			break
		}
	}

	return imageBuffer, nil
}
