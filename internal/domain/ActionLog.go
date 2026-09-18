package domain

import "encoding/json"

// ActionLogEntry 棋譜エントリ
type ActionLogEntry struct {
	TurnNumber   int               // ターン番号
	PlayerIdx    int               // プレイヤーインデックス (-1 = システム/ディーラー)
	ActionType   string            // アクション種別 (ゲーム固有)
	Detail       string            // 人間が読める説明
	DetailCode   string            // 翻訳に使う説明コード
	DetailParams map[string]string // 説明コードに渡す翻訳パラメータ
	Cards        []*Card           // 関連カード (常に公開)
}

// actionLogEntryJSON is the JSON wire format for ActionLogEntry.
type actionLogEntryJSON struct {
	T  int               `json:"t"`            // TurnNumber
	P  int               `json:"p"`            // PlayerIdx
	A  string            `json:"a"`            // ActionType
	D  string            `json:"d,omitempty"`  // Detail
	DC string            `json:"dc,omitempty"` // DetailCode
	DP map[string]string `json:"dp,omitempty"` // DetailParams
	C  []*Card           `json:"c"`            // Cards
}

// MarshalJSON implements json.Marshaler.
func (e *ActionLogEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(actionLogEntryJSON{
		T: e.TurnNumber, P: e.PlayerIdx, A: e.ActionType,
		D: e.Detail, DC: e.DetailCode, DP: e.DetailParams, C: e.Cards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *ActionLogEntry) UnmarshalJSON(data []byte) error {
	var j actionLogEntryJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	e.TurnNumber = j.T
	e.PlayerIdx = j.P
	e.ActionType = j.A
	e.Detail = j.D
	e.DetailCode = j.DC
	e.DetailParams = j.DP
	e.Cards = j.C
	return nil
}
