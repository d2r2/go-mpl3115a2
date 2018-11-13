NXP Semiconductors MPL3115A2 pressure and temperature sensor
============================================================

[![Build Status](https://travis-ci.org/d2r2/go-mpl3115a2.svg?branch=master)](https://travis-ci.org/d2r2/go-mpl3115a2)
[![Go Report Card](https://goreportcard.com/badge/github.com/d2r2/go-mpl3115a2)](https://goreportcard.com/report/github.com/d2r2/go-mpl3115a2)
[![GoDoc](https://godoc.org/github.com/d2r2/go-mpl3115a2?status.svg)](https://godoc.org/github.com/d2r2/go-mpl3115a2)
[![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

MPL3115A2 ([pdf reference](https://raw.github.com/d2r2/go-mpl3115a2/master/docs/mpl3115a2.pdf)) is a popular sensor among Arduino and Raspberry PI developers.
Sensor is a compact, piezoresistive, pressure and temperature sensor with an I2C digital interface:
![image](https://raw.github.com/d2r2/go-mpl3115a2/master/docs/mpl3115a2.jpg)

Here is a library written in [Go programming language](https://golang.org/) for Raspberry PI and counterparts, which gives you in the output temperature, atmospheric pressure and altitude values (making all necessary i2c-bus interacting and values computing).

Golang usage
------------


```go
func main() {
	// Create new connection to i2c-bus on 0 line with address 0x60.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(0x60, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer i2c.Close()

	sensor := mpl3115a2.NewMPL3115A2()

	// Oversample Ratio - define precision, from low(0) to high(7)
	osr := 3
	p, t, err := sensor.MeasurePressure(i2c, osr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Pressure = %v Pa, temperature = %v *C", p, t)

	p, t, err = sensor.MeasureAltitude(i2c, osr)
	if err != nil {
		lg.Fatal(err)
	}
	log.Printf("Altitude = %v m, temperature = %v *C", p, t)
}
```


Getting help
------------

GoDoc [documentation](http://godoc.org/github.com/d2r2/go-mpl3115a2)

Installation
------------

```bash
$ go get -u github.com/d2r2/go-mpl3115a2
```

Troubleshooting
--------------

- *How to obtain fresh Golang installation to RPi device (either any RPi clone):*
If your RaspberryPI golang installation taken by default from repository is outdated, you may consider
to install actual golang manually from official Golang [site](https://golang.org/dl/). Download
tar.gz file containing armv6l in the name. Follow installation instructions.

- *How to enable I2C bus on RPi device:*
If you employ RaspberryPI, use raspi-config utility to activate i2c-bus on the OS level.
Go to "Interfacing Options" menu, to active I2C bus.
Probably you will need to reboot to load i2c kernel module.
Finally you should have device like /dev/i2c-1 present in the system.

- *How to find I2C bus allocation and device address:*
Use i2cdetect utility in format "i2cdetect -y X", where X may vary from 0 to 5 or more,
to discover address occupied by peripheral device. To install utility you should run
`apt install i2c-tools` on debian-kind system. `i2cdetect -y 1` sample output:
	```
	     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
	00:          -- -- -- -- -- -- -- -- -- -- -- -- --
	10: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	20: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	30: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	40: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	50: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	60: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	70: -- -- -- -- -- -- 76 --    
	```

Contact
-------

Please use [Github issue tracker](https://github.com/d2r2/go-mpl3115a2/issues) for filing bugs or feature requests.


License
-------

Go-mpl3115a2 is licensed under MIT License.
