import type { SixBidSoloResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 人間の手番であることを示すフェーズ (2 = Play)。 */
const PLAY_PHASE = 2;

/** ミゼール系の宣言。3 = Misère, 5 = Spread Misère。取らないほうが勝つ。 */
const MISERE_KINDS: readonly number[] = [3, 5];

/**
 * Returns a frontend {@link HintResult} for Six-Bid Solo, or null when no
 * suggestion is available.
 *
 * `validPlays` comes from the server since following suit is compulsory.
 *
 * The declaration inverts the goal on two of the six bids: under a **misère**
 * the declarer wins by taking nothing, so both sides want the opposite of what
 * the other four bids reward. The bid kind is read from `highBid`, because
 * `trumpSuit` is 0 for both misères and cannot say which is live.
 */
export function getSixBidSoloHint(state: SixBidSoloResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== PLAY_PHASE) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  const plays = state.validPlays;
  if (plays.length === 0) return null;

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  if (plays.length === 1) {
    return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.sixbidsoloForced', confidence: 'strong' };
  }

  // **ミゼールは取らないほうが勝ち。**目標そのものが逆になる。
  if (state.highBid && MISERE_KINDS.includes(state.highBid.kind)) {
    const reason =
      human.id === state.declarerIdx ? 'frontendHint.sixbidsoloMisereDuck' : 'frontendHint.sixbidsoloMisereForce';
    return { targetAction: `card-${plays[0]}`, reason, confidence: 'moderate' };
  }

  return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.sixbidsoloChoose', confidence: 'moderate' };
}
