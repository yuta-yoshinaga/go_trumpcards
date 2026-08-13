//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// エラー値。復元時の検証で使う。
var (
	errBaseballNegativeChips = errors.New("baseballpoker: chips must not be negative")
	errBaseballNegativeBet   = errors.New("baseballpoker: bet must not be negative")
	errBaseballHandSize      = errors.New("baseballpoker: a hand cannot hold more cards than the deal allows")
	errBaseballFaceUpCount   = errors.New("baseballpoker: the face-up flags do not match the hand")
	errBaseballBonusCount    = errors.New("baseballpoker: bonus card count is out of range")
)

// BaseballPokerPlayer はベースボールポーカーの 1 席。
//
// **手札の枚数が席ごとに違う。** 表向きの 4 でボーナス札が配られるので、
// 7 枚で終わる席と 8 枚以上になる席が同じ卓に並ぶ ── 枚数を定数だと思って
// 書くと、ボーナスをもらった席の役が 1 枚ぶん足りない状態で評価される。
type BaseballPokerPlayer struct {
	Player
	ChipHolder
	bettingPlayerBase

	isHuman bool
	name    string
	// faceUp は cards と同じ長さで、その札が表向きかを持つ。
	//
	// **札そのものに向きを持たせない。** `Card` の裏表フラグは描画用で、
	// 「相手に見えているか」とは別物 ── 混ぜると伏せ札がワイヤに乗る。
	faceUp []bool
	// bonusCards は表の 4 で受け取った追加札の枚数。
	bonusCards int
	handRank   int
	usedWild   bool
	bestHand   []*Card
}

// NewBaseballPokerPlayer は BaseballPokerPlayer を構築する。
func NewBaseballPokerPlayer(name string, chips int, isHuman bool) *BaseballPokerPlayer {
	p := &BaseballPokerPlayer{
		Player:  Player{cards: make([]*Card, 0, BaseballBaseCards+BaseballMaxBonusCards)},
		faceUp:  make([]bool, 0, BaseballBaseCards+BaseballMaxBonusCards),
		isHuman: isHuman,
		name:    name,
	}
	p.SetChips(chips)
	return p
}

// NewBaseballPokerPlayersForTable は席数ぶんのプレイヤーを作る。席 0 が人間。
func NewBaseballPokerPlayersForTable(seats, chips int) []*BaseballPokerPlayer {
	players := make([]*BaseballPokerPlayer, 0, seats)
	for i := range seats {
		name := "CPU" + string(rune('0'+i))
		if i == 0 {
			name = "YOU"
		}
		players = append(players, NewBaseballPokerPlayer(name, chips, i == 0))
	}
	return players
}

// GetName は席の表示名を返す。
func (p *BaseballPokerPlayer) GetName() string { return p.name }

// GetIsHuman は人間の席かを返す。
func (p *BaseballPokerPlayer) GetIsHuman() bool { return p.isHuman }

// GetCards は手札を返す。
func (p *BaseballPokerPlayer) GetCards() []*Card { return p.cards }

// GetFaceUp は各札が表向きかを返す (cards と同じ並び)。
func (p *BaseballPokerPlayer) GetFaceUp() []bool { return p.faceUp }

// GetBonusCards は表の 4 で受け取った追加札の枚数を返す。
func (p *BaseballPokerPlayer) GetBonusCards() int { return p.bonusCards }

// GetHandRank は役のランクを返す。
func (p *BaseballPokerPlayer) GetHandRank() int { return p.handRank }

// GetUsedWild は役にワイルドを使ったかを返す。
func (p *BaseballPokerPlayer) GetUsedWild() bool { return p.usedWild }

// GetBestHand は選ばれた最良の 5 枚を返す。
func (p *BaseballPokerPlayer) GetBestHand() []*Card { return p.bestHand }

// AddDealtCard は札を 1 枚加える。faceUp が表向きかどうか。
func (p *BaseballPokerPlayer) AddDealtCard(c *Card, faceUp bool) {
	p.cards = append(p.cards, c)
	p.faceUp = append(p.faceUp, faceUp)
}

// AddBonusCard は表の 4 で得た伏せ札を 1 枚加える。
//
// **ボーナスは必ず伏せて配る。** 表で配ると、そのボーナス札がまた 4 や 3 で
// あった場合にイベントが連鎖して、1 枚のイベントが際限なく増える。
func (p *BaseballPokerPlayer) AddBonusCard(c *Card) {
	p.AddDealtCard(c, false)
	p.bonusCards++
}

// FaceUpCards は相手からも見えている札だけを返す。
func (p *BaseballPokerPlayer) FaceUpCards() []*Card {
	out := make([]*Card, 0, len(p.cards))
	for i, c := range p.cards {
		if i < len(p.faceUp) && p.faceUp[i] {
			out = append(out, c)
		}
	}
	return out
}

// EvaluateBest は手札全部から最良の 5 枚を選ぶ。
//
// **3 と 9 はワイルド。** 判定は `evalWildHand` に任せ、ここでは組合せだけを
// 回す ── 枚数が席ごとに違うので、`combinations` に実際の手札長を渡す。
func (p *BaseballPokerPlayer) EvaluateBest() int {
	if len(p.cards) < BaseballHandSize {
		p.handRank, p.usedWild, p.bestHand = PokerHandHighCard, false, nil
		return p.handRank
	}

	bestRank := -1
	bestWild := false
	var bestCards []*Card
	for _, combo := range combinations(p.cards, BaseballHandSize) {
		rank, usedWild := evalWildHand(combo, BaseballIsWild)
		if rank > bestRank || (rank == bestRank && compareHighCardsSlice(combo, bestCards) > 0) {
			bestRank, bestWild = rank, usedWild
			bestCards = make([]*Card, BaseballHandSize)
			copy(bestCards, combo)
		}
		// **ファイブカードより上は無い。** 手札は最大 11 枚まで増えるので、
		// 最良が確定したあとも残りの組合せを回すと、ワイルドの多い手ほど
		// 高くつく ── いちばん強い手がいちばん遅い、という逆立ちになる。
		if bestRank == PokerHandFiveOfAKind {
			break
		}
	}
	p.handRank, p.usedWild, p.bestHand = bestRank, bestWild, bestCards
	return p.handRank
}

// ResetForHand は次のハンドに向けて席の状態を戻す。
func (p *BaseballPokerPlayer) ResetForHand() {
	p.cards = p.cards[:0]
	p.faceUp = p.faceUp[:0]
	p.bonusCards = 0
	p.SetFolded(false)
	p.SetAllIn(false)
	p.SetCurrentBet(0)
	p.handRank = 0
	p.usedWild = false
	p.bestHand = nil
}

// baseballPlayerJSON は BaseballPokerPlayer の JSON 表現。
type baseballPlayerJSON struct {
	Chips      int     `json:"c"`
	CurrentBet int     `json:"b"`
	IsHuman    bool    `json:"h"`
	Name       string  `json:"n"`
	Folded     bool    `json:"f"`
	AllIn      bool    `json:"a"`
	Cards      []*Card `json:"cd"`
	FaceUp     []bool  `json:"fu"`
	BonusCards int     `json:"bn"`
	HandRank   int     `json:"hr"`
	UsedWild   bool    `json:"uw"`
	BestHand   []*Card `json:"bh"`
}

// MarshalJSON implements json.Marshaler.
func (p *BaseballPokerPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(baseballPlayerJSON{
		Chips: p.GetChips(), CurrentBet: p.GetCurrentBet(),
		IsHuman: p.isHuman, Name: p.name,
		Folded: p.GetFolded(), AllIn: p.GetAllIn(),
		Cards: p.cards, FaceUp: p.faceUp, BonusCards: p.bonusCards,
		HandRank: p.handRank, UsedWild: p.usedWild, BestHand: p.bestHand,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **向きの列は手札と同じ長さでなければならない。** ここがずれると、伏せ札を
// 表と誤って公開するか、表札を伏せて CPU の判断材料が消える ── どちらも
// 症状が出ないまま勝負だけが静かに変わる。枚数の範囲検査だけでは通ってしまう。
func (p *BaseballPokerPlayer) UnmarshalJSON(data []byte) error {
	var j baseballPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 {
		return errBaseballNegativeChips
	}
	if j.CurrentBet < 0 {
		return errBaseballNegativeBet
	}
	if len(j.Cards) > BaseballBaseCards+BaseballMaxBonusCards {
		return errBaseballHandSize
	}
	if len(j.FaceUp) != len(j.Cards) {
		return errBaseballFaceUpCount
	}
	if j.BonusCards < 0 || j.BonusCards > BaseballMaxBonusCards {
		return errBaseballBonusCount
	}
	// **ボーナスの枚数は手札の枚数とも整合していなければならない。**
	// 「7 枚しか無いのにボーナス 2 枚」は範囲検査を素通りする。
	if len(j.Cards) > 0 && j.BonusCards > max(0, len(j.Cards)-BaseballDownCards-1) {
		return errBaseballBonusCount
	}
	p.SetChips(j.Chips)
	p.SetCurrentBet(j.CurrentBet)
	p.isHuman = j.IsHuman
	p.name = j.Name
	p.SetFolded(j.Folded)
	p.SetAllIn(j.AllIn)
	p.cards = j.Cards
	p.faceUp = j.FaceUp
	p.bonusCards = j.BonusCards
	p.handRank = j.HandRank
	p.usedWild = j.UsedWild
	p.bestHand = j.BestHand
	return nil
}
