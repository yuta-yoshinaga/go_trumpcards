import type { SpoonsResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Pass', 'Grab', 'RoundEnd', 'GameEnd'];

/** Render the S-P-O-O-N-S letter progress for a given count (0–6). */
function formatLetters(letters: number): string {
  const word = 'SPOONS';
  return word
    .slice(0, Math.max(0, Math.min(letters, word.length)))
    .split('')
    .join('-');
}

/** Format a Spoons game state as terminal text. */
export function formatSpoonsState(state: SpoonsResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Spoons'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`spoons left: ${state.spoonsRemaining}  draw pile: ${state.drawPileSize}`);
  lines.push('');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    const status = p.eliminated ? 'OUT' : p.hasSpoon ? 'has spoon' : `${p.handSize} cards`;
    const letters = formatLetters(p.letters) || '-';
    lines.push(`${name}: letters=${letters} [${status}]`);
    if (p.hand.length > 0) {
      lines.push(`  ${formatIndexedCards(p.hand)}`);
    }
  });
  lines.push('----------');

  if (state.grabWindowOpen) {
    lines.push('(GRAB! type "g" to grab a spoon)');
  } else if (state.phase === 0 && state.isHumanTurn) {
    lines.push('(your turn — pass a card with "p <i>")');
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerIdx >= 0) {
    lines.push(`Game Over! Winner: ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
