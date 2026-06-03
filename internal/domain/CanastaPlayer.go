package domain

import "encoding/json"

// CanastaMeld テーブル上のメルド
type CanastaMeld struct {
	Cards     []*Card `json:"ca"`
	IsNatural bool    `json:"in"` // ワイルドカードを含まない場合 true
}

// IsCanasta 7枚以上のメルドかどうか
func (m *CanastaMeld) IsCanasta() bool {
	return len(m.Cards) >= 7
}

// IsBurraco は IsCanasta のエイリアス（Burraco モードでの呼称）。
func (m *CanastaMeld) IsBurraco() bool { return m.IsCanasta() }

// GetRank メルドのランク（ナチュラルカードのランク）を返す
func (m *CanastaMeld) GetRank() int {
	for _, c := range m.Cards {
		if !CanastaIsWild(c) {
			return c.GetValue()
		}
	}
	return 0
}

// CanastaPlayer カナスタプレイヤークラス
type CanastaPlayer struct {
	*GamePlayer
	RoundScoreHolder
	melds        []*CanastaMeld // テーブル上のメルド
	red3s        []*Card        // 場に出した赤3
	hasInitMeld  bool           // 初回メルド済みフラグ
	tookPozzetto bool           // ポゼット（予備手札）を獲得済みか (Burraco モードのみ)
}

// NewCanastaPlayer コンストラクタ
func NewCanastaPlayer(isHuman bool) *CanastaPlayer {
	return &CanastaPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		melds:      make([]*CanastaMeld, 0),
		red3s:      make([]*Card, 0),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・メルド・赤3を初期化）
func (p *CanastaPlayer) ResetRound() {
	p.SetRoundScore(0)
	p.Reset()
	p.SetIsFinished(false)
	p.melds = make([]*CanastaMeld, 0)
	p.red3s = make([]*Card, 0)
	p.hasInitMeld = false
	p.tookPozzetto = false
}

// GetMelds メルドを取得
func (p *CanastaPlayer) GetMelds() []*CanastaMeld { return p.melds }

// SetMelds メルドを設定 (テスト用)
func (p *CanastaPlayer) SetMelds(melds []*CanastaMeld) { p.melds = melds }

// AddMeld メルドを追加
func (p *CanastaPlayer) AddMeld(meld *CanastaMeld) {
	p.melds = append(p.melds, meld)
}

// GetRed3s 赤3を取得
func (p *CanastaPlayer) GetRed3s() []*Card { return p.red3s }

// SetRed3s 赤3を設定 (テスト用)
func (p *CanastaPlayer) SetRed3s(red3s []*Card) { p.red3s = red3s }

// AddRed3 赤3を追加
func (p *CanastaPlayer) AddRed3(card *Card) {
	p.red3s = append(p.red3s, card)
}

// GetHasInitMeld 初回メルド済みフラグ取得
func (p *CanastaPlayer) GetHasInitMeld() bool { return p.hasInitMeld }

// SetHasInitMeld 初回メルド済みフラグ設定
func (p *CanastaPlayer) SetHasInitMeld(v bool) { p.hasInitMeld = v }

// HasCanasta カナスタ（7枚以上のメルド）を持っているか
func (p *CanastaPlayer) HasCanasta() bool {
	for _, m := range p.melds {
		if m.IsCanasta() {
			return true
		}
	}
	return false
}

// HasBurraco は HasCanasta のエイリアス（Burraco モードでの呼称）。
func (p *CanastaPlayer) HasBurraco() bool { return p.HasCanasta() }

// GetTookPozzetto ポゼット獲得済みフラグ取得 (Burraco モードのみ)
func (p *CanastaPlayer) GetTookPozzetto() bool { return p.tookPozzetto }

// SetTookPozzetto ポゼット獲得済みフラグ設定 (テスト用)
func (p *CanastaPlayer) SetTookPozzetto(v bool) { p.tookPozzetto = v }

// canastaPlayerJSON is the JSON wire format for CanastaPlayer.
type canastaPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	Melds            []*CanastaMeld    `json:"ml"`
	Red3s            []*Card           `json:"r3"`
	HasInitMeld      bool              `json:"hi"`
	TookPozzetto     bool              `json:"tp,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (p *CanastaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(canastaPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		Melds:            p.melds,
		Red3s:            p.red3s,
		HasInitMeld:      p.hasInitMeld,
		TookPozzetto:     p.tookPozzetto,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CanastaPlayer) UnmarshalJSON(data []byte) error {
	var j canastaPlayerJSON
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
	p.melds = j.Melds
	if p.melds == nil {
		p.melds = make([]*CanastaMeld, 0)
	}
	p.red3s = j.Red3s
	if p.red3s == nil {
		p.red3s = make([]*Card, 0)
	}
	p.hasInitMeld = j.HasInitMeld
	p.tookPozzetto = j.TookPozzetto
	return nil
}
