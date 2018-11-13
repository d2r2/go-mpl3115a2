//--------------------------------------------------------------------------------------------------
//
// Copyright (c) 2018 Denis Dyakov
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

package mpl3115a2

import (
	"encoding/binary"
	"errors"
	"time"

	i2c "github.com/d2r2/go-i2c"
)

// Register map
const (
	// Alias for DR_STATUS or F_STATUS
	STATUS = 0x00

	// 20-bit realtime pressure sample
	OUT_PRES_MSB_CSB_LSB = 0x01
	OUT_PRES_BYTES       = 3

	// 12-bit realtime temperature sample
	OUT_TEMP_MSB_LSB = 0x04
	OUT_TEMP_BYTES   = 2

	// Data ready status information
	DR_STATUS = 0x06

	// 20-bit pressure change data
	OUT_PRES_DELTA_MSB_CSB_LSB = 0x07
	OUT_PRES_DELTA_BYTES       = 3

	// 12-bit temperature change data
	OUT_TEMP_DELTA_MSB_LSB = 0x0A
	OUT_TEMP_DELTA_BYTES   = 2

	// Fixed device ID number
	WHO_AM_I = 0x0C

	// FIFO status: no FIFO event detected
	F_STATUS = 0x0D

	// FIFO 8-bit data access
	F_DATA = 0x0E

	// FIFO setup
	F_SETUP = 0x0F

	// Time since FIFO overflow
	TIME_DLY = 0x10

	// Current system mode
	SYSMOD = 0x11

	// Interrupt status
	INT_SOURCE = 0x12

	// Data event flag configuration
	PT_DATA_CFG = 0x13

	// Barometric input for altitude calculation
	BAR_IN_MSB_LSB = 0x14
	BAR_IN_BYTES   = 2

	// Pressure/altitude target
	PRES_TGT_MSB_LSB = 0x16
	PRES_TGT_BYTES   = 2

	// Temperature target value
	T_TGT = 0x18

	// Pressure/altitude window
	PRES_WND_MSB_LSB = 0x19
	PRES_WND_BYTES   = 2

	// Temperature window
	TEMP_WND = 0x1B

	// Minimum pressure/altitude
	PRES_MIN_MSB_CSB_LSB = 0x1C
	PRES_MIN_BYTES       = 3

	// Minimum temperature
	TEMP_MIN_MSB_LSB = 0x1E
	TEMP_MIN_BYTES   = 2

	// Maximum pressure/altitude
	PRES_MAX_MSB_CSB_LSB = 0x21
	PRES_MAX_BYTES       = 3

	// Maximum temperature
	TEMP_MAX_MSB_LSB = 0x24
	TEMP_MAX_BYTES   = 2

	// Control register: Modes, oversampling
	CTRL_REG1 = 0x26

	// Control register: Acquisition time step
	CTRL_REG2 = 0x27

	// Control register: Interrupt pin configuration
	CTRL_REG3 = 0x28

	// Control register: Interrupt enables
	CTRL_REG4 = 0x29

	// Control register: Interrupt output pin assignment
	CTRL_REG5 = 0x2A

	// Pressure data offset
	OFF_PRES = 0x2B

	// Temperature data offset
	OFF_TEMP = 0x2C

	// Altitude data offset
	OFF_H = 0x2D
)

// Flag can keep any sensor register specific bit flags.
type Flag byte

const (
	// DR_STATUS flag: Pressure/altitude or temperature data ready.
	PRES_TEMP_DATA_READY Flag = 0x8
	// DR_STATUS flag: Pressure/altitude new data available.
	PRES_DATA_READY Flag = 0x4
	// DR_STATUS flag: Temperature new data available.
	TEMP_DATA_READY Flag = 0x2
)

// PressureType signify which type of
// pressure measurement in use.
type PressureType int

const (
	// Measure pressure in Pa
	Barometer PressureType = iota + 1
	// Measure altitude in m
	Altimeter
)

// RawPressure keeps raw pressure data received from sensor.
type RawPressure struct {
	PRES_MSB byte
	PRES_CSB byte
	PRES_LSB byte
}

// ConvertToSignedQ16Dot4 convert raw data to signed Q16.4,
// where integer and fraction parts returned in separate fields.
// Used for altimeter mode.
func (v *RawPressure) ConvertToSignedQ16Dot4() (int16, uint8) {
	presFrac := (v.PRES_LSB & 0xF0) >> 4
	presInt := int16((uint16(v.PRES_MSB) << 8) | uint16(v.PRES_CSB))
	return presInt, presFrac
}

// ConvertToUnsignedQ18Dot2 convert raw data to unsigned Q18.2,
// where integer and fraction parts returned in separate fields.
// Used for barometer mode.
func (v *RawPressure) ConvertToUnsignedQ18Dot2() (uint32, uint8) {
	presFrac := (v.PRES_LSB & 0xF0) >> 4
	presInt := (uint32(v.PRES_MSB) << 10) | (uint32(v.PRES_CSB) << 2) | uint32(presFrac>>2)
	presFrac &= 0x3
	return presInt, presFrac
}

// RawTemperature keeps raw temperature data received from sensor.
type RawTemperature struct {
	TEMP_MSB byte
	TEMP_LSB byte
}

// ConvertToSignedQ8Dot4 convert raw data to signed Q8.4,
// where integer and fraction parts returned in separate fields.
func (v *RawTemperature) ConvertToSignedQ8Dot4() (int8, uint8) {
	tempFrac := (v.TEMP_LSB & 0xF0) >> 4
	tempInt := int8(v.TEMP_MSB)
	return tempInt, tempFrac
}

// MPL3115A2 keeps sensor itself.
type MPL3115A2 struct {
}

// NewMPL3115A2 return new sensor instance.
func NewMPL3115A2() *MPL3115A2 {
	v := &MPL3115A2{}
	return v
}

// Oversample ratio should in range [0..7], where final value equal to 2^osr.
func (v *MPL3115A2) encodeCtrlOverSampleRatio(oversample int) (byte, error) {
	if oversample < 0 || oversample > 7 {
		return 0, errors.New("oversample ratio should be in range [0..7]")
	}
	b := byte(oversample) << 3
	return b, nil
}

// Define measure for altimeter/barometer mode.
// Altimeter mode return "pressure" value in meters,
// barometer mode in Pascals.
func (v *MPL3115A2) encodeCtrlAltimeterMode(altimeterMode bool) (byte, error) {
	if altimeterMode {
		return 0x80, nil
	}
	return 0, nil
}

// Activate/deactivate reset bit.
func (v *MPL3115A2) encodeCtrlResetBit(activateReset bool) (byte, error) {
	if activateReset {
		return 0x4, nil
	}
	return 0, nil
}

// Put sensor in ACTIVE/STANDBY mode.
func (v *MPL3115A2) encodeCtrlActiveStatus(activateSensor bool) (byte, error) {
	if activateSensor {
		return 0x1, nil
	}
	return 0, nil
}

// Read STATUS register.
func (v *MPL3115A2) readStatusReg(i2c *i2c.I2C) (Flag, error) {
	status, err := i2c.ReadRegU8(STATUS)
	if err != nil {
		return 0, err
	}
	return Flag(status), nil
}

// Write CTRL_REG1 register.
func (v *MPL3115A2) writeCtrlReg1(i2c *i2c.I2C, value byte) error {
	err := i2c.WriteRegU8(CTRL_REG1, value)
	if err != nil {
		return err
	}
	return nil
}

// Write PT_DATE_CFG register to define events.
func (v *MPL3115A2) writeEventMode(i2c *i2c.I2C,
	temperatureEvent, preasureEvent bool) error {

	var flags byte
	if temperatureEvent {
		flags |= 0x1
	}
	if preasureEvent {
		flags |= 0x2
	}
	if temperatureEvent || preasureEvent {
		flags |= 0x4
	}
	err := i2c.WriteRegU8(PT_DATA_CFG, flags)
	if err != nil {
		return err
	}
	return nil
}

// MeasureAltitude measure altitude in meters with specific
// precision defined by oversample ratio.
func (v *MPL3115A2) MeasureAltitude(i2c *i2c.I2C, oversampleRatio int) (float32, float32, error) {
	up, ut, err := v.measureRaw(i2c, oversampleRatio, Altimeter)
	if err != nil {
		return 0, 0, err
	}
	presInt, presFrac := up.ConvertToSignedQ16Dot4()
	tempInt, tempFrac := ut.ConvertToSignedQ8Dot4()
	alt := float32(presInt) + float32(presFrac)/(1<<4)
	t := float32(tempInt) + float32(tempFrac)/(1<<4)
	return alt, t, nil
}

// MeasurePressure measure pressure in Pa with specific
// precision defined by oversample ratio.
func (v *MPL3115A2) MeasurePressure(i2c *i2c.I2C, oversampleRation int) (float32, float32, error) {
	up, ut, err := v.measureRaw(i2c, oversampleRation, Barometer)
	if err != nil {
		return 0, 0, err
	}
	presInt, presFrac := up.ConvertToUnsignedQ18Dot2()
	tempInt, tempFrac := ut.ConvertToSignedQ8Dot4()
	pres := float32(presInt) + float32(presFrac)/(1<<2)
	t := float32(tempInt) + float32(tempFrac)/(1<<4)
	return pres, t, nil
}

// Initialize sensor and made raw measurement
// to read uncompensated pressure and temperature.
func (v *MPL3115A2) measureRaw(i2c *i2c.I2C, overampleRatio int,
	pressureType PressureType) (*RawPressure, *RawTemperature, error) {

	lg.Debug("Measurement pressure and temperature...")

	// enable Altimeter mode
	var barometerType bool
	if pressureType == Altimeter {
		barometerType = true
	}
	flags, err := v.encodeCtrlAltimeterMode(barometerType)
	if err != nil {
		return nil, nil, err
	}
	// define Oversample Ratio to 2^oversampleRatio
	b, err := v.encodeCtrlOverSampleRatio(overampleRatio)
	if err != nil {
		return nil, nil, err
	}
	flags |= b
	// activate Altimeter mode and set Oversample Ratio
	err = v.writeCtrlReg1(i2c, flags)
	if err != nil {
		return nil, nil, err
	}
	// enable events for temperature and pressure
	err = v.writeEventMode(i2c, true, true)
	if err != nil {
		return nil, nil, err
	}
	// get activate sensor bit
	b, err = v.encodeCtrlActiveStatus(true)
	if err != nil {
		return nil, nil, err
	}
	flags |= b
	// activate sensor
	err = v.writeCtrlReg1(i2c, flags)
	if err != nil {
		return nil, nil, err
	}
	// read status until measurement is done
	for {
		var n time.Duration = 1
		// n = 1 << overampleRatio
		time.Sleep(time.Millisecond * 2 * n)
		status, err := v.readStatusReg(i2c)
		if err != nil {
			return nil, nil, err
		}
		if status&PRES_TEMP_DATA_READY != 0 {
			break
		}
	}
	up, ut, err := v.readRawPressureTemperature(i2c)
	if err != nil {
		return nil, nil, err
	}
	return up, ut, nil
}

// Read uncompensated temperature and pressure sensor measurement.
func (v *MPL3115A2) readRawPressureTemperature(i2c *i2c.I2C) (*RawPressure, *RawTemperature, error) {
	_, err := i2c.WriteBytes([]byte{STATUS})
	if err != nil {
		return nil, nil, err
	}
	var data struct {
		STATUS byte
		RawPressure
		RawTemperature
	}
	err = readDataToStruct(i2c, 1+OUT_PRES_BYTES+OUT_TEMP_BYTES, binary.LittleEndian, &data)
	if err != nil {
		return nil, nil, err
	}
	// lg.Debugf("Data = %+v", data)
	return &data.RawPressure, &data.RawTemperature, nil
}

// ModifySeaLevelPressure call allow to change default sea level value 101326 Pa to custom one.
func (v *MPL3115A2) ModifySeaLevelPressure(i2c *i2c.I2C, pressureAtSeeLevel uint32) error {
	// divide by 2
	pressureAtSeeLevel = pressureAtSeeLevel / 2
	b := []byte{BAR_IN_MSB_LSB, byte(pressureAtSeeLevel >> 8), byte(pressureAtSeeLevel & 0xFF)}
	_, err := i2c.WriteBytes(b)
	if err != nil {
		return err
	}
	return nil
}

// GetDefaultSeaLevelPressure return average barometric pressure
// on the sea level defined as 101325 Pa.
func (v *MPL3115A2) GetDefaultSeaLevelPressure() uint32 {
	return 101326
}

// Reset reboot sensor and initialize some sensor registers.
func (v *MPL3115A2) Reset(i2c *i2c.I2C) error {
	lg.Debug("Reset sensor...")

	flags, err := v.encodeCtrlResetBit(true)
	if err != nil {
		return err
	}
	// activate reset bit
	err = v.writeCtrlReg1(i2c, flags)
	// ignore error, since sensor terminates i2c-connection
	return nil
}

// CompensateAltitude shift altitude from -128 to +127 meters.
// Default value is 0. Can be used for sensor calibration.
func (v *MPL3115A2) CompensateAltitude(i2c *i2c.I2C, shiftM int8) error {
	b := []byte{OFF_H, byte(shiftM)}
	_, err := i2c.WriteBytes(b)
	if err != nil {
		return err
	}
	return nil

}

// CompensatePressure shift pressure from -512 to +508 Pascal.
// Default value is 0. Can be user for sensor calibration.
func (v *MPL3115A2) CompensatePressure(i2c *i2c.I2C, shiftPa int16) error {
	if shiftPa > 508 || shiftPa < -512 {
		return errors.New("pressure compensation exceed range [-512..+508]")
	}
	// divide by 4
	shiftPa = shiftPa / 4
	b := []byte{OFF_PRES, byte(shiftPa)}
	_, err := i2c.WriteBytes(b)
	if err != nil {
		return err
	}
	return nil

}

// CompensateTemperature shift temperature from -8 to +7.9375 *C.
// Default value is 0. Can be used for sensor calibration.
func (v *MPL3115A2) CompensateTemperature(i2c *i2c.I2C, shiftTemp float32) error {
	if shiftTemp > 7.9375 || shiftTemp < -8 {
		return errors.New("temperature compensation exceed range [-8..+7.9375]")
	}
	// multiply by 16
	shiftTemp = shiftTemp * 16
	b := []byte{OFF_PRES, byte(shiftTemp)}
	_, err := i2c.WriteBytes(b)
	if err != nil {
		return err
	}
	return nil

}
