import { describe, expect, it } from 'vitest';
import { CASINOHOLDEM_HELP, parseCasinoholdemCommand } from './casinoholdemCommands';

describe('parseCasinoholdemCommand', () => {
  it('parses bet with ante only', () => {
    expect(parseCasinoholdemCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseCasinoholdemCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('parses bet with ante and AA bonus side bet', () => {
    expect(parseCasinoholdemCommand('b 100 10')).toEqual({ args: ['bet', 100, 10] });
    expect(parseCasinoholdemCommand('bet 200 25')).toEqual({ args: ['bet', 200, 25] });
  });

  it('returns error for bet without amount', () => {
    expect('error' in parseCasinoholdemCommand('b')).toBe(true);
    expect('error' in parseCasinoholdemCommand('bet')).toBe(true);
  });

  it('returns error for non-numeric ante', () => {
    expect('error' in parseCasinoholdemCommand('b abc')).toBe(true);
  });

  it('returns error for non-numeric bonus', () => {
    expect('error' in parseCasinoholdemCommand('b 100 xyz')).toBe(true);
  });

  it('parses call', () => {
    expect(parseCasinoholdemCommand('c')).toEqual({ args: ['call'] });
    expect(parseCasinoholdemCommand('call')).toEqual({ args: ['call'] });
  });

  it('parses fold', () => {
    expect(parseCasinoholdemCommand('f')).toEqual({ args: ['fold'] });
    expect(parseCasinoholdemCommand('fold')).toEqual({ args: ['fold'] });
  });

  it('parses log', () => {
    expect(parseCasinoholdemCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseCasinoholdemCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCasinoholdemCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parseCasinoholdemCommand('xyz')).toBe(true);
  });

  it('suggests a similar command for typos', () => {
    const result = parseCasinoholdemCommand('bett 100');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toMatch(/Did you mean/);
    }
  });
});

describe('CASINOHOLDEM_HELP', () => {
  it('lists all primary commands', () => {
    const joined = CASINOHOLDEM_HELP.join(' ');
    expect(joined).toMatch(/b /);
    expect(joined).toMatch(/c\/call/);
    expect(joined).toMatch(/f\/fold/);
    expect(joined).toMatch(/log/);
    expect(joined).toMatch(/r\/reset/);
  });
});
