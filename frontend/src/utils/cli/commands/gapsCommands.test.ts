import { describe, expect, it } from 'vitest';
import type { GapsMoveZone } from '../../../api/games/gaps';
import type { GapsResponse } from '../../../types/card';
import type { CliParseResult } from '../types';
import type { GapsCliArgs } from './gapsCommands';
import { GAPS_HELP, parseGapsCommand } from './gapsCommands';

function errorOf(result: CliParseResult<GapsCliArgs>): string | undefined {
  return 'error' in result ? result.error : undefined;
}

describe('parseGapsCommand', () => {
  it('maps indexed card movement to the same grid zones as the page', () => {
    const source: GapsMoveZone = { zone: 'grid', row: 1, col: 2 };
    const target: GapsMoveZone = { zone: 'grid', row: 0, col: 12 };
    expect(parseGapsCommand('p 15 0 12')).toEqual({ args: ['move', source, target] });
    expect(parseGapsCommand('play 15 0 12')).toEqual({ args: ['move', source, target] });
  });

  it('maps every page action and alias to gapsApi.exec arguments', () => {
    expect(parseGapsCommand('u')).toEqual({ args: ['undo'] });
    expect(parseGapsCommand('undo')).toEqual({ args: ['undo'] });
    expect(parseGapsCommand('d')).toEqual({ args: ['redeal'] });
    expect(parseGapsCommand('redeal')).toEqual({ args: ['redeal'] });
    expect(parseGapsCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseGapsCommand('giveup')).toEqual({ args: ['giveup'] });
    expect(parseGapsCommand('h')).toEqual({ args: ['hint'] });
    expect(parseGapsCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseGapsCommand('un 3')).toEqual({ args: ['undo_n', undefined, undefined, 3] });
    expect(parseGapsCommand('undo_n 3')).toEqual({ args: ['undo_n', undefined, undefined, 3] });
    expect(parseGapsCommand('r')).toEqual({ args: ['reset'] });
    expect(parseGapsCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('rejects missing and out-of-range movement and undo arguments', () => {
    for (const input of ['p', 'p 1', 'p 1 4 0', 'p 52 0 0', 'p 1 0 13', 'play x 0 0']) {
      expect(errorOf(parseGapsCommand(input))).toMatch(/^Usage: p/);
    }
    for (const input of ['un', 'un 0', 'un x']) expect(errorOf(parseGapsCommand(input))).toBe('Usage: undo_n <count>');
  });

  it('rejects an indexed source cell that is empty when state is available', () => {
    const grid = Array.from({ length: 4 }, () => Array(13).fill(null));
    const firstRow = grid[0];
    if (firstRow) firstRow[0] = { design: 'SPADE', value: 2 };
    const state = { grid, lockedPrefixLengths: [1, 0, 0, 0] } as GapsResponse;
    expect(errorOf(parseGapsCommand('p 1 0 4', state))).toBe('No card at idx 1');
    expect(parseGapsCommand('p 0 0 4', state)).toHaveProperty('error');
  });

  it('rejects moving a card within the locked prefix', () => {
    const grid = Array.from({ length: 4 }, () => Array(13).fill(null));
    const firstRow = grid[0];
    if (firstRow) {
      firstRow[1] = { design: 'HEART', value: 3 };
      firstRow[2] = { design: 'SPADE', value: 2 };
    }
    const state = { grid, lockedPrefixLengths: [2, 1, 0, 0] } as GapsResponse;
    expect(errorOf(parseGapsCommand('p 1 0 4', state))).toBe('Cannot move a locked card.');
    expect(parseGapsCommand('p 2 0 4', state)).toEqual({
      args: ['move', { zone: 'grid', row: 0, col: 2 }, { zone: 'grid', row: 0, col: 4 }],
    });
  });

  it('rejects unknown commands and documents the available actions', () => {
    expect(errorOf(parseGapsCommand('log'))).toContain('Unknown command: log');
    expect(GAPS_HELP.join('\n')).toContain('p <idx> <row> <col>');
  });
});
