//go:build !js || !wasm || casino

package controller

// PokerCommonInput holds the tournament/limit/rebuy/addon/tableSize fields
// shared by every tournament-style poker WebInput (Holdem, Pineapple,
// SevenCardStud, ...). It is meant to be embedded in those input structs
// so they can be parsed from the same JSON shape and applied via the same
// helpers without copy-pasting field declarations.
type PokerCommonInput struct {
	TournamentMode   *bool `json:"tournamentMode,omitempty"`
	BettingLimit     *int  `json:"bettingLimit,omitempty"`
	TableSize        *int  `json:"tableSize,omitempty"`
	RebuyEnabled     *bool `json:"rebuyEnabled,omitempty"`
	RebuyMaxCount    *int  `json:"rebuyMaxCount,omitempty"`
	RebuyChips       *int  `json:"rebuyChips,omitempty"`
	RebuyPeriodHands *int  `json:"rebuyPeriodHands,omitempty"`
	AddonEnabled     *bool `json:"addonEnabled,omitempty"`
	AddonChips       *int  `json:"addonChips,omitempty"`
	AddonAfterHand   *int  `json:"addonAfterHand,omitempty"`
}

// PokerBlindsInput holds the small/big blind fields shared by the
// Holdem-family poker WebInputs (Holdem, Pineapple, ...). SevenCardStud and
// other Stud variants do not embed this because they use Ante/BringIn/Bet
// fields instead.
type PokerBlindsInput struct {
	SmallBlind      *int `json:"smallBlind,omitempty"`
	BigBlind        *int `json:"bigBlind,omitempty"`
	BlindLevelHands *int `json:"blindLevelHands,omitempty"`
	BlindMultiplier *int `json:"blindMultiplier,omitempty"`
}
