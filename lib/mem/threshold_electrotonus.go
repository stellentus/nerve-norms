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
	Hyperpol40 *TEPair
	Hyperpol20 *TEPair
	Depol40    *TEPair
	Depol20    *TEPair
}

func (mem *Mem) ThresholdElectrotonus() (ThresholdElectrotonus, error) {
	te := ThresholdElectrotonus{}

	sec, err := mem.sectionContainingHeader("THRESHOLD ELECTROTONUS")
	if err != nil {
		return te, errors.New("Could not get threshold electrotonus: " + err.Error())
	}

	for i := range sec.Tables {
		pair := TEPair{Delay: Column([]float64{0, 9, 10, 11, 15, 20, 26, 33, 41, 50, 60, 70, 80, 90, 100, 109, 110, 111, 120, 130, 140, 150, 160, 170, 180, 190, 200, 210})}

		delay, err := sec.columnContainsName("Delay (ms)", i)
		if err != nil {
			return te, errors.New("Could not get threshold electrotonus: " + err.Error())
		}

		pair.ThreshReduction, err = sec.columnContainsName("Thresh redn. (%)", i)
		if err != nil {
			return te, errors.New("Could not get threshold electrotonus: " + err.Error())
		}

		pair.WasImputed = pair.ThreshReduction.ImputeWithValue(delay, pair.Delay, 0.01)

		current, err := sec.columnContainsName("Current (%)", i)
		if err != nil {
			return te, errors.New("Could not get threshold electrotonus: " + err.Error())
		}

		// This is a quick and simple way to parse the data we expect to see
		max := current.Maximum()
		min := current.Minimum()
		switch {
		case max > 34 && max < 46 && min == 0:
			te.Hyperpol40 = &pair
		case max > 14 && max < 26 && min == 0:
			te.Hyperpol20 = &pair
		case max == 0 && min < -34 && min > -46:
			te.Depol40 = &pair
		case max == 0 && min < -14 && min > -26:
			te.Depol20 = &pair
		default:
			fmt.Printf("TE contained unexpected current [%f, %f]\n", min, max)
		}
	}

	return te, nil
}

func (dat *ThresholdElectrotonus) MarshalJSON() ([]byte, error) {
	str := &struct {
		Columns []string         `json:"columns"`
		Data    map[string]Table `json:"data"`
	}{
		Columns: []string{"Delay (ms)", "Threshold Reduction (%)"},
		Data: map[string]Table{
			"Hyperpol40": []Column{dat.Hyperpol40.Delay, dat.Hyperpol40.ThreshReduction},
			"Hyperpol20": []Column{dat.Hyperpol20.Delay, dat.Hyperpol20.ThreshReduction},
			"Depol40":    []Column{dat.Depol40.Delay, dat.Depol40.ThreshReduction},
			"Depol20":    []Column{dat.Depol20.Delay, dat.Depol20.ThreshReduction},
		},
	}

	if dat.Hyperpol40.WasImputed != nil || dat.Hyperpol20.WasImputed != nil || dat.Depol40.WasImputed != nil || dat.Depol20.WasImputed != nil {
		str.Columns = append(str.Columns, "Was Imputed")
		str.Data["Hyperpol40"] = append(str.Data["Hyperpol40"], dat.Hyperpol40.WasImputed)
		str.Data["Hyperpol20"] = append(str.Data["Hyperpol20"], dat.Hyperpol20.WasImputed)
		str.Data["Depol40"] = append(str.Data["Depol40"], dat.Depol40.WasImputed)
		str.Data["Depol20"] = append(str.Data["Depol20"], dat.Depol20.WasImputed)
	}

	return json.Marshal(&str)
}
