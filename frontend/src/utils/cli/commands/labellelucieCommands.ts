import type { labellelucieApi } from '../../../api/gameApi';
import type { LaBelleLucieResponse } from '../../../types/card';
import { LaBelleLuciePhase } from '../../../types/phases';
import { parseIntArg, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type LaBelleLucieArgs = Parameters<typeof labellelucieApi.exec>;
export interface LaBelleLucieCommandState {
  /** Current fan count, which changes after a redeal. */
  fanCount: number;
  /** Current game phase. */
  phase: number;
  /** Whether the page currently permits undo. */
  canUndo: boolean;
}
/**
 * Maps the page state to what the parser checks. Before the state loads, no fan
 * exists and undo is refused (the terminal is not shown then anyway).
 */
export function toLaBelleLucieCommandState(state: LaBelleLucieResponse | null | undefined): LaBelleLucieCommandState {
  if (!state) return { fanCount: 0, phase: LaBelleLuciePhase.GAME_OVER, canUndo: false };
  return { fanCount: state.fans.length, phase: state.phase, canUndo: state.canUndo };
}

const VALID_COMMANDS = [
  'p',
  'play',
  'f',
  'foundation',
  'rd',
  'redeal',
  'ac',
  'autocomplete',
  'u',
  'undo',
  'undo_n',
  'g',
  'giveup',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a La Belle Lucie CLI command into API exec arguments. */
export function parseLaBelleLucieCommand(
  input: string,
  state?: LaBelleLucieCommandState,
): CliParseResult<LaBelleLucieArgs> {
  const { cmd, args } = splitCommand(input);
  const fanCount = state?.fanCount ?? 0;
  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length < 2) return { error: 'Usage: p <from> <to>' };
      const from = parseIntArg(args, 0);
      const to = parseIntArg(args, 1);
      if (
        'error' in from ||
        'error' in to ||
        from.value < 0 ||
        from.value >= fanCount ||
        to.value < 0 ||
        to.value >= fanCount
      ) {
        return { error: 'Usage: p <from> <to> (fan index out of range)' };
      }
      return { args: ['mf', from.value, to.value] };
    }
    case 'f':
    case 'foundation': {
      const from = parseIntArg(args, 0);
      if ('error' in from || from.value < 0 || from.value >= fanCount)
        return { error: 'Usage: f <idx> (fan index out of range)' };
      return { args: ['ff', from.value] };
    }
    case 'rd':
    case 'redeal':
      return { args: ['rd'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['ac'] };
    case 'u':
    case 'undo':
      if (!state || state.phase !== LaBelleLuciePhase.PLAYING || !state.canUndo) {
        return { error: 'Undo is not available.' };
      }
      return { args: ['u'] };
    case 'undo_n': {
      if (!state || state.phase !== LaBelleLuciePhase.PLAYING || !state.canUndo) {
        return { error: 'Undo is not available.' };
      }
      const n = parseIntArg(args, 0);
      if ('error' in n || n.value < 1) return { error: 'Usage: undo_n <count>' };
      return { args: ['undo_n', undefined, undefined, n.value] };
    }
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in La Belle Lucie's CLI terminal. */
export const LABELLELUCIE_HELP = [
  'p <from> <to> - Move a card between fan indices',
  'f <idx>       - Move a fan top to foundation',
  'rd / redeal   - Gather and redeal cards',
  'ac / autocomplete - Move all available cards to foundation',
  'u / undo      - Undo the last move',
  'undo_n <count> - Undo multiple moves',
  'g / giveup    - Give up',
  'h / hint      - Get a hint',
  'r / reset     - Reset game',
];
