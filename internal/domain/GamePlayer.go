package domain

import "encoding/json"

// GamePlayer isHuman/isFinished を持つプレイヤー共通構造体
type GamePlayer struct {
	Player
	isHuman    bool
	isFinished bool
}

// NewGamePlayer コンストラクタ
func NewGamePlayer(isHuman bool) *GamePlayer {
	return &GamePlayer{
		Player:  Player{cards: make([]*Card, 0)},
		isHuman: isHuman,
	}
}

// GetIsHuman 人間プレイヤーかどうか
func (gp *GamePlayer) GetIsHuman() bool { return gp.isHuman }

// GetIsFinished 上がっているかどうか
func (gp *GamePlayer) GetIsFinished() bool { return gp.isFinished }

// SetIsFinished 上がり状態設定
func (gp *GamePlayer) SetIsFinished(v bool) { gp.isFinished = v }

// RankedGamePlayer ランク付きプレイヤー共通構造体
type RankedGamePlayer struct {
	*GamePlayer
	rank int
}

// NewRankedGamePlayer コンストラクタ (rank defaults to -1)
func NewRankedGamePlayer(isHuman bool) *RankedGamePlayer {
	return &RankedGamePlayer{
		GamePlayer: NewGamePlayer(isHuman),
		rank:       -1,
	}
}

// GetRank ランク取得 (-1 = 未確定)
func (rp *RankedGamePlayer) GetRank() int { return rp.rank }

// SetRank ランク設定
func (rp *RankedGamePlayer) SetRank(r int) { rp.rank = r }

// gamePlayerJSON is the JSON wire format for GamePlayer.
type gamePlayerJSON struct {
	Player     *Player `json:"p"`
	IsHuman    bool    `json:"h"`
	IsFinished bool    `json:"f"`
}

// MarshalJSON implements json.Marshaler.
func (gp *GamePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(gamePlayerJSON{
		Player:     &gp.Player,
		IsHuman:    gp.isHuman,
		IsFinished: gp.isFinished,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (gp *GamePlayer) UnmarshalJSON(data []byte) error {
	var j gamePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player != nil {
		gp.Player = *j.Player
	}
	gp.isHuman = j.IsHuman
	gp.isFinished = j.IsFinished
	return nil
}

// rankedGamePlayerJSON is the JSON wire format for RankedGamePlayer.
type rankedGamePlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Rank       int         `json:"r"`
}

// MarshalJSON implements json.Marshaler.
func (rp *RankedGamePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(rankedGamePlayerJSON{
		GamePlayer: rp.GamePlayer,
		Rank:       rp.rank,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (rp *RankedGamePlayer) UnmarshalJSON(data []byte) error {
	var j rankedGamePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		rp.GamePlayer = j.GamePlayer
	} else {
		rp.GamePlayer = NewGamePlayer(false)
	}
	rp.rank = j.Rank
	return nil
}
