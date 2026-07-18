import type { Card, CardDesign } from '../types/card';

/**
 * Maps a Bezique trump-suit ordinal (1=♠ 2=♣ 3=♥ 4=♦ — the domain's
 * `CardDesign` constants in `internal/domain/Card.go`) to the card `design`
 * string sent on the wire. Returns `null` for an out-of-range ordinal (e.g. 0
 * before a trump has been fixed), which callers treat as "no trump".
 *
 * @param suit - The trump-suit ordinal from `BeziqueResponse.trumpSuit`.
 * @returns The matching `CardDesign`, or `null` when the ordinal is not 1-4.
 */
export function beziqueSuitDesign(suit: number): CardDesign | null {
  switch (suit) {
    case 1:
      return 'SPADE';
    case 2:
      return 'CLOVER';
    case 3:
      return 'HEART';
    case 4:
      return 'DIAMOND';
    default:
      return null;
  }
}

/**
 * In-suit rank strength (higher beats lower): A > 10 > K > Q > J > 9 > 8 > 7.
 * Mirrors `BeziqueRankOrder` in `internal/domain/Bezique.go`.
 *
 * @param value - The card's numeric value (1=A, 11=J, 12=Q, 13=K).
 * @returns A comparable strength score within a suit.
 */
function beziqueRank(value: number): number {
  switch (value) {
    case 1:
      return 8; // A
    case 10:
      return 7;
    case 13:
      return 6; // K
    case 12:
      return 5; // Q
    case 11:
      return 4; // J
    default:
      return value - 6; // 9→3, 8→2, 7→1
  }
}

/**
 * Whether `challenger` beats `currentBest` given the led suit and trump suit.
 * Mirrors `beziqueBeats` in `internal/domain/Bezique.go`: any trump beats any
 * non-trump, a non-lead non-trump card cannot win, and an exact tie is kept by
 * the earlier (led) card.
 *
 * @param challenger - The candidate card.
 * @param currentBest - The card currently winning the trick (the lead card).
 * @param leadSuit - The suit that was led.
 * @param trump - The trump suit design, or `null` when there is no trump.
 * @returns `true` when `challenger` would win over `currentBest`.
 */
function beziqueBeats(challenger: Card, currentBest: Card, leadSuit: CardDesign, trump: CardDesign | null): boolean {
  const cTrump = challenger.design === trump;
  const bTrump = currentBest.design === trump;
  if (cTrump && bTrump) return beziqueRank(challenger.value) > beziqueRank(currentBest.value);
  if (cTrump) return true;
  if (bTrump) return false;
  if (challenger.design !== leadSuit) return false;
  if (currentBest.design !== leadSuit) return true;
  return beziqueRank(challenger.value) > beziqueRank(currentBest.value);
}

/**
 * Whether `card` may be legally played by the follower during the Bezique
 * endgame (phase 2), when strict follow-suit + win-if-able rules apply. Mirrors
 * `Bezique.cardSatisfiesFollow` in `internal/domain/Bezique.go`:
 *
 * - Holding the led suit: the card must follow it, and if the hand can beat the
 *   lead in that suit the card must be one that beats it.
 * - Void in the led suit but holding trump: the card must be a trump.
 * - Void in both: any card is legal.
 *
 * This governs only the follower (a non-empty trick); the leader may play any
 * card, so callers should skip this check when leading.
 *
 * @param card - The candidate card from the follower's hand.
 * @param hand - The follower's full hand.
 * @param leadCard - The opponent's led card (the first card in the trick).
 * @param trump - The trump suit design, or `null` when there is no trump.
 * @returns `true` when the card satisfies the endgame follow rule.
 */
export function isBeziqueEndgameLegalPlay(card: Card, hand: Card[], leadCard: Card, trump: CardDesign | null): boolean {
  const leadSuit = leadCard.design;
  if (hand.some((c) => c.design === leadSuit)) {
    if (card.design !== leadSuit) return false;
    const hasWinner = hand.some((c) => c.design === leadSuit && beziqueBeats(c, leadCard, leadSuit, trump));
    return hasWinner ? beziqueBeats(card, leadCard, leadSuit, trump) : true;
  }
  if (trump != null && hand.some((c) => c.design === trump)) {
    return card.design === trump;
  }
  return true;
}

/**
 * Indices of the follower's hand cards that are legal to play during the
 * Bezique endgame, for highlighting. See {@link isBeziqueEndgameLegalPlay}.
 *
 * @param hand - The follower's full hand.
 * @param leadCard - The opponent's led card (the first card in the trick).
 * @param trump - The trump suit design, or `null` when there is no trump.
 * @returns The legal-to-play indices into `hand`.
 */
export function beziqueEndgameLegalIndices(hand: Card[], leadCard: Card, trump: CardDesign | null): number[] {
  return hand.reduce<number[]>((acc, card, idx) => {
    if (isBeziqueEndgameLegalPlay(card, hand, leadCard, trump)) acc.push(idx);
    return acc;
  }, []);
}
