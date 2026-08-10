package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// SpeedPlayerCnt プレイヤー数 (human + CPU)
const SpeedPlayerCnt = 2

// SpeedHandSize 手札の最大枚数
const SpeedHandSize = 4

// SpeedCenterPileCnt 台札の数
const SpeedCenterPileCnt = 2

// SpeedPhase ゲームフェーズ
type SpeedPhase int

// Speedのフェーズ定数
const (
	// SpeedPhasePlay プレイフェーズ
	SpeedPhasePlay SpeedPhase = 0
	// SpeedPhaseStuck 膠着フェーズ (めくりが必要)
	SpeedPhaseStuck SpeedPhase = 1
	// SpeedPhaseGameEnd ゲーム終了
	SpeedPhaseGameEnd SpeedPhase = 2
)

// SpeedCpuAction CPUが行ったアクション
type SpeedCpuAction struct {
	CardIndex int // 手札インデックス (出した時点での)
	PileIndex int // 台札インデックス
}

// Speed スピードゲームクラス
type Speed struct {
	trumpCards  *TrumpCards
	players     [SpeedPlayerCnt]*SpeedPlayer
	config      SpeedConfig
	phase       SpeedPhase
	centerPiles [SpeedCenterPileCnt]*Card // 各台札のトップカード
	gameEndFlag bool
	winnerIdx   int
	actionLogBase
}

// NewSpeed コンストラクタ
func NewSpeed(trumpCards *TrumpCards, players []*SpeedPlayer, config SpeedConfig) *Speed {
	s := &Speed{
		trumpCards: trumpCards,
		config:     config,
		winnerIdx:  -1,
	}
	for i := 0; i < SpeedPlayerCnt && i < len(players); i++ {
		s.players[i] = players[i]
	}
	return s
}

// NewDefaultSpeed returns Speed with the standard 2-player setup (1 human, 1 CPU)
// and DefaultSpeedConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultSpeed() *Speed {
	players := []*SpeedPlayer{
		NewSpeedPlayer(true),
		NewSpeedPlayer(false),
	}
	return NewSpeed(NewTrumpCards(0), players, DefaultSpeedConfig())
}

// Reset ゲームをリセットして新しいゲームを開始する
func (s *Speed) Reset() {
	s.phase = SpeedPhasePlay
	s.gameEndFlag = false
	s.winnerIdx = -1
	s.actionLog = nil

	// プレイヤーリセット
	for _, p := range s.players {
		p.Reset()
		p.ResetDrawPile()
		p.SetIsFinished(false)
	}

	// シャッフル
	s.trumpCards.Shuffle()

	// カード配布: 各プレイヤーに 手札4枚 + 台札1枚 + 山札21枚 = 26枚
	cards := make([]*Card, 0, CardCnt)
	for range CardCnt {
		c := s.trumpCards.DrawCard()
		if c != nil {
			cards = append(cards, c)
		}
	}

	half := len(cards) / 2
	for pi := range SpeedPlayerCnt {
		start := pi * half
		chunk := cards[start : start+half]

		// 最初の4枚を手札に
		for i := 0; i < SpeedHandSize && i < len(chunk); i++ {
			s.players[pi].AddCard(chunk[i])
		}
		// 次の1枚を台札に
		if len(chunk) > SpeedHandSize {
			s.centerPiles[pi] = chunk[SpeedHandSize]
		}
		// 残りを山札に
		if len(chunk) > SpeedHandSize+1 {
			s.players[pi].AddToDrawPile(chunk[SpeedHandSize+1:]...)
		}
	}

	s.updatePhase()
}

// isAdjacentRank 2つのランクが隣接しているか (K↔A ラップ)
func isAdjacentRank(a, b int) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff == 1 || diff == CardValueMax-1
}

// CanPlay 指定プレイヤーの手札カードが指定台札に出せるか
func (s *Speed) CanPlay(playerIdx, cardIdx, pileIdx int) bool {
	if playerIdx < 0 || playerIdx >= SpeedPlayerCnt {
		return false
	}
	if pileIdx < 0 || pileIdx >= SpeedCenterPileCnt {
		return false
	}
	card := s.players[playerIdx].GetCard(cardIdx)
	if card == nil {
		return false
	}
	pile := s.centerPiles[pileIdx]
	if pile == nil {
		return false
	}
	return isAdjacentRank(card.GetValue(), pile.GetValue())
}

// PlayerHasAnyPlay 指定プレイヤーに出せる手がある場合 true
func (s *Speed) PlayerHasAnyPlay(playerIdx int) bool {
	if playerIdx < 0 || playerIdx >= SpeedPlayerCnt {
		return false
	}
	p := s.players[playerIdx]
	for ci := 0; ci < p.GetCardsSize(); ci++ {
		for pi := range SpeedCenterPileCnt {
			if s.CanPlay(playerIdx, ci, pi) {
				return true
			}
		}
	}
	return false
}

// PlayerPlay 人間プレイヤーがカードを出す
func (s *Speed) PlayerPlay(cardIndex, pileIndex int) error {
	if s.phase != SpeedPhasePlay {
		return ErrWrongPhase
	}
	if pileIndex < 0 || pileIndex >= SpeedCenterPileCnt {
		return ErrInvalidPlay
	}
	p := s.players[0]
	card := p.GetCard(cardIndex)
	if card == nil {
		return ErrInvalidCard
	}
	pile := s.centerPiles[pileIndex]
	if pile == nil {
		return ErrInvalidPlay
	}
	if !isAdjacentRank(card.GetValue(), pile.GetValue()) {
		return ErrInvalidPlay
	}

	played := p.RemoveCard(cardIndex)
	s.centerPiles[pileIndex] = played
	p.RefillHand(SpeedHandSize)

	s.appendLog(0, "play", fmt.Sprintf("→ pile %d", pileIndex), []*Card{played})

	s.checkWin()
	return nil
}

// CpuPlay CPUがカードを出す (難易度に応じて複数枚)
func (s *Speed) CpuPlay() []*SpeedCpuAction {
	if s.phase != SpeedPhasePlay || s.gameEndFlag {
		return nil
	}

	switch s.config.CpuDifficulty {
	case SpeedCpuDifficultyEasy:
		return s.cpuPlayEasy()
	case SpeedCpuDifficultyHard:
		return s.cpuPlayHard()
	default:
		return s.cpuPlayGreedy()
	}
}

// cpuPlayEasy ランダムに1枚だけ出す
func (s *Speed) cpuPlayEasy() []*SpeedCpuAction {
	p := s.players[1]
	var valid []SpeedCpuAction
	for ci := 0; ci < p.GetCardsSize(); ci++ {
		for pi := range SpeedCenterPileCnt {
			if s.CanPlay(1, ci, pi) {
				valid = append(valid, SpeedCpuAction{CardIndex: ci, PileIndex: pi})
			}
		}
	}
	if len(valid) == 0 {
		return nil
	}
	chosen := valid[rand.Intn(len(valid))]
	played := p.RemoveCard(chosen.CardIndex)
	s.centerPiles[chosen.PileIndex] = played
	p.RefillHand(SpeedHandSize)

	s.appendLog(1, "play", fmt.Sprintf("→ pile %d", chosen.PileIndex), []*Card{played})
	s.checkWin()

	return []*SpeedCpuAction{&chosen}
}

// cpuPlayGreedy 貪欲に複数枚出す
func (s *Speed) cpuPlayGreedy() []*SpeedCpuAction {
	var actions []*SpeedCpuAction
	for !s.gameEndFlag {
		p := s.players[1]
		bestCI, bestPI := -1, -1
		bestDiff := -1
		for ci := 0; ci < p.GetCardsSize(); ci++ {
			for pi := range SpeedCenterPileCnt {
				if s.CanPlay(1, ci, pi) {
					card := p.GetCard(ci)
					pile := s.centerPiles[pi]
					diff := card.GetValue() - pile.GetValue()
					if diff < 0 {
						diff = -diff
					}
					if diff > bestDiff {
						bestDiff = diff
						bestCI = ci
						bestPI = pi
					}
				}
			}
		}
		if bestCI < 0 {
			break
		}
		played := p.RemoveCard(bestCI)
		s.centerPiles[bestPI] = played
		p.RefillHand(SpeedHandSize)

		s.appendLog(1, "play", fmt.Sprintf("→ pile %d", bestPI), []*Card{played})

		actions = append(actions, &SpeedCpuAction{CardIndex: bestCI, PileIndex: bestPI})
		s.checkWin()
	}
	return actions
}

// cpuPlayHard ブロッキングとコンボを考慮した戦略的プレイ
func (s *Speed) cpuPlayHard() []*SpeedCpuAction {
	var actions []*SpeedCpuAction
	for !s.gameEndFlag {
		p := s.players[1]
		bestCI, bestPI := -1, -1
		bestScore := -1000
		bestDiff := -1
		for ci := 0; ci < p.GetCardsSize(); ci++ {
			card := p.GetCard(ci)
			score := s.scoreHardPlay(card)
			for pi := range SpeedCenterPileCnt {
				if s.CanPlay(1, ci, pi) {
					pile := s.centerPiles[pi]
					diff := card.GetValue() - pile.GetValue()
					if diff < 0 {
						diff = -diff
					}
					if score > bestScore || (score == bestScore && diff > bestDiff) {
						bestScore = score
						bestDiff = diff
						bestCI = ci
						bestPI = pi
					}
				}
			}
		}
		if bestCI < 0 {
			break
		}
		played := p.RemoveCard(bestCI)
		s.centerPiles[bestPI] = played
		p.RefillHand(SpeedHandSize)

		s.appendLog(1, "play", fmt.Sprintf("→ pile %d", bestPI), []*Card{played})

		actions = append(actions, &SpeedCpuAction{CardIndex: bestCI, PileIndex: bestPI})
		s.checkWin()
	}
	return actions
}

// scoreHardPlay カードを出した場合のスコアを返す
func (s *Speed) scoreHardPlay(card *Card) int {
	newValue := card.GetValue()

	// 自分の残り手札で新しい台札値に隣接するカード数 (コンボ)
	ownFuture := s.countAdjacentCards(1, newValue)

	// 相手の手札で新しい台札値に隣接するカード数 (ブロッキング)
	opponentPlays := s.countAdjacentCards(0, newValue)

	return ownFuture*10 - opponentPlays*15
}

// countAdjacentCards 指定プレイヤーの手札のうち value に隣接するカードの枚数を返す
func (s *Speed) countAdjacentCards(playerIdx, value int) int {
	if playerIdx < 0 || playerIdx >= SpeedPlayerCnt {
		return 0
	}
	p := s.players[playerIdx]
	count := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c != nil && isAdjacentRank(c.GetValue(), value) {
			count++
		}
	}
	return count
}

// Flip 膠着時に台札をめくる
func (s *Speed) Flip() error {
	if s.phase != SpeedPhaseStuck {
		return ErrWrongPhase
	}

	flipped := false
	for pi := range SpeedCenterPileCnt {
		p := s.players[pi]
		if p.GetDrawPileSize() > 0 {
			ok := p.DrawToHand()
			if ok {
				// 手札の最後のカードを台札にする
				lastIdx := p.GetCardsSize() - 1
				card := p.RemoveCard(lastIdx)
				s.centerPiles[pi] = card
				s.appendLog(pi, "flip", "flip center", []*Card{card})
				flipped = true
			}
		}
	}

	if !flipped {
		// 山札が尽きて膠着: カード残数の少ないプレイヤーが勝ち
		s.resolveByCardCount()
		return nil
	}

	s.phase = SpeedPhasePlay
	s.updatePhase()
	return nil
}

// resolveByCardCount カード残数で勝敗を決める
func (s *Speed) resolveByCardCount() {
	cnt0 := s.players[0].GetCardsSize() + s.players[0].GetDrawPileSize()
	cnt1 := s.players[1].GetCardsSize() + s.players[1].GetDrawPileSize()
	s.gameEndFlag = true
	s.phase = SpeedPhaseGameEnd
	if cnt0 <= cnt1 {
		s.winnerIdx = 0
	} else {
		s.winnerIdx = 1
	}
}

// IsStuck 膠着状態かどうか
func (s *Speed) IsStuck() bool {
	return !s.PlayerHasAnyPlay(0) && !s.PlayerHasAnyPlay(1)
}

// checkWin 勝利判定
func (s *Speed) checkWin() {
	for i, p := range s.players {
		if !p.HasCards() {
			s.gameEndFlag = true
			s.winnerIdx = i
			s.phase = SpeedPhaseGameEnd
			p.SetIsFinished(true)
			return
		}
	}
}

// updatePhase フェーズ更新
func (s *Speed) updatePhase() {
	if s.gameEndFlag {
		return
	}
	if s.IsStuck() {
		s.phase = SpeedPhaseStuck
	} else {
		s.phase = SpeedPhasePlay
	}
}

// GetHint 人間プレイヤーへのヒントを返す
func (s *Speed) GetHint() (cardIdx, pileIdx int, found bool) {
	p := s.players[0]
	for ci := 0; ci < p.GetCardsSize(); ci++ {
		for pi := range SpeedCenterPileCnt {
			if s.CanPlay(0, ci, pi) {
				return ci, pi, true
			}
		}
	}
	return -1, -1, false
}

// --- Getters ---

// GetPhase フェーズ取得
func (s *Speed) GetPhase() SpeedPhase { return s.phase }

// GetGameEndFlag ゲーム終了フラグ取得
func (s *Speed) GetGameEndFlag() bool { return s.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得
func (s *Speed) GetWinnerIdx() int { return s.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (s *Speed) GetPlayerCnt() int { return SpeedPlayerCnt }

// GetPlayer プレイヤー取得
func (s *Speed) GetPlayer(i int) *SpeedPlayer {
	if i < 0 || i >= SpeedPlayerCnt {
		return nil
	}
	return s.players[i]
}

// GetCenterPile 台札取得
func (s *Speed) GetCenterPile(i int) *Card {
	if i < 0 || i >= SpeedCenterPileCnt {
		return nil
	}
	return s.centerPiles[i]
}

// GetConfig 設定取得
func (s *Speed) GetConfig() SpeedConfig { return s.config }

// SetConfig 設定更新
func (s *Speed) SetConfig(cfg SpeedConfig) { s.config = cfg }

// IsHumanTurn Speedでは常に人間のターン (同時プレイ)
func (s *Speed) IsHumanTurn() bool { return s.phase == SpeedPhasePlay }

// UpdatePhase 外部からフェーズ更新 (Interactor用)
func (s *Speed) UpdatePhase() {
	s.updatePhase()
}

// --- JSON ---

// speedJSON is the JSON wire format for Speed.
type speedJSON struct {
	TrumpCards  *TrumpCards                  `json:"tc"`
	Players     [SpeedPlayerCnt]*SpeedPlayer `json:"ps"`
	Config      SpeedConfig                  `json:"cf"`
	Phase       SpeedPhase                   `json:"ph"`
	CenterPiles [SpeedCenterPileCnt]*Card    `json:"cp"`
	GameEndFlag bool                         `json:"ge"`
	WinnerIdx   int                          `json:"wi"`
	ActionLog   []*ActionLogEntry            `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (s *Speed) MarshalJSON() ([]byte, error) {
	return json.Marshal(speedJSON{
		TrumpCards:  s.trumpCards,
		Players:     s.players,
		Config:      s.config,
		Phase:       s.phase,
		CenterPiles: s.centerPiles,
		GameEndFlag: s.gameEndFlag,
		WinnerIdx:   s.winnerIdx,
		ActionLog:   s.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Speed) UnmarshalJSON(data []byte) error {
	var j speedJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	s.trumpCards = j.TrumpCards
	s.players = j.Players
	s.config = j.Config
	s.phase = j.Phase
	s.centerPiles = j.CenterPiles
	s.gameEndFlag = j.GameEndFlag
	s.winnerIdx = j.WinnerIdx
	s.actionLog = j.ActionLog
	return nil
}
