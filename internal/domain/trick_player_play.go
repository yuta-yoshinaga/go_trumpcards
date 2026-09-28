package domain

// trickPlayerPlay plays the human seat's card at cardIndex. Callers must check
// that the hand is not over and the game is in its play phase first (so an
// invalid currentPlayerIdx is never used to index players). errKey is the
// game's i18n code for an out-of-range index.
func trickPlayerPlay(seat int, player *GamePlayer, cardIndex int, errKey string,
	validate func(seat int, c *Card) error, play func(seat int, c *Card)) error {
	if !player.GetIsHuman() {
		return ErrNotHumanTurn
	}
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainErrorCode(ErrInvalidCard, errKey, nil)
	}
	if err := validate(seat, player.GetCard(cardIndex)); err != nil {
		return err
	}
	play(seat, player.RemoveCard(cardIndex))
	return nil
}
