import {
  BJ_SUGGEST_DECLINE_INSURANCE,
  BJ_SUGGEST_DOUBLE,
  BJ_SUGGEST_DOUBLE_STAND,
  BJ_SUGGEST_HIT,
  BJ_SUGGEST_NONE,
  BJ_SUGGEST_SPLIT,
  BJ_SUGGEST_STAND,
  BJ_SUGGEST_SURRENDER,
} from '../../components/blackjack/bjConstants';
import type { BlackJackResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Maps backend suggestedAction to frontend hint action+reason. */
const HINT_MAP: Record<number, { targetAction: string; reason: string }> = {
  [BJ_SUGGEST_HIT]: { targetAction: 'hit', reason: 'hintReason.hitReason' },
  [BJ_SUGGEST_STAND]: { targetAction: 'stand', reason: 'hintReason.standReason' },
  [BJ_SUGGEST_DOUBLE]: { targetAction: 'double', reason: 'hintReason.doubleReason' },
  [BJ_SUGGEST_DOUBLE_STAND]: { targetAction: 'double', reason: 'hintReason.doubleStandReason' },
  [BJ_SUGGEST_SPLIT]: { targetAction: 'split', reason: 'hintReason.splitReason' },
  [BJ_SUGGEST_SURRENDER]: { targetAction: 'surrender', reason: 'hintReason.surrenderReason' },
  [BJ_SUGGEST_DECLINE_INSURANCE]: { targetAction: 'decline', reason: 'hintReason.declineReason' },
};

/** Returns a frontend HintResult from BlackJack game state, or null if no suggestion. */
export function getBlackjackHint(state: BlackJackResponse): HintResult | null {
  if (!state.hintEnabled) return null;
  if (state.suggestedAction === BJ_SUGGEST_NONE) return null;
  const mapped = HINT_MAP[state.suggestedAction];
  if (!mapped) return null;
  return { ...mapped, confidence: 'strong' };
}
