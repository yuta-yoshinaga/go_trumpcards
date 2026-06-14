import { describe, expect, it } from 'vitest';
import { MUS_HELP, parseMusCommand } from './musCommands';

describe('parseMusCommand', () => {
  it('parses mus aliases', () => {
    expect(parseMusCommand('m')).toEqual({ args: ['mus', { mus: true }] });
    expect(parseMusCommand('mus')).toEqual({ args: ['mus', { mus: true }] });
  });

  it('parses cut aliases', () => {
    expect(parseMusCommand('c')).toEqual({ args: ['mus', { mus: false }] });
    expect(parseMusCommand('cut')).toEqual({ args: ['mus', { mus: false }] });
    expect(parseMusCommand('corte')).toEqual({ args: ['mus', { mus: false }] });
  });

  it('parses discard with valid indices', () => {
    expect(parseMusCommand('d 0 2 3')).toEqual({ args: ['discard', { discardIndices: [0, 2, 3] }] });
    expect(parseMusCommand('discard 1')).toEqual({ args: ['discard', { discardIndices: [1] }] });
  });

  it('drops non-integer and negative discard indices', () => {
    expect(parseMusCommand('d 0 x -1 2')).toEqual({ args: ['discard', { discardIndices: [0, 2] }] });
  });

  it('parses discard with no indices as an empty list', () => {
    expect(parseMusCommand('d')).toEqual({ args: ['discard', { discardIndices: [] }] });
  });

  it('parses paso', () => {
    expect(parseMusCommand('paso')).toEqual({ args: ['bet', { betAction: 0, betAmount: 0 }] });
  });

  it('parses envido aliases with an amount', () => {
    expect(parseMusCommand('e 2')).toEqual({ args: ['bet', { betAction: 1, betAmount: 2 }] });
    expect(parseMusCommand('envido 5')).toEqual({ args: ['bet', { betAction: 1, betAmount: 5 }] });
  });

  it('errors on envido without an amount', () => {
    expect(parseMusCommand('e')).toHaveProperty('error');
  });

  it('errors on a non-numeric envido amount', () => {
    expect(parseMusCommand('envido foo')).toHaveProperty('error');
  });

  it('parses ordago', () => {
    expect(parseMusCommand('ordago')).toEqual({ args: ['bet', { betAction: 2, betAmount: 0 }] });
  });

  it('parses quiero', () => {
    expect(parseMusCommand('quiero')).toEqual({ args: ['bet', { betAction: 3, betAmount: 0 }] });
  });

  it('parses noquiero aliases', () => {
    expect(parseMusCommand('nq')).toEqual({ args: ['bet', { betAction: 4, betAmount: 0 }] });
    expect(parseMusCommand('noquiero')).toEqual({ args: ['bet', { betAction: 4, betAmount: 0 }] });
  });

  it('parses next aliases', () => {
    expect(parseMusCommand('n')).toEqual({ args: ['next'] });
    expect(parseMusCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses hint aliases', () => {
    expect(parseMusCommand('h')).toEqual({ args: ['hint'] });
    expect(parseMusCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset aliases', () => {
    expect(parseMusCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMusCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const result = parseMusCommand('envid');
    expect(result).toHaveProperty('error');
    if ('error' in result) {
      expect(result.error).toContain('Did you mean');
    }
  });

  it('errors with no suggestion for an unrelated command', () => {
    const result = parseMusCommand('zzzzzzzz');
    expect(result).toHaveProperty('error');
    if ('error' in result) {
      expect(result.error).not.toContain('Did you mean');
    }
  });
});

describe('MUS_HELP', () => {
  it('lists the core commands', () => {
    expect(MUS_HELP.length).toBeGreaterThan(0);
    const joined = MUS_HELP.join('\n');
    expect(joined).toContain('mus');
    expect(joined).toContain('ordago');
  });
});
