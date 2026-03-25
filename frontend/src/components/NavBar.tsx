import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useLocation } from 'react-router-dom';
import { gameCategories } from '../constants/gameRoutes';
import { SITE_NAME } from '../constants/site';
import { useIsMediumDesktop } from '../hooks/useCardDimensions';
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
  const isMediumDesktop = useIsMediumDesktop();
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
          {SITE_NAME}
        </Link>
        <div className="flex items-center gap-2">
          {soundMuted !== undefined && onSoundToggle && <SoundToggle muted={soundMuted} onToggle={onSoundToggle} />}
          {langToggle(currentLang, i18n, t)}
          <button
            ref={toggleRef}
            type="button"
            onClick={() => setIsOpen(!isOpen)}
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
        <div className="flex flex-col gap-1 sm:flex-row sm:flex-wrap sm:flex-1 sm:justify-end sm:gap-3 lg:flex-nowrap lg:gap-1">
          {gameCategories.map(({ labelKey, icon: catIcon, routes }) => {
            const hasActivePage = routes.some(({ path }) => path === pathname);
            return (
              <details
                key={labelKey}
                className="nav-category sm:flex sm:items-center"
                open={hasActivePage}
                onToggle={(e) => {
                  // On medium desktop (sm-lg), force details to stay open
                  if (isMediumDesktop && !e.currentTarget.open) {
                    e.currentTarget.open = true;
                  }
                }}
              >
                <summary className="text-gray-300 text-xs uppercase tracking-wider px-1 py-2 cursor-pointer select-none min-h-[44px] flex items-center gap-1 sm:cursor-default sm:py-0 sm:min-h-0 lg:cursor-pointer lg:py-2 lg:min-h-[44px] lg:hover:text-white lg:transition-colors shrink-0">
                  <span aria-hidden="true">{catIcon}</span> {t(labelKey)}
                </summary>
                <div className="nav-dropdown flex flex-col gap-1 pl-2 pb-1 sm:flex-row sm:pl-0 sm:pb-0 lg:flex-col">
                  {routes.map(({ path, labelKey: routeLabel, icon }) => (
                    <Link
                      key={path}
                      to={path}
                      aria-current={pathname === path ? 'page' : undefined}
                      onClick={() => setIsOpen(false)}
                      className={`inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded min-h-[44px] transition-[colors,box-shadow] duration-150${pathname === path ? ' bg-blue-600 text-white' : ' bg-gray-600 text-gray-200 hover:bg-gray-500 hover:shadow-md'}`}
                    >
                      <span aria-hidden="true">{icon}</span>
                      {t(routeLabel)}
                    </Link>
                  ))}
                </div>
              </details>
            );
          })}
        </div>
        <TutorialProgressPanel />
        <div className="hidden sm:flex sm:items-center sm:gap-2">
          {soundMuted !== undefined && onSoundToggle && <SoundToggle muted={soundMuted} onToggle={onSoundToggle} />}
          {langToggle(currentLang, i18n, t)}
        </div>
      </nav>
    </div>
  );
}
