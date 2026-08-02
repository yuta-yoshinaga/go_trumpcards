import type { KoenigrufenResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Call', 'Talon', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const CONTRACT_NAMES = ['-', 'Rufer'];
const OUTCOME_NAMES = ['-', 'Made (declarer team wins)', 'Failed (defenders win)'];
const SUIT_NAMES = ['-', 'Spades', 'Clubs', 'Hearts', 'Diamonds'];

/** Format a Königrufen (ケーニッヒルーフェン) game state as terminal text. */
export function formatKoenigrufenState(state: KoenigrufenResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Königrufen'));
  lines.push(
    `deal: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(
    `contract: ${CONTRACT_NAMES[state.contract] ?? state.contract}  highestBid: ${CONTRACT_NAMES[state.highestBid] ?? state.highestBid}`,
  );
  if (state.calledKing >= 1) {
    lines.push(`called King: ${SUIT_NAMES[state.calledKing] ?? state.calledKing}`);
  }
  lines.push(`scores: ${state.playerScores.map((s, i) => `P${i}=${s}`).join('  ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDeclarer ? 'Declarer' : state.partnerRevealed && p.isPartner ? 'Partner' : 'Defender';
    lines.push(`${name} (${role}): cards=${p.cardCount} tricks=${p.trickCount} pts=${p.cardPoints} score=${p.score}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.isHumanDiscard && state.talon.length > 0) {
    lines.push(`talon: ${state.talon.map((c) => formatCard(c)).join(' ')}`);
  }

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
