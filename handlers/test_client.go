package handlers

import (
	"context"
	"fmt"
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
		start := time.Now()
		end := start.Add(30 * time.Hour)
		p, _ := model.NewPostcode(fmt.Sprintf("N%d %dAA", i%100, i%10))
		postcodes := []model.Postcode{p}
		o := model.Outage{
			DNO:       t.Dno,
			ID:        strconv.Itoa(i),
			Start:     &start,
			End:       &end,
			Postcodes: postcodes,
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
