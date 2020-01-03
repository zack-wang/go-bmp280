//--------------------------------------------------------------------------------------------------
//
// Copyright (c) 2020 zack Wang
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
// associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial
// portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
//--------------------------------------------------------------------------------------------------

package bmp280

import (
	"log"
	"bytes"
	"encoding/binary"
	"time"
	"golang.org/x/exp/io/i2c"
)

// BMP280 sensors memory map
const (
	// BMP280 general registers
	BMP280_ID_REG        = 0xD0
	BMP280_STATUS_REG    = 0xF3
	BMP280_CNTR_MEAS_REG = 0xF4
	BMP280_CONFIG        = 0xF5
	BMP280_RESET         = 0xE0
	// BMP280 specific compensation register's block
	BMP280_COEF_START = 0x88
	BMP280_COEF_BYTES = 12 * 2
	// BMP280 specific 3-byte reading out temprature and preassure
	BMP280_PRES_3BYTES = 0xF7
	BMP280_TEMP_3BYTES  = 0xFA
)

// Unique BMP280 calibration coefficients
type BMP280COEFF struct {
	// Calibration data
	Dig_T1 uint16
	Dig_T2 int16
	Dig_T3 int16
	Dig_P1 uint16
	Dig_P2 int16
	Dig_P3 int16
	Dig_P4 int16
	Dig_P5 int16
	Dig_P6 int16
	Dig_P7 int16
	Dig_P8 int16
	Dig_P9 int16
}

var (
	Cal BMP280COEFF
	//// Private Variables
	b1=make([]byte,1)
	b2=make([]byte,2)
	b3=make([]byte,3)
	b4=make([]byte,4)
	b24=make([]byte,24)
)

// Verify BMP280
func VerifiySensorID(d string, a int) bool {
  i2c0,err:=i2c.Open(&i2c.Devfs{Dev:d},a)
  defer i2c0.Close()

  if err!=nil{
    log.Println("I2C Error")
    return false
  }else{
	  err = i2c0.ReadReg(BMP280_ID_REG,b1)
	  if err != nil {
		   return false
	  }else{
      if b1[0]!=byte(0x58){
        log.Println("Not BMP280")
        return false
      }else{
        return true
      }
    }
  }
}

// Read compensation coefficients, unique for each sensor.
func ReadCoeff(d string, a int) error {
	i2c0,err:=i2c.Open(&i2c.Devfs{Dev:d},a)
  defer i2c0.Close()

  if err!=nil{
    log.Println("I2C Error")
    return err
  }else{
		i2c0.ReadReg(BMP280_COEF_START,b24)
		buf:=bytes.NewReader(b24)
		err=binary.Read(buf, binary.LittleEndian, &Cal)
		return nil
	}
}


func getOversamplingRate(accuracy string) byte {
	switch accuracy {
	case "ULTRA_LOW":
		return byte(1)
	case "LOW":
		return byte(2)
	case "STANDARD":
		return byte(3)
	case "HIGH":
		return byte(4)
	case "ULTRA_HIGH":
		return byte(5)
	default:
		return byte(1)
	}
}

// Read Temprature from ADC.
func ReadUncompTemprature(d string, a int, accuracy string) (int32, error) {
	var power byte = 1 // Forced mode
	osr := getOversamplingRate(accuracy)
	i2c0,err:=i2c.Open(&i2c.Devfs{Dev:d},a)
  defer i2c0.Close()

  if err!=nil{
    log.Println("I2C BUS Error")
    return 0,err
  }else{
		err = i2c0.WriteReg(BMP280_CNTR_MEAS_REG, []byte{power|(osr<<5)})
		if err != nil {
			log.Println("0:I2C WRITE error")
			return 0, err
		}
		//// Wait for measurement finished
		for n:=0;n<30;n++{
			err=i2c0.ReadReg(BMP280_STATUS_REG,b1)
			if err!=nil{
				log.Println("I2C READ error")
				return 0,err
			}else{
				b1[0] = b1[0] & 0x8
				if b1[0] == byte(0){
					log.Println("Not Busy")
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
		}
		////
		err = i2c0.ReadReg(BMP280_TEMP_3BYTES, b3)
		if err != nil {
			return 0, err
		}
		ut := int32(b3[0])<<12 + int32(b3[1])<<4 + int32(b3[2]&0xF0)>>4
		return ut, nil
	}
}


// Read Pressure from ADC.
func ReadUncompPressure(d string, a int, accuracy string) (int32, error) {
	var power byte = 1 // Forced mode
	osr := getOversamplingRate(accuracy)
	i2c0,err:=i2c.Open(&i2c.Devfs{Dev:d},a)
  defer i2c0.Close()

  if err!=nil{
    log.Println("I2C BUS Error")
    return 0,err
  }else{

		err = i2c0.WriteReg(BMP280_CNTR_MEAS_REG, []byte{(power|(osr<<2))})
		if err != nil {
			log.Println("1:I2C WRITE error")
			return 0, err
		}
		//// Wait for measurement finished
		for n:=0;n<30;n++{
			err=i2c0.ReadReg(BMP280_STATUS_REG,b1)
			if err!=nil{
				log.Println("I2C READ error")
				return 0,err
			}else{
				b1[0] = b1[0] & 0x8
				//log.Println("Busy flag=", b1[0])
				if b1[0] == 0{
					log.Println("Not Busy")
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
		}
		////
		err = i2c0.ReadReg(BMP280_PRES_3BYTES, b3)
		if err != nil {
			return 0, err
		}
		up := int32(b3[0])<<12 + int32(b3[1])<<4 + int32(b3[2]&0xF0)>>4
		return up, nil
	}
}


// Read Pressure Multple by 10 in unit Pa .
func ReadPressurePa(d string, a int, accuracy string) (uint32, error) {
	adc_T, err :=ReadUncompTemprature(d, a, accuracy)
	if err != nil {
		return 0, err
	}
	adc_P, err := ReadUncompPressure(d, a, accuracy)
	if err != nil {
		return 0, err
	}
	log.Println("T_ADC=",adc_T,", P_ADC=", adc_P)

	err = ReadCoeff(d,a)
	if err != nil {
		return 0, err
	}
	log.Println("Calibration Coeff=",Cal)
	var var1, var2, T, t_fine int32
	var p uint32
	var1 = ((((adc_T>>3)-(int32(Cal.Dig_T1)<<1))) * (int32(Cal.Dig_T2))) >> 11
	var2 = (((((adc_T>>4)-(int32(Cal.Dig_T1))) * ((adc_T>>4)-(int32(Cal.Dig_T1)))) >> 12) * (int32(Cal.Dig_T3))) >> 14
	t_fine = var1 + var2
	T = (t_fine * 5 + 128) >> 8
	log.Println("T (fine)=",float32(T)/100)

	var1 = ((int32(t_fine))>>1)-64000
	var2 = (((var1>>2) * (var1>>2)) >> 11 ) * (int32(Cal.Dig_P6))
	var2 = var2 + ((var1*(int32(Cal.Dig_P5)))<<1)
	var2 = (var2>>2)+((int32(Cal.Dig_P4))<<16)
	var1 = (((int32(Cal.Dig_P3) * (((var1>>2) * (var1>>2)) >> 13 )) >> 3) + (((int32(Cal.Dig_P2)) * var1)>>1))>>18
	var1 =((((32768+var1))*(int32(Cal.Dig_P1)))>>15)
	if var1 == 0	{
		return 0,nil // avoid exception caused by division by zero
	}
	// p = ( ( (BMP280_U32_t)  (  ((BMP280_S32_t)1048576)-adc_P ) - (var2>>12) )  )*3125;
	p = uint32(3125) * uint32( (1048576-adc_P)-(var2>>12) )
	if p < 0x80000000	{
		p = (p << 1) / (uint32(var1))
	}else{
		p = (p / uint32(var1)) * 2
	}
	var1 = ((int32(Cal.Dig_P9)) * (int32((((p>>3)) * (p>>3))>>13)))>>12
	var2 = ((int32((p>>2))) * (int32(Cal.Dig_P8)))>>13
	p = uint32(int32(p) + ((var1 + var2 + int32(Cal.Dig_P7)) >> 4))
	log.Println("Pressure=",p)
	return p, nil
}
