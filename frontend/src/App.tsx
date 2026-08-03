import { type ComponentType, Suspense } from 'react';
import { useTranslation } from 'react-i18next';
import { HashRouter, Navigate, Route, Routes } from 'react-router-dom';
import { DesktopSidebar } from './components/DesktopSidebar';
import { ErrorBoundary } from './components/ErrorBoundary';
import { NavBar } from './components/NavBar';
import { RouteErrorBoundary } from './components/RouteErrorBoundary';
import { SkipNavLink } from './components/SkipNavLink';
import { SkeletonBar } from './components/skeleton/SkeletonBar';
import { gameRoutes } from './constants/gameRoutes';
import { resolvePageComponent } from './utils/resolvePageComponent';

// Vite resolves this glob at build time; each match becomes its own chunk
// because the importer is dynamic. Page components have heterogeneous prop
// shapes (e.g., BlackJackPage takes a `variant`), so the value type is
// `ComponentType<any>` — narrowed back to `ComponentType` (no props) in
// `resolvePageComponent` since we render each as `<LazyPage />`.
// biome-ignore lint/suspicious/noExplicitAny: Heterogeneous page prop shapes preclude a stricter generic here.
const pageModules = import.meta.glob<Record<string, ComponentType<any>>>('./pages/*Page.tsx');

const lazyPages = new Map<string, ComponentType>(
  gameRoutes.map(({ path, page }) => [path, resolvePageComponent(pageModules, path, page)]),
);

// The non-game pages match the glob above too, so importing them
// statically made them *both* dynamic and static imports — which a static
// import always wins. They were therefore inlined into the entry chunk that
// every visitor downloads, including visitors who never open Discover or hit
// a 404. Resolving them the same way as game pages is what actually splits
// them out (#4355). Each usage below needs its own Suspense boundary.
const DiscoverPage = resolvePageComponent(pageModules, '/discover', 'Discover');
const DiscoverResultPage = resolvePageComponent(pageModules, '/discover/result', 'DiscoverResult');
const LegalPage = resolvePageComponent(pageModules, '/legal', 'Legal');
const NotFoundPage = resolvePageComponent(pageModules, '*', 'NotFound');

/**
 * Placeholder rendered while a lazy game-page chunk downloads. Shows a
 * `SkeletonBar` so the user sees structure forming on cold-cache /
 * slow-network first paints, bridging visually to the page-specific
 * skeleton that mounts once the chunk resolves. Preserves the existing
 * `role="status"` / `aria-busy` contract for assistive tech and adds an
 * `sr-only` loading label mirroring `SkeletonShell`.
 */
export function RouteSuspenseFallback() {
  const { t } = useTranslation('common');
  return (
    <div role="status" aria-busy="true" className="flex-1 flex flex-col min-h-0">
      <SkeletonBar />
      <span className="sr-only">{t('skeleton.loading')}</span>
    </div>
  );
}

/** Root application component with router and game page routes. */
export default function App() {
  const { t } = useTranslation();
  return (
    <HashRouter>
      <ErrorBoundary>
        <div className="flex flex-col h-full lg:flex-row">
          <SkipNavLink targetId="main-content" label={t('nav.skipToContent')} />
          <DesktopSidebar />
          {/* `min-h-0` is load-bearing, not defensive: this is a flex item in a
              `h-full` column, and without it its automatic minimum size is its
              content height, so it refuses to shrink to the viewport and instead
              grows the document. That is what made 179 of 219 game pages scroll
              vertically on a 375x667 phone even though every page already has an
              internal `flex-1 overflow-y-auto` region meant to absorb the
              overflow. See issue #4373. */}
          <div className="flex flex-col flex-1 min-w-0 min-h-0">
            <NavBar />
            <main id="main-content" tabIndex={-1} className="flex-1 flex flex-col min-h-0">
              {/* Route-scoped boundary: a crash in one page is contained here so
                  the nav/sidebar stay usable; it resets on navigation. The outer
                  ErrorBoundary remains the last resort for chrome crashes (#4314). */}
              <RouteErrorBoundary>
                <Routes>
                  {gameRoutes.map(({ path }) => {
                    const LazyPage = lazyPages.get(path);
                    if (!LazyPage) return null;
                    return (
                      <Route
                        key={path}
                        path={path}
                        element={
                          <Suspense fallback={<RouteSuspenseFallback />}>
                            <LazyPage />
                          </Suspense>
                        }
                      />
                    );
                  })}
                  {/* AI Game Concierge — survey + recommendation result. */}
                  <Route
                    path="/discover"
                    element={
                      <Suspense fallback={<RouteSuspenseFallback />}>
                        <DiscoverPage />
                      </Suspense>
                    }
                  />
                  <Route
                    path="/discover/result"
                    element={
                      <Suspense fallback={<RouteSuspenseFallback />}>
                        <DiscoverResultPage />
                      </Suspense>
                    }
                  />
                  {/* Trademark notice and asset credits. Reachable from the nav
                      because the statement has to reach players, not only
                      readers of the repository. */}
                  <Route
                    path="/legal"
                    element={
                      <Suspense fallback={<RouteSuspenseFallback />}>
                        <LegalPage />
                      </Suspense>
                    }
                  />
                  {/* BlackJack lives at "/", but external links may use "/blackjack". */}
                  <Route path="/blackjack" element={<Navigate to="/" replace />} />
                  {/* Unknown hash routes (e.g., "#/notagame") render the 404
                    surface instead of silently redirecting home — #1902. */}
                  <Route
                    path="*"
                    element={
                      <Suspense fallback={<RouteSuspenseFallback />}>
                        <NotFoundPage />
                      </Suspense>
                    }
                  />
                </Routes>
              </RouteErrorBoundary>
            </main>
          </div>
        </div>
      </ErrorBoundary>
    </HashRouter>
  );
}
