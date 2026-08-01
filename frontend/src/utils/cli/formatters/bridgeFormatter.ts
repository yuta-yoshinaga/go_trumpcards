import type { BridgeResponse } from '../../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const SUIT_NAMES: Record<number, string> = { 1: 'Clubs', 2: 'Diamonds', 3: 'Hearts', 4: 'Spades', 5: 'NT' };
const BID_TYPE_NAMES: Record<number, string> = { 0: 'Pass', 1: 'Double', 2: 'Redouble', 3: 'Bid' };

/** Format a Bridge game state as terminal text. */
export function formatBridgeState(state: BridgeResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Contract Bridge'));
  lines.push(`round: ${state.roundNumber}  trick: ${state.trickNumber}`);
  if (state.contractLevel > 0) {
    const doubled = state.doubled === 1 ? ' X' : state.doubled === 2 ? ' XX' : '';
    lines.push(`contract: ${state.contractLevel}${SUIT_NAMES[state.contractSuit] ?? '?'}${doubled}`);
  }
  if (state.teamScores.length >= 2) {
    lines.push(`score: NS=${state.teamScores[0]} EW=${state.teamScores[1]}`);
    lines.push(`games: NS=${state.gamesWon[0]} EW=${state.gamesWon[1]}`);
  }
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const roles: string[] = [];
    if (state.declarerIdx === p.id) roles.push('Declarer');
    if (state.dummyIdx === p.id) roles.push('Dummy');
    const roleStr = roles.length > 0 ? ` [${roles.join(', ')}]` : '';
    lines.push(`${name}: team=${p.team} tricks=${p.trickCount}${roleStr}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }

  if (state.dummyHand && state.dummyHand.length > 0) {
    lines.push(`dummy: ${formatCardList(state.dummyHand)}`);
  }
  lines.push('----------');

  if (state.bidHistory.length > 0) {
    const bids = state.bidHistory.map((b) => {
      if (b.bidType !== 3) return BID_TYPE_NAMES[b.bidType] ?? '?';
      return `${b.level}${SUIT_NAMES[b.suit] ?? '?'}`;
    });
    lines.push(`bids: ${bids.join(', ')}`);
  }

  if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  if (state.hint && isRequestedHint(state)) lines.push(`HINT: ${state.hint.reason}`);

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push(`Game Over! Winner: Team ${state.winnerTeam}`);

  lines.push(formatSeparator());
  return lines.join('\n');
}
