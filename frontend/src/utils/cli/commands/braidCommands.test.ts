import { describe, expect, it } from 'vitest';
import { BRAID_HELP, parseBraidCommand } from './braidCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseBraidCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseBraidCommand', () => {
  it('parses the simple commands and their aliases', () => {
    expect(parseBraidCommand('d')).toEqual({ args: ['draw'] });
    expect(parseBraidCommand('draw')).toEqual({ args: ['draw'] });
    expect(parseBraidCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseBraidCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseBraidCommand('u')).toEqual({ args: ['undo'] });
    expect(parseBraidCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBraidCommand('log')).toEqual({ args: ['log'] });
    expect(parseBraidCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses the direction command', () => {
    for (const up of ['dir a', 'dir asc', 'dir up', 'direction a']) {
      expect(parseBraidCommand(up)).toEqual({ args: ['dir', undefined, undefined, undefined, true] });
    }
    for (const down of ['dir d', 'dir desc', 'dir down']) {
      expect(parseBraidCommand(down)).toEqual({ args: ['dir', undefined, undefined, undefined, false] });
    }
  });

  it('rejects a direction it does not recognise', () => {
    expect(errorOf('dir x')).toContain('dir a');
    expect(errorOf('dir')).toContain('dir a');
  });

  it('parses the moves', () => {
    expect(parseBraidCommand('m br f')).toEqual({ args: ['move', { zone: 'braid' }, { zone: 'foundation' }] });
    expect(parseBraidCommand('m fd2 f')).toEqual({
      args: ['move', { zone: 'field', col: 2 }, { zone: 'foundation' }],
    });
    expect(parseBraidCommand('m hp5 f')).toEqual({
      args: ['move', { zone: 'helper', col: 5 }, { zone: 'foundation' }],
    });
    expect(parseBraidCommand('m w f')).toEqual({ args: ['move', { zone: 'waste' }, { zone: 'foundation' }] });
    expect(parseBraidCommand('m w hp3')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'helper', col: 3 }],
    });
  });

  // Only the waste can fill a helper -- every other slot is foundation-only.
  it('refuses to fill a helper from anywhere but the waste', () => {
    expect(errorOf('m br hp0')).toContain('only the waste');
    expect(errorOf('m fd1 hp0')).toContain('only the waste');
    expect(errorOf('m hp1 hp0')).toContain('only the waste');
  });

  it('reports malformed moves', () => {
    expect(errorOf('m')).toContain('Usage');
    expect(errorOf('m br')).toContain('Usage');
    expect(errorOf('m zz f')).toContain('Invalid source');
    expect(errorOf('m fdx f')).toContain('Usage');
    expect(errorOf('m hpx f')).toContain('Usage');
    expect(errorOf('m br zz')).toContain('Invalid target');
    expect(errorOf('m w hpx')).toContain('Usage');
  });

  it('suggests a close command', () => {
    expect(errorOf('drw')).toContain('Did you mean');
  });

  it('reports an unknown command', () => {
    expect(errorOf('zzzzz')).toContain('Unknown command');
  });

  it('documents every move form in the help text', () => {
    expect(BRAID_HELP.some((line) => line.startsWith('dir a'))).toBe(true);
    expect(BRAID_HELP.some((line) => line.startsWith('m br f'))).toBe(true);
    expect(BRAID_HELP.some((line) => line.startsWith('m w hp<i>'))).toBe(true);
  });
});
