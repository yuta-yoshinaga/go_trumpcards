import type { SkatResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SkatPhase } from '../../types/phases';

/** Returns a Skat frontend hint or null.
 *
 * The Skat backend supplies a strategic hint via `state.hint` (computed only
 * when the human is the active actor in bidding/skat/discard/declaration/play).
 * We surface that hint to the user via the standard frontend HintTooltip when
 * hints are enabled. */
export function getSkatHint(state: SkatResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  if (!state.hint) return null;

  switch (state.phase) {
    case SkatPhase.BID:
      return {
        targetAction: state.hint.bid && state.hint.bid > 0 ? 'bid' : 'pass',
        reason: state.hint.reason || 'hint.bid',
        confidence: 'moderate',
      };
    case SkatPhase.SKAT_PICKUP:
      return {
        targetAction: state.hint.pickSkat ? 'pickSkat' : 'handGame',
        reason: state.hint.reason || 'hint.pickSkat',
        confidence: 'moderate',
      };
    case SkatPhase.DISCARD:
      return {
        targetAction: 'discard',
        reason: state.hint.reason || 'hint.discard',
        confidence: 'moderate',
      };
    case SkatPhase.GAME_DECLARATION:
      return {
        targetAction: 'declare',
        reason: state.hint.reason || 'hint.declare',
        confidence: 'moderate',
      };
    case SkatPhase.PLAY:
      return {
        targetAction: 'play',
        reason: state.hint.reason || 'hint.play',
        confidence: 'moderate',
      };
    default:
      return null;
  }
}
