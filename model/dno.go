package model

import "slices"

type Dno string

const (
	DnoEnergyNorthWest          Dno = "EnergyNorthWest"
	DnoNationalGridDistribution Dno = "NationalGridDistribution"
	DnoNorthernPowergrid        Dno = "NorthernPowergrid"
	DnoSPEnergy                 Dno = "SPEnergy"
	DnoSse                      Dno = "SSE"
	DnoUKPowerNetwork           Dno = "UKPowerNetwork"
)

var AllDnoList = []Dno{
	DnoEnergyNorthWest,
	DnoNationalGridDistribution,
	DnoNorthernPowergrid,
	DnoSPEnergy,
	DnoSse,
	DnoUKPowerNetwork,
}

func (dno Dno) isValid() bool {
	return slices.Contains(AllDnoList, dno)
}
