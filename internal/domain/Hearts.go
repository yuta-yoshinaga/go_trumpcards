package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// HeartsPlayerCnt ハーツプレイヤー数
const HeartsPlayerCnt = 4

// HeartsHandSize 各プレイヤーの手札枚数
const HeartsHandSize = 13

// HeartsMaxPoints シュート・ザ・ムーン時のペナルティ合計
const HeartsMaxPoints = 26

// HeartsPhase ゲームフェーズ
type HeartsPhase int

// Heartsのフェーズ定数
const (
	// HeartsPhasePass カード交換フェーズ
	HeartsPhasePass HeartsPhase = 0
	// HeartsPhasePlay トリックプレイフェーズ
	HeartsPhasePlay HeartsPhase = 1
	// HeartsPhaseTrickEnd トリック終了フェーズ
	HeartsPhaseTrickEnd HeartsPhase = 2
	// HeartsPhaseRoundEnd ラウンド終了フェーズ
	HeartsPhaseRoundEnd HeartsPhase = 3
	// HeartsPhaseGameEnd ゲーム終了フェーズ
	HeartsPhaseGameEnd HeartsPhase = 4
)

// HeartsPassDirection カード交換方向
type HeartsPassDirection int

// Heartsのカードパス方向定数
const (
	// HeartsPassLeft 左へ渡す
	HeartsPassLeft HeartsPassDirection = 0
	// HeartsPassRight 右へ渡す
	HeartsPassRight HeartsPassDirection = 1
	// HeartsPassAcross 向かいへ渡す
	HeartsPassAcross HeartsPassDirection = 2
	// HeartsPassNone 交換なし
	HeartsPassNone HeartsPassDirection = 3
)

// HeartsPassCardCount パスで渡すカード枚数
const HeartsPassCardCount = 3

// HeartsHint ヒント情報
type HeartsHint struct {
	CardIndices []int  // 推奨カードインデックス (プレイ時1枚, パス時3枚)
	Reason      string // ヒント理由キー
}

// Hearts ハーツゲームクラス
type Hearts struct {
	trumpCards       *TrumpCards
	players          []*HeartsPlayer
	config           HeartsConfig
	phase            HeartsPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	heartsBroken     bool
	passedCards      [HeartsPlayerCnt][]*Card
	passReady        [HeartsPlayerCnt]bool
	leadPlayerIdx    int
	gameEndFlag      bool
	winnerIdx        int
	actionLogBase
}

// NewHearts コンストラクタ
func NewHearts(trumpCards *TrumpCards, players []*HeartsPlayer, config HeartsConfig) *Hearts {
	return &Hearts{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
	}
}

// NewDefaultHearts returns Hearts with the standard 4-player setup (1 human, 3 CPU)
// and DefaultHeartsConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultHearts() *Hearts {
	players := []*HeartsPlayer{
		NewHeartsPlayer(true),
		NewHeartsPlayer(false),
		NewHeartsPlayer(false),
		NewHeartsPlayer(false),
	}
	return NewHearts(NewTrumpCards(0), players, DefaultHeartsConfig())
}

// Reset ゲーム初期化: デッキをシャッフルして配布し、最初のフェーズを設定
func (h *Hearts) Reset() {
	h.gameEndFlag = false
	h.winnerIdx = -1
	h.roundNumber = 1
	h.trickNumber = 0
	h.heartsBroken = false
	h.currentTrick = nil
	h.leadPlayerIdx = -1
	h.currentPlayerIdx = -1
	h.actionLog = nil

	// ラウンドスコアのみリセット（累積スコアも含めて全リセット）
	for _, p := range h.players {
		p.roundScore = 0
		p.cumulativeScore = 0
		p.tricksTaken = nil
		p.Reset()
		p.SetIsFinished(false)
	}

	h.passedCards = [HeartsPlayerCnt][]*Card{}
	h.passReady = [HeartsPlayerCnt]bool{}

	h.trumpCards.Shuffle()
	dealAllCards(h.trumpCards, h.players)
	h.sortAllHands()

	// Reset always sets roundNumber=1, so passDirection is always Left (pass phase)
	h.phase = HeartsPhasePass
}

// NextRound 次のラウンドを開始する
func (h *Hearts) NextRound() {
	if h.phase != HeartsPhaseRoundEnd {
		return
	}

	h.roundNumber++
	h.trickNumber = 0
	h.heartsBroken = false
	h.currentTrick = nil
	h.leadPlayerIdx = -1
	h.currentPlayerIdx = -1

	for _, p := range h.players {
		p.ResetRound()
	}

	h.passedCards = [HeartsPlayerCnt][]*Card{}
	h.passReady = [HeartsPlayerCnt]bool{}

	h.trumpCards.Shuffle()
	dealAllCards(h.trumpCards, h.players)
	h.sortAllHands()

	dir := h.GetPassDirection()
	if dir == HeartsPassNone {
		h.phase = HeartsPhasePlay
		h.startPlayPhase()
	} else {
		h.phase = HeartsPhasePass
	}
}

// PlayerPass 人間プレイヤーがカードを渡す
func (h *Hearts) PlayerPass(cardIndices []int) error {
	if h.gameEndFlag {
		return ErrGameEnded
	}
	if h.phase != HeartsPhasePass {
		return ErrWrongPhase
	}

	humanIdx := findHumanIdx(h.players)
	if humanIdx < 0 {
		return ErrNotHumanTurn
	}
	if h.passReady[humanIdx] {
		return NewDomainError(ErrInvalidPlay, "すでにカードを選択済みです")
	}
	if len(cardIndices) != HeartsPassCardCount {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("%d枚のカードを選択してください", HeartsPassCardCount))
	}

	// 重複チェック
	seen := make(map[int]bool, len(cardIndices))
	for _, idx := range cardIndices {
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "カードインデックスが重複しています")
		}
		seen[idx] = true
	}

	player := h.players[humanIdx]
	for _, idx := range cardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
	}

	cards := player.RemoveCards(cardIndices)
	h.passedCards[humanIdx] = cards
	h.passReady[humanIdx] = true

	return nil
}

// CpuPass すべてのCPUプレイヤーがカードを選択する
func (h *Hearts) CpuPass() {
	for i := 0; i < HeartsPlayerCnt; i++ {
		if h.players[i].GetIsHuman() || h.passReady[i] {
			continue
		}
		cards := h.cpuSelectPassCards(i)
		h.passedCards[i] = cards
		h.passReady[i] = true
	}
}

// ExecutePass カード交換を実行する
func (h *Hearts) ExecutePass() {
	if h.phase != HeartsPhasePass {
		return
	}
	// 全員が準備完了か確認
	for i := 0; i < HeartsPlayerCnt; i++ {
		if !h.passReady[i] {
			return
		}
	}

	dir := h.GetPassDirection()
	for i := 0; i < HeartsPlayerCnt; i++ {
		target := h.passTarget(i, dir)
		for _, card := range h.passedCards[i] {
			h.players[target].AddCard(card)
		}
	}

	h.appendLog(-1, "pass", fmt.Sprintf("round %d: cards passed (%s)", h.roundNumber, h.passDirectionStr(dir)), nil)

	h.sortAllHands()
	h.phase = HeartsPhasePlay
	h.startPlayPhase()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (h *Hearts) PlayerPlay(cardIndex int) error {
	if h.gameEndFlag {
		return ErrGameEnded
	}
	if h.phase != HeartsPhasePlay {
		return ErrWrongPhase
	}
	if !h.players[h.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := h.players[h.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := h.validatePlay(h.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	h.playCard(h.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (h *Hearts) CpuPlay() {
	if h.gameEndFlag || h.phase != HeartsPhasePlay {
		return
	}
	if h.players[h.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := h.players[h.currentPlayerIdx]
	cardIdx := h.cpuSelectPlayCard(h.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	h.playCard(h.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (h *Hearts) ResolveTrick() {
	if h.phase != HeartsPhaseTrickEnd || len(h.currentTrick) != HeartsPlayerCnt {
		return
	}

	winnerIdx := h.trickWinner()
	trickCards := make([]*Card, len(h.currentTrick))
	for i, tc := range h.currentTrick {
		trickCards[i] = tc.Card
	}

	h.players[winnerIdx].AddTrick(trickCards)

	// ポイント計算
	points := 0
	for _, tc := range h.currentTrick {
		points += cardPoints(tc.Card, h.config.OmnibusJD)
	}
	h.players[winnerIdx].roundScore += points

	winnerName := playerName(h.players, winnerIdx)
	h.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d (%+d pts)", winnerName, h.trickNumber, points), trickCards)

	h.leadPlayerIdx = winnerIdx

	if h.trickNumber >= HeartsHandSize {
		h.phase = HeartsPhaseRoundEnd
	} else {
		h.phase = HeartsPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (h *Hearts) NextTrick() {
	if h.phase != HeartsPhaseTrickEnd {
		return
	}
	h.currentTrick = nil
	h.currentPlayerIdx = h.leadPlayerIdx
	h.trickNumber++
	h.phase = HeartsPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (h *Hearts) ScoreRound() {
	if h.phase != HeartsPhaseRoundEnd {
		return
	}

	// シュート・ザ・ムーン判定
	// オムニバス時: 全ペナルティ(26) + J♦(-10) = 16 がムーンしきい値
	moonThreshold := HeartsMaxPoints
	if h.config.OmnibusJD {
		moonThreshold = HeartsMaxPoints - 10
	}
	moonShooter := -1
	for i := 0; i < HeartsPlayerCnt; i++ {
		if h.players[i].roundScore == moonThreshold {
			// オムニバス時: スコア16はJ♦なしでも到達できる（例: ハート3枚+Q♠=16）。
			// J♦を実際に取得したか確認する。J♦がある場合のみスコア16は正当なムーン。
			if h.config.OmnibusJD && !h.playerTookCard(i, CardDesignDiamond, 11) {
				continue
			}
			moonShooter = i
			break
		}
	}

	if moonShooter >= 0 {
		h.appendLog(moonShooter, "shoot_moon", fmt.Sprintf("%s shot the moon!", playerName(h.players, moonShooter)), nil)
		h.players[moonShooter].roundScore = 0
		for i := 0; i < HeartsPlayerCnt; i++ {
			if i != moonShooter {
				h.players[i].roundScore = HeartsMaxPoints
			}
		}
	}

	// 累積スコアに加算
	for i := 0; i < HeartsPlayerCnt; i++ {
		h.players[i].CommitRoundScore()
	}

	// スコアログ
	for i := 0; i < HeartsPlayerCnt; i++ {
		h.appendLog(i, "round_score", fmt.Sprintf("%s: round=%d, total=%d",
			playerName(h.players, i), h.players[i].roundScore, h.players[i].cumulativeScore), nil)
	}

	// ゲーム終了判定
	maxScore := 0
	for i := 0; i < HeartsPlayerCnt; i++ {
		if h.players[i].cumulativeScore > maxScore {
			maxScore = h.players[i].cumulativeScore
		}
	}

	if maxScore >= h.config.PointLimit {
		h.gameEndFlag = true
		h.phase = HeartsPhaseGameEnd

		// 最低スコアのプレイヤーが勝者
		minScore := h.players[0].cumulativeScore
		h.winnerIdx = 0
		for i := 1; i < HeartsPlayerCnt; i++ {
			if h.players[i].cumulativeScore < minScore {
				minScore = h.players[i].cumulativeScore
				h.winnerIdx = i
			}
		}
		h.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(h.players, h.winnerIdx)), nil)
	}
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (h *Hearts) GetPhase() HeartsPhase { return h.phase }

// SetPhase フェーズ設定 (テスト用)
func (h *Hearts) SetPhase(phase HeartsPhase) { h.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (h *Hearts) GetRoundNumber() int { return h.roundNumber }

// GetTrickNumber 現在のトリック番号取得
func (h *Hearts) GetTrickNumber() int { return h.trickNumber }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (h *Hearts) GetCurrentPlayerIdx() int { return h.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (h *Hearts) SetCurrentPlayerIdx(idx int) { h.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (h *Hearts) GetCurrentTrick() []*TrickCard { return h.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (h *Hearts) SetCurrentTrick(trick []*TrickCard) { h.currentTrick = trick }

// GetHeartsBroken ハーツブレイク状態取得
func (h *Hearts) GetHeartsBroken() bool { return h.heartsBroken }

// SetHeartsBroken ハーツブレイク状態設定 (テスト用)
func (h *Hearts) SetHeartsBroken(broken bool) { h.heartsBroken = broken }

// GetPassDirection 現在のラウンドの交換方向取得
func (h *Hearts) GetPassDirection() HeartsPassDirection {
	return HeartsPassDirection((h.roundNumber - 1) % 4)
}

// GetGameEndFlag ゲーム終了フラグ取得
func (h *Hearts) GetGameEndFlag() bool { return h.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (h *Hearts) GetWinnerIdx() int { return h.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (h *Hearts) GetPlayerCnt() int { return len(h.players) }

// GetPlayer プレイヤー取得
func (h *Hearts) GetPlayer(i int) *HeartsPlayer {
	return getPlayer(h.players, i)
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (h *Hearts) GetLeadPlayerIdx() int { return h.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (h *Hearts) SetLeadPlayerIdx(idx int) { h.leadPlayerIdx = idx }

// IsHumanTurn 現在の手番が人間かどうか
func (h *Hearts) IsHumanTurn() bool {
	return isHumanTurn(h.players, h.currentPlayerIdx)
}

// GetConfig 設定取得
func (h *Hearts) GetConfig() HeartsConfig { return h.config }

// SetConfig 設定変更
func (h *Hearts) SetConfig(cfg HeartsConfig) { h.config = cfg }

// GetPassReady パス準備状態取得
func (h *Hearts) GetPassReady() [HeartsPlayerCnt]bool { return h.passReady }

// GetPassedCards パス済みカード取得
func (h *Hearts) GetPassedCards() [HeartsPlayerCnt][]*Card { return h.passedCards }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (h *Hearts) SetRoundNumber(n int) { h.roundNumber = n }

// SetTrickNumber トリック番号設定 (テスト用)
func (h *Hearts) SetTrickNumber(n int) { h.trickNumber = n }

// --- Private methods ---

// startPlayPhase プレイフェーズ開始: 2♣を持つプレイヤーをリードに設定
func (h *Hearts) startPlayPhase() {
	if h.trickNumber == 0 {
		// 最初のトリック: 2♣を持つプレイヤーがリード
		starter := h.findTwoOfClubs()
		if starter >= 0 {
			h.leadPlayerIdx = starter
			h.currentPlayerIdx = starter
		} else {
			h.leadPlayerIdx = 0
			h.currentPlayerIdx = 0
		}
		h.trickNumber = 1
		h.currentTrick = nil
	} else {
		h.currentPlayerIdx = h.leadPlayerIdx
	}
}

// findTwoOfClubs 2♣を持つプレイヤーのインデックスを返す
func (h *Hearts) findTwoOfClubs() int {
	for i, p := range h.players {
		for j := 0; j < p.GetCardsSize(); j++ {
			card := p.GetCard(j)
			if card.GetDesign() == CardDesignClover && card.GetValue() == 2 {
				return i
			}
		}
	}
	return -1
}

// playCard カードをプレイする共通処理
func (h *Hearts) playCard(playerIdx int, card *Card) {
	h.currentTrick = append(h.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	// ハーツブレイクチェック
	if card.GetDesign() == CardDesignHeart {
		h.heartsBroken = true
	}

	h.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(h.players, playerIdx), cardStr(card)), []*Card{card})

	if len(h.currentTrick) == HeartsPlayerCnt {
		h.phase = HeartsPhaseTrickEnd
	} else {
		h.currentPlayerIdx = (h.currentPlayerIdx + 1) % HeartsPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
func (h *Hearts) validatePlay(playerIdx int, card *Card) error {
	player := h.players[playerIdx]

	// 最初のトリックの最初のカードは2♣でなければならない
	if h.trickNumber == 1 && len(h.currentTrick) == 0 {
		if card.GetDesign() != CardDesignClover || card.GetValue() != 2 {
			// 2♣を持っているか確認
			if h.playerHasCard(playerIdx, CardDesignClover, 2) {
				return NewDomainError(ErrInvalidPlay, "最初のトリックは2♣でリードしてください")
			}
		}
	}

	if len(h.currentTrick) == 0 {
		// リード: ハーツが壊れていない場合、ハーツでリードできない（他にカードがある場合）
		if !h.heartsBroken && card.GetDesign() == CardDesignHeart {
			if h.playerHasNonHeart(playerIdx) {
				return NewDomainError(ErrInvalidPlay, "ハーツはまだブレイクされていません")
			}
		}
		return nil
	}

	// フォロースート
	leadSuit := h.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit {
		// そのスートを持っていない場合のみ許可
		if h.playerHasSuit(playerIdx, leadSuit) {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
		// 最初のトリックではハートとQ♠を出せない（強制される場合を除く）
		if h.trickNumber == 1 {
			if isPointCard(card, h.config.OmnibusJD) && h.playerHasNonPointCard(player) {
				return NewDomainError(ErrInvalidPlay, "最初のトリックではポイントカードを出せません")
			}
		}
	}

	return nil
}

// playerHasCard プレイヤーが特定のカードを持っているか
func (h *Hearts) playerHasCard(playerIdx int, design, value int) bool {
	p := h.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == design && c.GetValue() == value {
			return true
		}
	}
	return false
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (h *Hearts) playerHasSuit(playerIdx int, design int) bool {
	p := h.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// playerHasNonHeart プレイヤーがハート以外のカードを持っているか
func (h *Hearts) playerHasNonHeart(playerIdx int) bool {
	p := h.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() != CardDesignHeart {
			return true
		}
	}
	return false
}

// playerHasNonPointCard プレイヤーがポイントカード以外のカードを持っているか
func (h *Hearts) playerHasNonPointCard(player *HeartsPlayer) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		if !isPointCard(player.GetCard(i), h.config.OmnibusJD) {
			return true
		}
	}
	return false
}

// playerTookCard プレイヤーが獲得したトリックに特定のカードが含まれるか確認する
func (h *Hearts) playerTookCard(playerIdx, design, value int) bool {
	for _, trick := range h.players[playerIdx].GetTricksTaken() {
		for _, card := range trick {
			if card.GetDesign() == design && card.GetValue() == value {
				return true
			}
		}
	}
	return false
}

// isPointCard ポイントカードかどうか (ハートまたはQ♠、オムニバス時はJ♦も含む)
func isPointCard(card *Card, omnibusJD bool) bool {
	return card.GetDesign() == CardDesignHeart ||
		(card.GetDesign() == CardDesignSpade && card.GetValue() == 12) ||
		(omnibusJD && card.GetDesign() == CardDesignDiamond && card.GetValue() == 11)
}

// cardPoints カードのポイント値
func cardPoints(card *Card, omnibusJD bool) int {
	switch card.GetDesign() {
	case CardDesignHeart:
		return 1
	case CardDesignSpade:
		if card.GetValue() == 12 {
			return 13
		}
	case CardDesignDiamond:
		if omnibusJD && card.GetValue() == 11 {
			return -10
		}
	}
	return 0
}

// isPenaltyCard ペナルティカードかどうか (ハートまたはQ♠のみ。J♦は含まない)
func isPenaltyCard(card *Card) bool {
	if card.GetDesign() == CardDesignHeart {
		return true
	}
	return card.GetDesign() == CardDesignSpade && card.GetValue() == 12
}

// trickWinner トリックの勝者を決定する
func (h *Hearts) trickWinner() int {
	return ResolveTrickWinner(h.currentTrick, -1, nil)
}

// passTarget パス先のプレイヤーインデックスを計算する
func (h *Hearts) passTarget(from int, dir HeartsPassDirection) int {
	switch dir {
	case HeartsPassLeft:
		return (from + 1) % HeartsPlayerCnt
	case HeartsPassRight:
		return (from + HeartsPlayerCnt - 1) % HeartsPlayerCnt
	case HeartsPassAcross:
		return (from + 2) % HeartsPlayerCnt
	default:
		return from
	}
}

// passDirectionStr パス方向の文字列表現
func (h *Hearts) passDirectionStr(dir HeartsPassDirection) string {
	switch dir {
	case HeartsPassLeft:
		return "left"
	case HeartsPassRight:
		return "right"
	case HeartsPassAcross:
		return "across"
	default:
		return "none"
	}
}

// sortAllHands 全プレイヤーの手札をソートする
func (h *Hearts) sortAllHands() {
	for _, p := range h.players {
		sortHand(p)
	}
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func sortHand(p *HeartsPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// cardStr カードの文字列表現
func cardStr(card *Card) string {
	suits := map[int]string{
		CardDesignSpade:   "♠",
		CardDesignClover:  "♣",
		CardDesignHeart:   "♥",
		CardDesignDiamond: "♦",
	}
	values := map[int]string{
		1: "A", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7",
		8: "8", 9: "9", 10: "10", 11: "J", 12: "Q", 13: "K",
	}
	s, ok := suits[card.GetDesign()]
	if !ok {
		s = "?"
	}
	v, ok := values[card.GetValue()]
	if !ok {
		v = "?"
	}
	return s + v
}

// GetHint ヒントを取得する
func (h *Hearts) GetHint() *HeartsHint {
	if h.phase == HeartsPhasePass && !h.passReady[0] {
		return h.getPassHint()
	}
	if h.phase == HeartsPhasePlay && h.currentPlayerIdx == 0 {
		validIndices := h.getValidPlayIndices(0)
		if len(validIndices) == 0 {
			return nil
		}
		idx := h.cpuPlayHard(0, validIndices)
		return &HeartsHint{CardIndices: []int{idx}, Reason: h.playHintReason(idx, validIndices)}
	}
	return nil
}

// getPassHint パスフェーズのヒントを生成する
func (h *Hearts) getPassHint() *HeartsHint {
	indices := h.scorePassIndices(h.players[0])
	return &HeartsHint{CardIndices: indices, Reason: "pass_high_risk_cards"}
}

// scorePassIndices パス対象カードのインデックスをスコア順に返す
func (h *Hearts) scorePassIndices(player *HeartsPlayer) []int {
	suitCounts := map[int]int{}
	for i := 0; i < player.GetCardsSize(); i++ {
		suitCounts[player.GetCard(i).GetDesign()]++
	}

	type cardScore struct {
		idx   int
		score int
	}
	scores := make([]cardScore, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		score := card.GetValue()
		if card.GetDesign() == CardDesignSpade && card.GetValue() == 12 {
			score += 100
		}
		if card.GetDesign() == CardDesignSpade && card.GetValue() >= 11 {
			score += 50
		}
		if card.GetDesign() == CardDesignHeart {
			score += 20
		}
		if suitCounts[card.GetDesign()] <= 3 && card.GetDesign() != CardDesignHeart {
			score += 30
		}
		if h.config.OmnibusJD && card.GetDesign() == CardDesignDiamond && card.GetValue() == 11 {
			score = -100
		}
		scores[i] = cardScore{idx: i, score: score}
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })

	indices := make([]int, HeartsPassCardCount)
	for i := 0; i < HeartsPassCardCount; i++ {
		indices[i] = scores[i].idx
	}
	return indices
}

// playHintReason プレイヒントの理由を判定する
func (h *Hearts) playHintReason(chosenIdx int, validIndices []int) string {
	player := h.players[0]
	card := player.GetCard(chosenIdx)

	if len(h.currentTrick) == 0 {
		return "lead_low"
	}

	leadSuit := h.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}

	if card.GetDesign() == CardDesignSpade && card.GetValue() == 12 {
		return "discard_queen_spades"
	}
	if card.GetDesign() == CardDesignHeart {
		return "discard_hearts"
	}
	return "discard_high"
}

// --- CPU AI ---

// cpuSelectPassCards CPU がパスするカードを選択する
func (h *Hearts) cpuSelectPassCards(playerIdx int) []*Card {
	player := h.players[playerIdx]

	switch h.config.CpuDifficulty {
	case HeartsCpuDifficultyHard:
		return h.cpuPassHard(player)
	case HeartsCpuDifficultyNormal:
		return h.cpuPassNormal(player)
	default:
		return h.cpuPassEasy(player)
	}
}

// cpuPassEasy ランダムに3枚選択
func (h *Hearts) cpuPassEasy(player *HeartsPlayer) []*Card {
	indices := make([]int, player.GetCardsSize())
	for i := range indices {
		indices[i] = i
	}
	rand.Shuffle(len(indices), func(i, j int) { indices[i], indices[j] = indices[j], indices[i] })
	return player.RemoveCards(indices[:HeartsPassCardCount])
}

// cpuPassNormal 高いカード・危険なカードを優先的にパス
func (h *Hearts) cpuPassNormal(player *HeartsPlayer) []*Card {
	type cardScore struct {
		idx   int
		score int
	}
	scores := make([]cardScore, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		score := card.GetValue()
		// Q♠は最優先で渡す
		if card.GetDesign() == CardDesignSpade && card.GetValue() == 12 {
			score += 100
		}
		// A♠, K♠も危険
		if card.GetDesign() == CardDesignSpade && card.GetValue() >= 11 {
			score += 50
		}
		// 高いハートは得点リスク
		if card.GetDesign() == CardDesignHeart {
			score += 20
		}
		// オムニバス時はJ♦を渡さない（有利なカード）
		if h.config.OmnibusJD && card.GetDesign() == CardDesignDiamond && card.GetValue() == 11 {
			score = -100
		}
		scores[i] = cardScore{idx: i, score: score}
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })

	indices := make([]int, HeartsPassCardCount)
	for i := 0; i < HeartsPassCardCount; i++ {
		indices[i] = scores[i].idx
	}
	return player.RemoveCards(indices)
}

// cpuPassHard 戦略的なパス (ボイド作成を意識)
func (h *Hearts) cpuPassHard(player *HeartsPlayer) []*Card {
	indices := h.scorePassIndices(player)
	return player.RemoveCards(indices)
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (h *Hearts) cpuSelectPlayCard(playerIdx int) int {
	validIndices := h.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0 // フォールバック
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch h.config.CpuDifficulty {
	case HeartsCpuDifficultyHard:
		return h.cpuPlayHard(playerIdx, validIndices)
	case HeartsCpuDifficultyNormal:
		return h.cpuPlayNormal(playerIdx, validIndices)
	default:
		return h.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (h *Hearts) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal ポイントを避けるプレイ
func (h *Hearts) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := h.players[playerIdx]

	if len(h.currentTrick) == 0 {
		// リード: 最も低いカードを出す
		bestIdx := validIndices[0]
		bestVal := player.GetCard(validIndices[0]).GetValue()
		for _, idx := range validIndices[1:] {
			card := player.GetCard(idx)
			if card.GetValue() < bestVal {
				bestVal = card.GetValue()
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// フォロー: ポイントを取らないように、可能な限り低いカードを出す
	leadSuit := h.currentTrick[0].Card.GetDesign()
	highestInTrick := 0
	for _, tc := range h.currentTrick {
		if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
			highestInTrick = tc.Card.GetValue()
		}
	}

	// リードスートを持っているか判定
	hasLeadSuit := false
	for _, idx := range validIndices {
		if player.GetCard(idx).GetDesign() == leadSuit {
			hasLeadSuit = true
			break
		}
	}

	if hasLeadSuit {
		// フォロースート: トリックに勝たないカード(アンダーカード)を探す
		underCards := []int{}
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			if card.GetDesign() == leadSuit && card.GetValue() < highestInTrick {
				underCards = append(underCards, idx)
			}
		}
		if len(underCards) > 0 {
			// アンダーカードのうち最も高いカードを出す(ポイントを避けつつ高いカードを消費)
			bestIdx := underCards[0]
			for _, idx := range underCards[1:] {
				if player.GetCard(idx).GetValue() > player.GetCard(bestIdx).GetValue() {
					bestIdx = idx
				}
			}
			return bestIdx
		}
		// アンダーカードがない場合、最も低いリードスートカードを出す
		bestIdx := validIndices[0]
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			if card.GetDesign() == leadSuit && (player.GetCard(bestIdx).GetDesign() != leadSuit || card.GetValue() < player.GetCard(bestIdx).GetValue()) {
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// ボイドの場合: ペナルティカードを最優先で捨てる、次に高いカード
	// オムニバス時はJ♦を捨てない（有利なカードのため）
	isOmnibusJD := func(c *Card) bool {
		return h.config.OmnibusJD && c.GetDesign() == CardDesignDiamond && c.GetValue() == 11
	}
	bestIdx := validIndices[0]
	for _, idx := range validIndices[1:] {
		card := player.GetCard(idx)
		bestCard := player.GetCard(bestIdx)
		// J♦はオムニバス時に有利なので最後の選択肢にする
		if isOmnibusJD(card) {
			continue
		}
		if isOmnibusJD(bestCard) {
			bestIdx = idx
			continue
		}
		cardPenalty := isPenaltyCard(card)
		bestPenalty := isPenaltyCard(bestCard)
		if cardPenalty && !bestPenalty {
			bestIdx = idx
		} else if cardPenalty == bestPenalty && card.GetValue() > bestCard.GetValue() {
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuPlayHard 高度な戦略プレイ
func (h *Hearts) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := h.players[playerIdx]

	if len(h.currentTrick) == 0 {
		// リード: 最も低い非ハートカードを出す。ハートしかない場合は最低ハート
		bestIdx := validIndices[0]
		bestVal := player.GetCard(validIndices[0]).GetValue()
		bestIsHeart := player.GetCard(validIndices[0]).GetDesign() == CardDesignHeart
		for _, idx := range validIndices[1:] {
			card := player.GetCard(idx)
			isHeart := card.GetDesign() == CardDesignHeart
			if bestIsHeart && !isHeart {
				bestIdx = idx
				bestVal = card.GetValue()
				bestIsHeart = false
			} else if isHeart == bestIsHeart && card.GetValue() < bestVal {
				bestIdx = idx
				bestVal = card.GetValue()
			}
		}
		return bestIdx
	}

	// フォロー or ディスカード
	leadSuit := h.currentTrick[0].Card.GetDesign()
	highestInTrick := 0
	trickPoints := 0
	for _, tc := range h.currentTrick {
		if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
			highestInTrick = tc.Card.GetValue()
		}
		trickPoints += cardPoints(tc.Card, h.config.OmnibusJD)
	}

	hasLeadSuit := false
	for _, idx := range validIndices {
		if player.GetCard(idx).GetDesign() == leadSuit {
			hasLeadSuit = true
			break
		}
	}

	if hasLeadSuit {
		// フォロースート: トリックに勝たないカード(アンダーカード)を探す
		underCards := []int{}
		overCards := []int{}
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			if card.GetDesign() != leadSuit {
				continue
			}
			if card.GetValue() < highestInTrick {
				underCards = append(underCards, idx)
			} else {
				overCards = append(overCards, idx)
			}
		}
		if len(underCards) > 0 {
			// アンダーカードのうち最も高いカードを出す
			bestIdx := underCards[0]
			for _, idx := range underCards[1:] {
				if player.GetCard(idx).GetValue() > player.GetCard(bestIdx).GetValue() {
					bestIdx = idx
				}
			}
			return bestIdx
		}
		// アンダーカードがない場合、最も低いオーバーカードを出す
		bestIdx := overCards[0]
		for _, idx := range overCards[1:] {
			if player.GetCard(idx).GetValue() < player.GetCard(bestIdx).GetValue() {
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// ディスカード: Q♠を最優先、次にハートの高いカード、次に高いカード
	// オムニバス時はJ♦を捨てない（有利なカードのため）
	bestIdx := validIndices[0]
	bestScore := -1
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		score := card.GetValue()
		if card.GetDesign() == CardDesignSpade && card.GetValue() == 12 {
			score += 200
		} else if card.GetDesign() == CardDesignHeart {
			score += 100
		}
		// J♦はオムニバス時に有利なので捨てない
		if h.config.OmnibusJD && card.GetDesign() == CardDesignDiamond && card.GetValue() == 11 {
			score = -100
		}
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (h *Hearts) getValidPlayIndices(playerIdx int) []int {
	player := h.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return h.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// heartsJSON is the JSON wire format for Hearts.
type heartsJSON struct {
	TrumpCards       *TrumpCards              `json:"tc"`
	Players          []*HeartsPlayer          `json:"ps"`
	Config           HeartsConfig             `json:"cf"`
	Phase            HeartsPhase              `json:"ph"`
	RoundNumber      int                      `json:"rn"`
	TrickNumber      int                      `json:"tn"`
	CurrentPlayerIdx int                      `json:"ci"`
	CurrentTrick     []*TrickCard             `json:"ct"`
	HeartsBroken     bool                     `json:"hb"`
	PassedCards      [HeartsPlayerCnt][]*Card `json:"pc"`
	PassReady        [HeartsPlayerCnt]bool    `json:"pr"`
	LeadPlayerIdx    int                      `json:"li"`
	GameEndFlag      bool                     `json:"ge"`
	WinnerIdx        int                      `json:"wi"`
	ActionLog        []*ActionLogEntry        `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (h *Hearts) MarshalJSON() ([]byte, error) {
	return json.Marshal(heartsJSON{
		TrumpCards:       h.trumpCards,
		Players:          h.players,
		Config:           h.config,
		Phase:            h.phase,
		RoundNumber:      h.roundNumber,
		TrickNumber:      h.trickNumber,
		CurrentPlayerIdx: h.currentPlayerIdx,
		CurrentTrick:     h.currentTrick,
		HeartsBroken:     h.heartsBroken,
		PassedCards:      h.passedCards,
		PassReady:        h.passReady,
		LeadPlayerIdx:    h.leadPlayerIdx,
		GameEndFlag:      h.gameEndFlag,
		WinnerIdx:        h.winnerIdx,
		ActionLog:        h.actionLog,
	})
}

// heartsMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const heartsMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (h *Hearts) UnmarshalJSON(data []byte) error {
	var j heartsJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > heartsMaxSliceLen || len(j.CurrentTrick) > heartsMaxSliceLen ||
		len(j.ActionLog) > heartsMaxSliceLen {
		return fmt.Errorf("hearts: input array exceeds maximum allowed size")
	}
	for i := range j.PassedCards {
		if len(j.PassedCards[i]) > heartsMaxSliceLen {
			return fmt.Errorf("hearts: input array exceeds maximum allowed size")
		}
	}
	h.trumpCards = j.TrumpCards
	if h.trumpCards == nil {
		h.trumpCards = NewTrumpCards(0)
	}
	h.players = j.Players
	if h.players == nil {
		h.players = make([]*HeartsPlayer, 0)
	}
	h.config = j.Config
	h.phase = j.Phase
	h.roundNumber = j.RoundNumber
	h.trickNumber = j.TrickNumber
	h.currentPlayerIdx = j.CurrentPlayerIdx
	h.currentTrick = j.CurrentTrick
	if h.currentTrick == nil {
		h.currentTrick = make([]*TrickCard, 0)
	}
	h.heartsBroken = j.HeartsBroken
	h.passedCards = j.PassedCards
	h.passReady = j.PassReady
	h.leadPlayerIdx = j.LeadPlayerIdx
	h.gameEndFlag = j.GameEndFlag
	h.winnerIdx = j.WinnerIdx
	h.actionLog = j.ActionLog
	if h.actionLog == nil {
		h.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
