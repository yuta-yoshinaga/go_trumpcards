import type { colourwhistApi } from '../../../api/gameApi';
import { ColourWhistContract } from '../../../types/phases';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ColourWhistArgs = Parameters<typeof colourwhistApi.exec>;

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

const BID_USAGE = 'Usage: bid <samen|alleen|miserie>, or pass (troel is dealt, not bid)';
const CALL_USAGE = 'Usage: call <spade|clover|heart|diamond>';

/**
 * Maps a contract token to its value, or null if unknown.
 *
 * **`troel` is deliberately not accepted.** It is forced at deal time by
 * holding three aces; offering it as a bid would only mislead.
 */
function parseContract(token: string | undefined): number | null {
  switch ((token ?? '').toLowerCase()) {
    case 'pass':
      return ColourWhistContract.NONE;
    case 'samen':
      return ColourWhistContract.SAMEN;
    case 'alleen':
      return ColourWhistContract.ALLEEN;
    case 'miserie':
    case 'mis':
      return ColourWhistContract.MISERIE;
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

/** Parse a Colour Whist CLI command into API exec arguments. */
export function parseColourWhistCommand(input: string): CliParseResult<ColourWhistArgs> {
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
    case 'pass':
      return { args: ['bid', undefined, ColourWhistContract.NONE] };
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

/** Help text for Colour Whist CLI mode. */
export const COLOURWHIST_CLI_HELP: string[] = [
  'play <idx>                   - Play the card at <idx>',
  'bid <samen|alleen|miserie>   - Declare a contract (troel is dealt, not bid)',
  'pass                         - Drop out of the bidding',
  'call <s|c|h|d>               - Choose trump',
  'next                         - Deal the next round',
  'giveup                       - Resign',
  'hint                         - Show a hint',
  'log                          - Show action log',
  'r/reset                      - Reset game',
];
