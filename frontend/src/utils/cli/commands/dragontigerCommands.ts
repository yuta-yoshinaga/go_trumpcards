import type { dragontigerApi } from '../../../api/gameApi';
import { DragonTigerBetType } from '../../../types/phases';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type DragonTigerArgs = Parameters<typeof dragontigerApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'd', 'dragon', 't', 'tiger', 'e', 'tie', 'clear', 'log', 'r', 'reset', 'help', '?'];

/** Maps a bet-target token to the DragonTigerBetType value, or null if unknown. */
function parseBetTarget(token: string | undefined): number | null {
  switch ((token ?? '').toLowerCase()) {
    case 'd':
    case 'dragon':
      return DragonTigerBetType.DRAGON;
    case 't':
    case 'tiger':
      return DragonTigerBetType.TIGER;
    case 'e':
    case 'tie':
      return DragonTigerBetType.TIE;
    default:
      return null;
  }
}

/** Builds a bet command for the given target and amount argument. */
function betArgs(betType: number, amountArg: string | undefined): CliParseResult<DragonTigerArgs> {
  if (amountArg === undefined || amountArg === '') {
    return { error: 'Usage: bet <dragon|tiger|tie> <amount>' };
  }
  const amount = parseIntArg([amountArg], 0);
  if ('error' in amount) return { error: 'Usage: bet <dragon|tiger|tie> <amount>' };
  return { args: ['bet', amount.value, betType] };
}

/** Parse a Dragon Tiger CLI command into API exec arguments. */
export function parseDragonTigerCommand(input: string): CliParseResult<DragonTigerArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const betType = parseBetTarget(args[0]);
      if (betType === null) return { error: 'Usage: bet <dragon|tiger|tie> <amount>' };
      return betArgs(betType, args[1]);
    }
    case 'd':
    case 'dragon':
      return betArgs(DragonTigerBetType.DRAGON, args[0]);
    case 't':
    case 'tiger':
      return betArgs(DragonTigerBetType.TIGER, args[0]);
    case 'e':
    case 'tie':
      return betArgs(DragonTigerBetType.TIE, args[0]);
    case 'clear':
      return { args: ['clear'] };
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

/** Help text for Dragon Tiger CLI mode. */
export const DRAGONTIGER_CLI_HELP: string[] = [
  'bet <dragon|tiger|tie> <amt> - Place a bet (or shortcuts: d/t/e <amt>)',
  'clear                        - Clear the current bet',
  'log                          - Show action log',
  'r/reset                      - Reset game',
];
