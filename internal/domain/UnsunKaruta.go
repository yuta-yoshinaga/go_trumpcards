//go:build !js || !wasm || classic

package domain

// うんすんカルタ「八人メリ」 — 熊本県人吉に伝わる 75 枚 5 スートの
// トリックテイキング。
//
//   - デッキ: 5 スート × 15 枚 = 75 枚。スートは パオ (棍棒) / イス (刀剣) /
//     コツ (聖杯) / オウル (金貨) / クル (巴紋)。位は 数札 1〜9 と、
//     ソウタ・ウマ・キリ・ウン・スン・ロバイ の 6 枚。
//   - 配札: 8 人が 2 組に分かれ、**敵味方が交互に座る**。3 枚ずつ 3 回配って
//     1 人 9 枚 (72 枚)、残る 3 枚は伏せて中央へ。**その 1 枚目を表に返し、
//     そのスートがそのディールの切り札**になる。
//   - 強さ: ウン > スン > ソウタ > ロバイ > キリ > ウマ > 数札。数札の並びは
//     **スートの種類で逆になる** ── 長物 (パオ・イス) は 9 が最強で 1 が最弱、
//     丸物 (コツ・オウル・クル) は **1 が最強で 9 が最弱**。うんすんカルタが
//     残した最も古い特徴で、これを取り違えると丸物のトリックが全部ひっくり返る。
//   - メリ / モンチ: **フォロー義務は宣言で生まれる。** 切り札でリードする人は
//     「メリ」、平札でリードする人は「モンチ」を宣言でき、宣言されたトリックは
//     全員が台札のスートに従う。宣言が無ければ好きな札を出してよい。
//   - トリック: 8 枚出揃った時点で、切り札が出ていればその最強、無ければ台札
//     スートの最強を出した席のチームが「コ」を 1 つ取る。
//   - 得点: ディールごとにチームの「コ」を数え、TargetDeals ディールの累計が
//     多いチームの勝ち。**ヤク (返した切り札による 2 コ / 5 コ の移動) は
//     実装していない** ── 宣言と席順の組み合わせで決まる別系統の規則で、
//     ここでは 1 トリック = 1 コ に揃えている。

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// デッキの形。
const (
	// UnsunKarutaDeckSize はデッキの総枚数。
	UnsunKarutaDeckSize = 75
	// UnsunKarutaSuitCnt はスート数。
	UnsunKarutaSuitCnt = 5
	// UnsunKarutaRankCnt は 1 スートの枚数。
	UnsunKarutaRankCnt = 15
	// UnsunKarutaPlayerCnt は席数 (8 人)。
	UnsunKarutaPlayerCnt = 8
	// UnsunKarutaTeamCnt はチーム数。
	UnsunKarutaTeamCnt = 2
	// UnsunKarutaHandSize は 1 人の手札枚数。
	UnsunKarutaHandSize = 9
	// UnsunKarutaTalonSize は中央に伏せる枚数。1 枚目を返して切り札を決める。
	UnsunKarutaTalonSize = 3
	// UnsunKarutaTrickCount は 1 ディールのトリック数。
	UnsunKarutaTrickCount = UnsunKarutaHandSize
)

// スート番号。**1〜2 が長物、3〜5 が丸物。** 数札の並びがここで反転する。
const (
	// UnsunKarutaSuitPao はパオ (棍棒)。長物。
	UnsunKarutaSuitPao = 1
	// UnsunKarutaSuitIsu はイス (刀剣)。長物。
	UnsunKarutaSuitIsu = 2
	// UnsunKarutaSuitKotsu はコツ (聖杯)。丸物。
	UnsunKarutaSuitKotsu = 3
	// UnsunKarutaSuitOru はオウル (金貨)。丸物。
	UnsunKarutaSuitOru = 4
	// UnsunKarutaSuitKuru はクル (巴紋)。丸物で、第 5 のスート。
	UnsunKarutaSuitKuru = 5
)

// 位の値。1〜9 は数札そのもの、10 以降が絵札。
const (
	// UnsunKarutaSota はソウタ (女官)。
	UnsunKarutaSota = 10
	// UnsunKarutaUma はウマ (騎馬)。
	UnsunKarutaUma = 11
	// UnsunKarutaKiri はキリ (王)。
	UnsunKarutaKiri = 12
	// UnsunKarutaUn はウン (福神)。最強。
	UnsunKarutaUn = 13
	// UnsunKarutaSun はスン (唐人)。
	UnsunKarutaSun = 14
	// UnsunKarutaRobai はロバイ (竜)。
	UnsunKarutaRobai = 15
)

// UnsunKarutaPhase はゲームの進行段階。
type UnsunKarutaPhase int

const (
	// UnsunKarutaPhasePlay はトリックプレイ中。
	UnsunKarutaPhasePlay UnsunKarutaPhase = iota
	// UnsunKarutaPhaseTrickEnd はトリックが揃って解決待ち。
	UnsunKarutaPhaseTrickEnd
	// UnsunKarutaPhaseRoundEnd はディールが終わって集計済み。
	UnsunKarutaPhaseRoundEnd
	// UnsunKarutaPhaseGameEnd はマッチ終了。
	UnsunKarutaPhaseGameEnd
)

// UnsunKarutaPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const UnsunKarutaPhaseMax = int(UnsunKarutaPhaseGameEnd)

// UnsunKarutaResult は人間チームから見たマッチ結果。
type UnsunKarutaResult int

const (
	// UnsunKarutaResultNone は未確定または引き分け。
	UnsunKarutaResultNone UnsunKarutaResult = iota
	// UnsunKarutaResultWin は人間のチームの勝ち。
	UnsunKarutaResultWin
	// UnsunKarutaResultLose は人間のチームの負け。
	UnsunKarutaResultLose
)

// unsunKarutaMaxSliceLen は復元時に許すスライス長の上限。
const unsunKarutaMaxSliceLen = 256

// エラー値 (`errors.Is` 用の番兵)。文言は i18n 側に持たせる。
var (
	errUnsunKarutaNotLeader = errors.New("unsunkaruta: only the leader may declare")
)

// UnsunKarutaHint はヒント (推奨札とその理由キー)。
type UnsunKarutaHint struct {
	CardIndices []int
	Reason      string
}

// UnsunKaruta はうんすんカルタ (八人メリ) の卓。
type UnsunKaruta struct {
	deck        []*Card
	deckDrawCnt int
	players     []*UnsunKarutaPlayer
	config      UnsunKarutaConfig

	phase            UnsunKarutaPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int

	// talon は中央に伏せた 3 枚。先頭が表を向いて切り札を決める。
	talon []*Card
	// trumpSuit はこのディールの切り札スート。
	trumpSuit int
	// mustFollow はこのトリックでフォロー義務があるか (メリ / モンチ の宣言)。
	mustFollow bool
	// declaredThisTrick は宣言が行われたか (画面表示用)。
	declaredThisTrick bool

	// teamTricks はディール中のチーム別「コ」。
	teamTricks [UnsunKarutaTeamCnt]int
	// teamScores はマッチ累計。
	teamScores      [UnsunKarutaTeamCnt]int
	lastTrickWinner int
	result          UnsunKarutaResult
	scored          bool
	gameEndFlag     bool
	winnerTeam      int
	actionLogBase
}

// NewUnsunKaruta コンストラクタ。
func NewUnsunKaruta(players []*UnsunKarutaPlayer, config UnsunKarutaConfig) *UnsunKaruta {
	return &UnsunKaruta{
		players:         players,
		config:          config,
		winnerTeam:      -1,
		lastTrickWinner: -1,
	}
}

// NewDefaultUnsunKaruta は既定の 8 人卓 (人間 1 + CPU 7) を返す。
func NewDefaultUnsunKaruta() *UnsunKaruta {
	return NewUnsunKaruta(newUnsunKarutaPlayers(), DefaultUnsunKarutaConfig())
}

// newUnsunKarutaPlayers は 8 席を作る (席 0 が人間)。
func newUnsunKarutaPlayers() []*UnsunKarutaPlayer {
	players := make([]*UnsunKarutaPlayer, 0, UnsunKarutaPlayerCnt)
	players = append(players, NewUnsunKarutaPlayer(true))
	for i := 1; i < UnsunKarutaPlayerCnt; i++ {
		players = append(players, NewUnsunKarutaPlayer(false))
	}
	return players
}

// UnsunKarutaTeamOf は席のチームを返す。
//
// **敵味方が交互に座る。** 偶数席が人間のチーム、奇数席が相手 ── 隣り合う席は
// 必ず敵同士になる。
func UnsunKarutaTeamOf(playerIdx int) int { return playerIdx % UnsunKarutaTeamCnt }

// buildUnsunKarutaDeck は 75 枚のデッキを組む。
func buildUnsunKarutaDeck() []*Card {
	deck := make([]*Card, 0, UnsunKarutaDeckSize)
	for suit := UnsunKarutaSuitPao; suit <= UnsunKarutaSuitKuru; suit++ {
		for val := 1; val <= UnsunKarutaRankCnt; val++ {
			deck = append(deck, NewCard(suit, val, false))
		}
	}
	return deck
}

// UnsunKarutaIsRoundSuit は丸物 (コツ・オウル・クル) かを返す。
//
// **数札の並びがここで決まる。** 丸物は 1 が最強で 9 が最弱、長物 (パオ・イス)
// はその逆。
func UnsunKarutaIsRoundSuit(suit int) bool {
	return suit == UnsunKarutaSuitKotsu || suit == UnsunKarutaSuitOru || suit == UnsunKarutaSuitKuru
}

// unsunKarutaMaxPip は数札の最大値 (9)。丸物の反転はここを軸にする。
const unsunKarutaMaxPip = 9

// unsunKarutaStrength は札の強さを返す (大きいほど強い)。同じスート同士の
// 比較にだけ意味がある。
func unsunKarutaStrength(c *Card) int {
	if c == nil {
		return -1
	}
	switch c.GetValue() {
	case UnsunKarutaUn:
		return 14
	case UnsunKarutaSun:
		return 13
	case UnsunKarutaSota:
		return 12
	case UnsunKarutaRobai:
		return 11
	case UnsunKarutaKiri:
		return 10
	case UnsunKarutaUma:
		return 9
	}
	// 数札 1〜9。**丸物は逆順** — 1 が最強 (8) で 9 が最弱 (0)。
	if UnsunKarutaIsRoundSuit(c.GetDesign()) {
		return unsunKarutaMaxPip - c.GetValue()
	}
	return c.GetValue() - 1
}

// Reset はマッチを初期化する。
func (g *UnsunKaruta) Reset() {
	if len(g.players) != UnsunKarutaPlayerCnt {
		g.players = newUnsunKarutaPlayers()
	}
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [UnsunKarutaTeamCnt]int{}
	g.result = UnsunKarutaResultNone
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.actionLog = nil
	g.startRound()
}

// NextRound は次のディールを始める。
func (g *UnsunKaruta) NextRound() {
	if g.phase != UnsunKarutaPhaseRoundEnd || g.gameEndFlag {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % UnsunKarutaPlayerCnt
	g.startRound()
}

// startRound は配って切り札を返し、プレイを始める。
func (g *UnsunKaruta) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.lastTrickWinner = -1
	g.teamTricks = [UnsunKarutaTeamCnt]int{}
	g.scored = false
	g.mustFollow = false
	g.declaredThisTrick = false
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	// **親の左隣がリードする。**
	g.leadPlayerIdx = (g.dealerIdx + 1) % UnsunKarutaPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = UnsunKarutaPhasePlay
	g.appendLog(-1, "deal", fmt.Sprintf("deal %d: trump is %s", g.roundNumber,
		UnsunKarutaSuitName(g.trumpSuit)), []*Card{g.TrumpCard()})
}

// deal は 3 枚ずつ 3 巡で 9 枚を配り、残り 3 枚を中央へ置いて切り札を返す。
func (g *UnsunKaruta) deal() {
	g.deck = buildUnsunKarutaDeck()
	rand.Shuffle(len(g.deck), func(i, j int) { g.deck[i], g.deck[j] = g.deck[j], g.deck[i] })
	g.deckDrawCnt = 0
	const packet = 3
	for r := 0; r < UnsunKarutaHandSize/packet; r++ {
		for j := 0; j < UnsunKarutaPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % UnsunKarutaPlayerCnt
			for k := 0; k < packet; k++ {
				if c := g.drawCard(); c != nil {
					g.players[idx].AddCard(c)
				}
			}
		}
	}
	g.talon = make([]*Card, 0, UnsunKarutaTalonSize)
	for g.deckDrawCnt < len(g.deck) {
		if c := g.drawCard(); c != nil {
			g.talon = append(g.talon, c)
		}
	}
	// **返した 1 枚がそのディールの切り札スート。**
	g.trumpSuit = 0
	if len(g.talon) > 0 && g.talon[0] != nil {
		g.trumpSuit = g.talon[0].GetDesign()
	}
}

// drawCard はデッキから 1 枚配る。
func (g *UnsunKaruta) drawCard() *Card { return drawFromDeck(g.deck, &g.deckDrawCnt) }

// TrumpCard は表に返した切り札札を返す。
func (g *UnsunKaruta) TrumpCard() *Card {
	if len(g.talon) == 0 {
		return nil
	}
	return g.talon[0]
}

// --- プレイ ---

// PlayerPlay は人間が 1 枚出す。リードのときは declare でメリ / モンチを宣言する。
func (g *UnsunKaruta) PlayerPlay(cardIndex int, declare bool) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != UnsunKarutaPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if declare && len(g.currentTrick) > 0 {
		return NewDomainErrorCode(errUnsunKarutaNotLeader, "unsunkaruta.errNotLeader", nil)
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainErrorCode(ErrInvalidCard, "unsunkaruta.errCardRange", nil)
	}
	card := player.GetCard(cardIndex)
	if err := g.validatePlay(g.currentPlayerIdx, card); err != nil {
		return err
	}
	if len(g.currentTrick) == 0 {
		g.setDeclaration(declare, card)
	}
	g.playCard(g.currentPlayerIdx, player.RemoveCard(cardIndex))
	return nil
}

// setDeclaration はリード時の宣言を記録する。
//
// **宣言してはじめてフォロー義務が生まれる。** 切り札リードなら「メリ」、
// 平札リードなら「モンチ」で、どちらも「以降その台札に従え」という同じ効果。
func (g *UnsunKaruta) setDeclaration(declare bool, lead *Card) {
	g.mustFollow = declare
	g.declaredThisTrick = declare
	if !declare {
		return
	}
	name := "monchi"
	if lead != nil && lead.GetDesign() == g.trumpSuit {
		name = "meri"
	}
	g.appendLog(g.currentPlayerIdx, "declare",
		fmt.Sprintf("%s declares %s", playerName(g.players, g.currentPlayerIdx), name), nil)
}

// CpuPlay は CPU が 1 枚出す。
func (g *UnsunKaruta) CpuPlay() {
	if g.gameEndFlag || g.phase != UnsunKarutaPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	card := g.players[idx].GetCard(cardIdx)
	if len(g.currentTrick) == 0 {
		g.setDeclaration(g.cpuDeclares(idx, card), card)
	}
	played := g.players[idx].RemoveCard(cardIdx)
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// playCard は 1 枚出す共通処理。
func (g *UnsunKaruta) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), UnsunKarutaCardName(card)),
		[]*Card{card})
	if len(g.currentTrick) == UnsunKarutaPlayerCnt {
		g.phase = UnsunKarutaPhaseTrickEnd
		return
	}
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % UnsunKarutaPlayerCnt
}

// ResolveTrick はトリックを解決する。
func (g *UnsunKaruta) ResolveTrick() {
	if g.phase != UnsunKarutaPhaseTrickEnd || len(g.currentTrick) != UnsunKarutaPlayerCnt {
		return
	}
	winner := g.trickWinner()
	cards := make([]*Card, 0, UnsunKarutaPlayerCnt)
	for _, tc := range g.currentTrick {
		if tc != nil {
			cards = append(cards, tc.Card)
		}
	}
	g.players[winner].AddTrick(cards)
	g.teamTricks[UnsunKarutaTeamOf(winner)]++
	g.lastTrickWinner = winner
	g.appendLog(winner, "trick_win",
		fmt.Sprintf("%s wins trick %d for team %d",
			playerName(g.players, winner), g.trickNumber, UnsunKarutaTeamOf(winner)), cards)

	g.leadPlayerIdx = winner
	if g.trickNumber >= UnsunKarutaTrickCount {
		g.phase = UnsunKarutaPhaseRoundEnd
		g.enterRoundEnd()
		return
	}
	g.phase = UnsunKarutaPhaseTrickEnd
}

// NextTrick は次のトリックを始める。
func (g *UnsunKaruta) NextTrick() {
	if g.phase != UnsunKarutaPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	// **宣言はトリックごと。** 前のトリックのフォロー義務を持ち越さない。
	g.mustFollow = false
	g.declaredThisTrick = false
	g.phase = UnsunKarutaPhasePlay
}

// trickWinner はトリックの勝者を返す。
func (g *UnsunKaruta) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	led := g.ledSuit()
	winner, best := g.currentTrick[0].PlayerIdx, -1
	for _, tc := range g.currentTrick {
		if tc == nil || tc.Card == nil {
			continue
		}
		rank := unsunKarutaWinRank(tc.Card, led, g.trumpSuit)
		if rank > best {
			best, winner = rank, tc.PlayerIdx
		}
	}
	return winner
}

// unsunKarutaWinRank は勝ち比べの順位を返す。切り札 > 台札 > それ以外。
func unsunKarutaWinRank(c *Card, led, trump int) int {
	if c == nil {
		return -1
	}
	switch c.GetDesign() {
	case trump:
		return 1000 + unsunKarutaStrength(c)
	case led:
		return 100 + unsunKarutaStrength(c)
	default:
		return -1
	}
}

// ledSuit は台札のスートを返す (誰も出していなければ -1)。
func (g *UnsunKaruta) ledSuit() int {
	for _, tc := range g.currentTrick {
		if tc != nil && tc.Card != nil {
			return tc.Card.GetDesign()
		}
	}
	return -1
}

// validatePlay は宣言のあるトリックでのフォロー義務を検証する。
func (g *UnsunKaruta) validatePlay(playerIdx int, card *Card) error {
	return validateCardIsPlayable(g.GetPlayableIndices(playerIdx), g.players[playerIdx], card)
}

// GetPlayableIndices は出せる手札のインデックスを返す。
//
// **宣言が無ければ全部出せる。** フォロー義務はメリ / モンチ を宣言した
// トリックにだけ生まれる ── ここを常時フォローにすると、宣言そのものが
// 意味を失う。
func (g *UnsunKaruta) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	player := g.players[playerIdx]
	all := make([]int, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		all = append(all, i)
	}
	led := g.ledSuit()
	if !g.mustFollow || led < 0 {
		return all
	}
	follow := make([]int, 0, len(all))
	for _, i := range all {
		if c := player.GetCard(i); c != nil && c.GetDesign() == led {
			follow = append(follow, i)
		}
	}
	if len(follow) == 0 {
		return all
	}
	return follow
}

// --- 集計 ---

// ScoreRound は RoundEnd での集計を行う (idempotent)。
func (g *UnsunKaruta) ScoreRound() {
	if g.phase != UnsunKarutaPhaseRoundEnd {
		return
	}
	g.enterRoundEnd()
}

// enterRoundEnd は一度だけ集計する。
func (g *UnsunKaruta) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	for i := range g.teamScores {
		g.teamScores[i] += g.teamTricks[i]
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("deal %d: tricks %v -> totals %v", g.roundNumber, g.teamTricks, g.teamScores), nil)
	g.checkGameEnd()
}

// checkGameEnd は規定ディール数で終局を判定する。
func (g *UnsunKaruta) checkGameEnd() {
	if g.roundNumber < g.config.TargetDeals {
		return
	}
	g.gameEndFlag = true
	g.phase = UnsunKarutaPhaseGameEnd
	switch {
	case g.teamScores[0] > g.teamScores[1]:
		g.winnerTeam = 0
	case g.teamScores[1] > g.teamScores[0]:
		g.winnerTeam = 1
	default:
		g.winnerTeam = -1
	}
	g.result = g.humanResult()
	if g.winnerTeam < 0 {
		g.appendLog(-1, "game_end", "the match ends in a draw", nil)
		return
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("team %d wins the match", g.winnerTeam), nil)
}

// humanResult は人間のチームから見た結果を返す。
func (g *UnsunKaruta) humanResult() UnsunKarutaResult {
	human := findHumanIdx(g.players)
	if human < 0 || g.winnerTeam < 0 {
		return UnsunKarutaResultNone
	}
	if UnsunKarutaTeamOf(human) == g.winnerTeam {
		return UnsunKarutaResultWin
	}
	return UnsunKarutaResultLose
}

// --- CPU ---

// cpuSelectPlayCard は CPU が出す札を選ぶ。
func (g *UnsunKaruta) cpuSelectPlayCard(playerIdx int) int {
	valid := g.GetPlayableIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == UnsunKarutaCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart は「味方が勝っているなら安く、そうでなければ取りにいく」。
func (g *UnsunKaruta) cpuPlaySmart(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	led := g.ledSuit()
	bestRank, bestSeat := -1, -1
	for _, tc := range g.currentTrick {
		if tc == nil {
			continue
		}
		if r := unsunKarutaWinRank(tc.Card, led, g.trumpSuit); r > bestRank {
			bestRank, bestSeat = r, tc.PlayerIdx
		}
	}
	// **味方が取っているトリックは奪わない。** 同じチームで競ると、安い札で
	// 済んだトリックに高い札を捨てることになる。
	friendLeads := bestSeat >= 0 && UnsunKarutaTeamOf(bestSeat) == UnsunKarutaTeamOf(playerIdx)

	winning, cheapest := -1, -1
	for _, idx := range valid {
		c := p.GetCard(idx)
		rank := unsunKarutaWinRank(c, led, g.trumpSuit)
		if rank > bestRank && (winning < 0 || rank < unsunKarutaWinRank(p.GetCard(winning), led, g.trumpSuit)) {
			winning = idx
		}
		if cheapest < 0 || unsunKarutaStrength(c) < unsunKarutaStrength(p.GetCard(cheapest)) {
			cheapest = idx
		}
	}
	if !friendLeads && winning >= 0 {
		return winning
	}
	if cheapest >= 0 {
		return cheapest
	}
	return valid[0]
}

// cpuDeclares は CPU がリードで宣言するかを決める。
//
// **強い台札でだけ宣言する。** フォロー義務は「その台札で押し切れる」ときの
// 武器で、弱い札で宣言すると相手に確実に取られる。
func (g *UnsunKaruta) cpuDeclares(playerIdx int, lead *Card) bool {
	if g.config.CpuDifficulty == UnsunKarutaCpuDifficultyEasy || lead == nil {
		return false
	}
	if lead.GetDesign() == g.trumpSuit {
		return unsunKarutaStrength(lead) >= 9 // 絵札の切り札
	}
	return unsunKarutaStrength(lead) >= 12 // ソウタ以上の平札
}

// GetHint は人間への推奨手を返す。
func (g *UnsunKaruta) GetHint() *UnsunKarutaHint {
	human := findHumanIdx(g.players)
	if human < 0 || g.gameEndFlag {
		return &UnsunKarutaHint{Reason: "none"}
	}
	switch g.phase {
	case UnsunKarutaPhasePlay:
		if g.currentPlayerIdx != human {
			return &UnsunKarutaHint{Reason: "none"}
		}
		reason := "follow_play"
		if len(g.currentTrick) == 0 {
			reason = "lead_strong"
		}
		// **助言は CPU の難易度に引きずられない。** cpuSelectPlayCard は Easy で
		// 合法手からランダムに選ぶので、そのまま使うと「Easy を選んだ人にだけ
		// でたらめな札を勧める」ことになる。
		valid := g.GetPlayableIndices(human)
		if len(valid) == 0 {
			return &UnsunKarutaHint{Reason: "none"}
		}
		return &UnsunKarutaHint{CardIndices: []int{g.cpuPlaySmart(human, valid)}, Reason: reason}
	case UnsunKarutaPhaseTrickEnd:
		return &UnsunKarutaHint{Reason: "next_trick"}
	case UnsunKarutaPhaseRoundEnd:
		return &UnsunKarutaHint{Reason: "next_round"}
	default:
		return &UnsunKarutaHint{Reason: "none"}
	}
}

// --- 参照 ---

// GetConfig はゲーム設定を返す。
func (g *UnsunKaruta) GetConfig() UnsunKarutaConfig { return g.config }

// SetConfig はゲーム設定を設定する。
func (g *UnsunKaruta) SetConfig(c UnsunKarutaConfig) { g.config = c }

// GetPhase は現在のフェーズを返す。
func (g *UnsunKaruta) GetPhase() UnsunKarutaPhase { return g.phase }

// GetPlayers は席を返す。
func (g *UnsunKaruta) GetPlayers() []*UnsunKarutaPlayer { return g.players }

// GetPlayer は指定席を返す (範囲外は nil)。
func (g *UnsunKaruta) GetPlayer(i int) *UnsunKarutaPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetPlayerCnt は席数を返す。
func (g *UnsunKaruta) GetPlayerCnt() int { return len(g.players) }

// GetRoundNumber は現在のディール番号を返す。
func (g *UnsunKaruta) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber は現在のトリック番号を返す。
func (g *UnsunKaruta) GetTrickNumber() int { return g.trickNumber }

// GetCurrentPlayerIdx は手番の席を返す。
func (g *UnsunKaruta) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// GetCurrentTrick は場の札を返す。
func (g *UnsunKaruta) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetLeadPlayerIdx はリードした席を返す。
func (g *UnsunKaruta) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// GetDealerIdx は親の席を返す。
func (g *UnsunKaruta) GetDealerIdx() int { return g.dealerIdx }

// GetTrumpSuit は切り札スートを返す。
func (g *UnsunKaruta) GetTrumpSuit() int { return g.trumpSuit }

// IsMustFollow はいまのトリックにフォロー義務があるかを返す。
func (g *UnsunKaruta) IsMustFollow() bool { return g.mustFollow }

// IsDeclaredThisTrick はいまのトリックで宣言が行われたかを返す。
func (g *UnsunKaruta) IsDeclaredThisTrick() bool { return g.declaredThisTrick }

// CanDeclare は人間がいま宣言できるか (リードの手番か) を返す。
func (g *UnsunKaruta) CanDeclare() bool {
	return g.phase == UnsunKarutaPhasePlay && len(g.currentTrick) == 0 && g.IsHumanTurn()
}

// GetTeamTricks はディール中のチーム別「コ」を返す。
func (g *UnsunKaruta) GetTeamTricks() []int { return g.teamTricks[:] }

// GetTeamScores はマッチ累計を返す。
func (g *UnsunKaruta) GetTeamScores() []int { return g.teamScores[:] }

// GetLastTrickWinner は直前のトリックを取った席を返す。
func (g *UnsunKaruta) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetResult は人間チームから見たマッチ結果を返す。
func (g *UnsunKaruta) GetResult() UnsunKarutaResult { return g.result }

// GetGameEndFlag は終局したかを返す。
func (g *UnsunKaruta) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam は勝ったチームを返す (引き分けは -1)。
func (g *UnsunKaruta) GetWinnerTeam() int { return g.winnerTeam }

// IsHumanTurn は人間の手番かを返す。
func (g *UnsunKaruta) IsHumanTurn() bool {
	return g.phase == UnsunKarutaPhasePlay &&
		g.currentPlayerIdx >= 0 && g.currentPlayerIdx < len(g.players) &&
		g.players[g.currentPlayerIdx].GetIsHuman()
}

// UnsunKarutaSuitName はスートの識別子を返す (i18n キーの一部に使う)。
func UnsunKarutaSuitName(suit int) string {
	switch suit {
	case UnsunKarutaSuitPao:
		return "pao"
	case UnsunKarutaSuitIsu:
		return "isu"
	case UnsunKarutaSuitKotsu:
		return "kotsu"
	case UnsunKarutaSuitOru:
		return "oru"
	case UnsunKarutaSuitKuru:
		return "kuru"
	default:
		return "unknown"
	}
}

// UnsunKarutaRankName は位の識別子を返す (数札は数字)。
func UnsunKarutaRankName(value int) string {
	switch value {
	case UnsunKarutaSota:
		return "sota"
	case UnsunKarutaUma:
		return "uma"
	case UnsunKarutaKiri:
		return "kiri"
	case UnsunKarutaUn:
		return "un"
	case UnsunKarutaSun:
		return "sun"
	case UnsunKarutaRobai:
		return "robai"
	default:
		return fmt.Sprintf("%d", value)
	}
}

// UnsunKarutaCardName は棋譜用の短い表記を返す。
func UnsunKarutaCardName(c *Card) string {
	if c == nil {
		return "??"
	}
	return UnsunKarutaSuitName(c.GetDesign()) + "-" + UnsunKarutaRankName(c.GetValue())
}

// sortAllHands は全員の手札を並べ替える。
func (g *UnsunKaruta) sortAllHands() {
	for _, p := range g.players {
		unsunKarutaSortHand(p)
	}
}

// unsunKarutaSortHand はスート順・強さ順に並べ替える。
//
// **並びは強さ順。** 数札の並びがスートで逆になるので、値の順に並べると
// 丸物だけ弱い札が先頭に来て、画面から強さが読めなくなる。
func unsunKarutaSortHand(p *UnsunKarutaPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := range cards {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(a, b int) bool {
		ca, cb := cards[a], cards[b]
		if ca == nil || cb == nil {
			return cb != nil
		}
		if ca.GetDesign() != cb.GetDesign() {
			return ca.GetDesign() < cb.GetDesign()
		}
		return unsunKarutaStrength(ca) > unsunKarutaStrength(cb)
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// --- JSON ---

// unsunKarutaJSON is the JSON wire format for UnsunKaruta.
type unsunKarutaJSON struct {
	Deck            []*Card                 `json:"dk"`
	DeckDrawCnt     int                     `json:"dc"`
	Players         []*UnsunKarutaPlayer    `json:"pl"`
	Config          UnsunKarutaConfig       `json:"cf"`
	Phase           UnsunKarutaPhase        `json:"ph"`
	RoundNumber     int                     `json:"rn"`
	TrickNumber     int                     `json:"tn"`
	CurrentPlayer   int                     `json:"cp"`
	CurrentTrick    []*TrickCard            `json:"ct"`
	LeadPlayerIdx   int                     `json:"lp"`
	DealerIdx       int                     `json:"di"`
	Talon           []*Card                 `json:"ta"`
	TrumpSuit       int                     `json:"ts"`
	MustFollow      bool                    `json:"mf"`
	Declared        bool                    `json:"dcl"`
	TeamTricks      [UnsunKarutaTeamCnt]int `json:"tt"`
	TeamScores      [UnsunKarutaTeamCnt]int `json:"tsc"`
	LastTrickWinner int                     `json:"lw"`
	Result          UnsunKarutaResult       `json:"rs"`
	Scored          bool                    `json:"sd"`
	GameEndFlag     bool                    `json:"ge"`
	WinnerTeam      int                     `json:"wt"`
	ActionLog       []*ActionLogEntry       `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *UnsunKaruta) MarshalJSON() ([]byte, error) {
	return json.Marshal(unsunKarutaJSON{
		Deck: g.deck, DeckDrawCnt: g.deckDrawCnt, Players: g.players, Config: g.config,
		Phase: g.phase, RoundNumber: g.roundNumber, TrickNumber: g.trickNumber,
		CurrentPlayer: g.currentPlayerIdx, CurrentTrick: g.currentTrick,
		LeadPlayerIdx: g.leadPlayerIdx, DealerIdx: g.dealerIdx, Talon: g.talon,
		TrumpSuit: g.trumpSuit, MustFollow: g.mustFollow, Declared: g.declaredThisTrick,
		TeamTricks: g.teamTricks, TeamScores: g.teamScores, LastTrickWinner: g.lastTrickWinner,
		Result: g.result, Scored: g.scored, GameEndFlag: g.gameEndFlag, WinnerTeam: g.winnerTeam,
		ActionLog: g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **席数と切り札まで検める。** 保存を書き換えれば 3 人の卓や存在しない
// スートの切り札が作れてしまい、次の 1 手で範囲外を読む。
func (g *UnsunKaruta) UnmarshalJSON(data []byte) error {
	var j unsunKarutaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Deck) > unsunKarutaMaxSliceLen || len(j.Players) > unsunKarutaMaxSliceLen ||
		len(j.ActionLog) > unsunKarutaMaxSliceLen || len(j.CurrentTrick) > unsunKarutaMaxSliceLen ||
		len(j.Talon) > unsunKarutaMaxSliceLen {
		return errors.New("unsunkaruta: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("unsunkaruta: invalid config: %w", err)
	}
	if len(j.Players) != UnsunKarutaPlayerCnt {
		return fmt.Errorf("unsunkaruta: seat count %d, want %d", len(j.Players), UnsunKarutaPlayerCnt)
	}
	if int(j.Phase) < int(UnsunKarutaPhasePlay) || int(j.Phase) > UnsunKarutaPhaseMax {
		return fmt.Errorf("unsunkaruta: invalid phase %d", j.Phase)
	}
	if j.TrumpSuit < UnsunKarutaSuitPao || j.TrumpSuit > UnsunKarutaSuitKuru {
		return fmt.Errorf("unsunkaruta: invalid trump suit %d", j.TrumpSuit)
	}
	if j.DealerIdx < 0 || j.DealerIdx >= UnsunKarutaPlayerCnt {
		return fmt.Errorf("unsunkaruta: dealer index %d out of range", j.DealerIdx)
	}
	if j.CurrentPlayer < 0 || j.CurrentPlayer >= UnsunKarutaPlayerCnt {
		return fmt.Errorf("unsunkaruta: current player %d out of range", j.CurrentPlayer)
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("unsunkaruta: round number %d out of range", j.RoundNumber)
	}

	g.deck = j.Deck
	g.deckDrawCnt = j.DeckDrawCnt
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayer
	g.currentTrick = j.CurrentTrick
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.talon = j.Talon
	g.trumpSuit = j.TrumpSuit
	g.mustFollow = j.MustFollow
	g.declaredThisTrick = j.Declared
	g.teamTricks = j.TeamTricks
	g.teamScores = j.TeamScores
	g.lastTrickWinner = j.LastTrickWinner
	g.result = j.Result
	g.scored = j.Scored
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
