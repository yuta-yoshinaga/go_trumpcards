import { describe, expect, it } from 'vitest';
import { parseNarcoticCommand } from './narcoticCommands';

describe('parseNarcoticCommand', () => {
  it('parses draw', () => {
    expect(parseNarcoticCommand('d')).toEqual({ args: ['draw'] });
    expect(parseNarcoticCommand('draw')).toEqual({ args: ['draw'] });
  });

  // **列を取らない。**揃った4枚をまとめて捨てるので、選ぶ余地が無い。
  // クローン元 (Aces Up) は `rm <col>` で、列を要求していた。
  it('parses remove without a column', () => {
    expect(parseNarcoticCommand('rm')).toEqual({ args: ['remove'] });
    expect(parseNarcoticCommand('remove')).toEqual({ args: ['remove'] });
  });

  it('parses redeal', () => {
    expect(parseNarcoticCommand('rd')).toEqual({ args: ['redeal'] });
    expect(parseNarcoticCommand('redeal')).toEqual({ args: ['redeal'] });
  });

  it('parses move with column', () => {
    expect(parseNarcoticCommand('mv 1')).toEqual({ args: ['move', 1] });
    expect(parseNarcoticCommand('move 3')).toEqual({ args: ['move', 3] });
  });

  it('returns error for move without column', () => {
    expect('error' in parseNarcoticCommand('mv')).toBe(true);
  });

  it('parses giveup', () => {
    expect(parseNarcoticCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseNarcoticCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses hint', () => {
    expect(parseNarcoticCommand('h')).toEqual({ args: ['hint'] });
    expect(parseNarcoticCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseNarcoticCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses undo', () => {
    expect(parseNarcoticCommand('u')).toEqual({ args: ['undo'] });
    expect(parseNarcoticCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses reset', () => {
    expect(parseNarcoticCommand('r')).toEqual({ args: ['reset'] });
    expect(parseNarcoticCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parseNarcoticCommand('xyz')).toBe(true);
  });

  it('suggests a close command for typos', () => {
    const result = parseNarcoticCommand('drw');
    expect('error' in result).toBe(true);
  });
});
