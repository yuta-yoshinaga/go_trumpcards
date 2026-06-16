import type { CourtPieceResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['TrumpDeclaration', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUIT_SYMBOLS = ['none', '♠', '♣', '♥', '♦'];

/** Format a Court Piece (Rang) game state as terminal text. */
export function formatCourtPieceState(state: CourtPieceResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Court Piece / Rang'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  const trumpText = state.trumpSuit === 0 ? 'undeclared' : (SUIT_SYMBOLS[state.trumpSuit] ?? '?');
  lines.push(`trump: ${trumpText}`);
  if (state.callerIdx >= 0) {
    const name = formatPlayerName(state.callerIdx, state.players[state.callerIdx]?.isHuman ?? false);
    const team = state.players[state.callerIdx]?.team ?? state.callerIdx % 2;
    lines.push(`caller: ${name} (team ${team})`);
  }
  lines.push(`game points: A=${state.teamScores[0] ?? 0}  B=${state.teamScores[1] ?? 0}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.id === state.callerIdx ? 'Caller' : 'Player';
    lines.push(`${name} (team ${p.team}, ${role}): cards=${p.cardCount} tricks=${p.trickCount}`);
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
    const teamATricks = state.players.filter((p) => p.team === 0).reduce((s, p) => s + p.trickCount, 0);
    const teamBTricks = state.players.filter((p) => p.team === 1).reduce((s, p) => s + p.trickCount, 0);
    lines.push(`round result: team A=${teamATricks} tricks  team B=${teamBTricks} tricks`);
    if (state.lastRoundCourt) lines.push('Court bonus scored!');
  }

  if (state.hint) {
    if (state.hint.trumpSuit != null) {
      lines.push(`HINT: declare trump ${SUIT_SYMBOLS[state.hint.trumpSuit] ?? '?'} (${state.hint.reason})`);
    } else if (state.hint.cardIndex != null) {
      lines.push(`HINT: play card index [${state.hint.cardIndex}] (${state.hint.reason})`);
    }
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerTeam >= 0) {
    lines.push(`Game Over! Winner: Team ${state.winnerTeam === 0 ? 'A' : 'B'}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
