import type { cirullaApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CirullaArgs = Parameters<typeof cirullaApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'nr',
  'nextround',
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

/** Lowest and highest target score the server accepts (sync: internal/domain/CirullaConfig.go). */
const MIN_TARGET = 11;
const MAX_TARGET = 51;

/**
 * Parse a Cirulla CLI command into API exec arguments.
 *
 * **The captured cards ride with the card played.** `p 0 1 2` plays hand card 0
 * and takes table cards 1 and 2 in one go; splitting it would allow a board
 * where a card is out but nothing has been taken.
 */
export function parseCirullaCommand(input: string): CliParseResult<CirullaArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx> [table...]' };
      const captures: number[] = [];
      for (const raw of args.slice(1)) {
        const n = Number.parseInt(raw, 10);
        if (Number.isNaN(n) || n < 0) return { error: 'Usage: p <idx> [table...]' };
        captures.push(n);
      }
      return {
        args: ['play', { handIndex: parsed.value, captureIndices: captures.length > 0 ? captures : undefined }],
      };
    }
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'sd':
    case 'setdifficulty': {
      const level = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(level) || level < 0 || level > 2) return { error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' };
      return { args: ['reset', { config: { cpuDifficulty: level } }] };
    }
    case 'st':
    case 'settarget': {
      const target = Number.parseInt(args[0] ?? '', 10);
      // **断る側も範囲を名指す。** 「11-51」と書かずに断ると、次にどの数字を
      // 打てばよいのか画面のどこにも書かれていない。
      if (Number.isNaN(target) || target < MIN_TARGET || target > MAX_TARGET) {
        return { error: `Usage: st <${MIN_TARGET}-${MAX_TARGET}>` };
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

/** Help text for Cirulla CLI mode. */
export const CIRULLA_HELP: string[] = [
  'p <idx> [table...]               - Play a card, listing table numbers to capture',
  'nr/nextround                     - Next round',
  'sd <0-2>                         - Set CPU difficulty (resets game)',
  'st <11-51>                       - Set the target score (resets game)',
  'h/hint                           - Show hint',
  'l/log                            - Show action log',
  'r/reset                          - Reset game',
];
