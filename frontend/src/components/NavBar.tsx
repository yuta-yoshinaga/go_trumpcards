import { useCallback, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useLocation } from 'react-router-dom';
import { gameCategories, gameRoutes } from '../constants/gameRoutes';
import { SITE_NAME } from '../constants/site';
import { useIsLargeDesktop, useIsMediumDesktop, useIsMobile } from '../hooks/useCardDimensions';
import { useDetailsOutsideClick } from '../hooks/useDetailsOutsideClick';
import { useFavoriteGames } from '../hooks/useFavoriteGames';
import { useGameRouteSearch } from '../hooks/useGameRouteSearch';
import { useNavFocusTrap } from '../hooks/useNavFocusTrap';
import { useRecentGames } from '../hooks/useRecentGames';
import { focusRingWhite } from '../styles/buttonStyles';
import { FavoriteToggleButton } from './nav/FavoriteToggleButton';
import { NavLangToggle } from './nav/NavLangToggle';
import { SoundToggle } from './SoundToggle';
import { TutorialProgressPanel } from './tutorial/TutorialProgressPanel';

/** SVG icon for the hamburger menu (open state). */
function MenuIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <line x1="4" x2="20" y1="12" y2="12" />
      <line x1="4" x2="20" y1="6" y2="6" />
      <line x1="4" x2="20" y1="18" y2="18" />
    </svg>
  );
}

/** SVG icon for the close button (menu dismiss). */
function CloseIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M18 6 6 18M6 6l12 12" />
    </svg>
  );
}

/** Lookup map from path to game route for recent/favorite rendering. */
const routeByPath = new Map(gameRoutes.map((r) => [r.path, r]));

/** Renders the top navigation bar with game links grouped by category and language toggle. */
export function NavBar() {
  const { pathname } = useLocation();
  const { t } = useTranslation('common');
  const [isOpen, setIsOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const isMobile = useIsMobile();
  const isMediumDesktop = useIsMediumDesktop();
  const isLargeDesktop = useIsLargeDesktop();
  const recentGames = useRecentGames(pathname);
  const { favorites, isFavorite, toggleFavorite } = useFavoriteGames();
  const navRef = useRef<HTMLElement>(null);
  const toggleRef = useRef<HTMLButtonElement>(null);

  useNavFocusTrap(navRef, toggleRef, isOpen, isMobile);
  useDetailsOutsideClick(navRef, !isMobile && !isMediumDesktop);

  const closeMenu = useCallback(() => {
    setIsOpen(false);
    setSearchTerm('');
  }, []);

  const handleNavKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      closeMenu();
    }
  };

  const { filteredRoutes } = useGameRouteSearch(searchTerm);

  return (
    <div className="glass-panel--dark lg:hidden relative z-30 pt-[env(safe-area-inset-top)] pl-[env(safe-area-inset-left)] pr-[env(safe-area-inset-right)]">
      <div className="flex items-center justify-between sm:hidden my-2 mx-2.5">
        <Link
          to="/"
          className="text-ds-text-primary font-display font-bold min-h-[44px] inline-flex items-center"
          onClick={closeMenu}
        >
          {SITE_NAME}
        </Link>
        <div className="flex items-center gap-2">
          <SoundToggle />
          <NavLangToggle />
          <button
            ref={toggleRef}
            type="button"
            onClick={() => (isOpen ? closeMenu() : setIsOpen(true))}
            aria-expanded={isOpen}
            aria-controls="main-nav"
            aria-label={isOpen ? t('nav.closeMenu') : t('nav.openMenu')}
            className={`text-ds-text-primary p-2 min-h-[44px] min-w-[44px] flex items-center justify-center ${focusRingWhite}`}
          >
            {isOpen ? <CloseIcon /> : <MenuIcon />}
          </button>
        </div>
      </div>

      <nav
        ref={navRef}
        id="main-nav"
        onKeyDown={handleNavKeyDown}
        className={`${isOpen ? 'flex' : 'hidden'} flex-col gap-2 mx-2.5 mb-2 sm:flex sm:flex-row sm:flex-wrap sm:items-start sm:justify-end sm:my-2`}
      >
        {isMobile && (
          <div className="flex items-center gap-1">
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
              className="flex-1 px-3 py-2 text-sm bg-ds-surface-elevated text-ds-text-primary rounded border border-ds-border-subtle placeholder-ds-text-muted min-h-[44px]"
            />
            {searchTerm && (
              <button
                type="button"
                aria-label={t('nav.searchClear')}
                onClick={() => setSearchTerm('')}
                className="text-ds-text-muted hover:text-ds-text-primary min-h-[44px] min-w-[44px] flex items-center justify-center text-sm"
              >
                ✕
              </button>
            )}
          </div>
        )}
        {!filteredRoutes && (
          <Link
            to="/discover"
            aria-current={pathname.startsWith('/discover') ? 'page' : undefined}
            onClick={closeMenu}
            className={`inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded min-h-[44px] border-l-2 border-ds-accent bg-gradient-to-r from-[rgba(212,168,83,0.12)] to-[rgba(212,168,83,0.04)] text-ds-text-primary hover:bg-[rgba(212,168,83,0.18)] transition-colors ${focusRingWhite}`}
          >
            <span aria-hidden="true">🎲</span>
            {t('nav.discover')}
          </Link>
        )}
        {isMobile && !filteredRoutes && favorites.length > 0 && (
          <div className="flex flex-col gap-1">
            <span className="text-ds-text-muted text-xs uppercase tracking-wider px-1 py-1">
              {t('nav.favoriteGames')}
            </span>
            {favorites.map((gamePath) => {
              const route = routeByPath.get(gamePath);
              if (!route) return null;
              return (
                <Link
                  key={`fav-${gamePath}`}
                  to={gamePath}
                  aria-current={pathname === gamePath ? 'page' : undefined}
                  onClick={closeMenu}
                  className={`inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded min-h-[44px] transition-[colors,box-shadow] duration-150${pathname === gamePath ? ' bg-ds-accent text-ds-text-on-accent' : ' bg-ds-surface-elevated text-ds-text-primary hover:bg-ds-surface-elevated-hover hover:shadow-md'}`}
                >
                  <span aria-hidden="true">{route.icon}</span>
                  {t(route.labelKey)}
                </Link>
              );
            })}
          </div>
        )}
        {isMobile && !filteredRoutes && recentGames.length > 0 && (
          <div className="flex flex-col gap-1">
            <span className="text-ds-text-muted text-xs uppercase tracking-wider px-1 py-1">
              {t('nav.recentGames')}
            </span>
            {recentGames.map((gamePath) => {
              const route = routeByPath.get(gamePath);
              if (!route) return null;
              return (
                <Link
                  key={`recent-${gamePath}`}
                  to={gamePath}
                  aria-current={pathname === gamePath ? 'page' : undefined}
                  onClick={closeMenu}
                  className={`inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded min-h-[44px] transition-[colors,box-shadow] duration-150${pathname === gamePath ? ' bg-ds-accent text-ds-text-on-accent' : ' bg-ds-surface-elevated text-ds-text-primary hover:bg-ds-surface-elevated-hover hover:shadow-md'}`}
                >
                  <span aria-hidden="true">{route.icon}</span>
                  {t(route.labelKey)}
                </Link>
              );
            })}
          </div>
        )}
        {isMobile && (
          <div aria-live="polite" className="sr-only">
            {searchTerm &&
              (filteredRoutes && filteredRoutes.length > 0
                ? t('nav.searchResultCount', { count: filteredRoutes.length })
                : t('nav.noResults'))}
          </div>
        )}
        {filteredRoutes ? (
          <div className="flex flex-col gap-1">
            {filteredRoutes.length === 0 ? (
              <span className="text-ds-text-muted text-xs px-3 py-2">{t('nav.noResults')}</span>
            ) : (
              filteredRoutes.map(({ path, labelKey: routeLabel, icon }) => (
                <Link
                  key={path}
                  to={path}
                  aria-current={pathname === path ? 'page' : undefined}
                  onClick={closeMenu}
                  className={`inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded min-h-[44px] transition-[colors,box-shadow] duration-150${pathname === path ? ' bg-ds-accent text-ds-text-on-accent' : ' bg-ds-surface-elevated text-ds-text-primary hover:bg-ds-surface-elevated-hover hover:shadow-md'}`}
                >
                  <span aria-hidden="true">{icon}</span>
                  {t(routeLabel)}
                </Link>
              ))
            )}
          </div>
        ) : (
          <div className="flex flex-col gap-1 sm:flex-row sm:flex-wrap sm:flex-1 sm:justify-end sm:gap-3">
            {gameCategories.map(({ labelKey, icon: catIcon, routes }) => {
              return (
                <details
                  key={labelKey}
                  className="nav-category sm:flex sm:items-center"
                  open
                  onToggle={(e) => {
                    // NavBar is lg:hidden, so we are always on mobile or
                    // medium desktop. Force categories to stay open on every
                    // applicable breakpoint so the full 77-game catalog is
                    // discoverable on first visit (#1698).
                    if (!e.currentTarget.open) {
                      e.currentTarget.open = true;
                    }
                  }}
                >
                  <summary className="text-ds-text-muted text-xs uppercase tracking-wider px-1 py-2 cursor-pointer select-none min-h-[44px] flex items-center gap-1 sm:cursor-default sm:py-0 sm:min-h-0 shrink-0">
                    <span aria-hidden="true">{catIcon}</span> {t(labelKey)}
                  </summary>
                  <div className="nav-dropdown flex flex-col gap-1 pl-2 pb-1 sm:flex-row sm:pl-0 sm:pb-0">
                    {routes.map(({ path, labelKey: routeLabel, icon }) => (
                      <div key={path} className="flex items-center gap-1">
                        <Link
                          to={path}
                          aria-current={pathname === path ? 'page' : undefined}
                          onClick={closeMenu}
                          className={`inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded min-h-[44px] transition-[colors,box-shadow] duration-150 flex-1${pathname === path ? ' bg-ds-accent text-ds-text-on-accent' : ' bg-ds-surface-elevated text-ds-text-primary hover:bg-ds-surface-elevated-hover hover:shadow-md'}`}
                        >
                          <span aria-hidden="true">{icon}</span>
                          {t(routeLabel)}
                        </Link>
                        {isMobile && (
                          <FavoriteToggleButton
                            path={path}
                            pressed={isFavorite(path)}
                            onToggle={toggleFavorite}
                            className={(pressed) =>
                              `min-h-[44px] min-w-[44px] flex items-center justify-center text-sm shrink-0 transition-colors ${
                                pressed ? 'text-ds-accent' : 'text-ds-text-muted hover:text-ds-accent'
                              }`
                            }
                          />
                        )}
                      </div>
                    ))}
                  </div>
                </details>
              );
            })}
          </div>
        )}
        <TutorialProgressPanel />
        {/* Rendered below the large-desktop breakpoint only, because
            DesktopSidebar (which is `hidden lg:flex`) carries the same link in
            its footer above it. Keying this on `isMobile` instead left tablet
            and small-desktop widths with no route to the notice at all. */}
        {!isLargeDesktop && (
          <Link
            to="/legal"
            aria-current={pathname === '/legal' ? 'page' : undefined}
            onClick={closeMenu}
            className={`flex items-center min-h-[44px] px-3 text-xs text-ds-text-muted hover:text-ds-text-primary transition-colors ${focusRingWhite}`}
          >
            {t('nav.legal')}
          </Link>
        )}
        <div className="hidden sm:flex sm:items-center sm:gap-2">
          <SoundToggle />
          <NavLangToggle />
        </div>
      </nav>
    </div>
  );
}
