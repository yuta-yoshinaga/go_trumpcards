import type { klaberjassApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by klaberjassApi.exec. */
export type KlaberjassCliArgs = Parameters<typeof klaberjassApi.exec>;

const VALID_COMMANDS = [
  'a',
  'accept',
  'c',
  'call',
  'ps',
  'pass',
  'sm',
  'schmeiss',
  'y',
  'yes',
  'no',
  'p',
  'play',
  'n',
  'next',
  'st',
  'settarget',
  'l',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Lowest and highest suit values on the wire (1=Spade … 4=Diamond). */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** Highest hand index — hands hold nine cards. */
const HAND_MAX_INDEX = 8;

/**
 * Parses a single CLI command line for the Klaberjass game into
 * {@link klaberjassApi}.exec arguments.
 *
 * `accept` takes the turn-up suit as trump and `call <1-4>` names any other.
 * `schmeiss` offers to throw the deal in; `y` agrees and `no` refuses — and a
 * refusal makes the *thrower* the maker, so it is not a free out. `play <i>`
 * plays a hand card, `next` deals again, and `st <n>` resets with a new target
 * because config is only accepted on reset.
 */
export function parseKlaberjassCommand(input: string): CliParseResult<KlaberjassCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'a':
    case 'accept':
      return { args: ['accept'] };
    case 'c':
    case 'call': {
      const suit = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(suit) || suit < SUIT_MIN || suit > SUIT_MAX) {
        return { error: 'Usage: c <1-4> (1=S 2=C 3=H 4=D)' };
      }
      return { args: ['call', { suit }] };
    }
    case 'ps':
    case 'pass':
      return { args: ['pass'] };
    case 'sm':
    case 'schmeiss':
      return { args: ['schmeiss'] };
    case 'y':
    case 'yes':
      return { args: ['answerschmeiss', { accept: true }] };
    case 'no':
      return { args: ['answerschmeiss', { accept: false }] };
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
    case 'st':
    case 'settarget': {
      const targetScore = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(targetScore) || targetScore < 100 || targetScore > 1000) {
        return { error: 'Usage: st <100-1000>' };
      }
      return { args: ['reset', { config: { targetScore } }] };
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

/** Help text shown in the CLI terminal for Klaberjass. */
export const KLABERJASS_HELP: string[] = [
  'a / accept          - Take the turn-up suit as trump',
  'c <1-4>             - Name a trump suit (1=S 2=C 3=H 4=D)',
  'ps / pass           - Pass',
  'sm / schmeiss       - Offer to throw the deal in',
  'y / no              - Answer a schmeiss (refusing makes the thrower the maker)',
  'p <0-8>             - Play a hand card',
  'n / next            - Deal the next hand',
  'st <100-1000>       - Set the target score (resets game)',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
