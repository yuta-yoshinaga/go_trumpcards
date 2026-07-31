import type { bostonApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by bostonApi.exec. */
export type BostonCliArgs = Parameters<typeof bostonApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'ps',
  'pass',
  'cp',
  'callpartner',
  'p',
  'play',
  'n',
  'next',
  'l',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Ladder bounds (sync: `BostonBidLevelCount`). Step 0 is a pass. */
const LEVEL_MIN = 1;
const LEVEL_MAX = 15;

/** Suit bounds on the wire (1=Spade … 4=Diamond). */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** Seats at the table (sync: `BostonPlayerCnt`). */
const SEAT_MAX = 3;

/** Cards per hand (sync: `BostonHandSize`). */
const HAND_MAX_INDEX = 12;

/**
 * Parses a single CLI command line for the Boston game into
 * {@link bostonApi}.exec arguments.
 *
 * `bid <step> [suit]` names a **ladder step**, not a trick count — the misère
 * bids sit between the trick bids, so a number of tricks would not identify a
 * bid uniquely. `callpartner <seat>` makes it two against two; `-1` plays alone
 * against three.
 */
export function parseBostonCommand(input: string): CliParseResult<BostonCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const level = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(level) || level < LEVEL_MIN || level > LEVEL_MAX) {
        return { error: `Usage: b <${LEVEL_MIN}-${LEVEL_MAX}> [suit] (ladder step, not a trick count)` };
      }
      if (args.length > 1) {
        const suit = Number.parseInt(args[1] ?? '', 10);
        if (Number.isNaN(suit) || suit < SUIT_MIN || suit > SUIT_MAX) {
          return { error: 'Usage: suit is 1-4 (1=S 2=C 3=H 4=D)' };
        }
        return { args: ['bid', { level, suit }] };
      }
      return { args: ['bid', { level }] };
    }
    case 'ps':
    case 'pass':
      return { args: ['pass'] };
    case 'cp':
    case 'callpartner': {
      const partner = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(partner) || partner < -1 || partner > SEAT_MAX) {
        return { error: `Usage: cp <0-${SEAT_MAX}> or cp -1 to play alone` };
      }
      return { args: ['callpartner', { partner }] };
    }
    case 'p':
    case 'play': {
      const cardIndex = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(cardIndex) || cardIndex < 0 || cardIndex > HAND_MAX_INDEX) {
        return { error: `Usage: p <0-${HAND_MAX_INDEX}>` };
      }
      return { args: ['play', { cardIndex }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
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

/** Help text shown in the CLI terminal for Boston. */
export const BOSTON_HELP: string[] = [
  'b <1-15> [suit]     - Bid by LADDER STEP (miseres sit between the trick bids)',
  'ps / pass           - Pass',
  'cp <seat|-1>        - Call a partner, or -1 to play alone against three',
  'p <0-12>            - Play a hand card',
  'n / next            - Deal the next hand',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
