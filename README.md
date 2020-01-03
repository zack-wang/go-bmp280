## input:
- d = "/dev/i2c-1"
- a = 0x76 (BMP280 on CJMCU8128 breakout board)
- accuracy = "ULTRA_LOW","LOW","STANDARD","HIGH","ULTRA_HIGH"

## output:
- p = uint32 ( pressure in unit Pa )

## Calibration:
Refer to https://cdn-shop.adafruit.com/datasheets/BST-BMP280-DS001-11.pdf

## Find your device address
````
i2cdetect -y 1
````

## Find BMP280 Device ID, should be 0x58 or 0x56 or 0x57
````
i2cget 1 0x76 0xD0
````

## Sample Code
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
