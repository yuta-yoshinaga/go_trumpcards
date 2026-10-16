import { expect, test } from '@playwright/test';
import { missingNapoleonRoundCheckpoints, type NapoleonRoundCheckpoint } from '../src/utils/napoleonPhaseCoverage';
import { gameButton, navigateTo, waitForLoaded } from './helpers';

test.describe('Napoleon E2E', () => {
  test('score table fits its scroll area at 1280×800', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await navigateTo(page, '/napoleon');
    const section = page.locator('[data-tutorial="np-score-table"]');
    const table = section.locator('table').last();
    await expect(table).toBeVisible();
    const dimensions = await table.evaluate((element) => {
      let ancestor = element.parentElement;
      while (ancestor && getComputedStyle(ancestor).overflowX === 'visible') ancestor = ancestor.parentElement;
      if (!ancestor) throw new Error('score table has no horizontal overflow ancestor');
      return { scrollWidth: ancestor.scrollWidth, clientWidth: ancestor.clientWidth };
    });
    expect(dimensions.scrollWidth).toBeLessThanOrEqual(dimensions.clientWidth);
  });

  test('navigates, resets, and plays through phase transitions', async ({ page }) => {
    await navigateTo(page, '/napoleon');

    // Click リセット to start (mid-game: confirm dialog)
    const midResetButton = page.getByRole('button', { name: 'リセット' });
    await expect(midResetButton).toBeVisible();
    await midResetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify round/trick info is visible
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();

    // Verify score table is visible
    await expect(page.getByText('スコア', { exact: true }).first()).toBeVisible();

    const bidButton = gameButton(page, 'ビッド');
    const passButton = gameButton(page, 'パス');
    const declareButton = gameButton(page, '宣言');
    // Matches both the idle "交換" label and the dynamic "<card> を捨てる" label shown once a card is selected.
    const exchangeButton = page.getByRole('button', { name: /交換|を捨てる/ });
    const playButton = page.getByRole('button', { name: '出す' });
    const nextTrickButton = page.getByRole('button', { name: '次のトリック' });
    const nextRoundButton = page.getByRole('button', { name: '次のラウンド' });
    const handCards = page.locator('button[aria-pressed]:has(img):not([aria-disabled="true"])');
    const anyResetButton = page.getByRole('button', { name: /リセット|次のゲーム/ });

    // Observe one complete round path, including optional phases when this deal uses them.
    const MAX_TURNS = 80;
    let interactions = 0;
    let playActions = 0;
    const required = new Set<NapoleonRoundCheckpoint>(['bid', 'play', 'round-boundary']);
    const observed = new Set<NapoleonRoundCheckpoint>();
    for (let turn = 0; turn < MAX_TURNS; turn++) {
      await expect
        .poll(
          async () => (await page.getByTestId('game-footer-actions').isVisible()) || (await anyResetButton.isVisible()),
          { timeout: 10_000 },
        )
        .toBe(true);

      const bidVisible = await bidButton.isVisible();
      const passVisible = await passButton.isVisible();
      const declareVisible = await declareButton.isVisible();
      const exchangeVisible = await exchangeButton.isVisible();
      const playVisible = await playButton.isVisible();
      const nextTrickVisible = await nextTrickButton.isVisible();
      const nextRoundVisible = await nextRoundButton.isVisible();

      // A round boundary follows the completed trick, either as the next round or the result screen.
      if (
        !bidVisible &&
        !passVisible &&
        !declareVisible &&
        !exchangeVisible &&
        !playVisible &&
        !nextTrickVisible &&
        !nextRoundVisible
      ) {
        observed.add('round-boundary');
        break;
      }

      // Bid phase: bid or pass
      if (bidVisible || passVisible) {
        observed.add('bid');
        interactions++;
        if (bidVisible) {
          const bidInput = page.locator('input[aria-label="bid-input"]');
          if (await bidInput.isVisible()) {
            await bidInput.fill('13');
          }
          await bidButton.click();
        } else {
          await passButton.click();
        }
        await waitForLoaded(page);
        continue;
      }

      // Trump declaration phase
      if (declareVisible) {
        required.add('declare');
        observed.add('declare');
        interactions++;
        await page.locator('[data-testid^="np-adjutant-option-"]').first().click();
        await expect(declareButton).toBeEnabled();
        await declareButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Kitty exchange phase
      if (exchangeVisible) {
        required.add('exchange');
        observed.add('exchange');
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        if (await exchangeButton.isEnabled()) {
          await exchangeButton.click();
          await waitForLoaded(page);
        }
        continue;
      }

      // Play phase: select a card and play
      if (playVisible) {
        observed.add('play');
        interactions++;
        const cardCount = await handCards.count();
        if (cardCount > 0) {
          await handCards.first().click();
        }
        if ((await playButton.isVisible()) && (await playButton.isEnabled())) {
          await playButton.click();
          playActions++;
          await waitForLoaded(page);
        }
        continue;
      }

      // Trick end
      if (nextTrickVisible) {
        required.add('trick-end');
        observed.add('trick-end');
        interactions++;
        await nextTrickButton.click();
        await waitForLoaded(page);
        continue;
      }

      // Round end
      if (nextRoundVisible) {
        observed.add('round-boundary');
        interactions++;
        await nextRoundButton.click();
        await waitForLoaded(page);
        break;
      }
    }

    const missing = missingNapoleonRoundCheckpoints(required, observed);
    expect(required.size, 'the expected round checkpoint set must not be empty').toBeGreaterThan(0);
    expect(missing, `Napoleon round checkpoints not observed: ${missing.join(', ')}`).toEqual([]);
    expect(interactions).toBeGreaterThan(0);
    expect(playActions, 'at least one play action must be completed').toBeGreaterThan(0);

    // Reset and verify game restarts. Button could be either mid-game (リセット) or end (次のゲーム).
    const midVisible = await midResetButton.isVisible();
    if (midVisible) {
      await midResetButton.click();
      await page.getByRole('button', { name: '確認' }).click();
    } else {
      await page.getByRole('button', { name: '次のゲーム' }).click();
    }
    await waitForLoaded(page);
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });

  test('settings: change CPU difficulty and point limit', async ({ page }) => {
    await navigateTo(page, '/napoleon');

    // Open settings
    const settingsToggle = page.locator('summary', { hasText: '設定' });
    await expect(settingsToggle).toBeVisible();
    await settingsToggle.click();

    // Change CPU difficulty
    const cpuSelect = page.locator('select').first();
    await expect(cpuSelect).toBeVisible();
    await cpuSelect.selectOption('2');

    // Reset with new settings
    const resetButton = page.getByRole('button', { name: 'リセット' });
    await resetButton.click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);

    // Verify game started with new settings
    await expect(page.getByText(/^ラウンド \d+$/).first()).toBeVisible();
  });
});
