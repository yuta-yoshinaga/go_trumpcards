import type { KaiserResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** 人間の手番であることを示すフェーズ (2 = Play)。 */
const PLAY_PHASE = 2;

/**
 * Returns a frontend {@link HintResult} for Kaiser, or null when no suggestion
 * is available.
 *
 * `validPlays` comes from the server — following suit is compulsory, trumping
 * is not — so the legal set is read rather than recomputed.
 *
 * The one piece of Kaiser-specific advice worth adding is the ♠3. It is worth
 * −3 to whoever takes the trick containing it, so holding it is a liability and
 * shedding it while you still have a choice is better than being forced to play
 * it into a trick you win. The ♥5 is the mirror image at +5, and the hint
 * deliberately does **not** push it out: leading it hands the opponents a
 * chance to take it.
 */
export function getKaiserHint(state: KaiserResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== PLAY_PHASE) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  const plays = state.validPlays;
  if (plays.length === 0) return null;

  // **札 0 も強制手になりうる。**真偽値で見ると先頭だけ落ちる。
  if (plays.length === 1) {
    return { targetAction: `card-${plays[0]}`, reason: 'frontendHint.kaiserForced', confidence: 'strong' };
  }

  // **♠3 は −3 点。**出せるうちに手放す。
  const spadeThree = plays.find((i) => human.cards[i]?.design === 'SPADE' && human.cards[i]?.value === 3);
  if (spadeThree !== undefined) {
    return { targetAction: `card-${spadeThree}`, reason: 'frontendHint.kaiserDumpSpadeThree', confidence: 'moderate' };
  }

  // **♥5 は +5 点。**自分から出すと相手に取られうるので、他に出せるなら避ける。
  const heartFive = (i: number) => human.cards[i]?.design === 'HEART' && human.cards[i]?.value === 5;
  const pick = plays.find((i) => !heartFive(i)) ?? plays[0];
  return { targetAction: `card-${pick}`, reason: 'frontendHint.kaiserChoose', confidence: 'moderate' };
}
