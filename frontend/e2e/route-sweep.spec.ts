import { expect, test } from '@playwright/test';
import { DESKTOP_VIEWPORT_COMPACT, DESKTOP_VIEWPORT_LARGE, PHONE_VIEWPORT } from './desktopViewports';
import { navigateTo } from './helpers';
import { checkAction } from './sweeps/action';
import { checkBoard } from './sweeps/board';
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
 * Check ordering:
 * None of the sweep checks mutate DOM or scroll state. They perform read-only layout and tree
 * measurements. In the 375x667 suite, checkFirstAction runs before checkHand to measure the
 * initial actionable state before finite card animations settle. checkHand and checkBoard
 * internally await finite animations (waitForAnimations) before layout measurement.
 * checkAction, checkCover, checkLabel, and checkDuplicate run subsequent read-only DOM and
 * computed-style scans.
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
      await checkHand(page, path, PHONE_VIEWPORT);
      await checkLabel(page, path, PHONE_VIEWPORT);
      await checkDuplicate(page, path, PHONE_VIEWPORT);
    });
  }
});

test.describe('desktop compact 1024x800', () => {
  test.use({ viewport: DESKTOP_VIEWPORT_COMPACT });
  for (const path of allGameRoutes) {
    test(`${path} passes the route sweep checks`, async ({ page }) => {
      await navigateTo(page, path);
      await checkBoard(page, path, DESKTOP_VIEWPORT_COMPACT);
      await checkAction(page, path, DESKTOP_VIEWPORT_COMPACT);
      await checkCover(page, path, DESKTOP_VIEWPORT_COMPACT);
    });
  }
});

test.describe('desktop large 1280x800', () => {
  test.use({ viewport: DESKTOP_VIEWPORT_LARGE });
  for (const path of allGameRoutes) {
    test(`${path} passes the route sweep checks`, async ({ page }) => {
      await navigateTo(page, path);
      await checkHand(page, path, DESKTOP_VIEWPORT_LARGE);
      await checkBoard(page, path, DESKTOP_VIEWPORT_LARGE);
      await checkAction(page, path, DESKTOP_VIEWPORT_LARGE);
      await checkCover(page, path, DESKTOP_VIEWPORT_LARGE);
      await checkLabel(page, path, DESKTOP_VIEWPORT_LARGE);
      await checkDuplicate(page, path, DESKTOP_VIEWPORT_LARGE);
    });
  }
});
