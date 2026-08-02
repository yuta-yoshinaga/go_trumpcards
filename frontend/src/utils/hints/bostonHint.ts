import type { BostonResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 人間の手番であることを示すフェーズ (2 = Play)。 */
const PLAY_PHASE = 2;

/**
 * Returns a frontend {@link HintResult} for Boston, or null when no suggestion
 * is available.
 *
 * `validPlays` comes from the server — following suit is compulsory — so the
 * legal set is read rather than recomputed, as in Klaberjass and Bourré.
 *
 * The extra thing worth saying is whose side the player is on. A called partner
 * counts toward the declarer's contract (`declarerTricks` is the pair's total),
 * so a defender and a partner want opposite outcomes from the same trick, and
 * "which side am I on" is the question a caller-partner game makes easy to lose
 * track of.
 */
export function getBostonHint(state: BostonResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== PLAY_PHASE) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  const plays = state.validPlays;
  if (plays.length === 0) return null;

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  if (plays.length === 1) {
    return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.bostonForced', confidence: 'strong' };
  }

  const onDeclaringSide = human.id === state.declarerIdx || human.id === state.partnerIdx;
  return {
    targetAction: `card-${plays[0]}`,
    reason: onDeclaringSide ? 'frontendHint.bostonPushContract' : 'frontendHint.bostonBreakContract',
    confidence: 'moderate',
  };
}
