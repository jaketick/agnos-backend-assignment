package dto

type PatientSearchRequest struct {
	NationalID  string `form:"national_id"`
	PassportID  string `form:"passport_id"`
	FirstName   string `form:"first_name"`
	MiddleName  string `form:"middle_name"`
	LastName    string `form:"last_name"`
	DateOfBirth string `form:"date_of_birth"`
	PhoneNumber string `form:"phone_number"`
	Email       string `form:"email"`
}

type PatientResponse struct {
	ID uint `json:"id"`

	PatientHN string `json:"patient_hn"`

	NationalID string `json:"national_id"`
	PassportID string `json:"passport_id"`

	FirstNameTH  string `json:"first_name_th"`
	MiddleNameTH string `json:"middle_name_th"`
	LastNameTH   string `json:"last_name_th"`

	FirstNameEN  string `json:"first_name_en"`
	MiddleNameEN string `json:"middle_name_en"`
	LastNameEN   string `json:"last_name_en"`

	DateOfBirth string `json:"date_of_birth"`

	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	Gender      string `json:"gender"`
}

type PatientSearchResponse struct {
	Count int               `json:"count"`
	Data  []PatientResponse `json:"data"`
}
