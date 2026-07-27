package http

type RecalculateRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	PeriodID  string `json:"period_id" binding:"required"`
}

type OverrideRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	PeriodID  string `json:"period_id" binding:"required"`
	Reason    string `json:"reason" binding:"required"`
}
