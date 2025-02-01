# jpleph: Go Library for JPL DE Ephemerides

[![GoDoc](https://godoc.org/github.com/mshafiee/jpleph?status.svg)](https://pkg.go.dev/github.com/mshafiee/jpleph) [![License](https://img.shields.io/badge/License-GPLv2%20or%20later-blue.svg)](https://www.gnu.org/licenses/gpl-2.0-standalone.html)

*Last updated: 2025 January 30.*

jpleph is a pure Go library designed to access, read, and compute positions and velocities from JPL Development Ephemeris (DE) binary files. Built for platform-independence and efficiency, the library automatically handles both little-endian and big-endian data formats, making it straightforward to integrate JPL’s high-precision solar system ephemerides into your Go applications.

**Repository:** [https://github.com/mshafiee/jpleph](https://github.com/mshafiee/jpleph)

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
    * [Basic Example](#basic-example)
    * [Loading Constants](#loading-constants)
	* [Accessing Constants](#accessing-constants)
    * [Error Handling](#error-handling)
* [What this Go library does](#what-does-this-go-library-do)
* [JPL DE basics](#jpl-de-basics)
* [JPL DE versions](#jpl-de-versions)
* [Finding DE files](#finding-de-files)
* [Other DE source code](#other-implementations)
* [Why this Go library exists](#why-this-go-library-exists)
* [Reasons to use JPL DE rather than analytic series](#reasons-to-use-jpl-de-rather-than-analytic-series)
* [Contributing](#contributing)
* [License](#license)
* [Copyright/legal issues (it's under the GPL)](#copyrightlegal-issues-its-under-the-gpl)
* [Changelog](#changelog)

## [Installation](#installation)

To install the `jpleph` Go library, use `go get`:

```bash
go get github.com/mshafiee/jpleph
```

Ensure that Go is installed and your `GOPATH` and `PATH` environment variables are set correctly.

## [Usage](#usage)

### [Basic Example](#basic-example)

Here's a simple example demonstrating how to use the `jpleph` library to calculate the position of Mars relative to the Sun at a specific Julian Ephemeris Date (JED):

```go
package main

import (
	"fmt"
	"log"
	"errors"

	"github.com/mshafiee/jpleph"
)

func main() {
	// Replace "path/to/your/de440.bin" with the actual path to your DE file.
	eph, err := jpleph.NewEphemeris("path/to/your/de440.bin", false)
	if err != nil {
		log.Fatalf("Failed to initialize ephemeris: %v", err)
	}
	defer eph.Close()

	et := 2451545.0 // Example JED (J2000.0)
	pos, vel, err := eph.CalculatePV(et, jpleph.Mars, jpleph.CenterSun, true)
	if err != nil {
		log.Fatalf("Error calculating PV: %v", err)
	}

	fmt.Println("Position of Mars relative to the Sun at JED", et)
	fmt.Printf("  Position (AU): X = %12.7f, Y = %12.7f, Z = %12.7f\n", pos.X, pos.Y, pos.Z)
	fmt.Printf("  Velocity (AU/day): DX = %12.7f, DY = %12.7f, DZ = %12.7f\n", vel.DX, vel.DY, vel.DZ)
}
```

Remember to replace `"path/to/your/de440.bin"` with the actual path to your JPL DE ephemeris file. The position is returned in Astronomical Units (AU) and velocity in AU/day.


### [Loading Constants](#loading-constants)

If you need to access constant values and names from the ephemeris file, you can load them during initialization by setting the `loadConstants` parameter to `true` when creating a new `Ephemeris` object. This will read all constants from the file into memory during initialization.

```go
eph, err := jpleph.NewEphemeris("path/to/your/de440.bin", true) // loadConstants = true
if err != nil {
	log.Fatalf("Failed to initialize ephemeris with constants: %v", err)
}
defer eph.Close()
```

Loading constants adds to the initialization time but allows for quick access to them later.

### [Accessing Constants](#accessing-constants)

Once constants are loaded, you can retrieve constant values and names using their index. The following constants are typically available in JPL DE ephemeris files:

* `DENUM`: **Planetary ephemeris number.** Identifies the specific planetary ephemeris (e.g., DE440).
* `LENUM`: **Lunar ephemeris number.** Identifies the lunar ephemeris associated with the planetary ephemeris.
* `TDATEF`, `TDATEB`: **Dates of the Forward and Backward Integrations.** These dates mark the time span of the numerical integration used to create the ephemeris. "Forward" and "Backward" refer to the direction of integration in time relative to a central epoch.
* `CLIGHT`: **Speed of light (km/s).** The speed of light in kilometers per second, used in relativistic calculations.
* `AU`: **Number of kilometers per astronomical unit.** The conversion factor between astronomical units (AU) and kilometers (km).
* `EMRAT`: **Earth-Moon mass ratio.** The ratio of the mass of the Earth to the mass of the Moon.
* `GMi`: **GM for ith planet [au<sup>3</sup>/day<sup>2</sup>].** The product of the Gravitational constant (G) and the mass (M) for the i-th planet, in units of astronomical units cubed per day squared. 'i' refers to the planet number (e.g., `GM3` is for Earth). GM is often used in celestial mechanics as it's known more precisely than G and M separately.
* `GMB`: **GM for the Earth-Moon Barycenter [au<sup>3</sup>/day<sup>2</sup>].** The GM value for the barycenter (center of mass) of the Earth-Moon system, in au<sup>3</sup>/day<sup>2</sup>.
* `GMS`: **Sun (= k<sup>2</sup>) [au<sup>3</sup>/day<sup>2</sup>].** The GM value for the Sun, which is numerically equal to the square of the Gaussian gravitational constant (k), in au<sup>3</sup>/day<sup>2</sup>.
* `X1, ..., ZD9`: **Initial conditions for the numerical integration, given at "JDEPOC", with respect to "CENTER".** These are the starting position and velocity vectors for the bodies in the solar system at the `JDEPOC` epoch, used as initial values for the numerical integration. They are given with respect to the `CENTER`.
* `JDEPOC`: **Epoch (JED) of initial conditions, normally JED 2440400.5.** The Julian Ephemeris Date (JED) at which the initial conditions (`X1, ..., ZD9`) are given. JED is a continuous count of days and fractions of a day since a specific epoch in the past, used for timekeeping in astronomy. 2440400.5 corresponds to 1969 December 19.0 TT.
* `CENTER`: **Reference center for the initial conditions (Sun: 11, Solar System Barycenter: 12).** Indicates the origin of the coordinate system for the initial conditions. It is typically either the Sun (number 11) or the Solar System Barycenter (number 12).
* `MAiiii`: **GM's of asteroid number iiii [au<sup>3</sup>/day<sup>2</sup>].** The GM values for specific asteroids, where 'iiii' is the asteroid number (e.g., `MA0001` for asteroid 1 Ceres). These constants are included for ephemerides that consider asteroid perturbations.
* `PHI, THT, PSI`: **Euler angles of the orientation of the lunar mantle.** A set of three angles (Euler angles) describing the orientation of the lunar mantle (the outer solid part of the Moon) with respect to a reference frame.
* `OMEGAX, ...`: **Rotational velocities of the lunar mantle.** Components of the angular velocity vector describing the rotation rate of the lunar mantle.
* `PHIC, THTC, PSIC`: **Euler angles of the orientation of the lunar core.** Euler angles describing the orientation of the lunar core (the inner, possibly fluid, part of the Moon).
* `OMGCX, ...`: **Rotational velocities of the lunar core.** Components of the angular velocity vector describing the rotation rate of the lunar core.

Here's how to access a constant by its index:

```go
constantValue, err := eph.GetConstantValue(0)
if err != nil {
	log.Fatalf("Error getting constant value: %v", err)
}
constantName, err := eph.GetConstantName(0)
if err != nil {
	log.Fatalf("Error getting constant name: %v", err)
}

fmt.Printf("Constant 0: Name = '%s', Value = %f\n", constantName, constantValue)
```

### [Error Handling](#error-handling)

The `jpleph` library uses standard Go error handling.  Functions return errors, which should be checked to ensure proper execution. Example of checking for specific errors:
```go
_, _, err := eph.CalculatePV(et, jpleph.Nutations, jpleph.CenterSun, true)
if err != nil {
	if errors.Is(err, jpleph.ErrQuantityNotInEphemeris) {
		fmt.Println("Error: Nutations data not available in this ephemeris file.")
	} else {
		log.Fatalf("Error calculating PV: %v", err)
	}
}
```

Refer to the [api.go](./api.go) file for a list of exported error variables.

## [What this Go library does](#what-does-this-go-library-do)

This Go library offers functionality for reading and computing positions from JPL DE-xxx binary ephemerides.  Similar to the original C/C++ implementation, this Go version is designed to handle both little-Endian and big-Endian ephemeris files automatically.  It determines the byte order of the ephemeris file upon first read and adjusts accordingly, eliminating the need for recompilation when switching between different ephemeris versions or byte orders.

Currently, this Go library supports a wide range of DE ephemerides, including (but not limited to) DE-405, DE-406, DE-422, DE-430, DE-431, DE-432, DE-435, DE-440, and DE-441. It is capable of handling the extended time range of DE-431 (years -13000 to +17000) and the TT-TDB time scale data present in DE-430t and DE-432t ephemerides.

This library is implemented in pure Go and aims to be platform-independent. It has been tested on Linux and macOS (arm64 and amd64 architecture).  It leverages Go's built-in `encoding/binary` package for efficient binary data handling. The core ephemeris functions are designed to be easily used without requiring deep knowledge of the underlying implementation. Error handling is implemented using Go's error return mechanism, allowing for robust integration into larger applications.

This Go library is based on the concepts and algorithms found in the C source code by Bill Gray, which itself was derived from Piotr A. Dybczyński's C and Fortran code. While the underlying logic is inspired by these sources, this Go implementation is a complete rewrite in Go, taking advantage of Go's language features and standard library.

## [JPL DE basics](#jpl-de-basics)

JPL Development Ephemerides (DE) are binary files containing Chebyshev polynomials that represent the positions and velocities of solar system bodies over time. These ephemerides are generated through high-precision numerical integration of the equations of motion, incorporating a wealth of observational data.

**Key Characteristics of JPL DE Files:**

*   **Chebyshev Polynomials:** Positions and velocities are represented using Chebyshev polynomials, which are efficient for interpolation within defined time intervals. Polynomials are provided for overlapping 32-day intervals to ensure accuracy across the entire ephemeris time span.
*   **Units:** Positions within the binary files are stored as Chebyshev coefficients in kilometers (km), but this Go library, by default, returns positions in astronomical units (AU) and velocities in AU/day. The conversion to AU is done using the value of the astronomical unit (AU) constant stored within the ephemeris file itself.
*   **Time Scale:** The integration time unit is days of Barycentric Dynamical Time (TDB).
*   **Binary Format:** DE files are binary for efficiency and are designed to be machine-independent, though endianness must be considered. This Go library handles endianness automatically.
*   **Data Content:** Most DE files include positions and velocities for major planets, the Sun, and the Moon. Many also contain:
    *   **Lunar Librations:** Chebyshev coefficients for lunar libration angles.
    *   **Nutation Series:**  Often includes the 1980 IAU nutation series (for backward compatibility, even though it's not the latest).

**Accessing DE Data:**

*   **SPICE Toolkit:** The recommended method for reading DE files is using NASA's SPICE toolkit, which provides comprehensive tools for ephemeris handling and ancillary data.
*   **Direct Reading (jpleph):** This Go library offers a direct way to read DE binary files, providing a lightweight alternative to SPICE for position and velocity calculations.

For a comprehensive introduction to JPL DE ephemerides, the best resource is the official [JPL Planetary and Lunar Ephemerides Export Information document](ftp://ssd.jpl.nasa.gov/pub/eph/planets/ascii/de_export.txt) (often found as `de_export.txt` or similar in the JPL ephemeris FTP directories), and the [Readme file from the JPL *Horizons* site](https://ssd.jpl.nasa.gov/planets/eph_export.html). These documents explain the fundamentals of DE ephemerides in detail.  The [Fortran Programs User Guide](ftp://ssd.jpl.nasa.gov/pub/eph/planets/fortran/userguide.txt) available at the JPL FTP site also provides valuable insights into the structure and usage of DE files.

 For a deeper understanding of the inner workings of DE ephemerides and how to convert their raw output into usable positions,  the book *Fundamental Ephemeris Computations* by Paul J. Heafner (Willmann-Bell) is highly recommended.

Further mathematical and algorithmic details regarding DE data usage can be found on [How to read the JPL Ephemeris and Perform Barycentering](http://lheawww.gsfc.nasa.gov/users/craigm/bary/) article, which is primarily aimed at pulsar timing applications but contains valuable information. However, for most users simply seeking to extract positions from DE files, this level of detail is generally not necessary.

## [JPL DE Ephemerides Overview](#jpl-de-versions)

JPL DE ephemerides are datasets containing the positions and velocities of solar system bodies computed through high-precision numerical integration.
They are continuously refined and updated by NASA's Jet Propulsion Laboratory (JPL).
DE ephemerides are widely used in astronomy, space missions and space research due to their accuracy and reliability. They provide precise positions and velocities of solar system bodies, which are essential for a wide range of applications, including space missions, astronomical observations, and scientific research.

JPL ephemerides have evolved through several "DE" series: the 100, 200, and 400 series. The DE-1xx series, referenced to the B1950 ecliptic, is largely of historical interest. The DE-2xx series transitioned to the J2000 system. The DE-4xx series represents the most current and accurate ephemerides, referenced to the International Celestial Reference Frame (ICRF).

A summary of key DE versions is provided below:

* **DE19**: (Before 1969)  The then-current JPL Export Ephemeris before DE69.
* **DE69**: (1969) Third release of the JPL Ephemeris Tapes, a special purpose, short-duration ephemeris.
* **DE96**: (November 1975) One of six ephemerides produced between 1975 and 1982 using modern techniques.
* **DE111**: (May 1980) One of six ephemerides produced between 1975 and 1982 using modern techniques.
* **DE102**: (September 1981) B1950 ecliptic, nutations, no librations. Covers JED 1206160.5 to 2817872.5. First numerically integrated Long Ephemeris, covering 1141 BC to AD 3001.
* **DE118**: (September 1981) B1950 ecliptic. Rarely seen, DE-200 is considered a J2000 rotation of DE-118.
* **DE200**: (1982) J2000 ecliptic, nutations, no librations. Covers JED 2305424.5 to 2513360.5. Used in the Astronomical Almanac (1984-2002). Adopted as the fundamental ephemeris for new almanacs starting in 1984.
* **DE202**: (October 1987) J2000 ecliptic, nutations and librations. Covers JED 2414992.5 to 2469808.5.
* **DE402**: (Released in 1995) Introduced coordinates referred to the International Celestial Reference Frame (ICRF). Quickly superseded by DE403.
* **DE403**: (May 1993, released in 1995) ICRF, nutations and librations. Covers JED 2305200.5 to 2524400.5. Fit to planetary and lunar laser ranging data. Included perturbations of 300 asteroids.
* **DE404**: (Released in 1996) Long Ephemeris, condensed version of DE403. Covered 3000 BC to AD 3000. Reduced accuracy, no nutation or libration.
* **DE405**: (Created May 1997, released 1998) ICRF, nutations and librations. Covers JED 2305424.50 to 2525008.50. Utilized in the Astronomical Almanac from 2003 until 2014.
* **DE406**: (Created May 1997, released 1998) ICRF, no nutations or librations. Spans JED 0625360.5 to 2816912.50. Long time span, reduced polynomial accuracy for smaller file size. Condensed version of DE405, covering 3000 BC to AD 3000.
* **DE409**: (Released in 2003) Improvements over DE405, especially for Mars and Saturn positions. Covers years 1901 to 2019.
* **DE410**: (April 2003, released in 2003) ICRF, nutations and librations. Covers JED 2415056.5 to 2458832.5. Used for Mars Exploration Rover navigation. Improvements in planetary masses. Covers 1901 - 2019.
* **DE411**: (Widely cited, not publicly released, before 2004)
* **DE412**: (Widely cited, not publicly released, before 2010)
* **DE413**: (November 2004, released in 2004) ICRF, nutations and librations. Covers JED 2414992.5 to 2469872.5. Pluto orbit update for Charon occultation planning.
* **DE408**: (Created in 2005, unreleased) Longer version of DE406, covering 20,000 years.
* **DE414**: (May 2005, released in 2006) ICRF, nutations and librations. Covers JED 2414992.5 to 2469872.5. Fit to MGS and Odyssey ranging data through 2005.
* **DE418**: (August 2007, released in 2007) ICRF, nutations and librations. Covers JED 2414864.5 to 2470192.5. For planning New Horizons mission. Improved lunar orbit and librations with lunar laser ranging data.
* **DE421**: (Feb 2008, released in 2008) ICRF, nutations and librations. Covers JED 2414864.5 to 2471184.5 (extended to 2200 later). Fit to planetary and lunar laser ranging data, Venus Express data.
* **DE422**: (September 2009, created in 2009) ICRF, nutations and librations. Covers JED 625648.5 to 2816816.5. Intended for MESSENGER mission, extended time range, successor to DE406. Long Ephemeris, covering 3000 BC to AD 3000.
* **DE423**: (February 2010, released in 2010) ICRF version 2.0, nutations and librations. Covers JED 2378480.5 to 2524624.5. Intended for MESSENGER mission. Fit to MESSENGER and Venus Express data.
* **DE424**: (Created in 2011) For Mars Science Laboratory mission support.
* **DE430**: (April 2013, created in 2013) ICRF version 2.0, librations, 1980 nutation. Covers JED 2287184.5 to 2688976.5. DE430t variant includes TT-TDB data. Utilized in the Astronomical Almanac from 2015 onwards.
* **DE431**: (April 2013, created in 2013) ICRF version 2.0, librations, 1980 nutation. Covers JED -0.3100015.5 to 8000016.5. Very long time span, large file size (2.8 GB). For historical observations.
* **DE432**: (April 2014, created April 2014) ICRF version 2.0, librations, no nutations. Covers JED 2287184.5 to 2688976.5. Update to DE430, intended for New Horizons Pluto targeting. DE432t variant includes TT-TDB data.
* **DE434**: (July 2016, created November 2015) ICRF version 2.0, librations and nutations. Covers JED 2287184.5 to 2688976.5. Improved Jupiter ephemeris for the Juno mission.
* **DE435**: (July 2016, created January 2016) ICRF version 2.0, librations and nutations. Covers JED 2287184.5 to 2688976.5. Improved Saturn ephemeris for the Cassini mission.
* **DE436**: (Created in 2016) Based on DE430, with improved Jupiter data for Juno mission.
* **DE433**: (July 2016) ICRF version 2.0, librations and nutations. Covers JED 2287184.5 to 2688976.5.
* **DE438**: (Created in 2018) Based on DE430, with improved data for Mercury, Mars, and Jupiter (MESSENGER, Mars Odyssey, MRO, Juno missions).
* **DE440**: (February 2022, created June 2020) ICRF version 2.0, librations and nutations. Covers JED 2287184.5 to 2688976.5. Adds about seven years of new data. Includes improved orbits for Jupiter, Saturn, and Pluto.
* **DE441**: (Created June 2020) ICRF version 2.0, librations and nutations. Covers JED -3100015.5 to 8000016.5. Very long time span, large file size (2.6 GB). Useful for historical observations outside DE440 span.

**DE440 vs DE441:**

*   **DE440:** The latest JPL ephemeris with a fully consistent treatment of planetary and lunar laser ranging data. It includes a frictional damping term between the lunar fluid core and elastic mantle, which makes it highly accurate for the time range it covers. However, this damping term limits its extrapolation accuracy beyond a few centuries from the present.
*   **DE441:** Integrated without the lunar core/mantle damping term to provide accurate positions over a much longer time span. For planets, DE441 agrees with DE440 to within one meter over the time covered by DE440. For the Moon, differences are more significant over long periods, mainly in the along-track position, which can differ by ~10 meters 100 years from the present, growing quadratically further in the past or future.

For most applications requiring the highest accuracy within the DE440 time range, **DE440 is recommended**. For applications needing a very long time span and where slight differences in lunar position over very long times are acceptable, **DE441 is the better choice**. If compatibility with older software is a concern or for less demanding applications, DE-405 or DE-406 might be sufficient.

**Note:**  Both DE440 and DE441 are referred to the International Celestial Reference Frame version 3.0, offering the latest and most accurate celestial reference system. Some recent ephemerides (DE-430t, DE-432t, and later) also include TT-TDB time scale transformation data, which is handled by this Go library if present.

## [Downloading DE Files](#finding-de-files)

JPL ephemeris files can be downloaded from the [JPL ftp site](https://ssd.jpl.nasa.gov/ftp/eph/planets/). The `Linux` directory contains little-Endian versions suitable for most PCs. The `SunOS` directory contains big-Endian versions.

**Note:** This Go library can handle both endianness files regardless of your system's architecture. You can directly use the `Linux` versions; the library will automatically detect and handle byte order differences if needed.

The JPL ftp site also provides ASCII versions of the ephemerides, test data, and inter-office memoranda describing each DE version.  For users interested in converting ASCII ephemerides to binary format or understanding the original Fortran tools, the `fortran` directory contains the `asc2eph.f` program for ASCII to binary conversion, and `testeph.f` for testing binary ephemeris files after conversion.  The `ascii` directory contains the ASCII ephemeris files and necessary header files.

**Direct Access via FTP:**

*   **Binary Ephemerides (Little-Endian):**  [ftp://ssd.jpl.nasa.gov/pub/eph/planets/Linux/](https://ssd.jpl.nasa.gov/ftp/eph/planets/Linux/) (Recommended for most users)
*   **Binary Ephemerides (Big-Endian):** [ftp://ssd.jpl.nasa.gov/pub/eph/planets/SunOS/](https://ssd.jpl.nasa.gov/ftp/eph/planets/SunOS/)
*   **ASCII Ephemerides:** [ftp://ssd.jpl.nasa.gov/pub/eph/planets/ascii/](https://ssd.jpl.nasa.gov/ftp/eph/planets/ascii/) (Requires conversion to binary format using tools like `asc2eph.f` from the `fortran` directory)
*   **Fortran Conversion and Reading Programs:** [ftp://ssd.jpl.nasa.gov/pub/eph/planets/fortran/](https://ssd.jpl.nasa.gov/ftp/eph/planets/fortran/) (For ASCII to binary conversion and binary reading, may require user tailoring as described in `userguide.txt`)
*   **ASCII Format Description:** [ftp://ssd.jpl.nasa.gov/pub/eph/planets/ascii/ascii_format.txt](https://ssd.jpl.nasa.gov/ftp/eph/planets/ascii/ascii_format.txt)
*   **Fortran Programs User Guide:** [ftp://ssd.jpl.nasa.gov/pub/eph/planets/fortran/userguide.txt](https://ssd.jpl.nasa.gov/ftp/eph/planets/fortran/userguide.txt)
*   **List of Other Readers (Non-JPL):** [ftp://ssd.jpl.nasa.gov/pub/eph/planets/other_readers.txt](https://ssd.jpl.nasa.gov/ftp/eph/planets/other_readers.txt)

For general use, especially with this Go library, downloading directly from the JPL ftp site is the most convenient method to obtain the latest pre-built binary ephemerides like DE-44x series from the `Linux` directory.

## [Other Implementations)](#other-implementations)

For context, here are some other language implementations for JPL ephemerides:

- **C/C++:** The original implementation by Bill Gray and Piotr Dybczyński is available on [GitHub](https://github.com/Bill-Gray/jpl_eph).
- **FORTRAN:** The official JPL FORTRAN source, including programs like `asc2eph.f` and `testeph.f`, is available on the JPL FTP site, though less suited for modern applications.
- **Java:** JPL also provides a Java version.

jpleph is unique in that it offers a fully native Go solution for these calculations.

## [Why this Go library exists](#why-this-go-library-exists)

Existing C/C++, FORTRAN, and Java implementations of DE ephemeris readers are available. However, a native Go implementation offers several advantages within the Go ecosystem:

* **Go Idiomatic:** This library is written in idiomatic Go, leveraging Go's concurrency, error handling, and package management features.
* **Performance:** Go provides excellent performance, often comparable to C/C++, while offering memory safety and ease of development.
* **Cross-platform:** Go's cross-compilation capabilities ensure that this library can be easily used on various platforms without complex build processes.
* **Integration:**  A pure-Go library simplifies integration into Go-based astronomy software, web services, and other applications.
* **Modernization:** Provides a modern, actively maintained alternative to older language implementations for Go developers.

This Go library aims to make JPL DE ephemeris data more accessible and usable within the growing Go programming community, particularly in fields like astronomy, space exploration, and related scientific computing domains.

## [Reasons to use JPL DE rather than analytic series](#reasons-to-use-jpl-de-rather-than-analytic-series)

While analytic series like VSOP-87, PS-1996, and ELP-82 offer a compact way to calculate planetary and lunar positions, JPL DE ephemerides provide superior accuracy and are based on the latest observational data.

* **Accuracy:** DE ephemerides are generated through numerical integration of highly sophisticated models of the solar system, fitted to a vast amount of observational data. They achieve much higher accuracy than analytic series, especially for long time spans.
* **Up-to-date:** JPL regularly updates DE ephemerides with new data, incorporating the latest observations from ground-based and space-based telescopes and missions. Analytic series are typically based on older DE versions and are not updated as frequently.
* **Time Range:**  While some analytic series have limited time ranges, DE ephemerides like DE-441 cover extremely long time spans (millennia), making them suitable for historical and long-term studies.
* **Speed:** Despite their complexity, accessing positions from DE ephemerides can be surprisingly fast, often faster than evaluating complex analytic series for high precision.

For applications requiring the highest possible accuracy, especially in modern astronomical research and space mission planning, JPL DE ephemerides are the preferred choice. This Go library makes it easy to leverage this accuracy within Go software.

## [Contributing](#contributing)

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests via the [repository](https://github.com/mshafiee/jpleph). For pull requests, please ensure your code adheres to Go coding standards and includes appropriate tests.

## [License](#license)

This project is licensed under the **GNU General Public License (GPL) Version 2 or later**. See the [LICENSE](LICENSE) file for the full license text.

## [Copyright/legal issues (it's under the GPL)](#copyrightlegal-issues-its-under-the-gpl)

This Go library is released under the **GNU General Public License (GPL) Version 2 or later**. This means:

* You are free to use, modify, and distribute this software, even for commercial purposes.
* If you distribute modified versions, you must also release your modifications under the GPL.
* The software is provided "as is" without warranty.

A copy of the GPL license should be included with the source code. Please review the license terms carefully before using or distributing this software.

## [Changelog](#changelog)

- **January 30, 2025:**  
  - Initial Go implementation and documentation.
  - Core ephemeris reading and position/velocity calculation functions implemented.
  - Supports multiple DE versions including DE405, DE406, DE422, DE430, DE431, DE432, DE435, DE440, and DE441.
  - Automatic endianness handling and basic error handling provided.
  - Example usage demonstrated.

*Future updates will be documented here.*
