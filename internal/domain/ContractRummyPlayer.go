//go:build !js || !wasm || extra

package domain

import "encoding/json"

// ContractRummyPlayer はコントラクトラミーのプレイヤー。手札・場のメルド・コントラクト達成状態を保持する。
type ContractRummyPlayer struct {
	*GamePlayer
	RoundScoreHolder
	melds         [][]*Card // 場に出したメルド
	contractMet   bool      // 当該ラウンドのコントラクトを達成済み（オープン済み）か
	contractIndex []int     // melds の各要素がコントラクトのどのスロットを満たすか（contractMet が true のときのみ意味を持つ）
}

// NewContractRummyPlayer コンストラクタ
func NewContractRummyPlayer(isHuman bool) *ContractRummyPlayer {
	return &ContractRummyPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		melds:      nil,
	}
}

// GetMelds 場に出したメルド一覧を返す
func (p *ContractRummyPlayer) GetMelds() [][]*Card { return p.melds }

// SetMelds メルドを差し替える（テスト用）
func (p *ContractRummyPlayer) SetMelds(m [][]*Card) { p.melds = m }

// GetMeldCount メルド数を取得
func (p *ContractRummyPlayer) GetMeldCount() int { return len(p.melds) }

// GetMeld 指定インデックスのメルドを取得
func (p *ContractRummyPlayer) GetMeld(i int) []*Card {
	if i < 0 || i >= len(p.melds) {
		return nil
	}
	return p.melds[i]
}

// AppendMeld 新しいメルドを追加する
func (p *ContractRummyPlayer) AppendMeld(meld []*Card) {
	cp := make([]*Card, len(meld))
	copy(cp, meld)
	p.melds = append(p.melds, cp)
}

// AddCardToMeld 指定インデックスのメルドに 1 枚追加する
func (p *ContractRummyPlayer) AddCardToMeld(i int, card *Card) bool {
	if i < 0 || i >= len(p.melds) {
		return false
	}
	p.melds[i] = append(p.melds[i], card)
	return true
}

// ClearMelds メルドを空にする
func (p *ContractRummyPlayer) ClearMelds() {
	p.melds = nil
	p.contractIndex = nil
	p.contractMet = false
}

// IsContractMet 当該ラウンドのコントラクト達成済みか
func (p *ContractRummyPlayer) IsContractMet() bool { return p.contractMet }

// SetContractMet コントラクト達成状態を設定する
func (p *ContractRummyPlayer) SetContractMet(met bool) { p.contractMet = met }

// GetContractIndex メルドとコントラクトスロットの対応を返す
func (p *ContractRummyPlayer) GetContractIndex() []int { return p.contractIndex }

// SetContractIndex メルドとコントラクトスロットの対応を設定する
func (p *ContractRummyPlayer) SetContractIndex(idx []int) {
	if idx == nil {
		p.contractIndex = nil
		return
	}
	cp := make([]int, len(idx))
	copy(cp, idx)
	p.contractIndex = cp
}

// ResetRound ラウンドをリセット（手札・スコア・メルド・終了状態を初期化）
func (p *ContractRummyPlayer) ResetRound() {
	resetRoundScored(p)
	p.ClearMelds()
}

// contractRummyPlayerJSON は ContractRummyPlayer の JSON 表現。
type contractRummyPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	Melds            [][]*Card         `json:"md"`
	ContractMet      bool              `json:"cm"`
	ContractIndex    []int             `json:"ci"`
}

// MarshalJSON implements json.Marshaler.
func (p *ContractRummyPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(contractRummyPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		Melds:            p.melds,
		ContractMet:      p.contractMet,
		ContractIndex:    p.contractIndex,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ContractRummyPlayer) UnmarshalJSON(data []byte) error {
	var j contractRummyPlayerJSON
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
