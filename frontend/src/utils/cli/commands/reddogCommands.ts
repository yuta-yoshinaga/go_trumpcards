import type { reddogApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type RedDogArgs = Parameters<typeof reddogApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'raise', 's', 'stay', 'log', 'r', 'reset', 'h', 'hint', 'help', '?'];

/** Parse a Red Dog CLI command into API exec arguments. */
export function parseReddogCommand(input: string): CliParseResult<RedDogArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount>' };
      return { args: ['bet', amount.value] };
    }
    case 'raise': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: raise <amount>' };
      return { args: ['raise', amount.value] };
    }
    case 's':
    case 'stay':
      return { args: ['stay'] };
    case 'log':
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

/** Help text for Red Dog CLI mode. */
export const REDDOG_HELP: string[] = [
  'b <amt>      - Place ante and deal cards',
  'raise <amt>  - Raise (up to ante) before drawing third card',
  's/stay       - Stay (no raise) before drawing third card',
  'log          - Show action log',
  'r/reset      - Reset game',
  'h/hint       - Get a hint',
];
