import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import App, { RouteSuspenseFallback } from './App';
import { renderWithProviders } from './test/renderWithProviders';

describe('RouteSuspenseFallback', () => {
  it('preserves the role="status" + aria-busy contract for assistive tech', () => {
    render(<RouteSuspenseFallback />);
    const status = screen.getByRole('status');
    expect(status).toHaveAttribute('aria-busy', 'true');
  });

  it('renders a SkeletonBar so the visual channel is not blank during chunk download', () => {
    render(<RouteSuspenseFallback />);
    // SkeletonBar marks itself aria-hidden so screen readers stay on the
    // sr-only loading label. Querying by that contract — not by class
    // name — keeps the test stable if styling changes.
    const status = screen.getByRole('status');
    expect(status.querySelector('[aria-hidden="true"]')).not.toBeNull();
  });

  it('exposes the localized loading label to screen readers', () => {
    render(<RouteSuspenseFallback />);
    // Accept either locale's translation of skeleton.loading; the test
    // setup defaults to ja but a contributor's env may differ.
    expect(screen.getByText(/^(Loading…|読み込み中…)$/)).toBeInTheDocument();
  });
});

// The three non-game routes are the only ones whose page component is not
// reached through `gameRoutes`, so nothing else asserted they render at all —
// there is no E2E spec for /discover, /discover/result or a 404 either.
//
// What these guard specifically: those pages are now pulled out of the
// `./pages/*Page.tsx` glob by module NAME (#4355), so a typo in the name
// passed to `resolvePageComponent` — or a page renamed without updating
// App.tsx — throws only when the route is actually mounted. Neither tsc nor
// the build can see it, since the name is a plain string looked up in a glob
// record. Verified by mutation: 'NotFound' → 'NotFund' fails the file with
// `no module at ./pages/NotFundPage.tsx`.
//
// They do NOT prove the Suspense boundaries are present: React 19 renders a
// lazy component with no boundary at all, silently and without a console
// error (probed directly). The boundary only decides whether the skeleton
// shows during the chunk download, so removing one is invisible here.
describe('App non-game routes', () => {
  afterEach(() => {
    window.location.hash = '';
  });

  it('renders the 404 page for an unknown hash route', async () => {
    window.location.hash = '#/notagame';
    renderWithProviders(<App />);
    // Matched by role rather than by the title string: this asserts "the 404
    // page mounted", and NotFoundPage's h1 is the only level-1 heading on the
    // route. Keying on the copy would tie the test to one locale, which is
    // what the sibling tests above avoid with a two-locale regex.
    expect(await screen.findByRole('heading', { level: 1 })).toBeInTheDocument();
  });

  it('renders the Discover survey at /discover', async () => {
    window.location.hash = '#/discover';
    renderWithProviders(<App />);
    expect(await screen.findByTestId('discover-survey')).toBeInTheDocument();
  });

  it('sends /discover/result back to the survey when the params are absent', async () => {
    // DiscoverResultPage redirects to /discover when it cannot parse the
    // search params. Reaching that redirect at all proves its own chunk
    // resolved under Suspense first, so this covers both pages.
    window.location.hash = '#/discover/result';
    renderWithProviders(<App />);
    expect(await screen.findByTestId('discover-survey')).toBeInTheDocument();
  });
});
