package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// SpiteAndMalicePhase Spite & Malice ゲームフェーズ
type SpiteAndMalicePhase int

// SpiteAndMalice のフェーズ定数
const (
	// SpiteAndMalicePhasePlaying プレイ中
	SpiteAndMalicePhasePlaying SpiteAndMalicePhase = iota
	// SpiteAndMalicePhaseGameOver 決着済み
	SpiteAndMalicePhaseGameOver
)

// Spite & Malice のゲーム構造定数
const (
	// SpiteAndMalicePlayerCnt プレイヤー数 (人間 + CPU)
	SpiteAndMalicePlayerCnt = 2
	// SpiteAndMaliceHumanIdx 人間プレイヤーのインデックス
	SpiteAndMaliceHumanIdx = 0
	// SpiteAndMaliceCpuIdx CPU プレイヤーのインデックス
	SpiteAndMaliceCpuIdx = 1
	// SpiteAndMaliceFoundationCnt 中央共有ファウンデーションの数
	SpiteAndMaliceFoundationCnt = 4
	// SpiteAndMaliceSideCnt 1プレイヤーあたりのサイドパイル数
	SpiteAndMaliceSideCnt = 4
	// SpiteAndMaliceHandMax 手札最大枚数（ターン開始時に補充する目標サイズ）
	SpiteAndMaliceHandMax = 5
	// SpiteAndMaliceWildValue ワイルドカード (K) の値
	SpiteAndMaliceWildValue = 13
	// SpiteAndMaliceFoundationMax 1ファウンデーションが完成する枚数 (A..Q = 12)
	SpiteAndMaliceFoundationMax = 12
)

// SpiteAndMaliceSource プレイ元の種別
type SpiteAndMaliceSource int

// SpiteAndMaliceSource 定数
const (
	// SpiteAndMaliceSourceHand 手札から
	SpiteAndMaliceSourceHand SpiteAndMaliceSource = iota
	// SpiteAndMaliceSourceGoal ゴールパイルのトップから
	SpiteAndMaliceSourceGoal
	// SpiteAndMaliceSourceSide サイドパイルのトップから
	SpiteAndMaliceSourceSide
)

// SpiteAndMaliceHint Spite & Malice のヒント（次の推奨手）。
// Source がプレイ元、Index がソース側の補助インデックス (hand/side のとき)、
// FoundationIdx が配置先中央ファウンデーション、Discard は true の場合は
// ディスカード推奨を表し、その場合 Index は hand の位置、
// FoundationIdx はサイドパイルの位置を表す。
type SpiteAndMaliceHint struct {
	Source        SpiteAndMaliceSource
	Index         int
	FoundationIdx int
	Discard       bool
}

// SpiteAndMalice Spite & Malice ゲームクラス
type SpiteAndMalice struct {
	trumpCards  *TrumpCards
	stock       []*Card
	completed   []*Card
	foundations [SpiteAndMaliceFoundationCnt][]*Card
	players     [SpiteAndMalicePlayerCnt]*SpiteAndMalicePlayer
	current     int
	phase       SpiteAndMalicePhase
	moveCount   int
	winner      int
	actionLog   []*ActionLogEntry
	config      SpiteAndMaliceConfig
}

// NewSpiteAndMalice コンストラクタ
func NewSpiteAndMalice(trumpCards *TrumpCards, cfg SpiteAndMaliceConfig) *SpiteAndMalice {
	g := &SpiteAndMalice{trumpCards: trumpCards, config: cfg, winner: -1}
	for i := range g.players {
		g.players[i] = NewSpiteAndMalicePlayer(i == SpiteAndMaliceCpuIdx)
	}
	return g
}

// NewDefaultSpiteAndMalice 既定デッキ (2 デッキ 104 枚) の Spite & Malice を返す
func NewDefaultSpiteAndMalice() *SpiteAndMalice {
	return NewSpiteAndMalice(NewTrumpCardsWithDecks(2, 0), DefaultSpiteAndMaliceConfig())
}

// Reset ゲーム初期化
func (s *SpiteAndMalice) Reset() {
	if err := s.config.Validate(); err != nil {
		// 不正設定の場合はデフォルトへフォールバック
		s.config = DefaultSpiteAndMaliceConfig()
	}
	s.trumpCards.Shuffle()
	s.phase = SpiteAndMalicePhasePlaying
	s.moveCount = 0
	s.winner = -1
	s.actionLog = nil
	s.completed = nil
	for i := range SpiteAndMaliceFoundationCnt {
		s.foundations[i] = nil
	}
	for _, p := range s.players {
		p.Reset()
	}

	// まず全カードを 1 枚ずつ引きストックへ
	s.stock = s.stock[:0]
	for s.trumpCards.GetRemainingCount() > 0 {
		s.stock = append(s.stock, s.trumpCards.DrawCard())
	}

	// 各プレイヤーにゴールパイルを配る (ストックの頭から)
	goalSize := s.config.GoalSize
	for i := range SpiteAndMalicePlayerCnt {
		for range goalSize {
			if len(s.stock) == 0 {
				break
			}
			card := s.stock[0]
			s.stock = s.stock[1:]
			s.players[i].AddToGoal(card)
		}
	}

	// 手札 5 枚を配る
	for i := range SpiteAndMalicePlayerCnt {
		s.drawToHand(i, SpiteAndMaliceHandMax)
	}

	s.current = SpiteAndMaliceHumanIdx
}

// --- Public API ---

// PlayFromHand 手札のカードをファウンデーションに出す
func (s *SpiteAndMalice) PlayFromHand(handIdx, foundationIdx int) error {
	if err := s.assertPlayable(); err != nil {
		return err
	}
	p := s.players[s.current]
	hand := p.GetHand()
	if handIdx < 0 || handIdx >= len(hand) {
		return errors.New("invalid hand index")
	}
	if foundationIdx < 0 || foundationIdx >= SpiteAndMaliceFoundationCnt {
		return errors.New("invalid foundation index")
	}
	card := hand[handIdx]
	if !s.canPlaceOnFoundation(card, foundationIdx) {
		return errors.New("cannot place card on foundation")
	}
	p.RemoveFromHand(handIdx)
	s.appendToFoundation(foundationIdx, card)
	s.moveCount++
	s.appendLog("playHand", fmt.Sprintf("プレイヤー%dが手札からファウンデーション%dへ", s.current, foundationIdx+1), []*Card{card})
	s.afterPlay()
	return nil
}

// PlayFromGoal ゴールパイルのトップをファウンデーションに出す
func (s *SpiteAndMalice) PlayFromGoal(foundationIdx int) error {
	if err := s.assertPlayable(); err != nil {
		return err
	}
	p := s.players[s.current]
	top := p.GoalTop()
	if top == nil {
		return errors.New("goal pile is empty")
	}
	if foundationIdx < 0 || foundationIdx >= SpiteAndMaliceFoundationCnt {
		return errors.New("invalid foundation index")
	}
	if !s.canPlaceOnFoundation(top, foundationIdx) {
		return errors.New("cannot place card on foundation")
	}
	card := p.PopGoal()
	s.appendToFoundation(foundationIdx, card)
	s.moveCount++
	s.appendLog("playGoal", fmt.Sprintf("プレイヤー%dがゴール→ファウンデーション%dへ", s.current, foundationIdx+1), []*Card{card})
	if p.GoalSize() == 0 {
		s.winner = s.current
		s.phase = SpiteAndMalicePhaseGameOver
		s.appendLog("win", fmt.Sprintf("プレイヤー%dがゴールパイルを出し切った", s.current), nil)
		return nil
	}
	s.afterPlay()
	return nil
}

// PlayFromSide サイドパイルのトップをファウンデーションに出す
func (s *SpiteAndMalice) PlayFromSide(sideIdx, foundationIdx int) error {
	if err := s.assertPlayable(); err != nil {
		return err
	}
	if sideIdx < 0 || sideIdx >= SpiteAndMaliceSideCnt {
		return errors.New("invalid side index")
	}
	if foundationIdx < 0 || foundationIdx >= SpiteAndMaliceFoundationCnt {
		return errors.New("invalid foundation index")
	}
	p := s.players[s.current]
	top := p.SideTop(sideIdx)
	if top == nil {
		return errors.New("side pile is empty")
	}
	if !s.canPlaceOnFoundation(top, foundationIdx) {
		return errors.New("cannot place card on foundation")
	}
	card := p.PopSide(sideIdx)
	s.appendToFoundation(foundationIdx, card)
	s.moveCount++
	s.appendLog("playSide", fmt.Sprintf("プレイヤー%dがサイド%d→ファウンデーション%dへ", s.current, sideIdx+1, foundationIdx+1), []*Card{card})
	s.afterPlay()
	return nil
}

// Discard 手札 1 枚をサイドパイルに置いてターンを終了する
func (s *SpiteAndMalice) Discard(handIdx, sideIdx int) error {
	if err := s.assertPlayable(); err != nil {
		return err
	}
	p := s.players[s.current]
	hand := p.GetHand()
	if handIdx < 0 || handIdx >= len(hand) {
		return errors.New("invalid hand index")
	}
	if sideIdx < 0 || sideIdx >= SpiteAndMaliceSideCnt {
		return errors.New("invalid side index")
	}
	card := hand[handIdx]
	p.RemoveFromHand(handIdx)
	p.PushSide(sideIdx, card)
	s.moveCount++
	s.appendLog("discard", fmt.Sprintf("プレイヤー%dがサイド%dへディスカードしてターン終了", s.current, sideIdx+1), []*Card{card})
	s.endTurn()
	return nil
}

// IsCpuTurn 現在のターンが CPU か
func (s *SpiteAndMalice) IsCpuTurn() bool {
	if s.phase == SpiteAndMalicePhaseGameOver {
		return false
	}
	return s.players[s.current].GetIsCpu()
}

// CpuStep CPU の手番を 1 ステップ進める
func (s *SpiteAndMalice) CpuStep() error {
	if !s.IsCpuTurn() {
		return errors.New("not cpu turn")
	}
	hint := s.bestCpuMove()
	if hint == nil {
		return errors.New("no cpu move available")
	}
	if hint.Discard {
		return s.Discard(hint.Index, hint.FoundationIdx)
	}
	switch hint.Source {
	case SpiteAndMaliceSourceGoal:
		return s.PlayFromGoal(hint.FoundationIdx)
	case SpiteAndMaliceSourceHand:
		return s.PlayFromHand(hint.Index, hint.FoundationIdx)
	case SpiteAndMaliceSourceSide:
		return s.PlayFromSide(hint.Index, hint.FoundationIdx)
	}
	return errors.New("invalid cpu hint")
}

// AutoComplete plays every immediately-legal foundation move from the human's
// own piles (goal → hand → side) without rotating the turn. It stops when no
// playable move remains or the game ends. Discards are intentionally skipped
// since those rotate the turn and may give the opponent a strategic opening,
// which the player should always confirm by hand.
//
// CanAutoComplete must guard the call site — it returns true only when the
// human is on turn and at least one safe foundation move exists.
func (s *SpiteAndMalice) AutoComplete() error {
	if s.phase != SpiteAndMalicePhasePlaying {
		return errors.New("game is over")
	}
	if s.IsCpuTurn() {
		return errors.New("not human turn")
	}
	// Bounded loop: AutoComplete only ever processes the human's piles, but
	// completed foundations refill from the shared stock so a single call can
	// chain more moves than CardCnt * 2 in pathological end-game states. Use
	// a generous cap (2 decks * full deck rotations) — the loop normally exits
	// via "no playable move" long before this and the cap is just defensive.
	const maxIterations = 1024
	for range maxIterations {
		if s.phase != SpiteAndMalicePhasePlaying || s.IsCpuTurn() {
			return nil
		}
		move := s.findPlayableMove(s.current)
		if move == nil {
			return nil
		}
		var err error
		switch move.Source {
		case SpiteAndMaliceSourceGoal:
			err = s.PlayFromGoal(move.FoundationIdx)
		case SpiteAndMaliceSourceHand:
			err = s.PlayFromHand(move.Index, move.FoundationIdx)
		case SpiteAndMaliceSourceSide:
			err = s.PlayFromSide(move.Index, move.FoundationIdx)
		default:
			return errors.New("invalid auto-complete move")
		}
		if err != nil {
			return err
		}
	}
	return errors.New("auto-complete exceeded iteration cap")
}

// CanAutoComplete reports whether the player has a playable foundation move
// available right now. The frontend uses this to enable / disable the button.
func (s *SpiteAndMalice) CanAutoComplete() bool {
	if s.phase != SpiteAndMalicePhasePlaying || s.IsCpuTurn() {
		return false
	}
	return s.findPlayableMove(s.current) != nil
}

// GetHint 現在ターンの推奨手を返す (人間向け)。ゲーム終了時は nil。
func (s *SpiteAndMalice) GetHint() *SpiteAndMaliceHint {
	if s.phase != SpiteAndMalicePhasePlaying {
		return nil
	}
	if move := s.findPlayableMove(s.current); move != nil {
		return move
	}
	// プレイ可能な手が無ければディスカード提案
	return s.findDiscardMove(s.current)
}

// --- Getters ---

// GetPhase フェーズ取得
func (s *SpiteAndMalice) GetPhase() SpiteAndMalicePhase { return s.phase }

// GetCurrent 現在ターンのプレイヤーインデックス
func (s *SpiteAndMalice) GetCurrent() int { return s.current }

// GetMoveCount 操作回数
func (s *SpiteAndMalice) GetMoveCount() int { return s.moveCount }

// GetWinner 勝者インデックス (-1 なら未決着)
func (s *SpiteAndMalice) GetWinner() int { return s.winner }

// GetStockSize 山札残り枚数
func (s *SpiteAndMalice) GetStockSize() int { return len(s.stock) }

// GetCompletedSize 完成済み山の枚数 (ファウンデーションが Q に達して回収されたもの)
func (s *SpiteAndMalice) GetCompletedSize() int { return len(s.completed) }

// GetFoundations ファウンデーションのスナップショット
func (s *SpiteAndMalice) GetFoundations() [SpiteAndMaliceFoundationCnt][]*Card {
	var out [SpiteAndMaliceFoundationCnt][]*Card
	for i := range SpiteAndMaliceFoundationCnt {
		out[i] = append([]*Card(nil), s.foundations[i]...)
	}
	return out
}

// GetFoundationTopValue ファウンデーションのトップ値 (空の場合 0)
func (s *SpiteAndMalice) GetFoundationTopValue(foundationIdx int) int {
	if foundationIdx < 0 || foundationIdx >= SpiteAndMaliceFoundationCnt {
		return 0
	}
	pile := s.foundations[foundationIdx]
	if len(pile) == 0 {
		return 0
	}
	return s.effectiveValue(pile, len(pile)-1)
}

// GetPlayer プレイヤーを取得
func (s *SpiteAndMalice) GetPlayer(idx int) *SpiteAndMalicePlayer {
	if idx < 0 || idx >= SpiteAndMalicePlayerCnt {
		return nil
	}
	return s.players[idx]
}

// GetActionLog 棋譜取得
func (s *SpiteAndMalice) GetActionLog() []*ActionLogEntry { return s.actionLog }

// GetGameEndFlag returns true once the game has left the playing phase.
func (s *SpiteAndMalice) GetGameEndFlag() bool { return s.phase != SpiteAndMalicePhasePlaying }

// GetConfig 設定取得
func (s *SpiteAndMalice) GetConfig() SpiteAndMaliceConfig { return s.config }

// SetConfig 設定更新 (ゲーム開始前に適用する)
func (s *SpiteAndMalice) SetConfig(cfg SpiteAndMaliceConfig) { s.config = cfg }

// --- Test-only setters (build tag は testhelpers 側で管理) ---

// --- Private helpers ---

func (s *SpiteAndMalice) assertPlayable() error {
	if s.phase != SpiteAndMalicePhasePlaying {
		return errors.New("game is over")
	}
	return nil
}

// IsGoalTopPlayable reports whether the given player's goal-pile top can go
// onto any foundation right now. Emptying the goal pile is how the game is won,
// so both UIs surface this every turn — the Web GUI rings the pile, the CUI
// marks the goal line (#4876).
func (s *SpiteAndMalice) IsGoalTopPlayable(playerIdx int) bool {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return false
	}
	top := s.players[playerIdx].GoalTop()
	if top == nil {
		return false
	}
	for i := range SpiteAndMaliceFoundationCnt {
		if s.canPlaceOnFoundation(top, i) {
			return true
		}
	}
	return false
}

func (s *SpiteAndMalice) canPlaceOnFoundation(card *Card, foundationIdx int) bool {
	if card == nil {
		return false
	}
	// K はワイルド: 完成 (Q) でない限りいつでも置ける
	if card.GetValue() == SpiteAndMaliceWildValue {
		return len(s.foundations[foundationIdx]) < SpiteAndMaliceFoundationMax
	}
	pile := s.foundations[foundationIdx]
	if len(pile) >= SpiteAndMaliceFoundationMax {
		return false
	}
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	topValue := s.effectiveValue(pile, len(pile)-1)
	return card.GetValue() == topValue+1
}

// effectiveValue は foundation の i 番目のカードの実効値を返す。K (ワイルド) の場合は
// その直前のカードの実効値 +1 (空の場合は 1)。
func (s *SpiteAndMalice) effectiveValue(pile []*Card, i int) int {
	if i < 0 || i >= len(pile) {
		return 0
	}
	card := pile[i]
	if card.GetValue() != SpiteAndMaliceWildValue {
		return card.GetValue()
	}
	if i == 0 {
		return 1
	}
	return s.effectiveValue(pile, i-1) + 1
}

func (s *SpiteAndMalice) appendToFoundation(foundationIdx int, card *Card) {
	s.foundations[foundationIdx] = append(s.foundations[foundationIdx], card)
	if len(s.foundations[foundationIdx]) == SpiteAndMaliceFoundationMax {
		// Q まで積まれたら完成済みに回収
		s.completed = append(s.completed, s.foundations[foundationIdx]...)
		s.foundations[foundationIdx] = nil
		s.appendLog("complete", fmt.Sprintf("ファウンデーション%dが完成", foundationIdx+1), nil)
	}
}

func (s *SpiteAndMalice) afterPlay() {
	if s.phase == SpiteAndMalicePhaseGameOver {
		return
	}
	// 手札を使い切ったら補充してプレイ継続
	p := s.players[s.current]
	if p.HandSize() == 0 {
		s.drawToHand(s.current, SpiteAndMaliceHandMax)
	}
}

func (s *SpiteAndMalice) endTurn() {
	if s.phase == SpiteAndMalicePhaseGameOver {
		return
	}
	s.current = (s.current + 1) % SpiteAndMalicePlayerCnt
	// 次プレイヤーのターン開始: 手札を 5 枚に補充
	s.drawToHand(s.current, SpiteAndMaliceHandMax)
}

func (s *SpiteAndMalice) drawToHand(playerIdx, target int) {
	p := s.players[playerIdx]
	for p.HandSize() < target {
		if len(s.stock) == 0 {
			s.refillStockFromCompleted()
			if len(s.stock) == 0 {
				return
			}
		}
		card := s.stock[0]
		s.stock = s.stock[1:]
		p.AddToHand(card)
	}
}

// refillStockFromCompleted はストックが空になったとき、完成済み山をシャッフルして
// ストックに戻す。完成済みが空なら何もしない。
func (s *SpiteAndMalice) refillStockFromCompleted() {
	if len(s.completed) == 0 {
		return
	}
	rest := make([]*Card, len(s.completed))
	copy(rest, s.completed)
	rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
	s.stock = append(s.stock, rest...)
	s.completed = nil
	s.appendLog("refill", "完成済み山をシャッフルしてストックへ戻した", nil)
}

func (s *SpiteAndMalice) appendLog(actionType, detail string, cards []*Card) {
	s.actionLog = append(s.actionLog, &ActionLogEntry{
		TurnNumber: s.moveCount,
		PlayerIdx:  s.current,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- CPU / Hint logic ---

// findPlayableMove はファウンデーションに出せる最善の手 (ゴール→手札→サイド) を返す。
func (s *SpiteAndMalice) findPlayableMove(playerIdx int) *SpiteAndMaliceHint {
	p := s.players[playerIdx]
	// 1) ゴールパイルのトップを優先
	if top := p.GoalTop(); top != nil {
		for f := range SpiteAndMaliceFoundationCnt {
			if s.canPlaceOnFoundation(top, f) {
				return &SpiteAndMaliceHint{Source: SpiteAndMaliceSourceGoal, Index: -1, FoundationIdx: f}
			}
		}
	}
	// 2) 手札 (ワイルドは最後に回す)
	hand := p.GetHand()
	for pass := range 2 {
		for i, c := range hand {
			if c == nil {
				continue
			}
			isWild := c.GetValue() == SpiteAndMaliceWildValue
			if (pass == 0 && isWild) || (pass == 1 && !isWild) {
				continue
			}
			for f := range SpiteAndMaliceFoundationCnt {
				if s.canPlaceOnFoundation(c, f) {
					return &SpiteAndMaliceHint{Source: SpiteAndMaliceSourceHand, Index: i, FoundationIdx: f}
				}
			}
		}
	}
	// 3) サイドパイルのトップ
	for side := range SpiteAndMaliceSideCnt {
		top := p.SideTop(side)
		if top == nil {
			continue
		}
		for f := range SpiteAndMaliceFoundationCnt {
			if s.canPlaceOnFoundation(top, f) {
				return &SpiteAndMaliceHint{Source: SpiteAndMaliceSourceSide, Index: side, FoundationIdx: f}
			}
		}
	}
	return nil
}

// findDiscardMove はディスカード候補 (hand → side) を返す。難易度で重み付けが変わる。
func (s *SpiteAndMalice) findDiscardMove(playerIdx int) *SpiteAndMaliceHint {
	p := s.players[playerIdx]
	hand := p.GetHand()
	if len(hand) == 0 {
		return nil
	}
	handIdx := s.pickDiscardHandIdx(playerIdx, hand)
	sideIdx := s.pickDiscardSide(p, hand[handIdx])
	return &SpiteAndMaliceHint{
		Source:        SpiteAndMaliceSourceHand,
		Index:         handIdx,
		FoundationIdx: sideIdx,
		Discard:       true,
	}
}

// pickDiscardHandIdx 難易度に応じて捨てる手札の位置を選ぶ。
// Easy: 先頭
// Normal: ワイルド (K) 以外のうち最も大きな値 (使い道が狭い)
// Hard: 相手ゴール top +1 と一致するカードを最優先で温存し、それ以外は Normal と同様
func (s *SpiteAndMalice) pickDiscardHandIdx(playerIdx int, hand []*Card) int {
	switch s.config.CpuDifficulty {
	case SpiteAndMaliceCpuDifficultyEasy:
		return 0
	case SpiteAndMaliceCpuDifficultyHard:
		opp := s.players[(playerIdx+1)%SpiteAndMalicePlayerCnt]
		var targetVal int
		if top := opp.GoalTop(); top != nil {
			v := top.GetValue()
			if v == SpiteAndMaliceWildValue {
				targetVal = 1
			} else {
				targetVal = v
			}
		}
		best := -1
		bestScore := -1
		for i, c := range hand {
			if c == nil {
				continue
			}
			if c.GetValue() == SpiteAndMaliceWildValue {
				continue
			}
			score := c.GetValue()
			// 相手の goal 送り値に一致するカードは温存 (-100 で最低スコア)
			if targetVal > 0 && c.GetValue() == targetVal {
				score -= 100
			}
			if score > bestScore {
				bestScore = score
				best = i
			}
		}
		if best == -1 {
			return 0
		}
		return best
	default: // Normal
		best := -1
		bestVal := -1
		for i, c := range hand {
			if c == nil || c.GetValue() == SpiteAndMaliceWildValue {
				continue
			}
			if c.GetValue() > bestVal {
				bestVal = c.GetValue()
				best = i
			}
		}
		if best == -1 {
			return 0
		}
		return best
	}
}

// pickDiscardSide ディスカード先サイドパイルを選ぶ。まず空、次にトップ値が近い順。
func (s *SpiteAndMalice) pickDiscardSide(p *SpiteAndMalicePlayer, card *Card) int {
	for side := range SpiteAndMaliceSideCnt {
		if p.SideSize(side) == 0 {
			return side
		}
	}
	// 全て非空: カードと同じ値またはそれに隣接する値がトップの山を優先
	best := 0
	bestDiff := 100
	for side := range SpiteAndMaliceSideCnt {
		// SideTop は前段のループで全てが非空であることを保証済み (空の山は既に return 済み)。
		top := p.SideTop(side)
		diff := intAbs(top.GetValue() - card.GetValue())
		if diff < bestDiff {
			bestDiff = diff
			best = side
		}
	}
	return best
}

// bestCpuMove 次の 1 手を返す。プレイ可能な手が無ければディスカード。
func (s *SpiteAndMalice) bestCpuMove() *SpiteAndMaliceHint {
	if move := s.findPlayableMove(s.current); move != nil {
		return move
	}
	return s.findDiscardMove(s.current)
}

func intAbs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// --- JSON ---

// spiteAndMaliceJSON is the JSON wire format for SpiteAndMalice.
type spiteAndMaliceJSON struct {
	TrumpCards  *TrumpCards                                    `json:"tc"`
	Stock       []*Card                                        `json:"st"`
	Completed   []*Card                                        `json:"cp"`
	Foundations [SpiteAndMaliceFoundationCnt][]*Card           `json:"fd"`
	Players     [SpiteAndMalicePlayerCnt]*SpiteAndMalicePlayer `json:"pl"`
	Current     int                                            `json:"cu"`
	Phase       SpiteAndMalicePhase                            `json:"ph"`
	MoveCount   int                                            `json:"mc"`
	Winner      int                                            `json:"wn"`
	ActionLog   []*ActionLogEntry                              `json:"al"`
	Config      SpiteAndMaliceConfig                           `json:"cf"`
}

// MarshalJSON implements json.Marshaler.
func (s *SpiteAndMalice) MarshalJSON() ([]byte, error) {
	return json.Marshal(spiteAndMaliceJSON{
		TrumpCards:  s.trumpCards,
		Stock:       s.stock,
		Completed:   s.completed,
		Foundations: s.foundations,
		Players:     s.players,
		Current:     s.current,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		Winner:      s.winner,
		ActionLog:   s.actionLog,
		Config:      s.config,
	})
}

// spiteAndMaliceMaxSliceLen caps slice sizes during deserialisation.
const spiteAndMaliceMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (s *SpiteAndMalice) UnmarshalJSON(data []byte) error {
	var j spiteAndMaliceJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > spiteAndMaliceMaxSliceLen ||
		len(j.Completed) > spiteAndMaliceMaxSliceLen ||
		len(j.ActionLog) > spiteAndMaliceMaxSliceLen {
		return fmt.Errorf("spiteandmalice: input array exceeds maximum allowed size")
	}
	for i := range SpiteAndMaliceFoundationCnt {
		if len(j.Foundations[i]) > spiteAndMaliceMaxSliceLen {
			return fmt.Errorf("spiteandmalice: foundation %d exceeds maximum allowed size", i)
		}
	}
	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCardsWithDecks(2, 0)
	}
	s.stock = j.Stock
	if s.stock == nil {
		s.stock = make([]*Card, 0)
	}
	s.completed = j.Completed
	if s.completed == nil {
		s.completed = make([]*Card, 0)
	}
	s.foundations = j.Foundations
	for i := range SpiteAndMaliceFoundationCnt {
		if s.foundations[i] == nil {
			s.foundations[i] = make([]*Card, 0)
		}
	}
	for i, p := range j.Players {
		if p != nil {
			s.players[i] = p
		} else {
			s.players[i] = NewSpiteAndMalicePlayer(i == SpiteAndMaliceCpuIdx)
		}
	}
	s.current = j.Current
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.winner = j.Winner
	if s.phase != SpiteAndMalicePhaseGameOver {
		s.winner = -1
	}
	s.actionLog = j.ActionLog
	if s.actionLog == nil {
		s.actionLog = make([]*ActionLogEntry, 0)
	}
	s.config = j.Config
	if err := s.config.Validate(); err != nil {
		s.config = DefaultSpiteAndMaliceConfig()
	}
	return nil
}
