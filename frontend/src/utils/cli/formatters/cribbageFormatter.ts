import type { CribbageResponse } from '../../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
} from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'DISCARD',
  1: 'CUT',
  2: 'PEGGING',
  3: 'SHOW',
  4: 'ROUND END',
  5: 'GAME END',
};

/** Format a Cribbage game state as terminal text. */
export function formatCribbageState(state: CribbageResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Cribbage'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  if (state.starter) lines.push(`starter: ${formatCard(state.starter)}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const dealer = state.dealerIdx === p.id ? ' [Dealer]' : '';
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore}${dealer}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.phase === 2) {
    lines.push(`peg count: ${state.pegCount}`);
    if (state.pegPlayedCards.length > 0) {
      lines.push(`played: ${formatCardList(state.pegPlayedCards)}`);
    }
  }

  if (state.crib.length > 0 && state.phase >= 3) {
    lines.push(`crib: ${formatCardList(state.crib)}`);
  }

  if (state.handScoreDetails) {
    for (let i = 0; i < state.handScoreDetails.length; i++) {
      const d = state.handScoreDetails[i];
      if (d) {
        const name = formatPlayerName(i, state.players[i]?.isHuman ?? false);
        lines.push(
          `${name} score: 15s=${d.fifteens} pairs=${d.pairs} runs=${d.runs} flush=${d.flush} nobs=${d.nobs} total=${d.total}`,
        );
      }
    }
  }

  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentPlayerIdx, state.players[state.currentPlayerIdx]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
