import { formatGanjifaSuit, type GanjifaResponse, isGanjifaStrongSuit } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];

/** Format a Ganjifa game state as terminal text. */
export function formatGanjifaState(state: GanjifaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Ganjifa'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`trump: ${formatGanjifaSuit(state.trumpSuit)}`);
  // Without this line the hand listing cannot be read: the same "3" is near the
  // bottom in a strong suit and near the top in a weak one.
  lines.push(
    isGanjifaStrongSuit(state.trumpSuit)
      ? 'trump group: strong — higher numbers are stronger (12 high)'
      : 'trump group: weak — lower numbers are stronger (1 high)',
  );
  lines.push(`scores: ${state.playerScores.map((s, i) => `P${i}=${s}`).join('  ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: cards=${p.cardCount} tricks=${p.trickCount} score=${p.score}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trickParts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${trickParts.join(', ')}`);
  }

  if (state.phase === 2 || state.phase === 3) {
    const tricks = state.roundTricks.map((v, i) => `P${i}=${v}`).join(' ');
    lines.push(`round result: tricks ${tricks}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    // A tie leaves winnerPlayer at -1; saying "Winner: Player -1" would be a lie.
    lines.push(state.winnerPlayer >= 0 ? `Game Over! Winner: Player ${state.winnerPlayer}` : 'Game Over! Draw.');
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
