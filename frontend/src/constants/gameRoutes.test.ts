import { describe, expect, it } from 'bun:test';
import { gameCategories, gameRoutes } from './gameRoutes';

describe('gameRoutes', () => {
  it('each entry has non-empty path and labelKey', () => {
    for (const route of gameRoutes) {
      expect(route.path.length).toBeGreaterThan(0);
      expect(route.labelKey.length).toBeGreaterThan(0);
    }
  });

  it('paths are unique', () => {
    const paths = gameRoutes.map((r) => r.path);
    expect(new Set(paths).size).toBe(paths.length);
  });

  it('first route is the home route', () => {
    expect(gameRoutes[0].path).toBe('/');
  });
});

describe('gameCategories', () => {
  it('each category has non-empty labelKey and at least one route', () => {
    for (const category of gameCategories) {
      expect(category.labelKey.length).toBeGreaterThan(0);
      expect(category.routes.length).toBeGreaterThanOrEqual(1);
    }
  });

  it('category labelKeys are unique', () => {
    const keys = gameCategories.map((c) => c.labelKey);
    expect(new Set(keys).size).toBe(keys.length);
  });

  it('gameRoutes is the flat list of all category routes', () => {
    const flat = gameCategories.flatMap((c) => c.routes);
    expect(gameRoutes).toEqual(flat);
  });
});
