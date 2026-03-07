package domain

import (
	"fmt"
	"math"
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

// DaifugoPendingAction ペンディングアクションの種類
type DaifugoPendingAction int

const (
	DaifugoPendingNone       DaifugoPendingAction = 0 // ペンディングなし
	DaifugoPendingSevenPass  DaifugoPendingAction = 1 // 7渡し待ち
	DaifugoPendingTenDiscard DaifugoPendingAction = 2 // 10捨て待ち
)

// DaifugoSortMode 手札ソートモード
type DaifugoSortMode int

const (
	DaifugoSortByStrength DaifugoSortMode = 0 // 強さ順 (デフォルト)
	DaifugoSortBySuit     DaifugoSortMode = 1 // スート順
	DaifugoSortByNumber   DaifugoSortMode = 2 // 数字順

	// jokerSortWeight ジョーカーをソート末尾に配置するための重み (最大値 4*100+13=413 を十分超える)
	jokerSortWeight = 10000
)

// DaifugoCpuDifficulty CPU難易度レベル
type DaifugoCpuDifficulty int

const (
	DaifugoDifficultyNormal DaifugoCpuDifficulty = 0 // 通常 (デフォルト、既存ロジック)
	DaifugoDifficultyEasy   DaifugoCpuDifficulty = 1 // 簡単 (単純なグリーディ)
	DaifugoDifficultyHard   DaifugoCpuDifficulty = 2 // 難しい (ヒューリスティックAI)
)

// DaifugoDifficultyNames 難易度名マップ
var DaifugoDifficultyNames = map[DaifugoCpuDifficulty]string{
	DaifugoDifficultyNormal: "Normal",
	DaifugoDifficultyEasy:   "Easy",
	DaifugoDifficultyHard:   "Hard",
}

// DaifugoConfig 大富豪ローカルルール設定
type DaifugoConfig struct {
	JokerCount                int                  // ジョーカー枚数 (default: 2)
	EightCutEnabled           bool                 // 8切り
	SuitLockEnabled           bool                 // スート縛り
	ElevenBackEnabled         bool                 // 11バック
	SequenceEnabled           bool                 // 階段
	CardExchangeEnabled       bool                 // カード交換
	FiveSkipEnabled           bool                 // 5飛び
	SevenPassEnabled          bool                 // 7渡し
	TenDiscardEnabled         bool                 // 10捨て
	SpadeThreeEnabled         bool                 // スペ3返し
	CapitalFallEnabled        bool                 // 都落ち
	NineReverseEnabled        bool                 // 9リバース
	CoupDetatEnabled          bool                 // クーデター (3枚の9で革命)
	IntenseLockEnabled        bool                 // 激シバ (連番縛り)
	SandstormEnabled          bool                 // 砂嵐 (3枚の3で場をクリア)
	EmperorEnabled            bool                 // エンペラー (4枚連番・全スート異なる→革命+場クリア)
	SequenceRevolutionEnabled bool                 // 階段革命 (4枚以上の階段で革命)
	IllegalFinishEnabled      bool                 // 反則上がり (8切り/ジョーカー/革命で上がりはペナルティ)
	CpuDifficulty             DaifugoCpuDifficulty // CPU難易度 (0=Normal, 1=Easy, 2=Hard)
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
		FiveSkipEnabled:     false,
		SevenPassEnabled:    false,
		TenDiscardEnabled:   false,
		SpadeThreeEnabled:   false,
		CapitalFallEnabled:  false,
		NineReverseEnabled:  false,
		CoupDetatEnabled:    false,
		IntenseLockEnabled:  false,
		SandstormEnabled:    false,
		EmperorEnabled:      false,
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
	trumpCards          *TrumpCards
	players             []*DaifugoPlayer
	currentTurn         int                      // 現在の手番プレイヤーインデックス
	tableCards          []*Card                  // 場に出されているカード (nil = 場はクリア)
	lastPlayPlayerIdx   int                      // 最後にカードを出したプレイヤーインデックス (-1 = なし)
	gameEndFlag         bool                     // ゲーム終了フラグ
	passCount           int                      // 最後の出し以降の連続パス数
	cpuActions          []*DaifugoCpuAction      // 人間ターン後のCPUの行動履歴
	humanAction         *DaifugoCpuAction        // 人間の最後の行動
	revolutionActive    bool                     // 革命フラグ (true = 革命中)
	config              DaifugoConfig            // ローカルルール設定
	suitLocked          bool                     // スート縛り発動中
	lockedSuit          int                      // 縛られているスート (CardDesignSpade等)
	elevenBackActive    bool                     // 11バック発動中
	tableIsSequence     bool                     // 場が階段プレイか
	exchangeActions     []*DaifugoExchangeAction // カード交換記録
	pendingActionType   DaifugoPendingAction     // ペンディングアクションの種類
	pendingActionTarget int                      // 7渡しの対象プレイヤーインデックス (-1 = なし)
	reverseDirection    bool                     // 9リバース: ターン方向が逆か
	numberLocked        bool                     // 激シバ: 連番縛り発動中
	sortMode            DaifugoSortMode          // 手札ソートモード
}

// NewDaifugo コンストラクタ
func NewDaifugo(trumpCards *TrumpCards, players []*DaifugoPlayer, config DaifugoConfig) *Daifugo {
	return &Daifugo{
		trumpCards:          trumpCards,
		players:             players,
		currentTurn:         0,
		tableCards:          nil,
		lastPlayPlayerIdx:   -1,
		gameEndFlag:         false,
		passCount:           0,
		cpuActions:          nil,
		humanAction:         nil,
		revolutionActive:    false,
		config:              config,
		suitLocked:          false,
		lockedSuit:          0,
		elevenBackActive:    false,
		tableIsSequence:     false,
		exchangeActions:     nil,
		pendingActionType:   DaifugoPendingNone,
		pendingActionTarget: -1,
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
	d.pendingActionType = DaifugoPendingNone
	d.pendingActionTarget = -1
	d.reverseDirection = false
	d.numberLocked = false
	// sortMode は意図的にリセットしない: ユーザーの好みをラウンド間で維持する

	// シャッフル
	d.trumpCards.Shuffle()

	// 全プレイヤーのカードリセット
	resetPlayers(d.players, func(p *DaifugoPlayer) {
		p.SetRank(-1)
		p.SetIllegalFinishPenalty(false)
	})

	// プレイ順をランダムにする
	rand.Shuffle(len(d.players), func(i, j int) {
		d.players[i], d.players[j] = d.players[j], d.players[i]
	})

	// 全カードを配る (ジョーカー含む)
	dealAllCards(d.trumpCards, d.players)

	// 各プレイヤーの手札をソート
	d.sortAllActiveHands()

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
	d.sortAllActiveHands()
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
// isSeq が true の場合は階段革命ルール (SequenceRevolutionEnabled) が有効な時のみ発動する
func (d *Daifugo) triggerRevolutionIfNeeded(cards []*Card, isSeq bool) {
	if len(cards) < 4 {
		return
	}
	if isSeq && !d.config.SequenceRevolutionEnabled {
		return
	}
	d.revolutionActive = !d.revolutionActive
	d.sortAllActiveHands()
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

// triggerSandstorm 砂嵐チェック: 3枚の非ジョーカー3が出されたら場をクリア
func (d *Daifugo) triggerSandstorm(cards []*Card) bool {
	if !d.config.SandstormEnabled {
		return false
	}
	if len(cards) != 3 {
		return false
	}
	for _, c := range cards {
		if IsJoker(c) || c.GetValue() != 3 {
			return false
		}
	}
	d.clearTableState()
	return true
}

// isValidEmperor エンペラー判定: 4枚の連番カード(全スート異なる)を場がクリアの時に出す
func (d *Daifugo) isValidEmperor(cards []*Card) bool {
	if !d.config.EmperorEnabled || len(cards) != 4 || d.tableCards != nil {
		return false
	}
	return d.isEmperorCards(cards)
}

// isEmperorCards カードがエンペラー条件を満たすか判定 (場の状態は見ない)
func (d *Daifugo) isEmperorCards(cards []*Card) bool {
	if len(cards) != 4 {
		return false
	}
	suits := make(map[int]bool)
	nonJokerValues := make([]int, 0, 4)
	jokerCount := 0
	for _, c := range cards {
		if IsJoker(c) {
			jokerCount++
			continue
		}
		if suits[c.GetDesign()] {
			return false // 同じスートが重複
		}
		suits[c.GetDesign()] = true
		nonJokerValues = append(nonJokerValues, d.cardStrength(c.GetValue()))
	}
	if len(nonJokerValues) == 0 {
		return false // 全ジョーカーは不可
	}
	sort.Ints(nonJokerValues)
	// 非ジョーカーの値が連続しているか確認
	gaps := 0
	for i := 1; i < len(nonJokerValues); i++ {
		diff := nonJokerValues[i] - nonJokerValues[i-1]
		if diff == 0 {
			return false // 重複値
		}
		gaps += diff - 1
	}
	// 非ジョーカー間の穴 + 両端への拡張で合計4枚になるか
	// 全体のスパン = max - min + 1 + 端に追加するジョーカー
	span := nonJokerValues[len(nonJokerValues)-1] - nonJokerValues[0] + 1
	remaining := jokerCount - gaps
	if remaining < 0 {
		return false // 穴を埋められない
	}
	totalSpan := span + remaining
	return totalSpan == 4
}

// triggerEmperor エンペラー発動: 革命を起こし場をクリア
func (d *Daifugo) triggerEmperor(cards []*Card) bool {
	if !d.config.EmperorEnabled {
		return false
	}
	if !d.isEmperorCards(cards) {
		return false
	}
	d.revolutionActive = !d.revolutionActive
	d.sortAllActiveHands()
	d.clearTableState()
	return true
}

// triggerElevenBack 11バックチェック: J(11)が出されたら11バック発動
func (d *Daifugo) triggerElevenBack(cards []*Card) {
	if !d.config.ElevenBackEnabled {
		return
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 11 {
			d.elevenBackActive = !d.elevenBackActive
			d.sortAllActiveHands()
			return
		}
	}
}

// updateSuitLock スート縛りの更新
func (d *Daifugo) updateSuitLock(cards []*Card) {
	if !d.config.SuitLockEnabled {
		return
	}
	if len(d.tableCards) == 0 {
		// 場がクリアだった → 縛りなし、今出したカードのスートを記録のみ
		d.suitLocked = false
		d.lockedSuit = 0
		return
	}
	// 前の場のカードと今出したカードのスートが一致するか確認 (ジョーカー除く)
	prevSuit := d.getNonJokerSuit(d.tableCards)
	newSuit := d.getNonJokerSuit(cards)
	if prevSuit > 0 && newSuit > 0 && prevSuit == newSuit {
		if !d.suitLocked {
			d.suitLocked = true
			d.lockedSuit = prevSuit
			// 激シバ: スート縛り発動時に連番かチェック
			if d.config.IntenseLockEnabled {
				prevBase := getBaseValue(d.tableCards)
				newBase := getBaseValue(cards)
				if prevBase > 0 && newBase > 0 {
					prevStr := d.cardStrength(prevBase)
					newStr := d.cardStrength(newBase)
					if newStr-prevStr == 1 {
						d.numberLocked = true
					}
				}
			}
		}
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
	return countPlayers(d.players, func(p *DaifugoPlayer) bool { return p.GetIsFinished() })
}

// getActivePlayerCnt アクティブ (未上がり) プレイヤー数取得
func (d *Daifugo) getActivePlayerCnt() int {
	return len(d.players) - d.countFinished()
}

// getNextActivePlayer fromの次のアクティブなプレイヤーインデックスを取得
func (d *Daifugo) getNextActivePlayer(from int) int {
	direction := 1
	if d.reverseDirection {
		direction = -1
	}
	return nextActivePlayer(d.players, from, direction)
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
		d.applyCapitalFall()
		d.applyIllegalFinishPenalty()
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
	d.numberLocked = false
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

	// スペ3返し特例: 場がジョーカー1枚のみ && スペードの3を1枚出す
	if d.isSpadeThreeCounter(cards) {
		return true
	}

	// プレイタイプ判定
	validGroup := isValidGroup(cards)
	validSeq := d.config.SequenceEnabled && d.isValidSequence(cards)
	validEmperor := d.isValidEmperor(cards)

	if !validGroup && !validSeq && !validEmperor {
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

	// 激シバ: 連番縛り発動中は強さの差が1でなければ出せない
	// ジョーカーは連番縛りをバイパスし、通常の強さ比較のみ適用
	if d.numberLocked && d.config.IntenseLockEnabled && d.config.SuitLockEnabled {
		if playBase > 0 && tableBase > 0 {
			return playStrength-tableStrength == 1
		}
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

	// ペンディングアクションがある場合はそちらを先に解決
	if d.pendingActionType != DaifugoPendingNone {
		return d.resolvePendingAction(indices)
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

	// スペ3返し判定 (isPlayable でチェック済み、ここで行動フラグを取得)
	spadeThree := d.isSpadeThreeCounter(selectedCards)

	// カードを出す
	cards := player.RemoveCards(indices)
	d.humanAction = &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: cards}
	d.playCards(d.currentTurn, cards, isSeq, spadeThree)
	return nil
}

// playCards はカードプレイ後の共通処理を実行する
func (d *Daifugo) playCards(playerIdx int, cards []*Card, isSeq bool, spadeThree bool) {
	d.tableCards = cards
	d.lastPlayPlayerIdx = playerIdx
	d.passCount = 0
	d.tableIsSequence = isSeq

	emperor := d.triggerEmperor(cards)
	if !emperor {
		d.triggerRevolutionIfNeeded(cards, isSeq)
	}
	d.triggerCoupDetatIfNeeded(cards)
	d.triggerElevenBack(cards)
	d.triggerNineReverseIfNeeded(cards, isSeq)

	if d.players[playerIdx].GetCardsSize() == 0 {
		if d.isIllegalFinish(cards, isSeq) {
			d.players[playerIdx].SetIllegalFinishPenalty(true)
		}
		d.finishPlayer(playerIdx)
	}

	eightCut := d.triggerEightCut(cards)
	sandstorm := d.triggerSandstorm(cards)

	fiveSkip := d.triggerFiveSkipIfNeeded(cards, isSeq)
	d.triggerSevenPassIfNeeded(cards, isSeq)
	d.triggerTenDiscardIfNeeded(cards, isSeq)

	if d.pendingActionType != DaifugoPendingNone {
		return
	}

	if !d.checkGameEnd() {
		if (!eightCut && !sandstorm && !emperor && !spadeThree) || d.players[playerIdx].GetIsFinished() {
			d.advanceTurn()
			if fiveSkip && !d.gameEndFlag {
				d.advanceTurn()
			}
			d.checkPassClear()
		}
	}
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (d *Daifugo) CpuPlay() {
	if d.gameEndFlag || d.players[d.currentTurn].GetIsHuman() {
		return
	}

	// ペンディングアクションがある場合はCPUが自動解決
	if d.pendingActionType != DaifugoPendingNone {
		d.cpuResolvePendingAction()
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

		// スペ3返し判定
		spadeThree := d.isSpadeThreeCounter(selectedCards)

		// スート縛り更新
		d.updateSuitLock(selectedCards)

		// 階段判定
		isSeq := d.config.SequenceEnabled && d.isValidSequence(selectedCards)

		cards := player.RemoveCards(playIndices)
		action := &DaifugoCpuAction{PlayerIdx: playerIdx, PlayedCards: cards}
		d.cpuActions = append(d.cpuActions, action)
		d.playCards(playerIdx, cards, isSeq, spadeThree)
	}
}

// wouldCauseIllegalFinish 指定カードで上がった場合に反則上がりになるか判定 (CPU AI用)
func (d *Daifugo) wouldCauseIllegalFinish(player *DaifugoPlayer, indices []int) bool {
	if !d.config.IllegalFinishEnabled || len(indices) == 0 {
		return false
	}
	if player.GetCardsSize() != len(indices) {
		return false // 上がりにならない
	}
	cards := make([]*Card, len(indices))
	for i, idx := range indices {
		cards[i] = player.GetCard(idx)
	}
	isSeq := d.config.SequenceEnabled && d.isValidSequence(cards)
	return d.isIllegalFinish(cards, isSeq)
}

// findBestPlay 難易度に応じたCPUプレイ戦略のディスパッチャー
func (d *Daifugo) findBestPlay(player *DaifugoPlayer) []int {
	switch d.config.CpuDifficulty {
	case DaifugoDifficultyEasy:
		return d.findBestPlayEasy(player)
	case DaifugoDifficultyHard:
		return d.findBestPlayHard(player)
	default:
		return d.findBestPlayNormal(player)
	}
}

// opponentMinCards 対戦相手の中で最小の手札枚数を返す
func (d *Daifugo) opponentMinCards(player *DaifugoPlayer) int {
	minCards := math.MaxInt
	for _, p := range d.players {
		if p == player || p.GetIsFinished() {
			continue
		}
		if p.GetCardsSize() < minCards {
			minCards = p.GetCardsSize()
		}
	}
	return minCards
}

// isUrgent 対戦相手が3枚以下の手札を持っているか (緊急モード判定)
func (d *Daifugo) isUrgent(player *DaifugoPlayer) bool {
	return d.opponentMinCards(player) <= 3
}

// shouldStrategicPass Hard AIの戦略的パス判定
// 場の強さが10以下で、出そうとしているカードがA(14)以上の強さで、手札が6枚以上の場合にパス
func (d *Daifugo) shouldStrategicPass(player *DaifugoPlayer, indices []int) bool {
	if player.GetCardsSize() <= 5 {
		return false
	}
	tableBase := getBaseValue(d.tableCards)
	var tableStrength int
	if tableBase < 0 {
		tableStrength = DaifugoJokerStrength
	} else {
		tableStrength = d.cardStrength(tableBase)
	}
	if tableStrength > 10 {
		return false
	}
	// 出そうとしているカードの強さを確認
	for _, idx := range indices {
		card := player.GetCard(idx)
		if IsJoker(card) {
			return true // ジョーカーは温存
		}
		if d.cardStrength(card.GetValue()) >= DaifugoCardStrength(1) { // A以上 (革命に依存しない固定閾値)
			return true
		}
	}
	return false
}

// calcTableStrength 場のカードの強さを計算する
func (d *Daifugo) calcTableStrength() int {
	tableBase := getBaseValue(d.tableCards)
	if tableBase < 0 {
		return DaifugoJokerStrength
	}
	return d.cardStrength(tableBase)
}

// cardSearchOpts カード検索のオプション
type cardSearchOpts struct {
	skipRevolution  bool // 革命防止: 4枚以上のグループをスキップ (Normal用)
	selectStrongest bool // 最強のグループを選択 (Hard urgent用)
}

// searchCardGroup 手札からカードグループを検索する共通ヘルパー
func (d *Daifugo) searchCardGroup(player *DaifugoPlayer, needed int, tableStrength int, opts cardSearchOpts) []int {
	jokerIndices := d.findJokerIndices(player)
	var bestIndices []int
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
			// 革命防止: 4枚以上かつ革命未発動かつ出した後も手札が残る場合はスキップ
			if opts.skipRevolution && count >= 4 && !d.revolutionActive && player.GetCardsSize() > count {
				i = j
				continue
			}
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
			if !opts.selectStrongest {
				return indices
			}
			bestIndices = indices
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
				if !opts.selectStrongest {
					return indices
				}
				bestIndices = indices
			}
		}
		i = j
	}
	return bestIndices
}

// findBestPlayNormal 通常難易度: 既存ロジック (最弱のカードを出す、8/ジョーカー温存)
func (d *Daifugo) findBestPlayNormal(player *DaifugoPlayer) []int {
	if d.tableCards == nil {
		// エンペラーを探す (場がクリアの時のみ)
		if emperorIndices := d.findEmperorPlay(player); emperorIndices != nil {
			if !d.wouldCauseIllegalFinish(player, emperorIndices) {
				return emperorIndices
			}
		}
		// 場がクリアなら最弱の1枚を出す (ジョーカーと8は温存)
		var fallbackIdx *int
		for i := 0; i < player.GetCardsSize(); i++ {
			card := player.GetCard(i)
			if !IsJoker(card) && card.GetValue() != 8 {
				if !d.wouldCauseIllegalFinish(player, []int{i}) {
					return []int{i}
				}
				if fallbackIdx == nil {
					v := i
					fallbackIdx = &v
				}
			}
		}
		// 8以外の非ジョーカーがない: 8を使う (ジョーカーより優先)
		for i := 0; i < player.GetCardsSize(); i++ {
			if !IsJoker(player.GetCard(i)) && player.GetCard(i).GetValue() == 8 {
				if !d.wouldCauseIllegalFinish(player, []int{i}) {
					return []int{i}
				}
				if fallbackIdx == nil {
					v := i
					fallbackIdx = &v
				}
			}
		}
		// ジョーカーしかない場合
		if player.GetCardsSize() > 0 {
			if !d.wouldCauseIllegalFinish(player, []int{0}) {
				return []int{0}
			}
			if fallbackIdx == nil {
				v := 0
				fallbackIdx = &v
			}
		}
		// 全ての選択肢が反則上がりの場合はfallbackを使用 (ペナルティ受け入れ)
		if fallbackIdx != nil {
			return []int{*fallbackIdx}
		}
		return nil
	}

	// 場が階段の場合
	if d.tableIsSequence && d.config.SequenceEnabled {
		return d.findBestSequencePlay(player)
	}

	needed := len(d.tableCards)
	tableStrength := d.calcTableStrength()

	// 手札は強さ順でソート済み。最弱のグループを探す (革命防止あり)。
	if indices := d.searchCardGroup(player, needed, tableStrength, cardSearchOpts{skipRevolution: true}); indices != nil {
		return indices
	}

	// ジョーカー単体で出す (場が1枚で、ジョーカーの方が強い場合)
	if needed == 1 {
		// 戦略的パス: 場が2またはジョーカーで手札が4枚以上の場合はジョーカーを温存
		if tableStrength >= DaifugoCardStrength(2) && player.GetCardsSize() > 3 {
			return nil
		}
		for i := 0; i < player.GetCardsSize(); i++ {
			if IsJoker(player.GetCard(i)) && DaifugoJokerStrength > tableStrength {
				return []int{i}
			}
		}
	}

	return nil
}

// findBestPlayEasy 簡単難易度: 単純に出せる最弱のカードを出す (8/ジョーカー温存なし、エンペラー探索なし、革命防止なし)
func (d *Daifugo) findBestPlayEasy(player *DaifugoPlayer) []int {
	if d.tableCards == nil {
		// 場がクリアなら最弱の1枚を出す (温存戦略なし、反則上がりチェックなし: Easy AIは戦略なしで失敗もする)
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
	tableStrength := d.calcTableStrength()

	// 最弱のグループを探す (革命防止なし、ジョーカー温存なし)
	if indices := d.searchCardGroup(player, needed, tableStrength, cardSearchOpts{}); indices != nil {
		return indices
	}

	// ジョーカー単体 (温存判定なし)
	if needed == 1 {
		for i := 0; i < player.GetCardsSize(); i++ {
			if IsJoker(player.GetCard(i)) && DaifugoJokerStrength > tableStrength {
				return []int{i}
			}
		}
	}

	return nil
}

// findBestPlayHard 難しい難易度: 対戦相手の手札状況を考慮したヒューリスティックAI
func (d *Daifugo) findBestPlayHard(player *DaifugoPlayer) []int {
	urgent := d.isUrgent(player)

	if d.tableCards == nil {
		// エンペラーを探す
		if emperorIndices := d.findEmperorPlay(player); emperorIndices != nil {
			if !d.wouldCauseIllegalFinish(player, emperorIndices) {
				return emperorIndices
			}
		}

		if urgent {
			// 緊急時: 最強の非ジョーカーカードを出す (手札末尾から探索)
			var fallbackIdx *int
			for i := player.GetCardsSize() - 1; i >= 0; i-- {
				card := player.GetCard(i)
				if !IsJoker(card) {
					if !d.wouldCauseIllegalFinish(player, []int{i}) {
						return []int{i}
					}
					if fallbackIdx == nil {
						v := i
						fallbackIdx = &v
					}
				}
			}
			// ジョーカーしかない場合
			if player.GetCardsSize() > 0 {
				if !d.wouldCauseIllegalFinish(player, []int{0}) {
					return []int{0}
				}
				if fallbackIdx == nil {
					v := 0
					fallbackIdx = &v
				}
			}
			if fallbackIdx != nil {
				return []int{*fallbackIdx}
			}
			return nil
		}

		// 非緊急時: Normalと同じ (最弱の1枚、8/ジョーカー温存)
		return d.findBestPlayNormal(player)
	}

	// 場が階段の場合
	if d.tableIsSequence && d.config.SequenceEnabled {
		return d.findBestSequencePlayHard(player)
	}

	needed := len(d.tableCards)
	tableStrength := d.calcTableStrength()

	if urgent {
		// 緊急時: 最強のグループを出す
		if indices := d.searchCardGroup(player, needed, tableStrength, cardSearchOpts{selectStrongest: true}); indices != nil {
			return indices
		}
		// ジョーカー単体 (緊急時は温存しない)
		if needed == 1 {
			for i := 0; i < player.GetCardsSize(); i++ {
				if IsJoker(player.GetCard(i)) && DaifugoJokerStrength > tableStrength {
					return []int{i}
				}
			}
		}
		return nil
	}

	// 非緊急時: 最弱のグループを探し、戦略的パスを検討
	normalIndices := d.findBestPlayNormal(player)
	if normalIndices != nil && d.shouldStrategicPass(player, normalIndices) {
		return nil
	}
	return normalIndices
}

// findBestSequencePlayHard Hard AIの階段モード
func (d *Daifugo) findBestSequencePlayHard(player *DaifugoPlayer) []int {
	if !d.isUrgent(player) {
		// 非緊急時: 通常の最弱階段
		return d.findBestSequencePlay(player)
	}
	// 緊急時: 最強の階段を出す
	needed := len(d.tableCards)
	tableMinStr := d.getSequenceMinStrength(d.tableCards)
	jokerIndices := d.findJokerIndices(player)

	var bestIndices []int
	for startIdx := 0; startIdx < player.GetCardsSize(); startIdx++ {
		card := player.GetCard(startIdx)
		if IsJoker(card) {
			continue
		}
		suit := card.GetDesign()
		startStrength := d.cardStrengthForCard(card)

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

		for si := 0; si < len(suitCards); si++ {
			indices := []int{suitCards[si].idx}
			lastStr := suitCards[si].strength
			jokersUsed := 0
			sci := si + 1

			for len(indices) < needed {
				targetStr := lastStr + 1
				found := false
				for sci < len(suitCards) {
					if suitCards[sci].strength == targetStr {
						indices = append(indices, suitCards[sci].idx)
						lastStr = targetStr
						sci++
						found = true
					}
					// suitCards are in ascending strength order, so once strength >= targetStr, stop
					break
				}
				if !found {
					if jokersUsed < len(jokerIndices) {
						indices = append(indices, jokerIndices[jokersUsed])
						jokersUsed++
						lastStr = targetStr
					} else {
						break
					}
				}
			}

			if len(indices) == needed {
				testCards := make([]*Card, len(indices))
				for i, idx := range indices {
					testCards[i] = player.GetCard(idx)
				}
				minStr := d.getSequenceMinStrength(testCards)
				if minStr > tableMinStr {
					sort.Ints(indices)
					bestIndices = indices // 後の方が強い → 上書き
				}
			}
		}
	}
	return bestIndices
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
					}
					// suitCards are in ascending strength order, so once strength >= targetStr, stop
					break
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

// findEmperorPlay エンペラーの組み合わせを探す (場がクリアの時のみ)
func (d *Daifugo) findEmperorPlay(player *DaifugoPlayer) []int {
	if !d.config.EmperorEnabled || d.tableCards != nil {
		return nil
	}
	n := player.GetCardsSize()
	if n < 4 {
		return nil
	}
	// 全4枚の組み合わせを探索 (C(n,4))
	for a := 0; a < n-3; a++ {
		for b := a + 1; b < n-2; b++ {
			for c := b + 1; c < n-1; c++ {
				for dd := c + 1; dd < n; dd++ {
					testCards := []*Card{
						player.GetCard(a),
						player.GetCard(b),
						player.GetCard(c),
						player.GetCard(dd),
					}
					if d.isEmperorCards(testCards) {
						return []int{a, b, c, dd}
					}
				}
			}
		}
	}
	return nil
}

// triggerFiveSkipIfNeeded 5飛びチェック: 非ジョーカーの5が出されたら次のプレイヤーをスキップ（階段時は無効）
func (d *Daifugo) triggerFiveSkipIfNeeded(cards []*Card, isSeq bool) bool {
	if !d.config.FiveSkipEnabled || isSeq {
		return false
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 5 {
			return true
		}
	}
	return false
}

// triggerSevenPassIfNeeded 7渡しチェック: 非ジョーカーの7が出されたらペンディングアクションをセット
func (d *Daifugo) triggerSevenPassIfNeeded(cards []*Card, isSeq bool) {
	if !d.config.SevenPassEnabled || isSeq {
		return
	}
	player := d.players[d.currentTurn]
	// 出した後に手札が残っている場合のみ
	if player.GetCardsSize() == 0 {
		return
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 7 {
			// 渡す対象: 次のアクティブなプレイヤー (自分以外)
			target := d.getNextActivePlayer(d.currentTurn)
			if target >= 0 && target != d.currentTurn {
				d.pendingActionType = DaifugoPendingSevenPass
				d.pendingActionTarget = target
			}
			return
		}
	}
}

// triggerTenDiscardIfNeeded 10捨てチェック: 非ジョーカーの10が出されたらペンディングアクションをセット
func (d *Daifugo) triggerTenDiscardIfNeeded(cards []*Card, isSeq bool) {
	if !d.config.TenDiscardEnabled || isSeq {
		return
	}
	player := d.players[d.currentTurn]
	// 出した後に手札が残っている場合のみ
	if player.GetCardsSize() == 0 {
		return
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 10 {
			d.pendingActionType = DaifugoPendingTenDiscard
			d.pendingActionTarget = -1
			return
		}
	}
}

// isSpadeThreeCounter スペ3返し判定: 場がジョーカー1枚でスペードの3を1枚出す場合
func (d *Daifugo) isSpadeThreeCounter(cards []*Card) bool {
	if !d.config.SpadeThreeEnabled {
		return false
	}
	// 場がジョーカー1枚のみ
	if len(d.tableCards) != 1 || !IsJoker(d.tableCards[0]) {
		return false
	}
	// 出すカードがスペードの3を1枚のみ
	if len(cards) != 1 {
		return false
	}
	c := cards[0]
	return !IsJoker(c) && c.GetDesign() == CardDesignSpade && c.GetValue() == 3
}

// resolvePendingAction ペンディングアクションを解決する
func (d *Daifugo) resolvePendingAction(indices []int) error {
	if len(indices) != 1 {
		return NewDomainError(ErrInvalidPlay, "pending action requires exactly 1 card index")
	}
	player := d.players[d.currentTurn]
	card := player.GetCard(indices[0])
	if card == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", indices[0]))
	}

	switch d.pendingActionType {
	case DaifugoPendingSevenPass:
		// 7渡し: カードを対象プレイヤーに渡す
		removed := player.RemoveCards([]int{indices[0]})
		target := d.players[d.pendingActionTarget]
		target.AddCard(removed[0])
		target.SortCardsByStrength(d.cardStrengthForCard)
		d.humanAction = &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: removed}
	case DaifugoPendingTenDiscard:
		// 10捨て: カードを捨てる
		removed := player.RemoveCards([]int{indices[0]})
		d.humanAction = &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: removed}
	}

	d.pendingActionType = DaifugoPendingNone
	d.pendingActionTarget = -1

	d.advanceTurn()
	d.checkPassClear()
	return nil
}

// findStrongestNonJokerIndex プレイヤーの手札中の最強の非ジョーカーカードのインデックスを返す (末尾から探索)
func (d *Daifugo) findStrongestNonJokerIndex(player *DaifugoPlayer) int {
	for i := player.GetCardsSize() - 1; i >= 0; i-- {
		if !IsJoker(player.GetCard(i)) {
			return i
		}
	}
	return 0
}

// findWeakestNonJokerIndex プレイヤーの手札中の最弱の非ジョーカーカードのインデックスを返す (先頭から探索)
func (d *Daifugo) findWeakestNonJokerIndex(player *DaifugoPlayer) int {
	for i := 0; i < player.GetCardsSize(); i++ {
		if !IsJoker(player.GetCard(i)) {
			return i
		}
	}
	return 0
}

// cpuResolvePendingAction CPUがペンディングアクションを自動解決する
func (d *Daifugo) cpuResolvePendingAction() {
	player := d.players[d.currentTurn]

	var idx int
	switch d.pendingActionType {
	case DaifugoPendingSevenPass:
		// 7渡し: 最強の非ジョーカーカードを渡す
		idx = d.findStrongestNonJokerIndex(player)
		removed := player.RemoveCards([]int{idx})
		target := d.players[d.pendingActionTarget]
		target.AddCard(removed[0])
		target.SortCardsByStrength(d.cardStrengthForCard)
		action := &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: removed}
		d.cpuActions = append(d.cpuActions, action)
	case DaifugoPendingTenDiscard:
		// 10捨て: 最弱の非ジョーカーカードを捨てる
		idx = d.findWeakestNonJokerIndex(player)
		removed := player.RemoveCards([]int{idx})
		action := &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: removed}
		d.cpuActions = append(d.cpuActions, action)
	}

	d.pendingActionType = DaifugoPendingNone
	d.pendingActionTarget = -1

	d.advanceTurn()
	d.checkPassClear()
}

// applyCapitalFall 都落ちを適用する
// 前回大富豪だったプレイヤーが今回1位でない場合、そのプレイヤーと最下位のプレイヤーのランクを入れ替える
func (d *Daifugo) applyCapitalFall() {
	if !d.config.CapitalFallEnabled {
		return
	}
	// 前回大富豪のプレイヤーを探す
	prevDaifugoIdx := -1
	for i, p := range d.players {
		if p.GetPrevRank() == DaifugoRankDaifugo {
			prevDaifugoIdx = i
			break
		}
	}
	// 前回大富豪がいない (初回ゲーム) か今回も1位の場合はスキップ
	if prevDaifugoIdx < 0 || d.players[prevDaifugoIdx].GetRank() == DaifugoRankDaifugo {
		return
	}
	// 最下位のプレイヤーを探す
	lowestRank := 0
	lowestIdx := -1
	for i, p := range d.players {
		if p.GetRank() > lowestRank {
			lowestRank = p.GetRank()
			lowestIdx = i
		}
	}
	if lowestIdx < 0 || lowestIdx == prevDaifugoIdx {
		return
	}
	// ランクを入れ替える
	prevRank := d.players[prevDaifugoIdx].GetRank()
	d.players[prevDaifugoIdx].SetRank(d.players[lowestIdx].GetRank())
	d.players[lowestIdx].SetRank(prevRank)
}

// isIllegalFinish 反則上がり判定: 8切り/ジョーカー/革命で上がりとなる手かどうか
func (d *Daifugo) isIllegalFinish(cards []*Card, isSeq bool) bool {
	if !d.config.IllegalFinishEnabled {
		return false
	}
	// 8切り上がり: 非ジョーカーの8が含まれている && 8切りが有効
	if d.config.EightCutEnabled {
		for _, c := range cards {
			if !IsJoker(c) && c.GetValue() == 8 {
				return true
			}
		}
	}
	// ジョーカー上がり
	for _, c := range cards {
		if IsJoker(c) {
			return true
		}
	}
	// 革命上がり: 4枚以上 (グループまたは階段)
	if len(cards) >= 4 {
		// 階段の場合は階段革命が有効な時のみ
		if isSeq && !d.config.SequenceRevolutionEnabled {
			return false
		}
		return true
	}
	return false
}

// applyIllegalFinishPenalty 反則上がりペナルティを適用する
// ペナルティを受けたプレイヤーを最下位に降格し、他のプレイヤーのランクを調整する
func (d *Daifugo) applyIllegalFinishPenalty() {
	if !d.config.IllegalFinishEnabled {
		return
	}
	penalized := make([]*DaifugoPlayer, 0)
	nonPenalized := make([]*DaifugoPlayer, 0)
	for _, p := range d.players {
		if p.GetIllegalFinishPenalty() {
			penalized = append(penalized, p)
		} else {
			nonPenalized = append(nonPenalized, p)
		}
	}
	if len(penalized) == 0 {
		return
	}
	sort.Slice(nonPenalized, func(i, j int) bool {
		return nonPenalized[i].GetRank() < nonPenalized[j].GetRank()
	})
	sort.Slice(penalized, func(i, j int) bool {
		return penalized[i].GetRank() < penalized[j].GetRank()
	})
	rank := 1
	for _, p := range nonPenalized {
		p.SetRank(rank)
		rank++
	}
	for _, p := range penalized {
		p.SetRank(rank)
		rank++
	}
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

// GetPendingActionType ペンディングアクションの種類取得
func (d *Daifugo) GetPendingActionType() DaifugoPendingAction { return d.pendingActionType }

// GetPendingActionTarget ペンディングアクションの対象プレイヤーインデックス取得
func (d *Daifugo) GetPendingActionTarget() int { return d.pendingActionTarget }

// HasPendingAction ペンディングアクションがあるか取得
func (d *Daifugo) HasPendingAction() bool { return d.pendingActionType != DaifugoPendingNone }

// SetConfig ローカルルール設定を変更（ResetWithConfig用）
func (d *Daifugo) SetConfig(config DaifugoConfig) { d.config = config }

// GetReverseDirection 9リバースの方向取得
func (d *Daifugo) GetReverseDirection() bool { return d.reverseDirection }

// GetNumberLocked 連番縛り発動中か取得
func (d *Daifugo) GetNumberLocked() bool { return d.numberLocked }

// GetSortMode 手札ソートモード取得
func (d *Daifugo) GetSortMode() DaifugoSortMode { return d.sortMode }

// triggerNineReverseIfNeeded 9リバースチェック: 非ジョーカーの9が出されたらターン方向を反転（階段時は無効）
func (d *Daifugo) triggerNineReverseIfNeeded(cards []*Card, isSeq bool) {
	if !d.config.NineReverseEnabled || isSeq {
		return
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 9 {
			d.reverseDirection = !d.reverseDirection
			return
		}
	}
}

// triggerCoupDetatIfNeeded クーデターチェック: 3枚の非ジョーカー9で革命を起こす
func (d *Daifugo) triggerCoupDetatIfNeeded(cards []*Card) {
	if !d.config.CoupDetatEnabled {
		return
	}
	if len(cards) != 3 {
		return
	}
	for _, c := range cards {
		if IsJoker(c) || c.GetValue() != 9 {
			return
		}
	}
	d.revolutionActive = !d.revolutionActive
	d.sortAllActiveHands()
}

// SortHumanHand 人間プレイヤーの手札を指定モードでソートする
func (d *Daifugo) SortHumanHand(mode DaifugoSortMode) error {
	if d.gameEndFlag {
		return ErrGameEnded
	}
	d.sortMode = mode
	for _, p := range d.players {
		if p.GetIsHuman() && !p.GetIsFinished() {
			d.sortPlayerCards(p)
			break
		}
	}
	return nil
}

// sortAllActiveHands 全アクティブプレイヤーの手札をソートする
// 人間プレイヤーは sortMode に従い、CPUは常に強さ順
func (d *Daifugo) sortAllActiveHands() {
	for _, p := range d.players {
		if p.GetIsFinished() {
			continue
		}
		if p.GetIsHuman() {
			d.sortPlayerCards(p)
		} else {
			p.SortCardsByStrength(d.cardStrengthForCard)
		}
	}
}

// sortPlayerCards プレイヤーの手札を sortMode に従ってソートする
func (d *Daifugo) sortPlayerCards(p *DaifugoPlayer) {
	switch d.sortMode {
	case DaifugoSortBySuit:
		d.sortBySuit(p)
	case DaifugoSortByNumber:
		d.sortByNumber(p)
	default:
		p.SortCardsByStrength(d.cardStrengthForCard)
	}
}

// sortBySuit スート順でソート (Spade < Clover < Heart < Diamond, 同スート内は値の昇順、ジョーカーは末尾)
func (d *Daifugo) sortBySuit(p *DaifugoPlayer) {
	p.SortCardsByStrength(func(c *Card) int {
		if IsJoker(c) {
			return jokerSortWeight // ジョーカーは末尾
		}
		return c.GetDesign()*100 + c.GetValue()
	})
}

// sortByNumber 数字順でソート (値の昇順、同値ならスートの昇順、ジョーカーは末尾)
func (d *Daifugo) sortByNumber(p *DaifugoPlayer) {
	p.SortCardsByStrength(func(c *Card) int {
		if IsJoker(c) {
			return jokerSortWeight // ジョーカーは末尾
		}
		return c.GetValue()*100 + c.GetDesign()
	})
}
