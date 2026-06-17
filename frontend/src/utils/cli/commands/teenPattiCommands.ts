import type { teenPattiApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by teenPattiApi.exec. */
export type TeenPattiCliArgs = Parameters<typeof teenPattiApi.exec>;

const VALID_COMMANDS = [
  's',
  'see',
  'b',
  'bet',
  'rs',
  'raise',
  'f',
  'fold',
  'sh',
  'show',
  'ss',
  'sideshow',
  'ac',
  'accept',
  'dc',
  'decline',
  'n',
  'next',
  'sd',
  'setdifficulty',
  'sa',
  'setante',
  'sc',
  'setchips',
  'h',
  'hint',
  'l',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Parses a single CLI command line for the Teen Patti game into
 * {@link teenPattiApi}.exec arguments.
 *
 * On the player's turn: `see` reveals the hand (Blind→Seen), `bet` calls the
 * stake, `raise <n>` raises the stake to `n`, `fold` drops out, `show` forces a
 * showdown when two players remain, and `sideshow` requests a private hand
 * comparison with the previous Seen player. `accept` / `decline` respond to a
 * Side Show requested of the human. `next` advances to the following deal.
 * `sd <0-2>`, `sa <n>`, and `sc <n>` reset the game with a new CPU difficulty /
 * ante / starting chips because config is only accepted on reset.
 */
export function parseTeenPattiCommand(input: string): CliParseResult<TeenPattiCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 's':
    case 'see':
      return { args: ['see'] };
    case 'b':
    case 'bet':
      return { args: ['bet'] };
    case 'rs':
    case 'raise': {
      const stake = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(stake) || stake < 1) return { error: 'Usage: rs <n> (raise stake, >= 1)' };
      return { args: ['raise', { raiseStake: stake }] };
    }
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'sh':
    case 'show':
      return { args: ['show'] };
    case 'ss':
    case 'sideshow':
      return { args: ['sideshow'] };
    case 'ac':
    case 'accept':
      return { args: ['respond', { accept: true }] };
    case 'dc':
    case 'decline':
      return { args: ['respond', { accept: false }] };
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'sd':
    case 'setdifficulty': {
      const level = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(level) || level < 0 || level > 2) return { error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' };
      return { args: ['reset', { config: { cpuDifficulty: level } }] };
    }
    case 'sa':
    case 'setante': {
      const ante = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(ante) || ante < 1) return { error: 'Usage: sa <n> (ante, >= 1)' };
      return { args: ['reset', { config: { ante } }] };
    }
    case 'sc':
    case 'setchips': {
      const chips = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(chips) || chips < 1) return { error: 'Usage: sc <n> (starting chips, >= 1)' };
      return { args: ['reset', { config: { startingChips: chips } }] };
    }
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text shown in the CLI terminal for Teen Patti. */
export const TEEN_PATTI_HELP: string[] = [
  's / see             - See your hand (Blind -> Seen)',
  'b / bet             - Bet (call the stake)',
  'rs / raise <n>      - Raise the stake to n',
  'f / fold            - Fold (drop out of the deal)',
  'sh / show           - Show (force a showdown, 2 players left)',
  'ss / sideshow       - Request a Side Show with the previous Seen player',
  'ac / accept         - Accept a Side Show request',
  'dc / decline        - Decline a Side Show request',
  'n / next            - Next deal',
  'sd <0-2>            - Set CPU difficulty (resets game)',
  'sa <n>              - Set ante (resets game)',
  'sc <n>              - Set starting chips (resets game)',
  'h / hint            - Show hint',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
