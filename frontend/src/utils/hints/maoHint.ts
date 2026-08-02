import type { Card, MaoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { MaoPhase } from '../../types/phases';

/** 8 は Mao でも万能札。いつでも出せる。 */
const WILD = 8;

/** スート番号から Card の design 名への対応 (`chosenSuit` は番号で届く)。 */
const SUIT_NUM_TO_DESIGN: Record<number, Card['design']> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/**
 * Returns a frontend {@link HintResult} for Mao, or null when no suggestion is
 * available.
 *
 * Unlike most games in this directory, Mao's hint is computed here rather than
 * mirrored from the Go backend: the server deliberately exposes no `hint` field
 * at all, only `hintUnlocked` and a partial `ruleHint`. That is the whole shape
 * of the game — the rules are secret and are learned by being penalised for
 * breaking them.
 *
 * So this hint stays strictly on the **visible** layer, which is Crazy Eights:
 * an 8 is wild; otherwise a card must follow the called suit, or match the top
 * card's suit or rank. It never touches the hidden rule, and it says nothing
 * during the declaration phase — filling that in would delete the game.
 */
export function getMaoHint(state: MaoResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const human = state.players[humanIdx];
  if (!human || human.cards.length === 0) return null;

  // **宣言フェーズには触れない。**言うべき言葉は隠しルールそのもので、
  // 解放済みの `ruleHint` が示す以上のことをここが漏らしてはいけない。
  if (state.phase === MaoPhase.MUST_DECLARE || state.awaitingWord) return null;

  if (state.phase === MaoPhase.CHOOSE_SUIT) {
    return { targetAction: 'chooseSuit', reason: 'frontendHint.maoChooseSuit', confidence: 'strong' };
  }

  if (state.phase !== MaoPhase.PLAY || state.currentPlayerIdx !== humanIdx) return null;
  return playHint(human.cards, state);
}

/** 表の（隠されていない）ルールだけで、出せる札を選ぶ。 */
function playHint(cards: Card[], state: MaoResponse): HintResult {
  const wilds = cards.filter((c) => c.value === WILD);
  const ordinary = cards.filter((c) => c.value !== WILD);
  const top = state.discardTop;

  if (!top) {
    return { targetAction: 'play', reason: 'frontendHint.maoMatchSuit', confidence: 'moderate' };
  }

  // **8 が出た後は、呼ばれたスートだけが通る。**表の札のランクに合わせても
  // 出せないので、ここで rank も見ると規則違反の手を勧めることになる。
  const called = state.chosenSuit > 0 ? SUIT_NUM_TO_DESIGN[state.chosenSuit] : undefined;
  const suitMatch = ordinary.some((c) => c.design === (called ?? top.design));
  const rankMatch = called === undefined && ordinary.some((c) => c.value === top.value);

  if ((suitMatch || rankMatch) && wilds.length > 0) {
    return { targetAction: 'play', reason: 'frontendHint.maoSaveWild', confidence: 'moderate' };
  }
  if (suitMatch) {
    return { targetAction: 'play', reason: 'frontendHint.maoMatchSuit', confidence: 'strong' };
  }
  if (rankMatch) {
    return { targetAction: 'play', reason: 'frontendHint.maoMatchRank', confidence: 'strong' };
  }
  if (wilds.length > 0) {
    return { targetAction: 'play', reason: 'frontendHint.maoPlayWild', confidence: 'strong' };
  }
  return { targetAction: 'draw', reason: 'frontendHint.maoDraw', confidence: 'moderate' };
}
