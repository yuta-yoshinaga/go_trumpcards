import type { CostlyColoursResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Costly Colours, or null when there
 * is nothing to suggest.
 *
 * The suggestion comes from the Go domain with a bare `hintReason` identifier
 * (`fifteen`, `twenty_five`, `thirty_one`, `safe`, `go`, `mog_accept`,
 * `mog_refuse`); this adapter re-maps it into the shared HintResult shape for
 * {@link hooks/useGameHint.useGameHint | useGameHint}.
 *
 * **The mog phase names no card.** `hintHandIdx` is -1 there, so this cannot
 * gate on the index the way the trick games do.
 */
export function getCostlyColoursHint(state: CostlyColoursResponse): HintResult | null {
  if (!state.hintReason || state.hintReason === 'none') return null;
  return {
    targetAction: state.hintReason.startsWith('mog_') ? 'select' : 'play',
    reason: `hint.${state.hintReason}`,
    confidence: 'moderate',
  };
}
