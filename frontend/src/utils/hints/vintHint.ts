import type { VintResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 人間の手番であることを示すフェーズ (1 = Play)。 */
const PLAY_PHASE = 1;

/**
 * Returns a frontend {@link HintResult} for Vint, or null when no suggestion is
 * available.
 *
 * `validPlays` comes from the server — following suit is compulsory — so the
 * legal set is read rather than recomputed.
 *
 * What Vint adds is that **both** sides score below the line for the tricks
 * they take, made contract or not. So there is no such thing as a trick that
 * does not matter, and a defender has a reason to take tricks rather than
 * merely deny them — which is the opposite of the habit most trick-takers
 * teach.
 */
export function getVintHint(state: VintResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== PLAY_PHASE) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  const plays = state.validPlays;
  if (plays.length === 0) return null;

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  if (plays.length === 1) {
    return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.vintForced', confidence: 'strong' };
  }

  const declaring = human.id === state.declarerIdx || partnerOf(state.declarerIdx) === human.id;
  return {
    targetAction: `card-${plays[0]}`,
    reason: declaring ? 'frontendHint.vintDeclaring' : 'frontendHint.vintDefending',
    confidence: 'moderate',
  };
}

/** 相方は対角の席。`team` は `id % 2` と同じ並び。 */
function partnerOf(seat: number): number {
  return seat < 0 ? -1 : (seat + 2) % 4;
}
