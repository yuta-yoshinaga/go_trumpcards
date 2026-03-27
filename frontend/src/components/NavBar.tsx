import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useLocation } from 'react-router-dom';
import { gameCategories, gameRoutes } from '../constants/gameRoutes';
import { SITE_NAME } from '../constants/site';
import { useIsMediumDesktop, useIsMobile } from '../hooks/useCardDimensions';
import { useFavoriteGames } from '../hooks/useFavoriteGames';
import { useRecentGames } from '../hooks/useRecentGames';
import { focusRingWhite } from '../styles/buttonStyles';
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

interface LangToggleProps {
  currentLang: string;
  i18n: { changeLanguage: (lng: string) => void };
  t: (key: string) => string;
}

/** Language toggle buttons (JA / EN). */
function LangToggle({ currentLang, i18n, t }: LangToggleProps) {
  return (
    <div className="flex gap-0.5">
      <button
        type="button"
        aria-label={t('nav.switchToJa')}
        aria-pressed={currentLang === 'ja'}
        onClick={() => i18n.changeLanguage('ja')}
        className={`px-3 py-2 text-xs font-bold rounded-l min-h-[44px] transition-colors ${currentLang === 'ja' ? 'bg-blue-500 text-white' : 'bg-gray-600 text-gray-200 hover:bg-gray-500'}`}
      >
        JA
      </button>
      <button
        type="button"
        aria-label={t('nav.switchToEn')}
        aria-pressed={currentLang === 'en'}
        onClick={() => i18n.changeLanguage('en')}
        className={`px-3 py-2 text-xs font-bold rounded-r min-h-[44px] transition-colors ${currentLang === 'en' ? 'bg-blue-500 text-white' : 'bg-gray-600 text-gray-200 hover:bg-gray-500'}`}
      >
        EN
      </button>
    </div>
  );
}

/** Lookup map from path to game route for recent/favorite rendering. */
const routeByPath = new Map(gameRoutes.map((r) => [r.path, r]));

interface NavBarProps {
  soundMuted?: boolean;
  onSoundToggle?: () => void;
}

/** Renders the top navigation bar with game links grouped by category and language toggle. */
export function NavBar({ soundMuted, onSoundToggle }: NavBarProps = {}) {
  const { pathname } = useLocation();
  const { t, i18n } = useTranslation('common');
  const currentLang = i18n.language;
  const [isOpen, setIsOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const isMobile = useIsMobile();
  const isMediumDesktop = useIsMediumDesktop();
  const recentGames = useRecentGames(pathname);
  const { favorites, isFavorite, toggleFavorite } = useFavoriteGames();
  const navRef = useRef<HTMLElement>(null);
  const toggleRef = useRef<HTMLButtonElement>(null);
  const wasOpen = useRef(false);

  useEffect(() => {
    if (isOpen && navRef.current) {
      const firstInteractive = navRef.current.querySelector<HTMLElement>('input, a');
      firstInteractive?.focus();
    }
    if (!isOpen && wasOpen.current && toggleRef.current) {
      toggleRef.current.focus();
    }
    wasOpen.current = isOpen;
  }, [isOpen]);

  // Close nav dropdown on outside click (large desktop only)
  useEffect(() => {
    const handleOutsideClick = (e: MouseEvent) => {
      if (!navRef.current) return;
      const openDetails = navRef.current.querySelectorAll('details[open]');
      for (const details of openDetails) {
        if (!details.contains(e.target as Node)) {
          details.removeAttribute('open');
        }
      }
    };
    document.addEventListener('mousedown', handleOutsideClick);
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, []);

  const closeMenu = useCallback(() => {
    setIsOpen(false);
    setSearchTerm('');
  }, []);

  const handleNavKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      closeMenu();
    }
  };

  /** Pre-compute bilingual names for search filtering.
   * Uses explicit lng overrides so the result is language-independent. */
  const searchableRoutes = useMemo(
    () =>
      gameRoutes.map((route) => ({
        route,
        ja: i18n.t(route.labelKey, { lng: 'ja', ns: 'common' }).toLowerCase(),
        en: i18n.t(route.labelKey, { lng: 'en', ns: 'common' }).toLowerCase(),
      })),
    [i18n.t],
  );

  /** Filter game routes by bilingual name match. */
  const filteredRoutes = useMemo(() => {
    if (!searchTerm) return null;
    const lower = searchTerm.toLowerCase();
    return searchableRoutes.filter(({ ja, en }) => ja.includes(lower) || en.includes(lower)).map(({ route }) => route);
  }, [searchTerm, searchableRoutes]);

  return (
    <div className="glass-panel--dark lg:hidden">
      <div className="flex items-center justify-between sm:hidden my-2 mx-2.5">
        <Link to="/" className="text-white font-bold min-h-[44px] inline-flex items-center" onClick={closeMenu}>
          {SITE_NAME}
        </Link>
        <div className="flex items-center gap-2">
          {soundMuted !== undefined && onSoundToggle && <SoundToggle muted={soundMuted} onToggle={onSoundToggle} />}
          <LangToggle currentLang={currentLang} i18n={i18n} t={t} />
          <button
            ref={toggleRef}
            type="button"
            onClick={() => (isOpen ? closeMenu() : setIsOpen(true))}
            aria-expanded={isOpen}
            aria-controls="main-nav"
            aria-label={isOpen ? t('nav.closeMenu') : t('nav.openMenu')}
            className={`text-white p-2 min-h-[44px] min-w-[44px] flex items-center justify-center ${focusRingWhite}`}
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
              className="flex-1 px-3 py-2 text-sm bg-gray-700 text-white rounded border border-gray-600 placeholder-gray-400 min-h-[44px]"
            />
            {searchTerm && (
              <button
                type="button"
                aria-label={t('nav.searchClear')}
                onClick={() => setSearchTerm('')}
                className="text-gray-400 hover:text-white min-h-[44px] min-w-[44px] flex items-center justify-center text-sm"
              >
                ✕
              </button>
            )}
          </div>
        )}
        {isMobile && !filteredRoutes && favorites.length > 0 && (
          <div className="flex flex-col gap-1">
            <span className="text-gray-300 text-xs uppercase tracking-wider px-1 py-1">{t('nav.favoriteGames')}</span>
            {favorites.map((gamePath) => {
              const route = routeByPath.get(gamePath);
              if (!route) return null;
              return (
                <Link
                  key={`fav-${gamePath}`}
                  to={gamePath}
                  aria-current={pathname === gamePath ? 'page' : undefined}
                  onClick={closeMenu}
                  className={`inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded min-h-[44px] transition-[colors,box-shadow] duration-150${pathname === gamePath ? ' bg-blue-600 text-white' : ' bg-gray-600 text-gray-200 hover:bg-gray-500 hover:shadow-md'}`}
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
            <span className="text-gray-300 text-xs uppercase tracking-wider px-1 py-1">{t('nav.recentGames')}</span>
            {recentGames.map((gamePath) => {
              const route = routeByPath.get(gamePath);
              if (!route) return null;
              return (
                <Link
                  key={`recent-${gamePath}`}
                  to={gamePath}
                  aria-current={pathname === gamePath ? 'page' : undefined}
                  onClick={closeMenu}
                  className={`inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded min-h-[44px] transition-[colors,box-shadow] duration-150${pathname === gamePath ? ' bg-blue-600 text-white' : ' bg-gray-600 text-gray-200 hover:bg-gray-500 hover:shadow-md'}`}
                >
                  <span aria-hidden="true">{route.icon}</span>
                  {t(route.labelKey)}
                </Link>
              );
            })}
          </div>
        )}
        {filteredRoutes ? (
          <div className="flex flex-col gap-1">
            {filteredRoutes.length === 0 ? (
              <span className="text-gray-400 text-xs px-3 py-2">{t('nav.noResults')}</span>
            ) : (
              filteredRoutes.map(({ path, labelKey: routeLabel, icon }) => (
                <Link
                  key={path}
                  to={path}
                  aria-current={pathname === path ? 'page' : undefined}
                  onClick={closeMenu}
                  className={`inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded min-h-[44px] transition-[colors,box-shadow] duration-150${pathname === path ? ' bg-blue-600 text-white' : ' bg-gray-600 text-gray-200 hover:bg-gray-500 hover:shadow-md'}`}
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
              const hasActivePage = routes.some(({ path }) => path === pathname);
              return (
                <details
                  key={labelKey}
                  className="nav-category sm:flex sm:items-center"
                  open={isMobile || hasActivePage}
                  onToggle={(e) => {
                    // On mobile or medium desktop (sm-lg), force details to stay open
                    if ((isMobile || isMediumDesktop) && !e.currentTarget.open) {
                      e.currentTarget.open = true;
                    }
                  }}
                >
                  <summary className="text-gray-300 text-xs uppercase tracking-wider px-1 py-2 cursor-pointer select-none min-h-[44px] flex items-center gap-1 sm:cursor-default sm:py-0 sm:min-h-0 shrink-0">
                    <span aria-hidden="true">{catIcon}</span> {t(labelKey)}
                  </summary>
                  <div className="nav-dropdown flex flex-col gap-1 pl-2 pb-1 sm:flex-row sm:pl-0 sm:pb-0">
                    {routes.map(({ path, labelKey: routeLabel, icon }) => (
                      <div key={path} className="flex items-center gap-1">
                        <Link
                          to={path}
                          aria-current={pathname === path ? 'page' : undefined}
                          onClick={closeMenu}
                          className={`inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded min-h-[44px] transition-[colors,box-shadow] duration-150 flex-1${pathname === path ? ' bg-blue-600 text-white' : ' bg-gray-600 text-gray-200 hover:bg-gray-500 hover:shadow-md'}`}
                        >
                          <span aria-hidden="true">{icon}</span>
                          {t(routeLabel)}
                        </Link>
                        {isMobile && (
                          <button
                            type="button"
                            aria-label={isFavorite(path) ? t('nav.removeFavorite') : t('nav.addFavorite')}
                            onClick={() => toggleFavorite(path)}
                            className="text-yellow-400 min-h-[44px] min-w-[44px] flex items-center justify-center text-sm shrink-0"
                          >
                            {isFavorite(path) ? '★' : '☆'}
                          </button>
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
        <div className="hidden sm:flex sm:items-center sm:gap-2">
          {soundMuted !== undefined && onSoundToggle && <SoundToggle muted={soundMuted} onToggle={onSoundToggle} />}
          <LangToggle currentLang={currentLang} i18n={i18n} t={t} />
        </div>
      </nav>
    </div>
  );
}
