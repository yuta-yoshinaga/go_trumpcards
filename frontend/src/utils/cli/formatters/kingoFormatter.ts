import type { KingoResponse } from '../../../types/card';
import { KingoPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const RANK_NAMES: Record<number, string> = {
  0: 'no combination',
  1: 'pair',
  2: 'arashi',
};

/** Format a Kingo game state as terminal text. */
export function formatKingoState(state: KingoResponse): string {
  const lines: string[] = [formatHeader('Kingo')];

  const phaseName = state.phase === KingoPhase.BET ? 'BET' : state.phase === KingoPhase.RESULT ? 'RESULT' : 'GAME END';
  lines.push(`Phase: ${phaseName}`);
  lines.push(`Round: ${state.roundNumber} of ${state.rounds}`);
  lines.push(`Banker: ${state.seats[state.bankerSeat]?.name ?? '?'}`);
  // **配当はサーバが送ってくる。** 画面で倍率を持たない。
  lines.push(`Arashi pays ${state.payoutArashi}x, a pair pays ${state.payoutPair}x`);

  lines.push(formatSeparator());
  state.seats.forEach((seat) => {
    const mark = seat.isBanker ? '*' : ' ';
    // **親は張らないので空欄。**
    const bet = seat.isBanker ? '-' : String(seat.bet || '-');
    // **配る前は手札が存在しない。** 隠しているのではなく、まだ無い。
    const cards = seat.cards.length > 0 ? seat.cards.map(formatCard).join(' ') : '-';
    const rank = seat.cards.length > 0 ? ` ${RANK_NAMES[seat.rank] ?? ''}` : '';
    lines.push(`${mark}${seat.name} chips ${seat.chips} / bet ${bet} : ${cards}${rank}`);
  });

  if (state.phase === KingoPhase.BET && !state.gameEndFlag) {
    lines.push(state.isHumanBanker ? 'You hold the bank — deal to continue' : 'Place a bet');
  }
  if (state.phase !== KingoPhase.BET) {
    lines.push(formatSeparator());
    state.seats.forEach((seat) => {
      if (seat.wonAmount !== 0) lines.push(`${seat.name}: ${seat.wonAmount}`);
    });
  }
  if (state.gameEndFlag) {
    lines.push(`Winner: ${state.seats[state.winnerSeat]?.name ?? '?'}`);
  }

  return lines.join('\n');
}
