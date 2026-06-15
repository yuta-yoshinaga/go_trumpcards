import type { FortyFivesResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUIT_SYMBOLS = ['none', '♠', '♣', '♥', '♦'];

/** Formats a contract value (0=Pass, otherwise the bid amount). */
function formatContract(contract: number): string {
  return contract === 0 ? 'Pass' : String(contract);
}

/** Returns the team index (0 or 1) for a seat; seats 0&2 = team 0, 1&3 = team 1. */
function teamOf(seat: number): number {
  return seat % 2;
}

/** Format an Auction Forty-Fives game state as terminal text. */
export function formatFortyFivesState(state: FortyFivesResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Auction Forty-Fives'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'}`);
  if (state.declarerIdx >= 0) {
    const name = formatPlayerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false);
    lines.push(`declarer: ${name} (team ${teamOf(state.declarerIdx)}) — ${formatContract(state.contract)}`);
  } else {
    lines.push(`bids: ${state.bids.map((b, i) => `P${i}=${formatContract(b)}`).join('  ')}`);
  }
  lines.push(`team scores: A=${state.teamScores[0] ?? 0}  B=${state.teamScores[1] ?? 0}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDeclarer ? 'Declarer' : 'Defender';
    lines.push(`${name} (team ${teamOf(p.id)}, ${role}): cards=${p.cardCount} tricks=${p.trickCount}`);
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

  if (state.phase === 3 || state.phase === 4) {
    lines.push(`round result: team A=${state.roundTeamPoints[0] ?? 0}  team B=${state.roundTeamPoints[1] ?? 0}`);
  }

  if (state.hint) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerTeam >= 0) {
    lines.push(`Game Over! Winner: Team ${state.winnerTeam === 0 ? 'A' : 'B'}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
