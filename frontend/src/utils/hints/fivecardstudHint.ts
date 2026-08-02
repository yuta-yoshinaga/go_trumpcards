import type { FiveCardStudResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { FiveCardStudPhase } from '../../types/phases';

/** ベットのあるストリート。ここ以外では賭けの助言をしない。 */
const BETTING_STREETS: readonly number[] = [
  FiveCardStudPhase.SECOND_STREET,
  FiveCardStudPhase.THIRD_STREET,
  FiveCardStudPhase.FOURTH_STREET,
  FiveCardStudPhase.FIFTH_STREET,
];

/**
 * ワンペア以上。`handRank` は `PokerHandNames` の添字で、0 = High Card、
 * 1 = One Pair (poker_hand_rank.go:25)。
 */
const MADE_HAND_FROM = 1;

/**
 * Returns a frontend {@link HintResult} for Five Card Stud, or null when no
 * suggestion is available.
 *
 * `handRank` is computed by the server for the human's own cards, so the hint
 * does not evaluate a hand here — it only turns that rank plus what is owed
 * into the four actions the page offers. It reads `maxBetAmount` before naming
 * a raise, so it never points at a control the betting limit has closed.
 *
 * It deliberately does not read the opponents' door cards. Judging a hand
 * against what is showing is the interesting part of Stud, and a hint that did
 * it badly would be worse than one that leaves it to the player.
 */
export function getFiveCardStudHint(state: FiveCardStudResponse): HintResult | null {
  if (state.gameEndFlag || !BETTING_STREETS.includes(state.phase)) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.folded || human.allIn || state.currentTurn !== human.id) return null;

  // **既に払い込んでいる分は差し引く。**同額まで出していれば負債はない。
  const owed = Math.max(0, state.lastBet - human.currentBet);

  if (human.handRank >= MADE_HAND_FROM) {
    // 上限に達していれば上げられない。押せない手を勧めない。
    if (state.maxBetAmount <= 0) {
      return owed > 0
        ? { targetAction: 'call', reason: 'frontendHint.fivecardstudCallMade', confidence: 'moderate' }
        : { targetAction: 'check', reason: 'frontendHint.fivecardstudCheckMade', confidence: 'moderate' };
    }
    return { targetAction: 'raise', reason: 'frontendHint.fivecardstudRaiseMade', confidence: 'moderate' };
  }

  return owed > 0
    ? { targetAction: 'fold', reason: 'frontendHint.fivecardstudFoldWeak', confidence: 'moderate' }
    : { targetAction: 'check', reason: 'frontendHint.fivecardstudCheckFree', confidence: 'moderate' };
}
