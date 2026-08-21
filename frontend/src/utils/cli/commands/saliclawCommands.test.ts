import { describe, expect, it } from 'vitest';
import { parseSalicLawCommand } from './saliclawCommands';

describe('parseSalicLawCommand', () => {
  it('parses draw', () => {
    expect(parseSalicLawCommand('d')).toEqual({ args: ['draw'] });
    expect(parseSalicLawCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses tableau moves', () => {
    expect(parseSalicLawCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
    expect(parseSalicLawCommand('m t0 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 5 }],
    });
  });

  // **捨て札も山札も移動元ではない。**Congress から引き継いだ `m w ...` /
  // `m s ...` をリネームだけして残すと、サーバが 400 で弾くまで分からない
  // 構文が CLI で通ってしまう。
  it.each(['m w f', 'm w t2', 'm s t3', 'm s f'])('rejects the removed syntax %s', (cmd) => {
    const r = parseSalicLawCommand(cmd);
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('Invalid source');
  });

  it('returns error for missing args', () => {
    expect('error' in parseSalicLawCommand('m')).toBe(true);
    expect('error' in parseSalicLawCommand('m t0')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseSalicLawCommand('m x0 f')).toBe(true);
    expect('error' in parseSalicLawCommand('m tz f')).toBe(true);
    expect('error' in parseSalicLawCommand('m t f')).toBe(true);
    expect('error' in parseSalicLawCommand('m t0 z')).toBe(true);
    expect('error' in parseSalicLawCommand('m t0 tz')).toBe(true);
    expect('error' in parseSalicLawCommand('m t0 t')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseSalicLawCommand('u')).toEqual({ args: ['undo'] });
    expect(parseSalicLawCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSalicLawCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseSalicLawCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseSalicLawCommand('log')).toEqual({ args: ['log'] });
    expect(parseSalicLawCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseSalicLawCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseSalicLawCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
