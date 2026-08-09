//go:build !js || !wasm || extra

// Package domain ガイゲル (Gaigel) のドメインモデル。
//
// Gaigel はドイツ・シュヴァーベン地方の 4 人 2 チーム制トリックテイキング
// ゲームで、Schnapsen / 66 ファミリーに属する point-trick ゲーム。
//
// デッキは 48 枚 = A,10,K,Q,J,7 の 24 枚 (6 ランク × 4 スート) を 2 組重ねたもの。
// 各プレイヤーに 5 枚ずつ配り (計 20 枚)、次の 1 枚を表向きの切り札表示カードと
// して場に置く (Schnapsen 方式、山札の最後の引き札として残る)。残り 27 枚が山札。
//
// 山札がある間 (第1フェーズ) はマストフォローが無く、何を出してもよい。
// トリック後、勝者から順に各プレイヤーが山札から 1 枚ずつ引いて 5 枚に補充する。
// 自分のリード番で同スートの K と Q を両方持っていれば「マリアージュ」を宣言して
// チームへボーナス点 (20 点、切り札なら 40 点) を獲得し、その K か Q をリードする。
//
// 山札が尽きると第2フェーズに移行し、マストフォロー (フォローできる場合は同スート、
// 無ければ切り札) が義務付けられる。スート内ランクは A>10>K>Q>J>7。同一カード
// (同スート・同ランク) が同一トリックに出た場合は先に出した側が勝つ。
//
// 全カードを出し終えたら各チームのカード点 (A=11,10=10,K=4,Q=3,J=2,7=0) と
// マリアージュ点を集計し、累計に加算する。先に 101 点に達したチームが勝利。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// GaigelPlayerCnt ガイゲルのプレイヤー数 (4人固定)
const GaigelPlayerCnt = 4

// GaigelTeamCnt チーム数
const GaigelTeamCnt = 2

// GaigelHandSize 各プレイヤーの手札枚数 (山札がある間は補充される)
const GaigelHandSize = 5

// GaigelMarriageBonus 通常スートのマリアージュ (K+Q) ボーナス
const GaigelMarriageBonus = 20

// GaigelRoyalMarriageBonus 切り札スートのマリアージュ (ロイヤルマリアージュ) ボーナス
const GaigelRoyalMarriageBonus = 40

// GaigelRoundCardPointsTotal 1ラウンドのカード合計点数
// (11+10+4+3+2+0) × 4スート × 2デッキ = 240
const GaigelRoundCardPointsTotal = 240

// GaigelPhase ゲームフェーズ
type GaigelPhase int

// Gaigelのフェーズ定数
const (
	// GaigelPhasePlay トリックプレイフェーズ
	GaigelPhasePlay GaigelPhase = iota
	// GaigelPhaseTrickEnd トリック終了フェーズ
	GaigelPhaseTrickEnd
	// GaigelPhaseRoundEnd ラウンド終了フェーズ
	GaigelPhaseRoundEnd
	// GaigelPhaseGameEnd ゲーム終了フェーズ
	GaigelPhaseGameEnd
)

// GaigelHint ヒント情報
type GaigelHint struct {
	CardIndex  *int   // 推奨カードインデックス
	Reason     string // ヒント理由キー
	IsMarriage bool   // 推奨アクションがマリアージュ宣言かどうか
}

// Gaigel ガイゲルゲームクラス
type Gaigel struct {
	trumpCards       *TrumpCards
	players          []*GaigelPlayer
	config           GaigelConfig
	phase            GaigelPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	trumpCard        *Card // 場に表向きで置かれる切り札表示カード (山札の最後)
	trumpSuit        int
	leadPlayerIdx    int
	dealerIdx        int
	teamScores       [GaigelTeamCnt]int      // 累計チームスコア
	roundPoints      [GaigelTeamCnt]int      // 当ラウンドのカード点 (トリック獲得分)
	roundMarriage    [GaigelTeamCnt]int      // 当ラウンドのマリアージュ点
	marriageDeclared [CardDesignMax + 1]bool // suit -> 当ラウンドで宣言済か
	lastTrickWinner  int
	gameEndFlag      bool
	winnerTeam       int // -1: 未確定
	actionLogBase
}

// NewGaigel コンストラクタ
func NewGaigel(trumpCards *TrumpCards, players []*GaigelPlayer, config GaigelConfig) *Gaigel {
	return &Gaigel{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerTeam:      -1,
		roundNumber:     0,
		dealerIdx:       0,
		lastTrickWinner: -1,
	}
}

// NewDefaultGaigel 標準4人パートナーシップ構成 (人間チーム0, CPUは交互配置)
// と DefaultGaigelConfig を組み合わせたデフォルト構築。CUI/Web/Worker 共通の SSoT。
func NewDefaultGaigel() *Gaigel {
	players := []*GaigelPlayer{
		NewGaigelPlayer(true, 0),
		NewGaigelPlayer(false, 1),
		NewGaigelPlayer(false, 0),
		NewGaigelPlayer(false, 1),
	}
	return NewGaigel(newGaigelDeck(), players, DefaultGaigelConfig())
}

// newGaigelDeck ガイゲル用48枚デッキを生成する。
// A,10,K,Q,J,7 (値: 1,10,13,12,11,7) × 4スート × 2デッキ = 48枚。
// TrumpCards.go を汚さないよう extra タグ付き Gaigel.go 内に自己完結させる
// (NewTrumpCardsSchnapsen と同じ要領でデッキを直接構築し、各 24 枚を 2 回追加する)。
func newGaigelDeck() *TrumpCards {
	gaigelValues := []int{1, 10, 13, 12, 11, 7} // A,10,K,Q,J,7
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(gaigelValues) * len(suits) * 2 // 48

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for deck := 0; deck < 2; deck++ {
		for _, suit := range suits {
			for _, val := range gaigelValues {
				t.deck = append(t.deck, NewCard(suit, val, false))
			}
		}
	}
	t.deckInit()
	return t
}

// GaigelCardPoints カードの得点を返す (A=11,10=10,K=4,Q=3,J=2,7=0)。
// SchnapsenCardPoints と同一の値だが、schnapsen は別ワーカー (solo) のビルドタグ
// 配下にあり extra ワーカーからは参照できないため、ここで同じ switch を再実装する。
// グローバルマップを避けて全 Cloudflare Worker WASM バイナリのサイズを抑える。
func GaigelCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1: // Ace
		return 11
	case 10: // Ten
		return 10
	case 13: // King
		return 4
	case 12: // Queen
		return 3
	case 11: // Jack
		return 2
	default: // 7
		return 0
	}
}

// GaigelRankOrder カードのスート内順位を返す (大きいほど強い; A>10>K>Q>J>7)。
func GaigelRankOrder(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1: // A
		return 6
	case 10: // 10
		return 5
	case 13: // K
		return 4
	case 12: // Q
		return 3
	case 11: // J
		return 2
	case 7: // 7
		return 1
	default:
		return 0
	}
}

// Reset ゲーム初期化
func (g *Gaigel) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.trickNumber = 0
	g.dealerIdx = 0
	g.teamScores = [GaigelTeamCnt]int{}
	g.actionLog = nil
	g.trumpCard = nil
	g.trumpSuit = 0

	for _, p := range g.players {
		p.ResetRound()
	}

	g.beginRound()
}

// NextRound 次のラウンドを開始する
func (g *Gaigel) NextRound() {
	if g.phase != GaigelPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % GaigelPlayerCnt
	g.trickNumber = 0
	g.trumpCard = nil
	g.trumpSuit = 0

	for _, p := range g.players {
		p.ResetRound()
	}

	g.beginRound()
}

// beginRound ラウンドの初期処理 (配布 + プレイフェーズ突入)
func (g *Gaigel) beginRound() {
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.roundPoints = [GaigelTeamCnt]int{}
	g.roundMarriage = [GaigelTeamCnt]int{}
	g.marriageDeclared = [CardDesignMax + 1]bool{}
	g.lastTrickWinner = -1

	g.trumpCards.Shuffle()
	g.dealInitial()
	g.sortAllHands()
	g.startPlayPhase()
}

// dealInitial 各プレイヤーに 5 枚配り、その次の 1 枚を表向きの切り札表示カードとして置く。
func (g *Gaigel) dealInitial() {
	for range GaigelHandSize {
		for i := range GaigelPlayerCnt {
			player := g.players[(g.dealerIdx+1+i)%GaigelPlayerCnt]
			if c := g.trumpCards.DrawCard(); c != nil {
				player.AddCard(c)
			}
		}
	}
	g.trumpCard = g.trumpCards.DrawCard()
	if g.trumpCard != nil {
		g.trumpSuit = g.trumpCard.GetDesign()
		g.appendLog(-1, "trump", fmt.Sprintf("Trump: %s", cardStr(g.trumpCard)), []*Card{g.trumpCard})
	}
}

// startPlayPhase プレイフェーズ開始: ディーラーの左隣がリード
func (g *Gaigel) startPlayPhase() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % GaigelPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = GaigelPhasePlay
}

// --- Play actions ---

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *Gaigel) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GaigelPhasePlay {
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

// PlayerDeclareMarriage 人間プレイヤーがマリアージュ (K+Q 同スート) を宣言し、
// 指定したカード (その K か Q) をリードする。リード番でのみ有効。
func (g *Gaigel) PlayerDeclareMarriage(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GaigelPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.declareMarriage(g.currentPlayerIdx, cardIndex)
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する
func (g *Gaigel) CpuPlay() {
	if g.gameEndFlag || g.phase != GaigelPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}

	// リード番で有益なマリアージュがあれば宣言してリードする
	if len(g.currentTrick) == 0 {
		if cardIdx, ok := g.cpuChooseMarriage(idx); ok {
			_ = g.declareMarriage(idx, cardIdx)
			return
		}
	}

	cardIdx := g.cpuSelectPlayCard(idx)
	played := g.players[idx].RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// declareMarriage マリアージュを宣言してチームへボーナス加点し、指定の K/Q をリードする共通処理。
// 同一スートのマリアージュは 1 ラウンドにつき 1 回のみ宣言できる (先着優先)。
func (g *Gaigel) declareMarriage(playerIdx, cardIndex int) error {
	if len(g.currentTrick) != 0 {
		return NewDomainError(ErrInvalidPlay, "マリアージュはリード時のみ宣言できます")
	}
	player := g.players[playerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if !g.isMarriageStarter(player, card) {
		return NewDomainError(ErrInvalidPlay, "そのカードでマリアージュは宣言できません")
	}

	suit := card.GetDesign()
	bonus := GaigelMarriageBonus
	if suit == g.trumpSuit {
		bonus = GaigelRoyalMarriageBonus
	}
	team := player.GetTeam()
	g.marriageDeclared[suit] = true
	g.roundMarriage[team] += bonus
	g.appendLog(playerIdx, "marriage",
		fmt.Sprintf("%s declares marriage in %s (+%d for team %d)",
			playerName(g.players, playerIdx), suitStr(suit), bonus, team), nil)

	played := player.RemoveCard(cardIndex)
	g.playCard(playerIdx, played)
	return nil
}

// isMarriageStarter card が「未宣言スートの K か Q で相方を手札に持つ」かを判定する。
func (g *Gaigel) isMarriageStarter(player *GaigelPlayer, card *Card) bool {
	v := card.GetValue()
	if v != 12 && v != 13 {
		return false
	}
	suit := card.GetDesign()
	if g.marriageDeclared[suit] {
		return false
	}
	partner := 12
	if v == 12 {
		partner = 13
	}
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() == suit && c.GetValue() == partner {
			return true
		}
	}
	return false
}

// playCard カードをプレイする共通処理
func (g *Gaigel) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)),
		[]*Card{card})

	if len(g.currentTrick) == GaigelPlayerCnt {
		g.phase = GaigelPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % GaigelPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する
func (g *Gaigel) ResolveTrick() {
	if g.phase != GaigelPhaseTrickEnd || len(g.currentTrick) != GaigelPlayerCnt {
		return
	}

	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	trickPoints := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		trickPoints += GaigelCardPoints(tc.Card)
	}

	g.players[winnerIdx].AddTrick(trickCards)
	g.roundPoints[g.players[winnerIdx].GetTeam()] += trickPoints

	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pt)", playerName(g.players, winnerIdx), g.trickNumber, trickPoints),
		trickCards)

	g.leadPlayerIdx = winnerIdx
	g.lastTrickWinner = winnerIdx
	g.phase = GaigelPhaseTrickEnd
}

// NextTrick 次のトリックを開始する。第1フェーズ (山札あり) では勝者から順に補充する。
// 全カードが尽きたらラウンド終了処理を実行する。
func (g *Gaigel) NextTrick() {
	if g.phase != GaigelPhaseTrickEnd {
		return
	}

	g.drawReplenish()

	if g.allHandsEmpty() {
		g.phase = GaigelPhaseRoundEnd
		return
	}

	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = GaigelPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (g *Gaigel) ScoreRound() {
	if g.phase != GaigelPhaseRoundEnd {
		return
	}

	for ti := range GaigelTeamCnt {
		total := g.roundPoints[ti] + g.roundMarriage[ti]
		g.teamScores[ti] += total
		g.appendLog(-1, "team_score",
			fmt.Sprintf("Team %d: %d card + %d marriage = %d (total %d)",
				ti, g.roundPoints[ti], g.roundMarriage[ti], total, g.teamScores[ti]), nil)
	}

	g.checkGameEnd()
}

// drawReplenish 第1フェーズで勝者から時計回りに各プレイヤーが 1 枚ずつ山札から引く。
// 山札が空 (切り札表示カードも引き終えた) なら第2フェーズなので何もしない。
func (g *Gaigel) drawReplenish() {
	if g.trumpCards.GetRemainingCount() == 0 && g.trumpCard == nil {
		return
	}
	for off := 0; off < GaigelPlayerCnt; off++ {
		idx := (g.leadPlayerIdx + off) % GaigelPlayerCnt
		if c := g.drawOne(); c != nil {
			g.players[idx].AddCard(c)
			g.sortHand(g.players[idx])
		}
	}
}

// drawOne 山札または切り札表示カードから 1 枚引く。優先順位は山札 → 切り札表示カード。
func (g *Gaigel) drawOne() *Card {
	return drawOrTakeTrump(g.trumpCards, &g.trumpCard)
}

// allHandsEmpty 全プレイヤーの手札が空かを返す
func (g *Gaigel) allHandsEmpty() bool {
	return allHandsEmpty(g.players)
}

// IsEndgame 第2フェーズ (山札と切り札表示カードが尽きてマストフォローになる) かを返す
func (g *Gaigel) IsEndgame() bool {
	return g.trumpCards.GetRemainingCount() == 0 && g.trumpCard == nil
}

// --- Validation / follow rules ---

// validatePlay カードのプレイがルール上有効かを検証する。
// 第1フェーズ・リード時は常に有効。第2フェーズの追随時のみマストフォローを課す。
func (g *Gaigel) validatePlay(playerIdx int, card *Card) error {
	return validateEndgameFollow(g.currentTrick, g, playerIdx, card)
}

// cardSatisfiesFollow 第2フェーズの追随時に card が合法かを返す。
// 1) リードスートを持つ: そのスートを出す。
// 2) リードスートを持たないが切り札を持つ: 切り札のみ可。
// 3) どちらも持たない: 任意。
func (g *Gaigel) cardSatisfiesFollow(playerIdx int, card *Card) bool {
	player := g.players[playerIdx]
	leadSuit := g.currentTrick[0].Card.GetDesign()

	if gaigelPlayerHasSuit(player, leadSuit) {
		return card.GetDesign() == leadSuit
	}
	if gaigelPlayerHasSuit(player, g.trumpSuit) {
		return card.GetDesign() == g.trumpSuit
	}
	return true
}

// legalPlayIndices validatePlay を満たすカードのインデックス集合を返す。
func (g *Gaigel) legalPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// gaigelPlayerHasSuit プレイヤーが指定スートのカードを持つか
func gaigelPlayerHasSuit(player *GaigelPlayer, suit int) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		if player.GetCard(i).GetDesign() == suit {
			return true
		}
	}
	return false
}

// trickWinner 現在のトリックの勝者インデックスを決定する。
// 同一カード (同スート・同ランク) が出た場合は先に出した側が勝つ (gaigelBeats は厳密 >)。
func (g *Gaigel) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerCard := g.currentTrick[0].Card

	for _, tc := range g.currentTrick[1:] {
		if gaigelBeats(tc.Card, winnerCard, leadSuit, g.trumpSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// gaigelBeats challenger が currentBest に厳密に勝つかを判定する。
// ・両者がトランプ: ランクの高い方が勝つ
// ・challenger のみトランプ: challenger が勝つ
// ・両者とも非トランプかつ同じリードスート: ランクの高い方が勝つ
// ・両者とも非トランプで challenger がリードスート以外: challenger は勝てない
// ランク同値 (ダブルデッキの同一カード) では false を返すため、先に出した方が勝つ。
func gaigelBeats(challenger, currentBest *Card, leadSuit, trumpSuit int) bool {
	cIsTrump := challenger.GetDesign() == trumpSuit
	bIsTrump := currentBest.GetDesign() == trumpSuit

	switch {
	case cIsTrump && bIsTrump:
		return GaigelRankOrder(challenger) > GaigelRankOrder(currentBest)
	case cIsTrump:
		return true
	case bIsTrump:
		return false
	}
	if challenger.GetDesign() != leadSuit {
		return false
	}
	if currentBest.GetDesign() != leadSuit {
		return true
	}
	return GaigelRankOrder(challenger) > GaigelRankOrder(currentBest)
}

// --- Game end ---

func (g *Gaigel) checkGameEnd() {
	for ti := range GaigelTeamCnt {
		if g.teamScores[ti] >= g.config.TargetScore {
			g.gameEndFlag = true
			g.phase = GaigelPhaseGameEnd
			if g.teamScores[0] >= g.teamScores[1] {
				g.winnerTeam = 0
			} else {
				g.winnerTeam = 1
			}
			g.appendLog(-1, "game_end",
				fmt.Sprintf("Team %d wins the game!", g.winnerTeam), nil)
			return
		}
	}
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (g *Gaigel) GetPhase() GaigelPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Gaigel) SetPhase(p GaigelPhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号取得
func (g *Gaigel) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Gaigel) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (g *Gaigel) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Gaigel) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Gaigel) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Gaigel) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Gaigel) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Gaigel) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetTrumpSuit 切り札スート取得
func (g *Gaigel) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Gaigel) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetTrumpCard 場に表向きで置かれている切り札表示カードを取得 (山札に残っていなければ nil)
func (g *Gaigel) GetTrumpCard() *Card { return g.trumpCard }

// SetTrumpCard 切り札表示カード設定 (テスト用)
func (g *Gaigel) SetTrumpCard(c *Card) { g.trumpCard = c }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Gaigel) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (g *Gaigel) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Gaigel) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Gaigel) GetPlayer(i int) *GaigelPlayer {
	return getPlayer(g.players, i)
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Gaigel) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Gaigel) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Gaigel) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *Gaigel) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetTeamScore チームスコア取得
func (g *Gaigel) GetTeamScore(team int) int {
	if team < 0 || team >= GaigelTeamCnt {
		return 0
	}
	return g.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (g *Gaigel) SetTeamScore(team, score int) {
	if team >= 0 && team < GaigelTeamCnt {
		g.teamScores[team] = score
	}
}

// GetRoundPoints 当ラウンドのチーム別カード点数取得
func (g *Gaigel) GetRoundPoints(team int) int {
	if team < 0 || team >= GaigelTeamCnt {
		return 0
	}
	return g.roundPoints[team]
}

// GetRoundMarriagePoints 当ラウンドのマリアージュ得点取得
func (g *Gaigel) GetRoundMarriagePoints(team int) int {
	if team < 0 || team >= GaigelTeamCnt {
		return 0
	}
	return g.roundMarriage[team]
}

// GetStockRemaining 山札の残り枚数 (場に出ている表向き切り札表示カードは含まない;
// それは GetTrumpCard() != nil の間 別カウントとして残る最後の 1 枚)。
func (g *Gaigel) GetStockRemaining() int {
	return g.trumpCards.GetRemainingCount()
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Gaigel) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Gaigel) GetConfig() GaigelConfig { return g.config }

// SetConfig 設定変更
func (g *Gaigel) SetConfig(cfg GaigelConfig) { g.config = cfg }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す。
// 第1フェーズは制約なし。第2フェーズはマストフォローを適用する。
func (g *Gaigel) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	return g.legalPlayIndices(playerIdx)
}

// GetMarriageIndices リード番でマリアージュ宣言を開始できるカード (未宣言スートの
// K または Q で、相方を手札に持つもの) のインデックスを返す。
func (g *Gaigel) GetMarriageIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	if len(g.currentTrick) != 0 || g.currentPlayerIdx != playerIdx {
		return nil
	}
	p := g.players[playerIdx]
	out := make([]int, 0)
	for i := 0; i < p.GetCardsSize(); i++ {
		if g.isMarriageStarter(p, p.GetCard(i)) {
			out = append(out, i)
		}
	}
	return out
}

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (g *Gaigel) CardRankPublic(card *Card) int { return GaigelRankOrder(card) }

// CardPointsPublic カード得点取得 (テスト用公開メソッド)
func (g *Gaigel) CardPointsPublic(card *Card) int { return GaigelCardPoints(card) }

// --- Hints ---

// GetHint 人間プレイヤー (idx 0) へのヒントを取得する
func (g *Gaigel) GetHint() *GaigelHint {
	if g.phase != GaigelPhasePlay || g.currentPlayerIdx != 0 {
		return nil
	}
	humanIdx := 0
	if g.players[humanIdx].GetCardsSize() == 0 {
		return nil
	}
	if len(g.currentTrick) == 0 {
		if idx, ok := g.cpuChooseMarriage(humanIdx); ok {
			i := idx
			return &GaigelHint{CardIndex: &i, Reason: "marriage", IsMarriage: true}
		}
	}
	idx := g.cpuSelectPlayCard(humanIdx)
	return &GaigelHint{CardIndex: &idx, Reason: g.playHintReason(humanIdx, idx)}
}

// playHintReason ヒント理由キーを判定する
func (g *Gaigel) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	pts := GaigelCardPoints(card)
	if len(g.currentTrick) == 0 {
		if card.GetDesign() == g.trumpSuit {
			return "lead_trump"
		}
		if pts == 0 {
			return "lead_low"
		}
		return "lead_value"
	}
	leadCard := g.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	if gaigelBeats(card, leadCard, leadSuit, g.trumpSuit) {
		if card.GetDesign() == g.trumpSuit && leadSuit != g.trumpSuit {
			return "follow_cut"
		}
		return "follow_win"
	}
	return "follow_dump"
}

// --- CPU AI ---

// cpuChooseMarriage CPU がリード時に宣言すべきマリアージュのカードインデックスを返す。
// 切り札マリアージュを優先し、宣言時はランクの低い Q をリードする。
func (g *Gaigel) cpuChooseMarriage(playerIdx int) (int, bool) {
	player := g.players[playerIdx]
	bestIdx := -1
	bestBonus := -1
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetValue() != 12 { // Q を起点に宣言 (低い方をリードして K を温存)
			continue
		}
		if !g.isMarriageStarter(player, c) {
			continue
		}
		bonus := GaigelMarriageBonus
		if c.GetDesign() == g.trumpSuit {
			bonus = GaigelRoyalMarriageBonus
		}
		if bonus > bestBonus {
			bestBonus = bonus
			bestIdx = i
		}
	}
	return bestIdx, bestIdx >= 0
}

// cpuSelectPlayCard CPU が出すべきカードのインデックスを選択する (合法手の中から)
func (g *Gaigel) cpuSelectPlayCard(playerIdx int) int {
	legal := g.legalPlayIndices(playerIdx)
	if len(legal) == 0 {
		return 0
	}
	if len(legal) == 1 {
		return legal[0]
	}
	if g.config.CpuDifficulty == GaigelCpuDifficultyEasy {
		return legal[rand.Intn(len(legal))]
	}
	if len(g.currentTrick) == 0 {
		return g.cpuLead(playerIdx, legal)
	}
	return g.cpuFollow(playerIdx, legal)
}

// cpuLead リード時の選択: 最も低い点数の非トランプを優先する。
func (g *Gaigel) cpuLead(playerIdx int, legal []int) int {
	player := g.players[playerIdx]
	bestIdx := legal[0]
	bestScore := gaigelLeadScore(player.GetCard(bestIdx), g.trumpSuit)
	for _, i := range legal[1:] {
		sc := gaigelLeadScore(player.GetCard(i), g.trumpSuit)
		if sc < bestScore {
			bestScore = sc
			bestIdx = i
		}
	}
	return bestIdx
}

// gaigelLeadScore 値が小さいほど「リードに適している」(トランプ・高得点札を温存する)
func gaigelLeadScore(c *Card, trumpSuit int) int {
	score := GaigelCardPoints(c)*10 + GaigelRankOrder(c)
	if c.GetDesign() == trumpSuit {
		score += 1000
	}
	return score
}

// cpuFollow 追随時の選択。パートナーが勝っていれば高点を載せ、そうでなければ
// 勝てる最小コストの札、無ければ最小ダンプ札を出す。
func (g *Gaigel) cpuFollow(playerIdx int, legal []int) int {
	player := g.players[playerIdx]
	leadCard := g.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	partnerIdx := (playerIdx + 2) % GaigelPlayerCnt
	partnerWinning := g.currentLeaderIdx() == partnerIdx

	if partnerWinning {
		best := legal[0]
		bestPts := -1
		for _, i := range legal {
			pts := GaigelCardPoints(player.GetCard(i))
			if pts > bestPts {
				bestPts = pts
				best = i
			}
		}
		return best
	}

	winIdx := -1
	winScore := 0
	dumpIdx := legal[0]
	dumpScoreVal := gaigelLeadScore(player.GetCard(legal[0]), g.trumpSuit)
	for _, i := range legal {
		c := player.GetCard(i)
		if gaigelBeats(c, g.currentBestCard(), leadSuit, g.trumpSuit) {
			sc := gaigelLeadScore(c, g.trumpSuit)
			if winIdx < 0 || sc < winScore {
				winIdx = i
				winScore = sc
			}
		}
		ds := gaigelLeadScore(c, g.trumpSuit)
		if ds < dumpScoreVal {
			dumpScoreVal = ds
			dumpIdx = i
		}
	}
	if winIdx >= 0 && GaigelCardPoints(leadCard) >= 10 {
		return winIdx
	}
	if !g.legalAllowsDump(playerIdx, legal) && winIdx >= 0 {
		return winIdx
	}
	return dumpIdx
}

// currentLeaderIdx 現時点のトリックの暫定勝者のプレイヤーインデックスを返す
func (g *Gaigel) currentLeaderIdx() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerCard := g.currentTrick[0].Card
	for _, tc := range g.currentTrick[1:] {
		if gaigelBeats(tc.Card, winnerCard, leadSuit, g.trumpSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// currentBestCard 現時点のトリックの暫定勝者カードを返す
func (g *Gaigel) currentBestCard() *Card {
	idx := g.currentLeaderIdx()
	for _, tc := range g.currentTrick {
		if tc.PlayerIdx == idx {
			return tc.Card
		}
	}
	if len(g.currentTrick) > 0 {
		return g.currentTrick[0].Card
	}
	return nil
}

// legalAllowsDump 合法手の中に「勝たないカード」が含まれるか (= 捨てる自由があるか) を返す。
func (g *Gaigel) legalAllowsDump(playerIdx int, legal []int) bool {
	player := g.players[playerIdx]
	best := g.currentBestCard()
	if best == nil {
		return true
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	for _, i := range legal {
		if !gaigelBeats(player.GetCard(i), best, leadSuit, g.trumpSuit) {
			return true
		}
	}
	return false
}

// --- Sorting / bookkeeping ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *Gaigel) sortAllHands() {
	sortEachHand(g.players, g.sortHand)
}

// sortHand プレイヤーの手札をスート (トランプ最後) → ランク でソートする
func (g *Gaigel) sortHand(p *GaigelPlayer) {
	trumpSuit := g.trumpSuit
	sortPlayerHand(p, func(ci, cj *Card) bool {
		iTrump := ci.GetDesign() == trumpSuit
		jTrump := cj.GetDesign() == trumpSuit
		if iTrump != jTrump {
			return !iTrump
		}
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return GaigelRankOrder(ci) < GaigelRankOrder(cj)
	})
}

// --- Test-only helpers ---

// AddRoundPointsForTest adds card points to a team for the current round (テスト用)。
func (g *Gaigel) AddRoundPointsForTest(team, pts int) {
	if team >= 0 && team < GaigelTeamCnt {
		g.roundPoints[team] += pts
	}
}

// RebeginRoundForTest re-deals and re-enters the play phase (テスト用)。
func (g *Gaigel) RebeginRoundForTest() {
	for _, p := range g.players {
		p.ResetRound()
	}
	g.beginRound()
}

// GetConfigDeckHelper returns a fresh 48-card Gaigel deck (テスト用コンストラクタ補助)。
func (g *Gaigel) GetConfigDeckHelper() *TrumpCards { return newGaigelDeck() }

// --- JSON ---

// gaigelJSON is the JSON wire format for Gaigel.
type gaigelJSON struct {
	TrumpCards       *TrumpCards             `json:"tc"`
	Players          []*GaigelPlayer         `json:"pl"`
	Config           GaigelConfig            `json:"cfg"`
	Phase            GaigelPhase             `json:"ph"`
	RoundNumber      int                     `json:"rn"`
	TrickNumber      int                     `json:"tn"`
	CurrentPlayerIdx int                     `json:"cp"`
	CurrentTrick     []*TrickCard            `json:"ct"`
	TrumpCard        *Card                   `json:"tu"`
	TrumpSuit        int                     `json:"ts"`
	LeadPlayerIdx    int                     `json:"li"`
	DealerIdx        int                     `json:"di"`
	TeamScores       [GaigelTeamCnt]int      `json:"sc"`
	RoundPoints      [GaigelTeamCnt]int      `json:"rp"`
	RoundMarriage    [GaigelTeamCnt]int      `json:"rm"`
	MarriageDeclared [CardDesignMax + 1]bool `json:"md"`
	LastTrickWinner  int                     `json:"lw"`
	GameEndFlag      bool                    `json:"ge"`
	WinnerTeam       int                     `json:"wt"`
	ActionLog        []*ActionLogEntry       `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Gaigel) MarshalJSON() ([]byte, error) {
	return json.Marshal(gaigelJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		TrumpCard:        g.trumpCard,
		TrumpSuit:        g.trumpSuit,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		TeamScores:       g.teamScores,
		RoundPoints:      g.roundPoints,
		RoundMarriage:    g.roundMarriage,
		MarriageDeclared: g.marriageDeclared,
		LastTrickWinner:  g.lastTrickWinner,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// gaigelMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const gaigelMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler. 列挙値・インデックス範囲・スライス要素を
// 検証し、不正な場合はエラーを返す (パニックさせない)。
func (g *Gaigel) UnmarshalJSON(data []byte) error {
	var j gaigelJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < GaigelPhasePlay || j.Phase > GaigelPhaseGameEnd {
		return NewDomainError(ErrInvalidPlay, "無効なフェーズです")
	}
	// trumpSuit=0 (未確定) は許可。確定済みの場合のみ範囲チェック。
	if j.TrumpSuit != 0 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return NewDomainError(ErrInvalidPlay, "無効な切り札スートです")
	}
	if len(j.Players) != GaigelPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "プレイヤー数が不正です")
	}
	// 直接インデックス参照される値の範囲検証 (不正な KV 状態でのパニック防止)。
	// currentPlayerIdx / dealerIdx は常に有効なプレイヤー。leadPlayerIdx /
	// lastTrickWinner / winnerTeam は未確定を表す -1 を許可する。
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= GaigelPlayerCnt ||
		j.DealerIdx < 0 || j.DealerIdx >= GaigelPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "プレイヤーインデックスが範囲外です")
	}
	if j.LeadPlayerIdx < -1 || j.LeadPlayerIdx >= GaigelPlayerCnt ||
		j.LastTrickWinner < -1 || j.LastTrickWinner >= GaigelPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "リード/勝者インデックスが範囲外です")
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= GaigelTeamCnt {
		return NewDomainError(ErrInvalidPlay, "勝者チームが範囲外です")
	}
	if len(j.CurrentTrick) > GaigelPlayerCnt || len(j.ActionLog) > gaigelMaxSliceLen {
		return NewDomainError(ErrInvalidPlay, "状態スライスが不正です")
	}
	for _, p := range j.Players {
		if p == nil {
			return NewDomainError(ErrInvalidPlay, "プレイヤーが nil です")
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return NewDomainError(ErrInvalidPlay, "トリックカードが nil です")
		}
	}
	for _, entry := range j.ActionLog {
		if entry == nil {
			return NewDomainError(ErrInvalidPlay, "棋譜エントリが nil です")
		}
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newGaigelDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	g.trumpCard = j.TrumpCard
	g.trumpSuit = j.TrumpSuit
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.teamScores = j.TeamScores
	g.roundPoints = j.RoundPoints
	g.roundMarriage = j.RoundMarriage
	g.marriageDeclared = j.MarriageDeclared
	g.lastTrickWinner = j.LastTrickWinner
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	return nil
}
