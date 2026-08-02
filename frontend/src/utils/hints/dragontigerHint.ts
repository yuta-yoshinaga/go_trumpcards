import type { DragonTigerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { DragonTigerBetType, DragonTigerPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Dragon Tiger, or null when no
 * suggestion is available.
 *
 * Dragon and Tiger are one card each with nothing to read, so there is no
 * advice to give about choosing between them and this offers none. The one
 * thing that is decidable is the tie: it pays 8:1 against roughly a one-in-
 * thirteen chance, so it loses money over time — and a Dragon or Tiger stake
 * only loses *half* to a tie rather than all of it.
 *
 * This was a stub returning null. Unlike the seven in #4637 there is no server
 * hint to surface here, so the advice is computed from the rules themselves.
 */
export function getDragontigerHint(state: DragonTigerResponse): HintResult | null {
  if (state.phase !== DragonTigerPhase.BET) return null;

  return state.betType === DragonTigerBetType.TIE
    ? { targetAction: 'bet', reason: 'frontendHint.dragontigerTieOdds', confidence: 'moderate' }
    : { targetAction: 'bet', reason: 'frontendHint.dragontigerEvenMoney', confidence: 'moderate' };
}
