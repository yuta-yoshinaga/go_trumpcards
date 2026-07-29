import type { PontoonHand, PontoonResponse } from '../../../types/card';
import { PontoonPhase, PontoonRank } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

/** Label for a hand's rank. Points hands rank by total and need no label. */
function rankLabel(rank: number): string {
  if (rank === PontoonRank.PONTOON) return ' PONTOON';
  if (rank === PontoonRank.FIVE_CARD) return ' FIVE-CARD';
  if (rank === PontoonRank.BUST) return ' BUST';
  return '';
}

/** Render one hand, hiding the cards while they are still face down. */
function formatHand(hand: PontoonHand, hide: boolean): string {
  if (hide) return hand.cards.map(() => '[??]').join(' ');
  const cards = hand.cards.map(formatCard).join(' ');
  return `${cards} (${hand.total})${rankLabel(hand.rank)}`;
}

/** Format a Pontoon game state as terminal text. */
export function formatPontoonState(state: PontoonResponse): string {
  const lines: string[] = [];
  const ended = state.phase === PontoonPhase.END;

  lines.push(formatHeader('Pontoon'));
  lines.push(`chips: ${state.chips}`);
  const bankerName = state.isHumanBanker ? 'you' : (state.seats[state.bankerIdx]?.name ?? '?');
  lines.push(`banker: ${bankerName}`);

  if (state.bankerHand) {
    // The banker's cards are the ones you cannot see -- that is the difference
    // from blackjack, so they stay hidden until the round settles.
    lines.push(`banker hand: ${formatHand(state.bankerHand, !ended && !state.isHumanBanker)}`);
  }
  lines.push('----------');

  state.seats.forEach((seat, seatIdx) => {
    if (seatIdx === state.bankerIdx) return;
    seat.hands.forEach((hand, handIdx) => {
      const onTurn =
        state.phase === PontoonPhase.PLAYER_TURN && seatIdx === state.activeSeat && handIdx === state.activeHand;
      const marker = onTurn ? '> ' : '  ';
      const payout = ended && hand.payout !== 0 ? ` -> ${hand.payout}` : '';
      lines.push(`${marker}${seat.name} bet ${hand.bet} ${formatHand(hand, !ended && seat.isCpu)}${payout}`);
    });
  });
  lines.push('----------');

  // Only the legal declarations are listed, so the prompt never suggests
  // sticking below 15 or buying after a twist.
  const options: string[] = [];
  if (state.canStick) options.push('s');
  if (state.canTwist) options.push('t');
  if (state.canBuy) options.push('buy');
  if (state.canSplit) options.push('sp');
  if (options.length > 0) lines.push(`available: ${options.join(' / ')}`);

  if (state.phase === PontoonPhase.BANKER_TURN) lines.push('banker: bt to draw, bs to stop');
  if (ended) lines.push(state.lastResult);
  if (ended && state.nextBanker >= 0) {
    lines.push(`${state.seats[state.nextBanker]?.name ?? '?'} takes the bank`);
  }
  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
