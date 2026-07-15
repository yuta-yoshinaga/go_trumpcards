import { describe, expect, it } from 'vitest';
import { parseSlapjackCommand, slapjackHelp } from './slapjackCommands';

describe('parseSlapjackCommand', () => {
  it('parses commands and their aliases', () => {
    expect(parseSlapjackCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSlapjackCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseSlapjackCommand('s')).toEqual({ args: ['step'] });
    expect(parseSlapjackCommand('step')).toEqual({ args: ['step'] });
    expect(parseSlapjackCommand('j')).toEqual({ args: ['slap'] });
    expect(parseSlapjackCommand('slap')).toEqual({ args: ['slap'] });
    expect(parseSlapjackCommand('tick')).toEqual({ args: ['tick'] });
    expect(parseSlapjackCommand('l')).toEqual({ args: ['log'] });
    expect(parseSlapjackCommand('log')).toEqual({ args: ['log'] });
  });

  it('is case-insensitive on the command word', () => {
    expect(parseSlapjackCommand('STEP')).toEqual({ args: ['step'] });
  });

  it('suggests the closest command for a near-miss typo', () => {
    expect(parseSlapjackCommand('slp')).toEqual({ error: '不明なコマンド: slp。もしかして: slap？' });
    expect(parseSlapjackCommand('rest')).toEqual({ error: '不明なコマンド: rest。もしかして: reset？' });
  });

  it('returns a plain unknown-command error when nothing is close', () => {
    expect(parseSlapjackCommand('zzzzzz')).toEqual({ error: '不明なコマンド: zzzzzz' });
  });
});

describe('slapjackHelp', () => {
  it('returns a non-empty list of localized help lines', () => {
    const help = slapjackHelp();
    expect(Array.isArray(help)).toBe(true);
    expect(help.length).toBe(5);
    expect(help.some((line) => line.includes('slap'))).toBe(true);
  });
});
