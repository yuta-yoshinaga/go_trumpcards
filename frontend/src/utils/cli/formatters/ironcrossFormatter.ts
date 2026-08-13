import type { IronCrossResponse } from '../../../types/card';
import { IronCrossPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [IronCrossPhase.BETTING]: 'BETTING',
  [IronCrossPhase.CHOOSE_LINE]: 'CHOOSE LINE',
  [IronCrossPhase.SHOWDOWN]: 'SHOWDOWN',
  [IronCrossPhase.GAME_END]: 'GAME END',
};

const LINE_NAMES: Record<number, string> = { 1: 'vertical', 2: 'horizontal' };

/** Cross positions, matching the Go domain's layout. */
const CENTER = 0;
const TOP = 1;
const BOTTOM = 2;
const LEFT = 3;
const RIGHT = 4;

/** Format an Iron Cross game state as terminal text. */
export function formatIronCrossState(state: IronCrossResponse): string {
  const lines: string[] = [formatHeader('Iron Cross')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Hand: ${state.handNumber} (pot: ${state.pot})`);

  lines.push(formatSeparator());
  // **十字は十字の形で出す。** 1 行に並べると、どれが縦でどれが横か分からない。
  const at = (i: number) => {
    const card = state.cross[i];
    return card ? formatCard(card) : '[??]';
  };
  lines.push(`    ${at(TOP)}`);
  lines.push(`${at(LEFT)} ${at(CENTER)} ${at(RIGHT)}`);
  lines.push(`    ${at(BOTTOM)}`);
  lines.push(`Cross: ${state.revealedCount} of ${state.crossTotal} shown`);

  state.seats.forEach((seat) => {
    const mark = seat.isTurn ? '*' : ' ';
    const status = seat.folded ? ' (folded)' : seat.allIn ? ' (all in)' : '';
    // **CPU の手札も選んだ列もサーバが送っていない。** 空なら伏せ表示。
    const cards = seat.cards.length > 0 ? seat.cards.map(formatCard).join(' ') : '(face down)';
    const line = LINE_NAMES[seat.line] ? ` [${LINE_NAMES[seat.line]}]` : '';
    lines.push(`${mark}${seat.name}${status} chips ${seat.chips} / bet ${seat.bet}${line} : ${cards}`);
    if (seat.wonAmount > 0) lines.push(`   won ${seat.wonAmount}`);
  });

  if (state.isChoosing) {
    lines.push('Choose vertical or horizontal');
  } else if (state.phase === IronCrossPhase.BETTING) {
    lines.push(state.toCall > 0 ? `${state.toCall} to call` : 'You may check');
  }
  if (state.gameEndFlag) {
    lines.push(`Winner: ${state.seats[state.winnerSeat]?.name ?? '?'}`);
  }

  return lines.join('\n');
}
