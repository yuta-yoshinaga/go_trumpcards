import type { FrenchTarotResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Chien', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const CONTRACT_NAMES = ['-', 'Petite', 'Garde', 'Garde Sans', 'Garde Contre'];
const OUTCOME_NAMES = ['-', 'Made (declarer wins)', 'Failed (defenders win)'];

/** Format a French Tarot (フレンチタロット) game state as terminal text. */
export function formatFrenchTarotState(state: FrenchTarotResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('French Tarot'));
  lines.push(
    `deal: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(
    `contract: ${CONTRACT_NAMES[state.contract] ?? state.contract}  highestBid: ${CONTRACT_NAMES[state.highestBid] ?? state.highestBid}`,
  );
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

  if (state.chienRevealed && state.chien.length > 0) {
    lines.push(`chien: ${state.chien.map((c) => formatCard(c)).join(' ')}`);
  }

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

  if (state.hint && isRequestedHint(state)) {
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
