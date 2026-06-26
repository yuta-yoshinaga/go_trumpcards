import type { DeuceToSevenResponse } from '../../../types/card';
import { DeuceToSevenPhase } from '../../../types/phases';
import { deuceToSevenBestIndices, isMadePatLow } from '../../deuceToSevenUtils';
import { formatCardList, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'INIT',
  1: 'DEAL',
  2: 'BET',
  3: 'DRAW',
  4: 'SHOWDOWN',
  5: 'END',
};

/** Format a 2-7 Triple Draw game state as terminal text. */
export function formatDeuceToSevenState(state: DeuceToSevenResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('2-7 Triple Draw'));
  const phaseName = PHASE_NAMES[state.phase] ?? 'UNKNOWN';
  lines.push(`phase: ${phaseName}  draw: ${state.drawIndex}  pot: ${state.pot}`);
  lines.push(`dealer: ${state.dealerIdx}  turn: ${state.currentTurn}  lastBet: ${state.lastBet}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const status: string[] = [];
    if (p.folded) status.push('FOLD');
    if (p.allIn) status.push('ALL-IN');
    const statusStr = status.length > 0 ? ` [${status.join(',')}]` : '';
    lines.push(`${name}${statusStr}: chips=${p.chips} bet=${p.currentBet}`);
    if (p.handName) lines.push(`  hand: ${p.handName}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }

  // Draw-phase stand-pat guidance for the human, mirroring the GUI banner.
  const human = state.players.find((p) => p.isHuman);
  if (state.phase === DeuceToSevenPhase.DRAW && human && state.currentTurn === human.id && human.cards.length === 5) {
    if (isMadePatLow(human.cards)) {
      lines.push(`Made low: ${formatCardList(human.cards)} (stand recommended)`);
    } else {
      const keep = deuceToSevenBestIndices(human.cards).map((i) => human.cards[i]);
      lines.push(`Best to keep: ${formatCardList(keep)} (draw the rest)`);
    }
  }

  if (state.roundResults.length > 0) {
    lines.push('--- results ---');
    for (const r of state.roundResults) {
      lines.push(`  player ${r.playerIdx}: ${r.handName} (won ${r.wonAmount})`);
    }
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Game Over');

  lines.push(formatSeparator());
  return lines.join('\n');
}
