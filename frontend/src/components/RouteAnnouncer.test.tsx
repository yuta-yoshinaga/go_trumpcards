import { act, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useNavigate } from 'react-router-dom';
import { afterEach, describe, expect, it } from 'vitest';
import { RouteAnnouncer } from './RouteAnnouncer';

/** A page that renders a button navigating to `to`, plus the focus target. */
function Harness({ initial = '/a' }: { initial?: string }) {
  return (
    <MemoryRouter initialEntries={[initial]}>
      <RouteAnnouncer />
      {/* The real app puts tabIndex={-1} on <main> for the skip link; the
          announcer reuses it, so the harness must provide it too. */}
      <main id="main-content" tabIndex={-1}>
        <Routes>
          <Route path="/a" element={<Nav to="/b" label="go b" title="Page A" />} />
          <Route path="/b" element={<Nav to="/a" label="go a" title="Page B" />} />
        </Routes>
      </main>
    </MemoryRouter>
  );
}

function Nav({ to, label, title }: { to: string; label: string; title: string }) {
  const navigate = useNavigate();
  document.title = title;
  return (
    <button type="button" onClick={() => navigate(to)}>
      {label}
    </button>
  );
}

afterEach(() => {
  document.title = '';
});

describe('RouteAnnouncer', () => {
  it('exposes a polite live region for assistive tech', () => {
    render(<Harness />);
    const region = screen.getByRole('status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    // Visually hidden: this is for screen readers, not a banner.
    expect(region).toHaveClass('sr-only');
  });

  // A first render is an ordinary page load, which assistive tech already
  // announces. Announcing again would duplicate it, and moving focus would
  // yank the user out of whatever they were doing (e.g. a deep link).
  it('says nothing on the first render', () => {
    render(<Harness />);
    expect(screen.getByRole('status')).toHaveTextContent('');
  });

  it('announces the new page after a navigation', () => {
    render(<Harness />);
    act(() => {
      fireEvent.click(screen.getByRole('button', { name: 'go b' }));
    });
    expect(screen.getByRole('status')).toHaveTextContent('Page B');
  });

  // Without this, a keyboard user lands back at the top of a 318-entry nav on
  // every game change and has to tab past all of it to reach the board.
  it('moves focus to the main region after a navigation', () => {
    render(<Harness />);
    act(() => {
      fireEvent.click(screen.getByRole('button', { name: 'go b' }));
    });
    expect(document.activeElement).toBe(document.getElementById('main-content'));
  });

  // The announcer must not assume the target exists: RouteAnnouncer is mounted
  // inside the router, and a crash boundary or a future layout change could
  // render without <main>.
  it('does not throw when the focus target is absent', () => {
    render(
      <MemoryRouter initialEntries={['/a']}>
        <RouteAnnouncer />
        <Routes>
          <Route path="/a" element={<Nav to="/b" label="go b" title="Page A" />} />
          <Route path="/b" element={<Nav to="/a" label="go a" title="Page B" />} />
        </Routes>
      </MemoryRouter>,
    );
    expect(() => {
      act(() => {
        fireEvent.click(screen.getByRole('button', { name: 'go b' }));
      });
    }).not.toThrow();
    expect(screen.getByRole('status')).toHaveTextContent('Page B');
  });
});
