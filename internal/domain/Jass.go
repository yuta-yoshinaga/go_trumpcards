//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// JassPlayerCnt ヤスプレイヤー数
const JassPlayerCnt = 4

// JassHandSize 各プレイヤーの手札枚数
const JassHandSize = 9

// JassTeamCnt チーム数
const JassTeamCnt = 2

// JassStockBonus Stöck (トランプの K+Q) ボーナス
const JassStockBonus = 20

// JassRoundCardPointsTotal 1ラウンドのカード合計点数 (最終トリックボーナス含まず)
const JassRoundCardPointsTotal = 152

// JassPhase ゲームフェーズ
type JassPhase int

// Jassのフェーズ定数
const (
	// JassPhaseBidTrump フォアハンドが切り札を選ぶ (またはパートナーへ Schieben)
	JassPhaseBidTrump JassPhase = 0
	// JassPhaseBidPartner Schieben 後にパートナーが切り札を選ぶ (必須)
	JassPhaseBidPartner JassPhase = 1
	// JassPhasePlay トリックプレイフェーズ
	JassPhasePlay JassPhase = 2
	// JassPhaseTrickEnd トリック終了フェーズ
	JassPhaseTrickEnd JassPhase = 3
	// JassPhaseRoundEnd ラウンド終了フェーズ
	JassPhaseRoundEnd JassPhase = 4
	// JassPhaseGameEnd ゲーム終了フェーズ
	JassPhaseGameEnd JassPhase = 5
)

// JassHint ヒント情報
type JassHint struct {
	CardIndex *int   // 推奨カードインデックス (プレイ時)
	Schieben  *bool  // Schieben すべきか (フォアハンドのビッド時)
	Suit      *int   // 推奨切り札スート (ビッド時)
	Reason    string // ヒント理由キー
}

// Jass ヤス(シーバー)ゲームクラス
type Jass struct {
	trumpCards       *TrumpCards
	players          []*JassPlayer
	config           JassConfig
	phase            JassPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	dealerIdx        int
	forehandIdx      int  // フォアハンド (ビッド開始 & 第1トリックのリード)
	trumpSuit        int  // 切り札スート (0 = 未確定)
	makerTeam        int  // 切り札を指名したチーム (0 or 1)
	makerPlayerIdx   int  // 切り札を指名したプレイヤー
	schieben         bool // フォアハンドが Schieben (パス) したか
	teamScores       [JassTeamCnt]int
	roundPoints      [JassTeamCnt]int // 当ラウンドの累計カード点数 (最終トリックボーナス含む)
	roundWeisPoints  [JassTeamCnt]int // 当ラウンドの Weis 得点
	roundStockPoints [JassTeamCnt]int // 当ラウンドの Stöck 得点
	weisResolved     bool             // Weis 比較が済んだか
	lastTrickWinner  int
	leadPlayerIdx    int
	bidPlayerIdx     int
	gameEndFlag      bool
	winnerTeam       int // 勝利チーム (-1 = 未確定)
	actionLogBase
}

// NewJass コンストラクタ
func NewJass(trumpCards *TrumpCards, players []*JassPlayer, config JassConfig) *Jass {
	return &Jass{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerTeam:      -1,
		roundNumber:     0,
		dealerIdx:       0,
		lastTrickWinner: -1,
	}
}

// NewDefaultJass 標準4人パートナーシップ構成 (人間チーム0, CPUは交互配置)
// と DefaultJassConfig を組み合わせたデフォルト構築。CUI/Web/Worker 共通の SSoT。
func NewDefaultJass() *Jass {
	players := []*JassPlayer{
		NewJassPlayer(true, 0),
		NewJassPlayer(false, 1),
		NewJassPlayer(false, 0),
		NewJassPlayer(false, 1),
	}
	return NewJass(newJassDeck(), players, DefaultJassConfig())
}

// newJassDeck ヤス用36枚デッキを生成する。
// 6,7,8,9,10,J,Q,K,A (値: 1,6,7,8,9,10,11,12,13) × 4スート = 36枚。
// TrumpCards.go を汚さないよう extra タグ付き Jass.go 内に自己完結させる
// (NewTrumpCardsBelote と同じ要領でデッキを直接構築する)。
func newJassDeck() *TrumpCards {
	jassValues := []int{1, 6, 7, 8, 9, 10, 11, 12, 13} // A,6,7,8,9,10,J,Q,K
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	totalCards := len(jassValues) * len(suits) // 36

	t := new(TrumpCards)
	t.deckCnt = totalCards
	t.deck = make([]*Card, 0, totalCards)
	for _, suit := range suits {
		for _, val := range jassValues {
			t.deck = append(t.deck, NewCard(suit, val, false))
		}
	}
	t.deckInit()
	return t
}

// Reset ゲーム初期化
func (g *Jass) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.trickNumber = 0
	g.dealerIdx = 0
	g.teamScores = [JassTeamCnt]int{}
	g.actionLog = nil
	g.trumpSuit = 0
	g.makerTeam = 0
	g.makerPlayerIdx = -1

	for _, p := range g.players {
		p.ResetRound()
	}

	g.beginRound()
}

// NextRound 次のラウンドを開始する
func (g *Jass) NextRound() {
	if g.phase != JassPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % JassPlayerCnt
	g.trickNumber = 0
	g.trumpSuit = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.makerPlayerIdx = -1

	for _, p := range g.players {
		p.ResetRound()
	}

	g.beginRound()
}

// beginRound ラウンドの初期処理 (配布 + ビッドフェーズ突入)
func (g *Jass) beginRound() {
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.roundPoints = [JassTeamCnt]int{}
	g.roundWeisPoints = [JassTeamCnt]int{}
	g.roundStockPoints = [JassTeamCnt]int{}
	g.weisResolved = false
	g.lastTrickWinner = -1
	g.schieben = false
	g.makerTeam = 0

	g.dealAll()
	g.forehandIdx = (g.dealerIdx + 1) % JassPlayerCnt
	g.bidPlayerIdx = g.forehandIdx
	g.phase = JassPhaseBidTrump
}

// dealAll 各プレイヤーに 9 枚配る (36 枚消費)
func (g *Jass) dealAll() {
	g.trumpCards.Shuffle()
	for range JassHandSize {
		for j := range JassPlayerCnt {
			card := g.trumpCards.DrawCard()
			if card != nil {
				g.players[j].AddCard(card)
			}
		}
	}
}

// --- Bid: Schieber ---

// PlayerChooseTrump 人間プレイヤーが切り札スートを指名する
func (g *Jass) PlayerChooseTrump(suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != JassPhaseBidTrump && g.phase != JassPhaseBidPartner {
		return ErrWrongPhase
	}
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= JassPlayerCnt || !g.players[g.bidPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "無効なスートです")
	}
	g.doChooseTrump(g.bidPlayerIdx, suit)
	return nil
}

// PlayerSchieben 人間プレイヤー(フォアハンド)がパートナーへ Schieben する
func (g *Jass) PlayerSchieben() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != JassPhaseBidTrump {
		return ErrWrongPhase
	}
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= JassPlayerCnt || !g.players[g.bidPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.doSchieben(g.bidPlayerIdx)
	return nil
}

// CpuBid CPUプレイヤーがビッド判断する (切り札選択 or Schieben)
func (g *Jass) CpuBid() {
	if g.gameEndFlag {
		return
	}
	if g.phase != JassPhaseBidTrump && g.phase != JassPhaseBidPartner {
		return
	}
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= JassPlayerCnt {
		return
	}
	if g.players[g.bidPlayerIdx].GetIsHuman() {
		return
	}

	suit, score := g.cpuBestTrump(g.bidPlayerIdx)
	// フォアハンド (PhaseBidTrump) で手札が弱ければパートナーへ回す。
	if g.phase == JassPhaseBidTrump && g.cpuShouldSchieben(score) {
		g.doSchieben(g.bidPlayerIdx)
		return
	}
	g.doChooseTrump(g.bidPlayerIdx, suit)
}

// doSchieben Schieben (パートナーへの委譲) を確定する
func (g *Jass) doSchieben(playerIdx int) {
	g.schieben = true
	g.appendLog(playerIdx, "schieben",
		fmt.Sprintf("%s schiebt (passes to partner)", playerName(g.players, playerIdx)), nil)
	g.bidPlayerIdx = (playerIdx + 2) % JassPlayerCnt
	g.phase = JassPhaseBidPartner
}

// doChooseTrump 切り札スートを確定し、プレイフェーズへ移行する
func (g *Jass) doChooseTrump(playerIdx, suit int) {
	g.trumpSuit = suit
	g.makerTeam = g.players[playerIdx].GetTeam()
	g.makerPlayerIdx = playerIdx
	g.appendLog(playerIdx, "choose_trump",
		fmt.Sprintf("%s chooses %s as trump", playerName(g.players, playerIdx), suitStr(suit)), nil)

	g.sortAllHands()
	g.resolveWeis()
	g.resolveStock()
	g.startPlayPhase()
}

// --- Play Phase ---

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *Jass) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != JassPhasePlay {
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

// CpuPlay CPUプレイヤーが1ターン実行
func (g *Jass) CpuPlay() {
	if g.gameEndFlag || g.phase != JassPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := g.players[g.currentPlayerIdx]
	cardIdx := g.cpuSelectPlayCard(g.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(g.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (g *Jass) ResolveTrick() {
	if g.phase != JassPhaseTrickEnd || len(g.currentTrick) != JassPlayerCnt {
		return
	}

	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	trickPoints := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		trickPoints += jassCardPoints(tc.Card, g.trumpSuit)
	}

	g.players[winnerIdx].AddTrick(trickCards)
	g.roundPoints[g.players[winnerIdx].GetTeam()] += trickPoints

	winnerName := playerName(g.players, winnerIdx)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pts)", winnerName, g.trickNumber, trickPoints),
		trickCards)

	g.leadPlayerIdx = winnerIdx
	g.lastTrickWinner = winnerIdx

	if g.trickNumber >= JassHandSize {
		// 最終トリックボーナス
		g.roundPoints[g.players[winnerIdx].GetTeam()] += g.config.LastTrickBonus
		g.appendLog(winnerIdx, "last_trick",
			fmt.Sprintf("%s wins last trick +%d", winnerName, g.config.LastTrickBonus), nil)
		g.phase = JassPhaseRoundEnd
	} else {
		g.phase = JassPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (g *Jass) NextTrick() {
	if g.phase != JassPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = JassPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (g *Jass) ScoreRound() {
	if g.phase != JassPhaseRoundEnd {
		return
	}

	for ti := range JassTeamCnt {
		total := g.roundPoints[ti] + g.roundWeisPoints[ti] + g.roundStockPoints[ti]
		g.teamScores[ti] += total
		g.appendLog(-1, "team_score",
			fmt.Sprintf("Team %d: %d card + %d weis + %d stock = %d (total %d)",
				ti, g.roundPoints[ti], g.roundWeisPoints[ti], g.roundStockPoints[ti],
				total, g.teamScores[ti]), nil)
	}

	g.checkGameEnd()
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (g *Jass) GetPhase() JassPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Jass) SetPhase(p JassPhase) { g.phase = p }

// GetRoundNumber 現在のラウンド番号取得
func (g *Jass) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Jass) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (g *Jass) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Jass) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Jass) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Jass) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Jass) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Jass) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Jass) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (g *Jass) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Jass) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Jass) GetPlayer(i int) *JassPlayer {
	return getPlayer(g.players, i)
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Jass) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Jass) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (g *Jass) GetBidPlayerIdx() int { return g.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (g *Jass) SetBidPlayerIdx(idx int) { g.bidPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Jass) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *Jass) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetForehandIdx フォアハンドインデックス取得
func (g *Jass) GetForehandIdx() int { return g.forehandIdx }

// GetTrumpSuit 切り札スート取得
func (g *Jass) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Jass) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetSchieben Schieben 状態取得
func (g *Jass) GetSchieben() bool { return g.schieben }

// GetMakerTeam メイカーチーム取得
func (g *Jass) GetMakerTeam() int { return g.makerTeam }

// SetMakerTeam メイカーチーム設定 (テスト用)
func (g *Jass) SetMakerTeam(team int) { g.makerTeam = team }

// GetMakerPlayerIdx メイカープレイヤー取得
func (g *Jass) GetMakerPlayerIdx() int { return g.makerPlayerIdx }

// GetTeamScore チームスコア取得
func (g *Jass) GetTeamScore(team int) int {
	if team < 0 || team >= JassTeamCnt {
		return 0
	}
	return g.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (g *Jass) SetTeamScore(team, score int) {
	if team >= 0 && team < JassTeamCnt {
		g.teamScores[team] = score
	}
}

// GetRoundPoints 当ラウンドのチーム別カード点数取得
func (g *Jass) GetRoundPoints(team int) int {
	if team < 0 || team >= JassTeamCnt {
		return 0
	}
	return g.roundPoints[team]
}

// GetRoundWeisPoints 当ラウンドの Weis 得点取得
func (g *Jass) GetRoundWeisPoints(team int) int {
	if team < 0 || team >= JassTeamCnt {
		return 0
	}
	return g.roundWeisPoints[team]
}

// GetRoundStockPoints 当ラウンドの Stöck 得点取得
func (g *Jass) GetRoundStockPoints(team int) int {
	if team < 0 || team >= JassTeamCnt {
		return 0
	}
	return g.roundStockPoints[team]
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Jass) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (g *Jass) IsHumanBidTurn() bool {
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.bidPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Jass) GetConfig() JassConfig { return g.config }

// SetConfig 設定変更
func (g *Jass) SetConfig(cfg JassConfig) { g.config = cfg }

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (g *Jass) CardRankPublic(card *Card) int { return g.cardRank(card) }

// CardPointsPublic カード得点取得 (テスト用公開メソッド)
func (g *Jass) CardPointsPublic(card *Card) int { return jassCardPoints(card, g.trumpSuit) }

// --- Test-only helpers (exercise unexported scoring paths deterministically) ---

// ResolveWeisForTest sets the trump/forehand and runs Weis resolution (テスト用)。
func (g *Jass) ResolveWeisForTest(trumpSuit, forehandIdx int) {
	g.trumpSuit = trumpSuit
	g.forehandIdx = forehandIdx
	g.weisResolved = false
	g.resolveWeis()
}

// ResolveStockForTest sets the trump and runs Stöck resolution (テスト用)。
func (g *Jass) ResolveStockForTest(trumpSuit int) {
	g.trumpSuit = trumpSuit
	g.resolveStock()
}

// AddRoundPointsForTest adds card points to a team for the current round (テスト用)。
func (g *Jass) AddRoundPointsForTest(team, pts int) {
	if team >= 0 && team < JassTeamCnt {
		g.roundPoints[team] += pts
	}
}

// SetBidPlayerIdxForTest sets the active bid player (テスト用)。
func (g *Jass) SetBidPlayerIdxForTest(idx int) { g.bidPlayerIdx = idx }

// SetDealerIdxForTest sets the dealer (テスト用)。
func (g *Jass) SetDealerIdxForTest(idx int) { g.dealerIdx = idx }

// RebeginRoundForTest re-deals and re-enters the bid phase (テスト用)。
func (g *Jass) RebeginRoundForTest() {
	for _, p := range g.players {
		p.ResetRound()
	}
	g.beginRound()
}

// GetConfigDeckHelper returns a fresh 36-card Jass deck (テスト用コンストラクタ補助)。
func (g *Jass) GetConfigDeckHelper() *TrumpCards { return newJassDeck() }

// --- Ranking + scoring helpers ---

// jassTrumpRank トランプスートのカードランク (高 = 強)
// J=9(highest) > 9=8 > A=7 > K=6 > Q=5 > 10=4 > 8=3 > 7=2 > 6=1
func jassTrumpRank(value int) int {
	switch value {
	case 11: // J
		return 9
	case 9:
		return 8
	case 1: // A
		return 7
	case 13: // K
		return 6
	case 12: // Q
		return 5
	case 10:
		return 4
	case 8:
		return 3
	case 7:
		return 2
	case 6:
		return 1
	}
	return 0
}

// jassNonTrumpRank 非トランプスートのカードランク (高 = 強)
// A=9(highest) > K=8 > Q=7 > J=6 > 10=5 > 9=4 > 8=3 > 7=2 > 6=1
func jassNonTrumpRank(value int) int {
	switch value {
	case 1: // A
		return 9
	case 13: // K
		return 8
	case 12: // Q
		return 7
	case 11: // J
		return 6
	case 10:
		return 5
	case 9:
		return 4
	case 8:
		return 3
	case 7:
		return 2
	case 6:
		return 1
	}
	return 0
}

// jassCardPoints トランプスートを踏まえたカード点数を返す
// 切り札: J=20, 9=14, A=11, 10=10, K=4, Q=3, 8/7/6=0
// 非切り札: A=11, 10=10, K=4, Q=3, J=2, 9/8/7/6=0
func jassCardPoints(c *Card, trumpSuit int) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == trumpSuit {
		switch c.GetValue() {
		case 11: // J
			return 20
		case 9:
			return 14
		case 1: // A
			return 11
		case 10:
			return 10
		case 13: // K
			return 4
		case 12: // Q
			return 3
		}
		return 0
	}
	switch c.GetValue() {
	case 1: // A
		return 11
	case 10:
		return 10
	case 13: // K
		return 4
	case 12: // Q
		return 3
	case 11: // J
		return 2
	}
	return 0
}

// cardRank トリック比較用ランクを返す (高い = 強い)
// 切り札スート: 200 + trumpRank, 非切り札: 100 + nonTrumpRank
func (g *Jass) cardRank(card *Card) int {
	if card.GetDesign() == g.trumpSuit {
		return 200 + jassTrumpRank(card.GetValue())
	}
	return 100 + jassNonTrumpRank(card.GetValue())
}

// --- Weis (melds) ---

// jassWeis 1プレイヤーの最良 Weis 情報
type jassWeis struct {
	points   int // 当該プレイヤーの全 Weis 合計点 (チーム加点用)
	bestVal  int // 比較用: 最良 Weis の点数
	bestLen  int // 比較用: 最良 Weis の長さ (同点時のタイブレーク)
	bestRank int // 比較用: 最良 Weis の最高位カードランク
}

// jassSeqRank シーケンス内のランク順 (6 < 7 < ... < A)
// 6=1,7=2,8=3,9=4,10=5,J=6,Q=7,K=8,A=9
func jassSeqRank(value int) int {
	switch value {
	case 6:
		return 1
	case 7:
		return 2
	case 8:
		return 3
	case 9:
		return 4
	case 10:
		return 5
	case 11:
		return 6
	case 12:
		return 7
	case 13:
		return 8
	case 1:
		return 9
	}
	return 0
}

// computeWeis 1プレイヤーの最良 Weis と全 Weis 合計点を計算する
func computeWeis(p *JassPlayer) jassWeis {
	w := jassWeis{}
	if p == nil {
		return w
	}

	// 各スートのシーケンスランク集合を作る
	suitRanks := map[int][]int{}
	valCount := map[int]int{}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		suitRanks[c.GetDesign()] = append(suitRanks[c.GetDesign()], jassSeqRank(c.GetValue()))
		valCount[c.GetValue()]++
	}

	// シーケンス検出
	for _, ranks := range suitRanks {
		sort.Ints(ranks)
		runStart := 0
		for i := 1; i <= len(ranks); i++ {
			if i < len(ranks) && ranks[i] == ranks[i-1]+1 {
				continue
			}
			runLen := i - runStart
			if runLen >= 3 {
				pts := jassSequencePoints(runLen)
				w.points += pts
				topRank := ranks[i-1]
				w.consider(pts, runLen, topRank)
			}
			runStart = i
		}
	}

	// 4枚同位 (four of a kind)
	for val, cnt := range valCount {
		if cnt < 4 {
			continue
		}
		pts := jassFourOfAKindPoints(val)
		if pts == 0 {
			continue
		}
		w.points += pts
		// 4枚同位の比較ランクは高位カード扱い (シーケンスより常に優先されるよう長さ大)
		w.consider(pts, 4, 100+val)
	}

	return w
}

// consider 比較用の最良 Weis を更新する
func (w *jassWeis) consider(pts, length, topRank int) {
	if pts > w.bestVal ||
		(pts == w.bestVal && length > w.bestLen) ||
		(pts == w.bestVal && length == w.bestLen && topRank > w.bestRank) {
		w.bestVal = pts
		w.bestLen = length
		w.bestRank = topRank
	}
}

// jassSequencePoints シーケンスの点数 (3=20, 4=50, 5+=100)
func jassSequencePoints(length int) int {
	switch {
	case length >= 5:
		return 100
	case length == 4:
		return 50
	case length == 3:
		return 20
	}
	return 0
}

// jassFourOfAKindPoints 4枚同位の点数
// J×4=200, 9×4=150, {A,K,Q,10}×4=100, それ以外(6,7,8)=0
func jassFourOfAKindPoints(value int) int {
	switch value {
	case 11: // J
		return 200
	case 9:
		return 150
	case 1, 13, 12, 10: // A,K,Q,10
		return 100
	}
	return 0
}

// resolveWeis 全プレイヤーの Weis を比較し、勝者チームへ全 Weis を加点する。
// タイブレーク: 同点・同長・同最高位ランクの場合、フォアハンドに近いプレイヤー
// (フォアハンドから時計回りに数えて先のプレイヤー) のチームを勝者とする。
func (g *Jass) resolveWeis() {
	if !g.config.EnableWeis || g.weisResolved {
		return
	}
	g.weisResolved = true

	weisByPlayer := make([]jassWeis, JassPlayerCnt)
	for i := range g.players {
		weisByPlayer[i] = computeWeis(g.players[i])
	}

	// 最良 Weis を持つプレイヤーを決定する (フォアハンド順で走査)
	bestPlayer := -1
	var best jassWeis
	for off := 0; off < JassPlayerCnt; off++ {
		idx := (g.forehandIdx + off) % JassPlayerCnt
		w := weisByPlayer[idx]
		if w.bestVal == 0 {
			continue
		}
		if bestPlayer < 0 || weisBeats(w, best) {
			best = w
			bestPlayer = idx
		}
	}

	if bestPlayer < 0 {
		return // 誰も Weis を持たない
	}

	winTeam := g.players[bestPlayer].GetTeam()
	for i := range g.players {
		if g.players[i].GetTeam() == winTeam {
			g.roundWeisPoints[winTeam] += weisByPlayer[i].points
		}
	}
	if g.roundWeisPoints[winTeam] > 0 {
		g.appendLog(bestPlayer, "weis",
			fmt.Sprintf("Team %d wins Weis (+%d)", winTeam, g.roundWeisPoints[winTeam]), nil)
	}
}

// weisBeats a が b に厳密に勝るか (フォアハンド順で先のものを best 初期化に使うため
// 同値タイは false を返し、先着 (フォアハンドに近い) が勝つ)
func weisBeats(a, b jassWeis) bool {
	if a.bestVal != b.bestVal {
		return a.bestVal > b.bestVal
	}
	if a.bestLen != b.bestLen {
		return a.bestLen > b.bestLen
	}
	return a.bestRank > b.bestRank
}

// resolveStock トランプの K+Q を持つプレイヤーのチームへ Stöck (+20) を加点する
func (g *Jass) resolveStock() {
	if g.trumpSuit == 0 {
		return
	}
	for i, p := range g.players {
		hasK := false
		hasQ := false
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			if c.GetDesign() != g.trumpSuit {
				continue
			}
			switch c.GetValue() {
			case 13:
				hasK = true
			case 12:
				hasQ = true
			}
		}
		if hasK && hasQ {
			team := p.GetTeam()
			g.roundStockPoints[team] += JassStockBonus
			g.appendLog(i, "stock",
				fmt.Sprintf("%s has Stöck (+%d for team %d)",
					playerName(g.players, i), JassStockBonus, team), nil)
		}
	}
}

// --- Trick play helpers ---

// startPlayPhase プレイフェーズを開始する
func (g *Jass) startPlayPhase() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = g.forehandIdx
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = JassPhasePlay
}

// playCard カードをプレイする共通処理
func (g *Jass) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == JassPlayerCnt {
		g.phase = JassPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % JassPlayerCnt
	}
}

// validatePlay カードのプレイが Jass のルールに従っているか検証する
// 義務: リードスート可能ならフォローする。フォロー不可なら任意 (トランプ含む)。
// 例外: リードがトランプの場合でも、手札がトランプの J (Bauer) 1枚のみのときは
// フォローを免除される (標準ヤスルール)。簡略化のため、ここでは
// 「フォロー可能ならフォローする」のみを必須とする。
func (g *Jass) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	player := g.players[playerIdx]
	leadSuit := g.currentTrick[0].Card.GetDesign()
	cardSuit := card.GetDesign()

	hasLead := g.playerHasSuit(player, leadSuit)
	if !hasLead {
		// フォロー不可: 任意のカード可 (トランプ含む)
		return nil
	}

	if leadSuit == g.trumpSuit {
		// リードがトランプ: トランプを持つならトランプを出す。
		// ただし手札のトランプが J (Bauer) 1枚のみの場合は任意のカード可。
		if g.onlyTrumpIsJack(player) {
			return nil
		}
		if cardSuit != g.trumpSuit {
			return NewDomainError(ErrInvalidPlay, "リードスート (切り札) に従ってください")
		}
		return nil
	}

	// リードが非トランプ かつ リードスートを持っている: フォロー必須。
	// (フォロー可能なときにトランプで切り上げるのは不可。トランプ切りは
	//  リードスートを持たない=void のときのみ — その分岐は上で処理済み)
	if cardSuit == leadSuit {
		return nil
	}
	return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
}

// onlyTrumpIsJack プレイヤーの手札中のトランプが J (Bauer) 1枚のみか
func (g *Jass) onlyTrumpIsJack(p *JassPlayer) bool {
	trumpCount := 0
	jackOnly := true
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c.GetDesign() != g.trumpSuit {
			continue
		}
		trumpCount++
		if c.GetValue() != 11 {
			jackOnly = false
		}
	}
	return trumpCount == 1 && jackOnly
}

// playerHasSuit プレイヤーが特定スートを持っているか
func (g *Jass) playerHasSuit(p *JassPlayer, suit int) bool {
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == suit {
			return true
		}
	}
	return false
}

// currentLeader 現在のトリック先頭時点での仮勝者を返す
func (g *Jass) currentLeader() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	winner := g.currentTrick[0].PlayerIdx
	winnerRank := g.cardRank(g.currentTrick[0].Card)
	winnerSuit := g.currentTrick[0].Card.GetDesign()
	for _, tc := range g.currentTrick[1:] {
		suit := tc.Card.GetDesign()
		rank := g.cardRank(tc.Card)
		if suit == g.trumpSuit && winnerSuit != g.trumpSuit {
			winner = tc.PlayerIdx
			winnerRank = rank
			winnerSuit = suit
			continue
		}
		if suit == winnerSuit && rank > winnerRank {
			winner = tc.PlayerIdx
			winnerRank = rank
		}
	}
	return winner
}

// trickWinner トリックの勝者を決定する
func (g *Jass) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	return g.currentLeader()
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (g *Jass) GetValidPlayIndices(playerIdx int) []int {
	return g.getValidPlayIndices(playerIdx)
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *Jass) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Game end + bookkeeping ---

func (g *Jass) checkGameEnd() {
	for ti := range JassTeamCnt {
		if g.teamScores[ti] >= g.config.TargetScore {
			g.gameEndFlag = true
			g.phase = JassPhaseGameEnd
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

// sortAllHands 全プレイヤーの手札をソートする (スート → ランク)
func (g *Jass) sortAllHands() {
	for _, p := range g.players {
		jassSortHand(p, g)
	}
}

func jassSortHand(p *JassPlayer, g *Jass) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		si := ci.GetDesign()
		sj := cj.GetDesign()
		if si != sj {
			return si < sj
		}
		return g.cardRank(ci) > g.cardRank(cj)
	})
}

// --- Hints ---

// GetHint 現フェーズのヒントを返す (人間プレイヤー視点)
func (g *Jass) GetHint() *JassHint {
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 {
		return nil
	}
	switch g.phase {
	case JassPhaseBidTrump:
		if g.bidPlayerIdx != humanIdx {
			return nil
		}
		suit, score := g.cpuBestTrump(humanIdx)
		if g.cpuShouldSchieben(score) {
			schieben := true
			return &JassHint{Schieben: &schieben, Reason: "schieben_recommended"}
		}
		return &JassHint{Suit: &suit, Reason: "strategic_trump"}
	case JassPhaseBidPartner:
		if g.bidPlayerIdx != humanIdx {
			return nil
		}
		suit, _ := g.cpuBestTrump(humanIdx)
		return &JassHint{Suit: &suit, Reason: "strategic_trump"}
	case JassPhasePlay:
		if g.currentPlayerIdx != humanIdx {
			return nil
		}
		valid := g.getValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlayChoose(humanIdx, valid)
		return &JassHint{CardIndex: &idx, Reason: g.playHintReason(idx)}
	}
	return nil
}

func (g *Jass) playHintReason(chosenIdx int) string {
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 {
		return ""
	}
	card := g.players[humanIdx].GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		if card.GetDesign() == g.trumpSuit {
			return "lead_trump"
		}
		return "lead_strong"
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == g.trumpSuit {
		return "trump_cut"
	}
	return "discard_weak"
}

// --- CPU AI ---

// cpuBestTrump CPUが評価する最良の切り札スートとそのスコアを返す
func (g *Jass) cpuBestTrump(playerIdx int) (int, int) {
	bestSuit := CardDesignSpade
	bestScore := -1
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		score := g.evalHandForTrump(playerIdx, suit)
		if score > bestScore {
			bestScore = score
			bestSuit = suit
		}
	}
	return bestSuit, bestScore
}

// cpuShouldSchieben スコアが弱ければ Schieben を推奨する (難易度で閾値変動)
func (g *Jass) cpuShouldSchieben(bestScore int) bool {
	threshold := 18
	switch g.config.CpuDifficulty {
	case JassCpuDifficultyEasy:
		threshold = 14
	case JassCpuDifficultyHard:
		threshold = 22
	}
	return bestScore < threshold
}

// evalHandForTrump 仮定したトランプスートに対する手札評価値を返す
// (高い = 強い: トランプ J/9/A、長いトランプ列、外スートの A をボーナス)
func (g *Jass) evalHandForTrump(playerIdx, trumpSuit int) int {
	p := g.players[playerIdx]
	score := 0
	trumpCount := 0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c.GetDesign() == trumpSuit {
			trumpCount++
			switch c.GetValue() {
			case 11: // J (Bauer)
				score += 14
			case 9: // Nell
				score += 10
			case 1: // A
				score += 6
			case 13: // K
				score += 3
			case 12: // Q
				score += 2
			default:
				score++
			}
			continue
		}
		switch c.GetValue() {
		case 1:
			score += 4 // 外スート A
		case 13:
			score += 2
		}
	}
	switch {
	case trumpCount >= 5:
		score += 8
	case trumpCount == 4:
		score += 5
	case trumpCount == 3:
		score += 2
	}
	return score
}

// cpuSelectPlayCard CPUがプレイするカードを選ぶ
func (g *Jass) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	switch g.config.CpuDifficulty {
	case JassCpuDifficultyEasy:
		return valid[rand.Intn(len(valid))]
	default:
		return g.cpuPlayChoose(playerIdx, valid)
	}
}

// cpuPlayChoose 標準ヒューリスティック:
//   - リード時: 強いトランプ (J/9) または高得点の非トランプ A を優先
//   - フォロー時: パートナーが勝者なら高点を載せ、そうでなければ勝てる最弱札、
//     勝てなければ最低点札を捨てる
func (g *Jass) cpuPlayChoose(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		best := valid[0]
		bestScore := -1
		for _, idx := range valid {
			c := player.GetCard(idx)
			s := 0
			if c.GetDesign() == g.trumpSuit {
				switch c.GetValue() {
				case 11:
					s = 25
				case 9:
					s = 15
				}
			} else {
				switch c.GetValue() {
				case 1:
					s = 30
				case 10:
					s = 20
				case 13:
					s = 5
				}
			}
			if s > bestScore {
				bestScore = s
				best = idx
			}
		}
		return best
	}

	winnerIdx := g.currentLeader()
	partnerIdx := (playerIdx + 2) % JassPlayerCnt
	partnerWinning := winnerIdx == partnerIdx

	if partnerWinning {
		best := valid[0]
		bestPts := -1
		for _, idx := range valid {
			pts := jassCardPoints(player.GetCard(idx), g.trumpSuit)
			if pts > bestPts {
				bestPts = pts
				best = idx
			}
		}
		return best
	}

	winnable := -1
	winnableRank := 9999
	for _, idx := range valid {
		c := player.GetCard(idx)
		if g.cardWouldWinTrick(c) {
			r := g.cardRank(c)
			if r < winnableRank {
				winnableRank = r
				winnable = idx
			}
		}
	}
	if winnable >= 0 {
		return winnable
	}

	worst := valid[0]
	worstPts := 9999
	for _, idx := range valid {
		pts := jassCardPoints(player.GetCard(idx), g.trumpSuit)
		if pts < worstPts {
			worstPts = pts
			worst = idx
		}
	}
	return worst
}

// cardWouldWinTrick 指定カードを今出した場合に現状の勝者を上回るか
func (g *Jass) cardWouldWinTrick(c *Card) bool {
	if len(g.currentTrick) == 0 {
		return true
	}
	winIdx := g.currentLeader()
	var winCard *Card
	for _, tc := range g.currentTrick {
		if tc.PlayerIdx == winIdx {
			winCard = tc.Card
			break
		}
	}
	if winCard == nil {
		return true
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	cSuit := c.GetDesign()
	wSuit := winCard.GetDesign()

	if cSuit == g.trumpSuit && wSuit != g.trumpSuit {
		return true
	}
	if cSuit == wSuit {
		return g.cardRank(c) > g.cardRank(winCard)
	}
	if wSuit == g.trumpSuit {
		return false
	}
	if cSuit == leadSuit {
		return g.cardRank(c) > g.cardRank(winCard)
	}
	return false
}

// --- JSON ---

// jassJSON Jass の JSON 表現
type jassJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*JassPlayer     `json:"pl"`
	Config           JassConfig        `json:"cfg"`
	Phase            JassPhase         `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"cp"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	DealerIdx        int               `json:"di"`
	ForehandIdx      int               `json:"fh"`
	TrumpSuit        int               `json:"ts"`
	MakerTeam        int               `json:"mt"`
	MakerPlayerIdx   int               `json:"mp"`
	Schieben         bool              `json:"sb"`
	TeamScores       [JassTeamCnt]int  `json:"sc"`
	RoundPoints      [JassTeamCnt]int  `json:"rp"`
	RoundWeisPoints  [JassTeamCnt]int  `json:"rw"`
	RoundStockPoints [JassTeamCnt]int  `json:"rs"`
	WeisResolved     bool              `json:"wr"`
	LastTrickWinner  int               `json:"lw"`
	LeadPlayerIdx    int               `json:"li"`
	BidPlayerIdx     int               `json:"bi"`
	GameEndFlag      bool              `json:"ge"`
	WinnerTeam       int               `json:"wt"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Jass) MarshalJSON() ([]byte, error) {
	return json.Marshal(jassJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		DealerIdx:        g.dealerIdx,
		ForehandIdx:      g.forehandIdx,
		TrumpSuit:        g.trumpSuit,
		MakerTeam:        g.makerTeam,
		MakerPlayerIdx:   g.makerPlayerIdx,
		Schieben:         g.schieben,
		TeamScores:       g.teamScores,
		RoundPoints:      g.roundPoints,
		RoundWeisPoints:  g.roundWeisPoints,
		RoundStockPoints: g.roundStockPoints,
		WeisResolved:     g.weisResolved,
		LastTrickWinner:  g.lastTrickWinner,
		LeadPlayerIdx:    g.leadPlayerIdx,
		BidPlayerIdx:     g.bidPlayerIdx,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler. 列挙値・インデックス範囲・スライス要素を
// 検証し、不正な場合はエラーを返す (パニックさせない)。
func (g *Jass) UnmarshalJSON(data []byte) error {
	var j jassJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < JassPhaseBidTrump || j.Phase > JassPhaseGameEnd {
		return NewDomainError(ErrInvalidPlay, "無効なフェーズです")
	}
	// trumpSuit=0 (未確定) は許可。確定済みの場合のみ範囲チェック。
	if j.TrumpSuit != 0 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return NewDomainError(ErrInvalidPlay, "無効な切り札スートです")
	}
	if len(j.Players) != JassPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "プレイヤー数が不正です")
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

	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	g.dealerIdx = j.DealerIdx
	g.forehandIdx = j.ForehandIdx
	g.trumpSuit = j.TrumpSuit
	g.makerTeam = j.MakerTeam
	g.makerPlayerIdx = j.MakerPlayerIdx
	g.schieben = j.Schieben
	g.teamScores = j.TeamScores
	g.roundPoints = j.RoundPoints
	g.roundWeisPoints = j.RoundWeisPoints
	g.roundStockPoints = j.RoundStockPoints
	g.weisResolved = j.WeisResolved
	g.lastTrickWinner = j.LastTrickWinner
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.bidPlayerIdx = j.BidPlayerIdx
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	return nil
}
