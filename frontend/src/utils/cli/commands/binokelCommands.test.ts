import { describe, expect, it } from 'vitest';
import { parseBinokelCommand } from './binokelCommands';

describe('parseBinokelCommand', () => {
  it('parses bid with valid amount (150 or more in steps of 10)', () => {
    expect(parseBinokelCommand('bid 150')).toEqual({ args: ['bid', undefined, undefined, 150] });
    expect(parseBinokelCommand('b 200')).toEqual({ args: ['bid', undefined, undefined, 200] });
  });

  it('returns error for bid below 150 or not in steps of 10', () => {
    const resLow = parseBinokelCommand('bid 25');
    expect('error' in resLow).toBe(true);
    const resStep = parseBinokelCommand('bid 155');
    expect('error' in resStep).toBe(true);
  });

  it('returns error for bid without amount', () => {
    const result = parseBinokelCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('parses pass and pa', () => {
    expect(parseBinokelCommand('pass')).toEqual({ args: ['pass'] });
    expect(parseBinokelCommand('pa')).toEqual({ args: ['pass'] });
  });

  it('parses discard with 3 indices', () => {
    expect(parseBinokelCommand('discard 0 1 2')).toEqual({
      args: ['discard', undefined, undefined, undefined, undefined, [0, 1, 2]],
    });
    expect(parseBinokelCommand('d 3 4 5')).toEqual({
      args: ['discard', undefined, undefined, undefined, undefined, [3, 4, 5]],
    });
  });

  it('returns error for discard with fewer than 3 indices', () => {
    const result = parseBinokelCommand('discard 0 1');
    expect('error' in result).toBe(true);
    const resultNone = parseBinokelCommand('d');
    expect('error' in resultNone).toBe(true);
  });

  it('parses trump with suit name or number', () => {
    expect(parseBinokelCommand('trump spade')).toEqual({ args: ['trump', undefined, undefined, undefined, 1] });
    expect(parseBinokelCommand('trump heart')).toEqual({ args: ['trump', undefined, undefined, undefined, 3] });
    expect(parseBinokelCommand('t 2')).toEqual({ args: ['trump', undefined, undefined, undefined, 2] });
  });

  it('returns error for trump without suit', () => {
    const result = parseBinokelCommand('trump');
    expect('error' in result).toBe(true);
  });

  it('returns error for trump with invalid suit', () => {
    const result = parseBinokelCommand('trump invalid');
    expect('error' in result).toBe(true);
  });

  it('parses meld and m', () => {
    expect(parseBinokelCommand('meld')).toEqual({ args: ['meld'] });
    expect(parseBinokelCommand('m')).toEqual({ args: ['meld'] });
  });

  it('parses log and l', () => {
    expect(parseBinokelCommand('log')).toEqual({ args: ['log'] });
    expect(parseBinokelCommand('l')).toEqual({ args: ['log'] });
  });

  it('parses play from shared trick commands', () => {
    expect(parseBinokelCommand('p 3')).toEqual({ args: ['play', 3] });
  });

  it('parses reset', () => {
    expect(parseBinokelCommand('r')).toEqual({ args: ['reset', undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseBinokelCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
