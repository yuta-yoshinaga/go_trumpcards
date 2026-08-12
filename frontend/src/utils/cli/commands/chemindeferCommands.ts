import type { chemindeferApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ChemindeFerArgs = Parameters<typeof chemindeferApi.exec>;

const VALID_COMMANDS = [
  'stake',
  's',
  'bet',
  'b',
  'draw',
  'stand',
  'pass',
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

const STAKE_USAGE = 'Usage: stake <amount>';
const BET_USAGE = 'Usage: bet <amount> (0 passes)';
const SIDE_USAGE = 'Usage: draw|stand [punter|banker]';

/** CLI help text for Chemin de Fer. */
export const CHEMINDEFER_CLI_HELP = [
  'stake <amount>   bank that amount (you must be the banker)',
  'bet <amount>     bet against the bank; 0 passes',
  'draw [p|b]       take a third card; side defaults to whoever is deciding',
  'stand [p|b]      stand pat; side defaults to whoever is deciding',
  'pass             give up the bank',
  'next             deal the next coup',
  'giveup           resign',
  'hint             show a hint',
  'log              show the action log',
  'reset (r)        restart',
];

/**
 * Maps a side token to its draw/stand command pair index, or null if unknown.
 *
 * An omitted side is **not** an error: the page sends the command for whichever
 * side is currently deciding, which the server resolves from the phase. Naming
 * the side is only for players who want to be explicit.
 */
function parseSide(token: string | undefined): 'punter' | 'banker' | null | 'invalid' {
  if (token === undefined || token === '') return null;
  switch (token.toLowerCase()) {
    case 'p':
    case 'punter':
      return 'punter';
    case 'b':
    case 'banker':
      return 'banker';
    default:
      return 'invalid';
  }
}

/** Parse a Chemin de Fer CLI command into API exec arguments. */
export function parseChemindeFerCommand(input: string): CliParseResult<ChemindeFerArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 's':
    case 'stake': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: STAKE_USAGE };
      return { args: ['stake', { stake: amount.value }] };
    }
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      // **0 is a pass, and a legal amount.** It must reach the server as a value.
      if ('error' in amount) return { error: BET_USAGE };
      return { args: ['bet', { amount: amount.value }] };
    }
    case 'draw':
    case 'stand': {
      const side = parseSide(args[0]);
      if (side === 'invalid') return { error: SIDE_USAGE };
      const drawing = cmd === 'draw';
      if (side === 'punter') return { args: [drawing ? 'pd' : 'ps'] };
      if (side === 'banker') return { args: [drawing ? 'bd' : 'bs'] };
      // Side omitted: send the phase-resolved command and let the server pick
      // the side from its own phase. The parser cannot see the game state, and
      // guessing here would resolve the wrong side's decision.
      return { args: [drawing ? 'd' : 'st'] };
    }
    case 'pass':
      return { args: ['pb'] };
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
