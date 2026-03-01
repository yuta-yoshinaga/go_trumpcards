package presenter

type errorResponse struct {
	Error string `json:"error"`
}

func internalServerErrorJSON() string {
	return `{"error":"internal server error"}`
}
