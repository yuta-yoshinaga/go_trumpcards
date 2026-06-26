import type { OpenFaceChinesePlayer, OpenFaceChineseResponse } from '../../../types/card';
import { OpenFaceChinesePhase } from '../../../types/phases';
import { formatCard, formatCardList, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [OpenFaceChinesePhase.PLACING]: 'PLACING',
  [OpenFaceChinesePhase.ROUND_END]: 'ROUND END',
  [OpenFaceChinesePhase.GAME_END]: 'GAME END',
};

/** Format one player's three rows plus any round-end scoring. */
function formatPlayer(p: OpenFaceChinesePlayer, showScores: boolean): string[] {
  const lines: string[] = [];
  const name = p.isHuman ? 'You' : `CPU ${p.id}`;
  const tags = [p.fouled ? 'FOULED' : '', p.fantasyland ? 'FANTASYLAND' : ''].filter(Boolean).join(' ');
  lines.push(`${name}${tags ? ` (${tags})` : ''}`);
  lines.push(`  front:  ${formatCardList(p.front) || '-'}`);
  lines.push(`  middle: ${formatCardList(p.middle) || '-'}`);
  lines.push(`  back:   ${formatCardList(p.back) || '-'}`);
  if (showScores) {
    const royalty = p.royalty > 0 ? ` (royalty +${p.royalty})` : '';
    lines.push(`  round: ${p.roundScore}${royalty}  total: ${p.totalScore}`);
  }
  return lines;
}

/** Format an Open Face Chinese Poker game state as terminal text. */
export function formatOpenfacechineseState(state: OpenFaceChineseResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Open Face Chinese Poker'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);

  if (state.currentCard && state.isHumanTurn) {
    lines.push(`Place this card: ${formatCard(state.currentCard)} (front/middle/back)`);
  }
  lines.push('');

  const showScores = state.phase !== OpenFaceChinesePhase.PLACING;
  for (const p of state.players) {
    lines.push(...formatPlayer(p, showScores));
    lines.push('');
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
