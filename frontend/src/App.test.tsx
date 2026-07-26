import { render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import App, { RouteSuspenseFallback } from './App';
import { DISCOVER_DRAFT_KEY } from './hooks/useSurveyDraft';
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
  beforeEach(() => {
    // A survey draft with every question answered makes /discover legitimately
    // NOT show the survey: DiscoverPage's submit effect fires immediately and
    // redirects to /discover/result with the stored answers. So the two
    // assertions below are asking for `discover-survey` in a state where its
    // absence is correct — which is why they failed intermittently, and only in
    // a full-suite run where an earlier test could leave such a draft behind.
    // Verified in a browser: seeding this key lands on
    // #/discover/result?m=0,0&... with the result page rendered.
    // `setup.ts` does not clear localStorage globally, and
    // `useSurveyDraft.test.ts` clears only in beforeEach, so a completed draft
    // can outlive it. Establish the precondition here instead of inheriting it.
    localStorage.removeItem(DISCOVER_DRAFT_KEY);
  });

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

  // Both assertions wait on a lazily-loaded route chunk, so they need more than
  // the 5s global asyncUtilTimeout from setup.ts. In a full-suite run every core
  // is busy compiling other files, and the wait is for a dynamic import to
  // resolve rather than for a state update — the second test below measured
  // 5810ms against the 5000ms budget and failed, while passing in ~1s when run
  // alone. Raising just these two keeps the global default tight.
  const CHUNK_TIMEOUT = { timeout: 20000 };

  it('renders the Discover survey at /discover', async () => {
    window.location.hash = '#/discover';
    renderWithProviders(<App />);
    expect(await screen.findByTestId('discover-survey', {}, CHUNK_TIMEOUT)).toBeInTheDocument();
  });

  it('sends /discover/result back to the survey when the params are absent', async () => {
    // DiscoverResultPage redirects to /discover when it cannot parse the
    // search params. Reaching that redirect at all proves its own chunk
    // resolved under Suspense first, so this covers both pages — which is also
    // why it is the slowest of the three: two dynamic imports resolve in
    // sequence, not one.
    window.location.hash = '#/discover/result';
    renderWithProviders(<App />);
    expect(await screen.findByTestId('discover-survey', {}, CHUNK_TIMEOUT)).toBeInTheDocument();
  });
});
