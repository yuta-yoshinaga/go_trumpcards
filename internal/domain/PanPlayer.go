//go:build !js || !wasm || extra

package domain

import "encoding/json"

// PanPlayer パングインゲのプレイヤー。手札・場に出したメルド・チップ・スコアを保持する。
type PanPlayer struct {
	*GamePlayer
	RoundScoreHolder
	// laidMelds はプレイヤーが自分の場に出したメルド一覧（各メルドは 3 枚以上）。
	laidMelds [][]*Card
	// chips はチップ残高（条件付き役でのやり取りの累積）。
	chips int
}

// NewPanPlayer コンストラクタ
func NewPanPlayer(isHuman bool) *PanPlayer {
	return &PanPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・ラウンドスコア・メルド・終了状態を初期化。チップは通算で保持）
func (p *PanPlayer) ResetRound() {
	p.SetRoundScore(0)
	p.Reset()
	p.laidMelds = nil
	p.SetIsFinished(false)
}

// GetLaidMelds 場に出したメルド一覧を取得する
func (p *PanPlayer) GetLaidMelds() [][]*Card { return p.laidMelds }

// SetLaidMelds メルド一覧を設定する（テスト用）
func (p *PanPlayer) SetLaidMelds(melds [][]*Card) { p.laidMelds = melds }

// AddLaidMeld 新しいメルドを場に追加する
func (p *PanPlayer) AddLaidMeld(meld []*Card) {
	cp := make([]*Card, len(meld))
	copy(cp, meld)
	p.laidMelds = append(p.laidMelds, cp)
}

// AppendToLaidMeld 既存メルドにカードを追加する（レイオフ）。範囲外なら false。
func (p *PanPlayer) AppendToLaidMeld(meldIdx int, card *Card) bool {
	if meldIdx < 0 || meldIdx >= len(p.laidMelds) {
		return false
	}
	p.laidMelds[meldIdx] = append(p.laidMelds[meldIdx], card)
	return true
}

// GetMeldedCardCount 場に出したカードの総枚数を返す（11 枚で「パン」あがり）。
func (p *PanPlayer) GetMeldedCardCount() int {
	total := 0
	for _, m := range p.laidMelds {
		total += len(m)
	}
	return total
}

// GetChips チップ残高を取得する
func (p *PanPlayer) GetChips() int { return p.chips }

// SetChips チップ残高を設定する（テスト用）
func (p *PanPlayer) SetChips(v int) { p.chips = v }

// AddChips チップ残高に加算する（負値で支払い）
func (p *PanPlayer) AddChips(delta int) { p.chips += delta }

// panPlayerJSON は PanPlayer の JSON 表現。
type panPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	LaidMelds        [][]*Card         `json:"lm"`
	Chips            int               `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (p *PanPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(panPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		LaidMelds:        p.laidMelds,
		Chips:            p.chips,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PanPlayer) UnmarshalJSON(data []byte) error {
	var j panPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.RoundScoreHolder != nil {
		p.RoundScoreHolder = *j.RoundScoreHolder
	}
	p.laidMelds = panSanitizeMelds(j.LaidMelds)
	p.chips = j.Chips
	return nil
}

// panSanitizeMelds は復元したメルド一覧から nil メルド・nil カードを除去する
// （後段のイテレーションでの nil デリファレンスパニックを防ぐ）。
func panSanitizeMelds(melds [][]*Card) [][]*Card {
	out := make([][]*Card, 0, len(melds))
	for _, m := range melds {
		if m == nil {
			continue
		}
		cleaned := make([]*Card, 0, len(m))
		for _, c := range m {
			if c != nil {
				cleaned = append(cleaned, c)
			}
		}
		if len(cleaned) > 0 {
			out = append(out, cleaned)
		}
	}
	return out
}
