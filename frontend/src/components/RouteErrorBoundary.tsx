import type { ReactNode } from 'react';
import { useLocation } from 'react-router-dom';
import { ErrorBoundary } from './ErrorBoundary';

/**
 * Route-scoped error boundary. Wraps the routed page in an {@link ErrorBoundary}
 * keyed on the current pathname, so a render crash in one game page is contained
 * to the `<main>` content area — the surrounding navigation and sidebar stay
 * usable — and the boundary automatically resets when the user navigates to a
 * different route (the changed `key` remounts it). The app-level
 * `ErrorBoundary` in App.tsx remains as the last-resort boundary for crashes in
 * the chrome itself. See issue #4314.
 */
export function RouteErrorBoundary({ children }: { children: ReactNode }) {
  const { pathname } = useLocation();
  return <ErrorBoundary key={pathname}>{children}</ErrorBoundary>;
}
