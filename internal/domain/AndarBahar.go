//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// アンダーバハールのフェーズ定数
const (
	// AndarBaharPhaseBet ジョーカーが公開され、ベットを待っている状態
	AndarBaharPhaseBet = 1
	// AndarBaharPhaseEnd 配布が終わり、精算が済んだ状態
	AndarBaharPhaseEnd = 2
)

// アンダーバハールのベット先 (列) 定数
const (
	// AndarBaharBetAndar アンダー (内側) の列
	AndarBaharBetAndar = 0
	// AndarBaharBetBahar バハール (外側) の列
	AndarBaharBetBahar = 1
)

// アンダーバハールのサイドベット (決着までの配布枚数) 定数
const (
	// AndarBaharSideNone サイドベットなし
	AndarBaharSideNone = -1
	// AndarBaharSideFirst 1 枚目で決着
	AndarBaharSideFirst = 0
	// AndarBaharSide2To5 2〜5 枚で決着
	AndarBaharSide2To5 = 1
	// AndarBaharSide6To10 6〜10 枚で決着
	AndarBaharSide6To10 = 2
	// AndarBaharSide11To15 11〜15 枚で決着
	AndarBaharSide11To15 = 3
	// AndarBaharSide16To25 16〜25 枚で決着
	AndarBaharSide16To25 = 4
	// AndarBaharSide26To35 26〜35 枚で決着
	AndarBaharSide26To35 = 5
	// AndarBaharSide36Plus 36 枚以上で決着
	AndarBaharSide36Plus = 6
)

// アンダーバハールの既定値
const (
	// AndarBaharDefaultChips 開始チップ
	AndarBaharDefaultChips = 1000
	// AndarBaharMinBet 最低ベット額 (兼ベット額の刻み)
	AndarBaharMinBet = 10
	// AndarBaharMaxBet 最大ベット額
	AndarBaharMaxBet = 10000
)

// AndarBaharPayoutScale は配当倍率の分母。
//
// **チップは int なので、0.9:1 を float で持つと必ず丸めが出ます。** 倍率を 1/10 単位の
// 整数で持ち、ベット額を 10 の倍数に限ることで、払い戻しは常に整数になります
// (`amount * 19 / 10` は amount が 10 の倍数なら誤差なし)。
const AndarBaharPayoutScale = 10

// AndarBaharFirstColumnPayout は**先に配る列**に賭けて当たったときの払戻総額の倍率
// (1/10 単位、賭け金の返還を含む)。0.9:1 なので 1 + 0.9 = 1.9 倍。
const AndarBaharFirstColumnPayout = 19

// AndarBaharSecondColumnPayout は**後に配る列**に賭けて当たったときの払戻総額の倍率
// (1/10 単位)。1:1 なので 2.0 倍。
const AndarBaharSecondColumnPayout = 20

// andarBaharSidePayouts はサイドベットの払戻総額の倍率 (1/10 単位、賭け金の返還を含む)。
//
// **公平倍率から約 11% のハウスマージンを引いた値です。** 添字は AndarBaharSide* 定数。
var andarBaharSidePayouts = [...]int{150, 42, 41, 52, 41, 90, 330}

// andarBaharSideBands はサイドベットの帯 (決着枚数の下限・上限、両端を含む)。
var andarBaharSideBands = [...][2]int{{1, 1}, {2, 5}, {6, 10}, {11, 15}, {16, 25}, {26, 35}, {36, 51}}

// AndarBaharMaxCards は 1 ラウンドで配りうる最大枚数。
//
// **必ず終わります。** ジョーカーを除いた 51 枚に同ランクが 3 枚あるので、いちばん早い
// 1 枚は遅くとも 49 枚目には出ます。
const AndarBaharMaxCards = 49

// andarBaharMaxSliceLen は復元時に受け付けるスライス長の上限。
const andarBaharMaxSliceLen = 1000

// AndarBahar はアンダーバハール (Andar Bahar / Katti) のゲーム本体。
//
// インド・バンガロール発祥のカジノゲーム。**ジョーカー (基準札) を 1 枚めくり、
// アンダーとバハールの 2 列へ交互に配って、基準札と同じランクが先に出た列を当てます。**
//
// # 非対称なのは「先に配る列」であって「1 枚目」ではない
//
// 基準札の色で配り始める列が決まります (黒ならアンダー、赤ならバハール)。**先に配る列は
// 1 枚多く配られる機会があるぶん有利**で、残り 51 枚に同ランクが 3 枚あるとき
// **51.50% 対 48.50%** です。だから先の列だけ 0.9:1 に下げます——業界標準の配当表が
// 非対称なのはこれが理由で、「最初の 1 枚だから」ではありません。
//
// 両方を 1:1 にすると先の列がプレイヤー有利 (+3.00%) になり、ハウスが成立しません。
//
// # 停止保証
//
// 51 枚に同ランクが 3 枚残っているので、配布は必ず 49 枚以内で止まります。
type AndarBahar struct {
	trumpCards *TrumpCards
	// joker は基準札。
	joker *Card
	// andar / bahar は各列に配られた札 (配った順)。
	andar []*Card
	bahar []*Card
	// firstColumn は基準札の色で決まる、先に配る列。
	firstColumn int
	chips       ChipHolder

	betAmount int
	betTarget int
	// sideAmount / sideBand はサイドベット (賭けていなければ 0 / AndarBaharSideNone)。
	sideAmount int
	sideBand   int

	phase       int
	gameEndFlag bool
	// winner は基準札と同ランクが出た列。
	winner int
	result GameResult
	payout int
	// mainPayout / sidePayout は payout の内訳。
	//
	// **サイドベットは別の賭け。** 合計だけでは、外したのがメインなのかサイドなのか
	// 画面から読めません (#5770)。合計は常に両者の和です。
	mainPayout int
	sidePayout int
	actionLogBase
	// history は罫線用の勝ち列の履歴。
	history []int
}

// NewAndarBahar はコンストラクタ。
func NewAndarBahar(trumpCards *TrumpCards) *AndarBahar {
	trumpCards.Shuffle()
	ab := &AndarBahar{
		trumpCards: trumpCards,
		phase:      AndarBaharPhaseBet,
		sideBand:   AndarBaharSideNone,
		winner:     -1,
	}
	ab.revealJoker()
	return ab
}

// NewDefaultAndarBahar は既定設定のアンダーバハールを返す。
func NewDefaultAndarBahar() *AndarBahar {
	ab := NewAndarBahar(NewTrumpCards(0))
	ab.chips.SetChips(AndarBaharDefaultChips)
	return ab
}

// Reset はラウンドを初期化する (罫線は保持)。
func (ab *AndarBahar) Reset() {
	ab.gameEndFlag = false
	ab.phase = AndarBaharPhaseBet
	ab.andar = nil
	ab.bahar = nil
	ab.betAmount = 0
	ab.betTarget = AndarBaharBetAndar
	ab.sideAmount = 0
	ab.sideBand = AndarBaharSideNone
	ab.winner = -1
	ab.result = 0
	ab.payout = 0
	ab.mainPayout = 0
	ab.sidePayout = 0
	ab.actionLog = nil
	if ab.chips.GetChips() < AndarBaharMinBet {
		ab.chips.SetChips(AndarBaharDefaultChips)
	}
	ab.trumpCards = NewTrumpCards(0)
	ab.trumpCards.Shuffle()
	ab.revealJoker()
}

// revealJoker は基準札をめくり、先に配る列を決める。
//
// **黒 (スペード/クローバー) ならアンダーから、赤 (ハート/ダイヤ) ならバハールから。**
func (ab *AndarBahar) revealJoker() {
	ab.joker = ab.trumpCards.DrawCard()
	ab.firstColumn = AndarBaharFirstColumnFor(ab.joker)
	ab.appendLog(-1, "joker",
		fmt.Sprintf("基準札 %s が公開されました", andarBaharColumnName(ab.firstColumn)),
		[]*Card{ab.joker})
}

// AndarBaharFirstColumnFor は基準札の色から先に配る列を返す。
func AndarBaharFirstColumnFor(joker *Card) int {
	if joker == nil {
		return AndarBaharBetAndar
	}
	switch joker.GetDesign() {
	case CardDesignHeart, CardDesignDiamond:
		return AndarBaharBetBahar
	default:
		return AndarBaharBetAndar
	}
}

// ClearHistory は罫線履歴を消す。
func (ab *AndarBahar) ClearHistory() { ab.history = nil }

// Bet はベットしてラウンドを最後まで進める。
//
// sideBand に AndarBaharSideNone を渡すとサイドベットなし (sideAmount は無視)。
func (ab *AndarBahar) Bet(amount, target, sideAmount, sideBand int) error {
	if ab.phase != AndarBaharPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if target != AndarBaharBetAndar && target != AndarBaharBetBahar {
		return NewDomainError(ErrInvalidPlay, "Invalid bet target.")
	}
	if err := andarBaharValidateAmount(amount); err != nil {
		return err
	}
	if sideBand == AndarBaharSideNone {
		sideAmount = 0
	} else {
		if sideBand < AndarBaharSideFirst || sideBand > AndarBaharSide36Plus {
			return NewDomainError(ErrInvalidPlay, "Invalid side bet band.")
		}
		if err := andarBaharValidateAmount(sideAmount); err != nil {
			return err
		}
	}
	if !ab.chips.SubtractChips(amount + sideAmount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}

	ab.betAmount = amount
	ab.betTarget = target
	ab.sideAmount = sideAmount
	ab.sideBand = sideBand
	ab.appendLog(0, "bet",
		fmt.Sprintf("%s に %d ベットしました", andarBaharColumnName(target), amount), nil)

	ab.deal()
	ab.judge()
	return nil
}

// andarBaharValidateAmount はベット額の刻みと範囲を検証する。
//
// **10 の倍数に限るのは配当の丸めを避けるため** (0.9:1 が整数で払えます)。
func andarBaharValidateAmount(amount int) error {
	if amount < AndarBaharMinBet || amount > AndarBaharMaxBet || amount%AndarBaharMinBet != 0 {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	return nil
}

// deal は基準札と同ランクが出るまで交互に配る。
func (ab *AndarBahar) deal() {
	target := andarBaharRank(ab.joker)
	col := ab.firstColumn
	for range AndarBaharMaxCards {
		c := ab.trumpCards.DrawCard()
		if c == nil {
			break
		}
		ab.push(col, c)
		if andarBaharRank(c) == target {
			ab.winner = col
			ab.appendLog(-1, "match",
				fmt.Sprintf("%s に基準札と同じランクが出ました (%d 枚目)",
					andarBaharColumnName(col), ab.DealtCount()), []*Card{c})
			return
		}
		col = andarBaharOtherColumn(col)
	}
}

// push は列に 1 枚積む。
func (ab *AndarBahar) push(col int, c *Card) {
	if col == AndarBaharBetAndar {
		ab.andar = append(ab.andar, c)
		return
	}
	ab.bahar = append(ab.bahar, c)
}

// judge は的中を判定して精算する。
func (ab *AndarBahar) judge() {
	if ab.winner == AndarBaharBetAndar || ab.winner == AndarBaharBetBahar {
		ab.history = append(ab.history, ab.winner)
	}
	if ab.winner == ab.betTarget {
		ab.result = GameResultWin
	} else {
		ab.result = GameResultLose
	}
	ab.mainPayout, ab.sidePayout = ab.calculatePayout()
	ab.payout = ab.mainPayout + ab.sidePayout
	ab.chips.AddChips(ab.payout)
	ab.appendLog(-1, "result",
		fmt.Sprintf("%s の勝ち。払い戻し %d", andarBaharColumnName(ab.winner), ab.payout), nil)

	ab.gameEndFlag = true
	ab.phase = AndarBaharPhaseEnd
}

// calculatePayout はメイン・サイドそれぞれの払戻 (賭け金の返還を含む) を返す。
//
// **先に配る列は 0.9:1、後の列は 1:1。** 先の列が 51.50% で勝つので、同じ配当にすると
// プレイヤー有利になってしまいます。
//
// **サイドベットは独立した賭け。** 内訳で返すのは、メインを外してサイドで取り返した
// 回とその逆を、画面が区別できるようにするためです (#5770)。
func (ab *AndarBahar) calculatePayout() (main, side int) {
	if ab.winner == ab.betTarget {
		rate := AndarBaharSecondColumnPayout
		if ab.betTarget == ab.firstColumn {
			rate = AndarBaharFirstColumnPayout
		}
		main = ab.betAmount * rate / AndarBaharPayoutScale
	}
	if ab.sideBand != AndarBaharSideNone && ab.sideBand == ab.SideBandOf(ab.DealtCount()) {
		side = ab.sideAmount * andarBaharSidePayouts[ab.sideBand] / AndarBaharPayoutScale
	}
	return main, side
}

// SideBandOf は決着枚数が属するサイドベットの帯を返す (該当なしは AndarBaharSideNone)。
func (ab *AndarBahar) SideBandOf(cards int) int {
	for i, b := range andarBaharSideBands {
		if cards >= b[0] && cards <= b[1] {
			return i
		}
	}
	return AndarBaharSideNone
}

// andarBaharRank はランクを 1..13 で返す。
func andarBaharRank(c *Card) int {
	if c == nil {
		return 0
	}
	return c.GetValue()
}

// andarBaharOtherColumn はもう一方の列を返す。
func andarBaharOtherColumn(col int) int {
	if col == AndarBaharBetAndar {
		return AndarBaharBetBahar
	}
	return AndarBaharBetAndar
}

// andarBaharColumnName は列の名前を返す。
func andarBaharColumnName(col int) string {
	switch col {
	case AndarBaharBetAndar:
		return "andar"
	case AndarBaharBetBahar:
		return "bahar"
	default:
		return "unknown"
	}
}

// --- Getters ---

// GetJoker は基準札を返す。
func (ab *AndarBahar) GetJoker() *Card { return ab.joker }

// GetAndarCards はアンダーに配られた札を返す。
func (ab *AndarBahar) GetAndarCards() []*Card { return ab.andar }

// GetBaharCards はバハールに配られた札を返す。
func (ab *AndarBahar) GetBaharCards() []*Card { return ab.bahar }

// GetFirstColumn は先に配る列を返す。
func (ab *AndarBahar) GetFirstColumn() int { return ab.firstColumn }

// DealtCount は決着までに配った枚数を返す。
func (ab *AndarBahar) DealtCount() int { return len(ab.andar) + len(ab.bahar) }

// GetPhase は現在のフェーズを返す。
func (ab *AndarBahar) GetPhase() int { return ab.phase }

// GetGameEndFlag は終了フラグを返す。
func (ab *AndarBahar) GetGameEndFlag() bool { return ab.gameEndFlag }

// GetBetAmount はメインベット額を返す。
func (ab *AndarBahar) GetBetAmount() int { return ab.betAmount }

// GetBetTarget はメインベット先の列を返す。
func (ab *AndarBahar) GetBetTarget() int { return ab.betTarget }

// GetSideAmount はサイドベット額を返す。
func (ab *AndarBahar) GetSideAmount() int { return ab.sideAmount }

// GetSideBand はサイドベットの帯を返す (AndarBaharSideNone = 賭けていない)。
func (ab *AndarBahar) GetSideBand() int { return ab.sideBand }

// GetWinner は基準札と同ランクが出た列を返す (-1 = 未決着)。
func (ab *AndarBahar) GetWinner() int { return ab.winner }

// GetResult はゲーム結果を返す。
func (ab *AndarBahar) GetResult() GameResult { return ab.result }

// GetPayout は払戻総額を返す。
func (ab *AndarBahar) GetPayout() int { return ab.payout }

// GetMainPayout はメインベットぶんの払戻を返す。外れたら 0。
func (ab *AndarBahar) GetMainPayout() int { return ab.mainPayout }

// GetSidePayout はサイドベットぶんの払戻を返す。張っていない・外れたら 0。
func (ab *AndarBahar) GetSidePayout() int { return ab.sidePayout }

// GetChips はチップを返す。
func (ab *AndarBahar) GetChips() int { return ab.chips.GetChips() }

// GetHistory は罫線履歴を返す。
func (ab *AndarBahar) GetHistory() []int { return ab.history }

// AndarBaharSideBand は帯 band の下限・上限を返す。
func AndarBaharSideBand(band int) (int, int, bool) {
	if band < AndarBaharSideFirst || band > AndarBaharSide36Plus {
		return 0, 0, false
	}
	b := andarBaharSideBands[band]
	return b[0], b[1], true
}

// AndarBaharSidePayout は帯 band の払戻倍率 (1/10 単位) を返す。
func AndarBaharSidePayout(band int) (int, bool) {
	if band < AndarBaharSideFirst || band > AndarBaharSide36Plus {
		return 0, false
	}
	return andarBaharSidePayouts[band], true
}

// GetHint は人間への助言を返す。
//
// **先に配る列のほうが 51.50% で有利**ですが、配当が 0.9:1 に下げられているぶん
// 期待値は -2.15% で、後の列 (-3.00%) よりまだ損が小さい、という助言をします。
func (ab *AndarBahar) GetHint() string {
	if ab.phase != AndarBaharPhaseBet {
		return "andarBaharHintWaitNextRound"
	}
	if ab.firstColumn == AndarBaharBetAndar {
		return "andarBaharHintAndarFirst"
	}
	return "andarBaharHintBaharFirst"
}

// --- Test helpers ---

// SetPhase はフェーズを設定する (テスト用)。
func (ab *AndarBahar) SetPhase(phase int) { ab.phase = phase }

// SetChips はチップを設定する (テスト用)。
func (ab *AndarBahar) SetChips(chips int) { ab.chips.SetChips(chips) }

// SetHistory は罫線履歴を設定する (テスト用)。
func (ab *AndarBahar) SetHistory(history []int) { ab.history = history }

// andarBaharJSON は AndarBahar の JSON ワイヤーフォーマット。
type andarBaharJSON struct {
	TrumpCards  *TrumpCards       `json:"tc"`
	Joker       *Card             `json:"jk"`
	Andar       []*Card           `json:"an"`
	Bahar       []*Card           `json:"bh"`
	FirstColumn int               `json:"fc"`
	Chips       *ChipHolder       `json:"ch"`
	BetAmount   int               `json:"ba"`
	BetTarget   int               `json:"bt"`
	SideAmount  int               `json:"sa"`
	SideBand    int               `json:"sb"`
	Phase       int               `json:"ps"`
	GameEndFlag bool              `json:"ge"`
	Winner      int               `json:"wn"`
	Result      GameResult        `json:"rs"`
	Payout      int               `json:"po"`
	MainPayout  int               `json:"pm"`
	SidePayout  int               `json:"pd"`
	ActionLog   []*ActionLogEntry `json:"al"`
	History     []int             `json:"hi"`
}

// MarshalJSON implements json.Marshaler.
func (ab *AndarBahar) MarshalJSON() ([]byte, error) {
	return json.Marshal(andarBaharJSON{
		TrumpCards:  ab.trumpCards,
		Joker:       ab.joker,
		Andar:       ab.andar,
		Bahar:       ab.bahar,
		FirstColumn: ab.firstColumn,
		Chips:       &ab.chips,
		BetAmount:   ab.betAmount,
		BetTarget:   ab.betTarget,
		SideAmount:  ab.sideAmount,
		SideBand:    ab.sideBand,
		Phase:       ab.phase,
		GameEndFlag: ab.gameEndFlag,
		Winner:      ab.winner,
		Result:      ab.result,
		Payout:      ab.payout,
		MainPayout:  ab.mainPayout,
		SidePayout:  ab.sidePayout,
		ActionLog:   ab.actionLog,
		History:     ab.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **範囲チェックだけでは足りません。** 「配った枚数が交互配布と食い違う」「基準札と
// 同ランクが途中に混ざっている」は、どの値も単独では範囲内なのに盤面としては
// あり得ない状態です。ここを通すと勝敗だけが静かに変わります。
func (ab *AndarBahar) UnmarshalJSON(data []byte) error {
	var j andarBaharJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := andarBaharValidateWire(&j); err != nil {
		return err
	}

	ab.trumpCards = j.TrumpCards
	if ab.trumpCards == nil {
		ab.trumpCards = NewTrumpCards(0)
	}
	ab.joker = j.Joker
	ab.andar = j.Andar
	ab.bahar = j.Bahar
	ab.firstColumn = j.FirstColumn
	if j.Chips != nil {
		ab.chips = *j.Chips
	}
	ab.betAmount = j.BetAmount
	ab.betTarget = j.BetTarget
	ab.sideAmount = j.SideAmount
	ab.sideBand = j.SideBand
	ab.phase = j.Phase
	ab.gameEndFlag = j.GameEndFlag
	ab.winner = j.Winner
	ab.result = j.Result
	ab.payout = j.Payout
	ab.mainPayout = j.MainPayout
	ab.sidePayout = j.SidePayout
	ab.actionLog = j.ActionLog
	if ab.actionLog == nil {
		ab.actionLog = make([]*ActionLogEntry, 0)
	}
	ab.history = j.History
	if ab.history == nil {
		ab.history = make([]int, 0)
	}
	return nil
}

// andarBaharValidateWire は復元しようとしている盤面の不変条件を検証する。
func andarBaharValidateWire(j *andarBaharJSON) error {
	if len(j.ActionLog) > andarBaharMaxSliceLen || len(j.History) > andarBaharMaxSliceLen {
		return errors.New("andarbahar: input array exceeds maximum allowed size")
	}
	if j.Phase != AndarBaharPhaseBet && j.Phase != AndarBaharPhaseEnd {
		return fmt.Errorf("andarbahar: phase out of range: %d", j.Phase)
	}
	if j.GameEndFlag != (j.Phase == AndarBaharPhaseEnd) {
		return fmt.Errorf("andarbahar: the game-end flag and the phase disagree (flag=%v, phase=%d)",
			j.GameEndFlag, j.Phase)
	}
	if j.FirstColumn != AndarBaharBetAndar && j.FirstColumn != AndarBaharBetBahar {
		return fmt.Errorf("andarbahar: first column out of range: %d", j.FirstColumn)
	}
	if j.BetTarget != AndarBaharBetAndar && j.BetTarget != AndarBaharBetBahar {
		return fmt.Errorf("andarbahar: bet target out of range: %d", j.BetTarget)
	}
	if err := andarBaharValidateBets(j); err != nil {
		return err
	}
	if j.Joker == nil {
		return errors.New("andarbahar: the joker must be face up")
	}
	// **先に配る列は基準札の色で決まります。** 保存データが両方を持っている以上、
	// 食い違いは改竄か壊れです。
	if want := AndarBaharFirstColumnFor(j.Joker); j.FirstColumn != want {
		return fmt.Errorf("andarbahar: first column %d does not match the joker's colour (want %d)",
			j.FirstColumn, want)
	}
	return andarBaharValidateColumns(j)
}

// andarBaharValidateBets はベット額とサイドベットの整合を検証する。
func andarBaharValidateBets(j *andarBaharJSON) error {
	if j.BetAmount < 0 || j.BetAmount > AndarBaharMaxBet || j.BetAmount%AndarBaharMinBet != 0 {
		return fmt.Errorf("andarbahar: bet amount out of range: %d", j.BetAmount)
	}
	if j.SideAmount < 0 || j.SideAmount > AndarBaharMaxBet || j.SideAmount%AndarBaharMinBet != 0 {
		return fmt.Errorf("andarbahar: side bet amount out of range: %d", j.SideAmount)
	}
	if j.SideBand != AndarBaharSideNone &&
		(j.SideBand < AndarBaharSideFirst || j.SideBand > AndarBaharSide36Plus) {
		return fmt.Errorf("andarbahar: side bet band out of range: %d", j.SideBand)
	}
	// **帯と金額は両方あるか両方無いかのどちらかです。** 片方だけの状態は `Bet` からは
	// 作れないので、保存データにあれば改竄か壊れです。0 は「1 枚目の帯」という有効な値
	// なので、**番号側を見ても「賭けていない」は判定できません**——金額側で決めます。
	if j.SideBand != AndarBaharSideNone && j.SideAmount == 0 {
		return fmt.Errorf("andarbahar: side bet band %d carries no stake", j.SideBand)
	}
	if j.SideBand == AndarBaharSideNone && j.SideAmount != 0 {
		return fmt.Errorf("andarbahar: %d staked on no side bet band", j.SideAmount)
	}
	if j.Payout < 0 {
		return fmt.Errorf("andarbahar: payout cannot be negative: %d", j.Payout)
	}
	if j.Phase == AndarBaharPhaseBet && j.Payout != 0 {
		return fmt.Errorf("andarbahar: %d paid out before the round was dealt", j.Payout)
	}
	// **内訳は合計と食い違えない。** 片方だけ書き換わった保存を通すと、画面が
	// 「メインは当たったのに合計は減っている」と表示してしまう。
	if j.MainPayout < 0 || j.SidePayout < 0 {
		return fmt.Errorf("andarbahar: payout breakdown cannot be negative: main=%d side=%d",
			j.MainPayout, j.SidePayout)
	}
	if j.MainPayout+j.SidePayout != j.Payout {
		return fmt.Errorf("andarbahar: payout breakdown %d+%d does not add up to %d",
			j.MainPayout, j.SidePayout, j.Payout)
	}
	if j.SidePayout != 0 && j.SideBand == AndarBaharSideNone {
		return fmt.Errorf("andarbahar: %d paid on a side bet that was never placed", j.SidePayout)
	}
	return nil
}

// andarBaharValidateColumns は 2 列の中身が交互配布の結果として辻褄が合うかを検証する。
func andarBaharValidateColumns(j *andarBaharJSON) error {
	total := len(j.Andar) + len(j.Bahar)
	// **ベット前はまだ 1 枚も配られていません。** 交互配布の検査より先に見ます——
	// でないと「1 枚だけ置かれた」改竄が枚数の食い違いとして報告され、原因を取り違えます。
	if j.Phase == AndarBaharPhaseBet && total != 0 {
		return fmt.Errorf("andarbahar: %d cards are already dealt in the bet phase", total)
	}
	if total > AndarBaharMaxCards {
		return fmt.Errorf("andarbahar: %d cards dealt, at most %d are possible",
			total, AndarBaharMaxCards)
	}

	first, second := j.Andar, j.Bahar
	if j.FirstColumn == AndarBaharBetBahar {
		first, second = j.Bahar, j.Andar
	}
	// **交互に配るので、先の列は後の列と同数か 1 枚多いかのどちらかです。**
	if d := len(first) - len(second); d != 0 && d != 1 {
		return fmt.Errorf("andarbahar: the columns do not alternate (first=%d, second=%d)",
			len(first), len(second))
	}

	if j.Phase != AndarBaharPhaseEnd {
		return nil
	}
	if total == 0 {
		return errors.New("andarbahar: the round ended without dealing a card")
	}
	if j.Winner != AndarBaharBetAndar && j.Winner != AndarBaharBetBahar {
		return fmt.Errorf("andarbahar: winner out of range: %d", j.Winner)
	}
	return andarBaharValidateMatch(j)
}

// andarBaharValidateMatch は「基準札と同ランクは決着の 1 枚だけ」を検証する。
func andarBaharValidateMatch(j *andarBaharJSON) error {
	rank := andarBaharRank(j.Joker)
	winning := j.Andar
	if j.Winner == AndarBaharBetBahar {
		winning = j.Bahar
	}
	// 決着した列の末尾が同ランクでなければ、そこで止まった説明がつきません。
	if len(winning) == 0 || andarBaharRank(winning[len(winning)-1]) != rank {
		return fmt.Errorf("andarbahar: column %d did not end on a card matching the joker", j.Winner)
	}
	// **途中に同ランクが混ざっていたら、そこで止まっていたはずです。** 末尾が同ランクなのは
	// 上で確かめたので、全体でちょうど 1 枚なら「決着の 1 枚だけ」と言えます。
	matches := 0
	for _, col := range [][]*Card{j.Andar, j.Bahar} {
		for _, c := range col {
			if andarBaharRank(c) == rank {
				matches++
			}
		}
	}
	if matches != 1 {
		return fmt.Errorf("andarbahar: %d cards match the joker, the round stops on the first", matches)
	}
	return nil
}
