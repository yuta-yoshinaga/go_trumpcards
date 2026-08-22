import type { dramahaApi } from '../../../api/gameApi';
import { splitCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';
import { BETTING_HELP, parseBettingCommand } from './sharedBettingCommands';

type DramahaArgs = Parameters<typeof dramahaApi.exec>;

/** Hole cards a Dramaha seat holds, and so the highest card number `d` takes. */
const HOLE_CARDS = 5;

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
  'd',
  'draw',
];

const DRAW_USAGE = `Usage: d [card numbers 1-${HOLE_CARDS}]`;

/**
 * Parse a Dramaha CLI command into API exec arguments.
 *
 * `d`/`draw` is handled before the shared betting parser because it is the one
 * command that carries a *list*: the shared parser's extra hook can only hand
 * back a single `amount`. Card numbers are **1-based**, matching what the hand
 * shows on screen and what the backend CUI accepts, and are converted to the
 * 0-based `indices` the web endpoint reads. A bare `d` stands pat.
 */
export function parseDramahaCommand(input: string): CliParseResult<DramahaArgs> {
  const { cmd, args } = splitCommand(input);
  if (cmd === 'd' || cmd === 'draw') {
    const indices = parseDrawIndices(args);
    if ('error' in indices) return { error: indices.error };
    return { args: ['draw', undefined, { indices: indices.value }] };
  }

  const result = parseBettingCommand(input, EXTRA_COMMANDS, (extraCmd) => {
    switch (extraCmd) {
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
  return { args: [result.command as DramahaArgs[0], result.amount] };
}

/** Convert `d`'s 1-based card numbers to 0-based indices, rejecting bad input. */
function parseDrawIndices(args: string[]): { value: number[] } | { error: string } {
  const seen = new Set<number>();
  const indices: number[] = [];
  for (const arg of args) {
    if (!/^\d+$/.test(arg)) return { error: DRAW_USAGE };
    const n = Number(arg);
    if (n < 1 || n > HOLE_CARDS) return { error: DRAW_USAGE };
    if (seen.has(n)) return { error: `Card ${n} is named twice` };
    seen.add(n);
    indices.push(n - 1);
  }
  return { value: indices };
}

/** Help text for Dramaha CLI mode. */
export const DRAMAHA_HELP: string[] = [
  ...BETTING_HELP,
  `d [n...]    - Draw round: exchange cards n (1-${HOLE_CARDS}); bare d stands pat`,
  'rb/rebuy    - Rebuy chips',
  'sr/skiprebuy- Skip rebuy',
  'ao/addon    - Add-on chips',
  'sa/skipaddon- Skip add-on',
  'mu/muck     - Muck hand',
  'sh/show     - Show hand',
];
