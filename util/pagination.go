package util

import (
	"net/http"
	"strconv"
)

const (
	_defaultCurPage  = 1
	_defaultPageSize = 10
	_maxPageSize     = 1000
)

const (
	curPageQueryParamKey  = "curPage"
	pageSizeQueryParamKey = "pageSize"
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

func NewPaginationFromRequest(r *http.Request, itemCount int) Pagination {
	cp, ps := parseQueryParam(r)
	return NewPagination(cp, ps, itemCount)
}

func parseQueryParam(r *http.Request) (curPage, pageSize int) {
	var err error
	if curPage, err = strconv.Atoi(r.URL.Query().Get(curPageQueryParamKey)); err != nil {
		curPage = _defaultCurPage
	}
	if pageSize, err = strconv.Atoi(r.URL.Query().Get(pageSizeQueryParamKey)); err != nil {
		pageSize = _defaultPageSize
	}

	return
}

func NewPagination(curPage, pageSize, itemCount int) Pagination {
	pg := Pagination{
		CurPage:   curPage,
		PageSize:  pageSize,
		ItemCount: itemCount,
	}

	// set page size
	if pageSize < 1 {
		pageSize = _defaultPageSize
	}

	if pageSize > _maxPageSize {
		pageSize = _maxPageSize
	}

	// set last page
	if itemCount == 0 {
		pg.LastPage = 1
	} else {
		if itemCount%pageSize == 0 {
			pg.LastPage = itemCount / pageSize
		} else {
			pg.LastPage = itemCount/pageSize + 1
		}
	}

	// set cur page
	if pg.CurPage < 1 {
		pg.CurPage = 1
	}

	if pg.CurPage > pg.LastPage {
		pg.CurPage = pg.LastPage
	}

	return pg
}
