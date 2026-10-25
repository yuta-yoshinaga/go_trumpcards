import type { BriscolaResponse } from '../../../types/card';
import { BriscolaPhase } from '../../../types/phases';
import { suitSymbolAt } from '../../cardAlt';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

/** Format a Briscola game state as terminal text. */
export function formatBriscolaState(state: BriscolaResponse): string {
  const lines = [formatHeader('Briscola')];
  const phase =
    state.phase === BriscolaPhase.PLAY ? 'Play' : state.phase === BriscolaPhase.TRICK_END ? 'TrickEnd' : 'GameEnd';
  lines.push(`trick: ${state.trickNumber}  phase: ${phase}`);
  lines.push(
    `points: You=${state.players.find((p) => p.isHuman)?.points ?? 0}  CPU=${state.players.find((p) => !p.isHuman)?.points ?? 0}`,
  );
  lines.push(
    `trump: ${state.trumpSuit > 0 ? suitSymbolAt(state.trumpSuit, '?') : 'none'}  stock: ${state.stockRemaining}`,
  );
  for (const player of state.players) {
    lines.push(
      `${formatPlayerName(player.id, player.isHuman)}: ${player.cardCount} cards, ${player.points} points, ${player.trickCount} tricks`,
    );
    if (player.isHuman && player.cards.length > 0)
      lines.push(`  ${formatIndexedCards(player.cards)}  (play with p <idx>)`);
  }
  if (state.currentTrick.length > 0) {
    lines.push(
      `trick cards: ${state.currentTrick.map(({ playerIdx, card }) => `${formatPlayerName(playerIdx, state.players[playerIdx]?.isHuman ?? false)}=${formatCard(card)}`).join(', ')}`,
    );
  }
  if (state.phase === BriscolaPhase.PLAY) lines.push('Available: p <idx>, h / hint, r / reset');
  if (state.phase === BriscolaPhase.TRICK_END) lines.push('Available: n / next, h / hint, r / reset');
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      `HINT: ${state.hint.cardIndex === undefined ? state.hint.reason : `play card [${state.hint.cardIndex}] (${state.hint.reason})`}`,
    );
  } else if (state.messageCode === 'briscola.noHint') {
    lines.push('No hint available.');
  }
  if (state.gameEndFlag)
    lines.push(state.winnerIdx === -1 ? 'Game Over! Draw.' : `${state.winnerIdx === 0 ? 'You win!' : 'CPU wins!'}`);
  lines.push(formatSeparator());
  return lines.join('\n');
}
