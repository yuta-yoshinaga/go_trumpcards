import { describe, expect, it } from 'vitest';
import { parseSakuraCommand, SAKURA_HELP } from './sakuraCommands';

describe('parseSakuraCommand', () => {
  it('parses play and field choice', () => {
    expect(parseSakuraCommand('p 1 3')).toEqual({ args: ['play', { cardIndex: 1, fieldIndex: 3 }] });
    expect(parseSakuraCommand('play 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
  });
  it('parses Sakura-specific round and setting commands', () => {
    expect(parseSakuraCommand('nr')).toEqual({ args: ['next'] });
    expect(parseSakuraCommand('ss 4')).toEqual({ args: ['reset', { config: { seats: 4 } }] });
    expect(parseSakuraCommand('sr 12')).toEqual({ args: ['reset', { config: { rounds: 12 } }] });
  });
  it('parses hint, log, and reset aliases', () => {
    expect(parseSakuraCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSakuraCommand('l')).toEqual({ args: ['log'] });
    expect(parseSakuraCommand('r')).toEqual({ args: ['reset'] });
  });
  it('reports malformed and unknown commands', () => {
    expect(parseSakuraCommand('p')).toEqual({ error: 'Usage: p <handIdx> [fieldIdx]' });
    expect(parseSakuraCommand('wat')).toEqual({ error: 'Unknown command: wat' });
  });
  it('exposes help text', () => expect(SAKURA_HELP.length).toBeGreaterThan(0));
});
