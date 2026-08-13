//go:build !js || !wasm || casino

package domain

import (
	"errors"
	"fmt"
)

// HorsePhase はオーケストレータの進行段階。
//
// **種目の中の進行は各実装が持っている。** ここが持つのは「ハンドが動いて
// いるか、終わって次を待っているか」だけ。
type HorsePhase int

const (
	// HorsePhaseHand はハンドが進行中。
	HorsePhaseHand HorsePhase = iota
	// HorsePhaseHandEnd はハンドが決着して次を待っている。
	HorsePhaseHandEnd
	// HorsePhaseGameEnd は 1 人を残して全員のチップが尽きた。
	HorsePhaseGameEnd
)

// HorsePhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const HorsePhaseMax = HorsePhaseGameEnd

// horseMaxSliceLen は復元時に許すスライス長の上限。
const horseMaxSliceLen = 512

// エラー値。
var (
	errHorseFinished   = errors.New("horse: game already finished")
	errHorseWrongPhase = errors.New("horse: not allowed in this phase")
	errHorseNoTable    = errors.New("horse: no hand in progress")
)

// horseTable は H.O.R.S.E. が種目に対して必要とする操作だけを切り出したもの。
//
// **メソッドの形は同じでも、フェーズ番号は同じではない。** Holdem / Omaha /
// SevenCardStud はどれも `Reset() error` と
// `PlayerAction(action, amount, humanPlayMs int) error` を同じ並びで持つので、
// 呼び出しは 1 つのインタフェースにまとめられる。しかし**フェーズ定数を共有
// しているのは Omaha だけ**で、Stud は独自の並びを持つ:
//
//	Holdem/Omaha : ... Showdown=5, End=6, Rebuy=7
//	SevenCardStud: ... Showdown=6, End=7, Rebuy=8   (ストリートが 1 つ多い)
//
// 素直に `HoldemPhaseEnd` と比べると、**Stud のショーダウン中を「終わった」と
// 誤読して精算前の残高を回収する** ── 実測でチップが増えた。番号を外から
// 決めつけず、終端は種目ごとに持たせる。
type horseTable interface {
	Reset() error
	PlayerAction(action, amount, humanPlayMs int) error
	GetPhase() int
	GetGameEndFlag() bool
	GetCurrentTurn() int
	GetPot() int
}

// horseEndPhase は種目ごとの「ハンドが終わった」フェーズ番号。
type horseEndPhase struct {
	end   int
	rebuy int
}

// horseSeat は 1 席の通算成績。**チップは種目をまたいで持ち回る。**
type horseSeat struct {
	chips   int
	isHuman bool
	name    string
}

// Horse は H.O.R.S.E. の卓 (5 種目のオーケストレータ)。
//
// **ゲームの規則は 1 つも書いていない。** ホールデムもレーズも既存の実装が
// そのまま担当し、ここがやるのは 3 つだけ:
//
//  1. H-O-R-S-E の順に種目を回す
//  2. 種目をまたいでチップを持ち回す
//  3. 1 人を残して全員が飛んだら終える
//
// **チップの持ち回しが唯一の難所。** 各種目のプレイヤー型は別物なので、
// ハンドの開始時に「正本の残高」を配り、終了時に回収する。正本はここが持ち、
// 種目側の残高は 1 ハンドのあいだだけ有効な写しである。
type Horse struct {
	config HorseConfig
	seats  []*horseSeat

	phase HorsePhase
	// discipline はいまの種目。**H-O-R-S-E の順に進む。**
	discipline HorseDiscipline
	// handInDiscipline はいまの種目で何ハンド目か (1 始まり)。
	handInDiscipline int
	// handNumber は通算ハンド数。
	handNumber int

	// table はいまのハンドを進めている種目の実装。
	table horseTable
	// endPhase はいまの種目の終端フェーズ番号 (種目ごとに違う)。
	endPhase horseEndPhase
	// seatMap は種目側の席番号 → 正本の席番号。
	//
	// **飛んだ席は座らせないので、番号が詰まる。** 種目側の 0 番が正本の 0 番とは
	// 限らない。
	seatMap []int
	// harvest はハンド終了時に種目側の残高を正本へ戻す。
	harvest func()

	gameEndFlag bool
	actionLog   []*ActionLogEntry
	turnNumber  int
}

// NewHorse は指定の設定で卓を構築する。席 0 が人間。
func NewHorse(config HorseConfig) *Horse {
	seats := make([]*horseSeat, 0, config.Seats)
	for i := range config.Seats {
		name := fmt.Sprintf("CPU%d", i)
		if i == 0 {
			name = "YOU"
		}
		seats = append(seats, &horseSeat{chips: config.InitialChips, isHuman: i == 0, name: name})
	}
	return &Horse{config: config, seats: seats, handInDiscipline: 1, handNumber: 1}
}

// NewDefaultHorse は既定の卓を構築する。
func NewDefaultHorse() *Horse { return NewHorse(DefaultHorseConfig()) }

// Reset はゲームを初期化する。
func (g *Horse) Reset() {
	for i, s := range g.seats {
		s.chips = g.config.InitialChips
		s.isHuman = i == 0
	}
	g.phase = HorsePhaseHand
	g.discipline = HorseHoldem
	g.handInDiscipline = 1
	g.handNumber = 1
	g.gameEndFlag = false
	g.actionLog = nil
	g.turnNumber = 0
	g.appendLog("reset", "game reset")
	g.startHand()
}

// --- 種目の切り替え ---

// startHand はいまの種目で 1 ハンド始める。
func (g *Horse) startHand() {
	g.table, g.endPhase, g.harvest = g.buildTable()
	if g.table == nil {
		g.finish()
		return
	}
	if err := g.table.Reset(); err != nil {
		// **種目が配れないならその卓は畳む。** 席が足りない等で起きうる。
		g.finish()
		return
	}
	g.phase = HorsePhaseHand
	g.appendLog("hand", fmt.Sprintf("%s hand %d (%s)",
		HorseDisciplineLetter(g.discipline), g.handNumber, HorseDisciplineName(g.discipline)))
}

// buildTable はいまの種目の卓と、その残高を回収する関数を返す。
//
// **正本の残高を配ってから作る。** 種目側の残高は 1 ハンドのあいだだけ有効な
// 写しで、回収するまで正本は動かない。
func (g *Horse) buildTable() (horseTable, horseEndPhase, func()) {
	g.seatMap = g.aliveSeatIndexes()
	if len(g.seatMap) < HorseMinSeats {
		return nil, horseEndPhase{}, nil
	}
	n := len(g.seatMap)
	switch g.discipline {
	case HorseHoldem:
		players := NewPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewHoldem(NewTrumpCards(0), players, DefaultHoldemConfig())
		return t, horseEndPhase{end: HoldemPhaseEnd, rebuy: HoldemPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return players[i].GetChips() }) }
	case HorseOmahaHiLo:
		players := NewOmahaPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewOmahaHiLo(NewTrumpCards(0), players, DefaultOmahaConfig())
		return t, horseEndPhase{end: HoldemPhaseEnd, rebuy: HoldemPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return players[i].GetChips() }) }
	case HorseRazz:
		players := NewSevenCardStudPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewRazz(NewTrumpCards(0), players, DefaultRazzConfig())
		return t, horseEndPhase{end: SevenCardStudPhaseEnd, rebuy: SevenCardStudPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return players[i].GetChips() }) }
	case HorseStud:
		players := NewSevenCardStudPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewSevenCardStud(NewTrumpCards(0), players, DefaultSevenCardStudConfig())
		return t, horseEndPhase{end: SevenCardStudPhaseEnd, rebuy: SevenCardStudPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return players[i].GetChips() }) }
	case HorseStudHiLo:
		players := NewSevenCardStudPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewSevenCardStudHiLo(NewTrumpCards(0), players, DefaultSevenCardStudConfig())
		return t, horseEndPhase{end: SevenCardStudPhaseEnd, rebuy: SevenCardStudPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return players[i].GetChips() }) }
	default:
		return nil, horseEndPhase{}, nil
	}
}

// aliveSeatIndexes はチップの残っている席の番号を並べて返す。
//
// **飛んだ席を種目に座らせない。** 各エンジンの `Reset` は残高 0 の席を
// `InitChips` まで**黙って積み直す** (単体のゲームなら卓を続けるための正しい
// 挙動)。ミックスゲームでそれをやると**チップが湧く** ── 実測で総量が
// 3000 → 3114 に増えた。座らせなければ積み直されない。
func (g *Horse) aliveSeatIndexes() []int {
	out := make([]int, 0, len(g.seats))
	for i, s := range g.seats {
		if s.chips > 0 {
			out = append(out, i)
		}
	}
	return out
}

// dealChipsTo は正本の残高を種目側へ配る。
func (g *Horse) dealChipsTo(set func(i, chips int)) {
	for tableIdx, seatIdx := range g.seatMap {
		set(tableIdx, g.seats[seatIdx].chips)
	}
}

// collectChipsFrom は種目側の残高を正本へ戻す。
func (g *Horse) collectChipsFrom(get func(i int) int) {
	for tableIdx, seatIdx := range g.seatMap {
		g.seats[seatIdx].chips = get(tableIdx)
	}
}

// NextHand はハンドを閉じて次へ進む。
//
// **種目が変わるのはここだけ。** ハンドの途中で切り替わることは無い。
func (g *Horse) NextHand() error {
	if g.gameEndFlag {
		return errHorseFinished
	}
	if g.phase != HorsePhaseHandEnd {
		return errHorseWrongPhase
	}
	if g.aliveSeats() < HorseMinSeats {
		g.finish()
		return nil
	}
	g.handNumber++
	if g.handInDiscipline >= g.config.HandsPerDiscipline {
		g.handInDiscipline = 1
		g.discipline = (g.discipline + 1) % HorseDisciplineCount
	} else {
		g.handInDiscipline++
	}
	g.startHand()
	return nil
}

// --- 進行 ---

// PlayerAction は人間の手をいまの種目へ渡す。
//
// **規則の判定はここでは一切しない。** レイズ上限もベット額の刻みも種目ごとに
// 違うので、そのまま渡して種目に決めさせる ── ここで真似ると 5 種目ぶんの規則を
// 二重に持つことになる。
func (g *Horse) PlayerAction(action, amount, humanPlayMs int) error {
	if g.gameEndFlag {
		return errHorseFinished
	}
	if g.phase != HorsePhaseHand {
		return errHorseWrongPhase
	}
	if g.table == nil {
		return errHorseNoTable
	}
	if err := g.table.PlayerAction(action, amount, humanPlayMs); err != nil {
		return err
	}
	g.settleIfHandOver()
	return nil
}

// settleIfHandOver は種目のハンドが終わっていれば残高を回収する。
//
// **回収はここ 1 か所。** 種目側の残高を正本に戻す経路を増やすと、二重に
// 回収して増える経路ができる。
func (g *Horse) settleIfHandOver() {
	if g.table == nil || !g.tableHandIsOver() {
		return
	}
	if g.harvest != nil {
		g.harvest()
	}
	g.phase = HorsePhaseHandEnd
	g.appendLog("handEnd", fmt.Sprintf("hand %d settled", g.handNumber))
	if g.aliveSeats() < HorseMinSeats {
		g.finish()
	}
}

// tableHandIsOver は種目がもう手を受け付けないかを返す。
//
// **番号は種目ごとに違う** ので `g.endPhase` を見る。ここを `HoldemPhaseEnd`
// 固定で書いて 2 つ踏んだ:
//
//   - Stud のリバイ待ちを終端と認めず、ハンドが閉じないまま次の手が
//     「Game already ended」で拒まれ続けた (盤面は動かずエラーだけ返る)
//   - 逆に Holdem の番号で Stud を読むと、**ショーダウン中を終端と誤読して
//     精算前の残高を回収し、チップが増えた**
//
// 終了フラグも併せて見る ── 番号を 1 つでも取りこぼすと固まる側に倒れる。
func (g *Horse) tableHandIsOver() bool {
	switch g.table.GetPhase() {
	case g.endPhase.end, g.endPhase.rebuy:
		return true
	}
	return g.table.GetGameEndFlag()
}

// aliveSeats はまだチップが残っている席の数を返す。
func (g *Horse) aliveSeats() int {
	n := 0
	for _, s := range g.seats {
		if s.chips > 0 {
			n++
		}
	}
	return n
}

// finish はゲームを終える。
func (g *Horse) finish() {
	g.gameEndFlag = true
	g.phase = HorsePhaseGameEnd
	g.appendLog("gameEnd", fmt.Sprintf("winner seat %d", g.WinnerSeat()))
}

// WinnerSeat はチップがいちばん多い席を返す。同点なら若い席。
func (g *Horse) WinnerSeat() int {
	best, bestChips := 0, -1
	for i, s := range g.seats {
		if s.chips > bestChips {
			best, bestChips = i, s.chips
		}
	}
	return best
}

// --- 参照 ---

// GetConfig はゲーム設定を返す。
func (g *Horse) GetConfig() HorseConfig { return g.config }

// SetConfig はゲーム設定を設定する。
func (g *Horse) SetConfig(c HorseConfig) { g.config = c }

// GetPhase は現在のフェーズを返す。
func (g *Horse) GetPhase() HorsePhase { return g.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Horse) GetGameEndFlag() bool { return g.gameEndFlag }

// GetDiscipline はいまの種目を返す。
func (g *Horse) GetDiscipline() HorseDiscipline { return g.discipline }

// GetDisciplineLetter はいまの種目の頭文字を返す。
func (g *Horse) GetDisciplineLetter() string { return HorseDisciplineLetter(g.discipline) }

// GetHandInDiscipline はいまの種目で何ハンド目かを返す。
func (g *Horse) GetHandInDiscipline() int { return g.handInDiscipline }

// GetHandNumber は通算ハンド数を返す。
func (g *Horse) GetHandNumber() int { return g.handNumber }

// GetSeatChips は席のチップ数を返す。
func (g *Horse) GetSeatChips(i int) int {
	if i < 0 || i >= len(g.seats) {
		return 0
	}
	return g.seats[i].chips
}

// SetSeatChips は席のチップ数を設定する。
func (g *Horse) SetSeatChips(i, chips int) {
	if i < 0 || i >= len(g.seats) {
		return
	}
	g.seats[i].chips = chips
}

// GetSeatName は席の表示名を返す。
func (g *Horse) GetSeatName(i int) string {
	if i < 0 || i >= len(g.seats) {
		return "?"
	}
	return g.seats[i].name
}

// GetSeatIsHuman は人間の席かを返す。
func (g *Horse) GetSeatIsHuman(i int) bool {
	if i < 0 || i >= len(g.seats) {
		return false
	}
	return g.seats[i].isHuman
}

// GetSeatCount は席数を返す。
func (g *Horse) GetSeatCount() int { return len(g.seats) }

// GetHumanSeat は人間の席を返す。
func (g *Horse) GetHumanSeat() int {
	for i, s := range g.seats {
		if s.isHuman {
			return i
		}
	}
	return 0
}

// GetCurrentTurn はいまの手番を**正本の席番号**で返す。ハンドが無ければ -1。
//
// **種目側の番号をそのまま出さない。** 飛んだ席は座らせないので番号が詰まって
// いて、種目の 0 番が正本の 0 番とは限らない。
func (g *Horse) GetCurrentTurn() int {
	if g.table == nil {
		return -1
	}
	return g.toCanonicalSeat(g.table.GetCurrentTurn())
}

// toCanonicalSeat は種目側の席番号を正本の席番号に直す。
func (g *Horse) toCanonicalSeat(tableIdx int) int {
	if tableIdx < 0 || tableIdx >= len(g.seatMap) {
		return -1
	}
	return g.seatMap[tableIdx]
}

// IsHumanTurn は人間の操作待ちかを返す。
func (g *Horse) IsHumanTurn() bool {
	return g.phase == HorsePhaseHand && g.table != nil && g.GetCurrentTurn() == g.GetHumanSeat()
}

// GetPot はいまの種目のポットを返す。
func (g *Horse) GetPot() int {
	if g.table == nil {
		return 0
	}
	return g.table.GetPot()
}

// GetTablePhase はいまの種目の内部フェーズを返す (Holdem のフェーズ定数)。
func (g *Horse) GetTablePhase() int {
	if g.table == nil {
		return HoldemPhaseInit
	}
	return g.table.GetPhase()
}

// GetActionLog は棋譜を返す。
func (g *Horse) GetActionLog() []*ActionLogEntry { return g.actionLog }

// appendLog は棋譜に 1 行足す。
func (g *Horse) appendLog(actionType, detail string) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  -1,
		ActionType: actionType,
		Detail:     detail,
	})
	if len(g.actionLog) > horseMaxSliceLen {
		g.actionLog = g.actionLog[len(g.actionLog)-horseMaxSliceLen:]
	}
}
