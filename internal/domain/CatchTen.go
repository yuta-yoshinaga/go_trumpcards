package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// CatchTenHandSize 各プレイヤーの手札枚数 (36枚 / 4人)
const CatchTenHandSize = 9

// CatchTenPhase ゲームフェーズ
type CatchTenPhase int

// CatchTenのフェーズ定数
const (
	// CatchTenPhasePlay トリックプレイフェーズ
	CatchTenPhasePlay CatchTenPhase = 0
	// CatchTenPhaseTrickEnd トリック終了フェーズ
	CatchTenPhaseTrickEnd CatchTenPhase = 1
	// CatchTenPhaseRoundEnd ラウンド終了フェーズ
	CatchTenPhaseRoundEnd CatchTenPhase = 2
	// CatchTenPhaseGameEnd ゲーム終了フェーズ
	CatchTenPhaseGameEnd CatchTenPhase = 3
)

// CatchTenHint ヒント情報
type CatchTenHint struct {
	CardIndex *int   // 推奨カードインデックス
	Reason    string // ヒント理由キー
}

// CatchTenTrickCard トリック中の1枚
type CatchTenTrickCard struct {
	PlayerIdx int   `json:"pi"`
	Card      *Card `json:"c"`
}

// CatchTen Catch the Ten (Scotch Whist) ゲームクラス
type CatchTen struct {
	trumpCards       *TrumpCards
	players          []*CatchTenPlayer
	config           CatchTenConfig
	phase            CatchTenPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*CatchTenTrickCard
	trumpSuit        int // トランプ（切り札）スート
	leadPlayerIdx    int
	dealerIdx        int
	teamScores       [CatchTenTeamCnt]int
	gameEndFlag      bool
	winnerTeam       int // -1 = 未確定 / -2 = 引き分け
	actionLog        []*ActionLogEntry
}

// CatchTenDrawTeam は両チームが同時にポイント上限へ到達した際の引き分けを表す。
const CatchTenDrawTeam = -2

// NewCatchTen コンストラクタ
func NewCatchTen(trumpCards *TrumpCards, players []*CatchTenPlayer, config CatchTenConfig) *CatchTen {
	return &CatchTen{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerTeam:  -1,
		roundNumber: 0,
	}
}

// NewDefaultCatchTen returns CatchTen with the standard 4-player team setup
// (human team 0, alternating CPU teams) and DefaultCatchTenConfig.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultCatchTen() *CatchTen {
	players := []*CatchTenPlayer{
		NewCatchTenPlayer(true, 0),
		NewCatchTenPlayer(false, 1),
		NewCatchTenPlayer(false, 0),
		NewCatchTenPlayer(false, 1),
	}
	return NewCatchTen(NewTrumpCardsShortDeck(), players, DefaultCatchTenConfig())
}

// catchTenTrumpRank は切り札スートでの強さ順位を返す。
// J が最強で、以下 A,K,Q,10,9,8,7,6 の順。数値が大きいほど強い。
func catchTenTrumpRank(value int) int {
	switch value {
	case 11: // J
		return 100
	case 1: // A
		return 90
	case 13: // K
		return 80
	case 12: // Q
		return 70
	case 10:
		return 60
	default: // 9,8,7,6
		return value
	}
}

// catchTenPlainRank は非切り札スートでの強さ順位を返す (A高 → 6低)。
func catchTenPlainRank(value int) int {
	if value == 1 { // A は最強
		return 14
	}
	return value
}

// catchTenHonorPoints はカードがトリックを獲得したチームに与える名誉点を返す。
// トランプの J=11, 10=10, A=4, K=3, Q=2。非トランプは0点。
func catchTenHonorPoints(card *Card, trumpSuit int) int {
	if card == nil || card.GetDesign() != trumpSuit {
		return 0
	}
	switch card.GetValue() {
	case 11: // J
		return 11
	case 10:
		return 10
	case 1: // A
		return 4
	case 13: // K
		return 3
	case 12: // Q
		return 2
	default:
		return 0
	}
}

// Reset ゲーム初期化
func (g *CatchTen) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.currentPlayerIdx = -1
	g.dealerIdx = 0
	g.teamScores = [CatchTenTeamCnt]int{}
	g.actionLog = nil

	for _, p := range g.players {
		p.SetRoundScore(0)
		p.ResetTricks()
		p.Reset()
		p.SetIsFinished(false)
	}

	g.trumpCards.Shuffle()
	g.dealAndSetTrump()
	g.sortAllHands()

	g.startPlayPhase()
}

// NextRound 次のラウンドを開始する
func (g *CatchTen) NextRound() {
	if g.phase != CatchTenPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.currentPlayerIdx = -1
	g.dealerIdx = (g.dealerIdx + 1) % CatchTenPlayerCnt

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealAndSetTrump()
	g.sortAllHands()

	g.startPlayPhase()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *CatchTen) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CatchTenPhasePlay {
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

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (g *CatchTen) CpuPlay() {
	if g.gameEndFlag || g.phase != CatchTenPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := g.players[g.currentPlayerIdx]
	cardIdx := g.cpuSelectPlayCard(g.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	g.playCard(g.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (g *CatchTen) ResolveTrick() {
	if g.phase != CatchTenPhaseTrickEnd || len(g.currentTrick) != CatchTenPlayerCnt {
		return
	}

	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	honor := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		honor += catchTenHonorPoints(tc.Card, g.trumpSuit)
	}

	g.players[winnerIdx].AddTrick(trickCards)
	// 名誉点を獲得者のラウンドスコアとして加算 (ScoreRound でチーム集計)
	if honor > 0 {
		winner := g.players[winnerIdx]
		winner.SetRoundScore(winner.GetRoundScore() + honor)
	}

	winnerName := g.playerName(winnerIdx)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (honors: %d)", winnerName, g.trickNumber, honor), trickCards)

	g.leadPlayerIdx = winnerIdx

	if g.trickNumber >= CatchTenHandSize {
		g.phase = CatchTenPhaseRoundEnd
	} else {
		g.phase = CatchTenPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (g *CatchTen) NextTrick() {
	if g.phase != CatchTenPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = CatchTenPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (g *CatchTen) ScoreRound() {
	if g.phase != CatchTenPhaseRoundEnd {
		return
	}

	// チームごとの名誉点 (各プレイヤーのラウンドスコアに既に積算済み) を集計
	teamHonor := [CatchTenTeamCnt]int{}
	for _, p := range g.players {
		teamHonor[p.GetTeam()] += p.GetRoundScore()
	}

	for ti := 0; ti < CatchTenTeamCnt; ti++ {
		g.teamScores[ti] += teamHonor[ti]
		g.appendLog(-1, "round_score",
			fmt.Sprintf("Team %d: %d honors (total: %d)", ti, teamHonor[ti], g.teamScores[ti]), nil)
	}

	g.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *CatchTen) GetPhase() CatchTenPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *CatchTen) SetPhase(phase CatchTenPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *CatchTen) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *CatchTen) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (g *CatchTen) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *CatchTen) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *CatchTen) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *CatchTen) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *CatchTen) GetCurrentTrick() []*CatchTenTrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *CatchTen) SetCurrentTrick(trick []*CatchTenTrickCard) { g.currentTrick = trick }

// GetTrumpSuit トランプスート取得
func (g *CatchTen) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit トランプスート設定 (テスト用)
func (g *CatchTen) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *CatchTen) GetGameEndFlag() bool { return g.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (g *CatchTen) SetGameEndFlag(flag bool) { g.gameEndFlag = flag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定, -2 = 引き分け)
func (g *CatchTen) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *CatchTen) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *CatchTen) GetPlayer(i int) *CatchTenPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *CatchTen) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *CatchTen) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *CatchTen) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *CatchTen) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetTeamScore チームスコア取得
func (g *CatchTen) GetTeamScore(team int) int {
	if team < 0 || team >= CatchTenTeamCnt {
		return 0
	}
	return g.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (g *CatchTen) SetTeamScore(team, score int) {
	if team >= 0 && team < CatchTenTeamCnt {
		g.teamScores[team] = score
	}
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *CatchTen) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *CatchTen) GetConfig() CatchTenConfig { return g.config }

// SetConfig 設定変更
func (g *CatchTen) SetConfig(cfg CatchTenConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *CatchTen) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (g *CatchTen) GetValidPlayIndices(playerIdx int) []int {
	return g.getValidPlayIndices(playerIdx)
}

// GetHint ヒントを取得する
func (g *CatchTen) GetHint() *CatchTenHint {
	if g.phase == CatchTenPhasePlay && g.currentPlayerIdx == 0 {
		validIndices := g.getValidPlayIndices(0)
		if len(validIndices) == 0 {
			return nil
		}
		idx := g.cpuPlayHard(0, validIndices)
		return &CatchTenHint{CardIndex: &idx, Reason: g.playHintReason(idx)}
	}
	return nil
}

// --- Private methods ---

// dealAndSetTrump カードを配布し、ディーラーの最後のカードのスートをトランプに設定する
func (g *CatchTen) dealAndSetTrump() {
	dealAllCards(g.trumpCards, g.players)

	// ディーラーの最後のカードのスートがトランプ
	dealer := g.players[g.dealerIdx]
	if dealer.GetCardsSize() > 0 {
		lastCard := dealer.GetCard(dealer.GetCardsSize() - 1)
		g.trumpSuit = lastCard.GetDesign()
	} else {
		g.trumpSuit = CardDesignSpade
	}

	g.appendLog(-1, "trump", fmt.Sprintf("Trump suit: %s", suitName(g.trumpSuit)), nil)
}

// startPlayPhase プレイフェーズ開始: ディーラーの左隣がリード
func (g *CatchTen) startPlayPhase() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % CatchTenPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = CatchTenPhasePlay
}

// playCard カードをプレイする共通処理
func (g *CatchTen) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &CatchTenTrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == CatchTenPlayerCnt {
		g.phase = CatchTenPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % CatchTenPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
func (g *CatchTen) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		// リード: 任意のカードが可能
		return nil
	}

	// フォロースート
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit {
		if g.playerHasSuit(playerIdx, leadSuit) {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
	}

	return nil
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (g *CatchTen) playerHasSuit(playerIdx int, design int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// cardStrength はカードのトリック内での強さ (大きいほど強い) を返す。
// 切り札はトランプ専用順位 + 1000 のオフセットで常にリードスートに勝つ。
func (g *CatchTen) cardStrength(card *Card, leadSuit int) int {
	if card.GetDesign() == g.trumpSuit {
		return 1000 + catchTenTrumpRank(card.GetValue())
	}
	if card.GetDesign() == leadSuit {
		return catchTenPlainRank(card.GetValue())
	}
	return -1 // リードにもトランプにも従わないカードは勝てない
}

// trickWinner トリックの勝者を決定する
func (g *CatchTen) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerStrength := g.cardStrength(g.currentTrick[0].Card, leadSuit)

	for _, tc := range g.currentTrick[1:] {
		s := g.cardStrength(tc.Card, leadSuit)
		if s > winnerStrength {
			winnerStrength = s
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// checkGameEnd ゲーム終了判定
func (g *CatchTen) checkGameEnd() {
	t0Reached := g.teamScores[0] >= g.config.PointLimit
	t1Reached := g.teamScores[1] >= g.config.PointLimit

	if !t0Reached && !t1Reached {
		return
	}

	g.gameEndFlag = true
	g.phase = CatchTenPhaseGameEnd

	switch {
	case t0Reached && t1Reached:
		// 同一ディールで両チームが上限到達。高スコアのチームが勝者。
		// 完全同点ならば引き分け。
		if g.teamScores[0] > g.teamScores[1] {
			g.winnerTeam = 0
		} else if g.teamScores[1] > g.teamScores[0] {
			g.winnerTeam = 1
		} else {
			g.winnerTeam = CatchTenDrawTeam
		}
	case t0Reached:
		g.winnerTeam = 0
	default:
		g.winnerTeam = 1
	}

	if g.winnerTeam == CatchTenDrawTeam {
		g.appendLog(-1, "game_end", "Game ends in a draw!", nil)
	} else {
		g.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game!", g.winnerTeam), nil)
	}
}

// sortAllHands 全プレイヤーの手札をソートする
func (g *CatchTen) sortAllHands() {
	for _, p := range g.players {
		catchTenSortHand(p)
	}
}

// catchTenSortHand プレイヤーの手札をスート→値の順にソートする
func catchTenSortHand(p *CatchTenPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return cards[i].GetValue() < cards[j].GetValue()
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す
func (g *CatchTen) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (g *CatchTen) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// playHintReason プレイヒントの理由を判定する
func (g *CatchTen) playHintReason(chosenIdx int) string {
	player := g.players[0]
	card := player.GetCard(chosenIdx)

	if len(g.currentTrick) == 0 {
		return "lead_strong"
	}

	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == g.trumpSuit {
		return "trump_cut"
	}
	return "discard_high"
}

// --- CPU AI ---

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (g *CatchTen) cpuSelectPlayCard(playerIdx int) int {
	validIndices := g.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch g.config.CpuDifficulty {
	case CatchTenCpuDifficultyHard:
		return g.cpuPlayHard(playerIdx, validIndices)
	case CatchTenCpuDifficultyNormal:
		return g.cpuPlayNormal(playerIdx, validIndices)
	default:
		return g.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (g *CatchTen) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal 基本戦略プレイ
func (g *CatchTen) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]

	if len(g.currentTrick) == 0 {
		// リード: 名誉点の高い切り札があれば温存し、強い平札を出す
		return g.highestRankIdx(playerIdx, validIndices)
	}

	leadSuit := g.currentTrick[0].Card.GetDesign()
	winningStrength := g.currentWinningStrength()

	// 勝てるカードのうち最小のものを探す
	bestWin := -1
	bestWinStrength := 1 << 30
	for _, idx := range validIndices {
		s := g.cardStrength(player.GetCard(idx), leadSuit)
		if s > winningStrength && s < bestWinStrength {
			bestWin = idx
			bestWinStrength = s
		}
	}
	if bestWin >= 0 {
		return bestWin
	}

	// 勝てない場合は最も弱い (名誉点の低い) カードを捨てる
	return g.lowestValueIdx(playerIdx, validIndices)
}

// cpuPlayHard 高度な戦略プレイ
func (g *CatchTen) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]

	if len(g.currentTrick) == 0 {
		// リード: 切り札のJ/10は温存し、強い平札のAでリード
		bestIdx := validIndices[0]
		bestScore := -1
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			score := catchTenPlainRank(card.GetValue())
			if card.GetDesign() == g.trumpSuit {
				// Save trumps: penalise enough that any plain card outranks
				// every trump, so trumps lead only when nothing else remains.
				score = catchTenTrumpRank(card.GetValue()) - 100
			}
			if score > bestScore {
				bestScore = score
				bestIdx = idx
			}
		}
		return bestIdx
	}

	leadSuit := g.currentTrick[0].Card.GetDesign()
	winningStrength := g.currentWinningStrength()
	partnerWinning := g.isPartnerWinning(playerIdx)

	// パートナーが勝っている場合: 名誉点を載せられるなら載せ、なければ最弱札
	if partnerWinning {
		// 名誉点のあるトランプを今のトリックに載せられるなら載せる (味方の獲得点アップ)
		honorIdx := g.honorDumpIdx(playerIdx, validIndices, leadSuit)
		if honorIdx >= 0 {
			return honorIdx
		}
		return g.lowestValueIdx(playerIdx, validIndices)
	}

	// 勝ちに行く: 勝てるカードのうち最小のもの
	bestWin := -1
	bestWinStrength := 1 << 30
	for _, idx := range validIndices {
		s := g.cardStrength(player.GetCard(idx), leadSuit)
		if s > winningStrength && s < bestWinStrength {
			bestWin = idx
			bestWinStrength = s
		}
	}
	if bestWin >= 0 {
		return bestWin
	}

	// 勝てない場合は最も弱い (名誉点の低い) カードを捨てる
	return g.lowestValueIdx(playerIdx, validIndices)
}

// currentWinningStrength は現在のトリックで最も強いカードの強さを返す。
func (g *CatchTen) currentWinningStrength() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	best := g.cardStrength(g.currentTrick[0].Card, leadSuit)
	for _, tc := range g.currentTrick[1:] {
		s := g.cardStrength(tc.Card, leadSuit)
		if s > best {
			best = s
		}
	}
	return best
}

// honorDumpIdx はパートナーが勝っている時に名誉点を載せられる手札のインデックスを返す (-1=なし)。
// 名誉点の最も高い札を優先する。
func (g *CatchTen) honorDumpIdx(playerIdx int, validIndices []int, leadSuit int) int {
	player := g.players[playerIdx]
	bestIdx := -1
	bestHonor := 0
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		// 自分のカードがリードスートでなくフォロー義務がなければ名誉点トランプは出せない場合がある。
		// validatePlay を通過したカードのみ来るので、ここでは名誉点判定のみ。
		_ = leadSuit
		h := catchTenHonorPoints(card, g.trumpSuit)
		if h > bestHonor {
			bestHonor = h
			bestIdx = idx
		}
	}
	return bestIdx
}

// highestRankIdx は validIndices の中で最も強い (平札ランクの高い) 札を返す。
func (g *CatchTen) highestRankIdx(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]
	bestIdx := validIndices[0]
	bestVal := -1
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		v := catchTenPlainRank(card.GetValue())
		if card.GetDesign() == g.trumpSuit {
			// Penalise trumps so a strong plain-suit card leads and trump
			// honours are saved; a trump only leads when nothing else remains.
			v = catchTenTrumpRank(card.GetValue()) - 100
		}
		if v > bestVal {
			bestVal = v
			bestIdx = idx
		}
	}
	return bestIdx
}

// lowestValueIdx は validIndices の中で最も名誉点が低く、かつ弱い札を返す。
func (g *CatchTen) lowestValueIdx(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]
	bestIdx := validIndices[0]
	bestScore := 1 << 30
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		// 名誉点を強く回避し、次に弱い札を優先
		score := catchTenHonorPoints(card, g.trumpSuit)*100 + card.GetValue()
		if card.GetDesign() == g.trumpSuit {
			score += 50 // 切り札は温存したい
		}
		if score < bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// isPartnerWinning パートナーが現在のトリックを勝っているかチェック
func (g *CatchTen) isPartnerWinning(playerIdx int) bool {
	if len(g.currentTrick) == 0 {
		return false
	}
	myTeam := g.players[playerIdx].GetTeam()
	winnerIdx := g.trickWinner()
	return g.players[winnerIdx].GetTeam() == myTeam
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *CatchTen) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// catchTenJSON is the JSON wire format for CatchTen.
type catchTenJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*CatchTenPlayer    `json:"ps"`
	Config           CatchTenConfig       `json:"cf"`
	Phase            CatchTenPhase        `json:"ph"`
	RoundNumber      int                  `json:"rn"`
	TrickNumber      int                  `json:"tn"`
	CurrentPlayerIdx int                  `json:"ci"`
	CurrentTrick     []*CatchTenTrickCard `json:"ct"`
	TrumpSuit        int                  `json:"ts"`
	LeadPlayerIdx    int                  `json:"li"`
	DealerIdx        int                  `json:"di"`
	TeamScores       [CatchTenTeamCnt]int `json:"sc"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerTeam       int                  `json:"wt"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *CatchTen) MarshalJSON() ([]byte, error) {
	return json.Marshal(catchTenJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		TrumpSuit:        g.trumpSuit,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		TeamScores:       g.teamScores,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// catchTenMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const catchTenMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler. It validates structural
// invariants (slice bounds, enum ranges, player/seat indices) so that a
// tampered or malformed KV payload cannot drive the game into an invalid
// state on restore.
func (g *CatchTen) UnmarshalJSON(data []byte) error {
	var j catchTenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > catchTenMaxSliceLen || len(j.CurrentTrick) > catchTenMaxSliceLen ||
		len(j.ActionLog) > catchTenMaxSliceLen {
		return fmt.Errorf("catchten: input array exceeds maximum allowed size")
	}

	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("catchten: invalid config: %w", err)
	}
	if j.Phase < CatchTenPhasePlay || j.Phase > CatchTenPhaseGameEnd {
		return fmt.Errorf("catchten: phase out of range: %d", j.Phase)
	}
	// A fully dealt game always has exactly CatchTenPlayerCnt players; an empty
	// (zero-value) state has none. Anything else is malformed.
	playerCnt := len(j.Players)
	if playerCnt != 0 && playerCnt != CatchTenPlayerCnt {
		return fmt.Errorf("catchten: player count must be %d, got %d", CatchTenPlayerCnt, playerCnt)
	}
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("catchten: player at index %d cannot be nil", i)
		}
	}
	// TrumpSuit 0 is the valid "unset" state (before the trump is turned up).
	if j.TrumpSuit != 0 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("catchten: trump suit out of range: %d", j.TrumpSuit)
	}

	if err := catchTenValidateSeat("currentPlayerIdx", j.CurrentPlayerIdx, playerCnt); err != nil {
		return err
	}
	if err := catchTenValidateSeat("leadPlayerIdx", j.LeadPlayerIdx, playerCnt); err != nil {
		return err
	}
	if err := catchTenValidateSeat("dealerIdx", j.DealerIdx, playerCnt); err != nil {
		return err
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return fmt.Errorf("catchten: trick card cannot be nil")
		}
		if tc.PlayerIdx < 0 || tc.PlayerIdx >= playerCnt {
			return fmt.Errorf("catchten: trick player index out of range: %d", tc.PlayerIdx)
		}
	}
	// winnerTeam は -1(未確定), -2(引き分け), 0..teamCnt-1 を許容する。
	if j.WinnerTeam < CatchTenDrawTeam || j.WinnerTeam >= CatchTenTeamCnt {
		return fmt.Errorf("catchten: winner team out of range: %d", j.WinnerTeam)
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsShortDeck()
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*CatchTenPlayer, 0)
	}
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*CatchTenTrickCard, 0)
	}
	g.trumpSuit = j.TrumpSuit
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.teamScores = j.TeamScores
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// catchTenValidateSeat は座席インデックスが [-1, playerCnt-1] の範囲かを検証する。
// playerCnt が 0 の場合 (空状態) は -1 のみ許容する。
func catchTenValidateSeat(name string, idx, playerCnt int) error {
	if idx < -1 || idx >= playerCnt {
		// 空のプレイヤー集合の場合は -1 のみ許可
		if playerCnt != 0 || idx != -1 {
			return fmt.Errorf("catchten: %s out of range: %d", name, idx)
		}
	}
	return nil
}
