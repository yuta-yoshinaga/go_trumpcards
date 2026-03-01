package domain

import "sort"

// SevensMaxJokerCount 7並べで許可されるジョーカー最大枚数
const SevensMaxJokerCount = 2

// SevensMaxPasses 7並べで各プレイヤーに許可されるパス回数
const SevensMaxPasses = 5

// SevensPlayer 7並べプレイヤークラス
type SevensPlayer struct {
	Player
	isHuman      bool
	isFinished   bool
	isEliminated bool // パス切れによる失格フラグ
	rank         int  // -1 = 未確定, 1 = 1位, ...
	passesUsed   int  // 使用済みパス回数
	maxPasses    int  // 最大パス回数
}

// NewSevensPlayer コンストラクタ
func NewSevensPlayer(isHuman bool) *SevensPlayer {
	return &SevensPlayer{
		Player:     Player{cards: make([]*Card, 0)},
		isHuman:    isHuman,
		isFinished: false,
		rank:       -1,
		passesUsed: 0,
		maxPasses:  SevensMaxPasses,
	}
}

// GetIsHuman 人間プレイヤーかどうか
func (p *SevensPlayer) GetIsHuman() bool { return p.isHuman }

// GetIsFinished 終了しているかどうか (上がりまたは失格)
func (p *SevensPlayer) GetIsFinished() bool { return p.isFinished }

// SetIsFinished 終了状態設定
func (p *SevensPlayer) SetIsFinished(v bool) { p.isFinished = v }

// GetIsEliminated 失格かどうか
func (p *SevensPlayer) GetIsEliminated() bool { return p.isEliminated }

// SetIsEliminated 失格フラグ設定
func (p *SevensPlayer) SetIsEliminated(v bool) { p.isEliminated = v }

// GetRank ランク取得 (-1 = 未確定)
func (p *SevensPlayer) GetRank() int { return p.rank }

// SetRank ランク設定
func (p *SevensPlayer) SetRank(r int) { p.rank = r }

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

// RemoveCard 指定インデックスのカードを手札から取り除いて返す
func (p *SevensPlayer) RemoveCard(idx int) *Card {
	if idx < 0 || idx >= len(p.cards) {
		return nil
	}
	card := p.cards[idx]
	p.cards = append(p.cards[:idx], p.cards[idx+1:]...)
	return card
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
