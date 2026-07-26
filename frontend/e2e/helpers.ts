import type { Locator, Page } from '@playwright/test';

/** Timeout for quick UI checks (button appeared after an action). */
export const TIMEOUT_QUICK = 1_000;
/** Timeout for standard action visibility (betting round transitions, card animations). */
export const TIMEOUT_ACTION = 3_000;
/** Timeout for slow transitions (tutorial dialogs, post-animation waits). */
export const TIMEOUT_TRANSITION = 5_000;
/** Timeout for game loop iterations (waiting for any actionable element). */
export const TIMEOUT_GAME_LOOP = 10_000;
/** Timeout for full page load (aria-busy). */
export const TIMEOUT_LOADED = 30_000;

/**
 * Check if a locator becomes visible within the given timeout.
 * Returns false on timeout instead of throwing. Use this for
 * conditional branching (e.g., "if insurance button appears, decline it").
 */
export async function isVisibleWithin(locator: Locator, timeout: number): Promise<boolean> {
  try {
    await locator.waitFor({ state: 'visible', timeout });
    return true;
  } catch {
    return false;
  }
}

/** Wait for the page to finish loading (aria-busy becomes "false"). */
export async function waitForLoaded(page: Page) {
  await page.waitForSelector('[aria-busy="false"]', { timeout: TIMEOUT_LOADED });
}

/** Navigate to a hash-routed page and wait for it to load. Suppresses the tutorial suggestion dialog. */
export async function navigateTo(page: Page, path: string) {
  // Suppress the first-visit tutorial suggestion dialog to avoid blocking game interaction
  await page.addInitScript(() => {
    localStorage.setItem('tutorial_no_suggest', 'true');
    // Skip the CPU replay animation. Specs assert on state, not on the
    // pacing of the animation, and at the default 'normal' speed every
    // CPU action costs 800ms — Sevens alone spent ~2.9s per turn waiting.
    localStorage.setItem('cpuReplaySpeed', 'instant');
  });
  await page.goto(`/#${path}`);
  await waitForLoaded(page);
}
