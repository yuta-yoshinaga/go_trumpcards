package domain

import (
	"fmt"
	"math"
	"math/rand"
)

// SevensPlayerCnt 7並べプレイヤー数
const SevensPlayerCnt = 4

// sevensNoScore 評価値の初期値 (未評価を表す最小値)
const sevensNoScore = math.MinInt

// SevensCpuAction CPUまたは人間の1ターン分の行動記録
type SevensCpuAction struct {
	PlayerIdx   int   // 行動したプレイヤーインデックス
	PlayedCard  *Card // 出したカード (nil = パスまたは失格)
	TargetSuit  int   // ジョーカー配置先スート (ジョーカー以外は0)
	TargetValue int   // ジョーカー配置先値 (ジョーカー以外は0)
}

// Sevens 7並べゲームクラス
// ボードは各スートごとにビットマスクで管理する (bit i = 値iが配置済み)
type Sevens struct {
	trumpCards  *TrumpCards
	players     []*SevensPlayer
	currentTurn int                // 現在の手番プレイヤーインデックス
	tablePlaced [5]uint16          // tablePlaced[suit] = ビットマスク (bit i = 値iが配置済み)
	config      SevensConfig       // ゲーム設定
	gameEndFlag bool               // ゲーム終了フラグ
	cpuActions  []*SevensCpuAction // 人間ターン後のCPUの行動履歴
	humanAction *SevensCpuAction   // 人間の最後の行動
}

// NewSevens コンストラクタ
func NewSevens(trumpCards *TrumpCards, players []*SevensPlayer, config SevensConfig) *Sevens {
	s := &Sevens{
		trumpCards:  trumpCards,
		players:     players,
		currentTurn: 0,
		config:      config,
		gameEndFlag: false,
		cpuActions:  nil,
		humanAction: nil,
	}
	// 初期状態: 7のみ配置
	for i := 1; i <= 4; i++ {
		s.tablePlaced[i] = 1 << 7
	}
	return s
}

// Reset ゲーム初期化
func (s *Sevens) Reset() {
	s.gameEndFlag = false
	s.currentTurn = 0
	s.cpuActions = nil
	s.humanAction = nil
	for i := 1; i <= 4; i++ {
		s.tablePlaced[i] = 1 << 7
	}
	s.tablePlaced[0] = 0

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

// isPositionPlaced 指定スート・値がボード上に配置済みか判定
func (s *Sevens) isPositionPlaced(suit, value int) bool {
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return false
	}
	if value < 1 || value > 13 {
		return false
	}
	return s.tablePlaced[suit]&(1<<uint(value)) != 0
}

// isPositionPlayable 指定スート・値がボード上に配置可能か判定
func (s *Sevens) isPositionPlayable(suit, value int) bool {
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return false
	}
	if value < 1 || value > 13 {
		return false
	}
	if s.isPositionPlaced(suit, value) {
		return false
	}
	// 隣接する値が配置済みか確認
	if s.isPositionPlaced(suit, value+1) {
		return true
	}
	if s.isPositionPlaced(suit, value-1) {
		return true
	}
	// トンネルルール: A(1)↔K(13)の循環
	if s.config.TunnelEnabled {
		if value == 1 && s.isPositionPlaced(suit, 13) {
			return true
		}
		if value == 13 && s.isPositionPlaced(suit, 1) {
			return true
		}
	}
	return false
}

// hasAnyPlayablePosition ボード上に配置可能なポジションがあるか判定
func (s *Sevens) hasAnyPlayablePosition() bool {
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		for value := 1; value <= 13; value++ {
			if s.isPositionPlayable(suit, value) {
				return true
			}
		}
	}
	return false
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

	// ジョーカー: ボード上に配置可能なポジションが1つでもあれば出せる
	if suit == CardDesignJoker {
		return s.hasAnyPlayablePosition()
	}

	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return false
	}

	return s.isPositionPlayable(suit, value)
}

// placeCard ボードにカードを置く (ビットマスクを更新)
func (s *Sevens) placeCard(card *Card) {
	suit := card.GetDesign()
	value := card.GetValue()
	if suit >= CardDesignSpade && suit <= CardDesignDiamond && value >= 1 && value <= 13 {
		s.tablePlaced[suit] |= 1 << uint(value)
	}
}

// placePosition 指定スート・値をボードに配置
func (s *Sevens) placePosition(suit, value int) {
	if suit >= CardDesignSpade && suit <= CardDesignDiamond && value >= 1 && value <= 13 {
		s.tablePlaced[suit] |= 1 << uint(value)
	}
}

// PlayerPlay 人間プレイヤーがカードを出す (または パスする)
// idx: 出すカードのインデックス。-1 の場合はパス。
func (s *Sevens) PlayerPlay(idx int) error {
	if s.gameEndFlag {
		return ErrGameEnded
	}
	if !s.players[s.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	// 人間のターン開始時にCPU行動履歴をリセット
	s.cpuActions = nil

	player := s.players[s.currentTurn]

	if idx < 0 {
		// パス
		if !player.CanPass() {
			return ErrCannotPass
		}
		player.IncrPassesUsed()
		s.humanAction = &SevensCpuAction{PlayerIdx: s.currentTurn, PlayedCard: nil}
		s.advanceTurn()
		return nil
	}

	// カードを出す
	card := player.GetCard(idx)
	if card == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", idx))
	}
	if !s.IsPlayable(card) {
		return NewDomainError(ErrInvalidPlay, "card cannot be played on the board")
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
	return nil
}

// PlayerPlayJoker 人間プレイヤーがジョーカーを指定ポジションに出す
// cardIdx: ジョーカーの手札インデックス
// targetSuit: 配置先スート, targetValue: 配置先値
func (s *Sevens) PlayerPlayJoker(cardIdx, targetSuit, targetValue int) error {
	if s.gameEndFlag {
		return ErrGameEnded
	}
	if !s.players[s.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	s.cpuActions = nil

	player := s.players[s.currentTurn]
	card := player.GetCard(cardIdx)
	if card == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", cardIdx))
	}
	if card.GetDesign() != CardDesignJoker {
		return NewDomainError(ErrInvalidCard, "card is not a joker")
	}
	if !s.isPositionPlayable(targetSuit, targetValue) {
		return NewDomainError(ErrInvalidPlay, "target position is not playable")
	}

	s.placePosition(targetSuit, targetValue)
	playedCard := player.RemoveCard(cardIdx)
	s.humanAction = &SevensCpuAction{
		PlayerIdx:   s.currentTurn,
		PlayedCard:  playedCard,
		TargetSuit:  targetSuit,
		TargetValue: targetValue,
	}

	if player.GetCardsSize() == 0 {
		s.assignRank(s.currentTurn)
	}
	if !s.checkGameEnd() {
		s.advanceTurn()
	}
	return nil
}

// sevensPlay CPUの1手分の情報
type sevensPlay struct {
	cardIdx     int
	targetSuit  int // ジョーカー用 (通常カードは0)
	targetValue int // ジョーカー用 (通常カードは0)
	score       int
}

// findBestPlay CPUにとって最適な1手を探す
// 戻り値: cardIdx, targetSuit, targetValue (-1 = 出せるカードなし)
func (s *Sevens) findBestPlay(player *SevensPlayer) (int, int, int) {
	if s.config.CpuStrategy {
		return s.findPlayableStrategic(player)
	}
	return s.findPlayableSimple(player)
}

// findPlayableSimple 最初に見つかった出せるカードを返す (戦略なし)
func (s *Sevens) findPlayableSimple(player *SevensPlayer) (int, int, int) {
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == CardDesignJoker {
			// ジョーカー: 最初に見つかった配置可能ポジションに出す
			suit, value := s.findFirstPlayablePosition()
			if suit > 0 {
				return i, suit, value
			}
			continue
		}
		if s.IsPlayable(card) {
			return i, 0, 0
		}
	}
	return -1, 0, 0
}

// findPlayableStrategic 戦略的に最適な1手を探す
func (s *Sevens) findPlayableStrategic(player *SevensPlayer) (int, int, int) {
	var plays []sevensPlay

	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == CardDesignJoker {
			// ジョーカー: 最も評価の高いポジションに配置
			bestSuit, bestValue, bestScore := s.evaluateJokerPlays(player)
			if bestSuit > 0 {
				plays = append(plays, sevensPlay{
					cardIdx:     i,
					targetSuit:  bestSuit,
					targetValue: bestValue,
					score:       bestScore,
				})
			}
			continue
		}
		if s.IsPlayable(card) {
			score := s.evaluatePlay(player, card)
			plays = append(plays, sevensPlay{
				cardIdx: i,
				score:   score,
			})
		}
	}

	if len(plays) == 0 {
		return -1, 0, 0
	}

	// 最高スコアの手を選択
	best := plays[0]
	for _, p := range plays[1:] {
		if p.score > best.score {
			best = p
		}
	}

	// 全ての手がマイナス評価で、パスの余裕がある場合はパスを選択
	if best.score < 0 && player.CanPass() && player.GetPassesUsed() < player.GetMaxPasses()-1 {
		return -1, 0, 0
	}

	return best.cardIdx, best.targetSuit, best.targetValue
}

// evaluatePlay 通常カードの1手を評価
// +2: 自分が次の延長カードを持っている
// -1: 自分が次の延長カードを持っていない (相手に道を開く)
func (s *Sevens) evaluatePlay(player *SevensPlayer, card *Card) int {
	score := 0
	suit := card.GetDesign()
	value := card.GetValue()

	// 下方向の次のカード
	nextLow := value - 1
	if s.config.TunnelEnabled && value == 1 {
		nextLow = 13
	}
	if nextLow >= 1 && nextLow <= 13 && !s.isPositionPlaced(suit, nextLow) {
		if s.playerHasCard(player, suit, nextLow) {
			score += 2
		} else {
			score--
		}
	}

	// 上方向の次のカード
	nextHigh := value + 1
	if s.config.TunnelEnabled && value == 13 {
		nextHigh = 1
	}
	if nextHigh >= 1 && nextHigh <= 13 && !s.isPositionPlaced(suit, nextHigh) {
		if s.playerHasCard(player, suit, nextHigh) {
			score += 2
		} else {
			score--
		}
	}

	return score
}

// evaluateJokerPlays ジョーカーの最適な配置先を評価
func (s *Sevens) evaluateJokerPlays(player *SevensPlayer) (int, int, int) {
	bestSuit := 0
	bestValue := 0
	bestScore := sevensNoScore

	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		for value := 1; value <= 13; value++ {
			if !s.isPositionPlayable(suit, value) {
				continue
			}
			// ジョーカーを置いた場合の評価: 自分がその先のカードを持っているか
			tmpCard := NewCard(suit, value, false)
			score := s.evaluatePlay(player, tmpCard)
			if score > bestScore {
				bestScore = score
				bestSuit = suit
				bestValue = value
			}
		}
	}

	return bestSuit, bestValue, bestScore
}

// playerHasCard プレイヤーが指定スート・値のカードを持っているか
func (s *Sevens) playerHasCard(player *SevensPlayer, suit, value int) bool {
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() == suit && c.GetValue() == value {
			return true
		}
	}
	return false
}

// findFirstPlayablePosition ボード上で最初の配置可能ポジションを返す
func (s *Sevens) findFirstPlayablePosition() (int, int) {
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		for value := 1; value <= 13; value++ {
			if s.isPositionPlayable(suit, value) {
				return suit, value
			}
		}
	}
	return 0, 0
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (s *Sevens) CpuPlay() {
	if s.gameEndFlag || s.players[s.currentTurn].GetIsHuman() {
		return
	}
	playerIdx := s.currentTurn
	player := s.players[playerIdx]

	playIdx, targetSuit, targetValue := s.findBestPlay(player)

	if playIdx >= 0 {
		card := player.GetCard(playIdx)
		if card.GetDesign() == CardDesignJoker {
			// ジョーカー: 指定ポジションに配置
			s.placePosition(targetSuit, targetValue)
		} else {
			// 通常カード
			s.placeCard(card)
		}
		playedCard := player.RemoveCard(playIdx)
		action := &SevensCpuAction{
			PlayerIdx:   playerIdx,
			PlayedCard:  playedCard,
			TargetSuit:  targetSuit,
			TargetValue: targetValue,
		}
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
	// ジョーカーはボードに配置しない (スキップ)
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() != CardDesignJoker {
			s.placeCard(card)
		}
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

// GetTablePlaced ボードのビットマスク取得
func (s *Sevens) GetTablePlaced() [5]uint16 { return s.tablePlaced }

// GetConfig ゲーム設定取得
func (s *Sevens) GetConfig() SevensConfig { return s.config }

// GetTableMinVals ボードの各スートの最小値取得 (ビットマスクから算出)
func (s *Sevens) GetTableMinVals() [5]int {
	var mins [5]int
	for suit := 1; suit <= 4; suit++ {
		mins[suit] = 7 // デフォルト (7は常に配置済み)
		for v := 1; v <= 13; v++ {
			if s.tablePlaced[suit]&(1<<uint(v)) != 0 {
				mins[suit] = v
				break
			}
		}
	}
	return mins
}

// GetTableMaxVals ボードの各スートの最大値取得 (ビットマスクから算出)
func (s *Sevens) GetTableMaxVals() [5]int {
	var maxs [5]int
	for suit := 1; suit <= 4; suit++ {
		maxs[suit] = 7 // デフォルト
		for v := 13; v >= 1; v-- {
			if s.tablePlaced[suit]&(1<<uint(v)) != 0 {
				maxs[suit] = v
				break
			}
		}
	}
	return maxs
}

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
