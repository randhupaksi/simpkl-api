package documentgen

type SchoolProfile struct {
	Name     string
	Type     string
	NPSN     string
	Address  string
	City     string
	Province string
	Phone    string
	Email    string
	Website  string
	Tagline  string
}

type Signatory struct {
	Name           string
	Title          string
	EmployeeNumber string
}

type Letter struct {
	Profile          SchoolProfile
	Signatory        Signatory
	Number           string
	Date             string
	Subject          string
	Recipient        string
	RecipientAddress string
	Body             string
}
