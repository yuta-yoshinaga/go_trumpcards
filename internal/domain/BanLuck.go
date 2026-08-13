//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BanLuckPhase はゲームの進行段階。
type BanLuckPhase int

const (
	// BanLuckPhaseBet は子が賭け金を置く段階。
	BanLuckPhaseBet BanLuckPhase = iota
	// BanLuckPhasePlay は各席が引くか止めるかを決める段階。
	BanLuckPhasePlay
	// BanLuckPhaseRoundEnd はラウンドの決着を見せる段階。
	BanLuckPhaseRoundEnd
	// BanLuckPhaseGameEnd は全ラウンドが終わった段階。
	BanLuckPhaseGameEnd
)

// BanLuckPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const BanLuckPhaseMax = BanLuckPhaseGameEnd

// banLuckMaxSliceLen は復元時に許すスライス長の上限。
const banLuckMaxSliceLen = 512

// banLuckMaxCpuSteps は CPU を進める 1 回あたりの上限。
//
// **止まらない規則を書いていないことを機械的に保証する。** 席ごとに最大 5 枚
// なので 6 席でも 30 手あれば足りるが、規則を書き換えたときに無限ループで
// 固まるのではなく上限で止まってほしいので余裕を持たせてある。
const banLuckMaxCpuSteps = 128

// エラー値。
var (
	errBanLuckFinished   = errors.New("banluck: game already finished")
	errBanLuckWrongPhase = errors.New("banluck: not allowed in this phase")
	errBanLuckBetRangeIn = errors.New("banluck: bet out of range")
	errBanLuckBetUnitIn  = errors.New("banluck: bet must be a multiple of the unit")
	errBanLuckNotEnough  = errors.New("banluck: not enough chips")
	errBanLuckNotYourRun = errors.New("banluck: not your turn")
	errBanLuckMustHit    = errors.New("banluck: the banker must hit below the minimum")
	errBanLuckHandFull   = errors.New("banluck: the hand already holds five cards")
)

// BanLuckSeatResult は 1 席のラウンド結果。
type BanLuckSeatResult struct {
	// Rank はその席の役。
	Rank BanLuckRank
	// Outcome は親から見た子の決着 (親自身は Push)。
	Outcome BanLuckOutcome
	// Bet はそのラウンドに置いた額。
	//
	// **精算のあとに席の `bet` を見てはいけない。** 使い終わった賭け金は
	// その場で 0 に戻す (親が回ってきた席に前ラウンドの賭け金が残ると、
	// 「親は賭けない」という不変条件が破れる) ので、表示に要る額はここに残す。
	Bet int
	// Delta はそのラウンドのチップ増減。
	Delta int
}

// BanLuck はバンラック (チャイニーズ・ブラックジャック) の卓。
//
// **親だけが 15 未満で引く義務を負う。** 子は自由に止められるのに親は止まれない
// ── これが「親は全員を相手に一度に勝てる」有利さを打ち消している仕掛けで、
// 片方だけ実装すると期待値が大きく傾く。
//
// **親は席を移る。** 誰が親かは卓が 1 つの添字で持つ。席の側に「親フラグ」を
// 置くと、移すたびに全席を書き換えることになり、しかも「親が 2 人いる」状態を
// 作れてしまう。
type BanLuck struct {
	deck    *TrumpCards
	players []*BanLuckPlayer
	config  BanLuckConfig

	phase    BanLuckPhase
	hands    []*BlackJackHand
	results  []BanLuckSeatResult
	settled  bool // このラウンドの精算が済んだか
	banker   int  // 親の席
	turn     int  // いま操作している席
	roundNum int

	gameEndFlag bool
	actionLog   []*ActionLogEntry
	turnNumber  int
}

// NewBanLuck は指定の山・席・設定で卓を構築する。
func NewBanLuck(deck *TrumpCards, players []*BanLuckPlayer, config BanLuckConfig) *BanLuck {
	return &BanLuck{deck: deck, players: players, config: config, roundNum: 1}
}

// NewDefaultBanLuck は既定の卓を構築する。席 0 が人間。
func NewDefaultBanLuck() *BanLuck {
	cfg := DefaultBanLuckConfig()
	players := make([]*BanLuckPlayer, 0, cfg.Seats)
	for i := range cfg.Seats {
		name := fmt.Sprintf("CPU%d", i)
		if i == 0 {
			name = "YOU"
		}
		players = append(players, NewBanLuckPlayer(name, cfg.InitialChips, i == 0))
	}
	return NewBanLuck(NewTrumpCards(0), players, cfg)
}

// Reset はゲームを初期化する。
func (g *BanLuck) Reset() {
	g.deck.Replenish()
	g.deck.Shuffle()
	for _, p := range g.players {
		p.SetChips(g.config.InitialChips)
		p.SetBet(0)
	}
	g.phase = BanLuckPhaseBet
	g.hands = nil
	g.results = nil
	g.settled = false
	g.banker = 0
	g.turn = 0
	g.roundNum = 1
	g.gameEndFlag = false
	g.actionLog = nil
	g.turnNumber = 0
	g.appendLog(-1, "reset", "game reset", nil)
}

// --- 進行 ---

// PlaceBet は人間の席の賭け金を置き、配って勝負を始める。
//
// **親は賭けない。** 親は全員の賭けを一手に受ける側なので、人間が親のラウンドは
// 額を取らずにそのまま配る。
func (g *BanLuck) PlaceBet(bet int) error {
	if g.gameEndFlag {
		return errBanLuckFinished
	}
	if g.phase != BanLuckPhaseBet {
		return errBanLuckWrongPhase
	}
	if !g.humanIsBanker() {
		if bet < BanLuckMinBet || bet > BanLuckMaxBet {
			return errBanLuckBetRangeIn
		}
		if bet%BanLuckBetUnit != 0 {
			return errBanLuckBetUnitIn
		}
		if g.players[g.humanSeat()].GetChips() < bet {
			return errBanLuckNotEnough
		}
	}
	g.takeBets(bet)
	g.deal()
	g.advanceCpu()
	return nil
}

// takeBets は子の賭け金を決める。CPU は既定額で、足りなければ持ち分すべて。
func (g *BanLuck) takeBets(humanBet int) {
	human := g.humanSeat()
	for i, p := range g.players {
		if i == g.banker {
			p.SetBet(0)
			continue
		}
		want := g.config.DefaultBet
		if i == human {
			want = humanBet
		}
		// **足りない席は持ち分すべてを賭ける。** 席を落とさずに続けるため。
		p.SetBet(min(want, p.GetChips()))
	}
}

// deal は全席に 2 枚ずつ配る。
func (g *BanLuck) deal() {
	g.ensureCards(2 * len(g.players))
	g.hands = make([]*BlackJackHand, len(g.players))
	g.results = make([]BanLuckSeatResult, len(g.players))
	for i := range g.players {
		h := NewBlackJackHand()
		h.AddCard(g.deck.DrawCard())
		h.AddCard(g.deck.DrawCard())
		g.hands[i] = h
		g.results[i] = BanLuckSeatResult{Rank: EvalBanLuckHand(h)}
	}
	g.settled = false
	g.phase = BanLuckPhasePlay
	g.appendLog(-1, "deal", fmt.Sprintf("round %d, banker seat %d", g.roundNum, g.banker), nil)
	// **手番は直接置かない。** 特別役の席は手番を持たないので、配った時点で
	// 「動ける席が 1 つも無い」ことがある (全席が Ban Luck など)。手番を素直に
	// 代入すると精算がどこからも呼ばれず、**盤面が PLAY のまま固まる** ──
	// 人間が親のラウンドでは CPU も走らないので、押せるものが何も無くなる。
	// 進めるか精算するかの判断は `advanceTurn` 1 か所に集める。
	g.advanceTurn()
}

// ensureCards は必要枚数を引けるように山を補充する。
//
// **1 デッキ 52 枚では 6 席 × 5 枚に足りない場面がある。** 引けない札を
// nil のまま手札に入れると、点数計算がそこで静かに壊れる。
func (g *BanLuck) ensureCards(need int) {
	if g.deck.GetRemainingCount() < need {
		g.deck.Replenish()
		g.deck.Shuffle()
	}
}

// seatDone はその席の手がもう動かないかを返す。
func (g *BanLuck) seatDone(i int) bool {
	h := g.hands[i]
	if h == nil {
		return true
	}
	// **特別役は配られた時点で確定。** 引く意味が無いので手番を作らない。
	if r := EvalBanLuckHand(h); r == BanLuckRankBanBan || r == BanLuckRankBanLuck {
		return true
	}
	return h.IsStood() || h.IsBusted() || h.GetCardsSize() >= BanLuckMaxHandCards
}

// Hit はいまの席が 1 枚引く。
func (g *BanLuck) Hit() error {
	if err := g.requireHumanTurn(); err != nil {
		return err
	}
	return g.hitSeat(g.turn)
}

// hitSeat は席 i に 1 枚配り、必要なら手番を進める。
func (g *BanLuck) hitSeat(i int) error {
	h := g.hands[i]
	if h.GetCardsSize() >= BanLuckMaxHandCards {
		return errBanLuckHandFull
	}
	g.ensureCards(1)
	c := g.deck.DrawCard()
	h.AddCard(c)
	if h.GetScore() > BanLuckTarget {
		h.SetBusted(true)
	}
	g.results[i].Rank = EvalBanLuckHand(h)
	g.appendLog(i, "hit", fmt.Sprintf("seat %d hits", i), []*Card{c})
	g.advanceTurn()
	return nil
}

// Stand はいまの席が打ち止めにする。
//
// **親は 15 未満で止まれない。** ここを弾かずに CPU 側だけで守ると、人間が
// 親のときだけ規則が消える。
func (g *BanLuck) Stand() error {
	if err := g.requireHumanTurn(); err != nil {
		return err
	}
	if g.turn == g.banker && g.bankerMustHit() {
		return errBanLuckMustHit
	}
	return g.standSeat(g.turn)
}

// standSeat は席 i を打ち止めにし、手番を進める。
func (g *BanLuck) standSeat(i int) error {
	g.hands[i].SetStood(true)
	g.appendLog(i, "stand", fmt.Sprintf("seat %d stands", i), nil)
	g.advanceTurn()
	return nil
}

// bankerMustHit は親が引く義務を負っているかを返す。
func (g *BanLuck) bankerMustHit() bool {
	h := g.hands[g.banker]
	if h == nil || h.GetCardsSize() >= BanLuckMaxHandCards {
		return false
	}
	return h.GetScore() < BanLuckBankerMustHitUnder
}

// MustHit は人間の席がいま引く義務を負っているかを返す (画面の案内に使う)。
func (g *BanLuck) MustHit() bool {
	if g.phase != BanLuckPhasePlay || g.turn != g.humanSeat() || g.turn != g.banker {
		return false
	}
	return g.bankerMustHit()
}

// requireHumanTurn は人間が操作できる状態かを検査する。
func (g *BanLuck) requireHumanTurn() error {
	if g.gameEndFlag {
		return errBanLuckFinished
	}
	if g.phase != BanLuckPhasePlay {
		return errBanLuckWrongPhase
	}
	if g.turn != g.humanSeat() {
		return errBanLuckNotYourRun
	}
	return nil
}

// advanceTurn は次に動く席へ手番を移し、全席が済んでいれば精算する。
//
// **配った直後の最初の手番もここが決める。** 「動ける席を探す」と「誰も動け
// なければ精算する」は必ず一組で、片方だけを別の場所に書くと固まる経路ができる。
//
// **親は最後。** 子が全員止まってから親が引くので、親は場の状況を見て決められる
// ── 義務ヒットと引き換えの、親の唯一の情報上の得。
func (g *BanLuck) advanceTurn() {
	for i := range g.players {
		if i == g.banker || g.seatDone(i) {
			continue
		}
		g.turn = i
		return
	}
	if !g.seatDone(g.banker) {
		g.turn = g.banker
		return
	}
	g.settle()
}

// CpuPlay は CPU の席を進める。
func (g *BanLuck) CpuPlay() { g.advanceCpu() }

// advanceCpu は人間の手番になるかラウンドが終わるまで CPU を進める。
func (g *BanLuck) advanceCpu() {
	for range banLuckMaxCpuSteps {
		if g.gameEndFlag || g.phase != BanLuckPhasePlay || g.IsHumanTurn() {
			return
		}
		if !g.stepCpu() {
			return
		}
	}
}

// stepCpu は CPU の席を 1 手だけ進める。進めたら true。
func (g *BanLuck) stepCpu() bool {
	i := g.turn
	if i == g.humanSeat() {
		return false
	}
	if g.cpuWantsHit(i) {
		return g.hitSeat(i) == nil
	}
	return g.standSeat(i) == nil
}

// cpuWantsHit は CPU が引くかを決める。
//
// **親の義務は戦略より先。** 15 未満なら好むと好まざるとにかかわらず引く。
func (g *BanLuck) cpuWantsHit(i int) bool {
	h := g.hands[i]
	if h.GetCardsSize() >= BanLuckMaxHandCards {
		return false
	}
	if i == g.banker {
		if g.bankerMustHit() {
			return true
		}
		return h.GetScore() < banLuckCpuBankerStand
	}
	// **4 枚目まで来たら Five Dragon を狙う価値がある。** 21 以下で 5 枚に
	// 届けば、合計に関係なく普通の手に勝てるため。
	if h.GetCardsSize() == BanLuckMaxHandCards-1 && h.GetScore() <= banLuckCpuDragonChase {
		return true
	}
	return h.GetScore() < banLuckCpuSeatStand
}

// CPU の打ち方の境目。
const (
	// banLuckCpuSeatStand は子の CPU が止まる下限。
	banLuckCpuSeatStand = 17
	// banLuckCpuBankerStand は親の CPU が止まる下限 (義務より上)。
	banLuckCpuBankerStand = 17
	// banLuckCpuDragonChase は 4 枚目から Five Dragon を狙う上限。
	banLuckCpuDragonChase = 15
)

// --- 精算 ---

// settle はラウンドを精算する。
//
// **チップは親と子の間だけで動く。** 子の負け分は親に入り、子の勝ち分は親から
// 出る。どこにも消えず、どこからも湧かない。
//
// **負けた席から先に集め、それから勝った席に払う。** 順序を逆にすると、
// 親の持ち分を各席が独立に見るので、3 席が同時に勝ったときに親が持っていない
// 額まで払い出して**残高が負になる**。集めてから払えば、払える上限は常に
// その時点の親の持ち分ひとつで決まる。
func (g *BanLuck) settle() {
	if g.settled {
		return
	}
	bankerHand := g.hands[g.banker]
	bankerRank := EvalBanLuckHand(bankerHand)
	bankerScore := bankerHand.GetScore()
	g.results[g.banker] = BanLuckSeatResult{Rank: bankerRank, Outcome: BanLuckOutcomePush}

	type pending struct {
		seat int
		mult int
	}
	var winners []pending
	bankerDelta := 0

	for i, p := range g.players {
		if i == g.banker {
			continue
		}
		rank := EvalBanLuckHand(g.hands[i])
		outcome, mult := CompareBanLuck(rank, bankerRank, g.hands[i].GetScore(), bankerScore)
		g.results[i] = BanLuckSeatResult{Rank: rank, Outcome: outcome, Bet: p.GetBet()}
		switch outcome {
		case BanLuckOutcomeLose:
			// **払えるぶんまでしか取らない。** 倍率が 3 倍まであるので、
			// 素直に bet*mult を引くと席の残高が負になる。
			amount := min(p.GetBet()*mult, p.GetChips())
			p.SubtractChips(amount)
			g.results[i].Delta = -amount
			bankerDelta += amount
		case BanLuckOutcomeWin:
			winners = append(winners, pending{seat: i, mult: mult})
		default:
		}
	}

	// 集め終わった時点の持ち分が、払える総額。
	avail := g.players[g.banker].GetChips() + bankerDelta
	for _, w := range winners {
		p := g.players[w.seat]
		amount := min(p.GetBet()*w.mult, avail)
		avail -= amount
		p.AddChips(amount)
		g.results[w.seat].Delta = amount
		bankerDelta -= amount
	}

	g.players[g.banker].AddChips(bankerDelta)
	g.results[g.banker].Delta = bankerDelta

	// **使い終わった賭け金は捨てる。** 親はこの直後に次の席へ移るので、
	// 残しておくと「親に賭け金が乗っている」状態が保存に出る (往復テストが
	// 実際にここで落ちた)。表示に要る額は結果側が持っている。
	for _, p := range g.players {
		p.SetBet(0)
	}

	g.settled = true
	g.phase = BanLuckPhaseRoundEnd
	g.appendLog(-1, "result", fmt.Sprintf("banker seat %d nets %d", g.banker, bankerDelta), nil)
	g.rotateBanker()
}

// rotateBanker は次のラウンドの親を決める。
//
// **親を倒した席が次の親になる。** 特別役で親を破った席があればその席、
// 無ければ次の席へ順に回す。「誰も親になれない」経路を作らないため、
// 最後は必ず次の席という既定に落ちる。
func (g *BanLuck) rotateBanker() {
	next := -1
	best := BanLuckRankBust
	for i := range g.players {
		if i == g.banker || g.results[i].Outcome != BanLuckOutcomeWin {
			continue
		}
		// 特別役で勝った席だけが親を奪える。普通の勝ちでは奪えない。
		if r := g.results[i].Rank; r > BanLuckRankPoint && r > best {
			best, next = r, i
		}
	}
	if next < 0 {
		next = (g.banker + 1) % len(g.players)
	}
	g.banker = next
}

// NextRound は次のラウンドを始める。
func (g *BanLuck) NextRound() error {
	if g.gameEndFlag {
		return errBanLuckFinished
	}
	if g.phase != BanLuckPhaseRoundEnd {
		return errBanLuckWrongPhase
	}
	// **人間が賭けられなくなったら終わる。** 最小額すら払えない席はどの額でも
	// `PlaceBet` に拒否されるので、他の席にチップが残っていると局が終わらず、
	// 画面はエラーを出し続けるだけの行き止まりになる。親のラウンドなら賭けずに
	// 打てるが、その次でまた詰まるので、ここで区切る。
	if g.roundNum >= g.config.Rounds || g.aliveSeats() < BanLuckMinSeats || g.humanIsBroke() {
		g.finish()
		return nil
	}
	g.roundNum++
	g.phase = BanLuckPhaseBet
	// **前のラウンドの盤面は持ち越さない。** 賭ける前に手札が残っていると、
	// 保存が「配る前なのに配られている」状態になる (往復テストが捕まえた)。
	// 次の `deal` が作り直すので、消しても失うものは無い。
	g.hands = nil
	g.results = nil
	g.settled = false
	for _, p := range g.players {
		p.SetBet(0)
	}
	return nil
}

// humanIsBroke は人間の席が最小の賭け金すら置けないかを返す。
func (g *BanLuck) humanIsBroke() bool {
	return g.players[g.humanSeat()].GetChips() < BanLuckMinBet
}

// aliveSeats はまだチップが残っている席の数を返す。
func (g *BanLuck) aliveSeats() int {
	n := 0
	for _, p := range g.players {
		if p.GetChips() > 0 {
			n++
		}
	}
	return n
}

// finish はゲームを終える。
func (g *BanLuck) finish() {
	g.gameEndFlag = true
	g.phase = BanLuckPhaseGameEnd
	g.appendLog(-1, "gameEnd", fmt.Sprintf("winner seat %d", g.WinnerSeat()), nil)
}

// WinnerSeat はチップがいちばん多い席を返す。同点なら若い席。
func (g *BanLuck) WinnerSeat() int {
	best, bestChips := 0, -1
	for i, p := range g.players {
		if p.GetChips() > bestChips {
			best, bestChips = i, p.GetChips()
		}
	}
	return best
}

// --- 参照 ---

// humanSeat は人間の席を返す。無ければ 0。
func (g *BanLuck) humanSeat() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return 0
}

// humanIsBanker は人間が親かを返す。
func (g *BanLuck) humanIsBanker() bool { return g.banker == g.humanSeat() }

// IsHumanTurn は人間の操作待ちかを返す。
func (g *BanLuck) IsHumanTurn() bool {
	return g.phase == BanLuckPhasePlay && g.turn == g.humanSeat()
}

// GetConfig はゲーム設定を返す。
func (g *BanLuck) GetConfig() BanLuckConfig { return g.config }

// SetConfig はゲーム設定を設定する。
func (g *BanLuck) SetConfig(c BanLuckConfig) { g.config = c }

// GetPhase は現在のフェーズを返す。
func (g *BanLuck) GetPhase() BanLuckPhase { return g.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *BanLuck) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayers は席の一覧を返す。
func (g *BanLuck) GetPlayers() []*BanLuckPlayer { return g.players }

// GetHands は席ごとの手札を返す。
func (g *BanLuck) GetHands() []*BlackJackHand { return g.hands }

// GetResults は席ごとのラウンド結果を返す。
func (g *BanLuck) GetResults() []BanLuckSeatResult { return g.results }

// GetBankerSeat は親の席を返す。
func (g *BanLuck) GetBankerSeat() int { return g.banker }

// GetTurnSeat はいま操作している席を返す。
func (g *BanLuck) GetTurnSeat() int { return g.turn }

// GetHumanSeat は人間の席を返す。
func (g *BanLuck) GetHumanSeat() int { return g.humanSeat() }

// GetRoundNumber は現在のラウンド数を返す。
func (g *BanLuck) GetRoundNumber() int { return g.roundNum }

// GetRemainingCards は山の残り枚数を返す。
func (g *BanLuck) GetRemainingCards() int { return g.deck.GetRemainingCount() }

// GetActionLog は棋譜を返す。
func (g *BanLuck) GetActionLog() []*ActionLogEntry { return g.actionLog }

// appendLog は棋譜に 1 行足す。
func (g *BanLuck) appendLog(seat int, actionType, detail string, cards []*Card) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  seat,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
	if len(g.actionLog) > banLuckMaxSliceLen {
		g.actionLog = g.actionLog[len(g.actionLog)-banLuckMaxSliceLen:]
	}
}

// --- 永続化 ---

// banLuckJSON is the JSON wire format for BanLuck.
type banLuckJSON struct {
	Deck        *TrumpCards       `json:"dk"`
	Players     []*BanLuckPlayer  `json:"pl"`
	Config      BanLuckConfig     `json:"cf"`
	Phase       int               `json:"ph"`
	Hands       []*BlackJackHand  `json:"hd"`
	Ranks       []int             `json:"rk"`
	Outcomes    []int             `json:"oc"`
	Deltas      []int             `json:"dl"`
	Settled     bool              `json:"st"`
	Banker      int               `json:"bk"`
	Turn        int               `json:"tu"`
	RoundNumber int               `json:"rn"`
	GameEndFlag bool              `json:"ge"`
	ActionLog   []*ActionLogEntry `json:"al"`
	TurnNumber  int               `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (g *BanLuck) MarshalJSON() ([]byte, error) {
	ranks := make([]int, 0, len(g.results))
	outcomes := make([]int, 0, len(g.results))
	deltas := make([]int, 0, len(g.results))
	for _, r := range g.results {
		ranks = append(ranks, int(r.Rank))
		outcomes = append(outcomes, int(r.Outcome))
		deltas = append(deltas, r.Delta)
	}
	return json.Marshal(banLuckJSON{
		Deck: g.deck, Players: g.players, Config: g.config,
		Phase: int(g.phase), Hands: g.hands,
		Ranks: ranks, Outcomes: outcomes, Deltas: deltas,
		Settled: g.settled, Banker: g.banker, Turn: g.turn,
		RoundNumber: g.roundNum, GameEndFlag: g.gameEndFlag,
		ActionLog: g.actionLog, TurnNumber: g.turnNumber,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **席の添字が 2 つある (親と手番) のが危ないところ。** どちらも範囲外なら
// その場で落ちるが、**範囲内でも局面と食い違う**ことがある ── 配る前なのに
// 手番があったり、親が居ない席を指していたり。範囲検査だけ書くと、そういう
// 保存が素通りして勝敗だけが静かに変わる。
func (g *BanLuck) UnmarshalJSON(data []byte) error {
	var j banLuckJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := banLuckValidate(&j); err != nil {
		return err
	}

	g.deck = j.Deck
	if g.deck == nil {
		g.deck = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = BanLuckPhase(j.Phase)
	g.hands = j.Hands
	g.results = make([]BanLuckSeatResult, 0, len(j.Ranks))
	for i := range j.Ranks {
		g.results = append(g.results, BanLuckSeatResult{
			Rank:    BanLuckRank(j.Ranks[i]),
			Outcome: BanLuckOutcome(j.Outcomes[i]),
			Delta:   j.Deltas[i],
		})
	}
	g.settled = j.Settled
	g.banker = j.Banker
	g.turn = j.Turn
	g.roundNum = j.RoundNumber
	g.gameEndFlag = j.GameEndFlag
	g.actionLog = j.ActionLog
	g.turnNumber = j.TurnNumber
	return nil
}

// banLuckValidate は保存データの範囲と整合を検証する。
func banLuckValidate(j *banLuckJSON) error {
	if err := j.Config.Validate(); err != nil {
		return err
	}
	seats := len(j.Players)
	if seats < BanLuckMinSeats || seats > BanLuckMaxSeats {
		return fmt.Errorf("banluck: %d seats out of range", seats)
	}
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("banluck: seat %d is missing", i)
		}
	}
	if j.Phase < int(BanLuckPhaseBet) || j.Phase > int(BanLuckPhaseMax) {
		return fmt.Errorf("banluck: phase out of range: %d", j.Phase)
	}
	if j.Banker < 0 || j.Banker >= seats {
		return fmt.Errorf("banluck: banker seat out of range: %d", j.Banker)
	}
	if j.Turn < 0 || j.Turn >= seats {
		return fmt.Errorf("banluck: turn seat out of range: %d", j.Turn)
	}
	if j.RoundNumber < 1 || j.RoundNumber > j.Config.Rounds {
		return fmt.Errorf("banluck: round number out of range: %d", j.RoundNumber)
	}
	if err := banLuckValidateSeatSlices(j, seats); err != nil {
		return err
	}
	return banLuckValidatePhase(j, seats)
}

// banLuckValidateSeatSlices は席ごとのスライスが揃っていることを見る。
//
// **手札・役・決着・増減は必ず席と同数。** ここがずれると、精算が別の席の
// 金を動かしても添字は範囲内のままなので、どこにも例外が出ない。
func banLuckValidateSeatSlices(j *banLuckJSON, seats int) error {
	if len(j.Hands) != 0 && len(j.Hands) != seats {
		return fmt.Errorf("banluck: %d hands for %d seats", len(j.Hands), seats)
	}
	for _, n := range []struct {
		name string
		got  int
	}{{"ranks", len(j.Ranks)}, {"outcomes", len(j.Outcomes)}, {"deltas", len(j.Deltas)}} {
		if n.got != len(j.Hands) {
			return fmt.Errorf("banluck: %d %s for %d hands", n.got, n.name, len(j.Hands))
		}
	}
	for _, r := range j.Ranks {
		if r < int(BanLuckRankBust) || r > int(BanLuckRankMax) {
			return fmt.Errorf("banluck: rank out of range: %d", r)
		}
	}
	for _, o := range j.Outcomes {
		if o < int(BanLuckOutcomeLose) || o > int(BanLuckOutcomeWin) {
			return fmt.Errorf("banluck: outcome out of range: %d", o)
		}
	}
	if len(j.ActionLog) > banLuckMaxSliceLen {
		return fmt.Errorf("banluck: action log too long: %d", len(j.ActionLog))
	}
	return nil
}

// banLuckValidatePhase はフェーズと盤面の整合を見る。
//
// **範囲チェックでは捕まらない食い違いがここ。** 「配る前なのに手札がある」
// 「配った後なのに手札が無い」「親に賭け金が乗っている」はどれも添字としては
// 正当で、通すと勝敗だけが静かに変わる。
func banLuckValidatePhase(j *banLuckJSON, seats int) error {
	dealt := len(j.Hands) == seats
	switch BanLuckPhase(j.Phase) {
	case BanLuckPhaseBet:
		if dealt {
			return fmt.Errorf("banluck: cards are dealt before the bets are placed")
		}
	case BanLuckPhasePlay, BanLuckPhaseRoundEnd:
		if !dealt {
			return fmt.Errorf("banluck: no cards dealt in phase %d", j.Phase)
		}
	default:
	}
	if BanLuckPhase(j.Phase) == BanLuckPhasePlay && j.Settled {
		return fmt.Errorf("banluck: the round is settled but still in play")
	}
	// **親は賭けない。** 賭け金の乗った親は精算で自分から取ることになる。
	if j.Players[j.Banker].GetBet() != 0 {
		return fmt.Errorf("banluck: the banker seat %d carries a bet", j.Banker)
	}
	return nil
}

// --- 助言 ---

// BanLuckHint は人間への助言。
type BanLuckHint struct {
	// Action は薦める操作 ("hit" / "stand")。
	Action string
	// Reason は理由の識別子 (i18n キーの一部)。
	Reason string
}

// GetHint は人間への助言を返す。判断どころでなければ nil。
//
// **義務のある場面では「引け」しか言わない。** 選べないところで戦略を語ると、
// 押せないボタンを薦めることになる。
func (g *BanLuck) GetHint() *BanLuckHint {
	if g.gameEndFlag || g.phase != BanLuckPhasePlay || !g.IsHumanTurn() {
		return nil
	}
	h := g.hands[g.turn]
	if h == nil {
		return nil
	}
	if g.turn == g.banker && g.bankerMustHit() {
		return &BanLuckHint{Action: "hit", Reason: "bankerMustHit"}
	}
	if h.GetCardsSize() >= BanLuckMaxHandCards {
		return &BanLuckHint{Action: "stand", Reason: "handFull"}
	}
	// **4 枚目まで来たら Five Dragon を狙う価値がある。** 5 枚で 21 以下なら
	// 合計に関係なく普通の手に勝てるので、17 でも引くほうが得な場面が出る。
	if h.GetCardsSize() == BanLuckMaxHandCards-1 && h.GetScore() <= banLuckCpuDragonChase {
		return &BanLuckHint{Action: "hit", Reason: "chaseFiveDragon"}
	}
	if h.GetScore() <= 11 {
		return &BanLuckHint{Action: "hit", Reason: "cannotBust"}
	}
	if h.GetScore() >= banLuckCpuSeatStand {
		return &BanLuckHint{Action: "stand", Reason: "standPat"}
	}
	return &BanLuckHint{Action: "hit", Reason: "chaseBanker"}
}
