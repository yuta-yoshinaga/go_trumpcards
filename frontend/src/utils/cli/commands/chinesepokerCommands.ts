import type { chinesepokerApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ChinesePokerArgs = Parameters<typeof chinesepokerApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 's', 'set', 'r', 'reset', 'log', 'l', 'h', 'hint'];

/** Parse a CLI input string into Chinese Poker API arguments. */
export function parseChinesepokerCommand(input: string): CliParseResult<ChinesePokerArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount>' };
      return { args: ['bet', amount.value] };
    }
    case 's':
    case 'set': {
      if (args.length < 8) return { error: 'Usage: s <f0 f1 f2 m0 m1 m2 m3 m4>' };
      const fi: number[] = [];
      for (let i = 0; i < 3; i++) {
        const v = parseIntArg(args, i);
        if ('error' in v) return { error: `Invalid front index at position ${i}` };
        fi.push(v.value);
      }
      const mi: number[] = [];
      for (let i = 3; i < 8; i++) {
        const v = parseIntArg(args, i);
        if ('error' in v) return { error: `Invalid middle index at position ${i - 3}` };
        mi.push(v.value);
      }
      return { args: ['set', undefined, fi, mi] };
    }
    case 'log':
    case 'l':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Chinese Poker CLI mode. */
export const CHINESEPOKER_HELP: string[] = [
  'b <amt>                           - Place bet',
  's <f0 f1 f2 m0 m1 m2 m3 m4>      - Set hands (3 front + 5 middle indices)',
  'r                                 - Reset / new game',
  'l                                 - Action log',
  'h/hint       - Get a hint',
];
