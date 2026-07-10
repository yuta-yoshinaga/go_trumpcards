import { describe, expect, it } from 'vitest';
import { parseWizardCommand } from './wizardCommands';

describe('parseWizardCommand', () => {
  it('parses bid with number', () => {
    expect(parseWizardCommand('bid 3')).toEqual({ args: ['bid', 3] });
  });

  it('returns error for bid without number', () => {
    const result = parseWizardCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('parses play with index', () => {
    expect(parseWizardCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
    expect(parseWizardCommand('play 5')).toEqual({ args: ['play', undefined, 5] });
  });

  it('returns error for play without index', () => {
    const result = parseWizardCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next', () => {
    expect(parseWizardCommand('n')).toEqual({ args: ['next', undefined, undefined] });
  });

  it('parses nextround', () => {
    expect(parseWizardCommand('nr')).toEqual({ args: ['nextround', undefined, undefined] });
  });

  it('parses hint', () => {
    expect(parseWizardCommand('h')).toEqual({ args: ['hint', undefined, undefined] });
  });

  it('parses reset', () => {
    expect(parseWizardCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseWizardCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
