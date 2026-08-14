import type { BotifarraResponse } from '../../../types/card';
import { BOTIFARRA_NO_TRUMP, BOTIFARRA_TOTAL_POINTS } from '../../../types/games/botifarra';
import { BotifarraPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [BotifarraPhase.DECLARE]: 'DECLARE',
  [BotifarraPhase.DELEGATED]: 'DELEGATED',
  [BotifarraPhase.DOUBLE]: 'DOUBLE',
  [BotifarraPhase.PLAY]: 'PLAY',
  [BotifarraPhase.ROUND_END]: 'ROUND END',
  [BotifarraPhase.GAME_END]: 'GAME END',
};

const TRUMP_NAMES: Record<number, string> = {
  1: 'Spades',
  2: 'Clubs',
  3: 'Hearts',
  4: 'Diamonds',
};

/** Renders the trump suit, naming no-trump rather than printing -1. */
function trumpName(suit: number): string {
  return suit === BOTIFARRA_NO_TRUMP ? 'No trump' : (TRUMP_NAMES[suit] ?? '?');
}

/** Format a Botifarra game state as terminal text. */
export function formatBotifarraState(state: BotifarraResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('Botifarra'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(
    `score: your team ${state.scores[0] ?? 0} / theirs ${state.scores[1] ?? 0}` +
      (state.config ? ` (game to ${state.config.targetScore})` : ''),
  );
  lines.push(`trump: ${trumpName(state.trumpSuit)} (x${state.multiplier})`);
  lines.push(`this round: ${state.roundPoints[0] ?? 0} / ${state.roundPoints[1] ?? 0} of ${BOTIFARRA_TOTAL_POINTS}`);

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
    // **出せる札に印を付ける。** 勝つ義務があるので大半が出せない場面がある。
    lines.push('only cards marked * are legal');
  }

  if (state.gameEndFlag) {
    lines.push(`winner: team ${state.winnerTeam}`);
  }
  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
