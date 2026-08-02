import type { Card, CuarentaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 人間の手番であることを示すフェーズ (0=Play)。 */
const PLAY_PHASE = 0;

/**
 * Returns a frontend {@link HintResult} for Cuarenta, or null when no
 * suggestion is available.
 *
 * The response carries no list of legal captures, so this works from what a
 * player sees: a hand card whose rank matches a table card takes it, and one
 * that matches the card just played takes it as a **caída**, which is worth
 * more. The run extension that follows a capture is deliberately not modelled —
 * it depends on the domain's ordering rules, and guessing at it here would put
 * a second copy of those rules in the frontend.
 */
export function getCuarentaHint(state: CuarentaResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== PLAY_PHASE) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentTurn !== human.id) return null;

  // **カイーダは直前に出された札と同じランク。**ふつうの取りより点が高い。
  const lastPlayed = lastPlayedCard(state);
  if (lastPlayed) {
    const idx = human.cards.findIndex((c) => c.value === lastPlayed.value);
    if (idx >= 0) {
      return { targetAction: `card-${idx}`, reason: 'frontendHint.cuarentaCaida', confidence: 'strong' };
    }
  }

  // **札 0 も取り手になりうる。**真偽値で見ると先頭だけ落ちる。
  const capture = human.cards.findIndex((c) => state.tableCards.some((t) => t.value === c.value));
  if (capture >= 0) {
    return { targetAction: `card-${capture}`, reason: 'frontendHint.cuarentaCapture', confidence: 'strong' };
  }

  return {
    targetAction: `card-${lowestIdx(human.cards)}`,
    reason: 'frontendHint.cuarentaLayLow',
    confidence: 'moderate',
  };
}

/** 直前に場へ出された札。誰も出していなければ null。 */
function lastPlayedCard(state: CuarentaResponse): Card | null {
  for (let i = state.cpuActions.length - 1; i >= 0; i -= 1) {
    const played = state.cpuActions[i].playedCard;
    if (played) return played;
  }
  return null;
}

/** 手札で一番小さいランクの位置。 */
function lowestIdx(hand: Card[]): number {
  let best = 0;
  for (let i = 1; i < hand.length; i += 1) {
    if (hand[i].value < hand[best].value) best = i;
  }
  return best;
}
