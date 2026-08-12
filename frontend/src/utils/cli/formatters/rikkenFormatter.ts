import type { RikkenResponse } from '../../../types/card';
import { RIKKEN_NO_TRUMP } from '../../../types/games/rikken';
import { RikkenContract, RikkenPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [RikkenPhase.BID]: 'BID',
  [RikkenPhase.CALL]: 'CALL',
  [RikkenPhase.PLAY]: 'PLAY',
  [RikkenPhase.ROUND_END]: 'ROUND END',
  [RikkenPhase.GAME_END]: 'GAME END',
};

const CONTRACT_NAMES: Record<number, string> = {
  [RikkenContract.NONE]: 'undecided',
  [RikkenContract.RIK]: 'Rik',
  [RikkenContract.MISERE]: 'Misere',
  [RikkenContract.SOLO]: 'Solo',
  [RikkenContract.OPEN_MISERE]: 'Open Misere',
};

const SUIT_NAMES: Record<number, string> = { 1: 'Spades', 2: 'Clubs', 3: 'Hearts', 4: 'Diamonds' };

/** Renders the trump, naming no-trump rather than printing -1. */
function trumpName(suit: number): string {
  return suit === RIKKEN_NO_TRUMP ? 'none' : (SUIT_NAMES[suit] ?? '?');
}

/** Format a Rikken game state as terminal text. */
export function formatRikkenState(state: RikkenResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('Rikken'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`round: ${state.roundNumber}${state.config ? ` / ${state.config.rounds}` : ''}`);
  lines.push(`contract: ${CONTRACT_NAMES[state.contract] ?? '?'} (trump: ${trumpName(state.trumpSuit)})`);
  if (state.declarerIdx >= 0) {
    lines.push(`declarer: seat ${state.declarerIdx} (${state.declarerTricks} tricks)`);
  }
  // **得点は負にもなります。** ゼロサムなので当然そうなります。
  lines.push(`scores: ${state.players.map((p) => `#${p.id}:${p.score}`).join(' ')}`);

  if (state.currentTrick.length > 0) {
    lines.push(formatSeparator());
    for (const tc of state.currentTrick) {
      lines.push(`  seat ${tc.playerIdx}: ${formatCard(tc.card)}`);
    }
  }

  const human = state.players.find((p) => p.isHuman);
  if (human && human.cards.length > 0) {
    lines.push(formatSeparator());
    const legal = new Set(state.validPlays);
    lines.push(`hand: ${human.cards.map((c, i) => `[${i}${legal.has(i) ? '*' : ' '}]${formatCard(c)}`).join(' ')}`);
  }

  if (state.gameEndFlag) lines.push(`winner: seat ${state.winnerIdx}`);
  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
