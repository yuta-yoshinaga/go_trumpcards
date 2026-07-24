import type { UltiResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Discard', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const CONTRACT_NAMES = ['-', 'Party', 'Betli', 'Durchmarsch', 'Ulti'];
const SUIT_NAMES = ['-', 'spade', 'club', 'heart', 'diamond'];
const OUTCOME_NAMES = ['-', 'Made (declarer wins)', 'Failed (coalition wins)'];

/** Format an Ulti (Ultimo) game state as terminal text. */
export function formatUltiState(state: UltiResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Ulti'));
  lines.push(
    `deal: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  const trumpName = state.trumpSuit >= 1 ? (SUIT_NAMES[state.trumpSuit] ?? '-') : '-';
  lines.push(`contract: ${CONTRACT_NAMES[state.contract] ?? state.contract}  trump: ${trumpName}`);
  lines.push(`coins: ${state.playerCoins.map((c, i) => `P${i}=${c}`).join('  ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDeclarer ? 'Declarer' : 'Coalition';
    lines.push(`${name} (${role}): cards=${p.cardCount} tricks=${p.trickCount} pts=${p.cardPoints} coins=${p.coins}`);
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

  if ((state.phase === 4 || state.phase === 5) && state.outcome > 0) {
    lines.push(`deal result: ${OUTCOME_NAMES[state.outcome] ?? state.outcome}`);
  }

  if (state.hint) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerPlayer >= 0) {
    lines.push(`Game Over! Winner: Player ${state.winnerPlayer}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
