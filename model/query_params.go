package model

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type QueryParams struct {
	PageSize  uint
	PageIndex uint
	Postcodes Postcodes
	Dnos      []Dno
}

func MakeDefaultQueryParams() QueryParams {
	return QueryParams{
		PageSize:  10,
		PageIndex: 0,
		Postcodes: []Postcode{},
		Dnos:      []Dno{},
	}
}

func ParseQueryParams(values url.Values) (QueryParams, error) {
	qp := MakeDefaultQueryParams()

	pageSize := values.Get("pageSize")
	if pageSize != "" {
		i, err := strconv.ParseUint(pageSize, 10, 0)
		if err != nil {
			return qp, fmt.Errorf("failed to parse pageSize '%s': %v", pageSize, err)
		}
		qp.PageSize = uint(i)
	}

	pageIndex := values.Get("pageIndex")
	if pageIndex != "" {
		i, err := strconv.ParseUint(pageIndex, 10, 0)
		if err != nil {
			return qp, fmt.Errorf("failed to parse pageIndex '%s': %v", pageIndex, err)
		}
		qp.PageIndex = uint(i)
	}

	postcodes := values.Get("postcodes")
	if postcodes != "" {
		postcodeStrings := strings.Split(string(postcodes), ",")
		p, err := ParsePostcodes(postcodeStrings, true)
		if err != nil {
			return qp, fmt.Errorf("failed to parse postcode: %v", err)
		}
		qp.Postcodes = p
	}

	for _, dno := range AllDnoList {
		err := checkDnoTarget(values, &qp.Dnos, dno)
		if err != nil {
			return qp, err
		}
	}
	if len(qp.Dnos) == 0 {
		return qp, errors.New("no DNOs targeted")
	}

	return qp, nil
}

func checkDnoTarget(values url.Values, targetDnos *[]Dno, dnoToCheck Dno) error {
	isTarget := values.Get(string(dnoToCheck))
	isTargetLower := strings.ToLower(isTarget)
	if isTargetLower == "" || isTargetLower == "true" {
		*targetDnos = append(*targetDnos, dnoToCheck)
		return nil
	}
	if isTargetLower == "false" {
		return nil
	}

	return fmt.Errorf("unexpected non-boolean value for DNO %s: %s", dnoToCheck, isTarget)
}
