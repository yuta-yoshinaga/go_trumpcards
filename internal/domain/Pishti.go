//go:build !js || !wasm || extra2

// Package domain (Pişti) implements Pişti — a Turkish fishing / capture game.
//
// # Rules summary
//
// Pişti is played by 2–4 players (seat 0 is the human) with a standard 52-card
// deck. A face-up pile sits in the centre. On a turn the active player plays one
// hand card onto the pile. The player CAPTURES the entire pile when:
//
//   - the played card's RANK equals the current pile-top card's rank, OR
//   - the played card is a Jack (value 11). A Jack is a wild capture card.
//
// On a capture the whole pile (including the just-played card) moves to the
// player's captured pile and the centre becomes empty. If there is no capture
// the played card simply becomes the new pile top.
//
// # Pişti bonus
//
// Capturing when the pile held exactly ONE card immediately before the play is a
// "Pişti" and awards a bonus:
//
//   - +10 when matching a lone single card (rank match or a Jack onto a lone
//     non-Jack), per PishtiBonusSingle.
//   - +20 when a Jack is played onto a lone Jack, per PishtiBonusJackOnJack.
//
// A Jack (or rank match) capturing a multi-card pile is a normal capture with no
// Pişti bonus.
//
// # Dealing
//
// At the start each player is dealt PishtiHandSize (4) cards and
// PishtiInitialPileSize (4) cards are placed on the centre pile. To keep play
// clean the issue specifies the starting top card must not be a Jack: if the top
// (4th) initial card is a Jack it is shuffled back into the stock and another
// card is drawn until a non-Jack tops the pile (or the stock empties). When all
// hands are empty 4 more cards are dealt to each player from the stock; this
// repeats until the stock is exhausted.
//
// # Final scoring (calcFinalScore)
//
// When the stock and all hands are exhausted, any remaining centre pile goes to
// the last player who captured. Each player then scores:
//
//   - +3 for the most captured cards (PishtiScoreMostCards; ties award nobody),
//   - +1 per Ace (PishtiScoreAce),
//   - +2 for the 2♣ (PishtiScoreTwoClubs),
//   - +3 for the 10♦ (PishtiScoreTenDiamonds),
//   - +1 per Jack (PishtiScoreJack),
//   - plus accumulated Pişti bonuses.
//
// The highest total wins. Ties on the final total yield multiple winners
// (GetWinners returns all tied seats).
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// PishtiPhase は Pişti のフェーズを表す。
type PishtiPhase string

// Pişti のフェーズ定数。
const (
	// PishtiPhasePlay プレイ中 (カードを場へ出す)
	PishtiPhasePlay PishtiPhase = "play"
	// PishtiPhaseRoundEnd ラウンド (配り直し) 締め処理。内部遷移用。
	PishtiPhaseRoundEnd PishtiPhase = "roundEnd"
	// PishtiPhaseGameEnd ゲーム終了
	PishtiPhaseGameEnd PishtiPhase = "gameEnd"
)

// pishtiMaxSliceLen はデシリアライズ時のスライス長上限。
const pishtiMaxSliceLen = 1000

// pishtiState はゲーム進行状態。
type pishtiState struct {
	phase          PishtiPhase
	currentTurn    int     // 現在の手番プレイヤー
	pile           []*Card // 場の山 (末尾が一番上)
	lastCaptureIdx int     // 最後に捕獲したプレイヤー (-1 = なし)
	gameEndFlag    bool
	winners        []int
	actionLogBase
}

// Pishti は Pişti ゲームの状態を保持する集約ルート。
type Pishti struct {
	trumpCards *TrumpCards
	players    []*PishtiPlayer
	config     PishtiConfig
	state      pishtiState
}

// NewPishti コンストラクタ。
func NewPishti(trumpCards *TrumpCards, players []*PishtiPlayer, config PishtiConfig) *Pishti {
	return &Pishti{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		state: pishtiState{
			phase:          PishtiPhasePlay,
			lastCaptureIdx: -1,
		},
	}
}

// NewDefaultPishti はデフォルト設定 (4人: 1 human + 3 CPU) の Pişti を返す。
func NewDefaultPishti() *Pishti {
	config := DefaultPishtiConfig()
	players := makePishtiPlayers(config.PlayerCnt)
	return NewPishti(NewTrumpCards(0), players, config)
}

// makePishtiPlayers は seat 0 を人間とした cnt 人のプレイヤースライスを作る。
func makePishtiPlayers(cnt int) []*PishtiPlayer {
	if cnt < PishtiMinPlayerCnt {
		cnt = PishtiMinPlayerCnt
	}
	if cnt > PishtiMaxPlayerCnt {
		cnt = PishtiMaxPlayerCnt
	}
	players := make([]*PishtiPlayer, cnt)
	players[0] = NewPishtiPlayer(true)
	for i := 1; i < cnt; i++ {
		players[i] = NewPishtiPlayer(false)
	}
	return players
}

// Reset は新しいゲームを開始する。
func (g *Pishti) Reset() {
	g.players = makePishtiPlayers(g.config.PlayerCnt)
	g.trumpCards = NewTrumpCards(0)
	g.trumpCards.Shuffle()
	g.state = pishtiState{
		phase:          PishtiPhasePlay,
		currentTurn:    0,
		lastCaptureIdx: -1,
		actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.dealHands()
	g.dealInitialPile()
}

// NextRound は Pişti では新規ゲーム開始 (Reset) と同義。
// 1 ゲームは山札を配り切るまでの 1 セッションで完結するため、終局後の続行は
// 新しいゲームの開始として扱う。
func (g *Pishti) NextRound() {
	g.Reset()
}

// dealHands は各プレイヤーへ PishtiHandSize 枚配る。山札が尽きたら途中で終わる。
func (g *Pishti) dealHands() {
	for k := 0; k < PishtiHandSize; k++ {
		for i := 0; i < len(g.players); i++ {
			card := g.trumpCards.DrawCard()
			if card == nil {
				return
			}
			g.players[i].AddCard(card)
		}
	}
}

// dealInitialPile はゲーム開始時に PishtiInitialPileSize 枚を場へ置く。
// 一番上 (最後に置く札) がジャックの場合は、続けて山札から引いた非ジャック札と
// 入れ替え、ジャックを山の下に潜らせて非ジャックを上にする (issue #2329 のルール:
// 開始時の一番上がジャックにならないようにする)。山札が尽きた場合は引き直しを
// 諦め、ジャックのまま開始する。TrumpCards は引いた札を山へ戻せないため、
// 「戻して引き直す」代わりに余分に 1 枚引いて入れ替える等価な処理とする。
func (g *Pishti) dealInitialPile() {
	for i := 0; i < PishtiInitialPileSize; i++ {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.state.pile = append(g.state.pile, card)
	}
	// 一番上がジャックなら、場札内の非ジャック札と入れ替えて上に出す。
	// 山札からは引かない (引くと総配布枚数が減り、最終ラウンドの配布が偏って
	// しまうため)。場札が全てジャックの稀なケースはそのまま開始する。
	if pishtiIsJack(g.pileTop()) {
		top := len(g.state.pile) - 1
		for i := top - 1; i >= 0; i-- {
			if !pishtiIsJack(g.state.pile[i]) {
				g.state.pile[i], g.state.pile[top] = g.state.pile[top], g.state.pile[i]
				break
			}
		}
	}
	g.appendLog(-1, "deal", fmt.Sprintf("dealt %d pile cards", len(g.state.pile)), append([]*Card(nil), g.state.pile...))
}

// pishtiIsJack はジャック (value 11) かどうか。
func pishtiIsJack(c *Card) bool {
	return c != nil && c.GetValue() == PishtiJackValue
}

// pileTop は場の一番上の札を返す (なければ nil)。
func (g *Pishti) pileTop() *Card {
	if len(g.state.pile) == 0 {
		return nil
	}
	return g.state.pile[len(g.state.pile)-1]
}

// allHandsEmpty は全員の手札が空かどうか。
func (g *Pishti) allHandsEmpty() bool {
	return allHandsEmpty(g.players)
}

// PlayerPlay は人間プレイヤーが手札 cardIndex を場へ出す。
func (g *Pishti) PlayerPlay(cardIndex int) error {
	if g.state.gameEndFlag {
		return ErrGameEnded
	}
	if g.state.phase != PishtiPhasePlay {
		return NewDomainError(ErrWrongPhase, "not in play phase")
	}
	if !g.players[g.state.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyPlay(g.state.currentTurn, cardIndex)
}

// CpuPlay は CPU のターンを 1 回進める。
func (g *Pishti) CpuPlay() {
	if g.state.gameEndFlag || g.state.phase != PishtiPhasePlay {
		return
	}
	p := g.players[g.state.currentTurn]
	if p.GetIsHuman() || p.GetCardsSize() == 0 {
		return
	}
	idx := g.chooseCpuCard(g.state.currentTurn)
	_ = g.applyPlay(g.state.currentTurn, idx)
}

// applyPlay は cardIndex の手札を場へ出し、捕獲判定・Pişti 判定・手番進行を行う。
func (g *Pishti) applyPlay(playerIdx, cardIndex int) error {
	player := g.players[playerIdx]
	card := player.GetCard(cardIndex)
	if card == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", cardIndex))
	}

	top := g.pileTop()
	pileSizeBefore := len(g.state.pile)
	captures := g.isCapture(card, top)

	_ = player.RemoveCard(cardIndex)
	g.state.pile = append(g.state.pile, card)

	if captures {
		// Pişti 判定: 直前に場が単独札だった場合。
		bonus := 0
		if pileSizeBefore == 1 {
			if pishtiIsJack(card) && pishtiIsJack(top) {
				bonus = PishtiBonusJackOnJack
			} else {
				bonus = PishtiBonusSingle
			}
		}
		captured := append([]*Card(nil), g.state.pile...)
		player.AddCaptured(captured)
		g.state.pile = g.state.pile[:0]
		g.state.lastCaptureIdx = playerIdx
		if bonus > 0 {
			player.AddPistiBonus(bonus)
			g.appendLog(playerIdx, "pisti", fmt.Sprintf("Pişti +%d (captured %d)", bonus, len(captured)), captured)
		} else {
			g.appendLog(playerIdx, "capture", fmt.Sprintf("captured %d card(s)", len(captured)), captured)
		}
	} else {
		g.appendLog(playerIdx, "play", "played onto pile", []*Card{card})
	}

	g.advanceTurn()
	return nil
}

// isCapture は card を top に重ねたとき捕獲が成立するかを返す。
// 場が空 (top == nil) の場合は捕獲不成立。
func (g *Pishti) isCapture(card, top *Card) bool {
	if top == nil {
		return false
	}
	if pishtiIsJack(card) {
		return true
	}
	return card.GetValue() == top.GetValue()
}

// advanceTurn は手番を次に進め、必要なら配り直し・終局処理を行う。
func (g *Pishti) advanceTurn() {
	g.state.currentTurn = (g.state.currentTurn + 1) % len(g.players)
	if !g.allHandsEmpty() {
		return
	}
	if g.trumpCards.GetRemainingCount() > 0 {
		g.dealHands()
		return
	}
	g.finishGame()
}

// finishGame は終局処理: 残りの場札を最後の捕獲者へ渡し、最終得点を確定する。
func (g *Pishti) finishGame() {
	g.state.phase = PishtiPhaseRoundEnd
	if g.state.lastCaptureIdx >= 0 && len(g.state.pile) > 0 {
		leftover := append([]*Card(nil), g.state.pile...)
		g.players[g.state.lastCaptureIdx].AddCaptured(leftover)
		g.appendLog(g.state.lastCaptureIdx, "lastTake", fmt.Sprintf("last-take: %d card(s)", len(leftover)), leftover)
	}
	g.state.pile = g.state.pile[:0]

	scores := g.calcFinalScore()
	maxScore := 0
	for _, s := range scores {
		if s > maxScore {
			maxScore = s
		}
	}
	winners := make([]int, 0)
	for i, s := range scores {
		if s == maxScore {
			winners = append(winners, i)
		}
	}
	g.state.winners = winners
	g.state.gameEndFlag = true
	g.state.phase = PishtiPhaseGameEnd
	g.appendLog(-1, "gameEnd", fmt.Sprintf("game ended (top score %d)", maxScore), nil)
}

// calcFinalScore は各プレイヤーの最終得点を計算する。
// インデックスはプレイヤーシートに対応する。
func (g *Pishti) calcFinalScore() []int {
	n := len(g.players)
	scores := make([]int, n)
	cardCounts := make([]int, n)

	// 最多枚数判定 (同点なら誰ももらえない)。
	mostIdx := g.mostCapturedSeat()
	for i, p := range g.players {
		cardCounts[i] = p.CapturedCount()
	}
	tie := mostIdx < 0
	mostVal := 0
	if mostIdx >= 0 {
		mostVal = g.players[mostIdx].CapturedCount()
	}

	for i, p := range g.players {
		s := 0
		for _, c := range p.GetCapturedCards() {
			s += pishtiCardPoints(c)
		}
		s += p.GetPistiBonus()
		scores[i] = s
	}
	if !tie && mostVal > 0 && mostIdx >= 0 {
		scores[mostIdx] += PishtiScoreMostCards
	}
	return scores
}

// mostCapturedSeat は最多捕獲の単独リーダーの席を返す。同数、または誰も
// 捕獲していなければ -1。
func (g *Pishti) mostCapturedSeat() int {
	best, bestSeat, tie := 0, -1, false
	for i, p := range g.players {
		cnt := p.CapturedCount()
		if bestSeat == -1 || cnt > best {
			best, bestSeat, tie = cnt, i, false
		} else if cnt == best {
			tie = true
		}
	}
	if tie || best == 0 {
		return -1
	}
	return bestSeat
}

// GetProvisionalScores は対局中の暫定スコアを返す。
//
// **カード点は含まない。**捕獲札の点数配分は最後に数えるので、途中で確実に
// 分かるのはピシュティ賞と最多捕獲の +3 だけ。近似であることは呼び出し側が
// 明示する (#4892)。**同数なら誰にも +3 は付かない。**
func (g *Pishti) GetProvisionalScores() []int {
	out := make([]int, len(g.players))
	leader := g.mostCapturedSeat()
	for i, p := range g.players {
		out[i] = p.GetPistiBonus()
		if i == leader {
			out[i] += PishtiScoreMostCards
		}
	}
	return out
}

// GetProvisionalLeader は暫定の最多捕獲リーダーの席を返す (同数なら -1)。
func (g *Pishti) GetProvisionalLeader() int { return g.mostCapturedSeat() }

// pishtiCardPoints は 1 枚のカードの基本得点を返す。
//
//	A (value 1)  → +1
//	2♣           → +2
//	10♦          → +3
//	J (value 11) → +1
//	その他       → 0
func pishtiCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	d := c.GetDesign()
	pts := 0
	if v == 1 {
		pts += PishtiScoreAce
	}
	if v == PishtiJackValue {
		pts += PishtiScoreJack
	}
	if v == 2 && d == CardDesignClover {
		pts += PishtiScoreTwoClubs
	}
	if v == 10 && d == CardDesignDiamond {
		pts += PishtiScoreTenDiamonds
	}
	return pts
}

// chooseCpuCard は CPU の手番で出す手札インデックスを選ぶ。
//   - Easy   : 合法手 (常に全札合法) からランダム。
//   - Normal : 捕獲できる札を優先、無ければ最も価値の低い札を捨てる。
//   - Hard   : Pişti / 高得点札を最優先で狙い、無ければ価値の低い札を捨てる。
func (g *Pishti) chooseCpuCard(playerIdx int) int {
	player := g.players[playerIdx]
	size := player.GetCardsSize()
	if size == 0 {
		return 0
	}
	top := g.pileTop()
	pileSize := len(g.state.pile)

	switch g.config.CpuDifficulty {
	case PishtiDifficultyEasy:
		return rand.Intn(size)
	case PishtiDifficultyHard:
		// Pişti が狙えるなら最優先。
		if pileSize == 1 {
			if idx := g.findCapturingCard(player, top); idx >= 0 {
				return idx
			}
		}
		if idx := g.findCapturingCard(player, top); idx >= 0 {
			return idx
		}
		return g.lowestValueCardIdx(player)
	default: // Normal
		if idx := g.findCapturingCard(player, top); idx >= 0 {
			return idx
		}
		return g.lowestValueCardIdx(player)
	}
}

// findCapturingCard は捕獲できる手札のインデックスを返す (なければ -1)。
// 複数ある場合はジャック以外の rank-match を優先し、ジャックは温存する。
func (g *Pishti) findCapturingCard(player *PishtiPlayer, top *Card) int {
	if top == nil {
		return -1
	}
	jackIdx := -1
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c == nil {
			continue
		}
		if pishtiIsJack(c) {
			if jackIdx == -1 {
				jackIdx = i
			}
			continue
		}
		if c.GetValue() == top.GetValue() {
			return i
		}
	}
	return jackIdx
}

// lowestValueCardIdx は最も得点価値の低い手札のインデックスを返す。
// 同価値なら最初に見つかったものを返す。
func (g *Pishti) lowestValueCardIdx(player *PishtiPlayer) int {
	best := 0
	bestPts := 1 << 30
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c == nil {
			continue
		}
		pts := pishtiCardPoints(c)
		// ジャックは捕獲ワイルドなので捨てたくない → 価値を底上げ。
		if pishtiIsJack(c) {
			pts += PishtiScoreJack
		}
		if pts < bestPts {
			bestPts = pts
			best = i
		}
	}
	return best
}

// appendLog は棋譜にエントリを追加する。
func (g *Pishti) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// --- 状態アクセサ ---

// IsHumanTurn は現在の手番が人間かどうかを返す。
func (g *Pishti) IsHumanTurn() bool {
	if g.state.gameEndFlag {
		return false
	}
	return g.players[g.state.currentTurn].GetIsHuman()
}

// GetCurrentTurn は現在の手番プレイヤーインデックスを返す。
func (g *Pishti) GetCurrentTurn() int { return g.state.currentTurn }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Pishti) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetPhase は現在のフェーズを返す。
func (g *Pishti) GetPhase() PishtiPhase { return g.state.phase }

// GetPile は場の山を返す (末尾が一番上)。
func (g *Pishti) GetPile() []*Card { return g.state.pile }

// GetPileTop は場の一番上の札を返す (なければ nil)。
func (g *Pishti) GetPileTop() *Card { return g.pileTop() }

// GetLastCaptureIdx は最後に捕獲したプレイヤーを返す (-1 = なし)。
func (g *Pishti) GetLastCaptureIdx() int { return g.state.lastCaptureIdx }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Pishti) GetPlayer(idx int) *PishtiPlayer {
	return getPlayer(g.players, idx)
}

// GetPlayerCnt はプレイヤー数を返す。
func (g *Pishti) GetPlayerCnt() int { return len(g.players) }

// GetRemainingDeck は山札の残り枚数を返す。
func (g *Pishti) GetRemainingDeck() int { return g.trumpCards.GetRemainingCount() }

// GetConfig は設定を返す。
func (g *Pishti) GetConfig() PishtiConfig { return g.config }

// SetConfig は設定を変更する。
func (g *Pishti) SetConfig(config PishtiConfig) { g.config = config }

// SetGameEndFlagForTest はテスト用に終了フラグを設定する。
func (g *Pishti) SetGameEndFlagForTest(v bool) { g.state.gameEndFlag = v }

// GetActionLog は棋譜を返す。
func (g *Pishti) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// GetWinners はゲーム終了時の勝者シートのリストを返す (同点なら複数)。
func (g *Pishti) GetWinners() []int { return g.state.winners }

// GetFinalScores は現在の最終得点を計算して返す (途中経過でも算出可能)。
func (g *Pishti) GetFinalScores() []int { return g.calcFinalScore() }

// --- JSON Serialization ---

// pishtiJSON is the JSON wire format for Pishti.
type pishtiJSON struct {
	TrumpCards     *TrumpCards       `json:"tc"`
	Players        []*PishtiPlayer   `json:"pl"`
	Config         PishtiConfig      `json:"cf"`
	Phase          PishtiPhase       `json:"ph"`
	CurrentTurn    int               `json:"ct"`
	Pile           []*Card           `json:"pi"`
	LastCaptureIdx int               `json:"lc"`
	GameEndFlag    bool              `json:"ge"`
	Winners        []int             `json:"wn"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Pishti) MarshalJSON() ([]byte, error) {
	return json.Marshal(pishtiJSON{
		TrumpCards:     g.trumpCards,
		Players:        g.players,
		Config:         g.config,
		Phase:          g.state.phase,
		CurrentTurn:    g.state.currentTurn,
		Pile:           g.state.pile,
		LastCaptureIdx: g.state.lastCaptureIdx,
		GameEndFlag:    g.state.gameEndFlag,
		Winners:        g.state.winners,
		ActionLog:      g.state.actionLog,
	})
}

// pishtiValidPhase は有効なフェーズかどうか。
func pishtiValidPhase(p PishtiPhase) bool {
	switch p {
	case PishtiPhasePlay, PishtiPhaseRoundEnd, PishtiPhaseGameEnd:
		return true
	default:
		return false
	}
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否するための
// バリデーションを行う。
func (g *Pishti) UnmarshalJSON(data []byte) error {
	var j pishtiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > pishtiMaxSliceLen || len(j.Pile) > pishtiMaxSliceLen ||
		len(j.ActionLog) > pishtiMaxSliceLen || len(j.Winners) > pishtiMaxSliceLen {
		return fmt.Errorf("pishti: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("pishti: invalid config: %w", err)
	}
	if j.TrumpCards == nil {
		return fmt.Errorf("pishti: missing trump cards in state")
	}
	if len(j.Players) < PishtiMinPlayerCnt || len(j.Players) > PishtiMaxPlayerCnt {
		return fmt.Errorf("pishti: player count out of range")
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("pishti: nil player in state")
		}
	}
	if !pishtiValidPhase(j.Phase) {
		return fmt.Errorf("pishti: invalid phase")
	}
	if j.CurrentTurn < 0 || j.CurrentTurn >= len(j.Players) {
		return fmt.Errorf("pishti: current turn out of range")
	}
	if j.LastCaptureIdx < -1 || j.LastCaptureIdx >= len(j.Players) {
		return fmt.Errorf("pishti: last capture index out of range")
	}
	for _, w := range j.Winners {
		if w < 0 || w >= len(j.Players) {
			return fmt.Errorf("pishti: winner index out of range")
		}
	}

	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.config = j.Config
	g.state = pishtiState{
		phase:          j.Phase,
		currentTurn:    j.CurrentTurn,
		pile:           j.Pile,
		lastCaptureIdx: j.LastCaptureIdx,
		gameEndFlag:    j.GameEndFlag,
		winners:        j.Winners,
		actionLogBase:  actionLogBase{actionLog: j.ActionLog},
	}
	if g.state.pile == nil {
		g.state.pile = make([]*Card, 0)
	}
	if g.state.actionLog == nil {
		g.state.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
