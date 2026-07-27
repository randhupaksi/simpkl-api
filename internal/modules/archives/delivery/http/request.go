package http

type ArchiveRequest struct {
	PeriodID string `json:"period_id" binding:"required"`
	Reason   string `json:"reason"`
}
