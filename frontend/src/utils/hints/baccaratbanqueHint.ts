import type { BaccaratBanqueResponse } from '../../types/card';
import { BACCARAT_BANQUE_PHASE } from '../../types/games/baccaratbanque';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Baccarat Banque, or null when there
 * is nothing to advise.
 *
 * **The recommendation comes from the server, not from the total.** The banker
 * is free at every total, so the right move depends on what the two tableaux
 * already drew — a calculation the domain does in `GetHint`. Re-deriving it
 * here would put the rule in a second place and let the two disagree.
 */
export function getBaccaratbanqueHint(state: BaccaratBanqueResponse): HintResult | null {
  if (state.gameEndFlag || !state.isHumanTurn) return null;
  if (state.phase !== BACCARAT_BANQUE_PHASE.BANKER) return null;
  if (!state.hintReason || state.hintReason === 'none') return null;

  return {
    targetAction: state.hintDraw ? 'draw' : 'stand',
    reason: `frontendHint.baccaratbanque_${state.hintReason}`,
    confidence: state.hintReason === 'behind_both' ? 'strong' : 'moderate',
  };
}
