import type { SetteEMezzoHand, SetteEMezzoResponse } from '../../../types/card';
import { SetteEMezzoPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

/**
 * Render one hand. Visibility comes from the server's `hidden` flag, never from
 * re-deriving it here: a hidden hand arrives with null cards, so a second
 * opinion about what may be shown could only ever disagree with the data.
 */
function formatHand(hand: SetteEMezzoHand): string {
  if (hand.hidden) return hand.cards.map(() => '[??]').join(' ');
  const cards = hand.cards.map((c) => (c ? formatCard(c) : '[??]')).join(' ');
  // The matta's current value is printed because it stays adjustable until the
  // hand stands -- a value you cannot see is one you cannot choose.
  const matta = hand.hasMatta ? ` [matta=${(hand.mattaHalves || 1) / 2}]` : '';
  return `${cards} (${hand.totalLabel})${matta}`;
}

/** Format a Sette e Mezzo game state as terminal text. */
export function formatSetteEMezzoState(state: SetteEMezzoResponse): string {
  const lines: string[] = [];
  const ended = state.phase === SetteEMezzoPhase.END;

  lines.push(formatHeader('Sette e Mezzo'));
  lines.push(`chips: ${state.chips}`);
  const bankerName = state.isHumanBanker ? 'you' : (state.seats[state.bankerIdx]?.name ?? '?');
  lines.push(`banker: ${bankerName}`);

  if (state.bankerHand) {
    lines.push(`banker hand: ${formatHand(state.bankerHand)}`);
  }
  lines.push('----------');

  state.seats.forEach((seat, seatIdx) => {
    if (seatIdx === state.bankerIdx || !seat.hand) return;
    const onTurn = state.phase === SetteEMezzoPhase.PLAYER_TURN && seatIdx === state.activeSeat;
    const marker = onTurn ? '> ' : '  ';
    const payout = ended && seat.hand.payout !== 0 ? ` -> ${seat.hand.payout}` : '';
    lines.push(`${marker}${seat.name} bet ${seat.hand.bet} ${formatHand(seat.hand)}${payout}`);
  });
  lines.push('----------');

  const options: string[] = [];
  if (state.canHit) options.push('h');
  if (state.canStand) options.push('s');
  if (state.canSetMatta) options.push('matta');
  if (options.length > 0) lines.push(`available: ${options.join(' / ')}`);

  if (state.phase === SetteEMezzoPhase.BANKER_TURN) lines.push('banker: bh to draw, bs to stop');
  if (ended) lines.push(state.lastResult);
  if (ended && state.nextBanker >= 0) {
    lines.push(`${state.seats[state.nextBanker]?.name ?? '?'} takes the bank with exactly 7.5`);
  }
  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
