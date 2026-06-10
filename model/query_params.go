package model

import (
	"fmt"
	"net/url"
	"strconv"
)

type QueryParams struct {
	PageSize  uint
	PageIndex uint
	Postcodes []string
}

func MakeDefaultQueryParams() QueryParams {
	return QueryParams{
		PageSize:  10,
		PageIndex: 0,
		Postcodes: []string{},
	}
}

func ParseQueryParams(values url.Values) (QueryParams, error) {
	qp := MakeDefaultQueryParams()

	pageSize := values.Get("pageSize")
	if pageSize != "" {
		i, err := strconv.ParseUint(pageSize, 10, 0)
		if err != nil {
			return qp, fmt.Errorf("failed to parse pageSize '%s': %v", pageSize, err)
		} else {
			qp.PageSize = uint(i)
		}
	}

	pageIndex := values.Get("pageIndex")
	if pageIndex != "" {
		i, err := strconv.ParseUint(pageIndex, 10, 0)
		if err != nil {
			return qp, fmt.Errorf("failed to parse pageIndex '%s': %v", pageIndex, err)
		} else {
			qp.PageIndex = uint(i)
		}
	}

	return qp, nil
}
