package pagination

import "math"

type Query struct {
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
	Search  string `form:"search"`
}
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func (q *Query) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 {
		q.PerPage = 20
	}
	if q.PerPage > 100 {
		q.PerPage = 100
	}
}
func (q Query) Offset() int { return (q.Page - 1) * q.PerPage }
func NewMeta(q Query, total int64) Meta {
	return Meta{q.Page, q.PerPage, total, int(math.Ceil(float64(total) / float64(q.PerPage)))}
}
