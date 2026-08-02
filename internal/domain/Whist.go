package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// WhistHandSize 各プレイヤーの手札枚数
const WhistHandSize = 13

// WhistBookThreshold トリック閾値 (この数を超えた分がチームポイント)
const WhistBookThreshold = 6

// WhistPhase ゲームフェーズ
type WhistPhase int

// Whistのフェーズ定数
const (
	// WhistPhasePlay トリックプレイフェーズ
	WhistPhasePlay WhistPhase = 0
	// WhistPhaseTrickEnd トリック終了フェーズ
	WhistPhaseTrickEnd WhistPhase = 1
	// WhistPhaseRoundEnd ラウンド終了フェーズ
	WhistPhaseRoundEnd WhistPhase = 2
	// WhistPhaseGameEnd ゲーム終了フェーズ
	WhistPhaseGameEnd WhistPhase = 3
)

// WhistHint ヒント情報
type WhistHint struct {
	CardIndex *int   // 推奨カードインデックス
	Reason    string // ヒント理由キー
}

// Whist ホイストゲームクラス
type Whist struct {
	trumpCards       *TrumpCards
	players          []*WhistPlayer
	config           WhistConfig
	phase            WhistPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	trumpSuit        int // トランプ（切り札）スート
	leadPlayerIdx    int
	dealerIdx        int
	teamScores       [WhistTeamCnt]int
	gameEndFlag      bool
	winnerTeam       int
	actionLog        []*ActionLogEntry
}

// NewWhist コンストラクタ
func NewWhist(trumpCards *TrumpCards, players []*WhistPlayer, config WhistConfig) *Whist {
	return &Whist{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerTeam:  -1,
		roundNumber: 0,
	}
}

// NewDefaultWhist returns Whist with the standard 4-player team setup
// (human team 0, alternating CPU teams) and DefaultWhistConfig.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultWhist() *Whist {
	players := []*WhistPlayer{
		NewWhistPlayer(true, 0),
		NewWhistPlayer(false, 1),
		NewWhistPlayer(false, 0),
		NewWhistPlayer(false, 1),
	}
	return NewWhist(NewTrumpCards(0), players, DefaultWhistConfig())
}

// Reset ゲーム初期化
func (w *Whist) Reset() {
	w.gameEndFlag = false
	w.winnerTeam = -1
	w.roundNumber = 1
	w.trickNumber = 0
	w.currentTrick = nil
	w.leadPlayerIdx = -1
	w.currentPlayerIdx = -1
	w.dealerIdx = 0
	w.teamScores = [WhistTeamCnt]int{}
	w.actionLog = nil

	for _, p := range w.players {
		p.roundScore = 0
		p.cumulativeScore = 0
		p.tricksTaken = nil
		p.Reset()
		p.SetIsFinished(false)
	}

	w.trumpCards.Shuffle()
	w.dealAndSetTrump()
	w.sortAllHands()

	w.startPlayPhase()
}

// NextRound 次のラウンドを開始する
func (w *Whist) NextRound() {
	if w.phase != WhistPhaseRoundEnd {
		return
	}

	w.roundNumber++
	w.trickNumber = 0
	w.currentTrick = nil
	w.leadPlayerIdx = -1
	w.currentPlayerIdx = -1
	w.dealerIdx = (w.dealerIdx + 1) % WhistPlayerCnt

	for _, p := range w.players {
		p.ResetRound()
	}

	w.trumpCards.Shuffle()
	w.dealAndSetTrump()
	w.sortAllHands()

	w.startPlayPhase()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (w *Whist) PlayerPlay(cardIndex int) error {
	if w.gameEndFlag {
		return ErrGameEnded
	}
	if w.phase != WhistPhasePlay {
		return ErrWrongPhase
	}
	if !w.players[w.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := w.players[w.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := w.validatePlay(w.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	w.playCard(w.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (w *Whist) CpuPlay() {
	if w.gameEndFlag || w.phase != WhistPhasePlay {
		return
	}
	if w.players[w.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := w.players[w.currentPlayerIdx]
	cardIdx := w.cpuSelectPlayCard(w.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	w.playCard(w.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (w *Whist) ResolveTrick() {
	if w.phase != WhistPhaseTrickEnd || len(w.currentTrick) != WhistPlayerCnt {
		return
	}

	winnerIdx := w.trickWinner()
	trickCards := make([]*Card, len(w.currentTrick))
	for i, tc := range w.currentTrick {
		trickCards[i] = tc.Card
	}

	w.players[winnerIdx].AddTrick(trickCards)

	winnerName := w.playerName(winnerIdx)
	w.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d", winnerName, w.trickNumber), trickCards)

	w.leadPlayerIdx = winnerIdx

	if w.trickNumber >= WhistHandSize {
		w.phase = WhistPhaseRoundEnd
	} else {
		w.phase = WhistPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (w *Whist) NextTrick() {
	if w.phase != WhistPhaseTrickEnd {
		return
	}
	w.currentTrick = nil
	w.currentPlayerIdx = w.leadPlayerIdx
	w.trickNumber++
	w.phase = WhistPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (w *Whist) ScoreRound() {
	if w.phase != WhistPhaseRoundEnd {
		return
	}

	// チームごとのトリック数を集計
	teamTricks := [WhistTeamCnt]int{}
	for _, p := range w.players {
		teamTricks[p.GetTeam()] += p.GetTrickCount()
	}

	// 6トリック超過分がチームポイント
	for ti := 0; ti < WhistTeamCnt; ti++ {
		points := 0
		if teamTricks[ti] > WhistBookThreshold {
			points = teamTricks[ti] - WhistBookThreshold
		}
		w.teamScores[ti] += points
		w.appendLog(-1, "round_score",
			fmt.Sprintf("Team %d: %d points (tricks: %d, total: %d)", ti, points, teamTricks[ti], w.teamScores[ti]), nil)
	}

	w.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (w *Whist) GetPhase() WhistPhase { return w.phase }

// SetPhase フェーズ設定 (テスト用)
func (w *Whist) SetPhase(phase WhistPhase) { w.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (w *Whist) GetRoundNumber() int { return w.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (w *Whist) SetRoundNumber(n int) { w.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (w *Whist) GetTrickNumber() int { return w.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (w *Whist) SetTrickNumber(n int) { w.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (w *Whist) GetCurrentPlayerIdx() int { return w.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (w *Whist) SetCurrentPlayerIdx(idx int) { w.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (w *Whist) GetCurrentTrick() []*TrickCard { return w.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (w *Whist) SetCurrentTrick(trick []*TrickCard) { w.currentTrick = trick }

// GetTrumpSuit トランプスート取得
func (w *Whist) GetTrumpSuit() int { return w.trumpSuit }

// SetTrumpSuit トランプスート設定 (テスト用)
func (w *Whist) SetTrumpSuit(suit int) { w.trumpSuit = suit }

// GetGameEndFlag ゲーム終了フラグ取得
func (w *Whist) GetGameEndFlag() bool { return w.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (w *Whist) SetGameEndFlag(flag bool) { w.gameEndFlag = flag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (w *Whist) GetWinnerTeam() int { return w.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (w *Whist) GetPlayerCnt() int { return len(w.players) }

// GetPlayer プレイヤー取得
func (w *Whist) GetPlayer(i int) *WhistPlayer {
	if i < 0 || i >= len(w.players) {
		return nil
	}
	return w.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (w *Whist) GetLeadPlayerIdx() int { return w.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (w *Whist) SetLeadPlayerIdx(idx int) { w.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (w *Whist) GetDealerIdx() int { return w.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (w *Whist) SetDealerIdx(idx int) { w.dealerIdx = idx }

// GetTeamScore チームスコア取得
func (w *Whist) GetTeamScore(team int) int {
	if team < 0 || team >= WhistTeamCnt {
		return 0
	}
	return w.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (w *Whist) SetTeamScore(team, score int) {
	if team >= 0 && team < WhistTeamCnt {
		w.teamScores[team] = score
	}
}

// IsHumanTurn 現在の手番が人間かどうか
func (w *Whist) IsHumanTurn() bool {
	if w.currentPlayerIdx < 0 || w.currentPlayerIdx >= len(w.players) {
		return false
	}
	return w.players[w.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (w *Whist) GetConfig() WhistConfig { return w.config }

// SetConfig 設定変更
func (w *Whist) SetConfig(cfg WhistConfig) { w.config = cfg }

// GetActionLog 棋譜取得
func (w *Whist) GetActionLog() []*ActionLogEntry { return w.actionLog }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (w *Whist) GetValidPlayIndices(playerIdx int) []int {
	return w.getValidPlayIndices(playerIdx)
}

// GetHint ヒントを取得する
func (w *Whist) GetHint() *WhistHint {
	if w.phase == WhistPhasePlay && w.currentPlayerIdx == 0 {
		validIndices := w.getValidPlayIndices(0)
		if len(validIndices) == 0 {
			return nil
		}
		idx := w.cpuPlayHard(0, validIndices)
		return &WhistHint{CardIndex: &idx, Reason: w.playHintReason(idx)}
	}
	return nil
}

// --- Private methods ---

// findHumanIdx 人間プレイヤーのインデックスを返す (-1=なし)
func (w *Whist) findHumanIdx() int {
	for i, p := range w.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// dealAndSetTrump カードを配布し、ディーラーの最後のカードのスートをトランプに設定する
func (w *Whist) dealAndSetTrump() {
	dealAllCards(w.trumpCards, w.players)

	// ディーラーの最後のカードのスートがトランプ
	dealer := w.players[w.dealerIdx]
	if dealer.GetCardsSize() > 0 {
		lastCard := dealer.GetCard(dealer.GetCardsSize() - 1)
		w.trumpSuit = lastCard.GetDesign()
	} else {
		w.trumpSuit = CardDesignSpade
	}

	w.appendLog(-1, "trump", fmt.Sprintf("Trump suit: %s", suitName(w.trumpSuit)), nil)
}

// startPlayPhase プレイフェーズ開始: ディーラーの左隣がリード
func (w *Whist) startPlayPhase() {
	w.trickNumber = 1
	w.currentTrick = nil
	w.leadPlayerIdx = (w.dealerIdx + 1) % WhistPlayerCnt
	w.currentPlayerIdx = w.leadPlayerIdx
	w.phase = WhistPhasePlay
}

// playCard カードをプレイする共通処理
func (w *Whist) playCard(playerIdx int, card *Card) {
	w.currentTrick = append(w.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	w.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", w.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(w.currentTrick) == WhistPlayerCnt {
		w.phase = WhistPhaseTrickEnd
	} else {
		w.currentPlayerIdx = (w.currentPlayerIdx + 1) % WhistPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
func (w *Whist) validatePlay(playerIdx int, card *Card) error {
	if len(w.currentTrick) == 0 {
		// リード: 任意のカードが可能
		return nil
	}

	// フォロースート
	leadSuit := w.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit {
		if w.playerHasSuit(playerIdx, leadSuit) {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
	}

	return nil
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (w *Whist) playerHasSuit(playerIdx int, design int) bool {
	p := w.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する
func (w *Whist) trickWinner() int {
	return ResolveTrickWinner(w.currentTrick, w.trumpSuit, nil)
}

// checkGameEnd ゲーム終了判定
func (w *Whist) checkGameEnd() {
	hasWinner := false
	for ti := 0; ti < WhistTeamCnt; ti++ {
		if w.teamScores[ti] >= w.config.PointLimit {
			hasWinner = true
			break
		}
	}

	if !hasWinner {
		return
	}

	w.gameEndFlag = true
	w.phase = WhistPhaseGameEnd

	// 最高スコアのチームが勝者 (同点ならチーム0)
	if w.teamScores[0] >= w.teamScores[1] {
		w.winnerTeam = 0
	} else {
		w.winnerTeam = 1
	}
	w.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game!", w.winnerTeam), nil)
}

// sortAllHands 全プレイヤーの手札をソートする
func (w *Whist) sortAllHands() {
	for _, p := range w.players {
		whistSortHand(p)
	}
}

// whistSortHand プレイヤーの手札をスート→値の順にソートする
func whistSortHand(p *WhistPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// playerName プレイヤー名を返す
func (w *Whist) playerName(idx int) string {
	if idx < 0 || idx >= len(w.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if w.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (w *Whist) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	w.actionLog = append(w.actionLog, &ActionLogEntry{
		TurnNumber: len(w.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// playHintReason プレイヒントの理由を判定する
func (w *Whist) playHintReason(chosenIdx int) string {
	player := w.players[0]
	card := player.GetCard(chosenIdx)

	if len(w.currentTrick) == 0 {
		return "lead_strong"
	}

	leadSuit := w.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == w.trumpSuit {
		return "trump_cut"
	}
	return "discard_high"
}

// --- CPU AI ---

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (w *Whist) cpuSelectPlayCard(playerIdx int) int {
	validIndices := w.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch w.config.CpuDifficulty {
	case WhistCpuDifficultyHard:
		return w.cpuPlayHard(playerIdx, validIndices)
	case WhistCpuDifficultyNormal:
		return w.cpuPlayNormal(playerIdx, validIndices)
	default:
		return w.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (w *Whist) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal 基本戦略プレイ
func (w *Whist) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := w.players[playerIdx]

	if len(w.currentTrick) == 0 {
		// リード: 高いカードを出す
		bestIdx := validIndices[0]
		bestVal := player.GetCard(validIndices[0]).GetValue()
		for _, idx := range validIndices[1:] {
			card := player.GetCard(idx)
			if card.GetValue() > bestVal {
				bestVal = card.GetValue()
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// フォロー
	leadSuit := w.currentTrick[0].Card.GetDesign()
	hasLeadSuit := false
	for _, idx := range validIndices {
		if player.GetCard(idx).GetDesign() == leadSuit {
			hasLeadSuit = true
			break
		}
	}

	if hasLeadSuit {
		highestInTrick := 0
		for _, tc := range w.currentTrick {
			if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
				highestInTrick = tc.Card.GetValue()
			}
		}

		// 勝てるカードがあれば最小の勝てるカード
		overCards := []int{}
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			if card.GetDesign() == leadSuit && card.GetValue() > highestInTrick {
				overCards = append(overCards, idx)
			}
		}
		if len(overCards) > 0 {
			bestIdx := overCards[0]
			for _, idx := range overCards[1:] {
				if player.GetCard(idx).GetValue() < player.GetCard(bestIdx).GetValue() {
					bestIdx = idx
				}
			}
			return bestIdx
		}

		// 勝てない場合は最も低いカード
		bestIdx := validIndices[0]
		bestVal := player.GetCard(validIndices[0]).GetValue()
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			if card.GetDesign() == leadSuit && card.GetValue() < bestVal {
				bestVal = card.GetValue()
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// ボイド: トランプでカット
	for _, idx := range validIndices {
		if player.GetCard(idx).GetDesign() == w.trumpSuit {
			return idx
		}
	}
	// 低いカードを捨てる
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

// cpuPlayHard 高度な戦略プレイ
func (w *Whist) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := w.players[playerIdx]

	if len(w.currentTrick) == 0 {
		// リード: A やトランプの高札でリード
		bestIdx := validIndices[0]
		bestScore := -1
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			score := card.GetValue()
			if card.GetDesign() == w.trumpSuit {
				score += 100
			}
			if score > bestScore {
				bestScore = score
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// フォロー
	leadSuit := w.currentTrick[0].Card.GetDesign()

	// トリックにトランプが含まれるか
	highestTrumpInTrick := 0
	hasTrumpInTrick := false
	highestInTrick := 0
	for _, tc := range w.currentTrick {
		if tc.Card.GetDesign() == w.trumpSuit {
			hasTrumpInTrick = true
			if tc.Card.GetValue() > highestTrumpInTrick {
				highestTrumpInTrick = tc.Card.GetValue()
			}
		}
		if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
			highestInTrick = tc.Card.GetValue()
		}
	}

	// パートナーが現在勝っているかチェック
	partnerWinning := w.isPartnerWinning(playerIdx)

	// フォロー可能なら勝ち札/受け札を選ぶ (フォロー時 validIndices は全てリードスート)。
	leadSuitIndices := filterByDesign(player, validIndices, leadSuit)
	if len(leadSuitIndices) > 0 {
		if partnerWinning {
			// パートナーが勝っているなら最も低いリードスート札を出す。
			return pickLowest(player, leadSuitIndices, nil)
		}
		// 勝ちに行く (トランプが出ていなければリードスートで勝てる)。
		if !hasTrumpInTrick || leadSuit == w.trumpSuit {
			threshold := highestInTrick
			if leadSuit == w.trumpSuit {
				threshold = highestTrumpInTrick
			}
			if over := filterAbove(player, leadSuitIndices, threshold, nil); len(over) > 0 {
				return pickLowest(player, over, nil)
			}
		}
		// アンダーカードのうち最も高い札で場に付き合う。
		if under := filterBelow(player, leadSuitIndices, highestInTrick, nil); len(under) > 0 {
			return pickHighest(player, under, nil)
		}
		// 勝てないなら最も低いリードスート札を出す。
		return pickLowest(player, leadSuitIndices, nil)
	}

	// ボイド: パートナーが勝っているなら最も低い札を捨てる。
	if partnerWinning {
		return pickLowest(player, validIndices, nil)
	}

	// トランプでカット (出ていれば勝てる最小のトランプ、なければ最小のトランプ)。
	trumpIndices := filterByDesign(player, validIndices, w.trumpSuit)
	if len(trumpIndices) > 0 {
		if hasTrumpInTrick {
			if over := filterAbove(player, trumpIndices, highestTrumpInTrick, nil); len(over) > 0 {
				return pickLowest(player, over, nil)
			}
		} else {
			return pickLowest(player, trumpIndices, nil)
		}
	}

	// 最も高い不要カードを捨てる
	bestIdx := validIndices[0]
	bestScore := -1
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		score := card.GetValue()
		// トランプは温存したい
		if card.GetDesign() == w.trumpSuit {
			score -= 100
		}
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// isPartnerWinning パートナーが現在のトリックを勝っているかチェック
func (w *Whist) isPartnerWinning(playerIdx int) bool {
	if len(w.currentTrick) == 0 {
		return false
	}
	myTeam := w.players[playerIdx].GetTeam()
	winnerIdx := w.trickWinner()
	return w.players[winnerIdx].GetTeam() == myTeam
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (w *Whist) getValidPlayIndices(playerIdx int) []int {
	player := w.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return w.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// whistJSON is the JSON wire format for Whist.
type whistJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*WhistPlayer    `json:"ps"`
	Config           WhistConfig       `json:"cf"`
	Phase            WhistPhase        `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	TrumpSuit        int               `json:"ts"`
	LeadPlayerIdx    int               `json:"li"`
	DealerIdx        int               `json:"di"`
	TeamScores       [WhistTeamCnt]int `json:"sc"`
	GameEndFlag      bool              `json:"ge"`
	WinnerTeam       int               `json:"wt"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (w *Whist) MarshalJSON() ([]byte, error) {
	return json.Marshal(whistJSON{
		TrumpCards:       w.trumpCards,
		Players:          w.players,
		Config:           w.config,
		Phase:            w.phase,
		RoundNumber:      w.roundNumber,
		TrickNumber:      w.trickNumber,
		CurrentPlayerIdx: w.currentPlayerIdx,
		CurrentTrick:     w.currentTrick,
		TrumpSuit:        w.trumpSuit,
		LeadPlayerIdx:    w.leadPlayerIdx,
		DealerIdx:        w.dealerIdx,
		TeamScores:       w.teamScores,
		GameEndFlag:      w.gameEndFlag,
		WinnerTeam:       w.winnerTeam,
		ActionLog:        w.actionLog,
	})
}

// whistMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const whistMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (w *Whist) UnmarshalJSON(data []byte) error {
	var j whistJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > whistMaxSliceLen || len(j.CurrentTrick) > whistMaxSliceLen ||
		len(j.ActionLog) > whistMaxSliceLen {
		return fmt.Errorf("whist: input array exceeds maximum allowed size")
	}
	w.trumpCards = j.TrumpCards
	if w.trumpCards == nil {
		w.trumpCards = NewTrumpCards(0)
	}
	w.players = j.Players
	if w.players == nil {
		w.players = make([]*WhistPlayer, 0)
	}
	w.config = j.Config
	w.phase = j.Phase
	w.roundNumber = j.RoundNumber
	w.trickNumber = j.TrickNumber
	w.currentPlayerIdx = j.CurrentPlayerIdx
	w.currentTrick = j.CurrentTrick
	if w.currentTrick == nil {
		w.currentTrick = make([]*TrickCard, 0)
	}
	w.trumpSuit = j.TrumpSuit
	w.leadPlayerIdx = j.LeadPlayerIdx
	w.dealerIdx = j.DealerIdx
	w.teamScores = j.TeamScores
	w.gameEndFlag = j.GameEndFlag
	w.winnerTeam = j.WinnerTeam
	w.actionLog = j.ActionLog
	if w.actionLog == nil {
		w.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
