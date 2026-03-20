import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useLocation } from 'react-router-dom';
import { gameRoutes } from '../constants/gameRoutes';

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
      className={`px-1.5 py-0.5 text-xs font-bold rounded-l transition-colors ${currentLang === 'ja' ? 'bg-blue-500 text-white' : 'bg-gray-600 text-gray-300 hover:bg-gray-500'}`}
    >
      JA
    </button>
    <button
      type="button"
      aria-label={t('nav.switchToEn')}
      aria-pressed={currentLang === 'en'}
      onClick={() => i18n.changeLanguage('en')}
      className={`px-1.5 py-0.5 text-xs font-bold rounded-r transition-colors ${currentLang === 'en' ? 'bg-blue-500 text-white' : 'bg-gray-600 text-gray-300 hover:bg-gray-500'}`}
    >
      EN
    </button>
  </div>
);

/** Renders the top navigation bar with game links and language toggle. */
export function NavBar() {
  const { pathname } = useLocation();
  const { t, i18n } = useTranslation('common');
  const currentLang = i18n.language;
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="bg-gray-800">
      <div className="flex items-center justify-between sm:hidden my-2 mx-2.5">
        <Link to="/" className="text-white font-bold" onClick={() => setIsOpen(false)}>
          Trump Cards
        </Link>
        <div className="flex items-center gap-2">
          {langToggle(currentLang, i18n, t)}
          <button
            type="button"
            onClick={() => setIsOpen(!isOpen)}
            aria-expanded={isOpen}
            aria-controls="main-nav"
            aria-label={isOpen ? t('nav.closeMenu') : t('nav.openMenu')}
            className="text-white p-2"
          >
            {isOpen ? '✕' : '☰'}
          </button>
        </div>
      </div>

      <nav
        id="main-nav"
        className={`${isOpen ? 'flex' : 'hidden'} flex-col gap-2 mx-2.5 mb-2 sm:flex sm:flex-row sm:flex-wrap sm:items-center sm:justify-end sm:my-2`}
      >
        <div className="flex flex-col gap-1 sm:flex-row sm:flex-wrap sm:flex-1 sm:justify-end">
          {gameRoutes.map(({ path, labelKey }) => (
            <Link
              key={path}
              to={path}
              aria-current={pathname === path ? 'page' : undefined}
              onClick={() => setIsOpen(false)}
              className={`inline-block px-2 py-0.5 text-xs font-medium rounded transition-colors${pathname === path ? ' bg-gray-400 text-white' : ' bg-gray-600 text-gray-200 hover:bg-gray-500'}`}
            >
              {t(labelKey)}
            </Link>
          ))}
        </div>
        <div className="hidden sm:flex">{langToggle(currentLang, i18n, t)}</div>
      </nav>
    </div>
  );
}
