import type { HorseResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Hand', 'HandEnd', 'GameEnd'];
const DISCIPLINE_NAMES: Record<string, string> = {
  holdem: "Texas Hold'em",
  omahaHiLo: 'Omaha Hi-Lo',
  razz: 'Razz',
  stud: 'Seven-Card Stud',
  studHiLo: 'Stud Hi-Lo',
};

/** Format a H.O.R.S.E. game state as terminal text. */
export function formatHorseState(state: HorseResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('H.O.R.S.E.'));
  lines.push(
    `${state.disciplineLetter} — ${DISCIPLINE_NAMES[state.disciplineName] ?? state.disciplineName}  ` +
      `hand ${state.handInDiscipline}/${state.config.handsPerDiscipline} (${state.handNumber} overall)  ` +
      `phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`pot: ${state.pot}  to call: ${state.toCall}`);
  if (state.communityCards.length > 0) {
    lines.push(`board: ${state.communityCards.map(formatCard).join(' ')}`);
  }
  lines.push(formatSeparator());

  for (const s of state.seats) {
    // **見えている札だけを並べる。** CPU の伏せ札はサーバが返さない。
    const cards = s.cards.length > 0 ? `  ${s.cards.map(formatCard).join(' ')}` : '';
    const turn = s.id === state.currentTurn && state.phase === 0 ? ' <-' : '';
    lines.push(`${s.name}${s.isHuman ? ' (you)' : ''}: ${s.chips} chips${cards}${turn}`);
  }

  if (state.gameEndFlag) {
    lines.push(formatSeparator());
    lines.push(`winner: ${state.seats[state.winnerSeat]?.name ?? '-'}`);
  }

  if (state.message) lines.push(state.message);
  return lines.join('\n');
}
