package entities

import "math/rand"

// SevensPlayerCnt 7並べプレイヤー数
const SevensPlayerCnt = 4

// SevensCpuAction CPUまたは人間の1ターン分の行動記録
type SevensCpuAction struct {
	PlayerIdx  int   // 行動したプレイヤーインデックス
	PlayedCard *Card // 出したカード (nil = パスまたは失格)
}

// Sevens 7並べゲームクラス
// ボードは各スートごとにmin値とmax値で管理する (初期値はすべて7)
type Sevens struct {
	trumpCards   *TrumpCards
	players      []*SevensPlayer
	currentTurn  int           // 現在の手番プレイヤーインデックス
	tableMinVals [5]int        // tableMinVals[suit] = そのスートのボード上の最小値 (1-4, 0番未使用)
	tableMaxVals [5]int        // tableMaxVals[suit] = そのスートのボード上の最大値
	gameEndFlag  bool          // ゲーム終了フラグ
	cpuActions   []*SevensCpuAction // 人間ターン後のCPUの行動履歴
	humanAction  *SevensCpuAction   // 人間の最後の行動
}

// NewSevens コンストラクタ
func NewSevens(trumpCards *TrumpCards, players []*SevensPlayer) *Sevens {
	s := &Sevens{
		trumpCards:  trumpCards,
		players:     players,
		currentTurn: 0,
		gameEndFlag: false,
		cpuActions:  nil,
		humanAction: nil,
	}
	for i := 0; i < 5; i++ {
		s.tableMinVals[i] = 7
		s.tableMaxVals[i] = 7
	}
	return s
}

// Reset ゲーム初期化
func (s *Sevens) Reset() {
	s.gameEndFlag = false
	s.currentTurn = 0
	s.cpuActions = nil
	s.humanAction = nil
	for i := 0; i < 5; i++ {
		s.tableMinVals[i] = 7
		s.tableMaxVals[i] = 7
	}

	// 全プレイヤーのリセット
	for _, p := range s.players {
		p.Reset()
		p.SetIsFinished(false)
		p.SetIsEliminated(false)
		p.SetRank(-1)
		p.ResetPasses()
	}

	// プレイ順をランダムにする
	rand.Shuffle(len(s.players), func(i, j int) {
		s.players[i], s.players[j] = s.players[j], s.players[i]
	})

	// シャッフルして配る
	s.trumpCards.Shuffle()
	idx := 0
	for {
		card := s.trumpCards.DrawCard()
		if card == nil {
			break
		}
		s.players[idx%SevensPlayerCnt].AddCard(card)
		idx++
	}

	// 全プレイヤーの7をボードに出す
	for _, p := range s.players {
		p.RemoveSevens()
	}

	// 手札をソート
	for _, p := range s.players {
		p.SortCards()
	}
}

// countFinished 終了済み (上がり・失格) プレイヤー数を返す
func (s *Sevens) countFinished() int {
	cnt := 0
	for _, p := range s.players {
		if p.GetIsFinished() {
			cnt++
		}
	}
	return cnt
}

// countNormalFinished 正常上がり (手札0枚) プレイヤー数を返す
func (s *Sevens) countNormalFinished() int {
	cnt := 0
	for _, p := range s.players {
		if p.GetIsFinished() && !p.GetIsEliminated() {
			cnt++
		}
	}
	return cnt
}

// countEliminated 失格プレイヤー数を返す
func (s *Sevens) countEliminated() int {
	cnt := 0
	for _, p := range s.players {
		if p.GetIsEliminated() {
			cnt++
		}
	}
	return cnt
}

// getActivePlayerCnt アクティブ (未終了) プレイヤー数取得
func (s *Sevens) getActivePlayerCnt() int {
	return len(s.players) - s.countFinished()
}

// getNextActivePlayer fromの次のアクティブなプレイヤーインデックスを取得
func (s *Sevens) getNextActivePlayer(from int) int {
	for i := 1; i <= SevensPlayerCnt; i++ {
		next := (from + i) % SevensPlayerCnt
		if !s.players[next].GetIsFinished() {
			return next
		}
	}
	return -1
}

// checkGameEnd ゲーム終了チェック (残り1人以下なら終了)
func (s *Sevens) checkGameEnd() bool {
	active := s.getActivePlayerCnt()
	if active <= 1 {
		for i, p := range s.players {
			if !p.GetIsFinished() {
				s.assignRank(i)
				break
			}
		}
		s.gameEndFlag = true
		return true
	}
	return false
}

// assignRank 正常上がりプレイヤーにランクを付与 (現在の正常上がり数+1)
func (s *Sevens) assignRank(idx int) {
	rank := s.countNormalFinished() + 1
	s.players[idx].SetIsFinished(true)
	s.players[idx].SetRank(rank)
}

// advanceTurn 手番を次のアクティブなプレイヤーへ進める
func (s *Sevens) advanceTurn() {
	if s.gameEndFlag {
		return
	}
	next := s.getNextActivePlayer(s.currentTurn)
	if next >= 0 {
		s.currentTurn = next
	}
}

// hasPlayableCard プレイヤーが出せるカードを持っているか確認
func (s *Sevens) hasPlayableCard(player *SevensPlayer) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		if s.IsPlayable(player.GetCard(i)) {
			return true
		}
	}
	return false
}

// HasAnyOption 指定プレイヤーが何らかの行動 (出す/パス) を取れるか
func (s *Sevens) HasAnyOption(playerIdx int) bool {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return false
	}
	player := s.players[playerIdx]
	return s.hasPlayableCard(player) || player.CanPass()
}

// IsPlayable カードが現在のボード状態に出せるか判定
func (s *Sevens) IsPlayable(card *Card) bool {
	if card == nil {
		return false
	}
	suit := card.GetDesign()
	value := card.GetValue()
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return false
	}
	// 最小値の左 (min-1) または最大値の右 (max+1) に出せる
	leftOk := value == s.tableMinVals[suit]-1 && s.tableMinVals[suit] > 1
	rightOk := value == s.tableMaxVals[suit]+1 && s.tableMaxVals[suit] < 13
	return leftOk || rightOk
}

// placeCard ボードにカードを置く (tableMinVals/tableMaxVals を更新)
func (s *Sevens) placeCard(card *Card) {
	suit := card.GetDesign()
	value := card.GetValue()
	if value < s.tableMinVals[suit] {
		s.tableMinVals[suit] = value
	} else {
		s.tableMaxVals[suit] = value
	}
}

// PlayerPlay 人間プレイヤーがカードを出す (または パスする)
// idx: 出すカードのインデックス。-1 の場合はパス。
// 成功した場合 true を返す。
func (s *Sevens) PlayerPlay(idx int) bool {
	if s.gameEndFlag || !s.players[s.currentTurn].GetIsHuman() {
		return false
	}
	// 人間のターン開始時にCPU行動履歴をリセット
	s.cpuActions = nil

	player := s.players[s.currentTurn]

	if idx < 0 {
		// パス
		if !player.CanPass() {
			return false
		}
		player.IncrPassesUsed()
		s.humanAction = &SevensCpuAction{PlayerIdx: s.currentTurn, PlayedCard: nil}
		s.advanceTurn()
		return true
	}

	// カードを出す
	card := player.GetCard(idx)
	if card == nil {
		return false
	}
	if !s.IsPlayable(card) {
		return false
	}

	s.placeCard(card)
	playedCard := player.RemoveCard(idx)
	s.humanAction = &SevensCpuAction{PlayerIdx: s.currentTurn, PlayedCard: playedCard}

	if player.GetCardsSize() == 0 {
		s.assignRank(s.currentTurn)
	}
	if !s.checkGameEnd() {
		s.advanceTurn()
	}
	return true
}

// findPlayableCard プレイヤーが出せる最初のカードのインデックスを返す
// 出せるカードがない場合は -1 を返す
func (s *Sevens) findPlayableCard(player *SevensPlayer) int {
	for i := 0; i < player.GetCardsSize(); i++ {
		if s.IsPlayable(player.GetCard(i)) {
			return i
		}
	}
	return -1
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (s *Sevens) CpuPlay() {
	if s.gameEndFlag || s.players[s.currentTurn].GetIsHuman() {
		return
	}
	playerIdx := s.currentTurn
	player := s.players[playerIdx]

	playIdx := s.findPlayableCard(player)

	if playIdx >= 0 {
		// カードを出す
		card := player.GetCard(playIdx)
		s.placeCard(card)
		playedCard := player.RemoveCard(playIdx)
		action := &SevensCpuAction{PlayerIdx: playerIdx, PlayedCard: playedCard}
		s.cpuActions = append(s.cpuActions, action)

		if player.GetCardsSize() == 0 {
			s.assignRank(playerIdx)
		}
		if !s.checkGameEnd() {
			s.advanceTurn()
		}
	} else if player.CanPass() {
		// パス
		player.IncrPassesUsed()
		action := &SevensCpuAction{PlayerIdx: playerIdx, PlayedCard: nil}
		s.cpuActions = append(s.cpuActions, action)
		s.advanceTurn()
	} else {
		// パスも不可 → 失格
		s.eliminatePlayer(playerIdx)
		if !s.checkGameEnd() {
			s.advanceTurn()
		}
	}
}

// eliminatePlayer プレイヤーを失格にする
// 残り手札をボードに強制配置して他プレイヤーのデッドロックを防ぎ、
// 失格プレイヤーには下位ランクを付与する (最初の失格=最下位)
func (s *Sevens) eliminatePlayer(idx int) {
	player := s.players[idx]
	// 残り手札をボードに強制配置してシーケンスのブロックを解除する
	for i := 0; i < player.GetCardsSize(); i++ {
		s.placeCard(player.GetCard(i))
	}
	// 手札をクリア
	for player.GetCardsSize() > 0 {
		player.RemoveCard(0)
	}
	// 失格ランクは下位から割り当て (最初の失格=最下位)
	rank := SevensPlayerCnt - s.countEliminated()
	player.SetIsEliminated(true)
	player.SetIsFinished(true)
	player.SetRank(rank)
}

// AutoHandleNoOption 現在のプレイヤーに選択肢がない場合の自動処理
// (出せるカードもパスも不可 → 失格)
func (s *Sevens) AutoHandleNoOption() {
	playerIdx := s.currentTurn
	action := &SevensCpuAction{PlayerIdx: playerIdx, PlayedCard: nil}
	if s.players[playerIdx].GetIsHuman() {
		s.humanAction = action
		s.cpuActions = nil
	} else {
		s.cpuActions = append(s.cpuActions, action)
	}
	s.eliminatePlayer(playerIdx)
	if !s.checkGameEnd() {
		s.advanceTurn()
	}
}

// IsHumanTurn 現在の手番が人間かどうか
func (s *Sevens) IsHumanTurn() bool {
	return s.players[s.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (s *Sevens) GetCurrentTurn() int { return s.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (s *Sevens) GetGameEndFlag() bool { return s.gameEndFlag }

// GetTableMinVals ボードの各スートの最小値取得
func (s *Sevens) GetTableMinVals() [5]int { return s.tableMinVals }

// GetTableMaxVals ボードの各スートの最大値取得
func (s *Sevens) GetTableMaxVals() [5]int { return s.tableMaxVals }

// GetPlayer プレイヤー取得
func (s *Sevens) GetPlayer(idx int) *SevensPlayer {
	if idx < 0 || idx >= len(s.players) {
		return nil
	}
	return s.players[idx]
}

// GetPlayerCnt プレイヤー数取得
func (s *Sevens) GetPlayerCnt() int { return len(s.players) }

// GetCpuActions CPUターンの行動履歴取得
func (s *Sevens) GetCpuActions() []*SevensCpuAction { return s.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (s *Sevens) GetHumanAction() *SevensCpuAction { return s.humanAction }
