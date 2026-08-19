//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"sort"
)

// PolignacPlayer ポリニャックのプレイヤー
type PolignacPlayer struct {
	*GamePlayer
	TrickHolder
	// score は全ラウンドの累計**失点**。**小さいほど良い**。
	// スロバーハンネスとは符号の向きが逆なので注意。
	score int
	// roundPenalty はこのラウンドで受けた失点。
	roundPenalty int
	// declaredCapot はこのラウンドで capot を宣言したか。
	declaredCapot bool
}

// NewPolignacPlayer コンストラクタ
func NewPolignacPlayer(isHuman bool) *PolignacPlayer {
	return &PolignacPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame ゲーム全体をリセットする（累計失点も 0 に戻す）
func (p *PolignacPlayer) ResetGame() {
	p.ResetRound()
	p.score = 0
}

// ResetRound 1 ラウンド分の状態だけを初期化する。累計失点は残す。
func (p *PolignacPlayer) ResetRound() {
	resetPlayerRound(p)
	p.roundPenalty = 0
	p.declaredCapot = false
}

// GetScore 累計失点を返す（小さいほど良い）
func (p *PolignacPlayer) GetScore() int { return p.score }

// AddScore 累計失点に加算する
func (p *PolignacPlayer) AddScore(n int) { p.score += n }

// SetScore 累計失点を設定する（復元・テスト用）
func (p *PolignacPlayer) SetScore(n int) { p.score = n }

// GetRoundPenalty このラウンドで受けた失点
func (p *PolignacPlayer) GetRoundPenalty() int { return p.roundPenalty }

// GetTakenJackSuits はこのラウンドで取ったジャックのスートを、
// スペードを先頭にした昇順で返す。
//
// **合計失点だけでは、♠J を踏んだのか他を 2 枚拾ったのかが分からない** (#5746)。
// 姉妹ゲームの Slobberhannes / Reversis は取った印付き札を個別に出している。
// 状態を増やさず、獲得済みトリックから数える (失点の計算と同じ札を見る)。
func (p *PolignacPlayer) GetTakenJackSuits() []int {
	suits := make([]int, 0, 4)
	for _, trick := range p.GetTricksTaken() {
		for _, c := range trick {
			if c != nil && c.GetValue() == PolignacJackValue {
				suits = append(suits, c.GetDesign())
			}
		}
	}
	sortPolignacSuits(suits)
	return suits
}

// sortPolignacSuits は ♠ を先頭に、残りをスート番号の昇順に並べる。
func sortPolignacSuits(suits []int) {
	sort.Slice(suits, func(i, j int) bool {
		if (suits[i] == CardDesignSpade) != (suits[j] == CardDesignSpade) {
			return suits[i] == CardDesignSpade
		}
		return suits[i] < suits[j]
	})
}

// AddRoundPenalty このラウンドの失点に加算する
func (p *PolignacPlayer) AddRoundPenalty(n int) { p.roundPenalty += n }

// SetRoundPenalty このラウンドの失点を設定する（テスト用）
func (p *PolignacPlayer) SetRoundPenalty(n int) { p.roundPenalty = n }

// GetDeclaredCapot capot を宣言しているか
func (p *PolignacPlayer) GetDeclaredCapot() bool { return p.declaredCapot }

// SetDeclaredCapot capot 宣言を設定する
func (p *PolignacPlayer) SetDeclaredCapot(b bool) { p.declaredCapot = b }

// polignacPlayerJSON is the JSON wire format for PolignacPlayer.
type polignacPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	// 失点・宣言は必ず往復させる。Worker はリクエストごとに KV から作り直すので、
	// 抜けると得点がラウンド途中で 0 に戻る (#4478)。
	Score         int  `json:"sc"`
	RoundPenalty  int  `json:"rp"`
	DeclaredCapot bool `json:"dc"`
}

// MarshalJSON implements json.Marshaler.
func (p *PolignacPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(polignacPlayerJSON{
		GamePlayer:    p.GamePlayer,
		TrickHolder:   &p.TrickHolder,
		Score:         p.score,
		RoundPenalty:  p.roundPenalty,
		DeclaredCapot: p.declaredCapot,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PolignacPlayer) UnmarshalJSON(data []byte) error {
	var j polignacPlayerJSON
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
	p.score = j.Score
	p.roundPenalty = j.RoundPenalty
	p.declaredCapot = j.DeclaredCapot
	return nil
}
