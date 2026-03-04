import type { Page } from '@playwright/test';

/** Wait for the page to finish loading (aria-busy becomes "false"). */
export async function waitForLoaded(page: Page) {
  await page.waitForSelector('[aria-busy="false"]', { timeout: 30_000 });
}

/** Navigate to a hash-routed page and wait for it to load. */
export async function navigateTo(page: Page, path: string) {
  await page.goto(`/#${path}`);
  await waitForLoaded(page);
}
