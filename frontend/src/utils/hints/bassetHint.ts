import type { BassetResponse } from '../../types/games/basset';
import type { HintResult } from '../../types/hint';

/** Suggests the required Basset decision after a winning player card. */
export function getBassetHint(state: BassetResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  if (state.phase === 3) return { targetAction: 'paroli', reason: 'hint.paroli', confidence: 'strong' };
  if (state.phase === 1 || state.phase === 2)
    return { targetAction: 'bet', reason: 'hint.bet', confidence: 'moderate' };
  return null;
}
