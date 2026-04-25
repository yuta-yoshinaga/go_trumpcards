import type { SpiteAndMaliceResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Spite & Malice phase: game over. */
const PHASE_GAME_OVER = 1;

/** Returns a Spite & Malice frontend hint derived from the backend hint, or null. */
export function getSpiteAndMaliceHint(state: SpiteAndMaliceResponse): HintResult | null {
  if (state.phase >= PHASE_GAME_OVER) return null;
  if (state.current !== 0) return null;
  if (!state.hint) return null;

  const { source, index, foundationIdx, discard } = state.hint;
  if (discard) {
    return {
      targetAction: `discard-${index}-${foundationIdx}`,
      reason: 'frontendHint.discard',
      confidence: 'moderate',
    };
  }
  switch (source) {
    case 'goal':
      return {
        targetAction: `goal-to-f${foundationIdx}`,
        reason: 'frontendHint.goalToFoundation',
        confidence: 'strong',
      };
    case 'hand':
      return {
        targetAction: `hand${index}-to-f${foundationIdx}`,
        reason: 'frontendHint.handToFoundation',
        confidence: 'strong',
      };
    case 'side':
      return {
        targetAction: `side${index}-to-f${foundationIdx}`,
        reason: 'frontendHint.sideToFoundation',
        confidence: 'strong',
      };
    default:
      return null;
  }
}
