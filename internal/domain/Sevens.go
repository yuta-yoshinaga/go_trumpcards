package domain

import (
	"encoding/json"
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
	PlayerIdx   int   `json:"pi"` // 行動したプレイヤーインデックス
	PlayedCard  *Card `json:"pc"` // 出したカード (nil = パスまたは失格)
	TargetSuit  int   `json:"ts"` // ジョーカー配置先スート (ジョーカー以外は0)
	TargetValue int   `json:"tv"` // ジョーカー配置先値 (ジョーカー以外は0)
	ForcedPass  bool  `json:"fp"` // true = 出せるカードがなくパスした
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
	jokerPlaced [5]uint16          // jokerPlaced[suit] = ジョーカーが配置されたポジションのビットマスク
	jokerCards  []*Card            // ボード上のジョーカーカードオブジェクト (回収用)
	actionLogBase
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

// NewDefaultSevens returns Sevens with the standard 4-player setup (1 human, 3 CPU)
// and DefaultSevensConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultSevens() *Sevens {
	config := DefaultSevensConfig()
	players := []*SevensPlayer{
		NewSevensPlayer(true),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
	}
	return NewSevens(NewTrumpCards(config.JokerCount), players, config)
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
	s.jokerPlaced = [5]uint16{}
	s.jokerCards = nil
	s.actionLog = nil

	// 全プレイヤーのリセット
	resetPlayers(s.players, func(p *SevensPlayer) {
		p.SetIsEliminated(false)
		p.SetRank(-1)
		p.ResetPasses()
		p.SetMaxPasses(s.config.MaxPasses)
		p.SetLastPlayedJoker(false)
	})

	// プレイ順をランダムにする
	rand.Shuffle(len(s.players), func(i, j int) {
		s.players[i], s.players[j] = s.players[j], s.players[i]
	})

	// デッキをジョーカー枚数に合わせて再生成しシャッフル
	s.trumpCards = NewTrumpCards(s.config.JokerCount)
	s.trumpCards.Shuffle()
	dealAllCards(s.trumpCards, s.players)

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
	return countPlayers(s.players, func(p *SevensPlayer) bool { return p.GetIsFinished() })
}

// countNormalFinished 正常上がり (手札0枚) プレイヤー数を返す
func (s *Sevens) countNormalFinished() int {
	return countPlayers(s.players, func(p *SevensPlayer) bool {
		return p.GetIsFinished() && !p.GetIsEliminated()
	})
}

// countEliminated 失格プレイヤー数を返す
func (s *Sevens) countEliminated() int {
	return countPlayers(s.players, func(p *SevensPlayer) bool { return p.GetIsEliminated() })
}

// getActivePlayerCnt アクティブ (未終了) プレイヤー数取得
func (s *Sevens) getActivePlayerCnt() int {
	return len(s.players) - s.countFinished()
}

// getNextActivePlayer fromの次のアクティブなプレイヤーインデックスを取得
func (s *Sevens) getNextActivePlayer(from int) int {
	return nextActivePlayer(s.players, from, 1)
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

// isEndStopped 片側ストップルールによりブロックされているか判定
func (s *Sevens) isEndStopped(suit, value int) bool {
	if !s.config.EndStopEnabled {
		return false
	}
	if value == 7 {
		return false
	}
	// 上側 (8-13): Aが配置済みならブロック
	if value > 7 && s.isPositionPlaced(suit, 1) {
		return true
	}
	// 下側 (1-6): Kが配置済みならブロック
	if value < 7 && s.isPositionPlaced(suit, 13) {
		return true
	}
	return false
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
		// ジョーカー回収が有効な場合、ジョーカーが置かれた場所はプレイ可能
		return s.config.JokerReclaimEnabled && (s.jokerPlaced[suit]&(1<<uint(value)) != 0)
	}
	// 片側ストップ: EndStopが有効でブロックされたポジションはプレイ不可
	if s.isEndStopped(suit, value) {
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
	// カスタムトンネル: ±TunnelSkipWidth の接続
	if s.config.TunnelSkipWidth >= 2 {
		low := value - s.config.TunnelSkipWidth
		high := value + s.config.TunnelSkipWidth
		if s.config.TunnelEnabled {
			low = wrapValue(low)
			high = wrapValue(high)
		}
		if low >= 1 && low <= 13 && s.isPositionPlaced(suit, low) { // range check needed when TunnelEnabled=false
			return true
		}
		if high >= 1 && high <= 13 && s.isPositionPlaced(suit, high) { // range check needed when TunnelEnabled=false
			return true
		}
	}
	return false
}

// wrapValue 値を1-13の循環範囲に収める
func wrapValue(v int) int {
	v = ((v - 1) % 13) + 1
	if v <= 0 {
		v += 13
	}
	return v
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

// hasOnlyJokers プレイヤーの手札がすべてジョーカーかどうか判定
func (s *Sevens) hasOnlyJokers(player *SevensPlayer) bool {
	if player.GetCardsSize() == 0 {
		return false
	}
	for i := 0; i < player.GetCardsSize(); i++ {
		if player.GetCard(i).GetDesign() != CardDesignJoker {
			return false
		}
	}
	return true
}

// isJokerBlockedByFinishRule ジョーカー上がり禁止ルールによりジョーカーが使用不可か判定
func (s *Sevens) isJokerBlockedByFinishRule(player *SevensPlayer) bool {
	return s.config.NoJokerFinish && s.hasOnlyJokers(player)
}

// isJokerBlockedByConsecutiveRule ジョーカー連続禁止ルールによりジョーカーが使用不可か判定
func (s *Sevens) isJokerBlockedByConsecutiveRule(player *SevensPlayer) bool {
	return s.config.JokerConsecutiveBanned && player.GetLastPlayedJoker()
}

// hasPlayableCard プレイヤーが出せるカードを持っているか確認
func (s *Sevens) hasPlayableCard(player *SevensPlayer) bool {
	jokerBlocked := s.isJokerBlockedByFinishRule(player) || s.isJokerBlockedByConsecutiveRule(player)
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if jokerBlocked && card.GetDesign() == CardDesignJoker {
			continue
		}
		if s.IsPlayable(card) {
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

// GetPlayableCardIndices は人間プレイヤーの手札のうち、いま出せるカードの
// インデックスを返す (#5479)。
//
// 判定は PlayerPlay が使う IsPlayable そのものを通す。トンネル・スキップ幅・
// ジョーカーの配置可否といった盤面の状態は全部そちらが見るので、ここで規則を
// 書き直さない。**別実装にすると「出せる」と印を付けた札が実際には弾かれる。**
//
// 戻り値は3状態を区別する:
//   - nil          … 判定していない (人間の手番でない)
//   - 空スライス    … 判定した上で1枚も出せない。7並べでは普通に起きる状態で、
//     そこでプレイヤーはパスする。**nil と同じ扱いにはできない。**
//   - 非空スライス  … 出せる手札のインデックス
func (s *Sevens) GetPlayableCardIndices() []int {
	if !s.IsHumanTurn() {
		return nil
	}
	player := s.players[s.currentTurn]
	idx := make([]int, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		if s.IsPlayable(player.GetCard(i)) {
			idx = append(idx, i)
		}
	}
	return idx
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

// recordJokerCard ジョーカー配置時にジョーカーカードオブジェクトを記録し、配置ビットマスクを更新
func (s *Sevens) recordJokerCard(card *Card, suit, value int) {
	if !s.config.JokerReclaimEnabled {
		return
	}
	s.jokerCards = append(s.jokerCards, card)
	if suit >= CardDesignSpade && suit <= CardDesignDiamond && value >= 1 && value <= 13 {
		s.jokerPlaced[suit] |= 1 << uint(value)
	}
}

// reclaimJokerIfNeeded 非ジョーカーカードでジョーカー配置済みポジションに置いた場合、ジョーカーを回収
func (s *Sevens) reclaimJokerIfNeeded(playerIdx, suit, value int) {
	if !s.config.JokerReclaimEnabled {
		return
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond || value < 1 || value > 13 {
		return
	}
	if s.jokerPlaced[suit]&(1<<uint(value)) == 0 {
		return
	}
	// ジョーカーマークをクリア
	s.jokerPlaced[suit] &^= 1 << uint(value)
	// jokerCards から1枚取り出してプレイヤーに戻す
	if len(s.jokerCards) > 0 {
		joker := s.jokerCards[len(s.jokerCards)-1]
		s.jokerCards = s.jokerCards[:len(s.jokerCards)-1]
		s.players[playerIdx].AddCard(joker)
		s.players[playerIdx].SortCards()
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
		player.SetLastPlayedJoker(false)
		s.appendLog(s.currentTurn, "pass", "pass", nil)
		s.humanAction = &SevensCpuAction{
			PlayerIdx:  s.currentTurn,
			PlayedCard: nil,
			ForcedPass: !s.hasPlayableCard(player),
		}
		s.advanceTurn()
		return nil
	}

	// カードを出す
	card := player.GetCard(idx)
	if card == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", idx))
	}
	if card.GetDesign() == CardDesignJoker {
		return NewDomainError(ErrInvalidPlay, "use PlayerPlayJoker to play a joker")
	}
	if !s.IsPlayable(card) {
		return NewDomainError(ErrInvalidPlay, "card cannot be played on the board")
	}

	s.placeCard(card)
	playedCard := player.RemoveCard(idx)
	s.reclaimJokerIfNeeded(s.currentTurn, card.GetDesign(), card.GetValue())
	player.SetLastPlayedJoker(false)
	s.appendLog(s.currentTurn, "play", fmt.Sprintf("played %s", cardLogStr(playedCard)), []*Card{playedCard})
	s.humanAction = &SevensCpuAction{PlayerIdx: s.currentTurn, PlayedCard: playedCard}

	if player.GetCardsSize() == 0 {
		s.assignRank(s.currentTurn)
		s.appendLog(-1, "finish", fmt.Sprintf("player %d finished (rank %d)", s.currentTurn, player.GetRank()), nil)
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
	if s.isJokerBlockedByFinishRule(player) {
		return NewDomainError(ErrInvalidPlay, "cannot finish with a joker")
	}
	if s.isJokerBlockedByConsecutiveRule(player) {
		return NewDomainError(ErrInvalidPlay, "cannot play joker on consecutive turns")
	}
	if !s.isPositionPlayable(targetSuit, targetValue) {
		return NewDomainError(ErrInvalidPlay, "target position is not playable")
	}

	s.placePosition(targetSuit, targetValue)
	playedCard := player.RemoveCard(cardIdx)
	s.recordJokerCard(playedCard, targetSuit, targetValue)
	player.SetLastPlayedJoker(true)
	s.appendLog(s.currentTurn, "joker", fmt.Sprintf("played joker as %s %d", suitLogStr(targetSuit), targetValue), []*Card{playedCard})
	s.humanAction = &SevensCpuAction{
		PlayerIdx:   s.currentTurn,
		PlayedCard:  playedCard,
		TargetSuit:  targetSuit,
		TargetValue: targetValue,
	}

	if player.GetCardsSize() == 0 {
		s.assignRank(s.currentTurn)
		s.appendLog(-1, "finish", fmt.Sprintf("player %d finished (rank %d)", s.currentTurn, player.GetRank()), nil)
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
	switch s.config.CpuStrategy {
	case SevensCpuStrategic:
		return s.findPlayableStrategic(player)
	case SevensCpuHarassment:
		return s.findPlayableHarassment(player)
	default:
		return s.findPlayableSimple(player)
	}
}

// findPlayableSimple 最初に見つかった出せるカードを返す (戦略なし)
func (s *Sevens) findPlayableSimple(player *SevensPlayer) (int, int, int) {
	jokerBlocked := s.isJokerBlockedByFinishRule(player) || s.isJokerBlockedByConsecutiveRule(player)
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == CardDesignJoker {
			if jokerBlocked {
				continue
			}
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
	jokerBlocked := s.isJokerBlockedByFinishRule(player) || s.isJokerBlockedByConsecutiveRule(player)

	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == CardDesignJoker {
			if jokerBlocked {
				continue
			}
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
	if best.score < 0 && (player.GetMaxPasses() == 0 || player.GetPassesUsed() < player.GetMaxPasses()-1) {
		return -1, 0, 0
	}

	return best.cardIdx, best.targetSuit, best.targetValue
}

// passUrgencyWeight パス残数に基づく緊急度重みを返す
func (s *Sevens) passUrgencyWeight(player *SevensPlayer) int {
	maxP := player.GetMaxPasses()
	if maxP == 0 {
		// 無制限パス → パス干上がりなし
		return 1
	}
	remaining := maxP - player.GetPassesUsed()
	switch {
	case remaining <= 1:
		return 3
	case remaining == 2:
		return 2
	default:
		return 1
	}
}

// countOpponentsHoldingCard 指定位置のカードを持っている相手の重み付きカウント (スキップ接続の評価用)
func (s *Sevens) countOpponentsHoldingCard(self *SevensPlayer, suit, value int) int {
	count := 0
	for _, p := range s.players {
		if p == self || p.GetIsFinished() {
			continue
		}
		if s.playerHasCard(p, suit, value) {
			count += s.passUrgencyWeight(p)
		}
	}
	return count
}

// countWeightedOpponentsBlocked パス残数で重み付けしたブロック相手数をカウント (±1方向の逐次スキャン用)
func (s *Sevens) countWeightedOpponentsBlocked(self *SevensPlayer, suit, fromValue, direction int) int {
	count := 0
	for _, p := range s.players {
		if p == self || p.GetIsFinished() {
			continue
		}
		v := fromValue
		for {
			v += direction
			if s.config.TunnelEnabled {
				if v < 1 {
					v = 13
				} else if v > 13 {
					v = 1
				}
			}
			if v < 1 || v > 13 {
				break
			}
			if s.isPositionPlaced(suit, v) {
				break
			}
			if s.playerHasCard(p, suit, v) {
				count += s.passUrgencyWeight(p)
				break
			}
			if s.config.TunnelEnabled && v == fromValue {
				break
			}
		}
	}
	return count
}

// evaluatePlay 通常カードの1手を評価
// +2: 自分が次の延長カードを持っている
// -(1+blockedOpponents): 自分が持っていない場合、ブロックしている相手が多いほど重いペナルティ
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
			score -= 1 + s.countWeightedOpponentsBlocked(player, suit, value, -1)
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
			score -= 1 + s.countWeightedOpponentsBlocked(player, suit, value, +1)
		}
	}

	// カスタムトンネル: ±TunnelSkipWidth 方向の評価
	// スキップ接続は単一位置のみ評価 (±1のような逐次スキャンではなく、距離Nの1点のみチェック)
	if s.config.TunnelSkipWidth >= 2 {
		skipLow := value - s.config.TunnelSkipWidth
		if s.config.TunnelEnabled {
			skipLow = wrapValue(skipLow)
		}
		if skipLow >= 1 && skipLow <= 13 && !s.isPositionPlaced(suit, skipLow) { // range check needed when TunnelEnabled=false
			if s.playerHasCard(player, suit, skipLow) {
				score += 2
			} else {
				score -= 1 + s.countOpponentsHoldingCard(player, suit, skipLow)
			}
		}
		skipHigh := value + s.config.TunnelSkipWidth
		if s.config.TunnelEnabled {
			skipHigh = wrapValue(skipHigh)
		}
		if skipHigh >= 1 && skipHigh <= 13 && !s.isPositionPlaced(suit, skipHigh) { // range check needed when TunnelEnabled=false
			if s.playerHasCard(player, suit, skipHigh) {
				score += 2
			} else {
				score -= 1 + s.countOpponentsHoldingCard(player, suit, skipHigh)
			}
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

// findPlayableHarassment 嫌がらせ特化で最適な1手を探す
func (s *Sevens) findPlayableHarassment(player *SevensPlayer) (int, int, int) {
	var plays []sevensPlay
	jokerBlocked := s.isJokerBlockedByFinishRule(player) || s.isJokerBlockedByConsecutiveRule(player)

	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == CardDesignJoker {
			if jokerBlocked {
				continue
			}
			bestSuit, bestValue, bestScore := s.evaluateJokerPlaysHarassment(player)
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
			score := s.evaluatePlayHarassment(player, card)
			plays = append(plays, sevensPlay{
				cardIdx: i,
				score:   score,
			})
		}
	}

	if len(plays) == 0 {
		return -1, 0, 0
	}

	best := plays[0]
	for _, p := range plays[1:] {
		if p.score > best.score {
			best = p
		}
	}

	// パス閾値: <= 0 (戦略モードの < 0 より攻撃的)
	if best.score <= 0 && (player.GetMaxPasses() == 0 || player.GetPassesUsed() < player.GetMaxPasses()-1) {
		return -1, 0, 0
	}

	return best.cardIdx, best.targetSuit, best.targetValue
}

// evaluatePlayHarassment 嫌がらせ特化の通常カード評価
// 相手の進行をブロックすることを重視し、相手を助けるプレイを避ける
func (s *Sevens) evaluatePlayHarassment(player *SevensPlayer, card *Card) int {
	score := 0
	suit := card.GetDesign()
	value := card.GetValue()

	// 下方向の次のカード
	nextLow := value - 1
	if s.config.TunnelEnabled && value == 1 {
		nextLow = 13
	}
	if nextLow >= 1 && nextLow <= 13 && !s.isPositionPlaced(suit, nextLow) {
		score += s.evaluateHarassmentDirection(player, suit, nextLow, -1)
	}

	// 上方向の次のカード
	nextHigh := value + 1
	if s.config.TunnelEnabled && value == 13 {
		nextHigh = 1
	}
	if nextHigh >= 1 && nextHigh <= 13 && !s.isPositionPlaced(suit, nextHigh) {
		score += s.evaluateHarassmentDirection(player, suit, nextHigh, +1)
	}

	// カスタムトンネル: ±TunnelSkipWidth 方向の評価
	if s.config.TunnelSkipWidth >= 2 {
		skipLow := value - s.config.TunnelSkipWidth
		if s.config.TunnelEnabled {
			skipLow = wrapValue(skipLow)
		}
		if skipLow >= 1 && skipLow <= 13 && !s.isPositionPlaced(suit, skipLow) {
			score += s.evaluateHarassmentSkipDirection(player, suit, skipLow)
		}
		skipHigh := value + s.config.TunnelSkipWidth
		if s.config.TunnelEnabled {
			skipHigh = wrapValue(skipHigh)
		}
		if skipHigh >= 1 && skipHigh <= 13 && !s.isPositionPlaced(suit, skipHigh) {
			score += s.evaluateHarassmentSkipDirection(player, suit, skipHigh)
		}
	}

	return score
}

// evaluateHarassmentPosition 嫌がらせ特化: プレイ候補先の評価
func (s *Sevens) evaluateHarassmentPosition(player *SevensPlayer, suit, value, blockedCount int) int {
	if s.anyOpponentHasCard(player, suit, value) {
		// 相手が次のカードを持っている → 出すと相手を助けてしまう
		urgency := s.passUrgencyWeight(player)
		return -3 * urgency
	}
	if s.playerHasCard(player, suit, value) {
		// 自分が次のカードを持っている
		if blockedCount > 0 {
			return 1
		}
		return 2
	}
	// 誰も持っていない → 相手がブロックされるので良い
	return 2 * blockedCount
}

// evaluateHarassmentDirection 嫌がらせ特化: ±1方向の評価
func (s *Sevens) evaluateHarassmentDirection(player *SevensPlayer, suit, nextValue, direction int) int {
	blockedCount := s.countWeightedOpponentsBlocked(player, suit, nextValue, direction)
	return s.evaluateHarassmentPosition(player, suit, nextValue, blockedCount)
}

// evaluateHarassmentSkipDirection 嫌がらせ特化: スキップ接続方向の評価
func (s *Sevens) evaluateHarassmentSkipDirection(player *SevensPlayer, suit, skipValue int) int {
	blockedCount := s.countOpponentsHoldingCard(player, suit, skipValue)
	return s.evaluateHarassmentPosition(player, suit, skipValue, blockedCount)
}

// evaluateJokerPlaysHarassment ジョーカーの最適な配置先を嫌がらせ特化で評価
func (s *Sevens) evaluateJokerPlaysHarassment(player *SevensPlayer) (int, int, int) {
	bestSuit := 0
	bestValue := 0
	bestScore := sevensNoScore

	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		for value := 1; value <= 13; value++ {
			if !s.isPositionPlayable(suit, value) {
				continue
			}
			tmpCard := NewCard(suit, value, false)
			score := s.evaluatePlayHarassment(player, tmpCard)
			if score > bestScore {
				bestScore = score
				bestSuit = suit
				bestValue = value
			}
		}
	}

	return bestSuit, bestValue, bestScore
}

// anyOpponentHasCard いずれかの対戦相手が指定カードを持っているか
func (s *Sevens) anyOpponentHasCard(self *SevensPlayer, suit, value int) bool {
	for _, p := range s.players {
		if p == self || p.GetIsFinished() {
			continue
		}
		if s.playerHasCard(p, suit, value) {
			return true
		}
	}
	return false
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
		if card.GetDesign() == CardDesignJoker {
			s.recordJokerCard(playedCard, targetSuit, targetValue)
			player.SetLastPlayedJoker(true)
			s.appendLog(playerIdx, "joker", fmt.Sprintf("played joker as %s %d", suitLogStr(targetSuit), targetValue), []*Card{playedCard})
		} else {
			s.reclaimJokerIfNeeded(playerIdx, card.GetDesign(), card.GetValue())
			player.SetLastPlayedJoker(false)
			s.appendLog(playerIdx, "play", fmt.Sprintf("played %s", cardLogStr(playedCard)), []*Card{playedCard})
		}
		action := &SevensCpuAction{
			PlayerIdx:   playerIdx,
			PlayedCard:  playedCard,
			TargetSuit:  targetSuit,
			TargetValue: targetValue,
		}
		s.cpuActions = append(s.cpuActions, action)

		if player.GetCardsSize() == 0 {
			s.assignRank(playerIdx)
			s.appendLog(-1, "finish", fmt.Sprintf("player %d finished (rank %d)", playerIdx, player.GetRank()), nil)
		}
		if !s.checkGameEnd() {
			s.advanceTurn()
		}
	} else if player.CanPass() {
		// パス
		player.IncrPassesUsed()
		player.SetLastPlayedJoker(false)
		s.appendLog(playerIdx, "pass", "pass", nil)
		action := &SevensCpuAction{
			PlayerIdx:  playerIdx,
			PlayedCard: nil,
			ForcedPass: !s.hasPlayableCard(player),
		}
		s.cpuActions = append(s.cpuActions, action)
		s.advanceTurn()
	} else {
		// パスも不可 → 失格
		s.eliminatePlayer(playerIdx)
		s.appendLog(-1, "finish", fmt.Sprintf("player %d finished (rank %d)", playerIdx, player.GetRank()), nil)
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
	action := &SevensCpuAction{PlayerIdx: playerIdx, PlayedCard: nil, ForcedPass: true}
	if s.players[playerIdx].GetIsHuman() {
		s.humanAction = action
		s.cpuActions = nil
	} else {
		s.cpuActions = append(s.cpuActions, action)
	}
	s.eliminatePlayer(playerIdx)
	s.appendLog(-1, "finish", fmt.Sprintf("player %d finished (rank %d)", playerIdx, s.players[playerIdx].GetRank()), nil)
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
	return getPlayer(s.players, idx)
}

// GetPlayerCnt プレイヤー数取得
func (s *Sevens) GetPlayerCnt() int { return len(s.players) }

// GetCpuActions CPUターンの行動履歴取得
func (s *Sevens) GetCpuActions() []*SevensCpuAction { return s.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (s *Sevens) GetHumanAction() *SevensCpuAction { return s.humanAction }

// SetConfig 設定変更
func (s *Sevens) SetConfig(config SevensConfig) {
	if config.JokerCount < 0 {
		config.JokerCount = 0
	}
	if config.JokerCount > SevensMaxJokerCount {
		config.JokerCount = SevensMaxJokerCount
	}
	if config.MaxPasses < 0 {
		config.MaxPasses = 0
	}
	if config.TunnelSkipWidth < 0 {
		config.TunnelSkipWidth = 0
	}
	if config.TunnelSkipWidth > 12 {
		config.TunnelSkipWidth = 12
	}
	if config.CpuStrategy < SevensCpuSimple || config.CpuStrategy > SevensCpuHarassment {
		config.CpuStrategy = SevensCpuSimple
	}
	s.config = config
}

// suitLogStr スートを棋譜用文字列に変換
func suitLogStr(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "spade"
	case CardDesignClover:
		return "clover"
	case CardDesignHeart:
		return "heart"
	case CardDesignDiamond:
		return "diamond"
	default:
		return "joker"
	}
}

// cardLogStr カードを棋譜用文字列に変換
func cardLogStr(card *Card) string {
	if card.GetDesign() == CardDesignJoker {
		return "joker"
	}
	return fmt.Sprintf("%s %d", suitLogStr(card.GetDesign()), card.GetValue())
}

// sevensJSON is the JSON wire format for Sevens.
type sevensJSON struct {
	TrumpCards  *TrumpCards        `json:"tc"`
	Players     []*SevensPlayer    `json:"ps"`
	CurrentTurn int                `json:"ct"`
	TablePlaced [5]uint16          `json:"tp"`
	Config      SevensConfig       `json:"cf"`
	GameEndFlag bool               `json:"ge"`
	CpuActions  []*SevensCpuAction `json:"ca"`
	HumanAction *SevensCpuAction   `json:"ha"`
	JokerPlaced [5]uint16          `json:"jp"`
	JokerCards  []*Card            `json:"jk"`
	ActionLog   []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (s *Sevens) MarshalJSON() ([]byte, error) {
	return json.Marshal(sevensJSON{
		TrumpCards:  s.trumpCards,
		Players:     s.players,
		CurrentTurn: s.currentTurn,
		TablePlaced: s.tablePlaced,
		Config:      s.config,
		GameEndFlag: s.gameEndFlag,
		CpuActions:  s.cpuActions,
		HumanAction: s.humanAction,
		JokerPlaced: s.jokerPlaced,
		JokerCards:  s.jokerCards,
		ActionLog:   s.actionLog,
	})
}

// sevensMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const sevensMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (s *Sevens) UnmarshalJSON(data []byte) error {
	var j sevensJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > sevensMaxSliceLen || len(j.CpuActions) > sevensMaxSliceLen ||
		len(j.JokerCards) > sevensMaxSliceLen || len(j.ActionLog) > sevensMaxSliceLen {
		return fmt.Errorf("sevens: input array exceeds maximum allowed size")
	}
	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCards(0)
	}
	s.players = j.Players
	if s.players == nil {
		s.players = make([]*SevensPlayer, 0)
	}
	s.currentTurn = j.CurrentTurn
	s.tablePlaced = j.TablePlaced
	s.config = j.Config
	s.gameEndFlag = j.GameEndFlag
	s.cpuActions = j.CpuActions
	if s.cpuActions == nil {
		s.cpuActions = make([]*SevensCpuAction, 0)
	}
	s.humanAction = j.HumanAction
	s.jokerPlaced = j.JokerPlaced
	s.jokerCards = j.JokerCards
	if s.jokerCards == nil {
		s.jokerCards = make([]*Card, 0)
	}
	s.actionLog = j.ActionLog
	if s.actionLog == nil {
		s.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
