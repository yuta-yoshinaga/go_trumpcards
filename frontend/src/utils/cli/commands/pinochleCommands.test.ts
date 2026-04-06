import { describe, expect, it } from 'vitest';
import { parsePinochleCommand } from './pinochleCommands';

describe('parsePinochleCommand', () => {
  it('parses bid with amount', () => {
    expect(parsePinochleCommand('bid 25')).toEqual({ args: ['bid', undefined, undefined, 25] });
  });

  it('returns error for bid without amount', () => {
    const result = parsePinochleCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('parses pass', () => {
    expect(parsePinochleCommand('pass')).toEqual({ args: ['pass'] });
  });

  it('parses trump with suit', () => {
    expect(parsePinochleCommand('trump spade')).toEqual({ args: ['trump', undefined, undefined, undefined, 1] });
    expect(parsePinochleCommand('trump heart')).toEqual({ args: ['trump', undefined, undefined, undefined, 3] });
  });

  it('returns error for trump without suit', () => {
    const result = parsePinochleCommand('trump');
    expect('error' in result).toBe(true);
  });

  it('returns error for trump with invalid suit', () => {
    const result = parsePinochleCommand('trump invalid');
    expect('error' in result).toBe(true);
  });

  it('parses meld', () => {
    expect(parsePinochleCommand('meld')).toEqual({ args: ['meld'] });
  });

  it('parses log', () => {
    expect(parsePinochleCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses play from shared trick commands', () => {
    expect(parsePinochleCommand('p 3')).toEqual({ args: ['play', 3] });
  });

  it('parses reset', () => {
    expect(parsePinochleCommand('r')).toEqual({ args: ['reset', undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parsePinochleCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
