//go:build !js || !wasm || extra3

// Package domain — シートグッベ (Skitgubbe) のドメインモデル。
//
// スウェーデンの 2 フェーズ構成のカードゲーム。52 枚、手札 3 枚から始まる。
// 最後まで手札が残った者が敗者 (Skitgubbe = 「汚いおっさん」)。
//
// # issue #4404 の仕様案との相違
//
// issue は**両フェーズの機構を取り違えている**ので、実ルール (pagat.com) を採る。
//
// 第1フェーズ:
//
//   - issue は「リードに対し**他家が**同スート以上で応じる」とするが、実際は
//     **常に 2 人の一騎打ち**で、**スートは無関係**。強い方が両方の札を取る
//   - issue は「勝者が 1 枚補充」とするが、実際は**両者が 3 枚に戻るまで引く**
//   - issue が触れていない **stunsa (バウンス)**: 同ランクなら 2 枚を場に残した
//     まま両者が 1 枚引き、**同じ人がもう一度リード**する。決着した時点で
//     たまった札を勝者が全部取る
//   - issue が触れていない: **切札は山札から最後に引かれた札**で決まる
//
// 第2フェーズ:
//
//   - issue は「**昇順**で出し切る」とするが、実際は**直前の札を上回る**
//     (同スートの上位か切札)。昇順ではない
//   - issue は「出せなければ**脱落**」とするが、実際は**場の札を手札へ引き取る**
//   - 弱い札を出して逃げることはできない ("It is never lawful to duck")
//
// 「昇順の出し切り」だと shithead 系の shedding になるが、実際は durak 系の
// beating で、ゲームの性質が丸ごと変わる。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// SkitgubbePlayerCnt はプレイヤー数 (3 人が標準)。
const SkitgubbePlayerCnt = 3

// SkitgubbeHandSize は第1フェーズで維持する手札枚数。
const SkitgubbeHandSize = 3

// SkitgubbeDeckSize は 52 枚。
const SkitgubbeDeckSize = 52

// SkitgubbeRankOrder はカードの強さ。A が最強、2 が最弱。
//
// A > K > Q > J > 10 > … > 2。値 1 (A) だけが順序と逆転するので表に切り出す。
func SkitgubbeRankOrder(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v == 1 {
		return 14 // A は最強
	}
	return c.GetValue()
}

// SkitgubbePhase はゲームフェーズ。
type SkitgubbePhase int

// Skitgubbeのフェーズ定数
const (
	// SkitgubbePhaseCollect 第1フェーズ (一騎打ちで札を集める)
	SkitgubbePhaseCollect SkitgubbePhase = iota
	// SkitgubbePhaseShed 第2フェーズ (集めた札を出し切る)
	SkitgubbePhaseShed
	// SkitgubbePhaseGameEnd 終局
	SkitgubbePhaseGameEnd
)

// newSkitgubbeDeck は 52 枚を生成する (シャッフル前)。
func newSkitgubbeDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, SkitgubbeDeckSize)
	for _, s := range suits {
		for v := 1; v <= 13; v++ {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// skitgubbeShuffle は Fisher-Yates。domain の shuffleCards は casino タグの
// ファイルにあり extra3 ビルドから見えないため、専用名で持つ。
func skitgubbeShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// Skitgubbe はシートグッベのゲームクラス。
type Skitgubbe struct {
	players   []*SkitgubbePlayer
	config    SkitgubbeConfig
	phase     SkitgubbePhase
	stock     []*Card
	trumpSuit int
	// duel は第1フェーズで場に出ている札。stunsa が続くと積み上がる。
	duel       []*Card
	duelLeader int
	currentIdx int
	collected  [][]*Card
	// pile は第2フェーズで場に出ている札。取れなかった者が引き取る。
	pile        []*Card
	pileLeader  int
	finished    []bool
	loserIdx    int
	gameEndFlag bool
	actionLog   []*ActionLogEntry
}

// NewSkitgubbe はコンストラクタ。
func NewSkitgubbe(players []*SkitgubbePlayer, config SkitgubbeConfig) *Skitgubbe {
	return &Skitgubbe{
		players:   players,
		config:    config,
		collected: make([][]*Card, len(players)),
		finished:  make([]bool, len(players)),
		loserIdx:  -1,
	}
}

// NewDefaultSkitgubbe は標準の 3 人セットアップを返す。
func NewDefaultSkitgubbe() *Skitgubbe {
	players := make([]*SkitgubbePlayer, 0, SkitgubbePlayerCnt)
	players = append(players, NewSkitgubbePlayer(true))
	for range SkitgubbePlayerCnt - 1 {
		players = append(players, NewSkitgubbePlayer(false))
	}
	return NewSkitgubbe(players, DefaultSkitgubbeConfig())
}

// Reset はゲームを初期化する。
func (s *Skitgubbe) Reset() {
	s.phase = SkitgubbePhaseCollect
	s.duel = nil
	s.pile = nil
	s.trumpSuit = -1
	s.loserIdx = -1
	s.gameEndFlag = false
	s.actionLog = nil
	s.finished = make([]bool, len(s.players))
	s.collected = make([][]*Card, len(s.players))
	for i := range s.collected {
		s.collected[i] = make([]*Card, 0, SkitgubbeDeckSize)
	}
	for _, p := range s.players {
		p.ResetGame()
	}

	deck := newSkitgubbeDeck()
	skitgubbeShuffle(deck)
	pos := 0
	for range SkitgubbeHandSize {
		for _, p := range s.players {
			p.AddCard(deck[pos])
			pos++
		}
	}
	s.stock = append([]*Card(nil), deck[pos:]...)

	s.duelLeader = 0
	s.currentIdx = 0
	s.addLog(-1, "deal", "cards dealt", nil)
}

// duelOpponent は第1フェーズでリード側の相手 (左隣) を返す。
func (s *Skitgubbe) duelOpponent() int {
	return (s.duelLeader + 1) % len(s.players)
}

// PlayCard は player が手札 handIdx の札を出す。
func (s *Skitgubbe) PlayCard(player, handIdx int) error {
	if s.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if player != s.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	p := s.GetPlayer(player)
	if p == nil {
		return fmt.Errorf("no such player: %d", player)
	}
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}
	if s.phase == SkitgubbePhaseShed && !s.canBeat(p.GetCard(handIdx)) {
		return fmt.Errorf("card index %d does not beat the pile; you may not duck", handIdx)
	}

	card := p.RemoveCard(handIdx)
	if card == nil {
		return fmt.Errorf("card index %d is empty", handIdx)
	}
	if s.phase == SkitgubbePhaseCollect {
		s.resolveDuelPlay(player, card)
		return nil
	}
	s.resolveShedPlay(player, card)
	return nil
}

// resolveDuelPlay は第1フェーズの 1 枚を処理する。
func (s *Skitgubbe) resolveDuelPlay(player int, card *Card) {
	s.duel = append(s.duel, card)
	s.addLog(player, "play", "plays a card", []*Card{card})

	// リード側が出した直後なら、相手の番。
	if player == s.duelLeader {
		s.currentIdx = s.duelOpponent()
		return
	}

	// 2 枚そろった。直近の 2 枚で比べる。
	lead := s.duel[len(s.duel)-2]
	resp := s.duel[len(s.duel)-1]
	leadRank, respRank := SkitgubbeRankOrder(lead), SkitgubbeRankOrder(resp)

	if leadRank == respRank {
		// stunsa: 札は場に残したまま、両者が引いて同じ人がもう一度リード。
		s.addLog(-1, "stunsa", "equal ranks -- the cards stay and the lead repeats", nil)
		s.drawUp(s.duelLeader)
		s.drawUp(player)
		s.currentIdx = s.duelLeader
		s.checkCollectEnd()
		return
	}

	winner := s.duelLeader
	if respRank > leadRank {
		winner = player
	}
	s.collected[winner] = append(s.collected[winner], s.duel...)
	s.addLog(winner, "win", fmt.Sprintf("takes %d card(s)", len(s.duel)), s.duel)
	s.duel = nil

	s.drawUp(s.duelLeader)
	s.drawUp(player)
	s.duelLeader = winner
	s.currentIdx = winner
	s.checkCollectEnd()
}

// drawUp は player の手札を 3 枚まで補充する。山札の最後の 1 枚が切札を決める。
func (s *Skitgubbe) drawUp(player int) {
	p := s.GetPlayer(player)
	if p == nil {
		return
	}
	for p.GetCardsSize() < SkitgubbeHandSize && len(s.stock) > 0 {
		card := s.stock[0]
		s.stock = s.stock[1:]
		p.AddCard(card)
		if len(s.stock) == 0 {
			// **山札から最後に引かれた札が切札を決める。** issue はこの規則に
			// 触れていないが、第2フェーズの強さがこれで決まる。
			s.trumpSuit = card.GetDesign()
			s.addLog(-1, "trump", fmt.Sprintf("the last card drawn sets trump to %d", s.trumpSuit), []*Card{card})
		}
	}
}

// checkCollectEnd は第1フェーズの終了条件を見て、満たしていれば移行する。
//
// 出典は「山札が尽き、**出せる人がいなくなったら**終わる」。全員の手札が空に
// なるのを待ってはいけない —— 一騎打ちは 2 人でしか行われないので、3 人戦では
// 一度もリードに絡まなかった人の手札が残り、その条件は永遠に満たされない
// (実際そうなった: 手番の人が 0 枚なのに他の 2 人が 3 枚と 1 枚を持ったまま
// 進行が止まった)。
func (s *Skitgubbe) checkCollectEnd() {
	if len(s.stock) > 0 {
		return
	}
	// 一騎打ちには**両者の手札**が要る。補充できなくなった以上、次の一騎打ちの
	// どちらかが尽きた時点で第1フェーズは終わりで、リード側だけを見てはいけない
	// (相手が 0 枚だと、リードは出せるのに応じ手がなく進行が止まる)。
	for _, idx := range []int{s.duelLeader, s.duelOpponent()} {
		if p := s.GetPlayer(idx); p == nil || p.GetCardsSize() == 0 {
			s.startShedPhase()
			return
		}
	}
}

// startShedPhase は第2フェーズへ移る。集めた札がそのまま手札になる。
func (s *Skitgubbe) startShedPhase() {
	// 決着のつかなかった duel の札は誰のものにもならない。stunsa の途中で
	// 第1フェーズが終わった場合にだけ起きる。
	s.duel = nil
	s.phase = SkitgubbePhaseShed
	s.pile = nil

	// 第2フェーズの手札は「集めた札 + 手元に残っている札」。第1フェーズは
	// 一騎打ちなので、一度もリードに絡まなかった人には手札が残る。それを捨てると
	// 山からカードが消えてしまうし、その人だけ極端に有利になる。
	for i, p := range s.players {
		leftover := make([]*Card, 0, p.GetCardsSize())
		for j := range p.GetCardsSize() {
			leftover = append(leftover, p.GetCard(j))
		}
		p.Reset()
		for _, c := range s.collected[i] {
			p.AddCard(c)
		}
		for _, c := range leftover {
			p.AddCard(c)
		}
		s.collected[i] = nil
		s.finished[i] = p.GetCardsSize() == 0
	}
	if s.trumpSuit < 0 {
		s.trumpSuit = CardDesignSpade
	}

	s.pileLeader = s.nextActive(len(s.players) - 1)
	s.currentIdx = s.pileLeader
	s.addLog(-1, "phase", "the shedding phase begins", nil)
	s.checkShedEnd()
}

// canBeat は card が場の一番上を上回るかを返す。
//
// **弱い札で逃げることはできない。**場が空ならどの札でも出せる。
func (s *Skitgubbe) canBeat(card *Card) bool {
	if card == nil {
		return false
	}
	if len(s.pile) == 0 {
		return true
	}
	top := s.pile[len(s.pile)-1]
	cardTrump := card.GetDesign() == s.trumpSuit
	topTrump := top.GetDesign() == s.trumpSuit
	switch {
	case cardTrump && !topTrump:
		return true
	case !cardTrump && topTrump:
		return false
	case card.GetDesign() != top.GetDesign():
		return false
	default:
		return SkitgubbeRankOrder(card) > SkitgubbeRankOrder(top)
	}
}

// GetValidPlayIndices は player が出せる手札の添字を返す。
func (s *Skitgubbe) GetValidPlayIndices(player int) []int {
	p := s.GetPlayer(player)
	if p == nil {
		return nil
	}
	var out []int
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		if s.phase == SkitgubbePhaseCollect || s.canBeat(c) {
			out = append(out, i)
		}
	}
	return out
}

// resolveShedPlay は第2フェーズの 1 枚を処理する。
func (s *Skitgubbe) resolveShedPlay(player int, card *Card) {
	s.pile = append(s.pile, card)
	s.addLog(player, "play", "beats the pile", []*Card{card})

	if s.GetPlayer(player).GetCardsSize() == 0 {
		s.finished[player] = true
		s.addLog(player, "out", "is out of cards", nil)
	}

	// 場の枚数が残っている人数に達したら、そのトリックは流す (avstick)。
	if len(s.pile) >= s.activeCount()+1 {
		s.pile = nil
		s.pileLeader = s.nextActive(player)
		s.currentIdx = s.pileLeader
		s.checkShedEnd()
		return
	}
	s.currentIdx = s.nextActive(player)
	s.checkShedEnd()
}

// PickUp は player が場の札を引き取る (上回れないとき)。
func (s *Skitgubbe) PickUp(player int) error {
	if s.gameEndFlag || s.phase != SkitgubbePhaseShed {
		return fmt.Errorf("nothing to pick up")
	}
	if player != s.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	if len(s.pile) == 0 {
		return fmt.Errorf("the pile is empty; you must play")
	}
	if len(s.GetValidPlayIndices(player)) > 0 {
		return fmt.Errorf("you can beat the pile, so you may not pick it up")
	}

	p := s.GetPlayer(player)
	for _, c := range s.pile {
		p.AddCard(c)
	}
	s.addLog(player, "pickup", fmt.Sprintf("picks up %d card(s)", len(s.pile)), s.pile)
	s.pile = nil
	s.finished[player] = false

	s.pileLeader = s.nextActive(player)
	s.currentIdx = s.pileLeader
	s.checkShedEnd()
	return nil
}

// nextActive は idx の次の、まだ手札を持つプレイヤーを返す。
func (s *Skitgubbe) nextActive(idx int) int {
	n := len(s.players)
	for i := 1; i <= n; i++ {
		j := (idx + i) % n
		if !s.finished[j] {
			return j
		}
	}
	return -1
}

// activeCount はまだ手札を持つ人数を返す。
func (s *Skitgubbe) activeCount() int {
	c := 0
	for i := range s.players {
		if !s.finished[i] {
			c++
		}
	}
	return c
}

// checkShedEnd は第2フェーズの終了条件を見る。最後の 1 人が敗者。
func (s *Skitgubbe) checkShedEnd() {
	if s.phase != SkitgubbePhaseShed {
		return
	}
	if s.activeCount() > 1 {
		if s.currentIdx < 0 || s.finished[s.currentIdx] {
			s.currentIdx = s.nextActive(s.currentIdx)
		}
		return
	}
	s.gameEndFlag = true
	s.phase = SkitgubbePhaseGameEnd
	s.currentIdx = -1
	for i := range s.players {
		if !s.finished[i] {
			s.loserIdx = i
			break
		}
	}
	s.addLog(s.loserIdx, "game", "is the Skitgubbe", nil)
}

// addLog は棋譜へ 1 行追加する。
func (s *Skitgubbe) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.actionLog = append(s.actionLog, &ActionLogEntry{
		TurnNumber: len(s.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// ---- CPU ----

// SkitgubbeCpuAction は CPU が選んだ手。
type SkitgubbeCpuAction struct {
	// PickUp が真なら場の札を引き取る (第2フェーズのみ)。
	PickUp bool
	// HandIdx は出す札の手札添字 (引き取るときは -1)。
	HandIdx int
}

// SkitgubbeCpuDecide は idx の CPU が取る手を決める。
//
// 第1フェーズは「勝てるなら最弱の勝てる札、勝てないなら最弱を捨てる」。取った札は
// 第2フェーズの手札になるので、**取りすぎない**方がよい —— ただし勝った側が
// 次にリードするので、主導権を握る価値もある。ここでは素直に勝ちに行く。
//
// 第2フェーズは「上回れる最弱の札、無ければ引き取る」。
func (s *Skitgubbe) SkitgubbeCpuDecide(idx int) SkitgubbeCpuAction {
	p := s.GetPlayer(idx)
	if p == nil || p.GetCardsSize() == 0 {
		return SkitgubbeCpuAction{HandIdx: -1}
	}
	valid := s.GetValidPlayIndices(idx)
	if len(valid) == 0 {
		return SkitgubbeCpuAction{PickUp: true, HandIdx: -1}
	}

	if s.phase == SkitgubbePhaseShed {
		return SkitgubbeCpuAction{HandIdx: s.weakestOf(idx, valid)}
	}

	// 第1フェーズ。相手が出していれば上回れる最弱の札、そうでなければ最弱。
	if len(s.duel) > 0 && idx != s.duelLeader {
		top := s.duel[len(s.duel)-1]
		best, bestRank := -1, 0
		for _, i := range valid {
			r := SkitgubbeRankOrder(p.GetCard(i))
			if r > SkitgubbeRankOrder(top) && (best == -1 || r < bestRank) {
				best, bestRank = i, r
			}
		}
		if best >= 0 {
			return SkitgubbeCpuAction{HandIdx: best}
		}
	}
	return SkitgubbeCpuAction{HandIdx: s.weakestOf(idx, valid)}
}

// weakestOf は候補のうち最も弱い札の添字を返す。
func (s *Skitgubbe) weakestOf(idx int, candidates []int) int {
	p := s.GetPlayer(idx)
	best, bestRank := candidates[0], 0
	for _, i := range candidates {
		if r := SkitgubbeRankOrder(p.GetCard(i)); bestRank == 0 || r < bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (s *Skitgubbe) GetPlayers() []*SkitgubbePlayer { return s.players }

// GetPlayer は idx のプレイヤーを返す。範囲外は nil。
func (s *Skitgubbe) GetPlayer(idx int) *SkitgubbePlayer {
	if idx < 0 || idx >= len(s.players) {
		return nil
	}
	return s.players[idx]
}

// GetPhase は現在のフェーズを返す。
func (s *Skitgubbe) GetPhase() SkitgubbePhase { return s.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (s *Skitgubbe) GetCurrentPlayerIdx() int { return s.currentIdx }

// GetStockCount は山札の残り枚数を返す。
func (s *Skitgubbe) GetStockCount() int { return len(s.stock) }

// GetTrumpSuit は切札スートを返す (未決は -1)。
func (s *Skitgubbe) GetTrumpSuit() int { return s.trumpSuit }

// GetDuel は第1フェーズで場に出ている札を返す。stunsa が続くと積み上がる。
func (s *Skitgubbe) GetDuel() []*Card { return s.duel }

// GetDuelLeader は第1フェーズのリード側を返す。
func (s *Skitgubbe) GetDuelLeader() int { return s.duelLeader }

// GetPile は第2フェーズで場に出ている札を返す。
func (s *Skitgubbe) GetPile() []*Card { return s.pile }

// GetCollectedCount は idx が第1フェーズで集めた枚数を返す。
func (s *Skitgubbe) GetCollectedCount(idx int) int {
	if idx < 0 || idx >= len(s.collected) {
		return 0
	}
	return len(s.collected[idx])
}

// IsFinished は idx が手札を出し切ったかを返す。
func (s *Skitgubbe) IsFinished(idx int) bool {
	return idx >= 0 && idx < len(s.finished) && s.finished[idx]
}

// GetGameEndFlag は終局しているかを返す。
func (s *Skitgubbe) GetGameEndFlag() bool { return s.gameEndFlag }

// GetLoserIdx は敗者 (Skitgubbe) の添字を返す。未確定は -1。
func (s *Skitgubbe) GetLoserIdx() int { return s.loserIdx }

// GetConfig はゲーム設定を返す。
func (s *Skitgubbe) GetConfig() SkitgubbeConfig { return s.config }

// SetConfig はゲーム設定を差し替える。
func (s *Skitgubbe) SetConfig(c SkitgubbeConfig) { s.config = c }

// GetActionLog は棋譜を返す。
func (s *Skitgubbe) GetActionLog() []*ActionLogEntry { return s.actionLog }

// SetTrumpSuitForTest は切札を設定する (テスト用)。
func (s *Skitgubbe) SetTrumpSuitForTest(suit int) { s.trumpSuit = suit }

// SetPhaseForTest はフェーズを設定する (テスト用)。
func (s *Skitgubbe) SetPhaseForTest(p SkitgubbePhase) { s.phase = p }

// SetPileForTest は第2フェーズの場札を設定する (テスト用)。
func (s *Skitgubbe) SetPileForTest(cards []*Card) { s.pile = cards }

// SetStockForTest は山札を設定する (テスト用)。
func (s *Skitgubbe) SetStockForTest(cards []*Card) { s.stock = cards }

// SetCurrentPlayerForTest は手番を設定する (テスト用)。
func (s *Skitgubbe) SetCurrentPlayerForTest(idx int) { s.currentIdx = idx }

// ---- JSON ----

// skitgubbeJSON は KV のワイヤ形式。Worker は毎リクエストここから組み直すので、
// ここに無いものは次のリクエストでは存在しない。
type skitgubbeJSON struct {
	Players     []*SkitgubbePlayer `json:"pl"`
	Config      SkitgubbeConfig    `json:"cf"`
	Phase       SkitgubbePhase     `json:"ph"`
	Stock       []*Card            `json:"st"`
	TrumpSuit   int                `json:"ts"`
	Duel        []*Card            `json:"du"`
	DuelLeader  int                `json:"dl"`
	CurrentIdx  int                `json:"ci"`
	Collected   [][]*Card          `json:"co"`
	Pile        []*Card            `json:"pi"`
	PileLeader  int                `json:"pk"`
	Finished    []bool             `json:"fi"`
	LoserIdx    int                `json:"lo"`
	GameEndFlag bool               `json:"ge"`
	ActionLog   []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (s *Skitgubbe) MarshalJSON() ([]byte, error) {
	return json.Marshal(skitgubbeJSON{
		Players:     s.players,
		Config:      s.config,
		Phase:       s.phase,
		Stock:       s.stock,
		TrumpSuit:   s.trumpSuit,
		Duel:        s.duel,
		DuelLeader:  s.duelLeader,
		CurrentIdx:  s.currentIdx,
		Collected:   s.collected,
		Pile:        s.pile,
		PileLeader:  s.pileLeader,
		Finished:    s.finished,
		LoserIdx:    s.loserIdx,
		GameEndFlag: s.gameEndFlag,
		ActionLog:   s.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// Worker はこれを KV の未検証バイト列に対して毎リクエスト実行する。添字は信用せず
// 丸め、スライスは人数ぶんに揃える。
func (s *Skitgubbe) UnmarshalJSON(data []byte) error {
	var j skitgubbeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) == 0 {
		return fmt.Errorf("skitgubbe: no players in snapshot")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("skitgubbe: %w", err)
	}
	s.players = j.Players
	s.config = j.Config
	s.phase = j.Phase
	s.stock = j.Stock
	s.trumpSuit = j.TrumpSuit
	s.duel = j.Duel
	s.pile = j.Pile
	s.gameEndFlag = j.GameEndFlag
	s.actionLog = j.ActionLog

	n := len(s.players)
	s.currentIdx = skitgubbeClampSeat(j.CurrentIdx, n)
	s.loserIdx = skitgubbeClampSeat(j.LoserIdx, n)
	s.duelLeader = skitgubbeClampSeat(j.DuelLeader, n)
	if s.duelLeader < 0 {
		s.duelLeader = 0
	}
	s.pileLeader = skitgubbeClampSeat(j.PileLeader, n)
	if s.pileLeader < 0 {
		s.pileLeader = 0
	}

	s.finished = make([]bool, n)
	copy(s.finished, j.Finished)
	s.collected = make([][]*Card, n)
	for i := range s.collected {
		if i < len(j.Collected) && j.Collected[i] != nil {
			s.collected[i] = j.Collected[i]
		} else {
			s.collected[i] = make([]*Card, 0, SkitgubbeDeckSize)
		}
	}
	return nil
}

// skitgubbeClampSeat は範囲外のプレイヤー添字を -1 に丸める。
func skitgubbeClampSeat(idx, n int) int {
	if idx < 0 || idx >= n {
		return -1
	}
	return idx
}
