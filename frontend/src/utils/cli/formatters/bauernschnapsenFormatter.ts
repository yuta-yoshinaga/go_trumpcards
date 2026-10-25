import type { BauernschnapsenResponse } from '../../../types/card';
import { BauernschnapsenPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const CONTRACTS: Record<number, string> = { 0: 'Undeclared', 1: 'Rufer', 2: 'Farbenzwang', 3: 'Bettel' };
const SUITS: Record<number, string> = { 1: 'Spades', 2: 'Clubs', 3: 'Hearts', 4: 'Diamonds' };
const PHASE_NAMES: Record<number, string> = {
  [BauernschnapsenPhase.CONTRACT]: 'Contract',
  [BauernschnapsenPhase.PLAY]: 'Play',
  [BauernschnapsenPhase.TRICK_END]: 'TrickEnd',
  [BauernschnapsenPhase.ROUND_END]: 'RoundEnd',
  [BauernschnapsenPhase.GAME_END]: 'GameEnd',
};

/** Format a Bauernschnapsen game state as terminal text. */
export function formatBauernschnapsenState(state: BauernschnapsenResponse): string {
  const lines = [
    formatHeader('Bauernschnapsen'),
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
    `contract: ${CONTRACTS[state.contract] ?? 'Unknown'}${state.trumpSuit ? `  trump: ${SUITS[state.trumpSuit] ?? '?'}` : ''}`,
    `score: Team0=${state.teamScores[0]} Team1=${state.teamScores[1]} (target ${state.config.targetScore})`,
    '',
  ];
  const currentPlayer = state.players[state.currentPlayerIdx];
  const humanTurn = currentPlayer?.isHuman ?? false;
  lines.splice(3, 0, `current player: ${formatPlayerName(state.currentPlayerIdx, humanTurn)}`);
  if (state.declarerIdx >= 0) {
    const declarer = state.players[state.declarerIdx];
    lines.splice(
      4,
      0,
      `declarer: ${formatPlayerName(state.declarerIdx, declarer?.isHuman ?? false)} — ${CONTRACTS[state.contract] ?? 'Unknown'}`,
    );
  }
  for (const p of state.players) {
    lines.push(`${formatPlayerName(p.id, p.isHuman)}: team=${p.team} tricks=${p.trickCount}`);
    if (p.isHuman && p.cards.length) {
      const markCards = humanTurn && state.validPlayIndices.length > 0;
      const markMarriages = humanTurn && state.marriageIndices.length > 0;
      const hand = p.cards
        .map((card, i) => {
          const mark = `${markMarriages && state.marriageIndices.includes(i) ? '*' : ''}${markCards && !state.validPlayIndices.includes(i) ? 'x' : ''}`;
          return `[${i}]${formatCard(card)}${mark}`;
        })
        .join('  ');
      lines.push(`  ${hand}`);
      if (humanTurn) {
        if (state.phase === BauernschnapsenPhase.PLAY) {
          lines.push(`  use p <idx>${state.marriageIndices.length ? ' or m <idx>' : ''}`);
        } else if (state.phase === BauernschnapsenPhase.CONTRACT) lines.push('  use c <0-3> [suit]');
      }
    }
  }
  if (humanTurn && state.phase === BauernschnapsenPhase.TRICK_END) lines.push('use n');
  if (humanTurn && state.phase === BauernschnapsenPhase.ROUND_END) lines.push('use nr');
  lines.push('----------');
  if (state.currentTrick.length)
    lines.push(
      `trick: ${state.currentTrick.map((tc) => `${formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false)}=${formatCard(tc.card)}`).join(', ')}`,
    );
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      state.hint.isMarriage
        ? `HINT: declare a marriage with card [${state.hint.cardIndex ?? '?'}] (${state.hint.reason})`
        : state.hint.cardIndex !== undefined
          ? `HINT: play card [${state.hint.cardIndex}] (${state.hint.reason})`
          : `HINT: ${state.hint.reason}`,
    );
  }
  if (state.messageCode === 'bauernschnapsen.noHint') lines.push('No hint available.');
  else if (state.message) lines.push(state.message);
  if (state.gameEndFlag)
    lines.push(state.winnerTeam >= 0 ? `Game Over! Team ${state.winnerTeam} wins!` : 'Game Over! Draw.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
