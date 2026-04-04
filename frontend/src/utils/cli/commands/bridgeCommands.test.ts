import { describe, expect, it } from 'vitest';
import { parseBridgeCommand } from './bridgeCommands';

describe('parseBridgeCommand', () => {
  it('parses pass', () => {
    expect(parseBridgeCommand('pass')).toEqual({ args: ['bid', undefined, 0] });
  });

  it('parses double', () => {
    expect(parseBridgeCommand('double')).toEqual({ args: ['bid', undefined, 1] });
    expect(parseBridgeCommand('dbl')).toEqual({ args: ['bid', undefined, 1] });
  });

  it('parses redouble', () => {
    expect(parseBridgeCommand('redouble')).toEqual({ args: ['bid', undefined, 2] });
    expect(parseBridgeCommand('rdbl')).toEqual({ args: ['bid', undefined, 2] });
  });

  it('parses log', () => {
    expect(parseBridgeCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses bid with level and suit', () => {
    expect(parseBridgeCommand('bid 1 clubs')).toEqual({ args: ['bid', undefined, 3, 1, 1] });
    expect(parseBridgeCommand('bid 2 hearts')).toEqual({ args: ['bid', undefined, 3, 2, 3] });
    expect(parseBridgeCommand('bid 3 nt')).toEqual({ args: ['bid', undefined, 3, 3, 5] });
  });

  it('returns error for bid without enough args', () => {
    const result = parseBridgeCommand('bid');
    expect('error' in result).toBe(true);
    const result2 = parseBridgeCommand('bid 1');
    expect('error' in result2).toBe(true);
  });

  it('returns error for bid with invalid suit', () => {
    const result = parseBridgeCommand('bid 1 invalid');
    expect('error' in result).toBe(true);
  });

  it('parses play from shared trick commands', () => {
    expect(parseBridgeCommand('p 3')).toEqual({ args: ['play', 3] });
  });

  it('parses next', () => {
    expect(parseBridgeCommand('n')).toEqual({ args: ['next', undefined] });
  });

  it('parses reset', () => {
    expect(parseBridgeCommand('r')).toEqual({ args: ['reset', undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseBridgeCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
