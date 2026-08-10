package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// OhHellPlayerCnt オー・ヘルプレイヤー数
const OhHellPlayerCnt = 4

// OhHellPhase ゲームフェーズ
type OhHellPhase int

// OhHellのフェーズ定数
const (
	// OhHellPhaseBid ビッドフェーズ
	OhHellPhaseBid OhHellPhase = 0
	// OhHellPhasePlay トリックプレイフェーズ
	OhHellPhasePlay OhHellPhase = 1
	// OhHellPhaseTrickEnd トリック終了フェーズ
	OhHellPhaseTrickEnd OhHellPhase = 2
	// OhHellPhaseRoundEnd ラウンド終了フェーズ
	OhHellPhaseRoundEnd OhHellPhase = 3
	// OhHellPhaseGameEnd ゲーム終了フェーズ
	OhHellPhaseGameEnd OhHellPhase = 4
)

// OhHellHint ヒント情報
type OhHellHint struct {
	CardIndex *int   // 推奨カードインデックス (ビッド時nil)
	Bid       *int   // 推奨ビッド値 (プレイ時nil)
	Reason    string // ヒント理由キー
}

// OhHell オー・ヘルゲームクラス
type OhHell struct {
	trumpCards       *TrumpCards
	players          []*OhHellPlayer
	config           OhHellConfig
	phase            OhHellPhase
	roundNumber      int
	totalRounds      int
	handSize         int // 現在のラウンドの手札枚数
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	bidPlayerIdx     int
	dealerIdx        int
	trumpCard        *Card // 切り札カード (nil = 切り札なし)
	trumpSuit        int   // 切り札スート (-1 = 切り札なし)
	gameEndFlag      bool
	winnerIdx        int
	actionLogBase
}

// NewOhHell コンストラクタ
func NewOhHell(trumpCards *TrumpCards, players []*OhHellPlayer, config OhHellConfig) *OhHell {
	return &OhHell{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		winnerIdx:  -1,
		trumpSuit:  -1,
	}
}

// NewDefaultOhHell returns OhHell with the standard 4-player setup (1 human, 3 CPU)
// and DefaultOhHellConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultOhHell() *OhHell {
	players := []*OhHellPlayer{
		NewOhHellPlayer(true),
		NewOhHellPlayer(false),
		NewOhHellPlayer(false),
		NewOhHellPlayer(false),
	}
	return NewOhHell(NewTrumpCards(0), players, DefaultOhHellConfig())
}

// Reset ゲーム初期化
func (o *OhHell) Reset() {
	o.gameEndFlag = false
	o.winnerIdx = -1
	o.dealerIdx = 0
	o.totalRounds = o.calcTotalRounds()
	o.roundNumber = 1
	o.trickNumber = 0
	o.currentTrick = nil
	o.leadPlayerIdx = -1
	o.currentPlayerIdx = -1
	o.actionLog = nil

	for _, p := range o.players {
		p.bid = -1
		p.roundScore = 0
		p.cumulativeScore = 0
		p.tricksTaken = nil
		p.Reset()
		p.SetIsFinished(false)
	}

	o.handSize = o.handSizeForRound(o.roundNumber)
	o.deal()
	o.phase = OhHellPhaseBid
	o.bidPlayerIdx = (o.dealerIdx + 1) % OhHellPlayerCnt
}

// NextRound 次のラウンドを開始する
func (o *OhHell) NextRound() {
	if o.phase != OhHellPhaseRoundEnd {
		return
	}

	o.roundNumber++
	o.dealerIdx = (o.dealerIdx + 1) % OhHellPlayerCnt
	o.trickNumber = 0
	o.currentTrick = nil
	o.leadPlayerIdx = -1
	o.currentPlayerIdx = -1

	for _, p := range o.players {
		p.ResetRound()
	}

	o.handSize = o.handSizeForRound(o.roundNumber)
	o.deal()
	o.phase = OhHellPhaseBid
	o.bidPlayerIdx = (o.dealerIdx + 1) % OhHellPlayerCnt
}

// PlayerBid 人間プレイヤーがビッドする
func (o *OhHell) PlayerBid(bid int) error {
	if o.gameEndFlag {
		return ErrGameEnded
	}
	if o.phase != OhHellPhaseBid {
		return ErrWrongPhase
	}

	humanIdx := findHumanIdx(o.players)
	if humanIdx < 0 {
		return ErrNotHumanTurn
	}
	if o.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if bid < 0 || bid > o.handSize {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは0〜%dで指定してください", o.handSize))
	}
	if o.isRestrictedBid(bid) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ディーラーはビッド%dを選べません（合計がトリック数と一致するため）", bid))
	}

	o.players[humanIdx].SetBid(bid)
	o.appendLog(humanIdx, "bid", fmt.Sprintf("%s bids %d", playerName(o.players, humanIdx), bid), nil)

	o.advanceBid()
	return nil
}

// CpuBid 現在のビッドプレイヤーがCPUの場合にビッドする
func (o *OhHell) CpuBid() {
	if o.gameEndFlag || o.phase != OhHellPhaseBid {
		return
	}
	if o.bidPlayerIdx < 0 || o.bidPlayerIdx >= OhHellPlayerCnt {
		return
	}
	if o.players[o.bidPlayerIdx].GetIsHuman() {
		return
	}

	bid := o.cpuSelectBid(o.bidPlayerIdx)
	o.players[o.bidPlayerIdx].SetBid(bid)
	o.appendLog(o.bidPlayerIdx, "bid", fmt.Sprintf("%s bids %d", playerName(o.players, o.bidPlayerIdx), bid), nil)

	o.advanceBid()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (o *OhHell) PlayerPlay(cardIndex int) error {
	if o.gameEndFlag {
		return ErrGameEnded
	}
	if o.phase != OhHellPhasePlay {
		return ErrWrongPhase
	}
	if !o.players[o.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := o.players[o.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := o.validatePlay(o.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	o.playCard(o.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (o *OhHell) CpuPlay() {
	if o.gameEndFlag || o.phase != OhHellPhasePlay {
		return
	}
	if o.players[o.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := o.players[o.currentPlayerIdx]
	cardIdx := o.cpuSelectPlayCard(o.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	o.playCard(o.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (o *OhHell) ResolveTrick() {
	if o.phase != OhHellPhaseTrickEnd || len(o.currentTrick) != OhHellPlayerCnt {
		return
	}

	winnerIdx := o.trickWinner()
	trickCards := make([]*Card, len(o.currentTrick))
	for i, tc := range o.currentTrick {
		trickCards[i] = tc.Card
	}

	o.players[winnerIdx].AddTrick(trickCards)

	winnerName := playerName(o.players, winnerIdx)
	o.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d", winnerName, o.trickNumber), trickCards)

	o.leadPlayerIdx = winnerIdx

	if o.trickNumber >= o.handSize {
		o.phase = OhHellPhaseRoundEnd
	} else {
		o.phase = OhHellPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (o *OhHell) NextTrick() {
	if o.phase != OhHellPhaseTrickEnd {
		return
	}
	o.currentTrick = nil
	o.currentPlayerIdx = o.leadPlayerIdx
	o.trickNumber++
	o.phase = OhHellPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (o *OhHell) ScoreRound() {
	if o.phase != OhHellPhaseRoundEnd {
		return
	}

	for i := range OhHellPlayerCnt {
		p := o.players[i]
		tricks := p.GetTrickCount()
		bid := p.GetBid()

		if tricks == bid {
			// ビッド的中: 10 + bid ポイント
			p.roundScore = 10 + bid
			o.appendLog(i, "bid_success", fmt.Sprintf("%s bid %d, took %d: +%d",
				playerName(o.players, i), bid, tricks, p.roundScore), nil)
		} else {
			switch o.config.ScoringVariant {
			case OhHellScoringPenalty:
				diff := tricks - bid
				if diff < 0 {
					diff = -diff
				}
				p.roundScore = -diff
				o.appendLog(i, "bid_fail", fmt.Sprintf("%s bid %d, took %d: %d",
					playerName(o.players, i), bid, tricks, p.roundScore), nil)
			default:
				p.roundScore = 0
				o.appendLog(i, "bid_fail", fmt.Sprintf("%s bid %d, took %d: 0",
					playerName(o.players, i), bid, tricks), nil)
			}
		}
	}

	// 累積スコアに加算
	for i := range OhHellPlayerCnt {
		o.players[i].CommitRoundScore()
	}

	// スコアログ
	for i := range OhHellPlayerCnt {
		o.appendLog(i, "cumulative_score", fmt.Sprintf("%s: total=%d",
			playerName(o.players, i), o.players[i].cumulativeScore), nil)
	}

	// ゲーム終了判定
	if o.roundNumber >= o.totalRounds {
		o.gameEndFlag = true
		o.phase = OhHellPhaseGameEnd
		o.determineWinner()
	}
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (o *OhHell) GetPhase() OhHellPhase { return o.phase }

// SetPhase フェーズ設定 (テスト用)
func (o *OhHell) SetPhase(phase OhHellPhase) { o.phase = phase }

// GetRoundNumber 現在のラ���ンド番号取得
func (o *OhHell) GetRoundNumber() int { return o.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (o *OhHell) SetRoundNumber(n int) { o.roundNumber = n }

// GetTotalRounds 総ラウンド数取得
func (o *OhHell) GetTotalRounds() int { return o.totalRounds }

// SetTotalRounds 総ラウンド数設定 (テスト用)
func (o *OhHell) SetTotalRounds(n int) { o.totalRounds = n }

// GetHandSize 現在のラウンドの手札枚数取得
func (o *OhHell) GetHandSize() int { return o.handSize }

// SetHandSize 手札枚数設定 (テスト用)
func (o *OhHell) SetHandSize(n int) { o.handSize = n }

// GetTrickNumber 現在のトリック番号取得
func (o *OhHell) GetTrickNumber() int { return o.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (o *OhHell) SetTrickNumber(n int) { o.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (o *OhHell) GetCurrentPlayerIdx() int { return o.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (o *OhHell) SetCurrentPlayerIdx(idx int) { o.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (o *OhHell) GetCurrentTrick() []*TrickCard { return o.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (o *OhHell) SetCurrentTrick(trick []*TrickCard) { o.currentTrick = trick }

// GetTrumpCard 切り札カード取得
func (o *OhHell) GetTrumpCard() *Card { return o.trumpCard }

// GetTrumpSuit 切り札スート取得 (-1 = 切り札なし)
func (o *OhHell) GetTrumpSuit() int { return o.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (o *OhHell) SetTrumpSuit(suit int) { o.trumpSuit = suit }

// GetDealerIdx ディーラーインデックス取得
func (o *OhHell) GetDealerIdx() int { return o.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (o *OhHell) SetDealerIdx(idx int) { o.dealerIdx = idx }

// GetGameEndFlag ゲーム終了フラグ取得
func (o *OhHell) GetGameEndFlag() bool { return o.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (o *OhHell) GetWinnerIdx() int { return o.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (o *OhHell) GetPlayerCnt() int { return len(o.players) }

// GetPlayer プレイヤー取得
func (o *OhHell) GetPlayer(i int) *OhHellPlayer {
	return getPlayer(o.players, i)
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (o *OhHell) GetLeadPlayerIdx() int { return o.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (o *OhHell) SetLeadPlayerIdx(idx int) { o.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (o *OhHell) GetBidPlayerIdx() int { return o.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (o *OhHell) SetBidPlayerIdx(idx int) { o.bidPlayerIdx = idx }

// IsHumanTurn 現在の手番が人間かどうか
func (o *OhHell) IsHumanTurn() bool {
	return isHumanTurn(o.players, o.currentPlayerIdx)
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (o *OhHell) IsHumanBidTurn() bool {
	return isHumanTurn(o.players, o.bidPlayerIdx)
}

// GetConfig 設定取得
func (o *OhHell) GetConfig() OhHellConfig { return o.config }

// SetConfig 設定変更
func (o *OhHell) SetConfig(cfg OhHellConfig) { o.config = cfg }

// GetRestrictedBid ディーラーが選択できないビッド値を返す (-1 = 制限なし)
func (o *OhHell) GetRestrictedBid() int {
	if o.phase != OhHellPhaseBid {
		return -1
	}
	if o.bidPlayerIdx != o.dealerIdx {
		return -1
	}
	// ディーラーの番: 合計がhandSizeにならないビッドを強制
	totalBids := 0
	for _, p := range o.players {
		if p.GetBid() >= 0 {
			totalBids += p.GetBid()
		}
	}
	restricted := o.handSize - totalBids
	if restricted < 0 || restricted > o.handSize {
		return -1
	}
	return restricted
}

// GetValidPlayIndices プレイ可���なカードのインデックスリストを返す
func (o *OhHell) GetValidPlayIndices(playerIdx int) []int {
	return o.getValidPlayIndices(playerIdx)
}

// --- Private methods ---

// calcTotalRounds 総ラウンド数を計算する
func (o *OhHell) calcTotalRounds() int {
	max := o.config.MaxHandSize
	if o.config.RoundDirection == OhHellRoundDownAndUp {
		return max*2 - 1 // max → 1 → max (例: 10→1→10 = 19ラウンド)
	}
	return max // max → 1 (例: 10→1 = 10ラウンド)
}

// handSizeForRound ラウンド番号から手札枚数を算出する
func (o *OhHell) handSizeForRound(round int) int {
	max := o.config.MaxHandSize
	if round <= max {
		// 下降フェーズ: max, max-1, ..., 1
		return max - round + 1
	}
	// 上昇フェーズ: 2, 3, ..., max
	return round - max + 1
}

// deal カードを配り、切り札を公開する
func (o *OhHell) deal() {
	o.trumpCards = NewTrumpCards(0)
	o.trumpCards.Shuffle()

	// 各プレイヤーにhandSize枚配る (ディーラーの左から)
	for range o.handSize {
		for j := range OhHellPlayerCnt {
			idx := (o.dealerIdx + 1 + j) % OhHellPlayerCnt
			card := o.trumpCards.DrawCard()
			if card != nil {
				o.players[idx].AddCard(card)
			}
		}
	}

	// 残りの山札から1枚めくって切り札を決定
	o.trumpCard = o.trumpCards.DrawCard()
	if o.trumpCard != nil {
		o.trumpSuit = o.trumpCard.GetDesign()
	} else {
		o.trumpSuit = -1
	}

	o.sortAllHands()
}

// advanceBid ビッドプレイヤーを次に進める
func (o *OhHell) advanceBid() {
	// 次のビッドプレイヤーへ（ディーラーの左から時計回り、ディーラーが最後）
	nextIdx := (o.bidPlayerIdx + 1) % OhHellPlayerCnt
	bidCount := 0
	for _, p := range o.players {
		if p.GetBid() >= 0 {
			bidCount++
		}
	}
	if bidCount >= OhHellPlayerCnt {
		// 全員ビッド完了 → プレイフェーズへ
		o.phase = OhHellPhasePlay
		o.startPlayPhase()
	} else {
		o.bidPlayerIdx = nextIdx
	}
}

// isRestrictedBid ディーラー制限（フックルール）の判定
func (o *OhHell) isRestrictedBid(bid int) bool {
	if o.bidPlayerIdx != o.dealerIdx {
		return false
	}
	totalBids := 0
	for _, p := range o.players {
		if p.GetBid() >= 0 {
			totalBids += p.GetBid()
		}
	}
	return totalBids+bid == o.handSize
}

// startPlayPhase プ���イフェーズ開始: ディーラーの左がリード
// adjustBidForRestriction ディーラー制限を考慮してビッドを調整する
func (o *OhHell) adjustBidForRestriction(bid int) int {
	if o.isRestrictedBid(bid) {
		if bid > 0 {
			bid--
		} else {
			bid++
		}
	}
	return bid
}

func (o *OhHell) startPlayPhase() {
	o.trickNumber = 1
	o.currentTrick = nil
	o.leadPlayerIdx = (o.dealerIdx + 1) % OhHellPlayerCnt
	o.currentPlayerIdx = o.leadPlayerIdx
}

// playCard カードをプレイする共通処理
func (o *OhHell) playCard(playerIdx int, card *Card) {
	o.currentTrick = append(o.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	o.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(o.players, playerIdx), cardStr(card)), []*Card{card})

	if len(o.currentTrick) == OhHellPlayerCnt {
		o.phase = OhHellPhaseTrickEnd
	} else {
		o.currentPlayerIdx = (o.currentPlayerIdx + 1) % OhHellPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
func (o *OhHell) validatePlay(playerIdx int, card *Card) error {
	if len(o.currentTrick) == 0 {
		// リード: 制限なし（Oh Hellはトランプでのリードに制限がない）
		return nil
	}

	// フォロースート
	leadSuit := o.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit {
		if o.playerHasSuit(playerIdx, leadSuit) {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
	}

	return nil
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (o *OhHell) playerHasSuit(playerIdx int, design int) bool {
	return handHasSuit(o.players[playerIdx], design)
}

// trickWinner トリックの勝者を決定する
func (o *OhHell) trickWinner() int {
	return ResolveTrickWinner(o.currentTrick, o.trumpSuit, nil)
}

// determineWinner 最終的な勝者を決定する
func (o *OhHell) determineWinner() {
	maxScore := o.players[0].cumulativeScore
	o.winnerIdx = 0
	for i := 1; i < OhHellPlayerCnt; i++ {
		if o.players[i].cumulativeScore > maxScore {
			maxScore = o.players[i].cumulativeScore
			o.winnerIdx = i
		}
	}
	o.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(o.players, o.winnerIdx)), nil)
}

// sortAllHands 全プレイヤーの手札をソートする
func (o *OhHell) sortAllHands() {
	for _, p := range o.players {
		ohHellSortHand(p)
	}
}

// ohHellSortHand プレイヤーの手札をスート→値の順にソートする
func ohHellSortHand(p *OhHellPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (o *OhHell) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(o.players[playerIdx], func(c *Card) bool { return o.validatePlay(playerIdx, c) == nil })
}

// GetHint ヒントを取得する
func (o *OhHell) GetHint() *OhHellHint {
	humanIdx := findHumanIdx(o.players)
	if humanIdx < 0 {
		return nil
	}
	if o.phase == OhHellPhaseBid && o.bidPlayerIdx == humanIdx {
		bid := o.cpuBidHard(humanIdx)
		bid = o.adjustBidForRestriction(bid)
		return &OhHellHint{Bid: &bid, Reason: "strategic_bid"}
	}
	if o.phase == OhHellPhasePlay && o.currentPlayerIdx == humanIdx {
		validIndices := o.getValidPlayIndices(humanIdx)
		if len(validIndices) == 0 {
			return nil
		}
		idx := o.cpuPlayHard(humanIdx, validIndices)
		return &OhHellHint{CardIndex: &idx, Reason: o.playHintReason(humanIdx, idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由を判定する
func (o *OhHell) playHintReason(playerIdx int, chosenIdx int) string {
	player := o.players[playerIdx]
	card := player.GetCard(chosenIdx)

	if len(o.currentTrick) == 0 {
		if player.GetTrickCount() < player.GetBid() {
			return "lead_strong"
		}
		return "lead_low"
	}

	leadSuit := o.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if o.trumpSuit >= 0 && card.GetDesign() == o.trumpSuit {
		return "trump_cut"
	}
	return "discard_high"
}

// --- CPU AI ---

// cpuSelectBid CPUがビッドを選択する
func (o *OhHell) cpuSelectBid(playerIdx int) int {
	var bid int
	switch o.config.CpuDifficulty {
	case OhHellCpuDifficultyHard:
		bid = o.cpuBidHard(playerIdx)
	case OhHellCpuDifficultyNormal:
		bid = o.cpuBidNormal(playerIdx)
	default:
		bid = o.cpuBidEasy(playerIdx)
	}

	return o.adjustBidForRestriction(bid)
}

// cpuBidEasy ランダムなビッド
func (o *OhHell) cpuBidEasy(_ int) int {
	if o.handSize == 0 {
		return 0
	}
	return rand.Intn(o.handSize + 1)
}

// cpuBidNormal カードの強さに基づくビッド
func (o *OhHell) cpuBidNormal(playerIdx int) int {
	player := o.players[playerIdx]
	bid := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if o.trumpSuit >= 0 && card.GetDesign() == o.trumpSuit {
			// トランプの高いカード
			if card.GetValue() >= 10 {
				bid++
			}
		} else if card.GetValue() >= 12 {
			// 他のスートのK, A
			bid++
		}
	}
	if bid > o.handSize {
		bid = o.handSize
	}
	return bid
}

// cpuBidHard 戦略的なビッド
func (o *OhHell) cpuBidHard(playerIdx int) int {
	player := o.players[playerIdx]
	bid := 0

	trumpCount := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if o.trumpSuit >= 0 && card.GetDesign() == o.trumpSuit {
			trumpCount++
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

	// トランプが多い場合は追加トリックを見込む
	if trumpCount >= 3 {
		bid++
	}

	if bid > o.handSize {
		bid = o.handSize
	}
	return bid
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (o *OhHell) cpuSelectPlayCard(playerIdx int) int {
	validIndices := o.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch o.config.CpuDifficulty {
	case OhHellCpuDifficultyHard:
		return o.cpuPlayHard(playerIdx, validIndices)
	case OhHellCpuDifficultyNormal:
		return o.cpuPlayNormal(playerIdx, validIndices)
	default:
		return o.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (o *OhHell) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal ビッドに近づくようにプレイ
func (o *OhHell) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := o.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	if len(o.currentTrick) == 0 {
		// リード
		if tricks < bid {
			return o.highestCardIdx(player, validIndices)
		}
		return o.lowestCardIdx(player, validIndices)
	}

	// フォロー
	leadSuit := o.currentTrick[0].Card.GetDesign()
	if tricks < bid {
		// 勝ちに行く
		return o.tryWinTrick(player, validIndices, leadSuit)
	}
	// 負けに行く
	return o.tryLoseTrick(player, validIndices, leadSuit)
}

// cpuPlayHard 高度な戦略プレイ
func (o *OhHell) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := o.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	needMore := tricks < bid
	exactBid := tricks == bid

	if len(o.currentTrick) == 0 {
		// リード
		if needMore {
			// トランプの高札やA, Kでリード
			return o.bestLeadCard(player, validIndices, true)
		}
		// もうトリックが十分: 低いカードでリード
		return o.bestLeadCard(player, validIndices, false)
	}

	// フォロー
	leadSuit := o.currentTrick[0].Card.GetDesign()

	if needMore {
		return o.tryWinTrick(player, validIndices, leadSuit)
	}
	if exactBid {
		return o.tryLoseTrick(player, validIndices, leadSuit)
	}
	// オーバートリック: 最も低いカードを出す
	return o.lowestCardIdx(player, validIndices)
}

// --- CPU helper methods ---

// highestCardIdx 最も高いカードのインデックスを返す
func (o *OhHell) highestCardIdx(player *OhHellPlayer, indices []int) int {
	best := indices[0]
	for _, idx := range indices[1:] {
		if player.GetCard(idx).GetValue() > player.GetCard(best).GetValue() {
			best = idx
		}
	}
	return best
}

// lowestCardIdx 最も低いカードのインデックスを返す
func (o *OhHell) lowestCardIdx(player *OhHellPlayer, indices []int) int {
	best := indices[0]
	for _, idx := range indices[1:] {
		if player.GetCard(idx).GetValue() < player.GetCard(best).GetValue() {
			best = idx
		}
	}
	return best
}

// bestLeadCard リード時に最適なカードを選ぶ
func (o *OhHell) bestLeadCard(player *OhHellPlayer, indices []int, wantWin bool) int {
	if wantWin {
		// トランプ→高い方を優先
		best := indices[0]
		bestScore := -1
		for _, idx := range indices {
			card := player.GetCard(idx)
			score := card.GetValue()
			if o.trumpSuit >= 0 && card.GetDesign() == o.trumpSuit {
				score += 100
			}
			if score > bestScore {
				bestScore = score
				best = idx
			}
		}
		return best
	}
	// 負けたい: 最も低い非トランプ
	best := indices[0]
	bestVal := player.GetCard(indices[0]).GetValue()
	bestIsTrump := o.trumpSuit >= 0 && player.GetCard(indices[0]).GetDesign() == o.trumpSuit
	for _, idx := range indices[1:] {
		card := player.GetCard(idx)
		isTrump := o.trumpSuit >= 0 && card.GetDesign() == o.trumpSuit
		if bestIsTrump && !isTrump {
			best = idx
			bestVal = card.GetValue()
			bestIsTrump = false
		} else if isTrump == bestIsTrump && card.GetValue() < bestVal {
			best = idx
			bestVal = card.GetValue()
		}
	}
	return best
}

// tryWinTrick 勝ちに行くカードを選ぶ
func (o *OhHell) tryWinTrick(player *OhHellPlayer, validIndices []int, leadSuit int) int {
	highestInTrick := 0
	highestTrumpInTrick := 0
	hasTrumpInTrick := false
	for _, tc := range o.currentTrick {
		if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
			highestInTrick = tc.Card.GetValue()
		}
		if o.trumpSuit >= 0 && tc.Card.GetDesign() == o.trumpSuit {
			hasTrumpInTrick = true
			if tc.Card.GetValue() > highestTrumpInTrick {
				highestTrumpInTrick = tc.Card.GetValue()
			}
		}
	}

	// リードスートのカードがあれば
	hasLead := false
	for _, idx := range validIndices {
		if player.GetCard(idx).GetDesign() == leadSuit {
			hasLead = true
			break
		}
	}

	if hasLead && !hasTrumpInTrick {
		// リードスートで勝てるか
		overCards := []int{}
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			if card.GetDesign() == leadSuit && card.GetValue() > highestInTrick {
				overCards = append(overCards, idx)
			}
		}
		if len(overCards) > 0 {
			// 最小の勝てるカード
			best := overCards[0]
			for _, idx := range overCards[1:] {
				if player.GetCard(idx).GetValue() < player.GetCard(best).GetValue() {
					best = idx
				}
			}
			return best
		}
	}

	if !hasLead {
		// ボイド: トランプでカット
		if o.trumpSuit >= 0 {
			trumpCards := []int{}
			for _, idx := range validIndices {
				if player.GetCard(idx).GetDesign() == o.trumpSuit {
					trumpCards = append(trumpCards, idx)
				}
			}
			if len(trumpCards) > 0 {
				if hasTrumpInTrick {
					// 勝てるトランプのみ
					winnable := []int{}
					for _, idx := range trumpCards {
						if player.GetCard(idx).GetValue() > highestTrumpInTrick {
							winnable = append(winnable, idx)
						}
					}
					if len(winnable) > 0 {
						best := winnable[0]
						for _, idx := range winnable[1:] {
							if player.GetCard(idx).GetValue() < player.GetCard(best).GetValue() {
								best = idx
							}
						}
						return best
					}
				} else {
					// 最小のトランプ
					best := trumpCards[0]
					for _, idx := range trumpCards[1:] {
						if player.GetCard(idx).GetValue() < player.GetCard(best).GetValue() {
							best = idx
						}
					}
					return best
				}
			}
		}
	}

	// 勝てない場合は最も低いカードを出す
	return o.lowestCardIdx(player, validIndices)
}

// tryLoseTrick 負けに行くカードを選ぶ
func (o *OhHell) tryLoseTrick(player *OhHellPlayer, validIndices []int, leadSuit int) int {
	highestInTrick := 0
	for _, tc := range o.currentTrick {
		if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
			highestInTrick = tc.Card.GetValue()
		}
	}

	// リードスートのアンダーカード
	underCards := []int{}
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		if card.GetDesign() == leadSuit && card.GetValue() < highestInTrick {
			underCards = append(underCards, idx)
		}
	}
	if len(underCards) > 0 {
		// 最も高いアンダーカード（リードスートを温存しつつ負ける）
		best := underCards[0]
		for _, idx := range underCards[1:] {
			if player.GetCard(idx).GetValue() > player.GetCard(best).GetValue() {
				best = idx
			}
		}
		return best
	}

	// アンダーカードがない場合、最も低いカード
	return o.lowestCardIdx(player, validIndices)
}

// --- JSON serialization ---

// ohHellJSON is the JSON wire format for OhHell.
type ohHellJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*OhHellPlayer   `json:"ps"`
	Config           OhHellConfig      `json:"cf"`
	Phase            OhHellPhase       `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TotalRounds      int               `json:"tr"`
	HandSize         int               `json:"hs"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	LeadPlayerIdx    int               `json:"li"`
	BidPlayerIdx     int               `json:"bi"`
	DealerIdx        int               `json:"di"`
	TrumpCard        *Card             `json:"tp"`
	TrumpSuit        int               `json:"ts"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (o *OhHell) MarshalJSON() ([]byte, error) {
	return json.Marshal(ohHellJSON{
		TrumpCards:       o.trumpCards,
		Players:          o.players,
		Config:           o.config,
		Phase:            o.phase,
		RoundNumber:      o.roundNumber,
		TotalRounds:      o.totalRounds,
		HandSize:         o.handSize,
		TrickNumber:      o.trickNumber,
		CurrentPlayerIdx: o.currentPlayerIdx,
		CurrentTrick:     o.currentTrick,
		LeadPlayerIdx:    o.leadPlayerIdx,
		BidPlayerIdx:     o.bidPlayerIdx,
		DealerIdx:        o.dealerIdx,
		TrumpCard:        o.trumpCard,
		TrumpSuit:        o.trumpSuit,
		GameEndFlag:      o.gameEndFlag,
		WinnerIdx:        o.winnerIdx,
		ActionLog:        o.actionLog,
	})
}

// ohHellMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const ohHellMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (o *OhHell) UnmarshalJSON(data []byte) error {
	var j ohHellJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > ohHellMaxSliceLen || len(j.CurrentTrick) > ohHellMaxSliceLen ||
		len(j.ActionLog) > ohHellMaxSliceLen {
		return fmt.Errorf("ohhell: input array exceeds maximum allowed size")
	}
	o.trumpCards = j.TrumpCards
	if o.trumpCards == nil {
		o.trumpCards = NewTrumpCards(0)
	}
	o.players = j.Players
	if o.players == nil {
		o.players = make([]*OhHellPlayer, 0)
	}
	// インデックスの範囲チェック
	pCnt := len(j.Players)
	if pCnt > 0 {
		if j.CurrentPlayerIdx < -1 || j.CurrentPlayerIdx >= pCnt {
			return fmt.Errorf("ohhell: currentPlayerIdx %d out of range [−1, %d)", j.CurrentPlayerIdx, pCnt)
		}
		if j.DealerIdx < 0 || j.DealerIdx >= pCnt {
			return fmt.Errorf("ohhell: dealerIdx %d out of range [0, %d)", j.DealerIdx, pCnt)
		}
		if j.BidPlayerIdx < -1 || j.BidPlayerIdx >= pCnt {
			return fmt.Errorf("ohhell: bidPlayerIdx %d out of range [−1, %d)", j.BidPlayerIdx, pCnt)
		}
		if j.LeadPlayerIdx < -1 || j.LeadPlayerIdx >= pCnt {
			return fmt.Errorf("ohhell: leadPlayerIdx %d out of range [−1, %d)", j.LeadPlayerIdx, pCnt)
		}
	}
	if j.TrickNumber < 0 {
		return fmt.Errorf("ohhell: trickNumber %d must be >= 0", j.TrickNumber)
	}
	o.config = j.Config
	o.phase = j.Phase
	o.roundNumber = j.RoundNumber
	o.totalRounds = j.TotalRounds
	o.handSize = j.HandSize
	o.trickNumber = j.TrickNumber
	o.currentPlayerIdx = j.CurrentPlayerIdx
	o.currentTrick = j.CurrentTrick
	if o.currentTrick == nil {
		o.currentTrick = make([]*TrickCard, 0)
	}
	o.leadPlayerIdx = j.LeadPlayerIdx
	o.bidPlayerIdx = j.BidPlayerIdx
	o.dealerIdx = j.DealerIdx
	o.trumpCard = j.TrumpCard
	o.trumpSuit = j.TrumpSuit
	o.gameEndFlag = j.GameEndFlag
	o.winnerIdx = j.WinnerIdx
	o.actionLog = j.ActionLog
	if o.actionLog == nil {
		o.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
