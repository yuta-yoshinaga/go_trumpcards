//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// RookPlayer ルーク(Rook)プレイヤークラス
type RookPlayer struct {
	*GamePlayer
	TrickHolder
	team       int  // チームインデックス (0 or 1)
	bid        int  // このラウンドの最終ビッド (0 = 未ビッド)
	passed     bool // このラウンドでパス済みか
	isDeclarer bool // 落札者(宣言者)か
	points     int  // このラウンドで獲得した得点札の合計
}

// NewRookPlayer コンストラクタ
func NewRookPlayer(isHuman bool, team int) *RookPlayer {
	return &RookPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *RookPlayer) GetTeam() int { return p.team }

// GetBid このラウンドのビッドを取得 (0 = 未ビッド)
func (p *RookPlayer) GetBid() int { return p.bid }

// SetBid ビッドを設定
func (p *RookPlayer) SetBid(bid int) { p.bid = bid }

// GetPassed パス済みかどうか
func (p *RookPlayer) GetPassed() bool { return p.passed }

// SetPassed パス状態を設定
func (p *RookPlayer) SetPassed(v bool) { p.passed = v }

// GetIsDeclarer 落札者かどうか
func (p *RookPlayer) GetIsDeclarer() bool { return p.isDeclarer }

// SetIsDeclarer 落札者状態を設定
func (p *RookPlayer) SetIsDeclarer(v bool) { p.isDeclarer = v }

// GetPoints このラウンドで獲得した得点札の合計を取得
func (p *RookPlayer) GetPoints() int { return p.points }

// AddPoints 得点札の合計に加算
func (p *RookPlayer) AddPoints(v int) { p.points += v }

// SetPoints 得点札の合計を設定 (テスト用)
func (p *RookPlayer) SetPoints(v int) { p.points = v }

// ResetRound ラウンドをリセット (トリック・手札・ビッド・得点状態を初期化)
func (p *RookPlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
	p.bid = 0
	p.passed = false
	p.isDeclarer = false
	p.points = 0
}

// rookPlayerJSON is the JSON wire format for RookPlayer.
type rookPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm"`
	Bid         int          `json:"bd"`
	Passed      bool         `json:"ps"`
	IsDeclarer  bool         `json:"dc"`
	Points      int          `json:"pt"`
}

// MarshalJSON implements json.Marshaler.
func (p *RookPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(rookPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
		Bid:         p.bid,
		Passed:      p.passed,
		IsDeclarer:  p.isDeclarer,
		Points:      p.points,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *RookPlayer) UnmarshalJSON(data []byte) error {
	var j rookPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.team = j.Team
	p.bid = j.Bid
	p.passed = j.Passed
	p.isDeclarer = j.IsDeclarer
	p.points = j.Points
	return nil
}
