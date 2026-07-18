import type { Card, MichiganBoodle } from '../types/card';

/** Per-boodle betting guidance derived from the human's hand and each boodle's claim state. */
export interface MichiganBoodleGuide {
  /**
   * Whether the human holds a card matching this boodle (same design and value),
   * so chips staked here are recoverable when that card is played.
   */
  collectible: boolean;
  /** Whether this boodle's chips have already been claimed by a player. */
  claimed: boolean;
}

/**
 * Computes betting guidance for each Michigan boodle so the player can bias their
 * chip distribution toward promising boodles rather than always splitting evenly.
 *
 * A boodle is "collectible" when the human's hand contains its exact card
 * (matching {@link Card.design} and {@link Card.value}); it is "claimed" once a
 * player has already taken its chips (`claimedBy >= 0`). Guidance is purely
 * advisory and never blocks betting.
 *
 * @param boodles - The four center boodle cards with their claim state.
 * @param handCards - The human's current hand (empty CPU hands yield no matches).
 * @returns One guide per boodle, index-aligned with `boodles`.
 */
export function michiganBoodleGuides(
  boodles: readonly MichiganBoodle[],
  handCards: readonly Card[],
): MichiganBoodleGuide[] {
  return boodles.map((boodle) => ({
    collectible: handCards.some((c) => c.design === boodle.card.design && c.value === boodle.card.value),
    claimed: boodle.claimedBy >= 0,
  }));
}
