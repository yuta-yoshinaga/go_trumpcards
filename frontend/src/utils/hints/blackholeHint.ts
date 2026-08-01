import type { BlackHoleResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Black Hole phase: playing. Later phases are game clear / game over. */
const PHASE_PLAYING = 0;

/**
 * Returns a Black Hole frontend hint derived from the backend hint, or null.
 *
 * The suggestion rides along with every state response (see
 * BlackHoleWebPresenter.Output, #4483). Black Hole has exactly one kind of
 * move -- play a fan's top card into the hole -- so the hint names the fan and
 * nothing else.
 */
export function getBlackHoleHint(state: BlackHoleResponse): HintResult | null {
  if (state.phase !== PHASE_PLAYING) return null;
  const hint = state.hint;
  if (!hint || hint.fan < 0) return null;

  return {
    targetAction: `fan-${hint.fan}`,
    reason: 'frontendHint.blackholePlayFan',
    confidence: 'strong',
  };
}
