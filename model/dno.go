package model

type Dno string

const (
	DnoEnergyNorthWest          Dno = "EnergyNorthWest"
	DnoNationalGridDistribution Dno = "NationalGridDistribution"
	DnoNorthernPowergrid        Dno = "NorthernPowergrid"
	DnoSPEnergy                 Dno = "SPEnergy"
	DnoSse                      Dno = "SSE"
	DnoUKPowerNetwork           Dno = "UKPowerNetwork"
)

var AllDnoList = [6]Dno{
	DnoEnergyNorthWest,
	DnoNationalGridDistribution,
	DnoNorthernPowergrid,
	DnoSPEnergy,
	DnoSse,
	DnoUKPowerNetwork,
}
