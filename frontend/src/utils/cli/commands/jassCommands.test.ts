import { describe, expect, it } from 'vitest';
import type { CliParseResult } from '../types';
import type { JassCliArgs } from './jassCommands';
import { parseJassCommand } from './jassCommands';

const error = (result: CliParseResult<JassCliArgs>) => ('error' in result ? result.error : undefined);

describe('parseJassCommand', () => {
  it('matches the action arguments dispatched by Jass page controls', () => {
    expect(parseJassCommand('p 0')).toEqual({ args: ['play', undefined, 0] });
    expect(parseJassCommand('play 8')).toEqual({ args: ['play', undefined, 8] });
    expect(parseJassCommand('ct 3')).toEqual({ args: ['calltrump', 3] });
    expect(parseJassCommand('ct 4')).toEqual({ args: ['calltrump', 4] });
    expect(parseJassCommand('calltrump 1')).toEqual({ args: ['calltrump', 1] });
    expect(parseJassCommand('s')).toEqual({ args: ['schieben'] });
    expect(parseJassCommand('schieben')).toEqual({ args: ['schieben'] });
    expect(parseJassCommand('n')).toEqual({ args: ['next'] });
    expect(parseJassCommand('next')).toEqual({ args: ['next'] });
    expect(parseJassCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseJassCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseJassCommand('h')).toEqual({ args: ['hint'] });
    expect(parseJassCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseJassCommand('r')).toEqual({ args: ['reset'] });
    expect(parseJassCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('rejects missing, malformed, and out of range arguments', () => {
    for (const input of ['p', 'p x', 'p -1', 'p 9', 'ct', 'ct x', 'ct 0', 'ct 5']) {
      expect(error(parseJassCommand(input))).toMatch(/^Usage:/);
    }
  });

  it('rejects commands not accepted by the Jass API', () => {
    expect(error(parseJassCommand('bid 1'))).toContain('Unknown command: bid');
    expect(error(parseJassCommand('discard 0'))).toContain('Unknown command: discard');
  });
});
