import { fireEvent, render, screen } from '@testing-library/react';
import { Link, MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { RouteErrorBoundary } from './RouteErrorBoundary';

function Boom(): never {
  throw new Error('page boom');
}
function OkPage() {
  return <div>ok page</div>;
}

describe('RouteErrorBoundary', () => {
  let errorSpy: ReturnType<typeof vi.spyOn>;
  beforeEach(() => {
    // React logs caught render errors; silence for a clean test output.
    errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
  });
  afterEach(() => {
    errorSpy.mockRestore();
  });

  it('contains a routed page crash in the boundary fallback', () => {
    render(
      <MemoryRouter initialEntries={['/boom']}>
        <RouteErrorBoundary>
          <Routes>
            <Route path="/boom" element={<Boom />} />
            <Route path="/ok" element={<OkPage />} />
          </Routes>
        </RouteErrorBoundary>
      </MemoryRouter>,
    );
    // The ErrorBoundary fallback (role="alert") is shown instead of crashing up.
    expect(screen.getByRole('alert')).toBeInTheDocument();
  });

  it('resets when navigating to a different route (pathname key remounts it)', () => {
    render(
      <MemoryRouter initialEntries={['/boom']}>
        {/* Nav lives outside the boundary so it survives the crashed page. */}
        <Link to="/ok">go ok</Link>
        <RouteErrorBoundary>
          <Routes>
            <Route path="/boom" element={<Boom />} />
            <Route path="/ok" element={<OkPage />} />
          </Routes>
        </RouteErrorBoundary>
      </MemoryRouter>,
    );
    expect(screen.getByRole('alert')).toBeInTheDocument();

    fireEvent.click(screen.getByText('go ok'));

    // Navigation changed the pathname key → boundary remounts fresh → the new
    // (non-crashing) page renders and the fallback is gone.
    expect(screen.getByText('ok page')).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});
