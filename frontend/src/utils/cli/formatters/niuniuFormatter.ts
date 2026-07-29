import type { NiuNiuHand, NiuNiuResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

/**
 * Render one hand. Visibility comes from the server's `hidden` flag, never from
 * re-deriving it here: a hidden hand arrives with null cards.
 *
 * The three cards that formed the bull are marked with `*`. Five cards and a
 * rank name with nothing connecting them cannot be read.
 */
function formatHand(hand: NiuNiuHand): string {
  if (hand.hidden) return hand.cards.map(() => '[??]').join(' ');
  const inCombo = new Set(hand.comboIdx);
  const cards = hand.cards.map((c, i) => `${inCombo.has(i) ? '*' : ''}${c ? formatCard(c) : '[??]'}`).join(' ');
  const mult = hand.multiplier > 1 ? ` (x${hand.multiplier})` : '';
  return `${cards} ${hand.rankLabel}${mult}`;
}

/** Format a Niu Niu game state as terminal text. */
export function formatNiuNiuState(state: NiuNiuResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Niu Niu'));
  lines.push(`chips: ${state.chips}`);

  if (state.bankerHand) {
    lines.push(`banker: ${formatHand(state.bankerHand)}`);
  }
  lines.push('----------');

  state.seats.forEach((seat, seatIdx) => {
    if (seatIdx === state.bankerIdx || !seat.hand) return;
    const payout = seat.hand.payout !== 0 ? ` -> ${seat.hand.payout}` : '';
    lines.push(`  ${seat.name} bet ${seat.hand.bet} ${formatHand(seat.hand)}${payout}`);
  });
  lines.push('----------');

  if (state.lastResult) lines.push(state.lastResult);
  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
