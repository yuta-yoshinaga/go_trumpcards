//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// おいちょかぶフェーズ定数
const (
	OichoKabuPhaseBet  = 1 // ベットフェーズ（掛け金を置く）
	OichoKabuPhaseDraw = 2 // 引くフェーズ（子が3枚目を引くか勝負するか）
	OichoKabuPhaseEnd  = 3 // 終了フェーズ（結果表示）
)

// OichoKabuResult は子（human）から見たラウンド結果。
// GameResult は共有ファイル internal/domain/game_result.go に移動したので到達可能に
// なったが、この型名は JSON ペイロードに出るため統合していない（#4462）。値は
// GameResult と同一。
type OichoKabuResult int

// OichoKabuResult 定数（値は GameResult と同一）
const (
	OichoKabuResultLose OichoKabuResult = -1 // 子の負け（親の勝ち）
	OichoKabuResultDraw OichoKabuResult = 0  // 引き分け（プッシュ、掛け金返却）
	OichoKabuResultWin  OichoKabuResult = 1  // 子の勝ち
)

// おいちょかぶデフォルト値
const (
	OichoKabuDefaultChips = 1000  // デフォルトチップ
	OichoKabuMinBet       = 10    // 最低ベット額
	OichoKabuMaxBet       = 10000 // 最大ベット額
	OichoKabuDeckSize     = 40    // カブ札（1〜10 を4枚ずつ）
	OichoKabuHandMax      = 3     // 1手の最大枚数
	OichoKabuHandInitial  = 2     // 初期配布枚数 (これを超えていれば追い引きした)
	OichoKabuCopies       = 4     // 各数字の枚数
	OichoKabuValueMax     = 10    // カブ札の最大数字

	// OichoKabuBankerDrawThreshold は親（胴元）の追い引きルール閾値。
	// ハウスルール: 親は自分のランクが 6 以下のとき3枚目を引き、7 以上なら止める。
	OichoKabuBankerDrawThreshold = 6
)

// OichoKabu おいちょかぶゲーム本体。
//
// おいちょかぶ（Oicho-Kabu）は 40 枚のカブ札（1〜10 を4枚ずつ）を使う
// 日本のバカラ系バンキングゲーム。1 人の「子」(human) が「親」(banker, 胴元)
// と勝負する。各札の点数は value%10（10 は 0 点）、手の目は点数合計の 1 の位。
// 9（カブ）が最強、0（ブタ）が最弱。高い目が勝ち、同点はプッシュ（掛け金返却）。
//
// MVP のため、しっぴん（四一）やあらし（同数3枚）といった特殊役は実装しない。
// 目は純粋に「点数合計 % 10」のみで判定する。
type OichoKabu struct {
	deck        []*Card
	playerHand  []*Card // 子（human）の手
	bankerHand  []*Card // 親（banker/胴元）の手
	chips       ChipHolder
	bet         int
	phase       int
	gameEndFlag bool
	result      OichoKabuResult
	totalPayout int
	actionLogBase
}

// NewOichoKabu コンストラクタ。40 枚のカブ札を組み立ててシャッフルする。
func NewOichoKabu() *OichoKabu {
	o := &OichoKabu{phase: OichoKabuPhaseBet}
	o.deck = buildOichoKabuDeck()
	o.shuffle()
	return o
}

// NewDefaultOichoKabu デフォルト設定のおいちょかぶを生成するファクトリ関数。
func NewDefaultOichoKabu() *OichoKabu {
	o := NewOichoKabu()
	o.chips.SetChips(OichoKabuDefaultChips)
	return o
}

// buildOichoKabuDeck は 40 枚のカブ札を組み立てる（1〜10 を4枚ずつ）。
// design=1..4 はコピー番号（一意化のため）、value=1..10 が数字。
func buildOichoKabuDeck() []*Card {
	deck := make([]*Card, 0, OichoKabuDeckSize)
	for copyIdx := 1; copyIdx <= OichoKabuCopies; copyIdx++ {
		for v := 1; v <= OichoKabuValueMax; v++ {
			deck = append(deck, NewCard(copyIdx, v, false))
		}
	}
	return deck
}

// shuffle 山札をシャッフルする。
func (o *OichoKabu) shuffle() {
	rand.Shuffle(len(o.deck), func(i, j int) {
		o.deck[i], o.deck[j] = o.deck[j], o.deck[i]
	})
}

// Reset ゲーム初期化。チップは持ち越し、手・掛け金・山札のみ初期化する。
func (o *OichoKabu) Reset() {
	o.gameEndFlag = false
	o.phase = OichoKabuPhaseBet
	o.playerHand = nil
	o.bankerHand = nil
	o.bet = 0
	o.result = 0
	o.totalPayout = 0
	o.actionLog = nil
	if o.chips.GetChips() < OichoKabuMinBet {
		o.chips.SetChips(OichoKabuDefaultChips)
	}
	o.deck = buildOichoKabuDeck()
	o.shuffle()
}

// Bet 掛け金を置き、子・親に 2 枚ずつ配る。親の手は結果まで伏せられる。
func (o *OichoKabu) Bet(amount int) error {
	if o.phase != OichoKabuPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if amount < OichoKabuMinBet || amount%OichoKabuMinBet != 0 || amount > OichoKabuMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	if !o.chips.SubtractChips(amount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	o.bet = amount
	o.appendLog(0, "bet", fmt.Sprintf("bet=%d", amount), nil)

	o.playerHand = make([]*Card, 0, OichoKabuHandMax)
	o.bankerHand = make([]*Card, 0, OichoKabuHandMax)
	for i := 0; i < OichoKabuHandInitial; i++ {
		o.playerHand = append(o.playerHand, o.drawCard())
		o.bankerHand = append(o.bankerHand, o.drawCard())
	}
	o.appendLog(-1, "deal", "dealt 2 cards each", o.playerHand)
	o.phase = OichoKabuPhaseDraw
	return nil
}

// Draw 子が 3 枚目を引く。引いた後、親が追い引きして勝負を決める。
func (o *OichoKabu) Draw() error {
	if o.phase != OichoKabuPhaseDraw {
		return NewDomainError(ErrWrongPhase, "Draw is only allowed during the draw phase.")
	}
	if len(o.playerHand) >= OichoKabuHandMax {
		return NewDomainError(ErrInvalidPlay, "The player already holds the maximum number of cards.")
	}
	c := o.drawCard()
	if c != nil {
		o.playerHand = append(o.playerHand, c)
	}
	o.appendLog(0, "draw", "player draws a third card", []*Card{c})
	o.resolve()
	return nil
}

// Stand 子が引かずに勝負する。親が追い引きして勝負を決める。
func (o *OichoKabu) Stand() error {
	if o.phase != OichoKabuPhaseDraw {
		return NewDomainError(ErrWrongPhase, "Stand is only allowed during the draw phase.")
	}
	o.appendLog(0, "stand", "player stands", nil)
	o.resolve()
	return nil
}

// resolve 親の追い引き（ハウスルール: ランク6以下で1枚引く）を行い、
// 目を比較して配当を確定する。高い目が勝ち、同点はプッシュ（掛け金返却）。
func (o *OichoKabu) resolve() {
	if o.bankerRank() <= OichoKabuBankerDrawThreshold && len(o.bankerHand) < OichoKabuHandMax {
		c := o.drawCard()
		if c != nil {
			o.bankerHand = append(o.bankerHand, c)
		}
		o.appendLog(-1, "draw", "banker draws a third card", []*Card{c})
	}

	pr, br := o.playerRank(), o.bankerRank()
	switch {
	case pr > br:
		o.result = OichoKabuResultWin
		o.totalPayout = o.bet * 2 // 掛け金返却 + 1:1 配当
		o.chips.AddChips(o.totalPayout)
		o.appendLog(-1, "result", "player wins", nil)
	case pr < br:
		o.result = OichoKabuResultLose
		o.totalPayout = 0
		o.appendLog(-1, "result", "banker wins", nil)
	default:
		o.result = OichoKabuResultDraw // プッシュ
		o.totalPayout = o.bet          // 掛け金返却
		o.chips.AddChips(o.totalPayout)
		o.appendLog(-1, "result", "push", nil)
	}
	o.gameEndFlag = true
	o.phase = OichoKabuPhaseEnd
}

// drawCard 山札の末尾から 1 枚引く（払い出しフラグを立てる）。
func (o *OichoKabu) drawCard() *Card {
	if len(o.deck) == 0 {
		return nil
	}
	c := o.deck[len(o.deck)-1]
	o.deck = o.deck[:len(o.deck)-1]
	c.SetDraw(true)
	return c
}

// oichoKabuCardPoint は 1 枚の点数（value % 10）を返す。10 は 0 点。
func oichoKabuCardPoint(c *Card) int {
	if c == nil {
		return 0
	}
	return c.GetValue() % 10
}

// oichoKabuHandRank は手の目（点数合計の 1 の位）を返す。
func oichoKabuHandRank(hand []*Card) int {
	sum := 0
	for _, c := range hand {
		sum += oichoKabuCardPoint(c)
	}
	return sum % 10
}

// playerRank 子の目
func (o *OichoKabu) playerRank() int { return oichoKabuHandRank(o.playerHand) }

// bankerRank 親の目
func (o *OichoKabu) bankerRank() int { return oichoKabuHandRank(o.bankerHand) }

// --- Getters ---

// GetPlayerHand 子の手
func (o *OichoKabu) GetPlayerHand() []*Card { return o.playerHand }

// GetBankerHand 親の手
func (o *OichoKabu) GetBankerHand() []*Card { return o.bankerHand }

// GetPlayerRank 子の目
func (o *OichoKabu) GetPlayerRank() int { return o.playerRank() }

// GetBankerRank 親の目
func (o *OichoKabu) GetBankerRank() int { return o.bankerRank() }

// GetPhase フェーズ
func (o *OichoKabu) GetPhase() int { return o.phase }

// GetGameEndFlag 終了フラグ
func (o *OichoKabu) GetGameEndFlag() bool { return o.gameEndFlag }

// GetBet 掛け金
func (o *OichoKabu) GetBet() int { return o.bet }

// GetResult 結果
func (o *OichoKabu) GetResult() OichoKabuResult { return o.result }

// GetTotalPayout 合計配当
func (o *OichoKabu) GetTotalPayout() int { return o.totalPayout }

// GetChips チップ
func (o *OichoKabu) GetChips() int { return o.chips.GetChips() }

// --- Test helpers ---

// SetPhase テスト用
func (o *OichoKabu) SetPhase(phase int) { o.phase = phase }

// SetPlayerHand テスト用
func (o *OichoKabu) SetPlayerHand(hand []*Card) { o.playerHand = hand }

// SetBankerHand テスト用
func (o *OichoKabu) SetBankerHand(hand []*Card) { o.bankerHand = hand }

// SetDeck テスト用（追い引きを決定的にするため）
func (o *OichoKabu) SetDeck(deck []*Card) { o.deck = deck }

// SetBet テスト用
func (o *OichoKabu) SetBet(bet int) { o.bet = bet }

// SetChips テスト用
func (o *OichoKabu) SetChips(chips int) { o.chips.SetChips(chips) }

// oichoKabuJSON は OichoKabu の JSON ワイヤーフォーマット。
type oichoKabuJSON struct {
	Deck        []*Card           `json:"dk"`
	PlayerHand  []*Card           `json:"ph"`
	BankerHand  []*Card           `json:"bh"`
	Chips       *ChipHolder       `json:"ch"`
	Bet         int               `json:"bt"`
	Phase       int               `json:"ps"`
	GameEndFlag bool              `json:"ge"`
	Result      OichoKabuResult   `json:"gr"`
	TotalPayout int               `json:"tp"`
	ActionLog   []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (o *OichoKabu) MarshalJSON() ([]byte, error) {
	return json.Marshal(oichoKabuJSON{
		Deck:        o.deck,
		PlayerHand:  o.playerHand,
		BankerHand:  o.bankerHand,
		Chips:       &o.chips,
		Bet:         o.bet,
		Phase:       o.phase,
		GameEndFlag: o.gameEndFlag,
		Result:      o.result,
		TotalPayout: o.totalPayout,
		ActionLog:   o.actionLog,
	})
}

// oichoKabuMaxSliceLen はデシリアライズ時のスライス長の上限。
const oichoKabuMaxSliceLen = 1000

// hasNilCard は札スライスに nil 要素が含まれるかを返す。
func hasNilCard(cards []*Card) bool {
	for _, c := range cards {
		if c == nil {
			return true
		}
	}
	return false
}

// UnmarshalJSON implements json.Unmarshaler.
// スライス長と手札枚数、nil 要素を防御的に検証する。
func (o *OichoKabu) UnmarshalJSON(data []byte) error {
	var j oichoKabuJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Deck) > oichoKabuMaxSliceLen || len(j.ActionLog) > oichoKabuMaxSliceLen {
		return fmt.Errorf("oichokabu: input array exceeds maximum allowed size")
	}
	if len(j.PlayerHand) > OichoKabuHandMax || len(j.BankerHand) > OichoKabuHandMax {
		return fmt.Errorf("oichokabu: a hand exceeds the maximum of %d cards", OichoKabuHandMax)
	}
	if hasNilCard(j.Deck) || hasNilCard(j.PlayerHand) || hasNilCard(j.BankerHand) {
		return fmt.Errorf("oichokabu: card slices must not contain nil elements")
	}

	o.deck = j.Deck
	if o.deck == nil {
		o.deck = make([]*Card, 0)
	}
	o.playerHand = j.PlayerHand
	o.bankerHand = j.BankerHand
	if j.Chips != nil {
		o.chips = *j.Chips
	}
	o.bet = j.Bet
	o.phase = j.Phase
	o.gameEndFlag = j.GameEndFlag
	o.result = j.Result
	o.totalPayout = j.TotalPayout
	o.actionLog = j.ActionLog
	if o.actionLog == nil {
		o.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
