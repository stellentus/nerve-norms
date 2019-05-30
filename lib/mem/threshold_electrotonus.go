package mem

import (
	"encoding/json"
	"errors"
	"fmt"
)

type TEPair struct {
	Delay           Column
	ThreshReduction Column
	WasImputed      Column
}

type ThresholdElectrotonus struct {
	Hyperpol40 TEPair
	Hyperpol20 TEPair
	Depol40    TEPair
	Depol20    TEPair
	Depol70    TEPair
	Depol100   TEPair
}

func (te *ThresholdElectrotonus) GetPair(name string) TEPair {
	switch name {
	case "h40":
		return te.Hyperpol40
	case "h20":
		return te.Hyperpol20
	case "d40":
		return te.Depol40
	case "d20":
		return te.Depol20
	case "d70":
		return te.Depol70
	case "d100":
		return te.Depol100
	default:
		return TEPair{}
	}
}

var TEDelay = Column([]float64{0, 9, 10, 11, 15, 20, 26, 30, 33, 30, 41, 50, 60, 70, 80, 90, 100, 109, 110, 111, 120, 130, 140, 150, 160, 170, 180, 190, 200, 210})

func (te *ThresholdElectrotonus) LoadFromMem(mem *rawMem) error {
	sec, err := mem.sectionContainingHeader("THRESHOLD ELECTROTONUS")
	if err != nil {
		return errors.New("Could not get threshold electrotonus: " + err.Error())
	}

	for i := range sec.Tables {
		pair := TEPair{Delay: TEDelay}

		delay, err := sec.columnContainsName("Delay (ms)", i)
		if err != nil {
			return errors.New("Could not get threshold electrotonus: " + err.Error())
		}

		pair.ThreshReduction, err = sec.columnContainsName("Thresh redn. (%)", i)
		if err != nil {
			return errors.New("Could not get threshold electrotonus: " + err.Error())
		}

		pair.WasImputed = pair.ThreshReduction.ImputeWithValue(delay, pair.Delay, 0.01, false)

		current, err := sec.columnContainsName("Current (%)", i)
		if err != nil {
			return errors.New("Could not get threshold electrotonus: " + err.Error())
		}

		// This is a quick and simple way to parse the data we expect to see
		max := current.Maximum()
		min := current.Minimum()
		switch {
		case max > 34 && max < 46 && min == 0:
			te.Hyperpol40 = pair
		case max > 14 && max < 26 && min == 0:
			te.Hyperpol20 = pair
		case max == 0 && min < -34 && min > -46:
			te.Depol40 = pair
		case max == 0 && min < -14 && min > -26:
			te.Depol20 = pair
		case max == 0 && min < -64 && min > -76:
			te.Depol70 = pair
		case max == 0 && min < -94 && min > -106:
			te.Depol100 = pair
		default:
			fmt.Printf("TE contained unexpected current [%f, %f]\n", min, max)
		}
	}

	return nil
}

type jsonThresholdElectrotonus struct {
	Columns []string         `json:"columns"`
	Data    map[string]Table `json:"data"`
}

func (dat *ThresholdElectrotonus) MarshalJSON() ([]byte, error) {
	str := &jsonThresholdElectrotonus{
		Columns: []string{"Delay (ms)", "Threshold Reduction (%)"},
		Data: map[string]Table{
			"h40": []Column{dat.Hyperpol40.Delay, dat.Hyperpol40.ThreshReduction},
			"h20": []Column{dat.Hyperpol20.Delay, dat.Hyperpol20.ThreshReduction},
			"d40": []Column{dat.Depol40.Delay, dat.Depol40.ThreshReduction},
			"d20": []Column{dat.Depol20.Delay, dat.Depol20.ThreshReduction},
		},
	}

	if dat.Hyperpol40.WasImputed != nil || dat.Hyperpol20.WasImputed != nil || dat.Depol40.WasImputed != nil || dat.Depol20.WasImputed != nil {
		str.Columns = append(str.Columns, "Was Imputed")
		str.Data["h40"] = append(str.Data["h40"], dat.Hyperpol40.WasImputed)
		str.Data["h20"] = append(str.Data["h20"], dat.Hyperpol20.WasImputed)
		str.Data["d40"] = append(str.Data["d40"], dat.Depol40.WasImputed)
		str.Data["d20"] = append(str.Data["d20"], dat.Depol20.WasImputed)
	}

	return json.Marshal(&str)
}

func (dat *ThresholdElectrotonus) UnmarshalJSON(value []byte) error {
	jsDat := jsonThresholdElectrotonus{}
	err := json.Unmarshal(value, &jsDat)
	if err != nil {
		return err
	}
	numCol := len(jsDat.Columns)

	if numCol < 2 || numCol > 3 {
		return errors.New("Incorrect number of ThresholdElectrotonus columns in JSON")
	}
	if jsDat.Columns[0] != "Delay (ms)" || jsDat.Columns[1] != "Threshold Reduction (%)" || (numCol == 3 && jsDat.Columns[2] != "Was Imputed") {
		return errors.New("Incorrect ThresholdElectrotonus column names in JSON")
	}

	dat.Hyperpol40.fromTable(jsDat.Data["h40"])
	dat.Hyperpol20.fromTable(jsDat.Data["h20"])
	dat.Depol40.fromTable(jsDat.Data["d40"])
	dat.Depol20.fromTable(jsDat.Data["d20"])

	return nil
}

func (tep *TEPair) fromTable(tab Table) {
	tep.Delay = tab[0]
	tep.ThreshReduction = tab[1]
	if len(tab) == 3 {
		tep.WasImputed = tab[2]
	}
}
