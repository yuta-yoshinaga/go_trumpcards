//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// MemoryPhase 神経衰弱ゲームフェーズ
type MemoryPhase int

// Memoryのフェーズ定数
const (
	// MemoryPhaseFlip1 1枚目を選択するフェーズ
	MemoryPhaseFlip1 MemoryPhase = iota
	// MemoryPhaseFlip2 2枚目を選択するフェーズ
	MemoryPhaseFlip2
	// MemoryPhaseResult マッチ結果表示フェーズ
	MemoryPhaseResult
	// MemoryPhaseGameEnd ゲーム終了
	MemoryPhaseGameEnd
)

// MemoryPlayerCnt 神経衰弱のプレイヤー数
const MemoryPlayerCnt = 4

// MemoryBoardSize ボードのカード枚数 (52枚)
const MemoryBoardSize = 52

// MemoryBoardCard ボード上のカード
type MemoryBoardCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
	Taken  bool  `json:"t"`
	// Visited is true once the card has been turned face up at least once,
	// powering the CUI/Web memory aids for previously-seen face-down cells.
	Visited bool `json:"v"`
}

// CPU記憶パラメータ (難易度別)
const (
	memoryRetentionEasy   = 0.3
	memoryRetentionNormal = 0.6
	memoryRetentionHard   = 0.95
	memoryDecayEasy       = 0.15
	memoryDecayNormal     = 0.05
	memoryDecayHard       = 0.01
)

// Memory 神経衰弱ゲームクラス
type Memory struct {
	trumpCards       *TrumpCards
	board            []*MemoryBoardCard
	players          []*MemoryPlayer
	config           MemoryConfig
	phase            MemoryPhase
	currentPlayerIdx int
	firstFlipPos     int
	secondFlipPos    int
	lastMatchResult  bool // 直前のフリップがマッチしたかどうか
	gameEndFlag      bool
	winnerIdx        int
	turnNumber       int
	actionLog        []*ActionLogEntry
}

// NewMemory コンストラクタ
func NewMemory(trumpCards *TrumpCards, players []*MemoryPlayer, config MemoryConfig) *Memory {
	m := &Memory{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		winnerIdx:  -1,
	}
	return m
}

// NewDefaultMemory returns Memory with the standard 4-player setup (1 human, 3 CPU)
// and DefaultMemoryConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultMemory() *Memory {
	players := []*MemoryPlayer{
		NewMemoryPlayer(true),
		NewMemoryPlayer(false),
		NewMemoryPlayer(false),
		NewMemoryPlayer(false),
	}
	return NewMemory(NewTrumpCards(0), players, DefaultMemoryConfig())
}

// Reset ゲームリセット
func (m *Memory) Reset() {
	m.trumpCards.Shuffle()
	m.phase = MemoryPhaseFlip1
	m.currentPlayerIdx = 0
	m.firstFlipPos = -1
	m.secondFlipPos = -1
	m.lastMatchResult = false
	m.gameEndFlag = false
	m.winnerIdx = -1
	m.turnNumber = 0
	m.actionLog = nil

	for _, p := range m.players {
		p.ResetGame()
	}

	// ボードにカードを配置。
	//
	// 一致判定はランク基準なので、単に上から N 枚引くと相方のいないランクが
	// 残り、allTaken() に到達できずゲームが終わらなくなる。フルデッキ 52 枚では
	// 各ランクが 4 枚で自然に偶数だったが、枚数を減らすとこの前提が崩れる。
	// そこで山から引いた札をランクごとに束ね、2 枚ずつの組にしてから必要な
	// ペア数だけ採用する。 (ADR-0035)
	pairCount := m.config.NormalizedPairCount()
	byRank := make(map[int][]*Card, MemoryMaxPairCount)
	for i := 0; i < MemoryBoardSize; i++ {
		card := m.trumpCards.DrawCard()
		if card == nil {
			continue
		}
		byRank[card.GetValue()] = append(byRank[card.GetValue()], card)
	}

	pairs := make([][2]*Card, 0, MemoryMaxPairCount)
	for _, cards := range byRank {
		for i := 0; i+1 < len(cards); i += 2 {
			pairs = append(pairs, [2]*Card{cards[i], cards[i+1]})
		}
	}
	// byRank の走査順はマップなので不定。並べる前に順序を決め直す必要はないが、
	// どのランクが採用されるかが偏らないようシャッフルしてから切り詰める。
	for i := len(pairs) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		pairs[i], pairs[j] = pairs[j], pairs[i]
	}
	if len(pairs) > pairCount {
		pairs = pairs[:pairCount]
	}

	m.board = make([]*MemoryBoardCard, 0, len(pairs)*2)
	for _, pair := range pairs {
		for _, card := range pair {
			m.board = append(m.board, &MemoryBoardCard{Card: card, FaceUp: false, Taken: false})
		}
	}
	// ここまでだとペアが隣り合ってしまうので、盤面上の位置を混ぜる。
	for i := len(m.board) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		m.board[i], m.board[j] = m.board[j], m.board[i]
	}
}

// PlayerFlip 人間プレイヤーがカードをめくる
func (m *Memory) PlayerFlip(pos int) error {
	if m.gameEndFlag {
		return errors.New("game is already over")
	}
	if !m.players[m.currentPlayerIdx].GetIsHuman() {
		return errors.New("it is not your turn")
	}
	return m.flip(pos)
}

// CpuFlip CPUプレイヤーがカードをめくる (1ターン分: flip1 + flip2)
func (m *Memory) CpuFlip() {
	if m.gameEndFlag {
		return
	}
	p := m.players[m.currentPlayerIdx]
	if p.GetIsHuman() {
		return
	}

	// まず記憶から既知のペアを探す
	pos1, pos2, found := p.FindAnyKnownPair()
	if found && !m.board[pos1].Taken && !m.board[pos2].Taken {
		_ = m.flip(pos1)
		_ = m.flip(pos2)
		return
	}

	// 1枚目: ランダムに未取得・未表面のカードを選ぶ
	first := m.randomAvailablePosition(-1)
	if first < 0 {
		return
	}
	_ = m.flip(first)

	// 1枚目のrankから既知のマッチを探す
	firstRank := m.board[first].Card.GetValue()
	_, matchPos, matchFound := p.FindKnownMatch(firstRank)
	if matchFound && matchPos != first && !m.board[matchPos].Taken {
		_ = m.flip(matchPos)
		return
	}

	// 記憶にない → ランダムに2枚目を選ぶ
	second := m.randomAvailablePosition(first)
	if second < 0 {
		return
	}
	_ = m.flip(second)

}

// ResolveFlip フリップ結果を解決する
func (m *Memory) ResolveFlip() {
	if m.phase != MemoryPhaseResult {
		return
	}

	card1 := m.board[m.firstFlipPos]
	card2 := m.board[m.secondFlipPos]

	if card1.Card.GetValue() == card2.Card.GetValue() {
		// マッチ → ペアを獲得、同じプレイヤーが続行
		m.lastMatchResult = true
		card1.Taken = true
		card2.Taken = true
		card1.FaceUp = false
		card2.FaceUp = false
		m.players[m.currentPlayerIdx].AddPair(card1.Card, card2.Card)

		// 全CPUプレイヤーから取られた位置の記憶を削除
		for _, p := range m.players {
			p.RemoveMemoryAt(m.firstFlipPos)
			p.RemoveMemoryAt(m.secondFlipPos)
		}

		m.appendLog("match", fmt.Sprintf("ペア獲得: 位置%d,位置%d", m.firstFlipPos, m.secondFlipPos),
			[]*Card{card1.Card, card2.Card})

		// ゲーム終了判定
		if m.allTaken() {
			m.determineWinner()
			m.phase = MemoryPhaseGameEnd
			m.gameEndFlag = true
			return
		}

		m.phase = MemoryPhaseFlip1
	} else {
		// 不一致 → カードを裏に戻す、次のプレイヤーへ
		m.lastMatchResult = false
		card1.FaceUp = false
		card2.FaceUp = false

		m.appendLog("miss", fmt.Sprintf("不一致: 位置%d,位置%d", m.firstFlipPos, m.secondFlipPos),
			[]*Card{card1.Card, card2.Card})

		m.advancePlayer()
		m.phase = MemoryPhaseFlip1
	}

	m.firstFlipPos = -1
	m.secondFlipPos = -1
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (m *Memory) GetPhase() MemoryPhase { return m.phase }

// SetPhase フェーズ設定 (テスト用)
func (m *Memory) SetPhase(phase MemoryPhase) { m.phase = phase }

// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得
func (m *Memory) GetCurrentPlayerIdx() int { return m.currentPlayerIdx }

// SetCurrentPlayerIdx 現在のプレイヤーインデックスを設定 (テスト用)
func (m *Memory) SetCurrentPlayerIdx(idx int) { m.currentPlayerIdx = idx }

// GetFirstFlipPos 1枚目のめくり位置を取得
func (m *Memory) GetFirstFlipPos() int { return m.firstFlipPos }

// GetSecondFlipPos 2枚目のめくり位置を取得
func (m *Memory) GetSecondFlipPos() int { return m.secondFlipPos }

// GetLastMatchResult 直前のマッチ結果を取得
func (m *Memory) GetLastMatchResult() bool { return m.lastMatchResult }

// GetGameEndFlag ゲーム終了フラグを取得
func (m *Memory) GetGameEndFlag() bool { return m.gameEndFlag }

// GetWinnerIdx 勝者インデックスを取得
func (m *Memory) GetWinnerIdx() int { return m.winnerIdx }

// GetPlayerCnt プレイヤー数を取得
func (m *Memory) GetPlayerCnt() int { return len(m.players) }

// GetPlayer プレイヤーを取得
func (m *Memory) GetPlayer(i int) *MemoryPlayer {
	return getPlayer(m.players, i)
}

// GetBoard ボードカードを取得
func (m *Memory) GetBoard() []*MemoryBoardCard { return m.board }

// GetBoardCard 指定位置のボードカードを取得
func (m *Memory) GetBoardCard(pos int) *MemoryBoardCard {
	if pos < 0 || pos >= len(m.board) {
		return nil
	}
	return m.board[pos]
}

// SetBoard ボード設定 (テスト用)
func (m *Memory) SetBoard(board []*MemoryBoardCard) { m.board = board }

// GetConfig 設定を取得
func (m *Memory) GetConfig() MemoryConfig { return m.config }

// SetConfig 設定を設定
func (m *Memory) SetConfig(cfg MemoryConfig) { m.config = cfg }

// IsHumanTurn 人間のターンかどうか
func (m *Memory) IsHumanTurn() bool {
	return m.players[m.currentPlayerIdx].GetIsHuman()
}

// GetTurnNumber ターン番号取得
func (m *Memory) GetTurnNumber() int { return m.turnNumber }

// GetActionLog 棋譜取得
func (m *Memory) GetActionLog() []*ActionLogEntry { return m.actionLog }

// --- Private helpers ---

// flip 1枚めくる
func (m *Memory) flip(pos int) error {
	if pos < 0 || pos >= len(m.board) {
		return errors.New("invalid position")
	}
	bc := m.board[pos]
	if bc.Taken {
		return errors.New("card already taken")
	}
	if bc.FaceUp {
		return errors.New("card already face up")
	}

	bc.FaceUp = true
	bc.Visited = true
	retentionChance := m.retentionChance()

	// 全CPUプレイヤーに公開されたカードを記憶させる
	for _, p := range m.players {
		if !p.GetIsHuman() {
			p.RecordRevealedCard(pos, bc.Card.GetValue(), retentionChance, m.turnNumber)
		}
	}

	switch m.phase {
	case MemoryPhaseFlip1:
		m.firstFlipPos = pos
		m.phase = MemoryPhaseFlip2
	case MemoryPhaseFlip2:
		m.secondFlipPos = pos
		m.phase = MemoryPhaseResult
		m.turnNumber++

		// CPUの記憶を減衰
		decayRate := m.decayRate()
		for _, p := range m.players {
			if !p.GetIsHuman() {
				p.DecayMemories(m.turnNumber, decayRate)
			}
		}
	}
	return nil
}

// retentionChance 難易度に応じた記憶確率を返す
func (m *Memory) retentionChance() float64 {
	switch m.config.CpuDifficulty {
	case MemoryCpuDifficultyEasy:
		return memoryRetentionEasy
	case MemoryCpuDifficultyHard:
		return memoryRetentionHard
	default:
		return memoryRetentionNormal
	}
}

// decayRate 難易度に応じた忘却率を返す
func (m *Memory) decayRate() float64 {
	switch m.config.CpuDifficulty {
	case MemoryCpuDifficultyEasy:
		return memoryDecayEasy
	case MemoryCpuDifficultyHard:
		return memoryDecayHard
	default:
		return memoryDecayNormal
	}
}

// randomAvailablePosition ランダムに利用可能な位置を返す (excludeを除く)
func (m *Memory) randomAvailablePosition(exclude int) int {
	available := []int{}
	for i := 0; i < len(m.board); i++ {
		if i != exclude && !m.board[i].Taken && !m.board[i].FaceUp {
			available = append(available, i)
		}
	}
	if len(available) == 0 {
		return -1
	}
	return available[rand.Intn(len(available))]
}

// advancePlayer 次のプレイヤーへ進む
func (m *Memory) advancePlayer() {
	m.currentPlayerIdx = (m.currentPlayerIdx + 1) % len(m.players)
}

// allTaken 全カードが取られたか判定
func (m *Memory) allTaken() bool {
	for i := 0; i < len(m.board); i++ {
		if !m.board[i].Taken {
			return false
		}
	}
	return true
}

// determineWinner 勝者を判定（ペア数最多のプレイヤー）
func (m *Memory) determineWinner() {
	maxPairs := -1
	winnerIdx := 0
	for i, p := range m.players {
		if p.GetPairCount() > maxPairs {
			maxPairs = p.GetPairCount()
			winnerIdx = i
		}
	}
	m.winnerIdx = winnerIdx
}

// appendLog 棋譜エントリを追加
func (m *Memory) appendLog(actionType, detail string, cards []*Card) {
	m.actionLog = append(m.actionLog, &ActionLogEntry{
		TurnNumber: m.turnNumber,
		PlayerIdx:  m.currentPlayerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// memoryJSON is the JSON wire format for Memory.
type memoryJSON struct {
	TrumpCards       *TrumpCards        `json:"tc"`
	Board            []*MemoryBoardCard `json:"bd"`
	Players          []*MemoryPlayer    `json:"pl"`
	Config           MemoryConfig       `json:"cf"`
	Phase            MemoryPhase        `json:"ps"`
	CurrentPlayerIdx int                `json:"ci"`
	FirstFlipPos     int                `json:"f1"`
	SecondFlipPos    int                `json:"f2"`
	LastMatchResult  bool               `json:"lm"`
	GameEndFlag      bool               `json:"ge"`
	WinnerIdx        int                `json:"wi"`
	TurnNumber       int                `json:"tn"`
	ActionLog        []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (m *Memory) MarshalJSON() ([]byte, error) {
	return json.Marshal(memoryJSON{
		TrumpCards:       m.trumpCards,
		Board:            m.board,
		Players:          m.players,
		Config:           m.config,
		Phase:            m.phase,
		CurrentPlayerIdx: m.currentPlayerIdx,
		FirstFlipPos:     m.firstFlipPos,
		SecondFlipPos:    m.secondFlipPos,
		LastMatchResult:  m.lastMatchResult,
		GameEndFlag:      m.gameEndFlag,
		WinnerIdx:        m.winnerIdx,
		TurnNumber:       m.turnNumber,
		ActionLog:        m.actionLog,
	})
}

// memoryMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const memoryMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (m *Memory) UnmarshalJSON(data []byte) error {
	var j memoryJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > memoryMaxSliceLen || len(j.ActionLog) > memoryMaxSliceLen {
		return fmt.Errorf("memory: input array exceeds maximum allowed size")
	}

	m.trumpCards = j.TrumpCards
	if m.trumpCards == nil {
		m.trumpCards = NewTrumpCards(0)
	}
	m.board = j.Board
	m.players = j.Players
	if m.players == nil {
		m.players = make([]*MemoryPlayer, 0)
	}
	m.config = j.Config
	m.phase = j.Phase
	m.currentPlayerIdx = j.CurrentPlayerIdx
	m.firstFlipPos = j.FirstFlipPos
	m.secondFlipPos = j.SecondFlipPos
	m.lastMatchResult = j.LastMatchResult
	m.gameEndFlag = j.GameEndFlag
	m.winnerIdx = j.WinnerIdx
	m.turnNumber = j.TurnNumber
	m.actionLog = j.ActionLog
	if m.actionLog == nil {
		m.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
