import type { escobaApi } from '../../../api/gameApi';
import type { EscobaResponse } from '../../../types/card';
import type { CliParseResult } from '../types';

/** Args tuple accepted by escobaApi.exec. */
export type EscobaCliArgs = Parameters<typeof escobaApi.exec>;

/** Parses a single CLI command line for the Escoba game.
 *
 * Accepted syntax:
 *   r / reset                   reset game
 *   n / next                    start next round
 *   p <hand> [tbl...]           play hand card; capture given table cards (sum 15)
 *                               (empty table list = lay the card on the table)
 *   l / log                     show action log
 */
export function parseEscobaCommand(input: string): CliParseResult<EscobaCliArgs> {
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

/** Renders an Escoba game state as CLI-friendly text. */
export function formatEscobaState(s: EscobaResponse): string {
  const lines: string[] = [];
  lines.push(`Phase: ${s.phase} / Turn: ${s.gameEndFlag ? 'End' : `Player ${s.currentTurn}`}`);
  for (const p of s.players) {
    const tag = p.isHuman ? `P${p.id} (You)` : `P${p.id}`;
    lines.push(`${tag}: hand=${p.handCount} captured=${p.capturedCount} escobas=${p.escobaCount} score=${p.score}`);
  }
  if (s.tableCards.length > 0) {
    lines.push(`Table: ${s.tableCards.map((c) => `${c.value}${c.design}`).join(' ')}`);
  } else {
    lines.push('Table: (empty)');
  }
  lines.push(`Stock: ${s.stockRemaining}`);
  if (s.message) lines.push(s.message);
  return lines.join('\n');
}

/** Help text shown in the CLI terminal for Escoba. */
export const ESCOBA_HELP = [
  'p <hand> [tbl...]  - Play hand card; capture table cards summing to 15 (empty = lay)',
  'n/next             - Start next round',
  'r/reset            - Reset game',
  'l/log              - Show action log',
  'h/hint             - Get a hint',
] as const;
