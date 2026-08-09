//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
)

// MaoPlayerCnt マオプレイヤー数
const MaoPlayerCnt = 4

// MaoHandSize 初期配布枚数
const MaoHandSize = 5

// MaoWildValue ワイルドカード値 (8: スート変更)
const MaoWildValue = 8

// MaoDrawTwoValue ドローツーのカード値 (2: 次のプレイヤーに2枚引かせる/重ね可)
const MaoDrawTwoValue = 2

// MaoSkipValue スキップのカード値 (A=1: 次のプレイヤーを飛ばす)
const MaoSkipValue = 1

// MaoDrawTwoAmount 2を1枚出すごとに累積するペナルティ枚数
const MaoDrawTwoAmount = 2

// MaoForgotPenalty マオ宣言を忘れた場合のペナルティ枚数
const MaoForgotPenalty = 2

// MaoRulePenalty 隠しルール違反のペナルティ枚数
const MaoRulePenalty = 1

// MaoHintThreshold 隠しルールのハーフヒントが解放されるまでの正解回数
const MaoHintThreshold = 3

// MaoPhase ゲームフェーズ
type MaoPhase int

// Maoのフェーズ定数
const (
	// MaoPhasePlay 通常プレイフェーズ
	MaoPhasePlay MaoPhase = 0
	// MaoPhaseChooseSuit 8を出した後のスート選択フェーズ
	MaoPhaseChooseSuit MaoPhase = 1
	// MaoPhaseMustDeclare 手札が1枚になった後の宣言待ちフェーズ
	MaoPhaseMustDeclare MaoPhase = 2
	// MaoPhaseRoundEnd ラウンド終了フェーズ
	MaoPhaseRoundEnd MaoPhase = 3
	// MaoPhaseGameEnd ゲーム終了フェーズ
	MaoPhaseGameEnd MaoPhase = 4
)

// MaoRuleTriggerKind は隠しルールのトリガー種別を表す。
type MaoRuleTriggerKind int

// 隠しルールのトリガー種別定数
const (
	// MaoTriggerSuit 特定スートを出したらトリガー
	MaoTriggerSuit MaoRuleTriggerKind = 0
	// MaoTriggerValue 特定ランクを出したらトリガー
	MaoTriggerValue MaoRuleTriggerKind = 1
)

// MaoHiddenRule は人間プレイヤーが推理すべき秘密のルールを表す。
// トリガー条件 (スート or ランク) が発火すると、人間は RequiredWord を
// 宣言しなければならない。違反すると +1 枚のペナルティを受ける。
type MaoHiddenRule struct {
	// TriggerKind トリガー種別 (スート or ランク)
	TriggerKind MaoRuleTriggerKind `json:"tk"`
	// TriggerValue トリガー値 (スートなら CardDesign*、ランクなら 1-13)
	TriggerValue int `json:"tv"`
	// RequiredWord 発火時に宣言すべき言葉 (大文字小文字を無視して比較)
	RequiredWord string `json:"rw"`
	// HintKey トリガーを示すぼかしたヒントの i18n キー (アクションは明かさない)。
	// **文言そのものではなくキーを持つ。**日本語を直書きすると `--lang en` でも
	// 日本語のまま出る (#4917)。`mao.` を前置して引く。
	HintKey string `json:"ht"`
}

// maoRuleSet は固定の隠しルール候補。ゲーム開始時に 1 つが決定的に選ばれる。
var maoRuleSet = []MaoHiddenRule{
	{TriggerKind: MaoTriggerSuit, TriggerValue: CardDesignSpade, RequiredWord: "spade", HintKey: "hintSuit"},
	{TriggerKind: MaoTriggerSuit, TriggerValue: CardDesignHeart, RequiredWord: "heart", HintKey: "hintSuit"},
	{TriggerKind: MaoTriggerValue, TriggerValue: 7, RequiredWord: "seven", HintKey: "hintNumber"},
	{TriggerKind: MaoTriggerValue, TriggerValue: 13, RequiredWord: "thank you", HintKey: "hintFace"},
	{TriggerKind: MaoTriggerValue, TriggerValue: 1, RequiredWord: "mao", HintKey: "hintRank"},
}

// Mao マオゲームクラス (クレイジーエイト + マジックカード + 秘密の隠しルール)
type Mao struct {
	trumpCards       *TrumpCards
	players          []*MaoPlayer
	config           MaoConfig
	phase            MaoPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	chosenSuit       int // -1 = 未選択
	penaltyDrawCount int // 累積ドローツー枚数 (0 = ペナルティなし)
	direction        int // +1 = 時計回り, -1 = 反時計回り
	pendingSkip      bool
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLogBase

	// --- 隠しルールシステム ---
	hiddenRule         MaoHiddenRule // ゲーム開始時に決定的に選ばれる
	awaitingWord       bool          // 人間の直近のプレイがトリガーを発火させ、宣言待ちか
	playerCorrectCount int           // 人間が累計で正しく従った回数
	hintUnlocked       bool          // ハーフヒントが解放されたか (3回正解で解放)
	rulePenaltyFlag    bool          // 直近のアクションで隠しルール違反ペナルティが発生したか
}

// NewMao コンストラクタ
func NewMao(trumpCards *TrumpCards, players []*MaoPlayer, config MaoConfig) *Mao {
	return &Mao{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
		chosenSuit:  -1,
		direction:   1,
	}
}

// NewDefaultMao returns Mao with the standard 4-player setup (1 human, 3 CPU)
// and DefaultMaoConfig. Used as the single source of truth for CUI, Web, and
// Worker construction sites.
func NewDefaultMao() *Mao {
	players := []*MaoPlayer{
		NewMaoPlayer(true),
		NewMaoPlayer(false),
		NewMaoPlayer(false),
		NewMaoPlayer(false),
	}
	return NewMao(NewTrumpCards(0), players, DefaultMaoConfig())
}

// Reset ゲーム初期化
func (g *Mao) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.chosenSuit = -1
	g.penaltyDrawCount = 0
	g.direction = 1
	g.pendingSkip = false
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.actionLog = nil
	g.awaitingWord = false
	g.playerCorrectCount = 0
	g.hintUnlocked = false
	g.rulePenaltyFlag = false

	for _, p := range g.players {
		p.roundScore = 0
		p.cumulativeScore = 0
		p.Reset()
		p.SetIsFinished(false)
		p.SetHasDeclared(false)
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.selectHiddenRule()
	g.sortAllHands()

	g.phase = MaoPhasePlay
}

// NextRound 次のラウンドを開始する
func (g *Mao) NextRound() {
	if g.phase != MaoPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.chosenSuit = -1
	g.penaltyDrawCount = 0
	g.direction = 1
	g.pendingSkip = false
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.awaitingWord = false
	g.rulePenaltyFlag = false

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	// The hidden rule persists across rounds for a single game; correct-count
	// and hint state are intentionally NOT reset so the human keeps progress.
	g.sortAllHands()

	g.phase = MaoPhasePlay
}

// dealInitialCards 初期配布: 各プレイヤーに5枚、1枚を捨て札に
func (g *Mao) dealInitialCards() {
	g.drawPile = make([]*Card, 0, g.trumpCards.GetTotalCount())
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.drawPile = append(g.drawPile, card)
	}

	rand.Shuffle(len(g.drawPile), func(i, j int) {
		g.drawPile[i], g.drawPile[j] = g.drawPile[j], g.drawPile[i]
	})

	for i := 0; i < MaoHandSize; i++ {
		for j := 0; j < MaoPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	// 最初の1枚を捨て札に (特殊カードでも初期の効果は発動しない)
	if len(g.drawPile) > 0 {
		firstCard := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, firstCard)
	}
}

// selectHiddenRule は初期デッキ順から決定的に隠しルールを選択する。
// Math.random や時刻は使わず、配布後の残り山札の先頭数枚の値の合計を
// 安定したシードとして使用する (シャッフル順はデッキに依存するが、once
// 確定したゲーム状態に対しては決定的)。
func (g *Mao) selectHiddenRule() {
	seed := 0
	limit := 5
	for i := 0; i < len(g.drawPile) && i < limit; i++ {
		seed += g.drawPile[i].GetDesign()*13 + g.drawPile[i].GetValue()
	}
	// 捨て札の先頭も加味して分散させる。
	if len(g.discardPile) > 0 {
		top := g.discardPile[len(g.discardPile)-1]
		seed += top.GetDesign()*13 + top.GetValue()
	}
	idx := seed % len(maoRuleSet)
	if idx < 0 {
		idx += len(maoRuleSet)
	}
	g.hiddenRule = maoRuleSet[idx]
}

// ruleTriggered は与えられたカードが隠しルールのトリガーを発火させるか判定する。
func (g *Mao) ruleTriggered(card *Card) bool {
	if card == nil {
		return false
	}
	switch g.hiddenRule.TriggerKind {
	case MaoTriggerSuit:
		return card.GetDesign() == g.hiddenRule.TriggerValue
	case MaoTriggerValue:
		return card.GetValue() == g.hiddenRule.TriggerValue
	}
	return false
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *Mao) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MaoPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	g.rulePenaltyFlag = false

	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if !g.isValidPlay(card) {
		return NewDomainError(ErrInvalidPlay, "そのカードは出せません")
	}

	// 選択カードを先に取り除く。宣言待ちペナルティは手札を引いて並べ替えるため、
	// 取り除きを後に回すと cardIndex が指すカードがずれてしまう。
	played := player.RemoveCard(cardIndex)
	// 直前のプレイで宣言待ちだったが別のアクションをした → ルール違反ペナルティ。
	// (rulePenaltyFlag は上で false に戻した後なので、ここでの違反が正しく残る)
	g.resolvePendingWord(g.currentPlayerIdx, false)
	humanTriggered := g.ruleTriggered(played)
	g.playCard(g.currentPlayerIdx, played)
	// 人間のプレイがトリガーを発火させた場合は宣言待ちにする。
	// (手札が空になりラウンド終了/ゲーム終了した場合を除く)
	if humanTriggered && (g.phase == MaoPhasePlay || g.phase == MaoPhaseChooseSuit || g.phase == MaoPhaseMustDeclare) {
		g.awaitingWord = true
	}
	return nil
}

// PlayerDeclareWord 人間プレイヤーが隠しルールに従って言葉を宣言する。
// プレイ直後にトリガーが発火していると awaitingWord が立つが、その時点で
// 手番は既に次のプレイヤーへ進んでいるため、手番チェックは行わない。
// 宣言は常に人間プレイヤー (seat 0) を対象として解決する。
func (g *Mao) PlayerDeclareWord(word string) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	humanIdx := g.humanPlayerIdx()
	if !g.awaitingWord {
		// 宣言待ちでないのに言葉を発した → ルール違反 (誤発言ペナルティ)
		g.applyRulePenalty(humanIdx)
		return nil
	}
	correct := strings.EqualFold(strings.TrimSpace(word), g.hiddenRule.RequiredWord)
	g.resolvePendingWord(humanIdx, correct)
	return nil
}

// humanPlayerIdx は人間プレイヤーのインデックスを返す (見つからなければ 0)。
func (g *Mao) humanPlayerIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return 0
}

// resolvePendingWord は宣言待ち状態を解決する。complied=true なら正解として
// カウントし、false なら違反ペナルティを科す。宣言待ちでなければ何もしない。
func (g *Mao) resolvePendingWord(playerIdx int, complied bool) {
	if !g.awaitingWord {
		return
	}
	g.awaitingWord = false
	if complied {
		g.playerCorrectCount++
		g.appendLog(playerIdx, "rule_ok", fmt.Sprintf("%s follows the secret rule", playerName(g.players, playerIdx)), nil)
		if g.playerCorrectCount >= MaoHintThreshold {
			g.hintUnlocked = true
		}
		return
	}
	g.applyRulePenalty(playerIdx)
}

// applyRulePenalty は隠しルール違反として +1 枚を引かせる。ルールは明かさない。
func (g *Mao) applyRulePenalty(playerIdx int) {
	g.playerCorrectCount = 0
	g.rulePenaltyFlag = true
	drawn := g.drawCards(playerIdx, MaoRulePenalty)
	g.appendLog(playerIdx, "penalty", fmt.Sprintf("%s receives a penalty (+%d card)", playerName(g.players, playerIdx), drawn), nil)
	g.sortHand(playerIdx)
}

// PlayerChooseSuit 人間プレイヤーがスートを選択する (8を出した後)
func (g *Mao) PlayerChooseSuit(suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MaoPhaseChooseSuit {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "スートは1〜4で指定してください")
	}

	g.chosenSuit = suit
	g.appendLog(g.currentPlayerIdx, "choose_suit", fmt.Sprintf("%s chooses %s", playerName(g.players, g.currentPlayerIdx), suitName(suit)), nil)

	g.finishTurn(g.currentPlayerIdx)
	return nil
}

// PlayerDraw 人間プレイヤーがカードを引く (ペナルティ中はスタックを引き受ける)
func (g *Mao) PlayerDraw() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MaoPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	// 宣言待ちのまま引いた → ルール違反ペナルティ
	g.resolvePendingWord(g.currentPlayerIdx, false)

	return g.drawCard(g.currentPlayerIdx)
}

// PlayerDeclare 人間プレイヤーが「マオ！」と宣言する
func (g *Mao) PlayerDeclare() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MaoPhaseMustDeclare {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.doDeclare(g.currentPlayerIdx)
	return nil
}

// PlayerDeclareMao は PlayerDeclare のエイリアス (「マオ！」宣言)。
func (g *Mao) PlayerDeclareMao() error {
	return g.PlayerDeclare()
}

// PlayerSkipDeclare 人間プレイヤーが宣言をスキップする（ペナルティを受ける）
func (g *Mao) PlayerSkipDeclare() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MaoPhaseMustDeclare {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.applyDeclarePenalty(g.currentPlayerIdx)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (g *Mao) CpuPlay() {
	if g.gameEndFlag || g.phase != MaoPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	cardIdx := g.cpuSelectPlayCard(g.currentPlayerIdx)
	if cardIdx >= 0 {
		player := g.players[g.currentPlayerIdx]
		played := player.RemoveCard(cardIdx)
		// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
		// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
		// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
		if played == nil {
			return
		}
		g.playCard(g.currentPlayerIdx, played)
	} else {
		// drawCard always returns nil today; the error is ignored intentionally.
		_ = g.drawCard(g.currentPlayerIdx)
	}
}

// CpuChooseSuit CPUがスートを選択する
func (g *Mao) CpuChooseSuit() {
	if g.gameEndFlag || g.phase != MaoPhaseChooseSuit {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	suit := g.cpuSelectSuit(g.currentPlayerIdx)
	g.chosenSuit = suit
	g.appendLog(g.currentPlayerIdx, "choose_suit", fmt.Sprintf("%s chooses %s", playerName(g.players, g.currentPlayerIdx), suitName(suit)), nil)
	g.finishTurn(g.currentPlayerIdx)
}

// CpuDeclare CPUが自動的に「マオ！」と宣言する
func (g *Mao) CpuDeclare() {
	if g.gameEndFlag || g.phase != MaoPhaseMustDeclare {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	// 難易度 Easy では一定確率で宣言を忘れる
	if g.config.CpuDifficulty == MaoCpuDifficultyEasy && rand.Intn(4) == 0 {
		g.applyDeclarePenalty(g.currentPlayerIdx)
		return
	}
	g.doDeclare(g.currentPlayerIdx)
}

// doDeclare 宣言処理の共通実装
func (g *Mao) doDeclare(playerIdx int) {
	g.players[playerIdx].SetHasDeclared(true)
	g.appendLog(playerIdx, "declare", fmt.Sprintf("%s declares Mao!", playerName(g.players, playerIdx)), nil)
	g.advanceTurn()
}

// applyDeclarePenalty 宣言忘れペナルティとして規定枚数を引かせる
func (g *Mao) applyDeclarePenalty(playerIdx int) {
	drawn := g.drawCards(playerIdx, MaoForgotPenalty)
	g.appendLog(playerIdx, "penalty", fmt.Sprintf("%s forgot to declare Mao! (+%d cards)", playerName(g.players, playerIdx), drawn), nil)
	g.sortHand(playerIdx)
	g.advanceTurn()
}

// ScoreRound ラウンドのスコアを確定する
func (g *Mao) ScoreRound() {
	if g.phase != MaoPhaseRoundEnd {
		return
	}

	winnerIdx := -1
	for i, p := range g.players {
		if p.GetCardsSize() == 0 {
			winnerIdx = i
			break
		}
	}

	if winnerIdx < 0 {
		return
	}

	totalScore := 0
	for i, p := range g.players {
		if i == winnerIdx {
			continue
		}
		score := 0
		for j := 0; j < p.GetCardsSize(); j++ {
			score += crazyEightsCardScore(p.GetCard(j))
		}
		totalScore += score
		g.appendLog(i, "hand_score", fmt.Sprintf("%s: %d points remaining", playerName(g.players, i), score), nil)
	}

	g.players[winnerIdx].roundScore = totalScore
	g.appendLog(winnerIdx, "round_win", fmt.Sprintf("%s wins round %d (+%d points)", playerName(g.players, winnerIdx), g.roundNumber, totalScore), nil)

	g.players[winnerIdx].CommitRoundScore()

	g.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Mao) GetPhase() MaoPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Mao) SetPhase(phase MaoPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Mao) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Mao) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Mao) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Mao) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Mao) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Mao) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *Mao) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Mao) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *Mao) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetChosenSuit 選択されたスート取得 (-1 = 未選択)
func (g *Mao) GetChosenSuit() int { return g.chosenSuit }

// SetChosenSuit スート設定 (テスト用)
func (g *Mao) SetChosenSuit(suit int) { g.chosenSuit = suit }

// GetPenaltyDrawCount 累積ドローツー枚数取得 (0 = ペナルティなし)
func (g *Mao) GetPenaltyDrawCount() int { return g.penaltyDrawCount }

// SetPenaltyDrawCount 累積ドローツー枚数設定 (テスト用)
func (g *Mao) SetPenaltyDrawCount(n int) { g.penaltyDrawCount = n }

// GetDirection プレイ方向取得 (+1 = 時計回り, -1 = 反時計回り)
func (g *Mao) GetDirection() int { return g.direction }

// SetDirection プレイ方向設定 (テスト用)
func (g *Mao) SetDirection(d int) { g.direction = d }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Mao) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *Mao) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Mao) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Mao) GetPlayer(i int) *MaoPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Mao) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Mao) GetConfig() MaoConfig { return g.config }

// SetConfig 設定変更
func (g *Mao) SetConfig(cfg MaoConfig) { g.config = cfg }

// --- Hidden-rule getters/setters ---

// GetAwaitingWord 人間が隠しルールに従って言葉を宣言すべき状態か
func (g *Mao) GetAwaitingWord() bool { return g.awaitingWord }

// SetAwaitingWord 宣言待ちフラグ設定 (テスト用)
func (g *Mao) SetAwaitingWord(v bool) { g.awaitingWord = v }

// GetPlayerCorrectCount 人間が隠しルールに正しく従った累計回数
func (g *Mao) GetPlayerCorrectCount() int { return g.playerCorrectCount }

// GetHintUnlocked ハーフヒントが解放されたか (3回正解で解放)
func (g *Mao) GetHintUnlocked() bool { return g.hintUnlocked }

// GetRulePenaltyFlag 直近のアクションで隠しルール違反ペナルティが発生したか
func (g *Mao) GetRulePenaltyFlag() bool { return g.rulePenaltyFlag }

// GetRuleHintKey 解放済みであればハーフヒントの i18n キーを返す。未解放なら空文字。
// 解放後もトリガーのみを示し、宣言すべき言葉そのものは明かさない。
// 翻訳は presenter 側で行う (ドメインは i18n に依存しない)。
func (g *Mao) GetRuleHintKey() string {
	if !g.hintUnlocked {
		return ""
	}
	return g.hiddenRule.HintKey
}

// GetHiddenRule 隠しルールを返す (ドメイン内部・テスト用。Web には公開しない)
func (g *Mao) GetHiddenRule() MaoHiddenRule { return g.hiddenRule }

// SetHiddenRule 隠しルールを設定する (テスト用)
func (g *Mao) SetHiddenRule(r MaoHiddenRule) { g.hiddenRule = r }

// --- Private methods ---

// isValidPlay カードがプレイ可能か判定
func (g *Mao) isValidPlay(card *Card) bool {
	// ペナルティ中は2のみ重ねられる
	if g.penaltyDrawCount > 0 {
		return card.GetValue() == MaoDrawTwoValue
	}

	// 8はいつでも出せる
	if card.GetValue() == MaoWildValue {
		return true
	}

	top := g.GetDiscardTop()
	if top == nil {
		return true
	}

	// chosenSuit が設定されている場合 (前の人が8を出した)
	if g.chosenSuit > 0 {
		return card.GetDesign() == g.chosenSuit
	}

	// スートまたはランクが一致
	return card.GetDesign() == top.GetDesign() || card.GetValue() == top.GetValue()
}

// playCard カードをプレイする共通処理
func (g *Mao) playCard(playerIdx int, card *Card) {
	g.discardPile = append(g.discardPile, card)
	g.chosenSuit = -1

	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	// マジックカードの状態更新
	switch card.GetValue() {
	case MaoDrawTwoValue:
		g.penaltyDrawCount += MaoDrawTwoAmount
		g.appendLog(playerIdx, "draw_two", fmt.Sprintf("Draw stack is now %d", g.penaltyDrawCount), nil)
	case MaoSkipValue:
		g.pendingSkip = true
		g.appendLog(playerIdx, "skip", "Next player is skipped", nil)
	}

	// 手札が空になったらラウンド終了
	if g.players[playerIdx].GetCardsSize() == 0 {
		g.players[playerIdx].SetIsFinished(true)
		g.phase = MaoPhaseRoundEnd
		return
	}

	// 8を出した場合はスート選択フェーズへ
	if card.GetValue() == MaoWildValue {
		g.phase = MaoPhaseChooseSuit
		return
	}

	g.finishTurn(playerIdx)
}

// finishTurn 宣言チェックを行い、必要なら宣言フェーズへ、そうでなければ次の手番へ
func (g *Mao) finishTurn(playerIdx int) {
	if g.players[playerIdx].GetCardsSize() == 1 && !g.players[playerIdx].GetHasDeclared() {
		g.phase = MaoPhaseMustDeclare
		return
	}
	g.advanceTurn()
}

// advanceTurn 次のプレイヤーへ (方向・スキップを反映)
func (g *Mao) advanceTurn() {
	steps := 1
	if g.pendingSkip {
		steps = 2
		g.pendingSkip = false
	}
	g.currentPlayerIdx = g.wrapIdx(g.currentPlayerIdx + steps*g.direction)
	g.phase = MaoPhasePlay
	// 手札が2枚以上に戻った場合は宣言フラグをリセット
	for _, p := range g.players {
		if p.GetCardsSize() >= 2 {
			p.SetHasDeclared(false)
		}
	}
}

// wrapIdx プレイヤーインデックスを 0..len(players)-1 に正規化する (負の方向にも対応)
func (g *Mao) wrapIdx(i int) int {
	n := len(g.players)
	if n == 0 {
		return 0
	}
	return ((i % n) + n) % n
}

// drawCard カードを引く (ペナルティ中はスタックを引き受けて手番終了)
func (g *Mao) drawCard(playerIdx int) error {
	if g.penaltyDrawCount > 0 {
		drawn := g.drawCards(playerIdx, g.penaltyDrawCount)
		g.penaltyDrawCount = 0
		g.appendLog(playerIdx, "take_penalty", fmt.Sprintf("%s takes %d penalty cards", playerName(g.players, playerIdx), drawn), nil)
		g.sortHand(playerIdx)
		g.advanceTurn()
		return nil
	}

	if len(g.drawPile) == 0 {
		g.recycleDrawPile()
	}

	if len(g.drawPile) == 0 {
		// 引けるカードがない→パス
		g.appendLog(playerIdx, "pass", fmt.Sprintf("%s passes (no cards to draw)", playerName(g.players, playerIdx)), nil)
		g.advanceTurn()
		return nil
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[playerIdx].AddCard(card)
	g.sortHand(playerIdx)

	g.appendLog(playerIdx, "draw", fmt.Sprintf("%s draws a card", playerName(g.players, playerIdx)), nil)

	// 引いたカードが出せないなら次へ
	if !g.hasPlayableCard(playerIdx) {
		g.advanceTurn()
	}

	return nil
}

// drawCards 指定枚数を引く (山札が尽きたら捨て札を再利用)。実際に引けた枚数を返す。
func (g *Mao) drawCards(playerIdx, n int) int {
	drawn := 0
	for i := 0; i < n; i++ {
		if len(g.drawPile) == 0 {
			g.recycleDrawPile()
		}
		if len(g.drawPile) == 0 {
			break
		}
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.players[playerIdx].AddCard(card)
		drawn++
	}
	return drawn
}

// recycleDrawPile 捨て札から山札を再構築する
func (g *Mao) recycleDrawPile() {
	if len(g.discardPile) <= 1 {
		return
	}

	top := g.discardPile[len(g.discardPile)-1]
	recycled := g.discardPile[:len(g.discardPile)-1]
	g.discardPile = []*Card{top}

	rand.Shuffle(len(recycled), func(i, j int) {
		recycled[i], recycled[j] = recycled[j], recycled[i]
	})

	g.drawPile = recycled
}

// hasPlayableCard プレイヤーが出せるカードを持っているか
func (g *Mao) hasPlayableCard(playerIdx int) bool {
	player := g.players[playerIdx]
	for i := 0; i < player.GetCardsSize(); i++ {
		if g.isValidPlay(player.GetCard(i)) {
			return true
		}
	}
	return false
}

// checkGameEnd ゲーム終了判定
func (g *Mao) checkGameEnd() {
	hasWinner := false
	for i := 0; i < MaoPlayerCnt; i++ {
		if g.players[i].cumulativeScore >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}

	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = MaoPhaseGameEnd

	// 最高スコアのプレイヤーが勝者
	maxScore := g.players[0].cumulativeScore
	g.winnerIdx = 0
	for i := 1; i < MaoPlayerCnt; i++ {
		if g.players[i].cumulativeScore > maxScore {
			maxScore = g.players[i].cumulativeScore
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, g.winnerIdx)), nil)
}

// sortAllHands 全プレイヤーの手札をソートする
func (g *Mao) sortAllHands() {
	sortHands(len(g.players), g)
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *Mao) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// --- CPU AI ---

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する (-1 = プレイ不可)
func (g *Mao) cpuSelectPlayCard(playerIdx int) int {
	validIndices := g.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return -1
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch g.config.CpuDifficulty {
	case MaoCpuDifficultyHard:
		return g.cpuPlayHard(playerIdx, validIndices)
	case MaoCpuDifficultyNormal:
		return g.cpuPlayNormal(playerIdx, validIndices)
	default:
		return g.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (g *Mao) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal 最も多いスートを優先、8を温存
func (g *Mao) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]

	nonWild := make([]int, 0)
	for _, idx := range validIndices {
		if player.GetCard(idx).GetValue() != MaoWildValue {
			nonWild = append(nonWild, idx)
		}
	}

	candidates := validIndices
	if len(nonWild) > 0 {
		candidates = nonWild
	}

	suitCount := g.countSuits(playerIdx)
	bestIdx := candidates[0]
	bestCount := suitCount[player.GetCard(candidates[0]).GetDesign()]
	for _, idx := range candidates[1:] {
		sc := suitCount[player.GetCard(idx).GetDesign()]
		if sc > bestCount {
			bestCount = sc
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuPlayHard 戦略的プレイ: 8を温存しつつ高得点カードを優先消費する
func (g *Mao) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]

	nonWild := make([]int, 0)
	for _, idx := range validIndices {
		if player.GetCard(idx).GetValue() != MaoWildValue {
			nonWild = append(nonWild, idx)
		}
	}

	// 手札が2枚以下なら8を使ってもOK
	if player.GetCardsSize() <= 2 {
		nonWild = validIndices
	}

	candidates := validIndices
	if len(nonWild) > 0 {
		candidates = nonWild
	}

	suitCount := g.countSuits(playerIdx)
	bestIdx := candidates[0]
	bestScore := g.cpuCardPriority(player.GetCard(candidates[0]), suitCount)
	for _, idx := range candidates[1:] {
		score := g.cpuCardPriority(player.GetCard(idx), suitCount)
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuCardPriority カードの優先度スコアを計算 (大きいほど先に出したい)
func (g *Mao) cpuCardPriority(card *Card, suitCount map[int]int) int {
	score := crazyEightsCardScore(card)  // 高得点カードを優先消費
	score += suitCount[card.GetDesign()] // 多いスートを優先
	return score
}

// cpuSelectSuit CPUがスートを選択する
func (g *Mao) cpuSelectSuit(playerIdx int) int {
	switch g.config.CpuDifficulty {
	case MaoCpuDifficultyHard, MaoCpuDifficultyNormal:
		return g.cpuSelectSuitSmart(playerIdx)
	default:
		return g.cpuSelectSuitRandom()
	}
}

// cpuSelectSuitRandom ランダムにスートを選択
func (g *Mao) cpuSelectSuitRandom() int {
	return rand.Intn(4) + 1 // 1-4 (Spade, Clover, Heart, Diamond)
}

// cpuSelectSuitSmart 手札で最も多いスートを選択
func (g *Mao) cpuSelectSuitSmart(playerIdx int) int {
	suitCount := g.countSuits(playerIdx)

	bestSuit := CardDesignSpade
	bestCount := 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if suitCount[suit] > bestCount {
			bestCount = suitCount[suit]
			bestSuit = suit
		}
	}
	return bestSuit
}

// countSuits プレイヤーの手札のスート別枚数をカウント (8は除外)
func (g *Mao) countSuits(playerIdx int) map[int]int {
	player := g.players[playerIdx]
	counts := make(map[int]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetValue() != MaoWildValue {
			counts[card.GetDesign()]++
		}
	}
	return counts
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *Mao) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.isValidPlay(c) })
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (g *Mao) GetValidPlayIndices(playerIdx int) []int {
	return g.getValidPlayIndices(playerIdx)
}

// maoJSON is the JSON wire format for Mao.
type maoJSON struct {
	TrumpCards         *TrumpCards       `json:"tc"`
	Players            []*MaoPlayer      `json:"pl"`
	Config             MaoConfig         `json:"cf"`
	Phase              MaoPhase          `json:"ps"`
	CurrentPlayerIdx   int               `json:"ci"`
	DiscardPile        []*Card           `json:"dp"`
	DrawPile           []*Card           `json:"wp"`
	ChosenSuit         int               `json:"cs"`
	PenaltyDrawCount   int               `json:"pd"`
	Direction          int               `json:"dr"`
	PendingSkip        bool              `json:"sk"`
	GameEndFlag        bool              `json:"ge"`
	WinnerIdx          int               `json:"wi"`
	RoundNumber        int               `json:"rn"`
	ActionLog          []*ActionLogEntry `json:"al"`
	HiddenRule         MaoHiddenRule     `json:"hr"`
	AwaitingWord       bool              `json:"aw"`
	PlayerCorrectCount int               `json:"pc"`
	HintUnlocked       bool              `json:"hu"`
	RulePenaltyFlag    bool              `json:"rp"`
}

// MarshalJSON implements json.Marshaler. The hidden rule IS included so the
// game state survives KV round-trips between HTTP requests. The WebPresenter is
// responsible for NOT leaking the rule to the client.
func (g *Mao) MarshalJSON() ([]byte, error) {
	return json.Marshal(maoJSON{
		TrumpCards:         g.trumpCards,
		Players:            g.players,
		Config:             g.config,
		Phase:              g.phase,
		CurrentPlayerIdx:   g.currentPlayerIdx,
		DiscardPile:        g.discardPile,
		DrawPile:           g.drawPile,
		ChosenSuit:         g.chosenSuit,
		PenaltyDrawCount:   g.penaltyDrawCount,
		Direction:          g.direction,
		PendingSkip:        g.pendingSkip,
		GameEndFlag:        g.gameEndFlag,
		WinnerIdx:          g.winnerIdx,
		RoundNumber:        g.roundNumber,
		ActionLog:          g.actionLog,
		HiddenRule:         g.hiddenRule,
		AwaitingWord:       g.awaitingWord,
		PlayerCorrectCount: g.playerCorrectCount,
		HintUnlocked:       g.hintUnlocked,
		RulePenaltyFlag:    g.rulePenaltyFlag,
	})
}

// maoMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const maoMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *Mao) UnmarshalJSON(data []byte) error {
	var j maoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > maoMaxSliceLen || len(j.DiscardPile) > maoMaxSliceLen ||
		len(j.DrawPile) > maoMaxSliceLen || len(j.ActionLog) > maoMaxSliceLen {
		return fmt.Errorf("mao: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("mao: invalid config: %w", err)
	}
	// Mao is strictly a 4-player game; reject malformed states that would
	// otherwise cause out-of-bounds panics during play (many methods iterate
	// with the MaoPlayerCnt bound or index g.players[g.currentPlayerIdx]).
	if len(j.Players) != MaoPlayerCnt {
		return fmt.Errorf("mao: invalid player count: expected %d, got %d", MaoPlayerCnt, len(j.Players))
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("mao: player slice contains nil element")
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= MaoPlayerCnt {
		return fmt.Errorf("mao: currentPlayerIdx %d out of range [0, %d)", j.CurrentPlayerIdx, MaoPlayerCnt)
	}
	if j.Phase < MaoPhasePlay || j.Phase > MaoPhaseGameEnd {
		return fmt.Errorf("mao: invalid phase: %d", j.Phase)
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.discardPile = j.DiscardPile
	if g.discardPile == nil {
		g.discardPile = make([]*Card, 0)
	}
	g.drawPile = j.DrawPile
	if g.drawPile == nil {
		g.drawPile = make([]*Card, 0)
	}
	g.chosenSuit = j.ChosenSuit
	g.penaltyDrawCount = j.PenaltyDrawCount
	if g.penaltyDrawCount < 0 {
		g.penaltyDrawCount = 0
	} else if g.penaltyDrawCount > maoMaxSliceLen {
		g.penaltyDrawCount = maoMaxSliceLen
	}
	g.direction = j.Direction
	if g.direction != 1 && g.direction != -1 {
		g.direction = 1
	}
	g.pendingSkip = j.PendingSkip
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundNumber = j.RoundNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.hiddenRule = j.HiddenRule
	g.awaitingWord = j.AwaitingWord
	g.playerCorrectCount = j.PlayerCorrectCount
	if g.playerCorrectCount < 0 {
		g.playerCorrectCount = 0
	} else if g.playerCorrectCount > maoMaxSliceLen {
		g.playerCorrectCount = maoMaxSliceLen
	}
	g.hintUnlocked = j.HintUnlocked
	g.rulePenaltyFlag = j.RulePenaltyFlag
	return nil
}
