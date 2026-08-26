import type { costlycoloursApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CostlyColoursArgs = Parameters<typeof costlycoloursApi.exec>;

const VALID_COMMANDS = [
  'mog',
  'm',
  'nomog',
  'nm',
  'p',
  'play',
  'nd',
  'nextdeal',
  'sd',
  'setdifficulty',
  'st',
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

/** Target-score bounds (sync: internal/domain/CostlyColoursConfig.go). */
const MIN_TARGET = 31;
const MAX_TARGET = 121;

/**
 * Parse a Costly Colours CLI command into API exec arguments.
 *
 * **Accepting and refusing the exchange are separate verbs.** Splitting on an
 * argument would let a typo turn a refusal into an exchange, and refusing pegs
 * a point for the opponent.
 */
export function parseCostlyColoursCommand(input: string): CliParseResult<CostlyColoursArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'mog':
    case 'm':
      return { args: ['mog', { accept: true }] };
    case 'nomog':
    case 'nm':
      return { args: ['mog', { accept: false }] };
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { handIndex: parsed.value }] };
    }
    case 'nd':
    case 'nextdeal':
      return { args: ['nextdeal'] };
    case 'sd':
    case 'setdifficulty': {
      const level = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(level) || level < 0 || level > 2) return { error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' };
      return { args: ['reset', { config: { cpuDifficulty: level } }] };
    }
    case 'st':
    case 'settarget': {
      const target = Number.parseInt(args[0] ?? '', 10);
      // **断る側も範囲を名指す。** 61 はコットン版、121 はパーレット版。
      if (Number.isNaN(target) || target < MIN_TARGET || target > MAX_TARGET) {
        return { error: `Usage: st <${MIN_TARGET}-${MAX_TARGET}> (61 = Cotton, 121 = Parlett)` };
      }
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

/** Help text for Costly Colours CLI mode. */
export const COSTLYCOLOURS_HELP: string[] = [
  'mog/m                            - Accept the exchange',
  'nomog/nm                         - Refuse it (opponent pegs 1)',
  'p <idx>                          - Play a card',
  'nd/nextdeal                      - Next deal',
  'sd <0-2>                         - Set CPU difficulty (resets game)',
  'st <31-121>                      - Set the target (61 = Cotton, 121 = Parlett)',
  'h/hint                           - Show hint',
  'l/log                            - Show action log',
  'r/reset                          - Reset game',
];
