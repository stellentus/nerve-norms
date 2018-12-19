package mem

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ExcitabilityVariables struct {
	Values          map[string]float64
	Program         string
	ThresholdMethod int
	SRMethod        int
}

func (exciteVar *ExcitabilityVariables) Parse(reader *Reader) error {
	// Find settings
	var err error
	exciteVar.Program, err = reader.ReadLineExtractingString(`^Program = (.*)`)
	if err != nil {
		return err
	}

	val, err := reader.ReadLineExtractingString(`^Threshold method = (\d+).*`)
	if err != nil {
		return err
	}
	exciteVar.ThresholdMethod, err = strconv.Atoi(val)
	if err != nil {
		return err
	}

	val, err = reader.ReadLineExtractingString(`^SR method = (\d+).*`)
	if err != nil {
		return err
	}
	exciteVar.SRMethod, err = strconv.Atoi(val)
	if err != nil {
		return err
	}

	// Read the main variables
	err = reader.parseLines(exciteVar)
	if err != nil {
		return err
	}

	// Now find any extra variables
	str, err := reader.skipNewlines()
	if err != nil {
		return err
	}

	if strings.Contains(str, "EXTRA VARIABLES") {
		err = reader.parseLines(&ExtraVariables{exciteVar})
	} else {
		// It looks like this header doesn't belong to us, so give it back
		reader.UnreadString(str)
	}

	return err
}

func (ev ExcitabilityVariables) String() string {
	return fmt.Sprintf("ExcitabilityVariables{%d values}", len(ev.Values))
}

func (exciteVar ExcitabilityVariables) ParseRegex() *regexp.Regexp {
	return regexp.MustCompile(`^ \d+\.\s+([-+]?\d*\.?\d+)\s+(.+)`)
}

func (exciteVar *ExcitabilityVariables) ParseLine(result []string) error {
	if len(result) != 3 {
		return errors.New("Incorrect ExVar line length")
	}

	val, err := strconv.ParseFloat(result[1], 64)
	if err != nil {
		return err
	}

	exciteVar.Values[result[2]] = val

	return nil
}

type ExtraVariables struct {
	*ExcitabilityVariables
}

func (extraVar ExtraVariables) ParseRegex() *regexp.Regexp {
	return regexp.MustCompile(`^(.+) = ([-+]?\d*\.?\d+)`)
}

func (extraVar *ExtraVariables) ParseLine(result []string) error {
	if len(result) != 3 {
		return errors.New("Incorrect ExtraVar line length")
	}

	val, err := strconv.ParseFloat(result[2], 64)
	if err != nil {
		return err
	}

	extraVar.Values[result[1]] = val

	return nil
}
