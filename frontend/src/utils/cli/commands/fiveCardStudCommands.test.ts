import { describe, expect, it } from 'vitest';
import { parseFiveCardStudCommand } from './fiveCardStudCommands';

describe('parseFiveCardStudCommand', () => {
  it('parses no-argument commands', () => {
    expect(parseFiveCardStudCommand('fold')).toEqual({ args: ['fold', undefined] });
    expect(parseFiveCardStudCommand('check')).toEqual({ args: ['check', undefined] });
    expect(parseFiveCardStudCommand('call')).toEqual({ args: ['call', undefined] });
    expect(parseFiveCardStudCommand('allin')).toEqual({ args: ['allin', undefined] });
    expect(parseFiveCardStudCommand('reset')).toEqual({ args: ['reset', undefined] });
  });

  it('parses the rebuy/addon lifecycle commands', () => {
    expect(parseFiveCardStudCommand('rebuy')).toEqual({ args: ['rebuy', undefined] });
    expect(parseFiveCardStudCommand('skiprebuy')).toEqual({ args: ['skiprebuy', undefined] });
    expect(parseFiveCardStudCommand('addon')).toEqual({ args: ['addon', undefined] });
    expect(parseFiveCardStudCommand('skipaddon')).toEqual({ args: ['skipaddon', undefined] });
    expect(parseFiveCardStudCommand('muck')).toEqual({ args: ['muck', undefined] });
    expect(parseFiveCardStudCommand('show')).toEqual({ args: ['show', undefined] });
  });

  it('parses bet and raise with an amount', () => {
    expect(parseFiveCardStudCommand('bet 50')).toEqual({ args: ['bet', 50] });
    expect(parseFiveCardStudCommand('raise 100')).toEqual({ args: ['raise', 100] });
  });

  it('is case-insensitive for the command token', () => {
    expect(parseFiveCardStudCommand('FOLD')).toEqual({ args: ['fold', undefined] });
    expect(parseFiveCardStudCommand('Bet 20')).toEqual({ args: ['bet', 20] });
  });

  it('errors on an unknown command', () => {
    expect(parseFiveCardStudCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
    // Aliases are not parsed (behavior preserved from the original inline parser).
    expect('error' in parseFiveCardStudCommand('f')).toBe(true);
  });
});
