import type { scoponeApi } from '../../../api/gameApi';
import type { ScoponeResponse } from '../../../types/card';
import type { CliParseResult } from '../types';

/** Args tuple accepted by scoponeApi.exec. */
export type ScoponeCliArgs = Parameters<typeof scoponeApi.exec>;

/** Parses a single CLI command line for the Scopone game.
 *
 * Accepted syntax:
 *   r / reset                   reset game
 *   n / next                    start next round
 *   p <hand> [tbl...]           play hand card; capture given table cards
 *                               (empty table list = lay the card on the table)
 *   l / log                     show action log
 */
export function parseScoponeCommand(input: string): CliParseResult<ScoponeCliArgs> {
  const parts = input.trim().toLowerCase().split(/\s+/).filter(Boolean);
  const cmd = parts[0] ?? '';
  if (cmd === 'reset' || cmd === 'r') return { args: ['r'] };
  if (cmd === 'next' || cmd === 'n') return { args: ['n'] };
  if (cmd === 'log' || cmd === 'l') return { args: ['log'] };

  if (cmd === 'p' || cmd === 'play') {
    const rest = parts.slice(1);
    if (rest.length === 0) return { error: 'Usage: p <hand> [tbl...]' };
    const hand = Number.parseInt(rest[0], 10);
    if (Number.isNaN(hand)) return { error: 'Invalid hand index' };
    const tableIndices = rest.slice(1).map((s) => Number.parseInt(s, 10));
    if (tableIndices.some((n) => Number.isNaN(n))) return { error: 'Invalid table index' };
    return { args: ['p', { handIndex: hand, tableIndices }] };
  }

  if (cmd === 'hint' || cmd === 'h') return { args: ['hint'] };

  return { error: `Unknown command: ${cmd}` };
}

/** Renders a Scopone game state as CLI-friendly text. */
export function formatScoponeState(s: ScoponeResponse): string {
  const lines: string[] = [];
  lines.push(`Phase: ${s.phase} / Turn: ${s.gameEndFlag ? 'End' : `Player ${s.currentTurn}`}`);
  for (const p of s.players) {
    const tag = p.isHuman ? `P${p.id} (You)` : `P${p.id}`;
    lines.push(`${tag} team${p.team}: hand=${p.handCount} captured=${p.capturedCount} scopas=${p.scopaCount}`);
  }
  if (s.tableCards.length > 0) {
    lines.push(`Table: ${s.tableCards.map((c) => `${c.value}${c.design}`).join(' ')}`);
  } else {
    lines.push('Table: (empty)');
  }
  lines.push(s.teamScores.map((sc, t) => `Team${t}: ${sc}`).join('  '));
  if (s.message) lines.push(s.message);
  return lines.join('\n');
}

/** Help text shown in the CLI terminal for Scopone. */
export const SCOPONE_HELP = [
  'p <hand> [tbl...]  - Play hand card; capture table cards (empty = lay)',
  'n/next             - Start next round',
  'r/reset            - Reset game',
  'l/log              - Show action log',
  'h/hint             - Get a hint',
] as const;
