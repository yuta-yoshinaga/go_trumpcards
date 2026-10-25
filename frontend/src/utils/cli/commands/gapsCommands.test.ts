import { describe, expect, it } from 'vitest';
import type { GapsMoveZone } from '../../../api/games/gaps';
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

  it('rejects unknown commands and documents the available actions', () => {
    expect(errorOf(parseGapsCommand('log'))).toContain('Unknown command: log');
    expect(GAPS_HELP.join('\n')).toContain('p <idx> <row> <col>');
  });
});
