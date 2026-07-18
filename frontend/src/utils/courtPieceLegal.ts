import type { Card, CourtPieceTrickCard } from '../types/card';

/**
 * Whether the hand contains at least one card of the given suit design.
 * Mirrors `playerHasSuit` in the Go domain (`internal/domain/CourtPiece.go`).
 */
function handHasSuit(hand: Card[], design: Card['design']): boolean {
  return hand.some((c) => c.design === design);
}

/**
 * Whether `card` may be legally played from `hand` given the current trick.
 *
 * Replicates the exact rule of the Go domain's `validatePlay`
 * (`internal/domain/CourtPiece.go`): Court Piece (Rang) enforces follow-suit
 * only. When leading (the trick is empty) any card is legal; when following,
 * the lead suit must be followed if the hand holds it, otherwise any card
 * (trump or discard) is legal.
 *
 * @param card - Candidate card from the player's hand.
 * @param hand - The full hand the card belongs to.
 * @param currentTrick - Cards played so far in the current trick.
 * @returns `true` when the card can be legally played.
 */
export function isCourtPieceLegalPlay(card: Card, hand: Card[], currentTrick: CourtPieceTrickCard[]): boolean {
  if (currentTrick.length === 0) return true;
  const leadSuit = currentTrick[0].card.design;
  if (card.design === leadSuit) return true;
  return !handHasSuit(hand, leadSuit);
}

/**
 * Indices of the cards in `hand` that are legal to play given the current trick.
 *
 * @param hand - The player's hand.
 * @param currentTrick - Cards played so far in the current trick.
 * @returns The subset of hand indices that {@link isCourtPieceLegalPlay} accepts.
 */
export function courtPieceLegalPlayIndices(hand: Card[], currentTrick: CourtPieceTrickCard[]): number[] {
  const indices: number[] = [];
  hand.forEach((card, idx) => {
    if (isCourtPieceLegalPlay(card, hand, currentTrick)) indices.push(idx);
  });
  return indices;
}
