import { renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { gameRoutes } from '../constants/gameRoutes';
import { useTutorialProgress } from './useTutorialProgress';

// **ゲーム数はここに数字で書かない。**書けばゲームを 1 本足すたびに無関係な
// テストが赤くなるだけで、何も守っていない (#4652)。
const TOTAL_GAMES = gameRoutes.length;

describe('useTutorialProgress', () => {
  afterEach(() => {
    localStorage.clear();
  });

  it('returns all games with 0 completed by default', () => {
    const { result } = renderHook(() => useTutorialProgress());
    expect(result.current.totalCount).toBe(TOTAL_GAMES);
    expect(result.current.completedCount).toBe(0);
    expect(result.current.games.every((g) => !g.completed)).toBe(true);
  });

  it('detects completed tutorials from localStorage', () => {
    localStorage.setItem('tutorial_completed_blackjack', 'true');
    localStorage.setItem('tutorial_completed_poker', 'true');
    const { result } = renderHook(() => useTutorialProgress());
    expect(result.current.completedCount).toBe(2);
    expect(result.current.games.find((g) => g.gameName === 'blackjack')?.completed).toBe(true);
    expect(result.current.games.find((g) => g.gameName === 'poker')?.completed).toBe(true);
    expect(result.current.games.find((g) => g.gameName === 'hearts')?.completed).toBe(false);
  });

  it('maps root path to blackjack', () => {
    const { result } = renderHook(() => useTutorialProgress());
    const bj = result.current.games.find((g) => g.gameName === 'blackjack');
    expect(bj).toBeDefined();
    expect(bj?.path).toBe('/');
  });

  it('maps non-root paths by removing leading slash', () => {
    const { result } = renderHook(() => useTutorialProgress());
    const poker = result.current.games.find((g) => g.gameName === 'poker');
    expect(poker).toBeDefined();
    expect(poker?.path).toBe('/poker');
  });

  it('includes all game routes with labelKey', () => {
    const { result } = renderHook(() => useTutorialProgress());
    for (const game of result.current.games) {
      expect(game.labelKey).toBeTruthy();
      expect(game.gameName).toBeTruthy();
    }
  });
});

// Guards the guard: if gameRoutes ever stops being the flattened game list,
// TOTAL_GAMES silently becomes a category count and the assertions above go
// vacuous. A floor is enough — it must not be a handful of categories.
describe('TOTAL_GAMES derivation', () => {
  it('counts games, not categories', () => {
    expect(TOTAL_GAMES).toBeGreaterThan(200);
  });
});
