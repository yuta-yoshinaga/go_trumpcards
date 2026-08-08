package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// TwoTenJackPlayerCnt ツーテンジャックプレイヤー数
const TwoTenJackPlayerCnt = 4

// TwoTenJackHandSize 各プレイヤーの手札枚数
const TwoTenJackHandSize = 13

// TwoTenJackRoundPool ラウンドの得点プール (Ace*4 + 10*40 + Jack*4 + 最終トリック2)
const TwoTenJackRoundPool = 50

// TwoTenJackMakeThreshold 宣言チームが契約達成に必要な点数 (50の過半数)
const TwoTenJackMakeThreshold = 26

// TwoTenJackPhase ゲームフェーズ
type TwoTenJackPhase int

// TwoTenJackのフェーズ定数
const (
	// TwoTenJackPhaseDeclare トランプ宣言フェーズ
	TwoTenJackPhaseDeclare TwoTenJackPhase = 0
	// TwoTenJackPhasePlay トリックプレイフェーズ
	TwoTenJackPhasePlay TwoTenJackPhase = 1
	// TwoTenJackPhaseTrickEnd トリック終了フェーズ
	TwoTenJackPhaseTrickEnd TwoTenJackPhase = 2
	// TwoTenJackPhaseRoundEnd ラウンド終了フェーズ
	TwoTenJackPhaseRoundEnd TwoTenJackPhase = 3
	// TwoTenJackPhaseGameEnd ゲーム終了フェーズ
	TwoTenJackPhaseGameEnd TwoTenJackPhase = 4
)

// TwoTenJackHint ヒント情報
type TwoTenJackHint struct {
	CardIndex *int   // 推奨カードインデックス (宣言時nil)
	TrumpSuit *int   // 推奨トランプスート (プレイ時nil)
	Reason    string // ヒント理由キー
}

// TwoTenJack ツーテンジャックゲームクラス
type TwoTenJack struct {
	trumpCards       *TrumpCards
	players          []*TwoTenJackPlayer
	config           TwoTenJackConfig
	phase            TwoTenJackPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	declarerIdx      int
	trumpSuit        int // -1 = 未宣言
	gameEndFlag      bool
	winnerTeam       int // -1 = 未確定, 0 = (0,2), 1 = (1,3)
	actionLogBase
}

// NewTwoTenJack コンストラクタ
func NewTwoTenJack(trumpCards *TrumpCards, players []*TwoTenJackPlayer, config TwoTenJackConfig) *TwoTenJack {
	return &TwoTenJack{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerTeam:  -1,
		roundNumber: 1,
		trumpSuit:   -1,
		declarerIdx: 0,
	}
}

// NewDefaultTwoTenJack returns TwoTenJack with the standard 4-player setup (1 human, 3 CPU)
// and DefaultTwoTenJackConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultTwoTenJack() *TwoTenJack {
	players := []*TwoTenJackPlayer{
		NewTwoTenJackPlayer(true),
		NewTwoTenJackPlayer(false),
		NewTwoTenJackPlayer(false),
		NewTwoTenJackPlayer(false),
	}
	return NewTwoTenJack(NewTrumpCards(0), players, DefaultTwoTenJackConfig())
}

// Reset ゲーム初期化
func (t *TwoTenJack) Reset() {
	t.gameEndFlag = false
	t.winnerTeam = -1
	t.roundNumber = 1
	t.trickNumber = 0
	t.currentTrick = nil
	t.leadPlayerIdx = -1
	t.currentPlayerIdx = -1
	t.declarerIdx = 0
	t.trumpSuit = -1
	t.actionLog = nil

	for _, p := range t.players {
		p.SetRoundScore(0)
		p.SetCumulativeScore(0)
		p.ResetTricks()
		p.Reset()
		p.SetIsFinished(false)
	}

	t.trumpCards.Shuffle()
	dealAllCards(t.trumpCards, t.players)
	t.sortAllHands()

	t.phase = TwoTenJackPhaseDeclare
}

// NextRound 次のラウンドを開始する
func (t *TwoTenJack) NextRound() {
	if t.phase != TwoTenJackPhaseRoundEnd {
		return
	}

	t.roundNumber++
	t.trickNumber = 0
	t.currentTrick = nil
	t.leadPlayerIdx = -1
	t.currentPlayerIdx = -1
	t.trumpSuit = -1
	t.declarerIdx = (t.declarerIdx + 1) % TwoTenJackPlayerCnt

	for _, p := range t.players {
		p.ResetRound()
	}

	t.trumpCards.Shuffle()
	dealAllCards(t.trumpCards, t.players)
	t.sortAllHands()

	t.phase = TwoTenJackPhaseDeclare
}

// PlayerDeclareTrump 人間プレイヤーがトランプを宣言する
func (t *TwoTenJack) PlayerDeclareTrump(suit int) error {
	if t.gameEndFlag {
		return ErrGameEnded
	}
	if t.phase != TwoTenJackPhaseDeclare {
		return ErrWrongPhase
	}
	if t.declarerIdx < 0 || t.declarerIdx >= TwoTenJackPlayerCnt {
		return ErrNotHumanTurn
	}
	if !t.players[t.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !isValidTwoTenJackSuit(suit) {
		return NewDomainError(ErrInvalidPlay, "トランプスートが不正です")
	}

	t.trumpSuit = suit
	t.appendLog(t.declarerIdx, "declare_trump", fmt.Sprintf("%s declares trump: %s", playerName(t.players, t.declarerIdx), twoTenJackSuitName(suit)), nil)
	t.startPlayPhase()
	return nil
}

// CpuDeclareTrump 現在の宣言者がCPUの場合にトランプを宣言する
func (t *TwoTenJack) CpuDeclareTrump() {
	if t.gameEndFlag || t.phase != TwoTenJackPhaseDeclare {
		return
	}
	if t.declarerIdx < 0 || t.declarerIdx >= TwoTenJackPlayerCnt {
		return
	}
	if t.players[t.declarerIdx].GetIsHuman() {
		return
	}
	suit := t.cpuSelectTrump(t.declarerIdx)
	t.trumpSuit = suit
	t.appendLog(t.declarerIdx, "declare_trump", fmt.Sprintf("%s declares trump: %s", playerName(t.players, t.declarerIdx), twoTenJackSuitName(suit)), nil)
	t.startPlayPhase()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (t *TwoTenJack) PlayerPlay(cardIndex int) error {
	if t.gameEndFlag {
		return ErrGameEnded
	}
	if t.phase != TwoTenJackPhasePlay {
		return ErrWrongPhase
	}
	if t.currentPlayerIdx < 0 || t.currentPlayerIdx >= TwoTenJackPlayerCnt {
		return ErrNotHumanTurn
	}
	if !t.players[t.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := t.players[t.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := t.validatePlay(t.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	t.playCard(t.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (t *TwoTenJack) CpuPlay() {
	if t.gameEndFlag || t.phase != TwoTenJackPhasePlay {
		return
	}
	if t.currentPlayerIdx < 0 || t.currentPlayerIdx >= TwoTenJackPlayerCnt {
		return
	}
	if t.players[t.currentPlayerIdx].GetIsHuman() {
		return
	}
	player := t.players[t.currentPlayerIdx]
	cardIdx := t.cpuSelectPlayCard(t.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	t.playCard(t.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (t *TwoTenJack) ResolveTrick() {
	if t.phase != TwoTenJackPhaseTrickEnd || len(t.currentTrick) != TwoTenJackPlayerCnt {
		return
	}

	winnerIdx := t.trickWinner()
	trickCards := make([]*Card, len(t.currentTrick))
	for i, tc := range t.currentTrick {
		trickCards[i] = tc.Card
	}

	t.players[winnerIdx].AddTrick(trickCards)

	winnerName := playerName(t.players, winnerIdx)
	t.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d", winnerName, t.trickNumber), trickCards)

	t.leadPlayerIdx = winnerIdx

	if t.trickNumber >= TwoTenJackHandSize {
		t.phase = TwoTenJackPhaseRoundEnd
	} else {
		t.phase = TwoTenJackPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (t *TwoTenJack) NextTrick() {
	if t.phase != TwoTenJackPhaseTrickEnd {
		return
	}
	t.currentTrick = nil
	t.currentPlayerIdx = t.leadPlayerIdx
	t.trickNumber++
	t.phase = TwoTenJackPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (t *TwoTenJack) ScoreRound() {
	if t.phase != TwoTenJackPhaseRoundEnd {
		return
	}

	// 各プレイヤーの点札合計
	playerPoints := make([]int, TwoTenJackPlayerCnt)
	for i := 0; i < TwoTenJackPlayerCnt; i++ {
		playerPoints[i] = t.players[i].GetCapturedPointCards()
	}

	// 最終トリック勝者に+2
	lastWinner := t.leadPlayerIdx
	if lastWinner >= 0 && lastWinner < TwoTenJackPlayerCnt {
		playerPoints[lastWinner] += 2
	}

	// チーム合計: team0 = (0,2), team1 = (1,3)
	teamPoints := [2]int{
		playerPoints[0] + playerPoints[2],
		playerPoints[1] + playerPoints[3],
	}

	// 宣言チーム
	declaringTeam := t.declarerIdx % 2

	// ラウンドスコア計算 (マージンスコア)
	// 宣言チームが26点以上取れば契約達成: 宣言チーム score = teamPoints - 25, 守備チーム 0
	// 達成失敗: 守備チーム score = 26 - teamPoints(宣言), 宣言チーム 0
	var teamRoundScore [2]int
	declTeamPts := teamPoints[declaringTeam]
	if declTeamPts >= TwoTenJackMakeThreshold {
		teamRoundScore[declaringTeam] = declTeamPts - (TwoTenJackMakeThreshold - 1)
		teamRoundScore[1-declaringTeam] = 0
	} else {
		teamRoundScore[declaringTeam] = 0
		teamRoundScore[1-declaringTeam] = TwoTenJackMakeThreshold - declTeamPts
	}

	// プレイヤーごとに roundScore を設定 (同じチームメンバーは同じ roundScore)
	for i := 0; i < TwoTenJackPlayerCnt; i++ {
		team := i % 2
		t.players[i].SetRoundScore(teamRoundScore[team])
	}

	t.appendLog(-1, "round_score", fmt.Sprintf("team0=%d team1=%d (declarer=team%d, declPts=%d)",
		teamRoundScore[0], teamRoundScore[1], declaringTeam, declTeamPts), nil)

	// 累積スコアに加算
	for i := 0; i < TwoTenJackPlayerCnt; i++ {
		t.players[i].CommitRoundScore()
	}

	for i := 0; i < TwoTenJackPlayerCnt; i++ {
		t.appendLog(i, "cumulative_score", fmt.Sprintf("%s: total=%d",
			playerName(t.players, i), t.players[i].GetCumulativeScore()), nil)
	}

	t.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (t *TwoTenJack) GetPhase() TwoTenJackPhase { return t.phase }

// SetPhase フェーズ設定 (テスト用)
func (t *TwoTenJack) SetPhase(phase TwoTenJackPhase) { t.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (t *TwoTenJack) GetRoundNumber() int { return t.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (t *TwoTenJack) SetRoundNumber(n int) { t.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (t *TwoTenJack) GetTrickNumber() int { return t.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (t *TwoTenJack) SetTrickNumber(n int) { t.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (t *TwoTenJack) GetCurrentPlayerIdx() int { return t.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (t *TwoTenJack) SetCurrentPlayerIdx(idx int) { t.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (t *TwoTenJack) GetCurrentTrick() []*TrickCard { return t.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (t *TwoTenJack) SetCurrentTrick(trick []*TrickCard) { t.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (t *TwoTenJack) GetLeadPlayerIdx() int { return t.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (t *TwoTenJack) SetLeadPlayerIdx(idx int) { t.leadPlayerIdx = idx }

// GetDeclarerIdx 宣言者インデックス取得
func (t *TwoTenJack) GetDeclarerIdx() int { return t.declarerIdx }

// SetDeclarerIdx 宣言者インデックス設定 (テスト用)
func (t *TwoTenJack) SetDeclarerIdx(idx int) { t.declarerIdx = idx }

// GetTrumpSuit トランプスート取得 (-1 = 未宣言)
func (t *TwoTenJack) GetTrumpSuit() int { return t.trumpSuit }

// SetTrumpSuit トランプスート設定 (テスト用)
func (t *TwoTenJack) SetTrumpSuit(suit int) { t.trumpSuit = suit }

// GetGameEndFlag ゲーム終了フラグ取得
func (t *TwoTenJack) GetGameEndFlag() bool { return t.gameEndFlag }

// GetWinnerTeam 勝利チームインデックス取得 (-1 = 未確定)
func (t *TwoTenJack) GetWinnerTeam() int { return t.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (t *TwoTenJack) GetPlayerCnt() int { return len(t.players) }

// GetPlayer プレイヤー取得
func (t *TwoTenJack) GetPlayer(i int) *TwoTenJackPlayer {
	if i < 0 || i >= len(t.players) {
		return nil
	}
	return t.players[i]
}

// IsHumanTurn 現在の手番が人間かどうか
func (t *TwoTenJack) IsHumanTurn() bool {
	return isHumanTurn(t.players, t.currentPlayerIdx)
}

// IsHumanDeclareTurn 現在の宣言手番が人間かどうか
func (t *TwoTenJack) IsHumanDeclareTurn() bool {
	if t.declarerIdx < 0 || t.declarerIdx >= len(t.players) {
		return false
	}
	return t.players[t.declarerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (t *TwoTenJack) GetConfig() TwoTenJackConfig { return t.config }

// SetConfig 設定変更
func (t *TwoTenJack) SetConfig(cfg TwoTenJackConfig) { t.config = cfg }

// --- Private methods ---

// startPlayPhase プレイフェーズ開始: 宣言者がリード
func (t *TwoTenJack) startPlayPhase() {
	t.phase = TwoTenJackPhasePlay
	t.leadPlayerIdx = t.declarerIdx
	t.currentPlayerIdx = t.declarerIdx
	t.trickNumber = 1
	t.currentTrick = nil
}

// playCard カードをプレイする共通処理
func (t *TwoTenJack) playCard(playerIdx int, card *Card) {
	t.currentTrick = append(t.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	t.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(t.players, playerIdx), cardStr(card)), []*Card{card})

	if len(t.currentTrick) == TwoTenJackPlayerCnt {
		t.phase = TwoTenJackPhaseTrickEnd
	} else {
		t.currentPlayerIdx = (t.currentPlayerIdx + 1) % TwoTenJackPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
func (t *TwoTenJack) validatePlay(playerIdx int, card *Card) error {
	if len(t.currentTrick) == 0 {
		return nil
	}
	leadSuit := t.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && t.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (t *TwoTenJack) playerHasSuit(playerIdx int, design int) bool {
	p := t.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する
func (t *TwoTenJack) trickWinner() int {
	return ResolveTrickWinner(t.currentTrick, t.trumpSuit, ttjEffectiveValue)
}

// ttjEffectiveValue トリック比較用の値。A=14, K=13, Q=12, J=11, 10=10, ..., 2=2
func ttjEffectiveValue(c *Card) int {
	v := c.GetValue()
	if v == 1 {
		return 14
	}
	return v
}

// checkGameEnd ゲーム終了判定
func (t *TwoTenJack) checkGameEnd() {
	team0 := t.players[0].GetCumulativeScore() + t.players[2].GetCumulativeScore()
	team1 := t.players[1].GetCumulativeScore() + t.players[3].GetCumulativeScore()

	if team0 < t.config.PointLimit && team1 < t.config.PointLimit {
		return
	}
	t.gameEndFlag = true
	t.phase = TwoTenJackPhaseGameEnd
	if team0 >= team1 {
		t.winnerTeam = 0
	} else {
		t.winnerTeam = 1
	}
	t.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game! (team0=%d, team1=%d)", t.winnerTeam, team0, team1), nil)
}

// sortAllHands 全プレイヤーの手札をソートする
func (t *TwoTenJack) sortAllHands() {
	for _, p := range t.players {
		twoTenJackSortHand(p)
	}
}

// twoTenJackSortHand プレイヤーの手札をスート→値の順にソートする
func twoTenJackSortHand(p *TwoTenJackPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// twoTenJackSuitName スート名を返す
func twoTenJackSuitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "Spade"
	case CardDesignHeart:
		return "Heart"
	case CardDesignDiamond:
		return "Diamond"
	case CardDesignClover:
		return "Club"
	default:
		return "?"
	}
}

// GetHint ヒントを取得する
func (t *TwoTenJack) GetHint() *TwoTenJackHint {
	if t.phase == TwoTenJackPhaseDeclare && t.declarerIdx == 0 && t.players[0].GetIsHuman() {
		suit := t.cpuTrumpHard(0)
		return &TwoTenJackHint{TrumpSuit: &suit, Reason: "strategic_trump"}
	}
	if t.phase == TwoTenJackPhasePlay && t.currentPlayerIdx == 0 && t.players[0].GetIsHuman() {
		validIndices := t.getValidPlayIndices(0)
		if len(validIndices) == 0 {
			return nil
		}
		idx := t.cpuPlayHard(0, validIndices)
		return &TwoTenJackHint{CardIndex: &idx, Reason: t.playHintReason(idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由を判定する
func (t *TwoTenJack) playHintReason(chosenIdx int) string {
	player := t.players[0]
	card := player.GetCard(chosenIdx)
	if len(t.currentTrick) == 0 {
		return "lead"
	}
	leadSuit := t.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == t.trumpSuit {
		return "trump_cut"
	}
	return "discard"
}

// --- CPU AI ---

// cpuSelectTrump CPUがトランプスートを選択する
func (t *TwoTenJack) cpuSelectTrump(playerIdx int) int {
	switch t.config.CpuDifficulty {
	case TwoTenJackCpuDifficultyHard:
		return t.cpuTrumpHard(playerIdx)
	case TwoTenJackCpuDifficultyNormal:
		return t.cpuTrumpNormal(playerIdx)
	default:
		return t.cpuTrumpEasy()
	}
}

// twoTenJackSuits 宣言可能なスート一覧
var twoTenJackSuits = [4]int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}

// isValidTwoTenJackSuit スートが有効か判定
func isValidTwoTenJackSuit(suit int) bool {
	for _, s := range twoTenJackSuits {
		if s == suit {
			return true
		}
	}
	return false
}

// cpuTrumpEasy ランダムなスート
func (t *TwoTenJack) cpuTrumpEasy() int {
	return twoTenJackSuits[rand.Intn(4)]
}

// cpuTrumpNormal 最も多いスートを選ぶ
func (t *TwoTenJack) cpuTrumpNormal(playerIdx int) int {
	counts := t.suitCounts(playerIdx)
	best := twoTenJackSuits[0]
	for _, s := range twoTenJackSuits[1:] {
		if counts[s] > counts[best] {
			best = s
		}
	}
	return best
}

// cpuTrumpHard 最も多いスート + 点札を考慮
func (t *TwoTenJack) cpuTrumpHard(playerIdx int) int {
	counts := t.suitCounts(playerIdx)
	honors := t.suitHonors(playerIdx)
	best := twoTenJackSuits[0]
	bestScore := -1
	for _, s := range twoTenJackSuits {
		score := counts[s]*2 + honors[s]*3
		if score > bestScore {
			bestScore = score
			best = s
		}
	}
	return best
}

// suitCounts スートごとのカード枚数 (キーはCardDesign*の値)
func (t *TwoTenJack) suitCounts(playerIdx int) map[int]int {
	counts := make(map[int]int, 4)
	p := t.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		counts[p.GetCard(i).GetDesign()]++
	}
	return counts
}

// suitHonors スートごとの点札(A/10/J)枚数 (キーはCardDesign*の値)
func (t *TwoTenJack) suitHonors(playerIdx int) map[int]int {
	honors := make(map[int]int, 4)
	p := t.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if TwoTenJackCardPoints(c) > 0 {
			honors[c.GetDesign()]++
		}
	}
	return honors
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (t *TwoTenJack) cpuSelectPlayCard(playerIdx int) int {
	validIndices := t.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}
	switch t.config.CpuDifficulty {
	case TwoTenJackCpuDifficultyHard:
		return t.cpuPlayHard(playerIdx, validIndices)
	case TwoTenJackCpuDifficultyNormal:
		return t.cpuPlayNormal(playerIdx, validIndices)
	default:
		return t.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (t *TwoTenJack) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal 簡易戦略: 勝てるなら高い点札を取り、負けなら低い非点札を捨てる
func (t *TwoTenJack) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := t.players[playerIdx]
	if len(t.currentTrick) == 0 {
		// リード: 最も低い非点札
		return pickLowestNonPoint(player, validIndices)
	}
	// フォロー: 現時点の勝者と自分のチームを判定
	currentWinnerIdx := t.currentTrickWinnerIdx()
	weCanWin := t.teamOf(currentWinnerIdx) == t.teamOf(playerIdx)
	if weCanWin {
		// 味方が勝っている: 点札を捨てるチャンス
		return pickHighestPoint(player, validIndices)
	}
	// 負けている: 可能なら勝てるカードを出す
	if idx, ok := pickLowestWinning(t, player, validIndices); ok {
		return idx
	}
	// 勝てない: 最も低い非点札を捨てる
	return pickLowestNonPoint(player, validIndices)
}

// cpuPlayHard より高度な戦略
func (t *TwoTenJack) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := t.players[playerIdx]
	if len(t.currentTrick) == 0 {
		// リード: 長いスートの高札を出す
		counts := t.suitCounts(playerIdx)
		bestIdx := validIndices[0]
		bestScore := -1
		for _, idx := range validIndices {
			c := player.GetCard(idx)
			score := counts[c.GetDesign()]*2 + ttjEffectiveValue(c)
			if score > bestScore {
				bestScore = score
				bestIdx = idx
			}
		}
		return bestIdx
	}
	currentWinnerIdx := t.currentTrickWinnerIdx()
	weCanWin := t.teamOf(currentWinnerIdx) == t.teamOf(playerIdx)
	if weCanWin {
		return pickHighestPoint(player, validIndices)
	}
	if idx, ok := pickLowestWinning(t, player, validIndices); ok {
		return idx
	}
	return pickLowestNonPoint(player, validIndices)
}

// pickLowestNonPoint 点札でない最も低い値のカードを選ぶ。なければ最も低い点札
func pickLowestNonPoint(player *TwoTenJackPlayer, validIndices []int) int {
	bestIdx := validIndices[0]
	bestScore := -1
	for _, idx := range validIndices {
		c := player.GetCard(idx)
		pts := TwoTenJackCardPoints(c)
		// prefer pts=0, then lower value
		score := 1000 - pts*100 - ttjEffectiveValue(c)
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// pickHighestPoint 最も点数の高いカードを捨てる
func pickHighestPoint(player *TwoTenJackPlayer, validIndices []int) int {
	bestIdx := validIndices[0]
	bestScore := -1
	for _, idx := range validIndices {
		c := player.GetCard(idx)
		score := TwoTenJackCardPoints(c)*10 + ttjEffectiveValue(c)
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// pickLowestWinning トリックに勝てる最小コストのカードを選ぶ
func pickLowestWinning(t *TwoTenJack, player *TwoTenJackPlayer, validIndices []int) (int, bool) {
	leadSuit := t.currentTrick[0].Card.GetDesign()
	// 現在のトリックの最高値
	maxTrumpVal := -1
	maxLeadVal := -1
	hasTrumpInTrick := false
	for _, tc := range t.currentTrick {
		v := ttjEffectiveValue(tc.Card)
		if tc.Card.GetDesign() == t.trumpSuit {
			hasTrumpInTrick = true
			if v > maxTrumpVal {
				maxTrumpVal = v
			}
		} else if tc.Card.GetDesign() == leadSuit {
			if v > maxLeadVal {
				maxLeadVal = v
			}
		}
	}

	winners := []int{}
	for _, idx := range validIndices {
		c := player.GetCard(idx)
		v := ttjEffectiveValue(c)
		isTrump := c.GetDesign() == t.trumpSuit
		if hasTrumpInTrick {
			if isTrump && v > maxTrumpVal {
				winners = append(winners, idx)
			}
		} else {
			if isTrump {
				winners = append(winners, idx)
			} else if c.GetDesign() == leadSuit && v > maxLeadVal {
				winners = append(winners, idx)
			}
		}
	}
	if len(winners) == 0 {
		return 0, false
	}
	bestIdx := winners[0]
	bestVal := ttjEffectiveValue(player.GetCard(bestIdx))
	for _, idx := range winners[1:] {
		v := ttjEffectiveValue(player.GetCard(idx))
		if v < bestVal {
			bestVal = v
			bestIdx = idx
		}
	}
	return bestIdx, true
}

// currentTrickWinnerIdx 現在のトリックの暫定勝者プレイヤーインデックス
func (t *TwoTenJack) currentTrickWinnerIdx() int {
	if len(t.currentTrick) == 0 {
		return -1
	}
	leadSuit := t.currentTrick[0].Card.GetDesign()
	winnerIdx := t.currentTrick[0].PlayerIdx
	winnerValue := ttjEffectiveValue(t.currentTrick[0].Card)
	winnerIsTrump := t.currentTrick[0].Card.GetDesign() == t.trumpSuit
	for _, tc := range t.currentTrick[1:] {
		isTrump := tc.Card.GetDesign() == t.trumpSuit
		v := ttjEffectiveValue(tc.Card)
		if isTrump && !winnerIsTrump {
			winnerIdx = tc.PlayerIdx
			winnerValue = v
			winnerIsTrump = true
		} else if isTrump && winnerIsTrump && v > winnerValue {
			winnerIdx = tc.PlayerIdx
			winnerValue = v
		} else if !isTrump && !winnerIsTrump && tc.Card.GetDesign() == leadSuit && v > winnerValue {
			winnerIdx = tc.PlayerIdx
			winnerValue = v
		}
	}
	return winnerIdx
}

// teamOf プレイヤーのチーム番号 (0 or 1)
func (t *TwoTenJack) teamOf(playerIdx int) int {
	if playerIdx < 0 {
		return -1
	}
	return playerIdx % 2
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (t *TwoTenJack) getValidPlayIndices(playerIdx int) []int {
	player := t.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return t.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (t *TwoTenJack) GetValidPlayIndices(playerIdx int) []int {
	return t.getValidPlayIndices(playerIdx)
}

// twoTenJackJSON is the JSON wire format for TwoTenJack.
type twoTenJackJSON struct {
	TrumpCards       *TrumpCards         `json:"tc"`
	Players          []*TwoTenJackPlayer `json:"ps"`
	Config           TwoTenJackConfig    `json:"cf"`
	Phase            TwoTenJackPhase     `json:"ph"`
	RoundNumber      int                 `json:"rn"`
	TrickNumber      int                 `json:"tn"`
	CurrentPlayerIdx int                 `json:"ci"`
	CurrentTrick     []*TrickCard        `json:"ct"`
	LeadPlayerIdx    int                 `json:"li"`
	DeclarerIdx      int                 `json:"di"`
	TrumpSuit        int                 `json:"ts"`
	GameEndFlag      bool                `json:"ge"`
	WinnerTeam       int                 `json:"wt"`
	ActionLog        []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (t *TwoTenJack) MarshalJSON() ([]byte, error) {
	return json.Marshal(twoTenJackJSON{
		TrumpCards:       t.trumpCards,
		Players:          t.players,
		Config:           t.config,
		Phase:            t.phase,
		RoundNumber:      t.roundNumber,
		TrickNumber:      t.trickNumber,
		CurrentPlayerIdx: t.currentPlayerIdx,
		CurrentTrick:     t.currentTrick,
		LeadPlayerIdx:    t.leadPlayerIdx,
		DeclarerIdx:      t.declarerIdx,
		TrumpSuit:        t.trumpSuit,
		GameEndFlag:      t.gameEndFlag,
		WinnerTeam:       t.winnerTeam,
		ActionLog:        t.actionLog,
	})
}

// twoTenJackMaxSliceLen caps slice sizes during deserialisation.
const twoTenJackMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (t *TwoTenJack) UnmarshalJSON(data []byte) error {
	var j twoTenJackJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != TwoTenJackPlayerCnt || len(j.CurrentTrick) > TwoTenJackPlayerCnt ||
		len(j.ActionLog) > twoTenJackMaxSliceLen {
		return fmt.Errorf("twotenjack: input array size is invalid")
	}
	t.trumpCards = j.TrumpCards
	if t.trumpCards == nil {
		t.trumpCards = NewTrumpCards(0)
	}
	t.players = j.Players
	if t.players == nil {
		t.players = make([]*TwoTenJackPlayer, 0)
	}
	t.config = j.Config
	t.phase = j.Phase
	t.roundNumber = j.RoundNumber
	t.trickNumber = j.TrickNumber
	t.currentPlayerIdx = j.CurrentPlayerIdx
	t.currentTrick = j.CurrentTrick
	if t.currentTrick == nil {
		t.currentTrick = make([]*TrickCard, 0)
	}
	t.leadPlayerIdx = j.LeadPlayerIdx
	t.declarerIdx = j.DeclarerIdx
	t.trumpSuit = j.TrumpSuit
	t.gameEndFlag = j.GameEndFlag
	t.winnerTeam = j.WinnerTeam
	t.actionLog = j.ActionLog
	if t.actionLog == nil {
		t.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
