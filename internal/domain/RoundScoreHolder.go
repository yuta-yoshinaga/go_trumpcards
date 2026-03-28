package domain

import "encoding/json"

// RoundScoreHolder ラウンドスコア管理の共通構造体
type RoundScoreHolder struct {
	roundScore      int // 現在のラウンドのスコア
	cumulativeScore int // 累積スコア
}

// GetRoundScore 現在のラウンドスコアを取得
func (h *RoundScoreHolder) GetRoundScore() int { return h.roundScore }

// SetRoundScore 現在のラウンドスコアを設定
func (h *RoundScoreHolder) SetRoundScore(score int) { h.roundScore = score }

// GetCumulativeScore 累積スコアを取得
func (h *RoundScoreHolder) GetCumulativeScore() int { return h.cumulativeScore }

// SetCumulativeScore 累積スコアを設定
func (h *RoundScoreHolder) SetCumulativeScore(score int) { h.cumulativeScore = score }

// CommitRoundScore ラウンドスコアを累積スコアに加算
func (h *RoundScoreHolder) CommitRoundScore() {
	h.cumulativeScore += h.roundScore
}

// roundScoreHolderJSON is the JSON wire format for RoundScoreHolder.
type roundScoreHolderJSON struct {
	RoundScore      int `json:"rs"` // roundScore
	CumulativeScore int `json:"cs"` // cumulativeScore
}

// MarshalJSON implements json.Marshaler.
func (h *RoundScoreHolder) MarshalJSON() ([]byte, error) {
	return json.Marshal(roundScoreHolderJSON{
		RoundScore:      h.roundScore,
		CumulativeScore: h.cumulativeScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (h *RoundScoreHolder) UnmarshalJSON(data []byte) error {
	var j roundScoreHolderJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	h.roundScore = j.RoundScore
	h.cumulativeScore = j.CumulativeScore
	return nil
}
