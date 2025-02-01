package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mshafiee/jpleph"
)

const nMasses = 16

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "'masses' takes the name of a JPL DE file as a command-line argument.\n")
		fmt.Fprintf(os.Stderr, "It will output a list of planetary masses in a table of the sort found\n")
		fmt.Fprintf(os.Stderr, "at the end of 'masses.go' (q.v.).\n")
		os.Exit(-1)
	}

	p, err := jpleph.NewEphemeris(os.Args[1], true)
	if err != nil {
		fmt.Printf("JPL data not loaded from '%s'\n", os.Args[1])
		fmt.Printf("Error: %v\n", err)
		os.Exit(-1)
	}
	defer p.Close()

	nConstants := int(p.GetEphemerisLong(jpleph.NumberOfConstants))
	masses := make([]float64, nMasses)
	names := [nMasses]string{"Sun ", "Merc", "Venu", "EMB ", "Mars",
		"Jupi", "Satu", "Uran", "Nept", "Plut", "Eart", "Moon",
		"Cere", "Pall", "Juno", "Vest"}

	var emrat, au_in_km float64 = 0.0, 0.0
	var gmb float64 = 0.0

	for i := 0; i < nConstants; i++ {
		constantName, err := p.GetConstantName(i)
		if err != nil {
			continue
		}
		ephemConstant, err := p.GetConstantValue(i)
		if err != nil {
			continue
		}

		// Check for GM constants (e.g., GMS, GMB, GM1)
		if len(constantName) >= 4 && strings.HasPrefix(constantName, "GM") && (constantName[3] == ' ' || len(constantName) == 3) {
			switch constantName[2] {
			case 'B':
				gmb = ephemConstant
			case 'S':
				masses[0] = ephemConstant
			case '1', '2', '4', '5', '6', '7', '8', '9':
				index, _ := strconv.Atoi(string(constantName[2]))
				masses[index] = ephemConstant
			}
		}

		// Trim spaces for comparison
		trimmedName := strings.TrimSpace(constantName)
		if trimmedName == "EMRAT" {
			emrat = ephemConstant
		} else if trimmedName == "AU" {
			au_in_km = ephemConstant
		}

		// Handle MA000x constants (Ceres, Pallas, Juno, Vesta)
		if len(constantName) >= 6 && strings.HasPrefix(constantName, "MA000") {
			idxStr := constantName[5:6]
			idx, _ := strconv.Atoi(idxStr)
			if idx >= 1 && idx <= 4 {
				masses[idx+11] = ephemConstant
			}
		}
	}

	// Correct Earth and Moon masses based on EMRAT
	masses[3] = gmb                // EMB
	masses[11] = gmb / (1 + emrat) // Moon
	masses[10] = gmb - masses[11]  // Earth

	seconds_per_day := 86400.0
	fmt.Printf("Data from %s\n", os.Args[1])
	fmt.Printf("%5s %21s %18s %19s %20s %20s\n",
		"Body",
		"mass(obj)/mass(sun)",
		"mass(sun)/mass(obj)",
		"GM (km³/s²)",
		"GM (AU³/day²)",
		"mass(obj)",
	)

	for i := 0; i < nMasses; i++ {
		massRatioSun := masses[i] / masses[0]
		sunRatioMass := masses[0] / masses[i]
		gmKM := masses[i] * au_in_km * au_in_km * au_in_km / (seconds_per_day * seconds_per_day)
		gmAU := (masses[i] * seconds_per_day * seconds_per_day) / (au_in_km * au_in_km * au_in_km)

		fmt.Printf("%5s %21.15e %21.15e %21.15e %21.15e %21.15e\n",
			names[i],
			massRatioSun,
			sunRatioMass,
			gmKM,
			gmAU,
			masses[i],
		)
	}
	os.Exit(0)
}

/*
Data from ./lnxm13000p17000.431
 Body   mass(obj)/mass(sun) mass(sun)/mass(obj)         GM (km³/s²)        GM (AU³/day²)            mass(obj)
 Sun  1.000000000000000e+00 1.000000000000000e+00 1.327124400419394e+11 6.598027659259623e-19 2.959122082855911e-04
 Merc 1.660114153054349e-07 6.023682155592479e+06 2.203178000000002e+04 1.095347909938096e-25 4.912480450364760e-11
 Venu 2.447838287784772e-06 4.085237186582997e+05 3.248585920000000e+05 1.615090472819864e-24 7.243452332644120e-10
 EMB  3.040432648022641e-06 3.289005598102475e+05 4.035032355022598e+05 2.006085870776937e-24 8.997011390199871e-10
 Mars 3.227156037554997e-07 3.098703590290707e+06 4.282837521400001e+04 2.129286479653456e-25 9.549548695550771e-11
 Jupi 9.547919152112403e-04 1.047348625463337e+03 1.267127648000002e+08 6.299743465401234e-22 2.825345840833870e-07
 Satu 2.858856727222417e-04 3.497901767786633e+03 3.794058520000000e+07 1.886281576007395e-22 8.459706073245031e-08
 Uran 4.366243735831270e-05 2.290298161308703e+04 5.794548600000009e+06 2.880859693608379e-23 1.292024825782960e-08
 Nept 5.151383772628674e-05 1.941225977597307e+04 6.836527100580023e+06 3.398897261526518e-23 1.524357347885110e-08
 Plut 7.361781606089468e-09 1.358366837686175e+08 9.770000000000007e+02 4.857323865840705e-27 2.178441051974180e-12
 Eart 3.003489614915764e-06 3.329460488339481e+05 3.986004354360960e+05 1.981710755351325e-24 8.887692445125634e-10
 Moon 3.694303310687700e-08 2.706870324120324e+07 4.902800066163795e+03 2.437511542561185e-26 1.093189450742367e-11
 Cere 4.732743418347629e-10 2.112939391819251e+09 6.280939271413429e+01 3.122677197843660e-28 1.400476556172344e-13
 Pall 1.049111226915838e-10 9.531877787065401e+09 1.392301107993935e+01 6.922064892830496e-29 3.104448198938713e-14
 Juno 1.222503910232921e-11 8.179932936242897e+10 1.622414768878230e+00 8.066114613269857e-30 3.617538317147937e-15
 Vest 1.302666831538261e-10 7.676559929135098e+09 1.728800937751447e+01 8.595031785289547e-29 3.854750187808810e-14
*/
