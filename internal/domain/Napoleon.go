//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// NapoleonPlayerCnt ナポレオンプレイヤー数
const NapoleonPlayerCnt = 4

// NapoleonHandSize 各プレイヤーの手札枚数
const NapoleonHandSize = 13

// NapoleonTotalCards デッキ総枚数 (52 + Joker 1)
const NapoleonTotalCards = 53

// NapoleonMaxPictureCards 絵札の最大枚数 (J,Q,K,A × 4スート + Joker = 17)
const NapoleonMaxPictureCards = 17

// NapoleonPhase ゲームフェーズ
type NapoleonPhase int

// ナポレオンのフェーズ定数
const (
	// NapoleonPhaseBid ビッドフェーズ
	NapoleonPhaseBid NapoleonPhase = 0
	// NapoleonPhaseTrumpDeclaration 切り札宣言＋副官指名フェーズ
	NapoleonPhaseTrumpDeclaration NapoleonPhase = 1
	// NapoleonPhaseKittyExchange 場札交換フェーズ
	NapoleonPhaseKittyExchange NapoleonPhase = 2
	// NapoleonPhasePlay トリックプレイフェーズ
	NapoleonPhasePlay NapoleonPhase = 3
	// NapoleonPhaseTrickEnd トリック終了フェーズ
	NapoleonPhaseTrickEnd NapoleonPhase = 4
	// NapoleonPhaseRoundEnd ラウンド終了フェーズ
	NapoleonPhaseRoundEnd NapoleonPhase = 5
	// NapoleonPhaseGameEnd ゲーム終了フェーズ
	NapoleonPhaseGameEnd NapoleonPhase = 6
)

// NapoleonWinnerTeam 勝利チーム
const (
	// NapoleonWinnerUndecided 未確定
	NapoleonWinnerUndecided = -1
	// NapoleonWinnerNapoleon ナポレオン軍の勝利
	NapoleonWinnerNapoleon = 0
	// NapoleonWinnerAllied 連合軍の勝利
	NapoleonWinnerAllied = 1
)

// NapoleonHint ヒント情報
type NapoleonHint struct {
	CardIndex     *int   // 推奨カードインデックス
	Bid           *int   // 推奨ビッド値
	TrumpSuit     *int   // 推奨切り札スート
	AdjutantSuit  *int   // 推奨副官カードスート
	AdjutantValue *int   // 推奨副官カード値
	DiscardIndex  *int   // 推奨捨てカードインデックス
	Reason        string // ヒント理由キー
}

// NapoleonTrickCard トリック中の1枚
type NapoleonTrickCard struct {
	PlayerIdx int
	Card      *Card
}

// napoleonRoundState ラウンドごとにリセットされる状態
type napoleonRoundState struct {
	phase            NapoleonPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*NapoleonTrickCard
	trumpSuit        int   // 切り札スート (CardDesignSpade〜CardDesignDiamond)
	adjutantCard     *Card // 副官指名カード
	napoleonIdx      int   // ナポレオンのプレイヤーインデックス
	adjutantIdx      int   // 副官のプレイヤーインデックス (-1 = 不明/自分自身)
	adjutantRevealed bool  // 副官が全体に公開されたか
	leadPlayerIdx    int
	bidPlayerIdx     int
	kitty            []*Card // 場札 (1枚)
	highestBid       int     // 現在の最高ビッド
	highestBidder    int     // 最高ビッドしたプレイヤー
	passCount        int     // パスしたプレイヤー数
	gameEndFlag      bool
	winnerTeam       int // NapoleonWinnerUndecided / NapoleonWinnerNapoleon / NapoleonWinnerAllied
	actionLog        []*ActionLogEntry
}

// Napoleon ナポレオンゲームクラス
type Napoleon struct {
	trumpCards *TrumpCards
	players    []*NapoleonPlayer
	config     NapoleonConfig
	round      napoleonRoundState
}

// NewNapoleon コンストラクタ
func NewNapoleon(trumpCards *TrumpCards, players []*NapoleonPlayer, config NapoleonConfig) *Napoleon {
	return &Napoleon{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: napoleonRoundState{
			winnerTeam:    NapoleonWinnerUndecided,
			napoleonIdx:   -1,
			adjutantIdx:   -1,
			highestBidder: -1,
		},
	}
}

// NewDefaultNapoleon returns Napoleon with the standard 4-player setup (1 human, 3 CPU)
// and DefaultNapoleonConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultNapoleon() *Napoleon {
	players := []*NapoleonPlayer{
		NewNapoleonPlayer(true),
		NewNapoleonPlayer(false),
		NewNapoleonPlayer(false),
		NewNapoleonPlayer(false),
	}
	return NewNapoleon(NewTrumpCards(1), players, DefaultNapoleonConfig())
}

// Reset ゲーム初期化
func (n *Napoleon) Reset() {
	n.round = napoleonRoundState{
		roundNumber:      1,
		leadPlayerIdx:    -1,
		currentPlayerIdx: -1,
		napoleonIdx:      -1,
		adjutantIdx:      -1,
		highestBidder:    -1,
		winnerTeam:       NapoleonWinnerUndecided,
	}

	for _, p := range n.players {
		p.bid = -1
		p.isNapoleon = false
		p.isAdjutant = false
		p.adjutantRevealed = false
		p.pictureCards = 0
		p.roundScore = 0
		p.cumulativeScore = 0
		p.tricksTaken = nil
		p.Reset()
		p.SetIsFinished(false)
	}

	n.dealCards()
	n.sortAllHands()

	n.round.phase = NapoleonPhaseBid
}

// NextRound 次のラウンドを開始する
func (n *Napoleon) NextRound() {
	if n.round.phase != NapoleonPhaseRoundEnd {
		return
	}

	prevRound := n.round.roundNumber
	n.round = napoleonRoundState{
		roundNumber:      prevRound + 1,
		leadPlayerIdx:    -1,
		currentPlayerIdx: -1,
		napoleonIdx:      -1,
		adjutantIdx:      -1,
		highestBidder:    -1,
		winnerTeam:       NapoleonWinnerUndecided,
	}

	for _, p := range n.players {
		p.ResetRound()
	}

	n.dealCards()
	n.sortAllHands()

	n.round.phase = NapoleonPhaseBid
}

// PlayerBid 人間プレイヤーがビッドする (0 = パス)
func (n *Napoleon) PlayerBid(bid int) error {
	if n.round.gameEndFlag {
		return ErrGameEnded
	}
	if n.round.phase != NapoleonPhaseBid {
		return ErrWrongPhase
	}

	humanIdx := n.findHumanIdx()
	if humanIdx < 0 || n.round.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}

	if bid != 0 {
		if bid < n.config.MinBid || bid > NapoleonMaxPictureCards {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは%d〜%dで指定してください（0でパス）", n.config.MinBid, NapoleonMaxPictureCards))
		}
		if bid <= n.round.highestBid {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("現在の最高ビッド%dより高い値を指定してください", n.round.highestBid))
		}
	}

	n.applyBid(humanIdx, bid)
	return nil
}

// CpuBid 現在のビッドプレイヤーがCPUの場合にビッドする
func (n *Napoleon) CpuBid() {
	if n.round.gameEndFlag || n.round.phase != NapoleonPhaseBid {
		return
	}
	if n.round.bidPlayerIdx >= NapoleonPlayerCnt {
		return
	}
	if n.players[n.round.bidPlayerIdx].GetIsHuman() {
		return
	}

	bid := n.cpuSelectBid(n.round.bidPlayerIdx)
	n.applyBid(n.round.bidPlayerIdx, bid)
}

// PlayerDeclareTrump 人間プレイヤーが切り札と副官を宣言する
func (n *Napoleon) PlayerDeclareTrump(suit int, adjSuit int, adjVal int) error {
	if n.round.gameEndFlag {
		return ErrGameEnded
	}
	if n.round.phase != NapoleonPhaseTrumpDeclaration {
		return ErrWrongPhase
	}
	if n.round.napoleonIdx < 0 || !n.players[n.round.napoleonIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "無効なスートです")
	}
	if adjSuit == CardDesignJoker {
		// ジョーカーを副官に指名
		if adjVal != 1 {
			return NewDomainError(ErrInvalidPlay, "ジョーカーのvalueは1です")
		}
	} else {
		if adjSuit < CardDesignSpade || adjSuit > CardDesignDiamond {
			return NewDomainError(ErrInvalidPlay, "無効な副官スートです")
		}
		if adjVal < 1 || adjVal > CardValueMax {
			return NewDomainError(ErrInvalidPlay, "無効な副官カード値です")
		}
	}

	n.applyDeclareTrump(suit, adjSuit, adjVal)
	return nil
}

// CpuDeclareTrump CPUナポレオンが切り札と副官を宣言する
func (n *Napoleon) CpuDeclareTrump() {
	if n.round.gameEndFlag || n.round.phase != NapoleonPhaseTrumpDeclaration {
		return
	}
	if n.round.napoleonIdx < 0 || n.players[n.round.napoleonIdx].GetIsHuman() {
		return
	}

	suit, adjSuit, adjVal := n.cpuSelectTrump(n.round.napoleonIdx)
	n.applyDeclareTrump(suit, adjSuit, adjVal)
}

// PlayerExchangeKitty 人間ナポレオンが場札を交換する (捨てるカードのインデックス指定)
func (n *Napoleon) PlayerExchangeKitty(discardIndex int) error {
	if n.round.gameEndFlag {
		return ErrGameEnded
	}
	if n.round.phase != NapoleonPhaseKittyExchange {
		return ErrWrongPhase
	}
	if n.round.napoleonIdx < 0 || !n.players[n.round.napoleonIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := n.players[n.round.napoleonIdx]
	// 場札はすでに手札に追加されている (14枚)
	if discardIndex < 0 || discardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	n.applyExchangeKitty(discardIndex)
	return nil
}

// CpuExchangeKitty CPUナポレオンが場札を交換する
func (n *Napoleon) CpuExchangeKitty() {
	if n.round.gameEndFlag || n.round.phase != NapoleonPhaseKittyExchange {
		return
	}
	if n.round.napoleonIdx < 0 || n.players[n.round.napoleonIdx].GetIsHuman() {
		return
	}

	discardIdx := n.cpuSelectDiscard(n.round.napoleonIdx)
	n.applyExchangeKitty(discardIdx)
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (n *Napoleon) PlayerPlay(cardIndex int) error {
	if n.round.gameEndFlag {
		return ErrGameEnded
	}
	if n.round.phase != NapoleonPhasePlay {
		return ErrWrongPhase
	}
	if !n.players[n.round.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := n.players[n.round.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := n.validatePlay(n.round.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	n.playCard(n.round.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (n *Napoleon) CpuPlay() {
	if n.round.gameEndFlag || n.round.phase != NapoleonPhasePlay {
		return
	}
	if n.players[n.round.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := n.players[n.round.currentPlayerIdx]
	cardIdx := n.cpuSelectPlayCard(n.round.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	n.playCard(n.round.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (n *Napoleon) ResolveTrick() {
	if n.round.phase != NapoleonPhaseTrickEnd || len(n.round.currentTrick) != NapoleonPlayerCnt {
		return
	}

	winnerIdx := n.trickWinner()
	trickCards := make([]*Card, len(n.round.currentTrick))
	for i, tc := range n.round.currentTrick {
		trickCards[i] = tc.Card
	}

	n.players[winnerIdx].AddTrick(trickCards)

	// 絵札カウント
	picCount := n.countPictureCards(trickCards)
	n.players[winnerIdx].pictureCards += picCount

	winnerName := n.playerName(winnerIdx)
	s := fmt.Sprintf("%s wins trick %d", winnerName, n.round.trickNumber)
	if picCount > 0 {
		s += fmt.Sprintf(" (+%d picture cards)", picCount)
	}
	n.appendLog(winnerIdx, "trick_win", s, trickCards)

	n.round.leadPlayerIdx = winnerIdx

	if n.round.trickNumber >= NapoleonHandSize {
		n.round.phase = NapoleonPhaseRoundEnd
	} else {
		n.round.phase = NapoleonPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (n *Napoleon) NextTrick() {
	if n.round.phase != NapoleonPhaseTrickEnd {
		return
	}
	n.round.currentTrick = nil
	n.round.currentPlayerIdx = n.round.leadPlayerIdx
	n.round.trickNumber++
	n.round.phase = NapoleonPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (n *Napoleon) ScoreRound() {
	if n.round.phase != NapoleonPhaseRoundEnd {
		return
	}

	// ナポレオン軍の絵札合計
	napoleonTeamPictures := 0
	for i := range NapoleonPlayerCnt {
		if n.players[i].isNapoleon || n.players[i].isAdjutant {
			napoleonTeamPictures += n.players[i].pictureCards
		}
	}

	bid := n.round.highestBid
	napoleonWon := napoleonTeamPictures >= bid

	if napoleonWon {
		n.round.winnerTeam = NapoleonWinnerNapoleon
		n.appendLog(-1, "round_result", fmt.Sprintf("Napoleon's team wins! (%d/%d picture cards)", napoleonTeamPictures, bid), nil)
	} else {
		n.round.winnerTeam = NapoleonWinnerAllied
		n.appendLog(-1, "round_result", fmt.Sprintf("Allied forces win! (%d/%d picture cards)", napoleonTeamPictures, bid), nil)
	}

	// スコア計算: ナポレオン軍勝利 → ナポレオン+bid, 副官+bid/2, 連合軍-bid
	//            連合軍勝利 → ナポレオン-bid*2, 副官-bid, 連合軍+bid
	for i := range NapoleonPlayerCnt {
		p := n.players[i]
		if napoleonWon {
			if p.isNapoleon {
				p.roundScore = bid
			} else if p.isAdjutant {
				p.roundScore = bid / 2
			} else {
				p.roundScore = -bid
			}
		} else {
			if p.isNapoleon {
				p.roundScore = -bid * 2
			} else if p.isAdjutant {
				p.roundScore = -bid
			} else {
				p.roundScore = bid
			}
		}
		n.appendLog(i, "round_score", fmt.Sprintf("%s: round=%d", n.playerName(i), p.roundScore), nil)
	}

	// 累積スコアに加算
	for i := range NapoleonPlayerCnt {
		n.players[i].CommitRoundScore()
	}

	// 累積スコアログ
	for i := range NapoleonPlayerCnt {
		n.appendLog(i, "cumulative_score", fmt.Sprintf("%s: total=%d",
			n.playerName(i), n.players[i].cumulativeScore), nil)
	}

	// ゲーム終了判定
	n.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (n *Napoleon) GetPhase() NapoleonPhase { return n.round.phase }

// SetPhase フェーズ設定 (テスト用)
func (n *Napoleon) SetPhase(phase NapoleonPhase) { n.round.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (n *Napoleon) GetRoundNumber() int { return n.round.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (n *Napoleon) SetRoundNumber(n2 int) { n.round.roundNumber = n2 }

// GetTrickNumber 現在のトリック番号取得
func (n *Napoleon) GetTrickNumber() int { return n.round.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (n *Napoleon) SetTrickNumber(t int) { n.round.trickNumber = t }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (n *Napoleon) GetCurrentPlayerIdx() int { return n.round.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (n *Napoleon) SetCurrentPlayerIdx(idx int) { n.round.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (n *Napoleon) GetCurrentTrick() []*NapoleonTrickCard { return n.round.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (n *Napoleon) SetCurrentTrick(trick []*NapoleonTrickCard) { n.round.currentTrick = trick }

// GetTrumpSuit 切り札スート取得
func (n *Napoleon) GetTrumpSuit() int { return n.round.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (n *Napoleon) SetTrumpSuit(suit int) { n.round.trumpSuit = suit }

// GetAdjutantCard 副官カード取得
func (n *Napoleon) GetAdjutantCard() *Card { return n.round.adjutantCard }

// SetAdjutantCard 副官カード設定 (テスト用)
func (n *Napoleon) SetAdjutantCard(card *Card) { n.round.adjutantCard = card }

// GetNapoleonIdx ナポレオンインデックス取得
func (n *Napoleon) GetNapoleonIdx() int { return n.round.napoleonIdx }

// SetNapoleonIdx ナポレオンインデックス設定 (テスト用)
func (n *Napoleon) SetNapoleonIdx(idx int) { n.round.napoleonIdx = idx }

// GetAdjutantIdx 副官インデックス取得
func (n *Napoleon) GetAdjutantIdx() int { return n.round.adjutantIdx }

// SetAdjutantIdx 副官インデックス設定 (テスト用)
func (n *Napoleon) SetAdjutantIdx(idx int) { n.round.adjutantIdx = idx }

// GetAdjutantRevealed 副官公開状態取得
func (n *Napoleon) GetAdjutantRevealed() bool { return n.round.adjutantRevealed }

// SetAdjutantRevealed 副官公開状態設定 (テスト用)
func (n *Napoleon) SetAdjutantRevealed(v bool) { n.round.adjutantRevealed = v }

// GetGameEndFlag ゲーム終了フラグ取得
func (n *Napoleon) GetGameEndFlag() bool { return n.round.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (n *Napoleon) SetGameEndFlag(flag bool) { n.round.gameEndFlag = flag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (n *Napoleon) GetWinnerTeam() int { return n.round.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (n *Napoleon) GetPlayerCnt() int { return len(n.players) }

// GetPlayer プレイヤー取得
func (n *Napoleon) GetPlayer(i int) *NapoleonPlayer {
	if i < 0 || i >= len(n.players) {
		return nil
	}
	return n.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (n *Napoleon) GetLeadPlayerIdx() int { return n.round.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (n *Napoleon) SetLeadPlayerIdx(idx int) { n.round.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (n *Napoleon) GetBidPlayerIdx() int { return n.round.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (n *Napoleon) SetBidPlayerIdx(idx int) { n.round.bidPlayerIdx = idx }

// GetKitty 場札取得
func (n *Napoleon) GetKitty() []*Card { return n.round.kitty }

// SetKitty 場札設定 (テスト用)
func (n *Napoleon) SetKitty(kitty []*Card) { n.round.kitty = kitty }

// GetHighestBid 現在の最高ビッド取得
func (n *Napoleon) GetHighestBid() int { return n.round.highestBid }

// SetHighestBid 最高ビッド設定 (テスト用)
func (n *Napoleon) SetHighestBid(bid int) { n.round.highestBid = bid }

// GetHighestBidder 最高ビッドプレイヤー取得
func (n *Napoleon) GetHighestBidder() int { return n.round.highestBidder }

// SetHighestBidder 最高ビッドプレイヤー設定 (テスト用)
func (n *Napoleon) SetHighestBidder(idx int) { n.round.highestBidder = idx }

// GetPassCount パス数取得
func (n *Napoleon) GetPassCount() int { return n.round.passCount }

// SetPassCount パス数設定 (テスト用)
func (n *Napoleon) SetPassCount(cnt int) { n.round.passCount = cnt }

// IsHumanTurn 現在の手番が人間かどうか
func (n *Napoleon) IsHumanTurn() bool {
	if n.round.currentPlayerIdx < 0 || n.round.currentPlayerIdx >= len(n.players) {
		return false
	}
	return n.players[n.round.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (n *Napoleon) IsHumanBidTurn() bool {
	if n.round.bidPlayerIdx < 0 || n.round.bidPlayerIdx >= len(n.players) {
		return false
	}
	return n.players[n.round.bidPlayerIdx].GetIsHuman()
}

// IsHumanDeclareTurn 切り札宣言が人間の番かどうか
func (n *Napoleon) IsHumanDeclareTurn() bool {
	if n.round.napoleonIdx < 0 || n.round.napoleonIdx >= len(n.players) {
		return false
	}
	return n.players[n.round.napoleonIdx].GetIsHuman()
}

// IsHumanExchangeTurn 場札交換が人間の番かどうか
func (n *Napoleon) IsHumanExchangeTurn() bool {
	if n.round.napoleonIdx < 0 || n.round.napoleonIdx >= len(n.players) {
		return false
	}
	return n.players[n.round.napoleonIdx].GetIsHuman()
}

// GetConfig 設定取得
func (n *Napoleon) GetConfig() NapoleonConfig { return n.config }

// SetConfig 設定変更
func (n *Napoleon) SetConfig(cfg NapoleonConfig) { n.config = cfg }

// GetActionLog 棋譜取得
func (n *Napoleon) GetActionLog() []*ActionLogEntry { return n.round.actionLog }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (n *Napoleon) GetValidPlayIndices(playerIdx int) []int {
	return n.getValidPlayIndices(playerIdx)
}

// GetHint ヒントを取得する
func (n *Napoleon) GetHint() *NapoleonHint {
	humanIdx := n.findHumanIdx()
	if humanIdx < 0 {
		return nil
	}

	switch n.round.phase {
	case NapoleonPhaseBid:
		if n.round.bidPlayerIdx != humanIdx {
			return nil
		}
		bid := n.cpuBidHard(humanIdx)
		return &NapoleonHint{Bid: &bid, Reason: "strategic_bid"}

	case NapoleonPhaseTrumpDeclaration:
		if n.round.napoleonIdx != humanIdx {
			return nil
		}
		suit, adjSuit, adjVal := n.cpuSelectTrumpHard(humanIdx)
		return &NapoleonHint{TrumpSuit: &suit, AdjutantSuit: &adjSuit, AdjutantValue: &adjVal, Reason: "strategic_declare"}

	case NapoleonPhaseKittyExchange:
		if n.round.napoleonIdx != humanIdx {
			return nil
		}
		discardIdx := n.cpuSelectDiscardHard(humanIdx)
		return &NapoleonHint{DiscardIndex: &discardIdx, Reason: "strategic_discard"}

	case NapoleonPhasePlay:
		if n.round.currentPlayerIdx != humanIdx {
			return nil
		}
		validIndices := n.getValidPlayIndices(humanIdx)
		if len(validIndices) == 0 {
			return nil
		}
		idx := n.cpuPlayHard(humanIdx, validIndices)
		return &NapoleonHint{CardIndex: &idx, Reason: n.playHintReason(idx)}
	}

	return nil
}

// --- Private methods ---

// findHumanIdx 人間プレイヤーのインデックスを返す (-1=なし)
func (n *Napoleon) findHumanIdx() int {
	for i, p := range n.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// dealCards 53枚を配る: 13枚×4人 + 場札1枚
func (n *Napoleon) dealCards() {
	n.trumpCards.Shuffle()
	// 13枚ずつラウンドロビンで配る
	for range NapoleonHandSize {
		for i := range NapoleonPlayerCnt {
			card := n.trumpCards.DrawCard()
			if card != nil {
				n.players[i].AddCard(card)
			}
		}
	}
	// 残り1枚を場札に
	kittyCard := n.trumpCards.DrawCard()
	if kittyCard != nil {
		n.round.kitty = []*Card{kittyCard}
	}
}

// applyBid ビッドを適用する共通処理
func (n *Napoleon) applyBid(playerIdx int, bid int) {
	n.players[playerIdx].SetBid(bid)

	if bid == 0 {
		n.appendLog(playerIdx, "bid", fmt.Sprintf("%s passes", n.playerName(playerIdx)), nil)
		n.round.passCount++
	} else {
		n.appendLog(playerIdx, "bid", fmt.Sprintf("%s bids %d", n.playerName(playerIdx), bid), nil)
		n.round.highestBid = bid
		n.round.highestBidder = playerIdx
	}

	n.round.bidPlayerIdx++
	n.checkBidComplete()
}

// checkBidComplete 全員がビッドしたかチェック
func (n *Napoleon) checkBidComplete() {
	if n.round.bidPlayerIdx < NapoleonPlayerCnt {
		return
	}

	if n.round.highestBidder < 0 {
		// 全員パス: 最初のプレイヤーが最低ビッドで強制ナポレオン
		n.round.highestBid = n.config.MinBid
		n.round.highestBidder = 0
		n.players[0].SetBid(n.config.MinBid)
		n.appendLog(0, "forced_bid", fmt.Sprintf("%s is forced to bid %d (all pass)", n.playerName(0), n.config.MinBid), nil)
	}

	n.round.napoleonIdx = n.round.highestBidder
	n.players[n.round.napoleonIdx].SetIsNapoleon(true)
	n.appendLog(n.round.napoleonIdx, "napoleon", fmt.Sprintf("%s becomes Napoleon (bid %d)", n.playerName(n.round.napoleonIdx), n.round.highestBid), nil)

	n.round.phase = NapoleonPhaseTrumpDeclaration
}

// applyDeclareTrump 切り札宣言を適用する
func (n *Napoleon) applyDeclareTrump(suit int, adjSuit int, adjVal int) {
	n.round.trumpSuit = suit
	n.round.adjutantCard = NewCard(adjSuit, adjVal, false)

	suitNames := map[int]string{
		CardDesignSpade: "Spade", CardDesignClover: "Club",
		CardDesignHeart: "Heart", CardDesignDiamond: "Diamond",
		CardDesignJoker: "Joker",
	}
	n.appendLog(n.round.napoleonIdx, "declare_trump",
		fmt.Sprintf("%s declares %s as trump", n.playerName(n.round.napoleonIdx), suitNames[suit]), nil)
	n.appendLog(n.round.napoleonIdx, "declare_adjutant",
		fmt.Sprintf("%s names %s as adjutant card", n.playerName(n.round.napoleonIdx), napoleonCardStr(n.round.adjutantCard)), nil)

	// 副官を特定
	n.round.adjutantIdx = n.findAdjutantHolder()
	if n.round.adjutantIdx >= 0 {
		n.players[n.round.adjutantIdx].SetIsAdjutant(true)
		// 自分自身を指名した場合
		if n.round.adjutantIdx == n.round.napoleonIdx {
			n.round.adjutantRevealed = true
			n.players[n.round.adjutantIdx].SetAdjutantRevealed(true)
		}
	}

	// 場札をナポレオンの手札に追加
	for _, c := range n.round.kitty {
		n.players[n.round.napoleonIdx].AddCard(c)
	}
	n.sortHand(n.players[n.round.napoleonIdx])

	n.round.phase = NapoleonPhaseKittyExchange
}

// applyExchangeKitty 場札交換を適用する
func (n *Napoleon) applyExchangeKitty(discardIndex int) {
	player := n.players[n.round.napoleonIdx]
	discarded := player.RemoveCard(discardIndex)
	n.round.kitty = []*Card{discarded}

	n.appendLog(n.round.napoleonIdx, "exchange", fmt.Sprintf("%s exchanges kitty card", n.playerName(n.round.napoleonIdx)), nil)

	n.sortHand(player)
	n.startPlayPhase()
}

// startPlayPhase プレイフェーズ開始: ナポレオンがリード
func (n *Napoleon) startPlayPhase() {
	n.round.leadPlayerIdx = n.round.napoleonIdx
	n.round.currentPlayerIdx = n.round.napoleonIdx
	n.round.trickNumber = 1
	n.round.currentTrick = nil
	n.round.phase = NapoleonPhasePlay
}

// playCard カードをプレイする共通処理
func (n *Napoleon) playCard(playerIdx int, card *Card) {
	n.round.currentTrick = append(n.round.currentTrick, &NapoleonTrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	n.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", n.playerName(playerIdx), napoleonCardStr(card)), []*Card{card})

	// 副官カードが出されたら公開
	n.checkAdjutantReveal(playerIdx, card)

	if len(n.round.currentTrick) == NapoleonPlayerCnt {
		n.round.phase = NapoleonPhaseTrickEnd
	} else {
		n.round.currentPlayerIdx = (n.round.currentPlayerIdx + 1) % NapoleonPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
func (n *Napoleon) validatePlay(playerIdx int, card *Card) error {
	if len(n.round.currentTrick) == 0 {
		// リード: 制限なし
		return nil
	}

	leadCard := n.round.currentTrick[0].Card
	// ジョーカーがリードされた場合、何でも出せる
	if leadCard.GetDesign() == CardDesignJoker {
		return nil
	}

	leadSuit := leadCard.GetDesign()

	// ジョーカーはいつでも出せる
	if card.GetDesign() == CardDesignJoker {
		return nil
	}

	// フォロースート
	if card.GetDesign() != leadSuit {
		if n.playerHasSuit(playerIdx, leadSuit) {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
	}

	return nil
}

// trickWinner トリックの勝者を決定する
func (n *Napoleon) trickWinner() int {
	if len(n.round.currentTrick) == 0 {
		return 0
	}

	// ジョーカーとスペード3の特殊判定
	jokerIdx := -1
	spade3Idx := -1
	for _, tc := range n.round.currentTrick {
		if tc.Card.GetDesign() == CardDesignJoker {
			jokerIdx = tc.PlayerIdx
		}
		if tc.Card.GetDesign() == CardDesignSpade && tc.Card.GetValue() == 3 {
			spade3Idx = tc.PlayerIdx
		}
	}

	// ジョーカーキラー: ジョーカーとスペード3が両方ある場合、スペード3が勝つ
	if jokerIdx >= 0 && spade3Idx >= 0 {
		return spade3Idx
	}
	// ジョーカー単独: ジョーカーが勝つ
	if jokerIdx >= 0 {
		return jokerIdx
	}

	// リードスートを特定
	leadSuit := n.round.currentTrick[0].Card.GetDesign()

	winnerIdx := n.round.currentTrick[0].PlayerIdx
	winnerValue := n.cardStrength(n.round.currentTrick[0].Card)
	winnerIsTrump := n.round.currentTrick[0].Card.GetDesign() == n.round.trumpSuit

	for _, tc := range n.round.currentTrick[1:] {
		isTrump := tc.Card.GetDesign() == n.round.trumpSuit
		value := n.cardStrength(tc.Card)

		if isTrump && !winnerIsTrump {
			// 切り札がリードスートに勝つ
			winnerIdx = tc.PlayerIdx
			winnerValue = value
			winnerIsTrump = true
		} else if isTrump && winnerIsTrump {
			// 切り札同士: 高い方が勝つ
			if value > winnerValue {
				winnerIdx = tc.PlayerIdx
				winnerValue = value
			}
		} else if !isTrump && !winnerIsTrump && tc.Card.GetDesign() == leadSuit && value > winnerValue {
			// 非切り札同士: リードスートの高い方が勝つ
			winnerIdx = tc.PlayerIdx
			winnerValue = value
		}
	}
	return winnerIdx
}

// cardStrength カードの強さを返す (A=14, K=13, ..., 2=2)
func (n *Napoleon) cardStrength(card *Card) int {
	if card.GetValue() == 1 {
		return 14 // Aは最強
	}
	return card.GetValue()
}

// isPictureCard 絵札かどうか (J, Q, K, A, Joker)
func (n *Napoleon) isPictureCard(card *Card) bool {
	if card.GetDesign() == CardDesignJoker {
		return true
	}
	v := card.GetValue()
	return v == 1 || v == 11 || v == 12 || v == 13 // A, J, Q, K
}

// countPictureCards カード配列中の絵札数をカウント
func (n *Napoleon) countPictureCards(cards []*Card) int {
	count := 0
	for _, c := range cards {
		if n.isPictureCard(c) {
			count++
		}
	}
	return count
}

// findAdjutantHolder 副官カードの所有者を探す
func (n *Napoleon) findAdjutantHolder() int {
	if n.round.adjutantCard == nil {
		return -1
	}
	for i, p := range n.players {
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c.GetDesign() == n.round.adjutantCard.GetDesign() && c.GetValue() == n.round.adjutantCard.GetValue() {
				return i
			}
		}
	}
	// 場札にあるかもしれない (まだ配られていない)
	for _, c := range n.round.kitty {
		if c.GetDesign() == n.round.adjutantCard.GetDesign() && c.GetValue() == n.round.adjutantCard.GetValue() {
			// 場札にある場合、ナポレオンが場札交換時に取得する
			return -1
		}
	}
	return -1
}

// checkAdjutantReveal 副官カードが出されたか確認
func (n *Napoleon) checkAdjutantReveal(playerIdx int, card *Card) {
	if n.round.adjutantRevealed || n.round.adjutantCard == nil {
		return
	}
	if card.GetDesign() == n.round.adjutantCard.GetDesign() && card.GetValue() == n.round.adjutantCard.GetValue() {
		n.round.adjutantRevealed = true
		n.players[playerIdx].SetAdjutantRevealed(true)
		// 副官が場札交換で入れ替わった可能性があるので再設定
		if !n.players[playerIdx].isAdjutant {
			// 前の副官フラグをクリア
			for _, p := range n.players {
				p.isAdjutant = false
			}
			n.round.adjutantIdx = playerIdx
			n.players[playerIdx].SetIsAdjutant(true)
		}
		n.appendLog(playerIdx, "adjutant_reveal", fmt.Sprintf("%s is revealed as the adjutant!", n.playerName(playerIdx)), []*Card{card})
	}
}

// playerHasSuit プレイヤーが特定のスートを持っているか (ジョーカーは除外)
func (n *Napoleon) playerHasSuit(playerIdx int, design int) bool {
	p := n.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == design {
			return true
		}
	}
	return false
}

// checkGameEnd ゲーム終了判定
func (n *Napoleon) checkGameEnd() {
	hasWinner := false
	for i := range NapoleonPlayerCnt {
		if n.players[i].cumulativeScore >= n.config.PointLimit {
			hasWinner = true
			break
		}
	}

	hasLoser := false
	for i := range NapoleonPlayerCnt {
		if n.players[i].cumulativeScore <= -n.config.PointLimit {
			hasLoser = true
			break
		}
	}

	if !hasWinner && !hasLoser {
		return
	}

	n.round.gameEndFlag = true
	n.round.phase = NapoleonPhaseGameEnd

	// 最高スコアのプレイヤーが勝者
	maxScore := n.players[0].cumulativeScore
	winnerIdx := 0
	for i := 1; i < NapoleonPlayerCnt; i++ {
		if n.players[i].cumulativeScore > maxScore {
			maxScore = n.players[i].cumulativeScore
			winnerIdx = i
		}
	}
	n.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", n.playerName(winnerIdx)), nil)
}

// sortAllHands 全プレイヤーの手札をソートする
func (n *Napoleon) sortAllHands() {
	for _, p := range n.players {
		n.sortHand(p)
	}
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (n *Napoleon) sortHand(p *NapoleonPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.Slice(cards, func(i, j int) bool {
		di, dj := cards[i].GetDesign(), cards[j].GetDesign()
		if di != dj {
			return di < dj
		}
		return cards[i].GetValue() < cards[j].GetValue()
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す
func (n *Napoleon) playerName(idx int) string {
	if idx < 0 || idx >= len(n.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if n.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (n *Napoleon) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	n.round.actionLog = append(n.round.actionLog, &ActionLogEntry{
		TurnNumber: len(n.round.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// napoleonCardStr カードの文字列表現 (ジョーカー対応)
func napoleonCardStr(card *Card) string {
	if card.GetDesign() == CardDesignJoker {
		return "Joker"
	}
	return cardStr(card)
}

// playHintReason プレイヒントの理由を判定する
func (n *Napoleon) playHintReason(chosenIdx int) string {
	player := n.players[n.findHumanIdx()]
	card := player.GetCard(chosenIdx)

	if len(n.round.currentTrick) == 0 {
		if n.isPictureCard(card) {
			return "lead_strong"
		}
		return "lead_low"
	}

	leadCard := n.round.currentTrick[0].Card
	if leadCard.GetDesign() != CardDesignJoker && card.GetDesign() == leadCard.GetDesign() {
		return "follow_suit"
	}
	if card.GetDesign() == n.round.trumpSuit {
		return "trump_cut"
	}
	if card.GetDesign() == CardDesignJoker {
		return "play_joker"
	}
	return "discard_low"
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (n *Napoleon) getValidPlayIndices(playerIdx int) []int {
	player := n.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return n.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- CPU AI ---

// cpuSelectBid CPUがビッドを選択する
func (n *Napoleon) cpuSelectBid(playerIdx int) int {
	switch n.config.CpuDifficulty {
	case NapoleonCpuDifficultyHard:
		return n.cpuBidHard(playerIdx)
	case NapoleonCpuDifficultyNormal:
		return n.cpuBidNormal(playerIdx)
	default:
		return n.cpuBidEasy(playerIdx)
	}
}

// cpuBidEasy ランダムにビッドまたはパス
func (n *Napoleon) cpuBidEasy(_ int) int {
	// 50%の確率でパス
	if rand.Intn(2) == 0 {
		return 0
	}
	// MinBid〜MinBid+2のランダムビッド
	bid := n.config.MinBid + rand.Intn(3)
	if bid <= n.round.highestBid {
		return 0 // 最高ビッドを超えられなければパス
	}
	if bid > NapoleonMaxPictureCards {
		bid = NapoleonMaxPictureCards
	}
	return bid
}

// cpuBidNormal カードの強さに基づくビッド
func (n *Napoleon) cpuBidNormal(playerIdx int) int {
	player := n.players[playerIdx]
	strength := 0

	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if n.isPictureCard(card) {
			strength++
		}
		if card.GetDesign() == CardDesignJoker {
			strength++ // ジョーカーは追加ボーナス
		}
	}

	// 手札の絵札数に基づいてビッド値を決定
	bid := n.config.MinBid + (strength - 4)
	if bid < n.config.MinBid {
		return 0 // 弱ければパス
	}
	if bid <= n.round.highestBid {
		return 0
	}
	if bid > NapoleonMaxPictureCards {
		bid = NapoleonMaxPictureCards
	}
	return bid
}

// cpuBidHard 戦略的なビッド
func (n *Napoleon) cpuBidHard(playerIdx int) int {
	player := n.players[playerIdx]
	strength := 0

	// 各スートの枚数をカウント
	suitCounts := map[int]int{}
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		suitCounts[card.GetDesign()]++

		if card.GetDesign() == CardDesignJoker {
			strength += 3 // ジョーカーは非常に強い
			continue
		}
		// A, K はトリックを取りやすい
		if card.GetValue() == 1 {
			strength += 2
		} else if card.GetValue() == 13 {
			strength += 2
		} else if card.GetValue() == 12 {
			strength++
		} else if card.GetValue() == 11 {
			if rand.Intn(2) == 0 {
				strength++
			}
		}
	}

	// 最も枚数の多いスートを切り札にした場合のボーナス
	maxSuitCount := 0
	for suit, cnt := range suitCounts {
		if suit != CardDesignJoker && cnt > maxSuitCount {
			maxSuitCount = cnt
		}
	}
	if maxSuitCount >= 5 {
		strength++
	}

	bid := n.config.MinBid + (strength - 5)
	if bid < n.config.MinBid {
		return 0
	}
	if bid <= n.round.highestBid {
		// 半分の確率でもう1上げる
		if n.round.highestBid+1 <= NapoleonMaxPictureCards && rand.Intn(2) == 0 && strength >= 6 {
			return n.round.highestBid + 1
		}
		return 0
	}
	if bid > NapoleonMaxPictureCards {
		bid = NapoleonMaxPictureCards
	}
	return bid
}

// cpuSelectTrump CPUが切り札と副官を選択する
func (n *Napoleon) cpuSelectTrump(playerIdx int) (int, int, int) {
	switch n.config.CpuDifficulty {
	case NapoleonCpuDifficultyHard:
		return n.cpuSelectTrumpHard(playerIdx)
	case NapoleonCpuDifficultyNormal:
		return n.cpuSelectTrumpNormal(playerIdx)
	default:
		return n.cpuSelectTrumpEasy(playerIdx)
	}
}

// cpuSelectTrumpEasy ランダムに切り札と副官を選択
func (n *Napoleon) cpuSelectTrumpEasy(_ int) (int, int, int) {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	trumpSuit := suits[rand.Intn(len(suits))]
	// 副官: 持っていないAを選ぶ（なければランダム）
	adjSuit := suits[rand.Intn(len(suits))]
	return trumpSuit, adjSuit, 1 // A
}

// cpuSelectTrumpNormal 手札に基づいて切り札と副官を選択
func (n *Napoleon) cpuSelectTrumpNormal(playerIdx int) (int, int, int) {
	player := n.players[playerIdx]

	// 最も枚数の多いスートを切り札に
	suitCounts := map[int]int{}
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() != CardDesignJoker {
			suitCounts[card.GetDesign()]++
		}
	}
	trumpSuit := CardDesignSpade
	maxCount := 0
	for suit, cnt := range suitCounts {
		if cnt > maxCount {
			maxCount = cnt
			trumpSuit = suit
		}
	}

	// 副官: 持っていない最強カードを選ぶ
	adjSuit, adjVal := n.selectAdjutantCard(playerIdx, trumpSuit)
	return trumpSuit, adjSuit, adjVal
}

// cpuSelectTrumpHard 戦略的に切り札と副官を選択
func (n *Napoleon) cpuSelectTrumpHard(playerIdx int) (int, int, int) {
	player := n.players[playerIdx]

	// 各スートのスコアを計算
	suitScores := map[int]int{}
	suitCounts := map[int]int{}
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == CardDesignJoker {
			continue
		}
		suitCounts[card.GetDesign()]++
		if card.GetValue() == 1 || card.GetValue() >= 11 {
			suitScores[card.GetDesign()] += 2
		} else if card.GetValue() >= 9 {
			suitScores[card.GetDesign()]++
		}
	}

	// 枚数+高札スコアが最大のスートを選択
	trumpSuit := CardDesignSpade
	bestScore := -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		score := suitCounts[suit]*2 + suitScores[suit]
		if score > bestScore {
			bestScore = score
			trumpSuit = suit
		}
	}

	adjSuit, adjVal := n.selectAdjutantCard(playerIdx, trumpSuit)
	return trumpSuit, adjSuit, adjVal
}

// selectAdjutantCard 副官カードを選択する (持っていない強カードを優先)
func (n *Napoleon) selectAdjutantCard(playerIdx int, trumpSuit int) (int, int) {
	player := n.players[playerIdx]
	handSet := map[[2]int]bool{}
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		handSet[[2]int{c.GetDesign(), c.GetValue()}] = true
	}

	// 優先順位: 切り札Aなし → 他スートAなし → 切り札Kなし → ...
	priorities := [][2]int{
		{trumpSuit, 1},  // 切り札のA
		{trumpSuit, 13}, // 切り札のK
	}
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if suit != trumpSuit {
			priorities = append(priorities, [2]int{suit, 1})  // 他スートA
			priorities = append(priorities, [2]int{suit, 13}) // 他スートK
		}
	}

	for _, p := range priorities {
		if !handSet[p] {
			return p[0], p[1]
		}
	}

	// 全部持っている場合、ジョーカーを指名
	if !handSet[[2]int{CardDesignJoker, 1}] {
		return CardDesignJoker, 1
	}

	// それでもなければ切り札Q
	return trumpSuit, 12
}

// cpuSelectDiscard CPUが捨てるカードを選択する
func (n *Napoleon) cpuSelectDiscard(playerIdx int) int {
	switch n.config.CpuDifficulty {
	case NapoleonCpuDifficultyHard:
		return n.cpuSelectDiscardHard(playerIdx)
	case NapoleonCpuDifficultyNormal:
		return n.cpuSelectDiscardNormal(playerIdx)
	default:
		return n.cpuSelectDiscardEasy(playerIdx)
	}
}

// cpuSelectDiscardEasy ランダムに捨てるカードを選択
func (n *Napoleon) cpuSelectDiscardEasy(playerIdx int) int {
	player := n.players[playerIdx]
	// 絵札以外からランダム
	var nonPicture []int
	for i := 0; i < player.GetCardsSize(); i++ {
		if !n.isPictureCard(player.GetCard(i)) {
			nonPicture = append(nonPicture, i)
		}
	}
	if len(nonPicture) > 0 {
		return nonPicture[rand.Intn(len(nonPicture))]
	}
	return rand.Intn(player.GetCardsSize())
}

// cpuSelectDiscardNormal 弱いカードを優先的に捨てる
func (n *Napoleon) cpuSelectDiscardNormal(playerIdx int) int {
	player := n.players[playerIdx]
	bestIdx := 0
	bestScore := 100

	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		score := n.cardStrength(card)

		// 絵札は捨てたくない
		if n.isPictureCard(card) {
			score += 50
		}
		// 切り札は捨てたくない
		if card.GetDesign() == n.round.trumpSuit {
			score += 30
		}
		// ジョーカーは絶対捨てない
		if card.GetDesign() == CardDesignJoker {
			score += 100
		}

		if score < bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	return bestIdx
}

// cpuSelectDiscardHard 戦略的に捨てるカードを選択
func (n *Napoleon) cpuSelectDiscardHard(playerIdx int) int {
	player := n.players[playerIdx]
	bestIdx := 0
	bestScore := 1000

	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		score := n.cardStrength(card)

		if n.isPictureCard(card) {
			score += 50
		}
		if card.GetDesign() == n.round.trumpSuit {
			score += 30
		}
		if card.GetDesign() == CardDesignJoker {
			score += 100
		}
		// 副官カードは絶対捨てない
		if n.round.adjutantCard != nil && card.GetDesign() == n.round.adjutantCard.GetDesign() && card.GetValue() == n.round.adjutantCard.GetValue() {
			score += 200
		}

		if score < bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	return bestIdx
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (n *Napoleon) cpuSelectPlayCard(playerIdx int) int {
	validIndices := n.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch n.config.CpuDifficulty {
	case NapoleonCpuDifficultyHard:
		return n.cpuPlayHard(playerIdx, validIndices)
	case NapoleonCpuDifficultyNormal:
		return n.cpuPlayNormal(playerIdx, validIndices)
	default:
		return n.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (n *Napoleon) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal ヒューリスティックでカードを選択
func (n *Napoleon) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := n.players[playerIdx]
	isNapoleonTeam := n.players[playerIdx].isNapoleon || n.players[playerIdx].isAdjutant

	if len(n.round.currentTrick) == 0 {
		// リード
		if isNapoleonTeam {
			// ナポレオン軍: 高いカード/絵札でリード
			bestIdx := validIndices[0]
			bestScore := -1
			for _, idx := range validIndices {
				card := player.GetCard(idx)
				score := n.cardStrength(card)
				if n.isPictureCard(card) {
					score += 20
				}
				if card.GetDesign() == n.round.trumpSuit {
					score += 10
				}
				if score > bestScore {
					bestScore = score
					bestIdx = idx
				}
			}
			return bestIdx
		}
		// 連合軍: 低いカードでリード
		bestIdx := validIndices[0]
		bestVal := n.cardStrength(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			val := n.cardStrength(player.GetCard(idx))
			if val < bestVal {
				bestVal = val
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// フォロー
	leadCard := n.round.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	if leadSuit == CardDesignJoker {
		// ジョーカーリード: 低いカードを出す
		bestIdx := validIndices[0]
		bestVal := n.cardStrength(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			val := n.cardStrength(player.GetCard(idx))
			if val < bestVal {
				bestVal = val
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// トリック内の絵札があるかチェック
	hasPictureInTrick := false
	for _, tc := range n.round.currentTrick {
		if n.isPictureCard(tc.Card) {
			hasPictureInTrick = true
			break
		}
	}

	if isNapoleonTeam && hasPictureInTrick {
		// ナポレオン軍: 絵札のあるトリックを取りに行く
		return n.cpuTryWinTrick(playerIdx, validIndices)
	}

	if !isNapoleonTeam && hasPictureInTrick {
		// 連合軍: 絵札のあるトリックを奪いに行く (ナポレオン軍に取らせない)
		return n.cpuTryWinTrick(playerIdx, validIndices)
	}

	// 低いカードを出す
	bestIdx := validIndices[0]
	bestVal := n.cardStrength(player.GetCard(validIndices[0]))
	for _, idx := range validIndices[1:] {
		val := n.cardStrength(player.GetCard(idx))
		if val < bestVal {
			bestVal = val
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuPlayHard 高度な戦略プレイ
func (n *Napoleon) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := n.players[playerIdx]
	isNapoleonTeam := n.players[playerIdx].isNapoleon || n.players[playerIdx].isAdjutant

	if len(n.round.currentTrick) == 0 {
		// リード
		if isNapoleonTeam {
			// 切り札でリードして絵札を回収
			bestIdx := validIndices[0]
			bestScore := -1
			for _, idx := range validIndices {
				card := player.GetCard(idx)
				score := n.cardStrength(card)
				if card.GetDesign() == n.round.trumpSuit {
					score += 100
				}
				if card.GetDesign() == CardDesignJoker {
					score += 200
				}
				if score > bestScore {
					bestScore = score
					bestIdx = idx
				}
			}
			return bestIdx
		}
		// 連合軍: 最も低い非切り札を出す
		bestIdx := validIndices[0]
		bestScore := 1000
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			score := n.cardStrength(card)
			if card.GetDesign() == n.round.trumpSuit {
				score += 100
			}
			if card.GetDesign() == CardDesignJoker {
				score += 200
			}
			if score < bestScore {
				bestScore = score
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// フォロー (絵札のあるトリックを制する戦略)
	hasPictureInTrick := false
	for _, tc := range n.round.currentTrick {
		if n.isPictureCard(tc.Card) {
			hasPictureInTrick = true
			break
		}
	}

	if hasPictureInTrick {
		return n.cpuTryWinTrick(playerIdx, validIndices)
	}

	// 絵札なしのトリック: 不要カードを捨てる
	bestIdx := validIndices[0]
	bestScore := 1000
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		score := n.cardStrength(card)
		if n.isPictureCard(card) {
			score += 50
		}
		if card.GetDesign() == n.round.trumpSuit {
			score += 30
		}
		if card.GetDesign() == CardDesignJoker {
			score += 200
		}
		if score < bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuTryWinTrick トリックを勝とうとする
func (n *Napoleon) cpuTryWinTrick(playerIdx int, validIndices []int) int {
	player := n.players[playerIdx]

	// 現在のトリック勝者を判定するために仮のトリックを見る
	leadCard := n.round.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	if leadSuit == CardDesignJoker {
		// ジョーカーリード: 低いカードを出す（勝てない）
		bestIdx := validIndices[0]
		bestVal := n.cardStrength(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			val := n.cardStrength(player.GetCard(idx))
			if val < bestVal {
				bestVal = val
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// 勝てるカードで最小のものを探す
	type candidate struct {
		idx   int
		score int
	}
	var winners []candidate

	for _, idx := range validIndices {
		card := player.GetCard(idx)
		// このカードで勝てるかシミュレート
		if n.wouldWinTrick(card) {
			winners = append(winners, candidate{idx, n.cardStrength(card)})
		}
	}

	if len(winners) > 0 {
		// 最小の勝てるカード
		best := winners[0]
		for _, w := range winners[1:] {
			if w.score < best.score {
				best = w
			}
		}
		return best.idx
	}

	// 勝てない: 最も低いカードを出す
	bestIdx := validIndices[0]
	bestVal := n.cardStrength(player.GetCard(validIndices[0]))
	for _, idx := range validIndices[1:] {
		val := n.cardStrength(player.GetCard(idx))
		if val < bestVal {
			bestVal = val
			bestIdx = idx
		}
	}
	return bestIdx
}

// wouldWinTrick このカードで現在のトリックに勝てるか判定
func (n *Napoleon) wouldWinTrick(card *Card) bool {
	if card.GetDesign() == CardDesignJoker {
		// ジョーカーは勝てるが、スペ3がすでに出ている場合は負ける
		for _, tc := range n.round.currentTrick {
			if tc.Card.GetDesign() == CardDesignSpade && tc.Card.GetValue() == 3 {
				return false
			}
		}
		return true
	}

	// スペ3でジョーカーに勝てるか
	if card.GetDesign() == CardDesignSpade && card.GetValue() == 3 {
		for _, tc := range n.round.currentTrick {
			if tc.Card.GetDesign() == CardDesignJoker {
				return true
			}
		}
	}

	leadSuit := n.round.currentTrick[0].Card.GetDesign()
	if leadSuit == CardDesignJoker {
		return false // ジョーカーリードには勝てない
	}

	// 現在の最強カードを特定
	currentWinnerIsTrump := false
	currentWinnerValue := 0
	for _, tc := range n.round.currentTrick {
		if tc.Card.GetDesign() == CardDesignJoker {
			return false // ジョーカーがいれば勝てない
		}
		isTrump := tc.Card.GetDesign() == n.round.trumpSuit
		val := n.cardStrength(tc.Card)
		if isTrump && !currentWinnerIsTrump {
			currentWinnerIsTrump = true
			currentWinnerValue = val
		} else if isTrump && val > currentWinnerValue {
			currentWinnerValue = val
		} else if !isTrump && !currentWinnerIsTrump && tc.Card.GetDesign() == leadSuit && val > currentWinnerValue {
			currentWinnerValue = val
		}
	}

	isTrump := card.GetDesign() == n.round.trumpSuit
	val := n.cardStrength(card)

	if isTrump && !currentWinnerIsTrump {
		return true
	}
	if isTrump && currentWinnerIsTrump && val > currentWinnerValue {
		return true
	}
	if !isTrump && !currentWinnerIsTrump && card.GetDesign() == leadSuit && val > currentWinnerValue {
		return true
	}
	return false
}

// --- JSON Serialization ---

// napoleonTrickCardJSON is the JSON wire format for NapoleonTrickCard.
type napoleonTrickCardJSON struct {
	PlayerIdx int   `json:"pi"`
	Card      *Card `json:"cd"`
}

// MarshalJSON implements json.Marshaler.
func (tc *NapoleonTrickCard) MarshalJSON() ([]byte, error) {
	return json.Marshal(napoleonTrickCardJSON{
		PlayerIdx: tc.PlayerIdx,
		Card:      tc.Card,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (tc *NapoleonTrickCard) UnmarshalJSON(data []byte) error {
	var j napoleonTrickCardJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	tc.PlayerIdx = j.PlayerIdx
	tc.Card = j.Card
	return nil
}

// napoleonJSON is the JSON wire format for Napoleon (flattens napoleonRoundState).
type napoleonJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*NapoleonPlayer    `json:"pl"`
	Config           NapoleonConfig       `json:"cf"`
	Phase            NapoleonPhase        `json:"ph"`
	RoundNumber      int                  `json:"rn"`
	TrickNumber      int                  `json:"tn"`
	CurrentPlayerIdx int                  `json:"ci"`
	CurrentTrick     []*NapoleonTrickCard `json:"ct"`
	TrumpSuit        int                  `json:"ts"`
	AdjutantCard     *Card                `json:"ac"`
	NapoleonIdx      int                  `json:"ni"`
	AdjutantIdx      int                  `json:"ai"`
	AdjutantRevealed bool                 `json:"ar"`
	LeadPlayerIdx    int                  `json:"li"`
	BidPlayerIdx     int                  `json:"bi"`
	Kitty            []*Card              `json:"ki"`
	HighestBid       int                  `json:"hb"`
	HighestBidder    int                  `json:"hd"`
	PassCount        int                  `json:"pc"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerTeam       int                  `json:"wt"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// napoleonMaxSliceLen caps slice sizes during deserialisation.
const napoleonMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (n *Napoleon) MarshalJSON() ([]byte, error) {
	return json.Marshal(napoleonJSON{
		TrumpCards:       n.trumpCards,
		Players:          n.players,
		Config:           n.config,
		Phase:            n.round.phase,
		RoundNumber:      n.round.roundNumber,
		TrickNumber:      n.round.trickNumber,
		CurrentPlayerIdx: n.round.currentPlayerIdx,
		CurrentTrick:     n.round.currentTrick,
		TrumpSuit:        n.round.trumpSuit,
		AdjutantCard:     n.round.adjutantCard,
		NapoleonIdx:      n.round.napoleonIdx,
		AdjutantIdx:      n.round.adjutantIdx,
		AdjutantRevealed: n.round.adjutantRevealed,
		LeadPlayerIdx:    n.round.leadPlayerIdx,
		BidPlayerIdx:     n.round.bidPlayerIdx,
		Kitty:            n.round.kitty,
		HighestBid:       n.round.highestBid,
		HighestBidder:    n.round.highestBidder,
		PassCount:        n.round.passCount,
		GameEndFlag:      n.round.gameEndFlag,
		WinnerTeam:       n.round.winnerTeam,
		ActionLog:        n.round.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *Napoleon) UnmarshalJSON(data []byte) error {
	var j napoleonJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > napoleonMaxSliceLen || len(j.CurrentTrick) > napoleonMaxSliceLen ||
		len(j.Kitty) > napoleonMaxSliceLen || len(j.ActionLog) > napoleonMaxSliceLen {
		return fmt.Errorf("napoleon: input array exceeds maximum allowed size")
	}
	n.trumpCards = j.TrumpCards
	if n.trumpCards == nil {
		n.trumpCards = NewTrumpCards(0)
	}
	n.players = j.Players
	if n.players == nil {
		n.players = make([]*NapoleonPlayer, 0)
	}
	n.config = j.Config
	n.round = napoleonRoundState{
		phase:            j.Phase,
		roundNumber:      j.RoundNumber,
		trickNumber:      j.TrickNumber,
		currentPlayerIdx: j.CurrentPlayerIdx,
		currentTrick:     j.CurrentTrick,
		trumpSuit:        j.TrumpSuit,
		adjutantCard:     j.AdjutantCard,
		napoleonIdx:      j.NapoleonIdx,
		adjutantIdx:      j.AdjutantIdx,
		adjutantRevealed: j.AdjutantRevealed,
		leadPlayerIdx:    j.LeadPlayerIdx,
		bidPlayerIdx:     j.BidPlayerIdx,
		kitty:            j.Kitty,
		highestBid:       j.HighestBid,
		highestBidder:    j.HighestBidder,
		passCount:        j.PassCount,
		gameEndFlag:      j.GameEndFlag,
		winnerTeam:       j.WinnerTeam,
		actionLog:        j.ActionLog,
	}
	if n.round.actionLog == nil {
		n.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
