import type { Card, ScoponeResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 人間の手番であることを示すフェーズ。 */
const PLAYER_TURN = 'playerTurn';

/**
 * Returns a frontend {@link HintResult} for Scopone, or null when no suggestion
 * is available.
 *
 * The legal captures are **not** recomputed here: `handCaptures` already lists,
 * per hand card, every table subset that is a legal take. Re-deriving that would
 * duplicate the domain's search. What this adds is the choice between them —
 * clearing the table is an escoba and worth a point on its own, so it outranks
 * any larger-but-incomplete capture.
 *
 * This is the same shape as `escobaHint`, because both games hand the frontend
 * the same `handCaptures` structure. They are deliberately not shared: Escoba's
 * captures must sum to fifteen while Scopone's match a rank or sum to it, so one
 * helper would have to know which rule it was under to describe itself
 * honestly.
 */
export function getScoponeHint(state: ScoponeResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== PLAYER_TURN || !state.isHumanTurn) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const tableSize = state.tableCards.length;

  // handCaptures は手札と同じ長さで届くが、短くても読める範囲だけ見る。
  const upto = Math.min(human.cards.length, state.handCaptures.length);
  let best = -1;
  let bestTaken = 0;
  let bestClears = false;
  for (let i = 0; i < upto; i += 1) {
    for (const set of state.handCaptures[i]) {
      // **場を全部さらう手が最優先。**スコパはそれ自体で 1 点。
      const clears = tableSize > 0 && set.length === tableSize;
      const better = bestClears ? clears && set.length > bestTaken : clears || set.length > bestTaken;
      if (!better) continue;
      best = i;
      bestTaken = set.length;
      bestClears = bestClears || clears;
    }
  }

  // **札 0 も取り手になりうる。**真偽値で見ると先頭だけ落ちる。
  if (best >= 0) {
    return bestClears
      ? { targetAction: `card-${best}`, reason: 'frontendHint.scoponeScopa', confidence: 'strong' }
      : { targetAction: `card-${best}`, reason: 'frontendHint.scoponeCaptureMost', confidence: 'moderate' };
  }

  // 取れないなら一番小さい札を置く。相手に取らせにくい。
  return {
    targetAction: `card-${lowestIdx(human.cards)}`,
    reason: 'frontendHint.scoponeLayLow',
    confidence: 'moderate',
  };
}

/** 手札で一番小さいランクの位置。 */
function lowestIdx(hand: Card[]): number {
  let best = 0;
  for (let i = 1; i < hand.length; i += 1) {
    if (hand[i].value < hand[best].value) best = i;
  }
  return best;
}
