//go:build !js || !wasm || extra3

package domain

import (
	"fmt"
	"math/rand"
)

// RamschPlayerCnt 3 active players
const RamschPlayerCnt = 3

// RamschHandSize cards dealt per player
const RamschHandSize = 10

// RamschSkatSize は伏せておく 2 枚 (the "Skat")。
//
// Skat 本編と違い、誰も取り上げも捨て替えもしない。**最終トリックの獲得者が
// 受け取る**ので、点は必ず誰かの負担になる（120 点がどこにも消えない）。
const RamschSkatSize = 2

// ramschTrumpSuitMarker は「切り札の一族」を表す擬似スート。ジャック 4 枚は
// 印刷されたスートに属さないので、フォロー判定でこの値に寄せる。
const ramschTrumpSuitMarker = -1

// ramschJackOrder は切り札 (ジャック) 4 枚の強さの順。**クラブが最強**で、
// スートの一般的な強弱とは別物。
var ramschJackOrder = []int{CardDesignClover, CardDesignSpade, CardDesignHeart, CardDesignDiamond}

// RamschTricksPerRound number of tricks per round
const RamschTricksPerRound = 10

// RamschPhase game phase
type RamschPhase int

// Ramsch phases.
//
// **入札も宣言も無い。** Skat から来た `Bid` / `Pickup` / `Discard` /
// `GameDeclaration` の 4 フェーズは、このゲームには存在しない ── 配ったら
// すぐプレイに入る。ジャック 4 枚が常に切り札で、選ぶ余地が無いのがこの
// ゲームの前提そのもの。
const (
	// RamschPhasePlay trick play phase
	RamschPhasePlay RamschPhase = 0
	// RamschPhaseTrickEnd trick end phase
	RamschPhaseTrickEnd RamschPhase = 1
	// RamschPhaseRoundEnd round end phase
	RamschPhaseRoundEnd RamschPhase = 2
	// RamschPhaseGameEnd game end phase
	RamschPhaseGameEnd RamschPhase = 3
)

// RamschTotalCardPoints は 1 ラウンドで動く点数の総和。
// A=11, 10=10, K=4, Q=3, J=2 の 4 スート分で 120 点。
const RamschTotalCardPoints = 120

// Ramsch card values 7..A use the standard project encoding
//
//	value 7 = the seven, 8 = eight, 9 = nine, 10 = ten,
//	11 = jack, 12 = queen, 13 = king, 1 = ace.
const (
	ramschValueSeven = 7
	ramschValueEight = 8
	ramschValueNine  = 9
	ramschValueTen   = 10
	ramschValueJack  = 11
	ramschValueQueen = 12
	ramschValueKing  = 13
	ramschValueAce   = 1
)

// RamschHint hint information for the human player
type RamschHint struct {
	CardIndex    *int  // recommended card index (play phase)
	Bid          *int  // recommended bid value
	GameType     *int  // recommended game type
	TrumpSuit    *int  // recommended trump suit (suit games only)
	PickRamsch   *bool // recommended ramsch pickup flag
	DiscardIndex *int  // recommended discard index
	Reason       string
}

// ramschRoundState round-scoped state
type ramschRoundState struct {
	phase            RamschPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int // rotates each round
	forehandIdx      int // leads the first trick (left of dealer)
	middlehandIdx    int
	rearhandIdx      int
	// skat は伏せておく 2 枚。**最終トリックの獲得者が受け取る**ので、
	// 「誰も取らない中立の 2 枚」ではなく最後まで勝敗に効く。
	skat []*Card
	// loserIdx は最も点を集めてしまったプレイヤー。Durchmarsch のときは -1。
	loserIdx int
	// durchmarsch は 1 人が全トリックを取ったか。取ると逆転勝ちになる。
	durchmarsch bool
	// durchmarschIdx は総取りしたプレイヤー (durchmarsch でなければ -1)。
	durchmarschIdx int
	// trickResolved はいまのトリックを精算済みか。**二重加算を防ぐための印。**
	// 精算は「CPU が最後を出したとき (usecase のループ)」と「人間が最後を出して
	// NextTrick が拾うとき」の 2 経路から来るので、フラグが無いと同じトリックを
	// 2 回数えて 120 点を超える。
	trickResolved bool
	gameEndFlag   bool
	actionLogBase
}

// Ramsch game class
type Ramsch struct {
	trumpCards *TrumpCards
	players    []*RamschPlayer
	config     RamschConfig
	round      ramschRoundState
}

// NewRamsch constructor
func NewRamsch(trumpCards *TrumpCards, players []*RamschPlayer, config RamschConfig) *Ramsch {
	return &Ramsch{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: ramschRoundState{
			loserIdx:       -1,
			durchmarschIdx: -1,
		},
	}
}

// NewDefaultRamsch returns Ramsch with the standard 3-player setup (1 human, 2 CPU).
func NewDefaultRamsch() *Ramsch {
	players := []*RamschPlayer{
		NewRamschPlayer(true),
		NewRamschPlayer(false),
		NewRamschPlayer(false),
	}
	return NewRamsch(newRamschDeck(), players, DefaultRamschConfig())
}

// newRamschDeck builds the 32-card Ramsch deck (7..A across the four standard suits).
//
// **NewTrumpCardsWithSuits(32, …) では作れない。** 枚数で打ち切る汎用
// コンストラクタなので、♠13 + ♣13 + ♥6 + ♦0 になり**ダイヤが 1 枚も入らない**。
// 切り札スートを選ぶゲームでこれをやると、選べるのに 1 枚も存在しない切り札が
// できる (#5296)。German 32 枚パックは Ramsch / Belote / Prsi で共通。
func newRamschDeck() *TrumpCards {
	return NewTrumpCards32()
}

// Reset initializes a new game session.
func (s *Ramsch) Reset() {
	s.round = ramschRoundState{
		roundNumber:    1,
		dealerIdx:      0,
		loserIdx:       -1,
		durchmarschIdx: -1,
	}

	for _, p := range s.players {
		p.SetCumulativeScore(0)
		p.SetRoundsWon(0)
		p.SetRoundsLost(0)
		p.ResetRound()
	}

	s.startRound()
}

// startRound resets per-round state, deals, and goes straight to trick play.
//
// **入札が無いので配ったらすぐプレイ。** Skat はここから 4 フェーズを挟むが、
// Ramsch は切り札が最初から決まっている（ジャック 4 枚）ので挟むものが無い。
func (s *Ramsch) startRound() {
	for _, p := range s.players {
		p.ResetRound()
	}

	s.trumpCards = newRamschDeck()
	s.dealCards()
	s.sortAllHands()

	// Positions: forehand = (dealer + 1) % 3, middlehand = +2, rearhand = +3.
	s.round.forehandIdx = (s.round.dealerIdx + 1) % RamschPlayerCnt
	s.round.middlehandIdx = (s.round.dealerIdx + 2) % RamschPlayerCnt
	s.round.rearhandIdx = (s.round.dealerIdx + 3) % RamschPlayerCnt

	s.round.loserIdx = -1
	s.round.durchmarsch = false
	s.round.durchmarschIdx = -1
	s.round.trickNumber = 1
	s.round.currentTrick = nil

	// フォアハンドが最初のトリックをリードする。
	s.round.leadPlayerIdx = s.round.forehandIdx
	s.round.currentPlayerIdx = s.round.forehandIdx
	s.round.phase = RamschPhasePlay

	s.appendLog(-1, "round_start",
		fmt.Sprintf("Round %d: dealer=%s", s.round.roundNumber, playerName(s.players, s.round.dealerIdx)), nil)

	s.runCpuTurns()
}

// NextRound advances to the next round.
func (s *Ramsch) NextRound() {
	if s.round.phase != RamschPhaseRoundEnd {
		return
	}
	prevRound := s.round.roundNumber
	prevDealer := s.round.dealerIdx
	s.round = ramschRoundState{
		roundNumber:    prevRound + 1,
		dealerIdx:      (prevDealer + 1) % RamschPlayerCnt,
		loserIdx:       -1,
		durchmarschIdx: -1,
	}
	s.startRound()
}

// dealCards deals 10/10/10 + 2 (the skat).
func (s *Ramsch) dealCards() {
	s.trumpCards.Shuffle()
	// Deal 3 then 2 skat cards then 4 then 3 (a common pattern). For simplicity
	// we just deal 10 cards round-robin, then 2 to the skat.
	for range RamschHandSize {
		for i := range RamschPlayerCnt {
			c := s.trumpCards.DrawCard()
			if c != nil {
				s.players[i].AddCard(c)
			}
		}
	}
	skat := []*Card{}
	for range RamschSkatSize {
		c := s.trumpCards.DrawCard()
		if c != nil {
			skat = append(skat, c)
		}
	}
	s.round.skat = skat
}

// PlayerPlay human plays a card.
func (s *Ramsch) PlayerPlay(cardIndex int) error {
	if s.round.gameEndFlag {
		return ErrGameEnded
	}
	if s.round.phase != RamschPhasePlay {
		return ErrWrongPhase
	}
	if !s.IsHumanTurn() {
		return ErrNotHumanTurn
	}
	player := s.players[s.round.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "Card index out of range.")
	}
	card := player.GetCard(cardIndex)
	if err := s.validatePlay(s.round.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	s.playCard(s.round.currentPlayerIdx, played)
	// 人間が出したら、次の人間の手番まで CPU を進める。ここを呼ばないと
	// 画面が「CPU の番」で止まり、操作する先が無くなる。
	s.runCpuTurns()
	return nil
}

// CpuPlay current CPU plays a card.
func (s *Ramsch) CpuPlay() {
	if s.round.gameEndFlag || s.round.phase != RamschPhasePlay {
		return
	}
	if s.players[s.round.currentPlayerIdx].GetIsHuman() {
		return
	}
	idx := s.cpuPickPlay(s.round.currentPlayerIdx)
	played := s.players[s.round.currentPlayerIdx].RemoveCard(idx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	s.playCard(s.round.currentPlayerIdx, played)
}

// playCard appends the card to the current trick and advances the turn.
func (s *Ramsch) playCard(playerIdx int, card *Card) {
	s.round.currentTrick = append(s.round.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	s.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(s.players, playerIdx), cardStr(card)), []*Card{card})
	if len(s.round.currentTrick) == RamschPlayerCnt {
		s.round.phase = RamschPhaseTrickEnd
		return
	}
	s.round.currentPlayerIdx = (s.round.currentPlayerIdx + 1) % RamschPlayerCnt
}

// ResolveTrick determines the trick winner and updates points.
func (s *Ramsch) ResolveTrick() {
	if s.round.phase != RamschPhaseTrickEnd || len(s.round.currentTrick) != RamschPlayerCnt {
		return
	}
	if s.round.trickResolved {
		return // 既に精算済み。二度数えない。
	}
	s.round.trickResolved = true
	winnerIdx := s.trickWinner()
	cards := make([]*Card, len(s.round.currentTrick))
	pts := 0
	for i, tc := range s.round.currentTrick {
		cards[i] = tc.Card
		pts += ramschCardPoints(tc.Card)
	}
	s.players[winnerIdx].AddTrick(cards)
	s.players[winnerIdx].SetCardPoints(s.players[winnerIdx].GetCardPoints() + pts)
	s.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d card points)", playerName(s.players, winnerIdx), s.round.trickNumber, pts), cards)
	s.round.leadPlayerIdx = winnerIdx

	if s.round.trickNumber < RamschTricksPerRound {
		s.round.phase = RamschPhaseTrickEnd
		return
	}

	// **最終トリックの獲得者が伏せ札 2 枚を受け取る。** 誰にも付けないと
	// 120 点のうち数点が宙に浮き、「取らなければ得」という前提が崩れる。
	skatPts := 0
	for _, c := range s.round.skat {
		skatPts += ramschCardPoints(c)
	}
	if skatPts > 0 || len(s.round.skat) > 0 {
		s.players[winnerIdx].SetCardPoints(s.players[winnerIdx].GetCardPoints() + skatPts)
		s.appendLog(winnerIdx, "skat_award",
			fmt.Sprintf("%s takes the skat (%d card points)", playerName(s.players, winnerIdx), skatPts),
			s.round.skat)
	}

	// **Durchmarsch**: 10 トリック全部を 1 人が取ったか。
	for i := range s.players {
		if s.players[i].GetTrickCount() == RamschTricksPerRound {
			s.round.durchmarsch = true
			s.round.durchmarschIdx = i
		}
	}

	s.round.phase = RamschPhaseRoundEnd
}

// NextTrick advances to the next trick, led by the player who took the last one.
//
// **まだ精算されていなければ先に精算する。** 人間が 3 枚目を出した場合、
// フェーズは TrickEnd になるが ResolveTrick を呼ぶ人がいない（CPU が最後を
// 出したときだけ usecase 側のループが呼ぶ）。ここで拾わないと、点が加算されない
// まま次のトリックへ進み、**120 点のうちいくつかが誰のものにもならない。**
func (s *Ramsch) NextTrick() {
	if s.round.phase != RamschPhaseTrickEnd {
		return
	}
	if !s.round.trickResolved && len(s.round.currentTrick) == RamschPlayerCnt {
		s.ResolveTrick()
		if s.round.phase != RamschPhaseTrickEnd {
			return // 最終トリックだった: ラウンド終了へ
		}
	}
	s.round.currentTrick = nil
	s.round.trickResolved = false
	s.round.trickNumber++
	s.round.currentPlayerIdx = s.round.leadPlayerIdx
	s.round.phase = RamschPhasePlay
	s.runCpuTurns()
}

// runCpuTurns plays out CPU turns until it is the human's move (or the trick
// is complete).
//
// **上限を必ず持つ。** 出せる札が無いなど、どこかが詰まったときに無限ループへ
// 落ちると Worker ごと止まる。1 トリックは 3 手なので、それを超えたら抜ける。
func (s *Ramsch) runCpuTurns() {
	for i := 0; i < RamschPlayerCnt; i++ {
		if s.round.gameEndFlag || s.round.phase != RamschPhasePlay {
			return
		}
		if s.round.currentPlayerIdx < 0 || s.round.currentPlayerIdx >= len(s.players) {
			return
		}
		if s.players[s.round.currentPlayerIdx].GetIsHuman() {
			return
		}
		before := len(s.round.currentTrick)
		s.CpuPlay()
		if len(s.round.currentTrick) == before {
			return // 進まなかった: 詰まっているので抜ける
		}
	}
}

// ScoreRound は 1 ラウンドを精算する。**得点は罰点で、多く取った人が負う。**
//
// これが Skat と逆になっている一点で、他は全部同じトリック機構に乗っている。
//
//   - 各プレイヤーが取ったカード点を数える (A=11, 10=10, K=4, Q=3, J=2、計 120)。
//   - **伏せてある 2 枚 (skat) は最終トリックの獲得者が受け取る。** 誰にも
//     付けないと 120 点のうち数点が宙に浮き、「取らなければ得」という
//     このゲームの前提が崩れる。
//   - 最も多く取ったプレイヤーがその点数を**失点**する。同点なら全員が負う
//     （罰を誰かに押し付ける根拠が無い）。
//   - **Durchmarsch**: 1 人が 10 トリック全部を取ったら逆転勝ちで、
//     **他の 2 人**が 120 点ずつ失点する。取らないゲームの唯一の攻めどころ。
func (s *Ramsch) ScoreRound() {
	if s.round.phase != RamschPhaseRoundEnd {
		return
	}

	if s.round.durchmarsch {
		winner := s.round.durchmarschIdx
		for i, p := range s.players {
			if i == winner {
				p.SetRoundScore(0)
				p.IncRoundsWon()
				continue
			}
			p.SetRoundScore(-RamschTotalCardPoints)
			p.IncRoundsLost()
		}
		s.appendLog(winner, "durchmarsch",
			fmt.Sprintf("%s takes every trick (Durchmarsch): the others lose %d each",
				playerName(s.players, winner), RamschTotalCardPoints), nil)
	} else {
		worst := 0
		for i := 1; i < RamschPlayerCnt; i++ {
			if s.GetCardPoints(i) > s.GetCardPoints(worst) {
				worst = i
			}
		}
		// 同点は全員が負う。「先に座っている人が損をする」は根拠が無い。
		tied := []int{}
		for i := 0; i < RamschPlayerCnt; i++ {
			if s.GetCardPoints(i) == s.GetCardPoints(worst) {
				tied = append(tied, i)
			}
		}
		s.round.loserIdx = worst
		if len(tied) > 1 {
			s.round.loserIdx = -1
		}
		for i, p := range s.players {
			penalised := false
			for _, idx := range tied {
				if idx == i {
					penalised = true
				}
			}
			if penalised {
				p.SetRoundScore(-s.GetCardPoints(i))
				p.IncRoundsLost()
				s.appendLog(i, "round_result",
					fmt.Sprintf("%s took the most card points (-%d)",
						playerName(s.players, i), s.GetCardPoints(i)), nil)
				continue
			}
			p.SetRoundScore(0)
			p.IncRoundsWon()
		}
	}

	for i, p := range s.players {
		p.CommitRoundScore()
		s.appendLog(i, "cumulative_score",
			fmt.Sprintf("%s total=%d", playerName(s.players, i), p.GetCumulativeScore()), nil)
	}

	s.checkGameEnd()
}

// validatePlay verifies that a card is legal to play.
func (s *Ramsch) validatePlay(playerIdx int, card *Card) error {
	if len(s.round.currentTrick) == 0 {
		return nil // any lead is legal
	}
	leadCard := s.round.currentTrick[0].Card
	leadIsTrump := s.isTrump(leadCard)
	cardIsTrump := s.isTrump(card)
	if leadIsTrump {
		// Must follow trump if possible.
		if !cardIsTrump && s.playerHasTrump(playerIdx) {
			return NewDomainError(ErrInvalidPlay, "You must follow trump.")
		}
		return nil
	}
	leadSuit := s.effectiveSuit(leadCard)
	cardSuit := s.effectiveSuit(card)
	// Must follow lead suit if possible (and the card should not be trump).
	if cardIsTrump && s.playerHasSuitNonTrump(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "You must follow suit.")
	}
	if !cardIsTrump && cardSuit != leadSuit && s.playerHasSuitNonTrump(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "You must follow suit.")
	}
	return nil
}

// effectiveSuit returns the suit a card belongs to. **ジャックは自分のスートに
// 属さない** ── 4 枚まとめて切り札の一族なので、♣J でクラブをフォローすることは
// できない。
//
// **このジャックの枝はテストでは観測できない。** 呼び出し側 (validatePlay /
// trickWinner / ramschBeats) がどれも先に `isTrump` で分岐するので、
// ジャックがここに届く経路が今は無く、印刷スートを返すよう壊しても
// 全テストが通る（ミューテーションを当てて確認済み）。規則の表明として
// 残してあるだけで、「テスト済み」ではない。`isTrump` を見ない呼び出しを
// 足したら、その瞬間からここが効き始める。
func (s *Ramsch) effectiveSuit(c *Card) int {
	if c.GetValue() == ramschValueJack {
		return ramschTrumpSuitMarker
	}
	return c.GetDesign()
}

// isTrump reports whether the card is a trump.
//
// **ジャック 4 枚だけ、常に。** 入札が無いので切り札を選ぶ余地が無く、これが
// Skat の Grand と同じ札順になる。ここに宣言由来の分岐は残さない ── 残すと
// 到達しない枝になり、「選べるように見えて選べない」ゲームになる。
func (s *Ramsch) isTrump(c *Card) bool {
	return c.GetValue() == ramschValueJack
}

// playerHasTrump reports whether the player holds any trump.
func (s *Ramsch) playerHasTrump(playerIdx int) bool {
	p := s.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if s.isTrump(p.GetCard(i)) {
			return true
		}
	}
	return false
}

// playerHasSuitNonTrump reports whether the player holds a non-trump card of
// the given suit.
func (s *Ramsch) playerHasSuitNonTrump(playerIdx int, suit int) bool {
	p := s.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if s.isTrump(c) {
			continue
		}
		if c.GetDesign() == suit {
			return true
		}
	}
	return false
}

// trickWinner determines the winner of the current trick.
func (s *Ramsch) trickWinner() int {
	if len(s.round.currentTrick) == 0 {
		return 0
	}
	leadCard := s.round.currentTrick[0].Card
	leadIsTrump := s.isTrump(leadCard)
	leadSuit := s.effectiveSuit(leadCard)
	winner := s.round.currentTrick[0]
	winnerStrength := s.cardStrength(leadCard)
	for _, tc := range s.round.currentTrick[1:] {
		isTrump := s.isTrump(tc.Card)
		if isTrump && !leadIsTrump {
			// First trump played beats any non-trump.
			if !s.isTrump(winner.Card) {
				winner = tc
				winnerStrength = s.cardStrength(tc.Card)
				continue
			}
			// Compare two trumps.
			str := s.cardStrength(tc.Card)
			if str > winnerStrength {
				winner = tc
				winnerStrength = str
			}
			continue
		}
		if !isTrump && !leadIsTrump && tc.Card.GetDesign() == leadSuit && !s.isTrump(winner.Card) {
			str := s.cardStrength(tc.Card)
			if str > winnerStrength {
				winner = tc
				winnerStrength = str
			}
			continue
		}
		if isTrump && leadIsTrump {
			str := s.cardStrength(tc.Card)
			if str > winnerStrength {
				winner = tc
				winnerStrength = str
			}
		}
	}
	return winner.PlayerIdx
}

// cardStrength returns the trick-comparison strength for trump/lead-suit cards.
//
// 切り札はジャック 4 枚だけで、その中の序列は **♣ > ♠ > ♥ > ♦**（Skat 系の
// 固定順で、スートの強さの順ではない）。ここは入札の結果に左右されないので、
// 表を毎回組み立てず定数として持っている。
func (s *Ramsch) cardStrength(c *Card) int {
	if c.GetValue() == ramschValueJack {
		for i, d := range ramschJackOrder {
			if d == c.GetDesign() {
				return 100 - i // 高いほど強い
			}
		}
		return 100
	}
	// 非切り札: A > 10 > K > Q > 9 > 8 > 7。**10 が K より強い**のが Skat 系。
	switch c.GetValue() {
	case ramschValueAce:
		return 7
	case ramschValueTen:
		return 6
	case ramschValueKing:
		return 5
	case ramschValueQueen:
		return 4
	case ramschValueNine:
		return 3
	case ramschValueEight:
		return 2
	case ramschValueSeven:
		return 1
	}
	return 0
}

// ramschCardPoints returns the card-point value used to score tricks.
func ramschCardPoints(c *Card) int {
	switch c.GetValue() {
	case ramschValueAce:
		return 11
	case ramschValueTen:
		return 10
	case ramschValueKing:
		return 4
	case ramschValueQueen:
		return 3
	case ramschValueJack:
		return 2
	}
	return 0
}

// checkGameEnd determines whether the game should end.
func (s *Ramsch) checkGameEnd() {
	for _, p := range s.players {
		if p.GetCumulativeScore() >= s.config.TargetScore {
			s.round.gameEndFlag = true
			s.round.phase = RamschPhaseGameEnd
			s.appendLog(-1, "game_end",
				fmt.Sprintf("%s reaches %d points and wins!", playerName(s.players, s.findIndex(p)), p.GetCumulativeScore()), nil)
			return
		}
	}
}

// findIndex returns the player's index.
func (s *Ramsch) findIndex(p *RamschPlayer) int {
	for i, q := range s.players {
		if p == q {
			return i
		}
	}
	return -1
}

// IsHumanTurn reports whether the current player is human (play phase).
func (s *Ramsch) IsHumanTurn() bool {
	return isHumanTurn(s.players, s.round.currentPlayerIdx)
}

// GetValidPlayIndices returns the indices of legally playable cards.
func (s *Ramsch) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return nil
	}
	p := s.players[playerIdx]
	var valid []int
	for i := 0; i < p.GetCardsSize(); i++ {
		if s.validatePlay(playerIdx, p.GetCard(i)) == nil {
			valid = append(valid, i)
		}
	}
	return valid
}

// GetHint returns a hint for the human player.
//
// 助言する場面はプレイの手番だけ。入札も宣言も無いので、Skat の 4 分岐は
// そもそも到達しない。
func (s *Ramsch) GetHint() *RamschHint {
	humanIdx := findHumanIdx(s.players)
	if humanIdx < 0 || s.round.phase != RamschPhasePlay {
		return nil
	}
	if s.round.currentPlayerIdx != humanIdx {
		return nil
	}
	if len(s.GetValidPlayIndices(humanIdx)) == 0 {
		return nil
	}
	idx := s.cpuPickPlay(humanIdx)
	return &RamschHint{CardIndex: &idx, Reason: ramschHintReasonFor(s, humanIdx, idx)}
}

// ramschHintReasonFor は助言の理由キーを返す。**「安く逃げる」のか「取りに行く」
// のかで文言が変わる** —— 取らないゲームで「これを出せ」とだけ言われても、
// 何を狙っているのかが読めない。
func ramschHintReasonFor(s *Ramsch, playerIdx, cardIdx int) string {
	p := s.players[playerIdx]
	if cardIdx < 0 || cardIdx >= p.GetCardsSize() {
		return "avoid_points"
	}
	if ramschCardPoints(p.GetCard(cardIdx)) == 0 {
		return "avoid_points"
	}
	if len(s.round.currentTrick) == 0 {
		return "lead_low"
	}
	return "forced_discard"
}

// appendLog appends an entry to the round action log.
func (s *Ramsch) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.round.appendLog(playerIdx, actionType, detail, cards)
}

// sortAllHands sorts every player's hand.
func (s *Ramsch) sortAllHands() {
	sortEachHand(s.players, s.sortHand)
}

// sortHand sorts the player's hand by suit then value.
func (s *Ramsch) sortHand(p *RamschPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		di, dj := ci.GetDesign(), cj.GetDesign()
		if di != dj {
			return di < dj
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// --- Getters / setters ---

// GetPhase returns the current phase.
func (s *Ramsch) GetPhase() RamschPhase { return s.round.phase }

// SetPhase sets the phase (test only).
func (s *Ramsch) SetPhase(p RamschPhase) { s.round.phase = p }

// GetRoundNumber returns the round number.
func (s *Ramsch) GetRoundNumber() int { return s.round.roundNumber }

// GetTrickNumber returns the current trick number.
func (s *Ramsch) GetTrickNumber() int { return s.round.trickNumber }

// GetCurrentPlayerIdx returns the current player index.
func (s *Ramsch) GetCurrentPlayerIdx() int { return s.round.currentPlayerIdx }

// SetCurrentPlayerIdx sets the current player index (test only).
func (s *Ramsch) SetCurrentPlayerIdx(idx int) { s.round.currentPlayerIdx = idx }

// GetCurrentTrick returns the current trick (in play order).
func (s *Ramsch) GetCurrentTrick() []*TrickCard { return s.round.currentTrick }

// SetCurrentTrick sets the current trick (test only).
func (s *Ramsch) SetCurrentTrick(trick []*TrickCard) { s.round.currentTrick = trick }

// GetForehandIdx returns the forehand index.
func (s *Ramsch) GetForehandIdx() int { return s.round.forehandIdx }

// GetMiddlehandIdx returns the middlehand index.
func (s *Ramsch) GetMiddlehandIdx() int { return s.round.middlehandIdx }

// GetRearhandIdx returns the rearhand index.
func (s *Ramsch) GetRearhandIdx() int { return s.round.rearhandIdx }

// GetDealerIdx returns the dealer index.
func (s *Ramsch) GetDealerIdx() int { return s.round.dealerIdx }

// GetSkat は伏せてある 2 枚を返す。最終トリックの獲得者が受け取る。
// otherwise; nil after pickup before discard).
func (s *Ramsch) GetSkat() []*Card { return s.round.skat }

// GetCardPoints は playerIdx がこれまでに集めた点を返す。**多いほど不利。**
//
// **出どころはプレイヤー 1 本。** ラウンド側にも同じ配列を持っていたが、
// 片方だけ更新される経路ができて presenter が古い値を読んだ。二重に持たない。
func (s *Ramsch) GetCardPoints(playerIdx int) int {
	if playerIdx < 0 || playerIdx >= len(s.players) || s.players[playerIdx] == nil {
		return 0
	}
	return s.players[playerIdx].GetCardPoints()
}

// GetLoserIdx は最も点を集めてしまったプレイヤーを返す。
// ラウンド終了前、同点、Durchmarsch のときは -1。
func (s *Ramsch) GetLoserIdx() int { return s.round.loserIdx }

// IsDurchmarsch は 1 人が全トリックを取ったか（逆転勝ち）を返す。
func (s *Ramsch) IsDurchmarsch() bool { return s.round.durchmarsch }

// GetDurchmarschIdx は総取りしたプレイヤーを返す（無ければ -1）。
func (s *Ramsch) GetDurchmarschIdx() int { return s.round.durchmarschIdx }

// GetGameEndFlag returns the game-end flag.
func (s *Ramsch) GetGameEndFlag() bool { return s.round.gameEndFlag }

// GetPlayerCnt returns the player count.
func (s *Ramsch) GetPlayerCnt() int { return len(s.players) }

// GetPlayer returns the i-th player (nil if out of range).
func (s *Ramsch) GetPlayer(i int) *RamschPlayer {
	return getPlayer(s.players, i)
}

// GetLeadPlayerIdx returns the lead player index.
func (s *Ramsch) GetLeadPlayerIdx() int { return s.round.leadPlayerIdx }

// GetConfig returns the config.
func (s *Ramsch) GetConfig() RamschConfig { return s.config }

// SetConfig sets the config.
func (s *Ramsch) SetConfig(c RamschConfig) { s.config = c }

// GetActionLog returns the action log.
func (s *Ramsch) GetActionLog() []*ActionLogEntry { return s.round.actionLog }

// --- CPU AI ---

// cpuPickPlay は CPU の出す札を選ぶ。**狙いは「取らない」こと。**
//
// Skat から持ってきた元の実装は「勝てるなら勝つ」だった。Ramsch でそれをやると
// 点を集めて自分が負けるので、判断が丸ごと裏返る:
//
//   - **勝てる札しか無いなら、いちばん安く勝つ。** 取るのが避けられない
//     トリックで、わざわざ A や 10 を捨てる理由は無い。
//   - **降りられるなら降りる。それも「いちばん高い負け札」で。**
//     手に残った高い札は後で必ず取らされるので、安全に降りられる回に
//     吐き出しておく。ここを「いちばん安い札」にすると、A が最後まで残って
//     終盤に 11 点を抱え込む。
//   - **リードは点の無い低い札から。** 自分から点を出さない。
//
// 難易度が Easy のときだけ、この判断を確率で崩して隙を作る。
func (s *Ramsch) cpuPickPlay(playerIdx int) int {
	valid := s.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if playerIdx < 0 || playerIdx >= len(s.players) || s.players[playerIdx] == nil {
		return 0
	}
	p := s.players[playerIdx]

	if s.config.CpuDifficulty == RamschCpuDifficultyEasy && rand.Intn(3) == 0 { //nolint:gosec // ゲーム AI の非暗号乱数
		return valid[rand.Intn(len(valid))] //nolint:gosec
	}

	// リード: 点の無い、いちばん弱い札から出す。
	if len(s.round.currentTrick) == 0 {
		best, bestKey := valid[0], 1<<30
		for _, i := range valid {
			c := p.GetCard(i)
			// 点をいちばん重く見る。同点なら弱い札。
			key := ramschCardPoints(c)*1000 + s.cardStrength(c)
			if key < bestKey {
				best, bestKey = i, key
			}
		}
		return best
	}

	// 追随: いま勝っている札を特定する（判定は trickWinner に一本化）。
	winnerIdx := s.trickWinner()
	var winnerCard *Card
	for _, tc := range s.round.currentTrick {
		if tc.PlayerIdx == winnerIdx {
			winnerCard = tc.Card
			break
		}
	}
	leadSuit := s.effectiveSuit(s.round.currentTrick[0].Card)

	duckIdx, duckKey := -1, -1  // 降りられる中で「いちばん高い札」
	winIdx, winKey := -1, 1<<30 // 取らざるを得ないとき「いちばん安い札」
	for _, i := range valid {
		c := p.GetCard(i)
		if s.ramschBeats(c, winnerCard, leadSuit) {
			key := ramschCardPoints(c)*1000 + s.cardStrength(c)
			if key < winKey {
				winIdx, winKey = i, key
			}
			continue
		}
		// 降りられる札。高い札から吐く（後で取らされる前に処分する）。
		key := s.cardStrength(c)*1000 + ramschCardPoints(c)
		if key > duckKey {
			duckIdx, duckKey = i, key
		}
	}
	if duckIdx >= 0 {
		return duckIdx
	}
	if winIdx >= 0 {
		return winIdx
	}
	return valid[0]
}

// ramschBeats は card が currentWinner に勝つかを返す。
// 比較規則は cardStrength / isTrump と同じ 1 本を通す。
func (s *Ramsch) ramschBeats(card, currentWinner *Card, leadSuit int) bool {
	if currentWinner == nil {
		return true
	}
	cardIsTrump := s.isTrump(card)
	winnerIsTrump := s.isTrump(currentWinner)
	switch {
	case cardIsTrump && !winnerIsTrump:
		return true
	case cardIsTrump && winnerIsTrump:
		return s.cardStrength(card) > s.cardStrength(currentWinner)
	case !cardIsTrump && winnerIsTrump:
		return false
	default:
		// どちらも非切り札: リードスートに追随している札同士でのみ比べる。
		return s.effectiveSuit(card) == leadSuit &&
			s.cardStrength(card) > s.cardStrength(currentWinner)
	}
}
