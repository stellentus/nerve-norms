package mef

import (
	"math"

	"gogs.bellstone.ca/james/jitter/lib/mem"
)

type LabelledTableFromMem func(*mem.Mem) *mem.LabelledTable

type NormTable struct {
	XValues mem.Column `json:"xvalues,omitempty"`
	Mean    mem.Column `json:"mean"`
	SD      mem.Column `json:"sd"`
	Num     mem.Column `json:"num"`
}

func NewNormTable(xv mem.Column, mef *Mef, ltfm LabelledTableFromMem) NormTable {
	numEl := len(ltfm(mef.mems[0]).XColumn)
	norm := NormTable{
		XValues: xv,
		Mean:    make(mem.Column, numEl),
		SD:      make(mem.Column, numEl),
		Num:     make(mem.Column, numEl),
	}

	// Sum the values
	for _, mm := range mef.mems {
		lt := ltfm(mm)
		col := lt.YColumn
		for rowN := range col {
			if !norm.wasImp(lt, rowN) {
				norm.Mean[rowN] += col[rowN]
				norm.Num[rowN]++
			}
		}
	}

	// Normalize to get mean
	for rowN := range norm.Mean {
		norm.Mean[rowN] /= norm.Num[rowN]
	}

	// Calculate SD
	for _, mm := range mef.mems {
		lt := ltfm(mm)
		col := lt.YColumn
		for rowN := range col {
			if !norm.wasImp(lt, rowN) {
				norm.SD[rowN] += math.Pow(col[rowN]-norm.Mean[rowN], 2)
			}
		}
	}

	// Normalize to get SD
	for rowN := range norm.Mean {
		norm.SD[rowN] = math.Sqrt(norm.SD[rowN] / norm.Num[rowN])
	}

	return norm
}

func (norm *NormTable) wasImp(lt *mem.LabelledTable, rowN int) bool {
	col := lt.WasImputed
	// Yes, this is terrible, but wasImputed is a float column
	return len(col) != 0 && col[rowN] > 0.5
}