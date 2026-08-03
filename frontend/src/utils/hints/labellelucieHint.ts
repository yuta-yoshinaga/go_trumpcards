import type { LaBelleLucieResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** La Belle Lucie phase: playing. Later phases are game clear / game over. */
const PHASE_PLAYING = 0;

/**
 * Returns a La Belle Lucie frontend hint derived from the backend hint, or null.
 *
 * The suggestion rides along with every state response (see
 * LaBelleLucieWebPresenter.Output, #4483). A move either starts or extends a
 * foundation, which is always progress, or shuffles a card between fans, which
 * only sometimes is -- hence the two confidences.
 */
export function getLaBelleLucieHint(state: LaBelleLucieResponse): HintResult | null {
  if (state.phase !== PHASE_PLAYING) return null;
  const hint = state.hint;
  if (!hint || hint.fromFan < 0) return null;

  if (hint.toFoundation) {
    return {
      targetAction: `fan-${hint.fromFan}`,
      reason: 'frontendHint.labellelucieToFoundation',
      confidence: 'strong',
    };
  }
  if (hint.toFan < 0) return null;
  return {
    targetAction: `fan-${hint.fromFan}`,
    reason: 'frontendHint.labellelucieToFan',
    confidence: 'moderate',
  };
}
