package main

import (
	i2c "github.com/d2r2/go-i2c"
	logger "github.com/d2r2/go-logger"
	mpl3115a2 "github.com/d2r2/go-mpl3115a2"
)

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

func main() {
	defer logger.FinalizeLogger()
	// Create new connection to i2c-bus on 1 line with address 0x60.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(0x60, 0)
	if err != nil {
		lg.Fatal(err)
	}
	defer i2c.Close()

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** !!! READ THIS !!!")
	lg.Notify("*** You can change verbosity of output, by modifying logging level of modules \"i2c\", \"mpl3115a2\".")
	lg.Notify("*** Uncomment/comment corresponding lines with call to ChangePackageLogLevel(...)")
	lg.Notify("*** !!! READ THIS !!!")
	lg.Notify("**********************************************************************************************")
	// Uncomment/comment next line to suppress/increase verbosity of output
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
	// logger.ChangePackageLogLevel("mpl3115a2", logger.InfoLevel)

	sensor := mpl3115a2.NewMPL3115A2()
	// Reset sensor
	err = sensor.Reset(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	// time.Sleep(50 * time.Millisecond)

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Measure pressure, altitude and temperature")
	lg.Notify("**********************************************************************************************")
	// Oversample Ratio - define precision, from low(0) to high(7)
	osr := 3
	p, t, err := sensor.MeasurePressure(i2c, osr)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Pressure = %v Pa, temperature = %v *C", p, t)

	p, t, err = sensor.MeasureAltitude(i2c, osr)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Altitude = %v m, temperature = %v *C", p, t)

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Calibrate sensor by shifting sea level, pressure, altitude")
	lg.Notify("**********************************************************************************************")
	lg.Infof("Shift pressure for %v Pa", 10)
	err = sensor.CompensatePressure(i2c, 10)
	if err != nil {
		lg.Fatal(err)
	}
	p, t, err = sensor.MeasurePressure(i2c, osr)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Pressure = %v Pa, temperature = %v *C", p, t)

	lg.Infof("Change sea level pressure to %v Pa, where default is %v Pa",
		90000, sensor.GetDefaultSeaLevelPressure())
	err = sensor.ModifySeaLevelPressure(i2c, 90000)
	if err != nil {
		lg.Fatal(err)
	}
	p, t, err = sensor.MeasureAltitude(i2c, osr)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Altitude = %v m, temperature = %v *C", p, t)

	lg.Infof("Shift altitude for %v m", -50)
	err = sensor.CompensateAltitude(i2c, -50)
	if err != nil {
		lg.Fatal(err)
	}
	p, t, err = sensor.MeasureAltitude(i2c, osr)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Altitude = %v m, temperature = %v *C", p, t)
}
