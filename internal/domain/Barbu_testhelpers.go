//go:build test

package domain

// Barbu_testhelpers.go は決定的なテストのための構築・設定ヘルパー。
// 実際の配札 (シャッフル) に依存せず、手札・場・トリックを直接組み立てる。

// BarbuTestNew はテスト用 Barbu を生成する (4 人、配札なし)。
// 既定では dealNumber=0, dealerIdx=0 (= human), phase=SelectContract。
func BarbuTestNew(config BarbuConfig) *Barbu {
	players := make([]*BarbuPlayer, BarbuPlayerCnt)
	players[0] = NewBarbuPlayer(true)
	for i := 1; i < BarbuPlayerCnt; i++ {
		players[i] = NewBarbuPlayer(false)
	}
	b := NewBarbu(NewTrumpCards(0), players, config)
	b.actionLog = make([]*ActionLogEntry, 0)
	return b
}

// BarbuTestSetHand はテスト用にプレイヤーの手札を直接設定する。
func (b *Barbu) BarbuTestSetHand(playerIdx int, cards []*Card) {
	p := b.players[playerIdx]
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// BarbuTestAddTrick はテスト用に捕獲トリックを追加する (得点テスト用)。
func (b *Barbu) BarbuTestAddTrick(playerIdx int, cards []*Card) {
	b.players[playerIdx].AddTrick(cards)
}

// BarbuTestSetPhase はテスト用にフェーズを設定する。
func (b *Barbu) BarbuTestSetPhase(phase string) { b.phase = phase }

// BarbuTestSetDealNumber はテスト用にディール番号を設定する。
func (b *Barbu) BarbuTestSetDealNumber(n int) { b.dealNumber = n }

// BarbuTestSetDealer はテスト用にディーラーを設定する。
func (b *Barbu) BarbuTestSetDealer(idx int) { b.dealerIdx = idx }

// BarbuTestSetContract はテスト用に現在のコントラクトと切り札を設定する。
func (b *Barbu) BarbuTestSetContract(contract, trumpSuit int) {
	b.currentContract = contract
	b.trumpSuit = trumpSuit
}

// BarbuTestSetCurrentPlayer はテスト用に手番を設定する。
func (b *Barbu) BarbuTestSetCurrentPlayer(idx int) { b.currentPlayer = idx }

// BarbuTestSetLeadPlayer はテスト用にリードプレイヤーを設定する。
func (b *Barbu) BarbuTestSetLeadPlayer(idx int) { b.leadPlayer = idx }

// BarbuTestSetTrickNumber はテスト用にトリック番号を設定する。
func (b *Barbu) BarbuTestSetTrickNumber(n int) { b.trickNumber = n }

// BarbuTestSetCurrentTrick はテスト用に進行中トリックを設定する。
func (b *Barbu) BarbuTestSetCurrentTrick(trick []*TrickCard) { b.currentTrick = trick }

// BarbuTestSetLastTrickWinner はテスト用に直前トリックの勝者を設定する。
func (b *Barbu) BarbuTestSetLastTrickWinner(idx int) { b.lastTrickWinner = idx }

// BarbuTestSetTablePlaced はテスト用に Dominoes の場を設定する。
func (b *Barbu) BarbuTestSetTablePlaced(table [5]uint16) { b.tablePlaced = table }

// BarbuTestSetGameEnd はテスト用にゲーム終了フラグを設定する。
func (b *Barbu) BarbuTestSetGameEnd(flag bool) { b.gameEndFlag = flag }

// BarbuTestSetUsedContract はテスト用に使用済みコントラクトを設定する。
func (b *Barbu) BarbuTestSetUsedContract(dealerIdx, contract int, used bool) {
	b.usedContracts[dealerIdx][contract] = used
}

// BarbuTestApplyTrickPlay は human/CPU の別を問わずトリックプレイを実行する
// (テスト用。手番ガードを経由せずトリック処理本体を直接呼ぶ)。
func (b *Barbu) BarbuTestApplyTrickPlay(playerIdx, handIdx int) error {
	return b.applyTrickPlay(playerIdx, handIdx)
}

// BarbuTestScoreDeal はテスト用に scoreDeal を直接呼ぶ。
func (b *Barbu) BarbuTestScoreDeal() *BarbuDealDetail { return b.scoreDeal() }

// BarbuTestFinishDeal はテスト用に finishDeal を直接呼ぶ。
func (b *Barbu) BarbuTestFinishDeal() { b.finishDeal() }

// BarbuTestAppendDealHistory はテスト用に完了ディールの内訳を積む。
// 7 ディール分を実際に打たせずに、集計表示だけを確かめられるようにする。
func (b *Barbu) BarbuTestAppendDealHistory(detail *BarbuDealDetail) {
	b.dealHistory = append(b.dealHistory, detail)
}
