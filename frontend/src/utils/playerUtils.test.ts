import { describe, expect, it } from 'bun:test';
import { findPlayerName, playerName } from './playerUtils';

describe('playerName', () => {
  it('returns あなた for human player', () => {
    expect(playerName(0, true)).toBe('あなた');
    expect(playerName(2, true)).toBe('あなた');
  });

  it('returns CPU {id} for non-human player', () => {
    expect(playerName(1, false)).toBe('CPU 1');
    expect(playerName(3, false)).toBe('CPU 3');
  });
});

describe('findPlayerName', () => {
  const players = [
    { id: 0, isHuman: false },
    { id: 1, isHuman: true },
    { id: 2, isHuman: false },
    { id: 3, isHuman: false },
  ];

  it('returns correct name by array index', () => {
    expect(findPlayerName(players, 0)).toBe('CPU 0');
    expect(findPlayerName(players, 1)).toBe('あなた');
    expect(findPlayerName(players, 2)).toBe('CPU 2');
    expect(findPlayerName(players, 3)).toBe('CPU 3');
  });

  it('returns fallback for out-of-range index', () => {
    expect(findPlayerName(players, 5)).toBe('Player 5');
    expect(findPlayerName(players, -1)).toBe('Player -1');
  });

  it('returns fallback for empty players array', () => {
    expect(findPlayerName([], 0)).toBe('Player 0');
  });
});
