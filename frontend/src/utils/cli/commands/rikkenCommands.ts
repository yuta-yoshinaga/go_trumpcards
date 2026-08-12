import type { rikkenApi } from '../../../api/gameApi';
import { RikkenContract } from '../../../types/phases';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type RikkenArgs = Parameters<typeof rikkenApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'bid',
  'pass',
  'call',
  'next',
  'giveup',
  'hint',
  'h',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

const BID_USAGE = 'Usage: bid <rik|misere|solo|open>, or pass';
const CALL_USAGE = 'Usage: call <spade|clover|heart|diamond>';

/** Maps a contract token to its ladder value, or null if unknown. */
function parseContract(token: string | undefined): number | null {
  switch ((token ?? '').toLowerCase()) {
    case 'pass':
      return RikkenContract.NONE;
    case 'rik':
      return RikkenContract.RIK;
    case 'misere':
    case 'mis':
      return RikkenContract.MISERE;
    case 'solo':
      return RikkenContract.SOLO;
    case 'open':
    case 'openmisere':
      return RikkenContract.OPEN_MISERE;
    default:
      return null;
  }
}

/** Maps a trump token to its suit value (1..4), or null if unknown. */
function parseSuit(token: string | undefined): number | null {
  switch ((token ?? '').toLowerCase()) {
    case 's':
    case 'spade':
      return 1;
    case 'c':
    case 'club':
    case 'clover':
      return 2;
    case 'h':
    case 'heart':
      return 3;
    case 'd':
    case 'diamond':
      return 4;
    default:
      return null;
  }
}

/** Parse a Rikken CLI command into API exec arguments. */
export function parseRikkenCommand(input: string): CliParseResult<RikkenArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: play <cardIndex>' };
      return { args: ['play', idx.value] };
    }
    case 'bid': {
      const contract = parseContract(args[0]);
      if (contract === null) return { error: BID_USAGE };
      return { args: ['bid', undefined, contract] };
    }
    // **パスは契約 0。** 別経路にせず、同じ値として送ります。
    case 'pass':
      return { args: ['bid', undefined, RikkenContract.NONE] };
    case 'call': {
      const suit = parseSuit(args[0]);
      if (suit === null) return { error: CALL_USAGE };
      return { args: ['call', undefined, undefined, suit] };
    }
    case 'next':
      return { args: ['next'] };
    case 'giveup':
      return { args: ['giveup'] };
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

/** Help text for Rikken CLI mode. */
export const RIKKEN_CLI_HELP: string[] = [
  'play <idx>                   - Play the card at <idx>',
  'bid <rik|misere|solo|open>   - Declare a contract',
  'pass                         - Drop out of the bidding',
  'call <s|c|h|d>               - Choose trump (Rik and Solo only)',
  'next                         - Deal the next round',
  'giveup                       - Resign',
  'hint                         - Show a hint',
  'log                          - Show action log',
  'r/reset                      - Reset game',
];
