//go:build !js || !wasm || extra3

package domain

import (
	"fmt"
	"math/rand"
	"sort"
)

// SkatPlayerCnt 3 active players
const SkatPlayerCnt = 3

// SkatHandSize cards dealt per player
const SkatHandSize = 10

// SkatSkatSize face-down cards (the "Skat")
const SkatSkatSize = 2

// SkatTricksPerRound number of tricks per round
const SkatTricksPerRound = 10

// SkatPhase game phase
type SkatPhase int

// Skat phases
const (
	// SkatPhaseBid bidding phase
	SkatPhaseBid SkatPhase = 0
	// SkatPhaseSkatPickup declarer decides whether to pick up the skat
	SkatPhaseSkatPickup SkatPhase = 1
	// SkatPhaseDiscard declarer discards two cards back to the skat
	SkatPhaseDiscard SkatPhase = 2
	// SkatPhaseGameDeclaration declarer chooses game type and (for suit games) trump
	SkatPhaseGameDeclaration SkatPhase = 3
	// SkatPhasePlay trick play phase
	SkatPhasePlay SkatPhase = 4
	// SkatPhaseTrickEnd trick end phase
	SkatPhaseTrickEnd SkatPhase = 5
	// SkatPhaseRoundEnd round end phase
	SkatPhaseRoundEnd SkatPhase = 6
	// SkatPhaseGameEnd game end phase
	SkatPhaseGameEnd SkatPhase = 7
)

// SkatGameType game type chosen by the declarer
type SkatGameType int

// Skat game types
const (
	// SkatGameNone no game declared yet
	SkatGameNone SkatGameType = 0
	// SkatGameSuit suit game (trump = chosen suit + jacks)
	SkatGameSuit SkatGameType = 1
	// SkatGameGrand grand game (trump = jacks only)
	SkatGameGrand SkatGameType = 2
	// SkatGameNull null game (no trump; declarer must lose every trick)
	SkatGameNull SkatGameType = 3
)

// SkatWinner declarer-vs-defenders outcome
const (
	// SkatWinnerUndecided undecided
	SkatWinnerUndecided = -1
	// SkatWinnerDeclarer declarer wins
	SkatWinnerDeclarer = 0
	// SkatWinnerDefenders defenders win
	SkatWinnerDefenders = 1
)

// Skat card values 7..A use the standard project encoding
//
//	value 7 = the seven, 8 = eight, 9 = nine, 10 = ten,
//	11 = jack, 12 = queen, 13 = king, 1 = ace.
const (
	skatValueSeven = 7
	skatValueEight = 8
	skatValueNine  = 9
	skatValueTen   = 10
	skatValueJack  = 11
	skatValueQueen = 12
	skatValueKing  = 13
	skatValueAce   = 1
)

// SkatHint hint information for the human player
type SkatHint struct {
	CardIndex    *int  // recommended card index (play phase)
	Bid          *int  // recommended bid value
	GameType     *int  // recommended game type
	TrumpSuit    *int  // recommended trump suit (suit games only)
	PickSkat     *bool // recommended skat pickup flag
	DiscardIndex *int  // recommended discard index
	Reason       string
}

// skatRoundState round-scoped state
type skatRoundState struct {
	phase            SkatPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int // rotates each round
	forehandIdx      int // first to bid response (left of dealer)
	middlehandIdx    int
	rearhandIdx      int
	bidderIdx        int // active caller during bidding
	responderIdx     int // active responder during bidding
	bidStep          int // index into the bid ladder for the active call
	currentBid       int // last announced bid value
	auctionRound     int // 1 = middle vs fore, 2 = rear vs survivor
	round1Winner     int // survivor of auction round 1 (-1 = unknown)
	declarerIdx      int // -1 if undecided / passed-out
	passedAtCall     [SkatPlayerCnt]bool
	pickedSkat       bool                // declarer picked up the skat
	gameType         SkatGameType        // chosen game type
	trumpSuit        int                 // chosen trump suit (suit games only)
	skat             []*Card             // 2 face-down cards; possibly post-discard
	originalSkat     []*Card             // pre-pickup skat snapshot for log
	declarerHand     []*Card             // declarer hand snapshot at start of play (used for matadors at scoring time)
	gameValue        int                 // game value (positive when declarer wins)
	breakdown        *SkatScoreBreakdown // 得点の内訳 (#5561)
	declarerCardPts  int
	defendersCardPts int
	winnerSide       int // SkatWinner*
	gameEndFlag      bool
	actionLogBase
}

// Skat game class
type Skat struct {
	trumpCards *TrumpCards
	players    []*SkatPlayer
	config     SkatConfig
	round      skatRoundState
}

// NewSkat constructor
func NewSkat(trumpCards *TrumpCards, players []*SkatPlayer, config SkatConfig) *Skat {
	return &Skat{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: skatRoundState{
			winnerSide:  SkatWinnerUndecided,
			declarerIdx: -1,
		},
	}
}

// NewDefaultSkat returns Skat with the standard 3-player setup (1 human, 2 CPU).
func NewDefaultSkat() *Skat {
	players := []*SkatPlayer{
		NewSkatPlayer(true),
		NewSkatPlayer(false),
		NewSkatPlayer(false),
	}
	return NewSkat(newSkatDeck(), players, DefaultSkatConfig())
}

// newSkatDeck builds the 32-card Skat deck (7..A across the four standard suits).
//
// **NewTrumpCardsWithSuits(32, …) では作れない。** 枚数で打ち切る汎用
// コンストラクタなので、♠13 + ♣13 + ♥6 + ♦0 になり**ダイヤが 1 枚も入らない**。
// 切り札スートを選ぶゲームでこれをやると、選べるのに 1 枚も存在しない切り札が
// できる (#5296)。German 32 枚パックは Skat / Belote / Prsi で共通。
func newSkatDeck() *TrumpCards {
	return NewTrumpCards32()
}

// Reset initializes a new game session.
func (s *Skat) Reset() {
	s.round = skatRoundState{
		roundNumber: 1,
		dealerIdx:   0,
		declarerIdx: -1,
		winnerSide:  SkatWinnerUndecided,
	}

	for _, p := range s.players {
		p.SetCumulativeScore(0)
		p.SetRoundsWon(0)
		p.SetRoundsLost(0)
		p.ResetRound()
	}

	s.startRound()
}

// startRound resets per-round state and deals.
func (s *Skat) startRound() {
	for _, p := range s.players {
		p.ResetRound()
	}

	// Build a fresh deck so Reset/NextRound start consistent.
	s.trumpCards = newSkatDeck()
	s.dealCards()
	s.sortAllHands()

	// Positions: forehand = (dealer + 1) % 3, middlehand = +2, rearhand = +3.
	s.round.forehandIdx = (s.round.dealerIdx + 1) % SkatPlayerCnt
	s.round.middlehandIdx = (s.round.dealerIdx + 2) % SkatPlayerCnt
	s.round.rearhandIdx = (s.round.dealerIdx + 3) % SkatPlayerCnt

	// Bidding starts: middlehand (bidder) calls forehand (responder).
	s.round.bidderIdx = s.round.middlehandIdx
	s.round.responderIdx = s.round.forehandIdx
	s.round.bidStep = 0
	s.round.currentBid = 0
	s.round.auctionRound = 1
	s.round.round1Winner = -1
	s.round.declarerIdx = -1
	s.round.passedAtCall = [SkatPlayerCnt]bool{}
	s.round.gameType = SkatGameNone
	s.round.trumpSuit = 0
	s.round.pickedSkat = false
	s.round.declarerHand = nil
	s.round.declarerCardPts = 0
	s.round.defendersCardPts = 0
	s.round.gameValue = 0
	s.round.breakdown = nil
	s.round.winnerSide = SkatWinnerUndecided
	s.round.trickNumber = 0
	s.round.currentTrick = nil
	s.round.currentPlayerIdx = -1
	s.round.leadPlayerIdx = -1

	s.round.phase = SkatPhaseBid

	s.appendLog(-1, "round_start", fmt.Sprintf("Round %d: dealer=%s", s.round.roundNumber, playerName(s.players, s.round.dealerIdx)), nil)
}

// NextRound advances to the next round.
func (s *Skat) NextRound() {
	if s.round.phase != SkatPhaseRoundEnd {
		return
	}
	prevRound := s.round.roundNumber
	prevDealer := s.round.dealerIdx
	s.round = skatRoundState{
		roundNumber: prevRound + 1,
		dealerIdx:   (prevDealer + 1) % SkatPlayerCnt,
		declarerIdx: -1,
		winnerSide:  SkatWinnerUndecided,
	}
	s.startRound()
}

// dealCards deals 10/10/10 + 2 (skat).
func (s *Skat) dealCards() {
	s.trumpCards.Shuffle()
	// Deal 3 then 2 skat cards then 4 then 3 (a common pattern). For simplicity
	// we just deal 10 cards round-robin, then 2 to the skat.
	for range SkatHandSize {
		for i := range SkatPlayerCnt {
			c := s.trumpCards.DrawCard()
			if c != nil {
				s.players[i].AddCard(c)
			}
		}
	}
	skat := []*Card{}
	for range SkatSkatSize {
		c := s.trumpCards.DrawCard()
		if c != nil {
			skat = append(skat, c)
		}
	}
	s.round.skat = skat
}

// SkatBidLadder is the simplified bid ladder used by this implementation.
// Real Skat uses 18, 20, 22, 23, 24, 27, 30, 33, 35, 36, ... — here we adopt a
// reduced subset that still preserves the auction's structure.
var SkatBidLadder = []int{18, 20, 22, 24, 27, 30, 33, 36, 40, 44, 48, 50, 60, 72}

// PlayerBid (bidder) calls the next ladder value.
// holdPrev=true keeps the current call alive (forehand "yes"); holdPrev=false passes.
func (s *Skat) PlayerBid(accept bool) error {
	if s.round.gameEndFlag {
		return ErrGameEnded
	}
	if s.round.phase != SkatPhaseBid {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(s.players)
	if humanIdx < 0 {
		return ErrNotHumanTurn
	}
	if !s.IsHumanBidTurn() {
		return ErrNotHumanTurn
	}
	s.applyBidStep(s.activeBidActorIdx(), accept)
	return nil
}

// CpuBid lets the active CPU bidder/responder act.
func (s *Skat) CpuBid() {
	if s.round.gameEndFlag || s.round.phase != SkatPhaseBid {
		return
	}
	idx := s.activeBidActorIdx()
	if idx < 0 || s.players[idx].GetIsHuman() {
		return
	}
	accept := s.cpuBidDecision(idx)
	s.applyBidStep(idx, accept)
}

// activeBidActorIdx returns the player index whose response is currently expected.
//   - If currentBid == 0 and bidStep == 0: middlehand calls 18 (or passes).
//   - Otherwise: the responder must accept or pass; if they pass the bidder
//     advances the ladder. After middlehand vs forehand resolves, rearhand
//     comes in against the survivor.
func (s *Skat) activeBidActorIdx() int {
	if s.round.declarerIdx >= 0 {
		return -1
	}
	if s.round.currentBid == 0 {
		return s.round.bidderIdx
	}
	return s.round.responderIdx
}

// IsHumanBidTurn reports whether the active bidder/responder is human.
func (s *Skat) IsHumanBidTurn() bool {
	idx := s.activeBidActorIdx()
	if idx < 0 || idx >= len(s.players) {
		return false
	}
	return s.players[idx].GetIsHuman()
}

// applyBidStep applies a yes/pass step from the active actor.
//
// Auction structure:
//   - Round 1: middlehand calls; forehand responds.
//   - Round 2: rearhand calls; the survivor of round 1 responds.
//   - When a bidder passes outright (currentBid==0), they drop out and the
//     responder is the round survivor. When a responder passes (currentBid>0),
//     the bidder is the round survivor.
//   - The declarer is the survivor of round 2, except when both rear and
//     middle passed without ever calling — then forehand is declarer at 18 if
//     someone called previously, otherwise the hand is passed out.
func (s *Skat) applyBidStep(actorIdx int, accept bool) {
	if actorIdx < 0 {
		return
	}
	if s.round.currentBid == 0 {
		// Bidder turn: call ladder[0] or pass.
		if accept {
			s.round.currentBid = SkatBidLadder[0]
			s.appendLog(actorIdx, "bid_call",
				fmt.Sprintf("%s calls %d", playerName(s.players, actorIdx), s.round.currentBid), nil)
			s.round.bidStep = 1
			return
		}
		// Bidder passed without calling.
		s.round.passedAtCall[actorIdx] = true
		s.appendLog(actorIdx, "bid_pass",
			fmt.Sprintf("%s passes", playerName(s.players, actorIdx)), nil)
		s.bidderDroppedOut()
		return
	}

	// Responder turn.
	if accept {
		s.appendLog(actorIdx, "bid_yes",
			fmt.Sprintf("%s answers yes at %d", playerName(s.players, actorIdx), s.round.currentBid), nil)
		if s.round.bidStep < len(SkatBidLadder) {
			s.round.currentBid = SkatBidLadder[s.round.bidStep]
			s.appendLog(s.round.bidderIdx, "bid_call",
				fmt.Sprintf("%s calls %d", playerName(s.players, s.round.bidderIdx), s.round.currentBid), nil)
			s.round.bidStep++
			return
		}
		// Ladder exhausted: bidder wins by default.
		s.responderDroppedOut()
		return
	}

	// Responder passes — bidder is round survivor.
	s.round.passedAtCall[actorIdx] = true
	s.appendLog(actorIdx, "bid_pass",
		fmt.Sprintf("%s passes", playerName(s.players, actorIdx)), nil)
	s.responderDroppedOut()
}

// bidderDroppedOut handles the bidder passing without calling.
func (s *Skat) bidderDroppedOut() {
	if s.round.auctionRound == 1 {
		// Forehand is round 1 survivor; rearhand challenges fore in round 2.
		s.round.round1Winner = s.round.forehandIdx
		s.round.auctionRound = 2
		s.round.bidderIdx = s.round.rearhandIdx
		s.round.responderIdx = s.round.forehandIdx
		s.round.bidStep = 0
		s.round.currentBid = 0
		return
	}
	// Round 2: rearhand passed without calling.
	if s.round.currentBid == 0 {
		// No call ever happened in round 2. Survivor of round 1 is declarer
		// at 18 if any bid happened anywhere; else passed-out.
		s.round.currentBid = SkatBidLadder[0]
		s.declareDeclarer(s.round.round1Winner)
		return
	}
	s.declareDeclarer(s.round.round1Winner)
}

// responderDroppedOut handles the responder passing — bidder is round survivor.
func (s *Skat) responderDroppedOut() {
	if s.round.auctionRound == 1 {
		s.round.round1Winner = s.round.bidderIdx
		s.round.auctionRound = 2
		s.round.bidderIdx = s.round.rearhandIdx
		s.round.responderIdx = s.round.round1Winner
		s.round.bidStep = s.bidStepForCurrent()
		return
	}
	// Round 2: rearhand wins.
	s.declareDeclarer(s.round.bidderIdx)
}

// bidStepForCurrent returns the next ladder step that exceeds currentBid.
func (s *Skat) bidStepForCurrent() int {
	for i, b := range SkatBidLadder {
		if b > s.round.currentBid {
			return i
		}
	}
	return len(SkatBidLadder)
}

// declareDeclarer locks the declarer in and advances to the skat-pickup phase.
func (s *Skat) declareDeclarer(idx int) {
	s.round.declarerIdx = idx
	s.players[idx].SetIsDeclarer(true)
	s.players[idx].SetBid(s.round.currentBid)
	s.appendLog(idx, "declarer",
		fmt.Sprintf("%s wins the auction at %d", playerName(s.players, idx), s.round.currentBid), nil)
	s.round.phase = SkatPhaseSkatPickup
}

// PlayerPickSkat declarer picks the skat (true) or plays a hand game (false).
func (s *Skat) PlayerPickSkat(pickup bool) error {
	if s.round.gameEndFlag {
		return ErrGameEnded
	}
	if s.round.phase != SkatPhaseSkatPickup {
		return ErrWrongPhase
	}
	if !s.IsHumanDeclarerTurn() {
		return ErrNotHumanTurn
	}
	s.applyPickSkat(pickup)
	return nil
}

// CpuPickSkat CPU declarer picks the skat.
func (s *Skat) CpuPickSkat() {
	if s.round.gameEndFlag || s.round.phase != SkatPhaseSkatPickup {
		return
	}
	if s.round.declarerIdx < 0 || s.players[s.round.declarerIdx].GetIsHuman() {
		return
	}
	pickup := s.cpuPickSkatDecision()
	s.applyPickSkat(pickup)
}

// applyPickSkat applies the pickup decision.
func (s *Skat) applyPickSkat(pickup bool) {
	declarer := s.players[s.round.declarerIdx]
	s.round.pickedSkat = pickup
	s.round.originalSkat = append([]*Card{}, s.round.skat...)

	if pickup {
		for _, c := range s.round.skat {
			declarer.AddCard(c)
		}
		s.sortHand(declarer)
		s.appendLog(s.round.declarerIdx, "pick_skat",
			fmt.Sprintf("%s picks up the skat", playerName(s.players, s.round.declarerIdx)), s.round.skat)
		s.round.skat = nil
		s.round.phase = SkatPhaseDiscard
		return
	}

	// Hand game — skat stays face-down; go straight to game declaration.
	s.appendLog(s.round.declarerIdx, "hand_game",
		fmt.Sprintf("%s plays a hand game", playerName(s.players, s.round.declarerIdx)), nil)
	s.round.phase = SkatPhaseGameDeclaration
}

// PlayerDiscard declarer discards two cards into the skat.
func (s *Skat) PlayerDiscard(idxA, idxB int) error {
	if s.round.gameEndFlag {
		return ErrGameEnded
	}
	if s.round.phase != SkatPhaseDiscard {
		return ErrWrongPhase
	}
	if !s.IsHumanDeclarerTurn() {
		return ErrNotHumanTurn
	}
	if idxA == idxB {
		return NewDomainError(ErrInvalidPlay, "Discard indices must differ.")
	}
	declarer := s.players[s.round.declarerIdx]
	if idxA < 0 || idxA >= declarer.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "Discard index out of range.")
	}
	if idxB < 0 || idxB >= declarer.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "Discard index out of range.")
	}
	s.applyDiscard(idxA, idxB)
	return nil
}

// CpuDiscard CPU declarer discards two cards.
func (s *Skat) CpuDiscard() {
	if s.round.gameEndFlag || s.round.phase != SkatPhaseDiscard {
		return
	}
	if s.round.declarerIdx < 0 || s.players[s.round.declarerIdx].GetIsHuman() {
		return
	}
	a, b := s.cpuPickDiscards()
	s.applyDiscard(a, b)
}

// applyDiscard performs the discard.
func (s *Skat) applyDiscard(idxA, idxB int) {
	declarer := s.players[s.round.declarerIdx]
	// Remove the higher index first to keep the other index valid.
	hi, lo := idxA, idxB
	if hi < lo {
		hi, lo = lo, hi
	}
	cardHi := declarer.RemoveCard(hi)
	cardLo := declarer.RemoveCard(lo)
	s.round.skat = []*Card{cardLo, cardHi}
	s.sortHand(declarer)
	s.appendLog(s.round.declarerIdx, "discard",
		fmt.Sprintf("%s discards 2 cards into the skat", playerName(s.players, s.round.declarerIdx)),
		[]*Card{cardLo, cardHi})
	s.round.phase = SkatPhaseGameDeclaration
}

// PlayerDeclareGame declarer chooses the game type (and trump suit for suit games).
func (s *Skat) PlayerDeclareGame(gameType SkatGameType, trumpSuit int) error {
	if s.round.gameEndFlag {
		return ErrGameEnded
	}
	if s.round.phase != SkatPhaseGameDeclaration {
		return ErrWrongPhase
	}
	if !s.IsHumanDeclarerTurn() {
		return ErrNotHumanTurn
	}
	if gameType != SkatGameSuit && gameType != SkatGameGrand && gameType != SkatGameNull {
		return NewDomainError(ErrInvalidPlay, "Invalid game type.")
	}
	if gameType == SkatGameSuit {
		if trumpSuit < CardDesignSpade || trumpSuit > CardDesignDiamond {
			return NewDomainError(ErrInvalidPlay, "Invalid trump suit.")
		}
	}
	s.applyGameDeclaration(gameType, trumpSuit)
	return nil
}

// CpuDeclareGame CPU declarer chooses a game.
func (s *Skat) CpuDeclareGame() {
	if s.round.gameEndFlag || s.round.phase != SkatPhaseGameDeclaration {
		return
	}
	if s.round.declarerIdx < 0 || s.players[s.round.declarerIdx].GetIsHuman() {
		return
	}
	gt, trump := s.cpuPickGame(s.round.declarerIdx)
	s.applyGameDeclaration(gt, trump)
}

// applyGameDeclaration locks in the game type and starts play.
func (s *Skat) applyGameDeclaration(gt SkatGameType, trumpSuit int) {
	s.round.gameType = gt
	if gt == SkatGameSuit {
		s.round.trumpSuit = trumpSuit
	} else {
		s.round.trumpSuit = 0
	}
	s.appendLog(s.round.declarerIdx, "declare_game",
		fmt.Sprintf("%s declares %s", playerName(s.players, s.round.declarerIdx), s.gameTypeName()), nil)
	s.startPlay()
}

// startPlay begins the trick-play phase.
func (s *Skat) startPlay() {
	s.round.leadPlayerIdx = s.round.forehandIdx
	s.round.currentPlayerIdx = s.round.forehandIdx
	s.round.trickNumber = 1
	s.round.currentTrick = nil
	// Snapshot the declarer's full 10-card hand so matadors can be counted
	// at scoring time, after the hand has been played out.
	if s.round.declarerIdx >= 0 {
		declarer := s.players[s.round.declarerIdx]
		hand := make([]*Card, 0, declarer.GetCardsSize())
		for i := 0; i < declarer.GetCardsSize(); i++ {
			hand = append(hand, declarer.GetCard(i))
		}
		s.round.declarerHand = hand
	}
	s.round.phase = SkatPhasePlay
}

// PlayerPlay human plays a card.
func (s *Skat) PlayerPlay(cardIndex int) error {
	if s.round.gameEndFlag {
		return ErrGameEnded
	}
	if s.round.phase != SkatPhasePlay {
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
	return nil
}

// CpuPlay current CPU plays a card.
func (s *Skat) CpuPlay() {
	if s.round.gameEndFlag || s.round.phase != SkatPhasePlay {
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
func (s *Skat) playCard(playerIdx int, card *Card) {
	s.round.currentTrick = append(s.round.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	s.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(s.players, playerIdx), cardStr(card)), []*Card{card})
	if len(s.round.currentTrick) == SkatPlayerCnt {
		s.round.phase = SkatPhaseTrickEnd
		return
	}
	s.round.currentPlayerIdx = (s.round.currentPlayerIdx + 1) % SkatPlayerCnt
}

// ResolveTrick determines the trick winner and updates points.
func (s *Skat) ResolveTrick() {
	if s.round.phase != SkatPhaseTrickEnd || len(s.round.currentTrick) != SkatPlayerCnt {
		return
	}
	winnerIdx := s.trickWinner()
	cards := make([]*Card, len(s.round.currentTrick))
	pts := 0
	for i, tc := range s.round.currentTrick {
		cards[i] = tc.Card
		pts += skatCardPoints(tc.Card)
	}
	s.players[winnerIdx].AddTrick(cards)
	s.players[winnerIdx].SetCardPoints(s.players[winnerIdx].GetCardPoints() + pts)
	s.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d card points)", playerName(s.players, winnerIdx), s.round.trickNumber, pts), cards)
	s.round.leadPlayerIdx = winnerIdx
	if s.round.trickNumber >= SkatTricksPerRound {
		s.round.phase = SkatPhaseRoundEnd
		return
	}
	s.round.phase = SkatPhaseTrickEnd
}

// NextTrick begins the next trick.
func (s *Skat) NextTrick() {
	if s.round.phase != SkatPhaseTrickEnd {
		return
	}
	s.round.currentTrick = nil
	s.round.currentPlayerIdx = s.round.leadPlayerIdx
	s.round.trickNumber++
	s.round.phase = SkatPhasePlay
}

// ScoreRound finalises the round score and checks for game end.
func (s *Skat) ScoreRound() {
	if s.round.phase != SkatPhaseRoundEnd {
		return
	}
	if s.round.declarerIdx < 0 {
		// Passed-out hand: nothing to score.
		s.checkGameEnd()
		return
	}

	declarerIdx := s.round.declarerIdx
	declarer := s.players[declarerIdx]

	// Count card points.
	declPts := declarer.GetCardPoints()
	if s.round.gameType != SkatGameNull {
		// Skat goes to the declarer.
		for _, c := range s.round.skat {
			declPts += skatCardPoints(c)
		}
	}
	defPts := 120 - declPts
	s.round.declarerCardPts = declPts
	s.round.defendersCardPts = defPts

	gameValue, declarerWon := s.computeRoundResult()
	s.round.gameValue = gameValue

	if declarerWon {
		s.round.winnerSide = SkatWinnerDeclarer
		declarer.SetRoundScore(gameValue)
		declarer.IncRoundsWon()
		s.appendLog(declarerIdx, "round_result",
			fmt.Sprintf("%s wins (+%d)", playerName(s.players, declarerIdx), gameValue), nil)
	} else {
		s.round.winnerSide = SkatWinnerDefenders
		declarer.SetRoundScore(-gameValue)
		declarer.IncRoundsLost()
		s.appendLog(declarerIdx, "round_result",
			fmt.Sprintf("%s loses (-%d)", playerName(s.players, declarerIdx), gameValue), nil)
	}

	declarer.CommitRoundScore()
	s.appendLog(declarerIdx, "cumulative_score",
		fmt.Sprintf("%s total=%d", playerName(s.players, declarerIdx), declarer.GetCumulativeScore()), nil)

	s.checkGameEnd()
}

// computeRoundResult computes the simplified Skat game value and outcome.
func (s *Skat) computeRoundResult() (int, bool) {
	if s.round.declarerIdx < 0 {
		return 0, false
	}
	if s.round.gameType == SkatGameNull {
		// Declarer must lose every trick.
		declarerLostNoTricks := s.players[s.round.declarerIdx].GetTrickCount() == 0
		// Simplified Null base value.
		base := 23
		s.round.breakdown = &SkatScoreBreakdown{Base: base, Multiplier: 1, Value: base, Null: true}
		if !declarerLostNoTricks {
			return base, false
		}
		return base, true
	}

	base := s.gameBaseValue()
	matadors := s.matadorsCount(s.round.declarerHand)
	multiplier := s.gameMultiplier()
	declPts := s.round.declarerCardPts

	won := declPts >= 61
	bd := &SkatScoreBreakdown{Base: base, Matadors: matadors, Hand: !s.round.pickedSkat}
	var value int
	// Schneider/Schwarz bonuses (simplified): +1 if 90+ pts (Schneider), +1 if all 10 tricks (Schwarz).
	if won {
		if declPts >= 90 {
			multiplier++
			bd.Schneider = true
		}
		if s.players[s.round.declarerIdx].GetTrickCount() == SkatTricksPerRound {
			multiplier++
			bd.Schwarz = true
		}
		value = base * multiplier
	} else {
		// Loser pays double the game value.
		value = base * multiplier * 2
		bd.Doubled = true
	}
	bd.Multiplier = multiplier
	bd.Value = value
	s.round.breakdown = bd

	// If the declarer overbid (final value < bid), they lose by twice the lowest
	// multiple of base that meets the bid (simplified to bid * 2).
	if won && value < s.round.currentBid {
		bd.Overbid = true
		bd.Bid = s.round.currentBid
		bd.Value = s.round.currentBid * 2
		return bd.Value, false
	}

	return value, won
}

// SkatScoreBreakdown はラウンド得点がどう積み上がったかの内訳。
//
// **なぜ 33 点で、別のラウンドは 66 点なのか。**マタドール (切り札の連続所持/
// 不所持) はスカートで最も分かりにくい規則なのに、どちらの UI も最終値しか
// 出していなかった (#5561)。ここに残しておけば、表示側が計算をやり直さずに済む。
type SkatScoreBreakdown struct {
	// Base は基礎点 (スート 9〜12 / グランド 24 / ヌル 23)。
	Base int
	// Matadors はマタドール数。乗数の出発点。
	Matadors int
	// Multiplier は最終的な乗数 (マタドール+1、ハンド・シュナイダー・シュヴァルツで加算)。
	Multiplier int
	// Hand / Schneider / Schwarz はそれぞれ乗数に +1 したか。
	Hand      bool
	Schneider bool
	Schwarz   bool
	// Doubled は敗北による 2 倍。Overbid はオーバービッドで bid*2 に置き換わったこと。
	//
	// **どちらも Base*Multiplier では最終得点にならない。**敗北は 2 倍、
	// オーバービッドは基礎点と無関係な bid*2 に置き換わる。表示側はこの 2 つを
	// 見て式そのものを変える必要がある (#5561 のレビュー指摘)。
	Doubled bool
	Overbid bool
	// Bid はそのラウンドの最終入札。Overbid のときだけ意味を持つ (Value = Bid*2)。
	Bid int
	// Value は最終得点。GetGameValue() と必ず一致する。
	Value int
	// Null はヌル契約 (乗数の概念が無い)。
	Null bool
}

// GetScoreBreakdown は直近ラウンドの得点内訳を返す。ラウンドが終わっていなければ nil。
func (s *Skat) GetScoreBreakdown() *SkatScoreBreakdown {
	return s.round.breakdown
}

// gameBaseValue returns the base value for the game type.
func (s *Skat) gameBaseValue() int {
	return skatBaseValueFor(s.round.gameType, s.round.trumpSuit)
}

// skatBaseValueFor は仮の契約に対する基礎点を返す。
func skatBaseValueFor(gameType SkatGameType, trumpSuit int) int {
	switch gameType {
	case SkatGameSuit:
		switch trumpSuit {
		case CardDesignDiamond:
			return 9
		case CardDesignHeart:
			return 10
		case CardDesignSpade:
			return 11
		case CardDesignClover:
			return 12
		}
	case SkatGameGrand:
		return 24
	}
	return 0
}

// gameMultiplier computes a simplified multiplier = matadors + 1 + (hand bonus).
func (s *Skat) gameMultiplier() int {
	matadors := s.matadorsCount(s.round.declarerHand)
	mult := matadors + 1
	if !s.round.pickedSkat {
		mult++ // hand bonus
	}
	return mult
}

// matadorsCount counts the consecutive trumps from the top of the trump order
// the declarer holds (with) or lacks (without). See Skat rules.
//
// The cards slice is the declarer's hand as it stood at the start of the play
// phase — we cannot read the live SkatPlayer hand here because matadors is
// computed at scoring time, after every card has been played out.
func (s *Skat) matadorsCount(cards []*Card) int {
	return skatMatadorsFor(cards, s.round.gameType, s.round.trumpSuit)
}

// skatMatadorsFor は仮の契約に対するマタドール数を返す。
func skatMatadorsFor(cards []*Card, gameType SkatGameType, trumpSuit int) int {
	order := skatTrumpOrderFor(gameType, trumpSuit)
	if len(order) == 0 {
		return 0
	}
	hand := map[[2]int]bool{}
	for _, c := range cards {
		if c == nil {
			continue
		}
		hand[[2]int{c.GetDesign(), c.GetValue()}] = true
	}
	// "With" matadors: consecutive top trumps the declarer holds.
	with := 0
	hasTop := hand[[2]int{order[0].design, order[0].value}]
	if hasTop {
		for _, t := range order {
			if hand[[2]int{t.design, t.value}] {
				with++
			} else {
				break
			}
		}
		return with
	}
	// "Without" matadors: consecutive top trumps the declarer lacks.
	without := 0
	for _, t := range order {
		if !hand[[2]int{t.design, t.value}] {
			without++
		} else {
			break
		}
	}
	return without
}

// trumpOrderEntry helper struct describing a trump card.
type trumpOrderEntry struct {
	design int
	value  int
}

// trumpOrder returns the trump cards in descending strength.
func (s *Skat) trumpOrder() []trumpOrderEntry {
	return skatTrumpOrderFor(s.round.gameType, s.round.trumpSuit)
}

// skatTrumpOrderFor は仮の契約に対する切札の序列を返す。
//
// **ビッド前の見積もりにも同じ序列が要る。**round の状態に縛られたままだと、
// 「この手札ならどの契約でいくつまで受けられるか」を計算できない (#4905)。
func skatTrumpOrderFor(gameType SkatGameType, trumpSuit int) []trumpOrderEntry {
	switch gameType {
	case SkatGameNull:
		return nil
	case SkatGameGrand:
		// Grand: jacks only, in suit order Clubs > Spades > Hearts > Diamonds.
		return []trumpOrderEntry{
			{CardDesignClover, skatValueJack},
			{CardDesignSpade, skatValueJack},
			{CardDesignHeart, skatValueJack},
			{CardDesignDiamond, skatValueJack},
		}
	case SkatGameSuit:
		out := []trumpOrderEntry{
			{CardDesignClover, skatValueJack},
			{CardDesignSpade, skatValueJack},
			{CardDesignHeart, skatValueJack},
			{CardDesignDiamond, skatValueJack},
		}
		ts := trumpSuit
		// Then suit cards in standard high→low order: A, T, K, Q, 9, 8, 7.
		order := []int{skatValueAce, skatValueTen, skatValueKing, skatValueQueen, skatValueNine, skatValueEight, skatValueSeven}
		for _, v := range order {
			out = append(out, trumpOrderEntry{ts, v})
		}
		return out
	}
	return nil
}

// validatePlay verifies that a card is legal to play.
func (s *Skat) validatePlay(playerIdx int, card *Card) error {
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

// effectiveSuit returns the suit a card belongs to (jacks are part of the
// trump "suit" in suit/grand games; in null games suits are literal).
func (s *Skat) effectiveSuit(c *Card) int {
	if s.round.gameType != SkatGameNull && c.GetValue() == skatValueJack {
		return -1 // marker for trump-suit
	}
	return c.GetDesign()
}

// isTrump reports whether the card is a trump.
func (s *Skat) isTrump(c *Card) bool {
	switch s.round.gameType {
	case SkatGameNull:
		return false
	case SkatGameGrand:
		return c.GetValue() == skatValueJack
	case SkatGameSuit:
		if c.GetValue() == skatValueJack {
			return true
		}
		return c.GetDesign() == s.round.trumpSuit
	}
	return false
}

// playerHasTrump reports whether the player holds any trump.
func (s *Skat) playerHasTrump(playerIdx int) bool {
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
func (s *Skat) playerHasSuitNonTrump(playerIdx int, suit int) bool {
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
func (s *Skat) trickWinner() int {
	if len(s.round.currentTrick) == 0 {
		return 0
	}
	if s.round.gameType == SkatGameNull {
		// Highest card of lead suit wins (null ranking: 7,8,9,10,J,Q,K,A).
		leadSuit := s.round.currentTrick[0].Card.GetDesign()
		winner := s.round.currentTrick[0]
		winnerVal := nullRank(s.round.currentTrick[0].Card)
		for _, tc := range s.round.currentTrick[1:] {
			if tc.Card.GetDesign() != leadSuit {
				continue
			}
			v := nullRank(tc.Card)
			if v > winnerVal {
				winnerVal = v
				winner = tc
			}
		}
		return winner.PlayerIdx
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
func (s *Skat) cardStrength(c *Card) int {
	if s.round.gameType == SkatGameNull {
		return nullRank(c)
	}
	// Trumps use the trump-order index.
	order := s.trumpOrder()
	for i, t := range order {
		if t.design == c.GetDesign() && t.value == c.GetValue() {
			return 100 - i // higher = stronger
		}
	}
	// Non-trump in suit/grand games: standard high-to-low order.
	switch c.GetValue() {
	case skatValueAce:
		return 7
	case skatValueTen:
		return 6
	case skatValueKing:
		return 5
	case skatValueQueen:
		return 4
	case skatValueNine:
		return 3
	case skatValueEight:
		return 2
	case skatValueSeven:
		return 1
	}
	return 0
}

// nullRank returns the null-game rank: 7,8,9,T,J,Q,K,A → 1..8.
func nullRank(c *Card) int {
	switch c.GetValue() {
	case skatValueSeven:
		return 1
	case skatValueEight:
		return 2
	case skatValueNine:
		return 3
	case skatValueTen:
		return 4
	case skatValueJack:
		return 5
	case skatValueQueen:
		return 6
	case skatValueKing:
		return 7
	case skatValueAce:
		return 8
	}
	return 0
}

// skatCardPoints returns the card-point value used to score tricks.
func skatCardPoints(c *Card) int {
	switch c.GetValue() {
	case skatValueAce:
		return 11
	case skatValueTen:
		return 10
	case skatValueKing:
		return 4
	case skatValueQueen:
		return 3
	case skatValueJack:
		return 2
	}
	return 0
}

// checkGameEnd determines whether the game should end.
func (s *Skat) checkGameEnd() {
	for _, p := range s.players {
		if p.GetCumulativeScore() >= s.config.TargetScore {
			s.round.gameEndFlag = true
			s.round.phase = SkatPhaseGameEnd
			s.appendLog(-1, "game_end",
				fmt.Sprintf("%s reaches %d points and wins!", playerName(s.players, s.findIndex(p)), p.GetCumulativeScore()), nil)
			return
		}
	}
}

// findIndex returns the player's index.
func (s *Skat) findIndex(p *SkatPlayer) int {
	for i, q := range s.players {
		if p == q {
			return i
		}
	}
	return -1
}

// IsHumanTurn reports whether the current player is human (play phase).
func (s *Skat) IsHumanTurn() bool {
	return isHumanTurn(s.players, s.round.currentPlayerIdx)
}

// IsHumanDeclarerTurn reports whether the declarer is human (used for skat
// pickup, discard, and game declaration phases).
func (s *Skat) IsHumanDeclarerTurn() bool {
	if s.round.declarerIdx < 0 {
		return false
	}
	return s.players[s.round.declarerIdx].GetIsHuman()
}

// GetValidPlayIndices returns the indices of legally playable cards.
func (s *Skat) GetValidPlayIndices(playerIdx int) []int {
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

// GetHint returns a hint for the human player based on the current phase.
func (s *Skat) GetHint() *SkatHint {
	humanIdx := findHumanIdx(s.players)
	if humanIdx < 0 {
		return nil
	}
	switch s.round.phase {
	case SkatPhaseBid:
		if s.activeBidActorIdx() != humanIdx {
			return nil
		}
		accept := s.cpuBidDecision(humanIdx)
		val := 0
		if accept {
			val = 1
		}
		return &SkatHint{Bid: &val, Reason: "strategic_bid"}
	case SkatPhaseSkatPickup:
		if s.round.declarerIdx != humanIdx {
			return nil
		}
		pick := s.cpuPickSkatDecision()
		return &SkatHint{PickSkat: &pick, Reason: "skat_pickup"}
	case SkatPhaseDiscard:
		if s.round.declarerIdx != humanIdx {
			return nil
		}
		a, _ := s.cpuPickDiscards()
		return &SkatHint{DiscardIndex: &a, Reason: "discard_low"}
	case SkatPhaseGameDeclaration:
		if s.round.declarerIdx != humanIdx {
			return nil
		}
		gt, trump := s.cpuPickGame(humanIdx)
		gtInt := int(gt)
		return &SkatHint{GameType: &gtInt, TrumpSuit: &trump, Reason: "game_choice"}
	case SkatPhasePlay:
		if s.round.currentPlayerIdx != humanIdx {
			return nil
		}
		valid := s.GetValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := s.cpuPickPlay(humanIdx)
		return &SkatHint{CardIndex: &idx, Reason: "best_play"}
	}
	return nil
}

// gameTypeName returns the human-readable game type.
func (s *Skat) gameTypeName() string {
	switch s.round.gameType {
	case SkatGameSuit:
		return fmt.Sprintf("Suit (trump=%s)", skatSuitName(s.round.trumpSuit))
	case SkatGameGrand:
		return "Grand"
	case SkatGameNull:
		return "Null"
	}
	return "None"
}

// skatSuitName returns the English suit name.
func skatSuitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "Spades"
	case CardDesignClover:
		return "Clubs"
	case CardDesignHeart:
		return "Hearts"
	case CardDesignDiamond:
		return "Diamonds"
	}
	return "?"
}

// appendLog appends an entry to the round action log.
func (s *Skat) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.round.appendLog(playerIdx, actionType, detail, cards)
}

// sortAllHands sorts every player's hand.
func (s *Skat) sortAllHands() {
	sortEachHand(s.players, s.sortHand)
}

// sortHand sorts the player's hand by suit then value.
func (s *Skat) sortHand(p *SkatPlayer) {
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
func (s *Skat) GetPhase() SkatPhase { return s.round.phase }

// SetPhase sets the phase (test only).
func (s *Skat) SetPhase(p SkatPhase) { s.round.phase = p }

// GetRoundNumber returns the round number.
func (s *Skat) GetRoundNumber() int { return s.round.roundNumber }

// GetTrickNumber returns the current trick number.
func (s *Skat) GetTrickNumber() int { return s.round.trickNumber }

// GetCurrentPlayerIdx returns the current player index.
func (s *Skat) GetCurrentPlayerIdx() int { return s.round.currentPlayerIdx }

// SetCurrentPlayerIdx sets the current player index (test only).
func (s *Skat) SetCurrentPlayerIdx(idx int) { s.round.currentPlayerIdx = idx }

// GetCurrentTrick returns the current trick (in play order).
func (s *Skat) GetCurrentTrick() []*TrickCard { return s.round.currentTrick }

// SetCurrentTrick sets the current trick (test only).
func (s *Skat) SetCurrentTrick(trick []*TrickCard) { s.round.currentTrick = trick }

// GetForehandIdx returns the forehand index.
func (s *Skat) GetForehandIdx() int { return s.round.forehandIdx }

// GetMiddlehandIdx returns the middlehand index.
func (s *Skat) GetMiddlehandIdx() int { return s.round.middlehandIdx }

// GetRearhandIdx returns the rearhand index.
func (s *Skat) GetRearhandIdx() int { return s.round.rearhandIdx }

// GetDealerIdx returns the dealer index.
func (s *Skat) GetDealerIdx() int { return s.round.dealerIdx }

// GetDeclarerIdx returns the declarer index (-1 if undetermined).
func (s *Skat) GetDeclarerIdx() int { return s.round.declarerIdx }

// SetDeclarerIdx sets the declarer index (test only).
func (s *Skat) SetDeclarerIdx(idx int) { s.round.declarerIdx = idx }

// GetCurrentBid returns the most recent bid call.
func (s *Skat) GetCurrentBid() int { return s.round.currentBid }

// GetActiveBidActorIdx returns the player whose response is currently expected.
func (s *Skat) GetActiveBidActorIdx() int { return s.activeBidActorIdx() }

// GetGameType returns the chosen game type.
func (s *Skat) GetGameType() SkatGameType { return s.round.gameType }

// GetTrumpSuit returns the chosen trump suit (suit games).
func (s *Skat) GetTrumpSuit() int { return s.round.trumpSuit }

// GetSkat returns the current skat (face-down before pickup, post-discard
// otherwise; nil after pickup before discard).
func (s *Skat) GetSkat() []*Card { return s.round.skat }

// GetOriginalSkat returns the pre-pickup skat snapshot.
func (s *Skat) GetOriginalSkat() []*Card { return s.round.originalSkat }

// GetDeclarerCardPoints returns declarer card points (post-round only).
func (s *Skat) GetDeclarerCardPoints() int { return s.round.declarerCardPts }

// GetDefendersCardPoints returns defenders card points (post-round only).
func (s *Skat) GetDefendersCardPoints() int { return s.round.defendersCardPts }

// GetWinnerSide returns the round outcome.
func (s *Skat) GetWinnerSide() int { return s.round.winnerSide }

// GetGameValue returns the round game value.
func (s *Skat) GetGameValue() int { return s.round.gameValue }

// GetGameEndFlag returns the game-end flag.
func (s *Skat) GetGameEndFlag() bool { return s.round.gameEndFlag }

// GetPlayerCnt returns the player count.
func (s *Skat) GetPlayerCnt() int { return len(s.players) }

// GetPlayer returns the i-th player (nil if out of range).
func (s *Skat) GetPlayer(i int) *SkatPlayer {
	return getPlayer(s.players, i)
}

// GetLeadPlayerIdx returns the lead player index.
func (s *Skat) GetLeadPlayerIdx() int { return s.round.leadPlayerIdx }

// GetConfig returns the config.
func (s *Skat) GetConfig() SkatConfig { return s.config }

// SetConfig sets the config.
func (s *Skat) SetConfig(c SkatConfig) { s.config = c }

// GetActionLog returns the action log.
func (s *Skat) GetActionLog() []*ActionLogEntry { return s.round.actionLog }

// PickedSkat reports whether the declarer picked up the skat.
func (s *Skat) PickedSkat() bool { return s.round.pickedSkat }

// --- CPU AI ---

// cpuBidDecision returns whether the CPU should accept the next bid step.
func (s *Skat) cpuBidDecision(playerIdx int) bool {
	strength := s.handStrength(playerIdx)
	current := s.round.currentBid
	step := SkatBidLadder[0]
	if s.round.bidStep < len(SkatBidLadder) {
		step = SkatBidLadder[s.round.bidStep]
	}
	threshold := step
	if current == 0 {
		threshold = SkatBidLadder[0]
	}
	switch s.config.CpuDifficulty {
	case SkatCpuDifficultyEasy:
		return strength*2 >= threshold && rand.Intn(2) == 0
	case SkatCpuDifficultyHard:
		return strength*2 >= threshold-2
	default:
		return strength*2 >= threshold
	}
}

// handStrength returns a coarse hand strength estimate (0..40).
func (s *Skat) handStrength(playerIdx int) int {
	p := s.players[playerIdx]
	jacks := 0
	suitCounts := map[int]int{}
	tens, aces := 0, 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetValue() == skatValueJack {
			jacks++
			continue
		}
		suitCounts[c.GetDesign()]++
		switch c.GetValue() {
		case skatValueAce:
			aces++
		case skatValueTen:
			tens++
		}
	}
	maxSuit := 0
	for _, n := range suitCounts {
		if n > maxSuit {
			maxSuit = n
		}
	}
	return jacks*5 + maxSuit*2 + aces*2 + tens
}

// cpuPickSkatDecision decides whether to pick up the skat (always true except
// when the hand is overwhelmingly strong; simplified to always pick up).
func (s *Skat) cpuPickSkatDecision() bool {
	strength := s.handStrength(s.round.declarerIdx)
	// Hard CPU plays hand games occasionally with very strong hands.
	if s.config.CpuDifficulty == SkatCpuDifficultyHard && strength >= 30 {
		return false
	}
	return true
}

// cpuPickDiscards picks the two lowest-trick-value cards (simple heuristic).
func (s *Skat) cpuPickDiscards() (int, int) {
	declarer := s.players[s.round.declarerIdx]
	type cardScore struct {
		idx   int
		score int
	}
	scored := make([]cardScore, declarer.GetCardsSize())
	for i := 0; i < declarer.GetCardsSize(); i++ {
		c := declarer.GetCard(i)
		scored[i] = cardScore{idx: i, score: skatCardPoints(c) + cardWeakness(c)}
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score < scored[j].score
	})
	return scored[0].idx, scored[1].idx
}

// cardWeakness returns a tiny bias so 7/8/9 cards are preferred for discard.
func cardWeakness(c *Card) int {
	switch c.GetValue() {
	case skatValueSeven:
		return -3
	case skatValueEight:
		return -2
	case skatValueNine:
		return -1
	}
	return 0
}

// cpuPickGame picks the best game type/trump suit for the declarer.
func (s *Skat) cpuPickGame(playerIdx int) (SkatGameType, int) {
	p := s.players[playerIdx]
	suitCounts := map[int]int{}
	jacks := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetValue() == skatValueJack {
			jacks++
			continue
		}
		suitCounts[c.GetDesign()]++
	}
	if jacks >= 3 {
		return SkatGameGrand, 0
	}
	bestSuit := CardDesignSpade
	maxCount := -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if suitCounts[suit] > maxCount {
			maxCount = suitCounts[suit]
			bestSuit = suit
		}
	}
	return SkatGameSuit, bestSuit
}

// cpuPickPlay picks a card to play.
func (s *Skat) cpuPickPlay(playerIdx int) int {
	valid := s.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if playerIdx < 0 || playerIdx >= len(s.players) || s.players[playerIdx] == nil {
		return 0
	}
	p := s.players[playerIdx]

	// Determine the current winning card once. Reusing trickWinner() here keeps
	// the comparison rules (trump beats non-trump, lead-suit-only otherwise,
	// nullRank for null games) in a single source of truth.
	var winnerCard *Card
	leadSuit := -1
	if len(s.round.currentTrick) > 0 {
		leadSuit = s.effectiveSuit(s.round.currentTrick[0].Card)
		wIdx := s.trickWinner()
		for _, tc := range s.round.currentTrick {
			if tc.PlayerIdx == wIdx {
				winnerCard = tc.Card
				break
			}
		}
	}

	// Heuristic: if we can win the trick, prefer the strongest winning card;
	// otherwise play the lowest-point card.
	bestWinIdx := -1
	bestWinStrength := -1
	worstIdx := valid[0]
	worstScore := 1 << 30
	for _, i := range valid {
		c := p.GetCard(i)
		score := skatCardPoints(c)
		if score < worstScore {
			worstScore = score
			worstIdx = i
		}
		if winnerCard == nil {
			continue
		}
		strength := s.cardStrength(c)
		cardIsTrump := s.isTrump(c)
		winnerIsTrump := s.isTrump(winnerCard)
		canBeat := false
		switch {
		case cardIsTrump && !winnerIsTrump:
			canBeat = true
		case cardIsTrump && winnerIsTrump:
			canBeat = strength > s.cardStrength(winnerCard)
		case !cardIsTrump && winnerIsTrump:
			canBeat = false
		default:
			// Both non-trump: must follow the lead suit to beat the winner.
			if leadSuit >= 0 && c.GetDesign() == leadSuit && strength > s.cardStrength(winnerCard) {
				canBeat = true
			}
		}
		if canBeat && strength > bestWinStrength {
			bestWinStrength = strength
			bestWinIdx = i
		}
	}
	if bestWinIdx >= 0 {
		if s.config.CpuDifficulty == SkatCpuDifficultyHard || rand.Intn(2) == 0 {
			return bestWinIdx
		}
	}
	return worstIdx
}

// SkatBidEstimate は 1 つの契約について、手札から安全に受けられるビッド額を表す。
type SkatBidEstimate struct {
	// GameType は契約の種別 (スート戦 / グランド)。
	GameType SkatGameType
	// TrumpSuit はスート戦の切札 (グランドでは 0)。
	TrumpSuit int
	// Base は基礎点。
	Base int
	// Matadors はマタドール数 (with / without のうち成立するほう)。
	Matadors int
	// Value は (マタドール + 1) × 基礎点。**追加の宣言 (ハント / シュナイダー /
	// シュヴァルツ / ウーヴェルト) を一切使わずに正当化できる上限**で、これを
	// 超えて落札するとオーバービッドで失う危険がある。
	Value int
}

// SkatBidEstimates は各契約についての安全ビッド上限を返す。
//
// **ヌル戦は含めない。**基礎点 23 が契約の種別だけで決まり、マタドールの
// 連なりで伸びないので、マタドールに基づく見積もりには乗らない。
func SkatBidEstimates(hand []*Card) []SkatBidEstimate {
	type spec struct {
		gameType SkatGameType
		trump    int
	}
	specs := []spec{
		{SkatGameSuit, CardDesignClover},
		{SkatGameSuit, CardDesignSpade},
		{SkatGameSuit, CardDesignHeart},
		{SkatGameSuit, CardDesignDiamond},
		{SkatGameGrand, 0},
	}
	out := make([]SkatBidEstimate, 0, len(specs))
	for _, sp := range specs {
		m := skatMatadorsFor(hand, sp.gameType, sp.trump)
		base := skatBaseValueFor(sp.gameType, sp.trump)
		out = append(out, SkatBidEstimate{
			GameType:  sp.gameType,
			TrumpSuit: sp.trump,
			Base:      base,
			Matadors:  m,
			Value:     (m + 1) * base,
		})
	}
	return out
}

// SkatBestBidEstimate は最も高い見積もりを返す。
//
// **手札が空でも 0 にはならない。**切札を 1 枚も持たないのは「without が最大」と
// いうことなので、かえって大きな値が出る。配り終えた 10 枚の手札に対してだけ
// 意味がある — 途中まで出したあとの手札に使ってはいけない。
func SkatBestBidEstimate(hand []*Card) SkatBidEstimate {
	best := SkatBidEstimate{}
	for _, e := range SkatBidEstimates(hand) {
		if e.Value > best.Value {
			best = e
		}
	}
	return best
}
