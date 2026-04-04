import type { SpeedResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

/** Format a Speed game state as terminal text. */
export function formatSpeedState(state: SpeedResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Speed'));

  // Center piles
  if (state.centerPiles.length >= 2) {
    lines.push(`center: [0]${formatCard(state.centerPiles[0])}  [1]${formatCard(state.centerPiles[1])}`);
  }
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: hand=${p.cardCount} draw=${p.drawPileSize}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.phase === 1) lines.push('Both stuck! Type "flip" to flip new cards.');

  if (state.hint) {
    if (state.hint.found) {
      lines.push(`HINT: play [${state.hint.cardIndex}] to pile [${state.hint.pileIndex}]`);
    } else {
      lines.push('HINT: no valid plays');
    }
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
