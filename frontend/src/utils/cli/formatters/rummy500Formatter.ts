import type { Rummy500Response } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASES = ['Draw', 'Play', 'RoundEnd', 'GameEnd'];

/** Format a Rummy 500 game state for terminal play. */
export function formatRummy500State(state: Rummy500Response): string {
  const lines = [
    formatHeader('Rummy 500'),
    `round: ${state.roundNumber}  phase: ${PHASES[state.phase] ?? state.phase}`,
    `stock: ${state.drawPileCount}  discard: ${state.discardPile.length}`,
  ];
  for (const player of state.players) {
    lines.push(
      `${formatPlayerName(player.id, player.isHuman)}: cards=${player.cardCount} score=${player.cumulativeScore} round=${player.roundScore}`,
    );
    if (player.isHuman && player.cards.length > 0) lines.push(`  ${formatIndexedCards(player.cards)}`);
    for (const [i, meld] of player.laidMelds.entries())
      lines.push(`  meld ${player.id}:${i} (${meld.score}): ${meld.cards.map(formatCard).join(' ')}`);
  }
  if (state.discardPile.length)
    lines.push(`discard pile: ${state.discardPile.map((c, i) => `[${i}]${formatCard(c)}`).join(' ')}`);
  lines.push(formatSeparator());
  lines.push(
    state.phase === 0
      ? 'Use d to draw from stock or take <idx> from discard.'
      : state.phase === 1
        ? `Use meld <idx> ... or discard <idx>.${state.layoffTargets?.some((targets) => targets.length > 0) ? ' Use layoff <hand-idx> <owner> <meld-idx> when available.' : ''}`
        : state.phase === 2
          ? 'Use nr to start the next round.'
          : 'Game complete.',
  );
  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerIdx >= 0 && state.winnerIdx < state.players.length)
    lines.push(
      `Game Over! ${formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false)} wins!`,
    );
  return lines.join('\n');
}
