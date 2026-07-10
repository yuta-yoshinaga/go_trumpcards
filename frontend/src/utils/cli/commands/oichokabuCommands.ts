import type { oichokabuApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type OichoKabuArgs = Parameters<typeof oichokabuApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'd', 'draw', 's', 'stand', 'log', 'r', 'reset', 'help', '?'];

/** Parse an Oicho-Kabu CLI command into API exec arguments. */
export function parseOichokabuCommand(input: string): CliParseResult<OichoKabuArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount>' };
      return { args: ['bet', amount.value] };
    }
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 's':
    case 'stand':
      return { args: ['stand'] };
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

/** Help text for Oicho-Kabu CLI mode. */
export const OICHOKABU_HELP: string[] = [
  'b <amt>      - Place your bet and deal two cards each',
  'd/draw       - Draw a third card',
  's/stand      - Stand and settle against the banker',
  'log          - Show action log',
  'r/reset      - Reset game',
];
