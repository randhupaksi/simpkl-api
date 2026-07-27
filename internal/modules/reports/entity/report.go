package entity

type PlacementRow struct {
	NIS            string `json:"nis"`
	StudentName    string `json:"student_name"`
	ClassName      string `json:"class_name"`
	MajorName      string `json:"major_name"`
	CompanyName    string `json:"company_name"`
	SupervisorName string `json:"supervisor_name"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	Status         string `json:"status"`
}

type Dashboard struct {
	TotalStudents       int64 `json:"total_students"`
	PlacedStudents      int64 `json:"placed_students"`
	UnplacedStudents    int64 `json:"unplaced_students"`
	ActivePlacements    int64 `json:"active_placements"`
	ActiveCompanies     int64 `json:"active_companies"`
	IncompleteDocuments int64 `json:"incomplete_documents"`
	StartingSoon        int64 `json:"starting_soon"`
	EndingSoon          int64 `json:"ending_soon"`
}

type ExpirationItem struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	ExpiresAt string `json:"expires_at"`
	DaysLeft  int    `json:"days_left"`
}
