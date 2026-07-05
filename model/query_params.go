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
	Status    []Status
}

func MakeDefaultQueryParams() QueryParams {
	return QueryParams{
		PageSize:  10,
		PageIndex: 0,
		Postcodes: []Postcode{},
		Dnos:      []Dno{},
		Status:    []Status{},
	}
}

func ParseQueryParams(values url.Values) (QueryParams, error) {
	qp := MakeDefaultQueryParams()

	pageSize := values.Get("pageSize")
	if pageSize != "" {
		i, err := strconv.ParseUint(pageSize, 10, 0)
		if err != nil {
			return qp, fmt.Errorf("failed to parse pageSize '%s': %w", pageSize, err)
		}
		qp.PageSize = uint(i)
	}

	pageIndex := values.Get("pageIndex")
	if pageIndex != "" {
		i, err := strconv.ParseUint(pageIndex, 10, 0)
		if err != nil {
			return qp, fmt.Errorf("failed to parse pageIndex '%s': %w", pageIndex, err)
		}
		qp.PageIndex = uint(i)
	}

	postcodes := values.Get("postcodes")
	if postcodes != "" {
		postcodeStrings := strings.Split(string(postcodes), ",")
		p, err := ParsePostcodes(postcodeStrings, true)
		if err != nil {
			return qp, fmt.Errorf("failed to parse postcode: %w", err)
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

	err := checkStatusTarget(values, &qp.Status)
	if err != nil {
		return qp, err
	}
	if len(qp.Status) == 0 {
		return qp, errors.New("no Status targeted")
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

func checkStatusTarget(values url.Values, targetStatus *[]Status) error {
	// Only Active is true by default.
	targetActive := values.Get(string(StatusActive))
	targetActiveLower := strings.ToLower(targetActive)
	if targetActiveLower == "" || targetActiveLower == "true" {
		*targetStatus = append(*targetStatus, StatusActive)
	} else if targetActiveLower != "false" {
		return fmt.Errorf("unexpected non-boolean value for status %s: %s", StatusActive, targetActive)
	}

	targetFuture := values.Get(string(StatusFuture))
	targetFutureLower := strings.ToLower(targetFuture)
	if targetFutureLower == "true" {
		*targetStatus = append(*targetStatus, StatusFuture)
	} else if targetFutureLower != "false" && targetFutureLower != "" {
		return fmt.Errorf("unexpected non-boolean value for status %s: %s", StatusFuture, targetFuture)
	}

	targetResolved := values.Get(string(StatusResolved))
	targetResolvedLower := strings.ToLower(targetResolved)
	if targetResolvedLower == "true" {
		*targetStatus = append(*targetStatus, StatusResolved)
	} else if targetResolvedLower != "false" && targetResolvedLower != "" {
		return fmt.Errorf("unexpected non-boolean value for status %s: %s", StatusResolved, targetResolved)
	}

	return nil
}
