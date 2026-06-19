import type { kempsApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by kempsApi.exec. */
export type KempsCliArgs = Parameters<typeof kempsApi.exec>;

const VALID_COMMANDS = [
  's',
  'swap',
  'p',
  'pass',
  'sig',
  'signal',
  'k',
  'kemps',
  'c',
  'counter',
  'n',
  'next',
  'sd',
  'setdifficulty',
  'l',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Parses a single CLI command line for the Kemps game into
 * {@link kempsApi}.exec arguments.
 *
 * `swap <h> <f>` swaps hand card h with field card f (both 0-based). `pass`
 * skips the swap (and declines in the declare window). `signal <n>` sets the
 * human's secret signal type (0=Sound, 1=Blink). `kemps` declares "Kemps!" and
 * `counter <seat>` declares "Counter-Kemps!" against an opponent seat. `next`
 * advances to the next round. `sd <0-2>` resets the game with a new CPU
 * difficulty because config is only accepted on reset.
 */
export function parseKempsCommand(input: string): CliParseResult<KempsCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 's':
    case 'swap': {
      const h = Number.parseInt(args[0] ?? '', 10);
      const f = Number.parseInt(args[1] ?? '', 10);
      if (Number.isNaN(h) || h < 0 || Number.isNaN(f) || f < 0)
        return { error: 'Usage: s <h> <f> (swap hand card h with field card f, both >= 0)' };
      return { args: ['swap', { handIndex: h, fieldIndex: f }] };
    }
    case 'p':
    case 'pass':
      return { args: ['pass'] };
    case 'sig':
    case 'signal': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 0 || n > 1) return { error: 'Usage: sig <0-1> (0=Sound 1=Blink)' };
      return { args: ['signal', { signalType: n }] };
    }
    case 'k':
    case 'kemps':
      return { args: ['kemps'] };
    case 'c':
    case 'counter': {
      const seat = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(seat) || seat < 0) return { error: 'Usage: c <seat> (counter-kemps on opponent seat, >= 0)' };
      return { args: ['counter', { targetSeat: seat }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'sd':
    case 'setdifficulty': {
      const level = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(level) || level < 0 || level > 2) return { error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' };
      return { args: ['reset', { config: { cpuDifficulty: level } }] };
    }
    case 'l':
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

/** Help text shown in the CLI terminal for Kemps. */
export const KEMPS_HELP: string[] = [
  's / swap <h> <f>    - Swap hand card h with field card f',
  'p / pass            - Skip the swap (or decline in the declare window)',
  'sig <0-1>           - Set signal type (0=Sound, 1=Blink)',
  'k / kemps           - Declare Kemps!',
  'c / counter <seat>  - Declare Counter-Kemps! on an opponent seat',
  'n / next            - Next round',
  'sd <0-2>            - Set CPU difficulty (resets game)',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
