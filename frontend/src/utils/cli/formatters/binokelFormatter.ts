import type { BinokelResponse } from '../../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const SUIT_NAMES: Record<number, string> = { 1: 'Spade', 2: 'Clover', 3: 'Heart', 4: 'Diamond' };

/** Format a Binokel game state as terminal text. */
export function formatBinokelState(state: BinokelResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Binokel'));
  lines.push(`round: ${state.roundNumber}  trick: ${state.trickNumber}`);
  if (state.trumpSuit > 0) {
    const suitName = SUIT_NAMES[state.trumpSuit] ?? '?';
    if (state.highestBidder >= 0) {
      const bidderName = formatPlayerName(state.highestBidder, state.players[state.highestBidder]?.isHuman ?? false);
      lines.push(`trump: ${suitName} (declarer: ${bidderName})`);
    } else {
      lines.push(`trump: ${suitName}`);
    }
  }
  if (state.highestBid > 0) {
    if (state.highestBidder >= 0) {
      const bidderName = formatPlayerName(state.highestBidder, state.players[state.highestBidder]?.isHuman ?? false);
      lines.push(`highest bid: ${state.highestBid} (${bidderName})`);
    } else {
      lines.push(`highest bid: ${state.highestBid}`);
    }
  }
  if (state.players && state.players.length > 0) {
    const scoreParts = state.players.map((p, i) => {
      const name = formatPlayerName(p.id, p.isHuman);
      const score = state.scores ? state.scores[i] : (p.score ?? 0);
      return `${name}=${score}`;
    });
    lines.push(`scores: ${scoreParts.join('  ')}`);
  }
  lines.push('');

  if (state.dabb && state.dabb.length > 0) {
    lines.push(`Dabb: ${formatCardList(state.dabb)}`);
  }

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const bidStr = p.hasPassed ? 'bid=[Passed]' : p.bid > 0 ? `bid=${p.bid}` : 'bid=-';
    const parts = [
      `score=${p.score ?? 0}`,
      bidStr,
      `meld=${p.meldScore}`,
      `tricks=${p.trickCount}T/${p.trickPoints}pts`,
    ];
    if (state.highestBidder === p.id) {
      parts.push('[Declarer]');
    }
    lines.push(`${name}: ${parts.join(' ')}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.playerMelds) {
    for (let i = 0; i < state.playerMelds.length; i++) {
      const melds = state.playerMelds[i];
      if (melds && melds.length > 0) {
        const name = formatPlayerName(i, state.players[i]?.isHuman ?? false);
        const meldStrs = melds.map((m) => `${m.points}pts(${formatCardList(m.cards)})`);
        lines.push(`${name} melds: ${meldStrs.join(', ')}`);
      }
    }
  }

  if (state.currentTrick && state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  if (state.hint && isRequestedHint(state)) lines.push(`HINT: ${state.hint.reason}`);

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winnerName = formatPlayerName(state.winnerPlayer, state.players[state.winnerPlayer]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winnerName}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
