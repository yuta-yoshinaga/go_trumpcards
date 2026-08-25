import type { dilotiApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type DilotiArgs = Parameters<typeof dilotiApi.exec>;

const VALID_COMMANDS = [
  't',
  'take',
  'd',
  'declare',
  'l2',
  'lay',
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

/** Lowest and highest target score the server accepts (sync: internal/domain/DilotiConfig.go). */
const MIN_TARGET = 21;
const MAX_TARGET = 101;
/** Declaration bounds (sync: internal/domain/DilotiRules.go). */
const MIN_DECL = 2;
const MAX_DECL = 10;

/**
 * Splits capture targets into loose table indices and declaration indices.
 *
 * **A `d` prefix names a declaration.** Without the split, `d0` would be read
 * as table card 0 and the wrong pile would be taken.
 */
function parseTargets(args: string[]): { tableIdxs: number[]; declIdxs: number[] } | { bad: string } {
  const tableIdxs: number[] = [];
  const declIdxs: number[] = [];
  for (const raw of args) {
    const token = raw.trim();
    if (token === '') continue;
    const isDecl = token[0] === 'd' || token[0] === 'D';
    const n = Number.parseInt(isDecl ? token.slice(1) : token, 10);
    // **読めない番号を黙って捨てない。** 捨てると取ったつもりの札が場に残る。
    if (Number.isNaN(n) || n < 0) return { bad: token };
    (isDecl ? declIdxs : tableIdxs).push(n);
  }
  return { tableIdxs, declIdxs };
}

/**
 * Parse a Diloti CLI command into API exec arguments.
 *
 * **The capture targets ride with the card played.** `t 0 1 2 d0` plays hand
 * card 0 and takes table cards 1 and 2 plus declaration 0 in one go; splitting
 * it would allow a board where a card is out but nothing has been taken.
 */
export function parseDilotiCommand(input: string): CliParseResult<DilotiArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 't':
    case 'take': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: t <idx> [table... dDECL...]' };
      const targets = parseTargets(args.slice(1));
      if ('bad' in targets) return { error: `Invalid target: ${targets.bad}` };
      return {
        args: [
          'play',
          {
            handIndex: parsed.value,
            action: 'capture',
            tableIndices: targets.tableIdxs.length > 0 ? targets.tableIdxs : undefined,
            declIndices: targets.declIdxs.length > 0 ? targets.declIdxs : undefined,
          },
        ],
      };
    }
    case 'd':
    case 'declare': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: d <idx> <value> [table... dDECL]' };
      const value = Number.parseInt(args[1] ?? '', 10);
      if (Number.isNaN(value) || value < MIN_DECL || value > MAX_DECL) {
        return { error: `Usage: d <idx> <${MIN_DECL}-${MAX_DECL}> [table... dDECL]` };
      }
      const targets = parseTargets(args.slice(2));
      if ('bad' in targets) return { error: `Invalid target: ${targets.bad}` };
      return {
        args: [
          'play',
          {
            handIndex: parsed.value,
            action: 'declare',
            tableIndices: targets.tableIdxs.length > 0 ? targets.tableIdxs : undefined,
            declIndices: targets.declIdxs.length > 0 ? targets.declIdxs : undefined,
            declValue: value,
          },
        ],
      };
    }
    case 'l2':
    case 'lay': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: l2 <idx>' };
      return { args: ['play', { handIndex: parsed.value, action: 'trail' }] };
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

/** Help text for Diloti CLI mode. */
export const DILOTI_HELP: string[] = [
  't <idx> [table... dDECL...]      - Take, listing table numbers and dN declarations',
  'd <idx> <2-10> [table... dDECL]  - Declare a value',
  'l2/lay <idx>                     - Lay the card on the table',
  'nr/nextround                     - Next round',
  'sd <0-2>                         - Set CPU difficulty (resets game)',
  'st <21-101>                      - Set the target score (resets game)',
  'h/hint                           - Show hint',
  'l/log                            - Show action log',
  'r/reset                          - Reset game',
];
