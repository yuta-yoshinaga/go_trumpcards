package entities

import "math/rand"

// OldMaidPlayerCnt ババ抜きプレイヤー数
const OldMaidPlayerCnt = 4

// OldMaidCpuAction CPUの1ターン分の行動記録
type OldMaidCpuAction struct {
	DrawPlayerIdx  int     // 引いたプレイヤーインデックス
	DrawFromIdx    int     // 引かれた相手のインデックス
	DrawnCard      *Card   // 引いたカード
	DiscardedPairs int     // 捨てたペア数
	DiscardedCards []*Card // 捨てたカード
}

// OldMaid ババ抜きゲームクラス
type OldMaid struct {
	trumpCards         *TrumpCards
	players            []*OldMaidPlayer
	currentTurn        int                 // 現在の手番プレイヤーインデックス
	gameEndFlag        bool                // ゲーム終了フラグ
	loserIdx           int                 // 負けたプレイヤーインデックス
	lastDrawPlayerIdx  int                 // 最後に引いたプレイヤーのインデックス (-1=まだなし)
	lastDrawFromIdx    int                 // 最後に引いた相手のインデックス (-1=まだなし)
	lastDrawCard       *Card               // 最後に引いたカード
	lastDiscardedPairs int                 // 最後に捨てたペア数
	lastDiscardedCards []*Card             // 最後に捨てたカード
	hasDrawn           bool                // 引きが発生したか
	cpuActions         []*OldMaidCpuAction // CPUターンの行動履歴 (人間のターン後にリセット)
	humanAction        *OldMaidCpuAction   // 人間プレイヤーの最後の行動記録
}

// NewOldMaid コンストラクタ
func NewOldMaid(trumpCards *TrumpCards, players []*OldMaidPlayer) *OldMaid {
	return &OldMaid{
		trumpCards:         trumpCards,
		players:            players,
		currentTurn:        0,
		gameEndFlag:        false,
		loserIdx:           -1,
		lastDrawPlayerIdx:  -1,
		lastDrawFromIdx:    -1,
		lastDrawCard:       nil,
		lastDiscardedPairs: 0,
		lastDiscardedCards: nil,
		hasDrawn:           false,
		cpuActions:         nil,
		humanAction:        nil,
	}
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

	// シャッフル
	for i := 0; i < 10; i++ {
		o.trumpCards.Shuffle()
	}

	// 全プレイヤーのカードリセット
	for _, p := range o.players {
		p.Reset()
		p.SetIsFinished(false)
	}

	// 全カードを配る
	idx := 0
	for {
		card := o.trumpCards.DrawCard()
		if card == nil {
			break
		}
		o.players[idx%OldMaidPlayerCnt].AddCard(card)
		idx++
	}

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
	if !o.gameEndFlag {
		start := o.currentTurn
		for o.players[o.currentTurn].GetIsFinished() {
			o.currentTurn = (o.currentTurn + 1) % OldMaidPlayerCnt
			if o.currentTurn == start {
				break
			}
		}
	}
}

// getNextActivePlayer fromの次のアクティブなプレイヤーインデックスを取得
func (o *OldMaid) getNextActivePlayer(from int) int {
	for i := 1; i <= OldMaidPlayerCnt; i++ {
		next := (from + i) % OldMaidPlayerCnt
		if !o.players[next].GetIsFinished() {
			return next
		}
	}
	return -1
}

// getActivePlayerCnt アクティブなプレイヤー数取得
func (o *OldMaid) getActivePlayerCnt() int {
	cnt := 0
	for _, p := range o.players {
		if !p.GetIsFinished() {
			cnt++
		}
	}
	return cnt
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

	// 最後の引き情報を更新
	o.lastDrawPlayerIdx = playerIdx
	o.lastDrawFromIdx = targetIdx
	o.lastDrawCard = card
	o.hasDrawn = true

	// ペアを捨てる
	discardedCards, discardedCount := player.DiscardPairs()
	o.lastDiscardedPairs = discardedCount
	o.lastDiscardedCards = discardedCards

	// 手が空になったプレイヤーを上がりにする
	if target.GetCardsSize() == 0 {
		target.SetIsFinished(true)
	}
	if player.GetCardsSize() == 0 {
		player.SetIsFinished(true)
	}

	// ゲーム終了チェック
	o.checkGameEnd()

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

// PlayerDraw 人間プレイヤーがカードを引く
// cardIdx: 引くカードのインデックス。-1 の場合はランダム選択。
func (o *OldMaid) PlayerDraw(cardIdx int) {
	if o.gameEndFlag || !o.players[o.currentTurn].GetIsHuman() {
		return
	}
	// 人間のターン開始時にCPU行動履歴をリセット
	o.cpuActions = nil
	o.drawCard(o.currentTurn, cardIdx)
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
}

// CpuDraw 現在の手番がCPUの場合に1ターン実行
func (o *OldMaid) CpuDraw() {
	if o.gameEndFlag || o.players[o.currentTurn].GetIsHuman() {
		return
	}
	playerIdx := o.currentTurn
	card := o.drawCard(playerIdx, -1)
	action := &OldMaidCpuAction{
		DrawPlayerIdx:  playerIdx,
		DrawFromIdx:    o.lastDrawFromIdx,
		DrawnCard:      card,
		DiscardedPairs: o.lastDiscardedPairs,
		DiscardedCards: o.lastDiscardedCards,
	}
	o.cpuActions = append(o.cpuActions, action)
	if !o.gameEndFlag {
		o.advanceTurn()
	}
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
