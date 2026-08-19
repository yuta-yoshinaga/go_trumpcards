//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// StealingBundlesPhase はゲームフェーズ。
type StealingBundlesPhase int

// スティーリングバンドルのフェーズ定数
const (
	// StealingBundlesPhasePlay プレイ中
	StealingBundlesPhasePlay StealingBundlesPhase = 0
	// StealingBundlesPhaseGameEnd ゲーム終了
	StealingBundlesPhaseGameEnd StealingBundlesPhase = 1
)

// StealingBundlesPhaseMin / Max はフェーズ列挙の範囲 (復元時の検証用)。
const (
	StealingBundlesPhaseMin = StealingBundlesPhasePlay
	StealingBundlesPhaseMax = StealingBundlesPhaseGameEnd
)

// stealingBundlesMaxSliceLen は復元時に受け付けるスライス長の上限。
const stealingBundlesMaxSliceLen = 2000

// 直前の獲得の種別。ロケール非依存のキーで、画面と CUI が印を出し分けます。
const (
	// StealingBundlesCaptureTake は場から合計を揃えて取った獲得。
	StealingBundlesCaptureTake = "take"
	// StealingBundlesCaptureSteal は相手の束を丸ごと奪った獲得。
	StealingBundlesCaptureSteal = "steal"
)

// StealingBundlesHint は人間への助言。
type StealingBundlesHint struct {
	// CardIndex は出すべき手札の位置。
	CardIndex *int
	// VictimIdx は略奪する相手の席 (-1 = 略奪ではない)。
	VictimIdx int
	// Reason は助言の理由。
	Reason string
}

// StealingBundles はスティーリングバンドル (Stealing Bundles) のゲームクラス。
//
// アメリカの子供向けフィッシングゲーム。場の同じランクの札を取るのが基本ですが、
// **相手が獲得済みの束は「一番上のランク」が弱点**で、そこに同じランクを出すと
// **束ごと丸ごと奪えます**。得点済みの資産を直接攻撃できる点が、場の札だけを
// 相手にする [Cassino] や [Basra] との違いです。
//
// # 取れるときは取る
//
// **取れる手があるのに場に置くことはできません。** 場から取るか、束を奪うか、
// どちらを選ぶかは自由ですが、どちらもできないときだけ場に置けます。issue の
// 仕様案は「取れなければ場に置く」の *あと* に略奪を書いていますが、略奪は
// 捨て札のあとに起きる別の処理ではなく、**取り方の選択肢のひとつ**です。
//
// # 配りと終わり
//
// 場に 4 枚公開し、各自に 4 枚ずつ配ります。全員の手札が尽きたら山札から
// 配り直し、山札が尽きたら終わりです。**枚数はきれいに割り切れます**——
// 2 人なら 52-4 = 48 枚を 8 枚ずつ 6 回、3 人なら 36 枚を 12 枚ずつ 3 回、
// 4 人なら 32 枚を 16 枚ずつ 2 回。
//
// 最後に場へ残った札は、**最後に取った人**が受け取ります。獲得枚数がいちばん
// 多い人の勝ちです。
type StealingBundles struct {
	trumpCards *TrumpCards
	players    []*StealingBundlesPlayer
	config     StealingBundlesConfig
	phase      StealingBundlesPhase

	tableCards []*Card
	// lastCaptureIdx は最後に取った席 (-1 = まだ誰も取っていない)。
	//
	// **場に残った札の行き先。** 盤面には痕跡が残らないので明示的に持ちます。
	lastCaptureIdx int
	// lastCaptureKind は直前の獲得が場からの「取り」("take") か、相手の束を
	// 丸ごと奪った「盗み」("steal") か。まだ誰も取っていなければ空。
	//
	// **盗みは相手の束を直接減らす。** 場から取るのとは重さが違うので、席に
	// 出す印も分けます (#5767)。
	lastCaptureKind string
	// lastCaptureVictimIdx は盗みの被害者の席 (盗み以外は -1)。
	lastCaptureVictimIdx int
	currentPlayerIdx     int
	packsDealt           int
	turnNumber           int
	gameEndFlag          bool
	winnerIdx            int
	actionLogBase
}

// NewStealingBundles はコンストラクタ。
func NewStealingBundles(players []*StealingBundlesPlayer, config StealingBundlesConfig) *StealingBundles {
	if config.PlayerCnt < StealingBundlesPlayerCntMin || config.PlayerCnt > StealingBundlesPlayerCntMax {
		config.PlayerCnt = StealingBundlesDefaultPlayerCnt
	}
	if players == nil {
		players = newStealingBundlesSeats(config.PlayerCnt)
	}
	return &StealingBundles{
		players:        players,
		config:         config,
		lastCaptureIdx: -1,
		winnerIdx:      -1,
	}
}

// newStealingBundlesSeats は席 0 を人間、以降を CPU とした座席を作る。
func newStealingBundlesSeats(n int) []*StealingBundlesPlayer {
	seats := make([]*StealingBundlesPlayer, n)
	for i := range seats {
		seats[i] = NewStealingBundlesPlayer(i == 0)
	}
	return seats
}

// NewDefaultStealingBundles は既定設定の StealingBundles を返す。
func NewDefaultStealingBundles() *StealingBundles {
	cfg := DefaultStealingBundlesConfig()
	return NewStealingBundles(newStealingBundlesSeats(cfg.PlayerCnt), cfg)
}

// Reset はゲームを初期化する。
func (s *StealingBundles) Reset() {
	for _, p := range s.players {
		p.ResetGame()
	}
	s.phase = StealingBundlesPhasePlay
	s.tableCards = nil
	s.lastCaptureIdx = -1
	s.lastCaptureKind = ""
	s.lastCaptureVictimIdx = -1
	s.currentPlayerIdx = 0
	s.packsDealt = 0
	s.turnNumber = 0
	s.gameEndFlag = false
	s.winnerIdx = -1
	s.actionLog = nil

	s.trumpCards = NewTrumpCards(0)
	s.trumpCards.Shuffle()

	// **場に 4 枚、各自に 4 枚。**
	for range StealingBundlesTableSize {
		if c := s.trumpCards.DrawCard(); c != nil {
			s.tableCards = append(s.tableCards, c)
		}
	}
	s.dealPack()

	s.addLog(-1, "start", fmt.Sprintf("スティーリングバンドルを開始しました（%d 人）",
		s.config.PlayerCnt), nil)
}

// dealPack は全員へ 1 パック (4 枚) 配る。
func (s *StealingBundles) dealPack() {
	for range StealingBundlesHandSize {
		for _, p := range s.players {
			if c := s.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}
	s.packsDealt++
	s.sortAllHands()
}

// sortAllHands は手札をランク・スート順に整える。
func (s *StealingBundles) sortAllHands() {
	for _, p := range s.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetValue() != cj.GetValue() {
				return ci.GetValue() < cj.GetValue()
			}
			return ci.GetDesign() < cj.GetDesign()
		})
	}
}

// GetTableMatches は手札 cardIndex で取れる場札の位置を返す。
func (s *StealingBundles) GetTableMatches(playerIdx, cardIndex int) []int {
	c := s.handCard(playerIdx, cardIndex)
	if c == nil {
		return nil
	}
	out := make([]int, 0, len(s.tableCards))
	for i, t := range s.tableCards {
		if t.GetValue() == c.GetValue() {
			out = append(out, i)
		}
	}
	return out
}

// GetStealTargets は手札 cardIndex で奪える相手の席を返す。
func (s *StealingBundles) GetStealTargets(playerIdx, cardIndex int) []int {
	c := s.handCard(playerIdx, cardIndex)
	if c == nil {
		return nil
	}
	out := make([]int, 0, len(s.players))
	for i, p := range s.players {
		if i == playerIdx {
			continue
		}
		// **狙えるのは一番上だけ。** 束の中身は関係ありません。
		if top := p.GetBundleTop(); top != nil && top.GetValue() == c.GetValue() {
			out = append(out, i)
		}
	}
	return out
}

// stealingBundlesContains は xs に v が含まれるかを返す。
//
// **他ゲームの同名ヘルパは使えません。** 近いものが LingerLonger にありますが、
// あちらは `extra` タグの中なので、このファイル (`extra3`) から呼ぶとホスト
// ビルドだけ通って Worker のビルドが落ちます。
func stealingBundlesContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// handCard は席 playerIdx の cardIndex 番目を返す (範囲外なら nil)。
func (s *StealingBundles) handCard(playerIdx, cardIndex int) *Card {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return nil
	}
	p := s.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return nil
	}
	return p.GetCard(cardIndex)
}

// CanCapture は席 playerIdx に取れる手があるかを返す。
func (s *StealingBundles) CanCapture(playerIdx int) bool {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return false
	}
	for i := range s.players[playerIdx].GetCardsSize() {
		if len(s.GetTableMatches(playerIdx, i)) > 0 || len(s.GetStealTargets(playerIdx, i)) > 0 {
			return true
		}
	}
	return false
}

// IsHumanTurn は現在の手番が人間かを返す。
func (s *StealingBundles) IsHumanTurn() bool {
	if s.gameEndFlag {
		return false
	}
	return s.players[s.currentPlayerIdx].GetIsHuman()
}

// PlayerTake は人間が場札を取る。
func (s *StealingBundles) PlayerTake(cardIndex int) error {
	if err := s.guardTurn(0); err != nil {
		return err
	}
	return s.take(0, cardIndex)
}

// PlayerSteal は人間が相手の束を奪う。
func (s *StealingBundles) PlayerSteal(cardIndex, victimIdx int) error {
	if err := s.guardTurn(0); err != nil {
		return err
	}
	return s.steal(0, cardIndex, victimIdx)
}

// PlayerTrail は人間が場に置く。
func (s *StealingBundles) PlayerTrail(cardIndex int) error {
	if err := s.guardTurn(0); err != nil {
		return err
	}
	return s.trail(0, cardIndex)
}

// guardTurn は手番と局面を検証する。
func (s *StealingBundles) guardTurn(playerIdx int) error {
	if s.gameEndFlag {
		return errors.New("ゲームは終了しています")
	}
	if s.currentPlayerIdx != playerIdx {
		return fmt.Errorf("いまは席 %d の番ではありません", playerIdx)
	}
	return nil
}

// take は場札を取る。
func (s *StealingBundles) take(playerIdx, cardIndex int) error {
	matches := s.GetTableMatches(playerIdx, cardIndex)
	if len(matches) == 0 {
		if s.handCard(playerIdx, cardIndex) == nil {
			return fmt.Errorf("手札の位置が範囲外です: %d", cardIndex)
		}
		return errors.New("その札で取れる場札はありません")
	}

	p := s.players[playerIdx]
	played := p.RemoveCard(cardIndex)
	taken := make([]*Card, 0, len(matches)+1)
	// 後ろから外して添字のずれを避ける。
	for i := len(matches) - 1; i >= 0; i-- {
		idx := matches[i]
		taken = append(taken, s.tableCards[idx])
		s.tableCards = append(s.tableCards[:idx], s.tableCards[idx+1:]...)
	}
	p.AddToBundle(taken...)
	// **出した札が一番上になる。** 次に狙われるランクはこれです。
	p.AddToBundle(played)
	s.lastCaptureIdx = playerIdx
	s.lastCaptureKind = StealingBundlesCaptureTake
	s.lastCaptureVictimIdx = -1
	s.addLog(playerIdx, "take", fmt.Sprintf("場から %d 枚取りました", len(taken)), taken)
	s.finishTurn()
	return nil
}

// steal は相手の束を丸ごと奪う。
func (s *StealingBundles) steal(playerIdx, cardIndex, victimIdx int) error {
	if victimIdx < 0 || victimIdx >= len(s.players) || victimIdx == playerIdx {
		return fmt.Errorf("奪う相手の席が正しくありません: %d", victimIdx)
	}
	if !stealingBundlesContains(s.GetStealTargets(playerIdx, cardIndex), victimIdx) {
		if s.handCard(playerIdx, cardIndex) == nil {
			return fmt.Errorf("手札の位置が範囲外です: %d", cardIndex)
		}
		return errors.New("その札では相手の束を奪えません")
	}

	p := s.players[playerIdx]
	played := p.RemoveCard(cardIndex)
	stolen := s.players[victimIdx].TakeBundle()
	p.AddToBundle(stolen...)
	p.AddToBundle(played)
	s.lastCaptureIdx = playerIdx
	s.lastCaptureKind = StealingBundlesCaptureSteal
	s.lastCaptureVictimIdx = victimIdx
	s.addLog(playerIdx, "steal", fmt.Sprintf("席 %d の束 %d 枚を奪いました", victimIdx, len(stolen)), stolen)
	s.finishTurn()
	return nil
}

// trail は場に置く。
//
// **取れる手があるときは置けません。**
func (s *StealingBundles) trail(playerIdx, cardIndex int) error {
	c := s.handCard(playerIdx, cardIndex)
	if c == nil {
		return fmt.Errorf("手札の位置が範囲外です: %d", cardIndex)
	}
	if s.CanCapture(playerIdx) {
		return errors.New("取れる手があるときは場に置けません")
	}
	s.players[playerIdx].RemoveCard(cardIndex)
	s.tableCards = append(s.tableCards, c)
	s.addLog(playerIdx, "trail", "場に置きました", []*Card{c})
	s.finishTurn()
	return nil
}

// finishTurn は手番を進め、必要なら配り直しと終局判定を行う。
func (s *StealingBundles) finishTurn() {
	s.turnNumber++
	if s.allHandsEmpty() {
		if s.trumpCards.GetRemainingCount() > 0 {
			s.dealPack()
		} else {
			s.finish()
			return
		}
	}
	s.currentPlayerIdx = (s.currentPlayerIdx + 1) % len(s.players)
}

// allHandsEmpty は全員の手札が尽きたかを返す。
func (s *StealingBundles) allHandsEmpty() bool {
	for _, p := range s.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// finish は場札を清算して終局する。
func (s *StealingBundles) finish() {
	// **場に残った札は最後に取った人のもの。** 誰も取っていなければ場に残します。
	if len(s.tableCards) > 0 && s.lastCaptureIdx >= 0 {
		s.players[s.lastCaptureIdx].AddToBundle(s.tableCards...)
		s.addLog(s.lastCaptureIdx, "sweep",
			fmt.Sprintf("場に残った %d 枚を受け取りました", len(s.tableCards)), s.tableCards)
		s.tableCards = nil
	}

	s.phase = StealingBundlesPhaseGameEnd
	s.gameEndFlag = true
	s.winnerIdx = s.leaderIdx()
	s.addLog(s.winnerIdx, "result",
		fmt.Sprintf("%d 枚でいちばん多く集めました", s.players[s.winnerIdx].GetBundleSize()), nil)
}

// leaderIdx は束がいちばん多い席を返す (同数なら若い席)。
func (s *StealingBundles) leaderIdx() int {
	best := 0
	for i, p := range s.players {
		if p.GetBundleSize() > s.players[best].GetBundleSize() {
			best = i
		}
	}
	return best
}

// GiveUp は投了する。
func (s *StealingBundles) GiveUp() {
	if s.gameEndFlag {
		return
	}
	s.phase = StealingBundlesPhaseGameEnd
	s.gameEndFlag = true
	// 人間以外で束のいちばん多い席を勝ちにする。
	best := -1
	for i := 1; i < len(s.players); i++ {
		if best < 0 || s.players[i].GetBundleSize() > s.players[best].GetBundleSize() {
			best = i
		}
	}
	if best < 0 {
		best = 0
	}
	s.winnerIdx = best
	s.addLog(0, "giveup", "投了しました", nil)
}

// CpuPlay は CPU が 1 手打つ。
func (s *StealingBundles) CpuPlay() {
	if s.gameEndFlag || s.IsHumanTurn() {
		return
	}
	idx := s.currentPlayerIdx
	card, victim := s.chooseCpuMove(idx)
	switch {
	case victim >= 0:
		_ = s.steal(idx, card, victim)
	case len(s.GetTableMatches(idx, card)) > 0:
		_ = s.take(idx, card)
	default:
		_ = s.trail(idx, card)
	}
}

// chooseCpuMove は CPU の手を選ぶ。返り値は (手札の位置, 奪う相手 / -1)。
//
// **いちばん大きい束を狙います。** 束の略奪は取れる枚数がそのまま得点差になる
// ので、場から 1 枚取るより優先します。
func (s *StealingBundles) chooseCpuMove(playerIdx int) (int, int) {
	bestCard, bestVictim, bestGain := -1, -1, 0
	for i := range s.players[playerIdx].GetCardsSize() {
		for _, v := range s.GetStealTargets(playerIdx, i) {
			if n := s.players[v].GetBundleSize(); n > bestGain {
				bestCard, bestVictim, bestGain = i, v, n
			}
		}
	}
	if bestVictim >= 0 {
		return bestCard, bestVictim
	}

	// 次に、場からいちばん多く取れる手。
	for i := range s.players[playerIdx].GetCardsSize() {
		if n := len(s.GetTableMatches(playerIdx, i)); n > bestGain {
			bestCard, bestGain = i, n
		}
	}
	if bestCard >= 0 {
		return bestCard, -1
	}

	// **取れないときは、いちばん狙われにくい札を置く。** 場に同じランクが
	// 増えるほど、次の手番の相手に取られやすくなります。
	pick, fewest := 0, -1
	for i := range s.players[playerIdx].GetCardsSize() {
		c := s.players[playerIdx].GetCard(i)
		n := 0
		for _, t := range s.tableCards {
			if t.GetValue() == c.GetValue() {
				n++
			}
		}
		if fewest < 0 || n < fewest {
			pick, fewest = i, n
		}
	}
	return pick, -1
}

// GetHint は人間への助言を返す。
func (s *StealingBundles) GetHint() *StealingBundlesHint {
	if s.gameEndFlag || !s.IsHumanTurn() {
		return nil
	}
	if s.players[0].GetCardsSize() == 0 {
		return nil
	}
	card, victim := s.chooseCpuMove(0)
	switch {
	case victim >= 0:
		return &StealingBundlesHint{CardIndex: &card, VictimIdx: victim, Reason: "stealingbundlesSteal"}
	case len(s.GetTableMatches(0, card)) > 0:
		return &StealingBundlesHint{CardIndex: &card, VictimIdx: -1, Reason: "stealingbundlesTake"}
	default:
		return &StealingBundlesHint{CardIndex: &card, VictimIdx: -1, Reason: "stealingbundlesTrail"}
	}
}

// addLog は棋譜に 1 行足す。
func (s *StealingBundles) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.appendLog(playerIdx, actionType, detail, cards)
}

// GetConfig は設定を返す。
func (s *StealingBundles) GetConfig() StealingBundlesConfig { return s.config }

// SetConfig は設定を更新する。
func (s *StealingBundles) SetConfig(cfg StealingBundlesConfig) {
	if err := cfg.Validate(); err != nil {
		return
	}
	if cfg.PlayerCnt != s.config.PlayerCnt {
		s.players = newStealingBundlesSeats(cfg.PlayerCnt)
	}
	s.config = cfg
}

// GetPhase は現在のフェーズを返す。
func (s *StealingBundles) GetPhase() StealingBundlesPhase { return s.phase }

// GetGameEndFlag は終局フラグを返す。
func (s *StealingBundles) GetGameEndFlag() bool { return s.gameEndFlag }

// GetTableCards は場札を返す。
func (s *StealingBundles) GetTableCards() []*Card { return s.tableCards }

// GetDeckRemaining は山札の残り枚数を返す。
func (s *StealingBundles) GetDeckRemaining() int {
	if s.trumpCards == nil {
		return 0
	}
	return s.trumpCards.GetRemainingCount()
}

// GetLastCaptureKind は直前の獲得の種別 ("take" / "steal"、まだなら "") を返す。
func (s *StealingBundles) GetLastCaptureKind() string { return s.lastCaptureKind }

// GetLastCaptureVictimIdx は直前の盗みの被害者席を返す。盗み以外は -1。
func (s *StealingBundles) GetLastCaptureVictimIdx() int { return s.lastCaptureVictimIdx }

// GetLastCaptureIdx は最後に取った席を返す (-1 = まだ)。
func (s *StealingBundles) GetLastCaptureIdx() int { return s.lastCaptureIdx }

// GetCurrentPlayerIdx は現在の手番を返す。
func (s *StealingBundles) GetCurrentPlayerIdx() int { return s.currentPlayerIdx }

// GetTurnNumber は打たれた手の数を返す。
func (s *StealingBundles) GetTurnNumber() int { return s.turnNumber }

// GetPacksDealt は配ったパック数を返す。
func (s *StealingBundles) GetPacksDealt() int { return s.packsDealt }

// GetPlayerCnt は人数を返す。
func (s *StealingBundles) GetPlayerCnt() int { return len(s.players) }

// GetPlayer は席 i のプレイヤーを返す。
func (s *StealingBundles) GetPlayer(i int) *StealingBundlesPlayer {
	if i < 0 || i >= len(s.players) {
		return nil
	}
	return s.players[i]
}

// GetWinnerIdx は勝者の席を返す (-1 = 未確定)。
func (s *StealingBundles) GetWinnerIdx() int { return s.winnerIdx }

// stealingBundlesJSON is the JSON wire format for StealingBundles.
type stealingBundlesJSON struct {
	Players              []*StealingBundlesPlayer `json:"pl"`
	Config               StealingBundlesConfig    `json:"cf"`
	Phase                StealingBundlesPhase     `json:"ph"`
	TableCards           []*Card                  `json:"tb"`
	Deck                 *TrumpCards              `json:"dk"`
	LastCaptureIdx       int                      `json:"lc"`
	LastCaptureKind      string                   `json:"lk"`
	LastCaptureVictimIdx int                      `json:"lv"`
	CurrentIdx           int                      `json:"ci"`
	PacksDealt           int                      `json:"pd"`
	TurnNumber           int                      `json:"tn"`
	GameEndFlag          bool                     `json:"ge"`
	WinnerIdx            int                      `json:"wi"`
	ActionLog            []*ActionLogEntry        `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (s *StealingBundles) MarshalJSON() ([]byte, error) {
	return json.Marshal(stealingBundlesJSON{
		Players:              s.players,
		Config:               s.config,
		Phase:                s.phase,
		TableCards:           s.tableCards,
		Deck:                 s.trumpCards,
		LastCaptureIdx:       s.lastCaptureIdx,
		LastCaptureKind:      s.lastCaptureKind,
		LastCaptureVictimIdx: s.lastCaptureVictimIdx,
		CurrentIdx:           s.currentPlayerIdx,
		PacksDealt:           s.packsDealt,
		TurnNumber:           s.turnNumber,
		GameEndFlag:          s.gameEndFlag,
		WinnerIdx:            s.winnerIdx,
		ActionLog:            s.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *StealingBundles) UnmarshalJSON(data []byte) error {
	var j stealingBundlesJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if len(j.Players) != j.Config.PlayerCnt {
		return fmt.Errorf("seat count %d does not match the configured %d", len(j.Players), j.Config.PlayerCnt)
	}
	if j.Phase < StealingBundlesPhaseMin || j.Phase > StealingBundlesPhaseMax {
		return fmt.Errorf("phase out of range: %d", j.Phase)
	}
	if len(j.ActionLog) > stealingBundlesMaxSliceLen {
		return fmt.Errorf("action log too long: %d", len(j.ActionLog))
	}
	if len(j.TableCards) > StealingBundlesDeckSize {
		return fmt.Errorf("the table holds more cards than the deck: %d", len(j.TableCards))
	}
	for i, c := range j.TableCards {
		if c == nil {
			return fmt.Errorf("table card %d is missing", i)
		}
	}
	if j.CurrentIdx < 0 || j.CurrentIdx >= len(j.Players) {
		return fmt.Errorf("current player index out of range: %d", j.CurrentIdx)
	}
	if j.LastCaptureIdx < -1 || j.LastCaptureIdx >= len(j.Players) {
		return fmt.Errorf("last capture index out of range: %d", j.LastCaptureIdx)
	}
	// **種別と席は一緒に決まる。** 盗みなのに被害者がいない (あるいはその逆) の
	// 復元は、印を誤って出すので弾きます。
	switch j.LastCaptureKind {
	case "":
		if j.LastCaptureIdx >= 0 {
			return fmt.Errorf("seat %d is recorded as having captured with no capture kind", j.LastCaptureIdx)
		}
	case StealingBundlesCaptureTake:
		if j.LastCaptureIdx < 0 {
			return fmt.Errorf("capture kind %q recorded with no capturing seat", j.LastCaptureKind)
		}
		if j.LastCaptureVictimIdx != -1 {
			return fmt.Errorf("a take cannot have a victim: %d", j.LastCaptureVictimIdx)
		}
	case StealingBundlesCaptureSteal:
		if j.LastCaptureIdx < 0 {
			return fmt.Errorf("capture kind %q recorded with no capturing seat", j.LastCaptureKind)
		}
		if j.LastCaptureVictimIdx < 0 || j.LastCaptureVictimIdx >= len(j.Players) ||
			j.LastCaptureVictimIdx == j.LastCaptureIdx {
			return fmt.Errorf("steal victim index out of range: %d", j.LastCaptureVictimIdx)
		}
	default:
		return fmt.Errorf("unknown capture kind: %q", j.LastCaptureKind)
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= len(j.Players) {
		return fmt.Errorf("winner index out of range: %d", j.WinnerIdx)
	}
	if j.GameEndFlag != (j.Phase == StealingBundlesPhaseGameEnd) {
		return fmt.Errorf("the game-end flag and the phase disagree (flag=%v, phase=%d)", j.GameEndFlag, j.Phase)
	}
	if j.GameEndFlag != (j.WinnerIdx >= 0) {
		return fmt.Errorf("a finished game has a winner and an unfinished one does not (flag=%v, winner=%d)",
			j.GameEndFlag, j.WinnerIdx)
	}
	if j.PacksDealt < 0 || j.TurnNumber < 0 {
		return fmt.Errorf("counters cannot be negative (packs=%d, turns=%d)", j.PacksDealt, j.TurnNumber)
	}

	// **52 枚は増えも減りもしません。** 手札・場・束・山札の合計が動いたら、
	// どこかが札を作ったか落としたということです。
	total := len(j.TableCards)
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("seat %d is missing", i)
		}
		total += p.GetCardsSize() + p.GetBundleSize()
	}
	if j.Deck != nil {
		total += j.Deck.GetRemainingCount()
	}
	if total != StealingBundlesDeckSize {
		return fmt.Errorf("the cards in play total %d, not %d", total, StealingBundlesDeckSize)
	}
	// **誰かが取ったなら、その痕跡は束に残ります。**
	if j.LastCaptureIdx < 0 {
		for i, p := range j.Players {
			if p.GetBundleSize() > 0 {
				return fmt.Errorf("seat %d holds a bundle but nobody is recorded as having captured", i)
			}
		}
	}

	s.players = j.Players
	s.config = j.Config
	s.phase = j.Phase
	s.tableCards = j.TableCards
	s.trumpCards = j.Deck
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCards(0)
	}
	s.lastCaptureIdx = j.LastCaptureIdx
	s.lastCaptureKind = j.LastCaptureKind
	s.lastCaptureVictimIdx = j.LastCaptureVictimIdx
	s.currentPlayerIdx = j.CurrentIdx
	s.packsDealt = j.PacksDealt
	s.turnNumber = j.TurnNumber
	s.gameEndFlag = j.GameEndFlag
	s.winnerIdx = j.WinnerIdx
	s.actionLog = j.ActionLog
	return nil
}
