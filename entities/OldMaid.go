package entities

import "math/rand"

// OldMaidPlayerCnt ババ抜きプレイヤー数
const OldMaidPlayerCnt = 4

// OldMaid ババ抜きゲームクラス
type OldMaid struct {
	trumpCards         *TrumpCards
	players            []*OldMaidPlayer
	currentTurn        int  // 現在の手番プレイヤーインデックス
	gameEndFlag        bool // ゲーム終了フラグ
	loserIdx           int  // 負けたプレイヤーインデックス
	lastDrawPlayerIdx  int  // 最後に引いたプレイヤーのインデックス (-1=まだなし)
	lastDrawFromIdx    int  // 最後に引いた相手のインデックス (-1=まだなし)
	lastDiscardedPairs int  // 最後に捨てたペア数
	hasDrawn           bool // 引きが発生したか
}

// NewOldMaid コンストラクタ
func NewOldMaid(trumpCards *TrumpCards, players []*OldMaidPlayer) *OldMaid {
	return &OldMaid{
		trumpCards:        trumpCards,
		players:           players,
		currentTurn:       0,
		gameEndFlag:       false,
		loserIdx:          -1,
		lastDrawPlayerIdx: -1,
		lastDrawFromIdx:   -1,
		lastDiscardedPairs: 0,
		hasDrawn:          false,
	}
}

// Reset ゲーム初期化
func (o *OldMaid) Reset() {
	o.gameEndFlag = false
	o.loserIdx = -1
	o.currentTurn = 0
	o.lastDrawPlayerIdx = -1
	o.lastDrawFromIdx = -1
	o.lastDiscardedPairs = 0
	o.hasDrawn = false

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
		p.DiscardPairs()
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
func (o *OldMaid) drawCard(playerIdx int) {
	if o.gameEndFlag {
		return
	}
	player := o.players[playerIdx]
	if player.GetIsFinished() {
		return
	}

	targetIdx := o.getNextActivePlayer(playerIdx)
	if targetIdx < 0 {
		return
	}
	target := o.players[targetIdx]
	if target.GetCardsSize() == 0 {
		return
	}

	// ランダムにカードを選ぶ
	randomIdx := rand.Intn(target.GetCardsSize())
	card := target.RemoveCard(randomIdx)
	player.AddCard(card)

	// 最後の引き情報を更新
	o.lastDrawPlayerIdx = playerIdx
	o.lastDrawFromIdx = targetIdx
	o.hasDrawn = true

	// ペアを捨てる
	discarded := player.DiscardPairs()
	o.lastDiscardedPairs = discarded

	// 手が空になったプレイヤーを上がりにする
	if target.GetCardsSize() == 0 {
		target.SetIsFinished(true)
	}
	if player.GetCardsSize() == 0 {
		player.SetIsFinished(true)
	}

	// ゲーム終了チェック
	o.checkGameEnd()
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
func (o *OldMaid) PlayerDraw() {
	if o.gameEndFlag || !o.players[o.currentTurn].GetIsHuman() {
		return
	}
	o.drawCard(o.currentTurn)
	if !o.gameEndFlag {
		o.advanceTurn()
	}
}

// CpuDraw 現在の手番がCPUの場合に1ターン実行
func (o *OldMaid) CpuDraw() {
	if o.gameEndFlag || o.players[o.currentTurn].GetIsHuman() {
		return
	}
	o.drawCard(o.currentTurn)
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

// GetLastDiscardedPairs 最後に捨てたペア数
func (o *OldMaid) GetLastDiscardedPairs() int {
	return o.lastDiscardedPairs
}

// GetHasDrawn 引きが発生したかどうか
func (o *OldMaid) GetHasDrawn() bool {
	return o.hasDrawn
}
