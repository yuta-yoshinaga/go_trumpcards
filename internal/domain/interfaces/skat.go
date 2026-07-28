//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SkatGame Skat game interface used by the use-case layer.
type SkatGame interface {
	BaseGame

	// Reset starts a new game session.
	Reset()
	// NextRound advances to the next round.
	NextRound()

	// PlayerBid applies the human's accept/pass at the active bid step.
	PlayerBid(accept bool) error
	// CpuBid lets the active CPU bidder/responder act.
	CpuBid()

	// PlayerPickSkat declarer picks up the skat (true) or plays a hand game (false).
	PlayerPickSkat(pickup bool) error
	// CpuPickSkat CPU declarer's skat-pickup choice.
	CpuPickSkat()

	// PlayerDiscard declarer discards two cards.
	PlayerDiscard(idxA, idxB int) error
	// CpuDiscard CPU declarer's discard.
	CpuDiscard()

	// PlayerDeclareGame declarer chooses game type and trump suit.
	PlayerDeclareGame(gameType domain.SkatGameType, trumpSuit int) error
	// CpuDeclareGame CPU declarer chooses game.
	CpuDeclareGame()

	// PlayerPlay human plays a card.
	PlayerPlay(cardIndex int) error
	// CpuPlay CPU plays a card.
	CpuPlay()

	// ResolveTrick resolves the current trick.
	ResolveTrick()
	// NextTrick begins the next trick.
	NextTrick()
	// ScoreRound finalises the round score.
	ScoreRound()

	// GetConfig returns the current configuration.
	GetConfig() domain.SkatConfig
	// SetConfig sets the configuration.
	SetConfig(cfg domain.SkatConfig)

	// GetGameEndFlag reports whether the game session has ended.
	GetGameEndFlag() bool
	// GetPhase returns the current phase.
	GetPhase() domain.SkatPhase
	// IsHumanTurn reports whether the current trick player is human.
	IsHumanTurn() bool
	// IsHumanBidTurn reports whether the active bid actor is human.
	IsHumanBidTurn() bool
	// IsHumanDeclarerTurn reports whether the declarer is human (skat pickup,
	// discard, and game-declaration phases).
	IsHumanDeclarerTurn() bool

	// GetRoundNumber returns the round number.
	GetRoundNumber() int
	// GetTrickNumber returns the trick number.
	GetTrickNumber() int
	// GetCurrentPlayerIdx returns the player whose turn it is to play a card.
	GetCurrentPlayerIdx() int
	// GetCurrentTrick returns the current trick (in play order).
	GetCurrentTrick() []*domain.TrickCard
	// GetForehandIdx returns the forehand index.
	GetForehandIdx() int
	// GetMiddlehandIdx returns the middlehand index.
	GetMiddlehandIdx() int
	// GetRearhandIdx returns the rearhand index.
	GetRearhandIdx() int
	// GetDealerIdx returns the dealer index.
	GetDealerIdx() int
	// GetDeclarerIdx returns the declarer index (-1 if not yet decided).
	GetDeclarerIdx() int
	// GetCurrentBid returns the most recent accepted bid.
	GetCurrentBid() int
	// GetActiveBidActorIdx returns the player whose response is currently
	// expected during bidding (-1 if bidding is over).
	GetActiveBidActorIdx() int
	// GetGameType returns the chosen game type.
	GetGameType() domain.SkatGameType
	// GetTrumpSuit returns the chosen trump suit (suit games only).
	GetTrumpSuit() int
	// GetSkat returns the current skat (face-down cards).
	GetSkat() []*domain.Card
	// GetOriginalSkat returns the pre-pickup skat snapshot.
	GetOriginalSkat() []*domain.Card
	// GetDeclarerCardPoints returns the declarer's card points (post-round).
	GetDeclarerCardPoints() int
	// GetDefendersCardPoints returns the defenders' card points (post-round).
	GetDefendersCardPoints() int
	// GetWinnerSide returns the round outcome.
	GetWinnerSide() int
	// GetGameValue returns the round's game value.
	GetGameValue() int
	// GetLeadPlayerIdx returns the lead player's index.
	GetLeadPlayerIdx() int
	// PickedSkat reports whether the declarer picked up the skat.
	PickedSkat() bool

	// GetPlayerCnt returns the player count.
	GetPlayerCnt() int
	// GetPlayer returns the i-th player.
	GetPlayer(i int) *domain.SkatPlayer

	// GetValidPlayIndices returns indices of legally playable cards.
	GetValidPlayIndices(playerIdx int) []int
	// GetHint returns a hint for the human player.
	GetHint() *domain.SkatHint
}
