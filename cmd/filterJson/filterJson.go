package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"gogs.bellstone.ca/james/jitter/lib/mef"
	"gogs.bellstone.ca/james/jitter/lib/mem"
)

var input = flag.String("input", "json/all.json", "path to the JSON that should be loaded")
var output = flag.String("output", "", "path to save the filtered JSON; otherwise, do nothing with it")
var norm = flag.String("norm", "json/norm.json", "path to save the norm JSON; otherwise, output to stdout")
var sexString = flag.String("sex", "", "only include participants of this sex (M/F)")
var minAge = flag.Int("minAge", 0, "only include participants at least this old")
var maxAge = flag.Int("maxAge", 200, "only include participants this age or younger")
var country = flag.String("country", "", "only include participants from this country (CA/PO/JP)")
var species = flag.String("species", "human", "only include participants from this species (human/rat)")
var nerve = flag.String("nerve", "median", "only include participants from this nerve (median/CP)")

func main() {
	flag.Parse()

	if *minAge > *maxAge {
		panic("minAge > maxAge is not valid")
	}

	sex, err := parseSex(*sexString)
	if err != nil {
		panic(err)
	}

	file, err := os.Open(*input)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var mefData mef.Mef

	err = json.Unmarshal(bytes, &mefData)
	if err != nil {
		panic(err)
	}

	mefData.Filter(mef.NewFilter().BySex(sex).ByAge(*minAge, *maxAge).ByCountry(*country).BySpecies(*species).ByNerve(*nerve))

	jsArray, err := json.Marshal(&mefData)
	if err != nil {
		panic("Could not concatenate JSON due to error: " + err.Error())
	}

	if *output != "" {
		err = ioutil.WriteFile(*output, jsArray, 0644)
		if err != nil {
			panic("Could not save JSON due to error: " + err.Error())
		}
	}

	jsNorm := mefData.Norm()
	jsNormArray, err := json.Marshal(&jsNorm)
	if err != nil {
		panic("Could not create norm JSON due to error: " + err.Error())
	}

	if *norm == "" {
		fmt.Println(string(jsNormArray))
	} else {
		err = ioutil.WriteFile(*norm, jsNormArray, 0644)
		if err != nil {
			panic("Could not save norm JSON due to error: " + err.Error())
		}
	}
}

func parseSex(sex string) (mem.Sex, error) {
	switch sex {
	case "male", "Male", "M", "m":
		return mem.MaleSex, nil
	case "female", "Female", "F", "f":
		return mem.MaleSex, nil
	case "":
		return mem.UnknownSex, nil
	default:
		return mem.UnknownSex, errors.New("Invalid sex '" + sex + "'")
	}
}
