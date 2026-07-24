import type { FourCardPokerResponse } from '../../../types/card';
import { FourCardPokerPhase } from '../../../types/phases';
import { formatCardList, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [FourCardPokerPhase.BET]: 'BET',
  [FourCardPokerPhase.ACTION]: 'ACTION',
  [FourCardPokerPhase.END]: 'END',
};

/** Format a Four Card Poker game state as terminal text. */
export function formatFourCardPokerState(state: FourCardPokerResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Four Card Poker'));
  lines.push(`chips: ${state.chips} | phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`ante: ${state.anteBet} | acesUp: ${state.acesUpBet} | play: ${state.playBet}`);
  lines.push('----------');

  if (state.phase === FourCardPokerPhase.BET) {
    lines.push('Place a bet: b <ante> [acesUp]');
  } else {
    lines.push(`your hand: ${formatCardList(state.playerHand) || '-'}`);
    // During the action phase only the dealer upcard is revealed.
    const dealerLabel = state.phase === FourCardPokerPhase.END ? 'dealer hand' : 'dealer up';
    lines.push(`${dealerLabel}: ${formatCardList(state.dealerHand) || '-'}`);

    if (state.phase === FourCardPokerPhase.ACTION) {
      lines.push('Make your play bet: p <1|2|3>, or fold (f)');
    }

    if (state.phase === FourCardPokerPhase.END) {
      lines.push('----------');
      lines.push(`your best:   ${formatCardList(state.playerBest) || '-'}`);
      lines.push(`dealer best: ${formatCardList(state.dealerBest) || '-'}`);
      lines.push('----------');
      lines.push(`ante payout: ${state.antePayout} | ante bonus: ${state.anteBonusPayout}`);
      lines.push(`play payout: ${state.playPayout} | acesUp: ${state.acesUpPayout}`);
      lines.push(`total payout: ${state.totalPayout}`);
    }
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
