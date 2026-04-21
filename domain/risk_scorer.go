package domain

// ScoreRisk assesses the risk of a commit based on its files and metadata.
// This is a stub — all commits return LOW until scoring rules are implemented.
func ScoreRisk(c Commit, files []FileChange) RiskAssessment {
	return RiskAssessment{
		Overall: RiskLow,
		Factors: []RiskFactor{
			{Description: "risk scoring not yet implemented", Level: RiskLow},
		},
	}
}
