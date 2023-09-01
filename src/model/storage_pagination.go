package model

import (
	"math"

	"gorm.io/gorm"
)

const DEFAULT_PAGE_LIMIT = 10
const DEFAULT_SORT_ORDER = "id desc"

type Pagination struct {
	Limit      int         `json:"limit,omitempty;query:limit"`
	Page       int         `json:"page,omitempty;query:page"`
	Sort       string      `json:"sort,omitempty;query:sort"`
	TotalRows  int64       `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
	Rows       interface{} `json:"rows"`
}

func NewPagiNation(limit int, page int) *Pagination {
	return &Pagination{
		Limit: limit,
		Page:  page,
	}
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = DEFAULT_PAGE_LIMIT
	}
	return p.Limit
}

func (p *Pagination) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

func (p *Pagination) GetTotalRows() int64 {
	return p.TotalRows
}

func (p *Pagination) GetTotalPages() int {
	return p.TotalPages
}

func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		p.Sort = DEFAULT_SORT_ORDER
	}
	return p.Sort
}

func (p *Pagination) SetTotalRows(totalRows int64) {
	p.TotalRows = totalRows
}

func (p *Pagination) SetSort(field string, vec string) {
	if len(p.Sort) == 0 {
		p.Sort += field + " " + vec
	} else {
		p.Sort += ", " + field + " " + vec
	}
}

func (p *Pagination) CalculateTotalPagesByTotalRows(totalRows int64) {
	p.TotalPages = int(math.Ceil(float64(totalRows) / float64(p.Limit)))
}

func paginate(db *gorm.DB, pagination *Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}
