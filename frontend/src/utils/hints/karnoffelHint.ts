import type { KarnoffelResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 人間の手番であることを示すフェーズ (0 = Play)。 */
const PLAY_PHASE = 0;

/**
 * Returns a frontend {@link HintResult} for Karnöffel, or null when no
 * suggestion is available.
 *
 * `validPlays` comes from the server, and the type says why it has to: there is
 * **no obligation to follow suit** here, but the devil cannot lead the first
 * trick, so legality is not something the client can read off the trick.
 *
 * The other thing worth saying is how close the hand is. Three tricks take a
 * hand out of five cards, so "one more trick wins this" arrives sooner than a
 * player used to full-length trick games expects.
 */
export function getKarnoffelHint(state: KarnoffelResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== PLAY_PHASE) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  const plays = state.validPlays;
  if (plays.length === 0) return null;

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  if (plays.length === 1) {
    return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.karnoffelForced', confidence: 'strong' };
  }

  // あと 1 トリックで手が決まる局面。5 枚のうち 3 トリックなので早く来る。
  const mine = state.teamTricks[human.team];
  if (mine === state.tricksToWin - 1) {
    return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.karnoffelOneAway', confidence: 'moderate' };
  }

  return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.karnoffelChoose', confidence: 'moderate' };
}
