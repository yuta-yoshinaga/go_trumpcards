package domain

// trickCpuPlay plays the card the CPU seat picks. It does nothing for a human
// player or when no card can be removed (the selector returns 0 on an empty
// hand and RemoveCard(0) is nil; passing nil on would crash the HTTP handler,
// #4606). Callers must check that the hand is not over and the game is in its
// play phase before accessing players, so currentPlayerIdx is never used to
// index players while invalid.
func trickCpuPlay(seat int, player *GamePlayer, choose func(seat int) int, play func(seat int, c *Card)) {
	if player.GetIsHuman() {
		return
	}
	played := player.RemoveCard(choose(seat))
	if played == nil {
		return
	}
	play(seat, played)
}
