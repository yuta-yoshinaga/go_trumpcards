//go:build !js || !wasm || casino

// Package domain コートピース (Court Piece / Rang / Hokm) のドメインモデル。
//
// Court Piece はパキスタン・イラン発祥の 4 人 2 チーム制トリックテイキング。52 枚デッキを
// 13 枚ずつ配る。ラウンド開始時、まず「呼び手 (Hakim)」に最初の 5 枚だけを配り、呼び手は
// その 5 枚を見て即座に切り札スートを宣言する。宣言後に残りのカードを配り切り、呼び手の
// リードでリードスート必従のトリックを 13 回行う。
//
// 1 ラウンドで 7 トリック以上取ったチームがラウンド勝利 (Sar)。勝利チームは 1 点を獲得し、
// 呼び手の権利を保持する (次ラウンドも同じ呼び手)。敗北した場合は呼び手が次の席へ移る。
// 同一チームがラウンドを連続で取ると「Court」ボーナスとなり 2 点を獲得する。全 13 トリックを
// 総取りした場合 (clean Court) も 2 点。先に PointLimit (デフォルト 7) に到達したチームが勝利。
//
// トリックの強さは Ace-high (A > K > Q > J > 10 > … > 2)。切り札は非切り札より常に強い。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// CourtPieceHandSize 各プレイヤーの手札枚数
const CourtPieceHandSize = 13

// CourtPiecePeekSize 呼び手が切り札宣言前に見る枚数 (最初の配り)
const CourtPiecePeekSize = 5

// CourtPieceTricksToWin ラウンド勝利 (Sar) に必要なトリック数 (過半数)
const CourtPieceTricksToWin = 7

// CourtPieceTrumpUndeclared トランプ未宣言を示すセンチネル値
const CourtPieceTrumpUndeclared = 0

// CourtPiecePhase ゲームフェーズ
type CourtPiecePhase int

// CourtPiece のフェーズ定数
const (
	// CourtPiecePhaseTrumpDeclaration トランプ宣言フェーズ (呼び手が 5 枚を見てスートを選ぶ)
	CourtPiecePhaseTrumpDeclaration CourtPiecePhase = 0
	// CourtPiecePhasePlay トリックプレイフェーズ
	CourtPiecePhasePlay CourtPiecePhase = 1
	// CourtPiecePhaseTrickEnd トリック終了フェーズ
	CourtPiecePhaseTrickEnd CourtPiecePhase = 2
	// CourtPiecePhaseRoundEnd ラウンド終了フェーズ
	CourtPiecePhaseRoundEnd CourtPiecePhase = 3
	// CourtPiecePhaseGameEnd ゲーム終了フェーズ
	CourtPiecePhaseGameEnd CourtPiecePhase = 4
)

// CourtPieceHint ヒント情報
type CourtPieceHint struct {
	CardIndex *int   // 推奨カードインデックス (トランプ宣言時 nil)
	TrumpSuit *int   // 推奨トランプスート (プレイ時 nil)
	Reason    string // ヒント理由キー
}

// CourtPiece Court Piece ゲームクラス
type CourtPiece struct {
	trumpCards       *TrumpCards
	players          []*CourtPiecePlayer
	config           CourtPieceConfig
	phase            CourtPiecePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	trumpSuit        int // 0 = 未宣言、それ以外は CardDesign 値
	callerIdx        int // 呼び手 (Hakim) のインデックス
	leadPlayerIdx    int
	teamScores       [CourtPieceTeamCnt]int
	lastWinnerTeam   int // 直前ラウンドの勝利チーム (-1 = なし)
	consecutiveWins  int // 同一チームの連続ラウンド勝利数
	lastRoundCourt   bool
	gameEndFlag      bool
	winnerTeam       int
	actionLogBase
}

// NewCourtPiece コンストラクタ
func NewCourtPiece(trumpCards *TrumpCards, players []*CourtPiecePlayer, config CourtPieceConfig) *CourtPiece {
	return &CourtPiece{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerTeam:     -1,
		lastWinnerTeam: -1,
		trumpSuit:      CourtPieceTrumpUndeclared,
	}
}

// NewDefaultCourtPiece は4人 (1人間 + 3 CPU) の標準セットアップを返す。
// CUI / Web / Worker 共通の構築 SSoT。
func NewDefaultCourtPiece() *CourtPiece {
	players := []*CourtPiecePlayer{
		NewCourtPiecePlayer(true, 0),
		NewCourtPiecePlayer(false, 1),
		NewCourtPiecePlayer(false, 0),
		NewCourtPiecePlayer(false, 1),
	}
	return NewCourtPiece(NewTrumpCards(0), players, DefaultCourtPieceConfig())
}

// Reset ゲーム初期化
func (c *CourtPiece) Reset() {
	c.gameEndFlag = false
	c.winnerTeam = -1
	c.roundNumber = 1
	c.trickNumber = 0
	c.currentTrick = nil
	c.leadPlayerIdx = -1
	c.currentPlayerIdx = -1
	c.callerIdx = 0
	c.teamScores = [CourtPieceTeamCnt]int{}
	c.lastWinnerTeam = -1
	c.consecutiveWins = 0
	c.lastRoundCourt = false
	c.actionLog = nil

	for _, p := range c.players {
		p.ResetRound()
	}

	c.startRound()
}

// NextRound 次のラウンドを開始する
func (c *CourtPiece) NextRound() {
	if c.phase != CourtPiecePhaseRoundEnd {
		return
	}

	c.roundNumber++
	c.trickNumber = 0
	c.currentTrick = nil
	c.leadPlayerIdx = -1
	c.currentPlayerIdx = -1

	for _, p := range c.players {
		p.ResetRound()
	}

	c.startRound()
}

// startRound デッキを切り、呼び手に 5 枚だけ配ってトランプ宣言フェーズを開始する。
func (c *CourtPiece) startRound() {
	c.trumpSuit = CourtPieceTrumpUndeclared
	c.trumpCards.Shuffle()

	// Stage 1: 呼び手 (Hakim) にのみ最初の CourtPiecePeekSize 枚を配る。
	caller := c.players[c.callerIdx]
	for i := 0; i < CourtPiecePeekSize; i++ {
		card := c.trumpCards.DrawCard()
		if card == nil {
			break
		}
		caller.AddCard(card)
	}
	courtPieceSortHand(caller)

	c.appendLog(c.callerIdx, "deal",
		fmt.Sprintf("%s peeks at the first %d cards", c.playerName(c.callerIdx), CourtPiecePeekSize), nil)
	c.currentPlayerIdx = c.callerIdx
	c.phase = CourtPiecePhaseTrumpDeclaration
}

// dealRemaining 残りのカードを全員が CourtPieceHandSize 枚になるまで配り切る。
func (c *CourtPiece) dealRemaining() {
	for {
		dealt := false
		for i := 0; i < CourtPiecePlayerCnt; i++ {
			p := c.players[(c.callerIdx+i)%CourtPiecePlayerCnt]
			if p.GetCardsSize() >= CourtPieceHandSize {
				continue
			}
			card := c.trumpCards.DrawCard()
			if card == nil {
				return
			}
			p.AddCard(card)
			dealt = true
		}
		if !dealt {
			return
		}
	}
}

// PlayerDeclareTrump 人間プレイヤー (呼び手) がトランプスートを宣言する。
func (c *CourtPiece) PlayerDeclareTrump(suit int) error {
	if c.gameEndFlag {
		return ErrGameEnded
	}
	if c.phase != CourtPiecePhaseTrumpDeclaration {
		return ErrWrongPhase
	}
	if c.callerIdx < 0 || !c.players[c.callerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !isValidSuit(suit) {
		return NewDomainError(ErrInvalidPlay, "トランプスートは ♠/♣/♥/♦ から選んでください")
	}
	c.applyTrumpDeclaration(suit)
	return nil
}

// CpuDeclareTrump 現在のトランプ宣言手番がCPUの場合に宣言する。
func (c *CourtPiece) CpuDeclareTrump() {
	if c.gameEndFlag || c.phase != CourtPiecePhaseTrumpDeclaration {
		return
	}
	if c.callerIdx < 0 || c.players[c.callerIdx].GetIsHuman() {
		return
	}
	suit := c.cpuSelectTrump(c.callerIdx)
	c.applyTrumpDeclaration(suit)
}

// applyTrumpDeclaration トランプスートを設定し、残りを配ってプレイフェーズへ遷移する。
func (c *CourtPiece) applyTrumpDeclaration(suit int) {
	c.trumpSuit = suit
	c.appendLog(c.callerIdx, "trump",
		fmt.Sprintf("%s declares %s as trump", c.playerName(c.callerIdx), suitName(suit)), nil)

	// Stage 2: 残りのカードを配り切る。
	c.dealRemaining()
	c.sortAllHands()

	c.leadPlayerIdx = c.callerIdx
	c.currentPlayerIdx = c.callerIdx
	c.trickNumber = 1
	c.currentTrick = nil
	c.phase = CourtPiecePhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (c *CourtPiece) PlayerPlay(cardIndex int) error {
	if c.gameEndFlag {
		return ErrGameEnded
	}
	if c.phase != CourtPiecePhasePlay {
		return ErrWrongPhase
	}
	if !c.players[c.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := c.players[c.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := c.validatePlay(c.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	c.playCard(c.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (c *CourtPiece) CpuPlay() {
	if c.gameEndFlag || c.phase != CourtPiecePhasePlay {
		return
	}
	if c.players[c.currentPlayerIdx].GetIsHuman() {
		return
	}
	player := c.players[c.currentPlayerIdx]
	cardIdx := c.cpuSelectPlayCard(c.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	c.playCard(c.currentPlayerIdx, played)
}

// ResolveTrick トリック解決
func (c *CourtPiece) ResolveTrick() {
	if c.phase != CourtPiecePhaseTrickEnd || len(c.currentTrick) != CourtPiecePlayerCnt {
		return
	}
	winnerIdx := c.trickWinner()
	trickCards := make([]*Card, len(c.currentTrick))
	for i, tc := range c.currentTrick {
		trickCards[i] = tc.Card
	}
	c.players[winnerIdx].AddTrick(trickCards)
	c.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", c.playerName(winnerIdx), c.trickNumber), trickCards)

	c.leadPlayerIdx = winnerIdx
	if c.trickNumber >= CourtPieceHandSize {
		c.phase = CourtPiecePhaseRoundEnd
	}
}

// NextTrick 次のトリック開始
func (c *CourtPiece) NextTrick() {
	if c.phase != CourtPiecePhaseTrickEnd {
		return
	}
	c.currentTrick = nil
	c.currentPlayerIdx = c.leadPlayerIdx
	c.trickNumber++
	c.phase = CourtPiecePhasePlay
}

// ScoreRound ラウンドのスコアを確定し、Sar / Court 判定とゲーム終了判定を行う。
func (c *CourtPiece) ScoreRound() {
	if c.phase != CourtPiecePhaseRoundEnd {
		return
	}
	teamTricks := [CourtPieceTeamCnt]int{}
	for _, p := range c.players {
		teamTricks[p.GetTeam()] += p.GetTrickCount()
	}
	// 13 トリック (奇数) なので、必ず一方が 7 以上を取る。
	winningTeam := 0
	if teamTricks[1] >= CourtPieceTricksToWin {
		winningTeam = 1
	}

	allThirteen := teamTricks[winningTeam] == CourtPieceHandSize
	if winningTeam == c.lastWinnerTeam {
		c.consecutiveWins++
	} else {
		c.consecutiveWins = 1
	}
	isCourt := allThirteen || c.consecutiveWins >= 2
	delta := 1
	if isCourt {
		delta = 2
	}
	c.teamScores[winningTeam] += delta
	c.lastWinnerTeam = winningTeam
	c.lastRoundCourt = isCourt

	// 呼び手の権利: 勝利チームが保持、敗北なら次席へ移る。
	if c.players[c.callerIdx].GetTeam() != winningTeam {
		c.callerIdx = (c.callerIdx + 1) % CourtPiecePlayerCnt
	}

	label := "Sar"
	if isCourt {
		label = "Court"
	}
	c.appendLog(-1, "round_score",
		fmt.Sprintf("Team %d wins the round (%s, tricks=%d, +%d, total=%d)",
			winningTeam, label, teamTricks[winningTeam], delta, c.teamScores[winningTeam]), nil)

	for _, p := range c.players {
		if p.GetTeam() == winningTeam {
			p.SetRoundScore(delta)
		} else {
			p.SetRoundScore(0)
		}
		p.CommitRoundScore()
	}

	c.checkGameEnd(winningTeam)
}

// checkGameEnd ゲーム終了判定。
func (c *CourtPiece) checkGameEnd(lastWinner int) {
	reached := false
	for ti := 0; ti < CourtPieceTeamCnt; ti++ {
		if c.teamScores[ti] >= c.config.PointLimit {
			reached = true
			break
		}
	}
	if !reached {
		return
	}
	c.gameEndFlag = true
	c.phase = CourtPiecePhaseGameEnd

	switch {
	case c.teamScores[0] > c.teamScores[1]:
		c.winnerTeam = 0
	case c.teamScores[1] > c.teamScores[0]:
		c.winnerTeam = 1
	default:
		c.winnerTeam = lastWinner
	}
	c.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game!", c.winnerTeam), nil)
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (c *CourtPiece) GetPhase() CourtPiecePhase { return c.phase }

// SetPhase フェーズ設定 (テスト用)
func (c *CourtPiece) SetPhase(phase CourtPiecePhase) { c.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (c *CourtPiece) GetRoundNumber() int { return c.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (c *CourtPiece) SetRoundNumber(n int) { c.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (c *CourtPiece) GetTrickNumber() int { return c.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (c *CourtPiece) SetTrickNumber(n int) { c.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (c *CourtPiece) GetCurrentPlayerIdx() int { return c.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (c *CourtPiece) SetCurrentPlayerIdx(idx int) { c.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (c *CourtPiece) GetCurrentTrick() []*TrickCard { return c.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (c *CourtPiece) SetCurrentTrick(trick []*TrickCard) { c.currentTrick = trick }

// GetTrumpSuit トランプスート取得 (0 = 未宣言)
func (c *CourtPiece) GetTrumpSuit() int { return c.trumpSuit }

// SetTrumpSuit トランプスート設定 (テスト用)
func (c *CourtPiece) SetTrumpSuit(suit int) { c.trumpSuit = suit }

// GetCallerIdx 呼び手 (Hakim) インデックス取得
func (c *CourtPiece) GetCallerIdx() int { return c.callerIdx }

// SetCallerIdx 呼び手インデックス設定 (テスト用)
func (c *CourtPiece) SetCallerIdx(idx int) { c.callerIdx = idx }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (c *CourtPiece) GetLeadPlayerIdx() int { return c.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (c *CourtPiece) SetLeadPlayerIdx(idx int) { c.leadPlayerIdx = idx }

// GetConsecutiveWins 連続ラウンド勝利数取得
func (c *CourtPiece) GetConsecutiveWins() int { return c.consecutiveWins }

// GetLastWinnerTeam 直前ラウンドの勝利チーム取得 (-1 = なし)
func (c *CourtPiece) GetLastWinnerTeam() int { return c.lastWinnerTeam }

// IsLastRoundCourt 直前ラウンドが Court ボーナスだったか
func (c *CourtPiece) IsLastRoundCourt() bool { return c.lastRoundCourt }

// GetTeamScore チームスコア取得
func (c *CourtPiece) GetTeamScore(team int) int {
	if team < 0 || team >= CourtPieceTeamCnt {
		return 0
	}
	return c.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (c *CourtPiece) SetTeamScore(team, score int) {
	if team >= 0 && team < CourtPieceTeamCnt {
		c.teamScores[team] = score
	}
}

// GetGameEndFlag ゲーム終了フラグ取得
func (c *CourtPiece) GetGameEndFlag() bool { return c.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (c *CourtPiece) GetWinnerTeam() int { return c.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (c *CourtPiece) GetPlayerCnt() int { return len(c.players) }

// GetPlayer プレイヤー取得
func (c *CourtPiece) GetPlayer(i int) *CourtPiecePlayer {
	if i < 0 || i >= len(c.players) {
		return nil
	}
	return c.players[i]
}

// GetConfig 設定取得
func (c *CourtPiece) GetConfig() CourtPieceConfig { return c.config }

// SetConfig 設定変更
func (c *CourtPiece) SetConfig(cfg CourtPieceConfig) { c.config = cfg }

// IsHumanTurn 現在の手番が人間かどうか
func (c *CourtPiece) IsHumanTurn() bool {
	if c.currentPlayerIdx < 0 || c.currentPlayerIdx >= len(c.players) {
		return false
	}
	return c.players[c.currentPlayerIdx].GetIsHuman()
}

// IsHumanTrumpTurn 現在のトランプ宣言手番が人間かどうか
func (c *CourtPiece) IsHumanTrumpTurn() bool {
	if c.callerIdx < 0 || c.callerIdx >= len(c.players) {
		return false
	}
	return c.players[c.callerIdx].GetIsHuman()
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (c *CourtPiece) GetValidPlayIndices(playerIdx int) []int {
	return c.getValidPlayIndices(playerIdx)
}

// GetHint ヒントを取得する
func (c *CourtPiece) GetHint() *CourtPieceHint {
	humanIdx := findHumanIdx(c.players)
	if humanIdx < 0 {
		return nil
	}
	switch c.phase {
	case CourtPiecePhaseTrumpDeclaration:
		if c.callerIdx != humanIdx {
			return nil
		}
		suit := c.cpuSelectTrumpHard(humanIdx)
		return &CourtPieceHint{TrumpSuit: &suit, Reason: "trump_longest"}
	case CourtPiecePhasePlay:
		if c.currentPlayerIdx != humanIdx {
			return nil
		}
		valid := c.getValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := c.cpuPlayHard(humanIdx, valid)
		return &CourtPieceHint{CardIndex: &idx, Reason: c.playHintReason(humanIdx, idx)}
	}
	return nil
}

// --- Private helpers ---

// playCard カードをプレイする共通処理
func (c *CourtPiece) playCard(playerIdx int, card *Card) {
	c.currentTrick = append(c.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	c.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", c.playerName(playerIdx), cardStr(card)), []*Card{card})
	if len(c.currentTrick) == CourtPiecePlayerCnt {
		c.phase = CourtPiecePhaseTrickEnd
		return
	}
	c.currentPlayerIdx = (c.currentPlayerIdx + 1) % CourtPiecePlayerCnt
}

// validatePlay カードのプレイがルール上有効か検証する。
// Court Piece はリードスート必従のみ。ボイドなら任意 (トランプ or 捨て札)。
func (c *CourtPiece) validatePlay(playerIdx int, card *Card) error {
	if len(c.currentTrick) == 0 {
		return nil
	}
	leadSuit := c.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && c.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// GetPlayableIndices は指定プレイヤーがいま出せる手札の位置を返す。
//
// **判定は validatePlay をそのまま使う。**マストフォローの規則をここで書き直すと、
// 「出せると表示したのにエラーになる」ずれが生まれる。プレイフェーズ以外や範囲外の
// プレイヤーでは空を返す。
func (c *CourtPiece) GetPlayableIndices(playerIdx int) []int {
	if c.phase != CourtPiecePhasePlay || playerIdx < 0 || playerIdx >= len(c.players) {
		return []int{}
	}
	p := c.players[playerIdx]
	out := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		if c.validatePlay(playerIdx, p.GetCard(i)) == nil {
			out = append(out, i)
		}
	}
	return out
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (c *CourtPiece) playerHasSuit(playerIdx, design int) bool {
	p := c.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// courtPieceRank converts a raw `Card.GetValue()` (1-13, where 1 = Ace) to
// Court Piece's comparison rank where the Ace is the highest card.
func courtPieceRank(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// trickWinner 現在のトリックの勝者を決定する。
//   - 任意のトランプが出ていれば最高トランプの勝ち。
//   - そうでなければリードスート最高の勝ち。
func (c *CourtPiece) trickWinner() int {
	return ResolveTrickWinner(c.currentTrick, c.trumpSuit, func(cd *Card) int { return courtPieceRank(cd.GetValue()) })
}

// sortAllHands 全プレイヤーの手札をソートする
func (c *CourtPiece) sortAllHands() {
	for _, p := range c.players {
		courtPieceSortHand(p)
	}
}

// courtPieceSortHand プレイヤーの手札をスート→値の順にソートする
func courtPieceSortHand(p *CourtPiecePlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return courtPieceRank(ci.GetValue()) < courtPieceRank(cj.GetValue())
	})
}

// playerName プレイヤー名を返す
func (c *CourtPiece) playerName(idx int) string {
	if idx < 0 || idx >= len(c.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if c.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// playHintReason プレイヒントの理由キー
func (c *CourtPiece) playHintReason(playerIdx, chosenIdx int) string {
	player := c.players[playerIdx]
	card := player.GetCard(chosenIdx)
	if len(c.currentTrick) == 0 {
		return "lead_strong"
	}
	leadSuit := c.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == c.trumpSuit {
		return "trump_cut"
	}
	return "discard_high"
}

// getValidPlayIndices プレイ可能なカードのインデックスリスト
func (c *CourtPiece) getValidPlayIndices(playerIdx int) []int {
	player := c.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return c.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- CPU AI ---

// cpuSelectTrump CPUがトランプを選ぶ。
func (c *CourtPiece) cpuSelectTrump(playerIdx int) int {
	switch c.config.CpuDifficulty {
	case CourtPieceCpuDifficultyHard:
		return c.cpuSelectTrumpHard(playerIdx)
	case CourtPieceCpuDifficultyNormal:
		return c.cpuSelectTrumpNormal(playerIdx)
	default:
		return c.cpuSelectTrumpEasy(playerIdx)
	}
}

// cpuSelectTrumpEasy ランダムに4スートのいずれかを返す。
func (c *CourtPiece) cpuSelectTrumpEasy(_ int) int {
	return CardDesignSpade + rand.Intn(4)
}

// cpuSelectTrumpNormal 最長スートを選ぶ。同数なら最初に見つけたスート。
func (c *CourtPiece) cpuSelectTrumpNormal(playerIdx int) int {
	counts := c.suitCounts(playerIdx)
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
func (c *CourtPiece) cpuSelectTrumpHard(playerIdx int) int {
	p := c.players[playerIdx]
	score := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		card := p.GetCard(i)
		score[card.GetDesign()]++
		switch card.GetValue() {
		case 1, 13:
			score[card.GetDesign()] += 3
		case 12:
			score[card.GetDesign()] += 2
		case 11:
			score[card.GetDesign()]++
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
func (c *CourtPiece) suitCounts(playerIdx int) map[int]int {
	p := c.players[playerIdx]
	counts := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		counts[p.GetCard(i).GetDesign()]++
	}
	return counts
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選ぶ。
func (c *CourtPiece) cpuSelectPlayCard(playerIdx int) int {
	valid := c.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	switch c.config.CpuDifficulty {
	case CourtPieceCpuDifficultyHard:
		return c.cpuPlayHard(playerIdx, valid)
	case CourtPieceCpuDifficultyNormal:
		return c.cpuPlayNormal(playerIdx, valid)
	default:
		return c.cpuPlayEasy(valid)
	}
}

// cpuPlayEasy ランダム。
func (c *CourtPiece) cpuPlayEasy(valid []int) int {
	return valid[rand.Intn(len(valid))]
}

// cpuPlayNormal リードでは最高値、フォローでは「勝てる最小」または「最も低い」。
func (c *CourtPiece) cpuPlayNormal(playerIdx int, valid []int) int {
	p := c.players[playerIdx]
	if len(c.currentTrick) == 0 {
		return c.pickHighest(p, valid)
	}
	leadSuit := c.currentTrick[0].Card.GetDesign()
	maxLead, _, maxTrump, hasTrumpInTrick := c.summariseTrick(leadSuit)
	leadIdxs := c.filterByDesign(p, valid, leadSuit)
	if len(leadIdxs) > 0 {
		if !hasTrumpInTrick {
			over := c.filterAbove(p, leadIdxs, maxLead)
			if len(over) > 0 {
				return c.pickLowest(p, over)
			}
		}
		return c.pickLowest(p, leadIdxs)
	}
	trumpIdxs := c.filterByDesign(p, valid, c.trumpSuit)
	if len(trumpIdxs) > 0 {
		if hasTrumpInTrick {
			over := c.filterAbove(p, trumpIdxs, maxTrump)
			if len(over) > 0 {
				return c.pickLowest(p, over)
			}
		} else {
			return c.pickLowest(p, trumpIdxs)
		}
	}
	return c.pickLowest(p, valid)
}

// cpuPlayHard パートナー考慮 + 切り札保全の高度な戦略。
func (c *CourtPiece) cpuPlayHard(playerIdx int, valid []int) int {
	p := c.players[playerIdx]
	if len(c.currentTrick) == 0 {
		bestIdx := valid[0]
		// Seed below any achievable score: a trump card scores rank*2-30 (can be
		// negative), so -1 would wrongly reject an all-trump hand.
		bestScore := -1000
		for _, idx := range valid {
			card := p.GetCard(idx)
			score := courtPieceRank(card.GetValue()) * 2
			if card.GetDesign() == c.trumpSuit {
				score -= 30
			}
			if score > bestScore {
				bestScore = score
				bestIdx = idx
			}
		}
		return bestIdx
	}
	leadSuit := c.currentTrick[0].Card.GetDesign()
	maxLead, _, maxTrump, hasTrumpInTrick := c.summariseTrick(leadSuit)
	partnerWinning := c.isPartnerCurrentlyWinning(playerIdx)

	leadIdxs := c.filterByDesign(p, valid, leadSuit)
	if len(leadIdxs) > 0 {
		if partnerWinning {
			return c.pickLowest(p, leadIdxs)
		}
		if !hasTrumpInTrick {
			over := c.filterAbove(p, leadIdxs, maxLead)
			if len(over) > 0 {
				return c.pickLowest(p, over)
			}
		}
		return c.pickLowest(p, leadIdxs)
	}
	if partnerWinning {
		return c.pickLowest(p, valid)
	}
	trumpIdxs := c.filterByDesign(p, valid, c.trumpSuit)
	if len(trumpIdxs) > 0 {
		if hasTrumpInTrick {
			over := c.filterAbove(p, trumpIdxs, maxTrump)
			if len(over) > 0 {
				return c.pickLowest(p, over)
			}
		} else {
			return c.pickLowest(p, trumpIdxs)
		}
	}
	bestIdx := valid[0]
	// Seed below any achievable score: a trump discard scores rank-100 (negative),
	// so -1 would wrongly reject an all-trump hand and always discard valid[0].
	bestScore := -1000
	for _, idx := range valid {
		card := p.GetCard(idx)
		score := courtPieceRank(card.GetValue())
		if card.GetDesign() == c.trumpSuit {
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
func (c *CourtPiece) isPartnerCurrentlyWinning(playerIdx int) bool {
	if len(c.currentTrick) == 0 {
		return false
	}
	winnerIdx := c.trickWinner()
	if winnerIdx == playerIdx {
		return false
	}
	return c.players[winnerIdx].GetTeam() == c.players[playerIdx].GetTeam()
}

// summariseTrick 現トリックの最高リードスート rank、リードスート所持フラグ、最高トランプ rank、トランプ所持フラグを返す。
func (c *CourtPiece) summariseTrick(leadSuit int) (maxLead int, hasLead bool, maxTrump int, hasTrump bool) {
	for _, tc := range c.currentTrick {
		d := tc.Card.GetDesign()
		r := courtPieceRank(tc.Card.GetValue())
		if d == leadSuit {
			hasLead = true
			if r > maxLead {
				maxLead = r
			}
		}
		if d == c.trumpSuit {
			hasTrump = true
			if r > maxTrump {
				maxTrump = r
			}
		}
	}
	return
}

// pickHighest courtPieceRank が最大のカードのインデックスを返す。
func (c *CourtPiece) pickHighest(p *CourtPiecePlayer, valid []int) int {
	if p == nil || len(valid) == 0 {
		return 0
	}
	best := valid[0]
	bestR := courtPieceRank(p.GetCard(best).GetValue())
	for _, idx := range valid[1:] {
		r := courtPieceRank(p.GetCard(idx).GetValue())
		if r > bestR {
			bestR = r
			best = idx
		}
	}
	return best
}

// pickLowest courtPieceRank が最小のカードのインデックスを返す。
func (c *CourtPiece) pickLowest(p *CourtPiecePlayer, valid []int) int {
	if p == nil || len(valid) == 0 {
		return 0
	}
	best := valid[0]
	bestR := courtPieceRank(p.GetCard(best).GetValue())
	for _, idx := range valid[1:] {
		r := courtPieceRank(p.GetCard(idx).GetValue())
		if r < bestR {
			bestR = r
			best = idx
		}
	}
	return best
}

// filterByDesign 指定スートのみのインデックスを返す
func (c *CourtPiece) filterByDesign(p *CourtPiecePlayer, valid []int, design int) []int {
	out := make([]int, 0, len(valid))
	for _, idx := range valid {
		if p.GetCard(idx).GetDesign() == design {
			out = append(out, idx)
		}
	}
	return out
}

// filterAbove rank が threshold より大きいインデックスのみを返す。
func (c *CourtPiece) filterAbove(p *CourtPiecePlayer, valid []int, threshold int) []int {
	out := make([]int, 0, len(valid))
	for _, idx := range valid {
		if courtPieceRank(p.GetCard(idx).GetValue()) > threshold {
			out = append(out, idx)
		}
	}
	return out
}

// courtPieceJSON is the JSON wire format for CourtPiece.
type courtPieceJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*CourtPiecePlayer    `json:"ps"`
	Config           CourtPieceConfig       `json:"cf"`
	Phase            CourtPiecePhase        `json:"ph"`
	RoundNumber      int                    `json:"rn"`
	TrickNumber      int                    `json:"tn"`
	CurrentPlayerIdx int                    `json:"ci"`
	CurrentTrick     []*TrickCard           `json:"ct"`
	TrumpSuit        int                    `json:"ts"`
	CallerIdx        int                    `json:"ka"`
	LeadPlayerIdx    int                    `json:"li"`
	TeamScores       [CourtPieceTeamCnt]int `json:"sc"`
	LastWinnerTeam   int                    `json:"lw"`
	ConsecutiveWins  int                    `json:"cw"`
	LastRoundCourt   bool                   `json:"lc"`
	GameEndFlag      bool                   `json:"ge"`
	WinnerTeam       int                    `json:"wt"`
	ActionLog        []*ActionLogEntry      `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (c *CourtPiece) MarshalJSON() ([]byte, error) {
	return json.Marshal(courtPieceJSON{
		TrumpCards:       c.trumpCards,
		Players:          c.players,
		Config:           c.config,
		Phase:            c.phase,
		RoundNumber:      c.roundNumber,
		TrickNumber:      c.trickNumber,
		CurrentPlayerIdx: c.currentPlayerIdx,
		CurrentTrick:     c.currentTrick,
		TrumpSuit:        c.trumpSuit,
		CallerIdx:        c.callerIdx,
		LeadPlayerIdx:    c.leadPlayerIdx,
		TeamScores:       c.teamScores,
		LastWinnerTeam:   c.lastWinnerTeam,
		ConsecutiveWins:  c.consecutiveWins,
		LastRoundCourt:   c.lastRoundCourt,
		GameEndFlag:      c.gameEndFlag,
		WinnerTeam:       c.winnerTeam,
		ActionLog:        c.actionLog,
	})
}

// courtPieceMaxSliceLen 復元時の上限 (DoS 防止)。
const courtPieceMaxSliceLen = 1000

// courtPieceMaxActionLogLen ActionLog の復元時上限。
const courtPieceMaxActionLogLen = 5000

// errCourtPieceInvalidState is returned when a restored index/state field is out of range.
var errCourtPieceInvalidState = newCourtPieceError("courtpiece: invalid state values in json")

// newCourtPieceError wraps a static message so the worker binary ships one sentinel.
func newCourtPieceError(msg string) error { return &courtPieceError{msg} }

type courtPieceError struct{ msg string }

func (e *courtPieceError) Error() string { return e.msg }

// UnmarshalJSON implements json.Unmarshaler.
func (c *CourtPiece) UnmarshalJSON(data []byte) error {
	var j courtPieceJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > courtPieceMaxSliceLen || len(j.CurrentTrick) > courtPieceMaxSliceLen ||
		len(j.ActionLog) > courtPieceMaxActionLogLen {
		return errCourtPieceInvalidState
	}
	if len(j.Players) != CourtPiecePlayerCnt {
		return errCourtPieceInvalidState
	}
	for _, p := range j.Players {
		if p == nil {
			return errCourtPieceInvalidState
		}
	}
	if j.CurrentPlayerIdx < -1 || j.CurrentPlayerIdx >= CourtPiecePlayerCnt ||
		j.CallerIdx < 0 || j.CallerIdx >= CourtPiecePlayerCnt ||
		j.LeadPlayerIdx < -1 || j.LeadPlayerIdx >= CourtPiecePlayerCnt ||
		j.WinnerTeam < -1 || j.WinnerTeam >= CourtPieceTeamCnt ||
		j.LastWinnerTeam < -1 || j.LastWinnerTeam >= CourtPieceTeamCnt ||
		j.TrumpSuit < 0 || j.TrumpSuit > CardDesignDiamond ||
		j.RoundNumber < 1 ||
		j.TrickNumber < 0 || j.TrickNumber > CourtPieceHandSize ||
		j.ConsecutiveWins < 0 ||
		j.Phase < CourtPiecePhaseTrumpDeclaration || j.Phase > CourtPiecePhaseGameEnd {
		return errCourtPieceInvalidState
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= CourtPiecePlayerCnt {
			return errCourtPieceInvalidState
		}
	}
	c.trumpCards = j.TrumpCards
	if c.trumpCards == nil {
		c.trumpCards = NewTrumpCards(0)
	}
	c.players = j.Players
	if c.players == nil {
		c.players = make([]*CourtPiecePlayer, 0)
	}
	c.config = j.Config
	c.phase = j.Phase
	c.roundNumber = j.RoundNumber
	c.trickNumber = j.TrickNumber
	c.currentPlayerIdx = j.CurrentPlayerIdx
	c.currentTrick = j.CurrentTrick
	if c.currentTrick == nil {
		c.currentTrick = make([]*TrickCard, 0)
	}
	c.trumpSuit = j.TrumpSuit
	c.callerIdx = j.CallerIdx
	c.leadPlayerIdx = j.LeadPlayerIdx
	c.teamScores = j.TeamScores
	c.lastWinnerTeam = j.LastWinnerTeam
	c.consecutiveWins = j.ConsecutiveWins
	c.lastRoundCourt = j.LastRoundCourt
	c.gameEndFlag = j.GameEndFlag
	c.winnerTeam = j.WinnerTeam
	c.actionLog = j.ActionLog
	if c.actionLog == nil {
		c.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
