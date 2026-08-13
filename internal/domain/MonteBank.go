//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// MonteBankPhase はゲームの進行段階。
type MonteBankPhase int

const (
	// MonteBankPhaseBet は場札を見て賭ける段階。
	MonteBankPhaseBet MonteBankPhase = iota
	// MonteBankPhaseResult はゲートをめくって決着した段階。
	MonteBankPhaseResult
	// MonteBankPhaseGameEnd は山かチップが尽きた段階。
	MonteBankPhaseGameEnd
)

// MonteBankPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const MonteBankPhaseMax = MonteBankPhaseGameEnd

// monteBankMaxSliceLen は復元時に許すスライス長の上限。
const monteBankMaxSliceLen = 512

// エラー値。
var (
	errMonteBankFinished   = errors.New("montebank: game already finished")
	errMonteBankWrongPhase = errors.New("montebank: not allowed in this phase")
	errMonteBankBetRange   = errors.New("montebank: bet out of range")
	errMonteBankBetUnit    = errors.New("montebank: bet must be a multiple of the unit")
	errMonteBankNotEnough  = errors.New("montebank: not enough chips")
	errMonteBankPickRange  = errors.New("montebank: layout index out of range")
)

// MonteBankResult は 1 ラウンドの決着。
type MonteBankResult int

const (
	// MonteBankResultNone はまだ決着していない。
	MonteBankResultNone MonteBankResult = iota
	// MonteBankResultWin はゲートが賭けた札のスートと一致した。
	MonteBankResultWin
	// MonteBankResultLose は一致しなかった。
	MonteBankResultLose
)

// MonteBankResultMax は最大の決着値 (復元時の範囲検査に使う)。
const MonteBankResultMax = MonteBankResultLose

// MonteBankResultName は決着の識別子を返す (i18n キーの一部に使う)。
func MonteBankResultName(r MonteBankResult) string {
	switch r {
	case MonteBankResultWin:
		return "win"
	case MonteBankResultLose:
		return "lose"
	default:
		return "none"
	}
}

// MonteBank はモンテバンク (スパニッシュ・モンテ) の卓。
//
// **スートだけを見るゲームで、ランクは一切使わない。** 山の上下から 2 枚ずつ
// めくって 4 枚の場札を並べ、プレイヤーはそのうち 1 枚に賭ける。次にめくる
// 1 枚 (ゲート) が賭けた札とスートで一致すれば 3:1。
//
// **控除率はすべてプレイヤーの選択から出る。** 場札に 1 枚しか出ていない
// スートに賭ければ 9/36 = 25% でちょうど互角、2 枚出ているスートなら 8/36 で
// 11.1% の損。だから「どれに賭けるか」だけが遊びになっていて、機械的に
// 賭け続けると必ず削られる。詳細は `MonteBankPayout` の説明にある。
type MonteBank struct {
	deck   *TrumpCards
	player *MonteBankPlayer
	config MonteBankConfig

	phase MonteBankPhase
	// layout は場札 4 枚。前半 2 枚が「上の組」、後半 2 枚が「下の組」の呼び名。
	layout []*Card
	gate   *Card
	// pick は賭けた場札の位置。賭ける前は -1。
	pick    int
	bet     int
	result  MonteBankResult
	payout  int
	roundNo int

	gameEndFlag bool
	actionLog   []*ActionLogEntry
	turnNumber  int
}

// NewMonteBank は指定の山・プレイヤー・設定で卓を構築する。
func NewMonteBank(deck *TrumpCards, player *MonteBankPlayer, config MonteBankConfig) *MonteBank {
	return &MonteBank{deck: deck, player: player, config: config, roundNo: 1, pick: -1}
}

// NewDefaultMonteBank は既定の卓を構築する。
//
// **ラテン 40 枚デッキは既存の無タグのコンストラクタで作れる。** issue は
// `Tute.go` からの流用を指しているが、Tute 自身が `NewTrumpCardsBriscola` を
// 呼んでいるだけで、そちらは `TrumpCards.go` にあってビルドタグを持たない。
// 経由する必要が無いぶん、このゲームのバケットは自由に選べる。
func NewDefaultMonteBank() *MonteBank {
	cfg := DefaultMonteBankConfig()
	return NewMonteBank(NewTrumpCardsBriscola(), NewMonteBankPlayer(cfg.InitialChips), cfg)
}

// Reset はゲームを初期化する。
func (g *MonteBank) Reset() {
	g.deck.Replenish()
	g.deck.Shuffle()
	g.player.SetChips(g.config.InitialChips)
	g.phase = MonteBankPhaseBet
	g.gate = nil
	g.pick = -1
	g.bet = 0
	g.result = MonteBankResultNone
	g.payout = 0
	g.roundNo = 1
	g.gameEndFlag = false
	g.actionLog = nil
	g.turnNumber = 0
	g.dealLayout()
	g.appendLog("reset", "game reset", nil)
}

// dealLayout は場札 4 枚を並べる。前半 2 枚が「上の組」、後半 2 枚が「下の組」。
//
// **上下という区別は表示のためだけに持っている。** 本来のモンテバンクは山の
// 両端からめくるが、**切ってある山では上から取るのも下から取るのも同じ確率
// 分布**なので、勝敗には一切影響しない。山の両端を実際に使うには共有の
// `TrumpCards` に取り出し口をもう 1 つ足すことになり、その JSON 形式は 307
// ゲームすべてが共有している ── 1 ゲームの演出のために動かす場所ではない。
// 呼び名だけを残して、確率の同じ操作で組み立てる。
func (g *MonteBank) dealLayout() {
	g.layout = make([]*Card, 0, MonteBankLayoutSize)
	for range MonteBankLayoutSize {
		if c := g.deck.DrawCard(); c != nil {
			g.layout = append(g.layout, c)
		}
	}
}

// --- 進行 ---

// PlaceBet は場札 idx に賭け、ゲートをめくって決着させる。
func (g *MonteBank) PlaceBet(idx, bet int) error {
	if g.gameEndFlag {
		return errMonteBankFinished
	}
	if g.phase != MonteBankPhaseBet {
		return errMonteBankWrongPhase
	}
	if idx < 0 || idx >= len(g.layout) {
		return errMonteBankPickRange
	}
	if bet < MonteBankMinBet || bet > MonteBankMaxBet {
		return errMonteBankBetRange
	}
	// **3:1 が整数で割り切れるように刻みを固定する。**
	if bet%MonteBankBetUnit != 0 {
		return errMonteBankBetUnit
	}
	if !g.player.SubtractChips(bet) {
		return errMonteBankNotEnough
	}

	g.pick = idx
	g.bet = bet
	g.appendLog("bet", fmt.Sprintf("bet %d on layout %d", bet, idx), []*Card{g.layout[idx]})
	g.revealGate()
	return nil
}

// revealGate はゲートをめくって精算する。
func (g *MonteBank) revealGate() {
	g.gate = g.deck.DrawCard()
	if g.gate != nil && g.gate.GetDesign() == g.layout[g.pick].GetDesign() {
		g.result = MonteBankResultWin
		// 賭け金の返却 + 3 倍の配当。
		g.payout = g.bet + g.bet*MonteBankPayout
	} else {
		g.result = MonteBankResultLose
		g.payout = 0
	}
	g.player.AddChips(g.payout)
	g.phase = MonteBankPhaseResult
	g.appendLog("gate", fmt.Sprintf("gate revealed, payout %d", g.payout), monteBankCardOrNil(g.gate))
}

// monteBankCardOrNil は棋譜に載せる札の並びを作る。
func monteBankCardOrNil(c *Card) []*Card {
	if c == nil {
		return nil
	}
	return []*Card{c}
}

// NextRound は次のラウンドを始める。
//
// **山が 1 ラウンドぶんに足りなければ終わる。** 途中で足りなくなると場札が
// 4 枚に満たないまま賭けさせることになり、選択肢の数が黙って変わる。
func (g *MonteBank) NextRound() error {
	if g.gameEndFlag {
		return errMonteBankFinished
	}
	if g.phase != MonteBankPhaseResult {
		return errMonteBankWrongPhase
	}
	if g.deck.GetRemainingCount() < MonteBankCardsPerRound || g.player.GetChips() < MonteBankMinBet {
		g.finish()
		return nil
	}
	g.roundNo++
	g.phase = MonteBankPhaseBet
	g.gate = nil
	g.pick = -1
	g.bet = 0
	g.result = MonteBankResultNone
	g.payout = 0
	g.dealLayout()
	return nil
}

// finish はゲームを終える。
func (g *MonteBank) finish() {
	g.gameEndFlag = true
	g.phase = MonteBankPhaseGameEnd
	g.appendLog("gameEnd", fmt.Sprintf("finished with %d chips", g.player.GetChips()), nil)
}

// --- 参照 ---

// SuitCountInLayout は場札に指定スートが何枚出ているかを返す。
//
// **これが賭けの良し悪しを決める唯一の数字。** 1 枚なら互角、2 枚以上なら
// 賭けるだけ損になる。画面もヒントもここを見る。
func (g *MonteBank) SuitCountInLayout(design int) int {
	n := 0
	for _, c := range g.layout {
		if c != nil && c.GetDesign() == design {
			n++
		}
	}
	return n
}

// RemainingOfSuit は場札を除いた残りに指定スートが何枚あるかを返す。
func (g *MonteBank) RemainingOfSuit(design int) int {
	return MonteBankSuitSize - g.SuitCountInLayout(design)
}

// GetConfig はゲーム設定を返す。
func (g *MonteBank) GetConfig() MonteBankConfig { return g.config }

// SetConfig はゲーム設定を設定する。
func (g *MonteBank) SetConfig(c MonteBankConfig) { g.config = c }

// GetPhase は現在のフェーズを返す。
func (g *MonteBank) GetPhase() MonteBankPhase { return g.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *MonteBank) GetGameEndFlag() bool { return g.gameEndFlag }

// GetLayout は場札を返す。
func (g *MonteBank) GetLayout() []*Card { return g.layout }

// GetGate はゲートの札を返す。まだめくっていなければ nil。
func (g *MonteBank) GetGate() *Card { return g.gate }

// GetPick は賭けた場札の位置を返す。賭ける前は -1。
func (g *MonteBank) GetPick() int { return g.pick }

// GetBet は賭け金を返す。
func (g *MonteBank) GetBet() int { return g.bet }

// GetResult はラウンドの決着を返す。
func (g *MonteBank) GetResult() MonteBankResult { return g.result }

// GetPayout はこのラウンドで戻ってきた総額を返す。
func (g *MonteBank) GetPayout() int { return g.payout }

// GetChips は保有チップ数を返す。
func (g *MonteBank) GetChips() int { return g.player.GetChips() }

// SetChips は保有チップ数を設定する。
func (g *MonteBank) SetChips(chips int) { g.player.SetChips(chips) }

// GetPlayer はプレイヤーを返す。
func (g *MonteBank) GetPlayer() *MonteBankPlayer { return g.player }

// GetRoundNumber は現在のラウンド数を返す。
func (g *MonteBank) GetRoundNumber() int { return g.roundNo }

// GetRemainingCards は山の残り枚数を返す。
func (g *MonteBank) GetRemainingCards() int { return g.deck.GetRemainingCount() }

// GetActionLog は棋譜を返す。
func (g *MonteBank) GetActionLog() []*ActionLogEntry { return g.actionLog }

// appendLog は棋譜に 1 行足す。
func (g *MonteBank) appendLog(actionType, detail string, cards []*Card) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
	if len(g.actionLog) > monteBankMaxSliceLen {
		g.actionLog = g.actionLog[len(g.actionLog)-monteBankMaxSliceLen:]
	}
}

// --- 助言 ---

// MonteBankHint は人間への助言。
type MonteBankHint struct {
	// PickIdx は薦める場札の位置。
	PickIdx int
	// Reason は理由の識別子 (i18n キーの一部)。
	Reason string
}

// GetHint は賭けどころの助言を返す。賭ける場面でなければ nil。
//
// **場札に 1 枚しか出ていないスートを選ぶ。** それがちょうど互角で、重複した
// スートを選ぶと 11% 以上の損になる ── このゲームで唯一の判断がここ。
func (g *MonteBank) GetHint() *MonteBankHint {
	if g.gameEndFlag || g.phase != MonteBankPhaseBet || len(g.layout) == 0 {
		return nil
	}
	best, bestCount := 0, MonteBankLayoutSize+1
	for i, c := range g.layout {
		if c == nil {
			continue
		}
		if n := g.SuitCountInLayout(c.GetDesign()); n < bestCount {
			best, bestCount = i, n
		}
	}
	if bestCount == 1 {
		return &MonteBankHint{PickIdx: best, Reason: "loneSuit"}
	}
	// **どれを選んでも重複している。** 少ないほうを選ぶしかないが、互角には
	// ならないことを言う ── 「得な手がある」と誤解させないため。
	return &MonteBankHint{PickIdx: best, Reason: "allDuplicated"}
}

// --- 永続化 ---

// monteBankJSON is the JSON wire format for MonteBank.
type monteBankJSON struct {
	Deck        *TrumpCards       `json:"dk"`
	Player      *MonteBankPlayer  `json:"pl"`
	Config      MonteBankConfig   `json:"cf"`
	Phase       int               `json:"ph"`
	Layout      []*Card           `json:"ly"`
	Gate        *Card             `json:"gt"`
	Pick        int               `json:"pk"`
	Bet         int               `json:"bt"`
	Result      int               `json:"rs"`
	Payout      int               `json:"po"`
	RoundNumber int               `json:"rn"`
	GameEndFlag bool              `json:"ge"`
	ActionLog   []*ActionLogEntry `json:"al"`
	TurnNumber  int               `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (g *MonteBank) MarshalJSON() ([]byte, error) {
	return json.Marshal(monteBankJSON{
		Deck: g.deck, Player: g.player, Config: g.config,
		Phase: int(g.phase), Layout: g.layout, Gate: g.gate,
		Pick: g.pick, Bet: g.bet, Result: int(g.result), Payout: g.payout,
		RoundNumber: g.roundNo, GameEndFlag: g.gameEndFlag,
		ActionLog: g.actionLog, TurnNumber: g.turnNumber,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **範囲だけでなく、フェーズと盤面の整合まで見る。** 賭ける前なのにゲートが
// 開いている / 決着後なのに賭けた札が無い、はどちらも添字としては正当で、
// 通すと配当だけが静かに変わる。
func (g *MonteBank) UnmarshalJSON(data []byte) error {
	var j monteBankJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player == nil {
		return fmt.Errorf("montebank: the player is missing")
	}
	if err := monteBankValidate(&j); err != nil {
		return err
	}

	g.deck = j.Deck
	if g.deck == nil {
		g.deck = NewTrumpCardsBriscola()
	}
	g.player = j.Player
	g.config = j.Config
	g.phase = MonteBankPhase(j.Phase)
	g.layout = j.Layout
	g.gate = j.Gate
	g.pick = j.Pick
	g.bet = j.Bet
	g.result = MonteBankResult(j.Result)
	g.payout = j.Payout
	g.roundNo = j.RoundNumber
	g.gameEndFlag = j.GameEndFlag
	g.actionLog = j.ActionLog
	g.turnNumber = j.TurnNumber
	return nil
}

// monteBankValidate は保存データの範囲と整合を検証する。
func monteBankValidate(j *monteBankJSON) error {
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < int(MonteBankPhaseBet) || j.Phase > int(MonteBankPhaseMax) {
		return fmt.Errorf("montebank: phase out of range: %d", j.Phase)
	}
	if j.Result < int(MonteBankResultNone) || j.Result > int(MonteBankResultMax) {
		return fmt.Errorf("montebank: result out of range: %d", j.Result)
	}
	if len(j.Layout) > MonteBankLayoutSize {
		return fmt.Errorf("montebank: %d layout cards exceeds %d", len(j.Layout), MonteBankLayoutSize)
	}
	if j.Bet < 0 {
		return fmt.Errorf("montebank: bet must not be negative: %d", j.Bet)
	}
	if j.Payout < 0 {
		return fmt.Errorf("montebank: payout must not be negative: %d", j.Payout)
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("montebank: round number out of range: %d", j.RoundNumber)
	}
	if len(j.ActionLog) > monteBankMaxSliceLen {
		return fmt.Errorf("montebank: action log too long: %d", len(j.ActionLog))
	}
	return monteBankValidatePhase(j)
}

// monteBankValidatePhase はフェーズと盤面の整合を見る。
func monteBankValidatePhase(j *monteBankJSON) error {
	// **賭けた札の位置は場札の中か、まだ賭けていないことを表す -1 だけ。**
	if j.Pick < -1 || j.Pick >= len(j.Layout) {
		return fmt.Errorf("montebank: pick out of range: %d", j.Pick)
	}
	switch MonteBankPhase(j.Phase) {
	case MonteBankPhaseBet:
		if j.Gate != nil {
			return fmt.Errorf("montebank: the gate is open before the bet")
		}
		if j.Pick >= 0 || j.Bet != 0 {
			return fmt.Errorf("montebank: a bet is recorded before the betting phase ends")
		}
		if len(j.Layout) != MonteBankLayoutSize {
			return fmt.Errorf("montebank: %d layout cards in the betting phase", len(j.Layout))
		}
	case MonteBankPhaseResult:
		if j.Pick < 0 {
			return fmt.Errorf("montebank: the round is settled with no bet placed")
		}
		if j.Bet == 0 {
			return fmt.Errorf("montebank: the round is settled with a zero bet")
		}
		if j.Result == int(MonteBankResultNone) {
			return fmt.Errorf("montebank: the round is settled with no result")
		}
	default:
	}
	// **勝ちなら必ず払い戻しがあり、負けなら無い。** 片方だけ書き換えた保存を弾く。
	if j.Result == int(MonteBankResultWin) && j.Payout == 0 {
		return fmt.Errorf("montebank: a winning round paid nothing")
	}
	if j.Result == int(MonteBankResultLose) && j.Payout != 0 {
		return fmt.Errorf("montebank: a losing round paid %d", j.Payout)
	}
	return nil
}
