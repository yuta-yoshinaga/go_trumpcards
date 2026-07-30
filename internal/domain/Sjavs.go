//go:build !js || !wasm || extra2

// Package domain — シャウス (Sjavs) のドメインモデル。
//
// フェロー諸島の国民的トランプゲーム。32 枚、2 対 2 の固定チーム、シャフコップ系の
// 常時切札を持つトリックテイキング。実装は pagat.com の記述に従う。
//
// # issue #4403 の仕様案との相違
//
//   - issue は常時切札を「♣Q ＞ ♠Q」の **2 枚**とするが、実際は **6 枚**:
//     ♣Q ＞ ♠Q ＞ ♣J ＞ ♠J ＞ ♥J ＞ ♦J。その下に切札スートの A,K,(Q),10,9,8,7
//   - issue はビッドを「**目標点数**の宣言」とするが、実際は「**自分の最長切札
//     スートの枚数**」の申告。5 枚に満たなければ発声できず、全員が届かなければ配り直し
//   - issue は「ビッド確定**後**に切札スートを宣言」とするが、同じ長さの候補に
//     ♣が含まれるなら**ビッドの時点で♣と宣言しなければならない**（同長では♣が勝つ）
//   - issue は「結果に応じてクロス(×)を 1 つ消す／先に全部消した側の勝ち」とするが、
//     1 ラバーは **24 点を 24 から引き算**していき、先に 0 以下へ達した側が勝つ。
//     クロスは**ラバーの勝利数の記録**であって 1 ハンドごとに消すものではない
//
// # 常時切札が 6 枚であることは切札の総数で検算できる
//
//   - 赤が切札: 常時 6 + そのスートの A,K,Q,10,9,8,7 の 7 = **13 枚**
//   - 黒が切札: そのスートの Q と J は既に常時切札なので 6 + 6 = **12 枚**
//
// pagat の «13 trumps if red, 12 if black» と一致する。常時切札を 2 枚だと
// 9 枚/8 枚にしかならず、この数字が合わない。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// SjavsPlayerCnt はプレイヤー数 (4 人固定、2 対 2)。
const SjavsPlayerCnt = 4

// SjavsHandSize は 1 人あたりの配札枚数 (32 / 4)。
const SjavsHandSize = 8

// SjavsDeckSize は 32 枚 (52 枚から 2〜6 を抜く)。
const SjavsDeckSize = 32

// SjavsTotalPoints はデッキの総点。A=11, 10=10, K=4, Q=3, J=2 の各 4 枚。
const SjavsTotalPoints = 120

// SjavsMinBid はビッドできる最短の切札スート長。これに満たなければ配り直し。
const SjavsMinBid = 5

// SjavsRubber は 1 ラバーの点数。ここから引き算していく。
const SjavsRubber = 24

// SjavsPhase はゲームフェーズ。
type SjavsPhase int

// Sjavs のフェーズ定数
const (
	// SjavsPhaseBid ビッド中
	SjavsPhaseBid SjavsPhase = iota
	// SjavsPhasePlay トリック進行中
	SjavsPhasePlay
	// SjavsPhaseHandEnd 1 ハンド終了 (精算済み)
	SjavsPhaseHandEnd
	// SjavsPhaseGameEnd ラバー決着
	SjavsPhaseGameEnd
)

// SjavsCardPoints は 1 枚の点数を返す。
func SjavsCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1:
		return 11
	case 10:
		return 10
	case 13:
		return 4
	case 12:
		return 3
	case 11:
		return 2
	default:
		return 0
	}
}

// sjavsPermanentTrumps は常時切札を強い順に並べたもの。
//
// **切札スートが何であってもこの 6 枚が最強**で、しかもこの順序は動かない。
// issue が 2 枚としているのはここ。
var sjavsPermanentTrumps = [][2]int{
	{CardDesignClover, 12},  // ♣Q
	{CardDesignSpade, 12},   // ♠Q
	{CardDesignClover, 11},  // ♣J
	{CardDesignSpade, 11},   // ♠J
	{CardDesignHeart, 11},   // ♥J
	{CardDesignDiamond, 11}, // ♦J
}

// sjavsPermanentRank は常時切札なら 0..5 の順位を、そうでなければ -1 を返す。
func sjavsPermanentRank(c *Card) int {
	if c == nil {
		return -1
	}
	for i, pt := range sjavsPermanentTrumps {
		if c.GetDesign() == pt[0] && c.GetValue() == pt[1] {
			return i
		}
	}
	return -1
}

// SjavsIsTrump は c が切札かを返す。常時切札は切札スートに関係なく切札。
func SjavsIsTrump(c *Card, trumpSuit int) bool {
	if c == nil {
		return false
	}
	if sjavsPermanentRank(c) >= 0 {
		return true
	}
	return c.GetDesign() == trumpSuit
}

// sjavsPlainRank は切札以外の平札の強さ (A が最強、7 が最弱)。
func sjavsPlainRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return 14 // A
	}
	return c.GetValue()
}

// sjavsTrumpStrength は切札同士の比較用の強さ。大きいほど強い。
//
// 常時切札 6 枚が最上位で、その下に切札スートの平札が来る。
func sjavsTrumpStrength(c *Card) int {
	if r := sjavsPermanentRank(c); r >= 0 {
		// 6 枚を最上位に置く。0 が最強なので反転する。
		return 100 + (len(sjavsPermanentTrumps) - r)
	}
	return sjavsPlainRank(c)
}

// SjavsTrumpCount は trumpSuit を切札にしたときの切札の総枚数を返す。
//
// 赤なら 13、黒なら 12。黒はスートの Q と J が既に常時切札に含まれるぶん減る。
func SjavsTrumpCount(trumpSuit int) int {
	n := len(sjavsPermanentTrumps)
	for v := 1; v <= 13; v++ {
		if v >= 2 && v <= 6 {
			continue // 32 枚デッキに無い
		}
		c := NewCard(trumpSuit, v, true)
		if sjavsPermanentRank(c) >= 0 {
			continue // 既に常時切札として数えた
		}
		n++
	}
	return n
}

// newSjavsDeck は 32 枚を生成する (シャッフル前)。2〜6 を抜いた 7,8,9,10,J,Q,K,A。
func newSjavsDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, SjavsDeckSize)
	for _, s := range suits {
		for _, v := range []int{1, 7, 8, 9, 10, 11, 12, 13} {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// sjavsShuffle は Fisher-Yates。domain の shuffleCards は casino タグのファイルに
// あり extra2 ビルドから見えないため、専用名で持つ。
func sjavsShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// SjavsTeamOf は席の所属チーム (0 or 1) を返す。向かい合わせが味方。
func SjavsTeamOf(idx int) int { return idx % 2 }

// SjavsTrickCard はトリックに出された 1 枚。
type SjavsTrickCard struct {
	PlayerIdx int
	Card      *Card
}

// Sjavs はシャウスのゲームクラス。
type Sjavs struct {
	players []*SjavsPlayer
	config  SjavsConfig
	phase   SjavsPhase

	dealerIdx  int
	currentIdx int

	// ---- ビッド ----
	bids       []int // 席ごとの申告枚数 (0 = パス/未申告)
	bidderIdx  int   // 現在の最高ビッド者 (-1: なし)
	bidLength  int   // その申告枚数
	bidIsClubs bool  // ♣で申告したか (同長では♣が勝つ)
	trumpSuit  int   // 確定した切札スート (-1: 未確定)

	// ---- プレイ ----
	trick     []SjavsTrickCard
	leadIdx   int
	trickNo   int
	tricksWon []int // 席ごとのトリック数
	points    []int // チームごとの獲得点

	// ---- 精算 ----
	// remaining[team] は 24 からの残り。0 以下でラバー勝ち。
	remaining []int
	// carryOver は 60-60 で流れたぶんの上乗せ。次ゲームの価値が +2 になる。
	carryOver  int
	crosses    []int // チームごとのラバー勝利数
	handResult *SjavsHandResult

	gameEndFlag bool
	winnerTeam  int
	actionLog   []*ActionLogEntry
}

// SjavsHandResult は 1 ハンドの精算結果。
type SjavsHandResult struct {
	// DeclarerTeam は切札を宣言した側のチーム。
	DeclarerTeam int
	// DeclarerPoints は宣言側が取った点。
	DeclarerPoints int
	// ScoringTeam は加点を得たチーム (-1: 60-60 で流れた)。
	ScoringTeam int
	// Amount は減算される点数。
	Amount int
	// Vol は全トリック獲得か。
	Vol bool
	// TrumpWasClubs は♣が切札だったか。倍額の根拠。
	TrumpWasClubs bool
}

// NewSjavs はコンストラクタ。
func NewSjavs(players []*SjavsPlayer, config SjavsConfig) *Sjavs {
	return &Sjavs{
		players:    players,
		config:     config,
		trumpSuit:  -1,
		bidderIdx:  -1,
		winnerTeam: -1,
		remaining:  []int{SjavsRubber, SjavsRubber},
		crosses:    []int{0, 0},
	}
}

// NewDefaultSjavs は標準の 4 人セットアップを返す。
func NewDefaultSjavs() *Sjavs {
	players := make([]*SjavsPlayer, 0, SjavsPlayerCnt)
	players = append(players, NewSjavsPlayer(true))
	for range SjavsPlayerCnt - 1 {
		players = append(players, NewSjavsPlayer(false))
	}
	return NewSjavs(players, DefaultSjavsConfig())
}

// Reset はラバーごと初期化する。
func (s *Sjavs) Reset() {
	s.dealerIdx = 0
	s.remaining = []int{SjavsRubber, SjavsRubber}
	s.crosses = []int{0, 0}
	s.carryOver = 0
	s.gameEndFlag = false
	s.winnerTeam = -1
	s.actionLog = nil
	s.dealHand()
}

// dealHand は 1 ハンドを配る。誰も 5 枚に届かなければ**配り直す**。
func (s *Sjavs) dealHand() {
	for range 100 {
		s.phase = SjavsPhaseBid
		s.trumpSuit = -1
		s.bidderIdx = -1
		s.bidLength = 0
		s.bidIsClubs = false
		s.bids = make([]int, len(s.players))
		s.trick = nil
		s.trickNo = 0
		s.tricksWon = make([]int, len(s.players))
		s.points = []int{0, 0}
		s.handResult = nil

		for _, p := range s.players {
			p.ResetGame()
		}
		deck := newSjavsDeck()
		sjavsShuffle(deck)
		for i, c := range deck {
			s.players[i%len(s.players)].AddCard(c)
		}
		s.currentIdx = (s.dealerIdx + 1) % len(s.players)
		s.leadIdx = s.currentIdx
		s.addLog(-1, "deal", "cards dealt", nil)

		// 誰か 1 人でも 5 枚のスートを持っていれば成立する。
		for i := range s.players {
			if s.LongestTrumpLength(i) >= SjavsMinBid {
				return
			}
		}
		s.addLog(-1, "redeal", "nobody can bid five, so the hand is redealt", nil)
	}
}

// suitLengths は player の各スートの「そのスートを切札にしたときの切札枚数」を返す。
//
// **常時切札は全部数える。**♣J は♥を切札にしても切札なので、♥の長さにも入る。
// 素のスート枚数を数えると、実際に持っている切札の枚数より少なく申告してしまう。
func (s *Sjavs) suitLengths(player int) map[int]int {
	out := map[int]int{}
	p := s.GetPlayer(player)
	if p == nil {
		return out
	}
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		n := 0
		for i := range p.GetCardsSize() {
			if SjavsIsTrump(p.GetCard(i), suit) {
				n++
			}
		}
		out[suit] = n
	}
	return out
}

// LongestTrumpLength は player の最長切札スート長を返す。
func (s *Sjavs) LongestTrumpLength(player int) int {
	best := 0
	for _, n := range s.suitLengths(player) {
		if n > best {
			best = n
		}
	}
	return best
}

// SjavsLongestSuits は player の最長スートの一覧を返す (同長が複数ありうる)。
func (s *Sjavs) SjavsLongestSuits(player int) []int {
	lengths := s.suitLengths(player)
	best := 0
	for _, n := range lengths {
		if n > best {
			best = n
		}
	}
	var out []int
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if lengths[suit] == best && best > 0 {
			out = append(out, suit)
		}
	}
	return out
}

// Bid は player が length 枚を申告する。length が 0 ならパス。
//
// 同長では♣が勝つので、**♣が最長候補に含まれるなら♣として申告する**。
func (s *Sjavs) Bid(player, length int) error {
	if s.phase != SjavsPhaseBid {
		return fmt.Errorf("bidding is over")
	}
	if player != s.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	if length != 0 {
		if length < SjavsMinBid {
			return fmt.Errorf("a bid must be at least %d cards", SjavsMinBid)
		}
		if got := s.LongestTrumpLength(player); length > got {
			return fmt.Errorf("you only hold %d cards in your longest trump suit", got)
		}
		clubs := false
		for _, suit := range s.SjavsLongestSuits(player) {
			if suit == CardDesignClover {
				clubs = true
			}
		}
		if !s.beatsStandingBid(length, clubs) {
			return fmt.Errorf("that does not beat the standing bid")
		}
		s.bids[player] = length
		s.bidderIdx = player
		s.bidLength = length
		s.bidIsClubs = clubs
		s.addLog(player, "bid", fmt.Sprintf("bids %d", length), nil)
	} else {
		s.addLog(player, "pass", "passes", nil)
	}

	s.currentIdx = (s.currentIdx + 1) % len(s.players)
	if s.currentIdx == (s.dealerIdx+1)%len(s.players) {
		s.finishBidding()
	}
	return nil
}

// beatsStandingBid は length/clubs が現在の最高ビッドを上回るかを返す。
//
// 長い方が勝ち、同長なら♣だけが勝つ。
func (s *Sjavs) beatsStandingBid(length int, clubs bool) bool {
	if s.bidderIdx < 0 {
		return true
	}
	if length > s.bidLength {
		return true
	}
	if length == s.bidLength && clubs && !s.bidIsClubs {
		return true
	}
	return false
}

// finishBidding はビッドを締めて切札を確定する。
func (s *Sjavs) finishBidding() {
	if s.bidderIdx < 0 {
		// 誰も申告しなかった。dealHand が 5 枚以上を保証しているので通常は
		// 起きないが、全員がパスを選ぶ余地はあるので配り直す。
		s.dealHand()
		return
	}
	// ♣が最長候補に含まれるなら♣が強制される。それ以外は最長のうち先頭。
	suits := s.SjavsLongestSuits(s.bidderIdx)
	chosen := suits[0]
	for _, suit := range suits {
		if suit == CardDesignClover {
			chosen = suit
		}
	}
	s.trumpSuit = chosen
	s.phase = SjavsPhasePlay
	s.currentIdx = (s.dealerIdx + 1) % len(s.players)
	s.leadIdx = s.currentIdx
	s.addLog(s.bidderIdx, "trump", fmt.Sprintf("declares trump suit %d", s.trumpSuit), nil)
}

// GetValidPlayIndices は player が出せる手札の添字を返す。
//
// リードのスートに従う。**切札は独立したスート**として扱うので、♣J は♥が切札
// のとき「♣」ではなく「切札」であり、♣がリードされても追随しない。
func (s *Sjavs) GetValidPlayIndices(player int) []int {
	p := s.GetPlayer(player)
	if p == nil {
		return nil
	}
	all := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if p.GetCard(i) != nil {
			all = append(all, i)
		}
	}
	if len(s.trick) == 0 {
		return all
	}
	lead := s.trick[0].Card
	leadIsTrump := SjavsIsTrump(lead, s.trumpSuit)

	var follow []int
	for _, i := range all {
		c := p.GetCard(i)
		if leadIsTrump {
			if SjavsIsTrump(c, s.trumpSuit) {
				follow = append(follow, i)
			}
			continue
		}
		if !SjavsIsTrump(c, s.trumpSuit) && c.GetDesign() == lead.GetDesign() {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// PlayCard は player が手札 handIdx の札を出す。
func (s *Sjavs) PlayCard(player, handIdx int) error {
	if s.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if s.phase != SjavsPhasePlay {
		return fmt.Errorf("it is not the play phase")
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
	legal := false
	for _, i := range s.GetValidPlayIndices(player) {
		if i == handIdx {
			legal = true
		}
	}
	if !legal {
		return fmt.Errorf("you must follow suit")
	}

	card := p.RemoveCard(handIdx)
	s.trick = append(s.trick, SjavsTrickCard{PlayerIdx: player, Card: card})
	s.addLog(player, "play", "plays a card", []*Card{card})

	if len(s.trick) < len(s.players) {
		s.currentIdx = (s.currentIdx + 1) % len(s.players)
		return nil
	}
	s.resolveTrick()
	return nil
}

// SjavsTrickWinner はトリックの勝者の席を返す。
func SjavsTrickWinner(trick []SjavsTrickCard, trumpSuit int) int {
	if len(trick) == 0 {
		return -1
	}
	best := 0
	for i := 1; i < len(trick); i++ {
		if sjavsBeats(trick[i].Card, trick[best].Card, trumpSuit) {
			best = i
		}
	}
	return trick[best].PlayerIdx
}

// sjavsBeats は challenger が leader を上回るかを返す。
func sjavsBeats(challenger, leader *Card, trumpSuit int) bool {
	cT, lT := SjavsIsTrump(challenger, trumpSuit), SjavsIsTrump(leader, trumpSuit)
	switch {
	case cT && !lT:
		return true
	case !cT && lT:
		return false
	case cT && lT:
		return sjavsTrumpStrength(challenger) > sjavsTrumpStrength(leader)
	case challenger.GetDesign() != leader.GetDesign():
		return false // リードのスートに追随していない = 勝てない
	default:
		return sjavsPlainRank(challenger) > sjavsPlainRank(leader)
	}
}

// resolveTrick はトリックを解決する。
func (s *Sjavs) resolveTrick() {
	winner := SjavsTrickWinner(s.trick, s.trumpSuit)
	pts := 0
	cards := make([]*Card, 0, len(s.trick))
	for _, tc := range s.trick {
		pts += SjavsCardPoints(tc.Card)
		cards = append(cards, tc.Card)
	}
	s.tricksWon[winner]++
	s.points[SjavsTeamOf(winner)] += pts
	s.addLog(winner, "trick", fmt.Sprintf("wins the trick for %d point(s)", pts), cards)

	s.trick = nil
	s.trickNo++
	s.leadIdx = winner
	s.currentIdx = winner

	if s.trickNo >= SjavsHandSize {
		s.settleHand()
	}
}

// settleHand は 1 ハンドを精算する。
//
// 精算表は pagat どおり。**60-60 は加点なしで、次ゲームの価値が +2 される**
// (carryOver)。宣言側が 1 トリックも取れなかった場合だけ、スートに関わらず 16。
func (s *Sjavs) settleHand() {
	declTeam := SjavsTeamOf(s.bidderIdx)
	declPts := s.points[declTeam]
	clubs := s.trumpSuit == CardDesignClover

	declTricks := 0
	for i := range s.players {
		if SjavsTeamOf(i) == declTeam {
			declTricks += s.tricksWon[i]
		}
	}

	res := &SjavsHandResult{
		DeclarerTeam:   declTeam,
		DeclarerPoints: declPts,
		ScoringTeam:    -1,
		TrumpWasClubs:  clubs,
	}

	double := func(n int) int {
		if clubs {
			return n * 2
		}
		return n
	}

	switch {
	case declTricks == SjavsHandSize:
		// **vol だけ♣が倍ではない。**12 → 16 であって 24 ではない。他の行は
		// すべて倍なので、ここも double() を通したくなるが原典と合わなくなる。
		res.Vol = true
		amount := 12
		if clubs {
			amount = 16
		}
		res.ScoringTeam, res.Amount = declTeam, amount
	case declPts >= 90:
		res.ScoringTeam, res.Amount = declTeam, double(4)
	case declPts >= 61:
		res.ScoringTeam, res.Amount = declTeam, double(2)
	case declPts == 60:
		// 引き分け。加点は無く、次ゲームの価値が上がる。
		s.carryOver += 2
		s.addLog(-1, "tie", "60-60: no score, the next game is worth two more", nil)
	case declTricks == 0:
		// スート不問で 16。ここだけ♣の倍額規則が効かない。
		res.ScoringTeam, res.Amount = 1-declTeam, 16
	case declPts >= 31:
		res.ScoringTeam, res.Amount = 1-declTeam, double(4)
	default:
		res.ScoringTeam, res.Amount = 1-declTeam, double(8)
	}

	if res.ScoringTeam >= 0 {
		res.Amount += s.carryOver
		s.carryOver = 0
		s.remaining[res.ScoringTeam] -= res.Amount
		s.addLog(-1, "score", fmt.Sprintf("team %d scores %d", res.ScoringTeam, res.Amount), nil)
	}

	s.handResult = res
	s.phase = SjavsPhaseHandEnd

	for t := range s.remaining {
		if s.remaining[t] <= 0 {
			s.crosses[t]++
			s.winnerTeam = t
			s.gameEndFlag = true
			s.phase = SjavsPhaseGameEnd
			s.addLog(-1, "rubber", fmt.Sprintf("team %d wins the rubber", t), nil)
			return
		}
	}
}

// NextHand は次のハンドを配る。ハンド終了時のみ。
func (s *Sjavs) NextHand() error {
	if s.gameEndFlag {
		return fmt.Errorf("the rubber is over")
	}
	if s.phase != SjavsPhaseHandEnd {
		return fmt.Errorf("the hand is still in progress")
	}
	s.dealerIdx = (s.dealerIdx + 1) % len(s.players)
	s.dealHand()
	return nil
}

// ---- CPU ----

// SjavsCpuAction は CPU が選んだ手。
type SjavsCpuAction struct {
	// BidLength はビッド枚数 (0 = パス)。プレイ中は無視される。
	BidLength int
	// HandIdx は出す札の手札添字 (出せないときは -1)。
	HandIdx int
}

// SjavsCpuDecide は idx の CPU が取る手を決める。
//
// ビッドは「上回れるなら自分の最長を申告する」。長さがそのまま切札の枚数なので、
// 申告できる = 切札が多いということであり、伏せる理由が薄い。
//
// プレイは「勝てるなら最弱の勝てる札、勝てないなら最も点の低い札」。
func (s *Sjavs) SjavsCpuDecide(idx int) SjavsCpuAction {
	if s.phase == SjavsPhaseBid {
		length := s.LongestTrumpLength(idx)
		if length < SjavsMinBid {
			return SjavsCpuAction{BidLength: 0, HandIdx: -1}
		}
		clubs := false
		for _, suit := range s.SjavsLongestSuits(idx) {
			if suit == CardDesignClover {
				clubs = true
			}
		}
		if !s.beatsStandingBid(length, clubs) {
			return SjavsCpuAction{BidLength: 0, HandIdx: -1}
		}
		return SjavsCpuAction{BidLength: length, HandIdx: -1}
	}

	p := s.GetPlayer(idx)
	valid := s.GetValidPlayIndices(idx)
	if p == nil || len(valid) == 0 {
		return SjavsCpuAction{HandIdx: -1}
	}
	if len(s.trick) == 0 {
		return SjavsCpuAction{HandIdx: s.strongestOf(idx, valid)}
	}

	// 現在の勝ち札を上回れる最弱の札を探す。
	leaderCard := s.trick[0].Card
	for _, tc := range s.trick[1:] {
		if sjavsBeats(tc.Card, leaderCard, s.trumpSuit) {
			leaderCard = tc.Card
		}
	}
	best, bestStrength := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if !sjavsBeats(c, leaderCard, s.trumpSuit) {
			continue
		}
		st := sjavsTrumpStrength(c)
		if !SjavsIsTrump(c, s.trumpSuit) {
			st = sjavsPlainRank(c)
		}
		if best == -1 || st < bestStrength {
			best, bestStrength = i, st
		}
	}
	if best >= 0 {
		return SjavsCpuAction{HandIdx: best}
	}
	return SjavsCpuAction{HandIdx: s.cheapestOf(idx, valid)}
}

// strongestOf は候補のうち最も強い札の添字を返す。
func (s *Sjavs) strongestOf(idx int, candidates []int) int {
	p := s.GetPlayer(idx)
	best, bestScore := candidates[0], -1
	for _, i := range candidates {
		c := p.GetCard(i)
		score := sjavsPlainRank(c)
		if SjavsIsTrump(c, s.trumpSuit) {
			score = 100 + sjavsTrumpStrength(c)
		}
		if score > bestScore {
			best, bestScore = i, score
		}
	}
	return best
}

// cheapestOf は候補のうち最も点の低い (同点なら弱い) 札の添字を返す。
func (s *Sjavs) cheapestOf(idx int, candidates []int) int {
	p := s.GetPlayer(idx)
	best, bestPts, bestRank := candidates[0], 99, 99
	for _, i := range candidates {
		c := p.GetCard(i)
		pts := SjavsCardPoints(c)
		rank := sjavsPlainRank(c)
		if pts < bestPts || (pts == bestPts && rank < bestRank) {
			best, bestPts, bestRank = i, pts, rank
		}
	}
	return best
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (s *Sjavs) GetPlayers() []*SjavsPlayer { return s.players }

// GetPlayer は idx のプレイヤーを返す。
func (s *Sjavs) GetPlayer(idx int) *SjavsPlayer {
	if idx < 0 || idx >= len(s.players) {
		return nil
	}
	return s.players[idx]
}

// GetPhase は現在のフェーズを返す。
func (s *Sjavs) GetPhase() SjavsPhase { return s.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (s *Sjavs) GetCurrentPlayerIdx() int { return s.currentIdx }

// GetDealerIdx は親の添字を返す。
func (s *Sjavs) GetDealerIdx() int { return s.dealerIdx }

// GetTrumpSuit は切札スートを返す (-1: 未確定)。
func (s *Sjavs) GetTrumpSuit() int { return s.trumpSuit }

// GetBidderIdx は切札を宣言した席を返す (-1: 未確定)。
func (s *Sjavs) GetBidderIdx() int { return s.bidderIdx }

// GetBidLength は確定したビッドの枚数を返す。
func (s *Sjavs) GetBidLength() int { return s.bidLength }

// GetBids は席ごとの申告枚数を返す。
func (s *Sjavs) GetBids() []int { return s.bids }

// GetTrick は現在のトリックを返す。
func (s *Sjavs) GetTrick() []SjavsTrickCard { return s.trick }

// GetTrickNumber は完了したトリック数を返す。
func (s *Sjavs) GetTrickNumber() int { return s.trickNo }

// GetTeamPoints は team の獲得点を返す。
func (s *Sjavs) GetTeamPoints(team int) int {
	if team < 0 || team >= len(s.points) {
		return 0
	}
	return s.points[team]
}

// GetRemaining は team の 24 からの残りを返す。
func (s *Sjavs) GetRemaining(team int) int {
	if team < 0 || team >= len(s.remaining) {
		return 0
	}
	return s.remaining[team]
}

// GetCrosses は team のラバー勝利数を返す。
func (s *Sjavs) GetCrosses(team int) int {
	if team < 0 || team >= len(s.crosses) {
		return 0
	}
	return s.crosses[team]
}

// GetCarryOver は 60-60 で持ち越された上乗せ点を返す。
func (s *Sjavs) GetCarryOver() int { return s.carryOver }

// GetHandResult は直近のハンド精算を返す (nil: 未精算)。
func (s *Sjavs) GetHandResult() *SjavsHandResult { return s.handResult }

// GetGameEndFlag はラバーが決着しているかを返す。
func (s *Sjavs) GetGameEndFlag() bool { return s.gameEndFlag }

// GetWinnerTeam は勝ったチームを返す (-1: 未決着)。
func (s *Sjavs) GetWinnerTeam() int { return s.winnerTeam }

// IsDoubleVictory は相手が 24 のまま残っての勝利かを返す。
func (s *Sjavs) IsDoubleVictory() bool {
	if s.winnerTeam < 0 {
		return false
	}
	return s.remaining[1-s.winnerTeam] == SjavsRubber
}

// GetConfig はゲーム設定を返す。
func (s *Sjavs) GetConfig() SjavsConfig { return s.config }

// SetConfig はゲーム設定をセットする。
func (s *Sjavs) SetConfig(c SjavsConfig) { s.config = c }

// GetActionLog は棋譜を返す。
func (s *Sjavs) GetActionLog() []*ActionLogEntry { return s.actionLog }

// SetTrumpSuitForTest はテスト用に切札を差し替える。
func (s *Sjavs) SetTrumpSuitForTest(suit int) { s.trumpSuit = suit }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (s *Sjavs) SetPhaseForTest(p SjavsPhase) { s.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (s *Sjavs) SetCurrentPlayerForTest(idx int) { s.currentIdx = idx }

// SetBidderForTest はテスト用にビッド者を差し替える。
func (s *Sjavs) SetBidderForTest(idx int) { s.bidderIdx = idx }

// SetTeamPointsForTest はテスト用にチーム得点を差し替える。
func (s *Sjavs) SetTeamPointsForTest(a, b int) { s.points = []int{a, b} }

// SetTricksWonForTest はテスト用にトリック数を差し替える。
func (s *Sjavs) SetTricksWonForTest(won []int) { s.tricksWon = won }

// SetRemainingForTest はテスト用に残り点を差し替える。
func (s *Sjavs) SetRemainingForTest(a, b int) { s.remaining = []int{a, b} }

// SettleHandForTest はテスト用に精算を走らせる。
func (s *Sjavs) SettleHandForTest() { s.settleHand() }

// addLog は棋譜に 1 件追加する。
func (s *Sjavs) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.actionLog = append(s.actionLog, &ActionLogEntry{
		TurnNumber: len(s.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// sjavsTrickCardJSON is the JSON wire format for SjavsTrickCard.
type sjavsTrickCardJSON struct {
	PlayerIdx int   `json:"p"`
	Card      *Card `json:"c"`
}

// sjavsJSON is the JSON wire format for Sjavs.
type sjavsJSON struct {
	Players    []*SjavsPlayer       `json:"pl"`
	Config     SjavsConfig          `json:"cfg"`
	Phase      SjavsPhase           `json:"ph"`
	Dealer     int                  `json:"dl"`
	Current    int                  `json:"cur"`
	Bids       []int                `json:"bd"`
	BidderIdx  int                  `json:"bi"`
	BidLength  int                  `json:"bl"`
	BidIsClubs bool                 `json:"bc"`
	TrumpSuit  int                  `json:"ts"`
	Trick      []sjavsTrickCardJSON `json:"tk"`
	LeadIdx    int                  `json:"li"`
	TrickNo    int                  `json:"tn"`
	TricksWon  []int                `json:"tw"`
	Points     []int                `json:"pt"`
	Remaining  []int                `json:"rm"`
	CarryOver  int                  `json:"co"`
	Crosses    []int                `json:"cr"`
	HandResult *SjavsHandResult     `json:"hr,omitempty"`
	GameEnd    bool                 `json:"ge"`
	WinnerTeam int                  `json:"wt"`
	ActionLog  []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (s *Sjavs) MarshalJSON() ([]byte, error) {
	trick := make([]sjavsTrickCardJSON, 0, len(s.trick))
	for _, tc := range s.trick {
		trick = append(trick, sjavsTrickCardJSON{PlayerIdx: tc.PlayerIdx, Card: tc.Card})
	}
	return json.Marshal(sjavsJSON{
		Players: s.players, Config: s.config, Phase: s.phase, Dealer: s.dealerIdx,
		Current: s.currentIdx, Bids: s.bids, BidderIdx: s.bidderIdx, BidLength: s.bidLength,
		BidIsClubs: s.bidIsClubs, TrumpSuit: s.trumpSuit, Trick: trick, LeadIdx: s.leadIdx,
		TrickNo: s.trickNo, TricksWon: s.tricksWon, Points: s.points, Remaining: s.remaining,
		CarryOver: s.carryOver, Crosses: s.crosses, HandResult: s.handResult,
		GameEnd: s.gameEndFlag, WinnerTeam: s.winnerTeam, ActionLog: s.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// KV から戻る生バイト列は信用できないので、席数に合わせて詰め直し、設定を検証する。
// **remaining と crosses が落ちるとラバーの進行が巻き戻る**ので、長さも固定する。
func (s *Sjavs) UnmarshalJSON(data []byte) error {
	var raw sjavsJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Players) != SjavsPlayerCnt {
		return fmt.Errorf("expected %d players, got %d", SjavsPlayerCnt, len(raw.Players))
	}
	if err := raw.Config.Validate(); err != nil {
		return err
	}
	if raw.Phase < SjavsPhaseBid || raw.Phase > SjavsPhaseGameEnd {
		return fmt.Errorf("unknown phase: %d", raw.Phase)
	}

	s.players = raw.Players
	s.config = raw.Config
	s.phase = raw.Phase
	s.bidLength = raw.BidLength
	s.bidIsClubs = raw.BidIsClubs
	s.trickNo = raw.TrickNo
	s.carryOver = raw.CarryOver
	s.handResult = raw.HandResult
	s.gameEndFlag = raw.GameEnd
	s.actionLog = raw.ActionLog

	s.dealerIdx = clampSjavsIdx(raw.Dealer, len(s.players))
	s.currentIdx = clampSjavsIdx(raw.Current, len(s.players))
	s.leadIdx = clampSjavsIdx(raw.LeadIdx, len(s.players))
	s.bidderIdx = raw.BidderIdx
	if s.bidderIdx < -1 || s.bidderIdx >= len(s.players) {
		s.bidderIdx = -1
	}
	s.winnerTeam = raw.WinnerTeam
	if s.winnerTeam < -1 || s.winnerTeam > 1 {
		s.winnerTeam = -1
	}
	s.trumpSuit = raw.TrumpSuit
	if s.trumpSuit < CardDesignSpade || s.trumpSuit > CardDesignDiamond {
		s.trumpSuit = -1
	}

	s.bids = padSjavsInts(raw.Bids, len(s.players))
	s.tricksWon = padSjavsInts(raw.TricksWon, len(s.players))
	s.points = padSjavsInts(raw.Points, 2)
	s.crosses = padSjavsInts(raw.Crosses, 2)
	s.remaining = padSjavsInts(raw.Remaining, 2)
	if len(raw.Remaining) < 2 {
		// 欠けた側は 0 ではなく満点から始める。0 だと即ラバー勝ちになる。
		for i := len(raw.Remaining); i < 2; i++ {
			s.remaining[i] = SjavsRubber
		}
	}

	s.trick = make([]SjavsTrickCard, 0, len(raw.Trick))
	for _, tc := range raw.Trick {
		if tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= len(s.players) {
			continue
		}
		s.trick = append(s.trick, SjavsTrickCard{PlayerIdx: tc.PlayerIdx, Card: tc.Card})
	}
	return nil
}

// clampSjavsIdx は席番号を 0..n-1 に収める。
func clampSjavsIdx(idx, n int) int {
	if idx < 0 || idx >= n {
		return 0
	}
	return idx
}

// padSjavsInts は長さ n に詰め直す。
func padSjavsInts(src []int, n int) []int {
	out := make([]int, n)
	copy(out, src)
	return out
}
