package domain

// RiskLevel is the assessed risk of a commit.
type RiskLevel string

const (
	RiskHigh   RiskLevel = "HIGH"
	RiskMedium RiskLevel = "MEDIUM"
	RiskLow    RiskLevel = "LOW"
)

// RiskFactor is a single reason contributing to a risk assessment.
type RiskFactor struct {
	Description string
	Level       RiskLevel
}

// RiskAssessment is the full result of scoring a commit.
type RiskAssessment struct {
	Overall RiskLevel
	Factors []RiskFactor
}
