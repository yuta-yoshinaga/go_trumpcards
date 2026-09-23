import type { BatakResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

/** Format a Batak game state as terminal text. */
export function formatBatakState(state: BatakResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Batak'));
  lines.push(
    `round: ${state.roundNumber}/${state.config.maxRounds}  trick: ${state.trickNumber}  spades broken: ${state.spadesBroken ? 'yes' : 'no'}`,
  );
  if (state.declarerIdx >= 0) {
    const declarerName = formatPlayerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false);
    lines.push(`declarer: ${declarerName}`);
  }
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const declarerTag = p.id === state.declarerIdx ? ' [declarer]' : '';
    const bidStr = p.bid < 0 ? '-' : p.bid === 0 ? 'pass' : String(p.bid);
    lines.push(
      `${name}${declarerTag}: total=${p.cumulativeScore} round=${p.roundScore} bid=${bidStr} tricks=${p.trickCount}`,
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

  if (state.phase === 0) lines.push('Bidding phase (5-13 or pass)');

  if (state.hint && isRequestedHint(state)) {
    if (state.hint.bid !== undefined) {
      const bidStr = state.hint.bid === 0 ? 'pass' : `bid ${state.hint.bid}`;
      lines.push(`HINT: ${bidStr} (${state.hint.reason})`);
    }
    if (state.hint.cardIndex !== undefined) lines.push(`HINT: play [${state.hint.cardIndex}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Game Over');

  lines.push(formatSeparator());
  return lines.join('\n');
}
