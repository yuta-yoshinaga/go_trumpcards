import type { pineappleApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import type { CliParseResult } from '../types';
import { BETTING_HELP, parseBettingCommand } from './sharedBettingCommands';

type PineappleArgs = Parameters<typeof pineappleApi.exec>;

const EXTRA_COMMANDS = [
  'rb',
  'rebuy',
  'sr',
  'skiprebuy',
  'ao',
  'addon',
  'sa',
  'skipaddon',
  'mu',
  'muck',
  'sh',
  'show',
  'dis',
  'discard',
];

/** Parse a Pineapple Poker CLI command into API exec arguments. */
export function parsePineappleCommand(input: string): CliParseResult<PineappleArgs> {
  const result = parseBettingCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    switch (cmd) {
      case 'rb':
      case 'rebuy':
        return { command: 'rebuy' };
      case 'sr':
      case 'skiprebuy':
        return { command: 'skiprebuy' };
      case 'ao':
      case 'addon':
        return { command: 'addon' };
      case 'sa':
      case 'skipaddon':
        return { command: 'skipaddon' };
      case 'mu':
      case 'muck':
        return { command: 'muck' };
      case 'sh':
      case 'show':
        return { command: 'show' };
      case 'dis':
      case 'discard': {
        const parsed = parseIntArg(args, 0);
        if ('error' in parsed) return { error: 'Usage: dis <cardIdx>' };
        return { command: 'discard', amount: parsed.value };
      }
      default:
        return null;
    }
  });

  if ('error' in result) return { error: result.error };
  if (result.command === 'discard') {
    return { args: ['discard', undefined, { cardIdx: result.amount }] };
  }
  return { args: [result.command as PineappleArgs[0], result.amount] };
}

/** Help text for Pineapple Poker CLI mode. */
export const PINEAPPLE_HELP: string[] = [
  ...BETTING_HELP,
  'dis <idx>   - Discard a card',
  'rb/rebuy    - Rebuy chips',
  'sr/skiprebuy- Skip rebuy',
  'ao/addon    - Add-on chips',
  'sa/skipaddon- Skip add-on',
  'mu/muck     - Muck hand',
  'sh/show     - Show hand',
];
