import { expect, test } from '@playwright/test';
import { DESKTOP_VIEWPORT_COMPACT, DESKTOP_VIEWPORT_LARGE, PHONE_VIEWPORT } from './desktopViewports';
import { navigateTo } from './helpers';
import { checkAction } from './sweeps/action';
import { checkBoard, PAGES as FIXED_BOARD_PAGES } from './sweeps/board';
import { checkCover } from './sweeps/cover';
import { checkDuplicate } from './sweeps/duplicate';
import { checkFirstAction } from './sweeps/firstAction';
import { checkHand, FOOTER_HAND_PAGES } from './sweeps/hand';
import { checkLabel } from './sweeps/label';
import { allGameRoutes } from './sweeps/routes';

/**
 * Route sweep tests (#11922): consolidates the six per-game-route E2E sweeps into
 * ONE page load per (route, viewport).
 *
 * Check ordering (#11922): checks that did not wait in the original per-spec sweeps run first,
 * immediately after navigation, to preserve their timing. Hand and Board checks run last because
 * they wait for animations; a wait timeout records a soft failure and cannot skip a non-waiting check.
 */

test('route list is complete', () => {
  expect(allGameRoutes.length, 'game route path declarations found').toBeGreaterThanOrEqual(360);
  expect(new Set(allGameRoutes).size, 'duplicate game routes removed').toBe(allGameRoutes.length);
  expect(
    allGameRoutes.every((path) => path.startsWith('/')),
    'every game route starts with /',
  ).toBe(true);
});

test('every strict footer-hand page is a game route', () => {
  const missingPages = FOOTER_HAND_PAGES.filter((path) => !allGameRoutes.includes(path));
  expect(missingPages, `footer-hand pages missing from game routes: ${missingPages.join(', ')}`).toEqual([]);
});

test.describe('phone 375x667', () => {
  test.use({ viewport: PHONE_VIEWPORT });
  for (const path of allGameRoutes) {
    test(`${path} passes the route sweep checks`, async ({ page }) => {
      await navigateTo(page, path);
      await checkFirstAction(page, path, PHONE_VIEWPORT);
      await checkLabel(page, path, PHONE_VIEWPORT);
      await checkDuplicate(page, path, PHONE_VIEWPORT);
      await checkHand(page, path, PHONE_VIEWPORT);
    });
  }
});

test.describe('desktop compact 1024x800', () => {
  test.use({ viewport: DESKTOP_VIEWPORT_COMPACT });
  for (const path of allGameRoutes) {
    test(`${path} passes the route sweep checks`, async ({ page }) => {
      await navigateTo(page, path);
      await checkAction(page, path, DESKTOP_VIEWPORT_COMPACT);
      await checkCover(page, path, DESKTOP_VIEWPORT_COMPACT);
      if (!FIXED_BOARD_PAGES.includes(path as (typeof FIXED_BOARD_PAGES)[number])) {
        await checkBoard(page, path, DESKTOP_VIEWPORT_COMPACT);
      }
    });
  }
});

test.describe('desktop large 1280x800', () => {
  test.use({ viewport: DESKTOP_VIEWPORT_LARGE });
  for (const path of allGameRoutes) {
    test(`${path} passes the route sweep checks`, async ({ page }) => {
      await navigateTo(page, path);
      await checkAction(page, path, DESKTOP_VIEWPORT_LARGE);
      await checkCover(page, path, DESKTOP_VIEWPORT_LARGE);
      await checkLabel(page, path, DESKTOP_VIEWPORT_LARGE);
      await checkDuplicate(page, path, DESKTOP_VIEWPORT_LARGE);
      await checkHand(page, path, DESKTOP_VIEWPORT_LARGE);
      if (!FIXED_BOARD_PAGES.includes(path as (typeof FIXED_BOARD_PAGES)[number])) {
        await checkBoard(page, path, DESKTOP_VIEWPORT_LARGE);
      }
    });
  }
});
