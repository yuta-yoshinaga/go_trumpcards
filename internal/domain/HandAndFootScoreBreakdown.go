//go:build !js || !wasm || extra7

package domain

// HandAndFootScoreBreakdown records the values used to settle one team's round score.
type HandAndFootScoreBreakdown struct {
	MeldCards    int `json:"meldCards"`
	RedCanasta   int `json:"redCanasta"`
	BlackCanasta int `json:"blackCanasta"`
	RedThrees    int `json:"redThrees"`
	GoingOut     int `json:"goingOut"`
	HandPenalty  int `json:"handPenalty"`
	FootPenalty  int `json:"footPenalty"`
}

// Total returns the settled score represented by the breakdown.
func (b HandAndFootScoreBreakdown) Total() int {
	return b.MeldCards + b.RedCanasta + b.BlackCanasta + b.RedThrees + b.GoingOut - b.HandPenalty - b.FootPenalty
}
