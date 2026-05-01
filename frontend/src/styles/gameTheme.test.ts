import { describe, expect, it } from 'vitest';
import { gameRoutes } from '../constants/gameRoutes';
import { gameTheme } from './gameTheme';

describe('gameTheme', () => {
  it('table games use green-bright theme', () => {
    expect(gameTheme.blackjack.bg).toContain('green-bright');
  });

  it('poker games use green-poker theme', () => {
    expect(gameTheme.holdem.bg).toContain('green-poker');
    expect(gameTheme.omaha.bg).toContain('green-poker');
  });

  it('trick-taking games use blue theme', () => {
    expect(gameTheme.hearts.bg).toContain('blue');
    expect(gameTheme.spades.bg).toContain('blue');
  });

  it('matching/pass games use green theme', () => {
    expect(gameTheme.doubt.bg).toContain('green');
    expect(gameTheme.crazyeights.bg).toContain('green');
  });

  it('solitaire games use casino theme', () => {
    expect(gameTheme.klondike.bg).toContain('casino');
    expect(gameTheme.memory.bg).toContain('casino');
  });

  it('counting/rummy games use blue theme', () => {
    expect(gameTheme.ginrummy.bg).toContain('blue');
    expect(gameTheme.cribbage.bg).toContain('blue');
  });

  it('every entry has bg and footer fields', () => {
    for (const [key, value] of Object.entries(gameTheme)) {
      expect(value.bg, `${key}.bg`).toBeTruthy();
      expect(value.footer, `${key}.footer`).toBeTruthy();
    }
  });

  it('covers every game listed in gameRoutes', () => {
    // Issue #1610: silent fallback on a missing key let new games ship without
    // their own theme. The strict GameKey union catches this at compile time;
    // this test guards against route additions that drift away from the union.
    const themeKeys = new Set(Object.keys(gameTheme));
    const missing = gameRoutes
      .map((r) => (r.path === '/' ? 'blackjack' : r.path.replace(/^\//, '')))
      .filter((key) => !themeKeys.has(key));
    expect(missing).toEqual([]);
  });
});
