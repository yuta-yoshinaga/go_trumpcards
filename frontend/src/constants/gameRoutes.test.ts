import { describe, expect, it } from 'vitest';
import { AXES, AXIS_KEYS, PROFILE_MAX } from './discoverAxes';
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

describe('gameRoutes profile (concierge SSoT)', () => {
  it('every route has a profile object', () => {
    for (const route of gameRoutes) {
      expect(route.profile).toBeDefined();
      expect(typeof route.profile).toBe('object');
    }
  });

  it('each axis vector length matches AXES[axis].profileLength', () => {
    for (const route of gameRoutes) {
      for (const axis of AXIS_KEYS) {
        expect(route.profile[axis].length).toBe(AXES[axis].profileLength);
      }
    }
  });

  it('every profile value is an integer in 0..PROFILE_MAX', () => {
    for (const route of gameRoutes) {
      for (const axis of AXIS_KEYS) {
        for (const v of route.profile[axis]) {
          expect(Number.isInteger(v)).toBe(true);
          expect(v).toBeGreaterThanOrEqual(0);
          expect(v).toBeLessThanOrEqual(PROFILE_MAX);
        }
      }
    }
  });

  it('every profile has at least one non-zero value on each axis', () => {
    for (const route of gameRoutes) {
      for (const axis of AXIS_KEYS) {
        const sum = route.profile[axis].reduce((a, b) => a + b, 0);
        expect(sum, `${route.path}.${axis} all-zero`).toBeGreaterThan(0);
      }
    }
  });
});
