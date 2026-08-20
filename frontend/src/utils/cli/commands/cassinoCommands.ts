import type { cassinoApi } from '../../../api/gameApi';
import type { CassinoResponse } from '../../../types/card';
import type { CliParseResult } from '../types';

/** Args tuple accepted by cassinoApi.exec. */
export type CassinoCliArgs = Parameters<typeof cassinoApi.exec>;

/** Parses a single CLI command line for the Cassino game.
 *
 * Accepted syntax:
 *   r / reset                             reset game
 *   n / next                              start next round
 *   t <hand> <tbl...>                     take (sum capture)
 *   t <hand> b <bi...>                    take builds only
 *   t <hand> <tbl...> b <bi...>           take table + builds
 *   b <hand> <value> <tbl...>             build declared value
 *   tr <hand>                             trail
 *   l / log                               show action log
 */
export function parseCassinoCommand(input: string): CliParseResult<CassinoCliArgs> {
  const parts = input.trim().toLowerCase().split(/\s+/).filter(Boolean);
  const cmd = parts[0] ?? '';
  if (cmd === 'reset' || cmd === 'r') return { args: ['reset'] };
  if (cmd === 'next' || cmd === 'n') return { args: ['next'] };
  if (cmd === 'log' || cmd === 'l') return { args: ['log'] };

  if (cmd === 't' || cmd === 'take') {
    const rest = parts.slice(1);
    if (rest.length === 0) return { error: 'Usage: t <hand> <tbl...> [b <bi...>]' };
    const hand = Number.parseInt(rest[0], 10);
    if (Number.isNaN(hand)) return { error: 'Invalid hand index' };
    let tableArgs = rest.slice(1);
    let buildArgs: string[] = [];
    const bIdx = tableArgs.indexOf('b');
    if (bIdx !== -1) {
      buildArgs = tableArgs.slice(bIdx + 1);
      tableArgs = tableArgs.slice(0, bIdx);
    }
    const tableIndices = tableArgs.map((s) => Number.parseInt(s, 10));
    const buildIndices = buildArgs.map((s) => Number.parseInt(s, 10));
    if ([...tableIndices, ...buildIndices].some((n) => Number.isNaN(n))) {
      return { error: 'Invalid index' };
    }
    return { args: ['take', { handIndex: hand, tableIndices, buildIndices }] };
  }

  if (cmd === 'b' || cmd === 'build') {
    const rest = parts.slice(1);
    if (rest.length < 3) return { error: 'Usage: b <hand> <value> <tbl...>' };
    const hand = Number.parseInt(rest[0], 10);
    const value = Number.parseInt(rest[1], 10);
    if (Number.isNaN(hand) || Number.isNaN(value)) return { error: 'Invalid index/value' };
    const tableIndices = rest.slice(2).map((s) => Number.parseInt(s, 10));
    if (tableIndices.some((n) => Number.isNaN(n))) return { error: 'Invalid table index' };
    return { args: ['build', { handIndex: hand, tableIndices, declaredValue: value }] };
  }

  if (cmd === 'tr' || cmd === 'trail') {
    const rest = parts.slice(1);
    if (rest.length < 1) return { error: 'Usage: tr <hand>' };
    const hand = Number.parseInt(rest[0], 10);
    if (Number.isNaN(hand)) return { error: 'Invalid hand index' };
    return { args: ['trail', { handIndex: hand }] };
  }

  if (cmd === 'hint' || cmd === 'h') return { args: ['hint'] };

  return { error: `Unknown command: ${cmd}` };
}

/** Renders a Cassino game state as CLI-friendly text. */
export function formatCassinoState(s: CassinoResponse): string {
  const lines: string[] = [];
  lines.push(`Phase: ${s.phase} / Turn: ${s.gameEndFlag ? 'End' : `Player ${s.currentTurn}`}`);
  for (const p of s.players) {
    const tag = p.isHuman ? 'You' : `CPU${p.id}`;
    lines.push(`${tag}: hand=${p.cardCount} captured=${p.capturedCount} sweeps=${p.sweepCount} score=${p.totalScore}`);
  }
  if (s.tableCards.length > 0) {
    lines.push(`Table: ${s.tableCards.map((c) => `${c.value}${c.design}`).join(' ')}`);
  }
  for (const b of s.builds) {
    const flat = b.groups.map((g) => g.map((c) => `${c.value}${c.design}`).join('+')).join(' | ');
    lines.push(`Build #${b.ownerIdx} val=${b.value} multi=${b.isMulti}: ${flat}`);
  }
  if (s.message) lines.push(s.message);
  return lines.join('\n');
}

/** Help text shown in the CLI terminal for Cassino. */
export const CASSINO_HELP = [
  't <hand> <tbl...> [b <bi...>] - Take table cards (and/or builds)',
  'b <hand> <value> <tbl...>     - Make a build of declared value',
  'tr <hand>                     - Trail (place hand card on table)',
  'n/next                        - Start next round',
  'r/reset                       - Reset game',
  'l/log                         - Show action log',
  'h/hint                        - Get a hint',
] as const;
