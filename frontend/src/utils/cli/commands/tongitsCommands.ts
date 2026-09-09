import type { tongitsApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TongitsArgs = Parameters<typeof tongitsApi.exec>;
const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'd',
  'dis',
  'discard',
  'm',
  'meld',
  'sp',
  'sapaw',
  'c',
  'challenge',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Tongits CLI command into API exec arguments. */
export function parseTongitsCommand(input: string): CliParseResult<TongitsArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'ds':
    case 'drawstock':
      return { args: ['drawstock'] };
    case 'dd':
    case 'drawdiscard':
      return { args: ['drawdiscard'] };
    case 'd':
    case 'dis':
    case 'discard': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: d <idx>' };
      return { args: ['discard', parsed.value] };
    }
    case 'm':
    case 'meld': {
      if (args.length < 3) return { error: 'Usage: m <idx> [idx ...]' };
      const indices: number[] = [];
      for (let index = 0; index < args.length; index += 1) {
        const parsed = parseIntArg(args, index);
        if ('error' in parsed) return { error: 'Usage: m <idx> [idx ...]' };
        indices.push(parsed.value);
      }
      return { args: ['meld', undefined, undefined, indices] };
    }
    case 'sp':
    case 'sapaw': {
      if (args.length !== 3) return { error: 'Usage: sp <player> <meld> <card>' };
      const values: number[] = [];
      for (let index = 0; index < args.length; index += 1) {
        const parsed = parseIntArg(args, index);
        if ('error' in parsed) return { error: 'Usage: sp <player> <meld> <card>' };
        values.push(parsed.value);
      }
      return { args: ['sapaw', values[2], undefined, undefined, values[0], values[1]] };
    }
    case 'c':
    case 'challenge':
      return { args: ['challenge', undefined, undefined, undefined, undefined, undefined, [true, true]] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'log':
      return { args: ['log'] };
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

/** Help text for Tongits CLI mode. */
export const TONGITS_HELP: string[] = [
  'ds/drawstock   - Draw from stock',
  'dd/drawdiscard - Draw from discard',
  'd/dis <idx>    - Discard a card',
  'm <idx> ...    - Meld selected cards',
  'sp <player> <meld> <card> - Sapaw a card',
  'c/challenge    - Declare a challenge',
  'nr/nextround   - Next round',
  'log            - Show action log',
  'r/reset        - Reset game',
];
