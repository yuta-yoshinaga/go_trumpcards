//go:build !js || !wasm || extra3

// Package domain ルー (Loo / Lanterloo) のドメインモデル。
//
// Loo は 18〜19 世紀イングランドで流行したポット系ギャンブル・トリックテイキング。
// 本実装は Five-card Loo (5 枚手札) の 4 人個人戦 (1 human + 3 CPU)。各ディールで全員が
// 一定額 (Ante) をポットへ支払い、山札のめくり札 (turn-up) のスートが切り札となる。
// 各プレイヤーは手札を見て「参加 (play)」か「降り (pass)」かを決める。参加者は 5 トリックを
// マストフォロー・マストヘッド (フォローできるなら勝てる札を出す義務、ボイドなら切り札を
// 出す義務) で戦う。1 トリックにつきポットの 1/5 を獲得し、参加したのに 1 トリックも
// 取れなかったプレイヤーは "looed" となりディール開始時のポット相当額を次ディールのポットへ
// 支払う (ペナルティ)。累積チップを積み上げるディール反復方式で、目標点レースはない。
//
// 本実装は extra ワーカーから到達可能なよう rank/trick/pot ロジックをすべてインラインで
// 持つ (Classic タグの Nap には依存しない)。extra 到達可能な NewTrumpCards(0) で 52 枚
// デッキを生成する。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// LooPlayerCnt はルーのプレイヤー数 (固定 4, 個人戦)。
const LooPlayerCnt = 4

// LooHandSize は各プレイヤーの手札枚数 (Five-card Loo)。
const LooHandSize = 5

// LooTrickCount は 1 ディールのトリック数。
const LooTrickCount = LooHandSize

// LooPhase はゲームフェーズ。
type LooPhase int

// Loo のフェーズ定数
const (
	// LooPhaseDecide 参加 (play/pass) 決定フェーズ
	LooPhaseDecide LooPhase = 0
	// LooPhasePlay トリックプレイフェーズ
	LooPhasePlay LooPhase = 1
	// LooPhaseTrickEnd トリック終了フェーズ
	LooPhaseTrickEnd LooPhase = 2
	// LooPhaseRoundEnd ディール終了フェーズ (ポット精算)
	LooPhaseRoundEnd LooPhase = 3
)

// LooDealDetail は 1 ディールの精算内訳。
type LooDealDetail struct {
	PotStart  int         // ディール開始時 (アンティ加算後) のポット額
	TrumpSuit int         // 切り札スート
	Playing   []bool      // 各プレイヤーが参加したか
	Tricks    map[int]int // プレイヤー別に獲得したトリック数
	Gained    map[int]int // プレイヤー別にこのディールで受け取ったチップ (獲得 - 支払い)
	Looed     []int       // looed (ペナルティを課された) プレイヤー
	PotCarry  int         // 次ディールへ繰り越すポット額 (ペナルティ計)
}

// LooHint はヒント情報。
type LooHint struct {
	CardIndices []int  // 推奨カードインデックス (play フェーズ)
	Decision    *bool  // 推奨 play(true)/pass(false) (decide フェーズ)
	Reason      string // ヒント理由キー
}

// Loo はルーゲームの状態を保持する集約ルート。
type Loo struct {
	trumpCards       *TrumpCards
	players          []*LooPlayer
	config           LooConfig
	phase            LooPhase
	roundNumber      int
	trickNumber      int
	dealerIdx        int
	currentPlayerIdx int
	decidePlayerIdx  int // 現在 play/pass を決めるプレイヤー
	decideDone       [LooPlayerCnt]bool
	currentTrick     []*TrickCard
	lastTrick        []*TrickCard
	lastTrickWinner  int
	leadPlayerIdx    int
	trumpSuit        int // 切り札スート (0=未確定)
	turnUp           *Card
	pot              int // 現在のポット額
	potStart         int // 現ディール開始時のポット額
	roundTricks      [LooPlayerCnt]int
	gameEndFlag      bool
	lastDealDetail   *LooDealDetail
	actionLogBase
	scored bool // 現ディールが精算済みか (二重精算防止)
}

// NewLoo はコンストラクタ。
func NewLoo(trumpCards *TrumpCards, players []*LooPlayer, config LooConfig) *Loo {
	return &Loo{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		lastTrickWinner: -1,
		dealerIdx:       LooPlayerCnt - 1,
		roundNumber:     0,
	}
}

// NewDefaultLoo は標準の 4 人構成 (1 human + 3 CPU) と DefaultLooConfig で Loo を生成する。
// CUI / Web / Worker の構築の単一情報源。
func NewDefaultLoo() *Loo {
	players := make([]*LooPlayer, LooPlayerCnt)
	players[0] = NewLooPlayer(true)
	for i := 1; i < LooPlayerCnt; i++ {
		players[i] = NewLooPlayer(false)
	}
	return NewLoo(newLooDeck(), players, DefaultLooConfig())
}

// newLooDeck はルー用 52 枚デッキを生成する。NewTrumpCards はビルドタグ無しの
// TrumpCards.go にあり extra ワーカーからも到達可能。
func newLooDeck() *TrumpCards {
	return NewTrumpCards(0)
}

// --- Loo 特有のランクヘルパー ---

// looRankValue はカードのランク値を返す (A 高: A=14 > K=13 > … > 2=2)。
func looRankValue(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。累計チップとポットもクリアする。
func (g *Loo) Reset() {
	g.gameEndFlag = false
	g.roundNumber = 1
	g.dealerIdx = LooPlayerCnt - 1
	g.pot = 0
	g.lastTrick = nil
	g.lastTrickWinner = -1
	g.lastDealDetail = nil
	g.actionLog = make([]*ActionLogEntry, 0)
	for _, p := range g.players {
		p.ResetDeal()
		p.ResetChips()
	}
	g.startDeal()
}

// NextRound は次のディールを開始する。
func (g *Loo) NextRound() {
	if g.phase != LooPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % LooPlayerCnt
	g.startDeal()
}

// startDeal はアンティを集め、手札を配り、切り札をめくって decide フェーズを開始する。
func (g *Loo) startDeal() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.lastTrick = nil
	g.lastTrickWinner = -1
	g.scored = false
	g.decideDone = [LooPlayerCnt]bool{}
	g.roundTricks = [LooPlayerCnt]int{}
	g.trumpSuit = 0
	g.turnUp = nil
	for _, p := range g.players {
		p.ResetDeal()
	}

	// アンティ: 各プレイヤーがポットへ支払う。
	ante := g.config.Ante
	for i := 0; i < LooPlayerCnt; i++ {
		g.players[i].AddChips(-ante)
		g.pot += ante
	}
	g.potStart = g.pot
	g.appendLog(-1, "ante", fmt.Sprintf("each antes %d; pot is %d", ante, g.pot), nil)

	g.trumpCards = newLooDeck()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	// 切り札をめくる (turn-up)。
	g.turnUp = g.trumpCards.DrawCard()
	if g.turnUp != nil {
		g.trumpSuit = g.turnUp.GetDesign()
		g.appendLog(-1, "trump_set", fmt.Sprintf("Trump is %s (turn-up %s)", suitName(g.trumpSuit), cardStr(g.turnUp)), []*Card{g.turnUp})
	}

	// decide 手番は forehand (dealer の次) から。
	g.decidePlayerIdx = (g.dealerIdx + 1) % LooPlayerCnt
	g.currentPlayerIdx = g.decidePlayerIdx
	g.phase = LooPhaseDecide
}

// deal は各プレイヤーへ LooHandSize 枚配る。
func (g *Loo) deal() {
	for i := 0; i < LooHandSize; i++ {
		for j := 0; j < LooPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % LooPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// --- Decide (play/pass) ---

// PlayerDecide は人間プレイヤーが参加 (play) / 降り (pass) を決める。
func (g *Loo) PlayerDecide(play bool) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != LooPhaseDecide {
		return ErrWrongPhase
	}
	if !g.players[g.decidePlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.applyDecide(g.decidePlayerIdx, play)
	return nil
}

// CpuDecide は decide フェーズで CPU が 1 件決定する。
func (g *Loo) CpuDecide() {
	if g.gameEndFlag || g.phase != LooPhaseDecide {
		return
	}
	idx := g.decidePlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	g.applyDecide(idx, g.cpuChooseDecide(idx))
}

// applyDecide は decide を記録し、次の決定者へ進める。全員決めたらプレイへ移る。
func (g *Loo) applyDecide(idx int, play bool) {
	g.players[idx].SetPlaying(play)
	g.decideDone[idx] = true
	if play {
		g.appendLog(idx, "decide", fmt.Sprintf("%s plays", playerName(g.players, idx)), nil)
	} else {
		g.appendLog(idx, "decide", fmt.Sprintf("%s passes", playerName(g.players, idx)), nil)
	}
	for k := 1; k <= LooPlayerCnt; k++ {
		ni := (idx + k) % LooPlayerCnt
		if !g.decideDone[ni] {
			g.decidePlayerIdx = ni
			g.currentPlayerIdx = ni
			return
		}
	}
	g.resolveDecide()
}

// resolveDecide は decide を締め、参加人数に応じてプレイ開始 or 即精算へ移る。
func (g *Loo) resolveDecide() {
	active := g.activePlayers()
	switch len(active) {
	case 0:
		// 誰も参加しなかった: ポットは繰り越し、このディールは精算 (looed なし)。
		g.appendLog(-1, "all_pass", "all players passed; pot carries over", nil)
		g.enterRoundEnd()
	case 1:
		// 1 人だけ参加: プレイせずにポットを総取り (トリックは戦われない)。
		winner := active[0]
		g.appendLog(winner, "walkover", fmt.Sprintf("%s is the only player and takes the pot", playerName(g.players, winner)), nil)
		g.enterRoundEnd()
	default:
		g.startPlayPhase(active[0])
	}
}

// enterRoundEnd はディール終了フェーズへ移行し、直ちにポットを精算する。
// これによりフロントエンドが「次のディール」操作を待つ間にディール結果
// (獲得チップ・looed・累計チップ) を表示できる。ScoreRound は scored フラグで
// 二重精算を防ぐため、後続の NextRound から再度呼ばれても副作用はない。
func (g *Loo) enterRoundEnd() {
	g.phase = LooPhaseRoundEnd
	g.ScoreRound()
}

// startPlayPhase はプレイフェーズを開始する: forehand の参加者がリードする。
func (g *Loo) startPlayPhase(firstLeader int) {
	g.leadPlayerIdx = firstLeader
	g.currentPlayerIdx = firstLeader
	g.trickNumber = 1
	g.currentTrick = nil
	g.phase = LooPhasePlay
}

// activePlayers は参加 (playing) しているプレイヤーのインデックスを forehand 順で返す。
func (g *Loo) activePlayers() []int {
	var out []int
	for k := 0; k < LooPlayerCnt; k++ {
		idx := (g.dealerIdx + 1 + k) % LooPlayerCnt
		if g.players[idx].GetPlaying() {
			out = append(out, idx)
		}
	}
	return out
}

// nextActive は idx の次に参加しているプレイヤーを返す (循環)。見つからなければ -1。
func (g *Loo) nextActive(idx int) int {
	for k := 1; k <= LooPlayerCnt; k++ {
		ni := (idx + k) % LooPlayerCnt
		if g.players[ni].GetPlaying() {
			return ni
		}
	}
	return -1
}

// --- Play ---

// PlayerPlay は人間プレイヤーがカードを出す。
func (g *Loo) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != LooPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if err := g.validatePlay(g.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// CpuPlay は現在の手番が CPU の場合に 1 ステップ実行する。decide / play の両フェーズを
// 進める。
func (g *Loo) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	switch g.phase {
	case LooPhaseDecide:
		g.CpuDecide()
	case LooPhasePlay:
		if g.players[g.currentPlayerIdx].GetIsHuman() {
			return
		}
		cardIdx := g.cpuSelectPlayCard(g.currentPlayerIdx)
		if cardIdx < 0 {
			return // 出せる札がない (状態機械上は到達しないが防御的に no-op)
		}
		played := g.players[g.currentPlayerIdx].RemoveCard(cardIdx)
		// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
		// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
		// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
		if played == nil {
			return
		}
		g.playCard(g.currentPlayerIdx, played)
	}
}

// playCard はカードをプレイする共通処理。
func (g *Loo) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == len(g.activePlayers()) {
		g.phase = LooPhaseTrickEnd
	} else {
		g.currentPlayerIdx = g.nextActive(g.currentPlayerIdx)
	}
}

// ResolveTrick はトリックを解決して勝者を決定する。
func (g *Loo) ResolveTrick() {
	if g.phase != LooPhaseTrickEnd || len(g.currentTrick) != len(g.activePlayers()) {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.roundTricks[winnerIdx]++
	g.lastTrick = g.currentTrick
	g.lastTrickWinner = winnerIdx
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= LooTrickCount {
		g.enterRoundEnd()
	} else {
		g.phase = LooPhaseTrickEnd
	}
}

// NextTrick は次のトリックを開始する。
func (g *Loo) NextTrick() {
	if g.phase != LooPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = LooPhasePlay
}

// trickWinner はトリックの勝者を決定する (切り札最高 > リード最高)。
func (g *Loo) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return g.leadPlayerIdx
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.looRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() != g.trumpSuit && tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.looRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// looRank はトリック比較用のランクを返す。切り札は非切り札より常に強い。A 高。
func (g *Loo) looRank(card *Card) int {
	base := looRankValue(card.GetValue())
	if g.trumpSuit != 0 && card.GetDesign() == g.trumpSuit {
		return 100 + base
	}
	return base
}

// --- Settle (pot) ---

// ScoreRound はディールのポットを精算する。参加者は取得トリックに応じてポットを分配し、
// 参加したのに 0 トリックのプレイヤーは looed としてペナルティを次ポットへ支払う。
func (g *Loo) ScoreRound() {
	if g.phase != LooPhaseRoundEnd || g.scored {
		return
	}
	g.scored = true
	active := g.activePlayers()
	gained := make(map[int]int, LooPlayerCnt)
	tricks := make(map[int]int, LooPlayerCnt)
	playing := make([]bool, LooPlayerCnt)
	for i := 0; i < LooPlayerCnt; i++ {
		gained[i] = 0
		tricks[i] = g.roundTricks[i]
		playing[i] = g.players[i].GetPlaying()
	}

	looed := make([]int, 0)
	potStart := g.potStart

	switch len(active) {
	case 0:
		// 誰も参加せず: ポットは繰り越し (変更なし)。
		g.appendLog(-1, "settle", fmt.Sprintf("no players; pot %d carries over", g.pot), nil)
	case 1:
		// 1 人だけ参加: プレイせずにポット総取り。
		winner := active[0]
		g.players[winner].AddChips(g.pot)
		gained[winner] += g.pot
		g.appendLog(winner, "settle", fmt.Sprintf("%s takes the whole pot %d", playerName(g.players, winner), g.pot), nil)
		g.pot = 0
	default:
		// 参加者でトリックに応じてポットを分配 (1 トリック = ポット/トリック数)。
		share := LooPerTrickShare(potStart)
		distributed := 0
		for _, idx := range active {
			win := g.roundTricks[idx] * share
			if win > 0 {
				g.players[idx].AddChips(win)
				gained[idx] += win
				distributed += win
			}
		}
		// 端数はそのままポットに残す (次ディールへ)。
		g.pot -= distributed

		// looed: 参加したのに 0 トリック → ペナルティを次ポットへ。
		penalty := potStart
		for _, idx := range active {
			if g.roundTricks[idx] == 0 {
				g.players[idx].AddChips(-penalty)
				gained[idx] -= penalty
				g.pot += penalty
				looed = append(looed, idx)
				g.appendLog(idx, "looed",
					fmt.Sprintf("%s is looed and pays %d to the pot", playerName(g.players, idx), penalty), nil)
			}
		}
	}

	g.lastDealDetail = &LooDealDetail{
		PotStart:  potStart,
		TrumpSuit: g.trumpSuit,
		Playing:   playing,
		Tricks:    tricks,
		Gained:    gained,
		Looed:     looed,
		PotCarry:  g.pot,
	}
	for i := 0; i < LooPlayerCnt; i++ {
		g.appendLog(i, "cumulative_chips",
			fmt.Sprintf("%s: total=%d", playerName(g.players, i), g.players[i].GetChips()), nil)
	}
}

// --- 検証 / プレイ可能判定 ---

// validatePlay はマストフォロー・マストヘッドを検証する。
// フォローできるなら (1) リードスートに従う。かつ (2) 現在勝っている札を上回れるなら
// 上回る札を出す (マストヘッド)。フォローできず切り札を持つならボイド時は切り札を出す。
func (g *Loo) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil // リードは任意
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	hasLead := g.playerHasSuit(playerIdx, leadSuit)
	hasTrump := g.playerHasSuit(playerIdx, g.trumpSuit)

	if hasLead {
		if card.GetDesign() != leadSuit {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
		// マストヘッド: 現在勝っている札を上回れるなら上回る義務がある。
		if g.canBeatWithSuit(playerIdx, leadSuit) && !g.beatsCurrentBest(card) {
			return NewDomainError(ErrInvalidPlay, "勝てる札を出す必要があります (マストヘッド)")
		}
		return nil
	}
	// リードスートなし。切り札を持つなら切り札を出さなければならない。
	if hasTrump {
		if card.GetDesign() != g.trumpSuit {
			return NewDomainError(ErrInvalidPlay, "切り札を出してください")
		}
		// マストヘッド: 切り札で現在の勝者を上回れるなら上回る義務がある。
		if g.canBeatWithSuit(playerIdx, g.trumpSuit) && !g.beatsCurrentBest(card) {
			return NewDomainError(ErrInvalidPlay, "勝てる切り札を出す必要があります (マストヘッド)")
		}
		return nil
	}
	// リードスートも切り札もない: 任意のカードを捨てられる。
	return nil
}

// beatsCurrentBest は card が現在のトリック勝者の札を上回るかを返す。
func (g *Loo) beatsCurrentBest(card *Card) bool {
	winnerIdx := g.trickWinner()
	return g.looRank(card) > g.trickTopRank(winnerIdx)
}

// canBeatWithSuit は playerIdx が指定スートの手札で現在の勝者を上回れるかを返す。
func (g *Loo) canBeatWithSuit(playerIdx, suit int) bool {
	winnerIdx := g.trickWinner()
	top := g.trickTopRank(winnerIdx)
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == suit && g.looRank(c) > top {
			return true
		}
	}
	return false
}

// playerHasSuit はプレイヤーが指定スートのカードを持っているか (suit=0 なら常に false)。
func (g *Loo) playerHasSuit(playerIdx, suit int) bool {
	if suit == 0 {
		return false
	}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == suit {
			return true
		}
	}
	return false
}

// getValidPlayIndices はプレイ可能なカードのインデックスリストを返す。
func (g *Loo) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// --- ヘルパー ---

func (g *Loo) sortAllHands() {
	for _, p := range g.players {
		looSortHand(p)
	}
}

// looSortHand はプレイヤーの手札をスート→ランクの順にソートする (表示用)。
func looSortHand(p *LooPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return looRankValue(cards[i].GetValue()) > looRankValue(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// indexOfPlayerInTrick は currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Loo) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// trickTopRank は現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *Loo) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.looRank(g.currentTrick[idx].Card)
}

// --- CPU AI ---

// looHandStrength はディール参加判断のための手札強度を見積もる (高いほど強い)。
// 切り札の枚数と高位札 (A/K/Q + 切り札) を重視する。
func (g *Loo) looHandStrength(playerIdx int) int {
	p := g.players[playerIdx]
	strength := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == g.trumpSuit {
			strength += 3 + looRankValue(c.GetValue())/4 // 切り札は価値大
		} else if c.GetValue() == 1 { // A
			strength += 3
		} else if c.GetValue() >= 12 { // K, Q
			strength += 1
		}
	}
	return strength
}

// cpuChooseDecide は CPU の play/pass を選ぶ。
func (g *Loo) cpuChooseDecide(playerIdx int) bool {
	if g.config.CpuDifficulty == LooCpuDifficultyEasy {
		return rand.Intn(2) == 0
	}
	strength := g.looHandStrength(playerIdx)
	threshold := 6
	if g.config.CpuDifficulty == LooCpuDifficultyHard {
		threshold = 8 // より厳しい参加基準
	}
	return strength >= threshold
}

// cpuSelectPlayCard は CPU がプレイするカードのインデックスを選ぶ。
func (g *Loo) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return -1 // 出せる札がない (呼び出し側で no-op)
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == LooCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart はトリック状況を意識した戦略プレイ。トリックを取りたい (looed 回避)。
func (g *Loo) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		// リード: 高い札で主導権を握る。
		return pickHighest(player, valid, func(c *Card) int { return g.looRank(c) })
	}
	winnerIdx := g.trickWinner()
	top := g.trickTopRank(winnerIdx)
	winners := looFilter(valid, func(idx int) bool { return g.looRank(player.GetCard(idx)) > top })
	if len(winners) > 0 {
		// 勝てるなら最小の勝ち札で勝つ。
		return pickLowest(player, winners, func(c *Card) int { return g.looRank(c) })
	}
	// 勝てない: 最小の札を捨てる。
	return pickLowest(player, valid, func(c *Card) int { return g.looRank(c) })
}

func looFilter(indices []int, pred func(int) bool) []int {
	var out []int
	for _, idx := range indices {
		if pred(idx) {
			out = append(out, idx)
		}
	}
	return out
}

// --- Hint ---

// GetHint は人間プレイヤーの手番における推奨を返す (decide / play フェーズ)。
func (g *Loo) GetHint() *LooHint {
	human := findHumanIdx(g.players)
	if human < 0 {
		return nil
	}
	switch g.phase {
	case LooPhaseDecide:
		if g.decidePlayerIdx != human {
			return nil
		}
		strength := g.looHandStrength(human)
		decision := strength >= 6
		reason := "decide_pass"
		if decision {
			reason = "decide_play"
		}
		d := decision
		return &LooHint{Decision: &d, Reason: reason}
	case LooPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &LooHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	default:
		return nil
	}
}

// playHintReason はプレイヒントの理由キーを判定する。
func (g *Loo) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_high"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	if g.beatsCurrentBest(card) {
		return "follow_win"
	}
	return "discard_low"
}

// --- 状態アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *Loo) GetPhase() LooPhase { return g.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *Loo) SetPhase(phase LooPhase) { g.phase = phase }

// GetRoundNumber はディール番号を返す。
func (g *Loo) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber はディール番号を設定する (テスト用)。
func (g *Loo) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber はトリック番号を返す。
func (g *Loo) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber はトリック番号を設定する (テスト用)。
func (g *Loo) SetTrickNumber(n int) { g.trickNumber = n }

// GetDealerIdx は親インデックスを返す。
func (g *Loo) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx は親インデックスを設定する (テスト用)。
func (g *Loo) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetCurrentTurn は現在の手番プレイヤーインデックスを返す。
func (g *Loo) GetCurrentTurn() int { return g.currentPlayerIdx }

// SetCurrentTurn は現在の手番を設定する (テスト用)。
func (g *Loo) SetCurrentTurn(idx int) { g.currentPlayerIdx = idx }

// GetDecidePlayerIdx は現在 play/pass を決めるプレイヤーインデックスを返す。
func (g *Loo) GetDecidePlayerIdx() int { return g.decidePlayerIdx }

// SetDecidePlayerIdx は decide 手番を設定する (テスト用)。
func (g *Loo) SetDecidePlayerIdx(idx int) { g.decidePlayerIdx = idx }

// GetCurrentTrick は進行中のトリックを返す。
func (g *Loo) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick は進行中のトリックを設定する (テスト用)。
func (g *Loo) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLastTrick は直前に完了したトリックを返す。
func (g *Loo) GetLastTrick() []*TrickCard { return g.lastTrick }

// GetLastTrickWinner は直前トリックの勝者を返す (-1=なし)。
func (g *Loo) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetLeadPlayerIdx はリードプレイヤーインデックスを返す。
func (g *Loo) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx はリードプレイヤーインデックスを設定する (テスト用)。
func (g *Loo) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetTrumpSuit は切り札スートを返す (0=未確定)。
func (g *Loo) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit は切り札スートを設定する (テスト用)。
func (g *Loo) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetTurnUp はめくり札 (切り札決定札) を返す (nil の場合もある)。
func (g *Loo) GetTurnUp() *Card { return g.turnUp }

// GetPot は現在のポット額を返す。
func (g *Loo) GetPot() int { return g.pot }

// SetPot はポット額を設定する (テスト用)。
func (g *Loo) SetPot(v int) { g.pot = v }

// GetPotStart は現ディール開始時のポット額を返す。
func (g *Loo) GetPotStart() int { return g.potStart }

// LooPerTrickShare は 1 トリックあたりの取り分を返す。
//
// **端数はポットに残る。**5 で割り切れないディールでは、全トリック取っても
// ポット全額は入らない (37 なら 7×5 = 35)。表示側もこの関数を通すこと —
// 別に割ると、表示より少ない額しか入らない案内になる (#4921)。
func LooPerTrickShare(potStart int) int {
	if potStart <= 0 {
		return 0
	}
	return potStart / LooTrickCount
}

// LooMaxWin は全トリック取ったときに実際に入る額を返す。
func LooMaxWin(potStart int) int { return LooPerTrickShare(potStart) * LooTrickCount }

// SetPotStart は現ディール開始時のポット額を設定する (テスト用)。
func (g *Loo) SetPotStart(v int) { g.potStart = v }

// GetRoundTricks は現ディールの獲得トリック数を返す。
func (g *Loo) GetRoundTricks() [LooPlayerCnt]int { return g.roundTricks }

// SetRoundTricks は現ディールの獲得トリック数を設定する (テスト用)。
func (g *Loo) SetRoundTricks(s [LooPlayerCnt]int) { g.roundTricks = s }

// GetGameEndFlag はゲーム終了フラグを返す (Loo は常に false; ディール反復方式)。
func (g *Loo) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayerCnt はプレイヤー数を返す。
func (g *Loo) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Loo) GetPlayer(i int) *LooPlayer {
	return getPlayer(g.players, i)
}

// GetLastDealDetail は直前ディールの精算内訳を返す (nil の場合もある)。
func (g *Loo) GetLastDealDetail() *LooDealDetail { return g.lastDealDetail }

// IsHumanTurn は現在の意思決定者が人間かを返す。
func (g *Loo) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	switch g.phase {
	case LooPhaseDecide:
		return g.decidePlayerIdx >= 0 && g.decidePlayerIdx < len(g.players) && g.players[g.decidePlayerIdx].GetIsHuman()
	case LooPhasePlay:
		return g.currentPlayerIdx >= 0 && g.currentPlayerIdx < len(g.players) && g.players[g.currentPlayerIdx].GetIsHuman()
	default:
		return false
	}
}

// GetConfig はローカルルール設定を返す。
func (g *Loo) GetConfig() LooConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *Loo) SetConfig(cfg LooConfig) { g.config = cfg }

// GetPlayableIndices はプレイフェーズでプレイ可能な手札インデックスを返す。
func (g *Loo) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != LooPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON Serialization ---

// looJSON is the JSON wire format for Loo.
type looJSON struct {
	TrumpCards       *TrumpCards        `json:"tc"`
	Players          []*LooPlayer       `json:"ps"`
	Config           LooConfig          `json:"cf"`
	Phase            LooPhase           `json:"ph"`
	RoundNumber      int                `json:"rn"`
	TrickNumber      int                `json:"tn"`
	DealerIdx        int                `json:"di"`
	CurrentPlayerIdx int                `json:"ci"`
	DecidePlayerIdx  int                `json:"dp"`
	DecideDone       [LooPlayerCnt]bool `json:"dd"`
	CurrentTrick     []*TrickCard       `json:"ct"`
	LastTrick        []*TrickCard       `json:"lt"`
	LastTrickWinner  int                `json:"lw"`
	LeadPlayerIdx    int                `json:"li"`
	TrumpSuit        int                `json:"ts"`
	TurnUp           *Card              `json:"tu"`
	Pot              int                `json:"po"`
	PotStart         int                `json:"pst"`
	RoundTricks      [LooPlayerCnt]int  `json:"rt"`
	GameEndFlag      bool               `json:"ge"`
	LastDealDetail   *LooDealDetail     `json:"ld"`
	ActionLog        []*ActionLogEntry  `json:"al"`
	Scored           bool               `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (g *Loo) MarshalJSON() ([]byte, error) {
	return json.Marshal(looJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		DealerIdx:        g.dealerIdx,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DecidePlayerIdx:  g.decidePlayerIdx,
		DecideDone:       g.decideDone,
		CurrentTrick:     g.currentTrick,
		LastTrick:        g.lastTrick,
		LastTrickWinner:  g.lastTrickWinner,
		LeadPlayerIdx:    g.leadPlayerIdx,
		TrumpSuit:        g.trumpSuit,
		TurnUp:           g.turnUp,
		Pot:              g.pot,
		PotStart:         g.potStart,
		RoundTricks:      g.roundTricks,
		GameEndFlag:      g.gameEndFlag,
		LastDealDetail:   g.lastDealDetail,
		ActionLog:        g.actionLog,
		Scored:           g.scored,
	})
}

// looMaxSliceLen caps slice sizes during deserialisation to prevent excessive
// memory allocation from malformed input.
const looMaxSliceLen = 1000

// looInRange reports whether v is in [0, LooPlayerCnt).
func looInRange(v int) bool { return v >= 0 && v < LooPlayerCnt }

// looInRangeOrUnset reports whether v is -1 (unset) or in [0, LooPlayerCnt).
func looInRangeOrUnset(v int) bool { return v == -1 || looInRange(v) }

// looValidateTrick は復元したトリック配列の各要素を検証する。
func looValidateTrick(trick []*TrickCard) error {
	for _, tc := range trick {
		if tc == nil || tc.Card == nil {
			return fmt.Errorf("loo: invalid trick card")
		}
		if !looInRange(tc.PlayerIdx) {
			return fmt.Errorf("loo: trick card player index out of range")
		}
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Loo) UnmarshalJSON(data []byte) error {
	var j looJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > looMaxSliceLen || len(j.CurrentTrick) > looMaxSliceLen ||
		len(j.LastTrick) > looMaxSliceLen || len(j.ActionLog) > looMaxSliceLen {
		return fmt.Errorf("loo: input array exceeds maximum allowed size")
	}
	if len(j.Players) != LooPlayerCnt {
		return fmt.Errorf("loo: invalid player count %d, expected %d", len(j.Players), LooPlayerCnt)
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("loo: nil player in state")
		}
	}
	if err := looValidateTrick(j.CurrentTrick); err != nil {
		return err
	}
	if err := looValidateTrick(j.LastTrick); err != nil {
		return err
	}
	switch j.Phase {
	case LooPhaseDecide, LooPhasePlay, LooPhaseTrickEnd, LooPhaseRoundEnd:
	default:
		return fmt.Errorf("loo: invalid phase %d", j.Phase)
	}
	if j.Pot < 0 || j.PotStart < 0 {
		return fmt.Errorf("loo: pot must be non-negative")
	}
	// 切り札スート: 0(未確定) 許容、それ以外は [Spade, Diamond]。
	if j.TrumpSuit != 0 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("loo: invalid trump suit %d", j.TrumpSuit)
	}
	if j.RoundNumber < 1 || j.TrickNumber < 1 || j.TrickNumber > LooTrickCount {
		return fmt.Errorf("loo: invalid round/trick number")
	}
	if !looInRange(j.DealerIdx) || !looInRange(j.DecidePlayerIdx) {
		return fmt.Errorf("loo: index field out of range")
	}
	if !looInRangeOrUnset(j.CurrentPlayerIdx) || !looInRangeOrUnset(j.LeadPlayerIdx) ||
		!looInRangeOrUnset(j.LastTrickWinner) {
		return fmt.Errorf("loo: sentinel index field out of range")
	}
	// フェーズが play 以降では切り札とリードが確定していなければならない。
	if j.Phase == LooPhasePlay || j.Phase == LooPhaseTrickEnd {
		if j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond {
			return fmt.Errorf("loo: trump must be set once play begins")
		}
		if !looInRange(j.CurrentPlayerIdx) || !looInRange(j.LeadPlayerIdx) {
			return fmt.Errorf("loo: play indices must be set once play begins")
		}
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newLooDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.dealerIdx = j.DealerIdx
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.decidePlayerIdx = j.DecidePlayerIdx
	g.decideDone = j.DecideDone
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.lastTrick = j.LastTrick
	g.lastTrickWinner = j.LastTrickWinner
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.trumpSuit = j.TrumpSuit
	g.turnUp = j.TurnUp
	g.pot = j.Pot
	g.potStart = j.PotStart
	g.roundTricks = j.RoundTricks
	g.gameEndFlag = j.GameEndFlag
	g.scored = j.Scored
	g.lastDealDetail = j.LastDealDetail
	if j.ActionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	} else {
		g.actionLog = j.ActionLog
	}
	return nil
}

// looDealDetailJSON is the JSON wire format for LooDealDetail.
type looDealDetailJSON struct {
	PotStart  int         `json:"ps"`
	TrumpSuit int         `json:"ts"`
	Playing   []bool      `json:"pl"`
	Tricks    map[int]int `json:"tr"`
	Gained    map[int]int `json:"gn"`
	Looed     []int       `json:"lo"`
	PotCarry  int         `json:"pc"`
}

// MarshalJSON implements json.Marshaler.
func (d *LooDealDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(looDealDetailJSON{
		PotStart:  d.PotStart,
		TrumpSuit: d.TrumpSuit,
		Playing:   d.Playing,
		Tricks:    d.Tricks,
		Gained:    d.Gained,
		Looed:     d.Looed,
		PotCarry:  d.PotCarry,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *LooDealDetail) UnmarshalJSON(data []byte) error {
	var j looDealDetailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.PotStart = j.PotStart
	d.TrumpSuit = j.TrumpSuit
	d.Playing = j.Playing
	d.Tricks = j.Tricks
	d.Gained = j.Gained
	d.Looed = j.Looed
	d.PotCarry = j.PotCarry
	return nil
}
