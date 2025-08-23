package storage

import (
	"database/sql"
	"time"
)

// User representa a tabela 'users' no banco de dados.
type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	UserType     string    `json:"user_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Patient contém todos os dados do formulário inicial e cadastrais.
type Patient struct {
	ID                        int          `json:"id"`
	AccessToken               sql.NullString `json:"access_token"`
	ConsentGivenAt            sql.NullTime `json:"consent_given_at"`
	ConsentDate               string       `form:"consent_date" json:"consent_date"`
	ConsentName               string       `form:"consent_name_inline" json:"consent_name"`
	ConsentCpfRg              string       `form:"consent_cpf_rg_inline" json:"consent_cpf_rg"`
	SignatureDate             string       `form:"signature_date" json:"signature_date"`
	SignatureLocation         string       `form:"signature_location" json:"signature_location"`
	Name                      string       `form:"client_name" json:"name"`
	AddressStreet             string       `form:"address_street" json:"address_street"`
	AddressNumber             string       `form:"address_number" json:"address_number"`
	AddressNeighborhood       string       `form:"address_neighborhood" json:"address_neighborhood"`
	AddressCity               string       `form:"address_city" json:"address_city"`
	AddressState              string       `form:"address_state" json:"address_state"`
	Phone                     string       `form:"phone" json:"phone"`
	Mobile                    string       `form:"mobile" json:"mobile"`
	DOB                       string       `form:"dob" json:"dob"`
	Age                       int          `form:"age" json:"age"`
	Gender                    string       `form:"gender" json:"gender"`
	MaritalStatus             string       `form:"marital_status" json:"marital_status"`
	Children                  string       `form:"children" json:"children"`
	NumChildren               int          `form:"num_children" json:"num_children"`
	Profession                string       `form:"profession" json:"profession"`
	Email                     string       `form:"email" json:"email"`
	EmergencyContact          string       `form:"emergency_contact_name" json:"emergency_contact"`
	EmergencyPhone            string       `form:"emergency_contact_phone" json:"emergency_phone"`
	EmergencyOther            string       `form:"emergency_contact_other" json:"emergency_other"`
	RepetitiveEffort          string       `form:"repetitive_effort" json:"repetitive_effort"`
	PhysicalActivity          string       `form:"physical_activity" json:"physical_activity"`
	Smoker                    string       `form:"smoker" json:"smoker"`
	Alcohol                   string       `form:"alcohol" json:"alcohol"`
	MentalDisorder            string       `form:"mental_disorder" json:"mental_disorder"`
	MentalDisorderTreatment   string       `form:"mental_disorder_treatment" json:"mental_disorder_treatment"`
	MentalDisorderDetails     string       `form:"mental_disorder_details" json:"mental_disorder_details"`
	Religion                  string       `form:"religion" json:"religion"`
	ReligionDetails           string       `form:"religion_details" json:"religion_details"`
	Medication                string       `form:"medication" json:"medication"`
	MedicationDetails         string       `form:"medication_details" json:"medication_details"`
	Surgery                   string       `form:"surgery" json:"surgery"`
	SurgeryDetails            string       `form:"surgery_details" json:"surgery_details"`
	Allergies                 string       `form:"allergies" json:"allergies"`
	AllergiesDetails          string       `form:"allergies_details" json:"allergies_details"`
	HowFound                  string       `form:"how_found" json:"how_found"`
	ReferralName              string       `form:"referral_name" json:"referral_name"`
	OtherSource               string       `form:"other_source" json:"other_source"`
	Notes                     string       `form:"notes" json:"notes"`
	AnxietyLevel              int          `form:"anxiety_level" json:"anxiety_level"`
	AngerLevel                int          `form:"anger_level" json:"anger_level"`
	FearLevel                 int          `form:"fear_level" json:"fear_level"`
	SadnessLevel              int          `form:"sadness_level" json:"sadness_level"`
	JoyLevel                  int          `form:"joy_level" json:"joy_level"`
	EnergyLevel               int          `form:"energy_level" json:"energy_level"`
	MainComplaint             string       `form:"main_complaint" json:"main_complaint"`
	ComplaintHistory          string       `form:"complaint_history" json:"complaint_history"`
	SignsSymptoms             string       `form:"signs_symptoms" json:"signs_symptoms"`
	CurrentTreatment          string       `form:"current_treatment" json:"current_treatment"`
	CreatedAt                 time.Time    `json:"created_at"`
	UpdatedAt                 time.Time    `json:"updated_at"`
}

// PatientRecord é a struct para o histórico clínico (prontuário).
type PatientRecord struct {
	ID               int       `json:"id"`
	PatientID        int       `form:"patient_id" json:"patient_id"`
	DoctorID         int       `form:"doctor_id" json:"doctor_id"`
	RecordDate       time.Time `json:"record_date"`
	AnxietyLevel     int       `form:"anxiety_level" json:"anxiety_level"`
	AngerLevel       int       `form:"anger_level" json:"anger_level"`
	FearLevel        int       `form:"fear_level" json:"fear_level"`
	SadnessLevel     int       `form:"sadness_level" json:"sadness_level"`
	JoyLevel         int       `form:"joy_level" json:"joy_level"`
	EnergyLevel      int       `form:"energy_level" json:"energy_level"`
	MainComplaint    string    `form:"main_complaint" json:"main_complaint"`
	ComplaintHistory string    `form:"complaint_history" json:"complaint_history"`
	SignsSymptoms    string    `form:"signs_symptoms" json:"signs_symptoms"`
	CurrentTreatment string    `form:"current_treatment" json:"current_treatment"`
	Notes            string    `form:"notes" json:"notes"`
	DoctorName       string    // Campo auxiliar para exibir o nome do médico
}

// Appointment representa a tabela 'appointments' no banco de dados.
type Appointment struct {
	ID        int       `json:"id"`
	PatientID int       `json:"patient_id"`
	DoctorID  int       `json:"doctor_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Notes     string    `json:"notes"`
	Status    string    `json:"status"`
	Price         float64 `json:"price"`
	PaymentStatus string  `json:"payment_status"`	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
// storage/models.go

// AuditLog representa a tabela 'audit_logs' no banco de dados.
type AuditLog struct {
	ID         int
	UserID     sql.NullInt64
	UserName   sql.NullString
	Action     string
	TargetType sql.NullString
	TargetID   sql.NullInt64
	CreatedAt  time.Time
}