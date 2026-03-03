import { describe, expect, it } from 'vitest';
import { gameRoutes } from './gameRoutes';

describe('gameRoutes', () => {
  it('each entry has non-empty path and label', () => {
    for (const route of gameRoutes) {
      expect(route.path.length).toBeGreaterThan(0);
      expect(route.label.length).toBeGreaterThan(0);
    }
  });

  it('paths are unique', () => {
    const paths = gameRoutes.map((r) => r.path);
    expect(new Set(paths).size).toBe(paths.length);
  });

  it('first entry is the home route', () => {
    expect(gameRoutes[0].path).toBe('/');
  });
});
