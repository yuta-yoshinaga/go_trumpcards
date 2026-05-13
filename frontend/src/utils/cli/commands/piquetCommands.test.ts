import { describe, expect, it } from 'vitest';
import { PIQUET_HELP, parsePiquetCommand } from './piquetCommands';

describe('parsePiquetCommand', () => {
  it.each([
    ['r', ['reset']],
    ['reset', ['reset']],
    ['d', ['d']],
    ['declare', ['d']],
    ['nd', ['nd']],
    ['nextdeal', ['nd']],
    ['h', ['h']],
    ['hint', ['h']],
    ['log', ['log']],
  ])('parses %s', (input, expected) => {
    const result = parsePiquetCommand(input);
    if ('error' in result) throw new Error(`unexpected error: ${result.error}`);
    expect(result.args).toEqual(expected);
  });

  it('parses elder exchange with comma-separated indices', () => {
    const result = parsePiquetCommand('e 0,1,2');
    if ('error' in result) throw new Error(result.error);
    expect(result.args).toEqual(['e', undefined, [0, 1, 2]]);
  });

  it('parses younger exchange (alone = pass)', () => {
    const result = parsePiquetCommand('y');
    if ('error' in result) throw new Error(result.error);
    expect(result.args).toEqual(['y', undefined, []]);
  });

  it('parses play with index', () => {
    const result = parsePiquetCommand('p 5');
    if ('error' in result) throw new Error(result.error);
    expect(result.args).toEqual(['p', 5]);
  });

  it('errors on unknown command', () => {
    const result = parsePiquetCommand('zzz');
    expect('error' in result).toBe(true);
  });

  it('errors on play without index', () => {
    const result = parsePiquetCommand('p');
    expect('error' in result).toBe(true);
  });

  it('errors on elder with no args', () => {
    const result = parsePiquetCommand('e');
    expect('error' in result).toBe(true);
  });

  it('errors on bad index', () => {
    const result = parsePiquetCommand('e abc');
    expect('error' in result).toBe(true);
  });

  it('suggests close commands', () => {
    const result = parsePiquetCommand('plays');
    expect('error' in result).toBe(true);
  });

  it('exports help text', () => {
    expect(PIQUET_HELP.length).toBeGreaterThan(0);
  });
});
