//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// LingerLongerPhase はリンガーロンガーのフェーズ。
type LingerLongerPhase int

const (
	// LingerLongerPhasePlay はプレイ中。
	LingerLongerPhasePlay LingerLongerPhase = iota
	// LingerLongerPhaseGameEnd は終局。
	LingerLongerPhaseGameEnd
)

// LingerLongerDeckSize は 1 デッキの枚数。
const LingerLongerDeckSize = 52

// lingerLongerMaxSliceLen は復元時に受け付ける棋譜の上限。
//
// **このゲームの手数には上限がある。** 毎トリック、場に人数ぶん出て補充は最大 1 枚
// なので手札の総数は単調に減り、52 枚を配り切ることはできても増えることはありません。
// 1 トリックにつき「プレイ人数 + 解決 + 補充 + 脱落」で高々 `人数 + 3` 行、
// トリック数は多くても 52 なので 52 × 9 + 開始 1 + 結果 1 に十分な余裕を取ります。
const lingerLongerMaxSliceLen = 1000

// LingerLongerHint はリンガーロンガーの助言。
type LingerLongerHint struct {
	CardIndex *int
	Reason    string
}

// lingerLongerRank は札の強さ。**A が最強。**
func lingerLongerRank(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// LingerLonger はリンガーロンガーのゲーム。
//
// **トリックを取っても得点にはならない。** 意味があるのは「山札から 1 枚補充できる」
// ことだけで、**手札が尽きた人から脱落**し、最後まで持ち続けた 1 人が勝ちます。
type LingerLonger struct {
	trumpCards *TrumpCards
	players    []*LingerLongerPlayer
	config     LingerLongerConfig
	phase      LingerLongerPhase
	actionLogBase

	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	trickNumber      int

	// eliminatedCnt は脱落した人数。
	eliminatedCnt int
	// lastDrawIdx は直前に補充した席（-1 = 無し）。表示用。
	lastDrawIdx int
	// discarded は場から抜けた札の枚数。
	//
	// **取ったトリックは得点にならず、手札にも戻りません。** 解決した時点で
	// 場から抜けるので、復元時に「手札 + 場札 + 山札 + 抜けた枚数 = 52」で
	// 突き合わせます——抜けた枚数を数えないと、正しい盤面を拒否します
	// （#5316 で踏んだ形）。
	discarded int

	gameEndFlag bool
	// winnerIdx は勝った席（-1 = 未確定）。
	winnerIdx int
}

// NewLingerLonger はコンストラクタ。
func NewLingerLonger(players []*LingerLongerPlayer, config LingerLongerConfig) *LingerLonger {
	if config.Validate() != nil {
		config = DefaultLingerLongerConfig()
	}
	if len(players) != config.PlayerCnt {
		players = newLingerLongerSeats(config.PlayerCnt)
	}
	return &LingerLonger{players: players, config: config, winnerIdx: -1, lastDrawIdx: -1}
}

// newLingerLongerSeats は標準の席（人間 1 + CPU）を返す。
func newLingerLongerSeats(n int) []*LingerLongerPlayer {
	seats := make([]*LingerLongerPlayer, 0, n)
	for i := range n {
		seats = append(seats, NewLingerLongerPlayer(i == 0))
	}
	return seats
}

// NewDefaultLingerLonger は標準セットアップを返す。
func NewDefaultLingerLonger() *LingerLonger {
	cfg := DefaultLingerLongerConfig()
	return NewLingerLonger(newLingerLongerSeats(cfg.PlayerCnt), cfg)
}

// Reset はゲームを初期化する。
func (l *LingerLonger) Reset() {
	for _, p := range l.players {
		p.ResetGame()
	}
	l.phase = LingerLongerPhasePlay
	l.currentTrick = nil
	l.trickNumber = 0
	l.eliminatedCnt = 0
	l.lastDrawIdx = -1
	l.discarded = 0
	l.gameEndFlag = false
	l.winnerIdx = -1
	l.actionLog = nil

	l.trumpCards = NewTrumpCards(0)
	l.trumpCards.Shuffle()
	// **配る枚数は人数と同じ。** 4 人なら 4 枚ずつ、6 人なら 6 枚ずつで、
	// 残りは山札。人数が増えるほど 1 人あたりも増えるので、序盤の余裕が変わります。
	for range l.config.PlayerCnt {
		for i := range l.config.PlayerCnt {
			if c := l.trumpCards.DrawCard(); c != nil {
				l.players[i].AddCard(c)
			}
		}
	}
	l.sortAllHands()

	l.leadPlayerIdx = 0
	l.currentPlayerIdx = 0
	l.addLog(-1, "start", fmt.Sprintf("リンガーロンガーを開始しました（%d 人、%d 枚ずつ）",
		l.config.PlayerCnt, l.config.PlayerCnt), nil)
}

// sortAllHands は手札をスート・ランク順に整える。
func (l *LingerLonger) sortAllHands() {
	for _, p := range l.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return lingerLongerRank(ci) < lingerLongerRank(cj)
		})
	}
}

// leadSuit はこのトリックのリードスートを返す（誰も出していなければ 0）。
func (l *LingerLonger) leadSuit() int {
	if len(l.currentTrick) == 0 {
		return 0
	}
	return l.currentTrick[0].Card.GetDesign()
}

// GetValidPlayIndices はプレイ可能な手札インデックスを返す。**フォロー義務あり。**
//
// **出せる札が無くても捨て札になります。** 脱落制なので、フォローできないことが
// 罰にはなりません——手札を減らせるぶん、むしろ脱落に近づくのが痛いところ。
func (l *LingerLonger) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= l.config.PlayerCnt {
		return nil
	}
	p := l.players[playerIdx]
	lead := l.leadSuit()
	all := make([]int, 0, p.GetCardsSize())
	follow := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
		if lead > 0 && p.GetCard(i).GetDesign() == lead {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// IsHumanTurn は人間の手番かを返す。
func (l *LingerLonger) IsHumanTurn() bool {
	if l.gameEndFlag || l.phase != LingerLongerPhasePlay {
		return false
	}
	return l.players[l.currentPlayerIdx].GetIsHuman()
}

// PlayerPlay は人間が札を出す。
func (l *LingerLonger) PlayerPlay(cardIndex int) error {
	if !l.IsHumanTurn() {
		return ErrNotHumanTurn
	}
	return l.play(l.currentPlayerIdx, cardIndex)
}

// CpuPlay は CPU が 1 枚出す。
func (l *LingerLonger) CpuPlay() {
	if l.gameEndFlag || l.phase != LingerLongerPhasePlay || l.IsHumanTurn() {
		return
	}
	_ = l.play(l.currentPlayerIdx, l.chooseCpuCard(l.currentPlayerIdx))
}

// play は 1 枚出す共通処理。
func (l *LingerLonger) play(playerIdx, cardIndex int) error {
	if l.gameEndFlag {
		return ErrGameEnded
	}
	if l.phase != LingerLongerPhasePlay {
		return ErrWrongPhase
	}
	if playerIdx != l.currentPlayerIdx {
		return ErrNotHumanTurn
	}
	if !lingerLongerContains(l.GetValidPlayIndices(playerIdx), cardIndex) {
		if cardIndex < 0 || cardIndex >= l.players[playerIdx].GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		return errors.New("must follow the led suit")
	}

	card := l.players[playerIdx].RemoveCard(cardIndex)
	if card == nil {
		return NewDomainError(ErrInvalidCard, "カードがありません")
	}
	l.currentTrick = append(l.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	l.addLog(playerIdx, "play", "カードを出しました", []*Card{card})

	if l.trickComplete() {
		l.resolveTrick()
		return nil
	}
	l.advanceTurn()
	return nil
}

// trickComplete はこのトリックで打つべき席が全員打ったかを返す。
//
// **在席数と枚数を比べない。** トリックの途中で誰かが脱落すると在席数がその場で
// 縮み、まだ出していない席を飛ばしてしまいます（#5316 で踏んだ形）。
func (l *LingerLonger) trickComplete() bool {
	if len(l.currentTrick) == 0 {
		return false
	}
	played := make(map[int]bool, len(l.currentTrick))
	for _, tc := range l.currentTrick {
		played[tc.PlayerIdx] = true
	}
	for i, p := range l.players {
		if !p.IsEliminated() && !played[i] {
			return false
		}
	}
	return true
}

// resolveTrick はトリックを解決する。
func (l *LingerLonger) resolveTrick() {
	winner := l.trickWinner()
	cards := make([]*Card, 0, len(l.currentTrick))
	for _, tc := range l.currentTrick {
		cards = append(cards, tc.Card)
	}
	l.players[winner].AddTrickWon()
	l.addLog(winner, "trick", "トリックを取りました（得点にはなりません）", cards)
	l.discarded += len(cards)
	l.currentTrick = nil
	l.trickNumber++

	// **勝った人だけが 1 枚補充できる。** これがこのゲームの全部で、
	// 勝ち続けるかぎり手札が減りません。山札が尽きたら誰も補充できません。
	l.lastDrawIdx = -1
	if c := l.trumpCards.DrawCard(); c != nil {
		l.players[winner].AddCard(c)
		l.lastDrawIdx = winner
		l.sortAllHands()
		l.addLog(winner, "draw", "山札から 1 枚補充しました", []*Card{c})
	}

	// **脱落判定は補充のあと。** 先に見ると、勝って補充する人まで落としてしまいます。
	l.eliminateEmptyHands(winner)
	if l.checkGameEnd(winner) {
		return
	}

	l.leadPlayerIdx = winner
	if l.players[winner].IsEliminated() {
		l.leadPlayerIdx = l.nextActive(winner)
	}
	l.currentPlayerIdx = l.leadPlayerIdx
}

// eliminateEmptyHands は手札が尽きた席を脱落させる。
//
// **勝者から時計回りに見る。** 同じトリックで複数が尽きたとき、脱落の順番が
// 呼び出し順で変わらないようにするためです（決定的）。
func (l *LingerLonger) eliminateEmptyHands(from int) {
	for step := range l.config.PlayerCnt {
		i := (from + step) % l.config.PlayerCnt
		p := l.players[i]
		if p.IsEliminated() || p.GetCardsSize() > 0 {
			continue
		}
		l.eliminatedCnt++
		p.SetEliminatedAt(l.eliminatedCnt)
		l.addLog(i, "eliminate", fmt.Sprintf("手札が尽きて %d 番目に脱落しました", l.eliminatedCnt), nil)
	}
}

// trickWinner はこのトリックの勝者を返す。**切り札はありません。**
func (l *LingerLonger) trickWinner() int {
	if len(l.currentTrick) == 0 {
		return l.leadPlayerIdx
	}
	lead := l.currentTrick[0].Card.GetDesign()
	best := l.currentTrick[0]
	for _, tc := range l.currentTrick[1:] {
		if tc.Card.GetDesign() != lead {
			continue
		}
		if lingerLongerRank(tc.Card) > lingerLongerRank(best.Card) {
			best = tc
		}
	}
	return best.PlayerIdx
}

// lingerLongerContains は xs が v を含むかを返す。
func lingerLongerContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// activeSeats はまだ脱落していない席を返す。
func (l *LingerLonger) activeSeats() []int {
	out := make([]int, 0, l.config.PlayerCnt)
	for i, p := range l.players {
		if !p.IsEliminated() {
			out = append(out, i)
		}
	}
	return out
}

// nextActive は i の次の、まだ脱落していない席を返す。
func (l *LingerLonger) nextActive(i int) int {
	for step := 1; step <= l.config.PlayerCnt; step++ {
		next := (i + step) % l.config.PlayerCnt
		if !l.players[next].IsEliminated() {
			return next
		}
	}
	return i
}

// advanceTurn は次の在席者へ手番を回す。
func (l *LingerLonger) advanceTurn() { l.currentPlayerIdx = l.nextActive(l.currentPlayerIdx) }

// checkGameEnd は決着したかを返す。
//
// **全員同時に尽きることがある。** 山札が空で全員の手札が 1 枚なら、そのトリックで
// 全員 0 になります。そのときは**最後のトリックを取った人**の勝ちにします
// ——「最後まで持ち続けた 1 人」が存在しないので、規則をそのまま当てはめられません。
func (l *LingerLonger) checkGameEnd(lastTrickWinner int) bool {
	active := l.activeSeats()
	switch {
	case len(active) == 1:
		l.finish(active[0])
		return true
	case len(active) == 0:
		l.finish(lastTrickWinner)
		return true
	default:
		return false
	}
}

// finish は終局処理。
func (l *LingerLonger) finish(winner int) {
	l.phase = LingerLongerPhaseGameEnd
	l.gameEndFlag = true
	l.winnerIdx = winner
	l.addLog(winner, "result", "最後まで手札を持ち続けました", nil)
}

// GiveUp は投了する。
func (l *LingerLonger) GiveUp() {
	if l.gameEndFlag {
		return
	}
	l.phase = LingerLongerPhaseGameEnd
	l.gameEndFlag = true
	// 人間（席 0）以外でいちばん手札の多い席を勝ちにする。
	best := -1
	for i := 1; i < l.config.PlayerCnt; i++ {
		if best < 0 || l.players[i].GetCardsSize() > l.players[best].GetCardsSize() {
			best = i
		}
	}
	l.winnerIdx = best
	l.addLog(0, "giveup", "投了しました", nil)
}

// chooseCpuCard は CPU の手。
//
// **取れば補充できるので、取りにいくのが基本。** 取れないなら安い札を捨てます。
func (l *LingerLonger) chooseCpuCard(playerIdx int) int {
	valid := l.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := l.players[playerIdx]

	canWin := false
	for _, i := range valid {
		if l.winsTrick(p.GetCard(i)) {
			canWin = true
			break
		}
	}

	pick, pickRank := valid[0], lingerLongerRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		r := lingerLongerRank(p.GetCard(i))
		switch {
		case canWin && l.winsTrick(p.GetCard(i)) && (!l.winsTrick(p.GetCard(pick)) || r > pickRank):
			pick, pickRank = i, r
		case !canWin && r < pickRank:
			pick, pickRank = i, r
		}
	}
	return pick
}

// winsTrick は card がいまの場を取れるかを返す（場が空ならリードなので取れる扱い）。
func (l *LingerLonger) winsTrick(card *Card) bool {
	if len(l.currentTrick) == 0 {
		return true
	}
	lead := l.currentTrick[0].Card.GetDesign()
	best := l.currentTrick[0].Card
	for _, tc := range l.currentTrick[1:] {
		if tc.Card.GetDesign() == lead && lingerLongerRank(tc.Card) > lingerLongerRank(best) {
			best = tc.Card
		}
	}
	return card.GetDesign() == lead && lingerLongerRank(card) > lingerLongerRank(best)
}

// GetHint は人間への助言を返す。
func (l *LingerLonger) GetHint() *LingerLongerHint {
	if l.gameEndFlag || !l.IsHumanTurn() {
		return nil
	}
	valid := l.GetValidPlayIndices(l.currentPlayerIdx)
	if len(valid) == 0 {
		return nil
	}
	idx := l.chooseCpuCard(l.currentPlayerIdx)
	reason := "lingerlongerDuck"
	if l.winsTrick(l.players[l.currentPlayerIdx].GetCard(idx)) {
		// **取れば補充できる。** 山札が空ならその価値は無い。
		reason = "lingerlongerWinTrick"
		if l.GetStockSize() == 0 {
			reason = "lingerlongerNoStock"
		}
	}
	return &LingerLongerHint{CardIndex: &idx, Reason: reason}
}

// addLog は棋譜に 1 行足す。
func (l *LingerLonger) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	l.appendLog(playerIdx, actionType, detail, cards)
}

// --- アクセサ ---------------------------------------------------------------

// GetConfig はゲーム設定を返す。
func (l *LingerLonger) GetConfig() LingerLongerConfig { return l.config }

// SetConfig はゲーム設定を設定する。**人数が変わると席も作り直す。**
func (l *LingerLonger) SetConfig(cfg LingerLongerConfig) {
	l.config = cfg
	if len(l.players) != cfg.PlayerCnt {
		l.players = newLingerLongerSeats(cfg.PlayerCnt)
	}
}

// GetPhase は現在のフェーズを返す。
func (l *LingerLonger) GetPhase() LingerLongerPhase { return l.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (l *LingerLonger) GetGameEndFlag() bool { return l.gameEndFlag }

// GetStockSize は山札の残り枚数を返す。**補充できるかがここで決まる。**
func (l *LingerLonger) GetStockSize() int {
	if l.trumpCards == nil {
		return 0
	}
	return l.trumpCards.GetRemainingCount()
}

// GetCurrentTrick は現在のトリックを返す。
func (l *LingerLonger) GetCurrentTrick() []*TrickCard { return l.currentTrick }

// GetCurrentPlayerIdx は現在の手番を返す。
func (l *LingerLonger) GetCurrentPlayerIdx() int { return l.currentPlayerIdx }

// GetLeadPlayerIdx はリード席を返す。
func (l *LingerLonger) GetLeadPlayerIdx() int { return l.leadPlayerIdx }

// GetTrickNumber は解決済みのトリック数を返す。
func (l *LingerLonger) GetTrickNumber() int { return l.trickNumber }

// GetLastDrawIdx は直前に補充した席を返す（-1 = 無し）。
func (l *LingerLonger) GetLastDrawIdx() int { return l.lastDrawIdx }

// GetDiscarded は場から抜けた札の枚数を返す。
func (l *LingerLonger) GetDiscarded() int { return l.discarded }

// GetEliminatedCnt は脱落した人数を返す。
func (l *LingerLonger) GetEliminatedCnt() int { return l.eliminatedCnt }

// GetPlayerCnt はプレイヤー数を返す。
func (l *LingerLonger) GetPlayerCnt() int { return l.config.PlayerCnt }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (l *LingerLonger) GetPlayer(i int) *LingerLongerPlayer {
	if i < 0 || i >= len(l.players) {
		return nil
	}
	return l.players[i]
}

// GetWinnerIdx は勝った席を返す（-1: 未確定）。
func (l *LingerLonger) GetWinnerIdx() int { return l.winnerIdx }

// lingerLongerJSON は KV スナップショットの表現。
type lingerLongerJSON struct {
	TrumpCards       *TrumpCards           `json:"tc"`
	Players          []*LingerLongerPlayer `json:"pl"`
	Config           LingerLongerConfig    `json:"cf"`
	Phase            LingerLongerPhase     `json:"ph"`
	CurrentTrick     []*TrickCard          `json:"ct"`
	CurrentPlayerIdx int                   `json:"ci"`
	LeadPlayerIdx    int                   `json:"li"`
	TrickNumber      int                   `json:"tn"`
	EliminatedCnt    int                   `json:"ec"`
	Discarded        int                   `json:"dc"`
	LastDrawIdx      int                   `json:"ld"`
	GameEndFlag      bool                  `json:"ge"`
	WinnerIdx        int                   `json:"wi"`
	ActionLog        []*ActionLogEntry     `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (l *LingerLonger) MarshalJSON() ([]byte, error) {
	return json.Marshal(&lingerLongerJSON{
		TrumpCards: l.trumpCards, Players: l.players, Config: l.config, Phase: l.phase,
		CurrentTrick: l.currentTrick, CurrentPlayerIdx: l.currentPlayerIdx,
		LeadPlayerIdx: l.leadPlayerIdx, TrickNumber: l.trickNumber,
		EliminatedCnt: l.eliminatedCnt, Discarded: l.discarded, LastDrawIdx: l.lastDrawIdx,
		GameEndFlag: l.gameEndFlag, WinnerIdx: l.winnerIdx, ActionLog: l.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元
//
// **10 PR 連続でここに実バグが出た**ので、書き込み側の関数が保っている関係を
// 先に写しています (#5302〜#5316)。範囲検査で捕まったものは 1 つもありません。
func (l *LingerLonger) UnmarshalJSON(data []byte) error {
	var j lingerLongerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < LingerLongerPhasePlay || j.Phase > LingerLongerPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.GameEndFlag != (j.Phase == LingerLongerPhaseGameEnd) {
		return fmt.Errorf("game end flag %v disagrees with phase %d", j.GameEndFlag, j.Phase)
	}
	if j.GameEndFlag != (j.WinnerIdx >= 0) {
		return fmt.Errorf("winner %d disagrees with game end flag %v", j.WinnerIdx, j.GameEndFlag)
	}
	if j.TrumpCards == nil {
		return errors.New("missing trump cards")
	}
	if len(j.Players) != j.Config.PlayerCnt {
		return fmt.Errorf("players has %d entries for %d seats", len(j.Players), j.Config.PlayerCnt)
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
	} {
		if idx < 0 || idx >= j.Config.PlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= j.Config.PlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	if j.LastDrawIdx < -1 || j.LastDrawIdx >= j.Config.PlayerCnt {
		return fmt.Errorf("invalid last draw: %d", j.LastDrawIdx)
	}
	if j.TrickNumber < 0 {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.EliminatedCnt < 0 || j.EliminatedCnt > j.Config.PlayerCnt {
		return fmt.Errorf("invalid eliminated count: %d", j.EliminatedCnt)
	}
	if len(j.CurrentTrick) > j.Config.PlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	// **枚数だけでなく中身も見る (#5310 の再発防止)。**
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= j.Config.PlayerCnt {
			return errors.New("invalid current trick entry")
		}
	}
	if len(j.ActionLog) > lingerLongerMaxSliceLen {
		return errors.New("lingerlonger: input array exceeds maximum allowed size")
	}

	// **脱落の順番は 1..eliminatedCnt の並べ替え。** 数だけ合っても、同じ順番が
	// 2 つあれば「その順で脱落した」経路がありません。
	seen := make(map[int]bool, j.EliminatedCnt)
	total := len(j.CurrentTrick)
	active := 0
	for i, p := range j.Players {
		if p == nil {
			return errors.New("nil player")
		}
		total += p.GetCardsSize()
		if !p.IsEliminated() {
			active++
			continue
		}
		if p.GetEliminatedAt() > j.EliminatedCnt || seen[p.GetEliminatedAt()] {
			return fmt.Errorf("seat %d has elimination order %d, which is not a place in 1..%d",
				i, p.GetEliminatedAt(), j.EliminatedCnt)
		}
		seen[p.GetEliminatedAt()] = true
	}
	if len(seen) != j.EliminatedCnt {
		return fmt.Errorf("%d seats are out but the count says %d", len(seen), j.EliminatedCnt)
	}
	// **札は 52 枚しかない（#5314 の形）。**
	//
	// **抜けた枚数を足して数える。** 解決したトリックは場から抜けるので、
	// 手札と場札と山札だけを足すと途中の盤面では必ず足りません——最初に書いた
	// 検証は、1 トリック解決した時点で正しい盤面を拒否しました。
	if j.Discarded < 0 {
		return fmt.Errorf("invalid discarded count: %d", j.Discarded)
	}
	if got := total + j.TrumpCards.GetRemainingCount() + j.Discarded; got != LingerLongerDeckSize {
		return fmt.Errorf("hands, the trick, the stock and the discards hold %d cards, want %d",
			got, LingerLongerDeckSize)
	}
	// **手番とリード席は脱落した席にならない。**
	if !j.GameEndFlag {
		if j.Players[j.CurrentPlayerIdx].IsEliminated() {
			return fmt.Errorf("seat %d is on turn but is out", j.CurrentPlayerIdx)
		}
		if j.Players[j.LeadPlayerIdx].IsEliminated() {
			return fmt.Errorf("seat %d leads but is out", j.LeadPlayerIdx)
		}
		// **場札は在席数より少ない。** 揃った時点で解決されるので残りません。
		if len(j.CurrentTrick) >= active {
			return fmt.Errorf("current trick holds %d cards with %d seats still in",
				len(j.CurrentTrick), active)
		}
	}

	l.trumpCards = j.TrumpCards
	l.players, l.config, l.phase = j.Players, j.Config, j.Phase
	l.currentTrick, l.currentPlayerIdx = j.CurrentTrick, j.CurrentPlayerIdx
	l.leadPlayerIdx, l.trickNumber = j.LeadPlayerIdx, j.TrickNumber
	l.eliminatedCnt, l.lastDrawIdx = j.EliminatedCnt, j.LastDrawIdx
	l.discarded = j.Discarded
	l.gameEndFlag, l.winnerIdx, l.actionLog = j.GameEndFlag, j.WinnerIdx, j.ActionLog
	return nil
}
