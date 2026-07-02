import type { trenteetquaranteApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TrenteEtQuaranteArgs = Parameters<typeof trenteetquaranteApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'n', 'next', 'nr', 'nextround', 'hint', 'log', 'r', 'reset', 'help', '?'];

/** Maps a bet-name token to its numeric bet type (0=Noir, 1=Rouge, 2=Couleur, 3=Inverse). */
const BET_TYPES: Record<string, number> = {
  noir: 0,
  black: 0,
  n: 0,
  rouge: 1,
  red: 1,
  couleur: 2,
  color: 2,
  c: 2,
  inverse: 3,
  i: 3,
};

const BET_USAGE = 'Usage: b <noir|rouge|couleur|inverse> <stake>';

/**
 * Parse a Trente et Quarante CLI command into API exec arguments.
 *
 * `bet` takes a bet name/type and a stake, e.g. `b rouge 100`, `bet 2 50`.
 * The command deals both rows and resolves the round in one step.
 */
export function parseTrenteEtQuaranteCommand(input: string): CliParseResult<TrenteEtQuaranteArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const token = (args[0] ?? '').toLowerCase();
      if (token === '') return { error: BET_USAGE };
      let betType = BET_TYPES[token];
      if (betType === undefined) {
        const asNum = Number(token);
        if (Number.isInteger(asNum) && asNum >= 0 && asNum <= 3) {
          betType = asNum;
        } else {
          return { error: BET_USAGE };
        }
      }
      const stake = parseIntArg(args, 1);
      if ('error' in stake) return { error: BET_USAGE };
      return { args: ['bet', betType, stake.value] };
    }
    case 'n':
    case 'next':
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'hint':
      return { args: ['hint'] };
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

/** Help text for Trente et Quarante CLI mode. */
export const TRENTEETQUARANTE_HELP: string[] = [
  'b <bet> <amt>  - Place a stake and deal (bet: noir|rouge|couleur|inverse)',
  'nr/nextround   - Start the next round',
  'hint           - Show a betting hint',
  'log            - Show action log',
  'r/reset        - Start a fresh game',
];
