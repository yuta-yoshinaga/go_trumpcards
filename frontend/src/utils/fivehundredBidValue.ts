/** Avondale base score for a 6-trick suit contract, keyed by suit id (1=♠,2=♣,3=♥,4=♦). */
const SUIT_BASE: Record<number, number> = { 1: 40, 2: 60, 3: 100, 4: 80 };

/** Fixed score for a Misère contract (declarer takes 0 tricks). */
export const FIVEHUNDRED_MISERE_VALUE = 250;
/** Fixed score for an Open Misère contract (hand revealed). */
export const FIVEHUNDRED_OPEN_MISERE_VALUE = 520;

/**
 * Avondale schedule score for a 500 suit/no-trump bid, mirroring the Go domain
 * (`FiveHundredBid.Value`): suit base + 100 per trick above 6; pass `suit = -1`
 * for No Trump (base 120). Misère / Open Misère are fixed
 * ({@link FIVEHUNDRED_MISERE_VALUE} / {@link FIVEHUNDRED_OPEN_MISERE_VALUE}).
 */
export function fivehundredBidValue(tricks: number, suit: number): number {
  const base = suit === -1 ? 120 : (SUIT_BASE[suit] ?? 0);
  return base + (tricks - 6) * 100;
}
