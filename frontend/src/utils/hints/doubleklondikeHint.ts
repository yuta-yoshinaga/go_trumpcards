import type { DoubleKlondikeResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Double Klondike phase: playing. Later phases are game clear / game over. */
const PHASE_PLAYING = 0;

/**
 * Returns a Double Klondike frontend hint derived from the backend hint, or null.
 *
 * The suggestion rides along with every state response (see
 * DoubleKlondikeWebPresenter.Output, #4483). A card reaching a foundation is
 * always progress; everything else is a rearrangement that may or may not be.
 */
export function getDoubleKlondikeHint(state: DoubleKlondikeResponse): HintResult | null {
  if (state.phase !== PHASE_PLAYING) return null;
  const hint = state.hint;
  if (!hint || !hint.fromZone || !hint.toZone) return null;

  // The waste has no column, so the action names the zone alone.
  const from = hint.fromCol >= 0 ? `${hint.fromZone}-${hint.fromCol}` : hint.fromZone;
  const toFoundation = hint.toZone === 'foundation';
  return {
    targetAction: from,
    reason: toFoundation ? 'frontendHint.dklondikeToFoundation' : 'frontendHint.dklondikeToTableau',
    confidence: toFoundation ? 'strong' : 'moderate',
  };
}
