package pagination

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type PaginationSerializer struct {
	Limit        int         `json:"limit"`
	Page         int         `json:"page"`
	TotalRecords int         `json:"total_records"`
	TotalPages   int         `json:"total_pages"`
	PreviousPage interface{} `json:"previous_page"`
	NextPage     interface{} `json:"next_page"`
	Data         interface{} `json:"data"`
}

func ListPaginationSerializer(limit int, page int, totalRecords int, totalPages int, nextPage interface{}, previsouPage interface{}, listData interface{}) PaginationSerializer {
	return PaginationSerializer{
		Limit:        limit,
		Page:         page,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		NextPage:     nextPage,
		PreviousPage: previsouPage,
		Data:         listData,
	}
}

func ValidatePagination(r *http.Request, count int) (int, int, int, int, interface{}, interface{}, error) {

	page := 1
	limit := 100

	// validate page and limit
	if r.URL.Query().Get("page") != "" {

		p, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err == nil {
			if p > 0 {
				page = p
			}
		}
	}

	if r.URL.Query().Get("limit") != "" {

		l, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err == nil {
			if l > 0 {
				limit = l
			}
		}
	}

	// calculate total pages
	totalPages := math.Ceil(float64(count) / float64(limit))

	// validate pagination
	if int(totalPages) > 0 && page > int(totalPages) {
		return 0, 0, 0, 0, nil, nil, errors.New("exceeded the number of pages")
	}

	// calculate next page
	var nextPage interface{}

	if page >= int(totalPages) {
		nextPage = nil

	} else {
		pathWithoutParams := strings.Split(fmt.Sprintf("%v", r.URL), "?")
		nextPage = os.Getenv("SERVER_URL") + pathWithoutParams[0] + "?limit=" + fmt.Sprintf("%v", limit) + "&page=" + fmt.Sprintf("%v", page+1)
	}

	// calculate previous page
	var previousPage interface{}

	if page <= 1 {
		previousPage = nil

	} else {
		pathWithoutParams := strings.Split(fmt.Sprintf("%v", r.URL), "?")
		previousPage = os.Getenv("SERVER_URL") + pathWithoutParams[0] + "?limit=" + fmt.Sprintf("%v", limit) + "&page=" + fmt.Sprintf("%v", page-1)
	}

	return ((page - 1) * limit), limit, page, int(totalPages), nextPage, previousPage, nil
}

func ValidatePaginationData(r *http.Request) (int, int) {

	page := 1
	limit := 100

	// validate page and limit
	if r.URL.Query().Get("page") != "" {

		p, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err == nil {
			if p > 0 {
				page = p
			}
		}
	}

	if r.URL.Query().Get("limit") != "" {

		l, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err == nil {
			if l > 0 {
				limit = l
			}
		}
	}

	// calculate next page

	return ((page - 1) * limit), limit
}
