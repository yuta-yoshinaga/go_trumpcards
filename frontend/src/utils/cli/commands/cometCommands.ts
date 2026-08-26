import type { cometApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CometArgs = Parameters<typeof cometApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'pass',
  'nr',
  'nextround',
  'sd',
  'setdifficulty',
  'sp',
  'setplayers',
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

/** Seat bounds (sync: internal/domain/CometConfig.go). */
const MIN_PLAYERS = 2;
const MAX_PLAYERS = 5;
/** Target-score bounds (sync: internal/domain/CometConfig.go). */
const MIN_TARGET = 20;
const MAX_TARGET = 200;

/**
 * Parse a Comet CLI command into API exec arguments.
 *
 * **Passing is a real move here.** Sequences stop on ranks that sit in the
 * dead hand or on the removed 8 of diamonds, so a seat with nothing playable
 * has to say so rather than being stuck.
 */
export function parseCometCommand(input: string): CliParseResult<CometArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { handIndex: parsed.value }] };
    }
    case 'pass':
      return { args: ['pass'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'sd':
    case 'setdifficulty': {
      const level = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(level) || level < 0 || level > 2) return { error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' };
      return { args: ['reset', { config: { cpuDifficulty: level } }] };
    }
    case 'sp':
    case 'setplayers': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < MIN_PLAYERS || n > MAX_PLAYERS) {
        return { error: `Usage: sp <${MIN_PLAYERS}-${MAX_PLAYERS}>` };
      }
      return { args: ['reset', { config: { players: n } }] };
    }
    case 'st':
    case 'settarget': {
      const target = Number.parseInt(args[0] ?? '', 10);
      // **断る側も範囲を名指す。** 書かずに断ると、次にどの数字を打てばよいのか
      // 画面のどこにも書かれていない。
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

/** Help text for Comet CLI mode. */
export const COMET_HELP: string[] = [
  'p <idx>                          - Play a card',
  'pass                             - Pass (only with nothing playable)',
  'nr/nextround                     - Next round',
  'sd <0-2>                         - Set CPU difficulty (resets game)',
  'sp <2-5>                         - Set the seat count (resets game)',
  'st <20-200>                      - Set the target score (resets game)',
  'h/hint                           - Show hint',
  'l/log                            - Show action log',
  'r/reset                          - Reset game',
];
