import { describe, expect, it } from 'vitest';
import { OICHOKABU_HELP, parseOichokabuCommand } from './oichokabuCommands';

describe('parseOichokabuCommand', () => {
  it('parses bet alias and full form', () => {
    expect(parseOichokabuCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseOichokabuCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('rejects bet without amount', () => {
    const r = parseOichokabuCommand('bet');
    expect('error' in r).toBe(true);
  });

  it('parses draw alias and full form', () => {
    expect(parseOichokabuCommand('d')).toEqual({ args: ['draw'] });
    expect(parseOichokabuCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses stand alias and full form', () => {
    expect(parseOichokabuCommand('s')).toEqual({ args: ['stand'] });
    expect(parseOichokabuCommand('stand')).toEqual({ args: ['stand'] });
  });

  it('parses reset and log', () => {
    expect(parseOichokabuCommand('r')).toEqual({ args: ['reset'] });
    expect(parseOichokabuCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseOichokabuCommand('log')).toEqual({ args: ['log'] });
  });

  it('suggests close commands on typos', () => {
    const r = parseOichokabuCommand('drawww');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toMatch(/draw/);
  });

  it('rejects unknown command', () => {
    const r = parseOichokabuCommand('nope');
    expect('error' in r).toBe(true);
  });

  it('exposes help lines', () => {
    expect(OICHOKABU_HELP.length).toBeGreaterThan(0);
  });
});
