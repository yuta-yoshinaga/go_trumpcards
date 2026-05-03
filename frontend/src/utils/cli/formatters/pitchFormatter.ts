import type { PitchResponse } from '../../../types/card';
import { PitchPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [PitchPhase.BID]: 'BID',
  [PitchPhase.PLAY]: 'PLAY',
  [PitchPhase.TRICK_END]: 'TRICK END',
  [PitchPhase.ROUND_END]: 'ROUND END',
  [PitchPhase.GAME_END]: 'GAME END',
};

const SUIT_NAMES: Record<number, string> = {
  0: '(unset)',
  1: 'Spades',
  2: 'Clubs',
  3: 'Hearts',
  4: 'Diamonds',
};

/** Format a Pitch game state as terminal text. */
export function formatPitchState(state: PitchResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('Pitch'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`,
  );
  lines.push(`dealer: ${state.dealerIdx}  bid: ${state.currentBid}  trump: ${SUIT_NAMES[state.trumpSuit] ?? '?'}`);
  if (state.bidWinnerIdx >= 0) {
    lines.push(`bidder: player ${state.bidWinnerIdx}`);
  }
  for (const p of state.players) {
    const bidStr = p.bid === -1 ? 'not bid' : p.bid === 0 ? 'pass' : String(p.bid);
    const tag = p.isHuman ? 'YOU' : `CPU${p.id}`;
    lines.push(
      `[${tag}] bid=${bidStr} tricks=${p.trickCount} round=${p.roundScore} total=${p.cumulativeScore} cards=${p.cardCount}`,
    );
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  hand: ${p.cards.map((c, i) => `[${i}]${formatCard(c)}`).join(' ')}`);
    }
  }
  if (state.currentTrick.length > 0) {
    lines.push(`trick: ${state.currentTrick.map((tc) => `P${tc.playerIdx}:${formatCard(tc.card)}`).join('  ')}`);
  }
  if (state.message) lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
