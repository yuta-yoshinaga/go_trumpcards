//go:build !js || !wasm || extra4

// Package domain シープスヘッド (Sheepshead) のドメインモデル。
//
// Sheepshead はドイツの Schafkopf を起源とするアメリカ中西部のトリックテイキング
// ゲーム。全クイーン・全ジャック・ダイヤ全札が固定切り札となる独自のランク体系を
// 持ち、5 人中 1 人の「ピッカー」が秘密の相棒（呼びカードの保持者）と組んで 120 点
// のカードポイントを争う。
//
// 切り札 (強い順): Q♣ > Q♠ > Q♥ > Q♦ > J♣ > J♠ > J♥ > J♦ > A♦ > 10♦ > K♦ > 9♦ > 8♦ > 7♦
// フェイル札 (各スート, 強い順): A > 10 > K > 9 > 8 > 7
// カードポイント: A=11, 10=10, K=4, Q=3, J=2, 9/8/7=0 (合計 120)
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// SheepsheadPlayerCnt プレイヤー数 (人間 1 + CPU 4)
const SheepsheadPlayerCnt = 5

// SheepsheadHandSize 各プレイヤーへ配る手札枚数
const SheepsheadHandSize = 6

// SheepsheadBlindSize ブラインド (場に伏せる札) の枚数
const SheepsheadBlindSize = 2

// SheepsheadBurySize ピッカーが埋める札の枚数
const SheepsheadBurySize = 2

// SheepsheadTrickCount 1 ラウンドのトリック数
const SheepsheadTrickCount = 6

// SheepsheadTotalPoints 1 ラウンドのカードポイント合計
const SheepsheadTotalPoints = 120

// SheepsheadPhase ゲームフェーズ
type SheepsheadPhase int

// Sheepshead のフェーズ定数
const (
	// SheepsheadPhasePick ピック (ブラインドを取るか降りるか) フェーズ
	SheepsheadPhasePick SheepsheadPhase = 0
	// SheepsheadPhaseBury ピッカーが 2 枚を埋めるフェーズ
	SheepsheadPhaseBury SheepsheadPhase = 1
	// SheepsheadPhaseCall ピッカーが相棒となる呼びカード (フェイル A) を指定するフェーズ
	SheepsheadPhaseCall SheepsheadPhase = 2
	// SheepsheadPhasePlay トリックプレイフェーズ
	SheepsheadPhasePlay SheepsheadPhase = 3
	// SheepsheadPhaseTrickEnd トリック終了フェーズ (解決済み・次トリック待ち)
	SheepsheadPhaseTrickEnd SheepsheadPhase = 4
	// SheepsheadPhaseRoundEnd ラウンド終了フェーズ
	SheepsheadPhaseRoundEnd SheepsheadPhase = 5
	// SheepsheadPhaseGameEnd ゲーム終了フェーズ
	SheepsheadPhaseGameEnd SheepsheadPhase = 6
)

// SheepsheadHint ヒント情報
type SheepsheadHint struct {
	CardIndices []int  // 推奨カードインデックス (プレイ・埋めフェーズ)
	Suit        int    // 推奨呼びスート (呼びフェーズ, それ以外は 0)
	Pick        bool   // 推奨ピック判断 (ピックフェーズ)
	Reason      string // ヒント理由キー
}

// Sheepshead シープスヘッドのゲームクラス
type Sheepshead struct {
	trumpCards       *TrumpCards
	players          []*SheepsheadPlayer
	config           SheepsheadConfig
	phase            SheepsheadPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	blind            []*Card // 場に伏せられた札 (ピック前)
	buried           []*Card // ピッカーが埋めた札 (得点はピッカー組へ)
	passCount        int     // 現ピックフェーズでパスした人数
	pickerIdx        int     // ピッカー (-1 = 未確定)
	partnerIdx       int     // 相棒 (-1 = 単独 or 未確定)
	calledSuit       int     // 呼びスート (0 = 未確定/単独)
	partnerRevealed  bool    // 呼びカードがプレイされ相棒が判明したか
	roundPickerPts   int     // 直近ラウンドのピッカー組得点
	roundMultiplier  int     // 直近ラウンドの倍率 (1/2/3)
	roundPickerWon   bool    // 直近ラウンドでピッカー組が勝ったか
	gameEndFlag      bool
	winnerIdx        int // ゲーム勝者 (-1 = 未確定)
	actionLogBase
}

// NewSheepshead コンストラクタ
func NewSheepshead(trumpCards *TrumpCards, players []*SheepsheadPlayer, config SheepsheadConfig) *Sheepshead {
	return &Sheepshead{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		pickerIdx:  -1,
		partnerIdx: -1,
		winnerIdx:  -1,
	}
}

// NewDefaultSheepshead 標準の 5 人構成 (人間 1, CPU 4) と既定設定で生成する。
// CUI / Web / Worker 構築の単一の真実源。
func NewDefaultSheepshead() *Sheepshead {
	cfg := DefaultSheepsheadConfig()
	players := make([]*SheepsheadPlayer, SheepsheadPlayerCnt)
	players[0] = NewSheepsheadPlayer(true, cfg.StartChips)
	for i := 1; i < SheepsheadPlayerCnt; i++ {
		players[i] = NewSheepsheadPlayer(false, cfg.StartChips)
	}
	return NewSheepshead(NewTrumpCardsBelote(), players, cfg)
}

// Reset ゲーム初期化: チップを開始値へ戻し最初のラウンドを開始する。
func (g *Sheepshead) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	for _, p := range g.players {
		p.SetChips(g.config.StartChips)
	}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Sheepshead) NextRound() {
	if g.phase != SheepsheadPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % SheepsheadPlayerCnt
	g.startRound()
}

// startRound 手札・ブラインドを配り、ピックフェーズを開始する。
func (g *Sheepshead) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.buried = nil
	g.pickerIdx = -1
	g.partnerIdx = -1
	g.calledSuit = 0
	g.partnerRevealed = false
	g.passCount = 0
	g.roundPickerPts = 0
	g.roundMultiplier = 0
	g.roundPickerWon = false

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()

	// ピックはディーラーの左隣から始まる。
	g.leadPlayerIdx = (g.dealerIdx + 1) % SheepsheadPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = SheepsheadPhasePick
}

// deal 各プレイヤーへ 6 枚、ブラインドへ 2 枚を配る。
func (g *Sheepshead) deal() {
	for i := 0; i < SheepsheadHandSize; i++ {
		for _, p := range g.players {
			if c := g.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}
	g.blind = make([]*Card, 0, SheepsheadBlindSize)
	for i := 0; i < SheepsheadBlindSize; i++ {
		if c := g.trumpCards.DrawCard(); c != nil {
			g.blind = append(g.blind, c)
		}
	}
}

// PlayerPick 人間プレイヤーがピック (true) かパス (false) を選択する。
func (g *Sheepshead) PlayerPick(pick bool) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SheepsheadPhasePick {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !pick && g.passCount >= SheepsheadPlayerCnt-1 {
		return NewDomainError(ErrInvalidPlay, "最後のプレイヤーはパスできません")
	}
	g.resolvePick(g.currentPlayerIdx, pick)
	return nil
}

// resolvePick ピック/パスを反映し、フェーズを進める。
func (g *Sheepshead) resolvePick(playerIdx int, pick bool) {
	// 4 人がパスした場合、最後のプレイヤーは強制的にピックする。
	if !pick && g.passCount >= SheepsheadPlayerCnt-1 {
		pick = true
	}
	if pick {
		g.becomePicker(playerIdx)
		return
	}
	g.passCount++
	g.appendLog(playerIdx, "pass", fmt.Sprintf("%s passes", playerName(g.players, playerIdx)), nil)
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % SheepsheadPlayerCnt
}

// becomePicker ピッカーを確定し、ブラインドを手札へ加えて埋めフェーズへ移る。
func (g *Sheepshead) becomePicker(playerIdx int) {
	g.pickerIdx = playerIdx
	picker := g.players[playerIdx]
	for _, c := range g.blind {
		picker.AddCard(c)
	}
	g.blind = nil
	sheepsheadSortHand(picker)
	g.appendLog(playerIdx, "pick", fmt.Sprintf("%s picks up the blind", playerName(g.players, playerIdx)), nil)
	g.currentPlayerIdx = playerIdx
	g.phase = SheepsheadPhaseBury
}

// PlayerBury 人間ピッカーが 2 枚を埋める。
func (g *Sheepshead) PlayerBury(indices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SheepsheadPhaseBury {
		return ErrWrongPhase
	}
	if g.currentPlayerIdx != g.pickerIdx || !g.players[g.pickerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if err := g.validateBury(indices); err != nil {
		return err
	}
	g.applyBury(indices)
	return nil
}

// validateBury 埋め札インデックスの妥当性を検証する。
func (g *Sheepshead) validateBury(indices []int) error {
	if len(indices) != SheepsheadBurySize {
		return NewDomainError(ErrInvalidPlay, "埋める札はちょうど 2 枚です")
	}
	picker := g.players[g.pickerIdx]
	seen := make(map[int]bool, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= picker.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidPlay, "同じ札を 2 回指定できません")
		}
		seen[idx] = true
	}
	return nil
}

// applyBury 埋め札を手札から取り除き、呼びフェーズ (または単独プレイ) へ進む。
func (g *Sheepshead) applyBury(indices []int) {
	picker := g.players[g.pickerIdx]
	g.buried = picker.RemoveCards(indices)
	g.appendLog(g.pickerIdx, "bury", fmt.Sprintf("%s buries %d cards", playerName(g.players, g.pickerIdx), len(g.buried)), nil)

	if len(g.callableSuits()) == 0 {
		// 呼べる札がない: ピッカーは単独で戦う。
		g.partnerIdx = -1
		g.appendLog(g.pickerIdx, "alone", fmt.Sprintf("%s plays alone", playerName(g.players, g.pickerIdx)), nil)
		g.beginPlay()
		return
	}
	g.phase = SheepsheadPhaseCall
}

// PlayerCall 人間ピッカーが呼びスートを指定する。
func (g *Sheepshead) PlayerCall(suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SheepsheadPhaseCall {
		return ErrWrongPhase
	}
	if g.currentPlayerIdx != g.pickerIdx || !g.players[g.pickerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !g.isCallableSuit(suit) {
		return NewDomainError(ErrInvalidPlay, "そのスートは呼べません")
	}
	g.applyCall(suit)
	return nil
}

// applyCall 呼びスートを確定し、相棒を決定してプレイフェーズへ進む。
func (g *Sheepshead) applyCall(suit int) {
	g.calledSuit = suit
	g.partnerIdx = g.holderOfCalledAce(suit)
	g.appendLog(g.pickerIdx, "call",
		fmt.Sprintf("%s calls the %s Ace", playerName(g.players, g.pickerIdx), suitStr(suit)), nil)
	g.beginPlay()
}

// beginPlay プレイフェーズを開始する。リードはピックフェーズの先頭プレイヤー。
func (g *Sheepshead) beginPlay() {
	g.leadPlayerIdx = (g.dealerIdx + 1) % SheepsheadPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber = 1
	g.currentTrick = nil
	g.phase = SheepsheadPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Sheepshead) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SheepsheadPhasePlay {
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

// CpuPlay 現在の手番が CPU の場合に 1 アクション実行する。フェーズに応じて
// ピック / 埋め / 呼び / プレイ を処理する。
func (g *Sheepshead) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	switch g.phase {
	case SheepsheadPhasePick:
		if g.players[g.currentPlayerIdx].GetIsHuman() {
			return
		}
		g.resolvePick(g.currentPlayerIdx, g.cpuDecidePick(g.currentPlayerIdx))
	case SheepsheadPhaseBury:
		if g.pickerIdx < 0 || g.players[g.pickerIdx].GetIsHuman() {
			return
		}
		g.applyBury(g.cpuSelectBury(g.pickerIdx))
	case SheepsheadPhaseCall:
		if g.pickerIdx < 0 || g.players[g.pickerIdx].GetIsHuman() {
			return
		}
		g.applyCall(g.cpuSelectCall(g.pickerIdx))
	case SheepsheadPhasePlay:
		if g.players[g.currentPlayerIdx].GetIsHuman() {
			return
		}
		player := g.players[g.currentPlayerIdx]
		idx := g.cpuSelectPlayCard(g.currentPlayerIdx)
		played := player.RemoveCard(idx)
		// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
		// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
		// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
		if played == nil {
			return
		}
		g.playCard(g.currentPlayerIdx, played)
	default:
		// 解決待ちのフェーズでは何もしない。
	}
}

// playCard カードをプレイする共通処理。
func (g *Sheepshead) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	// 呼びカードがプレイされたら相棒が判明する。
	if g.calledSuit != 0 && !g.partnerRevealed &&
		card.GetValue() == 1 && card.GetDesign() == g.calledSuit && !sheepsheadIsTrump(card) {
		g.partnerRevealed = true
		g.appendLog(playerIdx, "partner_reveal",
			fmt.Sprintf("%s is the picker's partner", playerName(g.players, playerIdx)), nil)
	}

	if len(g.currentTrick) == SheepsheadPlayerCnt {
		g.phase = SheepsheadPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % SheepsheadPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。
func (g *Sheepshead) ResolveTrick() {
	if g.phase != SheepsheadPhaseTrickEnd || len(g.currentTrick) != SheepsheadPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += sheepsheadCardPoints(tc.Card.GetValue())
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pts)", playerName(g.players, winnerIdx), g.trickNumber, pts), trickCards)

	g.leadPlayerIdx = winnerIdx
	// Clear the resolved trick so a spurious second ResolveTrick call cannot
	// double-count its points (defensive — NextTrick also clears it).
	g.currentTrick = nil
	if g.trickNumber >= SheepsheadTrickCount {
		g.phase = SheepsheadPhaseRoundEnd
	} else {
		g.phase = SheepsheadPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Sheepshead) NextTrick() {
	if g.phase != SheepsheadPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = SheepsheadPhasePlay
}

// ScoreRound ラウンドの得点を確定し、チップ精算とゲーム終了判定を行う。
func (g *Sheepshead) ScoreRound() {
	if g.phase != SheepsheadPhaseRoundEnd {
		return
	}

	pickerPts := g.pickerTeamPoints()
	defenderPts := SheepsheadTotalPoints - pickerPts
	pickerWon := pickerPts >= 61
	loserPts := defenderPts
	if !pickerWon {
		loserPts = pickerPts
	}
	mult := sheepsheadMultiplier(loserPts, g.loserTookNoTrick(pickerWon))

	g.roundPickerPts = pickerPts
	g.roundMultiplier = mult
	g.roundPickerWon = pickerWon
	g.settleChips(pickerWon, mult)

	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: picker team %d pts (%s, x%d)",
			g.roundNumber, pickerPts, sheepsheadOutcomeStr(pickerWon), mult), nil)

	if w := g.chipLeaderAtTarget(); w >= 0 {
		g.gameEndFlag = true
		g.winnerIdx = w
		g.phase = SheepsheadPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, w)), nil)
	}
}

// settleChips チップ精算 (ゼロサム)。
//
// ピッカー組勝利時: 各ディフェンダーが unit*mult を支払い、相棒ありなら
// ピッカー 2 : 相棒 1 の比で受け取る。単独なら全額ピッカーへ。敗北時は符号反転。
func (g *Sheepshead) settleChips(pickerWon bool, mult int) {
	unit := g.config.BaseChips * mult
	sign := 1
	if !pickerWon {
		sign = -1
	}
	for i := range g.players {
		if g.isPickerTeam(i) {
			continue
		}
		// ディフェンダー: 敗北側なら支払い、勝利側なら受け取り。
		g.players[i].AddChips(-sign * unit)
	}
	if g.partnerIdx >= 0 {
		g.players[g.pickerIdx].AddChips(sign * unit * 2)
		g.players[g.partnerIdx].AddChips(sign * unit)
	} else {
		g.players[g.pickerIdx].AddChips(sign * unit * 4)
	}
}

// --- Scoring helpers ---

// pickerTeamPoints ピッカー組 (ピッカー + 相棒 + 埋め札) の獲得カードポイント。
func (g *Sheepshead) pickerTeamPoints() int {
	pts := 0
	for i := range g.players {
		if g.isPickerTeam(i) {
			pts += sheepsheadTrickPoints(g.players[i].GetTricksTaken())
		}
	}
	for _, c := range g.buried {
		pts += sheepsheadCardPoints(c.GetValue())
	}
	return pts
}

// loserTookNoTrick 敗北側が 1 トリックも取れなかったか (シュバルツ判定用)。
func (g *Sheepshead) loserTookNoTrick(pickerWon bool) bool {
	for i := range g.players {
		onPickerTeam := g.isPickerTeam(i)
		loserSide := onPickerTeam != pickerWon // 敗北側のプレイヤー
		if loserSide && g.players[i].GetTrickCount() > 0 {
			return false
		}
	}
	return true
}

// chipLeaderAtTarget 目標チップに到達した最上位プレイヤーを返す (-1 = なし)。
func (g *Sheepshead) chipLeaderAtTarget() int {
	best, bestChips := -1, g.config.TargetChips-1
	for i, p := range g.players {
		if p.GetChips() >= g.config.TargetChips && p.GetChips() > bestChips {
			best, bestChips = i, p.GetChips()
		}
	}
	return best
}

// isPickerTeam playerIdx がピッカー組 (ピッカー or 相棒) か。
func (g *Sheepshead) isPickerTeam(playerIdx int) bool {
	return playerIdx == g.pickerIdx || (g.partnerIdx >= 0 && playerIdx == g.partnerIdx)
}

// --- Trick / play helpers ---

// validatePlay マストフォロー (リードスートに従う) を検証する。
func (g *Sheepshead) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := sheepsheadSuitID(g.currentTrick[0].Card)
	if sheepsheadSuitID(card) != leadSuit && g.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーがスート ID (切り札含む) のカードを持っているか。
func (g *Sheepshead) playerHasSuit(playerIdx, suitID int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if sheepsheadSuitID(p.GetCard(i)) == suitID {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、なければ
// リードスートの最強札が勝つ。
func (g *Sheepshead) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := sheepsheadSuitID(g.currentTrick[0].Card)
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerStrength := sheepsheadStrength(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		// 勝負に絡むのは「切り札」または「リードスートに従った札」のみ。
		// フェイルがリードされても切り札を出せば勝てる (切り札 > フェイル)。
		if !sheepsheadIsTrump(tc.Card) && sheepsheadSuitID(tc.Card) != leadSuit {
			continue
		}
		if s := sheepsheadStrength(tc.Card); s > winnerStrength {
			winnerIdx = tc.PlayerIdx
			winnerStrength = s
		}
	}
	return winnerIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Sheepshead) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// --- Call helpers ---

// callableSuits ピッカーが呼べるフェイル A のスート一覧を返す。ピッカーが
// 既に持つフェイル A は呼べない。
func (g *Sheepshead) callableSuits() []int {
	if g.pickerIdx < 0 {
		return nil
	}
	var suits []int
	for _, suit := range sheepsheadFailSuits() {
		if !g.pickerHoldsAce(suit) && g.holderOfCalledAce(suit) >= 0 {
			suits = append(suits, suit)
		}
	}
	return suits
}

// isCallableSuit 指定スートが呼び可能か。
func (g *Sheepshead) isCallableSuit(suit int) bool {
	for _, s := range g.callableSuits() {
		if s == suit {
			return true
		}
	}
	return false
}

// pickerHoldsAce ピッカーが指定フェイルスートの A を持っているか。
func (g *Sheepshead) pickerHoldsAce(suit int) bool {
	picker := g.players[g.pickerIdx]
	for i := 0; i < picker.GetCardsSize(); i++ {
		c := picker.GetCard(i)
		if c.GetValue() == 1 && c.GetDesign() == suit && !sheepsheadIsTrump(c) {
			return true
		}
	}
	return false
}

// holderOfCalledAce 指定フェイルスートの A を手札に持つプレイヤーを返す (-1=なし)。
func (g *Sheepshead) holderOfCalledAce(suit int) int {
	for i, p := range g.players {
		if i == g.pickerIdx {
			continue
		}
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c.GetValue() == 1 && c.GetDesign() == suit && !sheepsheadIsTrump(c) {
				return i
			}
		}
	}
	return -1
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Sheepshead) sortAllHands() {
	for _, p := range g.players {
		sheepsheadSortHand(p)
	}
}

// sheepsheadSortHand 手札を切り札を先頭に、スートごと強い順にソートする。
func sheepsheadSortHand(p *SheepsheadPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		si, sj := sheepsheadSuitID(ci), sheepsheadSuitID(cj)
		if si != sj {
			return si < sj
		}
		return sheepsheadStrength(ci) > sheepsheadStrength(cj)
	})
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Sheepshead) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// trickTopStrength 現在のトリック勝者 winnerIdx の札の強さを返す。防御的に、
// 勝者がトリック内に見つからない場合は極小値を返す (パニック回避)。
func (g *Sheepshead) trickTopStrength(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return sheepsheadStrength(g.currentTrick[idx].Card)
}

// --- Card classification ---

// sheepsheadIsTrump 切り札 (全 Q, 全 J, ダイヤ全札) か。
func sheepsheadIsTrump(card *Card) bool {
	return card.GetDesign() == CardDesignDiamond || card.GetValue() == 11 || card.GetValue() == 12
}

// sheepsheadSuitID トリック上のスート ID を返す。切り札は共通の ID
// (sheepsheadTrumpSuit) を持ち、フェイル札はスート定数をそのまま返す。
func sheepsheadSuitID(card *Card) int {
	if sheepsheadIsTrump(card) {
		return sheepsheadTrumpSuit
	}
	return card.GetDesign()
}

// sheepsheadTrumpSuit 切り札を表すスート ID。
const sheepsheadTrumpSuit = 0

// sheepsheadFailSuits フェイル (非切り札) スートの一覧。
func sheepsheadFailSuits() []int {
	return []int{CardDesignClover, CardDesignSpade, CardDesignHeart}
}

// sheepsheadStrength トリックでの強さ。切り札はすべてフェイル札より強い。
//
//	Q♣>Q♠>Q♥>Q♦ > J♣>J♠>J♥>J♦ > A♦>10♦>K♦>9♦>8♦>7♦ > (フェイル) A>10>K>9>8>7
func sheepsheadStrength(card *Card) int {
	const trumpBase = 100
	v := card.GetValue()
	if v == 12 { // Queen
		return trumpBase + 30 + sheepsheadTrumpSuitOrder(card.GetDesign())
	}
	if v == 11 { // Jack
		return trumpBase + 20 + sheepsheadTrumpSuitOrder(card.GetDesign())
	}
	if card.GetDesign() == CardDesignDiamond { // diamond trump (non Q/J)
		return trumpBase + sheepsheadFailRank(v)
	}
	return sheepsheadFailRank(v)
}

// sheepsheadTrumpSuitOrder Q/J の切り札内スート順位 (♣>♠>♥>♦)。
func sheepsheadTrumpSuitOrder(design int) int {
	switch design {
	case CardDesignClover:
		return 3
	case CardDesignSpade:
		return 2
	case CardDesignHeart:
		return 1
	default: // Diamond
		return 0
	}
}

// sheepsheadFailRank フェイル/ダイヤ非絵札の強さ順位。A>10>K>9>8>7。
func sheepsheadFailRank(value int) int {
	switch value {
	case 1: // Ace
		return 5
	case 10:
		return 4
	case 13: // King
		return 3
	case 9:
		return 2
	case 8:
		return 1
	default: // 7
		return 0
	}
}

// sheepsheadCardPoints カードポイント。A=11,10=10,K=4,Q=3,J=2,その他=0。
func sheepsheadCardPoints(value int) int {
	switch value {
	case 1:
		return 11
	case 10:
		return 10
	case 13:
		return 4
	case 12:
		return 3
	case 11:
		return 2
	default:
		return 0
	}
}

// sheepsheadTrickPoints 取得トリック群の合計カードポイント。
func sheepsheadTrickPoints(tricks [][]*Card) int {
	pts := 0
	for _, t := range tricks {
		for _, c := range t {
			pts += sheepsheadCardPoints(c.GetValue())
		}
	}
	return pts
}

// sheepsheadMultiplier チップ倍率。敗北側 0 トリック(シュバルツ)=3、
// 30 点以下(シュナイダー)=2、それ以外=1。
func sheepsheadMultiplier(loserPoints int, loserNoTrick bool) int {
	if loserNoTrick {
		return 3
	}
	if loserPoints <= 30 {
		return 2
	}
	return 1
}

// sheepsheadOutcomeStr 勝敗の表示文字列。
func sheepsheadOutcomeStr(pickerWon bool) string {
	if pickerWon {
		return "picker team wins"
	}
	return "defenders win"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Sheepshead) GetPhase() SheepsheadPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Sheepshead) SetPhase(phase SheepsheadPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Sheepshead) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Sheepshead) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (g *Sheepshead) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Sheepshead) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Sheepshead) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Sheepshead) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Sheepshead) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Sheepshead) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Sheepshead) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Sheepshead) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Sheepshead) GetDealerIdx() int { return g.dealerIdx }

// GetBlind ブラインド (ピック前の伏せ札) を取得
func (g *Sheepshead) GetBlind() []*Card { return g.blind }

// GetBuried 埋め札を取得
func (g *Sheepshead) GetBuried() []*Card { return g.buried }

// GetPickerIdx ピッカーのインデックス取得 (-1=未確定)
func (g *Sheepshead) GetPickerIdx() int { return g.pickerIdx }

// SetPickerIdx ピッカー設定 (テスト用)
func (g *Sheepshead) SetPickerIdx(idx int) { g.pickerIdx = idx }

// GetPartnerIdx 相棒のインデックス取得 (-1=単独/未確定)
func (g *Sheepshead) GetPartnerIdx() int { return g.partnerIdx }

// SetPartnerIdx 相棒設定 (テスト用)
func (g *Sheepshead) SetPartnerIdx(idx int) { g.partnerIdx = idx }

// GetCalledSuit 呼びスート取得 (0=未確定/単独)
func (g *Sheepshead) GetCalledSuit() int { return g.calledSuit }

// IsPartnerRevealed 相棒が判明済みか
func (g *Sheepshead) IsPartnerRevealed() bool { return g.partnerRevealed }

// GetPassCount 現ピックフェーズのパス人数取得
func (g *Sheepshead) GetPassCount() int { return g.passCount }

// GetRoundPickerPoints 直近ラウンドのピッカー組得点取得
func (g *Sheepshead) GetRoundPickerPoints() int { return g.roundPickerPts }

// GetRoundMultiplier 直近ラウンドの倍率取得
func (g *Sheepshead) GetRoundMultiplier() int { return g.roundMultiplier }

// GetRoundPickerWon 直近ラウンドでピッカー組が勝ったか取得
func (g *Sheepshead) GetRoundPickerWon() bool { return g.roundPickerWon }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Sheepshead) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1=未確定)
func (g *Sheepshead) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Sheepshead) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Sheepshead) GetPlayer(i int) *SheepsheadPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか。
func (g *Sheepshead) IsHumanTurn() bool {
	if g.phase == SheepsheadPhaseBury || g.phase == SheepsheadPhaseCall {
		return g.pickerIdx >= 0 && g.players[g.pickerIdx].GetIsHuman()
	}
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Sheepshead) GetConfig() SheepsheadConfig { return g.config }

// SetConfig 設定変更
func (g *Sheepshead) SetConfig(cfg SheepsheadConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Sheepshead) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	if g.phase != SheepsheadPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// GetCallableSuits ピッカーが呼べるフェイルスート一覧を返す。
func (g *Sheepshead) GetCallableSuits() []int { return g.callableSuits() }

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *Sheepshead) GetHint() *SheepsheadHint {
	human := findHumanIdx(g.players)
	if human < 0 {
		return nil
	}
	switch g.phase {
	case SheepsheadPhasePick:
		if g.currentPlayerIdx != human {
			return nil
		}
		pick := g.cpuDecidePick(human)
		reason := "pick_pass"
		if pick {
			reason = "pick_take"
		}
		return &SheepsheadHint{Pick: pick, Reason: reason}
	case SheepsheadPhaseBury:
		if g.pickerIdx != human {
			return nil
		}
		return &SheepsheadHint{CardIndices: g.cpuSelectBury(human), Reason: "bury_low"}
	case SheepsheadPhaseCall:
		if g.pickerIdx != human {
			return nil
		}
		return &SheepsheadHint{Suit: g.cpuSelectCall(human), Reason: "call_suit"}
	case SheepsheadPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &SheepsheadHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	default:
		return nil
	}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Sheepshead) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := sheepsheadSuitID(g.currentTrick[0].Card)
	if sheepsheadSuitID(card) != leadSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	topStrength := g.trickTopStrength(winnerIdx)
	if sheepsheadStrength(card) > topStrength {
		return "follow_win"
	}
	return "follow_duck"
}

// --- CPU AI ---

// cpuDecidePick CPU がピックするか判断する。切り札の枚数とクイーンで評価。
func (g *Sheepshead) cpuDecidePick(playerIdx int) bool {
	player := g.players[playerIdx]
	trumpCnt, queenCnt := 0, 0
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if sheepsheadIsTrump(c) {
			trumpCnt++
		}
		if c.GetValue() == 12 {
			queenCnt++
		}
	}
	if g.config.CpuDifficulty == SheepsheadCpuDifficultyEasy {
		return trumpCnt >= 4
	}
	return trumpCnt >= 3 || (trumpCnt >= 2 && queenCnt >= 1)
}

// cpuSelectBury CPU が埋める 2 枚のインデックスを選ぶ。得点の高い非切り札の
// 短いスートを優先的に埋め、切り札と A は温存する。
func (g *Sheepshead) cpuSelectBury(playerIdx int) []int {
	player := g.players[playerIdx]
	type cand struct {
		idx   int
		score int
	}
	cands := make([]cand, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		// 切り札は埋めたくない (低スコア), 非切り札の高得点札を優先 (高スコア)。
		score := sheepsheadCardPoints(c.GetValue())
		if sheepsheadIsTrump(c) {
			score -= 100
		}
		cands = append(cands, cand{i, score})
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].score > cands[j].score })
	out := []int{cands[0].idx, cands[1].idx}
	sort.Ints(out)
	return out
}

// cpuSelectCall CPU が呼ぶスートを選ぶ。手札に当該スートのフェイル札を持つ
// スートを優先する。
func (g *Sheepshead) cpuSelectCall(playerIdx int) int {
	callable := g.callableSuits()
	if len(callable) == 0 {
		return 0
	}
	player := g.players[playerIdx]
	for _, suit := range callable {
		for i := 0; i < player.GetCardsSize(); i++ {
			c := player.GetCard(i)
			if c.GetDesign() == suit && !sheepsheadIsTrump(c) {
				return suit
			}
		}
	}
	return callable[0]
}

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Sheepshead) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == SheepsheadCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点とチーム関係を意識した戦略プレイ。
func (g *Sheepshead) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]

	// リード: 得点・強さの低い札を出して温存する。
	if len(g.currentTrick) == 0 {
		return pickLowest(player, valid, func(c *Card) int {
			return sheepsheadCardPoints(c.GetValue())*100 + sheepsheadStrength(c)
		})
	}

	leadSuit := sheepsheadSuitID(g.currentTrick[0].Card)
	winnerIdx := g.trickWinner()
	topStrength := g.trickTopStrength(winnerIdx)
	partnerWinning := g.cpuSameTeam(playerIdx, winnerIdx)
	trickPts := 0
	for _, tc := range g.currentTrick {
		trickPts += sheepsheadCardPoints(tc.Card.GetValue())
	}

	var follows []int
	for _, idx := range valid {
		if sheepsheadSuitID(player.GetCard(idx)) == leadSuit {
			follows = append(follows, idx)
		}
	}

	if len(follows) == 0 {
		// ボイド: 味方が勝っていれば得点札を渡し、そうでなければ低得点札を捨てる。
		if partnerWinning {
			return pickHighest(player, valid, func(c *Card) int {
				if sheepsheadIsTrump(c) {
					return -sheepsheadStrength(c) // 切り札は温存
				}
				return sheepsheadCardPoints(c.GetValue())*100 - sheepsheadStrength(c)
			})
		}
		return pickLowest(player, valid, func(c *Card) int {
			return sheepsheadCardPoints(c.GetValue())*100 + sheepsheadStrength(c)
		})
	}

	winners := filterIndices(follows, func(idx int) bool {
		return sheepsheadStrength(player.GetCard(idx)) > topStrength
	})

	if partnerWinning {
		nonWinners := filterIndices(follows, func(idx int) bool {
			return sheepsheadStrength(player.GetCard(idx)) < topStrength
		})
		if len(nonWinners) > 0 {
			return pickHighest(player, nonWinners, func(c *Card) int {
				return sheepsheadCardPoints(c.GetValue())*100 - sheepsheadStrength(c)
			})
		}
		return pickLowest(player, follows, func(c *Card) int { return sheepsheadStrength(c) })
	}

	if trickPts > 0 && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return sheepsheadStrength(c) })
	}
	return pickLowest(player, follows, func(c *Card) int {
		return sheepsheadCardPoints(c.GetValue())*100 + sheepsheadStrength(c)
	})
}

// cpuSameTeam CPU 視点で 2 プレイヤーが同じ組か。相棒が未判明の場合でも、
// CPU はチーム構成を把握しているものとして扱う (簡易 AI)。
func (g *Sheepshead) cpuSameTeam(a, b int) bool {
	return g.isPickerTeam(a) == g.isPickerTeam(b)
}

// --- JSON ---

// sheepsheadJSON is the JSON wire format for Sheepshead.
type sheepsheadJSON struct {
	TrumpCards       *TrumpCards         `json:"tc"`
	Players          []*SheepsheadPlayer `json:"ps"`
	Config           SheepsheadConfig    `json:"cf"`
	Phase            SheepsheadPhase     `json:"ph"`
	RoundNumber      int                 `json:"rn"`
	TrickNumber      int                 `json:"tn"`
	CurrentPlayerIdx int                 `json:"ci"`
	CurrentTrick     []*TrickCard        `json:"ct"`
	LeadPlayerIdx    int                 `json:"li"`
	DealerIdx        int                 `json:"di"`
	Blind            []*Card             `json:"bl"`
	Buried           []*Card             `json:"bu"`
	PassCount        int                 `json:"pc"`
	PickerIdx        int                 `json:"pk"`
	PartnerIdx       int                 `json:"pt"`
	CalledSuit       int                 `json:"cs"`
	PartnerRevealed  bool                `json:"pr"`
	RoundPickerPts   int                 `json:"rp"`
	RoundMultiplier  int                 `json:"rm"`
	RoundPickerWon   bool                `json:"rw"`
	GameEndFlag      bool                `json:"ge"`
	WinnerIdx        int                 `json:"wi"`
	ActionLog        []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Sheepshead) MarshalJSON() ([]byte, error) {
	return json.Marshal(sheepsheadJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		Blind:            g.blind,
		Buried:           g.buried,
		PassCount:        g.passCount,
		PickerIdx:        g.pickerIdx,
		PartnerIdx:       g.partnerIdx,
		CalledSuit:       g.calledSuit,
		PartnerRevealed:  g.partnerRevealed,
		RoundPickerPts:   g.roundPickerPts,
		RoundMultiplier:  g.roundMultiplier,
		RoundPickerWon:   g.roundPickerWon,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		ActionLog:        g.actionLog,
	})
}

// sheepsheadMaxSliceLen caps slice sizes during deserialisation.
const sheepsheadMaxSliceLen = 1000

// errSheepsheadOversized is the single sentinel error for oversized input arrays.
var errSheepsheadOversized = errors.New("sheepshead: input array exceeds maximum allowed size")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Sheepshead) UnmarshalJSON(data []byte) error {
	var j sheepsheadJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > sheepsheadMaxSliceLen || len(j.CurrentTrick) > sheepsheadMaxSliceLen ||
		len(j.ActionLog) > sheepsheadMaxSliceLen {
		return errSheepsheadOversized
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsBelote()
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*SheepsheadPlayer, 0)
	}
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.blind = j.Blind
	g.buried = j.Buried
	g.passCount = j.PassCount
	g.pickerIdx = j.PickerIdx
	g.partnerIdx = j.PartnerIdx
	g.calledSuit = j.CalledSuit
	g.partnerRevealed = j.PartnerRevealed
	g.roundPickerPts = j.RoundPickerPts
	g.roundMultiplier = j.RoundMultiplier
	g.roundPickerWon = j.RoundPickerWon
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
