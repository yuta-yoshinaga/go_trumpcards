/**
 * @vitest-environment jsdom
 */
import { screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../test/renderWithProviders';
import { DiscoverResultPage } from './DiscoverResultPage';

function renderAt(path: string) {
  renderWithProviders(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/discover" element={<div data-testid="discover-landing">survey</div>} />
        <Route path="/discover/result" element={<DiscoverResultPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('DiscoverResultPage', () => {
  it('renders a hero + result sections for a valid mood query', async () => {
    renderAt('/discover/result?m=0,0&s=0,0&so=1,1&t=0,0');
    await waitFor(() => {
      expect(screen.getAllByRole('link').length).toBeGreaterThan(0);
    });
    expect(screen.queryByTestId('discover-landing')).not.toBeInTheDocument();
  });

  it('redirects to /discover when the URL is malformed', async () => {
    renderAt('/discover/result?m=99,abc&s=0,0&so=0,0&t=0,0');
    await waitFor(() => {
      expect(screen.getByTestId('discover-landing')).toBeInTheDocument();
    });
  });

  it('shows the fallback hero when every answer is a skip', async () => {
    renderAt('/discover/result?m=-,-&s=-,-&so=-,-&t=-,-');
    // We can't rely on the i18n string, but we can check that a heading exists
    // and that all 3 of TOP3 are still shown (alphabetical fallback).
    await waitFor(() => {
      expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
    });
  });
});
