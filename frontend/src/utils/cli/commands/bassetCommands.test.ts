import { describe, expect, it } from 'vitest';
import { parseBassetCommand } from './bassetCommands';

describe('parseBassetCommand', () => {
  it('parses paroli', () => expect(parseBassetCommand('paroli')).toEqual({ args: ['paroli'] }));
  it('parses a bet', () => expect(parseBassetCommand('b 7 10')).toEqual({ args: ['bet', { rank: 7, amount: 10 }] }));
});
