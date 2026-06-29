import type { KingAlbertResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { KingAlbertPhase } from '../../types/phases';

/** Returns a King Albert frontend hint or null.
 *
 * King Albert is a high-difficulty solitaire where every card sits face-up
 * from the start (nine open columns plus a seven-card one-way reserve), so the
 * strongest tactical advice is "always advance to the foundation if legal"
 * because freeing a column top or emptying a reserve cell opens future moves.
 * The backend supplies the next-move hint via `state.hint`; we surface it
 * unchanged. */
export function getKingalbertHint(state: KingAlbertResponse): HintResult | null {
  if (state.phase !== KingAlbertPhase.PLAYING) return null;
  if (state.isStalemate) return null;
  if (!state.hint) return null;
  const target = state.hint.toZone === 'foundation' ? 'play.foundation' : 'play.tableau';
  const reason = state.hint.toZone === 'foundation' ? 'hintReason.toFoundation' : 'hintReason.toTableau';
  return { targetAction: target, reason, confidence: 'strong' };
}
