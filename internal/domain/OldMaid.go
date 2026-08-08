package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// OldMaidPlayerCnt ババ抜きプレイヤー数
const OldMaidPlayerCnt = 4

// oldMaidShuffleCount シャッフル回数
const oldMaidShuffleCount = 10

// newShuffledDeck jokerCount枚のジョーカーを含むデッキを生成してシャッフルする
func newShuffledDeck(jokerCount int) *TrumpCards {
	tc := NewTrumpCards(jokerCount)
	for i := 0; i < oldMaidShuffleCount; i++ {
		tc.Shuffle()
	}
	return tc
}

// OldMaid固有の迷い時間ディレイ定数 (共通定数は hesitation.go)
const (
	oldMaidHesitationJokerMin = 1000
	oldMaidHesitationJokerMax = 1500
)

// OldMaidCpuAction CPUの1ターン分の行動記録
type OldMaidCpuAction struct {
	DrawPlayerIdx  int     // 引いたプレイヤーインデックス
	DrawFromIdx    int     // 引かれた相手のインデックス
	DrawnCard      *Card   // 引いたカード
	DiscardedPairs int     // 捨てたペア数
	DiscardedCards []*Card // 捨てたカード
	HesitationMs   int     // 迷い時間ディレイ (ミリ秒; 0=無効)
}

// OldMaidDrawHistoryEntry ゲーム全体の引き履歴の1エントリ
type OldMaidDrawHistoryEntry struct {
	DrawPlayerIdx  int  // 引いたプレイヤーインデックス
	DrawFromIdx    int  // 引かれた相手のインデックス
	DiscardedPairs int  // 捨てたペア数
	DrawerFinished bool // 引いた側が上がったか
	TargetFinished bool // 引かれた側が上がったか
}

// OldMaid ババ抜きゲームクラス
type OldMaid struct {
	trumpCards            *TrumpCards
	players               []*OldMaidPlayer
	currentTurn           int                        // 現在の手番プレイヤーインデックス
	gameEndFlag           bool                       // ゲーム終了フラグ
	loserIdx              int                        // 負けたプレイヤーインデックス
	lastDrawPlayerIdx     int                        // 最後に引いたプレイヤーのインデックス (-1=まだなし)
	lastDrawFromIdx       int                        // 最後に引いた相手のインデックス (-1=まだなし)
	lastDrawCard          *Card                      // 最後に引いたカード
	lastDiscardedPairs    int                        // 最後に捨てたペア数
	lastDiscardedCards    []*Card                    // 最後に捨てたカード
	hasDrawn              bool                       // 引きが発生したか
	cpuActions            []*OldMaidCpuAction        // CPUターンの行動履歴 (人間のターン後にリセット)
	humanAction           *OldMaidCpuAction          // 人間プレイヤーの最後の行動記録
	drawHistory           []*OldMaidDrawHistoryEntry // ゲーム全体の引き履歴
	config                OldMaidConfig              // ゲーム設定
	removedCard           *Card                      // ジジ抜き: 除外されたカード
	cpuHighlightedCardIdx int                        // CPU心理戦: 奇数カードの位置 (-1=なし)
	humanHandDirty        bool                       // 人間がシャッフル/並び替えしたフラグ
	humanProfile          *OldMaidHumanProfile       // メタAIプロファイル
	actionLogBase
}

// NewOldMaid コンストラクタ
func NewOldMaid(trumpCards *TrumpCards, players []*OldMaidPlayer) *OldMaid {
	return &OldMaid{
		trumpCards:            trumpCards,
		players:               players,
		currentTurn:           0,
		gameEndFlag:           false,
		loserIdx:              -1,
		lastDrawPlayerIdx:     -1,
		lastDrawFromIdx:       -1,
		lastDrawCard:          nil,
		lastDiscardedPairs:    0,
		lastDiscardedCards:    nil,
		hasDrawn:              false,
		cpuActions:            nil,
		humanAction:           nil,
		config:                DefaultOldMaidConfig(),
		removedCard:           nil,
		cpuHighlightedCardIdx: -1,
	}
}

// NewDefaultOldMaid returns OldMaid with the standard 4-player setup (1 human, 3 CPU)
// using a deck with 1 joker. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultOldMaid() *OldMaid {
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(true),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	return NewOldMaid(NewTrumpCards(1), players)
}

// SetConfig ゲーム設定をセット
func (o *OldMaid) SetConfig(config OldMaidConfig) { o.config = config }

// GetConfig ゲーム設定取得
func (o *OldMaid) GetConfig() OldMaidConfig { return o.config }

// GetRemovedCard ジジ抜きで除外されたカード取得
func (o *OldMaid) GetRemovedCard() *Card { return o.removedCard }

// GetCpuHighlightedCardIdx CPU心理戦で強調された奇数カードの位置取得 (-1=なし)
func (o *OldMaid) GetCpuHighlightedCardIdx() int { return o.cpuHighlightedCardIdx }

// GetHumanHandDirty 人間がシャッフル/並び替えしたか
func (o *OldMaid) GetHumanHandDirty() bool { return o.humanHandDirty }

// GetHumanProfile メタAIプロファイル取得
func (o *OldMaid) GetHumanProfile() *OldMaidHumanProfile { return o.humanProfile }

// ResetProfile メタAIプロファイルをリセットする
func (o *OldMaid) ResetProfile() { o.humanProfile = nil }

// ExportProfile メタAIプロファイルをエクスポートする (プロファイルがない場合はnil)
func (o *OldMaid) ExportProfile() interface{} {
	if o.humanProfile == nil {
		return nil
	}
	d := o.humanProfile.Export()
	return &d
}

// ImportProfile JSONバイトからメタAIプロファイルをインポートする
func (o *OldMaid) ImportProfile(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	d, err := ImportOldMaidHumanProfileJSON(data)
	if err != nil {
		return err
	}
	o.humanProfile = &OldMaidHumanProfile{}
	o.humanProfile.Import(d)
	return nil
}

// Reset ゲーム初期化
func (o *OldMaid) Reset() {
	o.gameEndFlag = false
	o.loserIdx = -1
	o.currentTurn = 0
	o.lastDrawPlayerIdx = -1
	o.lastDrawFromIdx = -1
	o.lastDrawCard = nil
	o.lastDiscardedPairs = 0
	o.lastDiscardedCards = nil
	o.hasDrawn = false
	o.cpuActions = nil
	o.humanAction = nil
	o.drawHistory = nil
	o.removedCard = nil
	o.cpuHighlightedCardIdx = -1
	o.humanHandDirty = false
	o.actionLog = nil

	// メタAIプロファイルの管理
	if o.config.CpuMetaAI {
		if o.humanProfile != nil {
			o.humanProfile.GamesPlayed++
		} else {
			o.humanProfile = &OldMaidHumanProfile{}
		}
	}

	// 全プレイヤーのカードリセット (記憶もクリア)
	resetPlayers(o.players, func(p *OldMaidPlayer) {
		p.ResetDrawMemory()
	})

	// プレイ順をランダムにする
	rand.Shuffle(len(o.players), func(i, j int) {
		o.players[i], o.players[j] = o.players[j], o.players[i]
	})

	// モードに応じてデッキを再構築し、カードを配る
	if o.config.Mode == OldMaidModeJijiNuki {
		// ジジ抜き: シャッフル済みデッキの先頭1枚を除外カードとし、残り51枚を配る
		o.trumpCards = newShuffledDeck(0)
		o.removedCard = o.trumpCards.DrawCard()
	} else {
		// ノーマル: ジョーカー1枚付き53枚
		o.trumpCards = newShuffledDeck(1)
	}
	dealAllCards(o.trumpCards, o.players)

	// 全プレイヤーのペアを捨てる
	for _, p := range o.players {
		_, _ = p.DiscardPairs()
		if p.GetCardsSize() == 0 {
			p.SetIsFinished(true)
		}
	}

	// ゲーム終了チェック
	o.checkGameEnd()

	// currentTurnがフィニッシュしていたら次へ
	o.advancePastFinished()
}

// getNextActivePlayer fromの次のアクティブなプレイヤーインデックスを取得
func (o *OldMaid) getNextActivePlayer(from int) int {
	return nextActivePlayer(o.players, from, 1)
}

// getActivePlayerCnt アクティブなプレイヤー数取得
func (o *OldMaid) getActivePlayerCnt() int {
	return countPlayers(o.players, func(p *OldMaidPlayer) bool { return !p.GetIsFinished() })
}

// checkGameEnd ゲーム終了チェック (残り1人なら負け確定)
func (o *OldMaid) checkGameEnd() bool {
	active := o.getActivePlayerCnt()
	if active <= 1 {
		for i, p := range o.players {
			if !p.GetIsFinished() {
				o.loserIdx = i
				break
			}
		}
		o.gameEndFlag = true
		return true
	}
	return false
}

// drawCard playerIdxがカードを引く (内部処理)
// cardIdx: 引くカードのインデックス。-1 の場合はランダム選択。
func (o *OldMaid) drawCard(playerIdx int, cardIdx int) *Card {
	if o.gameEndFlag {
		return nil
	}
	player := o.players[playerIdx]
	if player.GetIsFinished() {
		return nil
	}

	targetIdx := o.getNextActivePlayer(playerIdx)
	if targetIdx < 0 {
		return nil
	}
	target := o.players[targetIdx]
	if target.GetCardsSize() == 0 {
		return nil
	}

	// カードインデックスの決定
	idx := cardIdx
	if idx < 0 || idx >= target.GetCardsSize() {
		idx = rand.Intn(target.GetCardsSize())
	}
	card := target.RemoveCard(idx)
	player.AddCard(card)
	// 引いたカードの位置がわからないよう手札をシャッフル
	player.ShuffleCards()

	// 最後の引き情報を更新
	o.lastDrawPlayerIdx = playerIdx
	o.lastDrawFromIdx = targetIdx
	o.lastDrawCard = card
	o.hasDrawn = true

	// 棋譜: ドロー
	o.appendLog(playerIdx, "draw", fmt.Sprintf("drew from player %d", targetIdx), []*Card{card})

	// ペアを捨てる
	discardedCards, discardedCount := player.DiscardPairs()
	o.lastDiscardedPairs = discardedCount
	o.lastDiscardedCards = discardedCards

	// 棋譜: ペア捨て
	if discardedCount > 0 {
		o.appendLog(playerIdx, "discard", fmt.Sprintf("discarded %d pair(s)", discardedCount), discardedCards)
	}

	// 手が空になったプレイヤーを上がりにする
	if target.GetCardsSize() == 0 {
		target.SetIsFinished(true)
		o.appendLog(targetIdx, "finish", fmt.Sprintf("player %d finished", targetIdx), nil)
	}
	if player.GetCardsSize() == 0 {
		player.SetIsFinished(true)
		o.appendLog(playerIdx, "finish", fmt.Sprintf("player %d finished", playerIdx), nil)
	}

	// ゲーム終了チェック
	o.checkGameEnd()

	// 引き履歴に追加 (カード情報なし — プライバシー保護)
	o.drawHistory = append(o.drawHistory, &OldMaidDrawHistoryEntry{
		DrawPlayerIdx:  playerIdx,
		DrawFromIdx:    targetIdx,
		DiscardedPairs: discardedCount,
		DrawerFinished: player.GetIsFinished(),
		TargetFinished: target.GetIsFinished(),
	})

	return card
}

// advanceTurn 手番を次のアクティブなプレイヤーへ進める
func (o *OldMaid) advanceTurn() {
	if o.gameEndFlag {
		return
	}
	next := o.getNextActivePlayer(o.currentTurn)
	if next >= 0 {
		o.currentTurn = next
	}
}

// advancePastFinished currentTurnがフィニッシュしていたら次のアクティブプレイヤーへ進める
func (o *OldMaid) advancePastFinished() {
	if o.gameEndFlag {
		return
	}
	for i := 0; i < len(o.players); i++ {
		if !o.players[o.currentTurn].GetIsFinished() {
			break
		}
		o.currentTurn = (o.currentTurn + 1) % len(o.players)
	}
}

// PlayerDraw 人間プレイヤーがカードを引く
// cardIdx: 引くカードのインデックス。-1 の場合はランダム選択。
func (o *OldMaid) PlayerDraw(cardIdx int) error {
	if o.gameEndFlag {
		return ErrGameEnded
	}
	if !o.players[o.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	// メタAI: ピックと引きを記録
	if o.config.CpuMetaAI && o.humanProfile != nil {
		if cardIdx >= 0 {
			targetIdx := o.getNextActivePlayer(o.currentTurn)
			if targetIdx >= 0 {
				targetSize := o.players[targetIdx].GetCardsSize()
				if cardIdx < targetSize {
					o.humanProfile.RecordPick(cardIdx, targetSize)
				}
			}
		}
		o.humanProfile.RecordDraw()
	}
	// 人間のターン開始時にCPU行動履歴をリセット
	o.cpuActions = nil
	o.drawCard(o.currentTurn, cardIdx)
	// 人間が引いたらdirtyフラグをリセット (CPU記憶AI用)
	o.humanHandDirty = false
	// 人間の行動を記録
	o.humanAction = &OldMaidCpuAction{
		DrawPlayerIdx:  o.lastDrawPlayerIdx,
		DrawFromIdx:    o.lastDrawFromIdx,
		DrawnCard:      o.lastDrawCard,
		DiscardedPairs: o.lastDiscardedPairs,
		DiscardedCards: o.lastDiscardedCards,
	}
	if !o.gameEndFlag {
		o.advanceTurn()
	}
	return nil
}

// CPU戦略用定数
const (
	cpuEdgeSelectThreshold = 3  // 端のカードを選ぶ閾値 (30% = 3/10)
	cpuSelectTotalCases    = 10 // 乱数の全選択肢数
	cpuEdgeSides           = 2  // 先頭か末尾か
)

// CPU記憶AI用定数
const (
	cpuMemoryEdgeThreshold   = 5 // 人間がシャッフルした時の端選択閾値 (50% = 5/10)
	cpuMemoryAvoidThreshold  = 4 // 前回ペア不成立時の回避閾値 (40% = 4/10)
	cpuMemoryPreferThreshold = 4 // 前回ペア成立時の近傍選好閾値 (40% = 4/10)
)

// cpuSelectCardIdx 対象プレイヤーの手札枚数を受け取り、引くカードのインデックスを戦略的に選択する。
// 30%の確率で端のカード（先頭または末尾）を選択し、残りはランダム選択。
// 「誰から引くか」の解決は呼び出し元が担う。
func cpuSelectCardIdx(size int) int {
	if size <= 1 {
		return 0
	}
	// 30%の確率で端のカードを狙う
	if rand.Intn(cpuSelectTotalCases) < cpuEdgeSelectThreshold {
		if rand.Intn(cpuEdgeSides) == 0 {
			return 0 // 先頭のカード
		}
		return size - 1 // 末尾のカード
	}
	return rand.Intn(size)
}

// cpuSelectWithMemory 記憶AIを使ったカード選択
// playerIdx: CPU自身のインデックス, targetIdx: 引く相手のインデックス, size: 相手の手札枚数
func (o *OldMaid) cpuSelectWithMemory(playerIdx, targetIdx, size int) int {
	if size <= 1 {
		return 0
	}
	cpu := o.players[playerIdx]

	// 対象が人間 かつ humanHandDirty → 50%の確率で端を選択, 記憶を無効化
	if o.players[targetIdx].GetIsHuman() && o.humanHandDirty {
		cpu.ResetDrawMemory()
		if rand.Intn(cpuSelectTotalCases) < cpuMemoryEdgeThreshold {
			if rand.Intn(cpuEdgeSides) == 0 {
				return 0
			}
			return size - 1
		}
		return rand.Intn(size)
	}

	lastPos := cpu.GetMemLastDrawPos()
	// 有効な記憶がある場合 (前回引いた位置が現在の手札サイズ範囲内)
	if lastPos >= 0 && lastPos < size {
		if !cpu.GetMemGotPair() {
			// 前回ペア不成立 → 40%の確率でその位置を回避
			if rand.Intn(cpuSelectTotalCases) < cpuMemoryAvoidThreshold {
				// lastPosを除外してランダム選択
				idx := rand.Intn(size - 1)
				if idx >= lastPos {
					idx++
				}
				return idx
			}
			return rand.Intn(size)
		}
		// 前回ペア成立 → 40%の確率で近傍(±1)を選好
		if rand.Intn(cpuSelectTotalCases) < cpuMemoryPreferThreshold {
			candidates := make([]int, 0, 2)
			if lastPos-1 >= 0 {
				candidates = append(candidates, lastPos-1)
			}
			if lastPos+1 < size {
				candidates = append(candidates, lastPos+1)
			}
			return candidates[rand.Intn(len(candidates))]
		}
		return rand.Intn(size)
	}

	// 記憶なし → デフォルト戦略にフォールバック
	return cpuSelectCardIdx(size)
}

// CpuDraw 現在の手番がCPUの場合に1ターン実行
func (o *OldMaid) CpuDraw() error {
	if o.gameEndFlag {
		return ErrGameEnded
	}
	if o.players[o.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	playerIdx := o.currentTurn
	targetIdx := o.getNextActivePlayer(playerIdx)
	if targetIdx < 0 || targetIdx >= len(o.players) {
		return nil
	}
	var cardIdx int
	if o.config.CpuMemoryAI {
		cardIdx = o.cpuSelectWithMemory(playerIdx, targetIdx, o.players[targetIdx].GetCardsSize())
	} else {
		cardIdx = cpuSelectCardIdx(o.players[targetIdx].GetCardsSize())
	}
	card := o.drawCard(playerIdx, cardIdx)
	// CPU記憶AI: 引いた位置とペア結果を記録
	if o.config.CpuMemoryAI {
		o.players[playerIdx].SetMemLastDrawPos(cardIdx)
		o.players[playerIdx].SetMemGotPair(o.lastDiscardedPairs > 0)
	}
	action := &OldMaidCpuAction{
		DrawPlayerIdx:  playerIdx,
		DrawFromIdx:    o.lastDrawFromIdx,
		DrawnCard:      card,
		DiscardedPairs: o.lastDiscardedPairs,
		DiscardedCards: o.lastDiscardedCards,
	}
	if o.config.CpuHesitationEnabled {
		drewJoker := card != nil && card.GetDesign() == CardDesignJoker
		gotPair := o.lastDiscardedPairs > 0
		action.HesitationMs = calcOldMaidHesitationMs(gotPair, drewJoker)
	}
	o.cpuActions = append(o.cpuActions, action)
	if !o.gameEndFlag {
		o.advanceTurn()
	}
	return nil
}

// calcOldMaidHesitationMs 引いたカードの結果に応じた迷い時間(ミリ秒)を算出する
// Note: ジジ抜きモードではジョーカーがデッキに含まれないため drewJoker は常に false となり、
// pair/normal の2分岐のみが使われる。
func calcOldMaidHesitationMs(gotPair bool, drewJoker bool) int {
	if drewJoker {
		return oldMaidHesitationJokerMin + rand.Intn(oldMaidHesitationJokerMax-oldMaidHesitationJokerMin+1)
	}
	if gotPair {
		return hesitationFastMin + rand.Intn(hesitationFastMax-hesitationFastMin+1)
	}
	return hesitationMediumMin + rand.Intn(hesitationMediumMax-hesitationMediumMin+1)
}

// detectOddCardIdx プレイヤーの手札から奇数カードのインデックスを検出する (内部処理)
// Normal: ジョーカーのインデックスを返す (-1=なし)
// JijiNuki: 手札中で出現回数が奇数の値を持つ最初のカードのインデックスを返す (-1=なし)
func (o *OldMaid) detectOddCardIdx(player *OldMaidPlayer) int {
	switch o.config.Mode {
	case OldMaidModeNormal:
		for i := 0; i < player.GetCardsSize(); i++ {
			c := player.GetCard(i)
			if c != nil && c.GetDesign() == CardDesignJoker {
				return i
			}
		}
		return -1
	case OldMaidModeJijiNuki:
		// 手札内で各値の出現回数をカウント
		counts := make(map[int]int)
		for i := 0; i < player.GetCardsSize(); i++ {
			c := player.GetCard(i)
			if c != nil {
				counts[c.GetValue()]++
			}
		}
		// 出現回数が奇数の最初のカードのインデックスを返す
		for i := 0; i < player.GetCardsSize(); i++ {
			c := player.GetCard(i)
			if c != nil && counts[c.GetValue()]%2 != 0 {
				return i
			}
		}
		return -1
	default:
		return -1
	}
}

// ArrangeTargetForHumanDraw CPU心理戦: 人間が引く前に対象CPUの奇数カードを端に配置する
func (o *OldMaid) ArrangeTargetForHumanDraw() {
	o.cpuHighlightedCardIdx = -1
	if o.gameEndFlag || !o.IsHumanTurn() || !o.config.CpuPlacementStrategy {
		return
	}
	targetIdx := o.getNextActivePlayer(o.currentTurn)
	if targetIdx < 0 {
		return
	}
	target := o.players[targetIdx]
	if target.GetIsHuman() || target.GetCardsSize() <= 1 {
		return
	}
	oddIdx := o.detectOddCardIdx(target)
	if oddIdx < 0 {
		return
	}
	size := target.GetCardsSize()
	var position int
	// メタAI: 人間が最もピックしにくい位置に配置
	if o.config.CpuMetaAI && o.humanProfile != nil && o.humanProfile.AdaptStrength() >= metaAIMinAdaptForPlacement {
		position = o.humanProfile.StrategicPlacement(size)
	} else {
		if rand.Intn(2) == 0 {
			position = 0
		} else {
			position = size - 1
		}
	}
	card := target.RemoveCard(oddIdx)
	target.InsertCard(card, position)
	o.cpuHighlightedCardIdx = position
}

// findHumanPlayer 人間プレイヤーを検索する
func (o *OldMaid) findHumanPlayer() *OldMaidPlayer {
	for _, p := range o.players {
		if p.GetIsHuman() {
			return p
		}
	}
	return nil
}

// ShuffleHumanHand 人間プレイヤーの手札をシャッフルする
func (o *OldMaid) ShuffleHumanHand() error {
	if o.gameEndFlag {
		return ErrGameEnded
	}
	human := o.findHumanPlayer()
	if human == nil {
		return ErrNotHumanTurn
	}
	human.ShuffleCards()
	o.humanHandDirty = true
	// メタAI: シャッフルを記録
	if o.config.CpuMetaAI && o.humanProfile != nil {
		o.humanProfile.RecordShuffle()
	}
	return nil
}

// ReorderHumanHand 人間プレイヤーの手札を指定された順番に並び替える
func (o *OldMaid) ReorderHumanHand(indices []int) error {
	if o.gameEndFlag {
		return ErrGameEnded
	}
	human := o.findHumanPlayer()
	if human == nil {
		return ErrNotHumanTurn
	}
	err := human.ReorderCards(indices)
	if err == nil {
		o.humanHandDirty = true
	}
	return err
}

// IsHumanTurn 現在の手番が人間かどうか
func (o *OldMaid) IsHumanTurn() bool {
	return o.players[o.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (o *OldMaid) GetCurrentTurn() int {
	return o.currentTurn
}

// GetNextDrawTargetIdx 現在の手番プレイヤーが引く相手のインデックス取得
func (o *OldMaid) GetNextDrawTargetIdx() int {
	return o.getNextActivePlayer(o.currentTurn)
}

// GetGameEndFlag ゲーム終了フラグ取得
func (o *OldMaid) GetGameEndFlag() bool {
	return o.gameEndFlag
}

// GetLoserIdx 負けプレイヤーインデックス取得
func (o *OldMaid) GetLoserIdx() int {
	return o.loserIdx
}

// GetPlayer プレイヤー取得
func (o *OldMaid) GetPlayer(idx int) *OldMaidPlayer {
	if idx < 0 || idx >= len(o.players) {
		return nil
	}
	return o.players[idx]
}

// GetPlayerCnt プレイヤー数取得
func (o *OldMaid) GetPlayerCnt() int {
	return len(o.players)
}

// GetLastDrawPlayerIdx 最後に引いたプレイヤーのインデックス
func (o *OldMaid) GetLastDrawPlayerIdx() int {
	return o.lastDrawPlayerIdx
}

// GetLastDrawFromIdx 最後に引いた相手のインデックス
func (o *OldMaid) GetLastDrawFromIdx() int {
	return o.lastDrawFromIdx
}

// GetLastDrawCard 最後に引いたカード
func (o *OldMaid) GetLastDrawCard() *Card {
	return o.lastDrawCard
}

// GetLastDiscardedPairs 最後に捨てたペア数
func (o *OldMaid) GetLastDiscardedPairs() int {
	return o.lastDiscardedPairs
}

// GetLastDiscardedCards 最後に捨てたカード取得
func (o *OldMaid) GetLastDiscardedCards() []*Card {
	return o.lastDiscardedCards
}

// GetHasDrawn 引きが発生したかどうか
func (o *OldMaid) GetHasDrawn() bool {
	return o.hasDrawn
}

// GetCpuActions CPUターンの行動履歴取得
func (o *OldMaid) GetCpuActions() []*OldMaidCpuAction {
	return o.cpuActions
}

// GetHumanAction 人間プレイヤーの最後の行動記録取得
func (o *OldMaid) GetHumanAction() *OldMaidCpuAction {
	return o.humanAction
}

// GetDrawHistory ゲーム全体の引き履歴取得
func (o *OldMaid) GetDrawHistory() []*OldMaidDrawHistoryEntry {
	return o.drawHistory
}

// --- JSON Serialization ---

// oldMaidCpuActionJSON is the JSON wire format for OldMaidCpuAction.
type oldMaidCpuActionJSON struct {
	DrawPlayerIdx  int     `json:"dp"`
	DrawFromIdx    int     `json:"df"`
	DrawnCard      *Card   `json:"dc"`
	DiscardedPairs int     `json:"di"`
	DiscardedCards []*Card `json:"ds"`
	HesitationMs   int     `json:"hm"`
}

// MarshalJSON implements json.Marshaler.
func (a *OldMaidCpuAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(oldMaidCpuActionJSON{
		DrawPlayerIdx:  a.DrawPlayerIdx,
		DrawFromIdx:    a.DrawFromIdx,
		DrawnCard:      a.DrawnCard,
		DiscardedPairs: a.DiscardedPairs,
		DiscardedCards: a.DiscardedCards,
		HesitationMs:   a.HesitationMs,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *OldMaidCpuAction) UnmarshalJSON(data []byte) error {
	var j oldMaidCpuActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.DrawPlayerIdx = j.DrawPlayerIdx
	a.DrawFromIdx = j.DrawFromIdx
	a.DrawnCard = j.DrawnCard
	a.DiscardedPairs = j.DiscardedPairs
	a.DiscardedCards = j.DiscardedCards
	a.HesitationMs = j.HesitationMs
	return nil
}

// oldMaidDrawHistoryEntryJSON is the JSON wire format for OldMaidDrawHistoryEntry.
type oldMaidDrawHistoryEntryJSON struct {
	DrawPlayerIdx  int  `json:"dp"`
	DrawFromIdx    int  `json:"df"`
	DiscardedPairs int  `json:"di"`
	DrawerFinished bool `json:"dr"`
	TargetFinished bool `json:"tf"`
}

// MarshalJSON implements json.Marshaler.
func (e *OldMaidDrawHistoryEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(oldMaidDrawHistoryEntryJSON{
		DrawPlayerIdx:  e.DrawPlayerIdx,
		DrawFromIdx:    e.DrawFromIdx,
		DiscardedPairs: e.DiscardedPairs,
		DrawerFinished: e.DrawerFinished,
		TargetFinished: e.TargetFinished,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *OldMaidDrawHistoryEntry) UnmarshalJSON(data []byte) error {
	var j oldMaidDrawHistoryEntryJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	e.DrawPlayerIdx = j.DrawPlayerIdx
	e.DrawFromIdx = j.DrawFromIdx
	e.DiscardedPairs = j.DiscardedPairs
	e.DrawerFinished = j.DrawerFinished
	e.TargetFinished = j.TargetFinished
	return nil
}

// oldMaidJSON is the JSON wire format for OldMaid.
type oldMaidJSON struct {
	TrumpCards            *TrumpCards                `json:"tc"`
	Players               []*OldMaidPlayer           `json:"pl"`
	CurrentTurn           int                        `json:"ct"`
	GameEndFlag           bool                       `json:"ge"`
	LoserIdx              int                        `json:"li"`
	LastDrawPlayerIdx     int                        `json:"dp"`
	LastDrawFromIdx       int                        `json:"df"`
	LastDrawCard          *Card                      `json:"lc"`
	LastDiscardedPairs    int                        `json:"ld"`
	LastDiscardedCards    []*Card                    `json:"ls"`
	HasDrawn              bool                       `json:"hd"`
	CpuActions            []*OldMaidCpuAction        `json:"ca"`
	HumanAction           *OldMaidCpuAction          `json:"ha"`
	DrawHistory           []*OldMaidDrawHistoryEntry `json:"dh"`
	Config                OldMaidConfig              `json:"cf"`
	RemovedCard           *Card                      `json:"rc"`
	CpuHighlightedCardIdx int                        `json:"ch"`
	HumanHandDirty        bool                       `json:"hh"`
	Profile               *OldMaidHumanProfileData   `json:"pf,omitempty"`
	ActionLog             []*ActionLogEntry          `json:"al"`
}

// oldMaidMaxSliceLen caps slice sizes during deserialisation.
const oldMaidMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (o *OldMaid) MarshalJSON() ([]byte, error) {
	j := oldMaidJSON{
		TrumpCards:            o.trumpCards,
		Players:               o.players,
		CurrentTurn:           o.currentTurn,
		GameEndFlag:           o.gameEndFlag,
		LoserIdx:              o.loserIdx,
		LastDrawPlayerIdx:     o.lastDrawPlayerIdx,
		LastDrawFromIdx:       o.lastDrawFromIdx,
		LastDrawCard:          o.lastDrawCard,
		LastDiscardedPairs:    o.lastDiscardedPairs,
		LastDiscardedCards:    o.lastDiscardedCards,
		HasDrawn:              o.hasDrawn,
		CpuActions:            o.cpuActions,
		HumanAction:           o.humanAction,
		DrawHistory:           o.drawHistory,
		Config:                o.config,
		RemovedCard:           o.removedCard,
		CpuHighlightedCardIdx: o.cpuHighlightedCardIdx,
		HumanHandDirty:        o.humanHandDirty,
		ActionLog:             o.actionLog,
	}
	if o.humanProfile != nil {
		d := o.humanProfile.Export()
		j.Profile = &d
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *OldMaid) UnmarshalJSON(data []byte) error {
	var j oldMaidJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > oldMaidMaxSliceLen || len(j.LastDiscardedCards) > oldMaidMaxSliceLen ||
		len(j.CpuActions) > oldMaidMaxSliceLen || len(j.DrawHistory) > oldMaidMaxSliceLen ||
		len(j.ActionLog) > oldMaidMaxSliceLen {
		return fmt.Errorf("oldMaid: input array exceeds maximum allowed size")
	}
	o.trumpCards = j.TrumpCards
	if o.trumpCards == nil {
		o.trumpCards = NewTrumpCards(0)
	}
	o.players = j.Players
	if o.players == nil {
		o.players = make([]*OldMaidPlayer, 0)
	}
	o.currentTurn = j.CurrentTurn
	o.gameEndFlag = j.GameEndFlag
	o.loserIdx = j.LoserIdx
	o.lastDrawPlayerIdx = j.LastDrawPlayerIdx
	o.lastDrawFromIdx = j.LastDrawFromIdx
	o.lastDrawCard = j.LastDrawCard
	o.lastDiscardedPairs = j.LastDiscardedPairs
	o.lastDiscardedCards = j.LastDiscardedCards
	if o.lastDiscardedCards == nil {
		o.lastDiscardedCards = make([]*Card, 0)
	}
	o.hasDrawn = j.HasDrawn
	o.cpuActions = j.CpuActions
	o.humanAction = j.HumanAction
	o.drawHistory = j.DrawHistory
	if o.drawHistory == nil {
		o.drawHistory = make([]*OldMaidDrawHistoryEntry, 0)
	}
	o.config = j.Config
	o.removedCard = j.RemovedCard
	o.cpuHighlightedCardIdx = j.CpuHighlightedCardIdx
	o.humanHandDirty = j.HumanHandDirty
	if j.Profile != nil {
		o.humanProfile = &OldMaidHumanProfile{}
		o.humanProfile.Import(*j.Profile)
	}
	o.actionLog = j.ActionLog
	if o.actionLog == nil {
		o.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
