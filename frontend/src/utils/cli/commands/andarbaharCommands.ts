import type { andarbaharApi } from '../../../api/gameApi';
import { AndarBaharColumn, AndarBaharSideBand } from '../../../types/phases';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type AndarBaharArgs = Parameters<typeof andarbaharApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'a', 'andar', 'bahar', 'clear', 'hint', 'h', 'log', 'r', 'reset', 'help', '?'];

const BET_USAGE = 'Usage: bet <andar|bahar> <amount> [sideAmount sideBand]';

/** Maps a column token to its {@link AndarBaharColumn} value, or null if unknown. */
function parseColumn(token: string | undefined): number | null {
  switch ((token ?? '').toLowerCase()) {
    case 'a':
    case 'andar':
      return AndarBaharColumn.ANDAR;
    case 'b':
    case 'bahar':
      return AndarBaharColumn.BAHAR;
    default:
      return null;
  }
}

/**
 * Builds a bet command for the given column.
 *
 * **The side bet is only attached when both a stake and a band are given.**
 * Band 0 is a real band ("settles on the first card"), so a lone `sideBand`
 * must not be mistaken for one — the server would reject the round as a
 * zero-stake side bet.
 */
function betArgs(target: number, rest: string[]): CliParseResult<AndarBaharArgs> {
  const amountArg = rest[0];
  if (amountArg === undefined || amountArg === '') return { error: BET_USAGE };
  const amount = parseIntArg([amountArg], 0);
  if ('error' in amount) return { error: BET_USAGE };

  if (rest[1] === undefined || rest[2] === undefined) {
    return { args: ['bet', amount.value, target, 0, AndarBaharSideBand.NONE] };
  }
  const sideAmount = parseIntArg([rest[1]], 0);
  if ('error' in sideAmount) return { error: BET_USAGE };
  const sideBand = parseIntArg([rest[2]], 0);
  if ('error' in sideBand) return { error: BET_USAGE };
  if (sideBand.value < AndarBaharSideBand.FIRST || sideBand.value > AndarBaharSideBand.THIRTYSIX_PLUS) {
    return { error: 'Side bet band must be 0-6' };
  }
  return { args: ['bet', amount.value, target, sideAmount.value, sideBand.value] };
}

/** Parse an Andar Bahar CLI command into API exec arguments. */
export function parseAndarBaharCommand(input: string): CliParseResult<AndarBaharArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const target = parseColumn(args[0]);
      if (target === null) return { error: BET_USAGE };
      return betArgs(target, args.slice(1));
    }
    case 'a':
    case 'andar':
      return betArgs(AndarBaharColumn.ANDAR, args);
    case 'bahar':
      return betArgs(AndarBaharColumn.BAHAR, args);
    case 'clear':
      return { args: ['clear'] };
    case 'h':
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

/** Help text for Andar Bahar CLI mode. */
export const ANDARBAHAR_CLI_HELP: string[] = [
  'bet <andar|bahar> <amt>      - Place a bet (or shortcuts: andar/bahar <amt>)',
  'bet <andar|bahar> <amt> <s> <band> - Add a side bet on the card count (band 0-6)',
  'clear                        - Clear the road history',
  'hint                         - Which column is dealt first',
  'log                          - Show action log',
  'r/reset                      - Next round',
];
