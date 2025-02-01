// ./ephemeris.go
package jpleph

/*
Package jpleph provides functions for accessing JPL planetary and lunar ephemerides.

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

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

var debugFlag = false // Set to true to enable debug output

// GetDouble retrieves double-precision values from the ephemeris data structure.
// It takes an ephemeris interface and an integer value code as input.
// The value code specifies which parameter to retrieve (e.g., JPL_EPHEM_START_JD, JPL_EPHEM_AU_IN_KM).
// Returns the requested double-precision value. Returns -1 for invalid value codes.
func GetDouble(ephem *jplEphData, value int) float64 {
	var rval float64 = 0.0

	switch value {
	case JPL_EPHEM_START_JD: // Julian Ephemeris Date at start
		rval = ephem.ephemStart
	case JPL_EPHEM_END_JD: // Julian Ephemeris Date at end
		rval = ephem.ephemEnd
	case JPL_EPHEM_STEP: // Ephemeris time step (days)
		rval = ephem.ephemStep
	case JPL_EPHEM_AU_IN_KM: // Astronomical Unit in kilometers
		rval = ephem.au
	case JPL_EPHEM_EARTH_MOON_RATIO: // Earth-Moon mass ratio
		rval = ephem.emrat
	default:
		rval = -1 // Invalid value code
	}
	return rval
}

// GetLong retrieves integer (long) values from the ephemeris data structure.
// It takes an ephemeris interface and an integer value code as input.
// The value code specifies which parameter to retrieve (e.g., JPL_EPHEM_N_CONSTANTS, JPL_EPHEM_IPT_ARRAY).
// Returns the requested integer (int64) value. Returns -1 for invalid value codes or array indices.
func GetLong(ephem *jplEphData, value int) int64 {
	var rval int64

	switch value {
	case JPL_EPHEM_N_CONSTANTS: // Number of constants in ephemeris
		rval = int64(ephem.ncon)
	case JPL_EPHEM_EPHEMERIS_VERSION: // Ephemeris version (e.g., 405, 406)
		rval = int64(ephem.ephemerisVersion)
	case JPL_EPHEM_KERNEL_SIZE: // Size of the kernel in data units
		rval = int64(ephem.kernelSize)
	case JPL_EPHEM_KERNEL_RECORD_SIZE: // Size of a kernel record in bytes
		rval = int64(ephem.recsize)
	case JPL_EPHEM_KERNEL_NCOEFF: // Number of coefficients per record
		rval = int64(ephem.ncoeff)
	case JPL_EPHEM_KERNEL_SWAP_BYTES: // Flag indicating byte swapping is needed
		rval = int64(ephem.swapBytes)
	default:
		tval := value - JPL_EPHEM_IPT_ARRAY // Offset for IPT array access

		if tval >= 0 && tval < 45 { // IPT array indices range 0-44 (15x3)
			rval = int64(ephem.ipt[tval/3][tval%3]) // Access IPT array: ipt[row][column]
		} else {
			rval = -1 // Invalid IPT array index
			if rval == -1 {
				panic("Assertion failed: rval == -1 - Invalid JPL_EPHEM_IPT_ARRAY index") // Panic for assertion failure in Go
			}
		}
	}
	return rval
}

// Pleph calculates the position and velocity of a target body relative to a center body at a given time.
//
// Parameters:
//   - ephem:  ephemeris data.
//   - et: Julian Ephemeris Date (JED) at which to interpolate.
//   - ntarg: Target body index (1-17, see body numbering convention).
//   - ncent: Center body index (0-13, see body numbering convention).
//   - rrd: Output slice of 6 doubles to store position and velocity [x, y, z, dx, dy, dz] in AU and AU/day.
//   - calcVelocity: Flag (non-zero to calculate velocities, 0 for positions only).
//
// Body Numbering Convention:
//
//	1 = Mercury, 2 = Venus, 3 = Earth, 4 = Mars, 5 = Jupiter, 6 = Saturn, 7 = Uranus, 8 = Neptune,
//	9 = Pluto, 10 = Moon, 11 = Sun, 12 = Solar-system barycenter, 13 = Earth-moon barycenter,
//	14 = Nutations (longitude and obliquity), 15 = Librations, 16 = Lunar mantle omega_x,omega_y,omega_z,
//	17 = TT-TDB.
//
// Returns:
//   - 0 on success.
//   - JPL_EPH_QUANTITY_NOT_IN_EPHEMERIS if requested quantity (nutations, librations, TT-TDB) is not in the ephemeris file.
//   - JPL_EPH_INVALID_INDEX if target or center body index is invalid.
func Pleph(ephem *jplEphData, et float64, ntarg int, ncent int, calcVelocity int) ([]float64, error) {

	var pv [13][6]float64 // Position/velocity array for 13 bodies (0-12).
	// 0=Mercury, 1=Venus,..., 8=Pluto, 9=Moon, 10=Sun, 11=SSBary, 12=EMBary
	// First 10 elements (0-9) are filled by State(), all are adjusted here.

	listVal := 1 // Default to calculate positions only
	if calcVelocity != 0 {
		listVal = 2 // Calculate position and velocity if calcVelocity is non-zero
	}
	var i uint
	var list [14]int // List of bodies for which to calculate ephemeris values in State().
	// 0=Mercury, 1=Venus, 2=EMBary,..., 8=Pluto, 9=geocentric Moon, 10=nutations in
	// long. & obliq., 11= lunar librations, 12 = TT-TDB, 13=lunar mantle omegas

	// Initialize output array
	rrd := make([]float64, 6)

	if ntarg == ncent { // Relative position/velocity is zero if target and center are the same
		return rrd, nil
	}
	for i = 0; i < uint(len(list)); i++ {
		list[i] = 0
	}
	for i := 0; i < 4; i++ {
		if ntarg == int(i)+14 {
			if ephem.ipt[i+11][1] > 0 {
				list[i+10] = listVal
				err := State(ephem, et, list, &pv, rrd, 0)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, ErrQuantityNotInEphemeris
			}
			return rrd, nil
		}
	}
	if ntarg > 13 || ncent > 13 || ntarg < 1 || ncent < 1 {
		return nil, ErrInvalidIndex
	}

	// Prepare list for State call to get barycentric positions
	for i := 0; i < 2; i++ { // Iterate for target and center bodies
		k := uint((i*ncent + (1-i)*ntarg) - 1) // Calculate body index (0-based)

		if k <= 9 {
			list[k] = listVal // Major planets (Mercury to Pluto, Moon)
		}
		if k == 9 {
			list[2] = listVal // Moon requires Earth-Moon Barycenter state
		}
		if k == 2 {
			list[9] = listVal // Earth-Moon Barycenter requires Moon state
		}
		if k == 12 {
			list[2] = listVal // Earth-Moon Barycenter requires EMBary state (redundant, already set for Earth/Moon)
		}
	}

	// Call State to get barycentric positions and velocities
	// Handle Sun, Solar System Barycenter, and Earth-Moon Barycenter cases
	err := State(ephem, et, list, &pv, rrd, 1)
	if err != nil {
		return rrd, err
	}
	if ntarg == 11 || ncent == 11 {
		for i = 0; i < 6; i++ {
			pv[10][i] = ephem.pvsun[i] // Use pre-calculated Sun's state from State()
		}
	}

	if ntarg == 12 || ncent == 12 { // Solar System Barycenter is target or center
		for i = 0; i < 6; i++ {
			pv[11][i] = 0.0 // Solar System Barycenter position/velocity is defined as zero
		}
	}

	if ntarg == 13 || ncent == 13 { // Earth-Moon Barycenter is target or center
		for i = 0; i < 6; i++ {
			pv[12][i] = pv[2][i] // Earth-Moon Barycenter state is same as EMBary calculated by State()
		}
	}
	// Handle Earth-Moon and Moon-Earth cases for relative position
	if (ntarg*ncent) == 30 && (ntarg+ncent) == 13 { // Earth-Moon or Moon-Earth relative position
		for i = 0; i < 6; i++ {
			pv[2][i] = 0.0 // Earth's state is relative to Moon in this specific case (set to 0 for relative calculation)
		}
	} else {
		if list[2] != 0 { // Adjust Earth's state from EMBary to Earth-centric if needed
			for i = 0; i < uint(list[2]*3); i++ {
				pv[2][i] -= pv[9][i] / (1.0 + ephem.emrat) // Earth = EMBary - Moon/(1+emrat)
			}
		}

		if list[9] != 0 { // Calculate Moon's SSBary state if needed
			for i = 0; i < uint(list[9]*3); i++ {
				pv[9][i] += pv[2][i] // Moon = Moon(geocentric) + Earth(SSBary)
			}
		}
	}

	// Calculate relative position and velocity (target - center)
	for i = 0; i < uint(listVal*3); i++ {
		rrd[i] = pv[ntarg-1][i] - pv[ncent-1][i]
	}
	return rrd, nil
}

// interp interpolates Chebyshev coefficients to compute position, velocity, and optionally acceleration.
//
// Parameters:
//   - iinfo: Interpolation information struct to store/reuse Chebyshev polynomial values.
//   - coef: Slice of Chebyshev coefficients for position.
//   - t: Time parameters [fractional time in interval (0<=t<=1), interval length].
//   - ncf: Number of coefficients per component.
//   - ncm: Number of components per set of coefficients (e.g., 3 for x, y, z).
//   - na: Number of sets of coefficients in full array (number of sub-intervals).
//   - velocityFlag: Flag: 1=positions only, 2=pos and vel, 3=pos, vel, accel (for pvsun).
//   - posvel: Output slice to store interpolated quantities [position, velocity, acceleration (optional)].
func interp(iinfo *interpolationInfo, coef []float64, t [2]float64, ncf uint, ncm uint, na uint, velocityFlag int, posvel []float64) {
	if debugFlag {
		fmt.Println("interp: Entered")
		fmt.Printf("interp: t[0] = %f, t[1] = %f, ncf = %d, ncm = %d, na = %d, velocityFlag = %d\n", t[0], t[1], ncf, ncm, na, velocityFlag)
	}
	dna := float64(na) // Number of sub-intervals as float64
	temp := dna * t[0]
	intPart, fracPart := math.Modf(temp) // Integer and fractional parts of (na * t[0])
	l := uint(intPart)                   // Sub-interval index
	var vfac float64                     // Velocity scaling factor
	tc := 2.0*fracPart - 1.0             // Normalized time within sub-interval (-1 <= tc <= 1)
	var i, j uint

	if ncf >= maxCheby {
		panic("ncf must be less than maxCheby") // Panic if number of coefficients exceeds maxCheby
	}

	if l == na { // Handle edge case when t[0] is exactly 1.0
		l--
		tc = 1.0
	}

	if tc < -1.0 || tc > 1.0 {
		panic("tc must be between -1 and 1") // Panic if normalized time is out of bounds
	}

	// Recurrence relation for Chebyshev polynomials T_i(tc)
	if tc != iinfo.posnCoeff[1] { // Recompute Chebyshev polynomials if tc has changed
		iinfo.nPosnAvail = 2
		iinfo.nVelAvail = 2
		iinfo.posnCoeff[1] = tc
		iinfo.twot = tc + tc // 2*tc for efficiency in recurrence
		if debugFlag {
			fmt.Printf("interp: tc changed, iinfo.nPosnAvail = %d, iinfo.nVelAvail = %d, iinfo.posnCoeff[1] = %f, iinfo.twot = %f\n", iinfo.nPosnAvail, iinfo.nVelAvail, iinfo.posnCoeff[1], iinfo.twot)
		}
	}

	if iinfo.nPosnAvail < ncf { // Compute Chebyshev polynomials up to ncf if needed
		for i = 2; i < ncf; i++ {
			iinfo.posnCoeff[i] = iinfo.twot*iinfo.posnCoeff[i-1] - iinfo.posnCoeff[i-2] // T_{n+1} = 2tc*T_n - T_{n-1}
		}
		iinfo.nPosnAvail = ncf
		if debugFlag {
			fmt.Printf("interp: Updated iinfo.posnCoeff, iinfo.nPosnAvail = %d\n", iinfo.nPosnAvail)
		}
	}

	posvelIndex := 0
	for i = 0; i < ncm; i++ { // Interpolate position components
		coeffPtr := coef[ncf*(i+l*ncm):] // Pointer to coefficients for current component and sub-interval
		posvel[posvelIndex] = 0.0
		for j = 0; j < ncf; j++ {
			posvel[posvelIndex] += iinfo.posnCoeff[j] * coeffPtr[j] // Sum of coefficients * Chebyshev polynomials
		}
		posvelIndex++
		if debugFlag {
			fmt.Printf("interp: Calculated posvel[%d] = %f\n", posvelIndex-1, posvel[posvelIndex-1])
		}
	}

	if velocityFlag <= 1 { // Return if only position is needed
		if debugFlag {
			fmt.Println("interp: Returning after position calculation only")
		}
		return
	}

	// Recurrence relation for derivatives of Chebyshev polynomials T'_i(tc)
	if iinfo.nVelAvail < ncf { // Compute derivative Chebyshev polynomials up to ncf if needed
		for i = 2; i < ncf; i++ {
			iinfo.velCoeff[i] = iinfo.twot*iinfo.velCoeff[i-1] + 2*iinfo.posnCoeff[i-1] - iinfo.velCoeff[i-2] // T'_{n+1} = 2tc*T'_n + 2T_n - T'_{n-1}
		}
		iinfo.nVelAvail = ncf
		if debugFlag {
			fmt.Printf("interp: Updated iinfo.velCoeff, iinfo.nVelAvail = %d\n", iinfo.nVelAvail)
		}
	}

	vfac = (dna + dna) / t[1] // Velocity scaling factor: (2 * na) / interval length
	for i = 0; i < ncm; i++ { // Interpolate velocity components
		tval := 0.0
		coeffPtr := coef[ncf*(i+l*ncm):] // Pointer to coefficients for current component and sub-interval
		for j = 1; j < ncf; j++ {        // Sum of coefficients (starting from j=1) * derivative Chebyshev polynomials
			tval += iinfo.velCoeff[j] * coeffPtr[j]
		}
		posvel[posvelIndex] = tval * vfac // Scale velocity by vfac
		posvelIndex++
		if debugFlag {
			fmt.Printf("interp: Calculated posvel[%d] = %f\n", posvelIndex-1, posvel[posvelIndex-1])
		}
	}

	if velocityFlag == 3 { // Calculate acceleration if velocityFlag is 3 (for pvsun)
		accelCoeffs := make([]float64, maxCheby) // Array to store second derivatives of Chebyshev polynomials
		accelCoeffs[0] = 0.0
		accelCoeffs[1] = 0.0
		for i = 2; i < ncf; i++ {
			accelCoeffs[i] = 4.0*iinfo.velCoeff[i-1] + iinfo.twot*accelCoeffs[i-1] - accelCoeffs[i-2] // T''_{n+1} = 2tc*T''_n + 4T'_n - T''_{n-1}
		}
		for i = 0; i < ncm; i++ { // Interpolate acceleration components
			tval := 0.0
			coeffPtr := coef[ncf*(i+l*ncm):] // Pointer to coefficients for current component and sub-interval
			for j = 2; j < ncf; j++ {        // Sum of coefficients (starting from j=2) * second derivative Chebyshev polynomials
				tval += accelCoeffs[j] * coeffPtr[j]
			}
			posvel[posvelIndex] = tval * vfac * vfac // Scale acceleration by vfac^2
			posvelIndex++
			if debugFlag {
				fmt.Printf("interp: Calculated posvel[%d] = %f\n", posvelIndex-1, posvel[posvelIndex-1])
			}
		}
	}
	if debugFlag {
		fmt.Println("interp: Finished")
	}
}

// quantityDimension returns the dimension (number of components) for a given quantity index.
// Most ephemeris quantities have a dimension of 3 (x, y, z).
// Nutations have dimension 2 (longitude, obliquity), TT-TDB has dimension 1.
func quantityDimension(idx int) int {
	if idx == 11 { // Nutations
		return 2
	} else if idx == 14 { // TDT - TT
		return 1
	} else { // Planets, lunar mantle angles, librations
		return 3
	}
}

// State calculates and interpolates ephemeris data for specified bodies at a given time.
//
// Parameters:
//   - ephem: ephemeris data.
//   - et: Julian Ephemeris Date (JED) for interpolation.
//   - list: Array of flags (0, 1, or 2) indicating which bodies to interpolate (see body indices below).
//     list[i]=0: no interpolation for body i, 1: position only, 2: position and velocity.
//   - pv: Pointer to a [13][6] double array to store interpolated position and velocity vectors.
//     pv[i][0]=x, pv[i][1]=y, pv[i][2]=z, pv[i][3]=dx, pv[i][4]=dy, pv[i][5]=dz for body i.
//   - nut: Slice of 4 doubles to store nutations and rates if list[10] is set.
//     nut[0]=d psi (nutation in longitude), nut[1]=d epsilon (nutation in obliquity),
//     nut[2]=d psi dot, nut[3]=d epsilon dot.
//   - bary: Flag (non-zero to output heliocentric positions, 0 for solar-system barycentric).
//
// Body Indices for 'list' array:
//
//	0: Mercury, 1: Venus, 2: Earth-moon barycenter, 3: Mars, 4: Jupiter, 5: Saturn, 6: Uranus,
//	7: Neptune, 8: Pluto, 9: geocentric moon, 10: nutations, 11: lunar librations, 12: TT-TDB, 13: lunar mantle omegas.
//
// Returns:
//   - 0 on success.
//   - JPL_EPH_OUTSIDE_RANGE if the requested epoch is outside the ephemeris time range.
//   - JPL_EPH_FSEEK_ERROR if file seek operation fails.
//   - JPL_EPH_READ_ERROR if file read operation fails.
func State(ephem *jplEphData, et float64, list [14]int, pv *[13][6]float64, nut []float64, bary int) error {
	if debugFlag {
		fmt.Println("State: Entered")
		fmt.Printf("State: et = %f, list = %v, bary = %d\n", et, list, bary)
	}
	var i, j uint
	var nIntervals uint
	buf := ephem.cache                                    // Cache buffer for ephemeris data
	var t [2]float64                                      // Time parameters for interpolation
	blockLoc := (et - ephem.ephemStart) / ephem.ephemStep // Time block location in ephemeris file
	recomputePvsun := false                               // Flag to control recomputation of Sun's state
	aufac := 1.0 / ephem.au                               // Conversion factor from km to AU

	// Error return for epoch out of range
	if et < ephem.ephemStart || et > ephem.ephemEnd {
		if debugFlag {
			fmt.Println("State: Error - Epoch out of range")
		}
		return ErrOutsideRange
	}

	// Calculate record number and relative time within the interval
	nr := uint32(blockLoc)        // Record number (integer part of blockLoc)
	t[0] = blockLoc - float64(nr) // Fractional time within the interval (0 <= t[0] < 1)
	if t[0] == 0 && nr != 0 {     // Handle case when t[0] is exactly 0, except for the very first interval
		t[0] = 1.0
		nr--
	}
	if nr != ephem.currCacheLoc {
		ephem.currCacheLoc = nr
		_, err := ephem.ifile.Seek(int64((nr+2)*ephem.recsize), io.SeekStart)
		if err != nil {
			if debugFlag {
				fmt.Printf("State: Error - Seek error: %v\n", err)
			}
			return ErrFileSeek
		}
		err = binary.Read(ephem.ifile, defaultByteOrder, buf) // Read record into cache buffer
		if err != nil {
			if debugFlag {
				fmt.Printf("State: Error - Read error: %v\n", err)
			}
			return ErrFileRead
		}
		if ephem.swapBytes != 0 {
			swapBytes64Slice(buf) // Byte-swap if needed
		}
		if debugFlag {
			fmt.Println("State: Read block from file, first 10 values of buf:")
			for k := 0; k < 10 && k < len(buf); k++ {
				fmt.Printf("State: buf[%d] = %e\n", k, buf[k])
			}
		}
	}
	t[1] = ephem.ephemStep // Set interval length

	if ephem.pvsunT != et { // Check if Sun's state needs recomputation for the current time
		recomputePvsun = true // Recompute Sun's state if time has changed
		ephem.pvsunT = et     // Update last computed time for Sun's state
	} else {
		recomputePvsun = false // No need to recompute if time is the same
	}

	// Here, i loops through the "traditional" 14 listed items -- 10
	// solar system objects,  nutations,  librations,  lunar mantle angles,
	// and TT-TDT -- plus a fifteenth:  the solar system barycenter.  That
	// last is quite different:  it's computed 'as needed',  rather than
	// from list[];  the output goes to pvsun rather than the pv array;
	// and three quantities (position,  velocity,  acceleration) are
	// computed (nobody else gets accelerations at present.)
	for nIntervals = 1; nIntervals <= 8; nIntervals *= 2 {
		for i = 0; i < 15; i++ { // Loop through bodies and special quantities (15 total items)
			var quantities int
			var iptr *[3]uint32 // Pointer to IPT array entry for current body/quantity

			if i == 14 { // Special case for Solar System Barycenter (index 14 is SSB in this loop)
				if recomputePvsun { // Only compute if needed
					quantities = 3 // Position, velocity, acceleration for Sun
				}
				iptr = &ephem.ipt[10] // IPT entry for Sun
			} else {
				quantities = list[i] // Get interpolation flag from list
				if i < 10 {
					iptr = &ephem.ipt[i] // IPT entry for planets/moon
				} else {
					iptr = &ephem.ipt[i+1] // IPT entry for nutations, librations, TT-TDB, lunar omegas
				}
			}
			if nIntervals == uint((*iptr)[2]) && quantities != 0 { // Check if current interval matches IPT and interpolation is requested
				var dest []float64 // Destination slice for interpolated data

				if i < 10 {
					dest = pv[i][:] // Destination is pv array for planets/moon
				} else if i == 14 {
					dest = ephem.pvsun[:] // Destination is pvsun array for Sun
				} else {
					dest = nut // Destination is nut array for nutations
				}
				if debugFlag {
					fmt.Printf("State: Calling interp for body %d, iptr: %v, nIntervals: %d, quantities: %d\n", i+1, *iptr, nIntervals, quantities)
					fmt.Printf("State: coef slice start index: %d, ncf: %d, ncm: %d\n", (*iptr)[0]-1, uint((*iptr)[1]), uint(quantityDimension(int(i)+1)))
				}

				// Call Chebyshev interpolation function
				interp(&ephem.iinfo, buf[(*iptr)[0]-1:], t, uint((*iptr)[1]), uint(quantityDimension(int(i)+1)), nIntervals, quantities, dest)

				if i < 10 || i == 14 { // Convert km to AU for planets, moon, and sun
					for j = 0; j < uint(quantities*3); j++ {
						dest[j] *= aufac // Apply AU conversion factor
					}
				}
			}
		}
	}
	if bary == 0 { // Correct for solar system barycenter if barycentric output is requested (bary == 0)
		for i = 0; i < 9; i++ { // Loop through planets (Mercury to Pluto)
			for j = 0; j < uint(list[i]*3); j++ {
				pv[i][j] -= ephem.pvsun[j] // Subtract Sun's position/velocity from planet's SSB position/velocity
			}
		}
	}
	if debugFlag {
		fmt.Println("State: Finished")
	}
	return nil
}

// start400ThConstantName is the file offset to the names of constants beyond the first 400.
const start400ThConstantName = (84*3 + 400*6 + 5*8 + 41*4) // START_400TH_CONSTANT_NAME

// jplHeaderSize is the size of the JPL ephemeris header in bytes.
const jplHeaderSize = (5*8 + 41*4) // JPL_HEADER_SIZE

// initEphemeris initializes the JPL ephemeris data from a binary ephemeris file.
//
// Parameters:
//   - ephemerisFilename: Path to the binary ephemeris file (e.g., "de405.bin").
//   - nam: Optional [][6]byte array to store constant names (pass nil if not needed).
//   - val: Optional []float64 slice to store constant values (pass nil if not needed).
//
// Returns:
//   - Interface to the initialized ephemeris data (jplEphData) on success, nil on failure.
//   - Error if initialization fails (check InitErrorCode() for details).
func initEphemeris(ephemerisFilename string, nam [][6]byte, val []float64) (*jplEphData, error) {
	if debugFlag {
		fmt.Println("InitEphemeris: Entered, filename:", ephemerisFilename)
	}
	var i, j uint
	var deVersion int64
	title := make([]byte, 84)                // Buffer for ephemeris title
	ifile, err := os.Open(ephemerisFilename) // Open ephemeris file
	if err != nil {
		if debugFlag {
			fmt.Printf("InitEphemeris: Error opening file: %v\n", err)
		}
		return nil, fmt.Errorf("failed to open ephemeris file: %w", err)
	}

	rval := &jplEphData{ifile: ifile, pvsunT: -1e+80} // Allocate and initialize jplEphData structure
	tempData := rval                                  // Temporary pointer for easier access to struct fields

	// Read ephemeris title (first 84 bytes)
	n, err := ifile.Read(title)
	if n != 84 || (err != nil && !errors.Is(err, io.EOF)) {
		if debugFlag {
			fmt.Printf("InitEphemeris: Error reading title: %v\n", err)
		}
		return nil, fmt.Errorf("fread title failed: %w", err)
	}
	// Seek to header data location (byte 2652)
	_, err = ifile.Seek(2652, io.SeekStart)
	if err != nil {
		if debugFlag {
			fmt.Printf("InitEphemeris: Error seeking to header: %v\n", err)
		}
		return nil, fmt.Errorf("fseek failed: %w", err)
	}

	header := make([]byte, jplHeaderSize) // Buffer for header data
	// Read header data (jplHeaderSize bytes)
	n, err = ifile.Read(header)
	if n != len(header) || (err != nil && !errors.Is(err, io.EOF)) {
		if debugFlag {
			fmt.Printf("InitEphemeris: Error reading header: %v\n", err)
		}
		return nil, fmt.Errorf("fread header failed: %w", err)
	}
	// Parse header data
	tempData.ephemStart = float64FromBytes(header[0:8])  // Ephemeris start time (JD)
	tempData.ephemEnd = float64FromBytes(header[8:16])   // Ephemeris end time (JD)
	tempData.ephemStep = float64FromBytes(header[16:24]) // Ephemeris step size (days)
	tempData.ncon = uInt32FromBytes(header[24:28])       // Number of constants
	tempData.au = float64FromBytes(header[28:36])        // Astronomical Unit (km)
	tempData.emrat = float64FromBytes(header[36:44])     // Earth-Moon mass ratio

	// Parse IPT array (interpolation parameters table)
	for i := 0; i < 40; i++ {
		offset := 44 + i*4
		tempData.ipt[i/3][i%3] = uInt32FromBytes(header[offset : offset+4]) // IPT[row][column]
	}
	// Check if byte swapping is needed based on ncon value
	tempData.swapBytes = 0
	if tempData.ncon > 65536 { // Heuristic to detect wrong byte order
		tempData.swapBytes = 1 // Set swap flag
		swapBytes32(&tempData.ncon)
		swapBytes64(&tempData.ephemStart)
		swapBytes64(&tempData.ephemEnd)
		swapBytes64(&tempData.ephemStep)
		swapBytes64(&tempData.au)
		swapBytes64(&tempData.emrat)
	}
	// Parse DE version and ephemeris name from title string
	if bytes.HasPrefix(title, []byte("INPOP")) { // INPOP ephemeris format
		deVersionStr := strings.TrimLeft(string(title[5:30]), " ") // DE version string
		i := 0
		for ; i < len(deVersionStr); i++ { // Find end of version number in string
			if deVersionStr[i] < '0' || deVersionStr[i] > '9' {
				break
			}
		}
		var err error
		deVersion, err = strconv.ParseInt(deVersionStr[:i], 10, 64) // Convert version string to integer
		if err != nil {
			if debugFlag {
				fmt.Printf("InitEphemeris: Error parsing INPOP DE version: %v\n", err)
			}
			return nil, fmt.Errorf("atoi de_version (INPOP) failed for '%s': %w", deVersionStr[:i], err)
		}
		nameBytes := title[:30]                                      // Ephemeris name bytes
		if nullIdx := bytes.IndexByte(nameBytes, 0); nullIdx != -1 { // Remove null terminator if present
			nameBytes = nameBytes[:nullIdx]
		}
		nameStr := strings.TrimSpace(string(nameBytes))       // Trim whitespace from name
		if parts := strings.Fields(nameStr); len(parts) > 0 { // Extract first word as name
			copy(tempData.name[:], parts[0]) // Copy name to jplEphData struct
		}
	} else { // Standard JPL ephemeris format
		deVersionStr := strings.TrimLeft(string(title[26:54]), " ") // DE version string
		i := 0
		for ; i < len(deVersionStr); i++ { // Find end of version number in string
			if deVersionStr[i] < '0' || deVersionStr[i] > '9' {
				break
			}
		}
		var err error
		deVersion, err = strconv.ParseInt(deVersionStr[:i], 10, 64) // Convert version string to integer
		if err != nil {
			if debugFlag {
				fmt.Printf("InitEphemeris: Error parsing non-INPOP DE version: %v\n", err)
			}
			return nil, fmt.Errorf("atoi de_version failed for '%s': %w", deVersionStr[:i], err)
		}
		nameBytes := title[24:54]                                    // Ephemeris name bytes
		if nullIdx := bytes.IndexByte(nameBytes, 0); nullIdx != -1 { // Remove null terminator if present
			nameBytes = nameBytes[:nullIdx]
		}
		nameStr := strings.TrimSpace(string(nameBytes))       // Trim whitespace from name
		if parts := strings.Fields(nameStr); len(parts) > 0 { // Extract first word as name
			copy(tempData.name[:], parts[0]) // Copy name to jplEphData struct
		}
	}

	// Adjust IPT indices for lunar librations (historical quirk)
	tempData.ipt[12][0] = tempData.ipt[12][1]
	tempData.ipt[12][1] = tempData.ipt[12][2]
	tempData.ipt[12][2] = tempData.ipt[13][0]
	tempData.ephemerisVersion = uint64(deVersion) // Store DE version

	// Handle TT-TDB data (present in DE430 and later, but not always reliable)
	if deVersion >= 430 && tempData.ncon != 400 {
		// Seek past constants if more than 400
		if tempData.ncon > 400 {
			_, err = ifile.Seek(int64(tempData.ncon-400)*6, io.SeekCurrent)
			if err != nil {
				if debugFlag {
					fmt.Printf("InitEphemeris: Error seeking past 400 constants: %v\n", err)
				}
				return nil, fmt.Errorf("fseek failed after 400 constants: %w", err)
			}
		}
		ipt1314Header := make([]byte, 6*4) // Buffer for IPT[13] and IPT[14] data
		_, err = ifile.Read(ipt1314Header)
		if err != nil && !errors.Is(err, io.EOF) {
			if debugFlag {
				fmt.Printf("InitEphemeris: Error reading ipt[13][0]: %v\n", err)
			}
			return nil, fmt.Errorf("fread ipt[13][0] failed: %w", err)
		}
		ipt1314Reader := strings.NewReader(string(ipt1314Header))
		for i := 0; i < 6; i++ { // Read 6 integers for IPT[13] and IPT[14]
			val32, err := getUint32(ipt1314Reader) // Helper function to read uint32 from string reader
			if err != nil {
				if debugFlag {
					fmt.Printf("InitEphemeris: Error getting uint32 for ipt[%d][%d]: %v\n", (13+i)/3, (13+i)%3, err)
				}
				return nil, fmt.Errorf("getUint32 ipt[%d][%d] (13/14) failed: %w", (13+i)/3, (13+i)%3, err)
			}
			if i < 3 {
				tempData.ipt[13][i] = val32 // IPT[13]
			} else {
				tempData.ipt[14][i-3] = val32 // IPT[14]
			}
		}

	} else { // Mark IPT[13] and IPT[14] as invalid if DE version < 430 or ncon == 400
		tempData.ipt[13][0] = uint32(0) // Set to 0 as invalid
	}

	if tempData.swapBytes != 0 { // Byte swapping for IPT array (currently disabled)
		for j = 0; j < 3; j++ {
			for i = 0; i < 15; i++ {
				swapBytes32(&tempData.ipt[i][j])
			}
		}
	}
	// Sanity check for TT-TDB IPT data (cross-check indices)
	if tempData.ipt[13][0] != (tempData.ipt[12][0]+tempData.ipt[12][1]*tempData.ipt[12][2]*3) ||
		tempData.ipt[14][0] != (tempData.ipt[13][0]+tempData.ipt[13][1]*tempData.ipt[13][2]*3) {
		// Zero out IPT[13] and IPT[14] if sanity check fails (likely garbage data)
		for i = 13; i < 15; i++ {
			for j = 0; j < 3; j++ {
				tempData.ipt[i][j] = 0
			}
		}
	}
	// Sanity check for Earth-Moon mass ratio
	if tempData.emrat > 81.3008 || tempData.emrat < 81.30055 {
		if debugFlag {
			fmt.Printf("InitEphemeris: Error - Earth-Moon ratio out of range: %f\n", tempData.emrat)
		}
		return nil, fmt.Errorf("ephemeris file corrupt: Earth-Moon ratio out of range: %f", tempData.emrat)
	}

	// Calculate kernel size, record size, and number of coefficients
	tempData.kernelSize = 4 // Initial kernel size
	for i = 0; i < 15; i++ {
		tempData.kernelSize += 2 * tempData.ipt[i][1] * tempData.ipt[i][2] * uint32(quantityDimension(int(i))) // Sum of coefficients for each quantity
	}
	tempData.recsize = tempData.kernelSize * 4 // Record size in bytes (kernel size * 4 bytes/double)
	tempData.ncoeff = tempData.kernelSize / 2  // Number of coefficients (kernel size / 2 doubles/coefficient)

	// Allocate cache buffer for ephemeris data
	rval.cache = make([]float64, tempData.ncoeff)

	// Initialize interpolation info structure
	rval.iinfo.posnCoeff[0] = 1.0  // Initial Chebyshev polynomial values
	rval.iinfo.posnCoeff[1] = -2.0 // Bogus initial value, corrected in interp()
	rval.iinfo.velCoeff[0] = 0.0
	rval.iinfo.velCoeff[1] = 1.0
	rval.currCacheLoc = uint32(4294967295) // Initialize cache location to invalid value

	// Handle constant names beyond 400 (if present)
	if rval.ncon == 400 {
		buff := make([]byte, 6)                                   // Buffer for constant name
		_, err = ifile.Seek(start400ThConstantName, io.SeekStart) // Seek to start of extra constant names
		if err != nil {
			if debugFlag {
				fmt.Printf("InitEphemeris: Error seeking to 400th constant name: %v\n", err)
			}
			return nil, fmt.Errorf("fseek to 400th constant name failed: %w", err)
		}
		for { // Read constant names until EOF or read error
			n, err := ifile.Read(buff)
			if err != nil && errors.Is(err, io.EOF) {
				break // End of file
			}
			if err != nil {
				if debugFlag {
					fmt.Printf("InitEphemeris: Error reading constant name (400+): %v\n", err)
				}
				return nil, fmt.Errorf("fread constant name (400+) failed: %w", err)
			}
			if n != 6 { // Should read exactly 6 bytes for constant name
				break // Assume end of constant names if less than 6 bytes read
			}
			rval.ncon++ // Increment constant count for each extra name found
		}
	}

	if val != nil { // Read constant values if 'val' slice is provided
		_, err = ifile.Seek(int64(rval.recsize), io.SeekStart) // Seek to start of constant values
		if err != nil {
			if debugFlag {
				fmt.Printf("InitEphemeris: Error seeking to constant values: %v\n", err)
			}
			return nil, fmt.Errorf("fseek to constants values failed: %w", err)
		}
		err = binary.Read(ifile, defaultByteOrder, val[:rval.ncon]) // Read constant values into 'val' slice
		if err != nil && !errors.Is(err, io.EOF) {
			if debugFlag {
				fmt.Printf("InitEphemeris: Error reading constant values: %v\n", err)
			}
			return nil, fmt.Errorf("fread constant values failed: %w", err)
		}
		if rval.swapBytes != 0 { // Byte swap constant values if needed (currently disabled)
			swapBytes64Slice(val[:rval.ncon])
		}
	}

	if nam != nil { // Read constant names if 'nam' array is provided
		_, err = ifile.Seek(84*3, io.SeekStart) // Seek to start of constant names (after title lines)
		if err != nil {
			if debugFlag {
				fmt.Printf("InitEphemeris: Error seeking to constant names: %v\n", err)
			}
			return nil, fmt.Errorf("fseek to constant names failed: %w", err)
		}
		for i := uint(0); i < uint(rval.ncon); i++ { // Read constant names up to ncon
			if i == 400 { // Seek to start of extra constant names if index is 400
				_, err = ifile.Seek(start400ThConstantName, io.SeekStart)
				if err != nil {
					if debugFlag {
						fmt.Printf("InitEphemeris: Error seeking to 400+ constant names: %v\n", err)
					}
					return nil, fmt.Errorf("fseek to 400+ constant names failed: %w", err)
				}
			}
			_, err = ifile.Read(nam[i][:]) // Read constant name into 'nam' array
			if err != nil && !errors.Is(err, io.EOF) {
				if debugFlag {
					fmt.Printf("InitEphemeris: Error reading constant name [%d]: %v\n", i, err)
				}
				return nil, fmt.Errorf("fread constant name [%d] failed: %w", i, err)
			}
		}
	}
	if debugFlag {
		fmt.Println("InitEphemeris: Finished, ephemeris initialized successfully.")
	}
	return rval, nil
}

// closeEphemeris closes the ephemeris file associated with the given ephemeris data interface.
// It's important to call this function to release file resources when finished using the ephemeris.
func closeEphemeris(ephem *jplEphData) error {
	if debugFlag {
		fmt.Println("CloseEphemeris: Entered")
	}
	if ephem.ifile != nil {
		err := ephem.ifile.Close() // Close the ephemeris file
		if debugFlag {
			if err != nil {
				fmt.Printf("CloseEphemeris: Error closing file: %v\n", err)
			} else {
				fmt.Println("CloseEphemeris: File closed successfully")
			}
		}
		return err // Return any error from closing the file
	}
	if debugFlag {
		fmt.Println("CloseEphemeris: No file to close.")
	}
	return nil // Return nil if no file was open
}

// getConstant retrieves a specific JPL constant value by its index.
//
// Parameters:
//   - idx: Index of the constant to retrieve (0-based).
//   - ephem: ephemeris data.
//   - constantName: Byte slice of size 7 to store the constant name (optional, can be nil if name is not needed).
//
// Returns:
//   - The constant value as a float64. Returns 0 if index is invalid or read error occurs (check debug log for warnings).
func getConstant(idx int, ephem *jplEphData, constantName []byte) float64 {
	rval := 0.0

	if idx >= 0 && idx < int(ephem.ncon) { // Validate constant index
		var seekLoc int64
		if idx < 400 { // Calculate file offset for constant name based on index
			seekLoc = 84*3 + int64(idx)*6
		} else {
			seekLoc = start400ThConstantName + int64(idx-400)*6
		}

		_, err := ephem.ifile.Seek(seekLoc, io.SeekStart) // Seek to constant name location
		if err != nil {
			if debugFlag {
				fmt.Printf("GetConstant: Warning: fseek to constant name failed: %v\n", err) // Non-critical error, name might be unavailable
			}
			return 0 // Return 0 on seek error (constant name unavailable)
		}

		n, err := ephem.ifile.Read(constantName[:6]) // Read constant name (6 bytes)
		if err != nil && !errors.Is(err, io.EOF) {
			if debugFlag {
				fmt.Printf("GetConstant: Warning: fread constant name failed: %v\n", err) // Non-critical error, name might be unavailable
			}
			return 0 // Return 0 on read error (constant name unavailable)
		}
		if n == 6 { // If constant name was read successfully
			constantName[6] = 0                                                        // Null terminate the name (for C-style string compatibility, though Go doesn't need it)
			_, err = ephem.ifile.Seek(int64(ephem.recsize)+int64(idx)*8, io.SeekStart) // Seek to constant value location
			if err != nil {
				if debugFlag {
					fmt.Printf("GetConstant: Warning: fseek to constant value failed: %v\n", err) // Non-critical error, value might be unavailable
				}
				return 0 // Return 0 on seek error (constant value unavailable)
			}
			var val float64
			err = binary.Read(ephem.ifile, defaultByteOrder, &val) // Read constant value (double-precision)
			if err != nil && !errors.Is(err, io.EOF) {
				if debugFlag {
					fmt.Printf("GetConstant: Warning: fread constant value failed: %v\n", err) // Non-critical error, value might be unavailable
				}
				return 0 // Return 0 on read error (constant value unavailable)
			}
			rval = val                // Assign read constant value to return value
			if ephem.swapBytes != 0 { // Byte swap constant value if needed (currently disabled)
				swapBytes64(&rval)
			}
		}
	}
	return rval // Return retrieved constant value (or 0 if error)
}

// getEphemName returns the name of the ephemeris (e.g., "DE405").
func getEphemName(ephem *jplEphData) string {
	return string(ephem.name[:]) // Return ephemeris name as string
}

// setDebugFlag enables or disables debug print statements within the jpleph package.
// When enabled, debug information will be printed to the console.
func setDebugFlag(enable bool) {
	debugFlag = enable // Set the global debug flag
	if debugFlag {
		fmt.Println("Debug flag enabled")
	}
}

// GetCachePointer is an internal function to access the coefficient cache.
//
// Returns:
//   - []float64: A slice of float64 representing the coefficient cache.
func GetCachePointer(ephem *Ephemeris) []float64 {
	return ephem.ephemData.cache
}
