import type { CurdsAndWheyResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Simple Simon phase: playing. Later phases are game clear / game over. */
const PHASE_PLAYING = 0;

/**
 * Returns a Simple Simon frontend hint derived from the backend hint, or null.
 *
 * The suggestion rides along with every state response (see
 * CurdsAndWheyWebPresenter.Output, #4483). Every move is column-to-column, so
 * the hint only has to name the two columns.
 */
export function getCurdsAndWheyHint(state: CurdsAndWheyResponse): HintResult | null {
  if (state.phase !== PHASE_PLAYING) return null;
  const hint = state.hint;
  if (!hint || hint.fromCol < 0 || hint.toCol < 0) return null;

  return {
    targetAction: `col-${hint.fromCol}`,
    reason: 'frontendHint.curdsandwheyMove',
    confidence: 'moderate',
  };
}
