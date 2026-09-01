//go:build !js || !wasm || extra

package domain

import "encoding/json"

// SambaMeldKind はメルドの種類を表す。サンバではセットメルド（同ランク）と
// シーケンスメルド（同スート連番＝サンバ）の2種類がある。
type SambaMeldKind int

// SambaのメルドKind定数
const (
	// SambaMeldSet セットメルド (同ランク3枚以上、ワイルド可)
	SambaMeldSet SambaMeldKind = 0
	// SambaMeldSequence シーケンスメルド (同スートの連番3枚以上、ワイルド不可)
	SambaMeldSequence SambaMeldKind = 1
)

// SambaMeld テーブル上のメルド
type SambaMeld struct {
	Cards     []*Card       `json:"ca"`
	Kind      SambaMeldKind `json:"kd"`
	IsNatural bool          `json:"in"` // セットでワイルドを含まない場合 true (シーケンスは常に true)
}

// SambaCanastaSize はメルドがカナスタ／サンバとして完成する枚数。
// Web の `SAMBA_CANASTA_SIZE` と同じ値で、両者の食い違いは
// frontend/scripts/check-samba-canasta-size.mjs が落とす。
const SambaCanastaSize = 7

// IsCompleted は7枚以上の完成メルド（カナスタまたはサンバ）かどうかを返す。
// 上がり条件・チーム集計に使用する。
func (m *SambaMeld) IsCompleted() bool {
	return len(m.Cards) >= SambaCanastaSize
}

// IsCanasta は7枚以上のセットメルド（カナスタ）かどうかを返す。
func (m *SambaMeld) IsCanasta() bool {
	return m.Kind == SambaMeldSet && len(m.Cards) >= SambaCanastaSize
}

// IsSamba は7枚のシーケンスメルド（サンバ）かどうかを返す。
func (m *SambaMeld) IsSamba() bool {
	return m.Kind == SambaMeldSequence && len(m.Cards) >= SambaCanastaSize
}

// GetRank はセットメルドのランク（ナチュラルカードのランク）を返す。
// シーケンスメルドでは最初のカードの値を返す。
func (m *SambaMeld) GetRank() int {
	for _, c := range m.Cards {
		if !SambaIsWild(c) {
			return c.GetValue()
		}
	}
	return 0
}

// SuitDesign はシーケンスメルドのスートを返す（ワイルドを含まない先頭カードのスート）。
// セットメルドでは意味を持たない。
func (m *SambaMeld) SuitDesign() int {
	for _, c := range m.Cards {
		if !SambaIsWild(c) {
			return c.GetDesign()
		}
	}
	return -1
}

// SambaPlayer サンバプレイヤークラス
type SambaPlayer struct {
	*GamePlayer
	RoundScoreHolder
	team        int          // 所属チーム (0 = 席0・2, 1 = 席1・3)
	melds       []*SambaMeld // このプレイヤーが出したメルド
	red3s       []*Card      // 場に出した赤3
	hasInitMeld bool         // 初回メルド済みフラグ
}

// NewSambaPlayer コンストラクタ
func NewSambaPlayer(isHuman bool, team int) *SambaPlayer {
	return &SambaPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
		melds:      make([]*SambaMeld, 0),
		red3s:      make([]*Card, 0),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・メルド・赤3を初期化）
func (p *SambaPlayer) ResetRound() {
	p.SetRoundScore(0)
	p.Reset()
	p.SetIsFinished(false)
	p.melds = make([]*SambaMeld, 0)
	p.red3s = make([]*Card, 0)
	p.hasInitMeld = false
}

// GetTeam 所属チームを取得
func (p *SambaPlayer) GetTeam() int { return p.team }

// SetTeam 所属チームを設定 (テスト用)
func (p *SambaPlayer) SetTeam(team int) { p.team = team }

// GetMelds メルドを取得
func (p *SambaPlayer) GetMelds() []*SambaMeld { return p.melds }

// SetMelds メルドを設定 (テスト用)
func (p *SambaPlayer) SetMelds(melds []*SambaMeld) { p.melds = melds }

// AddMeld メルドを追加
func (p *SambaPlayer) AddMeld(meld *SambaMeld) {
	p.melds = append(p.melds, meld)
}

// GetRed3s 赤3を取得
func (p *SambaPlayer) GetRed3s() []*Card { return p.red3s }

// SetRed3s 赤3を設定 (テスト用)
func (p *SambaPlayer) SetRed3s(red3s []*Card) { p.red3s = red3s }

// AddRed3 赤3を追加
func (p *SambaPlayer) AddRed3(card *Card) {
	p.red3s = append(p.red3s, card)
}

// GetHasInitMeld 初回メルド済みフラグ取得
func (p *SambaPlayer) GetHasInitMeld() bool { return p.hasInitMeld }

// SetHasInitMeld 初回メルド済みフラグ設定
func (p *SambaPlayer) SetHasInitMeld(v bool) { p.hasInitMeld = v }

// CompletedMeldCount は完成メルド（7枚以上のカナスタまたはサンバ）の数を返す。
func (p *SambaPlayer) CompletedMeldCount() int {
	n := 0
	for _, m := range p.melds {
		if m.IsCompleted() {
			n++
		}
	}
	return n
}

// HasCanasta はセットのカナスタ（7枚以上）を持っているか。
func (p *SambaPlayer) HasCanasta() bool {
	for _, m := range p.melds {
		if m.IsCanasta() {
			return true
		}
	}
	return false
}

// HasSamba はシーケンスのサンバ（7枚）を持っているか。
func (p *SambaPlayer) HasSamba() bool {
	for _, m := range p.melds {
		if m.IsSamba() {
			return true
		}
	}
	return false
}

// sambaPlayerJSON is the JSON wire format for SambaPlayer.
type sambaPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	Team             int               `json:"tm"`
	Melds            []*SambaMeld      `json:"ml"`
	Red3s            []*Card           `json:"r3"`
	HasInitMeld      bool              `json:"hi"`
}

// MarshalJSON implements json.Marshaler.
func (p *SambaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(sambaPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		Team:             p.team,
		Melds:            p.melds,
		Red3s:            p.red3s,
		HasInitMeld:      p.hasInitMeld,
	})
}

// UnmarshalJSON implements json.Unmarshaler. Meld kinds are clamped to a valid
// value; the team is validated against the team count by Samba.UnmarshalJSON.
func (p *SambaPlayer) UnmarshalJSON(data []byte) error {
	var j sambaPlayerJSON
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
	p.team = j.Team
	// nil メルドは除外する (残すと CompletedMeldCount / HasCanasta / スコア計算で
	// nil ポインタ参照によりパニックする)。Kind は既知値へクランプする。
	p.melds = make([]*SambaMeld, 0, len(j.Melds))
	for _, m := range j.Melds {
		if m == nil {
			continue
		}
		if m.Kind != SambaMeldSet && m.Kind != SambaMeldSequence {
			m.Kind = SambaMeldSet
		}
		p.melds = append(p.melds, m)
	}
	p.red3s = j.Red3s
	if p.red3s == nil {
		p.red3s = make([]*Card, 0)
	}
	p.hasInitMeld = j.HasInitMeld
	return nil
}
