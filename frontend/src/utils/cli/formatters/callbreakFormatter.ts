import type { CallBreakResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

/** Render a Call Break int×10 score as "X.Y" / "-X.Y" for terminal display. */
function fmtScore(internal: number): string {
  const sign = internal < 0 ? '-' : '';
  const n = Math.abs(internal);
  return `${sign}${Math.trunc(n / 10)}.${n % 10}`;
}

/** Format a Call Break game state as terminal text. */
export function formatCallBreakState(state: CallBreakResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Call Break'));
  lines.push(
    `round: ${state.roundNumber}/${state.config.maxRounds}  trick: ${state.trickNumber}  spades broken: ${state.spadesBroken ? 'yes' : 'no'}`,
  );
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(
      `${name}: total=${fmtScore(p.cumulativeScore)} round=${fmtScore(p.roundScore)} bid=${p.bid} tricks=${p.trickCount}`,
    );
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  if (state.phase === 0) lines.push('Bidding phase (1-13, no Nil)');

  if (state.hint && isRequestedHint(state)) {
    if (state.hint.bid !== undefined) lines.push(`HINT: bid ${state.hint.bid} (${state.hint.reason})`);
    if (state.hint.cardIndex !== undefined) lines.push(`HINT: play [${state.hint.cardIndex}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Game Over');

  lines.push(formatSeparator());
  return lines.join('\n');
}
