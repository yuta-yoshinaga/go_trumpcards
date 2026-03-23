import type { Page } from '@playwright/test';

/** Wait for the page to finish loading (aria-busy becomes "false"). */
export async function waitForLoaded(page: Page) {
  await page.waitForSelector('[aria-busy="false"]', { timeout: 30_000 });
}

/** Navigate to a hash-routed page and wait for it to load. Suppresses the tutorial suggestion dialog. */
export async function navigateTo(page: Page, path: string) {
  // Suppress the first-visit tutorial suggestion dialog to avoid blocking game interaction
  await page.addInitScript(() => {
    localStorage.setItem('tutorial_no_suggest', 'true');
  });
  await page.goto(`/#${path}`);
  await waitForLoaded(page);
}
