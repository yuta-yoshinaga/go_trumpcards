import type { AccordionResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Accordion phase: game clear. */
const PHASE_GAME_CLEAR = 1;

/** Returns an Accordion frontend hint derived from the backend hint, or null. */
export function getAccordionHint(state: AccordionResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_CLEAR) return null;
  if (!state.hint) return null;

  const offset = state.hint.fromIdx - state.hint.toIdx;
  return {
    targetAction: offset === 3 ? 'mergeOffset3' : 'mergeOffset1',
    reason: offset === 3 ? 'frontendHint.accordionOffset3' : 'frontendHint.accordionOffset1',
    confidence: 'moderate',
  };
}
