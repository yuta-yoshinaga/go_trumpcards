package entities

// DaifugoPlayerCnt 大富豪プレイヤー数
const DaifugoPlayerCnt = 4

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

// DaifugoCpuAction CPUまたは人間の1ターン分の行動記録
type DaifugoCpuAction struct {
	PlayerIdx   int     // 行動したプレイヤーインデックス
	PlayedCards []*Card // 出したカード (nil = パス)
}

// Daifugo 大富豪ゲームクラス
type Daifugo struct {
	trumpCards        *TrumpCards
	players           []*DaifugoPlayer
	currentTurn       int               // 現在の手番プレイヤーインデックス
	tableCards        []*Card           // 場に出されているカード (nil = 場はクリア)
	lastPlayPlayerIdx int               // 最後にカードを出したプレイヤーインデックス (-1 = なし)
	gameEndFlag       bool              // ゲーム終了フラグ
	passCount         int               // 最後の出し以降の連続パス数
	cpuActions        []*DaifugoCpuAction // 人間ターン後のCPUの行動履歴
	humanAction       *DaifugoCpuAction   // 人間の最後の行動
}

// NewDaifugo コンストラクタ
func NewDaifugo(trumpCards *TrumpCards, players []*DaifugoPlayer) *Daifugo {
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
	}
}

// Reset ゲーム初期化
func (d *Daifugo) Reset() {
	d.gameEndFlag = false
	d.currentTurn = 0
	d.tableCards = nil
	d.lastPlayPlayerIdx = -1
	d.passCount = 0
	d.cpuActions = nil
	d.humanAction = nil

	// シャッフル
	for i := 0; i < 10; i++ {
		d.trumpCards.Shuffle()
	}

	// 全プレイヤーのカードリセット
	for _, p := range d.players {
		p.Reset()
		p.SetIsFinished(false)
		p.SetRank(-1)
	}

	// 全カードを配る (ジョーカーなしの52枚)
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
		p.SortCards()
	}
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
		d.tableCards = nil
		d.lastPlayPlayerIdx = -1
		d.passCount = 0
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
		d.tableCards = nil
		d.lastPlayPlayerIdx = -1
		d.passCount = 0
	}
}

// isPlayable 指定したカード (全て同じ値) が場のカードに対して出せるか判定
func (d *Daifugo) isPlayable(cardValues []int) bool {
	if len(cardValues) == 0 {
		return false
	}
	// 全て同じ値か確認
	for i := 1; i < len(cardValues); i++ {
		if cardValues[i] != cardValues[0] {
			return false
		}
	}
	if d.tableCards == nil {
		// 場がクリアなら何でも出せる
		return true
	}
	// 枚数が一致しているか
	if len(cardValues) != len(d.tableCards) {
		return false
	}
	// 場より強いか
	tableStrength := DaifugoCardStrength(d.tableCards[0].GetValue())
	proposedStrength := DaifugoCardStrength(cardValues[0])
	return proposedStrength > tableStrength
}

// PlayerPlay 人間プレイヤーがカードを出す (または パスする)
// indices: 出すカードのインデックス。空の場合はパス。
// 成功した場合 true を返す。
func (d *Daifugo) PlayerPlay(indices []int) bool {
	if d.gameEndFlag || !d.players[d.currentTurn].GetIsHuman() {
		return false
	}
	// 人間のターン開始時にCPU行動履歴をリセット
	d.cpuActions = nil

	if len(indices) == 0 {
		// パス
		d.passCount++
		d.humanAction = &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: nil}
		d.advanceTurn()
		d.checkPassClear()
		return true
	}

	// 指定カードの値を収集
	player := d.players[d.currentTurn]
	values := make([]int, len(indices))
	for i, idx := range indices {
		card := player.GetCard(idx)
		if card == nil {
			return false
		}
		values[i] = card.GetValue()
	}
	if !d.isPlayable(values) {
		return false
	}

	// カードを出す
	cards := player.RemoveCards(indices)
	d.tableCards = cards
	d.lastPlayPlayerIdx = d.currentTurn
	d.passCount = 0
	d.humanAction = &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: cards}

	if player.GetCardsSize() == 0 {
		d.finishPlayer(d.currentTurn)
	}
	if !d.checkGameEnd() {
		d.advanceTurn()
	}
	return true
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
		cards := player.RemoveCards(playIndices)
		d.tableCards = cards
		d.lastPlayPlayerIdx = playerIdx
		d.passCount = 0
		action := &DaifugoCpuAction{PlayerIdx: playerIdx, PlayedCards: cards}
		d.cpuActions = append(d.cpuActions, action)

		if player.GetCardsSize() == 0 {
			d.finishPlayer(playerIdx)
		}
		if !d.checkGameEnd() {
			d.advanceTurn()
		}
	}
}

// findBestPlay プレイヤーが出せる最弱のカードセットのインデックスを返す
// 出せるカードがない場合は nil を返す
func (d *Daifugo) findBestPlay(player *DaifugoPlayer) []int {
	if d.tableCards == nil {
		// 場がクリアなら最弱の1枚を出す
		if player.GetCardsSize() > 0 {
			return []int{0}
		}
		return nil
	}
	needed := len(d.tableCards)
	tableStrength := DaifugoCardStrength(d.tableCards[0].GetValue())

	// 手札は強さ順でソート済み。同じ値の連続するグループを探す。
	i := 0
	for i < player.GetCardsSize() {
		v := player.GetCard(i).GetValue()
		j := i
		for j < player.GetCardsSize() && player.GetCard(j).GetValue() == v {
			j++
		}
		count := j - i
		if count >= needed && DaifugoCardStrength(v) > tableStrength {
			indices := make([]int, needed)
			for k := 0; k < needed; k++ {
				indices[k] = i + k
			}
			return indices
		}
		i = j
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
