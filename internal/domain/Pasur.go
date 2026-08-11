//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// PasurPhase はパスールのフェーズ。
type PasurPhase int

const (
	// PasurPhasePlay は札を出すフェーズ。
	PasurPhasePlay PasurPhase = iota
	// PasurPhaseGameEnd は終局。
	PasurPhaseGameEnd
)

// PasurHandSize は 1 パックで各プレイヤーに配る枚数。
const PasurHandSize = 4

// PasurInitialTableSize はゲーム開始時に場へ置く枚数。
const PasurInitialTableSize = 4

// PasurCaptureSum は捕獲の合計値。**手札 1 枚 + 場の数札の合計がこれ。**
//
// **手札 1 枚だけでは絶対に取れない。** 数札は A=1..10 なので 11 ちょうどの札は
// 無く、どの捕獲にも必ず場の札が要ります。
const PasurCaptureSum = 11

// pasurMaxSliceLen は復元時に受け付けるスライスの上限。
const pasurMaxSliceLen = 1000

// 得点定数。**issue に得点表が無いので、全ケースが 1 つの式で一貫するように決めています。**
const (
	// PasurScoreClub はクラブ 1 枚あたりの点。
	PasurScoreClub = 1
	// PasurScoreTwoClubsExtra は 2♣ の上乗せ（クラブ分と合わせて 2 点）。
	PasurScoreTwoClubsExtra = 1
	// PasurScoreTenDiamonds は 10♦ の点。
	PasurScoreTenDiamonds = 3
	// PasurScoreAce はエース 1 枚あたりの点。
	PasurScoreAce = 1
	// PasurTotalScore は 1 デッキから出る得点の総和（スール倍化を除く）。
	//
	// クラブ 13 + 2♣ の上乗せ 1 + 10♦ 3 + エース 4 = 21。**毎ゲームちょうど
	// この点が分配される**ので、テストで固定できます。
	PasurTotalScore = 13*PasurScoreClub + PasurScoreTwoClubsExtra + PasurScoreTenDiamonds + 4*PasurScoreAce
	// PasurSoorMultiplier はスールで取った札にかかる倍率。
	PasurSoorMultiplier = 2
)

// pasurIsFace は絵札（J/Q/K）かを返す。**絵札は数値の合計に使いません。**
func pasurIsFace(c *Card) bool { return c.GetValue() >= 11 }

// pasurCardScore は 1 枚の得点を返す。
func pasurCardScore(c *Card) int {
	score := 0
	if c.GetDesign() == CardDesignClover {
		score += PasurScoreClub
		if c.GetValue() == 2 {
			score += PasurScoreTwoClubsExtra
		}
	}
	if c.GetDesign() == CardDesignDiamond && c.GetValue() == 10 {
		score += PasurScoreTenDiamonds
	}
	if c.GetValue() == 1 {
		score += PasurScoreAce
	}
	return score
}

// PasurHint はパスールの助言。
type PasurHint struct {
	// CardIndex は出すべき手札。
	CardIndex *int
	// TableIndices は取るべき場札（トレールなら空）。
	TableIndices []int
	Reason       string
}

// Pasur はパスールのゲーム。
type Pasur struct {
	trumpCards *TrumpCards
	players    []*PasurPlayer
	config     PasurConfig
	phase      PasurPhase
	actionLogBase

	// tableCards は場の札。
	tableCards []*Card
	// packsDealt はこれまでに配ったパック数。
	packsDealt int
	// lastCaptureIdx は最後に捕獲した席（-1 = まだ誰も取っていない）。
	//
	// **場に残った札はここへ行く。** 取り手がいないと札が消える。
	lastCaptureIdx   int
	currentPlayerIdx int

	gameEndFlag bool
	// winners は最高得点の席（同点なら複数）。
	winners []int
	// scores は確定した得点（終局まで nil）。
	scores []int
}

// NewPasur はコンストラクタ。
func NewPasur(players []*PasurPlayer, config PasurConfig) *Pasur {
	if config.Validate() != nil {
		config = DefaultPasurConfig()
	}
	if len(players) != config.PlayerCnt {
		players = newPasurSeats(config.PlayerCnt)
	}
	return &Pasur{players: players, config: config, lastCaptureIdx: -1}
}

// newPasurSeats は標準の席（人間 1 + CPU）を返す。
func newPasurSeats(n int) []*PasurPlayer {
	seats := make([]*PasurPlayer, 0, n)
	for i := range n {
		seats = append(seats, NewPasurPlayer(i == 0))
	}
	return seats
}

// NewDefaultPasur は標準セットアップを返す。
func NewDefaultPasur() *Pasur {
	cfg := DefaultPasurConfig()
	return NewPasur(newPasurSeats(cfg.PlayerCnt), cfg)
}

// Reset はゲームを初期化する。
func (p *Pasur) Reset() {
	for _, pl := range p.players {
		pl.ResetGame()
	}
	p.trumpCards = NewTrumpCards(0)
	p.trumpCards.Shuffle()
	p.phase = PasurPhasePlay
	p.packsDealt = 0
	p.lastCaptureIdx = -1
	p.currentPlayerIdx = 0
	p.gameEndFlag = false
	p.winners = nil
	p.scores = nil
	p.actionLog = nil

	// **場の 4 枚を先に引く。** 残り 48 枚が人数で割り切れるのはこのため。
	p.tableCards = nil
	for range PasurInitialTableSize {
		if c := p.trumpCards.DrawCard(); c != nil {
			p.tableCards = append(p.tableCards, c)
		}
	}
	p.dealPack()
	p.addLog(-1, "start", fmt.Sprintf("パスールを開始しました（%d 人）", p.config.PlayerCnt), nil)
}

// dealPack は全員に 1 パック（4 枚）配る。
func (p *Pasur) dealPack() {
	for range PasurHandSize {
		for i := range p.config.PlayerCnt {
			if c := p.trumpCards.DrawCard(); c != nil {
				p.players[i].AddCard(c)
			}
		}
	}
	p.packsDealt++
	p.sortAllHands()
}

// sortAllHands は手札をスート・ランク順に整える。
func (p *Pasur) sortAllHands() {
	for _, pl := range p.players {
		sortPlayerHand(pl, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return ci.GetValue() < cj.GetValue()
		})
	}
}

// GetCaptureOptions は手札 cardIndex で取れる場札の組み合わせを返す。
//
// **絵札は同ランクだけ、数札は合計 11 だけ。** 絵札を数値に混ぜないのがこの
// ゲームの肝で、混ぜると 11 の組み合わせが激増して別のゲームになります。
func (p *Pasur) GetCaptureOptions(playerIdx, cardIndex int) [][]int {
	if playerIdx < 0 || playerIdx >= p.config.PlayerCnt {
		return nil
	}
	pl := p.players[playerIdx]
	if cardIndex < 0 || cardIndex >= pl.GetCardsSize() {
		return nil
	}
	card := pl.GetCard(cardIndex)

	if pasurIsFace(card) {
		// 同ランクの場札をまとめて 1 通り。
		same := make([]int, 0, len(p.tableCards))
		for i, t := range p.tableCards {
			if t.GetValue() == card.GetValue() {
				same = append(same, i)
			}
		}
		if len(same) == 0 {
			return nil
		}
		return [][]int{same}
	}

	need := PasurCaptureSum - card.GetValue()
	// 場の**数札だけ**が対象。
	nums := make([]int, 0, len(p.tableCards))
	for i, t := range p.tableCards {
		if !pasurIsFace(t) {
			nums = append(nums, i)
		}
	}
	var out [][]int
	var walk func(start, remaining int, picked []int)
	walk = func(start, remaining int, picked []int) {
		if remaining == 0 && len(picked) > 0 {
			out = append(out, append([]int{}, picked...))
			return
		}
		for i := start; i < len(nums); i++ {
			v := p.tableCards[nums[i]].GetValue()
			if v > remaining {
				continue
			}
			walk(i+1, remaining-v, append(picked, nums[i]))
		}
	}
	walk(0, need, nil)
	return out
}

// PlayerPlay は人間が札を出す。tableIndices が空ならトレール（場に置く）。
func (p *Pasur) PlayerPlay(cardIndex int, tableIndices []int) error {
	if !p.IsHumanTurn() {
		return ErrNotHumanTurn
	}
	return p.play(p.currentPlayerIdx, cardIndex, tableIndices)
}

// CpuPlay は CPU が 1 枚出す。
func (p *Pasur) CpuPlay() {
	if p.gameEndFlag || p.phase != PasurPhasePlay || p.IsHumanTurn() {
		return
	}
	idx, table := p.chooseCpuMove(p.currentPlayerIdx)
	_ = p.play(p.currentPlayerIdx, idx, table)
}

// IsHumanTurn は人間の手番かを返す。
func (p *Pasur) IsHumanTurn() bool {
	if p.gameEndFlag || p.phase != PasurPhasePlay {
		return false
	}
	return p.players[p.currentPlayerIdx].GetIsHuman()
}

// play は 1 枚出す共通処理。
func (p *Pasur) play(playerIdx, cardIndex int, tableIndices []int) error {
	if p.gameEndFlag {
		return ErrGameEnded
	}
	if p.phase != PasurPhasePlay {
		return ErrWrongPhase
	}
	if playerIdx != p.currentPlayerIdx {
		return ErrNotHumanTurn
	}
	pl := p.players[playerIdx]
	if cardIndex < 0 || cardIndex >= pl.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	if err := p.validateCapture(playerIdx, cardIndex, tableIndices); err != nil {
		return err
	}

	card := pl.RemoveCard(cardIndex)
	if card == nil {
		return NewDomainError(ErrInvalidCard, "カードがありません")
	}

	if len(tableIndices) == 0 {
		// **トレール: 取れないので場に置く。**
		p.tableCards = append(p.tableCards, card)
		p.addLog(playerIdx, "trail", "場に置きました", []*Card{card})
	} else {
		p.capture(playerIdx, card, tableIndices)
	}

	p.advanceTurn()
	return nil
}

// validateCapture は指定の捕獲が合法かを検証する。
//
// **「取れるのにトレールする」も弾く。** 取れる手があるのに場に置けるようにすると、
// 点札を場に流して相手に取らせない戦術が成立してしまいます。
func (p *Pasur) validateCapture(playerIdx, cardIndex int, tableIndices []int) error {
	options := p.GetCaptureOptions(playerIdx, cardIndex)
	if len(tableIndices) == 0 {
		if len(options) > 0 {
			return errors.New("must capture when a capture is available")
		}
		return nil
	}
	if len(options) == 0 {
		return errors.New("no capture is available for that card")
	}
	for _, opt := range options {
		if pasurSameIndexSet(opt, tableIndices) {
			return nil
		}
	}
	return errors.New("that combination does not add up")
}

// pasurSameIndexSet は 2 つのインデックス集合が同じかを返す（順序は問わない）。
func pasurSameIndexSet(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[int]int, len(a))
	for _, v := range a {
		seen[v]++
	}
	for _, v := range b {
		if seen[v] == 0 {
			return false
		}
		seen[v]--
	}
	return true
}

// capture は場札を取る。
func (p *Pasur) capture(playerIdx int, card *Card, tableIndices []int) {
	taken := make([]*Card, 0, len(tableIndices)+1)
	taken = append(taken, card)
	remove := make(map[int]bool, len(tableIndices))
	for _, i := range tableIndices {
		remove[i] = true
	}
	rest := make([]*Card, 0, len(p.tableCards))
	for i, t := range p.tableCards {
		if remove[i] {
			taken = append(taken, t)
		} else {
			rest = append(rest, t)
		}
	}
	p.tableCards = rest
	p.lastCaptureIdx = playerIdx

	// **スールは「取った結果、場が空になった」こと。** 取った枚数ではありません。
	if len(p.tableCards) == 0 {
		p.players[playerIdx].AddSoorCaptured(taken)
		p.addLog(playerIdx, "soor",
			fmt.Sprintf("スール！ %d 枚を取り、場を空にしました（この札は %d 倍）",
				len(taken), PasurSoorMultiplier), taken)
		return
	}
	p.players[playerIdx].AddCaptured(taken)
	p.addLog(playerIdx, "capture", fmt.Sprintf("%d 枚を取りました", len(taken)), taken)
}

// advanceTurn は手番を進め、必要なら配り足し、山札が尽きたら精算する。
func (p *Pasur) advanceTurn() {
	p.currentPlayerIdx = (p.currentPlayerIdx + 1) % p.config.PlayerCnt

	if p.handsEmpty() {
		if p.trumpCards.GetRemainingCount() > 0 {
			p.dealPack()
			return
		}
		p.finishGame()
	}
}

// handsEmpty は全員の手札が尽きたかを返す。
func (p *Pasur) handsEmpty() bool {
	for _, pl := range p.players {
		if pl.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// finishGame は残り札を処理して得点を確定する。
func (p *Pasur) finishGame() {
	// **場に残った札は最後に取った人のもの。** 取り手がいなければ場に残したまま
	// ——どの席にも入れないので、総得点が 21 に満たないことがあります。
	if len(p.tableCards) > 0 && p.lastCaptureIdx >= 0 {
		p.players[p.lastCaptureIdx].AddCaptured(p.tableCards)
		p.addLog(p.lastCaptureIdx, "leftover",
			fmt.Sprintf("場の残り %d 枚を取りました", len(p.tableCards)), p.tableCards)
		p.tableCards = nil
	}

	p.scores = make([]int, p.config.PlayerCnt)
	best := -1
	for i, pl := range p.players {
		p.scores[i] = p.scoreOf(pl)
		if p.scores[i] > best {
			best = p.scores[i]
		}
	}
	p.winners = nil
	for i, s := range p.scores {
		if s == best {
			p.winners = append(p.winners, i)
		}
	}
	p.phase = PasurPhaseGameEnd
	p.gameEndFlag = true
	p.addLog(-1, "result", fmt.Sprintf("最終得点 %v", p.scores), nil)
}

// scoreOf は 1 人の得点を返す。**スールで取った札は倍。**
func (p *Pasur) scoreOf(pl *PasurPlayer) int {
	score := 0
	for _, c := range pl.GetCaptured() {
		score += pasurCardScore(c)
	}
	for _, c := range pl.GetSoorCaptured() {
		score += pasurCardScore(c) * PasurSoorMultiplier
	}
	return score
}

// GiveUp は投了する。
func (p *Pasur) GiveUp() {
	if p.gameEndFlag {
		return
	}
	p.phase = PasurPhaseGameEnd
	p.gameEndFlag = true
	p.scores = make([]int, p.config.PlayerCnt)
	for i, pl := range p.players {
		p.scores[i] = p.scoreOf(pl)
	}
	// 投了した席（0）以外の全員を勝者にする。
	p.winners = nil
	for i := 1; i < p.config.PlayerCnt; i++ {
		p.winners = append(p.winners, i)
	}
	p.addLog(0, "giveup", "投了しました", nil)
}

// chooseCpuMove は CPU の手を返す。
//
// **点になる札を優先して取る。** 取れる手が無ければいちばん点にならない札を置く。
func (p *Pasur) chooseCpuMove(playerIdx int) (int, []int) {
	pl := p.players[playerIdx]
	bestCard, bestTable, bestValue := -1, []int(nil), -1
	for i := range pl.GetCardsSize() {
		for _, opt := range p.GetCaptureOptions(playerIdx, i) {
			value := pasurCardScore(pl.GetCard(i))
			for _, ti := range opt {
				value += pasurCardScore(p.tableCards[ti])
			}
			// **場を空にできるならスール。** 倍になるので強く優先する。
			if len(opt) == len(p.tableCards) {
				value = value*PasurSoorMultiplier + 1
			}
			if value > bestValue {
				bestCard, bestTable, bestValue = i, opt, value
			}
		}
	}
	if bestCard >= 0 {
		return bestCard, bestTable
	}

	// 取れないので、いちばん点にならない札を置く。
	trail, trailScore := 0, -1
	for i := range pl.GetCardsSize() {
		s := pasurCardScore(pl.GetCard(i))
		if trailScore < 0 || s < trailScore {
			trail, trailScore = i, s
		}
	}
	return trail, nil
}

// GetHint は人間への助言を返す。
func (p *Pasur) GetHint() *PasurHint {
	if p.gameEndFlag || !p.IsHumanTurn() {
		return nil
	}
	if p.players[p.currentPlayerIdx].GetCardsSize() == 0 {
		return nil
	}
	idx, table := p.chooseCpuMove(p.currentPlayerIdx)
	reason := "pasurTrail"
	switch {
	case len(table) > 0 && len(table) == len(p.tableCards):
		reason = "pasurSoor"
	case len(table) > 0:
		reason = "pasurCapture"
	}
	return &PasurHint{CardIndex: &idx, TableIndices: table, Reason: reason}
}

// addLog は棋譜に 1 行足す。
func (p *Pasur) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	p.appendLog(playerIdx, actionType, detail, cards)
}

// --- アクセサ ---------------------------------------------------------------

// GetConfig はゲーム設定を返す。
func (p *Pasur) GetConfig() PasurConfig { return p.config }

// SetConfig はゲーム設定を設定する。
//
// **人数が変わると席も作り直す。** 席数と設定が食い違うと配りが崩れます。
func (p *Pasur) SetConfig(cfg PasurConfig) {
	p.config = cfg
	if len(p.players) != cfg.PlayerCnt {
		p.players = newPasurSeats(cfg.PlayerCnt)
	}
}

// GetPhase は現在のフェーズを返す。
func (p *Pasur) GetPhase() PasurPhase { return p.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (p *Pasur) GetGameEndFlag() bool { return p.gameEndFlag }

// GetTableCards は場の札を返す。
func (p *Pasur) GetTableCards() []*Card { return p.tableCards }

// GetDeckRemaining は山札の残り枚数を返す。
func (p *Pasur) GetDeckRemaining() int {
	if p.trumpCards == nil {
		return 0
	}
	return p.trumpCards.GetRemainingCount()
}

// GetPacksDealt は配ったパック数を返す。
func (p *Pasur) GetPacksDealt() int { return p.packsDealt }

// GetLastCaptureIdx は最後に捕獲した席を返す（-1 = なし）。
func (p *Pasur) GetLastCaptureIdx() int { return p.lastCaptureIdx }

// GetCurrentPlayerIdx は現在の手番を返す。
func (p *Pasur) GetCurrentPlayerIdx() int { return p.currentPlayerIdx }

// GetPlayerCnt はプレイヤー数を返す。
func (p *Pasur) GetPlayerCnt() int { return p.config.PlayerCnt }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (p *Pasur) GetPlayer(i int) *PasurPlayer {
	if i < 0 || i >= len(p.players) {
		return nil
	}
	return p.players[i]
}

// GetScore は席の得点を返す（終局前は現時点の暫定値）。
func (p *Pasur) GetScore(i int) int {
	if i < 0 || i >= len(p.players) {
		return 0
	}
	if p.scores != nil && i < len(p.scores) {
		return p.scores[i]
	}
	return p.scoreOf(p.players[i])
}

// GetWinners は勝った席を返す（同点なら複数、終局前は空）。
func (p *Pasur) GetWinners() []int { return p.winners }

// pasurJSON は KV スナップショットの表現。
type pasurJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*PasurPlayer    `json:"pl"`
	Config           PasurConfig       `json:"cf"`
	Phase            PasurPhase        `json:"ph"`
	TableCards       []*Card           `json:"tb"`
	PacksDealt       int               `json:"pd"`
	LastCaptureIdx   int               `json:"lc"`
	CurrentPlayerIdx int               `json:"ci"`
	GameEndFlag      bool              `json:"ge"`
	Winners          []int             `json:"wn"`
	Scores           []int             `json:"sr"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (p *Pasur) MarshalJSON() ([]byte, error) {
	return json.Marshal(&pasurJSON{
		TrumpCards: p.trumpCards, Players: p.players, Config: p.config, Phase: p.phase,
		TableCards: p.tableCards, PacksDealt: p.packsDealt,
		LastCaptureIdx: p.lastCaptureIdx, CurrentPlayerIdx: p.currentPlayerIdx,
		GameEndFlag: p.gameEndFlag, Winners: p.winners, Scores: p.scores,
		ActionLog: p.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元
//
// **7 PR 連続で「個々のフィールドは範囲内だが組み合わせがあり得ない」を通していた**
// ので、フェーズ × 各フィールドの表として書いています (#5302〜#5313)。とくに
// **まとめて立つフィールドの対**（終了フラグ・フェーズ・得点・勝者）を等値で見ます。
func (p *Pasur) UnmarshalJSON(data []byte) error {
	var j pasurJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < PasurPhasePlay || j.Phase > PasurPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	// **終了フラグとフェーズは対（#5313 で踏んだ形）。** 片方だけ立つと
	// すべての入口が早期 return する一方でフェーズが進まず、投了でも戻れません。
	if j.GameEndFlag != (j.Phase == PasurPhaseGameEnd) {
		return fmt.Errorf("game end flag %v disagrees with phase %d", j.GameEndFlag, j.Phase)
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= j.Config.PlayerCnt {
		return fmt.Errorf("invalid current player: %d", j.CurrentPlayerIdx)
	}
	if j.LastCaptureIdx < -1 || j.LastCaptureIdx >= j.Config.PlayerCnt {
		return fmt.Errorf("invalid last capture: %d", j.LastCaptureIdx)
	}
	if j.PacksDealt < 0 {
		return fmt.Errorf("invalid packs dealt: %d", j.PacksDealt)
	}
	if len(j.TableCards) > pasurMaxSliceLen {
		return errors.New("pasur: input array exceeds maximum allowed size")
	}
	// **枚数だけでなく中身も見る (#5310 の再発防止)。**
	for _, c := range j.TableCards {
		if c == nil {
			return errors.New("invalid table card")
		}
	}
	if len(j.ActionLog) > pasurMaxSliceLen {
		return errors.New("pasur: input array exceeds maximum allowed size")
	}
	// **得点と勝者は終局とセット。** 途中で立っていたら壊れています。
	if !j.GameEndFlag && (len(j.Scores) > 0 || len(j.Winners) > 0) {
		return errors.New("scores or winners before the game ended")
	}
	if j.GameEndFlag {
		if len(j.Scores) != j.Config.PlayerCnt {
			return fmt.Errorf("scores has %d entries for %d players", len(j.Scores), j.Config.PlayerCnt)
		}
		if len(j.Winners) == 0 || len(j.Winners) > j.Config.PlayerCnt {
			return fmt.Errorf("invalid winners: %v", j.Winners)
		}
		for _, w := range j.Winners {
			if w < 0 || w >= j.Config.PlayerCnt {
				return fmt.Errorf("invalid winner: %d", w)
			}
		}
	}
	if len(j.Players) != j.Config.PlayerCnt {
		return fmt.Errorf("players has %d entries for %d seats", len(j.Players), j.Config.PlayerCnt)
	}
	for _, pl := range j.Players {
		if pl == nil {
			return errors.New("nil player")
		}
	}

	if j.TrumpCards != nil {
		p.trumpCards = j.TrumpCards
	}
	p.players, p.config, p.phase = j.Players, j.Config, j.Phase
	p.tableCards, p.packsDealt = j.TableCards, j.PacksDealt
	p.lastCaptureIdx, p.currentPlayerIdx = j.LastCaptureIdx, j.CurrentPlayerIdx
	p.gameEndFlag, p.winners, p.scores = j.GameEndFlag, j.Winners, j.Scores
	p.actionLog = j.ActionLog
	return nil
}
