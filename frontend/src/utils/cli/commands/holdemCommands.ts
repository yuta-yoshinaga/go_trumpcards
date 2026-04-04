import type { holdemApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';
import { BETTING_HELP, parseBettingCommand } from './sharedBettingCommands';

type HoldemArgs = Parameters<typeof holdemApi.exec>;

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
  'log',
];

/** Parse a Texas Hold'em CLI command into API exec arguments. */
export function parseHoldemCommand(input: string): CliParseResult<HoldemArgs> {
  const result = parseBettingCommand(input, EXTRA_COMMANDS, (cmd, _args) => {
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
      default:
        return null;
    }
  });

  if ('error' in result) return { error: result.error };
  return { args: [result.command as HoldemArgs[0], result.amount] };
}

/** Help text for Texas Hold'em CLI mode. */
export const HOLDEM_HELP: string[] = [
  ...BETTING_HELP,
  'rb/rebuy    - Rebuy chips',
  'sr/skiprebuy- Skip rebuy',
  'ao/addon    - Add-on chips',
  'sa/skipaddon- Skip add-on',
  'mu/muck     - Muck hand',
  'sh/show     - Show hand',
];
