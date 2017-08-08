package main

import "github.com/hansj66/ucam-iii/ucam"
import "log"

func main() {

	// To be 100% sure of camera state, a hardware reset should be issued before communicating with the camera

	camera := ucam.NewCamera("/dev/cu.wchusbserial14120")
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
	err = camera.SetImageFormats(ucam.RAW16BitColourRGB, ucam.RAW128x128, ucam.JPEG640x480)
	if err != nil {
		log.Fatal(err)
	}
	err = camera.SetExposure(ucam.CNormal, ucam.BNormal, ucam.EZero)
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

	_, err = camera.GetPicture(ucam.Snapshot)
	if err != nil {
		log.Fatal(err)
	}

}

/*

Skriv som PPM-filer ?
Se på behov for sleep i mellom kall
Sjekk mulighet for å endre baudrate on the fly på bibliotekssiden

Se nederst side 9 mht om package size må settes eller ikke (for RAW)
SetPackageSize
Light
Snapshot / Get picture

lagre bildet som PNG

Send ACK i retur (sjekk spek)
*/
