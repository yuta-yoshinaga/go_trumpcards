import type { BidEuchreResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 人間の手番であることを示すフェーズ (2 = Play)。 */
const PLAY_PHASE = 2;

/** 宣言 5 = NoTrump LOW。ランキングが逆転し、9 が最強になる。 */
const NO_TRUMP_LOW = 5;

/**
 * Returns a frontend {@link HintResult} for Bid Euchre, or null when no
 * suggestion is available.
 *
 * `validPlays` comes from the server because the left bower counts as a trump,
 * which the trick alone does not show.
 *
 * The declaration worth calling out is **no-trump low**: the ranking reverses
 * and the nine becomes the highest card, so every instinct built on the other
 * five declarations is backwards. `trumpSuit` is 0 for both no-trump forms and
 * cannot tell them apart, so this reads `trump`.
 */
export function getBidEuchreHint(state: BidEuchreResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== PLAY_PHASE) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  const plays = state.validPlays;
  if (plays.length === 0) return null;

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  if (plays.length === 1) {
    return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.bideuchreForced', confidence: 'strong' };
  }

  // **no-trump low は序列が逆。**他の宣言で身につけた勘がそのまま裏目になる。
  if (state.trump === NO_TRUMP_LOW) {
    return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.bideuchreLowRanking', confidence: 'moderate' };
  }

  return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.bideuchreChoose', confidence: 'moderate' };
}
