package domain

import (
	"fmt"
	"math/rand"
	"sort"
)

// DaifugoPlayerCnt 大富豪プレイヤー数
const DaifugoPlayerCnt = 4

// DaifugoJokerCount 大富豪で使用するジョーカー枚数
const DaifugoJokerCount = 2

// DaifugoCardStrength カードの強さを返す (3が最弱、2が最強)
// 3 < 4 < 5 < 6 < 7 < 8 < 9 < 10 < J(11) < Q(12) < K(13) < A(1) < 2(2)
func DaifugoCardStrength(v int) int {
	if v == 1 {
		return 14 // Ace
	}
	if v == 2 {
		return 15 // 2 は最強
	}
	return v
}

// DaifugoCardStrengthRevolution 革命中のカードの強さを返す (2が最弱、3が最強)
// 2 < A(1) < K(13) < Q(12) < J(11) < 10 < 9 < 8 < 7 < 6 < 5 < 4 < 3
func DaifugoCardStrengthRevolution(v int) int {
	return 18 - DaifugoCardStrength(v)
}

// DaifugoJokerStrength ジョーカーの強さ (常に最強)
const DaifugoJokerStrength = 16

// ランク定数
const (
	DaifugoRankDaifugo   = 1 // 大富豪
	DaifugoRankFugo      = 2 // 富豪
	DaifugoRankHeimin    = 3 // 平民
	DaifugoRankDaihinmin = 4 // 大貧民
)

// カード交換枚数
const (
	DaifugoExchangeCountDaifugo = 2 // 大富豪↔大貧民: 2枚
	DaifugoExchangeCountFugo    = 1 // 富豪↔平民: 1枚
)

// DaifugoConfig 大富豪ローカルルール設定
type DaifugoConfig struct {
	JokerCount          int  // ジョーカー枚数 (default: 2)
	EightCutEnabled     bool // 8切り
	SuitLockEnabled     bool // スート縛り
	ElevenBackEnabled   bool // 11バック
	SequenceEnabled     bool // 階段
	CardExchangeEnabled bool // カード交換
}

// DefaultDaifugoConfig デフォルトのローカルルール設定 (全て有効)
func DefaultDaifugoConfig() DaifugoConfig {
	return DaifugoConfig{
		JokerCount:          DaifugoJokerCount,
		EightCutEnabled:     true,
		SuitLockEnabled:     true,
		ElevenBackEnabled:   true,
		SequenceEnabled:     true,
		CardExchangeEnabled: true,
	}
}

// DaifugoCpuAction CPUまたは人間の1ターン分の行動記録
type DaifugoCpuAction struct {
	PlayerIdx   int     // 行動したプレイヤーインデックス
	PlayedCards []*Card // 出したカード (nil = パス)
}

// DaifugoExchangeAction カード交換1件の記録
type DaifugoExchangeAction struct {
	FromPlayerIdx int     // 渡したプレイヤーインデックス
	ToPlayerIdx   int     // 受け取ったプレイヤーインデックス
	Cards         []*Card // 交換されたカード
}

// Daifugo 大富豪ゲームクラス
type Daifugo struct {
	trumpCards        *TrumpCards
	players           []*DaifugoPlayer
	currentTurn       int                      // 現在の手番プレイヤーインデックス
	tableCards        []*Card                  // 場に出されているカード (nil = 場はクリア)
	lastPlayPlayerIdx int                      // 最後にカードを出したプレイヤーインデックス (-1 = なし)
	gameEndFlag       bool                     // ゲーム終了フラグ
	passCount         int                      // 最後の出し以降の連続パス数
	cpuActions        []*DaifugoCpuAction      // 人間ターン後のCPUの行動履歴
	humanAction       *DaifugoCpuAction        // 人間の最後の行動
	revolutionActive  bool                     // 革命フラグ (true = 革命中)
	config            DaifugoConfig            // ローカルルール設定
	suitLocked        bool                     // スート縛り発動中
	lockedSuit        int                      // 縛られているスート (CardDesignSpade等)
	elevenBackActive  bool                     // 11バック発動中
	tableIsSequence   bool                     // 場が階段プレイか
	exchangeActions   []*DaifugoExchangeAction // カード交換記録
}

// NewDaifugo コンストラクタ
func NewDaifugo(trumpCards *TrumpCards, players []*DaifugoPlayer, config DaifugoConfig) *Daifugo {
	return &Daifugo{
		trumpCards:        trumpCards,
		players:           players,
		currentTurn:       0,
		tableCards:        nil,
		lastPlayPlayerIdx: -1,
		gameEndFlag:       false,
		passCount:         0,
		cpuActions:        nil,
		humanAction:       nil,
		revolutionActive:  false,
		config:            config,
		suitLocked:        false,
		lockedSuit:        0,
		elevenBackActive:  false,
		tableIsSequence:   false,
		exchangeActions:   nil,
	}
}

// Reset ゲーム初期化
func (d *Daifugo) Reset() {
	// 前回のランクをプレイヤーオブジェクトに保存 (カード交換に使用)
	// プレイヤーオブジェクトのポインタはシャッフル後も保持されるので安全
	hasPrevRanks := false
	for _, p := range d.players {
		if p.GetRank() > 0 {
			p.SetPrevRank(p.GetRank())
			hasPrevRanks = true
		} else {
			p.SetPrevRank(-1)
		}
	}

	d.gameEndFlag = false
	d.currentTurn = 0
	d.tableCards = nil
	d.lastPlayPlayerIdx = -1
	d.passCount = 0
	d.cpuActions = nil
	d.humanAction = nil
	d.revolutionActive = false
	d.suitLocked = false
	d.lockedSuit = 0
	d.elevenBackActive = false
	d.tableIsSequence = false
	d.exchangeActions = nil

	// シャッフル
	d.trumpCards.Shuffle()

	// 全プレイヤーのカードリセット
	for _, p := range d.players {
		p.Reset()
		p.SetIsFinished(false)
		p.SetRank(-1)
	}

	// プレイ順をランダムにする
	rand.Shuffle(len(d.players), func(i, j int) {
		d.players[i], d.players[j] = d.players[j], d.players[i]
	})

	// 全カードを配る (ジョーカー含む)
	idx := 0
	for {
		card := d.trumpCards.DrawCard()
		if card == nil {
			break
		}
		d.players[idx%DaifugoPlayerCnt].AddCard(card)
		idx++
	}

	// 各プレイヤーの手札をソート
	for _, p := range d.players {
		p.SortCardsByStrength(d.cardStrengthForCard)
	}

	// カード交換
	if d.config.CardExchangeEnabled && hasPrevRanks {
		d.performCardExchange()
	}
}

// performCardExchange 前回のランクに基づいてカード交換を行う
func (d *Daifugo) performCardExchange() {
	d.exchangeActions = make([]*DaifugoExchangeAction, 0)

	// 前回のランクからプレイヤーインデックスのマッピングを作成
	rankToPlayer := make(map[int]int) // rank → current player index
	for i, p := range d.players {
		if p.GetPrevRank() > 0 {
			rankToPlayer[p.GetPrevRank()] = i
		}
	}

	// 大富豪 ↔ 大貧民: 2枚交換
	if idx1, ok1 := rankToPlayer[DaifugoRankDaifugo]; ok1 {
		if idx4, ok4 := rankToPlayer[DaifugoRankDaihinmin]; ok4 {
			d.exchangeCardsBetween(idx1, idx4, DaifugoExchangeCountDaifugo)
		}
	}

	// 富豪 ↔ 平民: 1枚交換
	if idx2, ok2 := rankToPlayer[DaifugoRankFugo]; ok2 {
		if idx3, ok3 := rankToPlayer[DaifugoRankHeimin]; ok3 {
			d.exchangeCardsBetween(idx2, idx3, DaifugoExchangeCountFugo)
		}
	}

	// 交換後に再ソート
	for _, p := range d.players {
		p.SortCardsByStrength(d.cardStrengthForCard)
	}
}

// exchangeCardsBetween 上位プレイヤーと下位プレイヤー間でカード交換
// 下位→上位: 最強カードをcount枚渡す
// 上位→下位: 最弱カードをcount枚渡す
func (d *Daifugo) exchangeCardsBetween(upperIdx, lowerIdx, count int) {
	upper := d.players[upperIdx]
	lower := d.players[lowerIdx]

	if upper.GetCardsSize() < count || lower.GetCardsSize() < count {
		return
	}

	// 下位の最強カード(末尾)をcount枚取得
	lowerBestIndices := make([]int, count)
	for i := 0; i < count; i++ {
		lowerBestIndices[i] = lower.GetCardsSize() - count + i
	}
	lowerBestCards := lower.RemoveCards(lowerBestIndices)

	// 上位の最弱カード(先頭)をcount枚取得
	upperWorstIndices := make([]int, count)
	for i := 0; i < count; i++ {
		upperWorstIndices[i] = i
	}
	upperWorstCards := upper.RemoveCards(upperWorstIndices)

	// カードを交換
	for _, c := range lowerBestCards {
		upper.AddCard(c)
	}
	for _, c := range upperWorstCards {
		lower.AddCard(c)
	}

	// 交換記録
	d.exchangeActions = append(d.exchangeActions, &DaifugoExchangeAction{
		FromPlayerIdx: lowerIdx,
		ToPlayerIdx:   upperIdx,
		Cards:         lowerBestCards,
	})
	d.exchangeActions = append(d.exchangeActions, &DaifugoExchangeAction{
		FromPlayerIdx: upperIdx,
		ToPlayerIdx:   lowerIdx,
		Cards:         upperWorstCards,
	})
}

// cardStrength 現在の革命・11バック状態に応じたカード値の強さを返す
func (d *Daifugo) cardStrength(v int) int {
	// 革命と11バックのXOR: 両方有効なら打ち消し合う
	reversed := d.revolutionActive != d.elevenBackActive
	if reversed {
		return DaifugoCardStrengthRevolution(v)
	}
	return DaifugoCardStrength(v)
}

// cardStrengthForCard カードオブジェクトの強さを返す (ジョーカー対応)
func (d *Daifugo) cardStrengthForCard(card *Card) int {
	if card.GetDesign() == CardDesignJoker {
		return DaifugoJokerStrength
	}
	return d.cardStrength(card.GetValue())
}

// IsJoker カードがジョーカーかどうか判定
func IsJoker(card *Card) bool {
	return card.GetDesign() == CardDesignJoker
}

// triggerRevolutionIfNeeded 4枚出しで革命が起きるか判定し、起きた場合は革命フラグを切り替えて全プレイヤーの手札を再ソートする
func (d *Daifugo) triggerRevolutionIfNeeded(cards []*Card) {
	if len(cards) < 4 {
		return
	}
	d.revolutionActive = !d.revolutionActive
	for _, p := range d.players {
		if !p.GetIsFinished() {
			p.SortCardsByStrength(d.cardStrengthForCard)
		}
	}
}

// triggerEightCut 8切りチェック: 8が出されたら場をクリア
func (d *Daifugo) triggerEightCut(cards []*Card) bool {
	if !d.config.EightCutEnabled {
		return false
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 8 {
			d.clearTableState()
			return true
		}
	}
	return false
}

// triggerElevenBack 11バックチェック: J(11)が出されたら11バック発動
func (d *Daifugo) triggerElevenBack(cards []*Card) {
	if !d.config.ElevenBackEnabled {
		return
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 11 {
			d.elevenBackActive = !d.elevenBackActive
			// 手札を再ソート
			for _, p := range d.players {
				if !p.GetIsFinished() {
					p.SortCardsByStrength(d.cardStrengthForCard)
				}
			}
			return
		}
	}
}

// updateSuitLock スート縛りの更新
func (d *Daifugo) updateSuitLock(cards []*Card) {
	if !d.config.SuitLockEnabled {
		return
	}
	if d.tableCards == nil || len(d.tableCards) == 0 {
		// 場がクリアだった → 縛りなし、今出したカードのスートを記録のみ
		d.suitLocked = false
		d.lockedSuit = 0
		return
	}
	// 前の場のカードと今出したカードのスートが一致するか確認 (ジョーカー除く)
	prevSuit := d.getNonJokerSuit(d.tableCards)
	newSuit := d.getNonJokerSuit(cards)
	if prevSuit > 0 && newSuit > 0 && prevSuit == newSuit {
		d.suitLocked = true
		d.lockedSuit = prevSuit
	}
}

// getNonJokerSuit カード配列からジョーカー以外のスートを取得 (全て同じスートなら返す、混在なら0)
func (d *Daifugo) getNonJokerSuit(cards []*Card) int {
	suit := 0
	for _, c := range cards {
		if IsJoker(c) {
			continue
		}
		if suit == 0 {
			suit = c.GetDesign()
		} else if suit != c.GetDesign() {
			return 0 // 混在
		}
	}
	return suit
}

// countFinished 既に上がっているプレイヤー数を返す
func (d *Daifugo) countFinished() int {
	cnt := 0
	for _, p := range d.players {
		if p.GetIsFinished() {
			cnt++
		}
	}
	return cnt
}

// getActivePlayerCnt アクティブ (未上がり) プレイヤー数取得
func (d *Daifugo) getActivePlayerCnt() int {
	return len(d.players) - d.countFinished()
}

// getNextActivePlayer fromの次のアクティブなプレイヤーインデックスを取得
func (d *Daifugo) getNextActivePlayer(from int) int {
	for i := 1; i <= DaifugoPlayerCnt; i++ {
		next := (from + i) % DaifugoPlayerCnt
		if !d.players[next].GetIsFinished() {
			return next
		}
	}
	return -1
}

// checkGameEnd ゲーム終了チェック (残り1人以下なら終了)
func (d *Daifugo) checkGameEnd() bool {
	active := d.getActivePlayerCnt()
	if active <= 1 {
		for i, p := range d.players {
			if !p.GetIsFinished() {
				d.finishPlayer(i)
				break
			}
		}
		d.gameEndFlag = true
		return true
	}
	return false
}

// finishPlayer プレイヤーを上がりにしてランクを付与
// ランクは現在の上がり済みプレイヤー数 + 1 として計算する
func (d *Daifugo) finishPlayer(idx int) {
	rank := d.countFinished() + 1
	d.players[idx].SetIsFinished(true)
	d.players[idx].SetRank(rank)
	// 上がったプレイヤーが最後に出したプレイヤーなら場をクリア
	if d.lastPlayPlayerIdx == idx {
		d.clearTableState()
	}
}

// advanceTurn 手番を次のアクティブなプレイヤーへ進める
func (d *Daifugo) advanceTurn() {
	if d.gameEndFlag {
		return
	}
	next := d.getNextActivePlayer(d.currentTurn)
	if next >= 0 {
		d.currentTurn = next
	}
}

// checkPassClear 全員パスしたら場をクリアする
func (d *Daifugo) checkPassClear() {
	if d.tableCards == nil || d.lastPlayPlayerIdx < 0 {
		return
	}
	// 手番が最後に出したプレイヤーに戻ってきたら全員パス
	if d.currentTurn == d.lastPlayPlayerIdx {
		d.clearTableState()
	}
}

// clearTableState 場の状態をクリア (8切り、上がり時等に使用)
func (d *Daifugo) clearTableState() {
	d.tableCards = nil
	d.lastPlayPlayerIdx = -1
	d.passCount = 0
	d.suitLocked = false
	d.lockedSuit = 0
	d.elevenBackActive = false
	d.tableIsSequence = false
}

// getBaseValue ジョーカーを除いたカード配列の共通値を取得 (全ジョーカーなら-1)
func getBaseValue(cards []*Card) int {
	for _, c := range cards {
		if !IsJoker(c) {
			return c.GetValue()
		}
	}
	return -1 // 全てジョーカー
}

// isValidGroup カード配列がグループ (同じ値 + ジョーカーワイルド) として有効かチェック
func isValidGroup(cards []*Card) bool {
	base := getBaseValue(cards)
	if base < 0 {
		// 全てジョーカー → 有効 (ジョーカーだけのグループ)
		return true
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() != base {
			return false
		}
	}
	return true
}

// isValidSequence カード配列が階段 (同スートの連続した値 + ジョーカーワイルド) として有効かチェック
// 3枚以上が必要。革命・11バック状態に応じた強さを使用する。
func (d *Daifugo) isValidSequence(cards []*Card) bool {
	if len(cards) < 3 {
		return false
	}

	// ジョーカー以外のカードを収集
	suit := 0
	nonJokerValues := make([]int, 0)
	jokerCount := 0
	for _, c := range cards {
		if IsJoker(c) {
			jokerCount++
			continue
		}
		if suit == 0 {
			suit = c.GetDesign()
		} else if c.GetDesign() != suit {
			return false // スートが混在
		}
		nonJokerValues = append(nonJokerValues, d.cardStrength(c.GetValue()))
	}

	if len(nonJokerValues) == 0 {
		// 全ジョーカー → 階段としては不成立
		return false
	}

	sort.Ints(nonJokerValues)

	// ジョーカーで穴を埋められるか確認
	gaps := 0
	for i := 1; i < len(nonJokerValues); i++ {
		diff := nonJokerValues[i] - nonJokerValues[i-1]
		if diff == 0 {
			return false // 重複値
		}
		gaps += diff - 1
	}

	// 足りない分をジョーカーで埋める
	return gaps <= jokerCount
}

// getSequenceMinStrength 階段の最小強さ (最弱カードの強さ) を返す
func (d *Daifugo) getSequenceMinStrength(cards []*Card) int {
	minStr := DaifugoJokerStrength + 1
	for _, c := range cards {
		str := d.cardStrengthForCard(c)
		if str < minStr {
			minStr = str
		}
	}
	return minStr
}

// isPlayable 指定したカードが場のカードに対して出せるか判定
func (d *Daifugo) isPlayable(cards []*Card) bool {
	if len(cards) == 0 {
		return false
	}

	// プレイタイプ判定
	validGroup := isValidGroup(cards)
	validSeq := d.config.SequenceEnabled && d.isValidSequence(cards)

	if !validGroup && !validSeq {
		return false
	}

	if d.tableCards == nil {
		// 場がクリアなら何でも出せる
		return true
	}

	// 枚数が一致しているか
	if len(cards) != len(d.tableCards) {
		return false
	}

	// 場が階段の場合
	if d.tableIsSequence {
		if !validSeq {
			return false
		}
		tableMin := d.getSequenceMinStrength(d.tableCards)
		playMin := d.getSequenceMinStrength(cards)
		return playMin > tableMin
	}

	// グループプレイ
	if !validGroup {
		return false
	}

	// スート縛りチェック
	if d.suitLocked && d.config.SuitLockEnabled {
		newSuit := d.getNonJokerSuit(cards)
		if newSuit > 0 && newSuit != d.lockedSuit {
			return false
		}
	}

	// 強さ比較
	tableBase := getBaseValue(d.tableCards)
	playBase := getBaseValue(cards)

	var tableStrength, playStrength int
	if tableBase < 0 {
		tableStrength = DaifugoJokerStrength
	} else {
		tableStrength = d.cardStrength(tableBase)
	}
	if playBase < 0 {
		playStrength = DaifugoJokerStrength
	} else {
		playStrength = d.cardStrength(playBase)
	}
	return playStrength > tableStrength
}

// PlayerPlay 人間プレイヤーがカードを出す (または パスする)
// indices: 出すカードのインデックス。空の場合はパス。
func (d *Daifugo) PlayerPlay(indices []int) error {
	if d.gameEndFlag {
		return ErrGameEnded
	}
	if !d.players[d.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	// 人間のターン開始時にCPU行動履歴をリセット
	d.cpuActions = nil

	if len(indices) == 0 {
		// パス
		d.passCount++
		d.humanAction = &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: nil}
		d.advanceTurn()
		d.checkPassClear()
		return nil
	}

	// 重複インデックスを除去 (重複があると isPlayable の枚数チェックが狂うため)
	{
		cp := make([]int, len(indices))
		copy(cp, indices)
		sort.Ints(cp)
		unique := make([]int, 0, len(cp))
		for i, idx := range cp {
			if i == 0 || idx != cp[i-1] {
				unique = append(unique, idx)
			}
		}
		indices = unique
	}

	// 指定カードを収集
	player := d.players[d.currentTurn]
	selectedCards := make([]*Card, len(indices))
	for i, idx := range indices {
		card := player.GetCard(idx)
		if card == nil {
			return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", idx))
		}
		selectedCards[i] = card
	}
	if !d.isPlayable(selectedCards) {
		return NewDomainError(ErrInvalidPlay, "selected cards cannot be played")
	}

	// スート縛り更新 (場のカードがある場合、出す前にチェック)
	d.updateSuitLock(selectedCards)

	// 階段判定
	isSeq := d.config.SequenceEnabled && d.isValidSequence(selectedCards)

	// カードを出す
	cards := player.RemoveCards(indices)
	d.tableCards = cards
	d.lastPlayPlayerIdx = d.currentTurn
	d.passCount = 0
	d.humanAction = &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: cards}

	d.tableIsSequence = isSeq

	// 革命チェック
	d.triggerRevolutionIfNeeded(cards)

	// 11バックチェック
	d.triggerElevenBack(cards)

	// プレイヤー上がりチェック
	if player.GetCardsSize() == 0 {
		d.finishPlayer(d.currentTurn)
	}

	// 8切りチェック (上がりチェック後に実施)
	eightCut := d.triggerEightCut(cards)

	if !d.checkGameEnd() {
		// 8切り: 出したプレイヤーに手番が戻る (上がっていない場合のみ)
		if !eightCut || player.GetIsFinished() {
			d.advanceTurn()
		}
	}
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (d *Daifugo) CpuPlay() {
	if d.gameEndFlag || d.players[d.currentTurn].GetIsHuman() {
		return
	}
	playerIdx := d.currentTurn
	player := d.players[playerIdx]

	// 出せる最弱のカードセットを探す
	playIndices := d.findBestPlay(player)

	if len(playIndices) == 0 {
		// パス
		d.passCount++
		action := &DaifugoCpuAction{PlayerIdx: playerIdx, PlayedCards: nil}
		d.cpuActions = append(d.cpuActions, action)
		d.advanceTurn()
		d.checkPassClear()
	} else {
		// 出すカードを取得 (スート縛り更新用)
		selectedCards := make([]*Card, len(playIndices))
		for i, idx := range playIndices {
			selectedCards[i] = player.GetCard(idx)
		}

		// スート縛り更新
		d.updateSuitLock(selectedCards)

		// 階段判定
		isSeq := d.config.SequenceEnabled && d.isValidSequence(selectedCards)

		cards := player.RemoveCards(playIndices)
		d.tableCards = cards
		d.lastPlayPlayerIdx = playerIdx
		d.passCount = 0
		action := &DaifugoCpuAction{PlayerIdx: playerIdx, PlayedCards: cards}
		d.cpuActions = append(d.cpuActions, action)

		d.tableIsSequence = isSeq

		d.triggerRevolutionIfNeeded(cards)
		d.triggerElevenBack(cards)

		if player.GetCardsSize() == 0 {
			d.finishPlayer(playerIdx)
		}

		eightCut := d.triggerEightCut(cards)

		if !d.checkGameEnd() {
			// 8切り: 出したプレイヤーに手番が戻る (上がっていない場合のみ)
			if !eightCut || player.GetIsFinished() {
				d.advanceTurn()
			}
		}
	}
}

// findBestPlay プレイヤーが出せる最弱のカードセットのインデックスを返す
// 出せるカードがない場合は nil を返す
func (d *Daifugo) findBestPlay(player *DaifugoPlayer) []int {
	if d.tableCards == nil {
		// 場がクリアなら最弱の1枚を出す (ジョーカーは温存)
		for i := 0; i < player.GetCardsSize(); i++ {
			if !IsJoker(player.GetCard(i)) {
				return []int{i}
			}
		}
		// ジョーカーしかない場合
		if player.GetCardsSize() > 0 {
			return []int{0}
		}
		return nil
	}

	// 場が階段の場合
	if d.tableIsSequence && d.config.SequenceEnabled {
		return d.findBestSequencePlay(player)
	}

	needed := len(d.tableCards)
	tableBase := getBaseValue(d.tableCards)
	var tableStrength int
	if tableBase < 0 {
		tableStrength = DaifugoJokerStrength
	} else {
		tableStrength = d.cardStrength(tableBase)
	}

	// 手札は強さ順でソート済み。同じ値の連続するグループを探す。
	jokerIndices := d.findJokerIndices(player)
	i := 0
	for i < player.GetCardsSize() {
		card := player.GetCard(i)
		if IsJoker(card) {
			i++
			continue
		}
		v := card.GetValue()
		j := i
		for j < player.GetCardsSize() && !IsJoker(player.GetCard(j)) && player.GetCard(j).GetValue() == v {
			j++
		}
		count := j - i
		if count >= needed && d.cardStrength(v) > tableStrength {
			// スート縛りチェック
			if d.suitLocked && d.config.SuitLockEnabled {
				suit := player.GetCard(i).GetDesign()
				if suit != d.lockedSuit {
					i = j
					continue
				}
			}
			indices := make([]int, needed)
			for k := 0; k < needed; k++ {
				indices[k] = i + k
			}
			return indices
		}
		// ジョーカーで補完してグループを作れるか
		if count < needed && count > 0 && d.cardStrength(v) > tableStrength {
			if count+len(jokerIndices) >= needed {
				// スート縛りチェック
				if d.suitLocked && d.config.SuitLockEnabled {
					suit := player.GetCard(i).GetDesign()
					if suit != d.lockedSuit {
						i = j
						continue
					}
				}
				indices := make([]int, 0, needed)
				for k := 0; k < count && len(indices) < needed; k++ {
					indices = append(indices, i+k)
				}
				for _, ji := range jokerIndices {
					if len(indices) >= needed {
						break
					}
					indices = append(indices, ji)
				}
				sort.Ints(indices)
				return indices
			}
		}
		i = j
	}

	// ジョーカー単体で出す (場が1枚で、ジョーカーの方が強い場合)
	if needed == 1 {
		for i := 0; i < player.GetCardsSize(); i++ {
			if IsJoker(player.GetCard(i)) && DaifugoJokerStrength > tableStrength {
				return []int{i}
			}
		}
	}

	return nil
}

// findJokerIndices プレイヤーの手札中のジョーカーのインデックスを返す
func (d *Daifugo) findJokerIndices(player *DaifugoPlayer) []int {
	indices := make([]int, 0)
	for i := 0; i < player.GetCardsSize(); i++ {
		if IsJoker(player.GetCard(i)) {
			indices = append(indices, i)
		}
	}
	return indices
}

// findBestSequencePlay 階段モードで出せる最弱の階段を探す (ジョーカーで穴を埋められる)
func (d *Daifugo) findBestSequencePlay(player *DaifugoPlayer) []int {
	needed := len(d.tableCards)
	tableMinStr := d.getSequenceMinStrength(d.tableCards)
	jokerIndices := d.findJokerIndices(player)

	// 同スートの連続カードを探す
	for startIdx := 0; startIdx < player.GetCardsSize(); startIdx++ {
		card := player.GetCard(startIdx)
		if IsJoker(card) {
			continue
		}
		suit := card.GetDesign()
		startStrength := d.cardStrengthForCard(card)

		// startIdx から始まる同スートカードを強さ順に収集
		suitCards := []struct {
			idx      int
			strength int
		}{{startIdx, startStrength}}

		for nextIdx := startIdx + 1; nextIdx < player.GetCardsSize(); nextIdx++ {
			nextCard := player.GetCard(nextIdx)
			if IsJoker(nextCard) {
				continue
			}
			if nextCard.GetDesign() != suit {
				continue
			}
			nextStr := d.cardStrengthForCard(nextCard)
			if nextStr <= startStrength {
				continue
			}
			suitCards = append(suitCards, struct {
				idx      int
				strength int
			}{nextIdx, nextStr})
		}

		// suitCards + jokers で needed 枚の階段を構築できるか試みる
		// suitCards のサブセットを起点に、ジョーカーで穴を埋める
		for si := 0; si < len(suitCards); si++ {
			indices := []int{suitCards[si].idx}
			lastStr := suitCards[si].strength
			jokersUsed := 0
			sci := si + 1

			for len(indices) < needed {
				targetStr := lastStr + 1
				// 次の同スートカードがtargetStrを持つか探す
				found := false
				for sci < len(suitCards) {
					if suitCards[sci].strength == targetStr {
						indices = append(indices, suitCards[sci].idx)
						lastStr = targetStr
						sci++
						found = true
						break
					} else if suitCards[sci].strength > targetStr {
						break
					}
					sci++
				}
				if !found {
					// ジョーカーで埋める
					if jokersUsed < len(jokerIndices) {
						indices = append(indices, jokerIndices[jokersUsed])
						jokersUsed++
						lastStr = targetStr
					} else {
						break // ジョーカーが足りない
					}
				}
			}

			if len(indices) == needed {
				// この階段が場より強いかチェック
				testCards := make([]*Card, len(indices))
				for i, idx := range indices {
					testCards[i] = player.GetCard(idx)
				}
				minStr := d.getSequenceMinStrength(testCards)
				if minStr > tableMinStr {
					sort.Ints(indices)
					return indices
				}
			}
		}
	}
	return nil
}

// IsHumanTurn 現在の手番が人間かどうか
func (d *Daifugo) IsHumanTurn() bool {
	return d.players[d.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (d *Daifugo) GetCurrentTurn() int { return d.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (d *Daifugo) GetGameEndFlag() bool { return d.gameEndFlag }

// GetTableCards 場のカード取得 (nil = クリア)
func (d *Daifugo) GetTableCards() []*Card { return d.tableCards }

// GetLastPlayPlayerIdx 最後にカードを出したプレイヤーインデックス取得 (-1 = なし)
func (d *Daifugo) GetLastPlayPlayerIdx() int { return d.lastPlayPlayerIdx }

// GetPlayer プレイヤー取得
func (d *Daifugo) GetPlayer(idx int) *DaifugoPlayer {
	if idx < 0 || idx >= len(d.players) {
		return nil
	}
	return d.players[idx]
}

// GetPlayerCnt プレイヤー数取得
func (d *Daifugo) GetPlayerCnt() int { return len(d.players) }

// GetCpuActions CPUターンの行動履歴取得
func (d *Daifugo) GetCpuActions() []*DaifugoCpuAction { return d.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (d *Daifugo) GetHumanAction() *DaifugoCpuAction { return d.humanAction }

// GetPassCount 現在のパスカウント取得
func (d *Daifugo) GetPassCount() int { return d.passCount }

// GetRevolutionActive 革命フラグ取得
func (d *Daifugo) GetRevolutionActive() bool { return d.revolutionActive }

// GetConfig ローカルルール設定取得
func (d *Daifugo) GetConfig() DaifugoConfig { return d.config }

// GetSuitLocked スート縛り発動中か取得
func (d *Daifugo) GetSuitLocked() bool { return d.suitLocked }

// GetLockedSuit 縛られているスート取得
func (d *Daifugo) GetLockedSuit() int { return d.lockedSuit }

// GetElevenBackActive 11バック発動中か取得
func (d *Daifugo) GetElevenBackActive() bool { return d.elevenBackActive }

// GetTableIsSequence 場が階段プレイか取得
func (d *Daifugo) GetTableIsSequence() bool { return d.tableIsSequence }

// GetExchangeActions カード交換記録取得
func (d *Daifugo) GetExchangeActions() []*DaifugoExchangeAction { return d.exchangeActions }
