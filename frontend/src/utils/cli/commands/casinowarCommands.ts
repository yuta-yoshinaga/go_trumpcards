import type { casinowarApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CasinoWarArgs = Parameters<typeof casinowarApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'surrender', 'war', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Casino War CLI command into API exec arguments. */
export function parseCasinowarCommand(input: string): CliParseResult<CasinoWarArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount>' };
      return { args: ['bet', amount.value] };
    }
    case 'surrender':
      return { args: ['surrender'] };
    case 'war':
      return { args: ['war'] };
    case 'log':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text for Casino War CLI mode. */
export const CASINOWAR_HELP: string[] = [
  'b <amt>      - Place ante and deal initial cards',
  'surrender    - Tie: forfeit half ante and end the round',
  'war          - Tie: place an additional bet equal to ante and re-draw',
  'log          - Show action log',
  'r/reset      - Reset game',
];
