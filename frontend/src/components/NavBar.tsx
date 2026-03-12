import { useTranslation } from 'react-i18next';
import { Link, useLocation } from 'react-router-dom';
import { gameRoutes } from '../constants/gameRoutes';

export function NavBar() {
  const { pathname } = useLocation();
  const { t, i18n } = useTranslation('common');
  const currentLang = i18n.language;

  return (
    <nav style={{ textAlign: 'right', margin: '8px 10px' }} className="flex items-center justify-end gap-2 flex-wrap">
      <div className="flex flex-wrap gap-1 flex-1 justify-end">
        {gameRoutes.map(({ path, labelKey }) => (
          <Link
            key={path}
            to={path}
            aria-current={pathname === path ? 'page' : undefined}
            className={`inline-block px-2 py-0.5 text-xs font-medium rounded transition-colors${pathname === path ? ' bg-gray-400 text-white' : ' bg-gray-600 text-gray-200 hover:bg-gray-500'}`}
          >
            {t(labelKey)}
          </Link>
        ))}
      </div>
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
    </nav>
  );
}
