import type { ecarteApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by ecarteApi.exec. */
export type EcarteCliArgs = Parameters<typeof ecarteApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'pr',
  'propose',
  'st',
  'stand',
  'a',
  'accept',
  'rf',
  'refuse',
  'd',
  'discard',
  'n',
  'next',
  'sd',
  'setdifficulty',
  'tg',
  'settarget',
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
 * Parses a single CLI command line for the Écarté game into
 * {@link ecarteApi}.exec arguments.
 *
 * Écarté begins with an Exchange phase: the elder chooses `propose`/`stand`,
 * the dealer responds `accept`/`refuse`, then both `discard <i j k>` to draw
 * replacements. Once the stock empties, `play <idx>` resolves 5 must-follow
 * tricks; `next` advances to the following deal. `sd <0-2>` and `tg <n>` reset
 * the game with a new CPU difficulty / target score because config is only
 * accepted on reset.
 */
export function parseEcarteCommand(input: string): CliParseResult<EcarteCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(idx)) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: idx }] };
    }
    case 'pr':
    case 'propose':
      return { args: ['propose'] };
    case 'st':
    case 'stand':
      return { args: ['stand'] };
    case 'a':
    case 'accept':
      return { args: ['respond', { accept: true }] };
    case 'rf':
    case 'refuse':
      return { args: ['respond', { accept: false }] };
    case 'd':
    case 'discard': {
      const indices = args.map((a) => Number.parseInt(a, 10)).filter((n) => !Number.isNaN(n));
      return { args: ['discard', { discardIndices: indices }] };
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
    case 'tg':
    case 'settarget': {
      const target = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(target) || target < 1) return { error: 'Usage: tg <n> (target score, >= 1)' };
      return { args: ['reset', { config: { targetScore: target } }] };
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

/** Help text shown in the CLI terminal for Écarté. */
export const ECARTE_HELP: string[] = [
  'pr / propose        - Propose an exchange (elder, ElderDecide)',
  'st / stand          - Stand and play (elder, ElderDecide)',
  'a / accept          - Accept the exchange (dealer, DealerRespond)',
  'rf / refuse         - Refuse the exchange (dealer, DealerRespond)',
  'd <i j k>           - Discard cards and draw replacements',
  'p <idx>             - Play a card (Play phase)',
  'n / next            - Next deal',
  'sd <0-2>            - Set CPU difficulty (resets game)',
  'tg <n>              - Set target score (resets game)',
  'h / hint            - Show hint',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
