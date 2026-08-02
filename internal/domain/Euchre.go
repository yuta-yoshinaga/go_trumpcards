//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// EuchrePlayerCnt ユーカープレイヤー数
const EuchrePlayerCnt = 4

// EuchreHandSize 各プレイヤーの手札枚数
const EuchreHandSize = 5

// EuchreKittySize キティ枚数
const EuchreKittySize = 4

// EuchreTeamCnt チーム数
const EuchreTeamCnt = 2

// EuchrePhase ゲームフェーズ
type EuchrePhase int

// Euchreのフェーズ定数
const (
	// EuchrePhasePickUp ピックアップフェーズ (ラウンド1: 表向きカードのスートを切り札に指名するか)
	EuchrePhasePickUp EuchrePhase = 0
	// EuchrePhaseCallTrump コールトランプフェーズ (ラウンド2: 別のスートを切り札に指名するか)
	EuchrePhaseCallTrump EuchrePhase = 1
	// EuchrePhaseDiscard ディスカードフェーズ (ディーラーがカードを1枚捨てる)
	EuchrePhaseDiscard EuchrePhase = 2
	// EuchrePhasePlay トリックプレイフェーズ
	EuchrePhasePlay EuchrePhase = 3
	// EuchrePhaseTrickEnd トリック終了フェーズ
	EuchrePhaseTrickEnd EuchrePhase = 4
	// EuchrePhaseRoundEnd ラウンド終了フェーズ
	EuchrePhaseRoundEnd EuchrePhase = 5
	// EuchrePhaseGameEnd ゲーム終了フェーズ
	EuchrePhaseGameEnd EuchrePhase = 6
)

// EuchreHint ヒント情報
type EuchreHint struct {
	CardIndex *int   // 推奨カードインデックス (プレイ/ディスカード時)
	OrderUp   *bool  // ピックアップすべきか
	Suit      *int   // 推奨スート (コールトランプ時)
	GoAlone   *bool  // ゴーアローンすべきか
	Reason    string // ヒント理由キー
}

// Euchre ユーカーゲームクラス
type Euchre struct {
	trumpCards          *TrumpCards
	players             []*EuchrePlayer
	config              EuchreConfig
	phase               EuchrePhase
	roundNumber         int
	trickNumber         int
	currentPlayerIdx    int
	currentTrick        []*TrickCard
	dealerIdx           int
	trumpSuit           int // 切り札スート (CardDesignSpade等)
	faceUpCard          *Card
	kitty               []*Card
	makerTeam           int  // 切り札を指名したチーム (0 or 1)
	goingAlone          bool // ゴーアローン中か
	goingAlonePlayerIdx int  // ゴーアローンするプレイヤー
	teamScores          [EuchreTeamCnt]int
	leadPlayerIdx       int
	bidPlayerIdx        int // 現在のビッド手番
	gameEndFlag         bool
	winnerTeam          int // 勝利チーム (-1 = 未確定)
	actionLog           []*ActionLogEntry
}

// NewEuchre コンストラクタ
func NewEuchre(trumpCards *TrumpCards, players []*EuchrePlayer, config EuchreConfig) *Euchre {
	return &Euchre{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerTeam:  -1,
		roundNumber: 0,
		dealerIdx:   0,
	}
}

// NewDefaultEuchre returns Euchre with the standard 4-player team setup
// (human team 0, alternating CPU teams) and DefaultEuchreConfig.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultEuchre() *Euchre {
	players := []*EuchrePlayer{
		NewEuchrePlayer(true, 0),
		NewEuchrePlayer(false, 1),
		NewEuchrePlayer(false, 0),
		NewEuchrePlayer(false, 1),
	}
	return NewEuchre(NewTrumpCardsEuchre(), players, DefaultEuchreConfig())
}

// Reset ゲーム初期化
func (e *Euchre) Reset() {
	e.gameEndFlag = false
	e.winnerTeam = -1
	e.roundNumber = 1
	e.trickNumber = 0
	e.dealerIdx = 0
	e.teamScores = [EuchreTeamCnt]int{}
	e.actionLog = nil
	e.goingAlone = false
	e.goingAlonePlayerIdx = -1
	e.trumpSuit = 0
	e.makerTeam = 0

	for _, p := range e.players {
		p.ResetRound()
	}

	e.dealRound()
	e.phase = EuchrePhasePickUp
	e.bidPlayerIdx = (e.dealerIdx + 1) % EuchrePlayerCnt
}

// NextRound 次のラウンドを開始する
func (e *Euchre) NextRound() {
	if e.phase != EuchrePhaseRoundEnd {
		return
	}

	e.roundNumber++
	e.dealerIdx = (e.dealerIdx + 1) % EuchrePlayerCnt
	e.trickNumber = 0
	e.goingAlone = false
	e.goingAlonePlayerIdx = -1
	e.trumpSuit = 0
	e.currentTrick = nil
	e.leadPlayerIdx = -1

	for _, p := range e.players {
		p.ResetRound()
	}

	e.dealRound()
	e.phase = EuchrePhasePickUp
	e.bidPlayerIdx = (e.dealerIdx + 1) % EuchrePlayerCnt
}

// dealRound ラウンドのカードを配る (5枚ずつ + キティ4枚)
func (e *Euchre) dealRound() {
	e.trumpCards.Shuffle()
	e.kitty = nil
	e.faceUpCard = nil

	// 5枚ずつ配る
	for range EuchreHandSize {
		for j := range EuchrePlayerCnt {
			card := e.trumpCards.DrawCard()
			if card != nil {
				e.players[j].AddCard(card)
			}
		}
	}

	// 残り4枚がキティ
	for {
		card := e.trumpCards.DrawCard()
		if card == nil {
			break
		}
		e.kitty = append(e.kitty, card)
	}

	// キティのトップを表向きにする
	if len(e.kitty) > 0 {
		e.faceUpCard = e.kitty[0]
	}

	e.sortAllHands()
}

// --- Trump Selection: PickUp Phase ---

// PlayerPickUp 人間プレイヤーがピックアップ判断する
func (e *Euchre) PlayerPickUp(orderUp bool, goAlone bool) error {
	if e.gameEndFlag {
		return ErrGameEnded
	}
	if e.phase != EuchrePhasePickUp {
		return ErrWrongPhase
	}
	humanIdx := e.findHumanIdx()
	if humanIdx < 0 || e.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}

	if orderUp {
		e.doOrderUp(humanIdx, goAlone)
	} else {
		e.appendLog(humanIdx, "pass", fmt.Sprintf("%s passes", e.playerName(humanIdx)), nil)
		e.advanceBidPickUp()
	}
	return nil
}

// CpuPickUp CPUプレイヤーがピックアップ判断する
func (e *Euchre) CpuPickUp() {
	if e.gameEndFlag || e.phase != EuchrePhasePickUp {
		return
	}
	if e.bidPlayerIdx >= EuchrePlayerCnt {
		return
	}
	if e.players[e.bidPlayerIdx].GetIsHuman() {
		return
	}

	orderUp, goAlone := e.cpuSelectPickUp(e.bidPlayerIdx)
	if orderUp {
		e.doOrderUp(e.bidPlayerIdx, goAlone)
	} else {
		e.appendLog(e.bidPlayerIdx, "pass", fmt.Sprintf("%s passes", e.playerName(e.bidPlayerIdx)), nil)
		e.advanceBidPickUp()
	}
}

// doOrderUp オーダーアップ実行
func (e *Euchre) doOrderUp(playerIdx int, goAlone bool) {
	e.trumpSuit = e.faceUpCard.GetDesign()
	e.makerTeam = e.players[playerIdx].GetTeam()

	// ディーラーがフェイスアップカードを手札に追加
	e.players[e.dealerIdx].AddCard(e.faceUpCard)
	e.kitty = e.kitty[1:]

	if goAlone {
		e.goingAlone = true
		e.goingAlonePlayerIdx = playerIdx
		e.appendLog(playerIdx, "order_up_alone",
			fmt.Sprintf("%s orders up %s and goes alone", e.playerName(playerIdx), cardStr(e.faceUpCard)), []*Card{e.faceUpCard})
	} else {
		e.appendLog(playerIdx, "order_up",
			fmt.Sprintf("%s orders up %s", e.playerName(playerIdx), cardStr(e.faceUpCard)), []*Card{e.faceUpCard})
	}

	e.faceUpCard = nil

	// ディーラーがカードを捨てるフェーズへ
	e.phase = EuchrePhaseDiscard
	e.currentPlayerIdx = e.dealerIdx
}

// advanceBidPickUp ピックアップフェーズのビッドを進める
func (e *Euchre) advanceBidPickUp() {
	e.bidPlayerIdx = (e.bidPlayerIdx + 1) % EuchrePlayerCnt

	// 全員パスした場合 (ディーラーの次 = ディーラーの左 に戻ったら全員パス)
	startIdx := (e.dealerIdx + 1) % EuchrePlayerCnt
	if e.bidPlayerIdx == startIdx {
		// コールトランプフェーズへ
		e.phase = EuchrePhaseCallTrump
		e.bidPlayerIdx = startIdx
	}
}

// --- Trump Selection: CallTrump Phase ---

// PlayerCallTrump 人間プレイヤーがスートを指名する
func (e *Euchre) PlayerCallTrump(suit int, goAlone bool) error {
	if e.gameEndFlag {
		return ErrGameEnded
	}
	if e.phase != EuchrePhaseCallTrump {
		return ErrWrongPhase
	}
	humanIdx := e.findHumanIdx()
	if humanIdx < 0 || e.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if e.faceUpCard != nil && suit == e.faceUpCard.GetDesign() {
		return NewDomainError(ErrInvalidPlay, "表向きカードのスートは選べません")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "無効なスートです")
	}

	e.doCallTrump(humanIdx, suit, goAlone)
	return nil
}

// PlayerPassCall 人間プレイヤーがコールフェーズでパスする
func (e *Euchre) PlayerPassCall() error {
	if e.gameEndFlag {
		return ErrGameEnded
	}
	if e.phase != EuchrePhaseCallTrump {
		return ErrWrongPhase
	}
	humanIdx := e.findHumanIdx()
	if humanIdx < 0 || e.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	// スタックドディーラールール: ディーラーは必ず選ばなければならない
	if e.bidPlayerIdx == e.dealerIdx {
		return NewDomainError(ErrCannotPass, "ディーラーは必ずスートを選ばなければなりません")
	}

	e.appendLog(humanIdx, "pass", fmt.Sprintf("%s passes", e.playerName(humanIdx)), nil)
	e.advanceBidCallTrump()
	return nil
}

// CpuCallTrump CPUプレイヤーがコールトランプ判断する
func (e *Euchre) CpuCallTrump() {
	if e.gameEndFlag || e.phase != EuchrePhaseCallTrump {
		return
	}
	if e.players[e.bidPlayerIdx].GetIsHuman() {
		return
	}

	suit, goAlone := e.cpuSelectCallTrump(e.bidPlayerIdx)
	if suit > 0 {
		e.doCallTrump(e.bidPlayerIdx, suit, goAlone)
	} else {
		// スタックドディーラー: ディーラーは必ず選ぶ
		if e.bidPlayerIdx == e.dealerIdx {
			forcedSuit := e.cpuForceCallTrump(e.bidPlayerIdx)
			e.doCallTrump(e.bidPlayerIdx, forcedSuit, false)
		} else {
			e.appendLog(e.bidPlayerIdx, "pass", fmt.Sprintf("%s passes", e.playerName(e.bidPlayerIdx)), nil)
			e.advanceBidCallTrump()
		}
	}
}

// doCallTrump スートを指名する
func (e *Euchre) doCallTrump(playerIdx int, suit int, goAlone bool) {
	e.trumpSuit = suit
	e.makerTeam = e.players[playerIdx].GetTeam()

	suitName := suitStr(suit)
	if goAlone {
		e.goingAlone = true
		e.goingAlonePlayerIdx = playerIdx
		e.appendLog(playerIdx, "call_trump_alone",
			fmt.Sprintf("%s calls %s as trump and goes alone", e.playerName(playerIdx), suitName), nil)
	} else {
		e.appendLog(playerIdx, "call_trump",
			fmt.Sprintf("%s calls %s as trump", e.playerName(playerIdx), suitName), nil)
	}

	e.startPlayPhase()
}

// advanceBidCallTrump コールトランプフェーズのビッドを進める
func (e *Euchre) advanceBidCallTrump() {
	e.bidPlayerIdx = (e.bidPlayerIdx + 1) % EuchrePlayerCnt
	// スタックドディーラーにより全員パスは発生しない (ディーラーが強制的に選ぶ)
}

// --- Discard Phase ---

// PlayerDiscard 人間プレイヤー(ディーラー)がカードを1枚捨てる
func (e *Euchre) PlayerDiscard(cardIndex int) error {
	if e.gameEndFlag {
		return ErrGameEnded
	}
	if e.phase != EuchrePhaseDiscard {
		return ErrWrongPhase
	}
	if !e.players[e.dealerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := e.players[e.dealerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	discarded := player.RemoveCard(cardIndex)
	e.appendLog(e.dealerIdx, "discard", fmt.Sprintf("%s discards a card", e.playerName(e.dealerIdx)), []*Card{discarded})
	e.sortAllHands()
	e.startPlayPhase()
	return nil
}

// CpuDiscard CPUプレイヤー(ディーラー)がカードを1枚捨てる
func (e *Euchre) CpuDiscard() {
	if e.gameEndFlag || e.phase != EuchrePhaseDiscard {
		return
	}
	if e.players[e.dealerIdx].GetIsHuman() {
		return
	}

	idx := e.cpuSelectDiscard(e.dealerIdx)
	discarded := e.players[e.dealerIdx].RemoveCard(idx)
	e.appendLog(e.dealerIdx, "discard", fmt.Sprintf("%s discards a card", e.playerName(e.dealerIdx)), []*Card{discarded})
	e.sortAllHands()
	e.startPlayPhase()
}

// --- Play Phase ---

// PlayerPlay 人間プレイヤーがカードをプレイする
func (e *Euchre) PlayerPlay(cardIndex int) error {
	if e.gameEndFlag {
		return ErrGameEnded
	}
	if e.phase != EuchrePhasePlay {
		return ErrWrongPhase
	}
	if !e.players[e.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := e.players[e.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := e.validatePlay(e.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	e.playCard(e.currentPlayerIdx, played)
	return nil
}

// CpuPlay CPUプレイヤーが1ターン実行
func (e *Euchre) CpuPlay() {
	if e.gameEndFlag || e.phase != EuchrePhasePlay {
		return
	}
	if e.players[e.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := e.players[e.currentPlayerIdx]
	cardIdx := e.cpuSelectPlayCard(e.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	e.playCard(e.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (e *Euchre) ResolveTrick() {
	expectedCards := e.activePlayerCount()
	if e.phase != EuchrePhaseTrickEnd || len(e.currentTrick) != expectedCards {
		return
	}

	winnerIdx := e.trickWinner()
	trickCards := make([]*Card, len(e.currentTrick))
	for i, tc := range e.currentTrick {
		trickCards[i] = tc.Card
	}

	e.players[winnerIdx].AddTrick(trickCards)

	winnerName := e.playerName(winnerIdx)
	e.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d", winnerName, e.trickNumber), trickCards)

	e.leadPlayerIdx = winnerIdx

	if e.trickNumber >= EuchreHandSize {
		e.phase = EuchrePhaseRoundEnd
	} else {
		e.phase = EuchrePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (e *Euchre) NextTrick() {
	if e.phase != EuchrePhaseTrickEnd {
		return
	}
	e.currentTrick = nil
	e.currentPlayerIdx = e.leadPlayerIdx
	e.trickNumber++
	e.phase = EuchrePhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (e *Euchre) ScoreRound() {
	if e.phase != EuchrePhaseRoundEnd {
		return
	}

	// チームごとのトリック数を集計
	teamTricks := [EuchreTeamCnt]int{}
	for _, p := range e.players {
		teamTricks[p.GetTeam()] += p.GetTrickCount()
	}

	makerTricks := teamTricks[e.makerTeam]
	defenderTeam := 1 - e.makerTeam

	if makerTricks >= 3 {
		if makerTricks == EuchreHandSize {
			// マーチ (全5トリック獲得)
			if e.goingAlone {
				e.teamScores[e.makerTeam] += 4
				e.appendLog(-1, "march_alone",
					fmt.Sprintf("Team %d marches going alone! +4 points", e.makerTeam), nil)
			} else {
				e.teamScores[e.makerTeam] += 2
				e.appendLog(-1, "march",
					fmt.Sprintf("Team %d marches! +2 points", e.makerTeam), nil)
			}
		} else {
			// メイカー勝利 (3-4トリック)
			e.teamScores[e.makerTeam]++
			e.appendLog(-1, "maker_win",
				fmt.Sprintf("Team %d wins the round! +1 point", e.makerTeam), nil)
		}
	} else {
		// ユーカー (メイカーが3トリック未満)
		e.teamScores[defenderTeam] += 2
		e.appendLog(-1, "euchred",
			fmt.Sprintf("Team %d is euchred! Team %d +2 points", e.makerTeam, defenderTeam), nil)
	}

	// スコアログ
	for ti := range EuchreTeamCnt {
		e.appendLog(-1, "team_score",
			fmt.Sprintf("Team %d: %d points (tricks: %d)", ti, e.teamScores[ti], teamTricks[ti]), nil)
	}

	e.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (e *Euchre) GetPhase() EuchrePhase { return e.phase }

// SetPhase フェーズ設定 (テスト用)
func (e *Euchre) SetPhase(phase EuchrePhase) { e.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (e *Euchre) GetRoundNumber() int { return e.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (e *Euchre) SetRoundNumber(n int) { e.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (e *Euchre) GetTrickNumber() int { return e.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (e *Euchre) SetTrickNumber(n int) { e.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (e *Euchre) GetCurrentPlayerIdx() int { return e.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (e *Euchre) SetCurrentPlayerIdx(idx int) { e.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (e *Euchre) GetCurrentTrick() []*TrickCard { return e.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (e *Euchre) SetCurrentTrick(trick []*TrickCard) { e.currentTrick = trick }

// GetGameEndFlag ゲーム終了フラグ取得
func (e *Euchre) GetGameEndFlag() bool { return e.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (e *Euchre) GetWinnerTeam() int { return e.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (e *Euchre) GetPlayerCnt() int { return len(e.players) }

// GetPlayer プレイヤー取得
func (e *Euchre) GetPlayer(i int) *EuchrePlayer {
	if i < 0 || i >= len(e.players) {
		return nil
	}
	return e.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (e *Euchre) GetLeadPlayerIdx() int { return e.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (e *Euchre) SetLeadPlayerIdx(idx int) { e.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (e *Euchre) GetBidPlayerIdx() int { return e.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (e *Euchre) SetBidPlayerIdx(idx int) { e.bidPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (e *Euchre) GetDealerIdx() int { return e.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (e *Euchre) SetDealerIdx(idx int) { e.dealerIdx = idx }

// GetTrumpSuit 切り札スート取得
func (e *Euchre) GetTrumpSuit() int { return e.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (e *Euchre) SetTrumpSuit(suit int) { e.trumpSuit = suit }

// GetFaceUpCard 表向きカード取得 (nil = ピックアップ済み)
func (e *Euchre) GetFaceUpCard() *Card { return e.faceUpCard }

// SetFaceUpCard 表向きカード設定 (テスト用)
func (e *Euchre) SetFaceUpCard(card *Card) { e.faceUpCard = card }

// GetMakerTeam メイカーチーム取得
func (e *Euchre) GetMakerTeam() int { return e.makerTeam }

// SetMakerTeam メイカーチーム設定 (テスト用)
func (e *Euchre) SetMakerTeam(team int) { e.makerTeam = team }

// GetGoingAlone ゴーアローン状態取得
func (e *Euchre) GetGoingAlone() bool { return e.goingAlone }

// SetGoingAlone ゴーアローン設定 (テスト用)
func (e *Euchre) SetGoingAlone(alone bool) { e.goingAlone = alone }

// GetGoingAlonePlayerIdx ゴーアローンプレイヤーインデックス取得
func (e *Euchre) GetGoingAlonePlayerIdx() int { return e.goingAlonePlayerIdx }

// SetGoingAlonePlayerIdx ゴーアローンプレイヤーインデックス設定 (テスト用)
func (e *Euchre) SetGoingAlonePlayerIdx(idx int) { e.goingAlonePlayerIdx = idx }

// GetTeamScore チームスコア取得
func (e *Euchre) GetTeamScore(team int) int {
	if team < 0 || team >= EuchreTeamCnt {
		return 0
	}
	return e.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (e *Euchre) SetTeamScore(team, score int) {
	if team >= 0 && team < EuchreTeamCnt {
		e.teamScores[team] = score
	}
}

// GetKitty キティ取得
func (e *Euchre) GetKitty() []*Card { return e.kitty }

// IsHumanTurn 現在の手番が人間かどうか
func (e *Euchre) IsHumanTurn() bool {
	if e.currentPlayerIdx < 0 || e.currentPlayerIdx >= len(e.players) {
		return false
	}
	return e.players[e.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (e *Euchre) IsHumanBidTurn() bool {
	if e.bidPlayerIdx < 0 || e.bidPlayerIdx >= len(e.players) {
		return false
	}
	return e.players[e.bidPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (e *Euchre) GetConfig() EuchreConfig { return e.config }

// SetConfig 設定変更
func (e *Euchre) SetConfig(cfg EuchreConfig) { e.config = cfg }

// GetActionLog 棋譜取得
func (e *Euchre) GetActionLog() []*ActionLogEntry { return e.actionLog }

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (e *Euchre) CardRankPublic(card *Card) int { return e.cardRank(card) }

// EffectiveSuitPublic 実効スート取得 (テスト用公開メソッド)
func (e *Euchre) EffectiveSuitPublic(card *Card) int { return e.effectiveSuit(card) }

// --- Bower helpers ---

// sameColorSuit 同色スートを返す (Spade↔Clover, Heart↔Diamond)
func sameColorSuit(suit int) int {
	switch suit {
	case CardDesignSpade:
		return CardDesignClover
	case CardDesignClover:
		return CardDesignSpade
	case CardDesignHeart:
		return CardDesignDiamond
	case CardDesignDiamond:
		return CardDesignHeart
	}
	return suit
}

// isRightBower 右バウアー (切り札スートのJ) かどうか
func (e *Euchre) isRightBower(card *Card) bool {
	return card.GetValue() == 11 && card.GetDesign() == e.trumpSuit
}

// isLeftBower 左バウアー (同色スートのJ) かどうか
func (e *Euchre) isLeftBower(card *Card) bool {
	return card.GetValue() == 11 && card.GetDesign() == sameColorSuit(e.trumpSuit)
}

// effectiveSuit カードの実効スートを返す (左バウアーは切り札スート)
func (e *Euchre) effectiveSuit(card *Card) int {
	if e.isLeftBower(card) {
		return e.trumpSuit
	}
	return card.GetDesign()
}

// cardRank トリック比較用のカードランクを返す (高い = 強い)
// 切り札: RightBower(600) > LeftBower(500) > A(414) > K(413) > Q(412) > 10(410) > 9(409)
// 非切り札: A(114) > K(113) > Q(112) > J(111) > 10(110) > 9(109)
func (e *Euchre) cardRank(card *Card) int {
	if e.isRightBower(card) {
		return 600
	}
	if e.isLeftBower(card) {
		return 500
	}

	base := card.GetValue()
	// Aceは最強 (value=1 だが14として扱う)
	if base == 1 {
		base = 14
	}

	if e.effectiveSuit(card) == e.trumpSuit {
		return 400 + base
	}
	return 100 + base
}

// --- Private methods ---

// findHumanIdx 人間プレイヤーのインデックスを返す (-1=なし)
func (e *Euchre) findHumanIdx() int {
	for i, p := range e.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// startPlayPhase プレイフェーズを開始する
func (e *Euchre) startPlayPhase() {
	e.trickNumber = 1
	e.currentTrick = nil
	e.leadPlayerIdx = (e.dealerIdx + 1) % EuchrePlayerCnt
	// ゴーアローン時、リーダーがスキップ対象なら次へ
	if e.goingAlone && e.isSkippedPlayer(e.leadPlayerIdx) {
		e.leadPlayerIdx = e.nextActivePlayer(e.leadPlayerIdx)
	}
	e.currentPlayerIdx = e.leadPlayerIdx
	e.phase = EuchrePhasePlay
}

// activePlayerCount アクティブなプレイヤー数 (ゴーアローン時は3)
func (e *Euchre) activePlayerCount() int {
	if e.goingAlone {
		return EuchrePlayerCnt - 1
	}
	return EuchrePlayerCnt
}

// isSkippedPlayer ゴーアローン時にスキップされるプレイヤーか
func (e *Euchre) isSkippedPlayer(idx int) bool {
	if !e.goingAlone {
		return false
	}
	// ゴーアローンプレイヤーのパートナーをスキップ
	partnerIdx := (e.goingAlonePlayerIdx + 2) % EuchrePlayerCnt
	return idx == partnerIdx
}

// nextActivePlayer 次のアクティブプレイヤーを返す (スキップ対象を飛ばす)
func (e *Euchre) nextActivePlayer(idx int) int {
	next := (idx + 1) % EuchrePlayerCnt
	if e.isSkippedPlayer(next) {
		next = (next + 1) % EuchrePlayerCnt
	}
	return next
}

// playCard カードをプレイする共通処理
func (e *Euchre) playCard(playerIdx int, card *Card) {
	e.currentTrick = append(e.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	e.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", e.playerName(playerIdx), cardStr(card)), []*Card{card})

	expectedCards := e.activePlayerCount()
	if len(e.currentTrick) == expectedCards {
		e.phase = EuchrePhaseTrickEnd
	} else {
		e.currentPlayerIdx = e.nextActivePlayer(e.currentPlayerIdx)
	}
}

// validatePlay カードのプレイが有効か検証する
func (e *Euchre) validatePlay(playerIdx int, card *Card) error {
	if len(e.currentTrick) == 0 {
		return nil // リードは自由
	}

	// フォロースート (実効スートで判断)
	leadSuit := e.effectiveSuit(e.currentTrick[0].Card)
	if e.effectiveSuit(card) != leadSuit {
		if e.playerHasEffectiveSuit(playerIdx, leadSuit) {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
	}

	return nil
}

// playerHasEffectiveSuit プレイヤーが実効スートのカードを持っているか
func (e *Euchre) playerHasEffectiveSuit(playerIdx int, suit int) bool {
	p := e.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if e.effectiveSuit(p.GetCard(i)) == suit {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する
// 切り札が出ていれば最強の切り札が勝ち、なければリードスートの最強が勝つ
func (e *Euchre) trickWinner() int {
	if len(e.currentTrick) == 0 {
		return 0
	}

	winnerIdx := e.currentTrick[0].PlayerIdx
	winnerRank := e.cardRank(e.currentTrick[0].Card)
	winnerSuit := e.effectiveSuit(e.currentTrick[0].Card)
	leadSuit := winnerSuit

	for _, tc := range e.currentTrick[1:] {
		rank := e.cardRank(tc.Card)
		effSuit := e.effectiveSuit(tc.Card)

		if effSuit == e.trumpSuit && winnerSuit != e.trumpSuit {
			// 切り札が非切り札の勝者に勝つ
			winnerIdx = tc.PlayerIdx
			winnerRank = rank
			winnerSuit = effSuit
		} else if effSuit == winnerSuit && rank > winnerRank {
			// 同じスート同士: 高ランクが勝つ
			winnerIdx = tc.PlayerIdx
			winnerRank = rank
		} else if effSuit == leadSuit && winnerSuit != e.trumpSuit && winnerSuit != leadSuit && rank > winnerRank {
			// リードスートが非リード・非切り札の勝者に勝つ
			winnerIdx = tc.PlayerIdx
			winnerRank = rank
			winnerSuit = effSuit
		}
	}
	return winnerIdx
}

// checkGameEnd ゲーム終了判定
func (e *Euchre) checkGameEnd() {
	for ti := range EuchreTeamCnt {
		if e.teamScores[ti] >= e.config.PointLimit {
			e.gameEndFlag = true
			e.phase = EuchrePhaseGameEnd

			// 最高スコアのチームが勝者
			if e.teamScores[0] >= e.teamScores[1] {
				e.winnerTeam = 0
			} else {
				e.winnerTeam = 1
			}
			e.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game!", e.winnerTeam), nil)
			return
		}
	}
}

// sortAllHands 全プレイヤーの手札をソートする
func (e *Euchre) sortAllHands() {
	for _, p := range e.players {
		euchreSortHand(p, e)
	}
}

// euchreSortHand プレイヤーの手札を実効スート→ランクの順にソートする
func euchreSortHand(p *EuchrePlayer, e *Euchre) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		si := e.effectiveSuit(ci)
		sj := e.effectiveSuit(cj)
		if si != sj {
			return si < sj
		}
		return e.cardRank(ci) < e.cardRank(cj)
	})
}

// playerName プレイヤー名を返す
func (e *Euchre) playerName(idx int) string {
	if idx < 0 || idx >= len(e.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if e.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// suitStr スート名を返す

// appendLog 棋譜にエントリを追加する
func (e *Euchre) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	e.actionLog = append(e.actionLog, &ActionLogEntry{
		TurnNumber: len(e.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// GetHint ヒントを取得する
func (e *Euchre) GetHint() *EuchreHint {
	humanIdx := e.findHumanIdx()
	if humanIdx < 0 {
		return nil
	}

	switch e.phase {
	case EuchrePhasePickUp:
		if e.bidPlayerIdx != humanIdx {
			return nil
		}
		orderUp, goAlone := e.cpuEvalPickUp(humanIdx)
		return &EuchreHint{OrderUp: &orderUp, GoAlone: &goAlone, Reason: "strategic_pickup"}

	case EuchrePhaseCallTrump:
		if e.bidPlayerIdx != humanIdx {
			return nil
		}
		suit, goAlone := e.cpuEvalCallTrump(humanIdx)
		if suit > 0 {
			return &EuchreHint{Suit: &suit, GoAlone: &goAlone, Reason: "strategic_call"}
		}
		f := false
		return &EuchreHint{GoAlone: &f, Reason: "pass_recommended"}

	case EuchrePhaseDiscard:
		if e.dealerIdx != humanIdx {
			return nil
		}
		idx := e.cpuSelectDiscard(humanIdx)
		return &EuchreHint{CardIndex: &idx, Reason: "discard_weakest"}

	case EuchrePhasePlay:
		if e.currentPlayerIdx != humanIdx {
			return nil
		}
		validIndices := e.getValidPlayIndices(humanIdx)
		if len(validIndices) == 0 {
			return nil
		}
		idx := e.cpuPlayHard(humanIdx, validIndices)
		return &EuchreHint{CardIndex: &idx, Reason: e.playHintReason(idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由を判定する
func (e *Euchre) playHintReason(chosenIdx int) string {
	player := e.players[e.findHumanIdx()]
	card := player.GetCard(chosenIdx)

	if len(e.currentTrick) == 0 {
		if e.effectiveSuit(card) == e.trumpSuit {
			return "lead_trump"
		}
		return "lead_strong"
	}

	leadSuit := e.effectiveSuit(e.currentTrick[0].Card)
	if e.effectiveSuit(card) == leadSuit {
		return "follow_suit"
	}
	if e.effectiveSuit(card) == e.trumpSuit {
		return "trump_cut"
	}
	return "discard_weak"
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (e *Euchre) GetValidPlayIndices(playerIdx int) []int {
	return e.getValidPlayIndices(playerIdx)
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (e *Euchre) getValidPlayIndices(playerIdx int) []int {
	player := e.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return e.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- CPU AI ---

// cpuSelectPickUp CPUがピックアップするか判断する
func (e *Euchre) cpuSelectPickUp(playerIdx int) (orderUp bool, goAlone bool) {
	switch e.config.CpuDifficulty {
	case EuchreCpuDifficultyHard:
		return e.cpuEvalPickUp(playerIdx)
	case EuchreCpuDifficultyNormal:
		return e.cpuPickUpNormal(playerIdx)
	default:
		return e.cpuPickUpEasy(playerIdx)
	}
}

// cpuPickUpEasy ランダムにピックアップ判断 (30%の確率でオーダーアップ)
func (e *Euchre) cpuPickUpEasy(playerIdx int) (bool, bool) {
	if rand.Intn(100) < 30 {
		return true, false
	}
	// ディーラーは最後のチャンスなのでもう少し積極的に
	if playerIdx == e.dealerIdx && rand.Intn(100) < 50 {
		return true, false
	}
	return false, false
}

// cpuPickUpNormal カードの強さに基づくピックアップ判断
func (e *Euchre) cpuPickUpNormal(playerIdx int) (bool, bool) {
	score := e.evalHandForTrump(playerIdx, e.faceUpCard.GetDesign())
	if score >= 3 {
		return true, false
	}
	// ディーラーはボーナス (ピックアップで1枚交換できるため)
	if playerIdx == e.dealerIdx && score >= 2 {
		return true, false
	}
	return false, false
}

// cpuEvalPickUp 高度なピックアップ評価
func (e *Euchre) cpuEvalPickUp(playerIdx int) (bool, bool) {
	score := e.evalHandForTrump(playerIdx, e.faceUpCard.GetDesign())

	// ディーラーはフェイスアップカードを獲得するためボーナス
	if playerIdx == e.dealerIdx {
		score++
	}

	if score >= 5 {
		return true, true // ゴーアローン
	}
	if score >= 3 {
		return true, false
	}
	return false, false
}

// cpuSelectCallTrump CPUがコールトランプで指名するスートを選択する
func (e *Euchre) cpuSelectCallTrump(playerIdx int) (suit int, goAlone bool) {
	switch e.config.CpuDifficulty {
	case EuchreCpuDifficultyHard:
		return e.cpuEvalCallTrump(playerIdx)
	case EuchreCpuDifficultyNormal:
		return e.cpuCallTrumpNormal(playerIdx)
	default:
		return e.cpuCallTrumpEasy(playerIdx)
	}
}

// cpuCallTrumpEasy ランダムにコール判断
func (e *Euchre) cpuCallTrumpEasy(_ int) (int, bool) {
	if rand.Intn(100) < 40 {
		suits := e.availableCallSuits()
		if len(suits) > 0 {
			return suits[rand.Intn(len(suits))], false
		}
	}
	return 0, false
}

// cpuCallTrumpNormal カードの強さに基づくコール判断
func (e *Euchre) cpuCallTrumpNormal(playerIdx int) (int, bool) {
	bestSuit := 0
	bestScore := 0
	for _, suit := range e.availableCallSuits() {
		score := e.evalHandForTrump(playerIdx, suit)
		if score > bestScore {
			bestScore = score
			bestSuit = suit
		}
	}
	if bestScore >= 3 {
		return bestSuit, false
	}
	return 0, false
}

// cpuEvalCallTrump 高度なコール評価
func (e *Euchre) cpuEvalCallTrump(playerIdx int) (int, bool) {
	bestSuit := 0
	bestScore := 0
	for _, suit := range e.availableCallSuits() {
		score := e.evalHandForTrump(playerIdx, suit)
		if score > bestScore {
			bestScore = score
			bestSuit = suit
		}
	}
	if bestScore >= 5 {
		return bestSuit, true // ゴーアローン
	}
	if bestScore >= 3 {
		return bestSuit, false
	}
	return 0, false
}

// cpuForceCallTrump スタックドディーラー: 強制的にスートを選ぶ
func (e *Euchre) cpuForceCallTrump(playerIdx int) int {
	bestSuit := 0
	bestScore := -1
	for _, suit := range e.availableCallSuits() {
		score := e.evalHandForTrump(playerIdx, suit)
		if score > bestScore {
			bestScore = score
			bestSuit = suit
		}
	}
	if bestSuit == 0 {
		// フォールバック: 使用可能な最初のスート
		suits := e.availableCallSuits()
		if len(suits) > 0 {
			return suits[0]
		}
	}
	return bestSuit
}

// availableCallSuits コールトランプフェーズで選択可能なスートを返す
func (e *Euchre) availableCallSuits() []int {
	allSuits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	var available []int
	// faceUpCard が nil の場合（既にピックアップ済み）は全スート使用可能にはしないが、
	// 通常 CallTrump フェーズでは faceUpCard は残っている
	excludeSuit := -1
	if e.faceUpCard != nil {
		excludeSuit = e.faceUpCard.GetDesign()
	}
	for _, suit := range allSuits {
		if suit != excludeSuit {
			available = append(available, suit)
		}
	}
	return available
}

// evalHandForTrump 指定スートを切り札とした場合のハンド強度を評価する
func (e *Euchre) evalHandForTrump(playerIdx int, trumpSuit int) int {
	player := e.players[playerIdx]
	score := 0

	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)

		// Right Bower
		if card.GetValue() == 11 && card.GetDesign() == trumpSuit {
			score += 2
			continue
		}
		// Left Bower
		if card.GetValue() == 11 && card.GetDesign() == sameColorSuit(trumpSuit) {
			score += 2
			continue
		}
		// 切り札のA, K
		if card.GetDesign() == trumpSuit {
			if card.GetValue() == 1 || card.GetValue() == 13 {
				score++
			} else if card.GetValue() == 12 {
				// Q of trump
				if rand.Intn(2) == 0 {
					score++
				}
			}
			continue
		}
		// 他スートの A
		if card.GetValue() == 1 {
			if rand.Intn(2) == 0 {
				score++
			}
		}
	}
	return score
}

// cpuSelectDiscard CPUが捨てるカードのインデックスを選択する
func (e *Euchre) cpuSelectDiscard(playerIdx int) int {
	player := e.players[playerIdx]
	worstIdx := 0
	worstRank := e.cardRank(player.GetCard(0))

	for i := 1; i < player.GetCardsSize(); i++ {
		rank := e.cardRank(player.GetCard(i))
		if rank < worstRank {
			worstRank = rank
			worstIdx = i
		}
	}
	return worstIdx
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (e *Euchre) cpuSelectPlayCard(playerIdx int) int {
	validIndices := e.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch e.config.CpuDifficulty {
	case EuchreCpuDifficultyHard:
		return e.cpuPlayHard(playerIdx, validIndices)
	case EuchreCpuDifficultyNormal:
		return e.cpuPlayNormal(playerIdx, validIndices)
	default:
		return e.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (e *Euchre) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal 基本戦略プレイ
func (e *Euchre) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := e.players[playerIdx]

	if len(e.currentTrick) == 0 {
		// リード: 最も強いカード
		bestIdx := validIndices[0]
		bestRank := e.cardRank(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			rank := e.cardRank(player.GetCard(idx))
			if rank > bestRank {
				bestRank = rank
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// フォロー: 勝てるなら最小の勝てるカード、勝てないなら最小
	winnerRank := e.currentWinnerRank()
	overCards := []int{}
	for _, idx := range validIndices {
		if e.cardRank(player.GetCard(idx)) > winnerRank {
			overCards = append(overCards, idx)
		}
	}
	if len(overCards) > 0 {
		bestIdx := overCards[0]
		for _, idx := range overCards[1:] {
			if e.cardRank(player.GetCard(idx)) < e.cardRank(player.GetCard(bestIdx)) {
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// 勝てない: 最弱カード
	bestIdx := validIndices[0]
	bestRank := e.cardRank(player.GetCard(validIndices[0]))
	for _, idx := range validIndices[1:] {
		rank := e.cardRank(player.GetCard(idx))
		if rank < bestRank {
			bestRank = rank
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuPlayHard 高度な戦略プレイ
func (e *Euchre) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := e.players[playerIdx]

	if len(e.currentTrick) == 0 {
		// リード: パートナーが強いカードを持っている可能性を考慮
		// 切り札をリードして相手の切り札を引き出す or 強いサイドスートをリード
		bestIdx := validIndices[0]
		bestRank := e.cardRank(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			rank := e.cardRank(player.GetCard(idx))
			if rank > bestRank {
				bestRank = rank
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// 最後のプレイヤーか
	isLastPlayer := len(e.currentTrick) == e.activePlayerCount()-1

	winnerRank := e.currentWinnerRank()
	winnerTeam := e.players[e.currentTrickWinnerIdx()].GetTeam()
	myTeam := e.players[playerIdx].GetTeam()

	// パートナーが勝っている場合
	if winnerTeam == myTeam && isLastPlayer {
		// 最弱カードを出す
		bestIdx := validIndices[0]
		bestRank := e.cardRank(player.GetCard(validIndices[0]))
		for _, idx := range validIndices[1:] {
			rank := e.cardRank(player.GetCard(idx))
			if rank < bestRank {
				bestRank = rank
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// 勝ちに行く
	overCards := []int{}
	for _, idx := range validIndices {
		if e.cardRank(player.GetCard(idx)) > winnerRank {
			overCards = append(overCards, idx)
		}
	}
	if len(overCards) > 0 {
		bestIdx := overCards[0]
		for _, idx := range overCards[1:] {
			if e.cardRank(player.GetCard(idx)) < e.cardRank(player.GetCard(bestIdx)) {
				bestIdx = idx
			}
		}
		return bestIdx
	}

	// 勝てない: 最弱カードを捨てる
	bestIdx := validIndices[0]
	bestRank := e.cardRank(player.GetCard(validIndices[0]))
	for _, idx := range validIndices[1:] {
		rank := e.cardRank(player.GetCard(idx))
		if rank < bestRank {
			bestRank = rank
			bestIdx = idx
		}
	}
	return bestIdx
}

// currentWinnerRank 現在のトリックで最強カードのランクを返す
func (e *Euchre) currentWinnerRank() int {
	if len(e.currentTrick) == 0 {
		return 0
	}
	winnerRank := e.cardRank(e.currentTrick[0].Card)
	leadSuit := e.effectiveSuit(e.currentTrick[0].Card)

	for _, tc := range e.currentTrick[1:] {
		rank := e.cardRank(tc.Card)
		effSuit := e.effectiveSuit(tc.Card)

		if effSuit == e.trumpSuit || effSuit == leadSuit {
			if rank > winnerRank {
				winnerRank = rank
			}
		}
	}
	return winnerRank
}

// currentTrickWinnerIdx 現在のトリック内で暫定勝者のインデックスを返す
func (e *Euchre) currentTrickWinnerIdx() int {
	if len(e.currentTrick) == 0 {
		return 0
	}
	return e.trickWinner()
}

// euchreJSON is the JSON wire format for Euchre.
type euchreJSON struct {
	TrumpCards          *TrumpCards        `json:"tc"`
	Players             []*EuchrePlayer    `json:"ps"`
	Config              EuchreConfig       `json:"cf"`
	Phase               EuchrePhase        `json:"ph"`
	RoundNumber         int                `json:"rn"`
	TrickNumber         int                `json:"tn"`
	CurrentPlayerIdx    int                `json:"ci"`
	CurrentTrick        []*TrickCard       `json:"ct"`
	DealerIdx           int                `json:"di"`
	TrumpSuit           int                `json:"ts"`
	FaceUpCard          *Card              `json:"fu"`
	Kitty               []*Card            `json:"kt"`
	MakerTeam           int                `json:"mt"`
	GoingAlone          bool               `json:"ga"`
	GoingAlonePlayerIdx int                `json:"gi"`
	TeamScores          [EuchreTeamCnt]int `json:"sc"`
	LeadPlayerIdx       int                `json:"li"`
	BidPlayerIdx        int                `json:"bi"`
	GameEndFlag         bool               `json:"ge"`
	WinnerTeam          int                `json:"wt"`
	ActionLog           []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (e *Euchre) MarshalJSON() ([]byte, error) {
	return json.Marshal(euchreJSON{
		TrumpCards:          e.trumpCards,
		Players:             e.players,
		Config:              e.config,
		Phase:               e.phase,
		RoundNumber:         e.roundNumber,
		TrickNumber:         e.trickNumber,
		CurrentPlayerIdx:    e.currentPlayerIdx,
		CurrentTrick:        e.currentTrick,
		DealerIdx:           e.dealerIdx,
		TrumpSuit:           e.trumpSuit,
		FaceUpCard:          e.faceUpCard,
		Kitty:               e.kitty,
		MakerTeam:           e.makerTeam,
		GoingAlone:          e.goingAlone,
		GoingAlonePlayerIdx: e.goingAlonePlayerIdx,
		TeamScores:          e.teamScores,
		LeadPlayerIdx:       e.leadPlayerIdx,
		BidPlayerIdx:        e.bidPlayerIdx,
		GameEndFlag:         e.gameEndFlag,
		WinnerTeam:          e.winnerTeam,
		ActionLog:           e.actionLog,
	})
}

// euchreMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const euchreMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (e *Euchre) UnmarshalJSON(data []byte) error {
	var j euchreJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > euchreMaxSliceLen || len(j.CurrentTrick) > euchreMaxSliceLen ||
		len(j.Kitty) > euchreMaxSliceLen || len(j.ActionLog) > euchreMaxSliceLen {
		return fmt.Errorf("euchre: input array exceeds maximum allowed size")
	}
	e.trumpCards = j.TrumpCards
	if e.trumpCards == nil {
		e.trumpCards = NewTrumpCards(0)
	}
	e.players = j.Players
	if e.players == nil {
		e.players = make([]*EuchrePlayer, 0)
	}
	e.config = j.Config
	e.phase = j.Phase
	e.roundNumber = j.RoundNumber
	e.trickNumber = j.TrickNumber
	e.currentPlayerIdx = j.CurrentPlayerIdx
	e.currentTrick = j.CurrentTrick
	if e.currentTrick == nil {
		e.currentTrick = make([]*TrickCard, 0)
	}
	e.dealerIdx = j.DealerIdx
	e.trumpSuit = j.TrumpSuit
	e.faceUpCard = j.FaceUpCard
	e.kitty = j.Kitty
	if e.kitty == nil {
		e.kitty = make([]*Card, 0)
	}
	e.makerTeam = j.MakerTeam
	e.goingAlone = j.GoingAlone
	e.goingAlonePlayerIdx = j.GoingAlonePlayerIdx
	e.teamScores = j.TeamScores
	e.leadPlayerIdx = j.LeadPlayerIdx
	e.bidPlayerIdx = j.BidPlayerIdx
	e.gameEndFlag = j.GameEndFlag
	e.winnerTeam = j.WinnerTeam
	e.actionLog = j.ActionLog
	if e.actionLog == nil {
		e.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
