import { describe, expect, it } from 'vitest';
import { BASSET_HELP, parseBassetCommand } from './bassetCommands';

describe('parseBassetCommand', () => {
  it('parses bet with both aliases and numeric arguments', () => {
    expect(parseBassetCommand('b 7 10')).toEqual({ args: ['bet', { rank: 7, amount: 10 }] });
    expect(parseBassetCommand('bet 13 250')).toEqual({ args: ['bet', { rank: 13, amount: 250 }] });
  });

  it('rejects missing, non-numeric, and out-of-range bet arguments', () => {
    for (const input of ['b', 'bet 7', 'b x 10', 'b 7 nope', 'b 0 10', 'b 14 10', 'b 7 0']) {
      expect(parseBassetCommand(input)).toEqual({ error: 'Usage: b <rank 1-13> <amount>' });
    }
  });

  it('parses deal with both aliases', () => {
    expect(parseBassetCommand('d')).toEqual({ args: ['deal'] });
    expect(parseBassetCommand('deal')).toEqual({ args: ['deal'] });
  });

  it('parses take and paroli', () => {
    expect(parseBassetCommand('take')).toEqual({ args: ['take'] });
    expect(parseBassetCommand('paroli')).toEqual({ args: ['paroli'] });
  });

  it('parses next with both aliases', () => {
    expect(parseBassetCommand('n')).toEqual({ args: ['next'] });
    expect(parseBassetCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses log and reset with all aliases', () => {
    expect(parseBassetCommand('log')).toEqual({ args: ['log'] });
    expect(parseBassetCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBassetCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command and reports an unknown command', () => {
    const near = parseBassetCommand('bett');
    expect('error' in near).toBe(true);
    if ('error' in near) expect(near.error).toBe('Unknown command: bett. Did you mean bet?');
    const unknown = parseBassetCommand('xyz');
    expect(unknown).toEqual({ error: 'Unknown command: xyz' });
  });

  it('exposes help for each primary action', () => {
    expect(BASSET_HELP).toEqual([
      'b <rank> <amount> - place a wager',
      'deal - deal banker and player cards',
      'take - receive the current winnings',
      'paroli - leave the wager and increase the payout stage',
      'next - start the next deal',
      'reset - reset the game',
    ]);
  });
});
