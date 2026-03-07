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

// 記憶減衰率 (1ターンあたりの忘却ベースレート)
const (
	decayRateEasy   = 0.15
	decayRateNormal = 0.05
	decayRateHard   = 0.0
)

// 動的ブラフ確率の定数
const (
	bluffChanceBase          = 0.4
	bluffChanceLastCard      = 0.1
	bluffPenaltyLargeTable   = 0.15 // テーブル20枚以上
	bluffPenaltyMediumTable  = 0.10 // テーブル10枚以上
	bluffTableLargeThreshold = 20
	bluffTableMedThreshold   = 10
)

// mixedBluffChance 混合ブラフ確率 (intentBluff=false かつ numCards>1 のとき発動)
const mixedBluffChance = 0.3

// テル（緊張の兆候）表示確率の定数
const (
	tellChanceEasy   = 0.4
	tellChanceNormal = 0.2
	tellChanceHard   = 0.05
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
	HasTell      bool // テル（緊張の兆候）を見せているか
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
	turnCounter     int
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
	d.turnCounter = 0

	resetPlayers(d.players, func(p *DoubtPlayer) { p.ResetMemory() })

	d.trumpCards.Shuffle()
	dealAllCards(d.trumpCards, d.players)
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
	// 重複チェック
	seen := make(map[int]bool, len(cardIndices))
	for _, idx := range cardIndices {
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "カードインデックスが重複しています")
		}
		seen[idx] = true
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

	// 状況に応じた動的ブラフ確率 (除去後の値で計算して既存の確率分布を維持)
	postRemovalHandSize := player.GetCardsSize() - numCards
	postRemovalTableCount := len(d.tableCards) + numCards
	intentBluff := rand.Float64() < d.calcBluffChance(postRemovalHandSize, postRemovalTableCount)

	var played []*Card
	var claimedValue int

	if !intentBluff && numCards > 1 && rand.Float64() < mixedBluffChance {
		// 混合ブラフ: CPUが実際に持つ値を宣言しつつ不一致カードを混ぜて出す
		claimedValue = player.GetCard(0).GetValue()
		played = selectMixedCards(player, claimedValue, numCards)
	} else {
		// 意図的ブラフまたは正直プレイ: 先頭 numCards 枚を出す
		cardIndices := make([]int, numCards)
		for i := range cardIndices {
			cardIndices[i] = i
		}
		played = player.RemoveCards(cardIndices)

		if intentBluff {
			claimedValue = rand.Intn(13) + 1
		} else {
			claimedValue = played[0].GetValue()
		}
	}

	d.tableCards = append(d.tableCards, played...)

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
	if isActuallyBluff {
		cpuAction.HasTell = rand.Float64() < calcTellChance(d.config.CpuMemoryLevel)
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

	// 記憶の減衰処理
	rate := memoryDecayRate(d.config.CpuMemoryLevel)
	for i := 0; i < DoubtPlayerCnt; i++ {
		if d.players[i].GetIsHuman() || d.players[i].GetIsFinished() {
			continue
		}
		d.players[i].DecayMemories(d.turnCounter, rate)
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
			p.RecordRevealedCard(card.GetValue(), retentionChance, d.turnCounter)
		}
	}

	d.turnCounter++
	d.tableCards = nil
	d.currentTurn = (d.lastAction.PlayerIdx + 1) % DoubtPlayerCnt
	d.phase = DoubtPhasePlay
}

// SkipDoubt ダウトをスキップ: テーブルカードはそのまま残り、手番が進む
func (d *Doubt) SkipDoubt() {
	if d.phase != DoubtPhaseDoubt || d.lastAction == nil {
		return
	}
	d.turnCounter++
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

// GetTurnCounter ターンカウンター取得
func (d *Doubt) GetTurnCounter() int { return d.turnCounter }

// GetConfig ゲーム設定取得
func (d *Doubt) GetConfig() DoubtConfig { return d.config }

// SetConfig ゲーム設定変更
func (d *Doubt) SetConfig(cfg DoubtConfig) { d.config = cfg }

// memoryDecayRate 記憶力レベルに対応する記憶減衰率を返す
func memoryDecayRate(level DoubtMemoryLevel) float64 {
	switch level {
	case DoubtMemoryLevelEasy:
		return decayRateEasy
	case DoubtMemoryLevelNormal:
		return decayRateNormal
	case DoubtMemoryLevelHard:
		return decayRateHard
	default:
		return decayRateNormal
	}
}

// calcTellChance 記憶力レベルに対応するテル表示確率を返す
func calcTellChance(level DoubtMemoryLevel) float64 {
	switch level {
	case DoubtMemoryLevelEasy:
		return tellChanceEasy
	case DoubtMemoryLevelNormal:
		return tellChanceNormal
	case DoubtMemoryLevelHard:
		return tellChanceHard
	default:
		return tellChanceNormal
	}
}

// selectMixedCards は手札から claimedValue に一致するカードと一致しないカードを
// 混ぜて numCards 枚選択し、プレイヤーの手札から取り除いて返す。
// 不一致カードが0枚の場合は先頭 numCards 枚を返す。
// numCards >= 2 が前提 (呼び出し元で保証)。
func selectMixedCards(player *DoubtPlayer, claimedValue int, numCards int) []*Card {
	if numCards < 2 {
		return player.RemoveCards([]int{0})
	}

	var matching, nonMatching []int
	for i := 0; i < player.GetCardsSize(); i++ {
		if player.GetCard(i).GetValue() == claimedValue {
			matching = append(matching, i)
		} else {
			nonMatching = append(nonMatching, i)
		}
	}

	// 一致または不一致カードがない場合はフォールバック
	if len(matching) == 0 || len(nonMatching) == 0 {
		indices := make([]int, numCards)
		for i := range indices {
			indices[i] = i
		}
		return player.RemoveCards(indices)
	}

	// matchCount は [1, min(len(matching), numCards-1)]
	maxMatch := len(matching)
	if maxMatch > numCards-1 {
		maxMatch = numCards - 1
	}
	matchCount := rand.Intn(maxMatch) + 1
	nonMatchCount := numCards - matchCount

	// 不一致カードが足りない場合は調整
	if nonMatchCount > len(nonMatching) {
		nonMatchCount = len(nonMatching)
		matchCount = numCards - nonMatchCount
	}

	// ランダムに選択
	rand.Shuffle(len(matching), func(i, j int) { matching[i], matching[j] = matching[j], matching[i] })
	rand.Shuffle(len(nonMatching), func(i, j int) { nonMatching[i], nonMatching[j] = nonMatching[j], nonMatching[i] })

	selected := make([]int, 0, numCards)
	selected = append(selected, matching[:matchCount]...)
	selected = append(selected, nonMatching[:nonMatchCount]...)

	return player.RemoveCards(selected)
}

// calcBluffChance 手札枚数とテーブルカード枚数に応じた動的ブラフ確率を計算する
func (d *Doubt) calcBluffChance(handSize, tableCardCount int) float64 {
	if handSize <= 1 {
		return bluffChanceLastCard
	}
	chance := bluffChanceBase
	if tableCardCount >= bluffTableLargeThreshold {
		chance -= bluffPenaltyLargeTable
	} else if tableCardCount >= bluffTableMedThreshold {
		chance -= bluffPenaltyMediumTable
	}
	return chance
}
