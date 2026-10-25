import { describe, expect, it } from 'vitest';
import { parseBauernschnapsenCommand } from './bauernschnapsenCommands';

describe('parseBauernschnapsenCommand', () => {
  it('matches the page actions for all commands and aliases', () => {
    expect(parseBauernschnapsenCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
    expect(parseBauernschnapsenCommand('play 0')).toEqual({ args: ['play', undefined, 0] });
    expect(parseBauernschnapsenCommand('c 0')).toEqual({ args: ['contract', 0, undefined, undefined, 1] });
    expect(parseBauernschnapsenCommand('contract 1 3')).toEqual({ args: ['contract', 1, undefined, undefined, 3] });
    expect(parseBauernschnapsenCommand('m 1')).toEqual({ args: ['marriage', undefined, 1] });
    expect(parseBauernschnapsenCommand('marriage 4')).toEqual({ args: ['marriage', undefined, 4] });
    expect(parseBauernschnapsenCommand('n')).toEqual({ args: ['next'] });
    expect(parseBauernschnapsenCommand('next')).toEqual({ args: ['next'] });
    expect(parseBauernschnapsenCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseBauernschnapsenCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseBauernschnapsenCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBauernschnapsenCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseBauernschnapsenCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBauernschnapsenCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('rejects missing, invalid and out-of-range arguments', () => {
    for (const input of ['p', 'p x', 'p -1', 'p 5', 'c', 'c 4', 'c 1', 'c 1 5', 'c 1 x', 'm', 'm -1', 'm 5']) {
      expect(parseBauernschnapsenCommand(input)).toHaveProperty('error');
    }
    expect(parseBauernschnapsenCommand('wat')).toEqual({ error: 'Unknown command: wat' });
  });
});
