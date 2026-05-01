import { describe, expect, it } from 'vitest';
import { parseNertzCommand } from './nertzCommands';

describe('parseNertzCommand', () => {
  it('parses simple commands', () => {
    expect(parseNertzCommand('d')).toEqual({ args: ['d'] });
    expect(parseNertzCommand('t')).toEqual({ args: ['tick'] });
    expect(parseNertzCommand('nr')).toEqual({ args: ['nr'] });
    expect(parseNertzCommand('u')).toEqual({ args: ['u'] });
    expect(parseNertzCommand('h')).toEqual({ args: ['h'] });
    expect(parseNertzCommand('log')).toEqual({ args: ['log'] });
    expect(parseNertzCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses nertz-to-foundation move', () => {
    expect(parseNertzCommand('m n f')).toEqual({
      args: ['m', { from: { zone: 'nertz', cardIndex: undefined }, to: { zone: 'foundation' } }],
    });
  });

  it('parses waste-to-tableau move', () => {
    expect(parseNertzCommand('m w t2')).toEqual({
      args: ['m', { from: { zone: 'waste', cardIndex: undefined }, to: { zone: 'tableau', col: 2 } }],
    });
  });

  it('parses tableau-to-tableau with index', () => {
    expect(parseNertzCommand('m t0 3 t1')).toEqual({
      args: ['m', { from: { zone: 'tableau', col: 0, cardIndex: 3 }, to: { zone: 'tableau', col: 1 } }],
    });
  });

  it('parses tableau to specific foundation', () => {
    expect(parseNertzCommand('m t1 f2')).toEqual({
      args: ['m', { from: { zone: 'tableau', col: 1, cardIndex: undefined }, to: { zone: 'foundation', idx: 2 } }],
    });
  });

  it('rejects invalid source', () => {
    expect('error' in parseNertzCommand('m x f')).toBe(true);
  });

  it('rejects invalid target', () => {
    expect('error' in parseNertzCommand('m n x')).toBe(true);
  });

  it('rejects missing args', () => {
    expect('error' in parseNertzCommand('m')).toBe(true);
    expect('error' in parseNertzCommand('m n')).toBe(true);
  });

  it('suggests near-match for typos', () => {
    const result = parseNertzCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseNertzCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
