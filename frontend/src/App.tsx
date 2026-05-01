import { type ComponentType, lazy, Suspense } from 'react';
import { useTranslation } from 'react-i18next';
import { HashRouter, Route, Routes } from 'react-router-dom';
import { DesktopSidebar } from './components/DesktopSidebar';
import { ErrorBoundary } from './components/ErrorBoundary';
import { NavBar } from './components/NavBar';
import { SkipNavLink } from './components/SkipNavLink';
import { gameRoutes } from './constants/gameRoutes';

// Vite resolves this glob at build time; each match becomes its own chunk
// because the importer is dynamic. The `eager: false` is the default but stated
// explicitly here for the reader.
const pageModules = import.meta.glob<Record<string, ComponentType>>('./pages/*Page.tsx');

const lazyPages = new Map<string, ComponentType>(
  gameRoutes.map(({ path, page }) => {
    const importPath = `./pages/${page}Page.tsx`;
    const importer = pageModules[importPath];
    if (!importer) {
      throw new Error(`gameRoutes: no module at ${importPath} for path "${path}"`);
    }
    const exportName = `${page}Page`;
    const Lazy = lazy(async () => {
      const m = await importer();
      const Component = m[exportName];
      if (!Component) {
        throw new Error(`gameRoutes: ${importPath} has no export named ${exportName}`);
      }
      return { default: Component };
    });
    return [path, Lazy];
  }),
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
              </Routes>
            </main>
          </div>
        </div>
      </ErrorBoundary>
    </HashRouter>
  );
}
