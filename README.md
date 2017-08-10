# uCAM-III
Golang client / library for uCam-III

Eearly days yet, but JPEG seems to work ok.

## Example code (without error handling) 

```
import "github.com/hansj66/ucam-iii/ucam"
import "log"
import "io/ioutil"

func main() {
  camera := ucam.NewCamera("/dev/cu.wchusbserial14520", 9600)
	camera.Log(true)

	camera.Connect()
	camera.DisableSleepTimeout()
	camera.SetImageFormats(ucam.JPEG16Bit, ucam.RAW128x128, ucam.JPEG640x480)
	camera.SetExposure(ucam.CNormal, ucam.BNormal, ucam.EZero)
	camera.SetLightFrequency(50)
	camera.SetPackageSize(512)
	camera.Snapshot(ucam.JPEG)
	image, _ := camera.GetPicture(ucam.Snapshot)

	ioutil.WriteFile("./test.jpg", image, 0644)

}
```
