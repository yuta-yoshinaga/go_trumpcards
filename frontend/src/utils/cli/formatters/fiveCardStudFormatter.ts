import type { FiveCardStudResponse } from '../../../types/card';

/**
 * Format a Five Card Stud game state as terminal text.
 *
 * @param s - The current game state.
 * @param phaseNames - Phase-index → localized phase label lookup (from `usePhaseNames`).
 */
export function formatFiveCardStudState(s: FiveCardStudResponse, phaseNames: Record<number, string>): string {
  const lines: string[] = [];
  lines.push(`Phase: ${phaseNames[s.phase] ?? 'Init'} | Pot: ${s.pot}`);
  for (const p of s.players) {
    const tag = p.isHuman ? 'You' : `CPU ${p.id}`;
    const door = p.doorCards.map((c) => `${c.design[0]}${c.value}`).join(' ');
    const hole = p.holeCards.map((c) => `${c.design[0]}${c.value}`).join(' ');
    lines.push(
      `${tag}: chips=${p.chips} door=[${door}] hole=[${hole}]${p.folded ? ' FOLDED' : ''}${p.allIn ? ' ALL-IN' : ''}`,
    );
  }
  if (s.message) lines.push(s.message);
  return lines.join('\n');
}
