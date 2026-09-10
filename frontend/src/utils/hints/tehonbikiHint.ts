import type { TehonbikiResponse } from '../../types/games/tehonbiki';
import type { HintResult } from '../../types/hint';

/** Suggests wagering while a Tehonbiki round is open. */
export function getTehonbikiHint(state: TehonbikiResponse): HintResult | null {
  return state.gameEndFlag || state.phase !== 0
    ? null
    : { targetAction: 'bet', reason: 'frontendHint.tehonbiki', confidence: 'moderate' };
}
