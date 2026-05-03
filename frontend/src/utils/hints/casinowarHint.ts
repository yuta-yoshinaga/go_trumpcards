import type { CasinoWarResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CasinoWarPhase } from '../../types/phases';

/**
 * Returns a Casino War hint for the tie decision phase.
 *
 * Going to war has a smaller house edge (~2.88%) than surrendering (~3.7%)
 * because the war bet pays even money on a tie or win, while surrender always
 * loses half the ante. Recommend war whenever the player is in TieDecision.
 */
export function getCasinowarHint(state: CasinoWarResponse): HintResult | null {
  if (state.phase !== CasinoWarPhase.TIE_DECISION) return null;
  return { targetAction: 'war', reason: 'hint.warEv', confidence: 'strong' };
}
