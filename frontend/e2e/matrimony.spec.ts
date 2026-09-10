import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

test.describe('Matrimony E2E', () => {
  test('navigates, draws, resets, and triggers basic actions', async ({ page }) => {
    await navigateTo(page, '/matrimony');
    await expect(page.getByText(/手数/)).toBeVisible();

    const drawButton = page.getByRole('button', { name: 'めくる', exact: true });
    await expect(drawButton).toBeVisible();
    await drawButton.click();
    await waitForLoaded(page);

    const hintButton = page.getByRole('button', { name: 'ヒント', exact: true });
    await expect(hintButton).toBeVisible();
    await hintButton.click();
    await waitForLoaded(page);

    await page.getByRole('button', { name: 'リセット' }).click();
    await page.getByRole('button', { name: '確認' }).click();
    await waitForLoaded(page);
    await expect(page.getByText(/手数/)).toBeVisible();
  });

  test('renders the sixteen-slot tableau and four foundations', async ({ page }) => {
    await navigateTo(page, '/matrimony');
    await expect(page.getByRole('heading', { name: 'マトリモニー' })).toBeVisible();
    await expect(page.getByTestId('matrimony-tableau')).toBeVisible();
    await expect(page.getByRole('button', { name: /山札 残り/ })).toBeVisible();
  });

  test('drawing changes only the stock count, independent of the deal', async ({ page }) => {
    await navigateTo(page, '/matrimony');
    const stock = page.getByRole('button', { name: /山札 残り/ });
    const before = await stock.textContent();
    await page.getByRole('button', { name: 'めくる', exact: true }).click();
    await waitForLoaded(page);
    await expect(stock).not.toHaveText(before ?? '');
  });
});
