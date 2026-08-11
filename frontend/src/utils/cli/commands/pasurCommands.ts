import type { pasurApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PasurArgs = Parameters<typeof pasurApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Pasur CLI command into API exec arguments (indices are 0-based). */
export function parsePasurCommand(input: string): CliParseResult<PasurArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx> [tableIdx...]' };
      // **場札の指定は 0 個でもよい（場に置く）。** 引数不足ではない。
      const table: number[] = [];
      for (let i = 1; i < args.length; i++) {
        const t = parseIntArg(args, i);
        if ('error' in t) return { error: 'Usage: p <cardIdx> [tableIdx...]' };
        table.push(t.value);
      }
      return { args: ['play', idx.value, undefined, table] as PasurArgs };
    }
    case 'h':
    case 'hint':
      return { args: ['hint'] as PasurArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as PasurArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as PasurArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as PasurArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Pasur CLI mode. */
export const PASUR_HELP: string[] = [
  'p <i> [t...] - Play hand card i, taking table cards t... (omit t to lay it down)',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
