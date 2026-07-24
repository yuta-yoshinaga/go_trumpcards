import { describe, expect, it } from 'vitest';
import { parsePigtailCommand, pigtailHelp } from './pigtailCommands';

describe('parsePigtailCommand', () => {
  it('parses draw and reset with their aliases', () => {
    expect(parsePigtailCommand('r')).toEqual({ args: ['reset'] });
    expect(parsePigtailCommand('reset')).toEqual({ args: ['reset'] });
    expect(parsePigtailCommand('d')).toEqual({ args: ['draw'] });
    expect(parsePigtailCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('is case-insensitive on the command word', () => {
    expect(parsePigtailCommand('DRAW')).toEqual({ args: ['draw'] });
  });

  it('suggests the closest command for a near-miss typo', () => {
    expect(parsePigtailCommand('drw')).toEqual({ error: '不明なコマンド: drw。もしかして: draw？' });
    expect(parsePigtailCommand('rest')).toEqual({ error: '不明なコマンド: rest。もしかして: reset？' });
  });

  it('returns a plain unknown-command error when nothing is close', () => {
    expect(parsePigtailCommand('zzzzzz')).toEqual({ error: '不明なコマンド: zzzzzz' });
  });
});

describe('pigtailHelp', () => {
  it('returns a non-empty list of localized help lines', () => {
    const help = pigtailHelp();
    expect(Array.isArray(help)).toBe(true);
    expect(help.length).toBe(2);
    expect(help.some((line) => line.includes('draw'))).toBe(true);
  });
});
