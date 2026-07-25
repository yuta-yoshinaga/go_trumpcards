package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// SpadesPlayerCnt スペードプレイヤー数
const SpadesPlayerCnt = 4

// SpadesHandSize 各プレイヤーの手札枚数
const SpadesHandSize = 13

// SpadesPhase ゲームフェーズ
type SpadesPhase int

// Spadesのフェーズ定数
const (
	// SpadesPhaseBid ビッドフェーズ
	SpadesPhaseBid SpadesPhase = 0
	// SpadesPhasePlay トリックプレイフェーズ
	SpadesPhasePlay SpadesPhase = 1
	// SpadesPhaseTrickEnd トリック終了フェーズ
	SpadesPhaseTrickEnd SpadesPhase = 2
	// SpadesPhaseRoundEnd ラウンド終了フェーズ
	SpadesPhaseRoundEnd SpadesPhase = 3
	// SpadesPhaseGameEnd ゲーム終了フェーズ
	SpadesPhaseGameEnd SpadesPhase = 4
)

// SpadesHint ヒント情報
type SpadesHint struct {
	CardIndex *int   // 推奨カードインデックス (ビッド時nil)
	Bid       *int   // 推奨ビッド値 (プレイ時nil)
	Reason    string // ヒント理由キー
}

// SpadesLoseThreshold 負け閾値 (-200点以下で負け)
const SpadesLoseThreshold = -200

// Spades スペードゲームクラス
type Spades struct {
	trumpCards       *TrumpCards
	players          []*SpadesPlayer
	config           SpadesConfig
	phase            SpadesPhase
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

// NewSpades コンストラクタ
func NewSpades(trumpCards *TrumpCards, players []*SpadesPlayer, config SpadesConfig) *Spades {
	return &Spades{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
	}
}

// NewDefaultSpades returns Spades with the standard 4-player setup (1 human, 3 CPU)
// and DefaultSpadesConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultSpades() *Spades {
	players := []*SpadesPlayer{
		NewSpadesPlayer(true),
		NewSpadesPlayer(false),
		NewSpadesPlayer(false),
		NewSpadesPlayer(false),
	}
	return NewSpades(NewTrumpCards(0), players, DefaultSpadesConfig())
}

// Reset ゲーム初期化
func (s *Spades) Reset() {
	s.gameEndFlag = false
	s.winnerIdx = -1
	s.roundNumber = 1
	s.trickNumber = 0
	s.spadesBroken = false
	s.currentTrick = nil
	s.leadPlayerIdx = -1
	s.currentPlayerIdx = -1
	s.bidPlayerIdx = 0
	s.actionLog = nil

	for _, p := range s.players {
		p.bid = -1
		p.roundScore = 0
		p.cumulativeScore = 0
		p.bags = 0
		p.tricksTaken = nil
		p.Reset()
		p.SetIsFinished(false)
	}

	s.trumpCards.Shuffle()
	dealAllCards(s.trumpCards, s.players)
	s.sortAllHands()

	s.phase = SpadesPhaseBid
}

// NextRound 次のラウンドを開始する
func (s *Spades) NextRound() {
	if s.phase != SpadesPhaseRoundEnd {
		return
	}

	s.roundNumber++
	s.trickNumber = 0
	s.spadesBroken = false
	s.currentTrick = nil
	s.leadPlayerIdx = -1
	s.currentPlayerIdx = -1
	s.bidPlayerIdx = 0

	for _, p := range s.players {
		p.ResetRound()
	}

	s.trumpCards.Shuffle()
	dealAllCards(s.trumpCards, s.players)
	s.sortAllHands()

	s.phase = SpadesPhaseBid
}

// PlayerBid 人間プレイヤーがビッドする
func (s *Spades) PlayerBid(bid int) error {
	if s.gameEndFlag {
		return ErrGameEnded
	}
	if s.phase != SpadesPhaseBid {
		return ErrWrongPhase
	}

	humanIdx := s.findHumanIdx()
	if humanIdx < 0 {
		return ErrNotHumanTurn
	}
	if s.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if bid < 0 || bid > SpadesHandSize {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは0〜%dで指定してください", SpadesHandSize))
	}

	s.players[humanIdx].SetBid(bid)
	bidStr := fmt.Sprintf("%d", bid)
	if bid == 0 {
		bidStr = "Nil"
	}
	s.appendLog(humanIdx, "bid", fmt.Sprintf("%s bids %s", s.playerName(humanIdx), bidStr), nil)

	s.bidPlayerIdx++
	s.checkBidComplete()
	return nil
}

// CpuBid 現在のビッドプレイヤーがCPUの場合にビッドする
func (s *Spades) CpuBid() {
	if s.gameEndFlag || s.phase != SpadesPhaseBid {
		return
	}
	if s.bidPlayerIdx >= SpadesPlayerCnt {
		return
	}
	if s.players[s.bidPlayerIdx].GetIsHuman() {
		return
	}

	bid := s.cpuSelectBid(s.bidPlayerIdx)
	s.players[s.bidPlayerIdx].SetBid(bid)
	bidStr := fmt.Sprintf("%d", bid)
	if bid == 0 {
		bidStr = "Nil"
	}
	s.appendLog(s.bidPlayerIdx, "bid", fmt.Sprintf("%s bids %s", s.playerName(s.bidPlayerIdx), bidStr), nil)

	s.bidPlayerIdx++
	s.checkBidComplete()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (s *Spades) PlayerPlay(cardIndex int) error {
	if s.gameEndFlag {
		return ErrGameEnded
	}
	if s.phase != SpadesPhasePlay {
		return ErrWrongPhase
	}
	if !s.players[s.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := s.players[s.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := s.validatePlay(s.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	s.playCard(s.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (s *Spades) CpuPlay() {
	if s.gameEndFlag || s.phase != SpadesPhasePlay {
		return
	}
	if s.players[s.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := s.players[s.currentPlayerIdx]
	cardIdx := s.cpuSelectPlayCard(s.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	s.playCard(s.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (s *Spades) ResolveTrick() {
	if s.phase != SpadesPhaseTrickEnd || len(s.currentTrick) != SpadesPlayerCnt {
		return
	}

	winnerIdx := s.trickWinner()
	trickCards := make([]*Card, len(s.currentTrick))
	for i, tc := range s.currentTrick {
		trickCards[i] = tc.Card
	}

	s.players[winnerIdx].AddTrick(trickCards)

	winnerName := s.playerName(winnerIdx)
	s.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d", winnerName, s.trickNumber), trickCards)

	s.leadPlayerIdx = winnerIdx

	if s.trickNumber >= SpadesHandSize {
		s.phase = SpadesPhaseRoundEnd
	} else {
		s.phase = SpadesPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (s *Spades) NextTrick() {
	if s.phase != SpadesPhaseTrickEnd {
		return
	}
	s.currentTrick = nil
	s.currentPlayerIdx = s.leadPlayerIdx
	s.trickNumber++
	s.phase = SpadesPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (s *Spades) ScoreRound() {
	if s.phase != SpadesPhaseRoundEnd {
		return
	}

	for i := 0; i < SpadesPlayerCnt; i++ {
		p := s.players[i]
		tricks := p.GetTrickCount()
		bid := p.GetBid()

		if bid == 0 {
			// ニルビッド
			if tricks == 0 {
				p.roundScore = s.config.NilBonus
				s.appendLog(i, "nil_success", fmt.Sprintf("%s nil success! +%d", s.playerName(i), s.config.NilBonus), nil)
			} else {
				p.roundScore = -s.config.NilBonus
				s.appendLog(i, "nil_fail", fmt.Sprintf("%s nil failed! -%d (%d tricks taken)", s.playerName(i), s.config.NilBonus, tricks), nil)
			}
		} else if tricks >= bid {
			// ビッド成功
			overtricks := tricks - bid
			p.roundScore = bid*10 + overtricks
			p.bags += overtricks

			// バッグペナルティ判定
			if s.config.BagPenaltyThreshold > 0 && p.bags >= s.config.BagPenaltyThreshold {
				penalty := (p.bags / s.config.BagPenaltyThreshold) * 100
				p.roundScore -= penalty
				p.bags %= s.config.BagPenaltyThreshold
				s.appendLog(i, "bag_penalty", fmt.Sprintf("%s bag penalty! -%d", s.playerName(i), penalty), nil)
			}
		} else {
			// ビッド失敗
			p.roundScore = -bid * 10
		}

		s.appendLog(i, "round_score", fmt.Sprintf("%s: bid=%d tricks=%d round=%d bags=%d",
			s.playerName(i), bid, tricks, p.roundScore, p.bags), nil)
	}

	// 累積スコアに加算
	for i := 0; i < SpadesPlayerCnt; i++ {
		s.players[i].CommitRoundScore()
	}

	// スコアログ
	for i := 0; i < SpadesPlayerCnt; i++ {
		s.appendLog(i, "cumulative_score", fmt.Sprintf("%s: total=%d",
			s.playerName(i), s.players[i].cumulativeScore), nil)
	}

	// ゲーム終了判定
	s.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (s *Spades) GetPhase() SpadesPhase { return s.phase }

// SetPhase フェーズ設定 (テスト用)
func (s *Spades) SetPhase(phase SpadesPhase) { s.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (s *Spades) GetRoundNumber() int { return s.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (s *Spades) SetRoundNumber(n int) { s.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (s *Spades) GetTrickNumber() int { return s.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (s *Spades) SetTrickNumber(n int) { s.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (s *Spades) GetCurrentPlayerIdx() int { return s.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (s *Spades) SetCurrentPlayerIdx(idx int) { s.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (s *Spades) GetCurrentTrick() []*TrickCard { return s.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (s *Spades) SetCurrentTrick(trick []*TrickCard) { s.currentTrick = trick }

// GetSpadesBroken スペードブレイク状態取得
func (s *Spades) GetSpadesBroken() bool { return s.spadesBroken }

// SetSpadesBroken スペードブレイク状態設定 (テスト用)
func (s *Spades) SetSpadesBroken(broken bool) { s.spadesBroken = broken }

// GetGameEndFlag ゲーム終了フラグ取得
func (s *Spades) GetGameEndFlag() bool { return s.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (s *Spades) GetWinnerIdx() int { return s.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (s *Spades) GetPlayerCnt() int { return len(s.players) }

// GetPlayer プレイヤー取得
func (s *Spades) GetPlayer(i int) *SpadesPlayer {
	if i < 0 || i >= len(s.players) {
		return nil
	}
	return s.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (s *Spades) GetLeadPlayerIdx() int { return s.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (s *Spades) SetLeadPlayerIdx(idx int) { s.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (s *Spades) GetBidPlayerIdx() int { return s.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (s *Spades) SetBidPlayerIdx(idx int) { s.bidPlayerIdx = idx }

// IsHumanTurn 現在の手番が人間かどうか
func (s *Spades) IsHumanTurn() bool {
	if s.currentPlayerIdx < 0 || s.currentPlayerIdx >= len(s.players) {
		return false
	}
	return s.players[s.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (s *Spades) IsHumanBidTurn() bool {
	if s.bidPlayerIdx < 0 || s.bidPlayerIdx >= len(s.players) {
		return false
	}
	return s.players[s.bidPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (s *Spades) GetConfig() SpadesConfig { return s.config }

// SetConfig 設定変更
func (s *Spades) SetConfig(cfg SpadesConfig) { s.config = cfg }

// GetActionLog 棋譜取得
func (s *Spades) GetActionLog() []*ActionLogEntry { return s.actionLog }

// --- Private methods ---

// findHumanIdx 人間プレイヤーのインデックスを返す (-1=なし)
func (s *Spades) findHumanIdx() int {
	for i, p := range s.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// checkBidComplete 全員がビッドしたかチェックし、プレイフェーズに移行
func (s *Spades) checkBidComplete() {
	if s.bidPlayerIdx >= SpadesPlayerCnt {
		s.phase = SpadesPhasePlay
		s.startPlayPhase()
	}
}

// startPlayPhase プレイフェーズ開始: 2♣を持つプレイヤーをリードに設定
func (s *Spades) startPlayPhase() {
	if s.trickNumber == 0 {
		starter := s.findTwoOfClubs()
		if starter >= 0 {
			s.leadPlayerIdx = starter
			s.currentPlayerIdx = starter
		} else {
			s.leadPlayerIdx = 0
			s.currentPlayerIdx = 0
		}
		s.trickNumber = 1
		s.currentTrick = nil
	} else {
		s.currentPlayerIdx = s.leadPlayerIdx
	}
}

// findTwoOfClubs 2♣を持つプレイヤーのインデックスを返す
func (s *Spades) findTwoOfClubs() int {
	for i, p := range s.players {
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
func (s *Spades) playCard(playerIdx int, card *Card) {
	s.currentTrick = append(s.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	// スペードブレイクチェック
	if card.GetDesign() == CardDesignSpade {
		s.spadesBroken = true
	}

	s.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", s.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(s.currentTrick) == SpadesPlayerCnt {
		s.phase = SpadesPhaseTrickEnd
	} else {
		s.currentPlayerIdx = (s.currentPlayerIdx + 1) % SpadesPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
func (s *Spades) validatePlay(playerIdx int, card *Card) error {
	// 最初のトリックの最初のカードは2♣でなければならない
	if s.trickNumber == 1 && len(s.currentTrick) == 0 {
		if card.GetDesign() != CardDesignClover || card.GetValue() != 2 {
			if s.playerHasCard(playerIdx, CardDesignClover, 2) {
				return NewDomainError(ErrInvalidPlay, "最初のトリックは2♣でリードしてください")
			}
		}
	}

	if len(s.currentTrick) == 0 {
		// リード: スペードが壊れていない場合、スペードでリードできない（他にカードがある場合）
		if !s.spadesBroken && card.GetDesign() == CardDesignSpade {
			if s.playerHasNonSpade(playerIdx) {
				return NewDomainError(ErrInvalidPlay, "スペードはまだブレイクされていません")
			}
		}
		return nil
	}

	// フォロースート
	leadSuit := s.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit {
		if s.playerHasSuit(playerIdx, leadSuit) {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
	}

	return nil
}

// playerHasCard プレイヤーが特定のカードを持っているか
func (s *Spades) playerHasCard(playerIdx int, design, value int) bool {
	p := s.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == design && c.GetValue() == value {
			return true
		}
	}
	return false
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (s *Spades) playerHasSuit(playerIdx int, design int) bool {
	p := s.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// playerHasNonSpade プレイヤーがスペード以外のカードを持っているか
func (s *Spades) playerHasNonSpade(playerIdx int) bool {
	p := s.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() != CardDesignSpade {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する (スペードがトランプ)
func (s *Spades) trickWinner() int {
	return ResolveTrickWinner(s.currentTrick, CardDesignSpade, nil)
}

// checkGameEnd ゲーム終了判定
func (s *Spades) checkGameEnd() {
	// 勝利判定: PointLimit以上のプレイヤーがいるか
	hasWinner := false
	for i := 0; i < SpadesPlayerCnt; i++ {
		if s.players[i].cumulativeScore >= s.config.PointLimit {
			hasWinner = true
			break
		}
	}

	// 敗北判定: LoseThreshold以下のプレイヤーがいるか
	hasLoser := false
	for i := 0; i < SpadesPlayerCnt; i++ {
		if s.players[i].cumulativeScore <= SpadesLoseThreshold {
			hasLoser = true
			break
		}
	}

	if !hasWinner && !hasLoser {
		return
	}

	s.gameEndFlag = true
	s.phase = SpadesPhaseGameEnd

	// 最高スコアのプレイヤーが勝者
	maxScore := s.players[0].cumulativeScore
	s.winnerIdx = 0
	for i := 1; i < SpadesPlayerCnt; i++ {
		if s.players[i].cumulativeScore > maxScore {
			maxScore = s.players[i].cumulativeScore
			s.winnerIdx = i
		}
	}
	s.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", s.playerName(s.winnerIdx)), nil)
}

// sortAllHands 全プレイヤーの手札をソートする
func (s *Spades) sortAllHands() {
	for _, p := range s.players {
		spadeSortHand(p)
	}
}

// spadeSortHand プレイヤーの手札をスート→値の順にソートする
func spadeSortHand(p *SpadesPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// playerName プレイヤー名を返す
func (s *Spades) playerName(idx int) string {
	if idx < 0 || idx >= len(s.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if s.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (s *Spades) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.actionLog = append(s.actionLog, &ActionLogEntry{
		TurnNumber: len(s.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// GetHint ヒントを取得する
func (s *Spades) GetHint() *SpadesHint {
	if s.phase == SpadesPhaseBid && s.bidPlayerIdx == 0 {
		bid := s.cpuBidHard(0)
		return &SpadesHint{Bid: &bid, Reason: "strategic_bid"}
	}
	if s.phase == SpadesPhasePlay && s.currentPlayerIdx == 0 {
		validIndices := s.getValidPlayIndices(0)
		if len(validIndices) == 0 {
			return nil
		}
		idx := s.cpuPlayHard(0, validIndices)
		return &SpadesHint{CardIndex: &idx, Reason: s.playHintReason(idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由を判定する
func (s *Spades) playHintReason(chosenIdx int) string {
	player := s.players[0]
	card := player.GetCard(chosenIdx)

	if len(s.currentTrick) == 0 {
		if player.GetTrickCount() < player.GetBid() {
			return "lead_strong"
		}
		return "lead_low"
	}

	leadSuit := s.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == CardDesignSpade {
		return "trump_cut"
	}
	return "discard_high"
}

// --- CPU AI ---

// cpuSelectBid CPUがビッドを選択する
func (s *Spades) cpuSelectBid(playerIdx int) int {
	switch s.config.CpuDifficulty {
	case SpadesCpuDifficultyHard:
		return s.cpuBidHard(playerIdx)
	case SpadesCpuDifficultyNormal:
		return s.cpuBidNormal(playerIdx)
	default:
		return s.cpuBidEasy(playerIdx)
	}
}

// cpuBidEasy ランダムに1〜5のビッド
func (s *Spades) cpuBidEasy(playerIdx int) int {
	return rand.Intn(5) + 1
}

// cpuBidNormal カードの強さに基づくビッド
func (s *Spades) cpuBidNormal(playerIdx int) int {
	player := s.players[playerIdx]
	bid := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == CardDesignSpade {
			// スペードの高いカードをカウント
			if card.GetValue() >= 10 {
				bid++
			}
		} else if card.GetValue() >= 12 {
			// 他のスートのK, A
			bid++
		}
	}
	if bid < 1 {
		bid = 1
	}
	return bid
}

// cpuBidHard 戦略的なビッド
func (s *Spades) cpuBidHard(playerIdx int) int {
	player := s.players[playerIdx]
	bid := 0

	// スペードの枚数と高札をカウント
	spadeCount := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == CardDesignSpade {
			spadeCount++
			if card.GetValue() >= 10 {
				bid++
			}
		} else {
			// A, K は確実なトリック
			if card.GetValue() == 1 || card.GetValue() == 13 {
				bid++
			} else if card.GetValue() == 12 {
				// Qは半確実
				if rand.Intn(2) == 0 {
					bid++
				}
			}
		}
	}

	// スペードの枚数が多い場合は追加トリックを見込む
	if spadeCount >= 5 {
		bid++
	}

	if bid < 1 {
		bid = 1
	}
	if bid > SpadesHandSize {
		bid = SpadesHandSize
	}
	return bid
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (s *Spades) cpuSelectPlayCard(playerIdx int) int {
	validIndices := s.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch s.config.CpuDifficulty {
	case SpadesCpuDifficultyHard:
		return s.cpuPlayHard(playerIdx, validIndices)
	case SpadesCpuDifficultyNormal:
		return s.cpuPlayNormal(playerIdx, validIndices)
	default:
		return s.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (s *Spades) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal ビッドに近づくようにプレイ
func (s *Spades) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := s.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	if len(s.currentTrick) == 0 {
		// リード
		if tricks < bid {
			// トリックが足りない: 高いカードでリード
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
		// 十分なトリック: 最も低いカードを出す
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

	// フォロー
	leadSuit := s.currentTrick[0].Card.GetDesign()
	hasLeadSuit := false
	for _, idx := range validIndices {
		if player.GetCard(idx).GetDesign() == leadSuit {
			hasLeadSuit = true
			break
		}
	}

	if hasLeadSuit {
		highestInTrick := 0
		for _, tc := range s.currentTrick {
			if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
				highestInTrick = tc.Card.GetValue()
			}
		}

		if tricks < bid {
			// 勝ちに行く: 最小の勝てるカード
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
		}
		// アンダーカードを出す
		underCards := []int{}
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			if card.GetDesign() == leadSuit && card.GetValue() < highestInTrick {
				underCards = append(underCards, idx)
			}
		}
		if len(underCards) > 0 {
			bestIdx := underCards[0]
			for _, idx := range underCards[1:] {
				if player.GetCard(idx).GetValue() > player.GetCard(bestIdx).GetValue() {
					bestIdx = idx
				}
			}
			return bestIdx
		}
		// 最も低いリードスートカード
		bestIdx := validIndices[0]
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			if card.GetDesign() == leadSuit && (player.GetCard(bestIdx).GetDesign() != leadSuit || card.GetValue() < player.GetCard(bestIdx).GetValue()) {
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// ボイド: トリックが足りなければスペード(トランプ)を使う
	if tricks < bid {
		for _, idx := range validIndices {
			if player.GetCard(idx).GetDesign() == CardDesignSpade {
				return idx
			}
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
func (s *Spades) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := s.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	if len(s.currentTrick) == 0 {
		// リード
		if tricks < bid {
			// Aやスペードの高札でリード
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

	// フォロー
	leadSuit := s.currentTrick[0].Card.GetDesign()

	// トリックにスペード(トランプ)が含まれるか
	highestSpadeInTrick := 0
	hasSpadeInTrick := false
	highestInTrick := 0
	for _, tc := range s.currentTrick {
		if tc.Card.GetDesign() == CardDesignSpade {
			hasSpadeInTrick = true
			if tc.Card.GetValue() > highestSpadeInTrick {
				highestSpadeInTrick = tc.Card.GetValue()
			}
		}
		if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
			highestInTrick = tc.Card.GetValue()
		}
	}

	// フォロー可能なら勝ち札/受け札を選ぶ (フォロー時 validIndices は全てリードスート)。
	leadSuitIndices := filterByDesign(player, validIndices, leadSuit)
	if len(leadSuitIndices) > 0 {
		// 勝ちに行く (トランプが出ていなければリードスートで勝てる)。
		if tricks < bid && (!hasSpadeInTrick || leadSuit == CardDesignSpade) {
			threshold := highestInTrick
			if leadSuit == CardDesignSpade {
				threshold = highestSpadeInTrick
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

	// ボイド: スペードでカット or 低いカードを捨てる。
	spadeIndices := filterByDesign(player, validIndices, CardDesignSpade)
	if tricks < bid && len(spadeIndices) > 0 {
		if hasSpadeInTrick {
			// すでにスペードが出ている場合、勝てる最小のスペードのみ。
			if over := filterAbove(player, spadeIndices, highestSpadeInTrick, nil); len(over) > 0 {
				return pickLowest(player, over, nil)
			}
		} else {
			return pickLowest(player, spadeIndices, nil)
		}
	}

	// 最も高い不要カードを捨てる
	bestIdx := validIndices[0]
	bestScore := -1
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		score := card.GetValue()
		// スペードは温存したい
		if card.GetDesign() == CardDesignSpade {
			score -= 100
		}
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (s *Spades) getValidPlayIndices(playerIdx int) []int {
	player := s.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return s.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (s *Spades) GetValidPlayIndices(playerIdx int) []int {
	return s.getValidPlayIndices(playerIdx)
}

// spadesJSON is the JSON wire format for Spades.
type spadesJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*SpadesPlayer   `json:"ps"`
	Config           SpadesConfig      `json:"cf"`
	Phase            SpadesPhase       `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	SpadesBroken     bool              `json:"sb"`
	LeadPlayerIdx    int               `json:"li"`
	BidPlayerIdx     int               `json:"bi"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (s *Spades) MarshalJSON() ([]byte, error) {
	return json.Marshal(spadesJSON{
		TrumpCards:       s.trumpCards,
		Players:          s.players,
		Config:           s.config,
		Phase:            s.phase,
		RoundNumber:      s.roundNumber,
		TrickNumber:      s.trickNumber,
		CurrentPlayerIdx: s.currentPlayerIdx,
		CurrentTrick:     s.currentTrick,
		SpadesBroken:     s.spadesBroken,
		LeadPlayerIdx:    s.leadPlayerIdx,
		BidPlayerIdx:     s.bidPlayerIdx,
		GameEndFlag:      s.gameEndFlag,
		WinnerIdx:        s.winnerIdx,
		ActionLog:        s.actionLog,
	})
}

// spadesMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const spadesMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (s *Spades) UnmarshalJSON(data []byte) error {
	var j spadesJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > spadesMaxSliceLen || len(j.CurrentTrick) > spadesMaxSliceLen ||
		len(j.ActionLog) > spadesMaxSliceLen {
		return fmt.Errorf("spades: input array exceeds maximum allowed size")
	}
	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCards(0)
	}
	s.players = j.Players
	if s.players == nil {
		s.players = make([]*SpadesPlayer, 0)
	}
	s.config = j.Config
	s.phase = j.Phase
	s.roundNumber = j.RoundNumber
	s.trickNumber = j.TrickNumber
	s.currentPlayerIdx = j.CurrentPlayerIdx
	s.currentTrick = j.CurrentTrick
	if s.currentTrick == nil {
		s.currentTrick = make([]*TrickCard, 0)
	}
	s.spadesBroken = j.SpadesBroken
	s.leadPlayerIdx = j.LeadPlayerIdx
	s.bidPlayerIdx = j.BidPlayerIdx
	s.gameEndFlag = j.GameEndFlag
	s.winnerIdx = j.WinnerIdx
	s.actionLog = j.ActionLog
	if s.actionLog == nil {
		s.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
