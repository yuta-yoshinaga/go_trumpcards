import type { AndarBaharResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { AndarBaharPhase, AndarBaharSideBand } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Andar Bahar, or null when there is
 * nothing to advise.
 *
 * There is no card counting to do here — the odds are fixed the moment the
 * joker is turned up. What *is* decidable is which side of the table costs
 * less: the first-dealt column wins 51.50% of the time and pays 0.9:1 for a
 * house edge of 2.15%, against 3.00% on the other column. So the advice is to
 * back the column that is dealt first, and to treat the side bet — roughly an
 * 11% margin — as the expensive bet it is.
 */
export function getAndarbaharHint(state: AndarBaharResponse): HintResult | null {
  if (state.phase !== AndarBaharPhase.BET) return null;

  return state.sideBand !== AndarBaharSideBand.NONE
    ? { targetAction: 'bet', reason: 'frontendHint.andarBaharSideBet', confidence: 'moderate' }
    : { targetAction: 'bet', reason: 'frontendHint.andarBaharFirstColumn', confidence: 'moderate' };
}
