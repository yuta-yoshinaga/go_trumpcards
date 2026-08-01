import type { guandanApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by guandanApi.exec. */
export type GuandanCliArgs = Parameters<typeof guandanApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'ps', 'pass', 't', 'tribute', 'n', 'next', 'l', 'log', 'r', 'reset', 'help', '?'];

/** Cards dealt to each seat (sync: `GuandanHandSize`). */
const HAND_MAX = 26;

/**
 * Parses a single CLI command line for the Guandan game into
 * {@link guandanApi}.exec arguments.
 *
 * `play` takes **several** indexes, because a turn is a combination, not a
 * card: pairs, triples, straights, tubes, plates and bombs are all played at
 * once. Which combinations beat which — and that the level cards outrank aces
 * while the heart ones are wild — is settled on the server, where the level
 * and the whole hand are known.
 */
export function parseGuandanCommand(input: string): CliParseResult<GuandanCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length === 0) {
        return { error: 'Usage: p <index> [<index> ...] (e.g. p 0 1 2 for a triple)' };
      }
      const cardIndexes: number[] = [];
      for (const a of args) {
        const v = Number.parseInt(a, 10);
        if (Number.isNaN(v) || v < 0 || v > HAND_MAX) {
          return { error: `Usage: every index is 0-${HAND_MAX}` };
        }
        // **同じ札を 2 回数えられない。**通すと 1 枚からペアが作れてしまう。
        if (cardIndexes.includes(v)) {
          return { error: `Usage: index ${v} was given twice` };
        }
        cardIndexes.push(v);
      }
      return { args: ['play', { cardIndexes }] };
    }
    case 'ps':
    case 'pass':
      return { args: ['pass'] };
    case 't':
    case 'tribute': {
      const cardIndex = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(cardIndex) || cardIndex < 0 || cardIndex > HAND_MAX) {
        return { error: `Usage: t <index 0-${HAND_MAX}>` };
      }
      return { args: ['tribute', { cardIndex }] };
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

/** Help text shown in the CLI terminal for Guandan. */
export const GUANDAN_HELP: string[] = [
  'p <index...>          - Play a combination (e.g. p 0 1 2 for a triple)',
  'ps / pass             - Pass',
  't <index>             - Hand a card back as the return tribute',
  'n / next              - Deal the next hand',
  'l / log               - Show action log',
  'r / reset             - Reset game',
];
