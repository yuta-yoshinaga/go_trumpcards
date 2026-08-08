package domain

import (
	"encoding/json"
	"fmt"
)

// WarPlayerCnt 戦争ゲームのプレイヤー数 (人間 + CPU)
const WarPlayerCnt = 2

// WarBurialCount 戦争時に伏せる枚数
const WarBurialCount = 3

// WarPhase 戦争ゲームのフェーズ
type WarPhase int

// 戦争ゲームのフェーズ定数
const (
	// WarPhaseReveal 次のめくりを待つ状態
	WarPhaseReveal WarPhase = 0
	// WarPhaseResolved 直近のラウンドで勝者が決まり、場札を勝者に渡すのを待つ状態
	WarPhaseResolved WarPhase = 1
	// WarPhaseWarBury 戦争発生。次のステップで伏せ札 + 表札を出す
	WarPhaseWarBury WarPhase = 2
	// WarPhaseGameEnd ゲーム終了
	WarPhaseGameEnd WarPhase = 3
)

// warRank 戦争ゲームでの強さランクを返す (エース=14 の high-ace ルール)
func warRank(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v == 1 {
		return 14
	}
	return v
}

// War 戦争ゲームクラス
type War struct {
	trumpCards      *TrumpCards
	players         [WarPlayerCnt]*WarPlayer
	config          WarConfig
	phase           WarPhase
	warPot          []*Card
	playerRevealed  *Card
	cpuRevealed     *Card
	lastWinnerIdx   int
	lastBurialCount int
	roundsPlayed    int
	gameEndFlag     bool
	winnerIdx       int
	actionLogBase
}

// NewWar コンストラクタ
func NewWar(trumpCards *TrumpCards, players []*WarPlayer, config WarConfig) *War {
	w := &War{
		trumpCards:    trumpCards,
		config:        config,
		lastWinnerIdx: -1,
		winnerIdx:     -1,
	}
	for i := 0; i < WarPlayerCnt && i < len(players); i++ {
		w.players[i] = players[i]
	}
	return w
}

// NewDefaultWar returns War with the standard 2-player setup (1 human, 1 CPU)
// and DefaultWarConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultWar() *War {
	players := []*WarPlayer{
		NewWarPlayer(true),
		NewWarPlayer(false),
	}
	return NewWar(NewTrumpCards(0), players, DefaultWarConfig())
}

// Reset ゲームをリセットして新しいゲームを開始する
func (w *War) Reset() {
	w.phase = WarPhaseReveal
	w.warPot = nil
	w.playerRevealed = nil
	w.cpuRevealed = nil
	w.lastWinnerIdx = -1
	w.lastBurialCount = 0
	w.roundsPlayed = 0
	w.gameEndFlag = false
	w.winnerIdx = -1
	w.actionLog = nil

	for _, p := range w.players {
		p.Reset()
		p.ResetPiles()
		p.SetIsFinished(false)
	}

	w.trumpCards.Shuffle()

	cards := make([]*Card, 0, CardCnt)
	for range CardCnt {
		c := w.trumpCards.DrawCard()
		if c != nil {
			cards = append(cards, c)
		}
	}

	half := len(cards) / 2
	for pi := range WarPlayerCnt {
		start := pi * half
		w.players[pi].AddToDrawPile(cards[start : start+half]...)
	}
}

// Step 状態機械を1ステップ進める
func (w *War) Step() error {
	if w.gameEndFlag {
		return ErrGameEnded
	}
	switch w.phase {
	case WarPhaseReveal:
		return w.stepReveal()
	case WarPhaseResolved:
		return w.stepResolved()
	case WarPhaseWarBury:
		return w.stepWarBury()
	}
	return ErrWrongPhase
}

// warAutoPlayMaxSteps caps the AutoPlay loop. With MaxRounds defaulting to 500
// and a generous safety factor for nested wars, this is well above any
// realistic play-out length while preventing runaway iteration. Declared as a
// var (not const) so tests can lower it to exercise the cap-hit branch.
var warAutoPlayMaxSteps = 100000

// AutoPlay 自動プレイ（決着まで Step を繰り返す）
func (w *War) AutoPlay() error {
	if w.gameEndFlag {
		return ErrGameEnded
	}
	for i := range warAutoPlayMaxSteps {
		if w.gameEndFlag {
			return nil
		}
		if err := w.Step(); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	return fmt.Errorf("auto-play reached maximum steps (%d) without finishing", warAutoPlayMaxSteps)
}

// stepReveal 両者の山札から1枚ずつめくる
func (w *War) stepReveal() error {
	w.lastBurialCount = 0
	pc := w.players[0].DrawOne()
	cc := w.players[1].DrawOne()
	if pc == nil || cc == nil {
		if pc != nil {
			w.warPot = append(w.warPot, pc)
		}
		if cc != nil {
			w.warPot = append(w.warPot, cc)
		}
		w.playerRevealed = pc
		w.cpuRevealed = cc
		w.finishByTotal()
		return nil
	}
	w.playerRevealed = pc
	w.cpuRevealed = cc
	w.warPot = append(w.warPot, pc, cc)
	w.appendLog(0, "reveal", "turn up", []*Card{pc, cc})
	w.resolveCompare(pc, cc)
	return nil
}

// stepResolved 場札の勝者への受け渡し・次ラウンド開始
func (w *War) stepResolved() error {
	if w.lastWinnerIdx >= 0 && w.lastWinnerIdx < WarPlayerCnt && len(w.warPot) > 0 {
		w.players[w.lastWinnerIdx].AddToDiscardPile(w.warPot...)
		w.appendLog(w.lastWinnerIdx, "collect", fmt.Sprintf("+%d cards", len(w.warPot)), nil)
	}
	w.warPot = nil
	w.playerRevealed = nil
	w.cpuRevealed = nil
	w.lastBurialCount = 0
	w.roundsPlayed++

	if w.roundsPlayed >= w.config.MaxRounds {
		w.finishByTotal()
		return nil
	}
	if !w.players[0].HasCards() || !w.players[1].HasCards() {
		w.finishByTotal()
		return nil
	}
	w.phase = WarPhaseReveal
	return nil
}

// stepWarBury 伏せ札をプールに追加し、新たな表札を出して比較する
func (w *War) stepWarBury() error {
	// 既存の表札は warPot に既に含まれている。ここでは伏せ札 + 新しい表札を追加する。
	w.playerRevealed = nil
	w.cpuRevealed = nil

	pBurn := w.buryFor(0)
	cBurn := w.buryFor(1)
	w.lastBurialCount = max(pBurn, cBurn)

	pc := w.players[0].DrawOne()
	cc := w.players[1].DrawOne()
	if pc == nil || cc == nil {
		if pc != nil {
			w.warPot = append(w.warPot, pc)
		}
		if cc != nil {
			w.warPot = append(w.warPot, cc)
		}
		w.playerRevealed = pc
		w.cpuRevealed = cc
		w.finishByTotal()
		return nil
	}
	w.playerRevealed = pc
	w.cpuRevealed = cc
	w.warPot = append(w.warPot, pc, cc)
	w.appendLog(0, "war", fmt.Sprintf("buried %d each, turn up", w.lastBurialCount), []*Card{pc, cc})
	w.resolveCompare(pc, cc)
	return nil
}

// buryFor 指定プレイヤーから最大 WarBurialCount 枚を伏せ札として warPot に追加する。
// ただし最後の1枚は表札用に残す。戻り値は実際に伏せた枚数。
func (w *War) buryFor(idx int) int {
	p := w.players[idx]
	target := WarBurialCount
	if p.TotalCards() <= target {
		target = p.TotalCards() - 1
	}
	if target < 0 {
		target = 0
	}
	buried := 0
	for range target {
		c := p.DrawOne()
		if c == nil {
			break
		}
		w.warPot = append(w.warPot, c)
		buried++
	}
	return buried
}

// resolveCompare 表札の比較結果に応じて phase を更新する
func (w *War) resolveCompare(pc, cc *Card) {
	pr := warRank(pc)
	cr := warRank(cc)
	switch {
	case pr > cr:
		w.lastWinnerIdx = 0
		w.phase = WarPhaseResolved
	case pr < cr:
		w.lastWinnerIdx = 1
		w.phase = WarPhaseResolved
	default:
		w.lastWinnerIdx = -1
		w.phase = WarPhaseWarBury
	}
}

// finishByTotal 保有枚数の多い方を勝者としてゲームを終了する (同数なら 0)
func (w *War) finishByTotal() {
	w.gameEndFlag = true
	w.phase = WarPhaseGameEnd
	t0 := w.players[0].TotalCards()
	t1 := w.players[1].TotalCards()
	switch {
	case t0 > t1:
		w.winnerIdx = 0
	case t1 > t0:
		w.winnerIdx = 1
	default:
		w.winnerIdx = 0
	}
	w.players[w.winnerIdx].SetIsFinished(true)
}

// --- Getters ---

// GetPhase フェーズ取得
func (w *War) GetPhase() WarPhase { return w.phase }

// GetGameEndFlag ゲーム終了フラグ取得
func (w *War) GetGameEndFlag() bool { return w.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得
func (w *War) GetWinnerIdx() int { return w.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (w *War) GetPlayerCnt() int { return WarPlayerCnt }

// GetPlayer プレイヤー取得
func (w *War) GetPlayer(i int) *WarPlayer {
	if i < 0 || i >= WarPlayerCnt {
		return nil
	}
	return w.players[i]
}

// GetPlayerRevealed 人間側の表札取得
func (w *War) GetPlayerRevealed() *Card { return w.playerRevealed }

// GetCpuRevealed CPU側の表札取得
func (w *War) GetCpuRevealed() *Card { return w.cpuRevealed }

// GetWarPotSize 場に出ている総枚数
func (w *War) GetWarPotSize() int { return len(w.warPot) }

// GetLastWinnerIdx 直近ラウンドの勝者 (-1 = 未確定)
func (w *War) GetLastWinnerIdx() int { return w.lastWinnerIdx }

// GetLastBurialCount 直近の戦争で伏せた枚数
func (w *War) GetLastBurialCount() int { return w.lastBurialCount }

// GetRoundsPlayed 消化ラウンド数
func (w *War) GetRoundsPlayed() int { return w.roundsPlayed }

// GetConfig 設定取得
func (w *War) GetConfig() WarConfig { return w.config }

// SetConfig 設定更新
func (w *War) SetConfig(cfg WarConfig) { w.config = cfg }

// IsHumanTurn 戦争は常に人間入力待ち
func (w *War) IsHumanTurn() bool { return !w.gameEndFlag }

// --- JSON ---

// warJSON is the JSON wire format for War.
type warJSON struct {
	TrumpCards      *TrumpCards              `json:"tc"`
	Players         [WarPlayerCnt]*WarPlayer `json:"ps"`
	Config          WarConfig                `json:"cf"`
	Phase           WarPhase                 `json:"ph"`
	WarPot          []*Card                  `json:"wp"`
	PlayerRevealed  *Card                    `json:"pr"`
	CpuRevealed     *Card                    `json:"cr"`
	LastWinnerIdx   int                      `json:"lw"`
	LastBurialCount int                      `json:"lb"`
	RoundsPlayed    int                      `json:"rp"`
	GameEndFlag     bool                     `json:"ge"`
	WinnerIdx       int                      `json:"wi"`
	ActionLog       []*ActionLogEntry        `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (w *War) MarshalJSON() ([]byte, error) {
	return json.Marshal(warJSON{
		TrumpCards:      w.trumpCards,
		Players:         w.players,
		Config:          w.config,
		Phase:           w.phase,
		WarPot:          w.warPot,
		PlayerRevealed:  w.playerRevealed,
		CpuRevealed:     w.cpuRevealed,
		LastWinnerIdx:   w.lastWinnerIdx,
		LastBurialCount: w.lastBurialCount,
		RoundsPlayed:    w.roundsPlayed,
		GameEndFlag:     w.gameEndFlag,
		WinnerIdx:       w.winnerIdx,
		ActionLog:       w.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (w *War) UnmarshalJSON(data []byte) error {
	var j warJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	w.trumpCards = j.TrumpCards
	w.players = j.Players
	w.config = j.Config
	w.phase = j.Phase
	w.warPot = j.WarPot
	w.playerRevealed = j.PlayerRevealed
	w.cpuRevealed = j.CpuRevealed
	w.lastWinnerIdx = j.LastWinnerIdx
	w.lastBurialCount = j.LastBurialCount
	w.roundsPlayed = j.RoundsPlayed
	w.gameEndFlag = j.GameEndFlag
	w.winnerIdx = j.WinnerIdx
	w.actionLog = j.ActionLog
	return nil
}
