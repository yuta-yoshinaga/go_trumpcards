/**
 * @vitest-environment jsdom
 */
import { screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../test/renderWithProviders';

// Hoisted mock: pretend the bundle is still loading so we exercise the
// `if (!bundleReady) return <DiscoverSkeleton />` branch in DiscoverPage.
vi.mock('../hooks/useDiscoverI18nBundle', () => ({
  useDiscoverI18nBundle: () => false,
}));

// Importing after vi.mock so the page sees the mocked hook.
import { DiscoverPage } from './DiscoverPage';

describe('DiscoverPage skeleton branch', () => {
  it('renders the DiscoverSkeleton placeholder while the bundle is loading', () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/discover']}>
        <Routes>
          <Route path="/discover" element={<DiscoverPage />} />
        </Routes>
      </MemoryRouter>,
    );
    const statuses = screen.getAllByRole('status');
    expect(statuses.length).toBeGreaterThan(0);
    expect(statuses.some((s) => s.getAttribute('aria-busy') === 'true')).toBe(true);
  });
});
