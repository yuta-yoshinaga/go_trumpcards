import type { MadrassoResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

/** Suit glyphs by design value (1=Spade … 4=Diamond); index 0 is "not yet decided". */
const SUIT_NAMES = ['-', '♠', '♣', '♥', '♦'];

/** Format a Madrasso game state as terminal text. */
export function formatMadrassoState(state: MadrassoResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Madrasso'));
  // 切り札は配りで決まる。クローン元 (トレセッテ) に無い概念なので、出さないと
  // 盤面から切り札が読めない — CUI はこれを madrasso.trumpLine として出している
  // のに、Web の CLI モードだけが落としていた (#6615)。
  lines.push(`round: ${state.roundNumber}  trick: ${state.trickNumber}  trump: ${SUIT_NAMES[state.trumpSuit] ?? '-'}`);
  lines.push(
    `Team A: ${state.teamScores[0]} pts (this round ${state.teamRoundPoints[0]}/3)  ` +
      `Team B: ${state.teamScores[1]} pts (this round ${state.teamRoundPoints[1]}/3)`,
  );
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name} [Team ${p.teamId === 0 ? 'A' : 'B'}]: cards=${p.cardCount} tricks=${p.trickCount}`);
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

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    lines.push(`Game Over! Winner: Team ${state.winnerTeam === 0 ? 'A' : 'B'}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
