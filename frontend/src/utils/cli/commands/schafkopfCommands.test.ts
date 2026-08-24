import { describe, expect, it } from 'vitest';
import { parseSchafkopfCommand, SCHAFKOPF_HELP } from './schafkopfCommands';

describe('parseSchafkopfCommand', () => {
  it('maps each contract to its own pick payload', () => {
    expect(parseSchafkopfCommand('pick')).toEqual({ args: ['pick', { pick: true, contract: 0 }] });
    expect(parseSchafkopfCommand('wenz')).toEqual({ args: ['pick', { pick: true, contract: 1 }] });
    expect(parseSchafkopfCommand('w')).toEqual({ args: ['pick', { pick: true, contract: 1 }] });
    expect(parseSchafkopfCommand('pass')).toEqual({ args: ['pick', { pick: false }] });
  });

  it('carries the suit a Solo names', () => {
    // Each suit, not just one: a parser that hardcoded a suit would pass a
    // single-value check.
    expect(parseSchafkopfCommand('solo 1')).toEqual({ args: ['pick', { pick: true, contract: 2, soloSuit: 1 }] });
    expect(parseSchafkopfCommand('so 3')).toEqual({ args: ['pick', { pick: true, contract: 2, soloSuit: 3 }] });
    expect(parseSchafkopfCommand('solo 4')).toEqual({ args: ['pick', { pick: true, contract: 2, soloSuit: 4 }] });
  });

  it('rejects a Solo without a real suit instead of defaulting', () => {
    // A default here would silently turn a mistyped Solo into a spade Solo.
    for (const input of ['solo', 'solo 0', 'solo 5', 'solo x']) {
      const result = parseSchafkopfCommand(input);
      expect(result, input).toHaveProperty('error');
    }
  });

  it('parses the remaining phase commands', () => {
    expect(parseSchafkopfCommand('call 2')).toEqual({ args: ['call', { callSuit: 2 }] });
    expect(parseSchafkopfCommand('p 3')).toEqual({ args: ['play', { cardIndex: 3 }] });
    expect(parseSchafkopfCommand('n')).toEqual({ args: ['next'] });
    expect(parseSchafkopfCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseSchafkopfCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSchafkopfCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command', () => {
    expect(parseSchafkopfCommand('wenzz')).toHaveProperty('error', expect.stringContaining('wenz'));
    expect(parseSchafkopfCommand('zzz')).toHaveProperty('error', 'Unknown command: zzz');
  });

  it('does not advertise a command it cannot run', () => {
    // The help block is a promise: every line in it has to parse.
    for (const line of SCHAFKOPF_HELP) {
      const name = line.split(/\s+/)[0].split('/')[0];
      expect(parseSchafkopfCommand(name), name).not.toHaveProperty('error', `Unknown command: ${name}`);
    }
  });
});
