import type { Card, ChinesePokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ChinesePokerPhase } from '../../types/phases';
import { chinesePokerIsFoul } from '../chinesePokerFoul';

/** 前列 3 枚・中列 5 枚・後列 5 枚 (sync: internal/domain/ChinesePoker.go)。 */
const FRONT_SIZE = 3;
const MIDDLE_SIZE = 5;
const FULL_HAND = 13;

/**
 * Returns a frontend {@link HintResult} for Chinese Poker, or null when no
 * suggestion is available.
 *
 * The page renders only `reason` — it has no `data-hint-action` targets — so the
 * advice has to be readable on its own rather than pointing at a control.
 *
 * **The arrangement being staged is not in the response.** `assignments` lives in
 * `ChinesePokerPage`'s own state and never reaches the server until `set`, so this
 * cannot comment on what the player has selected so far; it can only talk about
 * the thirteen cards they were dealt. That is why the set-hands advice is phrased
 * as a suggested split rather than a critique of the current one.
 *
 * The split it suggests is the rank-ordered one — five highest at the back, next
 * five in the middle, three lowest in front — checked against the shared
 * {@link chinesePokerIsFoul}, which is a 1:1 port of the server's `cpValidateHands`.
 * That check is not decoration: sorting by rank does **not** guarantee a legal
 * arrangement, because categories do not follow high cards. Three low cards of a
 * kind in front beat a middle row of unpaired high cards, and that arrangement
 * fouls. When the naive split fouls, the hint says so instead of recommending a
 * losing arrangement — it does not search for a better one, because an exhaustive
 * split is 13C3 x 10C5 = 72,072 evaluations on every response.
 */
export function getChinesePokerHint(state: ChinesePokerResponse): HintResult | null {
  if (state.phase === ChinesePokerPhase.BET) {
    // **チップが賭け金に足りているかだけは確かなことが言える。**
    return state.chips <= 0
      ? null
      : { targetAction: 'bet', reason: 'frontendHint.chinesepokerBet', confidence: 'moderate' };
  }

  if (state.phase !== ChinesePokerPhase.SET_HANDS) return null;
  if (state.playerCards.length !== FULL_HAND) return null;

  // **サーバーの案があるならそれを使う。**ドメインは同じ計算を CUI 向けに
  // 出しており (#4717)、ここで並べ直すと同じ手札で違う分け方を勧めることに
  // なる。targetIndices を付けるので、ページはどの札かを名指しできる (#5615)。
  const suggested = state.suggestedArrangement;
  if (suggested) {
    return {
      targetAction: 'setHands',
      reason: suggested.foul ? 'frontendHint.chinesepokerFoulRisk' : 'frontendHint.chinesepokerSplit',
      confidence: 'moderate',
      targetIndices: suggested.front,
    };
  }

  const sorted = [...state.playerCards].sort((a, b) => rank(b) - rank(a));
  const back = sorted.slice(0, MIDDLE_SIZE);
  const middle = sorted.slice(MIDDLE_SIZE, MIDDLE_SIZE * 2);
  const front = sorted.slice(MIDDLE_SIZE * 2);

  if (front.length !== FRONT_SIZE) return null;

  return chinesePokerIsFoul(front, middle, back)
    ? { targetAction: 'setHands', reason: 'frontendHint.chinesepokerFoulRisk', confidence: 'moderate' }
    : { targetAction: 'setHands', reason: 'frontendHint.chinesepokerSplit', confidence: 'moderate' };
}

/** A は最強。`value` は 1 が A なので 14 に読み替える。 */
function rank(c: Card): number {
  return c.value === 1 ? 14 : c.value;
}
