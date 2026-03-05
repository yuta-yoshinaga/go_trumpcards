package domain

import "sort"

// SevensMaxJokerCount 7並べで許可されるジョーカー最大枚数
const SevensMaxJokerCount = 2

// SevensMaxPasses 7並べで各プレイヤーに許可されるパス回数
const SevensMaxPasses = 5

// SevensPlayer 7並べプレイヤークラス
type SevensPlayer struct {
	*RankedGamePlayer
	isEliminated bool // パス切れによる失格フラグ
	passesUsed   int  // 使用済みパス回数
	maxPasses    int  // 最大パス回数
}

// NewSevensPlayer コンストラクタ
func NewSevensPlayer(isHuman bool) *SevensPlayer {
	return &SevensPlayer{
		RankedGamePlayer: NewRankedGamePlayer(isHuman),
		maxPasses:        SevensMaxPasses,
	}
}

// GetIsEliminated 失格かどうか
func (p *SevensPlayer) GetIsEliminated() bool { return p.isEliminated }

// SetIsEliminated 失格フラグ設定
func (p *SevensPlayer) SetIsEliminated(v bool) { p.isEliminated = v }

// GetPassesUsed 使用済みパス回数取得
func (p *SevensPlayer) GetPassesUsed() int { return p.passesUsed }

// IncrPassesUsed パス回数をインクリメント
func (p *SevensPlayer) IncrPassesUsed() { p.passesUsed++ }

// GetMaxPasses 最大パス回数取得
func (p *SevensPlayer) GetMaxPasses() int { return p.maxPasses }

// SetMaxPasses 最大パス回数設定 (0 = 無制限)
func (p *SevensPlayer) SetMaxPasses(n int) { p.maxPasses = n }

// CanPass パス可能かどうか (maxPasses == 0 は無制限, それ以外は使用済みパス < 最大パス)
func (p *SevensPlayer) CanPass() bool {
	if p.maxPasses == 0 {
		return true
	}
	return p.passesUsed < p.maxPasses
}

// ResetPasses パス回数リセット
func (p *SevensPlayer) ResetPasses() {
	p.passesUsed = 0
}

// RemoveSevens 手札からすべての7を取り除いて返す
func (p *SevensPlayer) RemoveSevens() []*Card {
	removed := make([]*Card, 0)
	newCards := make([]*Card, 0, len(p.cards))
	for _, c := range p.cards {
		if c.GetValue() == 7 {
			removed = append(removed, c)
		} else {
			newCards = append(newCards, c)
		}
	}
	p.cards = newCards
	return removed
}

// SortCards カードをスートごと・値の昇順でソート
func (p *SevensPlayer) SortCards() {
	sort.Slice(p.cards, func(i, j int) bool {
		if p.cards[i].GetDesign() != p.cards[j].GetDesign() {
			return p.cards[i].GetDesign() < p.cards[j].GetDesign()
		}
		return p.cards[i].GetValue() < p.cards[j].GetValue()
	})
}
