//go:build !js || !wasm || extra

package domain

import "encoding/json"

// CariocaPlayer はカリオカのプレイヤー。手札・場のメルド・コントラクト達成状態を保持する。
type CariocaPlayer struct {
	*GamePlayer
	RoundScoreHolder
	melds         [][]*Card // 場に出したメルド
	contractMet   bool      // 当該ラウンドのコントラクトを達成済み（オープン済み）か
	contractIndex []int     // melds の各要素がコントラクトのどのスロットを満たすか（contractMet が true のときのみ意味を持つ）
}

// NewCariocaPlayer コンストラクタ
func NewCariocaPlayer(isHuman bool) *CariocaPlayer {
	return &CariocaPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		melds:      nil,
	}
}

// GetMelds 場に出したメルド一覧を返す
func (p *CariocaPlayer) GetMelds() [][]*Card { return p.melds }

// SetMelds メルドを差し替える（テスト用）
func (p *CariocaPlayer) SetMelds(m [][]*Card) { p.melds = m }

// GetMeldCount メルド数を取得
func (p *CariocaPlayer) GetMeldCount() int { return len(p.melds) }

// GetMeld 指定インデックスのメルドを取得
func (p *CariocaPlayer) GetMeld(i int) []*Card {
	if i < 0 || i >= len(p.melds) {
		return nil
	}
	return p.melds[i]
}

// AppendMeld 新しいメルドを追加する
func (p *CariocaPlayer) AppendMeld(meld []*Card) {
	cp := make([]*Card, len(meld))
	copy(cp, meld)
	p.melds = append(p.melds, cp)
}

// AddCardToMeld 指定インデックスのメルドに 1 枚追加する
func (p *CariocaPlayer) AddCardToMeld(i int, card *Card) bool {
	if i < 0 || i >= len(p.melds) {
		return false
	}
	p.melds[i] = append(p.melds[i], card)
	return true
}

// ClearMelds メルドを空にする
func (p *CariocaPlayer) ClearMelds() {
	p.melds = nil
	p.contractIndex = nil
	p.contractMet = false
}

// IsContractMet 当該ラウンドのコントラクト達成済みか
func (p *CariocaPlayer) IsContractMet() bool { return p.contractMet }

// SetContractMet コントラクト達成状態を設定する
func (p *CariocaPlayer) SetContractMet(met bool) { p.contractMet = met }

// GetContractIndex メルドとコントラクトスロットの対応を返す
func (p *CariocaPlayer) GetContractIndex() []int { return p.contractIndex }

// SetContractIndex メルドとコントラクトスロットの対応を設定する
func (p *CariocaPlayer) SetContractIndex(idx []int) {
	if idx == nil {
		p.contractIndex = nil
		return
	}
	cp := make([]int, len(idx))
	copy(cp, idx)
	p.contractIndex = cp
}

// ResetRound ラウンドをリセット（手札・スコア・メルド・終了状態を初期化）
func (p *CariocaPlayer) ResetRound() {
	resetRoundScored(p)
	p.ClearMelds()
}

// cariocaPlayerJSON は CariocaPlayer の JSON 表現。
type cariocaPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	Melds            [][]*Card         `json:"md"`
	ContractMet      bool              `json:"cm"`
	ContractIndex    []int             `json:"ci"`
}

// MarshalJSON implements json.Marshaler.
func (p *CariocaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(cariocaPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		Melds:            p.melds,
		ContractMet:      p.contractMet,
		ContractIndex:    p.contractIndex,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CariocaPlayer) UnmarshalJSON(data []byte) error {
	var j cariocaPlayerJSON
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
	p.contractMet = j.ContractMet
	p.contractIndex = j.ContractIndex
	return nil
}
