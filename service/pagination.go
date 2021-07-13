package service

import (
	"net/http"
	"strconv"
)

const (
	defaultPageSize = 10
	maxPageSize = 1000
)

type Pagination struct {
	CurPage   int `json:"curPage"`
	PageSize  int `json:"pageSize"`
	LastPage  int `json:"lastPage"`
	ItemCount int `json:"itemCount"`
}

func (p Pagination) Offset() int {
	return (p.CurPage - 1) * p.PageSize
}

func (p Pagination) Limit() int {
	return p.PageSize
}

func GetPaginationFromUrl(r *http.Request, itemCount int) Pagination {
	cp, ps := ExtractPageDataFromUrl(r)
	return GetPagination(cp, ps, itemCount)
}

func ExtractPageDataFromUrl(r *http.Request) (curPage, pageSize int) {
	var err error
	if curPage, err = strconv.Atoi(r.URL.Query().Get("curPage")); err != nil {
		curPage = 1
	}
	if pageSize, err = strconv.Atoi(r.URL.Query().Get("pageSize")); err != nil {
		pageSize = defaultPageSize
	}

	return
}

func GetPagination(curPage, pageSize, itemCount int) Pagination {
	pg := Pagination{
		CurPage: curPage,
		PageSize: pageSize,
		ItemCount: itemCount,
	}

	if pageSize < 1 {
		pageSize = defaultPageSize
	}

	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	if itemCount == 0 {
		pg.LastPage = 1
	} else {
		if itemCount % pageSize == 0 {
			pg.LastPage = itemCount / pageSize
		} else {
			pg.LastPage = itemCount / pageSize + 1
		}
	}

	if pg.CurPage < 1 {
		pg.CurPage = 1
	}

	if pg.CurPage > pg.LastPage {
		pg.CurPage = pg.LastPage
	}

	return pg
}
