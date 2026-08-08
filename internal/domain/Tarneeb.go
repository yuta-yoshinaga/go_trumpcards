//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// TarneebHandSize 各プレイヤーの手札枚数
const TarneebHandSize = 13

// TarneebMaxBid 最大ビッド (デッキ枚数の上限)
const TarneebMaxBid = 13

// TarneebTrumpUndeclared トランプ未宣言を示すセンチネル値
const TarneebTrumpUndeclared = 0

// TarneebPhase ゲームフェーズ
type TarneebPhase int

// Tarneebのフェーズ定数
const (
	// TarneebPhaseBid ビッドフェーズ (各プレイヤーが7-13またはパス)
	TarneebPhaseBid TarneebPhase = 0
	// TarneebPhaseTrumpDeclaration トランプ宣言フェーズ (ビッド勝者がスートを選ぶ)
	TarneebPhaseTrumpDeclaration TarneebPhase = 1
	// TarneebPhasePlay トリックプレイフェーズ
	TarneebPhasePlay TarneebPhase = 2
	// TarneebPhaseTrickEnd トリック終了フェーズ
	TarneebPhaseTrickEnd TarneebPhase = 3
	// TarneebPhaseRoundEnd ラウンド終了フェーズ
	TarneebPhaseRoundEnd TarneebPhase = 4
	// TarneebPhaseGameEnd ゲーム終了フェーズ
	TarneebPhaseGameEnd TarneebPhase = 5
)

// TarneebHint ヒント情報
type TarneebHint struct {
	CardIndex *int   // 推奨カードインデックス (ビッド/トランプ宣言時 nil)
	Bid       *int   // 推奨ビッド値 (プレイ/トランプ宣言時 nil)
	TrumpSuit *int   // 推奨トランプスート (ビッド/プレイ時 nil)
	Reason    string // ヒント理由キー
}

// Tarneeb Tarneebゲームクラス
//
// ルール概要:
//   - 4人2チーム制 (idx 0+2 vs 1+3)、52枚デッキ、各13枚配布。
//   - ビッドはディーラーの左隣から1人1回。7〜13の数字またはパス(0)。
//     2人目以降は現在の最高ビッドより厳密に大きい値のみ。全員パスなら再配布。
//   - 最高ビッド者がトランプスートを宣言した後、その者がリードでプレイ開始。
//   - プレイは「リードスート必従、ボイドなら自由」 (Spadesと違いトランプブレイク制限なし)。
//   - スコア: ビッドチームのトリック合計 >= ビッドで +トリック数、不足で -ビッド数。
//     防衛チームは常に +トリック数。先に PointLimit (デフォルト31) に到達したチームの勝ち。
type Tarneeb struct {
	trumpCards       *TrumpCards
	players          []*TarneebPlayer
	config           TarneebConfig
	phase            TarneebPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	trumpSuit        int // 0 = 未宣言、それ以外は CardDesign 値
	bidPlayerIdx     int // 次にビッドするプレイヤー (ビッドフェーズ中)
	bidWinnerIdx     int // 最高ビッド者 (-1 = 未確定)
	highestBid       int // 現在の最高ビッド値
	redealCount      int // 全員パスによる再配布回数 (このラウンドのみ)
	leadPlayerIdx    int
	dealerIdx        int
	teamScores       [TarneebTeamCnt]int
	gameEndFlag      bool
	winnerTeam       int
	actionLogBase
}

// NewTarneeb コンストラクタ
func NewTarneeb(trumpCards *TrumpCards, players []*TarneebPlayer, config TarneebConfig) *Tarneeb {
	return &Tarneeb{
		trumpCards:   trumpCards,
		players:      players,
		config:       config,
		bidWinnerIdx: -1,
		winnerTeam:   -1,
		trumpSuit:    TarneebTrumpUndeclared,
	}
}

// NewDefaultTarneeb は4人 (1人間 + 3 CPU) の標準セットアップを返す。
// CUI / Web / Worker 共通の構築 SSoT。
func NewDefaultTarneeb() *Tarneeb {
	players := []*TarneebPlayer{
		NewTarneebPlayer(true, 0),
		NewTarneebPlayer(false, 1),
		NewTarneebPlayer(false, 0),
		NewTarneebPlayer(false, 1),
	}
	return NewTarneeb(NewTrumpCards(0), players, DefaultTarneebConfig())
}

// Reset ゲーム初期化
func (t *Tarneeb) Reset() {
	t.gameEndFlag = false
	t.winnerTeam = -1
	t.roundNumber = 1
	t.trickNumber = 0
	t.currentTrick = nil
	t.leadPlayerIdx = -1
	t.currentPlayerIdx = -1
	t.dealerIdx = 0
	t.teamScores = [TarneebTeamCnt]int{}
	t.actionLog = nil
	t.redealCount = 0

	for _, p := range t.players {
		p.ResetRound()
	}

	t.startBidRound()
}

// NextRound 次のラウンドを開始する
func (t *Tarneeb) NextRound() {
	if t.phase != TarneebPhaseRoundEnd {
		return
	}

	t.roundNumber++
	t.trickNumber = 0
	t.currentTrick = nil
	t.leadPlayerIdx = -1
	t.currentPlayerIdx = -1
	t.dealerIdx = (t.dealerIdx + 1) % TarneebPlayerCnt
	t.redealCount = 0

	for _, p := range t.players {
		p.ResetRound()
	}

	t.startBidRound()
}

// startBidRound デッキを切ってビッドフェーズを開始する。
func (t *Tarneeb) startBidRound() {
	t.bidWinnerIdx = -1
	t.highestBid = 0
	t.trumpSuit = TarneebTrumpUndeclared

	t.trumpCards.Shuffle()
	dealAllCards(t.trumpCards, t.players)
	t.sortAllHands()

	t.bidPlayerIdx = (t.dealerIdx + 1) % TarneebPlayerCnt
	t.phase = TarneebPhaseBid
}

// PlayerBid 人間プレイヤーがビッドする (0 = パス、7-13 = ビッド)
func (t *Tarneeb) PlayerBid(bid int) error {
	if t.gameEndFlag {
		return ErrGameEnded
	}
	if t.phase != TarneebPhaseBid {
		return ErrWrongPhase
	}
	if t.bidPlayerIdx < 0 || t.bidPlayerIdx >= TarneebPlayerCnt {
		return ErrWrongPhase
	}
	if !t.players[t.bidPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if err := t.validateBid(bid); err != nil {
		return err
	}
	t.applyBid(t.bidPlayerIdx, bid)
	return nil
}

// CpuBid 現在のビッド手番がCPUの場合にビッドする
func (t *Tarneeb) CpuBid() {
	if t.gameEndFlag || t.phase != TarneebPhaseBid {
		return
	}
	if t.bidPlayerIdx < 0 || t.bidPlayerIdx >= TarneebPlayerCnt {
		return
	}
	if t.players[t.bidPlayerIdx].GetIsHuman() {
		return
	}
	bid := t.cpuSelectBid(t.bidPlayerIdx)
	t.applyBid(t.bidPlayerIdx, bid)
}

// validateBid ビッド値がルール上有効か検証する。
//   - 0 (パス) は常に許容。
//   - 非パスは [config.MinBid, TarneebMaxBid] の範囲内かつ現在の最高ビッドより厳密に大きい必要がある。
//   - ただし、まだ誰もビッドしていない状態で最後の手番なら最低ビッドのみ可能。
func (t *Tarneeb) validateBid(bid int) error {
	if bid == TarneebPassBid {
		return nil
	}
	if bid < t.config.MinBid || bid > TarneebMaxBid {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは %d〜%d またはパス(0) で指定してください", t.config.MinBid, TarneebMaxBid))
	}
	if bid <= t.highestBid {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは現在の最高 %d より大きくする必要があります", t.highestBid))
	}
	return nil
}

// applyBid ビッドを適用し、必要なら次のフェーズへ遷移する。
func (t *Tarneeb) applyBid(playerIdx, bid int) {
	t.players[playerIdx].SetBid(bid)
	bidLabel := fmt.Sprintf("%d", bid)
	if bid == TarneebPassBid {
		bidLabel = "Pass"
	} else if bid > t.highestBid {
		t.highestBid = bid
		t.bidWinnerIdx = playerIdx
	}
	t.appendLog(playerIdx, "bid", fmt.Sprintf("%s bids %s", t.playerName(playerIdx), bidLabel), nil)

	t.bidPlayerIdx = (t.bidPlayerIdx + 1) % TarneebPlayerCnt
	// 4人ビッドし終えたらディーラーの左隣に戻り、フェーズ遷移を判定する。
	if t.bidPlayerIdx == (t.dealerIdx+1)%TarneebPlayerCnt {
		t.finishBidPhase()
	}
}

// finishBidPhase ビッドラウンドを終え、トランプ宣言フェーズへ進むか再配布する。
func (t *Tarneeb) finishBidPhase() {
	if t.bidWinnerIdx < 0 {
		t.redealCount++
		t.appendLog(-1, "redeal", fmt.Sprintf("All passed (redeal #%d)", t.redealCount), nil)
		for _, p := range t.players {
			p.ResetRound()
		}
		t.startBidRound()
		return
	}
	t.appendLog(t.bidWinnerIdx, "bid_win",
		fmt.Sprintf("%s wins the auction with %d", t.playerName(t.bidWinnerIdx), t.highestBid), nil)
	t.phase = TarneebPhaseTrumpDeclaration
}

// PlayerDeclareTrump 人間プレイヤーがトランプスートを宣言する。
func (t *Tarneeb) PlayerDeclareTrump(suit int) error {
	if t.gameEndFlag {
		return ErrGameEnded
	}
	if t.phase != TarneebPhaseTrumpDeclaration {
		return ErrWrongPhase
	}
	if t.bidWinnerIdx < 0 || !t.players[t.bidWinnerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !isValidSuit(suit) {
		return NewDomainError(ErrInvalidPlay, "トランプスートは ♠/♣/♥/♦ から選んでください")
	}
	t.applyTrumpDeclaration(suit)
	return nil
}

// CpuDeclareTrump 現在のトランプ宣言手番がCPUの場合に宣言する。
func (t *Tarneeb) CpuDeclareTrump() {
	if t.gameEndFlag || t.phase != TarneebPhaseTrumpDeclaration {
		return
	}
	if t.bidWinnerIdx < 0 || t.players[t.bidWinnerIdx].GetIsHuman() {
		return
	}
	suit := t.cpuSelectTrump(t.bidWinnerIdx)
	t.applyTrumpDeclaration(suit)
}

// applyTrumpDeclaration トランプスートを設定し、プレイフェーズへ遷移する。
func (t *Tarneeb) applyTrumpDeclaration(suit int) {
	t.trumpSuit = suit
	t.appendLog(t.bidWinnerIdx, "trump",
		fmt.Sprintf("%s declares %s as trump", t.playerName(t.bidWinnerIdx), suitName(suit)), nil)
	t.leadPlayerIdx = t.bidWinnerIdx
	t.currentPlayerIdx = t.bidWinnerIdx
	t.trickNumber = 1
	t.currentTrick = nil
	t.phase = TarneebPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (t *Tarneeb) PlayerPlay(cardIndex int) error {
	if t.gameEndFlag {
		return ErrGameEnded
	}
	if t.phase != TarneebPhasePlay {
		return ErrWrongPhase
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
func (t *Tarneeb) CpuPlay() {
	if t.gameEndFlag || t.phase != TarneebPhasePlay {
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

// ResolveTrick トリック解決
func (t *Tarneeb) ResolveTrick() {
	if t.phase != TarneebPhaseTrickEnd || len(t.currentTrick) != TarneebPlayerCnt {
		return
	}
	winnerIdx := t.trickWinner()
	trickCards := make([]*Card, len(t.currentTrick))
	for i, tc := range t.currentTrick {
		trickCards[i] = tc.Card
	}
	t.players[winnerIdx].AddTrick(trickCards)
	t.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", t.playerName(winnerIdx), t.trickNumber), trickCards)

	t.leadPlayerIdx = winnerIdx
	if t.trickNumber >= TarneebHandSize {
		t.phase = TarneebPhaseRoundEnd
	}
}

// NextTrick 次のトリック開始
func (t *Tarneeb) NextTrick() {
	if t.phase != TarneebPhaseTrickEnd {
		return
	}
	t.currentTrick = nil
	t.currentPlayerIdx = t.leadPlayerIdx
	t.trickNumber++
	t.phase = TarneebPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う。
func (t *Tarneeb) ScoreRound() {
	if t.phase != TarneebPhaseRoundEnd {
		return
	}
	teamTricks := [TarneebTeamCnt]int{}
	for _, p := range t.players {
		teamTricks[p.GetTeam()] += p.GetTrickCount()
	}
	bidderTeam := t.players[t.bidWinnerIdx].GetTeam()
	defenderTeam := 1 - bidderTeam
	bid := t.highestBid

	var bidderDelta int
	if teamTricks[bidderTeam] >= bid {
		bidderDelta = teamTricks[bidderTeam]
	} else {
		bidderDelta = -bid
	}
	defenderDelta := teamTricks[defenderTeam]

	t.teamScores[bidderTeam] += bidderDelta
	t.teamScores[defenderTeam] += defenderDelta

	t.appendLog(-1, "round_score",
		fmt.Sprintf("Team %d (bidder): bid=%d tricks=%d delta=%+d total=%d",
			bidderTeam, bid, teamTricks[bidderTeam], bidderDelta, t.teamScores[bidderTeam]), nil)
	t.appendLog(-1, "round_score",
		fmt.Sprintf("Team %d (defender): tricks=%d delta=%+d total=%d",
			defenderTeam, teamTricks[defenderTeam], defenderDelta, t.teamScores[defenderTeam]), nil)

	// roundScore はプレイヤー単位の表示用 (チームスコアは teamScores に集約)。
	for _, p := range t.players {
		if p.GetTeam() == bidderTeam {
			p.SetRoundScore(bidderDelta)
		} else {
			p.SetRoundScore(defenderDelta)
		}
		p.CommitRoundScore()
	}

	t.checkGameEnd(bidderTeam)
}

// checkGameEnd ゲーム終了判定。タイブレークではビッドチームが勝つ。
func (t *Tarneeb) checkGameEnd(bidderTeam int) {
	reachedAny := false
	for ti := 0; ti < TarneebTeamCnt; ti++ {
		if t.teamScores[ti] >= t.config.PointLimit {
			reachedAny = true
			break
		}
	}
	if !reachedAny {
		return
	}
	t.gameEndFlag = true
	t.phase = TarneebPhaseGameEnd

	switch {
	case t.teamScores[0] > t.teamScores[1]:
		t.winnerTeam = 0
	case t.teamScores[1] > t.teamScores[0]:
		t.winnerTeam = 1
	default:
		// 同点: ビッドチームが勝つ
		t.winnerTeam = bidderTeam
	}
	t.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game!", t.winnerTeam), nil)
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (t *Tarneeb) GetPhase() TarneebPhase { return t.phase }

// SetPhase フェーズ設定 (テスト用)
func (t *Tarneeb) SetPhase(phase TarneebPhase) { t.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (t *Tarneeb) GetRoundNumber() int { return t.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (t *Tarneeb) SetRoundNumber(n int) { t.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (t *Tarneeb) GetTrickNumber() int { return t.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (t *Tarneeb) SetTrickNumber(n int) { t.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (t *Tarneeb) GetCurrentPlayerIdx() int { return t.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (t *Tarneeb) SetCurrentPlayerIdx(idx int) { t.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (t *Tarneeb) GetCurrentTrick() []*TrickCard { return t.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (t *Tarneeb) SetCurrentTrick(trick []*TrickCard) { t.currentTrick = trick }

// GetTrumpSuit トランプスート取得 (0 = 未宣言)
func (t *Tarneeb) GetTrumpSuit() int { return t.trumpSuit }

// SetTrumpSuit トランプスート設定 (テスト用)
func (t *Tarneeb) SetTrumpSuit(suit int) { t.trumpSuit = suit }

// GetBidPlayerIdx 次のビッド手番取得
func (t *Tarneeb) GetBidPlayerIdx() int { return t.bidPlayerIdx }

// SetBidPlayerIdx 次のビッド手番設定 (テスト用)
func (t *Tarneeb) SetBidPlayerIdx(idx int) { t.bidPlayerIdx = idx }

// GetBidWinnerIdx 最高ビッド者インデックス取得 (-1 = 未確定)
func (t *Tarneeb) GetBidWinnerIdx() int { return t.bidWinnerIdx }

// SetBidWinnerIdx 最高ビッド者インデックス設定 (テスト用)
func (t *Tarneeb) SetBidWinnerIdx(idx int) { t.bidWinnerIdx = idx }

// GetHighestBid 現在の最高ビッド値取得
func (t *Tarneeb) GetHighestBid() int { return t.highestBid }

// SetHighestBid 最高ビッド値設定 (テスト用)
func (t *Tarneeb) SetHighestBid(bid int) { t.highestBid = bid }

// GetRedealCount 当ラウンドの再配布回数取得
func (t *Tarneeb) GetRedealCount() int { return t.redealCount }

// GetDealerIdx ディーラーインデックス取得
func (t *Tarneeb) GetDealerIdx() int { return t.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (t *Tarneeb) SetDealerIdx(idx int) { t.dealerIdx = idx }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (t *Tarneeb) GetLeadPlayerIdx() int { return t.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (t *Tarneeb) SetLeadPlayerIdx(idx int) { t.leadPlayerIdx = idx }

// GetTeamScore チームスコア取得
func (t *Tarneeb) GetTeamScore(team int) int {
	if team < 0 || team >= TarneebTeamCnt {
		return 0
	}
	return t.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (t *Tarneeb) SetTeamScore(team, score int) {
	if team >= 0 && team < TarneebTeamCnt {
		t.teamScores[team] = score
	}
}

// GetGameEndFlag ゲーム終了フラグ取得
func (t *Tarneeb) GetGameEndFlag() bool { return t.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (t *Tarneeb) GetWinnerTeam() int { return t.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (t *Tarneeb) GetPlayerCnt() int { return len(t.players) }

// GetPlayer プレイヤー取得
func (t *Tarneeb) GetPlayer(i int) *TarneebPlayer {
	if i < 0 || i >= len(t.players) {
		return nil
	}
	return t.players[i]
}

// GetConfig 設定取得
func (t *Tarneeb) GetConfig() TarneebConfig { return t.config }

// SetConfig 設定変更
func (t *Tarneeb) SetConfig(cfg TarneebConfig) { t.config = cfg }

// IsHumanTurn 現在の手番が人間かどうか
func (t *Tarneeb) IsHumanTurn() bool {
	if t.currentPlayerIdx < 0 || t.currentPlayerIdx >= len(t.players) {
		return false
	}
	return t.players[t.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (t *Tarneeb) IsHumanBidTurn() bool {
	if t.bidPlayerIdx < 0 || t.bidPlayerIdx >= len(t.players) {
		return false
	}
	return t.players[t.bidPlayerIdx].GetIsHuman()
}

// IsHumanTrumpTurn 現在のトランプ宣言手番が人間かどうか
func (t *Tarneeb) IsHumanTrumpTurn() bool {
	if t.bidWinnerIdx < 0 || t.bidWinnerIdx >= len(t.players) {
		return false
	}
	return t.players[t.bidWinnerIdx].GetIsHuman()
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (t *Tarneeb) GetValidPlayIndices(playerIdx int) []int {
	return t.getValidPlayIndices(playerIdx)
}

// GetHint ヒントを取得する
func (t *Tarneeb) GetHint() *TarneebHint {
	humanIdx := t.findHumanIdx()
	if humanIdx < 0 {
		return nil
	}
	switch t.phase {
	case TarneebPhaseBid:
		if t.bidPlayerIdx != humanIdx {
			return nil
		}
		bid := t.cpuBidHard(humanIdx)
		return &TarneebHint{Bid: &bid, Reason: hintBidReason(bid)}
	case TarneebPhaseTrumpDeclaration:
		if t.bidWinnerIdx != humanIdx {
			return nil
		}
		suit := t.cpuSelectTrumpHard(humanIdx)
		return &TarneebHint{TrumpSuit: &suit, Reason: "trump_longest"}
	case TarneebPhasePlay:
		if t.currentPlayerIdx != humanIdx {
			return nil
		}
		valid := t.getValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := t.cpuPlayHard(humanIdx, valid)
		return &TarneebHint{CardIndex: &idx, Reason: t.playHintReason(humanIdx, idx)}
	}
	return nil
}

// --- Private helpers ---

func hintBidReason(bid int) string {
	if bid == TarneebPassBid {
		return "bid_pass"
	}
	return "bid_estimate"
}

// findHumanIdx 人間プレイヤーのインデックス (-1 = なし)
func (t *Tarneeb) findHumanIdx() int {
	for i, p := range t.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// playCard カードをプレイする共通処理
func (t *Tarneeb) playCard(playerIdx int, card *Card) {
	t.currentTrick = append(t.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	t.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", t.playerName(playerIdx), cardStr(card)), []*Card{card})
	if len(t.currentTrick) == TarneebPlayerCnt {
		t.phase = TarneebPhaseTrickEnd
		return
	}
	t.currentPlayerIdx = (t.currentPlayerIdx + 1) % TarneebPlayerCnt
}

// validatePlay カードのプレイがルール上有効か検証する。
// Tarneeb はリードスート必従のみ。ボイドなら任意 (トランプ or 捨て札)。
func (t *Tarneeb) validatePlay(playerIdx int, card *Card) error {
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
func (t *Tarneeb) playerHasSuit(playerIdx, design int) bool {
	p := t.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// tarneebRank converts a raw `Card.GetValue()` (1-13, where 1 = Ace) to
// Tarneeb's comparison rank where the Ace is the highest card. Used across
// trickWinner, hand sort, and the CPU AI so that every site agrees on the
// canonical ordering A > K > Q > J > 10 > … > 2.
func tarneebRank(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// trickWinner 現在のトリックの勝者を決定する。
//   - 任意のトランプが出ていれば最高トランプの勝ち。
//   - そうでなければリードスート最高の勝ち。
//
// Aces compare as the strongest rank (see tarneebRank).
func (t *Tarneeb) trickWinner() int {
	return ResolveTrickWinner(t.currentTrick, t.trumpSuit, func(cd *Card) int { return tarneebRank(cd.GetValue()) })
}

// sortAllHands 全プレイヤーの手札をソートする
func (t *Tarneeb) sortAllHands() {
	for _, p := range t.players {
		tarneebSortHand(p)
	}
}

// tarneebSortHand プレイヤーの手札をスート→値の順にソートする
func tarneebSortHand(p *TarneebPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		// Aces sort to the right (largest) so the hand display matches Tarneeb's
		// A > K > … ranking.
		return tarneebRank(ci.GetValue()) < tarneebRank(cj.GetValue())
	})
}

// playerName プレイヤー名を返す
func (t *Tarneeb) playerName(idx int) string {
	if idx < 0 || idx >= len(t.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if t.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// isValidSuit 4スートのうちいずれかか
func isValidSuit(suit int) bool {
	return suit >= CardDesignSpade && suit <= CardDesignDiamond
}

// playHintReason プレイヒントの理由キー
func (t *Tarneeb) playHintReason(playerIdx, chosenIdx int) string {
	player := t.players[playerIdx]
	card := player.GetCard(chosenIdx)
	if len(t.currentTrick) == 0 {
		return "lead_strong"
	}
	leadSuit := t.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == t.trumpSuit {
		return "trump_cut"
	}
	return "discard_high"
}

// getValidPlayIndices プレイ可能なカードのインデックスリスト
func (t *Tarneeb) getValidPlayIndices(playerIdx int) []int {
	player := t.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return t.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- CPU AI ---

// cpuSelectBid CPUの難易度に応じてビッド値を選ぶ。
func (t *Tarneeb) cpuSelectBid(playerIdx int) int {
	switch t.config.CpuDifficulty {
	case TarneebCpuDifficultyHard:
		return t.cpuBidHard(playerIdx)
	case TarneebCpuDifficultyNormal:
		return t.cpuBidNormal(playerIdx)
	default:
		return t.cpuBidEasy(playerIdx)
	}
}

// cpuBidEasy ランダムにパス or 最低ビッド。
func (t *Tarneeb) cpuBidEasy(playerIdx int) int {
	// 簡易: 高札が3枚以上ならビッド、それ以外はパス。
	high := t.countHighCards(playerIdx)
	if high < 3 {
		return TarneebPassBid
	}
	candidate := t.config.MinBid + rand.Intn(2)
	return t.adjustToValidBid(candidate)
}

// cpuBidNormal 高札カウントに基づくビッド。
func (t *Tarneeb) cpuBidNormal(playerIdx int) int {
	estimate := t.estimateTricks(playerIdx, false)
	if estimate < t.config.MinBid {
		return TarneebPassBid
	}
	return t.adjustToValidBid(estimate)
}

// cpuBidHard 最長スート + 高札の総合評価。
func (t *Tarneeb) cpuBidHard(playerIdx int) int {
	estimate := t.estimateTricks(playerIdx, true)
	if estimate < t.config.MinBid {
		return TarneebPassBid
	}
	return t.adjustToValidBid(estimate)
}

// adjustToValidBid 候補ビッドを現在の最高ビッドを上回るように切り上げる。
// 上限を超える場合はパスにする。
func (t *Tarneeb) adjustToValidBid(candidate int) int {
	if candidate <= t.highestBid {
		candidate = t.highestBid + 1
	}
	if candidate < t.config.MinBid {
		candidate = t.config.MinBid
	}
	if candidate > TarneebMaxBid {
		return TarneebPassBid
	}
	return candidate
}

// countHighCards Q以上 (12,13,1=A) のカード枚数。
func (t *Tarneeb) countHighCards(playerIdx int) int {
	p := t.players[playerIdx]
	cnt := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		v := p.GetCard(i).GetValue()
		if v == 1 || v >= 12 {
			cnt++
		}
	}
	return cnt
}

// estimateTricks ハンドからおおよそ取れるトリック数を見積もる。
// useLength=true なら長スート補正を加算する。
func (t *Tarneeb) estimateTricks(playerIdx int, useLength bool) int {
	p := t.players[playerIdx]
	suitCnt := map[int]int{}
	bid := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		suitCnt[c.GetDesign()]++
		switch c.GetValue() {
		case 1, 13:
			bid++
		case 12:
			// Q は半確実
			bid++
		}
	}
	if useLength {
		longest := 0
		for _, n := range suitCnt {
			if n > longest {
				longest = n
			}
		}
		if longest >= 5 {
			bid += longest - 4
		}
	}
	return bid
}

// cpuSelectTrump CPUがトランプを選ぶ。
func (t *Tarneeb) cpuSelectTrump(playerIdx int) int {
	switch t.config.CpuDifficulty {
	case TarneebCpuDifficultyHard:
		return t.cpuSelectTrumpHard(playerIdx)
	case TarneebCpuDifficultyNormal:
		return t.cpuSelectTrumpNormal(playerIdx)
	default:
		return t.cpuSelectTrumpEasy(playerIdx)
	}
}

// cpuSelectTrumpEasy ランダムに4スートのいずれかを返す。
func (t *Tarneeb) cpuSelectTrumpEasy(_ int) int {
	return CardDesignSpade + rand.Intn(4)
}

// cpuSelectTrumpNormal 最長スートを選ぶ。同数なら最初に見つけたスート。
func (t *Tarneeb) cpuSelectTrumpNormal(playerIdx int) int {
	counts := t.suitCounts(playerIdx)
	best := CardDesignSpade
	bestCnt := counts[CardDesignSpade]
	for _, s := range []int{CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if counts[s] > bestCnt {
			bestCnt = counts[s]
			best = s
		}
	}
	return best
}

// cpuSelectTrumpHard 長さ + 高札 (A,K,Q) の総合点で選ぶ。
func (t *Tarneeb) cpuSelectTrumpHard(playerIdx int) int {
	p := t.players[playerIdx]
	score := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		score[c.GetDesign()]++
		switch c.GetValue() {
		case 1, 13:
			score[c.GetDesign()] += 3
		case 12:
			score[c.GetDesign()] += 2
		case 11:
			score[c.GetDesign()]++
		}
	}
	best := CardDesignSpade
	bestScore := score[CardDesignSpade]
	for _, s := range []int{CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if score[s] > bestScore {
			bestScore = score[s]
			best = s
		}
	}
	return best
}

// suitCounts プレイヤーの手札のスート別枚数
func (t *Tarneeb) suitCounts(playerIdx int) map[int]int {
	p := t.players[playerIdx]
	counts := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		counts[p.GetCard(i).GetDesign()]++
	}
	return counts
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選ぶ。
func (t *Tarneeb) cpuSelectPlayCard(playerIdx int) int {
	valid := t.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	switch t.config.CpuDifficulty {
	case TarneebCpuDifficultyHard:
		return t.cpuPlayHard(playerIdx, valid)
	case TarneebCpuDifficultyNormal:
		return t.cpuPlayNormal(playerIdx, valid)
	default:
		return t.cpuPlayEasy(valid)
	}
}

// cpuPlayEasy ランダム。
func (t *Tarneeb) cpuPlayEasy(valid []int) int {
	return valid[rand.Intn(len(valid))]
}

// cpuPlayNormal リードでは最高値、フォローでは「勝てる最小」または「最も低い」。
func (t *Tarneeb) cpuPlayNormal(playerIdx int, valid []int) int {
	p := t.players[playerIdx]
	if len(t.currentTrick) == 0 {
		return t.pickHighest(p, valid)
	}
	leadSuit := t.currentTrick[0].Card.GetDesign()
	maxLead, _, maxTrump, hasTrumpInTrick := t.summariseTrick(leadSuit)
	leadIdxs := t.filterByDesign(p, valid, leadSuit)
	if len(leadIdxs) > 0 {
		// リードスート所持: トランプが出ていなければリード最高値を超える最小、出ていれば最小値で逃げる。
		if !hasTrumpInTrick {
			over := t.filterAbove(p, leadIdxs, maxLead)
			if len(over) > 0 {
				return t.pickLowest(p, over)
			}
		}
		return t.pickLowest(p, leadIdxs)
	}
	// ボイド: 勝てるトランプがあれば最小トランプを切る、なければ最低カードを捨てる。
	trumpIdxs := t.filterByDesign(p, valid, t.trumpSuit)
	if len(trumpIdxs) > 0 {
		if hasTrumpInTrick {
			over := t.filterAbove(p, trumpIdxs, maxTrump)
			if len(over) > 0 {
				return t.pickLowest(p, over)
			}
		} else {
			return t.pickLowest(p, trumpIdxs)
		}
	}
	return t.pickLowest(p, valid)
}

// cpuPlayHard パートナー考慮 + 切り札保全の高度な戦略。
func (t *Tarneeb) cpuPlayHard(playerIdx int, valid []int) int {
	p := t.players[playerIdx]
	if len(t.currentTrick) == 0 {
		// リード: A や K で取りに行く。トランプはなるべく温存。
		// tarneebRank() で Ace(=14) を高位として比較する。
		bestIdx := valid[0]
		bestScore := -1
		for _, idx := range valid {
			c := p.GetCard(idx)
			score := tarneebRank(c.GetValue()) * 2
			if c.GetDesign() == t.trumpSuit {
				score -= 30 // 温存
			}
			if score > bestScore {
				bestScore = score
				bestIdx = idx
			}
		}
		return bestIdx
	}
	leadSuit := t.currentTrick[0].Card.GetDesign()
	maxLead, _, maxTrump, hasTrumpInTrick := t.summariseTrick(leadSuit)
	partnerWinning := t.isPartnerCurrentlyWinning(playerIdx)

	leadIdxs := t.filterByDesign(p, valid, leadSuit)
	if len(leadIdxs) > 0 {
		if partnerWinning {
			return t.pickLowest(p, leadIdxs)
		}
		if !hasTrumpInTrick {
			over := t.filterAbove(p, leadIdxs, maxLead)
			if len(over) > 0 {
				return t.pickLowest(p, over)
			}
		}
		return t.pickLowest(p, leadIdxs)
	}
	// ボイド
	if partnerWinning {
		return t.pickLowest(p, valid)
	}
	trumpIdxs := t.filterByDesign(p, valid, t.trumpSuit)
	if len(trumpIdxs) > 0 {
		if hasTrumpInTrick {
			over := t.filterAbove(p, trumpIdxs, maxTrump)
			if len(over) > 0 {
				return t.pickLowest(p, over)
			}
		} else {
			return t.pickLowest(p, trumpIdxs)
		}
	}
	// 捨て札は最高値カード (温存したくない非トランプ)。Ace-high で評価する。
	bestIdx := valid[0]
	bestScore := -1
	for _, idx := range valid {
		c := p.GetCard(idx)
		score := tarneebRank(c.GetValue())
		if c.GetDesign() == t.trumpSuit {
			score -= 100
		}
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// isPartnerCurrentlyWinning パートナーが現在のトリックを勝っているか
func (t *Tarneeb) isPartnerCurrentlyWinning(playerIdx int) bool {
	if len(t.currentTrick) == 0 {
		return false
	}
	winnerIdx := t.trickWinner()
	if winnerIdx == playerIdx {
		return false
	}
	return t.players[winnerIdx].GetTeam() == t.players[playerIdx].GetTeam()
}

// summariseTrick 現トリックの最高リードスート rank、リードスート所持フラグ、最高トランプ rank、トランプ所持フラグを返す。
// max* 値は tarneebRank() を通した「Ace-high」順位なので、A=14 として King(13) より大きい。
func (t *Tarneeb) summariseTrick(leadSuit int) (maxLead int, hasLead bool, maxTrump int, hasTrump bool) {
	for _, tc := range t.currentTrick {
		d := tc.Card.GetDesign()
		r := tarneebRank(tc.Card.GetValue())
		if d == leadSuit {
			hasLead = true
			if r > maxLead {
				maxLead = r
			}
		}
		if d == t.trumpSuit {
			hasTrump = true
			if r > maxTrump {
				maxTrump = r
			}
		}
	}
	return
}

// pickHighest tarneebRank が最大のカードのインデックスを返す。Ace は最強として扱う。
func (t *Tarneeb) pickHighest(p *TarneebPlayer, valid []int) int {
	if p == nil || len(valid) == 0 {
		return 0
	}
	best := valid[0]
	bestR := tarneebRank(p.GetCard(best).GetValue())
	for _, idx := range valid[1:] {
		r := tarneebRank(p.GetCard(idx).GetValue())
		if r > bestR {
			bestR = r
			best = idx
		}
	}
	return best
}

// pickLowest tarneebRank が最小のカードのインデックスを返す。Ace は最強として扱うため、
// 最小に選ばれにくい。
func (t *Tarneeb) pickLowest(p *TarneebPlayer, valid []int) int {
	if p == nil || len(valid) == 0 {
		return 0
	}
	best := valid[0]
	bestR := tarneebRank(p.GetCard(best).GetValue())
	for _, idx := range valid[1:] {
		r := tarneebRank(p.GetCard(idx).GetValue())
		if r < bestR {
			bestR = r
			best = idx
		}
	}
	return best
}

// filterByDesign 指定スートのみのインデックスを返す
func (t *Tarneeb) filterByDesign(p *TarneebPlayer, valid []int, design int) []int {
	out := make([]int, 0, len(valid))
	for _, idx := range valid {
		if p.GetCard(idx).GetDesign() == design {
			out = append(out, idx)
		}
	}
	return out
}

// filterAbove rank が threshold より大きいインデックスのみを返す。Ace-high なので
// A は K(13) より大きい 14 として比較される。`threshold` は呼び出し側で
// `tarneebRank()` を通した値を渡すこと (`summariseTrick` の出力をそのまま使える)。
func (t *Tarneeb) filterAbove(p *TarneebPlayer, valid []int, threshold int) []int {
	out := make([]int, 0, len(valid))
	for _, idx := range valid {
		if tarneebRank(p.GetCard(idx).GetValue()) > threshold {
			out = append(out, idx)
		}
	}
	return out
}

// tarneebJSON is the JSON wire format for Tarneeb.
type tarneebJSON struct {
	TrumpCards       *TrumpCards         `json:"tc"`
	Players          []*TarneebPlayer    `json:"ps"`
	Config           TarneebConfig       `json:"cf"`
	Phase            TarneebPhase        `json:"ph"`
	RoundNumber      int                 `json:"rn"`
	TrickNumber      int                 `json:"tn"`
	CurrentPlayerIdx int                 `json:"ci"`
	CurrentTrick     []*TrickCard        `json:"ct"`
	TrumpSuit        int                 `json:"ts"`
	BidPlayerIdx     int                 `json:"bi"`
	BidWinnerIdx     int                 `json:"bw"`
	HighestBid       int                 `json:"hb"`
	RedealCount      int                 `json:"rd"`
	LeadPlayerIdx    int                 `json:"li"`
	DealerIdx        int                 `json:"di"`
	TeamScores       [TarneebTeamCnt]int `json:"sc"`
	GameEndFlag      bool                `json:"ge"`
	WinnerTeam       int                 `json:"wt"`
	ActionLog        []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (t *Tarneeb) MarshalJSON() ([]byte, error) {
	return json.Marshal(tarneebJSON{
		TrumpCards:       t.trumpCards,
		Players:          t.players,
		Config:           t.config,
		Phase:            t.phase,
		RoundNumber:      t.roundNumber,
		TrickNumber:      t.trickNumber,
		CurrentPlayerIdx: t.currentPlayerIdx,
		CurrentTrick:     t.currentTrick,
		TrumpSuit:        t.trumpSuit,
		BidPlayerIdx:     t.bidPlayerIdx,
		BidWinnerIdx:     t.bidWinnerIdx,
		HighestBid:       t.highestBid,
		RedealCount:      t.redealCount,
		LeadPlayerIdx:    t.leadPlayerIdx,
		DealerIdx:        t.dealerIdx,
		TeamScores:       t.teamScores,
		GameEndFlag:      t.gameEndFlag,
		WinnerTeam:       t.winnerTeam,
		ActionLog:        t.actionLog,
	})
}

// tarneebMaxSliceLen 復元時の上限 (DoS 防止)。
const tarneebMaxSliceLen = 1000

// tarneebMaxActionLogLen ActionLog の復元時上限。
// 1ラウンドあたり ~70 エントリ × 50ラウンド程度の余裕を確保。
const tarneebMaxActionLogLen = 5000

// UnmarshalJSON implements json.Unmarshaler.
func (t *Tarneeb) UnmarshalJSON(data []byte) error {
	var j tarneebJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > tarneebMaxSliceLen || len(j.CurrentTrick) > tarneebMaxSliceLen ||
		len(j.ActionLog) > tarneebMaxActionLogLen {
		return fmt.Errorf("tarneeb: input array exceeds maximum allowed size")
	}
	t.trumpCards = j.TrumpCards
	if t.trumpCards == nil {
		t.trumpCards = NewTrumpCards(0)
	}
	t.players = j.Players
	if t.players == nil {
		t.players = make([]*TarneebPlayer, 0)
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
	t.trumpSuit = j.TrumpSuit
	t.bidPlayerIdx = j.BidPlayerIdx
	t.bidWinnerIdx = j.BidWinnerIdx
	t.highestBid = j.HighestBid
	t.redealCount = j.RedealCount
	t.leadPlayerIdx = j.LeadPlayerIdx
	t.dealerIdx = j.DealerIdx
	t.teamScores = j.TeamScores
	t.gameEndFlag = j.GameEndFlag
	t.winnerTeam = j.WinnerTeam
	t.actionLog = j.ActionLog
	if t.actionLog == nil {
		t.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
