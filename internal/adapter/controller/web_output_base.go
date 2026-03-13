package controller

// WebOutputBase holds fields common to all game WebOutput structs.
type WebOutputBase struct {
	Message       string            `json:"message"`
	MessageCode   string            `json:"messageCode,omitempty"`
	MessageParams map[string]string `json:"messageParams,omitempty"`
}
