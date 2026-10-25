import type { GapsMoveZone, gapsApi } from '../../../api/games/gaps';
import type { GapsResponse } from '../../../types/card';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by gapsApi.exec. */
export type GapsCliArgs = Parameters<typeof gapsApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'u',
  'undo',
  'd',
  'redeal',
  'g',
  'giveup',
  'h',
  'hint',
  'un',
  'undo_n',
  'r',
  'reset',
];

/** Parse Gaps CLI input into arguments accepted by gapsApi.exec. */
export function parseGapsCommand(input: string, state?: GapsResponse | null): CliParseResult<GapsCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      const [idx, row, col] = args.map((arg) => Number(arg));
      if (
        args.length < 3 ||
        !Number.isInteger(idx) ||
        idx < 0 ||
        idx > 51 ||
        !Number.isInteger(row) ||
        row < 0 ||
        row > 3 ||
        !Number.isInteger(col) ||
        col < 0 ||
        col > 12
      ) {
        return { error: 'Usage: p <idx> <row> <col> (idx 0-51, row 0-3, col 0-12)' };
      }
      const source: GapsMoveZone = { zone: 'grid', row: Math.floor(idx / 13), col: idx % 13 };
      if (state && source.col < (state.lockedPrefixLengths[source.row] ?? 0)) {
        return { error: 'Cannot move a locked card.' };
      }
      const target: GapsMoveZone = { zone: 'grid', row, col };
      return { args: ['move', source, target] };
    }
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'd':
    case 'redeal':
      return { args: ['redeal'] };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'un':
    case 'undo_n': {
      const n = Number(args[0]);
      if (!args.length || !Number.isInteger(n) || n < 1) return { error: 'Usage: undo_n <count>' };
      return { args: ['undo_n', undefined, undefined, n] };
    }
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Gaps CLI terminal. */
export const GAPS_HELP = [
  'p <idx> <row> <col> - Move indexed card to an empty cell (idx 0-51, row 0-3, col 0-12)',
  'u / undo            - Undo the last move',
  'd / redeal          - Redeal while keeping locked cards',
  'g / giveup          - Give up the game',
  'h / hint            - Request a hint',
  'un <count>          - Undo multiple moves',
  'r / reset           - Reset game',
];
