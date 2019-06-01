package mef

import (
	"math"

	"gogs.bellstone.ca/james/jitter/lib/mem"
)

type Matrix interface {
	NColumns() int
	NRows() int
	Column(int) mem.Column
	WasImputed(int) mem.Column
}

type LabelledTableFromMem func(*mem.Mem) *mem.LabelledTable

type GenericNorm struct {
	XValues mem.Column `json:"xvalues,omitempty"`
	MatNorm `json:"norms"`
	ltfm    LabelledTableFromMem
	mef     *Mef
}

func (norm GenericNorm) NColumns() int {
	return len(norm.mef.mems)
}

func (norm GenericNorm) NRows() int {
	return len(norm.ltfm(norm.mef.mems[0]).XColumn)
}

func (norm GenericNorm) Column(i int) mem.Column {
	return norm.ltfm(norm.mef.mems[i]).YColumn
}

func (norm GenericNorm) WasImputed(i int) mem.Column {
	return norm.ltfm(norm.mef.mems[i]).WasImputed
}

type MatNorm struct {
	Mean mem.Column `json:"mean"`
	SD   mem.Column `json:"sd"`
	Num  mem.Column `json:"num"`
}

func MatrixNorm(mat Matrix) MatNorm {
	numEl := mat.NRows()
	norm := MatNorm{
		Mean: make(mem.Column, numEl),
		SD:   make(mem.Column, numEl),
		Num:  make(mem.Column, numEl),
	}

	// Sum the values
	for colN := 0; colN < mat.NColumns(); colN++ {
		col := mat.Column(colN)
		for rowN := range col {
			if !wasImp(mat, colN, rowN) {
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
	for colN := 0; colN < mat.NColumns(); colN++ {
		col := mat.Column(colN)
		for rowN := range col {
			if !wasImp(mat, colN, rowN) {
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

func wasImp(mat Matrix, colN, rowN int) bool {
	col := mat.WasImputed(colN)
	// Yes, this is terrible, but wasImputed is a float column
	return len(col) != 0 && col[rowN] > 0.5
}
