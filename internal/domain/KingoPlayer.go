//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
)

// エラー値。復元時の検証で使う。
var (
	errKingoNegativeChips = errors.New("kingo: chips must not be negative")
	errKingoNegativeBet   = errors.New("kingo: bet must not be negative")
	errKingoHandSize      = errors.New("kingo: a hand must hold three cards")
	errKingoRankMismatch  = errors.New("kingo: the recorded rank does not match the hand")
)

// KingoPlayer はキンゴの 1 席。
//
// **親も子も同じ型。** 親は席が回るだけで、持ち物が違うわけではない ──
// 親専用の型を作ると、親交代のたびに状態を移し替えることになる。
type KingoPlayer struct {
	Player
	ChipHolder

	isHuman bool
	name    string
	// bet はこのラウンドの張り額。**親は張らないので 0。**
	bet int
	// wonAmount はこのラウンドの収支 (負けなら負)。
	wonAmount int
}

// NewKingoPlayer は KingoPlayer を構築する。
func NewKingoPlayer(name string, chips int, isHuman bool) *KingoPlayer {
	p := &KingoPlayer{
		Player:  Player{cards: make([]*Card, 0, KingoHandSize)},
		isHuman: isHuman,
		name:    name,
	}
	p.SetChips(chips)
	return p
}

// NewKingoPlayersForTable は席数ぶんのプレイヤーを作る。席 0 が人間。
func NewKingoPlayersForTable(seats, chips int) []*KingoPlayer {
	players := make([]*KingoPlayer, 0, seats)
	for i := range seats {
		name := "CPU" + string(rune('0'+i))
		if i == 0 {
			name = "YOU"
		}
		players = append(players, NewKingoPlayer(name, chips, i == 0))
	}
	return players
}

// GetName は席の表示名を返す。
func (p *KingoPlayer) GetName() string { return p.name }

// GetIsHuman は人間の席かを返す。
func (p *KingoPlayer) GetIsHuman() bool { return p.isHuman }

// GetCards は手札を返す。
func (p *KingoPlayer) GetCards() []*Card { return p.cards }

// GetBet はこのラウンドの張り額を返す。
func (p *KingoPlayer) GetBet() int { return p.bet }

// SetBet は張り額を設定する。
func (p *KingoPlayer) SetBet(v int) { p.bet = v }

// GetWonAmount はこのラウンドの収支を返す。
func (p *KingoPlayer) GetWonAmount() int { return p.wonAmount }

// SetWonAmount は収支を設定する。
func (p *KingoPlayer) SetWonAmount(v int) { p.wonAmount = v }

// GetRank は手役を返す。
//
// **役は手札から毎回数える。** 別に持たせて配るたびに更新すると、更新を
// 1 か所忘れただけで画面の役と勝敗が食い違う。
func (p *KingoPlayer) GetRank() KingoRank { return KingoHandRank(p.cards) }

// ResetForRound は次のラウンドに向けて席の状態を戻す。
func (p *KingoPlayer) ResetForRound() {
	p.cards = p.cards[:0]
	p.bet = 0
	p.wonAmount = 0
}

// KingoHandRank は手札の役を返す。
//
// **同じ数字を何枚そろえたかだけを見る。** おいちょかぶは合計の下一桁で
// competing するが、キンゴはそろい方で決まる ── 合計を持ち込むと別のゲームに
// なる。
func KingoHandRank(cards []*Card) KingoRank {
	best := 0
	for _, n := range kingoValueCounts(cards) {
		best = max(best, n)
	}
	switch {
	case best >= 3:
		return KingoRankArashi
	case best == 2:
		return KingoRankPair
	default:
		return KingoRankNone
	}
}

// KingoMatchedValue はそろえた数字を返す (役なしなら最大の数字)。
//
// **同じ役どうしは、そろえた数字の大きいほうが勝つ。** ここが無いと、
// 同じ役の対戦がすべて引き分けになり、3 枚配る意味がほとんど消える。
func KingoMatchedValue(cards []*Card) int {
	counts := kingoValueCounts(cards)
	bestCount, bestValue := 0, 0
	for v, n := range counts {
		if n > bestCount || (n == bestCount && v > bestValue) {
			bestCount, bestValue = n, v
		}
	}
	return bestValue
}

// KingoKicker は役に使わなかった札のうち最大の数字を返す (無ければ 0)。
func KingoKicker(cards []*Card) int {
	matched := KingoMatchedValue(cards)
	best := 0
	for _, c := range cards {
		if c == nil || c.GetValue() == matched {
			continue
		}
		best = max(best, c.GetValue())
	}
	return best
}

// KingoCompare は 2 つの手を比べる (a が強ければ正、弱ければ負、同じなら 0)。
func KingoCompare(a, b []*Card) int {
	if ra, rb := KingoHandRank(a), KingoHandRank(b); ra != rb {
		return int(ra) - int(rb)
	}
	if va, vb := KingoMatchedValue(a), KingoMatchedValue(b); va != vb {
		return va - vb
	}
	return KingoKicker(a) - KingoKicker(b)
}

// kingoValueCounts は数字ごとの枚数を返す。
func kingoValueCounts(cards []*Card) map[int]int {
	counts := make(map[int]int, len(cards))
	for _, c := range cards {
		if c == nil {
			continue
		}
		counts[c.GetValue()]++
	}
	return counts
}

// kingoPlayerJSON は KingoPlayer の JSON 表現。
type kingoPlayerJSON struct {
	Chips     int     `json:"c"`
	Bet       int     `json:"b"`
	IsHuman   bool    `json:"h"`
	Name      string  `json:"n"`
	Cards     []*Card `json:"cd"`
	WonAmount int     `json:"w"`
	// Rank は表示用の控え。**復元時に手札と突き合わせて検証する。**
	Rank int `json:"rk"`
}

// MarshalJSON implements json.Marshaler.
func (p *KingoPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(kingoPlayerJSON{
		Chips: p.GetChips(), Bet: p.bet,
		IsHuman: p.isHuman, Name: p.name,
		Cards: p.cards, WonAmount: p.wonAmount,
		Rank: int(p.GetRank()),
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **控えの役が手札と合わない保存は通さない。** 役は本来 `GetRank` が手札から
// 毎回数えるので、控えがずれていても勝敗は正しく出る ── つまり**画面にだけ
// 嘘の役が出る**保存が作れてしまう。範囲検査では絶対に見つからない。
func (p *KingoPlayer) UnmarshalJSON(data []byte) error {
	var j kingoPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 {
		return errKingoNegativeChips
	}
	if j.Bet < 0 {
		return errKingoNegativeBet
	}
	if len(j.Cards) != 0 && len(j.Cards) != KingoHandSize {
		return errKingoHandSize
	}
	if j.Rank < int(KingoRankNone) || j.Rank > int(KingoRankMax) {
		return errKingoRankMismatch
	}
	if len(j.Cards) == KingoHandSize && KingoRank(j.Rank) != KingoHandRank(j.Cards) {
		return errKingoRankMismatch
	}
	p.SetChips(j.Chips)
	p.bet = j.Bet
	p.isHuman = j.IsHuman
	p.name = j.Name
	p.cards = j.Cards
	p.wonAmount = j.WonAmount
	return nil
}
