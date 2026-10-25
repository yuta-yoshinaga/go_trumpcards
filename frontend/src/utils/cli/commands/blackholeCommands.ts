import type { blackholeApi } from '../../../api/gameApi';
import type { Card } from '../../../types/card';
import type { BlackHoleResponse } from '../../../types/games/blackhole';
import { BlackHolePhase } from '../../../types/phases';
import { parseIntArg, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by blackholeApi.exec. */
export type BlackHoleCliArgs = Parameters<typeof blackholeApi.exec>;

export const VALID_COMMANDS = [
  'p',
  'play',
  'mb',
  'm',
  'move',
  'u',
  'undo',
  'g',
  'giveup',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Returns the ranks (1-13) that the black hole accepts next (adjacent ±1, no A-K wrap).
 */
export function getAcceptableRanks(holeTop: Card | null | undefined): number[] {
  if (!holeTop) return [];
  return [holeTop.value - 1, holeTop.value + 1].filter((v) => v >= 1 && v <= 13);
}

/**
 * Returns the set of fan indices whose top card can be legally moved to the black hole.
 * A fan's top card is playable when its rank is exactly ±1 the black hole's top rank (no A-K wrap).
 */
export function getLegalFans(state: Pick<BlackHoleResponse, 'fans' | 'blackHole'>): Set<number> {
  const holeTop = state.blackHole.length > 0 ? state.blackHole[state.blackHole.length - 1] : null;
  if (!holeTop) return new Set();
  const legal = new Set<number>();
  for (let idx = 0; idx < state.fans.length; idx++) {
    const fan = state.fans[idx];
    if (!fan || fan.length === 0) continue;
    const top = fan[fan.length - 1];
    if (top && Math.abs(top.value - holeTop.value) === 1) {
      legal.add(idx);
    }
  }
  return legal;
}

/**
 * Parses a single CLI command line for Black Hole into blackholeApi.exec arguments.
 */
export function parseBlackHoleCommand(input: string, state?: BlackHoleResponse): CliParseResult<BlackHoleCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play':
    case 'mb':
    case 'm':
    case 'move': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <fan>' };
      const fan = parsed.value;
      if (state) {
        if (state.phase !== BlackHolePhase.PLAYING) {
          return { error: 'Game is not in playing phase' };
        }
        if (fan < 0 || fan >= state.fans.length) {
          return { error: `Fan index out of range (0-${state.fans.length - 1})` };
        }
        if (state.fans[fan].length === 0) {
          return { error: `Fan ${fan} is empty` };
        }
        const legalFans = getLegalFans(state);
        if (!legalFans.has(fan)) {
          return { error: `Fan ${fan} cannot be moved to the black hole` };
        }
      } else if (fan < 0) {
        return { error: 'Invalid fan index' };
      }
      return { args: ['mb', { fan }] };
    }
    case 'u':
    case 'undo': {
      if (state) {
        if (state.phase !== BlackHolePhase.PLAYING) {
          return { error: 'Game is not in playing phase' };
        }
        if (!state.canUndo) {
          return { error: 'No moves to undo' };
        }
      }
      return { args: ['u'] };
    }
    case 'g':
    case 'giveup': {
      if (state && state.phase !== BlackHolePhase.PLAYING) {
        return { error: 'Game is not in playing phase' };
      }
      return { args: ['g'] };
    }
    case 'h':
    case 'hint': {
      if (state && state.phase !== BlackHolePhase.PLAYING) {
        return { error: 'Game is not in playing phase' };
      }
      return { args: ['hint'] };
    }
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the CLI terminal for Black Hole. */
export const BLACKHOLE_HELP: string[] = [
  'p <fan> / mb <fan> - Move top card of fan to the black hole (0-based)',
  'u / undo          - Undo last move',
  'h / hint          - Get a hint',
  'g / giveup        - Give up current game',
  'r / reset         - Reset game',
];
