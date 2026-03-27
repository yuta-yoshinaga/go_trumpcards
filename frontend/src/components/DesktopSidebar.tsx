import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useLocation } from 'react-router-dom';
import { gameCategories, gameRoutes } from '../constants/gameRoutes';
import { SITE_NAME } from '../constants/site';
import { useFavoriteGames } from '../hooks/useFavoriteGames';
import { SoundToggle } from './SoundToggle';
import { TutorialProgressPanel } from './tutorial/TutorialProgressPanel';

/** Lookup map from path to game route for favorite rendering. */
const routeByPath = new Map(gameRoutes.map((r) => [r.path, r]));

interface DesktopSidebarProps {
  soundMuted?: boolean;
  onSoundToggle?: () => void;
}

/** Persistent left sidebar navigation for large desktop (≥1024px) with search, favorites, categories, and tutorial progress. */
export function DesktopSidebar({ soundMuted, onSoundToggle }: DesktopSidebarProps) {
  const { pathname } = useLocation();
  const { t, i18n } = useTranslation('common');
  const currentLang = i18n.language;
  const [searchTerm, setSearchTerm] = useState('');
  const { favorites, isFavorite, toggleFavorite } = useFavoriteGames();

  /** Pre-compute bilingual names for search filtering. */
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
  const filteredPaths = useMemo(() => {
    if (!searchTerm) return null;
    const lower = searchTerm.toLowerCase();
    return new Set(
      searchableRoutes.filter(({ ja, en }) => ja.includes(lower) || en.includes(lower)).map(({ route }) => route.path),
    );
  }, [searchTerm, searchableRoutes]);

  return (
    <aside
      className="hidden lg:flex lg:flex-col lg:w-60 lg:shrink-0 glass-panel--dark overflow-y-auto"
      aria-label={t('nav.sidebar', { defaultValue: 'Game navigation' })}
    >
      {/* Site name */}
      <div className="px-3 py-3 border-b border-white/10">
        <Link to="/" className="text-white font-bold text-sm inline-flex items-center gap-1.5">
          {SITE_NAME}
        </Link>
      </div>

      {/* Search */}
      <div className="px-3 py-2 border-b border-white/10">
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
            className="flex-1 px-2 py-1.5 text-xs bg-gray-700 text-white rounded border border-gray-600 placeholder-gray-400 min-w-0"
          />
          {searchTerm && (
            <button
              type="button"
              aria-label={t('nav.searchClear')}
              onClick={() => setSearchTerm('')}
              className="text-gray-400 hover:text-white flex items-center justify-center text-xs shrink-0"
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
            <span className="text-gray-400 text-[10px] uppercase tracking-wider px-1 font-semibold">
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
                    className={`inline-flex items-center gap-1.5 px-2 py-1.5 text-xs rounded transition-colors ${pathname === gamePath ? 'bg-blue-600 text-white' : 'text-gray-200 hover:bg-white/10'}`}
                  >
                    <span aria-hidden="true">{route.icon}</span>
                    {t(route.labelKey)}
                  </Link>
                );
              })}
            </div>
          </div>
        )}

        {/* Search results or category list */}
        {filteredPaths ? (
          <div className="flex flex-col gap-0.5">
            {filteredPaths.size === 0 ? (
              <span className="text-gray-400 text-xs px-2 py-1">{t('nav.noResults')}</span>
            ) : (
              gameRoutes
                .filter((r) => filteredPaths.has(r.path))
                .map(({ path, labelKey: routeLabel, icon }) => (
                  <div key={path} className="flex items-center gap-0.5">
                    <Link
                      to={path}
                      aria-current={pathname === path ? 'page' : undefined}
                      className={`inline-flex items-center gap-1.5 px-2 py-1.5 text-xs rounded transition-colors flex-1 ${pathname === path ? 'bg-blue-600 text-white' : 'text-gray-200 hover:bg-white/10'}`}
                    >
                      <span aria-hidden="true">{icon}</span>
                      {t(routeLabel)}
                    </Link>
                    <button
                      type="button"
                      aria-label={isFavorite(path) ? t('nav.removeFavorite') : t('nav.addFavorite')}
                      onClick={() => toggleFavorite(path)}
                      className="text-yellow-400 w-6 h-6 flex items-center justify-center text-xs shrink-0 hover:scale-110 transition-transform"
                    >
                      {isFavorite(path) ? '★' : '☆'}
                    </button>
                  </div>
                ))
            )}
          </div>
        ) : (
          gameCategories.map(({ labelKey, icon: catIcon, routes }) => (
            <div key={labelKey} className="mb-2">
              <span className="text-gray-400 text-[10px] uppercase tracking-wider px-1 font-semibold flex items-center gap-1">
                <span aria-hidden="true">{catIcon}</span> {t(labelKey)}
              </span>
              <div className="mt-1 flex flex-col gap-0.5">
                {routes.map(({ path, labelKey: routeLabel, icon }) => (
                  <div key={path} className="flex items-center gap-0.5">
                    <Link
                      to={path}
                      aria-current={pathname === path ? 'page' : undefined}
                      className={`inline-flex items-center gap-1.5 px-2 py-1.5 text-xs rounded transition-colors flex-1 ${pathname === path ? 'bg-blue-600 text-white' : 'text-gray-200 hover:bg-white/10'}`}
                    >
                      <span aria-hidden="true">{icon}</span>
                      {t(routeLabel)}
                    </Link>
                    <button
                      type="button"
                      aria-label={isFavorite(path) ? t('nav.removeFavorite') : t('nav.addFavorite')}
                      onClick={() => toggleFavorite(path)}
                      className="text-yellow-400 w-6 h-6 flex items-center justify-center text-xs shrink-0 hover:scale-110 transition-transform"
                    >
                      {isFavorite(path) ? '★' : '☆'}
                    </button>
                  </div>
                ))}
              </div>
            </div>
          ))
        )}
      </nav>

      {/* Tutorial progress */}
      <div className="border-t border-white/10 px-2 py-2">
        <TutorialProgressPanel />
      </div>

      {/* Language + Sound controls */}
      <div className="border-t border-white/10 px-3 py-2 flex items-center justify-between">
        <div className="flex gap-0.5">
          <button
            type="button"
            aria-label={t('nav.switchToJa')}
            aria-pressed={currentLang === 'ja'}
            onClick={() => i18n.changeLanguage('ja')}
            className={`px-2 py-1 text-xs font-bold rounded-l transition-colors ${currentLang === 'ja' ? 'bg-blue-500 text-white' : 'bg-gray-600 text-gray-200 hover:bg-gray-500'}`}
          >
            JA
          </button>
          <button
            type="button"
            aria-label={t('nav.switchToEn')}
            aria-pressed={currentLang === 'en'}
            onClick={() => i18n.changeLanguage('en')}
            className={`px-2 py-1 text-xs font-bold rounded-r transition-colors ${currentLang === 'en' ? 'bg-blue-500 text-white' : 'bg-gray-600 text-gray-200 hover:bg-gray-500'}`}
          >
            EN
          </button>
        </div>
        {soundMuted !== undefined && onSoundToggle && <SoundToggle muted={soundMuted} onToggle={onSoundToggle} />}
      </div>
    </aside>
  );
}
