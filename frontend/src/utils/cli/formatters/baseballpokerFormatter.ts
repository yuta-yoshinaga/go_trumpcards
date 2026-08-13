import type { BaseballPokerResponse } from '../../../types/card';
import { BaseballPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [BaseballPhase.BETTING]: 'BETTING',
  [BaseballPhase.BUY_IN]: 'BUY THE POT',
  [BaseballPhase.SHOWDOWN]: 'SHOWDOWN',
  [BaseballPhase.GAME_END]: 'GAME END',
};

/** Format a Baseball Poker game state as terminal text. */
export function formatBaseballPokerState(state: BaseballPokerResponse): string {
  const lines: string[] = [formatHeader('Baseball Poker')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Hand: ${state.handNumber} (pot: ${state.pot})`);
  lines.push(`Street: ${state.street} of ${state.streetTotal}`);
  // **ワイルドとイベントの値はサーバが送ってくる。** 画面で 3 と 9 を持たない。
  lines.push(
    `Wild: ${state.wildValues.join(', ')} | face-up ${state.bonusValue} pays a card, face-up ${state.buyInValue} buys the pot`,
  );

  lines.push(formatSeparator());
  state.seats.forEach((seat, i) => {
    const mark = seat.isBuying ? '$' : seat.isTurn ? '*' : ' ';
    const status = seat.folded ? ' (folded)' : seat.allIn ? ' (all in)' : '';
    const bonus = seat.bonusCards > 0 ? ` (+${seat.bonusCards} bonus)` : '';
    // **届いていない札だけを伏せる。** 表札は全席ぶん届いている。
    const cards = seat.cards.map((card) => (card ? formatCard(card) : '[??]')).join(' ') || '-';
    lines.push(`${mark}${seat.name}${status} chips ${seat.chips} / bet ${seat.bet}${bonus} : ${cards}`);
    if (seat.wonAmount > 0) {
      lines.push(`   won ${seat.wonAmount}${seat.usedWild ? ' with wilds' : ''}`);
    }
    if (i === state.seats.length - 1) lines.push('');
  });

  if (state.isBuying) {
    lines.push(`A face-up ${state.buyInValue}. Pay ${state.buyCost} to stay in, or drop out`);
  } else if (state.phase === BaseballPhase.BETTING) {
    lines.push(state.toCall > 0 ? `${state.toCall} to call` : 'You may check');
  }
  if (state.gameEndFlag) {
    lines.push(`Winner: ${state.seats[state.winnerSeat]?.name ?? '?'}`);
  }

  return lines.join('\n');
}
