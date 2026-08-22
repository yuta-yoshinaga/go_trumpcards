import { describe, expect, it } from 'vitest';
import { DRAMAHA_HELP, parseDramahaCommand } from './dramahaCommands';

describe('parseDramahaCommand', () => {
  it('parses fold', () => {
    expect(parseDramahaCommand('f')).toEqual({ args: ['fold', undefined] });
  });

  it('parses check', () => {
    expect(parseDramahaCommand('ck')).toEqual({ args: ['check', undefined] });
  });

  it('parses call', () => {
    expect(parseDramahaCommand('c')).toEqual({ args: ['call', undefined] });
  });

  it('parses bet with amount', () => {
    expect(parseDramahaCommand('b 100')).toEqual({ args: ['bet', 100] });
  });

  it('rejects bet without amount', () => {
    const result = parseDramahaCommand('b');
    expect(result).toHaveProperty('error');
  });

  it('parses raise with amount', () => {
    expect(parseDramahaCommand('ra 200')).toEqual({ args: ['raise', 200] });
  });

  it('parses allin', () => {
    expect(parseDramahaCommand('a')).toEqual({ args: ['allin', undefined] });
  });

  it('parses reset', () => {
    expect(parseDramahaCommand('r')).toEqual({ args: ['reset', undefined] });
  });

  it('parses muck and show', () => {
    expect(parseDramahaCommand('mu')).toEqual({ args: ['muck', undefined] });
    expect(parseDramahaCommand('sh')).toEqual({ args: ['show', undefined] });
  });

  it('returns an error for unknown commands', () => {
    const result = parseDramahaCommand('xyz');
    expect(result).toHaveProperty('error');
  });
});

describe('parseDramahaCommand — the draw round', () => {
  it('converts the 1-based card numbers on screen to 0-based indices', () => {
    expect(parseDramahaCommand('d 1 3')).toEqual({ args: ['draw', undefined, { indices: [0, 2] }] });
  });

  it('accepts the long form and keeps the order given', () => {
    expect(parseDramahaCommand('draw 5 2')).toEqual({ args: ['draw', undefined, { indices: [4, 1] }] });
  });

  it('treats a bare d as standing pat', () => {
    expect(parseDramahaCommand('d')).toEqual({ args: ['draw', undefined, { indices: [] }] });
  });

  it('rejects card 0 — the numbers on screen start at 1', () => {
    expect(parseDramahaCommand('d 0')).toHaveProperty('error');
  });

  it('rejects a card number past the five in hand', () => {
    expect(parseDramahaCommand('d 6')).toHaveProperty('error');
  });

  it('rejects a non-numeric card number', () => {
    expect(parseDramahaCommand('d two')).toHaveProperty('error');
  });

  it('rejects the same card named twice', () => {
    expect(parseDramahaCommand('d 2 2')).toHaveProperty('error');
  });

  it('exchanges all five when every card is named', () => {
    expect(parseDramahaCommand('d 1 2 3 4 5')).toEqual({ args: ['draw', undefined, { indices: [0, 1, 2, 3, 4] }] });
  });
});

describe('DRAMAHA_HELP', () => {
  it('documents the draw command it accepts', () => {
    expect(DRAMAHA_HELP.some((line) => line.startsWith('d '))).toBe(true);
  });

  it('advertises no command the parser rejects', () => {
    for (const line of DRAMAHA_HELP) {
      const alias = line.split(/[\s/]/)[0];
      const probe = alias === 'b' || alias === 'ra' ? `${alias} 10` : alias;
      expect(parseDramahaCommand(probe), `help advertises "${alias}"`).not.toHaveProperty('error');
    }
  });
});
