import type { OichoKabuResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { OichoKabuPhase } from '../../types/phases';

/**
 * Returns an Oicho-Kabu hint for the draw decision phase.
 *
 * Rank is the sum of card values mod 10 (9 is best, 0 is worst). A low rank has
 * a lot of upside and little to lose, so drawing a third card is favorable;
 * with a high rank the extra card usually only spoils it, so standing is the
 * basic play. Returns null outside the draw phase.
 */
export function getOichokabuHint(state: OichoKabuResponse): HintResult | null {
  if (state.phase !== OichoKabuPhase.DRAW) return null;
  if (state.playerRank <= 3) {
    return { targetAction: 'draw', reason: 'hint.drawLow', confidence: 'moderate' };
  }
  return { targetAction: 'stand', reason: 'hint.standHigh', confidence: 'moderate' };
}
