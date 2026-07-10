import type { gostopApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type GoStopArgs = Parameters<typeof gostopApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'g',
  'go',
  'st',
  'stop',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Go-Stop (ゴーストップ) CLI command into API exec arguments. */
export function parseGoStopCommand(input: string): CliParseResult<GoStopArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <handIdx> [fieldIdx]' };
      if (args.length > 1) {
        const field = Number.parseInt(args[1], 10);
        if (Number.isNaN(field)) return { error: `Invalid field index: ${args[1]}` };
        return { args: ['play', { cardIndex: parsed.value, fieldIndex: field }] };
      }
      return { args: ['play', { cardIndex: parsed.value }] };
    }
    case 'g':
    case 'go':
      return { args: ['go'] };
    case 'st':
    case 'stop':
      return { args: ['stop'] };
    case 'n':
    case 'next':
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text for Go-Stop (ゴーストップ) CLI mode. */
export const GOSTOP_HELP: string[] = [
  'p <h> [f]    - Play hand card h, capturing field card f (omit f unless a 2-way match)',
  'g/go         - Call go (continue the round for more points, raising the stakes)',
  'st/stop      - Call stop (bank the current score and end the round)',
  'nr/nextround - Deal the next round',
  'h/hint       - Show hint',
  'r/reset      - Reset game',
];
