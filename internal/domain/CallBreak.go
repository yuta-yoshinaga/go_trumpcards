package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// CallBreakPlayerCnt Call Break のプレイヤー数
const CallBreakPlayerCnt = 4

// CallBreakHandSize 各プレイヤーの手札枚数
const CallBreakHandSize = 13

// CallBreakPhase ゲームフェーズ
type CallBreakPhase int

// CallBreak のフェーズ定数
const (
	// CallBreakPhaseBid ビッドフェーズ
	CallBreakPhaseBid CallBreakPhase = 0
	// CallBreakPhasePlay トリックプレイフェーズ
	CallBreakPhasePlay CallBreakPhase = 1
	// CallBreakPhaseTrickEnd トリック終了フェーズ
	CallBreakPhaseTrickEnd CallBreakPhase = 2
	// CallBreakPhaseRoundEnd ラウンド終了フェーズ
	CallBreakPhaseRoundEnd CallBreakPhase = 3
	// CallBreakPhaseGameEnd ゲーム終了フェーズ
	CallBreakPhaseGameEnd CallBreakPhase = 4
)

// CallBreakHint ヒント情報
type CallBreakHint struct {
	CardIndex *int   // 推奨カードインデックス (ビッド時 nil)
	Bid       *int   // 推奨ビッド値 (プレイ時 nil)
	Reason    string // ヒント理由キー
}

// CallBreak Call Break ゲームクラス
//
// ルール上の差分メモ (Spades との比較):
//   - スコアは ×10 された整数。bid 達成時は bid*10 + overtricks、未達時は -bid*10。
//   - Nil ビッドは存在しない (最小 1)。
//   - リード時にスペードを切れる条件は Spades と同じ (ブレイク済み or 手札がスペードのみ)。
//   - リードスートに従えない場合は「必ずトランプ (スペード) を切らなければならない」。
//     ボイドかつスペードを持たない場合のみ任意のカードを捨てられる。
//   - ゲームは MaxRounds に達した時点で終了し、累積スコアが最大のプレイヤーが勝者。
type CallBreak struct {
	trumpCards       *TrumpCards
	players          []*CallBreakPlayer
	config           CallBreakConfig
	phase            CallBreakPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	spadesBroken     bool
	leadPlayerIdx    int
	bidPlayerIdx     int
	gameEndFlag      bool
	winnerIdx        int
	actionLog        []*ActionLogEntry
}

// NewCallBreak コンストラクタ
func NewCallBreak(trumpCards *TrumpCards, players []*CallBreakPlayer, config CallBreakConfig) *CallBreak {
	return &CallBreak{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
	}
}

// NewDefaultCallBreak は 4 人 (1 人間 + 3 CPU) の標準セットアップを返す。
// CUI / Web / Worker 共通の構築 SSoT。
func NewDefaultCallBreak() *CallBreak {
	players := []*CallBreakPlayer{
		NewCallBreakPlayer(true),
		NewCallBreakPlayer(false),
		NewCallBreakPlayer(false),
		NewCallBreakPlayer(false),
	}
	return NewCallBreak(NewTrumpCards(0), players, DefaultCallBreakConfig())
}

// Reset ゲーム初期化
func (cb *CallBreak) Reset() {
	cb.gameEndFlag = false
	cb.winnerIdx = -1
	cb.roundNumber = 1
	cb.trickNumber = 0
	cb.spadesBroken = false
	cb.currentTrick = nil
	cb.leadPlayerIdx = 0
	cb.currentPlayerIdx = -1
	cb.bidPlayerIdx = 0
	cb.actionLog = nil

	for _, p := range cb.players {
		p.bid = -1
		p.SetRoundScore(0)
		p.ResetTricks()
		p.Reset()
		p.SetIsFinished(false)
	}
	// 累積スコアはラウンド跨ぎで保持する値だが、Reset はゲーム開始時のみ呼ばれるので 0 に戻す
	for _, p := range cb.players {
		p.SetCumulativeScore(0)
	}

	cb.trumpCards.Shuffle()
	dealAllCards(cb.trumpCards, cb.players)
	cb.sortAllHands()

	cb.phase = CallBreakPhaseBid
}

// NextRound 次のラウンドを開始する
func (cb *CallBreak) NextRound() {
	if cb.phase != CallBreakPhaseRoundEnd {
		return
	}

	cb.roundNumber++
	cb.trickNumber = 0
	cb.spadesBroken = false
	cb.currentTrick = nil
	cb.leadPlayerIdx = 0
	cb.currentPlayerIdx = -1
	cb.bidPlayerIdx = 0

	for _, p := range cb.players {
		p.ResetRound()
	}

	cb.trumpCards.Shuffle()
	dealAllCards(cb.trumpCards, cb.players)
	cb.sortAllHands()

	cb.phase = CallBreakPhaseBid
}

// PlayerBid 人間プレイヤーがビッドする
func (cb *CallBreak) PlayerBid(bid int) error {
	if cb.gameEndFlag {
		return ErrGameEnded
	}
	if cb.phase != CallBreakPhaseBid {
		return ErrWrongPhase
	}

	humanIdx := cb.findHumanIdx()
	if humanIdx < 0 {
		return ErrNotHumanTurn
	}
	if cb.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if bid < CallBreakMinBid || bid > CallBreakHandSize {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは %d〜%d で指定してください", CallBreakMinBid, CallBreakHandSize))
	}

	cb.players[humanIdx].SetBid(bid)
	cb.appendLog(humanIdx, "bid", fmt.Sprintf("%s bids %d", cb.playerName(humanIdx), bid), nil)

	cb.bidPlayerIdx++
	cb.checkBidComplete()
	return nil
}

// CpuBid 現在のビッドプレイヤーが CPU の場合にビッドする
func (cb *CallBreak) CpuBid() {
	if cb.gameEndFlag || cb.phase != CallBreakPhaseBid {
		return
	}
	if cb.bidPlayerIdx >= CallBreakPlayerCnt {
		return
	}
	if cb.players[cb.bidPlayerIdx].GetIsHuman() {
		return
	}

	bid := cb.cpuSelectBid(cb.bidPlayerIdx)
	cb.players[cb.bidPlayerIdx].SetBid(bid)
	cb.appendLog(cb.bidPlayerIdx, "bid", fmt.Sprintf("%s bids %d", cb.playerName(cb.bidPlayerIdx), bid), nil)

	cb.bidPlayerIdx++
	cb.checkBidComplete()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (cb *CallBreak) PlayerPlay(cardIndex int) error {
	if cb.gameEndFlag {
		return ErrGameEnded
	}
	if cb.phase != CallBreakPhasePlay {
		return ErrWrongPhase
	}
	if !cb.players[cb.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := cb.players[cb.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := cb.validatePlay(cb.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	cb.playCard(cb.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行
func (cb *CallBreak) CpuPlay() {
	if cb.gameEndFlag || cb.phase != CallBreakPhasePlay {
		return
	}
	if cb.players[cb.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := cb.players[cb.currentPlayerIdx]
	cardIdx := cb.cpuSelectPlayCard(cb.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	cb.playCard(cb.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (cb *CallBreak) ResolveTrick() {
	if cb.phase != CallBreakPhaseTrickEnd || len(cb.currentTrick) != CallBreakPlayerCnt {
		return
	}

	winnerIdx := cb.trickWinner()
	trickCards := make([]*Card, len(cb.currentTrick))
	for i, tc := range cb.currentTrick {
		trickCards[i] = tc.Card
	}

	cb.players[winnerIdx].AddTrick(trickCards)
	cb.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d", cb.playerName(winnerIdx), cb.trickNumber), trickCards)

	cb.leadPlayerIdx = winnerIdx

	if cb.trickNumber >= CallBreakHandSize {
		cb.phase = CallBreakPhaseRoundEnd
	} else {
		cb.phase = CallBreakPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (cb *CallBreak) NextTrick() {
	if cb.phase != CallBreakPhaseTrickEnd {
		return
	}
	cb.currentTrick = nil
	cb.currentPlayerIdx = cb.leadPlayerIdx
	cb.trickNumber++
	cb.phase = CallBreakPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
//
// スコアは内部的に「×10 された整数」で計算する。
//   - bid 達成: roundScore = bid*10 + overtricks
//   - bid 未達: roundScore = -bid*10
//
// 例: bid=4 で 5 トリック → 41 (表示 4.1) / bid=4 で 3 トリック → -40 (表示 -4.0)
func (cb *CallBreak) ScoreRound() {
	if cb.phase != CallBreakPhaseRoundEnd {
		return
	}

	for i := 0; i < CallBreakPlayerCnt; i++ {
		p := cb.players[i]
		tricks := p.GetTrickCount()
		bid := p.GetBid()

		var score int
		if tricks >= bid {
			score = bid*10 + (tricks - bid)
		} else {
			score = -bid * 10
		}
		p.SetRoundScore(score)

		cb.appendLog(i, "round_score", fmt.Sprintf("%s: bid=%d tricks=%d round=%s",
			cb.playerName(i), bid, tricks, FormatCallBreakScore(score)), nil)
	}

	// 累積スコアに加算
	for i := 0; i < CallBreakPlayerCnt; i++ {
		cb.players[i].CommitRoundScore()
	}

	// スコアログ
	for i := 0; i < CallBreakPlayerCnt; i++ {
		cb.appendLog(i, "cumulative_score", fmt.Sprintf("%s: total=%s",
			cb.playerName(i), FormatCallBreakScore(cb.players[i].GetCumulativeScore())), nil)
	}

	cb.checkGameEnd()
}

// FormatCallBreakScore は内部値 (×10) を "X.Y" / "-X.Y" 形式の文字列にする。
// 41 → "4.1", -40 → "-4.0", 0 → "0.0"。
func FormatCallBreakScore(internal int) string {
	if internal < 0 {
		n := -internal
		return fmt.Sprintf("-%d.%d", n/10, n%10)
	}
	return fmt.Sprintf("%d.%d", internal/10, internal%10)
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (cb *CallBreak) GetPhase() CallBreakPhase { return cb.phase }

// SetPhase フェーズ設定 (テスト用)
func (cb *CallBreak) SetPhase(phase CallBreakPhase) { cb.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (cb *CallBreak) GetRoundNumber() int { return cb.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (cb *CallBreak) SetRoundNumber(n int) { cb.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (cb *CallBreak) GetTrickNumber() int { return cb.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (cb *CallBreak) SetTrickNumber(n int) { cb.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (cb *CallBreak) GetCurrentPlayerIdx() int { return cb.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (cb *CallBreak) SetCurrentPlayerIdx(idx int) { cb.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (cb *CallBreak) GetCurrentTrick() []*TrickCard { return cb.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (cb *CallBreak) SetCurrentTrick(trick []*TrickCard) { cb.currentTrick = trick }

// GetSpadesBroken スペードブレイク状態取得
func (cb *CallBreak) GetSpadesBroken() bool { return cb.spadesBroken }

// SetSpadesBroken スペードブレイク状態設定 (テスト用)
func (cb *CallBreak) SetSpadesBroken(broken bool) { cb.spadesBroken = broken }

// GetGameEndFlag ゲーム終了フラグ取得
func (cb *CallBreak) GetGameEndFlag() bool { return cb.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (cb *CallBreak) GetWinnerIdx() int { return cb.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (cb *CallBreak) GetPlayerCnt() int { return len(cb.players) }

// GetPlayer プレイヤー取得
func (cb *CallBreak) GetPlayer(i int) *CallBreakPlayer {
	if i < 0 || i >= len(cb.players) {
		return nil
	}
	return cb.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (cb *CallBreak) GetLeadPlayerIdx() int { return cb.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (cb *CallBreak) SetLeadPlayerIdx(idx int) { cb.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (cb *CallBreak) GetBidPlayerIdx() int { return cb.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (cb *CallBreak) SetBidPlayerIdx(idx int) { cb.bidPlayerIdx = idx }

// IsHumanTurn 現在の手番が人間かどうか
func (cb *CallBreak) IsHumanTurn() bool {
	if cb.currentPlayerIdx < 0 || cb.currentPlayerIdx >= len(cb.players) {
		return false
	}
	return cb.players[cb.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (cb *CallBreak) IsHumanBidTurn() bool {
	if cb.bidPlayerIdx < 0 || cb.bidPlayerIdx >= len(cb.players) {
		return false
	}
	return cb.players[cb.bidPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (cb *CallBreak) GetConfig() CallBreakConfig { return cb.config }

// SetConfig 設定変更
func (cb *CallBreak) SetConfig(cfg CallBreakConfig) { cb.config = cfg }

// GetActionLog 棋譜取得
func (cb *CallBreak) GetActionLog() []*ActionLogEntry { return cb.actionLog }

// --- Private methods ---

// findHumanIdx 人間プレイヤーのインデックスを返す (-1 = なし)
func (cb *CallBreak) findHumanIdx() int {
	for i, p := range cb.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// checkBidComplete 全員がビッドしたかチェックし、プレイフェーズに移行
func (cb *CallBreak) checkBidComplete() {
	if cb.bidPlayerIdx >= CallBreakPlayerCnt {
		cb.phase = CallBreakPhasePlay
		cb.startPlayPhase()
	}
}

// startPlayPhase プレイフェーズ開始。リードプレイヤーはラウンド先頭は 0、
// それ以降は前のトリックの勝者が引き継ぐ (leadPlayerIdx として保存済み)。
func (cb *CallBreak) startPlayPhase() {
	cb.trickNumber = 1
	cb.currentTrick = nil
	cb.currentPlayerIdx = cb.leadPlayerIdx
}

// playCard カードをプレイする共通処理
func (cb *CallBreak) playCard(playerIdx int, card *Card) {
	cb.currentTrick = append(cb.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	if card.GetDesign() == CardDesignSpade {
		cb.spadesBroken = true
	}

	cb.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", cb.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(cb.currentTrick) == CallBreakPlayerCnt {
		cb.phase = CallBreakPhaseTrickEnd
	} else {
		cb.currentPlayerIdx = (cb.currentPlayerIdx + 1) % CallBreakPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
func (cb *CallBreak) validatePlay(playerIdx int, card *Card) error {
	if len(cb.currentTrick) == 0 {
		// リード: スペード未ブレイクの場合、スペードでリードできない (他にカードがある場合)
		if !cb.spadesBroken && card.GetDesign() == CardDesignSpade {
			if cb.playerHasNonSpade(playerIdx) {
				return NewDomainError(ErrInvalidPlay, "スペードはまだブレイクされていません")
			}
		}
		return nil
	}

	leadSuit := cb.currentTrick[0].Card.GetDesign()

	// フォロースート優先
	if cb.playerHasSuit(playerIdx, leadSuit) {
		if card.GetDesign() != leadSuit {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
		return nil
	}

	// ボイド: スペード (トランプ) を持っている場合は必ず切る必要がある
	if leadSuit != CardDesignSpade && cb.playerHasSuit(playerIdx, CardDesignSpade) {
		if card.GetDesign() != CardDesignSpade {
			return NewDomainError(ErrInvalidPlay, "リードスートが無い場合はスペードで切らなければなりません")
		}
	}
	return nil
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (cb *CallBreak) playerHasSuit(playerIdx, design int) bool {
	p := cb.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// playerHasNonSpade プレイヤーがスペード以外のカードを持っているか
func (cb *CallBreak) playerHasNonSpade(playerIdx int) bool {
	p := cb.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() != CardDesignSpade {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する (スペードがトランプ)
func (cb *CallBreak) trickWinner() int {
	return ResolveTrickWinner(cb.currentTrick, CardDesignSpade, nil)
}

// checkGameEnd ゲーム終了判定: MaxRounds 到達でゲーム終了、最高スコアが勝者
func (cb *CallBreak) checkGameEnd() {
	if cb.roundNumber < cb.config.MaxRounds {
		return
	}

	cb.gameEndFlag = true
	cb.phase = CallBreakPhaseGameEnd

	maxScore := cb.players[0].GetCumulativeScore()
	cb.winnerIdx = 0
	for i := 1; i < CallBreakPlayerCnt; i++ {
		if cb.players[i].GetCumulativeScore() > maxScore {
			maxScore = cb.players[i].GetCumulativeScore()
			cb.winnerIdx = i
		}
	}
	cb.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", cb.playerName(cb.winnerIdx)), nil)
}

// sortAllHands 全プレイヤーの手札をソートする (スート → 値)
func (cb *CallBreak) sortAllHands() {
	for _, p := range cb.players {
		callBreakSortHand(p)
	}
}

// callBreakSortHand プレイヤーの手札をスート → 値の順にソートする
func callBreakSortHand(p *CallBreakPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// playerName プレイヤー名を返す
func (cb *CallBreak) playerName(idx int) string {
	if idx < 0 || idx >= len(cb.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if cb.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (cb *CallBreak) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	cb.actionLog = append(cb.actionLog, &ActionLogEntry{
		TurnNumber: len(cb.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// GetHint ヒントを取得する
func (cb *CallBreak) GetHint() *CallBreakHint {
	if cb.phase == CallBreakPhaseBid && cb.bidPlayerIdx == 0 {
		bid := cb.cpuBidHard(0)
		return &CallBreakHint{Bid: &bid, Reason: "strategic_bid"}
	}
	if cb.phase == CallBreakPhasePlay && cb.currentPlayerIdx == 0 {
		validIndices := cb.getValidPlayIndices(0)
		if len(validIndices) == 0 {
			return nil
		}
		idx := cb.cpuPlayHard(0, validIndices)
		return &CallBreakHint{CardIndex: &idx, Reason: cb.playHintReason(idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由を判定する
func (cb *CallBreak) playHintReason(chosenIdx int) string {
	player := cb.players[0]
	card := player.GetCard(chosenIdx)

	if len(cb.currentTrick) == 0 {
		if player.GetTrickCount() < player.GetBid() {
			return "lead_strong"
		}
		return "lead_low"
	}

	leadSuit := cb.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == CardDesignSpade {
		return "trump_cut"
	}
	return "discard_high"
}

// --- CPU AI ---

// cpuSelectBid CPU がビッドを選択する
func (cb *CallBreak) cpuSelectBid(playerIdx int) int {
	switch cb.config.CpuDifficulty {
	case CallBreakCpuDifficultyHard:
		return cb.cpuBidHard(playerIdx)
	case CallBreakCpuDifficultyNormal:
		return cb.cpuBidNormal(playerIdx)
	default:
		return cb.cpuBidEasy(playerIdx)
	}
}

// cpuBidEasy ランダムに 1〜5 のビッド (最低 CallBreakMinBid)
func (cb *CallBreak) cpuBidEasy(_ int) int {
	return rand.Intn(5) + CallBreakMinBid
}

// cpuBidNormal カードの強さに基づくビッド
func (cb *CallBreak) cpuBidNormal(playerIdx int) int {
	player := cb.players[playerIdx]
	bid := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == CardDesignSpade {
			if card.GetValue() >= 10 {
				bid++
			}
		} else if card.GetValue() >= 12 {
			// 他スートの Q, K
			bid++
		}
	}
	return cb.clampBid(bid)
}

// cpuBidHard 戦略的なビッド
func (cb *CallBreak) cpuBidHard(playerIdx int) int {
	player := cb.players[playerIdx]
	bid := 0

	spadeCount := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == CardDesignSpade {
			spadeCount++
			if card.GetValue() >= 10 {
				bid++
			}
		} else if card.GetValue() == 13 {
			// K は概ね確実なトリック
			bid++
		} else if card.GetValue() == 12 {
			if rand.Intn(2) == 0 {
				bid++
			}
		}
	}

	if spadeCount >= 5 {
		bid++
	}

	return cb.clampBid(bid)
}

// clampBid ビッド値を [CallBreakMinBid, CallBreakHandSize] に収める
func (cb *CallBreak) clampBid(bid int) int {
	if bid < CallBreakMinBid {
		return CallBreakMinBid
	}
	if bid > CallBreakHandSize {
		return CallBreakHandSize
	}
	return bid
}

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選択する
func (cb *CallBreak) cpuSelectPlayCard(playerIdx int) int {
	validIndices := cb.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch cb.config.CpuDifficulty {
	case CallBreakCpuDifficultyHard:
		return cb.cpuPlayHard(playerIdx, validIndices)
	case CallBreakCpuDifficultyNormal:
		return cb.cpuPlayNormal(playerIdx, validIndices)
	default:
		return cb.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (cb *CallBreak) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal ビッドに近づくようにプレイ
func (cb *CallBreak) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := cb.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	if len(cb.currentTrick) == 0 {
		if tricks < bid {
			return pickHighest(player, validIndices, nil)
		}
		return pickLowest(player, validIndices, nil)
	}

	leadSuit := cb.currentTrick[0].Card.GetDesign()
	highestInTrick := 0
	for _, tc := range cb.currentTrick {
		if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
			highestInTrick = tc.Card.GetValue()
		}
	}

	leadSuitIndices := filterByDesign(player, validIndices, leadSuit)
	if len(leadSuitIndices) > 0 {
		if tricks < bid {
			over := filterAbove(player, leadSuitIndices, highestInTrick, nil)
			if len(over) > 0 {
				return pickLowest(player, over, nil)
			}
		}
		return pickLowest(player, leadSuitIndices, nil)
	}

	// ボイド: validatePlay により残ったカードはルール上有効なものだけ。
	// スペードを必ず切るルールがある場合、validIndices はスペードのみのはず。
	if tricks < bid {
		return pickLowest(player, validIndices, nil)
	}
	return pickLowest(player, validIndices, nil)
}

// cpuPlayHard 高度な戦略プレイ
func (cb *CallBreak) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := cb.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	if len(cb.currentTrick) == 0 {
		if tricks < bid {
			bestIdx := validIndices[0]
			bestScore := -1
			for _, idx := range validIndices {
				card := player.GetCard(idx)
				score := card.GetValue()
				if card.GetDesign() == CardDesignSpade {
					score += 100
				}
				if score > bestScore {
					bestScore = score
					bestIdx = idx
				}
			}
			return bestIdx
		}
		// 十分なトリック: 最も低い非スペードを出す
		bestIdx := validIndices[0]
		bestVal := player.GetCard(validIndices[0]).GetValue()
		bestIsSpade := player.GetCard(validIndices[0]).GetDesign() == CardDesignSpade
		for _, idx := range validIndices[1:] {
			card := player.GetCard(idx)
			isSpade := card.GetDesign() == CardDesignSpade
			if bestIsSpade && !isSpade {
				bestIdx = idx
				bestVal = card.GetValue()
				bestIsSpade = false
			} else if isSpade == bestIsSpade && card.GetValue() < bestVal {
				bestIdx = idx
				bestVal = card.GetValue()
			}
		}
		return bestIdx
	}

	leadSuit := cb.currentTrick[0].Card.GetDesign()
	highestSpadeInTrick, hasSpadeInTrick, highestInTrick := cb.summariseTrick(leadSuit)

	leadSuitIndices := filterByDesign(player, validIndices, leadSuit)
	if len(leadSuitIndices) > 0 {
		if tricks < bid && (!hasSpadeInTrick || leadSuit == CardDesignSpade) {
			threshold := highestInTrick
			if leadSuit == CardDesignSpade {
				threshold = highestSpadeInTrick
			}
			over := filterAbove(player, leadSuitIndices, threshold, nil)
			if len(over) > 0 {
				return pickLowest(player, over, nil)
			}
		}
		return pickLowest(player, leadSuitIndices, nil)
	}

	// ボイドかつ validIndices にスペードが含まれている場合は最小のスペードでカットを試みる。
	spadeIndices := filterByDesign(player, validIndices, CardDesignSpade)
	if len(spadeIndices) > 0 {
		if tricks < bid {
			if hasSpadeInTrick {
				if over := filterAbove(player, spadeIndices, highestSpadeInTrick, nil); len(over) > 0 {
					return pickLowest(player, over, nil)
				}
			} else {
				return pickLowest(player, spadeIndices, nil)
			}
		}
		// 余裕がある場合は最小スペードを温存気味に出す
		return pickLowest(player, spadeIndices, nil)
	}

	// スペードを持たないボイド: 最も高い不要カードを捨てる
	bestIdx := validIndices[0]
	bestScore := -1
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		score := card.GetValue()
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// summariseTrick 現在のトリックの状態 (最高スペード、スペード有無、リードスート最高) を返す
func (cb *CallBreak) summariseTrick(leadSuit int) (highestSpade int, hasSpade bool, highestLead int) {
	for _, tc := range cb.currentTrick {
		if tc.Card.GetDesign() == CardDesignSpade {
			hasSpade = true
			if tc.Card.GetValue() > highestSpade {
				highestSpade = tc.Card.GetValue()
			}
		}
		if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestLead {
			highestLead = tc.Card.GetValue()
		}
	}
	return
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (cb *CallBreak) getValidPlayIndices(playerIdx int) []int {
	player := cb.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return cb.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web 用)
func (cb *CallBreak) GetValidPlayIndices(playerIdx int) []int {
	return cb.getValidPlayIndices(playerIdx)
}

// callBreakJSON is the JSON wire format for CallBreak.
type callBreakJSON struct {
	TrumpCards       *TrumpCards        `json:"tc"`
	Players          []*CallBreakPlayer `json:"ps"`
	Config           CallBreakConfig    `json:"cf"`
	Phase            CallBreakPhase     `json:"ph"`
	RoundNumber      int                `json:"rn"`
	TrickNumber      int                `json:"tn"`
	CurrentPlayerIdx int                `json:"ci"`
	CurrentTrick     []*TrickCard       `json:"ct"`
	SpadesBroken     bool               `json:"sb"`
	LeadPlayerIdx    int                `json:"li"`
	BidPlayerIdx     int                `json:"bi"`
	GameEndFlag      bool               `json:"ge"`
	WinnerIdx        int                `json:"wi"`
	ActionLog        []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (cb *CallBreak) MarshalJSON() ([]byte, error) {
	return json.Marshal(callBreakJSON{
		TrumpCards:       cb.trumpCards,
		Players:          cb.players,
		Config:           cb.config,
		Phase:            cb.phase,
		RoundNumber:      cb.roundNumber,
		TrickNumber:      cb.trickNumber,
		CurrentPlayerIdx: cb.currentPlayerIdx,
		CurrentTrick:     cb.currentTrick,
		SpadesBroken:     cb.spadesBroken,
		LeadPlayerIdx:    cb.leadPlayerIdx,
		BidPlayerIdx:     cb.bidPlayerIdx,
		GameEndFlag:      cb.gameEndFlag,
		WinnerIdx:        cb.winnerIdx,
		ActionLog:        cb.actionLog,
	})
}

// callBreakMaxSliceLen caps small fixed-size slice fields during deserialisation
// (players: max 4, currentTrick: max 4) to prevent excessive memory allocation.
const callBreakMaxSliceLen = 1000

// callBreakMaxActionLogLen caps the ActionLog slice during deserialisation.
// ActionLog grows by ~70 entries per round (52 plays + 4 bids + 13 trick winners +
// 1 score entry); 5000 accommodates ~71 rounds while still bounding DoS risk.
const callBreakMaxActionLogLen = 5000

// UnmarshalJSON implements json.Unmarshaler.
func (cb *CallBreak) UnmarshalJSON(data []byte) error {
	var j callBreakJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > callBreakMaxSliceLen || len(j.CurrentTrick) > callBreakMaxSliceLen ||
		len(j.ActionLog) > callBreakMaxActionLogLen {
		return fmt.Errorf("callbreak: input array exceeds maximum allowed size")
	}
	cb.trumpCards = j.TrumpCards
	if cb.trumpCards == nil {
		cb.trumpCards = NewTrumpCards(0)
	}
	cb.players = j.Players
	if cb.players == nil {
		cb.players = make([]*CallBreakPlayer, 0)
	}
	cb.config = j.Config
	cb.phase = j.Phase
	cb.roundNumber = j.RoundNumber
	cb.trickNumber = j.TrickNumber
	cb.currentPlayerIdx = j.CurrentPlayerIdx
	cb.currentTrick = j.CurrentTrick
	if cb.currentTrick == nil {
		cb.currentTrick = make([]*TrickCard, 0)
	}
	cb.spadesBroken = j.SpadesBroken
	cb.leadPlayerIdx = j.LeadPlayerIdx
	cb.bidPlayerIdx = j.BidPlayerIdx
	cb.gameEndFlag = j.GameEndFlag
	cb.winnerIdx = j.WinnerIdx
	cb.actionLog = j.ActionLog
	if cb.actionLog == nil {
		cb.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
