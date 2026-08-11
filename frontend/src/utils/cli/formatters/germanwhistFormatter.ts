import type { GermanWhistResponse } from '../../../types/card';
import { GermanWhistPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [GermanWhistPhase.DRAW]: 'FIRST HALF (no score)',
  [GermanWhistPhase.SCORING]: 'SECOND HALF (scoring)',
  [GermanWhistPhase.GAME_END]: 'GAME END',
};

// trumpSuit is a 1-based suit code, as elsewhere in this repo.
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Format a German Whist game state as terminal text. */
export function formatGermanWhistState(state: GermanWhistResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('German Whist'));
  lines.push(`trick ${state.trickNumber}/26 | ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'} | stock: ${state.stockCount}`);
  // The face-up card is the whole point of the first half; say so when it's gone.
  lines.push(`face-up: ${state.upCard ? formatCard(state.upCard) : '(none — stock exhausted)'}`);
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trick = state.currentTrick
      .map((tc) => `${formatPlayerName(tc.playerIdx, false)}:${formatCard(tc.card)}`)
      .join('  ');
    lines.push(`trick: ${trick}`);
    lines.push('----------');
  }

  state.players.forEach((p) => {
    const marker = p.id === state.currentPlayerIdx && !state.gameEndFlag ? '>' : ' ';
    lines.push(
      `${marker}${formatPlayerName(p.id, p.isHuman)}: ${p.scoringTricks} scoring | ${p.trickCount} total | ${p.cardCount} cards`,
    );
  });

  const human = state.players.find((p) => p.isHuman);
  if (human) {
    lines.push('----------');
    const hand = human.cards
      .map((c, i) => `[${i}]${formatCard(c)}${state.validPlays.includes(i) ? '*' : ''}`)
      .join('  ');
    lines.push(`your hand: ${hand || '(empty)'}`);
  }

  if (state.gameEndFlag) {
    lines.push('----------');
    lines.push(
      state.winnerIdx >= 0
        ? `game over — winner: ${formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false)}`
        : 'game over — tie',
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
