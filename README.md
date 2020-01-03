## input:
- d = "/dev/i2c-1"
- a = 0x76 (BMP280 on CJMCU8128 breakout board)
- accuracy = "ULTRA_LOW","LOW","STANDARD","HIGH","ULTRA_HIGH"

## output:
- p = uint32 ( pressure in unit Pa )

## Calibration:
Refer to https://cdn-shop.adafruit.com/datasheets/BST-BMP280-DS001-11.pdf

````
package main
import(
	"log"
	"github.com/zack-wang/go-bmp280"

)
func main(){
// Atmosphere Pressure
	p,err:=bmp280.ReadPressurePa("/dev/i2c-1",0x76,"LOW")
	if err!=nil{
	  log.Fatal("Not BMP280",err)
	}else{
	  log.Println("p=",p)
	}
}
````
