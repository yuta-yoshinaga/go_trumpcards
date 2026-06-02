import type { tienlenApi } from '../../../api/gameApi';
import type { TienLenResponse } from '../../../types/card';
import { parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by tienlenApi.exec. */
export type TienLenCliArgs = Parameters<typeof tienlenApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'r', 'reset', 'help', '?'];

/** Parses a single CLI command line for the Tien Len game. */
export function parseTienLenCommand(input: string): CliParseResult<TienLenCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length === 0) return { args: ['play'] }; // pass
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: `Invalid indices: ${parsed.error}` };
      return { args: ['play', parsed.values] };
    }
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Renders a Tien Len game state as CLI-friendly text. */
export function formatTienLenState(s: TienLenResponse): string {
  const lines: string[] = [];
  lines.push(`Turn: ${s.gameEndFlag ? 'End' : `Player ${s.currentTurn}`}`);
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

/** Help text shown in the CLI terminal for Tien Len. */
export const TIENLEN_HELP = [
  'p <idx...>  - Play cards at indices (e.g., p 0 2)',
  'p           - Pass',
  'r/reset     - Reset game',
] as const;
