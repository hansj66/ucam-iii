package main

import "github.com/hansj66/ucam-iii/ucam"
import "log"
import "io/ioutil"

func main() {

	// To be 100% sure of camera state, a hardware reset should be issued before communicating with the camera. (The reset pin is active low.)

	camera := ucam.NewCamera("/dev/cu.wchusbserial14520", 9600)
	camera.Log(true)

	var err error

	err = camera.Connect()
	if err != nil {
		log.Fatal(err)
	}
	err = camera.DisableSleepTimeout()
	if err != nil {
		log.Fatal(err)
	}
	err = camera.SetImageFormats(ucam.JPEG16Bit, ucam.RAW128x128, ucam.JPEG640x480)
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
	err = camera.SetPackageSize(512)
	if err != nil {
		log.Fatal(err)
	}
	err = camera.Snapshot(ucam.JPEG)
	if err != nil {
		log.Fatal(err)
	}
	image, errPic := camera.GetPicture(ucam.Snapshot)
	if errPic != nil {
		log.Fatal(errPic)
	}

	err = ioutil.WriteFile("./test.jpg", image, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
