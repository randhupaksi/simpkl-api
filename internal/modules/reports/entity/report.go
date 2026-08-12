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
	Period              DashboardPeriod       `json:"period"`
	TotalStudents       int64                 `json:"total_students"`
	PlacedStudents      int64                 `json:"placed_students"`
	UnplacedStudents    int64                 `json:"unplaced_students"`
	ActivePlacements    int64                 `json:"active_placements"`
	ActiveCompanies     int64                 `json:"active_companies"`
	IncompleteDocuments int64                 `json:"incomplete_documents"`
	StartingSoon        int64                 `json:"starting_soon"`
	EndingSoon          int64                 `json:"ending_soon"`
	PlacementStatuses   []DashboardBreakdown  `json:"placement_statuses"`
	MajorProgress       []DashboardMajor      `json:"major_progress"`
	Readiness           DashboardReadiness    `json:"readiness"`
	CompanyCapacity     DashboardCapacity     `json:"company_capacity"`
	Priorities          []DashboardPriority   `json:"priorities"`
	Agenda              []DashboardAgendaItem `json:"agenda"`
	RecentActivities    []DashboardActivity   `json:"recent_activities"`
}

type DashboardPeriod struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	AcademicYear string `json:"academic_year"`
	Semester     string `json:"semester"`
	Status       string `json:"status"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
}

type DashboardBreakdown struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type DashboardMajor struct {
	MajorID        string `json:"major_id"`
	MajorCode      string `json:"major_code"`
	MajorName      string `json:"major_name"`
	TotalStudents  int64  `json:"total_students"`
	PlacedStudents int64  `json:"placed_students"`
	ActiveStudents int64  `json:"active_students"`
}

type DashboardReadiness struct {
	Total      int64   `json:"total"`
	Ready      int64   `json:"ready"`
	Attention  int64   `json:"attention"`
	Incomplete int64   `json:"incomplete"`
	Average    float64 `json:"average"`
}

type DashboardCapacity struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
}

type DashboardPriority struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Value       int64  `json:"value"`
	Tone        string `json:"tone"`
	Path        string `json:"path"`
}

type DashboardAgendaItem struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        string `json:"date"`
	DaysLeft    int    `json:"days_left"`
	Tone        string `json:"tone"`
	Path        string `json:"path"`
}

type DashboardActivity struct {
	ID        string `json:"id"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	ActorName string `json:"actor_name"`
	CreatedAt string `json:"created_at"`
}

type ExpirationItem struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	ExpiresAt string `json:"expires_at"`
	DaysLeft  int    `json:"days_left"`
}
