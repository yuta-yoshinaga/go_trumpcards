import type { SomersetResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SomersetPhase } from '../../types/phases';

/** Returns a Somerset frontend hint or null.
 *
 * Somerset is a high-difficulty solitaire where every card sits face-up
 * from the start, so the strongest tactical advice is "always advance to the
 * foundation if legal" because freeing a column top opens future moves. The
 * backend supplies the next-move hint via `state.hint`; we surface it unchanged. */
export function getSomersetHint(state: SomersetResponse): HintResult | null {
  if (state.phase !== SomersetPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  const target = state.hint.toZone === 'foundation' ? 'play.foundation' : 'play.tableau';
  const reason = state.hint.toZone === 'foundation' ? 'hintReason.toFoundation' : 'hintReason.toTableau';
  return { targetAction: target, reason, confidence: 'strong' };
}
