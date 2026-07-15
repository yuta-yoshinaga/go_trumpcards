import type { EasthavenResponse } from '../../../types/card';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Format Easthaven state for CLI display. */
export function formatEasthavenState(state: EasthavenResponse): string {
  const lines: string[] = [];
  lines.push(`Stock: ${state.stockCount}`);
  lines.push('Foundation:');
  for (let i = 0; i < state.foundation.length; i++) {
    const pile = state.foundation[i];
    const top = pile.length > 0 ? `${pile[pile.length - 1].design}-${pile[pile.length - 1].value}` : 'empty';
    lines.push(`  ${FOUNDATION_SUITS[i]}: ${top} (${pile.length})`);
  }
  lines.push('');
  lines.push('Tableau:');
  for (let col = 0; col < state.tableau.length; col++) {
    const cards = state.tableau[col]
      .map((tc, i) => (tc.faceUp && tc.card ? `[${i}]${tc.card.design}-${tc.card.value}` : `[${i}]??`))
      .join(' ');
    lines.push(`  ${col}: ${cards || '(empty)'}`);
  }
  lines.push('');
  lines.push(`Moves: ${state.moveCount}  Phase: ${state.phase}`);
  return lines.join('\n');
}
