import type { CometResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Comet, or null when there is
 * nothing to suggest.
 *
 * The suggestion comes from the Go domain as a hand index plus a bare
 * `hintReason` identifier (`go_out`, `follow`, `comet`, `king`, `pass`); this
 * adapter re-maps it into the shared HintResult shape for
 * {@link hooks/useGameHint.useGameHint | useGameHint}.
 */
export function getCometHint(state: CometResponse): HintResult | null {
  if (state.hintHandIdx < 0 || !state.hintReason || state.hintReason === 'none') return null;
  return {
    targetAction: 'play',
    reason: `hint.${state.hintReason}`,
    confidence: 'moderate',
  };
}
