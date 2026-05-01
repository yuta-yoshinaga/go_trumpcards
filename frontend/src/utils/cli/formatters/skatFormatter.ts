import type { SkatResponse } from '../../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
} from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'BID',
  1: 'SKAT_PICKUP',
  2: 'DISCARD',
  3: 'GAME_DECLARATION',
  4: 'PLAY',
  5: 'TRICK_END',
  6: 'ROUND_END',
  7: 'GAME_END',
};

const GAME_TYPE_NAMES: Record<number, string> = {
  0: 'NONE',
  1: 'SUIT',
  2: 'GRAND',
  3: 'NULL',
};

const SUIT_NAMES: Record<number, string> = {
  0: '-',
  1: '♣',
  2: '♠',
  3: '♥',
  4: '♦',
};

/** Format a Skat game state as terminal text. */
export function formatSkatState(state: SkatResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Skat'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`,
  );
  lines.push(
    `bid: ${state.currentBid}  game: ${GAME_TYPE_NAMES[state.gameType] ?? '?'}  trump: ${SUIT_NAMES[state.trumpSuit] ?? '?'}`,
  );
  if (state.declarerIdx >= 0) {
    lines.push(
      `declarer: P${state.declarerIdx}  declarerPts: ${state.declarerCardPoints}  defendersPts: ${state.defendersCardPoints}`,
    );
  }
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const turnMark = p.id === state.currentPlayerIdx ? ' *' : '';
    const role = p.isDeclarer ? ' [DECL]' : '';
    lines.push(
      `${name}${turnMark}${role}: cumScore=${p.cumulativeScore} round=${p.roundScore} tricks=${p.trickCount} bid=${p.bid}`,
    );
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  hand: ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trick = state.currentTrick.map((tc) => `P${tc.playerIdx}:${formatCard(tc.card)}`).join('  ');
    lines.push(`trick: ${trick}`);
  }

  if (state.skat && state.skat.length > 0 && state.pickedSkat) {
    lines.push(`skat: ${formatCardList(state.skat)}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(`Game Over! winner side: ${state.winnerSide}  game value: ${state.gameValue}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
