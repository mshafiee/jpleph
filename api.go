// ./api.go

/*
Package jpleph provides functions for accessing JPL planetary and lunar ephemerides.

This package is a Go translation of the original C code, which in turn was a translation of Fortran-77 code.
It allows users to read and interpolate data from JPL ephemeris files, such as DE405, DE430, etc.,
to obtain positions and velocities of solar system bodies at specified times.

Key Features:
  - Initialization and closing of ephemeris files.
  - Calculation of position and velocity vectors for planets and other celestial bodies.
  - Access to ephemeris metadata and constants.
  - Go-friendly types for planets, center bodies, and value types.
  - Robust error handling with standard Go error types.

Usage:
To use this package, you need a JPL ephemeris file (e.g., de405.bin).

 1. Initialize the ephemeris data using NewEphemeris:
    ```go
    ephem, err := jpleph.NewEphemeris("de405.bin", true) // Load constants
    if err != nil {
        log.Fatal(err)
    }
    defer ephem.Close()
    ```

 2. Calculate position and velocity:
    ```go
    et := 2451545.0 // Julian Ephemeris Date for J2000.0
    pos, vel, err := ephem.CalculatePV(et, jpleph.Mars, jpleph.Sun, true)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Position of Mars relative to Sun at J2000.0:\n")
    fmt.Printf("X: %f AU, Y: %f AU, Z: %f AU\n", pos.X, pos.Y, pos.Z)
    fmt.Printf("Velocity of Mars relative to Sun at J2000.0:\n")
    fmt.Printf("DX: %f AU/day, DY: %f AU/day, DZ: %f AU/day\n", vel.DX, vel.DY, vel.DZ)
    ```

 3. Access ephemeris information:
    ```go
    startDate := ephem.GetEphemerisDouble(jpleph.EphemerisStartJD)
    endDate := ephem.GetEphemerisDouble(jpleph.EphemerisEndJD)
    fmt.Printf("Ephemeris time range: JD %.1f to JD %.1f\n", startDate, endDate)
    ```

 4. Access constants (if loaded during initialization):
    ```go
    constantName, err := ephem.GetConstantName(0)
    if err != nil {
        log.Println(err)
    } else {
        constantValue, err := ephem.GetConstantValue(0)
        if err != nil {
            log.Println(err)
        } else {
            fmt.Printf("Constant %s: %f\n", constantName, constantValue)
        }
    }
    ```

License:
This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA
02110-1301, USA.

Authorship:
Mohammad Shafiee authored this Go code as a translation of the original C code.
The C version was a translation of Fortran-77 code originally written by
Piotr A. Dybczynski and later revised by Bill J Gray.
*/

// Package jpleph provides functions for accessing JPL planetary and lunar ephemerides.
package jpleph

import (
	"bytes"
	"errors"
	"fmt"
)

// ErrQuantityNotInEphemeris is returned when the requested quantity is not available in the ephemeris file.
var ErrQuantityNotInEphemeris = errors.New("quantity not available in ephemeris file")

// ErrInvalidIndex is returned when an invalid target or center body index is used.
var ErrInvalidIndex = errors.New("invalid target or center body index")

// ErrOutsideRange is returned when the requested time is outside the ephemeris time range.
var ErrOutsideRange = errors.New("requested time is outside ephemeris time range")

// ErrFileSeek is returned when there is an error seeking in the ephemeris file.
var ErrFileSeek = errors.New("error seeking in ephemeris file")

// ErrFileRead is returned when there is an error reading from the ephemeris file.
var ErrFileRead = errors.New("error reading from ephemeris file")

// ErrInitialization is returned when the ephemeris initialization fails. It wraps more specific initialization errors.
var ErrInitialization = errors.New("ephemeris initialization error") // For wrapping InitErrorCode

// ErrConstantNotFound is returned when a requested constant is not found in the ephemeris data.
var ErrConstantNotFound = errors.New("constant not found")

// Planet represents the celestial bodies available as targets in the ephemeris.
type Planet int

const (
	// Mercury represents the planet Mercury.
	Mercury Planet = 1
	// Venus represents the planet Venus.
	Venus Planet = 2
	// Earth represents the planet Earth.
	Earth Planet = 3
	// Mars represents the planet Mars.
	Mars Planet = 4
	// Jupiter represents the planet Jupiter.
	Jupiter Planet = 5
	// Saturn represents the planet Saturn.
	Saturn Planet = 6
	// Uranus represents the planet Uranus.
	Uranus Planet = 7
	// Neptune represents the planet Neptune.
	Neptune Planet = 8
	// Pluto represents the dwarf planet Pluto.
	Pluto Planet = 9
	// Moon represents the Earth's Moon.
	Moon Planet = 10
	// Sun represents the Sun.
	Sun Planet = 11
	// SolarSystemBarycenter represents the Solar System Barycenter.
	SolarSystemBarycenter Planet = 12
	// EarthMoonBarycenter represents the Earth-Moon Barycenter.
	EarthMoonBarycenter Planet = 13
	// Nutations represents nutations (used for high-precision calculations).
	Nutations Planet = 14
	// Librations represents lunar librations (used for high-precision lunar calculations).
	Librations Planet = 15
	// LunarMantleOmega represents Lunar Mantle Omega (used for lunar orientation).
	LunarMantleOmega Planet = 16
	// TT_TDB represents the time conversion factor between Terrestrial Time (TT) and Barycentric Dynamical Time (TDB).
	TT_TDB Planet = 17
)

// CenterBody represents the celestial bodies that can be used as the center of motion for calculations.
type CenterBody int

const (
	// CenterMercury represents Mercury as the center body.
	CenterMercury CenterBody = 1
	// CenterVenus represents Venus as the center body.
	CenterVenus CenterBody = 2
	// CenterEarth represents Earth as the center body.
	CenterEarth CenterBody = 3
	// CenterMars represents Mars as the center body.
	CenterMars CenterBody = 4
	// CenterJupiter represents Jupiter as the center body.
	CenterJupiter CenterBody = 5
	// CenterSaturn represents Saturn as the center body.
	CenterSaturn CenterBody = 6
	// CenterUranus represents Uranus as the center body.
	CenterUranus CenterBody = 7
	// CenterNeptune represents Neptune as the center body.
	CenterNeptune CenterBody = 8
	// CenterPluto represents Pluto as the center body.
	CenterPluto CenterBody = 9
	// CenterMoon represents the Moon as the center body.
	CenterMoon CenterBody = 10
	// CenterSun represents the Sun as the center body.
	CenterSun CenterBody = 11
	// CenterSolarSystemBarycenter represents the Solar System Barycenter as the center body.
	CenterSolarSystemBarycenter CenterBody = 12
	// CenterEarthMoonBarycenter represents the Earth-Moon Barycenter as the center body.
	CenterEarthMoonBarycenter CenterBody = 13
)

// ValueType represents the type of value to retrieve from ephemeris data using GetDouble or GetLong.
type ValueType int

const (
	// EphemerisStartJD represents the Julian Date of the start of the ephemeris time range.
	EphemerisStartJD ValueType = JPL_EPHEM_START_JD
	// EphemerisEndJD represents the Julian Date of the end of the ephemeris time range.
	EphemerisEndJD ValueType = JPL_EPHEM_END_JD
	// EphemerisStep represents the time step (in days) used in the ephemeris data.
	EphemerisStep ValueType = JPL_EPHEM_STEP
	// AUinKM represents the number of kilometers in one Astronomical Unit (AU).
	AUinKM ValueType = JPL_EPHEM_AU_IN_KM
	// EarthMoonMassRatio represents the mass ratio of the Earth to the Moon.
	EarthMoonMassRatio ValueType = JPL_EPHEM_EARTH_MOON_RATIO
	// NumberOfConstants represents the total number of constants available in the ephemeris.
	NumberOfConstants ValueType = JPL_EPHEM_N_CONSTANTS
	// EphemerisVersion represents the version number of the ephemeris (e.g., 405 for DE405).
	EphemerisVersion ValueType = JPL_EPHEM_EPHEMERIS_VERSION
	// KernelSize represents the total size of the ephemeris kernel file in bytes.
	KernelSize ValueType = JPL_EPHEM_KERNEL_SIZE
	// KernelRecordSize represents the size of each data record in the ephemeris kernel file in bytes.
	KernelRecordSize ValueType = JPL_EPHEM_KERNEL_RECORD_SIZE
	// KernelNCoeff represents the number of coefficients used in each Chebyshev polynomial record.
	KernelNCoeff ValueType = JPL_EPHEM_KERNEL_NCOEFF
	// KernelSwapBytes indicates whether byte swapping is needed for this kernel (1 for swap, 0 for no swap).
	KernelSwapBytes ValueType = JPL_EPHEM_KERNEL_SWAP_BYTES
	// IPTArrayOffset represents the offset to access the IPT array elements. Use IPTArrayOffset + index to access individual IPT array elements.
	IPTArrayOffset ValueType = JPL_EPHEM_IPT_ARRAY // Use IPTArrayOffset + index to access IPT array elements
)

// Position represents a 3D position vector in Astronomical Units (AU).
type Position struct {
	// X is the X component of the position in AU.
	X float64 // X component in AU
	// Y is the Y component of the position in AU.
	Y float64 // Y component in AU
	// Z is the Z component of the position in AU.
	Z float64 // Z component in AU
}

// Velocity represents a 3D velocity vector in Astronomical Units per day (AU/day).
type Velocity struct {
	// DX is the X component of the velocity in AU/day.
	DX float64 // DX component in AU/day
	// DY is the Y component of the velocity in AU/day.
	DY float64 // DY component in AU/day
	// DZ is the Z component of the velocity in AU/day.
	DZ float64 // DZ component in AU/day
}

// Ephemeris is a wrapper struct holding the ephemeris data interface and optional caches for constants.
// It provides methods to access ephemeris data and perform calculations.
type Ephemeris struct {
	ephemData   *jplEphData // Holds the underlying jplEphData directly
	constNames  [][]byte    // Cache for constant names (optional)
	constValues []float64   // Cache for constant values (optional)
}

// newEphemeris creates a new Ephemeris instance from a jplEphData interface.
// This is an internal constructor and should not be used directly.
// Use NewEphemeris to initialize an Ephemeris instance from a file.
func newEphemeris(data *jplEphData) *Ephemeris {
	return &Ephemeris{ephemData: data}
}

// NewEphemeris initializes the JPL ephemeris data from a binary ephemeris file and returns an Ephemeris wrapper.
// It opens the specified ephemeris file, reads necessary header information, and prepares the data for calculations.
// Optionally, it can load and cache constant names and values if `loadConstants` is true.
//
// Parameters:
//   - ephemerisFilename: Path to the binary ephemeris file (e.g., "de405.bin").
//   - loadConstants: Boolean flag to indicate whether to load and cache constant names and values.
//     Setting this to true can improve performance for repeated access to constants.
//
// Returns:
//   - *Ephemeris: Pointer to the initialized Ephemeris wrapper on success, nil on failure.
//   - error: Standard Go error if initialization fails. The error can be checked using errors.Is for specific error types
//     like ErrFileRead, ErrFileSeek, ErrInitialization.
func NewEphemeris(ephemerisFilename string, loadConstants bool) (*Ephemeris, error) {
	setDebugFlag(false)                                          // Disable debug flag by default
	ephemData, err := initEphemeris(ephemerisFilename, nil, nil) // Initialize ephemeris data
	if err != nil {
		return nil, fmt.Errorf("initialization failed: %w", err)
	}

	ephemWrapper := newEphemeris(ephemData) // Create Ephemeris wrapper
	if loadConstants {                      // Load constants if requested
		numConstants := ephemWrapper.GetEphemerisLong(NumberOfConstants)
		if numConstants <= 0 {
			return nil, fmt.Errorf("initialization failed: invalid number of constants: %d", numConstants)
		}
		ephemWrapper.constNames = make([][]byte, numConstants)   // Initialize slice for constant names
		ephemWrapper.constValues = make([]float64, numConstants) // Initialize slice for constant values
		for i := 0; i < int(numConstants); i++ {
			nameBuf := make([]byte, 7) // Buffer to read constant name
			value := getConstant(i, ephemData, nameBuf)
			ephemWrapper.constValues[i] = value
			ephemWrapper.constNames[i] = bytes.TrimRight(nameBuf[:6], "\x00") // Store name without null terminator
		}
	}
	return ephemWrapper, nil
}

// Close closes the ephemeris file associated with the Ephemeris data.
// It releases resources and ensures that the ephemeris file is properly closed.
// It is important to call Close when you are finished using the Ephemeris to avoid resource leaks.
//
// Returns:
//   - error: nil on success, or an error if closing the file fails.
func (e *Ephemeris) Close() error {
	return closeEphemeris(e.ephemData)
}

// CalculatePV calculates the position and optionally velocity of a target Planet relative to a CenterBody at a given time.
// The time is specified as Julian Ephemeris Date (JED).
// The function returns the position and velocity vectors in Astronomical Units (AU) and AU/day, respectively.
//
// Parameters:
//   - et: Julian Ephemeris Date (JED) at which to interpolate.
//   - target: Target Planet for which to calculate position and velocity. Use Planet constants (e.g., jpleph.Mars).
//   - center: Center CenterBody relative to which the position and velocity are calculated. Use CenterBody constants (e.g., jpleph.Sun).
//   - calcVelocity: Flag to indicate whether to calculate velocities. Set to true to calculate velocities, false for positions only.
//
// Returns:
//   - Position: Calculated position vector.
//   - Velocity: Calculated velocity vector (will be a zero vector if calcVelocity is false).
//   - error: nil on success, or a standard Go error if the underlying Pleph function returns an error code.
//     The error can be checked using errors.Is() to determine the specific error type, such as:
//     ErrQuantityNotInEphemeris, ErrInvalidIndex, ErrOutsideRange, ErrFileSeek, ErrFileRead.
func (e *Ephemeris) CalculatePV(et float64, target Planet, center CenterBody, calcVelocity bool) (Position, Velocity, error) {
	velFlag := 0
	if calcVelocity {
		velFlag = 2
	}
	rrd, err := Pleph(e.ephemData, et, int(target), int(center), velFlag)
	if err != nil {
		return Position{}, Velocity{}, err
	}
	pos := Position{X: rrd[0], Y: rrd[1], Z: rrd[2]}
	vel := Velocity{} // Initialize to zero velocity in case velocity is not calculated
	if calcVelocity {
		vel = Velocity{DX: rrd[3], DY: rrd[4], DZ: rrd[5]}
	}

	return pos, vel, nil
}

// GetEphemerisDouble retrieves a double-precision (float64) value from the ephemeris data structure.
// This function is used to access metadata and parameters stored in the ephemeris file as double-precision numbers.
//
// Parameters:
//   - valueType: ValueType specifying which parameter to retrieve. Use ValueType constants (e.g., jpleph.EphemerisStartJD).
//
// Returns:
//   - float64: The requested double-precision value. Returns -1 if the ValueType is invalid or an error occurs.
func (e *Ephemeris) GetEphemerisDouble(valueType ValueType) float64 {
	return GetDouble(e.ephemData, int(valueType))
}

// GetEphemerisLong retrieves an integer (int64) value from the ephemeris data structure.
// This function is used to access metadata and parameters stored in the ephemeris file as integer numbers.
//
// Parameters:
//   - valueType: ValueType specifying which parameter to retrieve. Use ValueType constants (e.g., jpleph.NumberOfConstants).
//
// Returns:
//   - int64: The requested integer (int64) value. Returns -1 if the ValueType is invalid or an error occurs.
func (e *Ephemeris) GetEphemerisLong(valueType ValueType) int64 {
	return GetLong(e.ephemData, int(valueType))
}

// GetIPTArrayValue retrieves a value from the IPT (Interpolation Parameter Table) array at the given index.
// The IPT array contains metadata about the Chebyshev polynomial interpolation scheme used in the ephemeris.
// Valid indices are in the range 0-44.
//
// Parameters:
//   - index: Index of the IPT array element to retrieve (0-44).
//
// Returns:
//   - int64: The requested IPT array value. Returns -1 if the index is invalid (out of range).
func (e *Ephemeris) GetIPTArrayValue(index int) int64 {
	if index < 0 || index > 44 {
		return -1 // Invalid index
	}
	return GetLong(e.ephemData, JPL_EPHEM_IPT_ARRAY+index)
}

// GetEphemName returns the name of the ephemeris file as stored in the kernel.
// This name typically includes the ephemeris series (e.g., DE405) and the date range.
//
// Returns:
//   - string: The name of the ephemeris.
func (e *Ephemeris) GetEphemName() string {
	return getEphemName(e.ephemData)
}

// GetConstantName retrieves the name of a constant at the given index from the ephemeris data.
// Constant names are typically short strings (e.g., "GM_Sun", "AU").
// This function only works correctly if constants were loaded during initialization (NewEphemeris with loadConstants=true).
//
// Parameters:
//   - index: Index of the constant (0-based). The valid range depends on the ephemeris file.
//
// Returns:
//   - string: Constant name.
//   - error: ErrConstantNotFound if the index is out of range or if constant name retrieval fails.
func (e *Ephemeris) GetConstantName(index int) (string, error) {
	if index < 0 || index >= len(e.constNames) {
		return "", fmt.Errorf("get constant name failed: %w: index %d out of range", ErrConstantNotFound, index)
	}
	return string(e.constNames[index]), nil
}

// GetConstantValue retrieves the value of a constant at the given index from the ephemeris data.
// Constant values are typically physical constants used in ephemeris calculations.
// This function only works correctly if constants were loaded during initialization (NewEphemeris with loadConstants=true).
//
// Parameters:
//   - index: Index of the constant (0-based). The valid range depends on the ephemeris file.
//
// Returns:
//   - float64: Constant value.
//   - error: ErrConstantNotFound if the index is out of range or if constant value retrieval fails.
func (e *Ephemeris) GetConstantValue(index int) (float64, error) {
	if index < 0 || index >= len(e.constValues) {
		return 0.0, fmt.Errorf("get constant value failed: %w: index %d out of range", ErrConstantNotFound, index)
	}
	return e.constValues[index], nil
}
