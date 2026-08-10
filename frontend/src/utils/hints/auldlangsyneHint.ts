import type { AuldLangSyneResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Auld Lang Syne phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/**
 * Returns an Auld Lang Syne frontend hint derived from the backend hint, or null.
 *
 * There is only one shape of hint here -- waste top to foundation. Sir Tommy's
 * second case (stock to foundation) has no counterpart, because the stock is
 * reached by dealing rather than by playing a card off it.
 */
export function getAuldLangSyneHint(state: AuldLangSyneResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  return {
    targetAction: `waste${state.hint.wasteIdx}-to-f${state.hint.foundationIdx}`,
    reason: 'frontendHint.auldlangsyneWaste',
    confidence: 'strong',
  };
}
