/**
 * @vitest-environment jsdom
 */
import { screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
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
    // The fallback hero shows the "もう少しヒントをください" title; verify by
    // looking up the translation directly so the assertion does not depend on
    // the static blurb content.
    await waitFor(() => {
      expect(screen.getByRole('heading', { level: 1, name: i18n.t('discover:fallback.title') })).toBeInTheDocument();
    });
  });

  it('shows real recommendations (not the fallback) when at least one answer is given', async () => {
    // Only one out of eight questions answered — every other axis is skipped.
    // Per design, partial signal should still produce real recommendations
    // (the skipped axes contribute a 0.5 neutral via `axisScore`).
    renderAt('/discover/result?m=0,-&s=-,-&so=-,-&t=-,-');
    await waitFor(() => {
      expect(
        screen.queryByRole('heading', { level: 1, name: i18n.t('discover:fallback.title') }),
      ).not.toBeInTheDocument();
    });
    // A hero recommendation heading is rendered instead.
    expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
  });
});
