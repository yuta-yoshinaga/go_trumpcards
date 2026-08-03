import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useLocation } from 'react-router-dom';
import { gameCategories, gameRoutes } from '../constants/gameRoutes';
import { SITE_NAME } from '../constants/site';
import { useCategoryExpansion } from '../hooks/useCategoryExpansion';
import { useFavoriteGames } from '../hooks/useFavoriteGames';
import { useGameRouteSearch } from '../hooks/useGameRouteSearch';
import { useRecentGames } from '../hooks/useRecentGames';
import { focusRingWhite } from '../styles/buttonStyles';
import { FavoriteToggleButton } from './nav/FavoriteToggleButton';
import { NavLangToggle } from './nav/NavLangToggle';
import { SoundToggle } from './SoundToggle';
import { TutorialProgressPanel } from './tutorial/TutorialProgressPanel';

/** Lookup map from path to game route for favorite rendering. */
const routeByPath = new Map(gameRoutes.map((r) => [r.path, r]));

/** Persistent left sidebar navigation for large desktop (≥1024px) with search, favorites, categories, and tutorial progress. */
export function DesktopSidebar() {
  const { pathname } = useLocation();
  const { t } = useTranslation('common');
  const [searchTerm, setSearchTerm] = useState('');
  const { favorites, isFavorite, toggleFavorite } = useFavoriteGames();
  const { isExpanded, setExpanded } = useCategoryExpansion();
  const recentGames = useRecentGames(pathname);
  const { filteredPaths } = useGameRouteSearch(searchTerm);

  return (
    <aside
      className="hidden lg:flex lg:flex-col lg:w-60 lg:shrink-0 glass-panel--dark overflow-y-auto pt-[env(safe-area-inset-top)] pl-[env(safe-area-inset-left)]"
      aria-label={t('nav.sidebar', { defaultValue: 'Game navigation' })}
    >
      {/* Site name */}
      <div className="px-3 py-3 border-b border-ds-border-subtle">
        <Link to="/" className="text-ds-text-primary font-display font-bold text-sm inline-flex items-center gap-1.5">
          {SITE_NAME}
        </Link>
      </div>

      {/* AI Game Concierge CTA */}
      <div className="px-3 py-2 border-b border-ds-border-subtle">
        <Link
          to="/discover"
          aria-current={pathname.startsWith('/discover') ? 'page' : undefined}
          className={`block px-3 py-3 text-sm font-medium rounded-md border-l-2 border-ds-accent bg-gradient-to-r from-[rgba(212,168,83,0.12)] to-[rgba(212,168,83,0.04)] text-ds-text-primary hover:bg-[rgba(212,168,83,0.18)] transition-colors min-h-[44px] ${focusRingWhite}`}
        >
          <span aria-hidden="true" className="mr-1.5">
            🎲
          </span>
          {t('nav.discover', { defaultValue: 'おすすめを探す' })}
        </Link>
      </div>

      {/* Search */}
      <div className="px-3 py-2 border-b border-ds-border-subtle">
        <div className="flex items-center gap-1">
          <svg
            data-testid="search-icon"
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className="text-ds-text-muted shrink-0"
            aria-hidden="true"
          >
            <circle cx="11" cy="11" r="8" />
            <line x1="21" x2="16.65" y1="21" y2="16.65" />
          </svg>
          <input
            type="search"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Escape' && searchTerm) {
                e.stopPropagation();
                setSearchTerm('');
              }
            }}
            placeholder={t('nav.searchPlaceholder')}
            aria-label={t('nav.searchPlaceholder')}
            className="flex-1 px-2 py-2 text-sm bg-ds-surface-elevated text-ds-text-primary rounded border border-ds-border-subtle placeholder-ds-text-muted min-w-0 min-h-[44px]"
          />
          {searchTerm && (
            <button
              type="button"
              aria-label={t('nav.searchClear')}
              onClick={() => setSearchTerm('')}
              className={`text-ds-text-muted hover:text-ds-text-primary min-w-[44px] min-h-[44px] flex items-center justify-center text-xs shrink-0 ${focusRingWhite}`}
            >
              ✕
            </button>
          )}
        </div>
      </div>

      {/* Scrollable game list */}
      <nav className="flex-1 overflow-y-auto px-2 py-2" id="sidebar-nav">
        {/* Favorites */}
        {!filteredPaths && favorites.length > 0 && (
          <div className="mb-2">
            <span className="text-ds-text-muted text-[10px] uppercase tracking-wider px-1 font-semibold">
              {t('nav.favoriteGames')}
            </span>
            <div className="mt-1 flex flex-col gap-0.5">
              {favorites.map((gamePath) => {
                const route = routeByPath.get(gamePath);
                if (!route) return null;
                return (
                  <Link
                    key={`fav-${gamePath}`}
                    to={gamePath}
                    aria-current={pathname === gamePath ? 'page' : undefined}
                    className={`inline-flex items-center gap-1.5 px-2 py-2 text-sm rounded transition-colors min-h-[44px] ${pathname === gamePath ? 'bg-ds-accent text-ds-text-on-accent' : 'text-ds-text-primary hover:bg-ds-surface-elevated-hover'}`}
                  >
                    <span aria-hidden="true">{route.icon}</span>
                    {t(route.labelKey)}
                  </Link>
                );
              })}
            </div>
          </div>
        )}

        {/* Recent Games */}
        {!filteredPaths && recentGames.length > 0 && (
          <div className="mb-2">
            <span className="text-ds-text-muted text-[10px] uppercase tracking-wider px-1 font-semibold">
              {t('nav.recentGames')}
            </span>
            <div className="mt-1 flex flex-col gap-0.5">
              {recentGames.map((gamePath) => {
                const route = routeByPath.get(gamePath);
                if (!route) return null;
                return (
                  <Link
                    key={`recent-${gamePath}`}
                    to={gamePath}
                    aria-current={pathname === gamePath ? 'page' : undefined}
                    className={`inline-flex items-center gap-1.5 px-2 py-2 text-sm rounded transition-colors min-h-[44px] ${pathname === gamePath ? 'bg-ds-accent text-ds-text-on-accent' : 'text-ds-text-primary hover:bg-ds-surface-elevated-hover'}`}
                  >
                    <span aria-hidden="true">{route.icon}</span>
                    {t(route.labelKey)}
                  </Link>
                );
              })}
            </div>
          </div>
        )}

        <div aria-live="polite" className="sr-only">
          {searchTerm &&
            (filteredPaths && filteredPaths.size > 0
              ? t('nav.searchResultCount', { count: filteredPaths.size })
              : t('nav.noResults'))}
        </div>
        {/* Search results or category list */}
        {filteredPaths ? (
          <div className="flex flex-col gap-0.5">
            {filteredPaths.size === 0 ? (
              <span className="text-ds-text-muted text-xs px-2 py-1">{t('nav.noResults')}</span>
            ) : (
              gameRoutes
                .filter((r) => filteredPaths.has(r.path))
                .map(({ path, labelKey: routeLabel, icon }) => (
                  <div key={path} className="flex items-center gap-0.5">
                    <Link
                      to={path}
                      aria-current={pathname === path ? 'page' : undefined}
                      className={`inline-flex items-center gap-1.5 px-2 py-2 text-sm rounded transition-colors flex-1 min-h-[44px] ${pathname === path ? 'bg-ds-accent text-ds-text-on-accent' : 'text-ds-text-primary hover:bg-ds-surface-elevated-hover'}`}
                    >
                      <span aria-hidden="true">{icon}</span>
                      {t(routeLabel)}
                    </Link>
                    <FavoriteToggleButton
                      path={path}
                      pressed={isFavorite(path)}
                      onToggle={toggleFavorite}
                      className={() =>
                        `text-ds-accent min-w-[44px] min-h-[44px] flex items-center justify-center text-sm shrink-0 hover:scale-110 transition-transform ${focusRingWhite}`
                      }
                    />
                  </div>
                ))
            )}
          </div>
        ) : (
          gameCategories.map(({ labelKey, icon: catIcon, routes }) => {
            return (
              <details
                key={labelKey}
                className="sidebar-category mb-1"
                open={isExpanded(labelKey)}
                onToggle={(e) => setExpanded(labelKey, e.currentTarget.open)}
              >
                <summary className="text-ds-text-muted text-[10px] uppercase tracking-wider px-1 py-1 font-semibold flex items-center gap-1 cursor-pointer select-none list-none [&::-webkit-details-marker]:hidden">
                  <span aria-hidden="true">{catIcon}</span> {t(labelKey)}
                  <span className="ml-auto text-[8px] text-ds-text-muted sidebar-category-chevron" aria-hidden="true">
                    ▶
                  </span>
                </summary>
                <div className="mt-0.5 flex flex-col gap-0.5 pl-1">
                  {routes.map(({ path, labelKey: routeLabel, icon }) => (
                    <div key={path} className="flex items-center gap-0.5">
                      <Link
                        to={path}
                        aria-current={pathname === path ? 'page' : undefined}
                        className={`inline-flex items-center gap-1.5 px-2 py-2 text-sm rounded transition-colors flex-1 min-h-[44px] ${pathname === path ? 'bg-ds-accent text-ds-text-on-accent' : 'text-ds-text-primary hover:bg-ds-surface-elevated-hover'}`}
                      >
                        <span aria-hidden="true">{icon}</span>
                        {t(routeLabel)}
                      </Link>
                      <FavoriteToggleButton
                        path={path}
                        pressed={isFavorite(path)}
                        onToggle={toggleFavorite}
                        className={() =>
                          `text-ds-accent min-w-[44px] min-h-[44px] flex items-center justify-center text-sm shrink-0 hover:scale-110 transition-transform ${focusRingWhite}`
                        }
                      />
                    </div>
                  ))}
                </div>
              </details>
            );
          })
        )}
      </nav>

      {/* Tutorial progress */}
      <div className="border-t border-ds-border-subtle px-2 py-2">
        <TutorialProgressPanel />
      </div>

      {/* Language + Sound controls */}
      <div className="border-t border-ds-border-subtle px-3 py-2 flex items-center justify-between">
        <NavLangToggle size="sm" />
        <SoundToggle />
      </div>

      {/* Trademark notice and asset credits. Kept in the persistent chrome
          rather than a single page's footer: the non-affiliation statement
          only does its job if a player can reach it from anywhere. */}
      <div className="border-t border-ds-border-subtle px-3 py-1">
        <Link
          to="/legal"
          aria-current={pathname === '/legal' ? 'page' : undefined}
          className={`flex items-center min-h-[44px] text-xs text-ds-text-muted hover:text-ds-text-primary transition-colors ${focusRingWhite}`}
        >
          {t('nav.legal')}
        </Link>
      </div>
    </aside>
  );
}
