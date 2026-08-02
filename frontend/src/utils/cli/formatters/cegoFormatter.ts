import type { CegoResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Contract', 'Exchange', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const CONTRACT_TYPE_NAMES = ['-', 'Cego', 'Handspiel'];
const OUTCOME_NAMES = ['-', 'Made (declarer wins)', 'Failed (defenders win)'];

/** Format a Cego (チェゴ) game state as terminal text. */
export function formatCegoState(state: CegoResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Cego'));
  lines.push(
    `deal: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`contract: ${CONTRACT_TYPE_NAMES[state.contractType] ?? state.contractType}  blind: ${state.blindCount}`);
  lines.push(`scores: ${state.playerScores.map((s, i) => `P${i}=${s}`).join('  ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDeclarer ? 'Declarer' : 'Defender';
    lines.push(`${name} (${role}): cards=${p.cardCount} tricks=${p.trickCount} pts=${p.cardPoints} score=${p.score}`);
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

  if (state.phase === 5 && state.outcome > 0) {
    lines.push(`deal result: ${OUTCOME_NAMES[state.outcome] ?? state.outcome}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(state.winnerPlayer >= 0 ? `Game Over! Winner: Player ${state.winnerPlayer}` : 'Game Over! Draw!');
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
