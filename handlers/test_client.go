package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/k-lomer/lights-out/clients"
	"github.com/k-lomer/lights-out/model"
)

type TestDnoClient struct {
	Dno        model.Dno
	NumOutages int
}

func NewTestDnoClient(dno model.Dno, numOutages int) TestDnoClient {
	return TestDnoClient{
		dno,
		numOutages,
	}
}

func (t TestDnoClient) GetDno() model.Dno {
	return t.Dno
}

func (t TestDnoClient) ListOutages(ctx context.Context) ([]model.Outage, error) {
	outages := make([]model.Outage, 0, t.NumOutages)
	for i := range t.NumOutages {
		end := time.Now().Add(24 * time.Hour)
		o := model.Outage{
			DNO:       t.Dno,
			ID:        strconv.Itoa(i),
			Start:     time.Now(),
			End:       &end,
			Postcodes: []string{"WC2N 5EH"},
		}
		outages = append(outages, o)
	}

	return outages, nil
}

func NewTestDnoClients() map[model.Dno]clients.DnoClient {
	numOutages := 10
	return map[model.Dno]clients.DnoClient{
		model.DnoEnergyNorthWest:          NewTestDnoClient(model.DnoEnergyNorthWest, numOutages),
		model.DnoNationalGridDistribution: NewTestDnoClient(model.DnoNationalGridDistribution, numOutages),
		model.DnoNorthernPowergrid:        NewTestDnoClient(model.DnoNorthernPowergrid, numOutages),
		model.DnoSPEnergy:                 NewTestDnoClient(model.DnoSPEnergy, numOutages),
		model.DnoSse:                      NewTestDnoClient(model.DnoSse, numOutages),
		model.DnoUKPowerNetwork:           NewTestDnoClient(model.DnoUKPowerNetwork, numOutages),
	}
}
