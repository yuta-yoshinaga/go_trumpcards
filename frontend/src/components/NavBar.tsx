import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useLocation } from 'react-router-dom';
import { gameCategories } from '../constants/gameRoutes';

const langToggle = (
  currentLang: string,
  i18n: { changeLanguage: (lng: string) => void },
  t: (key: string) => string,
) => (
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

/** Renders the top navigation bar with game links grouped by category and language toggle. */
export function NavBar() {
  const { pathname } = useLocation();
  const { t, i18n } = useTranslation('common');
  const currentLang = i18n.language;
  const [isOpen, setIsOpen] = useState(false);
  const navRef = useRef<HTMLElement>(null);
  const toggleRef = useRef<HTMLButtonElement>(null);
  const wasOpen = useRef(false);

  useEffect(() => {
    if (isOpen && navRef.current) {
      const firstLink = navRef.current.querySelector<HTMLElement>('a');
      firstLink?.focus();
    }
    if (!isOpen && wasOpen.current && toggleRef.current) {
      toggleRef.current.focus();
    }
    wasOpen.current = isOpen;
  }, [isOpen]);

  const handleNavKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      setIsOpen(false);
    }
  };

  return (
    <div className="glass-panel--dark">
      <div className="flex items-center justify-between sm:hidden my-2 mx-2.5">
        <Link
          to="/"
          className="text-white font-bold min-h-[44px] inline-flex items-center"
          onClick={() => setIsOpen(false)}
        >
          Trump Cards
        </Link>
        <div className="flex items-center gap-2">
          {langToggle(currentLang, i18n, t)}
          <button
            ref={toggleRef}
            type="button"
            onClick={() => setIsOpen(!isOpen)}
            aria-expanded={isOpen}
            aria-controls="main-nav"
            aria-label={isOpen ? t('nav.closeMenu') : t('nav.openMenu')}
            className="text-white p-2 min-h-[44px] min-w-[44px] flex items-center justify-center"
          >
            {isOpen ? '✕' : '☰'}
          </button>
        </div>
      </div>

      <nav
        ref={navRef}
        id="main-nav"
        onKeyDown={handleNavKeyDown}
        className={`${isOpen ? 'flex' : 'hidden'} flex-col gap-2 mx-2.5 mb-2 sm:flex sm:flex-row sm:flex-wrap sm:items-start sm:justify-end sm:my-2`}
      >
        <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:flex-1 sm:justify-end sm:gap-3">
          {gameCategories.map(({ labelKey, routes }) => (
            <div key={labelKey} className="flex flex-col gap-1 sm:flex-row sm:items-center">
              <span className="text-gray-400 text-[10px] uppercase tracking-wider px-1 shrink-0">{t(labelKey)}</span>
              <div className="flex flex-col gap-1 sm:flex-row">
                {routes.map(({ path, labelKey: routeLabel }) => (
                  <Link
                    key={path}
                    to={path}
                    aria-current={pathname === path ? 'page' : undefined}
                    onClick={() => setIsOpen(false)}
                    className={`inline-flex items-center px-3 py-2 text-xs font-medium rounded min-h-[44px] transition-colors${pathname === path ? ' bg-blue-600 text-white' : ' bg-gray-600 text-gray-200 hover:bg-gray-500'}`}
                  >
                    {t(routeLabel)}
                  </Link>
                ))}
              </div>
            </div>
          ))}
        </div>
        <div className="hidden sm:flex">{langToggle(currentLang, i18n, t)}</div>
      </nav>
    </div>
  );
}
