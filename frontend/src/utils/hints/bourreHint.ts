import type { BourreResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 勝負に残る目安。切り札がこれだけあれば 1 トリックは望める。 */
const STAY_FROM_TRUMPS = 3;

/**
 * Returns a frontend {@link HintResult} for Bourré, or null when no suggestion
 * is available.
 *
 * The legal plays come from `validPlays` rather than being recomputed. Bourré
 * compels following suit and, when able, winning the trick, so the legal set
 * often collapses to a single card — the thing a player misses. The other
 * decision worth a hint is whether to stay in at all: failing to take a trick
 * after staying costs the pot.
 */
export function getBourreHint(state: BourreResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.folded || human.cards.length === 0) return null;
  if (state.currentPlayerIdx !== human.id) return null;

  if (state.phase === 'decide') {
    // 切り札が決まっていない間は数えようがない。
    if (!state.trumpSuit) return null;
    const trumps = human.cards.filter((c) => c.design === state.trumpSuit).length;
    return trumps >= STAY_FROM_TRUMPS
      ? { targetAction: 'stay', reason: 'frontendHint.bourreStay', confidence: 'moderate' }
      : { targetAction: 'fold', reason: 'frontendHint.bourreFold', confidence: 'moderate' };
  }

  if (state.phase !== 'play') return null;

  const plays = state.validPlays;
  if (plays.length === 0) return null;

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  return plays.length === 1
    ? { targetAction: `card-${plays[0]}`, reason: 'frontendHint.bourreForced', confidence: 'strong' }
    : { targetAction: `card-${plays[0]}`, reason: 'frontendHint.bourreChoose', confidence: 'moderate' };
}
