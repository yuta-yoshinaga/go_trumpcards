//go:build !js || !wasm || casino

package domain

import "encoding/json"

// OpenFaceChinesePlayer オープンフェイス・チャイニーズポーカーのプレイヤー。
//
// プレイヤーは 3 つの段（front=3 枚, middle=5 枚, back=5 枚）を 1 枚ずつ表向きに置いて
// 構築する。pending は配られた／引いた未配置のカードのバッファ（通常は 1 枚、最初の
// ディールとファンタジーランドでは複数枚）。一度置いたカードは動かせない。
type OpenFaceChinesePlayer struct {
	*GamePlayer
	front       []*Card // 上段（最大 3 枚）
	middle      []*Card // 中段（最大 5 枚）
	back        []*Card // 下段（最大 5 枚）
	pending     []*Card // 未配置カードのバッファ
	roundScore  int     // 直近ラウンドの対戦相手合計に対する得点（ロイヤリティ込み）
	royalty     int     // 直近ラウンドのロイヤリティ点
	fouled      bool    // 直近ラウンドでファウルしたか
	fantasyland bool    // 次ラウンドにファンタジーランド権を持つか
	totalScore  int     // マッチ累積得点
}

// NewOpenFaceChinesePlayer コンストラクタ
func NewOpenFaceChinesePlayer(isHuman bool) *OpenFaceChinesePlayer {
	return &OpenFaceChinesePlayer{
		GamePlayer: NewGamePlayer(isHuman),
		front:      make([]*Card, 0, OpenFaceChineseFrontSize),
		middle:     make([]*Card, 0, OpenFaceChineseMiddleSize),
		back:       make([]*Card, 0, OpenFaceChineseBackSize),
		pending:    make([]*Card, 0, OpenFaceChineseHandSize),
	}
}

// ResetRound ラウンド開始時に段・バッファ・ラウンド状態を初期化する（累積得点と
// fantasyland 権は保持する）。
func (p *OpenFaceChinesePlayer) ResetRound() {
	p.front = make([]*Card, 0, OpenFaceChineseFrontSize)
	p.middle = make([]*Card, 0, OpenFaceChineseMiddleSize)
	p.back = make([]*Card, 0, OpenFaceChineseBackSize)
	p.pending = make([]*Card, 0, OpenFaceChineseHandSize)
	p.roundScore = 0
	p.royalty = 0
	p.fouled = false
	p.Reset()
}

// GetFront 上段（最大 3 枚）を取得
func (p *OpenFaceChinesePlayer) GetFront() []*Card { return p.front }

// GetMiddle 中段（最大 5 枚）を取得
func (p *OpenFaceChinesePlayer) GetMiddle() []*Card { return p.middle }

// GetBack 下段（最大 5 枚）を取得
func (p *OpenFaceChinesePlayer) GetBack() []*Card { return p.back }

// GetPending 未配置カードのバッファを取得
func (p *OpenFaceChinesePlayer) GetPending() []*Card { return p.pending }

// GetRoundScore 直近ラウンドの得点を取得
func (p *OpenFaceChinesePlayer) GetRoundScore() int { return p.roundScore }

// GetRoyalty 直近ラウンドのロイヤリティ点を取得
func (p *OpenFaceChinesePlayer) GetRoyalty() int { return p.royalty }

// GetFouled 直近ラウンドでファウルしたかを取得
func (p *OpenFaceChinesePlayer) GetFouled() bool { return p.fouled }

// GetFantasyland 次ラウンドにファンタジーランド権を持つかを取得
func (p *OpenFaceChinesePlayer) GetFantasyland() bool { return p.fantasyland }

// GetTotalScore マッチ累積得点を取得
func (p *OpenFaceChinesePlayer) GetTotalScore() int { return p.totalScore }

// SetFront 上段を設定（テスト用）
func (p *OpenFaceChinesePlayer) SetFront(cards []*Card) { p.front = cards }

// SetMiddle 中段を設定（テスト用）
func (p *OpenFaceChinesePlayer) SetMiddle(cards []*Card) { p.middle = cards }

// SetBack 下段を設定（テスト用）
func (p *OpenFaceChinesePlayer) SetBack(cards []*Card) { p.back = cards }

// SetPending バッファを設定（テスト用）
func (p *OpenFaceChinesePlayer) SetPending(cards []*Card) { p.pending = cards }

// SetRoundScore 直近ラウンドの得点を設定（テスト用）
func (p *OpenFaceChinesePlayer) SetRoundScore(v int) { p.roundScore = v }

// SetRoyalty 直近ラウンドのロイヤリティ点を設定（テスト用）
func (p *OpenFaceChinesePlayer) SetRoyalty(v int) { p.royalty = v }

// SetFouled ファウルフラグを設定（テスト用）
func (p *OpenFaceChinesePlayer) SetFouled(v bool) { p.fouled = v }

// SetFantasyland ファンタジーランド権を設定（テスト用）
func (p *OpenFaceChinesePlayer) SetFantasyland(v bool) { p.fantasyland = v }

// SetTotalScore マッチ累積得点を設定（テスト用）
func (p *OpenFaceChinesePlayer) SetTotalScore(v int) { p.totalScore = v }

// placedCount 既に配置済みのカード総数を返す。
func (p *OpenFaceChinesePlayer) placedCount() int {
	return len(p.front) + len(p.middle) + len(p.back)
}

// rowFull 指定段が満杯か。
func (p *OpenFaceChinesePlayer) rowFull(row int) bool {
	switch row {
	case OpenFaceChineseRowFront:
		return len(p.front) >= OpenFaceChineseFrontSize
	case OpenFaceChineseRowMiddle:
		return len(p.middle) >= OpenFaceChineseMiddleSize
	default:
		return len(p.back) >= OpenFaceChineseBackSize
	}
}

// placeCard 先頭の pending カードを指定段に置く。段が満杯か pending が空ならエラー。
// 一度置いたカードは動かせない（pending から取り除くだけで戻す手段は提供しない）。
func (p *OpenFaceChinesePlayer) placeCard(row int) error {
	if len(p.pending) == 0 {
		return NewDomainError(ErrInvalidPlay, "no pending card to place")
	}
	if row < OpenFaceChineseRowFront || row > OpenFaceChineseRowBack {
		return NewDomainError(ErrInvalidPlay, "invalid row")
	}
	if p.rowFull(row) {
		return NewDomainError(ErrInvalidPlay, "row is full")
	}
	card := p.pending[0]
	p.pending[0] = nil // 参照を切って GC を妨げないようにする
	p.pending = p.pending[1:]
	switch row {
	case OpenFaceChineseRowFront:
		p.front = append(p.front, card)
	case OpenFaceChineseRowMiddle:
		p.middle = append(p.middle, card)
	default:
		p.back = append(p.back, card)
	}
	return nil
}

// openFaceChinesePlayerJSON is the JSON wire format for OpenFaceChinesePlayer.
type openFaceChinesePlayerJSON struct {
	GamePlayer  *GamePlayer `json:"gp"`
	Front       []*Card     `json:"fr"`
	Middle      []*Card     `json:"md"`
	Back        []*Card     `json:"bk"`
	Pending     []*Card     `json:"pd"`
	RoundScore  int         `json:"rs"`
	Royalty     int         `json:"ry"`
	Fouled      bool        `json:"fo"`
	Fantasyland bool        `json:"fl"`
	TotalScore  int         `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (p *OpenFaceChinesePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(openFaceChinesePlayerJSON{
		GamePlayer:  p.GamePlayer,
		Front:       p.front,
		Middle:      p.middle,
		Back:        p.back,
		Pending:     p.pending,
		RoundScore:  p.roundScore,
		Royalty:     p.royalty,
		Fouled:      p.fouled,
		Fantasyland: p.fantasyland,
		TotalScore:  p.totalScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler. Row buffers are capped at their
// face-up sizes so a tampered payload cannot inject oversized rows.
func (p *OpenFaceChinesePlayer) UnmarshalJSON(data []byte) error {
	var j openFaceChinesePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Front) > OpenFaceChineseFrontSize || len(j.Middle) > OpenFaceChineseMiddleSize ||
		len(j.Back) > OpenFaceChineseBackSize || len(j.Pending) > OpenFaceChineseHandSize {
		return errOpenFaceChineseInvalidState
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.front = ofcNilToEmptyCards(j.Front)
	p.middle = ofcNilToEmptyCards(j.Middle)
	p.back = ofcNilToEmptyCards(j.Back)
	p.pending = ofcNilToEmptyCards(j.Pending)
	p.roundScore = j.RoundScore
	p.royalty = j.Royalty
	p.fouled = j.Fouled
	p.fantasyland = j.Fantasyland
	p.totalScore = j.TotalScore
	return nil
}

// ofcNilToEmptyCards returns an empty (non-nil) slice when cards is nil.
func ofcNilToEmptyCards(cards []*Card) []*Card {
	if cards == nil {
		return make([]*Card, 0)
	}
	return cards
}
