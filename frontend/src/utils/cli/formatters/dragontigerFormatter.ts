import type { DragonTigerResponse } from '../../../types/card';
import { DragonTigerBetType, DragonTigerPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [DragonTigerPhase.BET]: 'BET',
  [DragonTigerPhase.END]: 'END',
};

const BET_NAMES: Record<number, string> = {
  [DragonTigerBetType.DRAGON]: 'Dragon',
  [DragonTigerBetType.TIGER]: 'Tiger',
  [DragonTigerBetType.TIE]: 'Tie',
};

// Domain GameResult: 1=Dragon wins, 2=Tiger wins, 3=Tie.
const RESULT_NAMES: Record<number, string> = { 1: 'Dragon', 2: 'Tiger', 3: 'Tie' };

/** Format a Dragon Tiger game state as terminal text. */
export function formatDragonTigerState(state: DragonTigerResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('Dragon Tiger'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);

  if (state.betAmount > 0) {
    lines.push(`bet: ${state.betAmount} on ${BET_NAMES[state.betType] ?? '?'}`);
  }

  if (state.dragonCard || state.tigerCard) {
    lines.push(`Dragon: ${state.dragonCard ? formatCard(state.dragonCard) : '??'}`);
    lines.push(`Tiger: ${state.tigerCard ? formatCard(state.tigerCard) : '??'}`);
  }

  if (state.phase === DragonTigerPhase.END) {
    lines.push(`result: ${RESULT_NAMES[state.result] ?? 'UNKNOWN'}  payout: ${state.payout}`);
  }

  if (state.history.length > 0) {
    lines.push(`history: ${state.history.map((h) => BET_NAMES[h] ?? '?').join(' ')}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
