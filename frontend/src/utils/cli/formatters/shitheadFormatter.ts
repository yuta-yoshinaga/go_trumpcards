import type { ShitheadResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

/** Format a Shithead game state as terminal text. */
export function formatShitheadState(state: ShitheadResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Shithead'));
  const top = state.discardPile.length > 0 ? formatCard(state.discardPile[state.discardPile.length - 1]) : '[  ]';
  lines.push(`stock: ${state.stockSize}  discard top: ${top}  pile: ${state.discardPile.length}`);
  if (state.skipNext) lines.push('next player skipped (8)');
  if (state.sevenActive) lines.push('seven active (next must play <= 7)');
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const turnMark = p.id === state.currentTurn ? ' *' : '';
    const status = p.isFinished ? ` [done #${p.rank}]` : '';
    lines.push(
      `${name}${turnMark}${status}: hand=${p.handCount} faceUp=${p.faceUpCards.length} faceDown=${p.faceDownCount}`,
    );
    if (p.isHuman) {
      if (p.handCards.length > 0) {
        lines.push(`  hand: ${formatIndexedCards(p.handCards)}`);
      }
      if (p.faceUpCards.length > 0) {
        lines.push(`  faceUp: ${formatIndexedCards(p.faceUpCards)}`);
      }
    }
  }
  lines.push('----------');

  lines.push(`turn: ${state.currentTurn}  source: ${state.currentSource || 'hand'}`);

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const loser = state.players.find((p) => !p.isFinished);
    lines.push(`Game Over! ${loser ? `${formatPlayerName(loser.id, loser.isHuman)} is the Shithead.` : ''}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
