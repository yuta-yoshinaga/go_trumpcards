import type { presidentApi } from '../../../api/gameApi';
import type { PresidentResponse } from '../../../types/card';
import type { CliParseResult } from '../types';

/** Args tuple accepted by presidentApi.exec. */
export type PresidentCliArgs = Parameters<typeof presidentApi.exec>;

/** Parses a single CLI command line for the President game. */
export function parsePresidentCommand(input: string): CliParseResult<PresidentCliArgs> {
  const parts = input.trim().toLowerCase().split(/\s+/).filter(Boolean);
  const cmd = parts[0] ?? '';
  if (cmd === 'reset' || cmd === 'r') return { args: ['reset'] };
  if (cmd === 'p' || cmd === 'play') {
    const indices = parts.slice(1).map((s) => Number.parseInt(s, 10));
    if (indices.some((n) => Number.isNaN(n))) {
      return { error: 'Usage: p [idx ...]' };
    }
    return { args: ['play', indices] };
  }
  if (cmd === 'log' || cmd === 'l') return { args: ['log'] };
  if (cmd === 'hint' || cmd === 'h') return { args: ['hint'] };
  return { error: `Unknown command: ${cmd}` };
}

/** Renders a President game state as CLI-friendly text. */
export function formatPresidentState(s: PresidentResponse): string {
  const lines: string[] = [];
  lines.push(`Turn: ${s.gameEndFlag ? 'End' : `Player ${s.currentTurn}`}`);
  if (s.revolutionActive) lines.push('*** REVOLUTION ACTIVE ***');
  for (const p of s.players) {
    const tag = p.isHuman ? 'You' : `CPU${p.id}`;
    const status = p.isFinished ? `rank=${p.rank}` : `${p.cardCount} cards`;
    lines.push(`${tag}: ${status}`);
  }
  if (s.tableCards.length > 0) {
    lines.push(`Table: ${s.tableCards.map((c) => `${c.value}${c.design}`).join(' ')}`);
  }
  if (s.message) lines.push(s.message);
  return lines.join('\n');
}

/** Help text shown in the CLI terminal for President. */
export const PRESIDENT_HELP = [
  'p [idx ...] - Play cards at indices (no index = pass)',
  'r/reset    - Reset game',
  'l/log      - Show action log',
  'h/hint     - Get a hint',
] as const;
