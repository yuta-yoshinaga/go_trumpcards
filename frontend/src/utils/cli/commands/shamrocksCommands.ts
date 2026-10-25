import type { shamrocksApi } from '../../../api/gameApi';
import type { ShamrocksResponse } from '../../../types/card';
import { ShamrocksPhase } from '../../../types/phases';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by shamrocksApi.exec. */
export type ShamrocksCliArgs = Parameters<typeof shamrocksApi.exec>;

export const VALID_COMMANDS = [
  'p',
  'play',
  'f',
  'foundation',
  'm',
  'move',
  'mf',
  'ff',
  'ac',
  'autocomplete',
  'u',
  'undo',
  'h',
  'hint',
  'g',
  'giveup',
  'r',
  'reset',
  'help',
  '?',
];

function parseFanToken(tok: string): number | null {
  if (/^fan-?\d+$/i.test(tok)) {
    const num = Number.parseInt(tok.replace(/^fan-?/i, ''), 10);
    return Number.isNaN(num) || num < 0 ? null : num;
  }
  if (/^\d+$/.test(tok)) {
    const num = Number.parseInt(tok, 10);
    return Number.isNaN(num) || num < 0 ? null : num;
  }
  return null;
}

function validateFanMove(from: number, to: number, state?: ShamrocksResponse): CliParseResult<ShamrocksCliArgs> {
  if (state) {
    if (state.phase !== ShamrocksPhase.PLAYING) {
      return { error: 'Game is already over' };
    }
    if (from < 0 || from >= state.fans.length) {
      return { error: `Invalid source fan index: ${from} (valid: 0-${state.fans.length - 1})` };
    }
    if ((state.fans[from]?.length ?? 0) === 0) {
      return { error: `Fan ${from} is empty` };
    }
    if (from === to) {
      return { error: 'Cannot move a fan onto itself' };
    }
    if (to < 0 || to >= state.fans.length) {
      return { error: `Invalid destination fan index: ${to} (valid: 0-${state.fans.length - 1})` };
    }
    if ((state.fans[to]?.length ?? 0) >= 3) {
      return { error: `Fan ${to} is full` };
    }
    const legal = state.legalTargets[`fan-${from}`];
    if (!legal?.fan.includes(to)) {
      return { error: `Illegal move: cannot move fan ${from} to fan ${to}` };
    }
  }
  return { args: ['mf', from, to] };
}

function validateFoundationMove(from: number, state?: ShamrocksResponse): CliParseResult<ShamrocksCliArgs> {
  if (state) {
    if (state.phase !== ShamrocksPhase.PLAYING) {
      return { error: 'Game is already over' };
    }
    if (from < 0 || from >= state.fans.length) {
      return { error: `Invalid source fan index: ${from} (valid: 0-${state.fans.length - 1})` };
    }
    if ((state.fans[from]?.length ?? 0) === 0) {
      return { error: `Fan ${from} is empty` };
    }
    const legal = state.legalTargets[`fan-${from}`];
    if (!legal || legal.foundation.length === 0) {
      return { error: `Card from fan ${from} cannot be moved to foundation` };
    }
  }
  return { args: ['ff', from] };
}

/**
 * Parses a single CLI command line for Shamrocks into {@link shamrocksApi}.exec arguments.
 */
export function parseShamrocksCommand(input: string, state?: ShamrocksResponse): CliParseResult<ShamrocksCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length < 2) {
        return { error: 'Usage: p <from> <to>' };
      }
      const from = parseFanToken(args[0]);
      if (from === null) {
        return { error: 'Invalid source fan: use <index> (e.g., 0, fan-0)' };
      }
      const targetTok = args[1].toLowerCase();
      if (targetTok === 'f' || targetTok === 'foundation') {
        return validateFoundationMove(from, state);
      }
      const to = parseFanToken(targetTok);
      if (to === null) {
        return { error: 'Invalid target: use <fan> or f (foundation)' };
      }
      return validateFanMove(from, to, state);
    }
    case 'f':
    case 'foundation': {
      if (args.length < 1) {
        return { error: 'Usage: f <idx>' };
      }
      const from = parseFanToken(args[0]);
      if (from === null) {
        return { error: 'Usage: f <idx>' };
      }
      return validateFoundationMove(from, state);
    }
    case 'm':
    case 'move': {
      if (args.length < 2) {
        return { error: 'Usage: m <from> <to> or m <from> f' };
      }
      const from = parseFanToken(args[0]);
      if (from === null) {
        return { error: 'Invalid source fan: use <index> (e.g., 0, fan-0)' };
      }
      const targetTok = args[1].toLowerCase();
      if (targetTok === 'f' || targetTok === 'foundation') {
        return validateFoundationMove(from, state);
      }
      const to = parseFanToken(targetTok);
      if (to === null) {
        return { error: 'Invalid target: use <fan> or f (foundation)' };
      }
      return validateFanMove(from, to, state);
    }
    case 'mf': {
      if (args.length < 2) {
        return { error: 'Usage: mf <from> <to>' };
      }
      const from = parseFanToken(args[0]);
      const to = parseFanToken(args[1]);
      if (from === null || to === null) {
        return { error: 'Usage: mf <from> <to>' };
      }
      return validateFanMove(from, to, state);
    }
    case 'ff': {
      if (args.length < 1) {
        return { error: 'Usage: ff <from>' };
      }
      const from = parseFanToken(args[0]);
      if (from === null) {
        return { error: 'Usage: ff <from>' };
      }
      return validateFoundationMove(from, state);
    }
    case 'ac':
    case 'autocomplete': {
      if (state && state.phase !== ShamrocksPhase.PLAYING) {
        return { error: 'Game is already over' };
      }
      return { args: ['ac'] };
    }
    case 'u':
    case 'undo': {
      if (state) {
        if (state.phase !== ShamrocksPhase.PLAYING) {
          return { error: 'Game is already over' };
        }
        if (!state.canUndo) {
          return { error: 'Cannot undo' };
        }
      }
      return { args: ['u'] };
    }
    case 'h':
    case 'hint': {
      if (state && state.phase !== ShamrocksPhase.PLAYING) {
        return { error: 'Game is already over' };
      }
      return { args: ['hint'] };
    }
    case 'g':
    case 'giveup': {
      if (state && state.phase !== ShamrocksPhase.PLAYING) {
        return { error: 'Game is already over' };
      }
      return { args: ['giveup'] };
    }
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the CLI terminal for Shamrocks. */
export const SHAMROCKS_HELP: string[] = [
  'p <from> <to>    - Move top card from fan to fan',
  'f <idx>          - Move top card from fan to foundation',
  'm <from> <to>    - Move top card from fan to fan',
  'm <from> f       - Move top card from fan to foundation',
  'mf <from> <to>   - Move top card from fan to fan',
  'ff <from>        - Move top card from fan to foundation',
  'ac/autocomplete - Auto-complete cards to foundation',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'r/reset         - Reset game',
];
