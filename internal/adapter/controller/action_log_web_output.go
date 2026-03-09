package controller

// ActionLogWebOutput 棋譜Webアウトプット
type ActionLogWebOutput struct {
	Entries []*ActionLogWebEntry `json:"entries"`
}

// ActionLogWebEntry 棋譜エントリWebアウトプット
type ActionLogWebEntry struct {
	TurnNumber int              `json:"turnNumber"`
	PlayerIdx  int              `json:"playerIdx"`
	ActionType string           `json:"actionType"`
	Detail     string           `json:"detail"`
	Cards      []*WebOutputCard `json:"cards,omitempty"`
}
