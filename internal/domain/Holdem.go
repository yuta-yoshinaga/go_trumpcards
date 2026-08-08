//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// フェーズ定数
const (
	HoldemPhaseInit     = 0 // 初期状態
	HoldemPhasePreFlop  = 1 // プリフロップ
	HoldemPhaseFlop     = 2 // フロップ
	HoldemPhaseTurn     = 3 // ターン
	HoldemPhaseRiver    = 4 // リバー
	HoldemPhaseShowdown = 5 // ショーダウン
	HoldemPhaseEnd      = 6 // ゲーム終了
	HoldemPhaseRebuy    = 7 // リバイ/アドオン待ち
)

// アクション定数 (共通定数のエイリアス)
const (
	HoldemActionFold  = bettingActionFold  // フォールド
	HoldemActionCheck = bettingActionCheck // チェック
	HoldemActionCall  = bettingActionCall  // コール
	HoldemActionBet   = bettingActionBet   // ベット
	HoldemActionRaise = bettingActionRaise // レイズ
	HoldemActionAllIn = bettingActionAllIn // オールイン
)

// holdemDefaultMaxRaises Fixed/PotLimit時のデフォルト最大レイズ回数
const holdemDefaultMaxRaises = bettingMaxRaisesPerRound

// HoldemResult ショーダウン結果
//
// Omaha Hi-Lo (8 or Better) はショーダウン時にハイハンドとローハンドで
// ポットを分割するため、Low* フィールドはオプショナル (Hi-Loルールでない
// 場合は空/0)。`WonAmount` はハイ + ローの合計、`HiWonAmount`/`LowWonAmount`
// は内訳。`LowQualifies` は 8 以下5枚で構成された有効なローが成立したか。
type HoldemResult struct {
	PlayerIdx    int     // プレイヤーインデックス
	HandRank     int     // ハンドランク
	HandName     string  // ハンド名
	BestHand     []*Card // ベスト5枚
	Kickers      []int   // キッカーカード値
	WonAmount    int     // 獲得チップ (Hi-Loでは Hi + Lo の合計)
	Mucked       bool    // マックしたかどうか
	LowQualifies bool    `json:",omitempty"` // 有効なローハンドが成立したか
	LowBestHand  []*Card `json:",omitempty"` // ローベスト5枚
	LowKickers   []int   `json:",omitempty"` // ローキッカーカード値
	HiWonAmount  int     `json:",omitempty"` // ハイサイド獲得チップ
	LowWonAmount int     `json:",omitempty"` // ローサイド獲得チップ
}

// HoldemCpuAction CPU行動記録
type HoldemCpuAction struct {
	PlayerIdx int // プレイヤーインデックス
	Action    int // アクション
	Amount    int // 金額
}

// リバイフェーズ種別定数
const (
	HoldemRebuyPhaseNone  = 0 // なし
	HoldemRebuyPhaseRebuy = 1 // リバイ待ち
	HoldemRebuyPhaseAddon = 2 // アドオン待ち
)

// Holdem テキサスホールデムクラス
type Holdem struct {
	communityCardBettingBase
	tournamentBase  // handCount / rebuyCounts / addonUsed, shared across poker variants (issue #1463)
	trumpCards      *TrumpCards
	players         []*HoldemPlayer
	communityCards  []*Card
	sidePots        []SidePot
	dealerIdx       int
	currentTurn     int
	phase           int
	config          HoldemConfig
	roundResults    []HoldemResult
	cpuActions      []HoldemCpuAction
	startingChips   []int
	vpipTracked     []bool // 当該ハンドでVPIP済みかどうか
	pfrTracked      []bool // 当該ハンドでPFR済みかどうか
	threeBetTracked []bool // 当該ハンドで3Bet追跡済みかどうか
	lastCpuError    error  // CPU行動エラーの最後のフォールバック記録 (テスト検出用)
	rebuyPhaseType  int    // 0=none, 1=rebuy pending, 2=addon pending
	actionLogBase
	humanProfile    *BettingHumanProfile
	lastHumanPlayMs int
}

// NewHoldem コンストラクタ
func NewHoldem(trumpCards *TrumpCards, players []*HoldemPlayer, config HoldemConfig) *Holdem {
	h := &Holdem{
		communityCardBettingBase: communityCardBettingBase{
			actedFlags: make([]bool, len(players)),
		},
		trumpCards:      trumpCards,
		players:         players,
		communityCards:  make([]*Card, 0),
		sidePots:        make([]SidePot, 0),
		roundResults:    make([]HoldemResult, 0),
		cpuActions:      make([]HoldemCpuAction, 0),
		startingChips:   make([]int, len(players)),
		vpipTracked:     make([]bool, len(players)),
		pfrTracked:      make([]bool, len(players)),
		threeBetTracked: make([]bool, len(players)),
		config:          config,
		phase:           HoldemPhaseInit,
	}
	h.initTournamentState(len(players))
	return h
}

// NewDefaultHoldem returns Holdem with the default table size and DefaultHoldemConfig.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultHoldem() *Holdem {
	cfg := DefaultHoldemConfig()
	return NewHoldem(NewTrumpCards(0), NewPlayersForTable(cfg.TableSize), cfg)
}

// Reset ゲーム初期化
func (h *Holdem) Reset() error {
	h.phase = HoldemPhaseInit
	h.pot = 0
	h.sidePots = make([]SidePot, 0)
	h.communityCards = make([]*Card, 0)
	h.gameEndFlag = false
	h.lastBet = 0
	h.minRaise = h.config.BigBlind
	h.raiseCount = 0
	h.actedFlags = make([]bool, len(h.players))
	h.roundResults = make([]HoldemResult, 0)
	h.cpuActions = make([]HoldemCpuAction, 0)
	h.rebuyPhaseType = HoldemRebuyPhaseNone
	h.actionLog = nil
	h.lastHumanPlayMs = 0

	// メタAI: プロファイル初期化
	if h.config.CpuMetaAI {
		if h.humanProfile != nil {
			h.humanProfile.GamesPlayed++
		} else {
			h.humanProfile = &BettingHumanProfile{}
		}
	}

	h.trumpCards.Shuffle()
	for _, p := range h.players {
		p.Reset()
		p.SetFolded(false)
		p.SetAllIn(false)
		p.SetCurrentBet(0)
		p.handRank = 0
		p.bestHand = nil
		if p.GetChips() <= 0 && !h.config.RebuyEnabled {
			p.SetChips(h.config.InitChips)
		}
		p.IncrementTotalHands()
	}

	// HUDスタッツ追跡フラグをリセット
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))

	// トーナメントモード: ブラインドエスカレーション
	if h.config.TournamentMode && h.config.BlindLevelHands > 0 && h.handCount > 0 && h.handCount%h.config.BlindLevelHands == 0 {
		h.config.SmallBlind = h.config.SmallBlind * h.config.BlindMultiplier / 100
		h.config.BigBlind = h.config.BigBlind * h.config.BlindMultiplier / 100
		if h.config.SmallBlind < 1 {
			h.config.SmallBlind = 1
		}
		if h.config.BigBlind < 2 {
			h.config.BigBlind = 2
		}
	}
	h.handCount++

	// リバイチェック: リバイ有効 & リバイ期間内
	if h.config.RebuyEnabled && h.handCount <= h.config.RebuyPeriodHands {
		needHumanRebuy := false
		for i, p := range h.players {
			if p.GetChips() <= 0 && h.rebuyCounts[i] < h.config.RebuyMaxCount {
				if p.GetIsHuman() {
					needHumanRebuy = true
				} else {
					// CPU自動リバイ
					p.AddChips(h.config.RebuyChips)
					h.rebuyCounts[i]++
				}
			}
		}
		if needHumanRebuy {
			h.phase = HoldemPhaseRebuy
			h.rebuyPhaseType = HoldemRebuyPhaseRebuy
			return nil
		}
	}

	// アドオンチェック: アドオン有効 & アドオンハンド番号に到達
	if h.checkAndTransitionAddon() {
		return nil
	}

	return h.continueReset()
}

// continueReset ディール以降のリセット処理 (リバイ/アドオン判定後に実行)
func (h *Holdem) continueReset() error {
	// ハンド開始時のチップを記録 (サイドポット計算用)
	h.startingChips = make([]int, len(h.players))
	for i, p := range h.players {
		h.startingChips[i] = p.GetChips()
	}

	// ブラインド投入
	h.postBlinds()

	// ホールカード配布
	for i := 0; i < 2; i++ {
		for j := 0; j < len(h.players); j++ {
			idx := (h.dealerIdx + 1 + j) % len(h.players)
			card := h.trumpCards.DrawCard()
			if card != nil {
				h.players[idx].AddCard(card)
			}
		}
	}

	h.phase = HoldemPhasePreFlop
	// UTG (ビッグブラインドの次) から開始
	h.currentTurn = (h.dealerIdx + 3) % len(h.players)

	// CPUプリフロップアクション実行
	if err := h.runCpuActions(); err != nil {
		return fmt.Errorf("runCpuActions failed during Reset: %w", err)
	}
	return nil
}

// postBlinds ブラインド投入
func (h *Holdem) postBlinds() {
	sbIdx := (h.dealerIdx + 1) % len(h.players)
	bbIdx := (h.dealerIdx + 2) % len(h.players)

	sbAmount := h.config.SmallBlind
	if h.players[sbIdx].GetChips() < sbAmount {
		sbAmount = h.players[sbIdx].GetChips()
	}
	h.players[sbIdx].SubtractChips(sbAmount)
	h.players[sbIdx].SetCurrentBet(sbAmount)
	h.pot += sbAmount
	h.appendLog(sbIdx, "blind", fmt.Sprintf("posts small blind %d", sbAmount), nil)

	bbAmount := h.config.BigBlind
	if h.players[bbIdx].GetChips() < bbAmount {
		bbAmount = h.players[bbIdx].GetChips()
	}
	h.players[bbIdx].SubtractChips(bbAmount)
	h.players[bbIdx].SetCurrentBet(bbAmount)
	h.pot += bbAmount
	h.appendLog(bbIdx, "blind", fmt.Sprintf("posts big blind %d", bbAmount), nil)

	h.lastBet = bbAmount

	// チップが0になったらオールイン
	if h.players[sbIdx].GetChips() == 0 {
		h.players[sbIdx].SetAllIn(true)
		h.actedFlags[sbIdx] = true
	}
	if h.players[bbIdx].GetChips() == 0 {
		h.players[bbIdx].SetAllIn(true)
		h.actedFlags[bbIdx] = true
	}
}

// PlayerAction 人間プレイヤーのアクション実行
// humanPlayMs: 迷い時間(ms, 0=計測なし)
func (h *Holdem) PlayerAction(action, amount, humanPlayMs int) error {
	if h.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if h.phase < HoldemPhasePreFlop || h.phase > HoldemPhaseRiver {
		return NewDomainError(ErrWrongPhase, "Action is not allowed now.")
	}
	if !h.players[h.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	// メタAI: 人間アクション記録
	h.lastHumanPlayMs = humanPlayMs
	if h.config.CpuMetaAI && h.humanProfile != nil {
		pl := h.players[h.currentTurn]
		handRank := pl.EvalBestHand(h.communityCards)
		h.humanProfile.RecordAction(handRank, action)
		h.humanProfile.RecordHesitation(humanPlayMs)
		if h.lastBet > pl.GetCurrentBet() {
			h.humanProfile.RecordFoldToBet(action == HoldemActionFold)
		}
	}

	err := h.executeAction(h.currentTurn, action, amount)
	if err != nil {
		return err
	}

	h.advanceTurn()
	return h.runCpuActions()
}

// executeAction 指定プレイヤーのアクション実行
func (h *Holdem) executeAction(playerIdx, action, amount int) error {
	// HUDスタッツ追跡 (Holdem固有)
	h.trackPreFlopStats(playerIdx, action)
	h.trackPostFlopStats(playerIdx, action)

	bp := toBettingPlayers(h.players)
	state := h.bettingState()
	maxRaises, maxBetAmount := h.bettingLimits()
	err := ExecuteBettingAction(bp, state, playerIdx, action, amount, h.config.BigBlind, maxRaises, maxBetAmount)
	h.syncBettingState(state)
	if err != nil {
		return err
	}

	h.logAction(playerIdx, action, amount)

	if h.countActivePlayers() == 1 {
		h.resolveLastPlayer()
	}
	return nil
}

// advanceTurn 次のプレイヤーに進める
func (h *Holdem) advanceTurn() {
	if h.gameEndFlag {
		return
	}

	bp := toBettingPlayers(h.players)
	if h.isBettingRoundComplete(bp) {
		h.advancePhase()
		return
	}

	if next := h.findNextActiveTurn(h.currentTurn, bp); next >= 0 {
		h.currentTurn = next
		return
	}

	h.advancePhase()
}

// advancePhase 次のフェーズに進める
func (h *Holdem) advancePhase() {
	// ラウンドベットリセット
	for _, p := range h.players {
		p.SetCurrentBet(0)
	}
	h.lastBet = 0
	h.minRaise = h.config.BigBlind
	h.raiseCount = 0
	h.actedFlags = make([]bool, len(h.players))
	// フォールド・オールインプレイヤーはacted扱い
	for i, p := range h.players {
		if p.GetFolded() || p.GetAllIn() {
			h.actedFlags[i] = true
		}
	}

	switch h.phase {
	case HoldemPhasePreFlop:
		h.phase = HoldemPhaseFlop
		for i := 0; i < 3; i++ {
			card := h.trumpCards.DrawCard()
			if card != nil {
				h.communityCards = append(h.communityCards, card)
			}
		}
		h.appendLog(-1, "deal", "dealt flop", h.communityCards)
	case HoldemPhaseFlop:
		h.phase = HoldemPhaseTurn
		card := h.trumpCards.DrawCard()
		if card != nil {
			h.communityCards = append(h.communityCards, card)
		}
		h.appendLog(-1, "deal", "dealt turn", h.communityCards[3:])
	case HoldemPhaseTurn:
		h.phase = HoldemPhaseRiver
		card := h.trumpCards.DrawCard()
		if card != nil {
			h.communityCards = append(h.communityCards, card)
		}
		h.appendLog(-1, "deal", "dealt river", h.communityCards[4:])
	case HoldemPhaseRiver:
		h.phase = HoldemPhaseShowdown
		h.appendLog(-1, "showdown", "showdown", nil)
		h.resolveShowdown()
		return
	}

	// アクティブプレイヤーが0-1人ならショーダウンへ
	activeCnt := 0
	for _, p := range h.players {
		if !p.GetFolded() && !p.GetAllIn() {
			activeCnt++
		}
	}
	if activeCnt <= 1 {
		// 全員オールインまたはフォールド → 残りのコミュニティカードを配ってショーダウン
		h.dealRemainingCommunity()
		h.phase = HoldemPhaseShowdown
		h.resolveShowdown()
		return
	}

	// ディーラーの次のアクティブプレイヤーから開始
	h.currentTurn = h.findNextActive(h.dealerIdx)
}

// dealRemainingCommunity 残りのコミュニティカードを全て配る
func (h *Holdem) dealRemainingCommunity() {
	for len(h.communityCards) < 5 {
		card := h.trumpCards.DrawCard()
		if card == nil {
			break
		}
		h.communityCards = append(h.communityCards, card)
	}
}

// findNextActive 指定インデックスの次のアクティブ (フォールド・オールインでない) プレイヤーを探す
func (h *Holdem) findNextActive(fromIdx int) int {
	for i := 1; i <= len(h.players); i++ {
		next := (fromIdx + i) % len(h.players)
		if !h.players[next].GetFolded() && !h.players[next].GetAllIn() {
			return next
		}
	}
	return (fromIdx + 1) % len(h.players)
}

// countActivePlayers フォールドしていないプレイヤー数を返す
func (h *Holdem) countActivePlayers() int {
	cnt := 0
	for _, p := range h.players {
		if !p.GetFolded() {
			cnt++
		}
	}
	return cnt
}

// bettingLimits ベッティングリミット設定からmaxRaisesとmaxBetAmountを計算
func (h *Holdem) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(h.config.BettingLimit, h.pot, h.lastBet)
}

// boardTexture ボードテクスチャ分析結果
type boardTexture struct {
	paired       bool // ペアボード (同じ数字が2枚以上)
	flushDraw    bool // フラッシュドロー (同スート3枚以上)
	straightDraw bool // ストレートドロー (5幅ウィンドウ内に3枚以上)
	wet          bool // ウェットボード (flushDraw || straightDraw)
	highCards    int  // ハイカード数 (10以上 or A)
}

// evalBoardTexture コミュニティカードからボードテクスチャを分析
func evalBoardTexture(communityCards []*Card) boardTexture {
	bt := boardTexture{}
	if len(communityCards) == 0 {
		return bt
	}

	// ペア判定
	valCount := make(map[int]int)
	suitCount := make(map[int]int)
	for _, c := range communityCards {
		valCount[c.GetValue()]++
		suitCount[c.GetDesign()]++
	}
	for _, cnt := range valCount {
		if cnt >= 2 {
			bt.paired = true
			break
		}
	}

	// フラッシュドロー判定
	for _, cnt := range suitCount {
		if cnt >= 3 {
			bt.flushDraw = true
			break
		}
	}

	// ストレートドロー判定: 5幅ウィンドウ内に3枚以上
	vals := make(map[int]bool)
	for _, c := range communityCards {
		v := c.GetValue()
		vals[v] = true
		if v == 1 {
			vals[14] = true // A=14としても扱う
		}
	}
	for base := 1; base <= 10; base++ {
		cnt := 0
		for v := base; v < base+5; v++ {
			if vals[v] {
				cnt++
			}
		}
		if cnt >= 3 {
			bt.straightDraw = true
			break
		}
	}

	bt.wet = bt.flushDraw || bt.straightDraw

	// ハイカード数 (10以上 or A)
	for _, c := range communityCards {
		v := c.GetValue()
		if v >= 10 || v == 1 {
			bt.highCards++
		}
	}

	return bt
}

// clamp 値を範囲内に収める
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// GetEquity エクイティ計算結果を返す (PreFlop-Riverフェーズで人間がフォールドしていない場合のみ)
func (h *Holdem) GetEquity() *HoldemEquityResult {
	if h.phase < HoldemPhasePreFlop || h.phase > HoldemPhaseRiver {
		return nil
	}
	// 人間プレイヤーを探す
	var humanPlayer *HoldemPlayer
	for _, p := range h.players {
		if p.GetIsHuman() {
			humanPlayer = p
			break
		}
	}
	if humanPlayer == nil || humanPlayer.GetFolded() {
		return nil
	}
	humanCards := make([]*Card, humanPlayer.GetCardsSize())
	for i := 0; i < humanPlayer.GetCardsSize(); i++ {
		humanCards[i] = humanPlayer.GetCard(i)
	}
	// フォールドしていないアクティブ相手プレイヤー数
	activePlayers := 0
	for _, p := range h.players {
		if !p.GetIsHuman() && !p.GetFolded() {
			activePlayers++
		}
	}
	result := CalcEquity(humanCards, h.communityCards, activePlayers, holdemEquitySimulations, nil)
	return &result
}

// GetPotOdds ポットオッズを返す (PreFlop-Riverフェーズのみ)
func (h *Holdem) GetPotOdds() float64 {
	if h.phase < HoldemPhasePreFlop || h.phase > HoldemPhaseRiver {
		return 0.0
	}
	// 人間プレイヤーの現在ベットを取得
	humanCurrentBet := 0
	for _, p := range h.players {
		if p.GetIsHuman() {
			humanCurrentBet = p.GetCurrentBet()
			break
		}
	}
	callAmount := h.lastBet - humanCurrentBet
	if callAmount < 0 {
		callAmount = 0
	}
	return CalcPotOdds(h.pot, callAmount)
}

// --- ゲッター ---

// GetPhase フェーズ取得
func (h *Holdem) GetPhase() int { return h.phase }

// GetPlayers プレイヤー一覧取得
func (h *Holdem) GetPlayers() []*HoldemPlayer { return h.players }

// GetPlayer 指定プレイヤー取得
func (h *Holdem) GetPlayer(i int) *HoldemPlayer {
	if i >= 0 && i < len(h.players) {
		return h.players[i]
	}
	return nil
}

// GetPlayerCnt プレイヤー数取得
func (h *Holdem) GetPlayerCnt() int { return len(h.players) }

// GetCommunityCards コミュニティカード取得
func (h *Holdem) GetCommunityCards() []*Card { return h.communityCards }

// GetPot ポット取得
func (h *Holdem) GetPot() int { return h.pot }

// GetSidePots サイドポット取得
func (h *Holdem) GetSidePots() []SidePot { return h.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (h *Holdem) GetDealerIdx() int { return h.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (h *Holdem) GetCurrentTurn() int { return h.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (h *Holdem) GetGameEndFlag() bool { return h.gameEndFlag }

// GetLastBet 最後のベット取得
func (h *Holdem) GetLastBet() int { return h.lastBet }

// GetMinRaise 最小レイズ額取得
func (h *Holdem) GetMinRaise() int { return h.minRaise }

// GetRaiseCount 現在のレイズ回数取得
func (h *Holdem) GetRaiseCount() int { return h.raiseCount }

// GetRoundResults ラウンド結果取得
func (h *Holdem) GetRoundResults() []HoldemResult { return h.roundResults }

// GetCpuActions CPU行動記録取得
func (h *Holdem) GetCpuActions() []HoldemCpuAction { return h.cpuActions }

// GetLastCpuError 最後のCPUアクションエラー取得 (テスト・デバッグ用)
func (h *Holdem) GetLastCpuError() error { return h.lastCpuError }

// GetHumanProfile メタAIプロファイル取得
func (h *Holdem) GetHumanProfile() *BettingHumanProfile { return h.humanProfile }

// ResetProfile メタAIプロファイルをリセットする
func (h *Holdem) ResetProfile() { h.humanProfile = nil }

// ExportProfile メタAIプロファイルをエクスポートする (プロファイルがない場合はnil)
func (h *Holdem) ExportProfile() interface{} {
	if h.humanProfile == nil {
		return nil
	}
	d := h.humanProfile.Export()
	return &d
}

// ImportProfile JSONバイトからメタAIプロファイルをインポートする
func (h *Holdem) ImportProfile(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	d, err := ImportBettingHumanProfileJSON(data)
	if err != nil {
		return err
	}
	h.humanProfile = &BettingHumanProfile{}
	h.humanProfile.Import(d)
	return nil
}

// GetConfig 設定取得
func (h *Holdem) GetConfig() HoldemConfig { return h.config }

// SetConfig 設定変更
func (h *Holdem) SetConfig(cfg HoldemConfig) { h.config = cfg }

// IsHumanTurn 人間のターンかチェック
func (h *Holdem) IsHumanTurn() bool {
	if h.currentTurn >= 0 && h.currentTurn < len(h.players) {
		return h.players[h.currentTurn].GetIsHuman()
	}
	return false
}

// GetActedFlags actedフラグ取得
func (h *Holdem) GetActedFlags() []bool {
	result := make([]bool, len(h.actedFlags))
	copy(result, h.actedFlags)
	return result
}

// GetHandCount ハンド数取得
func (h *Holdem) GetHandCount() int { return h.handCount }

// logAction ベッティングアクションを棋譜に記録する
func (h *Holdem) logAction(playerIdx, action, amount int) {
	switch action {
	case HoldemActionFold:
		h.appendLog(playerIdx, "fold", "fold", nil)
	case HoldemActionCheck:
		h.appendLog(playerIdx, "check", "check", nil)
	case HoldemActionCall:
		h.appendLog(playerIdx, "call", fmt.Sprintf("call %d", h.players[playerIdx].GetCurrentBet()), nil)
	case HoldemActionBet:
		h.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", amount), nil)
	case HoldemActionRaise:
		h.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", amount), nil)
	case HoldemActionAllIn:
		h.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", h.players[playerIdx].GetCurrentBet()), nil)
	}
}

// holdemJSON is the JSON wire format for Holdem.
type holdemJSON struct {
	TrumpCards      *TrumpCards              `json:"tc"`
	Players         []*HoldemPlayer          `json:"pl"`
	CommunityCards  []*Card                  `json:"cc"`
	Pot             int                      `json:"pt"`
	SidePots        []SidePot                `json:"sp"`
	DealerIdx       int                      `json:"di"`
	CurrentTurn     int                      `json:"ct"`
	Phase           int                      `json:"ph"`
	Config          HoldemConfig             `json:"cf"`
	GameEndFlag     bool                     `json:"ge"`
	LastBet         int                      `json:"lb"`
	MinRaise        int                      `json:"mr"`
	RaiseCount      int                      `json:"rc"`
	ActedFlags      []bool                   `json:"af"`
	RoundResults    []HoldemResult           `json:"rr"`
	CpuActions      []HoldemCpuAction        `json:"ca"`
	StartingChips   []int                    `json:"sc"`
	VPIPTracked     []bool                   `json:"vt"`
	PFRTracked      []bool                   `json:"ft"`
	ThreeBetTracked []bool                   `json:"tt"`
	HandCount       int                      `json:"hc"`
	RebuyCounts     []int                    `json:"rb"`
	AddonUsed       []bool                   `json:"au"`
	RebuyPhaseType  int                      `json:"rp"`
	ActionLog       []*ActionLogEntry        `json:"al"`
	Profile         *BettingHumanProfileData `json:"pf,omitempty"`
	LastHumanPlayMs int                      `json:"hm"`
}

// holdemMaxSliceLen caps slice sizes during deserialisation.
const holdemMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (h *Holdem) MarshalJSON() ([]byte, error) {
	j := holdemJSON{
		TrumpCards:      h.trumpCards,
		Players:         h.players,
		CommunityCards:  h.communityCards,
		Pot:             h.pot,
		SidePots:        h.sidePots,
		DealerIdx:       h.dealerIdx,
		CurrentTurn:     h.currentTurn,
		Phase:           h.phase,
		Config:          h.config,
		GameEndFlag:     h.gameEndFlag,
		LastBet:         h.lastBet,
		MinRaise:        h.minRaise,
		RaiseCount:      h.raiseCount,
		ActedFlags:      h.actedFlags,
		RoundResults:    h.roundResults,
		CpuActions:      h.cpuActions,
		StartingChips:   h.startingChips,
		VPIPTracked:     h.vpipTracked,
		PFRTracked:      h.pfrTracked,
		ThreeBetTracked: h.threeBetTracked,
		HandCount:       h.handCount,
		RebuyCounts:     h.rebuyCounts,
		AddonUsed:       h.addonUsed,
		RebuyPhaseType:  h.rebuyPhaseType,
		ActionLog:       h.actionLog,
		LastHumanPlayMs: h.lastHumanPlayMs,
	}
	if h.humanProfile != nil {
		d := h.humanProfile.Export()
		j.Profile = &d
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (h *Holdem) UnmarshalJSON(data []byte) error {
	var j holdemJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > holdemMaxSliceLen || len(j.CommunityCards) > holdemMaxSliceLen ||
		len(j.SidePots) > holdemMaxSliceLen || len(j.ActedFlags) > holdemMaxSliceLen ||
		len(j.RoundResults) > holdemMaxSliceLen || len(j.CpuActions) > holdemMaxSliceLen ||
		len(j.StartingChips) > holdemMaxSliceLen || len(j.ActionLog) > holdemMaxSliceLen {
		return fmt.Errorf("holdem: input array exceeds maximum allowed size")
	}
	h.trumpCards = j.TrumpCards
	if h.trumpCards == nil {
		h.trumpCards = NewTrumpCards(0)
	}
	h.players = j.Players
	if h.players == nil {
		h.players = make([]*HoldemPlayer, 0)
	}
	h.communityCards = j.CommunityCards
	if h.communityCards == nil {
		h.communityCards = make([]*Card, 0)
	}
	h.pot = j.Pot
	h.sidePots = j.SidePots
	if h.sidePots == nil {
		h.sidePots = make([]SidePot, 0)
	}
	h.dealerIdx = j.DealerIdx
	h.currentTurn = j.CurrentTurn
	h.phase = j.Phase
	h.config = j.Config
	h.gameEndFlag = j.GameEndFlag
	h.lastBet = j.LastBet
	h.minRaise = j.MinRaise
	h.raiseCount = j.RaiseCount
	h.actedFlags = j.ActedFlags
	if h.actedFlags == nil {
		h.actedFlags = make([]bool, 0)
	}
	h.roundResults = j.RoundResults
	if h.roundResults == nil {
		h.roundResults = make([]HoldemResult, 0)
	}
	h.cpuActions = j.CpuActions
	if h.cpuActions == nil {
		h.cpuActions = make([]HoldemCpuAction, 0)
	}
	h.startingChips = j.StartingChips
	if h.startingChips == nil {
		h.startingChips = make([]int, 0)
	}
	h.vpipTracked = j.VPIPTracked
	if h.vpipTracked == nil {
		h.vpipTracked = make([]bool, 0)
	}
	h.pfrTracked = j.PFRTracked
	if h.pfrTracked == nil {
		h.pfrTracked = make([]bool, 0)
	}
	h.threeBetTracked = j.ThreeBetTracked
	if h.threeBetTracked == nil {
		h.threeBetTracked = make([]bool, 0)
	}
	h.handCount = j.HandCount
	h.rebuyCounts = j.RebuyCounts
	if h.rebuyCounts == nil {
		h.rebuyCounts = make([]int, 0)
	}
	h.addonUsed = j.AddonUsed
	if h.addonUsed == nil {
		h.addonUsed = make([]bool, 0)
	}
	h.rebuyPhaseType = j.RebuyPhaseType
	h.actionLog = j.ActionLog
	if h.actionLog == nil {
		h.actionLog = make([]*ActionLogEntry, 0)
	}
	h.lastHumanPlayMs = j.LastHumanPlayMs
	if j.Profile != nil {
		h.humanProfile = &BettingHumanProfile{}
		h.humanProfile.Import(*j.Profile)
	}
	return nil
}

// Resize プレイヤースライスを差し替え、プレイヤー数依存スライスを再初期化する
func (h *Holdem) Resize(players []*HoldemPlayer) {
	h.players = players
	n := len(players)
	h.actedFlags = make([]bool, n)
	h.startingChips = make([]int, n)
	h.vpipTracked = make([]bool, n)
	h.pfrTracked = make([]bool, n)
	h.threeBetTracked = make([]bool, n)
	h.initTournamentState(n)
}
