import type { MonteBankResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { MonteBankPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Monte Bank, or null when there is
 * nothing to advise.
 *
 * **The only decision in this game is which layout card to back**, and the
 * whole house edge comes out of it: a suit showing once is exactly break-even
 * at 3:1, while two, three or four cost 11.1%, 22.2% and 33.3%. So the advice
 * is simply "back the least duplicated suit" — and when nothing is even, say
 * so rather than implying the choice is safe.
 */
export function getMontebankHint(state: MonteBankResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  if (state.phase !== MonteBankPhase.BET || state.layout.length === 0) return null;

  const best = state.layout.reduce((a, b) => (b.suitCount < a.suitCount ? b : a));
  if (best.isEven) {
    return { targetAction: 'bet', reason: 'frontendHint.monteBankLoneSuit', confidence: 'strong' };
  }
  return { targetAction: 'bet', reason: 'frontendHint.monteBankAllDuplicated', confidence: 'moderate' };
}
