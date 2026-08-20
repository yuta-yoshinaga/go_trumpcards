//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// crazyFourPokerMaxSliceLen はデシリアライズ時のスライス長上限。
const crazyFourPokerMaxSliceLen = 1000

// エラー値。
var (
	errCrazyFourPokerWrongPhase   = errors.New("crazyfourpoker: action not allowed in this phase")
	errCrazyFourPokerAnteRange    = errors.New("crazyfourpoker: ante is out of range")
	errCrazyFourPokerAnteUnit     = errors.New("crazyfourpoker: ante must be a multiple of the betting unit")
	errCrazyFourPokerSideRange    = errors.New("crazyfourpoker: side bet is out of range")
	errCrazyFourPokerNotEnough    = errors.New("crazyfourpoker: not enough chips for that wager")
	errCrazyFourPokerMultiplier   = errors.New("crazyfourpoker: that play multiplier needs a pair of aces or better")
	errCrazyFourPokerGameFinished = errors.New("crazyfourpoker: the player is out of chips")
)

// CrazyFourPokerResult は 1 ラウンドの決着。
type CrazyFourPokerResult int

// 決着の種類。
const (
	// CrazyFourPokerResultNone まだ決着していない
	CrazyFourPokerResultNone CrazyFourPokerResult = iota
	// CrazyFourPokerResultFold 降りた
	CrazyFourPokerResultFold
	// CrazyFourPokerResultWin プレイヤーの勝ち
	CrazyFourPokerResultWin
	// CrazyFourPokerResultLose ディーラーの勝ち
	CrazyFourPokerResultLose
	// CrazyFourPokerResultPush 引き分け
	CrazyFourPokerResultPush
	// CrazyFourPokerResultDealerNotQualified ディーラー不成立
	CrazyFourPokerResultDealerNotQualified
)

// CrazyFourPokerResultMax は最大の決着値 (復元時の範囲検査に使う)。
const CrazyFourPokerResultMax = CrazyFourPokerResultDealerNotQualified

// CrazyFourPokerResultName は決着の識別子を返す (i18n キーの一部に使う)。
func CrazyFourPokerResultName(r CrazyFourPokerResult) string {
	switch r {
	case CrazyFourPokerResultFold:
		return "fold"
	case CrazyFourPokerResultWin:
		return "win"
	case CrazyFourPokerResultLose:
		return "lose"
	case CrazyFourPokerResultPush:
		return "push"
	case CrazyFourPokerResultDealerNotQualified:
		return "dealerNotQualified"
	default:
		return "none"
	}
}

// CrazyFourPokerHint は人間への助言。
type CrazyFourPokerHint struct {
	// Multiplier は薦めるプレイベットの倍率。0 ならフォールドを薦める。
	Multiplier int
	// Reason は理由の識別子 (i18n キーの一部)。
	Reason string
}

// CrazyFourPoker はクレイジー 4 ポーカーの卓。
//
// プレイヤーとディーラーに 5 枚ずつ配り、**それぞれ最良の 4 枚**で勝負する。
// 名前の由来は「手が強いと分かった時点でプレイベットを 3 倍まで乗せられる」ところで、
// **倍率を動かせること自体がエースのペア以上の特典**。
//
// 賭けは 3 本立て:
//   - アンティと Super Bonus は**同額で必須**
//   - Queens Up は任意のサイドベット (自分の役だけで決まる)
//   - プレイベットは手札を見てから置く
type CrazyFourPoker struct {
	trumpCards *TrumpCards
	player     *CrazyFourPokerPlayer
	config     CrazyFourPokerConfig

	phase       CrazyFourPokerPhase
	playerHand  []*Card
	dealerHand  []*Card
	playerBest  []*Card
	dealerBest  []*Card
	anteBet     int
	superBet    int
	queensUpBet int
	playBet     int
	playMult    int
	result      CrazyFourPokerResult
	// payout はこのラウンドで戻ってきた総額 (賭け金の返却を含む)。
	payout      int
	roundNumber int
	gameEndFlag bool
	actionLog   []*ActionLogEntry
	turnNumber  int
}

// NewCrazyFourPoker は指定のデッキ・プレイヤー・設定で卓を構築する。
func NewCrazyFourPoker(trumpCards *TrumpCards, player *CrazyFourPokerPlayer, config CrazyFourPokerConfig) *CrazyFourPoker {
	return &CrazyFourPoker{
		trumpCards:  trumpCards,
		player:      player,
		config:      config,
		roundNumber: 1,
	}
}

// NewDefaultCrazyFourPoker は既定の卓を構築する。
func NewDefaultCrazyFourPoker() *CrazyFourPoker {
	cfg := DefaultCrazyFourPokerConfig()
	return NewCrazyFourPoker(NewTrumpCards(0), NewCrazyFourPokerPlayer(cfg.InitialChips), cfg)
}

// Reset はゲームを初期状態に戻す。
func (g *CrazyFourPoker) Reset() {
	g.player.SetChips(g.config.InitialChips)
	g.roundNumber = 1
	g.gameEndFlag = false
	g.actionLog = nil
	g.turnNumber = 0
	g.appendLog("start", "crazy 4 poker begins", nil)
	g.startRound()
}

// startRound は 1 ラウンド分の状態を初期化する。
func (g *CrazyFourPoker) startRound() {
	g.trumpCards = NewTrumpCards(0)
	g.trumpCards.Shuffle()
	g.playerHand = nil
	g.dealerHand = nil
	g.playerBest = nil
	g.dealerBest = nil
	g.anteBet = 0
	g.superBet = 0
	g.queensUpBet = 0
	g.playBet = 0
	g.playMult = 0
	g.result = CrazyFourPokerResultNone
	g.payout = 0
	g.phase = CrazyFourPokerPhaseBet
}

// NextRound は次のラウンドを始める。
func (g *CrazyFourPoker) NextRound() error {
	if g.gameEndFlag {
		return errCrazyFourPokerGameFinished
	}
	if g.phase != CrazyFourPokerPhaseResult {
		return errCrazyFourPokerWrongPhase
	}
	g.roundNumber++
	g.startRound()
	// **最低額のアンティも置けなくなったら終わり。**
	if g.player.GetChips() < g.minTotalWager() {
		g.gameEndFlag = true
		g.appendLog("gameEnd", "out of chips", nil)
	}
	return nil
}

// minTotalWager は 1 ラウンドを始めるのに最低限必要なチップを返す。
//
// **アンティと Super Bonus は同額で必須**なので、最低でもアンティ 2 口が要る。
// さらにプレイベットで最低 1 口を置くので、実際には 3 口ないと最後まで打てない。
func (g *CrazyFourPoker) minTotalWager() int {
	return CrazyFourPokerAnteMin * 3
}

// GetMinTotalWager は 1 ラウンドに最低限必要なチップを返す。
func (g *CrazyFourPoker) GetMinTotalWager() int { return g.minTotalWager() }

// --- 賭け ---

// PlaceBet はアンティ (と同額の Super Bonus)、任意の Queens Up を置いて配る。
//
// **Super Bonus は選べない。** アンティと同額で必ず付く。
func (g *CrazyFourPoker) PlaceBet(ante, queensUp int) error {
	if g.gameEndFlag {
		return errCrazyFourPokerGameFinished
	}
	if g.phase != CrazyFourPokerPhaseBet {
		return errCrazyFourPokerWrongPhase
	}
	if ante < CrazyFourPokerAnteMin || ante > CrazyFourPokerAnteMax {
		return errCrazyFourPokerAnteRange
	}
	// **配当に 1.5:1 があるので刻みを固定する。** 端数を丸めると控除率が動く。
	if ante%CrazyFourPokerAnteUnit != 0 {
		return errCrazyFourPokerAnteUnit
	}
	if queensUp < 0 || queensUp > CrazyFourPokerAnteMax {
		return errCrazyFourPokerSideRange
	}
	if queensUp != 0 && queensUp%CrazyFourPokerAnteUnit != 0 {
		return errCrazyFourPokerAnteUnit
	}

	total := ante*2 + queensUp // アンティ + Super Bonus + Queens Up
	if !g.player.SubtractChips(total) {
		return errCrazyFourPokerNotEnough
	}
	g.anteBet = ante
	g.superBet = ante
	g.queensUpBet = queensUp
	g.appendLog("bet", fmt.Sprintf("ante %d, super %d, queensUp %d", ante, ante, queensUp), nil)

	g.deal()
	return nil
}

// deal は 5 枚ずつ配り、最良の 4 枚を確定させる。
func (g *CrazyFourPoker) deal() {
	for range CrazyFourPokerHandSize {
		g.playerHand = append(g.playerHand, g.trumpCards.DrawCard())
		g.dealerHand = append(g.dealerHand, g.trumpCards.DrawCard())
	}
	g.playerBest = pickBestFour(g.playerHand)
	g.dealerBest = pickBestFour(g.dealerHand)
	g.appendLog("deal", "cards dealt", g.playerHand)
	g.phase = CrazyFourPokerPhaseDecide
}

// --- 判断 ---

// PlayerHasAcesOrBetter はプレイヤーの最良 4 枚が**エースのペア以上**かを返す。
//
// これが 3 倍まで乗せられる条件。
func (g *CrazyFourPoker) PlayerHasAcesOrBetter() bool {
	return crazyFourPokerPairAtLeast(g.playerBest, CrazyFourPokerSuperBonusMinPair)
}

// MaxPlayMultiplier はいま置けるプレイベットの上限倍率を返す。
func (g *CrazyFourPoker) MaxPlayMultiplier() int {
	return CrazyFourPokerMaxPlayMultiplier(g.PlayerHasAcesOrBetter())
}

// Play はプレイベットを置いて決着させる。multiplier はアンティに対する倍率。
func (g *CrazyFourPoker) Play(multiplier int) error {
	if g.gameEndFlag {
		return errCrazyFourPokerGameFinished
	}
	if g.phase != CrazyFourPokerPhaseDecide {
		return errCrazyFourPokerWrongPhase
	}
	if multiplier < CrazyFourPokerPlayMin {
		return errCrazyFourPokerMultiplier
	}
	// **上限はドメインだけが決める。** 手役を見ずに 2 倍を通すと 3 倍ルールが消える。
	if multiplier > g.MaxPlayMultiplier() {
		return errCrazyFourPokerMultiplier
	}
	bet := g.anteBet * multiplier
	if !g.player.SubtractChips(bet) {
		return errCrazyFourPokerNotEnough
	}
	g.playBet = bet
	g.playMult = multiplier
	g.appendLog("play", fmt.Sprintf("play %d (x%d)", bet, multiplier), nil)
	g.resolve()
	return nil
}

// Fold は降りる。**アンティと Super Bonus を失う。**
func (g *CrazyFourPoker) Fold() error {
	if g.gameEndFlag {
		return errCrazyFourPokerGameFinished
	}
	if g.phase != CrazyFourPokerPhaseDecide {
		return errCrazyFourPokerWrongPhase
	}
	g.result = CrazyFourPokerResultFold
	g.appendLog("fold", "folded", nil)
	// 降りても Queens Up は自分の役だけで決まるので生きている。
	g.payout = g.queensUpPayout()
	g.player.AddChips(g.payout)
	g.phase = CrazyFourPokerPhaseResult
	return nil
}

// --- 決着 ---

// PlayerQualifies はプレイヤーの手がキング以上かを返す。
//
// **画面のヒントがこれを自分で計算しないための窓口。** 札の値から「キング以上か」を
// 引き直すと、ディーラーの成立条件と同じ規則がフロントにもう 1 つできてずれる。
func (g *CrazyFourPoker) PlayerQualifies() bool { return crazyFourPokerQualifies(g.playerBest) }

// DealerQualifies はディーラーの手が成立している (キング以上) かを返す。
//
// **成立しなければプレイベットはプッシュ**で、アンティだけが 1:1 で払われる。
func (g *CrazyFourPoker) DealerQualifies() bool {
	return crazyFourPokerQualifies(g.dealerBest)
}

// crazyFourPokerQualifies は 4 枚の手がキング以上で成立しているかを返す。
func crazyFourPokerQualifies(best []*Card) bool {
	if len(best) == 0 {
		return false
	}
	if evalFourCardHand(best) > FourCardHandHighCard {
		return true // ペア以上は無条件で成立
	}
	for _, c := range best {
		if crazyFourPokerRankOrder(c.GetValue()) >= CrazyFourPokerDealerQualifyValue {
			return true
		}
	}
	return false
}

// crazyFourPokerRankOrder は A を最強 (14) とする順位を返す。
func crazyFourPokerRankOrder(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// crazyFourPokerPairAtLeast は「ペア以上の役」かつ「ちょうどペアなら minPair 以上」を返す。
func crazyFourPokerPairAtLeast(best []*Card, minPair int) bool {
	if len(best) == 0 {
		return false
	}
	rank := evalFourCardHand(best)
	if rank > FourCardHandPair {
		return true
	}
	if rank != FourCardHandPair {
		return false
	}
	pv := fourCardPairSortedValues(best)
	if len(pv) == 0 {
		return false
	}
	return crazyFourPokerRankOrder(pv[0]) >= crazyFourPokerRankOrder(minPair)
}

// resolve は勝敗を決め、すべての賭けを精算する。
func (g *CrazyFourPoker) resolve() {
	cmp := compareFourCardHands(g.playerBest, g.dealerBest)
	qualifies := g.DealerQualifies()

	switch {
	case !qualifies:
		g.result = CrazyFourPokerResultDealerNotQualified
	case cmp > 0:
		g.result = CrazyFourPokerResultWin
	case cmp < 0:
		g.result = CrazyFourPokerResultLose
	default:
		g.result = CrazyFourPokerResultPush
	}

	g.payout = g.antePlayPayout(cmp, qualifies) + g.superBonusPayout(cmp, qualifies) + g.queensUpPayout()
	g.player.AddChips(g.payout)
	g.appendLog("result", CrazyFourPokerResultName(g.result), nil)
	g.phase = CrazyFourPokerPhaseResult
}

// antePlayPayout はアンティとプレイベットの払い戻し (賭け金の返却を含む) を返す。
func (g *CrazyFourPoker) antePlayPayout(cmp int, qualifies bool) int {
	switch {
	case !qualifies:
		// **アンティは 1:1、プレイベットはプッシュ。**
		return g.anteBet*2 + g.playBet
	case cmp > 0:
		return (g.anteBet + g.playBet) * 2
	case cmp == 0:
		return g.anteBet + g.playBet
	default:
		return 0
	}
}

// superBonusPayout は Super Bonus の払い戻し (賭け金の返却を含む) を返す。
//
// **エースのペア以上なら勝敗に関係なく配当表どおりに払う。** それ未満のときは
// 勝ち・引き分けで返却、負けで没収。ここがこのゲームでいちばん誤解されやすい賭けで、
// 「強い手を持っていれば負けても報われる」という設計になっている。
func (g *CrazyFourPoker) superBonusPayout(cmp int, qualifies bool) int {
	if g.superBet == 0 {
		return 0
	}
	if mult, ok := g.superBonusMultiplier(); ok {
		return g.superBet + g.superBet*mult/CrazyFourPokerPayoutScale
	}
	// AA 未満。ディーラー不成立は「負けていない」ので返却する。
	if !qualifies || cmp >= 0 {
		return g.superBet
	}
	return 0
}

// superBonusMultiplier は Super Bonus の配当倍率 (10 倍した整数) を返す。
func (g *CrazyFourPoker) superBonusMultiplier() (int, bool) {
	if !crazyFourPokerPairAtLeast(g.playerBest, CrazyFourPokerSuperBonusMinPair) {
		return 0, false
	}
	rank := evalFourCardHand(g.playerBest)
	// **fourCardQuadValue はエースを 14 で返す** (エース高の慣習)。1 と比べると
	// 4 枚のエースが別枠に入らず、静かに 30:1 で払われる。
	if rank == FourCardHandFourOfAKind &&
		fourCardQuadValue(g.playerBest) == crazyFourPokerRankOrder(1) {
		return CrazyFourPokerFourAcesPayout, true // 4 枚のエースは別枠
	}
	mult, ok := crazyFourPokerSuperBonusPayouts[rank]
	return mult, ok
}

// queensUpPayout は Queens Up の払い戻し (賭け金の返却を含む) を返す。
//
// **自分の役だけで決まる。** ディーラーの手も、降りたかどうかも関係しない。
func (g *CrazyFourPoker) queensUpPayout() int {
	if g.queensUpBet == 0 {
		return 0
	}
	if !crazyFourPokerPairAtLeast(g.playerBest, CrazyFourPokerQueensUpMinPair) {
		return 0
	}
	pay, ok := crazyFourPokerQueensUpPayouts[evalFourCardHand(g.playerBest)]
	if !ok {
		return 0
	}
	return g.queensUpBet + g.queensUpBet*pay
}

// --- ヒント ---

// GetHint は人間への助言を返す。判断どころでなければ nil。
//
// **助言はドメインの規則からしか作らない。** 上限倍率も MaxPlayMultiplier に聞く。
func (g *CrazyFourPoker) GetHint() *CrazyFourPokerHint {
	if g.gameEndFlag || g.phase != CrazyFourPokerPhaseDecide {
		return nil
	}
	if g.PlayerHasAcesOrBetter() {
		return &CrazyFourPokerHint{Multiplier: g.MaxPlayMultiplier(), Reason: "acesOrBetter"}
	}
	// エース未満でも、ディーラーが降りない以上フォールドは損になりやすい。
	// クイーン以上のハイカードがあれば乗る。
	if crazyFourPokerQualifies(g.playerBest) {
		return &CrazyFourPokerHint{Multiplier: CrazyFourPokerPlayMin, Reason: "marginal"}
	}
	return &CrazyFourPokerHint{Multiplier: 0, Reason: "fold"}
}

// --- ログ ---

// appendLog は行動ログを 1 行足す。
func (g *CrazyFourPoker) appendLog(actionType, detail string, cards []*Card) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
	if len(g.actionLog) > crazyFourPokerMaxSliceLen {
		g.actionLog = g.actionLog[len(g.actionLog)-crazyFourPokerMaxSliceLen:]
	}
}

// --- アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *CrazyFourPoker) GetPhase() CrazyFourPokerPhase { return g.phase }

// GetPlayerHand は配られた 5 枚を返す。
func (g *CrazyFourPoker) GetPlayerHand() []*Card { return g.playerHand }

// GetDealerHand はディーラーの 5 枚を返す。
func (g *CrazyFourPoker) GetDealerHand() []*Card { return g.dealerHand }

// GetPlayerBest はプレイヤーの最良 4 枚を返す。
func (g *CrazyFourPoker) GetPlayerBest() []*Card { return g.playerBest }

// GetDealerBest はディーラーの最良 4 枚を返す。
func (g *CrazyFourPoker) GetDealerBest() []*Card { return g.dealerBest }

// GetPlayerHandRank はプレイヤーの役を返す。
func (g *CrazyFourPoker) GetPlayerHandRank() int { return crazyFourPokerRankOf(g.playerBest) }

// GetDealerHandRank はディーラーの役を返す。
func (g *CrazyFourPoker) GetDealerHandRank() int { return crazyFourPokerRankOf(g.dealerBest) }

// crazyFourPokerRankOf は 4 枚の役を返す (未配なら 0)。
func crazyFourPokerRankOf(best []*Card) int {
	if len(best) == 0 {
		return 0
	}
	return evalFourCardHand(best)
}

// GetAnteBet はアンティ額を返す。
func (g *CrazyFourPoker) GetAnteBet() int { return g.anteBet }

// GetSuperBet は Super Bonus 額を返す。**常にアンティと同額。**
func (g *CrazyFourPoker) GetSuperBet() int { return g.superBet }

// GetQueensUpBet は Queens Up 額を返す。
func (g *CrazyFourPoker) GetQueensUpBet() int { return g.queensUpBet }

// GetPlayBet はプレイベット額を返す。
func (g *CrazyFourPoker) GetPlayBet() int { return g.playBet }

// GetPlayMultiplier はプレイベットの倍率を返す。
func (g *CrazyFourPoker) GetPlayMultiplier() int { return g.playMult }

// GetResult はラウンドの決着を返す。
func (g *CrazyFourPoker) GetResult() CrazyFourPokerResult { return g.result }

// GetPayout はこのラウンドで戻ってきた総額 (賭け金の返却を含む) を返す。
func (g *CrazyFourPoker) GetPayout() int { return g.payout }

// GetChips は保有チップ数を返す。
func (g *CrazyFourPoker) GetChips() int { return g.player.GetChips() }

// SetChips は保有チップ数を設定する。
func (g *CrazyFourPoker) SetChips(chips int) { g.player.SetChips(chips) }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *CrazyFourPoker) GetRoundNumber() int { return g.roundNumber }

// GetGameEndFlag はゲームが終了したかを返す。
func (g *CrazyFourPoker) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayer はプレイヤーを返す。
func (g *CrazyFourPoker) GetPlayer() *CrazyFourPokerPlayer { return g.player }

// GetConfig は設定を返す。
func (g *CrazyFourPoker) GetConfig() CrazyFourPokerConfig { return g.config }

// SetConfig は設定を差し替える。
func (g *CrazyFourPoker) SetConfig(c CrazyFourPokerConfig) { g.config = c }

// GetActionLog は行動ログを返す。
func (g *CrazyFourPoker) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetRemainingCards はデッキの残り枚数を返す。
func (g *CrazyFourPoker) GetRemainingCards() int { return g.trumpCards.GetRemainingCount() }

// --- 永続化 ---

// crazyFourPokerJSON is the JSON wire format for CrazyFourPoker.
type crazyFourPokerJSON struct {
	TrumpCards  *TrumpCards           `json:"tc"`
	Player      *CrazyFourPokerPlayer `json:"pl"`
	Config      CrazyFourPokerConfig  `json:"cf"`
	Phase       int                   `json:"ph"`
	PlayerHand  []*Card               `json:"ph5"`
	DealerHand  []*Card               `json:"dh5"`
	PlayerBest  []*Card               `json:"pb"`
	DealerBest  []*Card               `json:"db"`
	AnteBet     int                   `json:"an"`
	SuperBet    int                   `json:"sb"`
	QueensUpBet int                   `json:"qu"`
	PlayBet     int                   `json:"pbt"`
	PlayMult    int                   `json:"pm"`
	Result      int                   `json:"rs"`
	Payout      int                   `json:"po"`
	RoundNumber int                   `json:"rn"`
	GameEndFlag bool                  `json:"ge"`
	ActionLog   []*ActionLogEntry     `json:"al"`
	TurnNumber  int                   `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (g *CrazyFourPoker) MarshalJSON() ([]byte, error) {
	return json.Marshal(crazyFourPokerJSON{
		TrumpCards:  g.trumpCards,
		Player:      g.player,
		Config:      g.config,
		Phase:       int(g.phase),
		PlayerHand:  g.playerHand,
		DealerHand:  g.dealerHand,
		PlayerBest:  g.playerBest,
		DealerBest:  g.dealerBest,
		AnteBet:     g.anteBet,
		SuperBet:    g.superBet,
		QueensUpBet: g.queensUpBet,
		PlayBet:     g.playBet,
		PlayMult:    g.playMult,
		Result:      int(g.result),
		Payout:      g.payout,
		RoundNumber: g.roundNumber,
		GameEndFlag: g.gameEndFlag,
		ActionLog:   g.actionLog,
		TurnNumber:  g.turnNumber,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **範囲チェックだけでは足りない。** 個々の値が範囲内でも、組み合わせとして到達
// できない盤面 (Super Bonus がアンティと違う、賭ける前に配札済み、エース未満なのに
// 3 倍が置かれている、決着済みなのに判断フェーズ) はいくらでも作れる。壊れた保存
// データが静かに配当を変えるのを防ぐため、**値どうしの整合**まで見る。
func (g *CrazyFourPoker) UnmarshalJSON(data []byte) error {
	var j crazyFourPokerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Player == nil {
		return fmt.Errorf("crazyfourpoker: the player is missing")
	}
	if err := crazyFourPokerValidateScalars(&j); err != nil {
		return err
	}
	if err := crazyFourPokerValidateConsistency(&j); err != nil {
		return err
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.player = j.Player
	g.config = j.Config
	g.phase = CrazyFourPokerPhase(j.Phase)
	g.playerHand = j.PlayerHand
	g.dealerHand = j.DealerHand
	g.playerBest = j.PlayerBest
	g.dealerBest = j.DealerBest
	g.anteBet = j.AnteBet
	g.superBet = j.SuperBet
	g.queensUpBet = j.QueensUpBet
	g.playBet = j.PlayBet
	g.playMult = j.PlayMult
	g.result = CrazyFourPokerResult(j.Result)
	g.payout = j.Payout
	g.roundNumber = j.RoundNumber
	g.gameEndFlag = j.GameEndFlag
	g.actionLog = j.ActionLog
	g.turnNumber = j.TurnNumber
	return nil
}

// crazyFourPokerValidateScalars は単独で範囲を判定できる値を検証する。
func crazyFourPokerValidateScalars(j *crazyFourPokerJSON) error {
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < int(CrazyFourPokerPhaseBet) || j.Phase > int(CrazyFourPokerPhaseMax) {
		return fmt.Errorf("crazyfourpoker: phase out of range: %d", j.Phase)
	}
	if j.Result < int(CrazyFourPokerResultNone) || j.Result > int(CrazyFourPokerResultMax) {
		return fmt.Errorf("crazyfourpoker: result out of range: %d", j.Result)
	}
	// **順序を固定して見る。** map を range すると走査順が実行ごとに変わるので、
	// 2 つ以上の額が同時に負の保存では**返るエラーが毎回違う**。利用者から見れば
	// 「同じ壊れた保存を読み込むたびに別の理由を言われる」ことになり、テストから
	// 見れば数回に 1 回だけ落ちるフレークになる (CI で実際に落ちた)。
	for _, f := range []struct {
		name  string
		value int
	}{
		{"ante", j.AnteBet},
		{"super bonus", j.SuperBet},
		{"queens up", j.QueensUpBet},
		{"play", j.PlayBet},
		{"payout", j.Payout},
	} {
		if f.value < 0 {
			return fmt.Errorf("crazyfourpoker: %s must not be negative: %d", f.name, f.value)
		}
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("crazyfourpoker: round number out of range: %d", j.RoundNumber)
	}
	if len(j.PlayerHand) > CrazyFourPokerHandSize || len(j.DealerHand) > CrazyFourPokerHandSize {
		return fmt.Errorf("crazyfourpoker: a hand holds more than %d cards", CrazyFourPokerHandSize)
	}
	if len(j.PlayerBest) > CrazyFourPokerBestSize || len(j.DealerBest) > CrazyFourPokerBestSize {
		return fmt.Errorf("crazyfourpoker: a best hand holds more than %d cards", CrazyFourPokerBestSize)
	}
	if len(j.ActionLog) > crazyFourPokerMaxSliceLen {
		return fmt.Errorf("crazyfourpoker: action log too long: %d", len(j.ActionLog))
	}
	return nil
}

// crazyFourPokerValidateConsistency は値どうしの整合を検証する。
func crazyFourPokerValidateConsistency(j *crazyFourPokerJSON) error {
	// **Super Bonus は常にアンティと同額。** ここを素通しすると、賭けていない額に
	// 配当が付く保存データが作れる。
	if j.SuperBet != j.AnteBet {
		return fmt.Errorf("crazyfourpoker: the super bonus (%d) does not match the ante (%d)",
			j.SuperBet, j.AnteBet)
	}
	if j.Phase == int(CrazyFourPokerPhaseBet) &&
		(len(j.PlayerHand) > 0 || len(j.DealerHand) > 0) {
		return fmt.Errorf("crazyfourpoker: cards are dealt before the ante is placed")
	}
	if j.AnteBet == 0 && j.PlayBet > 0 {
		return fmt.Errorf("crazyfourpoker: a play bet without an ante")
	}
	if j.PlayMult > 0 && j.PlayBet != j.AnteBet*j.PlayMult {
		return fmt.Errorf("crazyfourpoker: play bet %d is not %d x the ante %d",
			j.PlayBet, j.PlayMult, j.AnteBet)
	}
	if j.PlayMult > CrazyFourPokerPlayAcesMax {
		return fmt.Errorf("crazyfourpoker: play multiplier out of range: %d", j.PlayMult)
	}
	// **3 倍はエースのペア以上でしか置けない。** 倍率だけ書き換えた保存データを弾く。
	if j.PlayMult > CrazyFourPokerPlayNormalMax &&
		!crazyFourPokerPairAtLeast(j.PlayerBest, CrazyFourPokerSuperBonusMinPair) {
		return fmt.Errorf("crazyfourpoker: multiplier %d needs a pair of aces or better", j.PlayMult)
	}
	if j.Result != int(CrazyFourPokerResultNone) && j.Phase != int(CrazyFourPokerPhaseResult) {
		return fmt.Errorf("crazyfourpoker: the round is decided but the phase is %d", j.Phase)
	}
	return nil
}
