package domain

import "math/rand"

// DoubtPlayerCnt ダウトプレイヤー数
const DoubtPlayerCnt = 4

// MinClaimedValue claimed value の最小値 (A)
const MinClaimedValue = 1

// MaxClaimedValue claimed value の最大値 (K)
const MaxClaimedValue = 13

// randomDoubtChance CPUがランダムにダウトを宣言する確率
const randomDoubtChance = 0.3

// retentionChanceEasy / Normal / Hard 各記憶レベルのカード保持確率
const (
	retentionChanceEasy   = 0.3
	retentionChanceNormal = 0.7
	retentionChanceHard   = 1.0
)

// DoubtPhase ゲームフェーズ
type DoubtPhase int

const (
	// DoubtPhasePlay プレイフェーズ (カードを出す番)
	DoubtPhasePlay DoubtPhase = 0
	// DoubtPhaseDoubt ダウトフェーズ (ダウトするか判断する番)
	DoubtPhaseDoubt DoubtPhase = 1
	// DoubtPhaseEnd ゲーム終了フェーズ
	DoubtPhaseEnd DoubtPhase = 2
)

// DoubtAction プレイヤーが出したカードの情報 (実際のカードを含む)
type DoubtAction struct {
	PlayerIdx    int     // 出したプレイヤーインデックス
	ClaimedValue int     // 宣言した値 (1-13)
	CardCount    int     // 出した枚数
	PlayedCards  []*Card // 実際に出したカード (ダウト時に公開)
}

// DoubtCpuAction 1ターン分の行動記録 (表示用; 実際のカードは含まない)
type DoubtCpuAction struct {
	PlayerIdx    int  // 行動したプレイヤーインデックス
	ClaimedValue int  // 宣言した値
	CardCount    int  // 出した枚数
	IsBluff      bool // ブラフかどうか (CPU のみ追跡)
}

// DoubtDoubtResult ダウト解決結果
type DoubtDoubtResult struct {
	DoubterIdx    int     // ダウトしたプレイヤーインデックス
	CardPlayerIdx int     // カードを出したプレイヤーインデックス
	WasLying      bool    // カードを出したプレイヤーが嘘をついていたか
	LoserIdx      int     // 負けたプレイヤーインデックス (テーブルカードを引き取る)
	CardCount     int     // 引き取ったカード枚数
	RevealedCards []*Card // 公開されたカード
}

// cardsPerValue 標準52枚デッキにおける各値のカード枚数
const cardsPerValue = 4

// Doubt ダウトゲームクラス
type Doubt struct {
	trumpCards      *TrumpCards
	players         []*DoubtPlayer
	currentTurn     int
	phase           DoubtPhase
	tableCards      []*Card
	lastAction      *DoubtAction
	gameEndFlag     bool
	winnerIdx       int
	cpuDoubters     []int
	cpuActions      []*DoubtCpuAction
	humanAction     *DoubtCpuAction
	lastDoubtResult *DoubtDoubtResult
	config          DoubtConfig
}

// NewDoubt コンストラクタ
func NewDoubt(trumpCards *TrumpCards, players []*DoubtPlayer) *Doubt {
	return &Doubt{
		trumpCards:  trumpCards,
		players:     players,
		currentTurn: 0,
		phase:       DoubtPhasePlay,
		gameEndFlag: false,
		winnerIdx:   -1,
		config:      DefaultDoubtConfig(),
	}
}

// Reset ゲーム初期化: シャッフルして各プレイヤーに均等配布
func (d *Doubt) Reset() {
	d.gameEndFlag = false
	d.currentTurn = 0
	d.phase = DoubtPhasePlay
	d.tableCards = nil
	d.lastAction = nil
	d.lastDoubtResult = nil
	d.cpuDoubters = nil
	d.cpuActions = nil
	d.humanAction = nil
	d.winnerIdx = -1

	for _, p := range d.players {
		p.Reset()
		p.SetIsFinished(false)
		p.ResetMemory()
	}

	d.trumpCards.Shuffle()
	idx := 0
	for {
		card := d.trumpCards.DrawCard()
		if card == nil {
			break
		}
		d.players[idx%DoubtPlayerCnt].AddCard(card)
		idx++
	}
}

// PlayerPlay 人間プレイヤーがカードを出す
// cardIndices: 出すカードのインデックス, claimedValue: 宣言する値 (1-13)
func (d *Doubt) PlayerPlay(cardIndices []int, claimedValue int) error {
	if d.gameEndFlag {
		return ErrGameEnded
	}
	if d.phase != DoubtPhasePlay {
		return ErrWrongPhase
	}
	if !d.players[d.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if claimedValue < MinClaimedValue || claimedValue > MaxClaimedValue {
		return NewDomainError(ErrInvalidPlay, "宣言する値は1から13の範囲で指定してください")
	}
	if len(cardIndices) == 0 {
		return NewDomainError(ErrInvalidPlay, "1枚以上のカードを指定してください")
	}

	player := d.players[d.currentTurn]
	for _, idx := range cardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
	}

	played := player.RemoveCards(cardIndices)
	d.tableCards = append(d.tableCards, played...)

	d.lastAction = &DoubtAction{
		PlayerIdx:    d.currentTurn,
		ClaimedValue: claimedValue,
		CardCount:    len(played),
		PlayedCards:  played,
	}
	d.humanAction = &DoubtCpuAction{
		PlayerIdx:    d.currentTurn,
		ClaimedValue: claimedValue,
		CardCount:    len(played),
		IsBluff:      false,
	}
	d.cpuActions = nil

	if player.GetCardsSize() == 0 {
		player.SetIsFinished(true)
		d.winnerIdx = d.currentTurn
		d.gameEndFlag = true
		return nil
	}

	d.phase = DoubtPhaseDoubt
	d.decideCpuDoubters()
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (d *Doubt) CpuPlay() {
	if d.gameEndFlag || d.players[d.currentTurn].GetIsHuman() {
		return
	}
	if d.phase != DoubtPhasePlay {
		return
	}

	playerIdx := d.currentTurn
	player := d.players[playerIdx]

	// 出す枚数をランダム決定 (1〜手持ち枚数)
	numCards := 1
	if player.GetCardsSize() > 1 {
		numCards = rand.Intn(player.GetCardsSize()) + 1
	}

	// 先頭 numCards 枚を出す
	cardIndices := make([]int, numCards)
	for i := range cardIndices {
		cardIndices[i] = i
	}
	played := player.RemoveCards(cardIndices)
	d.tableCards = append(d.tableCards, played...)

	// 40% の確率でブラフ
	intentBluff := rand.Float64() < 0.4
	var claimedValue int
	if intentBluff {
		claimedValue = rand.Intn(13) + 1
	} else {
		claimedValue = played[0].GetValue()
	}

	// IsBluff は宣言値と実際のカードが一致しないかで判定 (checkLying と一致させる)
	isActuallyBluff := false
	for _, card := range played {
		if card.GetValue() != claimedValue {
			isActuallyBluff = true
			break
		}
	}

	d.lastAction = &DoubtAction{
		PlayerIdx:    playerIdx,
		ClaimedValue: claimedValue,
		CardCount:    numCards,
		PlayedCards:  played,
	}
	cpuAction := &DoubtCpuAction{
		PlayerIdx:    playerIdx,
		ClaimedValue: claimedValue,
		CardCount:    numCards,
		IsBluff:      isActuallyBluff,
	}
	d.cpuActions = append(d.cpuActions, cpuAction)

	if player.GetCardsSize() == 0 {
		player.SetIsFinished(true)
		d.winnerIdx = playerIdx
		d.gameEndFlag = true
		return
	}

	d.phase = DoubtPhaseDoubt
	d.decideCpuDoubters()
}

// memoryRetentionChance 記憶力レベルに対応するカード記憶確率を返す
func memoryRetentionChance(level DoubtMemoryLevel) float64 {
	switch level {
	case DoubtMemoryLevelEasy:
		return retentionChanceEasy
	case DoubtMemoryLevelNormal:
		return retentionChanceNormal
	case DoubtMemoryLevelHard:
		return retentionChanceHard
	default:
		return retentionChanceNormal
	}
}

// decideCpuDoubters カードを出した後、CPUがダウトするか決定する
// 宣言が物理的に不可能な場合は100%ダウト、それ以外は30%の確率でダウト
func (d *Doubt) decideCpuDoubters() {
	d.cpuDoubters = nil
	if d.lastAction == nil {
		return
	}
	cardPlayerIdx := d.lastAction.PlayerIdx
	claimedValue := d.lastAction.ClaimedValue
	claimedCount := d.lastAction.CardCount
	for i := 0; i < DoubtPlayerCnt; i++ {
		if i == cardPlayerIdx {
			continue
		}
		if d.players[i].GetIsHuman() {
			continue
		}
		if d.players[i].GetIsFinished() {
			continue
		}
		known := d.players[i].CountKnownCards(claimedValue)
		if known+claimedCount > cardsPerValue {
			// 物理的に不可能な宣言 → 100%ダウト
			d.cpuDoubters = append(d.cpuDoubters, i)
		} else if rand.Float64() < randomDoubtChance {
			d.cpuDoubters = append(d.cpuDoubters, i)
		}
	}
}

// findHighestPriorityDoubter ダウトの優先順位が最も高い doubter を返す
// カード出したプレイヤーの次から順に最初に見つかった doubter を返す。
// 見つからない場合は -1。
func (d *Doubt) findHighestPriorityDoubter(doubterIndices []int) int {
	if len(doubterIndices) == 0 || d.lastAction == nil {
		return -1
	}
	cardPlayerIdx := d.lastAction.PlayerIdx
	for i := 1; i <= DoubtPlayerCnt; i++ {
		idx := (cardPlayerIdx + i) % DoubtPlayerCnt
		for _, doubter := range doubterIndices {
			if doubter == idx {
				return idx
			}
		}
	}
	return -1
}

// checkLying カードを出したプレイヤーが嘘をついているか確認
func (d *Doubt) checkLying() bool {
	if d.lastAction == nil {
		return false
	}
	for _, card := range d.lastAction.PlayedCards {
		if card.GetValue() != d.lastAction.ClaimedValue {
			return true
		}
	}
	return false
}

// ResolveDoubt ダウト解決: 最優先 doubter がカードを公開し、負けた方がテーブルカードを引き取る
// doubterIndices: ダウトしたプレイヤーインデックスのリスト (人間 + CPU)
func (d *Doubt) ResolveDoubt(doubterIndices []int) {
	if d.phase != DoubtPhaseDoubt || d.lastAction == nil {
		return
	}

	doubter := d.findHighestPriorityDoubter(doubterIndices)
	if doubter < 0 {
		d.SkipDoubt()
		return
	}

	wasLying := d.checkLying()
	var loserIdx int
	if wasLying {
		loserIdx = d.lastAction.PlayerIdx
	} else {
		loserIdx = doubter
	}

	cardCount := len(d.tableCards)
	revealedCards := d.lastAction.PlayedCards

	for _, card := range d.tableCards {
		d.players[loserIdx].AddCard(card)
	}

	d.lastDoubtResult = &DoubtDoubtResult{
		DoubterIdx:    doubter,
		CardPlayerIdx: d.lastAction.PlayerIdx,
		WasLying:      wasLying,
		LoserIdx:      loserIdx,
		CardCount:     cardCount,
		RevealedCards: revealedCards,
	}

	// 非敗者のCPUがカードを記憶する
	retentionChance := memoryRetentionChance(d.config.CpuMemoryLevel)
	for i, p := range d.players {
		if p.GetIsHuman() || i == loserIdx {
			continue
		}
		for _, card := range revealedCards {
			p.RecordRevealedCard(card.GetValue(), retentionChance)
		}
	}

	d.tableCards = nil
	d.currentTurn = (d.lastAction.PlayerIdx + 1) % DoubtPlayerCnt
	d.phase = DoubtPhasePlay
}

// SkipDoubt ダウトをスキップ: テーブルカードはそのまま残り、手番が進む
func (d *Doubt) SkipDoubt() {
	if d.phase != DoubtPhaseDoubt || d.lastAction == nil {
		return
	}
	d.lastDoubtResult = nil
	d.currentTurn = (d.lastAction.PlayerIdx + 1) % DoubtPlayerCnt
	d.phase = DoubtPhasePlay
}

// IsHumanTurn 現在の手番が人間かどうか
func (d *Doubt) IsHumanTurn() bool {
	return d.players[d.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (d *Doubt) GetCurrentTurn() int { return d.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (d *Doubt) GetGameEndFlag() bool { return d.gameEndFlag }

// GetPhase 現在のフェーズ取得
func (d *Doubt) GetPhase() DoubtPhase { return d.phase }

// GetPlayerCnt プレイヤー数取得
func (d *Doubt) GetPlayerCnt() int { return len(d.players) }

// GetPlayer プレイヤー取得
func (d *Doubt) GetPlayer(i int) *DoubtPlayer {
	if i < 0 || i >= len(d.players) {
		return nil
	}
	return d.players[i]
}

// GetTableCardCount テーブルカード枚数取得
func (d *Doubt) GetTableCardCount() int { return len(d.tableCards) }

// GetTableCards テーブルカード取得
func (d *Doubt) GetTableCards() []*Card { return d.tableCards }

// GetLastAction 最後のプレイアクション取得
func (d *Doubt) GetLastAction() *DoubtAction { return d.lastAction }

// GetCpuDoubters CPUダウターインデックスリスト取得
func (d *Doubt) GetCpuDoubters() []int { return d.cpuDoubters }

// GetWinnerIdx 勝者インデックス取得 (-1 = まだ決まっていない)
func (d *Doubt) GetWinnerIdx() int { return d.winnerIdx }

// GetCpuActions CPUターンの行動履歴取得
func (d *Doubt) GetCpuActions() []*DoubtCpuAction { return d.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (d *Doubt) GetHumanAction() *DoubtCpuAction { return d.humanAction }

// GetLastDoubtResult 最後のダウト結果取得
func (d *Doubt) GetLastDoubtResult() *DoubtDoubtResult { return d.lastDoubtResult }

// SetTableCards テーブルカード設定 (テスト用)
func (d *Doubt) SetTableCards(cards []*Card) { d.tableCards = cards }

// SetLastAction 最後のプレイアクション設定 (テスト用)
func (d *Doubt) SetLastAction(action *DoubtAction) { d.lastAction = action }

// SetPhase フェーズ設定 (テスト用)
func (d *Doubt) SetPhase(phase DoubtPhase) { d.phase = phase }

// SetCpuDoubters CPUダウターリスト設定 (テスト用)
func (d *Doubt) SetCpuDoubters(doubters []int) { d.cpuDoubters = doubters }

// SetCpuActions CPU行動設定 (テスト用)
func (d *Doubt) SetCpuActions(actions []*DoubtCpuAction) { d.cpuActions = actions }

// SetHumanAction 人間の行動設定 (テスト用)
func (d *Doubt) SetHumanAction(action *DoubtCpuAction) { d.humanAction = action }

// SetLastDoubtResult ダウト結果設定 (テスト用)
func (d *Doubt) SetLastDoubtResult(result *DoubtDoubtResult) { d.lastDoubtResult = result }

// SetWinnerIdx 勝者インデックス設定 (テスト用)
func (d *Doubt) SetWinnerIdx(idx int) { d.winnerIdx = idx }

// GetConfig ゲーム設定取得
func (d *Doubt) GetConfig() DoubtConfig { return d.config }

// SetConfig ゲーム設定変更
func (d *Doubt) SetConfig(cfg DoubtConfig) { d.config = cfg }
