//go:build !js || !wasm || solo

// Package domain クオドリベット (Quodlibet) のドメインモデル。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// QuodlibetPlayerCnt はプレイヤー数。
const QuodlibetPlayerCnt = 4

// QuodlibetHandSize は 1 人の手札枚数 (32 枚 ÷ 4 人)。
const QuodlibetHandSize = 8

// QuodlibetRoundCnt は「輪」(Rad) の数。
const QuodlibetRoundCnt = 3

// QuodlibetContractsPerRound は 1 つの輪に属するコントラクト数。
const QuodlibetContractsPerRound = 4

// QuodlibetTotalDeals は 1 ゲームの総ディール数 (3 輪 × 4 種目)。
//
// **ディーラー 1 人につき 1 種目ではない。** 12 ディールを 3 つの輪に分け、
// 輪ごとにその輪の 4 種目を 1 回ずつ消化する。
const QuodlibetTotalDeals = QuodlibetRoundCnt * QuodlibetContractsPerRound

// QuodlibetContractCnt は全コントラクト数。
const QuodlibetContractCnt = QuodlibetTotalDeals

// コントラクト定数。輪の順に並ぶ (0-3 が第 1 の輪、4-7 が第 2、8-11 が第 3)。
const (
	// QuodlibetPlus プラス。取れなかったトリック 1 つにつき罰点。
	QuodlibetPlus = 0
	// QuodlibetMinus マイナス。取ったトリック 1 つにつき罰点。
	QuodlibetMinus = 1
	// QuodlibetBadNeighbour 悪い隣人。マイナスと同じ点が「右隣」に付く。
	QuodlibetBadNeighbour = 2
	// QuodlibetAlarich アラリック。K♥ と Q♦ を取ると罰点。
	QuodlibetAlarich = 3
	// QuodlibetFirstThreeAndLast 1238。最初の 3 トリックと最後のトリックに罰点。
	QuodlibetFirstThreeAndLast = 4
	// QuodlibetNoReds 赤なし。ハートを取ると罰点 (**低い札のほうが重い**)。
	QuodlibetNoReds = 5
	// QuodlibetOberUnter オーバー / ウンター。Q と J を取ると罰点。
	QuodlibetOberUnter = 6
	// QuodlibetBribe 賄賂。トリックを取っても、最低の札を出しても罰点。
	QuodlibetBribe = 7
	// QuodlibetOpen 開いたズボン。自分の手札だけが見えない状態で打つ。
	QuodlibetOpen = 8
	// QuodlibetHunt 狩猟。全員の手札が見える状態で打つ。
	QuodlibetHunt = 9
	// QuodlibetQuadrature 四分。同じスートのちょうど 3 つ上でしか重ねられない。
	QuodlibetQuadrature = 10
	// QuodlibetSnack 小食い。7 並べで手札を早く出し切る。
	QuodlibetSnack = 11
)

// フェーズ。
const (
	// QuodlibetPhaseSelectContract ディーラーがコントラクトを選択中。
	QuodlibetPhaseSelectContract = "selectContract"
	// QuodlibetPhasePlay プレイ中。
	QuodlibetPhasePlay = "play"
	// QuodlibetPhaseDealEnd 1 ディール終了 (罰点確定、次ディール待ち)。
	QuodlibetPhaseDealEnd = "dealEnd"
	// QuodlibetPhaseGameEnd ゲーム終了 (12 ディール完了)。
	QuodlibetPhaseGameEnd = "gameEnd"
)

// QuodlibetDealDetail は 1 ディールの罰点内訳。
type QuodlibetDealDetail struct {
	// Contract は打たれたコントラクト。
	Contract int
	// Round は輪の番号 (0-2)。
	Round int
	// DealerIdx はディーラー (Bierkönig)。
	DealerIdx int
	// Points はプレイヤー別のこのディールの罰点。
	Points map[int]int
}

// Quodlibet はクオドリベットの状態を保持する集約ルート。
type Quodlibet struct {
	trumpCards      *TrumpCards
	players         []*QuodlibetPlayer
	config          QuodlibetConfig
	phase           string
	dealNumber      int // 0..QuodlibetTotalDeals
	dealerIdx       int
	currentContract int // -1 = 未選択
	usedContracts   [QuodlibetContractCnt]bool
	currentPlayer   int
	leadPlayer      int
	trickNumber     int
	currentTrick    []*TrickCard
	lastTrick       []*TrickCard
	lastTrickWinner int
	trickWinners    []int          // このディールのトリック勝者 (index = トリック番号 - 1)
	trickRecord     [][]*TrickCard // このディールの全トリック (席つき。賄賂の採点に要る)
	tablePlaced     [5]uint16      // Snack の場 (index 1-4 = スート)
	stack           []*Card        // Quadrature の現在の重ね
	passCount       [QuodlibetPlayerCnt]int
	outCount        int // シェディング系で上がった人数
	gameEndFlag     bool
	lastDealDetail  *QuodlibetDealDetail
	dealHistory     []*QuodlibetDealDetail
	actionLogBase
}

// NewQuodlibet はコンストラクタ。
func NewQuodlibet(trumpCards *TrumpCards, players []*QuodlibetPlayer, config QuodlibetConfig) *Quodlibet {
	return &Quodlibet{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		phase:           QuodlibetPhaseSelectContract,
		currentContract: -1,
		lastTrickWinner: -1,
	}
}

// NewDefaultQuodlibet は標準の 4 人構成 (1 human + 3 CPU) を生成する。
func NewDefaultQuodlibet() *Quodlibet {
	players := make([]*QuodlibetPlayer, QuodlibetPlayerCnt)
	players[0] = NewQuodlibetPlayer(true)
	for i := 1; i < QuodlibetPlayerCnt; i++ {
		players[i] = NewQuodlibetPlayer(false)
	}
	return NewQuodlibet(NewTrumpCards32(), players, DefaultQuodlibetConfig())
}

// QuodlibetRoundOf はコントラクトが属する輪 (0-2) を返す。
func QuodlibetRoundOf(contract int) int {
	if contract < 0 || contract >= QuodlibetContractCnt {
		return -1
	}
	return contract / QuodlibetContractsPerRound
}

// QuodlibetRightOf は「右隣」の席を返す。
//
// **反時計回りに進む卓なので、右隣は次の手番の人。** 悪い隣人 (Böser Nachbar)
// の罰点はここへ流れるので、向きを取り違えると点が真逆の人に付く。
func QuodlibetRightOf(playerIdx int) int {
	return (playerIdx + 1) % QuodlibetPlayerCnt
}

// QuodlibetIsSheddingContract はトリックを取らないコントラクトかを返す。
func QuodlibetIsSheddingContract(contract int) bool {
	return contract == QuodlibetQuadrature || contract == QuodlibetSnack
}

// Reset は新しいゲームを開始する。累計罰点もクリアする。
func (q *Quodlibet) Reset() {
	for _, p := range q.players {
		p.ResetDeal()
		p.ResetPenalty()
	}
	q.dealNumber = 0
	q.gameEndFlag = false
	q.usedContracts = [QuodlibetContractCnt]bool{}
	q.lastDealDetail = nil
	q.dealHistory = make([]*QuodlibetDealDetail, 0, QuodlibetTotalDeals)
	q.actionLog = make([]*ActionLogEntry, 0)
	q.startDeal()
}

// NextDeal は次のディールを開始する。12 ディール完了済みなら終局する。
func (q *Quodlibet) NextDeal() {
	if q.gameEndFlag || q.phase != QuodlibetPhaseDealEnd {
		return
	}
	q.dealNumber++
	if q.dealNumber >= QuodlibetTotalDeals {
		q.gameEndFlag = true
		q.phase = QuodlibetPhaseGameEnd
		q.appendLog(-1, "gameEnd", "all 12 deals completed", nil)
		return
	}
	q.startDeal()
}

// startDeal はディーラーを決め、8 枚ずつ配り、コントラクト選択フェーズへ移る。
func (q *Quodlibet) startDeal() {
	q.dealerIdx = q.dealNumber % QuodlibetPlayerCnt
	q.currentContract = -1
	q.currentTrick = nil
	q.lastTrick = nil
	q.lastTrickWinner = -1
	q.trickNumber = 0
	q.trickWinners = nil
	q.trickRecord = nil
	q.tablePlaced = [5]uint16{}
	q.stack = nil
	q.passCount = [QuodlibetPlayerCnt]int{}
	q.outCount = 0
	for _, p := range q.players {
		p.ResetDeal()
	}
	q.trumpCards = NewTrumpCards32()
	q.trumpCards.Shuffle()
	dealAllCards(q.trumpCards, q.players)
	for _, p := range q.players {
		quodlibetSortHand(p)
	}
	q.phase = QuodlibetPhaseSelectContract
	q.appendLog(-1, "deal", fmt.Sprintf("deal %d/%d, round %d, dealer=%d",
		q.dealNumber+1, QuodlibetTotalDeals, q.GetRoundNumber(), q.dealerIdx), nil)
	if q.config.AutoSelectContract {
		_ = q.applySelectContract(q.autoContract())
	}
}

// GetRoundNumber は現在の輪 (1-3) を返す。
func (q *Quodlibet) GetRoundNumber() int {
	return q.dealNumber/QuodlibetContractsPerRound + 1
}

// GetAvailableContracts はこの輪でまだ打たれていないコントラクトを返す。
//
// **選べるのはその輪の 4 種目だけ。** 輪をまたいで選べるようにすると、
// 「3 つの輪で性格の違う 4 種目ずつを消化する」という構成が崩れる。
func (q *Quodlibet) GetAvailableContracts() []int {
	round := q.dealNumber / QuodlibetContractsPerRound
	if round < 0 || round >= QuodlibetRoundCnt {
		return nil
	}
	out := make([]int, 0, QuodlibetContractsPerRound)
	for i := 0; i < QuodlibetContractsPerRound; i++ {
		c := round*QuodlibetContractsPerRound + i
		if !q.usedContracts[c] {
			out = append(out, c)
		}
	}
	return out
}

// autoContract は自動選択でのコントラクトを返す (輪の残りの先頭)。
func (q *Quodlibet) autoContract() int {
	avail := q.GetAvailableContracts()
	if len(avail) == 0 {
		return -1
	}
	return avail[0]
}

// SelectContract はディーラーがコントラクトを選択する。
func (q *Quodlibet) SelectContract(contract int) error {
	if q.gameEndFlag {
		return ErrGameEnded
	}
	if q.phase != QuodlibetPhaseSelectContract {
		return ErrWrongPhase
	}
	if !q.players[q.dealerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return q.applySelectContract(contract)
}

// applySelectContract はコントラクトを確定してプレイフェーズへ移す。
func (q *Quodlibet) applySelectContract(contract int) error {
	avail := q.GetAvailableContracts()
	ok := false
	for _, c := range avail {
		if c == contract {
			ok = true
			break
		}
	}
	if !ok {
		return NewDomainErrorCode(ErrInvalidPlay, "quodlibet.errContractUnavailable",
			map[string]string{"contract": QuodlibetContractName(contract)})
	}
	q.currentContract = contract
	q.usedContracts[contract] = true
	q.phase = QuodlibetPhasePlay
	q.trickNumber = 1
	// **リードはディーラーの右隣。** 反時計回りなので次の席から始まる。
	q.leadPlayer = QuodlibetRightOf(q.dealerIdx)
	q.currentPlayer = q.leadPlayer
	q.appendLog(q.dealerIdx, "contract",
		fmt.Sprintf("dealer %d chooses %s", q.dealerIdx, QuodlibetContractName(contract)), nil)
	return nil
}

// CpuSelectContract は CPU のディーラーがコントラクトを選ぶ。
func (q *Quodlibet) CpuSelectContract() {
	if q.gameEndFlag || q.phase != QuodlibetPhaseSelectContract {
		return
	}
	if q.players[q.dealerIdx].GetIsHuman() {
		return
	}
	_ = q.applySelectContract(q.cpuPickContract())
}

// PlayerPlay は人間が 1 手指す。handIdx == -1 はパス (Snack で合法手が無い場合)。
func (q *Quodlibet) PlayerPlay(handIdx int) error {
	if q.gameEndFlag {
		return ErrGameEnded
	}
	if q.phase != QuodlibetPhasePlay {
		return ErrWrongPhase
	}
	if !q.players[q.currentPlayer].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if QuodlibetIsSheddingContract(q.currentContract) {
		return q.applySheddingPlay(q.currentPlayer, handIdx)
	}
	return q.applyTrickPlay(q.currentPlayer, handIdx)
}

// CpuPlay は CPU が 1 手指す。
func (q *Quodlibet) CpuPlay() {
	if q.gameEndFlag || q.phase != QuodlibetPhasePlay {
		return
	}
	idx := q.currentPlayer
	if q.players[idx].GetIsHuman() {
		return
	}
	if QuodlibetIsSheddingContract(q.currentContract) {
		_ = q.applySheddingPlay(idx, q.cpuSheddingChoice(idx))
		return
	}
	_ = q.applyTrickPlay(idx, q.cpuTrickChoice(idx))
}

// applyTrickPlay はトリックへ 1 枚出す。
func (q *Quodlibet) applyTrickPlay(playerIdx, handIdx int) error {
	player := q.players[playerIdx]
	if handIdx < 0 || handIdx >= player.GetCardsSize() {
		return NewDomainErrorCode(ErrInvalidCard, "quodlibet.errCardRange", nil)
	}
	card := player.GetCard(handIdx)
	if err := validateCardIsPlayable(q.GetPlayableIndices(playerIdx), player, card); err != nil {
		return err
	}
	played := player.RemoveCard(handIdx)
	q.currentTrick = append(q.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: played})
	q.appendLog(playerIdx, "play",
		fmt.Sprintf("player %d plays %s", playerIdx, cardStr(played)), []*Card{played})

	if len(q.currentTrick) < QuodlibetPlayerCnt {
		q.currentPlayer = QuodlibetRightOf(q.currentPlayer)
		return nil
	}
	q.resolveTrick()
	return nil
}

// GetPlayableIndices は出せる手札のインデックスを返す。
//
// **フォロー義務はコントラクトによらず常にある** ── ただし Quadrature だけは
// トリックではなく重ねなので、専用の判定を使う。
func (q *Quodlibet) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(q.players) {
		return nil
	}
	if QuodlibetIsSheddingContract(q.currentContract) {
		return q.GetSheddingPlayableIndices(playerIdx)
	}
	player := q.players[playerIdx]
	all := make([]int, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		all = append(all, i)
	}
	if len(q.currentTrick) == 0 {
		return all
	}
	lead := q.currentTrick[0].Card.GetDesign()
	follow := make([]int, 0, len(all))
	for _, i := range all {
		if c := player.GetCard(i); c != nil && c.GetDesign() == lead {
			follow = append(follow, i)
		}
	}
	if len(follow) == 0 {
		return all
	}
	return follow
}

// resolveTrick はトリックを解決して次へ進める。
func (q *Quodlibet) resolveTrick() {
	winner := q.trickWinner()
	cards := make([]*Card, 0, QuodlibetPlayerCnt)
	for _, tc := range q.currentTrick {
		if tc != nil {
			cards = append(cards, tc.Card)
		}
	}
	q.players[winner].AddTrick(cards)
	q.trickWinners = append(q.trickWinners, winner)
	q.trickRecord = append(q.trickRecord, q.currentTrick)
	q.lastTrick = q.currentTrick
	q.lastTrickWinner = winner
	q.appendLog(winner, "trick_win",
		fmt.Sprintf("player %d wins trick %d", winner, q.trickNumber), cards)
	q.currentTrick = nil

	if q.trickNumber >= QuodlibetHandSize {
		q.finishDeal()
		return
	}
	q.trickNumber++
	q.leadPlayer = winner
	q.currentPlayer = winner
}

// trickWinner は台札スートの最強札を出した席を返す (切り札は無い)。
func (q *Quodlibet) trickWinner() int {
	if len(q.currentTrick) == 0 {
		return q.leadPlayer
	}
	lead := q.currentTrick[0].Card.GetDesign()
	winner, best := q.currentTrick[0].PlayerIdx, -1
	for _, tc := range q.currentTrick {
		if tc == nil || tc.Card == nil || tc.Card.GetDesign() != lead {
			continue
		}
		if s := QuodlibetCardStrength(tc.Card); s > best {
			best, winner = s, tc.PlayerIdx
		}
	}
	return winner
}

// QuodlibetCardStrength は 32 枚デッキでの札の強さを返す (A が最強、7 が最弱)。
func QuodlibetCardStrength(c *Card) int {
	if c == nil {
		return -1
	}
	if c.GetValue() == 1 {
		return 14 // エース
	}
	return c.GetValue()
}

// finishDeal はディールを締めて罰点を確定する。
func (q *Quodlibet) finishDeal() {
	detail := q.scoreDeal()
	for i, p := range q.players {
		p.SetDealPoints(detail.Points[i])
		p.AddPenalty(detail.Points[i])
	}
	q.lastDealDetail = detail
	q.dealHistory = append(q.dealHistory, detail)
	q.phase = QuodlibetPhaseDealEnd
	q.appendLog(-1, "dealEnd",
		fmt.Sprintf("deal %d scored: %s", q.dealNumber+1, quodlibetPointsStr(detail.Points)), nil)
}

// quodlibetPointsStr は罰点内訳をログ用の文字列にする。
func quodlibetPointsStr(points map[int]int) string {
	parts := make([]string, 0, len(points))
	for i := 0; i < QuodlibetPlayerCnt; i++ {
		parts = append(parts, fmt.Sprintf("p%d=%d", i, points[i]))
	}
	return fmt.Sprint(parts)
}

// quodlibetSortHand はスート別・強い順に手札を並べる。
func quodlibetSortHand(p *QuodlibetPlayer) {
	cards := make([]*Card, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards = append(cards, p.GetCard(i))
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return QuodlibetCardStrength(cards[i]) > QuodlibetCardStrength(cards[j])
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// IsHumanTurn は人間の手番かを返す。
func (q *Quodlibet) IsHumanTurn() bool {
	if q.gameEndFlag {
		return false
	}
	if q.phase == QuodlibetPhaseSelectContract {
		return q.players[q.dealerIdx].GetIsHuman()
	}
	if q.phase != QuodlibetPhasePlay {
		return false
	}
	if q.currentPlayer < 0 || q.currentPlayer >= len(q.players) {
		return false
	}
	return q.players[q.currentPlayer].GetIsHuman()
}

// QuodlibetHint は人間への推奨手。
type QuodlibetHint struct {
	// CardIndices は勧める手札のインデックス。
	CardIndices []int
	// Contract は勧めるコントラクト (選択フェーズのみ、それ以外は -1)。
	Contract int
	// Reason は理由の識別子。
	Reason string
}

// GetHint は人間への推奨手を返す。
func (q *Quodlibet) GetHint() *QuodlibetHint {
	human := findHumanIdx(q.players)
	if human < 0 || q.gameEndFlag {
		return &QuodlibetHint{Contract: -1, Reason: "none"}
	}
	switch q.phase {
	case QuodlibetPhaseSelectContract:
		if !q.players[q.dealerIdx].GetIsHuman() {
			return &QuodlibetHint{Contract: -1, Reason: "none"}
		}
		return &QuodlibetHint{Contract: q.smartPickContractFor(q.dealerIdx), Reason: "pick_contract"}
	case QuodlibetPhasePlay:
		if q.currentPlayer != human {
			return &QuodlibetHint{Contract: -1, Reason: "none"}
		}
		// **助言は CPU の難易度に引きずらせない。** cpuTrickChoice /
		// cpuSheddingChoice は Easy だと合法手からランダムに選ぶので、そのまま
		// 使うと「Easy を選んだ人にだけでたらめな札を勧める」ことになる。
		if QuodlibetIsSheddingContract(q.currentContract) {
			valid := q.GetSheddingPlayableIndices(human)
			if len(valid) == 0 {
				return &QuodlibetHint{Contract: -1, Reason: "pass"}
			}
			return &QuodlibetHint{CardIndices: []int{q.smartSheddingChoice(human, valid)}, Contract: -1, Reason: "shed_low"}
		}
		valid := q.GetPlayableIndices(human)
		if len(valid) == 0 {
			return &QuodlibetHint{Contract: -1, Reason: "none"}
		}
		return &QuodlibetHint{
			CardIndices: []int{q.quodlibetSmartTrickChoice(human, valid)},
			Contract:    -1,
			Reason:      "avoid_penalty",
		}
	case QuodlibetPhaseDealEnd:
		return &QuodlibetHint{Contract: -1, Reason: "next_deal"}
	default:
		return &QuodlibetHint{Contract: -1, Reason: "none"}
	}
}

// QuodlibetContractName はコントラクトの安定した識別名を返す (i18n キー用)。
func QuodlibetContractName(c int) string {
	switch c {
	case QuodlibetPlus:
		return "plus"
	case QuodlibetMinus:
		return "minus"
	case QuodlibetBadNeighbour:
		return "badNeighbour"
	case QuodlibetAlarich:
		return "alarich"
	case QuodlibetFirstThreeAndLast:
		return "firstThreeAndLast"
	case QuodlibetNoReds:
		return "noReds"
	case QuodlibetOberUnter:
		return "oberUnter"
	case QuodlibetBribe:
		return "bribe"
	case QuodlibetOpen:
		return "open"
	case QuodlibetHunt:
		return "hunt"
	case QuodlibetQuadrature:
		return "quadrature"
	case QuodlibetSnack:
		return "snack"
	default:
		return "unknown"
	}
}

// QuodlibetHandVisibility は手札の見え方を返す。
//
// **第 3 の輪では手札の見え方そのものが規則。** 「開いたズボン」は自分の手札
// だけが見えず、「狩猟」は全員の手札が見える。
func QuodlibetHandVisibility(contract, viewerIdx, ownerIdx int) bool {
	switch contract {
	case QuodlibetOpen:
		return viewerIdx != ownerIdx
	case QuodlibetHunt:
		return true
	default:
		return viewerIdx == ownerIdx
	}
}

// GetWinners は罰点が最も少ないプレイヤー (複数可) を返す。
func (q *Quodlibet) GetWinners() []int {
	if len(q.players) == 0 {
		return nil
	}
	best := q.players[0].GetPenalty()
	for _, p := range q.players[1:] {
		if p.GetPenalty() < best {
			best = p.GetPenalty()
		}
	}
	winners := make([]int, 0, len(q.players))
	for i, p := range q.players {
		if p.GetPenalty() == best {
			winners = append(winners, i)
		}
	}
	return winners
}

// --- 参照 ---

// GetConfig はゲーム設定を返す。
func (q *Quodlibet) GetConfig() QuodlibetConfig { return q.config }

// SetConfig はゲーム設定を設定する。
func (q *Quodlibet) SetConfig(c QuodlibetConfig) { q.config = c }

// GetPhase は現在のフェーズを返す。
func (q *Quodlibet) GetPhase() string { return q.phase }

// GetPlayerCnt はプレイヤー数を返す。
func (q *Quodlibet) GetPlayerCnt() int { return len(q.players) }

// GetPlayer は席のプレイヤーを返す。
func (q *Quodlibet) GetPlayer(i int) *QuodlibetPlayer {
	if i < 0 || i >= len(q.players) {
		return nil
	}
	return q.players[i]
}

// GetPlayers は全プレイヤーを返す。
func (q *Quodlibet) GetPlayers() []*QuodlibetPlayer { return q.players }

// GetDealNumber は現在のディール index (0 始まり) を返す。
func (q *Quodlibet) GetDealNumber() int { return q.dealNumber }

// GetDealerIdx はディーラー (Bierkönig) の席を返す。
func (q *Quodlibet) GetDealerIdx() int { return q.dealerIdx }

// GetCurrentContract は現在のコントラクトを返す (-1 = 未選択)。
func (q *Quodlibet) GetCurrentContract() int { return q.currentContract }

// GetCurrentTurn は手番の席を返す。
func (q *Quodlibet) GetCurrentTurn() int { return q.currentPlayer }

// GetCurrentPlayerIdx は手番の席を返す (共通インタフェース用の別名)。
func (q *Quodlibet) GetCurrentPlayerIdx() int { return q.currentPlayer }

// GetLeadPlayerIdx はリードの席を返す。
func (q *Quodlibet) GetLeadPlayerIdx() int { return q.leadPlayer }

// GetTrickNumber は現在のトリック番号 (1 始まり) を返す。
func (q *Quodlibet) GetTrickNumber() int { return q.trickNumber }

// GetCurrentTrick は進行中のトリックを返す。
func (q *Quodlibet) GetCurrentTrick() []*TrickCard { return q.currentTrick }

// GetLastTrick は直前に完了したトリックを返す。
func (q *Quodlibet) GetLastTrick() []*TrickCard { return q.lastTrick }

// GetLastTrickWinner は直前トリックの勝者を返す (-1 = なし)。
func (q *Quodlibet) GetLastTrickWinner() int { return q.lastTrickWinner }

// GetTablePlaced は Snack の場を返す (index 1-4 = スート)。
func (q *Quodlibet) GetTablePlaced() [5]uint16 { return q.tablePlaced }

// GetStack は Quadrature の現在の重ねを返す。
func (q *Quodlibet) GetStack() []*Card { return q.stack }

// GetUsedContracts は消化済みコントラクトを返す。
func (q *Quodlibet) GetUsedContracts() [QuodlibetContractCnt]bool { return q.usedContracts }

// GetLastDealDetail は直近ディールの罰点内訳を返す。
func (q *Quodlibet) GetLastDealDetail() *QuodlibetDealDetail { return q.lastDealDetail }

// GetDealHistory は完了した各ディールの罰点内訳を返す。
func (q *Quodlibet) GetDealHistory() []*QuodlibetDealDetail { return q.dealHistory }

// GetGameEndFlag は終局フラグを返す。
func (q *Quodlibet) GetGameEndFlag() bool { return q.gameEndFlag }

// --- 永続化 ---

// quodlibetJSON is the JSON wire format for Quodlibet.
type quodlibetJSON struct {
	TrumpCards      *TrumpCards                `json:"tc"`
	Players         []*QuodlibetPlayer         `json:"pl"`
	Config          QuodlibetConfig            `json:"cf"`
	Phase           string                     `json:"ph"`
	DealNumber      int                        `json:"dn"`
	DealerIdx       int                        `json:"di"`
	CurrentContract int                        `json:"cc"`
	UsedContracts   [QuodlibetContractCnt]bool `json:"uc"`
	CurrentPlayer   int                        `json:"cp"`
	LeadPlayer      int                        `json:"lp"`
	TrickNumber     int                        `json:"tn"`
	CurrentTrick    []*TrickCard               `json:"ct"`
	LastTrick       []*TrickCard               `json:"lt"`
	LastTrickWinner int                        `json:"lw"`
	TrickWinners    []int                      `json:"tw"`
	TrickRecord     [][]*TrickCard             `json:"tr"`
	TablePlaced     [5]uint16                  `json:"tp"`
	Stack           []*Card                    `json:"st"`
	PassCount       [QuodlibetPlayerCnt]int    `json:"pc"`
	OutCount        int                        `json:"oc"`
	GameEndFlag     bool                       `json:"ge"`
	LastDealDetail  *QuodlibetDealDetail       `json:"ld"`
	DealHistory     []*QuodlibetDealDetail     `json:"dh"`
	ActionLog       []*ActionLogEntry          `json:"al"`
}

// quodlibetMaxSliceLen は復元時に受け入れるスライスの上限。
const quodlibetMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。** KV に
// 保存した盤が空で返ってくるので、ここは必ず自前で書く。
func (q *Quodlibet) MarshalJSON() ([]byte, error) {
	return json.Marshal(quodlibetJSON{
		TrumpCards:      q.trumpCards,
		Players:         q.players,
		Config:          q.config,
		Phase:           q.phase,
		DealNumber:      q.dealNumber,
		DealerIdx:       q.dealerIdx,
		CurrentContract: q.currentContract,
		UsedContracts:   q.usedContracts,
		CurrentPlayer:   q.currentPlayer,
		LeadPlayer:      q.leadPlayer,
		TrickNumber:     q.trickNumber,
		CurrentTrick:    q.currentTrick,
		LastTrick:       q.lastTrick,
		LastTrickWinner: q.lastTrickWinner,
		TrickWinners:    q.trickWinners,
		TrickRecord:     q.trickRecord,
		TablePlaced:     q.tablePlaced,
		Stack:           q.stack,
		PassCount:       q.passCount,
		OutCount:        q.outCount,
		GameEndFlag:     q.gameEndFlag,
		LastDealDetail:  q.lastDealDetail,
		DealHistory:     q.dealHistory,
		ActionLog:       q.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (q *Quodlibet) UnmarshalJSON(data []byte) error {
	var j quodlibetJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > quodlibetMaxSliceLen || len(j.CurrentTrick) > quodlibetMaxSliceLen ||
		len(j.LastTrick) > quodlibetMaxSliceLen || len(j.TrickRecord) > quodlibetMaxSliceLen ||
		len(j.ActionLog) > quodlibetMaxSliceLen || len(j.DealHistory) > quodlibetMaxSliceLen {
		return fmt.Errorf("quodlibet: input array exceeds maximum allowed size")
	}
	if len(j.Players) != QuodlibetPlayerCnt {
		return fmt.Errorf("quodlibet: invalid player count %d, expected %d", len(j.Players), QuodlibetPlayerCnt)
	}
	q.trumpCards = j.TrumpCards
	if q.trumpCards == nil {
		q.trumpCards = NewTrumpCards32()
	}
	q.players = j.Players
	q.config = j.Config
	q.phase = j.Phase
	q.dealNumber = j.DealNumber
	q.dealerIdx = j.DealerIdx
	q.currentContract = j.CurrentContract
	q.usedContracts = j.UsedContracts
	q.currentPlayer = j.CurrentPlayer
	q.leadPlayer = j.LeadPlayer
	q.trickNumber = j.TrickNumber
	q.currentTrick = j.CurrentTrick
	q.lastTrick = j.LastTrick
	q.lastTrickWinner = j.LastTrickWinner
	q.trickWinners = j.TrickWinners
	q.trickRecord = j.TrickRecord
	q.tablePlaced = j.TablePlaced
	q.stack = j.Stack
	q.passCount = j.PassCount
	q.outCount = j.OutCount
	q.gameEndFlag = j.GameEndFlag
	q.lastDealDetail = j.LastDealDetail
	q.dealHistory = j.DealHistory
	if q.dealHistory == nil {
		q.dealHistory = make([]*QuodlibetDealDetail, 0, QuodlibetTotalDeals)
	}
	q.actionLog = j.ActionLog
	if q.actionLog == nil {
		q.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// quodlibetDealDetailJSON is the JSON wire format for QuodlibetDealDetail.
type quodlibetDealDetailJSON struct {
	Contract  int         `json:"c"`
	Round     int         `json:"r"`
	DealerIdx int         `json:"d"`
	Points    map[int]int `json:"p"`
}

// MarshalJSON implements json.Marshaler.
func (d *QuodlibetDealDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(quodlibetDealDetailJSON{
		Contract:  d.Contract,
		Round:     d.Round,
		DealerIdx: d.DealerIdx,
		Points:    d.Points,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *QuodlibetDealDetail) UnmarshalJSON(data []byte) error {
	var j quodlibetDealDetailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.Contract = j.Contract
	d.Round = j.Round
	d.DealerIdx = j.DealerIdx
	d.Points = j.Points
	if d.Points == nil {
		d.Points = map[int]int{}
	}
	return nil
}

// quodlibetRandIntn は rand.Intn の薄いラッパ (n <= 0 を握りつぶす)。
func quodlibetRandIntn(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}
