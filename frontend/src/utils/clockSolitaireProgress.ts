/**
 * Clock Solitaire progress: how many of the 13 piles are finished.
 *
 * The count lived inline in the page's CLI formatter, so "how many piles are
 * still open" was visible only to players who opened the terminal (#5523).
 * Extracted so the header and the terminal read the same number.
 */

/** Cards in a completed pile — mirrors domain.ClockSolitaireCardsPerPile. */
export const CLOCK_CARDS_PER_PILE = 4;

/** Piles on the clock, including the centre kings pile — domain.ClockSolitairePileCount. */
export const CLOCK_PILE_COUNT = 13;

/** Number of piles whose four cards are all face up. */
export function completedClockPiles(faceUpCount: readonly number[]): number {
  return faceUpCount.filter((c) => c >= CLOCK_CARDS_PER_PILE).length;
}
