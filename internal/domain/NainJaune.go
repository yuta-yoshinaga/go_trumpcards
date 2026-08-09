//go:build !js || !wasm || extra3

// Package domain — ル・ナン・ジョーヌ (Le Nain Jaune / Yellow Dwarf) の
// ドメインモデル。
//
// フランスのストップス系。52 枚、専用盤の 5 区画。
//
// # issue #4380 の仕様案との相違
//
// 5 区画という骨格は合っているが、**区画の中身・アンティ・並べ方・精算の
// すべてに誤りがある**。
//
//   - issue は区画を「♦7 / **♠10** / **♥Q** / ♣J / **シーケンス**」とするが、
//     正しくは **♦10 / ♣J / ♠Q / ♥K / ♦7**。♦10 を「♠10」、♠Q を「♥Q」と
//     スートを取り違えており、**♥K が抜けている**。そして**「シーケンス」区画は
//     存在しない**
//   - issue は「5 区画に事前アンティを積む」としか書かないが、枚数は区画ごとに
//     決まっている。**♦10 に 1、♣J に 2、♠Q に 3、♥K に 4、♦7 に 5**
//   - issue は「シーケンス区画は各スートの最後の 1 枚 (K など) を出した者が
//     獲得する」とするが、その区画が無い。**K は「並びを止めて好きな札で次を
//     始める」札**である
//   - issue は「最小ランクから連続するカードを場に出していく」とするが、最初の
//     リードは**好きな札**。そして以降は **スートを問わない**
//     («Suits are irrelevant and so there is no requirement to follow suit")
//   - issue は「残り手札**枚数**分を支払う」とするが、**点数**で払う。
//     A=1 / 2〜10 は額面 / J・Q・K は各 10
//   - issue が触れていない: **talon (配り切らない残り) がある**。そして
//     **取られなかった区画は次のディールへ持ち越される**
//
// # Pope Joan との違い
//
// 同じストップス系で盤も似ているが、**並びの作り方が正反対**である。
// Pope Joan は同じスートの次に高い札 (♦8 を抜いてあるので ♦7 で必ず止まる)、
// こちらはスート無関係にランクが 1 つずつ上がり、**K で止まる** (次が無い)。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// NainJaunePlayerCnt はプレイヤー数 (4 人。原典は 3〜8 人)。
const NainJaunePlayerCnt = 4

// NainJauneHandSize は 4 人のときの配札枚数。
//
// 原典は人数で変える (3 人なら 15 枚、8 人なら 6 枚)。4 人では 12 枚配り、
// 残り 4 枚が talon になる。
const NainJauneHandSize = 12

// NainJaunePhase はゲームフェーズ。
type NainJaunePhase int

// Nain Jaune のフェーズ定数
const (
	// NainJaunePhasePlay 手番進行中
	NainJaunePhasePlay NainJaunePhase = iota
	// NainJaunePhaseDealEnd 1 ディール終了
	NainJaunePhaseDealEnd
	// NainJaunePhaseGameEnd 決着
	NainJaunePhaseGameEnd
)

// NainJauneAward は区画が誰にいくら渡ったかの記録。
type NainJauneAward struct {
	Box    NainJauneBox
	Player int
	Chips  int
}

// newNainJauneDeck は 52 枚を生成する。
func newNainJauneDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, 52)
	for _, s := range suits {
		for v := 1; v <= 13; v++ {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// nainJauneShuffle は Fisher-Yates。
func nainJauneShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// NainJaune はル・ナン・ジョーヌのゲームクラス。
type NainJaune struct {
	players []*NainJaunePlayer
	config  NainJauneConfig
	phase   NainJaunePhase
	board   NainJauneBoard

	// talon は配り切らなかった残り。誰も使わない。
	talon []*Card
	// awards はこのディールで区画が動いた記録。
	awards []*NainJauneAward

	currentIdx int
	dealerIdx  int
	dealNo     int

	// runRank は今の並びの最高ランク。0 なら好きな札で始められる。
	// **スートは持たない。**この game は無関係。
	runRank int
	// playedPile は場に出た札。
	playedPile []*Card

	dealWinner  int
	gameEndFlag bool
	winnerIdx   int
	actionLog   []*ActionLogEntry
}

// NewNainJaune はコンストラクタ。
func NewNainJaune(players []*NainJaunePlayer, config NainJauneConfig) *NainJaune {
	return &NainJaune{
		players:    players,
		config:     config,
		dealWinner: -1,
		winnerIdx:  -1,
	}
}

// NewDefaultNainJaune は標準の 4 人セットアップを返す。
func NewDefaultNainJaune() *NainJaune {
	players := make([]*NainJaunePlayer, 0, NainJaunePlayerCnt)
	players = append(players, NewNainJaunePlayer(true))
	for range NainJaunePlayerCnt - 1 {
		players = append(players, NewNainJaunePlayer(false))
	}
	return NewNainJaune(players, DefaultNainJauneConfig())
}

// Reset はゲーム全体を初期化する。
func (n *NainJaune) Reset() {
	n.board = NainJauneBoard{}
	n.dealerIdx = 0
	n.dealNo = 0
	n.gameEndFlag = false
	n.winnerIdx = -1
	n.actionLog = nil
	for _, p := range n.players {
		p.AddChips(-p.GetChips())
	}
	n.dealRound()
}

// dealRound は 1 ディールを配る。
func (n *NainJaune) dealRound() {
	for _, p := range n.players {
		p.ResetDeal()
	}
	n.awards = nil
	n.runRank = 0
	n.playedPile = nil
	n.dealWinner = -1

	// **全員が区画ごとに決まった枚数を置く。**取られなかったぶんは残っている
	// ので、その上に積み増される。
	n.board.Ante(len(n.players))
	for _, p := range n.players {
		p.AddChips(-NainJauneAnteTotal)
	}

	deck := newNainJauneDeck()
	nainJauneShuffle(deck)
	pos := 0
	for range NainJauneHandSize {
		for _, p := range n.players {
			p.AddCard(deck[pos])
			pos++
		}
	}
	// **talon。**配り切らない残りで、誰も使わない。
	n.talon = append([]*Card(nil), deck[pos:]...)

	n.currentIdx = (n.dealerIdx + 1) % len(n.players)
	n.phase = NainJaunePhasePlay
	n.addLog(-1, "deal", fmt.Sprintf("%d cards left in the talon", len(n.talon)), nil)
}

// Play は手札 1 枚を出す。
//
// 並びが止まっていれば好きな札で始められる。並びの途中なら**ランクが 1 つ上の
// 札**でなければならない -- **スートは問わない**。
func (n *NainJaune) Play(player, handIdx int) error {
	if n.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if n.phase != NainJaunePhasePlay {
		return fmt.Errorf("the deal is not in progress")
	}
	if player != n.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	p := n.GetPlayer(player)
	if p == nil || handIdx < 0 || handIdx >= p.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}
	card := p.GetCard(handIdx)
	if !n.playable(card) {
		return fmt.Errorf("you must play the next rank up, of any suit")
	}

	p.RemoveCard(handIdx)
	n.playedPile = append(n.playedPile, card)
	n.runRank = card.GetValue()
	n.addLog(player, "play", "plays a card", []*Card{card})
	n.payForCard(player, card)

	if p.GetCardsSize() == 0 {
		n.finishDeal(player)
		return nil
	}
	// **K は並びを止める。**出した本人が好きな札で次を始める。
	if card.GetValue() == 13 {
		n.runRank = 0
		n.addLog(player, "stop", "a king ends the run; the same player leads again", nil)
		return nil
	}
	n.advance(player)
	return nil
}

// playable は card が今出せるかを返す。**スートは見ない。**
func (n *NainJaune) playable(card *Card) bool {
	if card == nil {
		return false
	}
	if n.runRank == 0 {
		return true
	}
	return card.GetValue() == n.runRank+1
}

// NainJauneValidPlays は player が今出せる手札インデックスを返す。
//
// **判定は playable をそのまま呼ぶ。**規則を書き写すと、示した手が拒否される
// ようになる (#4935)。手番でない場合は nil。
func (n *NainJaune) NainJauneValidPlays(player int) []int {
	if player != n.currentIdx {
		return nil
	}
	pl := n.GetPlayer(player)
	if pl == nil {
		return nil
	}
	out := make([]int, 0, pl.GetCardsSize())
	for i := range pl.GetCardsSize() {
		if n.playable(pl.GetCard(i)) {
			out = append(out, i)
		}
	}
	return out
}

// payForCard は出した札が区画を取るかを判定する。
func (n *NainJaune) payForCard(player int, card *Card) {
	box, ok := NainJauneBoxForCard(card)
	if !ok {
		return
	}
	chips := n.board.Take(box)
	if chips == 0 {
		return
	}
	n.players[player].AddChips(chips)
	n.awards = append(n.awards, &NainJauneAward{Box: box, Player: player, Chips: chips})
	n.addLog(player, "award", fmt.Sprintf("takes the %s box (%d)", box, chips), []*Card{card})
}

// advance は次に出せる人へ手番を回す。誰も次を持っていなければ、**最後に札を
// 出した人**が好きな札から再開する。
func (n *NainJaune) advance(lastPlayer int) {
	for i := 1; i <= len(n.players); i++ {
		s := (lastPlayer + i) % len(n.players)
		if n.hasNextCard(s) {
			n.currentIdx = s
			return
		}
	}
	n.runRank = 0
	n.currentIdx = lastPlayer
	n.addLog(lastPlayer, "stop", "nobody can continue; the run restarts here", nil)
}

// hasNextCard は seat が並びの続きを持っているかを返す。
func (n *NainJaune) hasNextCard(seat int) bool {
	if n.runRank == 0 {
		return false
	}
	p := n.GetPlayer(seat)
	if p == nil {
		return false
	}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c != nil && c.GetValue() == n.runRank+1 {
			return true
		}
	}
	return false
}

// finishDeal はディールを精算する。
//
// 出し切った人が、他家から**残り札の点数ぶん**を受け取る。**枚数ではない。**
func (n *NainJaune) finishDeal(winner int) {
	paid := 0
	for i, p := range n.players {
		if i == winner {
			continue
		}
		pts := 0
		for j := range p.GetCardsSize() {
			pts += NainJaunePoints(p.GetCard(j))
		}
		if pts == 0 {
			continue
		}
		p.AddChips(-pts)
		paid += pts
	}
	n.players[winner].AddChips(paid)
	n.dealWinner = winner
	n.addLog(winner, "deal_end", fmt.Sprintf("goes out and collects %d points", paid), nil)

	n.dealNo++
	n.phase = NainJaunePhaseDealEnd
	if n.dealNo >= n.config.TargetDeals {
		n.finishGame()
	}
}

// finishGame は最終集計する。チップが最も多い人の勝ち。
func (n *NainJaune) finishGame() {
	best := 0
	for i := 1; i < len(n.players); i++ {
		if n.players[i].GetChips() > n.players[best].GetChips() {
			best = i
		}
	}
	n.winnerIdx = best
	n.gameEndFlag = true
	n.phase = NainJaunePhaseGameEnd
	n.addLog(best, "game_end", "finishes with the most chips", nil)
}

// NextDeal は次のディールを配る。
func (n *NainJaune) NextDeal() error {
	if n.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if n.phase != NainJaunePhaseDealEnd {
		return fmt.Errorf("the deal is still in progress")
	}
	n.dealerIdx = (n.dealerIdx + 1) % len(n.players)
	n.dealRound()
	return nil
}

// NainJauneCpuDecide は idx の CPU が出す手札の添字を返す (-1: 出せない)。
//
// 区画を取れる札があればそれを優先する。無ければ最も低い札から出す。
func (n *NainJaune) NainJauneCpuDecide(idx int) int {
	p := n.GetPlayer(idx)
	if p == nil {
		return -1
	}
	best, bestRank := -1, 1<<30
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if !n.playable(c) {
			continue
		}
		// 区画つきの札は最優先。♦7 は 5 枚積まれているので取り逃したくない。
		if _, ok := NainJauneBoxForCard(c); ok && n.board.Get(mustBox(c)) > 0 {
			return i
		}
		if c.GetValue() < bestRank {
			best, bestRank = i, c.GetValue()
		}
	}
	return best
}

// mustBox は NainJauneBoxForCard が真を返す札の区画を取り出す。
func mustBox(c *Card) NainJauneBox {
	box, _ := NainJauneBoxForCard(c)
	return box
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (n *NainJaune) GetPlayers() []*NainJaunePlayer { return n.players }

// GetPlayer は idx のプレイヤーを返す。
func (n *NainJaune) GetPlayer(idx int) *NainJaunePlayer {
	return getPlayer(n.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (n *NainJaune) GetPhase() NainJaunePhase { return n.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (n *NainJaune) GetCurrentPlayerIdx() int { return n.currentIdx }

// GetBoard は 5 区画の残高を返す。
func (n *NainJaune) GetBoard() NainJauneBoard { return n.board }

// GetTalonCount は talon の枚数を返す。
func (n *NainJaune) GetTalonCount() int { return len(n.talon) }

// GetAwards はこのディールで区画が動いた記録を返す。
func (n *NainJaune) GetAwards() []*NainJauneAward { return n.awards }

// GetPlayedPile は場に出た札を返す。
func (n *NainJaune) GetPlayedPile() []*Card { return n.playedPile }

// GetRunRank は今の並びの最高ランクを返す (0: 好きな札で始められる)。
func (n *NainJaune) GetRunRank() int { return n.runRank }

// GetDealNumber は完了したディール数を返す。
func (n *NainJaune) GetDealNumber() int { return n.dealNo }

// GetDealWinner は直近ディールで出し切った席を返す (-1: なし)。
func (n *NainJaune) GetDealWinner() int { return n.dealWinner }

// GetGameEndFlag は決着しているかを返す。
func (n *NainJaune) GetGameEndFlag() bool { return n.gameEndFlag }

// GetWinnerIdx は勝者の添字を返す (-1: 未決着)。
func (n *NainJaune) GetWinnerIdx() int { return n.winnerIdx }

// GetConfig はゲーム設定を返す。
func (n *NainJaune) GetConfig() NainJauneConfig { return n.config }

// SetConfig はゲーム設定をセットする。
func (n *NainJaune) SetConfig(c NainJauneConfig) { n.config = c }

// GetActionLog は棋譜を返す。
func (n *NainJaune) GetActionLog() []*ActionLogEntry { return n.actionLog }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (n *NainJaune) SetPhaseForTest(p NainJaunePhase) { n.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (n *NainJaune) SetCurrentPlayerForTest(idx int) { n.currentIdx = idx }

// SetBoardForTest はテスト用に盤を差し替える。
func (n *NainJaune) SetBoardForTest(b NainJauneBoard) { n.board = b }

// SetRunRankForTest はテスト用に並びの状態を差し替える。
func (n *NainJaune) SetRunRankForTest(rank int) { n.runRank = rank }

// SetDealNumberForTest はテスト用にディール数を差し替える。
func (n *NainJaune) SetDealNumberForTest(d int) { n.dealNo = d }

// addLog は棋譜に 1 件追加する。
func (n *NainJaune) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	n.actionLog = append(n.actionLog, &ActionLogEntry{
		TurnNumber: len(n.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// nainJauneJSON is the JSON wire format for NainJaune.
type nainJauneJSON struct {
	Players    []*NainJaunePlayer `json:"pl"`
	Config     NainJauneConfig    `json:"cfg"`
	Phase      NainJaunePhase     `json:"ph"`
	Board      NainJauneBoard     `json:"bd"`
	Talon      []*Card            `json:"tl"`
	Awards     []*NainJauneAward  `json:"aw"`
	Current    int                `json:"cur"`
	Dealer     int                `json:"dl"`
	DealNo     int                `json:"dn"`
	RunRank    int                `json:"rr"`
	PlayedPile []*Card            `json:"pi"`
	DealWinner int                `json:"dw"`
	GameEnd    bool               `json:"ge"`
	WinnerIdx  int                `json:"wi"`
	ActionLog  []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (n *NainJaune) MarshalJSON() ([]byte, error) {
	return json.Marshal(nainJauneJSON{
		Players: n.players, Config: n.config, Phase: n.phase, Board: n.board,
		Talon: n.talon, Awards: n.awards, Current: n.currentIdx, Dealer: n.dealerIdx,
		DealNo: n.dealNo, RunRank: n.runRank, PlayedPile: n.playedPile,
		DealWinner: n.dealWinner, GameEnd: n.gameEndFlag, WinnerIdx: n.winnerIdx,
		ActionLog: n.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// KV から戻る生バイト列は信用できないので、席数に合わせて詰め直し、設定を
// 検証する。**盤のチップは持ち越しそのもの**なので、そのまま復元する。
func (n *NainJaune) UnmarshalJSON(data []byte) error {
	var raw nainJauneJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Players) != NainJaunePlayerCnt {
		return fmt.Errorf("expected %d players, got %d", NainJaunePlayerCnt, len(raw.Players))
	}
	if err := raw.Config.Validate(); err != nil {
		return err
	}
	if raw.Phase < NainJaunePhasePlay || raw.Phase > NainJaunePhaseGameEnd {
		return fmt.Errorf("unknown phase: %d", raw.Phase)
	}
	// **runRank が範囲外だと出せる札が無くなって固まる。**0 は「好きな札で
	// 始められる」という意味を持つので潰さない。
	//
	// 上限は 13 ではなく **12**。K (13) を出した瞬間に Play が runRank を 0 へ
	// 戻すので、13 は対局中に観測されない値である。復元でだけ入り込むと 14 の
	// 札は存在しないため playable() が全員永久に偽になり、そのディールが詰む。
	if raw.RunRank < 0 || raw.RunRank > 12 {
		return fmt.Errorf("bad run rank: %d", raw.RunRank)
	}

	n.players = raw.Players
	n.config = raw.Config
	n.phase = raw.Phase
	n.board = raw.Board
	n.talon = raw.Talon
	n.dealNo = raw.DealNo
	n.runRank = raw.RunRank
	n.playedPile = raw.PlayedPile
	n.gameEndFlag = raw.GameEnd
	n.actionLog = raw.ActionLog

	n.currentIdx = clampNainJauneIdx(raw.Current, len(n.players))
	n.dealerIdx = clampNainJauneIdx(raw.Dealer, len(n.players))
	n.dealWinner = clampNainJauneSeatOrNone(raw.DealWinner, len(n.players))
	n.winnerIdx = clampNainJauneSeatOrNone(raw.WinnerIdx, len(n.players))

	n.awards = make([]*NainJauneAward, 0, len(raw.Awards))
	for _, a := range raw.Awards {
		if a == nil || a.Box < 0 || a.Box >= NainJauneBoxCount {
			continue
		}
		if a.Player < 0 || a.Player >= len(n.players) {
			continue
		}
		n.awards = append(n.awards, a)
	}
	return nil
}

// clampNainJauneIdx は席番号を 0..n-1 に収める。
func clampNainJauneIdx(idx, cnt int) int {
	if idx < 0 || idx >= cnt {
		return 0
	}
	return idx
}

// clampNainJauneSeatOrNone は席番号を -1..n-1 に収める。
func clampNainJauneSeatOrNone(idx, cnt int) int {
	if idx < -1 || idx >= cnt {
		return -1
	}
	return idx
}
