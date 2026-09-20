package controller

// ActionLogWebOutput 棋譜Webアウトプット
type ActionLogWebOutput struct {
	Entries []*ActionLogWebEntry `json:"entries"`
}

// ActionLogWebEntry 棋譜エントリWebアウトプット
type ActionLogWebEntry struct {
	TurnNumber   int               `json:"turnNumber"`
	PlayerIdx    int               `json:"playerIdx"`
	ActionType   string            `json:"actionType"`
	DetailCode   string            `json:"detailCode,omitempty"`
	DetailParams map[string]string `json:"detailParams,omitempty"`
	Cards        []*WebOutputCard  `json:"cards,omitempty"`
}
