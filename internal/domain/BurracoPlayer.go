package domain

import "encoding/json"

// BurracoMeld テーブル上のメルド
type BurracoMeld struct {
	Cards     []*Card `json:"ca"`
	IsNatural bool    `json:"in"` // ワイルドカードを含まない場合 true
}

// IsBurraco 7枚以上のメルドかどうか
func (m *BurracoMeld) IsBurraco() bool {
	return len(m.Cards) >= 7
}

// GetRank メルドのランク（ナチュラルカードのランク）を返す
func (m *BurracoMeld) GetRank() int {
	for _, c := range m.Cards {
		if !BurracoIsWild(c) {
			return c.GetValue()
		}
	}
	return 0
}

// BurracoPlayer ブラーコプレイヤークラス
type BurracoPlayer struct {
	*GamePlayer
	RoundScoreHolder
	melds        []*BurracoMeld // テーブル上のメルド
	red3s        []*Card        // 場に出した赤3
	hasInitMeld  bool           // 初回メルド済みフラグ
	tookPozzetto bool           // ポゼット（予備手札）を獲得済みか
}

// NewBurracoPlayer コンストラクタ
func NewBurracoPlayer(isHuman bool) *BurracoPlayer {
	return &BurracoPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		melds:      make([]*BurracoMeld, 0),
		red3s:      make([]*Card, 0),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・メルド・赤3を初期化）
func (p *BurracoPlayer) ResetRound() {
	p.SetRoundScore(0)
	p.Reset()
	p.SetIsFinished(false)
	p.melds = make([]*BurracoMeld, 0)
	p.red3s = make([]*Card, 0)
	p.hasInitMeld = false
	p.tookPozzetto = false
}

// GetMelds メルドを取得
func (p *BurracoPlayer) GetMelds() []*BurracoMeld { return p.melds }

// SetMelds メルドを設定 (テスト用)
func (p *BurracoPlayer) SetMelds(melds []*BurracoMeld) { p.melds = melds }

// AddMeld メルドを追加
func (p *BurracoPlayer) AddMeld(meld *BurracoMeld) {
	p.melds = append(p.melds, meld)
}

// GetRed3s 赤3を取得
func (p *BurracoPlayer) GetRed3s() []*Card { return p.red3s }

// SetRed3s 赤3を設定 (テスト用)
func (p *BurracoPlayer) SetRed3s(red3s []*Card) { p.red3s = red3s }

// AddRed3 赤3を追加
func (p *BurracoPlayer) AddRed3(card *Card) {
	p.red3s = append(p.red3s, card)
}

// GetHasInitMeld 初回メルド済みフラグ取得
func (p *BurracoPlayer) GetHasInitMeld() bool { return p.hasInitMeld }

// SetHasInitMeld 初回メルド済みフラグ設定
func (p *BurracoPlayer) SetHasInitMeld(v bool) { p.hasInitMeld = v }

// GetTookPozzetto ポゼット獲得済みフラグ取得
func (p *BurracoPlayer) GetTookPozzetto() bool { return p.tookPozzetto }

// SetTookPozzetto ポゼット獲得済みフラグ設定 (テスト用)
func (p *BurracoPlayer) SetTookPozzetto(v bool) { p.tookPozzetto = v }

// HasBurraco ブラーコ（7枚以上のメルド）を持っているか
func (p *BurracoPlayer) HasBurraco() bool {
	for _, m := range p.melds {
		if m.IsBurraco() {
			return true
		}
	}
	return false
}

// burracoPlayerJSON is the JSON wire format for BurracoPlayer.
type burracoPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	Melds            []*BurracoMeld    `json:"ml"`
	Red3s            []*Card           `json:"r3"`
	HasInitMeld      bool              `json:"hi"`
	TookPozzetto     bool              `json:"tp"`
}

// MarshalJSON implements json.Marshaler.
func (p *BurracoPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(burracoPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		Melds:            p.melds,
		Red3s:            p.red3s,
		HasInitMeld:      p.hasInitMeld,
		TookPozzetto:     p.tookPozzetto,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BurracoPlayer) UnmarshalJSON(data []byte) error {
	var j burracoPlayerJSON
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
		p.melds = make([]*BurracoMeld, 0)
	}
	p.red3s = j.Red3s
	if p.red3s == nil {
		p.red3s = make([]*Card, 0)
	}
	p.hasInitMeld = j.HasInitMeld
	p.tookPozzetto = j.TookPozzetto
	return nil
}
