import type { Card } from '../types/card';

/** Combo an unpaired hand forms once the retourne is counted as a shared card. */
export type RetourneNoteKey = 'favori' | 'carre';

/** Analysis of how the shared retourne card relates to the human's hand. */
export interface RetourneMatch {
  /** Hand indices whose rank equals the retourne's rank. */
  matchingIndices: number[];
  /**
   * Retourne-completed combo, or `null` when none:
   * - `'favori'`: the hand holds a pair the retourne turns into a brelan.
   * - `'carre'`: the hand is already a brelan and the retourne matches its rank.
   */
  noteKey: RetourneNoteKey | null;
}

/**
 * Determines which of the human's hand cards share the retourne's rank and
 * whether they combine into a retourne-completed brelan. Ranks are compared by
 * card value (Bouillotte uses a standard deck, so equal value means equal rank).
 * Returns no matches when the retourne has not yet been dealt.
 *
 * @param hand - The human player's cards.
 * @param retourne - The shared turned-up card, or `null` before it is dealt.
 * @returns The matching hand indices and any retourne-completed combo note key.
 */
export function analyzeRetourneMatch(hand: readonly Card[], retourne: Card | null | undefined): RetourneMatch {
  if (!retourne) return { matchingIndices: [], noteKey: null };
  const matchingIndices: number[] = [];
  for (let i = 0; i < hand.length; i++) {
    if (hand[i].value === retourne.value) matchingIndices.push(i);
  }
  let noteKey: RetourneNoteKey | null = null;
  if (matchingIndices.length === 2) noteKey = 'favori';
  else if (matchingIndices.length >= 3) noteKey = 'carre';
  return { matchingIndices, noteKey };
}
