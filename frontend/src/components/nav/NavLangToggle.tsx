import { useTranslation } from 'react-i18next';

/** Size variant — `'md'` matches the NavBar (44 px touch target), `'sm'` matches the desktop sidebar. */
export type NavLangToggleSize = 'sm' | 'md';

/** Props for {@link NavLangToggle}. */
export interface NavLangToggleProps {
  /** Visual size; defaults to `'md'` to preserve mobile/tablet touch target. */
  size?: NavLangToggleSize;
}

const SIZE_CLASS: Record<NavLangToggleSize, string> = {
  sm: 'px-2 py-1',
  md: 'px-3 py-2 min-h-[44px]',
};

/** JA / EN language toggle pair, shared between NavBar and DesktopSidebar. */
export function NavLangToggle({ size = 'md' }: NavLangToggleProps) {
  const { t, i18n } = useTranslation('common');
  const currentLang = i18n.language;
  const sizeCls = SIZE_CLASS[size];
  return (
    <div className="flex gap-0.5">
      <button
        type="button"
        aria-label={t('nav.switchToJa')}
        aria-pressed={currentLang === 'ja'}
        onClick={() => i18n.changeLanguage('ja')}
        className={`${sizeCls} text-xs font-bold rounded-l transition-colors ${currentLang === 'ja' ? 'bg-ds-accent text-ds-text-on-accent' : 'bg-ds-surface-elevated text-ds-text-primary hover:bg-ds-surface-elevated-hover'}`}
      >
        JA
      </button>
      <button
        type="button"
        aria-label={t('nav.switchToEn')}
        aria-pressed={currentLang === 'en'}
        onClick={() => i18n.changeLanguage('en')}
        className={`${sizeCls} text-xs font-bold rounded-r transition-colors ${currentLang === 'en' ? 'bg-ds-accent text-ds-text-on-accent' : 'bg-ds-surface-elevated text-ds-text-primary hover:bg-ds-surface-elevated-hover'}`}
      >
        EN
      </button>
    </div>
  );
}
