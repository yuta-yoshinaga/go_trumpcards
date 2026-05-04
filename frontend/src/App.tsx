import { type ComponentType, Suspense } from 'react';
import { useTranslation } from 'react-i18next';
import { HashRouter, Navigate, Route, Routes } from 'react-router-dom';
import { DesktopSidebar } from './components/DesktopSidebar';
import { ErrorBoundary } from './components/ErrorBoundary';
import { NavBar } from './components/NavBar';
import { SkipNavLink } from './components/SkipNavLink';
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

/** Minimal `aria-busy` placeholder shown while a lazy game-page chunk loads. */
function RouteSuspenseFallback() {
  return <div role="status" aria-busy="true" className="flex-1" />;
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
          <div className="flex flex-col flex-1 min-w-0">
            <NavBar />
            <main id="main-content" tabIndex={-1} className="flex-1 flex flex-col min-h-0">
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
                {/* BlackJack lives at "/", but external links may use "/blackjack". */}
                <Route path="/blackjack" element={<Navigate to="/" replace />} />
                {/* Unknown hash routes (e.g., "#/notagame") fall back to home
                    instead of rendering an empty <main>. */}
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </main>
          </div>
        </div>
      </ErrorBoundary>
    </HashRouter>
  );
}
