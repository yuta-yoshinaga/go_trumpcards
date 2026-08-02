import type { Card, KlaberjassResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { KlaberjassPhase } from '../../types/phases';

/** ベラは切り札の K と Q。 */
const KING = 13;
const QUEEN = 12;

/** 表向きの札を取るかどうかの目安。手札にそのスートが何枚あれば取るか。 */
const TAKE_TURN_UP_FROM = 3;

/**
 * Returns a frontend {@link HintResult} for Klaberjass, or null when no
 * suggestion is available.
 *
 * The legal plays are **not** recomputed here. The response's own comment says
 * why: following suit, trumping when void and overtrumping a trump lead are all
 * compulsory, and re-deriving that on the client always drifts. So the hint
 * reads `validPlays` and stays out of the ranking business — what it adds is the
 * two things a player actually misses: that the compulsory rules have left them
 * exactly one card, and that they are holding an unscored bela.
 */
export function getKlaberjassHint(state: KlaberjassResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human) return null;

  if (state.phase === KlaberjassPhase.BID_TURN_UP) {
    return turnUpHint(state, human.cards);
  }

  if (state.phase !== KlaberjassPhase.PLAY || state.currentPlayerIdx !== human.id) return null;

  // **ベラは 2 枚目を出すときに宣言する。**持っているのに気づかず崩すと
  // 20 点がそのまま消える。
  if (state.belaHolder === human.id && !state.belaScored && holdsBela(human.cards, state.trumpSuit)) {
    return { targetAction: 'bela', reason: 'frontendHint.klaberjassBela', confidence: 'strong' };
  }

  const plays = state.validPlays;
  if (plays.length === 0) return null;

  // **札 0 も強制手になりうる。**真偽値で見ると先頭の札だけ落ちる。
  if (plays.length === 1) {
    return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.klaberjassForced', confidence: 'strong' };
  }
  return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.klaberjassChoose', confidence: 'moderate' };
}

/** 表向きの札を取るか。手札にそのスートが十分あるかだけを見る。 */
function turnUpHint(state: KlaberjassResponse, hand: Card[]): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || state.bidPlayerIdx !== human.id || !state.turnUpCard) return null;

  const suit = state.turnUpCard.design;
  const inSuit = hand.filter((c) => c.design === suit).length;
  return inSuit >= TAKE_TURN_UP_FROM
    ? { targetAction: 'accept', reason: 'frontendHint.klaberjassTakeTurnUp', confidence: 'moderate' }
    : { targetAction: 'pass', reason: 'frontendHint.klaberjassPassTurnUp', confidence: 'moderate' };
}

/** 手札に切り札の K と Q が揃っているか。 */
function holdsBela(hand: Card[], trumpSuit: number): boolean {
  const suits: Record<number, Card['design']> = { 1: 'SPADE', 2: 'CLOVER', 3: 'HEART', 4: 'DIAMOND' };
  const design = suits[trumpSuit];
  if (!design) return false;
  return (
    hand.some((c) => c.design === design && c.value === KING) &&
    hand.some((c) => c.design === design && c.value === QUEEN)
  );
}
