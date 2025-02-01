// ./cmd/jpleph_test/main.go
package main

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
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/mshafiee/jpleph"
)

// printResult is a helper function to format and print position results.
func printResult(label string, pos jpleph.Position) {
	fmt.Printf("  %s: [%12.5e, %12.5e, %12.5e]\n",
		label, pos.X, pos.Y, pos.Z)
}

// printVelocityResult is a helper function to format and print velocity results.
func printVelocityResult(label string, vel jpleph.Velocity) {
	fmt.Printf("  %s: [%12.5e, %12.5e, %12.5e]\n",
		label, vel.DX, vel.DY, vel.DZ)
}

// magnitude calculates the magnitude of a Position vector.
func magnitude(pos jpleph.Position) float64 {
	return math.Sqrt(pos.X*pos.X + pos.Y*pos.Y + pos.Z*pos.Z)
}

// testBody performs tests for a specific celestial body.
func testBody(eph *jpleph.Ephemeris, et float64, body jpleph.Planet, bodyName string) {
	fmt.Printf("\nTesting %s:\n", bodyName)

	pos, vel, err := eph.CalculatePV(et, body, jpleph.CenterSolarSystemBarycenter, true)
	if err != nil {
		fmt.Printf("  Error calculating barycentric position: %v\n", err)
		return
	}
	printResult("Barycentric Position (AU)", pos)
	printVelocityResult("Barycentric Velocity (AU/day)", vel)

	if body >= jpleph.Mercury && body <= jpleph.Pluto && body != jpleph.Sun {
		pos, _, err = eph.CalculatePV(et, body, jpleph.CenterSun, true) // Corrected: Using jpleph.CenterSun
		if err == nil {
			printResult("Heliocentric Position (AU)", pos)
		} else {
			fmt.Printf("  Error calculating heliocentric position: %v\n", err)
		}
	}
}

// testSpecial performs tests for special quantities.
func testSpecial(eph *jpleph.Ephemeris, et float64, spec jpleph.Planet, specName string) {
	fmt.Printf("\nTesting %s:\n", specName)

	pos, vel, err := eph.CalculatePV(et, spec, jpleph.CenterSun, true)
	if err != nil {
		fmt.Printf("  Error calculating %s: %v\n", specName, err)
		return
	}

	switch spec {
	case jpleph.Nutations:
		fmt.Printf("  Δψ: %.5e rad\n", pos.X)
		fmt.Printf("  Δε: %.5e rad\n", pos.Y)
		fmt.Printf("  dΔψ/dt: %.5e rad/day\n", vel.DX)
		fmt.Printf("  dΔε/dt: %.5e rad/day\n", vel.DY)

	case jpleph.Librations:
		printResult("Libration Angles (rad)", pos)
		printVelocityResult("Angular Rates (rad/day)", vel)
	case jpleph.LunarMantleOmega:
		printVelocityResult("Angular Velocity (rad/day)", vel)
	case jpleph.TT_TDB:
		fmt.Printf("  TT-TDB: %.5e seconds\n", pos.X*86400)
	}
}

// testEarthMoonSystem performs tests specific to the Earth-Moon system.
func testEarthMoonSystem(eph *jpleph.Ephemeris, et float64) {
	au := eph.GetEphemerisDouble(jpleph.AUinKM)
	var embPos, earthEmbPos, moonEmbPos, earthMoonPos jpleph.Position
	var err error

	fmt.Printf("\n=== Earth-Moon System ===\n")

	embPos, _, err = eph.CalculatePV(et, jpleph.EarthMoonBarycenter, jpleph.CenterSolarSystemBarycenter, true)
	if err != nil {
		fmt.Printf("  Error calculating EMB position: %v\n", err)
		return
	}
	printResult("EMB Position (AU)", embPos)

	earthEmbPos, _, err = eph.CalculatePV(et, jpleph.Earth, jpleph.CenterEarthMoonBarycenter, true)
	if err != nil {
		fmt.Printf("  Error calculating Earth relative to EMB: %v\n", err)
		return
	}
	printResult("Earth relative to EMB (AU)", earthEmbPos)

	moonEmbPos, _, err = eph.CalculatePV(et, jpleph.Moon, jpleph.CenterEarthMoonBarycenter, true)
	if err != nil {
		fmt.Printf("  Error calculating Moon relative to EMB: %v\n", err)
		return
	}
	printResult("Moon relative to EMB (AU)", moonEmbPos)

	earthMoonPos, _, err = eph.CalculatePV(et, jpleph.Moon, jpleph.CenterEarthMoonBarycenter, true)
	if err != nil {
		fmt.Printf("  Error calculating Earth-Moon relative position: %v\n", err)
		return
	}
	distance := magnitude(earthMoonPos)
	fmt.Printf("\nEarth-Moon Distance: %.5f AU (%.3f km)\n",
		distance, distance*au)
}

// testConstants tests constant name and value retrieval.
func testConstants(eph *jpleph.Ephemeris) {
	fmt.Printf("\n=== Constant Tests ===\n")
	numConstants := int(eph.GetEphemerisLong(jpleph.NumberOfConstants))
	if numConstants <= 0 {
		fmt.Println("  Warning: No constants to test.")
		return
	}

	fmt.Printf("  Testing retrieval of %d constants...\n", numConstants)
	for i := 0; i < numConstants; i++ {
		name, err := eph.GetConstantName(i)
		if err != nil {
			fmt.Printf("  Error getting constant name at index %d: %v\n", i, err)
			continue
		}
		value, err := eph.GetConstantValue(i)
		if err != nil {
			fmt.Printf("  Error getting constant value at index %d: %v\n", i, err)
			continue
		}
		if i < 10 { // Print only first 10 for brevity
			fmt.Printf("  Constant %d: Name='%s', Value=%.2f\n", i, name, value)
		}
	}
	fmt.Println("  Constant tests finished.")
}

// main is the entry point of the jpleph test program.
func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <path_to_ephemeris_file>\n", os.Args[0])
		os.Exit(1)
	}

	ephemFilename := os.Args[1]
	if !filepath.IsAbs(ephemFilename) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current working directory:", err)
			os.Exit(1)
		}
		ephemFilename = filepath.Join(cwd, ephemFilename)
	}

	// Initialize the JPL ephemeris using the wrapper, loading constants
	eph, err := jpleph.NewEphemeris(ephemFilename, true) // loadConstants is true now
	if err != nil {
		fmt.Printf("Failed to open ephemeris: %v\n", err)
		os.Exit(1)
	}
	defer eph.Close() // Use Close method on Ephemeris instance

	fmt.Printf("=== Ephemeris Header Tests ===\n")
	fmt.Printf("Ephemeris name: %s\n", eph.GetEphemName()) // Use GetEphemName method
	fmt.Printf("Time range: %.1f to %.1f JD (step %.2f days)\n",
		eph.GetEphemerisDouble(jpleph.EphemerisStartJD), // Use GetEphemerisDouble method
		eph.GetEphemerisDouble(jpleph.EphemerisEndJD),
		eph.GetEphemerisDouble(jpleph.EphemerisStep))
	fmt.Printf("AU value: %.8f km\n", eph.GetEphemerisDouble(jpleph.AUinKM))
	fmt.Printf("Earth-Moon ratio: %.5f\n", eph.GetEphemerisDouble(jpleph.EarthMoonMassRatio))

	const testTime = 2451545.0
	fmt.Printf("\nTesting positions at JD %.3f\n", testTime)

	bodies := []struct {
		Planet jpleph.Planet
		Name   string
	}{
		{jpleph.Mercury, "Mercury"}, {jpleph.Venus, "Venus"}, {jpleph.Earth, "Earth"},
		{jpleph.Mars, "Mars"}, {jpleph.Jupiter, "Jupiter"}, {jpleph.Saturn, "Saturn"},
		{jpleph.Uranus, "Uranus"}, {jpleph.Neptune, "Neptune"}, {jpleph.Pluto, "Pluto"},
		{jpleph.Moon, "Moon"}, {jpleph.Sun, "Sun"}, {jpleph.SolarSystemBarycenter, "Solar System Barycenter"},
		{jpleph.EarthMoonBarycenter, "Earth-Moon Barycenter"},
	}

	fmt.Printf("\n=== Standard Celestial Bodies ===\n")
	for _, body := range bodies {
		testBody(eph, testTime, body.Planet, body.Name) // Pass Ephemeris instance
	}

	specials := []struct {
		Planet jpleph.Planet
		Name   string
	}{
		{jpleph.Nutations, "Nutations"},
		{jpleph.Librations, "Librations"},
		{jpleph.LunarMantleOmega, "Lunar Mantle Rotation"},
		{jpleph.TT_TDB, "TT-TDB"},
	}

	fmt.Printf("\n=== Special Quantities ===\n")
	for _, special := range specials {
		testSpecial(eph, testTime, special.Planet, special.Name) // Pass Ephemeris instance
	}

	testEarthMoonSystem(eph, testTime) // Pass Ephemeris instance

	testConstants(eph) // Test constant retrieval

	fmt.Println("Program finished successfully.")
}
