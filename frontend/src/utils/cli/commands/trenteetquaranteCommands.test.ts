import { describe, expect, it } from 'vitest';
import { parseTrenteEtQuaranteCommand, TRENTEETQUARANTE_HELP } from './trenteetquaranteCommands';

describe('parseTrenteEtQuaranteCommand', () => {
  it('parses bet with a named bet type', () => {
    expect(parseTrenteEtQuaranteCommand('b noir 100')).toEqual({ args: ['bet', 0, 100] });
    expect(parseTrenteEtQuaranteCommand('bet rouge 50')).toEqual({ args: ['bet', 1, 50] });
    expect(parseTrenteEtQuaranteCommand('bet couleur 20')).toEqual({ args: ['bet', 2, 20] });
    expect(parseTrenteEtQuaranteCommand('bet inverse 20')).toEqual({ args: ['bet', 3, 20] });
  });

  it('parses bet aliases and numeric bet types', () => {
    expect(parseTrenteEtQuaranteCommand('b red 100')).toEqual({ args: ['bet', 1, 100] });
    expect(parseTrenteEtQuaranteCommand('b 3 100')).toEqual({ args: ['bet', 3, 100] });
  });

  it('rejects bet without a stake', () => {
    expect('error' in parseTrenteEtQuaranteCommand('bet')).toBe(true);
    expect('error' in parseTrenteEtQuaranteCommand('bet noir')).toBe(true);
  });

  it('rejects bet with an invalid bet type', () => {
    expect('error' in parseTrenteEtQuaranteCommand('bet purple 100')).toBe(true);
    expect('error' in parseTrenteEtQuaranteCommand('bet 9 100')).toBe(true);
  });

  it('parses nextround, hint, reset, and log', () => {
    expect(parseTrenteEtQuaranteCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseTrenteEtQuaranteCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseTrenteEtQuaranteCommand('next')).toEqual({ args: ['nextround'] });
    expect(parseTrenteEtQuaranteCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseTrenteEtQuaranteCommand('r')).toEqual({ args: ['reset'] });
    expect(parseTrenteEtQuaranteCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseTrenteEtQuaranteCommand('log')).toEqual({ args: ['log'] });
  });

  it('suggests close commands on typos', () => {
    const r = parseTrenteEtQuaranteCommand('rese');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toMatch(/reset/);
  });

  it('rejects unknown command', () => {
    expect('error' in parseTrenteEtQuaranteCommand('nope')).toBe(true);
  });

  it('exposes help lines', () => {
    expect(TRENTEETQUARANTE_HELP.length).toBeGreaterThan(0);
  });
});
