//go:build !js || !wasm || extra

package domain

import "encoding/json"

// BoliviaMeldKind はメルドの種類を表す。
//
// **3 種類ある。** サンバはセットとシーケンスの 2 つだが、ボリビアは
// **ワイルドだけのメルド**を認めるのが唯一無二の点で、それが 7 枚揃ったものを
// 「ボリビア」と呼ぶ ── ゲーム名の由来。
type BoliviaMeldKind int

// BoliviaのメルドKind定数
const (
	// BoliviaMeldSet セットメルド (同ランク3枚以上、ワイルド可)
	BoliviaMeldSet BoliviaMeldKind = 0
	// BoliviaMeldEscalera エスカレラ (同スートの連番3枚以上、**ワイルド不可**)。
	// スペイン語の「梯子」。**3 だけで作るメルドではない。**
	BoliviaMeldEscalera BoliviaMeldKind = 1
	// BoliviaMeldWild ワイルドのみのメルド (2 とジョーカーだけ 3 枚以上)。
	// 7 枚揃うと「ボリビア」。
	BoliviaMeldWild BoliviaMeldKind = 2
)

// BoliviaMeld テーブル上のメルド
type BoliviaMeld struct {
	Cards     []*Card         `json:"ca"`
	Kind      BoliviaMeldKind `json:"kd"`
	IsNatural bool            `json:"in"` // セットでワイルドを含まない場合 true (シーケンスは常に true)
}

// IsCompleted は7枚以上の完成メルド（カナスタまたはボリビア）かどうかを返す。
// 上がり条件・チーム集計に使用する。
func (m *BoliviaMeld) IsCompleted() bool {
	return len(m.Cards) >= 7
}

// IsCanasta は7枚以上のセットメルド（カナスタ）かどうかを返す。
func (m *BoliviaMeld) IsCanasta() bool {
	return m.Kind == BoliviaMeldSet && len(m.Cards) >= 7
}

// IsEscalera は7枚のエスカレラ（ワイルド無しの同スート連番）かどうかを返す。
//
// **上がるには最低 1 つこれが要る。** カナスタ 2 つでは上がれない。
func (m *BoliviaMeld) IsEscalera() bool {
	return m.Kind == BoliviaMeldEscalera && len(m.Cards) >= BoliviaCanastaSize
}

// IsBoliviaCanasta は7枚のワイルドメルド（ボリビア）かどうかを返す。
func (m *BoliviaMeld) IsBoliviaCanasta() bool {
	return m.Kind == BoliviaMeldWild && len(m.Cards) >= BoliviaCanastaSize
}

// GetRank はセットメルドのランク（ナチュラルカードのランク）を返す。
// シーケンスメルドでは最初のカードの値を返す。
func (m *BoliviaMeld) GetRank() int {
	for _, c := range m.Cards {
		if !BoliviaIsWild(c) {
			return c.GetValue()
		}
	}
	return 0
}

// SuitDesign はシーケンスメルドのスートを返す（ワイルドを含まない先頭カードのスート）。
// セットメルドでは意味を持たない。
func (m *BoliviaMeld) SuitDesign() int {
	for _, c := range m.Cards {
		if !BoliviaIsWild(c) {
			return c.GetDesign()
		}
	}
	return -1
}

// BoliviaPlayer ボリビアプレイヤークラス
type BoliviaPlayer struct {
	*GamePlayer
	RoundScoreHolder
	team        int            // 所属チーム (0 = 席0・2, 1 = 席1・3)
	melds       []*BoliviaMeld // このプレイヤーが出したメルド
	red3s       []*Card        // 場に出した赤3
	hasInitMeld bool           // 初回メルド済みフラグ
}

// NewBoliviaPlayer コンストラクタ
func NewBoliviaPlayer(isHuman bool, team int) *BoliviaPlayer {
	return &BoliviaPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
		melds:      make([]*BoliviaMeld, 0),
		red3s:      make([]*Card, 0),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・メルド・赤3を初期化）
func (p *BoliviaPlayer) ResetRound() {
	p.SetRoundScore(0)
	p.Reset()
	p.SetIsFinished(false)
	p.melds = make([]*BoliviaMeld, 0)
	p.red3s = make([]*Card, 0)
	p.hasInitMeld = false
}

// GetTeam 所属チームを取得
func (p *BoliviaPlayer) GetTeam() int { return p.team }

// SetTeam 所属チームを設定 (テスト用)
func (p *BoliviaPlayer) SetTeam(team int) { p.team = team }

// GetMelds メルドを取得
func (p *BoliviaPlayer) GetMelds() []*BoliviaMeld { return p.melds }

// SetMelds メルドを設定 (テスト用)
func (p *BoliviaPlayer) SetMelds(melds []*BoliviaMeld) { p.melds = melds }

// AddMeld メルドを追加
func (p *BoliviaPlayer) AddMeld(meld *BoliviaMeld) {
	p.melds = append(p.melds, meld)
}

// GetRed3s 赤3を取得
func (p *BoliviaPlayer) GetRed3s() []*Card { return p.red3s }

// SetRed3s 赤3を設定 (テスト用)
func (p *BoliviaPlayer) SetRed3s(red3s []*Card) { p.red3s = red3s }

// AddRed3 赤3を追加
func (p *BoliviaPlayer) AddRed3(card *Card) {
	p.red3s = append(p.red3s, card)
}

// GetHasInitMeld 初回メルド済みフラグ取得
func (p *BoliviaPlayer) GetHasInitMeld() bool { return p.hasInitMeld }

// SetHasInitMeld 初回メルド済みフラグ設定
func (p *BoliviaPlayer) SetHasInitMeld(v bool) { p.hasInitMeld = v }

// CompletedMeldCount は完成メルド（7枚以上のカナスタまたはボリビア）の数を返す。
func (p *BoliviaPlayer) CompletedMeldCount() int {
	n := 0
	for _, m := range p.melds {
		if m.IsCompleted() {
			n++
		}
	}
	return n
}

// HasCanasta はセットのカナスタ（7枚以上）を持っているか。
func (p *BoliviaPlayer) HasCanasta() bool {
	for _, m := range p.melds {
		if m.IsCanasta() {
			return true
		}
	}
	return false
}

// HasEscalera は完成したエスカレラ (ワイルド無しの同スート 7 枚連番) を
// 持っているかを返す。**上がりにはチームで最低 1 本要る。**
func (p *BoliviaPlayer) HasEscalera() bool {
	for _, m := range p.melds {
		if m.IsEscalera() {
			return true
		}
	}
	return false
}

// HasBolivia は完成したボリビア (ワイルド 7 枚) を持っているかを返す。
//
// **エスカレラと混同しないこと。** 上がりに要るのはエスカレラのほうで、
// ボリビアは点が重いだけ。
func (p *BoliviaPlayer) HasBolivia() bool {
	for _, m := range p.melds {
		if m.IsBoliviaCanasta() {
			return true
		}
	}
	return false
}

// boliviaPlayerJSON is the JSON wire format for BoliviaPlayer.
type boliviaPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	Team             int               `json:"tm"`
	Melds            []*BoliviaMeld    `json:"ml"`
	Red3s            []*Card           `json:"r3"`
	HasInitMeld      bool              `json:"hi"`
}

// MarshalJSON implements json.Marshaler.
func (p *BoliviaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(boliviaPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		Team:             p.team,
		Melds:            p.melds,
		Red3s:            p.red3s,
		HasInitMeld:      p.hasInitMeld,
	})
}

// UnmarshalJSON implements json.Unmarshaler. Meld kinds are clamped to a valid
// value; the team is validated against the team count by Bolivia.UnmarshalJSON.
func (p *BoliviaPlayer) UnmarshalJSON(data []byte) error {
	var j boliviaPlayerJSON
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
	p.melds = make([]*BoliviaMeld, 0, len(j.Melds))
	for _, m := range j.Melds {
		if m == nil {
			continue
		}
		// **種類は 3 つ。** クローン元は 2 つしか無いので、この許可リストに
		// ワイルドを足し忘れると、**保存して読み戻すたびにボリビアが
		// セットに化ける** ── Worker は毎リクエストで復元するので、
		// 2500 点のメルドが次の 1 手で 300 点になる (レビュー指摘)。
		if m.Kind != BoliviaMeldSet && m.Kind != BoliviaMeldEscalera && m.Kind != BoliviaMeldWild {
			m.Kind = BoliviaMeldSet
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
