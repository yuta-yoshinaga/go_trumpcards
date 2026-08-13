import type { CincinnatiResponse } from '../../../types/card';
import { CincinnatiPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [CincinnatiPhase.DEAL]: 'DEAL',
  [CincinnatiPhase.BETTING]: 'BETTING',
  [CincinnatiPhase.SHOWDOWN]: 'SHOWDOWN',
  [CincinnatiPhase.GAME_END]: 'GAME END',
};

/** Format a Cincinnati game state as terminal text. */
export function formatCincinnatiState(state: CincinnatiResponse): string {
  const lines: string[] = [formatHeader('Cincinnati')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Hand: ${state.handNumber} (pot: ${state.pot})`);

  lines.push(formatSeparator());
  // **あと何枚めくれるかを出す。** 残りの回数だけベットラウンドがある。
  const board = state.community.length > 0 ? state.community.map(formatCard).join(' ') : '-';
  lines.push(`Board: ${board}  (${state.revealedCount} of ${state.communityTotal} shown)`);

  state.seats.forEach((seat) => {
    const mark = seat.isTurn ? '*' : ' ';
    const state_ = seat.folded ? ' (folded)' : seat.allIn ? ' (all in)' : '';
    // **CPU の手札はサーバが送っていない。** 空なら伏せ表示。
    const cards = seat.cards.length > 0 ? seat.cards.map(formatCard).join(' ') : '(face down)';
    lines.push(`${mark}${seat.name}${state_} chips ${seat.chips} / bet ${seat.bet} : ${cards}`);
    if (seat.wonAmount > 0) lines.push(`   won ${seat.wonAmount}`);
  });

  if (state.phase === CincinnatiPhase.BETTING) {
    lines.push(state.toCall > 0 ? `${state.toCall} to call` : 'You may check');
  }
  if (state.gameEndFlag) {
    lines.push(`Winner: ${state.seats[state.winnerSeat]?.name ?? '?'}`);
  }

  return lines.join('\n');
}
