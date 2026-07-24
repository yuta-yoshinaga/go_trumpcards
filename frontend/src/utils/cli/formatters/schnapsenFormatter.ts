import type { SchnapsenResponse } from '../../../types/card';
import { SchnapsenPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [SchnapsenPhase.PLAY]: 'PLAY',
  [SchnapsenPhase.TRICK_END]: 'TRICK END',
  [SchnapsenPhase.GAME_END]: 'GAME END',
};

// Schnapsen trumpSuit is a 1-based suit code (see DESIGN_TO_SUIT on the page).
const SUIT_SYMBOLS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Format a Schnapsen game state as terminal text. */
export function formatSchnapsenState(state: SchnapsenResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Schnapsen'));
  lines.push(
    `trick ${state.trickNumber} | phase: ${PHASE_NAMES[state.phase] ?? state.phase}${state.isEndgame ? ' (endgame)' : ''}`,
  );
  const trump = state.trumpCard ? formatCard(state.trumpCard) : (SUIT_SYMBOLS[state.trumpSuit] ?? '?');
  lines.push(`trump: ${trump} | stock: ${state.stockRemaining}`);
  lines.push('----------');

  // Current trick (cards played so far this trick).
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
      `${marker}${formatPlayerName(p.id, p.isHuman)}: ${p.points} pts | ${p.trickCount} tricks | ${p.cardCount} cards`,
    );
  });

  // Human hand, indexed; '*' = legal to play, 'M' = can declare a marriage.
  const human = state.players.find((p) => p.isHuman);
  if (human) {
    lines.push('----------');
    const hand = human.cards
      .map((c, i) => {
        const tags = `${state.validPlays.includes(i) ? '*' : ''}${state.marriagePlays.includes(i) ? 'M' : ''}`;
        return `[${i}]${formatCard(c)}${tags}`;
      })
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
