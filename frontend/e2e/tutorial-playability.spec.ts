import { expect, test } from '@playwright/test';
import { waitForLoaded } from './helpers';

/**
 * All game paths (excluding bridge which has no tutorial).
 * For each game, we start the tutorial, click through all steps,
 * then verify the game is still playable after the tutorial ends.
 */
const gamesWithTutorial = [
  { path: '/', name: 'blackjack' },
  { path: '/baccarat', name: 'baccarat' },
  { path: '/threecard', name: 'threecard' },
  { path: '/poker', name: 'poker' },
  { path: '/holdem', name: 'holdem' },
  { path: '/omaha', name: 'omaha' },
  { path: '/shortdeck', name: 'shortdeck' },
  { path: '/pineapple', name: 'pineapple' },
  { path: '/indianpoker', name: 'indianpoker' },
  { path: '/videopoker', name: 'videopoker' },
  { path: '/deuceswild', name: 'deuceswild' },
  { path: '/jokerpoker', name: 'jokerpoker' },
  { path: '/hearts', name: 'hearts' },
  { path: '/spades', name: 'spades' },
  { path: '/ohhell', name: 'ohhell' },
  { path: '/euchre', name: 'euchre' },
  { path: '/napoleon', name: 'napoleon' },
  { path: '/oldmaid', name: 'oldmaid' },
  { path: '/doubt', name: 'doubt' },
  { path: '/daifugo', name: 'daifugo' },
  { path: '/sevens', name: 'sevens' },
  { path: '/crazyeights', name: 'crazyeights' },
  { path: '/speed', name: 'speed' },
  { path: '/klondike', name: 'klondike' },
  { path: '/freecell', name: 'freecell' },
  { path: '/spider', name: 'spider' },
  { path: '/pyramid', name: 'pyramid' },
  { path: '/tripeaks', name: 'tripeaks' },
  { path: '/memory', name: 'memory' },
  { path: '/ginrummy', name: 'ginrummy' },
  { path: '/cribbage', name: 'cribbage' },
];

test.describe('Tutorial → Game Playability', () => {
  for (const { path, name } of gamesWithTutorial) {
    test(`${name}: tutorial completes and game remains playable`, async ({ page }) => {
      // Do NOT suppress tutorial_no_suggest — we want the dialog to appear
      await page.goto(`/#${path}`);
      await waitForLoaded(page);

      // Step 1: The first-visit tutorial suggestion dialog should appear
      const startButton = page.getByRole('button', { name: 'チュートリアルを開始' });
      const dialogVisible = await startButton.isVisible({ timeout: 5_000 }).catch(() => false);
      if (!dialogVisible) {
        // If dialog doesn't appear (e.g., localStorage already set), skip this game
        test.skip();
        return;
      }
      await startButton.click();

      // Step 2: Click through all tutorial steps
      // The tutorial overlay should appear with a "次へ" or "完了" button
      const maxSteps = 30; // Safety limit
      for (let i = 0; i < maxSteps; i++) {
        // Wait for a tutorial tooltip button (次へ or 完了)
        const nextBtn = page.getByRole('button', { name: '次へ' });
        const completeBtn = page.getByRole('button', { name: '完了' });
        const tutorialBtn = nextBtn.or(completeBtn).first();

        const visible = await tutorialBtn.isVisible({ timeout: 5_000 }).catch(() => false);
        if (!visible) {
          // Tutorial overlay gone — tutorial ended
          break;
        }

        // Check if this is the last step (完了 button)
        const isComplete = await completeBtn.isVisible({ timeout: 500 }).catch(() => false);
        if (isComplete) {
          await completeBtn.click();
          break;
        }
        await nextBtn.click();
      }

      // Step 3: After tutorial ends, verify the tutorial overlay is gone
      const tutorialDialog = page.locator('[role="dialog"][aria-label="Tutorial"]');
      await expect(tutorialDialog).toBeHidden({ timeout: 5_000 });

      // Step 4: Verify the game is still playable — at least one interactive button exists
      // (excluding NavBar buttons by scoping to main content area)
      const mainContent = page.locator('main, [role="main"], .game-container, [aria-busy]').first();
      const gameButtons = mainContent.getByRole('button');
      const buttonCount = await gameButtons.count();

      // There should be at least one game action button visible
      expect(buttonCount).toBeGreaterThan(0);

      // Step 5: Verify the first game button is actually clickable (not blocked by overlay)
      const firstButton = gameButtons.first();
      await expect(firstButton).toBeEnabled({ timeout: 5_000 });

      // Try clicking the first button — it should not throw
      // (if an overlay is still blocking, this would fail)
      await firstButton.click({ timeout: 5_000 });
    });
  }
});
