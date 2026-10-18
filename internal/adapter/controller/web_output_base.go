package controller

// WebOutputRanking carries an untranslated player rank for frontend localization.
type WebOutputRanking struct {
	Seat     int  `json:"seat"`
	Rank     int  `json:"rank"`
	Shithead bool `json:"shithead,omitempty"`
}

// WebOutputBase holds fields common to all game WebOutput structs.
type WebOutputBase struct {
	Message       string            `json:"message"`
	MessageCode   string            `json:"messageCode,omitempty"`
	MessageParams map[string]string `json:"messageParams,omitempty"`
}
