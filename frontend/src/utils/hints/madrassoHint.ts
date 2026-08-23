import type { Card, MadrassoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { MadrassoPhase } from '../../types/phases';

/**
 * Madrasso trick strength (higher beats lower within the led suit).
 *
 * Order: **A > 3 > K > Q > J > 7 > 6 > 5 > 4 > 2** — the Briscola-family order,
 * where the highest-scoring card is also the strongest. This must stay in step
 * with `madrassoStrength` in `internal/domain/Madrasso.go`; the clone source
 * (Tressette) ranks 3 and 2 above the Ace, which is a different game.
 */
function strength(value: number): number {
  switch (value) {
    case 1: // Asso — strongest
      return 9;
    case 3: // Tre
      return 8;
    case 13: // Re
      return 7;
    case 12: // Cavallo
      return 6;
    case 11: // Fante
      return 5;
    case 7:
      return 4;
    case 6:
      return 3;
    case 5:
      return 2;
    case 4:
      return 1;
    default: // 2 — weakest
      return 0;
  }
}

/**
 * Reports whether `challenger` beats `best`.
 *
 * Mirrors `madrassoBeats` in `internal/domain/Madrasso.go`: trump beats the led
 * suit, the led suit beats everything else, and ties break on rank strength.
 */
function beats(challenger: Card, best: Card, ledSuit: string, trumpSuit: number): boolean {
  const cTrump = suitCode(challenger.design) === trumpSuit;
  const bTrump = suitCode(best.design) === trumpSuit;
  if (cTrump !== bTrump) return cTrump;
  if (cTrump && bTrump) return strength(challenger.value) > strength(best.value);
  if (challenger.design !== ledSuit) return false;
  if (best.design !== ledSuit) return true;
  return strength(challenger.value) > strength(best.value);
}

/** Maps a card design to the numeric suit code the API uses (1=♠ 2=♣ 3=♥ 4=♦). */
function suitCode(design: Card['design']): number {
  switch (design) {
    case 'SPADE':
      return 1;
    case 'CLOVER':
      return 2;
    case 'HEART':
      return 3;
    default:
      return 4;
  }
}

/** Returns a frontend HintResult for Madrasso, or null if no suggestion. */
export function getMadrassoHint(state: MadrassoResponse): HintResult | null {
  if (state.phase !== MadrassoPhase.PLAY) return null;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx < 0 || state.currentPlayerIdx !== humanIdx) return null;
  const human = state.players[humanIdx];
  if (!human || human.cards.length === 0) return null;

  const trick = state.currentTrick;
  if (trick.length === 0) {
    return { targetAction: 'play', reason: 'hint.leadLow', confidence: 'moderate' };
  }

  const ledSuit = trick[0].card.design;
  const trumpSuit = state.trumpSuit;
  const suitCards = human.cards.filter((c: Card) => c.design === ledSuit);

  // **切り札がある。** クローン元 (トレセッテ) には無いので、その形のままだと
  // リードスートを持たないときに「捨てろ」としか言えず、切り札で取れる場面を
  // 見逃す。追従できないときだけ切り札を切れる。
  if (suitCards.length === 0) {
    const hasTrump = human.cards.some((c: Card) => suitCode(c.design) === trumpSuit);
    if (hasTrump) {
      return { targetAction: 'play', reason: 'hint.trumpIn', confidence: 'moderate' };
    }
    return { targetAction: 'play', reason: 'hint.discardLow', confidence: 'moderate' };
  }

  // 場を勝っている札。切り札はどの平札にも勝つ。
  let topCard = trick[0].card;
  let topPlayerIdx = trick[0].playerIdx;
  for (const tc of trick) {
    if (beats(tc.card, topCard, ledSuit, trumpSuit)) {
      topCard = tc.card;
      topPlayerIdx = tc.playerIdx;
    }
  }
  const partnerWinning = topPlayerIdx % 2 === humanIdx % 2;
  if (partnerWinning) {
    return { targetAction: 'play', reason: 'hint.givePartner', confidence: 'moderate' };
  }
  const canWin = suitCards.some((c: Card) => beats(c, topCard, ledSuit, trumpSuit));
  if (canWin) {
    return { targetAction: 'play', reason: 'hint.followWin', confidence: 'strong' };
  }
  return { targetAction: 'play', reason: 'hint.followDuck', confidence: 'moderate' };
}
