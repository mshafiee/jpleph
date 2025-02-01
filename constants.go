// ./constants.go
package jpleph

/*
Package jpleph provides constants for accessing JPL ephemeris data.

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

// Constants for jpleph package.
//
// This file defines constants used in the jpleph package, originally
// defined in the C header file `jpleph.h`.

// Constants used in GetDouble() and GetLong() functions to specify
// which ephemeris parameter to retrieve.
const (
	JPL_EPHEM_START_JD           = 0   // Start Julian Date of ephemeris data
	JPL_EPHEM_END_JD             = 8   // End Julian Date of ephemeris data
	JPL_EPHEM_STEP               = 16  // Time step (in days) of ephemeris data
	JPL_EPHEM_N_CONSTANTS        = 24  // Number of constants in the ephemeris file
	JPL_EPHEM_AU_IN_KM           = 28  // Astronomical Unit in kilometers (km/AU)
	JPL_EPHEM_EARTH_MOON_RATIO   = 36  // Earth-Moon mass ratio
	JPL_EPHEM_IPT_ARRAY          = 44  // Base offset for IPT (interpolation parameters table) array
	JPL_EPHEM_EPHEMERIS_VERSION  = 224 // Ephemeris version (e.g., 405, 406)
	JPL_EPHEM_KERNEL_SIZE        = 228 // Size of the ephemeris kernel in data units (doubles)
	JPL_EPHEM_KERNEL_RECORD_SIZE = 232 // Size of a single ephemeris record in bytes
	JPL_EPHEM_KERNEL_NCOEFF      = 236 // Number of coefficients per data record
	JPL_EPHEM_KERNEL_SWAP_BYTES  = 240 // Flag indicating if byte swapping is needed (non-zero if yes)
)

// Error codes returned by State() and Pleph() functions.
const (
	JPL_EPH_OUTSIDE_RANGE             = -1 // Requested Julian Date is outside the ephemeris time range
	JPL_EPH_READ_ERROR                = -2 // Error occurred during file read operation
	JPL_EPH_QUANTITY_NOT_IN_EPHEMERIS = -3 // Requested quantity (e.g., nutations, librations) is not available in the ephemeris file
	JPL_EPH_INVALID_INDEX             = -5 // Invalid target or center body index provided
	JPL_EPH_FSEEK_ERROR               = -6 // Error occurred during file seek operation
)

// Error codes returned by InitErrorCode() after calling InitEphemeris().
const (
	JPL_INIT_NO_ERROR       = 0   // No error during initialization
	JPL_INIT_FILE_NOT_FOUND = -1  // Ephemeris file not found at the specified path
	JPL_INIT_FSEEK_FAILED   = -2  // File seek operation failed during initialization
	JPL_INIT_FREAD_FAILED   = -3  // Initial file read operation failed during initialization
	JPL_INIT_FREAD2_FAILED  = -4  // Second file read operation failed during initialization
	JPL_INIT_FREAD5_FAILED  = -10 // Fifth file read operation failed during initialization (IPT data)
	JPL_INIT_FILE_CORRUPT   = -5  // Ephemeris file is likely corrupt or invalid
	JPL_INIT_MEMORY_FAILURE = -6  // Memory allocation failed during initialization
	JPL_INIT_FREAD3_FAILED  = -7  // Third file read operation failed during initialization (constant values)
	JPL_INIT_FREAD4_FAILED  = -8  // Fourth file read operation failed during initialization (constant names)
	JPL_INIT_NOT_CALLED     = -9  // InitEphemeris() has not been called yet, or initialization failed
)
