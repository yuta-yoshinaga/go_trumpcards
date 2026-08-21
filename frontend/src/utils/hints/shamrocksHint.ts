import type { ShamrocksResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Shamrocks phase: playing. Later phases are game clear / game over. */
const PHASE_PLAYING = 0;

/**
 * Returns a Shamrocks frontend hint derived from the backend hint, or null.
 *
 * The suggestion rides along with every state response (see
 * ShamrocksWebPresenter.Output, #4483). A move either starts or extends a
 * foundation, which is always progress, or shuffles a card between fans, which
 * only sometimes is -- hence the two confidences.
 */
export function getShamrocksHint(state: ShamrocksResponse): HintResult | null {
  if (state.phase !== PHASE_PLAYING) return null;
  const hint = state.hint;
  if (!hint || hint.fromFan < 0) return null;

  if (hint.toFoundation) {
    return {
      targetAction: `fan-${hint.fromFan}`,
      reason: 'frontendHint.shamrocksToFoundation',
      confidence: 'strong',
    };
  }
  if (hint.toFan < 0) return null;
  return {
    targetAction: `fan-${hint.fromFan}`,
    reason: 'frontendHint.shamrocksToFan',
    confidence: 'moderate',
  };
}
