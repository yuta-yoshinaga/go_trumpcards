import type { GapsResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { GapsPhase } from '../../types/phases';

/**
 * Returns a frontend HintResult for Gaps. The backend already exposes the
 * authoritative next-move suggestion in `state.hint`, so this helper just
 * surfaces an i18n-friendly toast whenever a hint is available.
 */
export function getGapsHint(state: GapsResponse): HintResult | null {
  if (state.phase !== GapsPhase.PLAYING) return null;
  if (!state.hint) return null;
  return {
    targetAction: 'move',
    reason: 'frontendHint.validMove',
    confidence: 'strong',
  };
}
