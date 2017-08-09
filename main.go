package main

import "github.com/hansj66/ucam-iii/ucam"
import "log"

func main() {

	// To be 100% sure of camera state, a hardware reset should be issued before communicating with the camera. (The reset pin is active low.)

	camera := ucam.NewCamera("/dev/cu.wchusbserial14110", 9600)
	camera.Log(true)

	var err error
	//var buf []byte

	err = camera.Connect()
	if err != nil {
		log.Fatal(err)
	}
	err = camera.DisableSleepTimeout()
	if err != nil {
		log.Fatal(err)
	}

	err = camera.SetBaudRate(9600)
	if err != nil {
		log.Fatal(err)
	}

	err = camera.SetImageFormats(ucam.JPEG16Bit, ucam.RAW128x128, ucam.JPEG160x128)
	if err != nil {
		log.Fatal(err)
	}
	err = camera.SetExposure(ucam.CNormal, ucam.BNormal, ucam.EZero)
	if err != nil {
		log.Fatal(err)
	}
	err = camera.SetLightFrequency(50)
	if err != nil {
		log.Fatal(err)
	}
	err = camera.SetPackageSize(128)
	if err != nil {
		log.Fatal(err)
	}
	err = camera.Snapshot(ucam.JPEG)
	if err != nil {
		log.Fatal(err)
	}

	_, err = camera.GetPicture(ucam.Snapshot)
	if err != nil {
		log.Fatal(err)
	}

}
