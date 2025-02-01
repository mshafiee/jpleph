// ./internal_types.go
package jpleph

import "io"

/*
Package jpleph provides internal definitions for JPL ephemeris functions.

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

// Internal definitions for JPL ephemeris functions.
//
// This file contains internal data structures and constants used by the
// jpleph package. These are not intended for direct use by users of the package.

// File structure notes (as of March 25, 2014):
//
// This section provides notes on the binary file structure of JPL ephemeris files,
// based on analysis and reverse engineering.  Offsets and sizes are approximate and may vary slightly between different DE versions.
// For detailed and definitive information, refer to JPL's official documentation for each ephemeris version.
//
// Header (Located at the beginning of the file):
//
// Bytes 0-83:    First title line: "JPL Planetary Ephemeris DExxx/LExxx" (84 bytes)
// Bytes 84-167:   Second title line: "Start Epoch: JED = ttttttt.t yyyy-MMM-dd 00:00:00" (84 bytes)
// Bytes 168-251:  Third title line: "Final Epoch: JED = ttttttt.t yyyy-MMM-dd 00:00:00" (84 bytes)
// Bytes 252-257:  Name of the 0th constant: e.g., "DENUM " (6 bytes)
// Bytes 258-263:  Name of the 1st constant (6 bytes)
// ...
// Bytes 252+6n to 257+6n: Name of the nth constant (6 bytes), for constants 0 to 399.
// Bytes 2646-2651: Name of the 399th (400th) constant (6 bytes)
//
// Numerical Header Data (Following Constant Names):
//
// Bytes 2652-2659:  ephem_start (start Julian Ephemeris Date, double-precision float64, 8 bytes)
// Bytes 2660-2667:  ephem_end (end Julian Ephemeris Date, double-precision float64, 8 bytes)
// Bytes 2668-2675:  ephem_step (ephemeris time step in days, double-precision float64, 8 bytes)
// Bytes 2676-2679:  ncon (number of constants, 32-bit integer, 4 bytes)
// Bytes 2680-2687:  AU in km (Astronomical Unit in kilometers, double-precision float64, 8 bytes) - approximately 149597870.700000 km
// Bytes 2688-2695:  Earth/moon mass ratio (double-precision float64, 8 bytes) - approximately 81.300569
// Bytes 2696-2851:  ipt array (Interpolation Parameters Table, 15x3 array of 32-bit integers, 15 * 3 * 4 = 180 bytes) - ipt[0][0] to ipt[14][2]
// Bytes 2852-2855:  ephemeris version (e.g., 405, 430, etc., 32-bit integer, 4 bytes)
//
// IPT Array Details and Special Cases:
//
// Note: In the original JPL FORTRAN code, the IPT array is further subdivided into:
//   - lpt[0..2]:  ipt[12][0..2] - Lunar libration offsets
//   - rpt[0..2]:  ipt[13][0..2] - Lunar Euler angle rate offsets (new in DE-430t and later)
//   - tpt[0..2]:  ipt[14][0..2] - TT-TDB offsets (new in DE-430t and later)
//
// For DE versions prior to DE-430t, ipt[13] and ipt[14] are typically zero or invalid.
//
// Constant Names (Beyond 400, if present):
//
// Bytes 2856-2861 onwards: If the number of constants (ncon) is greater than 400, the names of constants 400 and above follow sequentially in 6-byte chunks.
//   - Bytes 2856-2861: Name of the 400th (401st) constant (6 bytes)
//   - Bytes 2862-2867: Name of the 401st (402nd) constant (6 bytes)
//   - ... and so on, until all constant names are listed.
//
// IPT[13] and IPT[14] Data Location:
//
// After the last constant name (or immediately after byte 2855 if ncon <= 400), the data for ipt[13][0..2] and ipt[14][0..2] is stored in 24 bytes:
//
// - If n_constants <= 400, these bytes immediately follow the header at bytes 2856-2879.
// - If n_constants > 400, the offset to these bytes is calculated by adding (n_constants - 400) * 6 to 2856.
//
// Constant Values (Following Header and IPT Data):
//
// Starting at byte offset 'recsize' (record size), the actual values of the constants are stored as double-precision floats (8 bytes each).
// The total size of the constant values section is 8 * ncon bytes.
//
// Data Records (Following Constant Values):
//
// The ephemeris data records themselves follow the constant values section.
// Each record has a size of 'recsize' bytes, where recsize = 8 * ncoeff, and ncoeff is the number of coefficients per record.
//
// Examples of ncoeff and Time Ranges for Various DE Ephemerides:
//
// DE-102: ncoeff = 773
// DE-200 & 202: ncoeff = 826
// DE-403, 405, 410, 413, 414, 418, 421, 422, 423, 424, 430, 431, 433, 434, 435, 436, 438, 440, 441: ncoeff = 1018
// DE-404, 406: ncoeff = 728
// DE-432: ncoeff = 938
// DE-430t, 432t: ncoeff = 982
// DE-436t, 440t: ncoeff = 1122
// DE-438t: ncoeff = 1042
//
// Example Time Ranges (Julian Dates and Approximate Years):
//
// DE-406: JD 625360.500 to 2816848.500 (years -2999.821 to 3000.146)
// DE-410: JD 2436912.500 to 2458832.500 (years 1959.938 to 2019.952)
// DE-422: JD 625648.500 to 2816816.500 (years -2999.032 to 3000.059)
// DE-430: JD 2287184.500 to 2688976.500 (years 1550.005 to 2650.052)
// DE-431: JD -3027215.500 to 7930192.500 (years -13000.029 to 16999.719)
// DE-432 to DE-440t: JD 2287184.500 to 2688976.500 (years 1550.005 to 2650.052)
// DE-441: JD -3100015.500 to 8000016.5 (years -13200 to 17191)

// maxCheby defines the maximum number of Chebyshev coefficients used in the ephemeris.
// Currently set to 18, which is sufficient for all known JPL DE ephemerides.
// An assertion in the code will trigger if this value needs to be increased for future ephemerides.
const maxCheby = 18

// interpolationInfo struct holds data required for Chebyshev interpolation.
// Used to optimize interpolation by storing and reusing Chebyshev polynomial values.
type interpolationInfo struct {
	posnCoeff  [maxCheby]float64 // posnCoeff stores Chebyshev polynomial values T_i(tc).
	velCoeff   [maxCheby]float64 // velCoeff stores derivatives of Chebyshev polynomials T'_i(tc).
	nPosnAvail uint              // nPosnAvail indicates the number of position Chebyshev polynomials already computed and available in posnCoeff.
	nVelAvail  uint              // nVelAvail indicates the number of velocity Chebyshev polynomial derivatives already computed and available in velCoeff.
	twot       float64           // twot stores 2 * tc, used as an optimization in Chebyshev recurrence relations.
}

// jplEphData struct encapsulates data to access and interpolate a JPL ephemeris file.
// Instances are returned by InitEphemeris() and passed to other jpleph functions.
type jplEphData struct {
	ephemStart       float64       // ephemStart is the starting Julian Ephemeris Date of the ephemeris data.
	ephemEnd         float64       // ephemEnd is the ending Julian Ephemeris Date of the ephemeris data.
	ephemStep        float64       // ephemStep is the time step (in days) between data records in the ephemeris.
	ncon             uint32        // ncon is the number of constants in the ephemeris file.
	au               float64       // au is the value of the Astronomical Unit in kilometers, as defined in the ephemeris.
	emrat            float64       // emrat is the Earth-Moon mass ratio used in the ephemeris.
	ipt              [15][3]uint32 // ipt is the Interpolation Parameters Table, a 15x3 array of integers controlling interpolation.
	ephemerisVersion uint64        // ephemerisVersion indicates the JPL ephemeris version (e.g., 405, 406, 430).

	// Internal data computed and used by the jpleph package.
	kernelSize   uint32            // kernelSize is the size of the ephemeris kernel in doubles (number of doubles per record).
	recsize      uint32            // recsize is the size of a single ephemeris data record in bytes.
	ncoeff       uint32            // ncoeff is the number of Chebyshev coefficients per data record (kernelSize / 2).
	swapBytes    uint32            // swapBytes is a flag indicating if byte swapping is needed when reading the ephemeris file (non-zero if yes).
	currCacheLoc uint32            // currCacheLoc stores the record number of the currently cached data block.
	pvsun        [9]float64        // pvsun stores the position, velocity, and acceleration of the Sun (Solar System Barycentric).
	pvsunT       float64           // pvsunT stores the Julian Ephemeris Date for which pvsun was last computed, for caching purposes.
	cache        []float64         // cache is a buffer to store a single ephemeris data record, read from the file.
	iinfo        interpolationInfo // iinfo is an instance of interpolationInfo, used to store Chebyshev interpolation data for optimization.
	ifile        io.ReadSeekCloser // ifile is an interface representing the opened ephemeris file.
	name         [32]byte          // name stores the name of the ephemeris (e.g., "DE405", "INPOP-19a").
}
