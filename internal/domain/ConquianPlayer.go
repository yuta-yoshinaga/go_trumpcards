//go:build !js || !wasm || extra

package domain

import "encoding/json"

// ConquianPlayer コンキャンプレイヤークラス
//
// 手札 (GamePlayer 内) に加えて、テーブル上に表向きで並べたメルド (melds) と、
// マッチ通算勝利数 (wins) を保持する。
type ConquianPlayer struct {
	*GamePlayer
	melds [][]*Card // テーブル上の表向きメルド
	wins  int       // マッチ通算勝利数
}

// NewConquianPlayer コンストラクタ
func NewConquianPlayer(isHuman bool) *ConquianPlayer {
	return &ConquianPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		melds:      make([][]*Card, 0),
	}
}

// GetMelds テーブル上のメルド一覧を取得
func (p *ConquianPlayer) GetMelds() [][]*Card { return p.melds }

// SetMelds テーブル上のメルドを設定 (テスト用)
func (p *ConquianPlayer) SetMelds(melds [][]*Card) {
	if melds == nil {
		melds = make([][]*Card, 0)
	}
	p.melds = melds
}

// AddMeld 新しいメルドをテーブルに追加
func (p *ConquianPlayer) AddMeld(meld []*Card) {
	p.melds = append(p.melds, meld)
}

// GetWins マッチ通算勝利数を取得
func (p *ConquianPlayer) GetWins() int { return p.wins }

// SetWins マッチ通算勝利数を設定
func (p *ConquianPlayer) SetWins(w int) { p.wins = w }

// AddWin マッチ通算勝利数を1加算
func (p *ConquianPlayer) AddWin() { p.wins++ }

// ResetRound ラウンドをリセット（手札・メルド・終了状態を初期化、勝利数は維持）
func (p *ConquianPlayer) ResetRound() {
	p.Reset()
	p.melds = make([][]*Card, 0)
	p.SetIsFinished(false)
}

// conquianPlayerJSON is the JSON wire format for ConquianPlayer.
type conquianPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Melds      [][]*Card   `json:"md"`
	Wins       int         `json:"wn"`
}

// MarshalJSON implements json.Marshaler.
func (p *ConquianPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(conquianPlayerJSON{
		GamePlayer: p.GamePlayer,
		Melds:      p.melds,
		Wins:       p.wins,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ConquianPlayer) UnmarshalJSON(data []byte) error {
	var j conquianPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.melds = j.Melds
	if p.melds == nil {
		p.melds = make([][]*Card, 0)
	}
	p.wins = j.Wins
	return nil
}
