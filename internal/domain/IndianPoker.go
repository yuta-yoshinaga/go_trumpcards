//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// フェーズ定数
const (
	IndianPokerPhaseInit     = 0 // 初期状態
	IndianPokerPhaseAnte     = 1 // アンティ投入済み
	IndianPokerPhaseBetting  = 2 // ベッティングラウンド
	IndianPokerPhaseShowdown = 3 // ショーダウン
	IndianPokerPhaseEnd      = 4 // ゲーム終了
)

// アクション定数 (共通定数のエイリアス)
const (
	IndianPokerActionFold  = bettingActionFold  // フォールド
	IndianPokerActionCheck = bettingActionCheck // チェック
	IndianPokerActionCall  = bettingActionCall  // コール
	IndianPokerActionBet   = bettingActionBet   // ベット
	IndianPokerActionRaise = bettingActionRaise // レイズ
	IndianPokerActionAllIn = bettingActionAllIn // オールイン
)

// IndianPokerResult ショーダウン結果
type IndianPokerResult struct {
	PlayerIdx int   // プレイヤーインデックス
	Card      *Card // 公開カード
	CardRank  int   // カードランク (2-14, Ace=14)
	WonAmount int   // 獲得チップ
}

// IndianPokerCpuAction CPU行動記録
type IndianPokerCpuAction struct {
	PlayerIdx int // プレイヤーインデックス
	Action    int // アクション
	Amount    int // 金額
}

// IndianPoker インディアンポーカークラス
type IndianPoker struct {
	trumpCards    *TrumpCards
	players       []*IndianPokerPlayer
	pot           int
	sidePots      []SidePot
	dealerIdx     int
	currentTurn   int
	phase         int
	config        IndianPokerConfig
	gameEndFlag   bool
	lastBet       int
	minRaise      int
	raiseCount    int
	actedFlags    []bool
	roundResults  []IndianPokerResult
	cpuActions    []IndianPokerCpuAction
	startingChips []int
	handCount     int
	lastCpuError  error
	actionLogBase
	humanProfile    *IndianPokerHumanProfile
	lastHumanPlayMs int
}

// NewIndianPoker コンストラクタ
func NewIndianPoker(trumpCards *TrumpCards, players []*IndianPokerPlayer, config IndianPokerConfig) *IndianPoker {
	return &IndianPoker{
		trumpCards:    trumpCards,
		players:       players,
		sidePots:      make([]SidePot, 0),
		actedFlags:    make([]bool, len(players)),
		roundResults:  make([]IndianPokerResult, 0),
		cpuActions:    make([]IndianPokerCpuAction, 0),
		startingChips: make([]int, len(players)),
		config:        config,
		phase:         IndianPokerPhaseInit,
	}
}

// NewDefaultIndianPoker returns IndianPoker with the standard player setup and
// DefaultIndianPokerConfig. Used as the single source of truth for CUI, Web,
// and Worker construction sites.
func NewDefaultIndianPoker() *IndianPoker {
	return NewIndianPoker(NewTrumpCards(0), NewIndianPokerPlayers(), DefaultIndianPokerConfig())
}

// Reset ゲーム初期化
func (ip *IndianPoker) Reset() error {
	ip.phase = IndianPokerPhaseInit
	ip.pot = 0
	ip.sidePots = make([]SidePot, 0)
	ip.gameEndFlag = false
	ip.lastBet = 0
	ip.minRaise = ip.config.Ante
	ip.raiseCount = 0
	ip.actedFlags = make([]bool, len(ip.players))
	ip.roundResults = make([]IndianPokerResult, 0)
	ip.cpuActions = make([]IndianPokerCpuAction, 0)
	ip.actionLog = nil
	ip.lastHumanPlayMs = 0

	// メタAI: プロファイル初期化
	if ip.config.CpuMetaAI {
		if ip.humanProfile != nil {
			ip.humanProfile.GamesPlayed++
		} else {
			ip.humanProfile = &IndianPokerHumanProfile{}
		}
	}

	ip.trumpCards.Shuffle()
	for _, p := range ip.players {
		p.Reset()
		p.SetFolded(false)
		p.SetAllIn(false)
		p.SetCurrentBet(0)
		p.handRank = 0
		if p.GetChips() <= 0 {
			p.SetChips(ip.config.InitChips)
		}
	}

	ip.handCount++

	// ハンド開始時のチップを記録 (サイドポット計算用)
	ip.startingChips = make([]int, len(ip.players))
	for i, p := range ip.players {
		ip.startingChips[i] = p.GetChips()
	}

	// アンティ投入
	ip.postAntes()

	// カード配布 (1枚ずつ)
	for i := 0; i < len(ip.players); i++ {
		idx := (ip.dealerIdx + 1 + i) % len(ip.players)
		card := ip.trumpCards.DrawCard()
		if card != nil {
			ip.players[idx].AddCard(card)
		}
	}

	// ハンドランク設定
	for _, p := range ip.players {
		if p.GetCardsSize() > 0 {
			p.SetHandRank(indianPokerCardRank(p.GetCard(0)))
		}
	}

	ip.phase = IndianPokerPhaseBetting
	// ディーラーの左 (ディーラー+1) から開始
	ip.currentTurn = (ip.dealerIdx + 1) % len(ip.players)

	// CPUアクション実行
	if err := ip.runCpuActions(); err != nil {
		return fmt.Errorf("runCpuActions failed during Reset: %w", err)
	}
	return nil
}

// postAntes アンティ投入
func (ip *IndianPoker) postAntes() {
	ip.phase = IndianPokerPhaseAnte
	for i, p := range ip.players {
		anteAmount := min(ip.config.Ante, p.GetChips())
		p.SubtractChips(anteAmount)
		p.SetCurrentBet(anteAmount)
		ip.pot += anteAmount
		ip.appendLog(i, "ante", fmt.Sprintf("posts ante %d", anteAmount), nil)

		// チップが0になったらオールイン
		if p.GetChips() == 0 {
			p.SetAllIn(true)
			ip.actedFlags[i] = true
		}
	}
	ip.lastBet = ip.config.Ante
}

// PlayerAction 人間プレイヤーのアクション実行
// humanPlayMs: 迷い時間(ms, 0=計測なし)
func (ip *IndianPoker) PlayerAction(action, amount, humanPlayMs int) error {
	if ip.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if ip.phase != IndianPokerPhaseBetting {
		return NewDomainError(ErrWrongPhase, "Action is not allowed now.")
	}
	if !ip.players[ip.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	// メタAI: 人間アクション記録 (CPUは人間のカードが見える)
	ip.lastHumanPlayMs = humanPlayMs
	if ip.config.CpuMetaAI && ip.humanProfile != nil {
		pl := ip.players[ip.currentTurn]
		cardRank := indianPokerCardRank(pl.GetCard(0))
		ip.humanProfile.RecordAction(cardRank, action)
		ip.humanProfile.RecordHesitation(humanPlayMs)
		if ip.lastBet > pl.GetCurrentBet() {
			ip.humanProfile.RecordFoldToBet(action == IndianPokerActionFold)
		}
	}

	err := ip.executeAction(ip.currentTurn, action, amount)
	if err != nil {
		return err
	}

	ip.advanceTurn()
	return ip.runCpuActions()
}

// bettingPlayers BettingPlayerスライスを生成
func (ip *IndianPoker) bettingPlayers() []BettingPlayer {
	bp := make([]BettingPlayer, len(ip.players))
	for i, pl := range ip.players {
		bp[i] = pl
	}
	return bp
}

// executeAction 指定プレイヤーのアクション実行
func (ip *IndianPoker) executeAction(playerIdx, action, amount int) error {
	bp := ip.bettingPlayers()
	state := &BettingState{
		Pot: ip.pot, LastBet: ip.lastBet, MinRaise: ip.minRaise,
		RaiseCount: ip.raiseCount, ActedFlags: ip.actedFlags,
	}
	maxRaises, maxBetAmount := ip.bettingLimits()
	err := ExecuteBettingAction(bp, state, playerIdx, action, amount, ip.config.Ante, maxRaises, maxBetAmount)
	ip.pot = state.Pot
	ip.lastBet = state.LastBet
	ip.minRaise = state.MinRaise
	ip.raiseCount = state.RaiseCount
	if err != nil {
		return err
	}

	// 棋譜記録
	ip.logAction(playerIdx, action, amount)

	// フォールドでアクティブプレイヤーが1人になったらチェック
	if ip.countActivePlayers() == 1 {
		ip.resolveLastPlayer()
	}
	return nil
}

// advanceTurn 次のプレイヤーに進める
func (ip *IndianPoker) advanceTurn() {
	if ip.gameEndFlag {
		return
	}

	// ベッティングラウンド終了チェック
	if ip.isBettingRoundComplete() {
		ip.phase = IndianPokerPhaseShowdown
		ip.appendLog(-1, "showdown", "showdown", nil)
		ip.resolveShowdown()
		return
	}

	// 次のアクティブプレイヤーを探す
	for i := 1; i <= len(ip.players); i++ {
		next := (ip.currentTurn + i) % len(ip.players)
		if !ip.players[next].GetFolded() && !ip.players[next].GetAllIn() && !ip.actedFlags[next] {
			ip.currentTurn = next
			return
		}
	}

	// 全員行動済みならショーダウン
	ip.phase = IndianPokerPhaseShowdown
	ip.appendLog(-1, "showdown", "showdown", nil)
	ip.resolveShowdown()
}

// isBettingRoundComplete ベッティングラウンドが完了したかチェック
func (ip *IndianPoker) isBettingRoundComplete() bool {
	for i, p := range ip.players {
		if p.GetFolded() || p.GetAllIn() {
			continue
		}
		if !ip.actedFlags[i] {
			return false
		}
	}
	return true
}

// bettingLimits ベッティングリミット設定からmaxRaisesとmaxBetAmountを計算
func (ip *IndianPoker) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(ip.config.BettingLimit, ip.pot, ip.lastBet)
}

// countActivePlayers フォールドしていないプレイヤー数を返す
func (ip *IndianPoker) countActivePlayers() int {
	cnt := 0
	for _, p := range ip.players {
		if !p.GetFolded() {
			cnt++
		}
	}
	return cnt
}

// resolveLastPlayer 最後の1人が残った場合のポット配分
func (ip *IndianPoker) resolveLastPlayer() {
	for i, p := range ip.players {
		if !p.GetFolded() {
			p.AddChips(ip.pot)
			ip.roundResults = []IndianPokerResult{{
				PlayerIdx: i,
				WonAmount: ip.pot,
			}}
			if p.GetCardsSize() > 0 {
				ip.roundResults[0].Card = p.GetCard(0)
				ip.roundResults[0].CardRank = indianPokerCardRank(p.GetCard(0))
			}
			ip.pot = 0
			break
		}
	}
	ip.phase = IndianPokerPhaseEnd
	ip.gameEndFlag = true
	ip.dealerIdx = (ip.dealerIdx + 1) % len(ip.players)
}

// resolveShowdown ショーダウン: カード比較・ポット配分
func (ip *IndianPoker) resolveShowdown() {
	// サイドポット計算・配分 (スートタイブレーク付き勝者判定)
	bp := ip.bettingPlayers()
	ip.sidePots = CalculateSidePots(bp, ip.pot, ip.startingChips)
	wonAmounts := DistributePotsWithWinnerFunc(bp, ip.sidePots, indianPokerFindWinners)

	// 結果を構築
	ip.roundResults = make([]IndianPokerResult, 0)
	for i, p := range ip.players {
		if p.GetFolded() {
			continue
		}
		result := IndianPokerResult{
			PlayerIdx: i,
			WonAmount: wonAmounts[i],
		}
		if p.GetCardsSize() > 0 {
			result.Card = p.GetCard(0)
			result.CardRank = indianPokerCardRank(p.GetCard(0))
		}
		ip.roundResults = append(ip.roundResults, result)
	}

	ip.phase = IndianPokerPhaseEnd
	ip.gameEndFlag = true
	ip.dealerIdx = (ip.dealerIdx + 1) % len(ip.players)
}

// indianPokerCardRank カードランクを返す (Ace=14, それ以外=額面値)
func indianPokerCardRank(c *Card) int {
	v := c.GetValue()
	if v == 1 {
		return 14
	}
	return v
}

// indianPokerSuitRank スートランクを返す (Spade=4 > Heart=3 > Diamond=2 > Club=1)
func indianPokerSuitRank(c *Card) int {
	switch c.GetDesign() {
	case CardDesignSpade:
		return 4
	case CardDesignHeart:
		return 3
	case CardDesignDiamond:
		return 2
	default: // Clover
		return 1
	}
}

// indianPokerFindWinners インディアンポーカー用の勝者判定 (カードランク比較、同値時はスートタイブレーク)
func indianPokerFindWinners(players []BettingPlayer, eligible []int) []int {
	bestRank := -1
	bestSuit := -1
	var winners []int

	for _, idx := range eligible {
		pl := players[idx]
		if pl.GetFolded() {
			continue
		}
		cards := pl.GetComparisonCards()
		if len(cards) == 0 {
			continue
		}
		rank := indianPokerCardRank(cards[0])
		suit := indianPokerSuitRank(cards[0])

		if rank > bestRank || (rank == bestRank && suit > bestSuit) {
			bestRank = rank
			bestSuit = suit
			winners = []int{idx}
		}
	}

	return winners
}

// --- CPU AI ---

// indianPokerCpuStyleParams CPUスタイル別パラメータ
type indianPokerCpuStyleParams struct {
	aggressive     bool // Aggressive or Passive
	bluffRate      int  // ブラフ率(%)
	foldThreshold  int  // 推定強度 < this → fold評価
	raiseThreshold int  // 推定強度 >= this → raise
	raisePotPct    int  // raise = pot * this / 100
}

// indianPokerStyleParamsMap スタイルごとのパラメータ
var indianPokerStyleParamsMap = map[HoldemPlayStyle]indianPokerCpuStyleParams{
	HoldemStyleTAG: {
		aggressive: true, bluffRate: 15,
		foldThreshold: 45, raiseThreshold: 70, raisePotPct: 75,
	},
	HoldemStyleLAP: {
		aggressive: false, bluffRate: 5,
		foldThreshold: 20, raiseThreshold: 80, raisePotPct: 50,
	},
	HoldemStyleTAP: {
		aggressive: false, bluffRate: 5,
		foldThreshold: 35, raiseThreshold: 85, raisePotPct: 50,
	},
	HoldemStyleLAG: {
		aggressive: true, bluffRate: 30,
		foldThreshold: 20, raiseThreshold: 50, raisePotPct: 100,
	},
}

// runCpuActions CPUプレイヤーのアクションを実行
func (ip *IndianPoker) runCpuActions() error {
	if ip.gameEndFlag {
		return nil
	}
	const maxIterations = 100
	iterations := 0
	for !ip.gameEndFlag && ip.phase == IndianPokerPhaseBetting {
		iterations++
		if iterations > maxIterations {
			return fmt.Errorf("maxIterations reached in runCpuActions, possible infinite loop")
		}
		if ip.players[ip.currentTurn].GetIsHuman() {
			return nil
		}
		if ip.players[ip.currentTurn].GetFolded() || ip.players[ip.currentTurn].GetAllIn() {
			ip.advanceTurn()
			continue
		}
		action, amount := ip.cpuDecide(ip.currentTurn)
		ip.cpuActions = append(ip.cpuActions, IndianPokerCpuAction{
			PlayerIdx: ip.currentTurn,
			Action:    action,
			Amount:    amount,
		})
		err := ip.executeAction(ip.currentTurn, action, amount)
		if err != nil {
			ip.handleCpuActionError(ip.currentTurn, action, err)
		}
		if ip.gameEndFlag {
			return nil
		}
		ip.advanceTurn()
	}
	return nil
}

// handleCpuActionError CPUアクション失敗時のフォールバック処理
func (ip *IndianPoker) handleCpuActionError(playerIdx, action int, err error) {
	ip.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", playerIdx, action, err)
	callAmt := ip.lastBet - ip.players[playerIdx].GetCurrentBet()
	if callAmt > 0 {
		_ = ip.executeAction(playerIdx, IndianPokerActionFold, 0)
	} else {
		_ = ip.executeAction(playerIdx, IndianPokerActionCheck, 0)
	}
}

// cpuDecide CPUプレイヤーの意思決定
func (ip *IndianPoker) cpuDecide(idx int) (int, int) {
	p := ip.players[idx]
	callAmount := ip.lastBet - p.GetCurrentBet()

	params, ok := indianPokerStyleParamsMap[p.GetPlayStyle()]
	if !ok {
		return CpuCallOrCheck(callAmount)
	}

	// 推定自分のカード強度 (0-100) を計算
	strength := ip.estimateOwnStrength(idx)

	// メタAI: ブラフ率を調整
	if ip.config.CpuMetaAI && ip.humanProfile != nil {
		adjusted := ip.humanProfile.AdjustedBluffChance(float64(params.bluffRate))
		params.bluffRate = int(math.Round(adjusted))
	}

	var action, amount int

	// Fold評価
	if strength < params.foldThreshold {
		if params.aggressive {
			// Aggressive: ブラフの可能性
			if rand.Intn(100) < params.bluffRate { //nolint:gosec // non-crypto random for game AI
				betAmt := ip.cpuPotBet(params.raisePotPct)
				action, amount = CpuRaiseOrBet(p.GetChips(), callAmount, betAmt)
			} else {
				action, amount = CpuFoldOrCheck(callAmount)
			}
		} else {
			action, amount = CpuFoldOrCheck(callAmount)
		}
	} else if strength >= params.raiseThreshold || (params.aggressive && rand.Intn(100) < params.bluffRate) { //nolint:gosec // non-crypto random for game AI
		// Raise
		betAmt := ip.cpuPotBet(params.raisePotPct)
		action, amount = CpuRaiseOrBet(p.GetChips(), callAmount, betAmt)
	} else {
		// Call or Check
		action, amount = CpuCallOrCheck(callAmount)
	}

	// メタAI: 人間のベット/レイズに対してコール確率を調整
	if ip.config.CpuMetaAI && ip.humanProfile != nil && ip.lastHumanPlayMs > 0 {
		if action == IndianPokerActionFold && callAmount > 0 {
			// CPUは人間のカードが見える → 人間のカードランクブラケットで判定
			humanIdx := ip.findHumanIdx()
			if humanIdx >= 0 && ip.players[humanIdx].GetCardsSize() > 0 {
				humanCardRank := indianPokerCardRank(ip.players[humanIdx].GetCard(0))
				bracket := indianPokerCardBracket(humanCardRank)
				adjustedCall := ip.humanProfile.AdjustedCallChance(0.0, bracket, ip.lastHumanPlayMs)
				if adjustedCall > 0 && rand.Float64() < adjustedCall { //nolint:gosec // non-crypto random for game AI
					action = IndianPokerActionCall
					amount = 0
				}
			}
		}
	}

	maxRaises, maxBetAmount := ip.bettingLimits()

	// PotLimit: CPUベット額をポットサイズに制限
	if maxBetAmount > 0 && amount > maxBetAmount {
		amount = maxBetAmount
	}

	// レイズ上限に達したら、レイズ/ベットをコール/チェックに変更
	if maxRaises > 0 && ip.raiseCount >= maxRaises {
		if action == IndianPokerActionRaise || action == IndianPokerActionBet {
			if callAmount > 0 {
				return IndianPokerActionCall, 0
			}
			return IndianPokerActionCheck, 0
		}
	}

	return action, amount
}

// estimateOwnStrength CPUが自分のカード強度を推定する (0-100)
// 見えている他プレイヤーのカードから、残りの49枚中何枚に勝てるかを計算
func (ip *IndianPoker) estimateOwnStrength(idx int) int {
	// 見えているカードを収集 (自分以外のプレイヤーのカード)
	visibleCards := make([]*Card, 0, 3)
	for i, p := range ip.players {
		if i == idx {
			continue
		}
		if p.GetCardsSize() > 0 {
			visibleCards = append(visibleCards, p.GetCard(0))
		}
	}

	// 見えているカードのランクセット
	visibleRanks := make(map[int]int) // rank -> count
	for _, c := range visibleCards {
		visibleRanks[indianPokerCardRank(c)]++
	}

	// 残りのカード (52 - 見えている他プレイヤーのカード枚数)
	// 自分のカードは不明なので候補に含まれる
	totalRemaining := 52 - len(visibleCards)

	// 各ランクのカードは4枚 (2-14の13種 × 4枚 = 52枚)
	// 見えているランクの残り枚数を考慮して、相手の最高ランクに勝てる確率を計算
	maxVisibleRank := 0
	for _, c := range visibleCards {
		r := indianPokerCardRank(c)
		if r > maxVisibleRank {
			maxVisibleRank = r
		}
	}

	// 自分のカードが maxVisibleRank より大きい確率
	// = (maxVisibleRank より大きいランクの残りカード枚数) / totalRemaining
	cardsAbove := 0
	for rank := maxVisibleRank + 1; rank <= 14; rank++ {
		cardsAbove += max(4-visibleRanks[rank], 0)
	}

	// 同ランクでもスートで勝つ可能性があるため、同ランクの一部も加算
	sameRankCards := 4 - visibleRanks[maxVisibleRank]
	// スート勝率は概算50%
	cardsAbove += sameRankCards / 2

	if totalRemaining <= 0 {
		return 50
	}

	strength := min(cardsAbove*100/totalRemaining, 100)
	return strength
}

// findHumanIdx 人間プレイヤーのインデックスを返す (-1 = 見つからない)
func (ip *IndianPoker) findHumanIdx() int {
	for i, p := range ip.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// cpuPotBet ポット比率ベースのベット額を計算 (最低Ante, 最低minRaise)
func (ip *IndianPoker) cpuPotBet(potPct int) int {
	bet := max(ip.pot*potPct/100, ip.config.Ante)
	return max(bet, ip.minRaise)
}

// --- ゲッター ---

// GetPhase フェーズ取得
func (ip *IndianPoker) GetPhase() int { return ip.phase }

// GetEstimatedStrength は player idx の推定勝率 (0-100) を返す。自分のカードが
// 見えないインディアンポーカーで、見えている相手札から算出した勝率であり、CUI
// のエクイティ表示と CPU 判断で共有する単一ロジック。
func (ip *IndianPoker) GetEstimatedStrength(idx int) int { return ip.estimateOwnStrength(idx) }

// GetPlayers プレイヤー一覧取得
func (ip *IndianPoker) GetPlayers() []*IndianPokerPlayer { return ip.players }

// GetPlayer 指定プレイヤー取得
func (ip *IndianPoker) GetPlayer(i int) *IndianPokerPlayer {
	if i >= 0 && i < len(ip.players) {
		return ip.players[i]
	}
	return nil
}

// GetPlayerCnt プレイヤー数取得
func (ip *IndianPoker) GetPlayerCnt() int { return len(ip.players) }

// GetPot ポット取得
func (ip *IndianPoker) GetPot() int { return ip.pot }

// GetSidePots サイドポット取得
func (ip *IndianPoker) GetSidePots() []SidePot { return ip.sidePots }

// GetDealerIdx ディーラーインデックス取得
func (ip *IndianPoker) GetDealerIdx() int { return ip.dealerIdx }

// GetCurrentTurn 現在のターン取得
func (ip *IndianPoker) GetCurrentTurn() int { return ip.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (ip *IndianPoker) GetGameEndFlag() bool { return ip.gameEndFlag }

// GetLastBet 最後のベット取得
func (ip *IndianPoker) GetLastBet() int { return ip.lastBet }

// GetMinRaise 最小レイズ額取得
func (ip *IndianPoker) GetMinRaise() int { return ip.minRaise }

// GetRaiseCount 現在のレイズ回数取得
func (ip *IndianPoker) GetRaiseCount() int { return ip.raiseCount }

// GetRoundResults ラウンド結果取得
func (ip *IndianPoker) GetRoundResults() []IndianPokerResult { return ip.roundResults }

// GetCpuActions CPU行動記録取得
func (ip *IndianPoker) GetCpuActions() []IndianPokerCpuAction { return ip.cpuActions }

// GetLastCpuError 最後のCPUアクションエラー取得 (テスト・デバッグ用)
func (ip *IndianPoker) GetLastCpuError() error { return ip.lastCpuError }

// GetHumanProfile メタAIプロファイル取得
func (ip *IndianPoker) GetHumanProfile() *IndianPokerHumanProfile { return ip.humanProfile }

// ResetProfile メタAIプロファイルをリセットする
func (ip *IndianPoker) ResetProfile() { ip.humanProfile = nil }

// ExportProfile メタAIプロファイルをエクスポートする (プロファイルがない場合はnil)
func (ip *IndianPoker) ExportProfile() interface{} {
	if ip.humanProfile == nil {
		return nil
	}
	d := ip.humanProfile.Export()
	return &d
}

// ImportProfile JSONバイトからメタAIプロファイルをインポートする
func (ip *IndianPoker) ImportProfile(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	d, err := ImportIndianPokerHumanProfileJSON(data)
	if err != nil {
		return err
	}
	ip.humanProfile = &IndianPokerHumanProfile{}
	ip.humanProfile.Import(d)
	return nil
}

// GetConfig 設定取得
func (ip *IndianPoker) GetConfig() IndianPokerConfig { return ip.config }

// SetConfig 設定変更
func (ip *IndianPoker) SetConfig(cfg IndianPokerConfig) { ip.config = cfg }

// IsHumanTurn 人間のターンかチェック
func (ip *IndianPoker) IsHumanTurn() bool {
	if ip.currentTurn >= 0 && ip.currentTurn < len(ip.players) {
		return ip.players[ip.currentTurn].GetIsHuman()
	}
	return false
}

// GetActedFlags actedフラグ取得
func (ip *IndianPoker) GetActedFlags() []bool {
	result := make([]bool, len(ip.actedFlags))
	copy(result, ip.actedFlags)
	return result
}

// GetHandCount ハンド数取得
func (ip *IndianPoker) GetHandCount() int { return ip.handCount }

// logAction ベッティングアクションを棋譜に記録する
func (ip *IndianPoker) logAction(playerIdx, action, amount int) {
	switch action {
	case IndianPokerActionFold:
		ip.appendLog(playerIdx, "fold", "fold", nil)
	case IndianPokerActionCheck:
		ip.appendLog(playerIdx, "check", "check", nil)
	case IndianPokerActionCall:
		ip.appendLog(playerIdx, "call", fmt.Sprintf("call %d", ip.players[playerIdx].GetCurrentBet()), nil)
	case IndianPokerActionBet:
		ip.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", amount), nil)
	case IndianPokerActionRaise:
		ip.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", amount), nil)
	case IndianPokerActionAllIn:
		ip.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", ip.players[playerIdx].GetCurrentBet()), nil)
	}
}

// indianPokerJSON is the JSON wire format for IndianPoker.
type indianPokerJSON struct {
	TrumpCards      *TrumpCards                  `json:"tc"`
	Players         []*IndianPokerPlayer         `json:"pl"`
	Pot             int                          `json:"pt"`
	SidePots        []SidePot                    `json:"sp"`
	DealerIdx       int                          `json:"di"`
	CurrentTurn     int                          `json:"ct"`
	Phase           int                          `json:"ph"`
	Config          IndianPokerConfig            `json:"cf"`
	GameEndFlag     bool                         `json:"ge"`
	LastBet         int                          `json:"lb"`
	MinRaise        int                          `json:"mr"`
	RaiseCount      int                          `json:"rc"`
	ActedFlags      []bool                       `json:"af"`
	RoundResults    []IndianPokerResult          `json:"rr"`
	CpuActions      []IndianPokerCpuAction       `json:"ca"`
	StartingChips   []int                        `json:"sc"`
	HandCount       int                          `json:"hc"`
	ActionLog       []*ActionLogEntry            `json:"al"`
	Profile         *IndianPokerHumanProfileData `json:"pf,omitempty"`
	LastHumanPlayMs int                          `json:"hm"`
}

// indianPokerMaxSliceLen caps slice sizes during deserialisation.
const indianPokerMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (ip *IndianPoker) MarshalJSON() ([]byte, error) {
	j := indianPokerJSON{
		TrumpCards:      ip.trumpCards,
		Players:         ip.players,
		Pot:             ip.pot,
		SidePots:        ip.sidePots,
		DealerIdx:       ip.dealerIdx,
		CurrentTurn:     ip.currentTurn,
		Phase:           ip.phase,
		Config:          ip.config,
		GameEndFlag:     ip.gameEndFlag,
		LastBet:         ip.lastBet,
		MinRaise:        ip.minRaise,
		RaiseCount:      ip.raiseCount,
		ActedFlags:      ip.actedFlags,
		RoundResults:    ip.roundResults,
		CpuActions:      ip.cpuActions,
		StartingChips:   ip.startingChips,
		HandCount:       ip.handCount,
		ActionLog:       ip.actionLog,
		LastHumanPlayMs: ip.lastHumanPlayMs,
	}
	if ip.humanProfile != nil {
		d := ip.humanProfile.Export()
		j.Profile = &d
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (ip *IndianPoker) UnmarshalJSON(data []byte) error {
	var j indianPokerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > indianPokerMaxSliceLen || len(j.SidePots) > indianPokerMaxSliceLen ||
		len(j.ActedFlags) > indianPokerMaxSliceLen || len(j.RoundResults) > indianPokerMaxSliceLen ||
		len(j.CpuActions) > indianPokerMaxSliceLen || len(j.StartingChips) > indianPokerMaxSliceLen ||
		len(j.ActionLog) > indianPokerMaxSliceLen {
		return fmt.Errorf("indianpoker: input array exceeds maximum allowed size")
	}
	ip.trumpCards = j.TrumpCards
	if ip.trumpCards == nil {
		ip.trumpCards = NewTrumpCards(0)
	}
	ip.players = j.Players
	if ip.players == nil {
		ip.players = make([]*IndianPokerPlayer, 0)
	}
	ip.pot = j.Pot
	ip.sidePots = j.SidePots
	if ip.sidePots == nil {
		ip.sidePots = make([]SidePot, 0)
	}
	ip.dealerIdx = j.DealerIdx
	ip.currentTurn = j.CurrentTurn
	ip.phase = j.Phase
	ip.config = j.Config
	ip.gameEndFlag = j.GameEndFlag
	ip.lastBet = j.LastBet
	ip.minRaise = j.MinRaise
	ip.raiseCount = j.RaiseCount
	ip.actedFlags = j.ActedFlags
	if ip.actedFlags == nil {
		ip.actedFlags = make([]bool, 0)
	}
	ip.roundResults = j.RoundResults
	if ip.roundResults == nil {
		ip.roundResults = make([]IndianPokerResult, 0)
	}
	ip.cpuActions = j.CpuActions
	if ip.cpuActions == nil {
		ip.cpuActions = make([]IndianPokerCpuAction, 0)
	}
	ip.startingChips = j.StartingChips
	if ip.startingChips == nil {
		ip.startingChips = make([]int, 0)
	}
	ip.handCount = j.HandCount
	ip.actionLog = j.ActionLog
	if ip.actionLog == nil {
		ip.actionLog = make([]*ActionLogEntry, 0)
	}
	ip.lastHumanPlayMs = j.LastHumanPlayMs
	if j.Profile != nil {
		ip.humanProfile = &IndianPokerHumanProfile{}
		ip.humanProfile.Import(*j.Profile)
	}
	return nil
}

// --- テスト用セッター ---

// SetPhase フェーズ設定 (テスト用)
func (ip *IndianPoker) SetPhase(phase int) { ip.phase = phase }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (ip *IndianPoker) SetGameEndFlag(flag bool) { ip.gameEndFlag = flag }

// SetCurrentTurn 現在のターン設定 (テスト用)
func (ip *IndianPoker) SetCurrentTurn(turn int) { ip.currentTurn = turn }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (ip *IndianPoker) SetDealerIdx(idx int) { ip.dealerIdx = idx }

// SetLastBet 最後のベット設定 (テスト用)
func (ip *IndianPoker) SetLastBet(bet int) { ip.lastBet = bet }

// SetPot ポット設定 (テスト用)
func (ip *IndianPoker) SetPot(pot int) { ip.pot = pot }

// SetActedFlags actedフラグ設定 (テスト用)
func (ip *IndianPoker) SetActedFlags(flags []bool) { ip.actedFlags = flags }

// SetCpuActions CPU行動記録設定 (テスト用)
func (ip *IndianPoker) SetCpuActions(actions []IndianPokerCpuAction) { ip.cpuActions = actions }

// SetRoundResults ラウンド結果設定 (テスト用)
func (ip *IndianPoker) SetRoundResults(results []IndianPokerResult) { ip.roundResults = results }

// SetMinRaise 最小レイズ額設定 (テスト用)
func (ip *IndianPoker) SetMinRaise(mr int) { ip.minRaise = mr }

// SetRaiseCount レイズ回数設定 (テスト用)
func (ip *IndianPoker) SetRaiseCount(rc int) { ip.raiseCount = rc }

// SetStartingChips ハンド開始時チップ設定 (テスト用)
func (ip *IndianPoker) SetStartingChips(chips []int) { ip.startingChips = chips }

// SetSidePots サイドポット設定 (テスト用)
func (ip *IndianPoker) SetSidePots(pots []SidePot) { ip.sidePots = pots }

// SetHandCount ハンド数設定 (テスト用)
func (ip *IndianPoker) SetHandCount(count int) { ip.handCount = count }
