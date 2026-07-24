import type { Card, HeartsTrickCard } from '../types/card';

/**
 * Game context needed to evaluate whether a Hearts play is legal.
 *
 * Mirrors the fields the Go domain's `validatePlay` reads
 * (`internal/domain/Hearts.go`): the current trick, whether hearts have been
 * broken, the 1-based trick number, and the Omnibus (J♦) option that widens the
 * "point card" set used by the first-trick restriction.
 */
export interface HeartsPlayContext {
  /** Cards played so far in the current trick (empty when leading). */
  currentTrick: HeartsTrickCard[];
  /** Whether a heart (or Q♠) has been played, unlocking hearts leads. */
  heartsBroken: boolean;
  /** 1-based trick number; the first-trick rules apply only when this is 1. */
  trickNumber: number;
  /** Omnibus option: when true, J♦ is also treated as a point card. */
  omnibusJD: boolean;
}

/** Whether `card` is the two of clubs (the forced first-trick lead). */
function isTwoOfClubs(card: Card): boolean {
  return card.design === 'CLOVER' && card.value === 2;
}

/**
 * Whether `card` is a penalty/point card for first-trick discard purposes.
 * Hearts and Q♠ always count; J♦ counts only under the Omnibus option.
 * Mirrors `isPointCard` in the Go domain.
 */
function isPointCard(card: Card, omnibusJD: boolean): boolean {
  if (card.design === 'HEART') return true;
  if (card.design === 'SPADE' && card.value === 12) return true;
  return omnibusJD && card.design === 'DIAMOND' && card.value === 11;
}

/** Whether the hand contains the two of clubs. */
function handHasTwoOfClubs(hand: Card[]): boolean {
  return hand.some(isTwoOfClubs);
}

/** Whether the hand contains at least one non-heart card. */
function handHasNonHeart(hand: Card[]): boolean {
  return hand.some((c) => c.design !== 'HEART');
}

/** Whether the hand contains at least one card of the given suit design. */
function handHasSuit(hand: Card[], design: Card['design']): boolean {
  return hand.some((c) => c.design === design);
}

/** Whether the hand contains at least one non-point card. */
function handHasNonPointCard(hand: Card[], omnibusJD: boolean): boolean {
  return hand.some((c) => !isPointCard(c, omnibusJD));
}

/**
 * Whether `card` may be legally played from `hand` in the given context.
 *
 * Replicates the exact rule sequence of the Go domain's `validatePlay`
 * (`internal/domain/Hearts.go`):
 * 1. The very first card of the first trick must be the 2♣ (when the hand holds it).
 * 2. When leading, hearts may not be led before they are broken unless the hand
 *    holds nothing but hearts.
 * 3. When following, the lead suit must be followed if the hand can; when void,
 *    point cards (hearts / Q♠ / J♦-Omnibus) may not be discarded on the first
 *    trick while a non-point card remains.
 *
 * @param card - Candidate card from the player's hand.
 * @param hand - The full hand the card belongs to.
 * @param ctx - Current trick / hearts-broken / trick-number / Omnibus context.
 * @returns `true` when the card can be legally played.
 */
export function isHeartsLegalPlay(card: Card, hand: Card[], ctx: HeartsPlayContext): boolean {
  const { currentTrick, heartsBroken, trickNumber, omnibusJD } = ctx;
  const leading = currentTrick.length === 0;

  // First card of the first trick must be the 2♣ when the hand holds it.
  if (trickNumber === 1 && leading && !isTwoOfClubs(card) && handHasTwoOfClubs(hand)) {
    return false;
  }

  if (leading) {
    // Hearts cannot be led before they are broken, unless the hand is all hearts.
    if (!heartsBroken && card.design === 'HEART' && handHasNonHeart(hand)) {
      return false;
    }
    return true;
  }

  const leadSuit = currentTrick[0].card.design;
  if (card.design !== leadSuit) {
    // Must follow the lead suit if able.
    if (handHasSuit(hand, leadSuit)) return false;
    // Void in the lead suit: no point cards on the first trick while a safe card remains.
    if (trickNumber === 1 && isPointCard(card, omnibusJD) && handHasNonPointCard(hand, omnibusJD)) {
      return false;
    }
  }
  return true;
}

/**
 * Indices of the cards in `hand` that are legal to play in the given context.
 *
 * @param hand - The player's hand.
 * @param ctx - Current play context.
 * @returns The subset of hand indices that {@link isHeartsLegalPlay} accepts.
 */
export function heartsLegalPlayIndices(hand: Card[], ctx: HeartsPlayContext): number[] {
  const indices: number[] = [];
  hand.forEach((card, idx) => {
    if (isHeartsLegalPlay(card, hand, ctx)) indices.push(idx);
  });
  return indices;
}

/**
 * The i18n key describing why some cards are currently illegal, or `null` when
 * the player has a free choice. At any decision point every illegal card shares
 * a single governing reason, so one key suffices for the restriction tooltip.
 *
 * @param hand - The player's hand.
 * @param ctx - Current play context.
 * @returns An `illegalReason.*` translation key, or `null` when unrestricted.
 */
export function heartsIllegalReasonKey(hand: Card[], ctx: HeartsPlayContext): string | null {
  const { currentTrick, heartsBroken, trickNumber, omnibusJD } = ctx;
  const leading = currentTrick.length === 0;

  if (leading) {
    if (trickNumber === 1 && handHasTwoOfClubs(hand)) {
      return 'illegalReason.mustLeadTwoClubs';
    }
    if (!heartsBroken && handHasNonHeart(hand)) {
      return 'illegalReason.heartsNotBroken';
    }
    return null;
  }

  const leadSuit = currentTrick[0].card.design;
  if (handHasSuit(hand, leadSuit)) {
    return 'illegalReason.mustFollowSuit';
  }
  if (trickNumber === 1 && handHasNonPointCard(hand, omnibusJD)) {
    return 'illegalReason.noPointsFirstTrick';
  }
  return null;
}
