import type { zhengApi } from '../../../api/gameApi';
import type { ZhengResponse } from '../../../types/card';
import { parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by zhengApi.exec. */
export type ZhengCliArgs = Parameters<typeof zhengApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'r', 'reset', 'help', '?'];

/** Parses a single CLI command line for the Zheng Shangyou game. */
export function parseZhengCommand(input: string): CliParseResult<ZhengCliArgs> {
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

/** Renders a Zheng Shangyou game state as CLI-friendly text. */
export function formatZhengState(s: ZhengResponse): string {
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

/** Help text shown in the CLI terminal for Zheng Shangyou. */
export const ZHENG_HELP = [
  'p <idx...>  - Play cards at indices (e.g., p 0 2)',
  'p           - Pass (not allowed on a lead)',
  'r/reset     - Reset game',
] as const;
