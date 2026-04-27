import type { NertzResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Nertz hint stub. Hints are computed server-side and surfaced via
 * `state.hint`; the frontend does not run a separate calculator. Returns
 * a HintResult only when the backend provided one.
 */
export function getNertzHint(state: NertzResponse): HintResult | null {
  if (!state.hint) return null;
  const { fromZone, fromCol, cardIndex, toZone, toCol } = state.hint;
  const target = `${fromZone}${fromCol >= 0 ? `-c${fromCol}` : ''}${cardIndex >= 0 ? `-i${cardIndex}` : ''}-to-${toZone}-${toCol}`;
  return {
    targetAction: target,
    reason: 'nertz.messages.selectTarget',
    confidence: 'moderate',
  };
}
