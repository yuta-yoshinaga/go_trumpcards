package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// GongZhuPlayerCnt 拱猪（Gong Zhu）プレイヤー数
const GongZhuPlayerCnt = 4

// GongZhuHandSize 各プレイヤーの手札枚数
const GongZhuHandSize = 13

// GongZhuAllHeartsBonus 全ハート獲得（シュート・ザ・ムーン相当）時の得点
const GongZhuAllHeartsBonus = 200

// GongZhuPhase ゲームフェーズ
type GongZhuPhase int

// Gong Zhuのフェーズ定数
const (
	// GongZhuPhaseExpose カード公開（明牌）フェーズ
	GongZhuPhaseExpose GongZhuPhase = 0
	// GongZhuPhasePlay トリックプレイフェーズ
	GongZhuPhasePlay GongZhuPhase = 1
	// GongZhuPhaseTrickEnd トリック終了フェーズ
	GongZhuPhaseTrickEnd GongZhuPhase = 2
	// GongZhuPhaseRoundEnd ラウンド終了フェーズ
	GongZhuPhaseRoundEnd GongZhuPhase = 3
	// GongZhuPhaseGameEnd ゲーム終了フェーズ
	GongZhuPhaseGameEnd GongZhuPhase = 4
)

// GongZhuHint ヒント情報
type GongZhuHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// GongZhuExposure 公開（明牌）されたポイントカードの集合。誰が獲得しても得点が倍増する。
type GongZhuExposure struct {
	Pig     bool `json:"pg"` // ♠Q
	Sheep   bool `json:"sh"` // ♦J
	Ace     bool `json:"ac"` // ♥A（全ハートが倍）
	Doubler bool `json:"db"` // ♣10（変圧器が ×4 / 単体 +100）
}

// GongZhu 拱猪（Gong Zhu）ゲームクラス
type GongZhu struct {
	trumpCards       *TrumpCards
	players          []*GongZhuPlayer
	config           GongZhuConfig
	phase            GongZhuPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	heartsBroken     bool
	exposed          GongZhuExposure
	exposeReady      [GongZhuPlayerCnt]bool
	leadPlayerIdx    int
	gameEndFlag      bool
	winnerIdx        int
	actionLog        []*ActionLogEntry
}

// NewGongZhu コンストラクタ
func NewGongZhu(trumpCards *TrumpCards, players []*GongZhuPlayer, config GongZhuConfig) *GongZhu {
	return &GongZhu{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
	}
}

// NewDefaultGongZhu returns Gong Zhu with the standard 4-player setup (1 human, 3 CPU)
// and DefaultGongZhuConfig. Single source of truth for CUI, Web, and Worker construction.
func NewDefaultGongZhu() *GongZhu {
	players := []*GongZhuPlayer{
		NewGongZhuPlayer(true),
		NewGongZhuPlayer(false),
		NewGongZhuPlayer(false),
		NewGongZhuPlayer(false),
	}
	return NewGongZhu(NewTrumpCards(0), players, DefaultGongZhuConfig())
}

// Reset ゲーム初期化: デッキをシャッフルして配布し、最初のフェーズを設定
func (g *GongZhu) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.trickNumber = 0
	g.heartsBroken = false
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.currentPlayerIdx = -1
	g.actionLog = nil
	g.exposed = GongZhuExposure{}
	g.exposeReady = [GongZhuPlayerCnt]bool{}

	for _, p := range g.players {
		p.SetRoundScore(0)
		p.SetCumulativeScore(0)
		p.ResetTricks()
		p.Reset()
		p.SetIsFinished(false)
	}

	g.trumpCards.Shuffle()
	dealAllCards(g.trumpCards, g.players)
	g.sortAllHands()

	g.phase = GongZhuPhaseExpose
}

// NextRound 次のラウンドを開始する
func (g *GongZhu) NextRound() {
	if g.phase != GongZhuPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.trickNumber = 0
	g.heartsBroken = false
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.currentPlayerIdx = -1
	g.exposed = GongZhuExposure{}
	g.exposeReady = [GongZhuPlayerCnt]bool{}

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	dealAllCards(g.trumpCards, g.players)
	g.sortAllHands()

	g.phase = GongZhuPhaseExpose
}

// PlayerExpose 人間プレイヤーが公開するポイントカードを選択する。
// cardIndices は手札のインデックス（公開できるカードのみ）。空スライスは「公開なし」。
func (g *GongZhu) PlayerExpose(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GongZhuPhaseExpose {
		return ErrWrongPhase
	}

	humanIdx := g.findHumanIdx()
	if humanIdx < 0 {
		return ErrNotHumanTurn
	}
	if g.exposeReady[humanIdx] {
		return NewDomainError(ErrInvalidPlay, "すでに公開選択は完了しています")
	}

	player := g.players[humanIdx]
	seen := make(map[int]bool, len(cardIndices))
	for _, idx := range cardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "カードインデックスが重複しています")
		}
		seen[idx] = true
		if !gzIsSpecial(player.GetCard(idx)) {
			return NewDomainError(ErrInvalidPlay, "公開できるのはポイントカード（♠Q, ♦J, ♥A, ♣10）のみです")
		}
	}

	for _, idx := range cardIndices {
		g.markExposed(player.GetCard(idx))
	}
	g.exposeReady[humanIdx] = true
	return nil
}

// CpuExpose すべてのCPUプレイヤーの公開選択を行う（CPUは公開しない）
func (g *GongZhu) CpuExpose() {
	for i := 0; i < GongZhuPlayerCnt; i++ {
		if g.players[i].GetIsHuman() || g.exposeReady[i] {
			continue
		}
		g.exposeReady[i] = true
	}
}

// ExecuteExpose 全員の公開選択が完了したらプレイフェーズへ移行する
func (g *GongZhu) ExecuteExpose() {
	if g.phase != GongZhuPhaseExpose {
		return
	}
	for i := 0; i < GongZhuPlayerCnt; i++ {
		if !g.exposeReady[i] {
			return
		}
	}

	g.appendLog(-1, "expose", fmt.Sprintf("round %d: %s", g.roundNumber, g.exposureSummary()), nil)
	g.phase = GongZhuPhasePlay
	g.startPlayPhase()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *GongZhu) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GongZhuPhasePlay {
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
func (g *GongZhu) CpuPlay() {
	if g.gameEndFlag || g.phase != GongZhuPhasePlay {
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
func (g *GongZhu) ResolveTrick() {
	if g.phase != GongZhuPhaseTrickEnd || len(g.currentTrick) != GongZhuPlayerCnt {
		return
	}

	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(trickCards)

	rawPts := 0
	for _, c := range trickCards {
		rawPts += gzCardRawPoints(c)
	}
	g.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d (raw %+d)", g.playerName(winnerIdx), g.trickNumber, rawPts), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= GongZhuHandSize {
		g.phase = GongZhuPhaseRoundEnd
	} else {
		g.phase = GongZhuPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (g *GongZhu) NextTrick() {
	if g.phase != GongZhuPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = GongZhuPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (g *GongZhu) ScoreRound() {
	if g.phase != GongZhuPhaseRoundEnd {
		return
	}

	for i := 0; i < GongZhuPlayerCnt; i++ {
		if g.playerHeartCount(i) == GongZhuHandSize {
			g.appendLog(i, "all_hearts", fmt.Sprintf("%s collected all hearts!", g.playerName(i)), nil)
		}
		g.players[i].SetRoundScore(g.scoreForPlayer(i))
	}

	for i := 0; i < GongZhuPlayerCnt; i++ {
		g.players[i].CommitRoundScore()
	}

	for i := 0; i < GongZhuPlayerCnt; i++ {
		g.appendLog(i, "round_score", fmt.Sprintf("%s: round=%+d, total=%+d",
			g.playerName(i), g.players[i].GetRoundScore(), g.players[i].GetCumulativeScore()), nil)
	}

	ended := false
	for i := 0; i < GongZhuPlayerCnt; i++ {
		if g.players[i].GetCumulativeScore() <= -g.config.PointLimit {
			ended = true
			break
		}
	}

	if ended {
		g.gameEndFlag = true
		g.phase = GongZhuPhaseGameEnd
		// 最高スコアのプレイヤーが勝者
		maxScore := g.players[0].GetCumulativeScore()
		g.winnerIdx = 0
		for i := 1; i < GongZhuPlayerCnt; i++ {
			if g.players[i].GetCumulativeScore() > maxScore {
				maxScore = g.players[i].GetCumulativeScore()
				g.winnerIdx = i
			}
		}
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", g.playerName(g.winnerIdx)), nil)
	}
}

// scoreForPlayer プレイヤーのラウンド得点を算出する
func (g *GongZhu) scoreForPlayer(playerIdx int) int {
	heartCount := 0
	heartsSum := 0
	hasPig, hasSheep, hasDoubler := false, false, false

	for _, trick := range g.players[playerIdx].GetTricksTaken() {
		for _, c := range trick {
			switch {
			case c.GetDesign() == CardDesignHeart:
				heartCount++
				heartsSum -= gzHeartPenalty(c.GetValue())
			case gzIsPig(c):
				hasPig = true
			case gzIsSheep(c):
				hasSheep = true
			case gzIsDoubler(c):
				hasDoubler = true
			}
		}
	}

	if heartCount == GongZhuHandSize {
		heartsSum = GongZhuAllHeartsBonus
	}
	if g.exposed.Ace {
		heartsSum *= 2
	}

	base := heartsSum
	if hasPig {
		if g.exposed.Pig {
			base -= 200
		} else {
			base -= 100
		}
	}
	if hasSheep {
		if g.exposed.Sheep {
			base += 200
		} else {
			base += 100
		}
	}

	if hasDoubler {
		mult, standalone := 2, 50
		if g.exposed.Doubler {
			mult, standalone = 4, 100
		}
		if base == 0 {
			base = standalone
		} else {
			base *= mult
		}
	}
	return base
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *GongZhu) GetPhase() GongZhuPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *GongZhu) SetPhase(phase GongZhuPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *GongZhu) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号取得
func (g *GongZhu) GetTrickNumber() int { return g.trickNumber }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *GongZhu) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *GongZhu) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *GongZhu) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *GongZhu) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetHeartsBroken ハーツブレイク状態取得
func (g *GongZhu) GetHeartsBroken() bool { return g.heartsBroken }

// SetHeartsBroken ハーツブレイク状態設定 (テスト用)
func (g *GongZhu) SetHeartsBroken(broken bool) { g.heartsBroken = broken }

// GetExposure 公開状態取得
func (g *GongZhu) GetExposure() GongZhuExposure { return g.exposed }

// SetExposure 公開状態設定 (テスト用)
func (g *GongZhu) SetExposure(e GongZhuExposure) { g.exposed = e }

// GetExposeReady 公開準備状態取得
func (g *GongZhu) GetExposeReady() [GongZhuPlayerCnt]bool { return g.exposeReady }

// GetExposableIndices プレイヤーが公開できるカードのインデックス一覧を返す
func (g *GongZhu) GetExposableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	p := g.players[playerIdx]
	var indices []int
	for i := 0; i < p.GetCardsSize(); i++ {
		if gzIsSpecial(p.GetCard(i)) {
			indices = append(indices, i)
		}
	}
	return indices
}

// GetGameEndFlag ゲーム終了フラグ取得
func (g *GongZhu) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *GongZhu) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *GongZhu) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *GongZhu) GetPlayer(i int) *GongZhuPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *GongZhu) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *GongZhu) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// IsHumanTurn 現在の手番が人間かどうか
func (g *GongZhu) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *GongZhu) GetConfig() GongZhuConfig { return g.config }

// SetConfig 設定変更
func (g *GongZhu) SetConfig(cfg GongZhuConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *GongZhu) GetActionLog() []*ActionLogEntry { return g.actionLog }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *GongZhu) SetRoundNumber(n int) { g.roundNumber = n }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *GongZhu) SetTrickNumber(n int) { g.trickNumber = n }

// --- Private methods ---

// findHumanIdx 人間プレイヤーのインデックスを返す (-1=なし)
func (g *GongZhu) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// markExposed 公開対象カードに応じて公開フラグを立てる
func (g *GongZhu) markExposed(c *Card) {
	switch {
	case gzIsPig(c):
		g.exposed.Pig = true
	case gzIsSheep(c):
		g.exposed.Sheep = true
	case gzIsAceHeart(c):
		g.exposed.Ace = true
	case gzIsDoubler(c):
		g.exposed.Doubler = true
	}
}

// exposureSummary 公開状況の文字列表現
func (g *GongZhu) exposureSummary() string {
	var parts []string
	if g.exposed.Pig {
		parts = append(parts, "♠Q")
	}
	if g.exposed.Sheep {
		parts = append(parts, "♦J")
	}
	if g.exposed.Ace {
		parts = append(parts, "♥A")
	}
	if g.exposed.Doubler {
		parts = append(parts, "♣10")
	}
	if len(parts) == 0 {
		return "no cards exposed"
	}
	out := "exposed:"
	for i, p := range parts {
		if i > 0 {
			out += ","
		}
		out += " " + p
	}
	return out
}

// startPlayPhase プレイフェーズ開始: ♣2を持つプレイヤーをリードに設定
func (g *GongZhu) startPlayPhase() {
	if g.trickNumber == 0 {
		starter := g.findTwoOfClubs()
		if starter >= 0 {
			g.leadPlayerIdx = starter
			g.currentPlayerIdx = starter
		} else {
			g.leadPlayerIdx = 0
			g.currentPlayerIdx = 0
		}
		g.trickNumber = 1
		g.currentTrick = nil
	} else {
		g.currentPlayerIdx = g.leadPlayerIdx
	}
}

// findTwoOfClubs ♣2を持つプレイヤーのインデックスを返す
func (g *GongZhu) findTwoOfClubs() int {
	for i, p := range g.players {
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
func (g *GongZhu) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	if card.GetDesign() == CardDesignHeart {
		g.heartsBroken = true
	}

	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == GongZhuPlayerCnt {
		g.phase = GongZhuPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % GongZhuPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
func (g *GongZhu) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		// リード: ハーツが壊れていない場合、ハーツでリードできない（他にカードがある場合）
		if !g.heartsBroken && card.GetDesign() == CardDesignHeart {
			if g.playerHasNonHeart(playerIdx) {
				return NewDomainError(ErrInvalidPlay, "ハーツはまだブレイクされていません")
			}
		}
		return nil
	}

	// フォロースート
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && g.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (g *GongZhu) playerHasSuit(playerIdx int, design int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// playerHasNonHeart プレイヤーがハート以外のカードを持っているか
func (g *GongZhu) playerHasNonHeart(playerIdx int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() != CardDesignHeart {
			return true
		}
	}
	return false
}

// playerHeartCount プレイヤーが獲得したハートの枚数を返す
func (g *GongZhu) playerHeartCount(playerIdx int) int {
	count := 0
	for _, trick := range g.players[playerIdx].GetTricksTaken() {
		for _, c := range trick {
			if c.GetDesign() == CardDesignHeart {
				count++
			}
		}
	}
	return count
}

// trickWinner トリックの勝者を決定する
func (g *GongZhu) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerValue := gzRankValue(g.currentTrick[0].Card)

	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() == leadSuit && gzRankValue(tc.Card) > winnerValue {
			winnerIdx = tc.PlayerIdx
			winnerValue = gzRankValue(tc.Card)
		}
	}
	return winnerIdx
}

// sortAllHands 全プレイヤーの手札をソートする
func (g *GongZhu) sortAllHands() {
	for _, p := range g.players {
		gzSortHand(p)
	}
}

// gzSortHand プレイヤーの手札をスート→値の順にソートする
func gzSortHand(p *GongZhuPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// playerName プレイヤー名を返す
func (g *GongZhu) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (g *GongZhu) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Card helpers ---

// gzIsPig ♠Q（豚）かどうか
func gzIsPig(c *Card) bool { return c.GetDesign() == CardDesignSpade && c.GetValue() == 12 }

// gzIsSheep ♦J（羊）かどうか
func gzIsSheep(c *Card) bool { return c.GetDesign() == CardDesignDiamond && c.GetValue() == 11 }

// gzIsDoubler ♣10（変圧器）かどうか
func gzIsDoubler(c *Card) bool { return c.GetDesign() == CardDesignClover && c.GetValue() == 10 }

// gzIsAceHeart ♥A かどうか
func gzIsAceHeart(c *Card) bool { return c.GetDesign() == CardDesignHeart && c.GetValue() == 1 }

// gzIsSpecial 公開対象のポイントカード（♠Q, ♦J, ♥A, ♣10）かどうか
func gzIsSpecial(c *Card) bool {
	return gzIsPig(c) || gzIsSheep(c) || gzIsAceHeart(c) || gzIsDoubler(c)
}

// gzHeartPenalty ハートの得点の絶対値を返す（A=50,K=40,Q=30,J=20,5〜10=10,2〜4=0）
func gzHeartPenalty(value int) int {
	switch value {
	case 1:
		return 50
	case 13:
		return 40
	case 12:
		return 30
	case 11:
		return 20
	case 5, 6, 7, 8, 9, 10:
		return 10
	default:
		return 0
	}
}

// gzCardRawPoints トリック表示用の素点（倍率・公開を考慮しない）を返す
func gzCardRawPoints(c *Card) int {
	switch {
	case c.GetDesign() == CardDesignHeart:
		return -gzHeartPenalty(c.GetValue())
	case gzIsPig(c):
		return -100
	case gzIsSheep(c):
		return 100
	default:
		return 0
	}
}

// gzRankValue トリックの強さ比較用の値。Aは最強(14)として扱う。
func gzRankValue(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// --- Hint ---

// GetHint ヒントを取得する
func (g *GongZhu) GetHint() *GongZhuHint {
	human := g.findHumanIdx()
	if human < 0 {
		return nil
	}
	if g.phase == GongZhuPhaseExpose && !g.exposeReady[human] {
		return g.getExposeHint(human)
	}
	if g.phase == GongZhuPhasePlay && g.currentPlayerIdx == human {
		validIndices := g.getValidPlayIndices(human)
		if len(validIndices) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, validIndices)
		return &GongZhuHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// getExposeHint 公開フェーズのヒント（♦J を持っていれば公開を推奨）
func (g *GongZhu) getExposeHint(playerIdx int) *GongZhuHint {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if gzIsSheep(p.GetCard(i)) {
			return &GongZhuHint{CardIndices: []int{i}, Reason: "expose_sheep"}
		}
	}
	return &GongZhuHint{CardIndices: []int{}, Reason: "expose_none"}
}

// playHintReason プレイヒントの理由を判定する
func (g *GongZhu) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if gzIsPig(card) {
		return "discard_pig"
	}
	if card.GetDesign() == CardDesignHeart {
		return "discard_hearts"
	}
	return "discard_high"
}

// --- CPU AI ---

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (g *GongZhu) cpuSelectPlayCard(playerIdx int) int {
	validIndices := g.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}
	if g.config.CpuDifficulty == GongZhuCpuDifficultyEasy {
		return validIndices[rand.Intn(len(validIndices))]
	}
	return g.cpuPlaySmart(playerIdx, validIndices)
}

// cpuPlaySmart ポイントを意識した戦略プレイ
func (g *GongZhu) cpuPlaySmart(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]

	if len(g.currentTrick) == 0 {
		// リード: 危険度の低いカード（低い数値・非ポイントカード）を出す
		bestIdx := validIndices[0]
		bestScore := gzLeadDanger(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			s := gzLeadDanger(player.GetCard(idx))
			if s < bestScore {
				bestScore = s
				bestIdx = idx
			}
		}
		return bestIdx
	}

	leadSuit := g.currentTrick[0].Card.GetDesign()
	topValue := 0
	trickHasSheep := false
	trickNegative := false
	for _, tc := range g.currentTrick {
		if tc.Card.GetDesign() == leadSuit && gzRankValue(tc.Card) > topValue {
			topValue = gzRankValue(tc.Card)
		}
		if gzIsSheep(tc.Card) {
			trickHasSheep = true
		}
		if gzCardRawPoints(tc.Card) < 0 {
			trickNegative = true
		}
	}

	var follows []int
	for _, idx := range validIndices {
		if player.GetCard(idx).GetDesign() == leadSuit {
			follows = append(follows, idx)
		}
	}

	if len(follows) > 0 {
		return g.cpuFollow(player, follows, topValue, trickHasSheep, trickNegative)
	}
	return g.cpuDiscard(player, validIndices)
}

// cpuFollow フォロースート時のカード選択
func (g *GongZhu) cpuFollow(player *GongZhuPlayer, follows []int, topValue int, trickHasSheep, trickNegative bool) int {
	if trickHasSheep {
		// 羊（+100）を奪いに行く: 勝てる最高札を出す
		winners := gzFilter(follows, func(idx int) bool { return gzRankValue(player.GetCard(idx)) > topValue })
		if len(winners) > 0 {
			return gzHighest(player, winners)
		}
	}
	if trickNegative {
		// マイナスのトリックは避ける: 勝たない最高札（ダックの最大）を出す
		ducks := gzFilter(follows, func(idx int) bool { return gzRankValue(player.GetCard(idx)) < topValue })
		if len(ducks) > 0 {
			return gzHighest(player, ducks)
		}
		// ダックできない場合は最低札で損失を抑える
		return gzLowest(player, follows)
	}
	// 中立のトリック: 低い札を温存しつつ最低札を出す
	return gzLowest(player, follows)
}

// cpuDiscard ボイド時の捨て札選択（豚 > 高いハート > 高札、羊は温存）
func (g *GongZhu) cpuDiscard(player *GongZhuPlayer, validIndices []int) int {
	bestIdx := validIndices[0]
	bestScore := gzDiscardScore(player.GetCard(validIndices[0]))
	for _, idx := range validIndices[1:] {
		s := gzDiscardScore(player.GetCard(idx))
		if s > bestScore {
			bestScore = s
			bestIdx = idx
		}
	}
	return bestIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *GongZhu) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// gzLeadDanger リード時のカード危険度（低いほど安全）
func gzLeadDanger(c *Card) int {
	score := gzRankValue(c)
	if gzIsPig(c) {
		score += 100
	}
	if gzIsSheep(c) {
		score += 80
	}
	if c.GetDesign() == CardDesignHeart {
		score += 30
	}
	return score
}

// gzDiscardScore 捨て札の優先度（高いほど捨てたい）
func gzDiscardScore(c *Card) int {
	if gzIsSheep(c) {
		return -100 // 羊は温存
	}
	score := gzRankValue(c)
	if gzIsPig(c) {
		score += 200
	} else if c.GetDesign() == CardDesignHeart {
		score += 100 + gzHeartPenalty(c.GetValue())
	}
	return score
}

// gzFilter 述語を満たすインデックスを抽出する
func gzFilter(indices []int, pred func(int) bool) []int {
	var out []int
	for _, idx := range indices {
		if pred(idx) {
			out = append(out, idx)
		}
	}
	return out
}

// gzHighest 指定インデックス群のうち最も強いカードのインデックスを返す
func gzHighest(player *GongZhuPlayer, indices []int) int {
	best := indices[0]
	for _, idx := range indices[1:] {
		if gzRankValue(player.GetCard(idx)) > gzRankValue(player.GetCard(best)) {
			best = idx
		}
	}
	return best
}

// gzLowest 指定インデックス群のうち最も弱いカードのインデックスを返す
func gzLowest(player *GongZhuPlayer, indices []int) int {
	best := indices[0]
	for _, idx := range indices[1:] {
		if gzRankValue(player.GetCard(idx)) < gzRankValue(player.GetCard(best)) {
			best = idx
		}
	}
	return best
}

// --- JSON ---

// gongZhuJSON is the JSON wire format for GongZhu.
type gongZhuJSON struct {
	TrumpCards       *TrumpCards            `json:"tc"`
	Players          []*GongZhuPlayer       `json:"ps"`
	Config           GongZhuConfig          `json:"cf"`
	Phase            GongZhuPhase           `json:"ph"`
	RoundNumber      int                    `json:"rn"`
	TrickNumber      int                    `json:"tn"`
	CurrentPlayerIdx int                    `json:"ci"`
	CurrentTrick     []*TrickCard           `json:"ct"`
	HeartsBroken     bool                   `json:"hb"`
	Exposed          GongZhuExposure        `json:"ex"`
	ExposeReady      [GongZhuPlayerCnt]bool `json:"er"`
	LeadPlayerIdx    int                    `json:"li"`
	GameEndFlag      bool                   `json:"ge"`
	WinnerIdx        int                    `json:"wi"`
	ActionLog        []*ActionLogEntry      `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *GongZhu) MarshalJSON() ([]byte, error) {
	return json.Marshal(gongZhuJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		HeartsBroken:     g.heartsBroken,
		Exposed:          g.exposed,
		ExposeReady:      g.exposeReady,
		LeadPlayerIdx:    g.leadPlayerIdx,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		ActionLog:        g.actionLog,
	})
}

// gongZhuMaxSliceLen caps slice sizes during deserialisation.
const gongZhuMaxSliceLen = 1000

// errGongZhuOversized is the single sentinel error for oversized input arrays.
var errGongZhuOversized = errors.New("gongzhu: input array exceeds maximum allowed size")

// UnmarshalJSON implements json.Unmarshaler.
func (g *GongZhu) UnmarshalJSON(data []byte) error {
	var j gongZhuJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > gongZhuMaxSliceLen || len(j.CurrentTrick) > gongZhuMaxSliceLen ||
		len(j.ActionLog) > gongZhuMaxSliceLen {
		return errGongZhuOversized
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*GongZhuPlayer, 0)
	}
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.heartsBroken = j.HeartsBroken
	g.exposed = j.Exposed
	g.exposeReady = j.ExposeReady
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
