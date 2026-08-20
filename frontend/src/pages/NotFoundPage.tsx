import { useTranslation } from 'react-i18next';
import { Link, useLocation } from 'react-router-dom';
import { useDocumentTitle } from '../hooks/useDocumentTitle';
import { btnPrimary, btnSecondary, focusRingWhite } from '../styles/buttonStyles';

/**
 * 404 fallback for unknown hash routes (e.g. `#/notagame`,
 * `#/games/xyz`). Replaces the previous silent `<Navigate to="/" />`
 * redirect so users get a clear "this page doesn't exist" surface
 * with paths back to the Discover survey or the home page.
 */
export function NotFoundPage() {
  const { t } = useTranslation('common');
  useDocumentTitle(t('notFound.title'));
  const location = useLocation();
  return (
    <div className="flex-1 flex flex-col items-center justify-center px-6 py-12 gap-6 bg-ds-surface text-ds-text-primary">
      <p className="text-xs uppercase tracking-[0.18em] text-ds-accent">{t('notFound.eyebrow')}</p>
      <h1 className="font-serif text-[32px] leading-[1.15] tracking-[-0.01em] text-center">{t('notFound.title')}</h1>
      <p className="text-sm text-ds-text-muted text-center max-w-md">{t('notFound.body')}</p>
      <p className="text-xs text-ds-text-muted font-mono break-all">
        {t('notFound.requestedPath')}: {location.pathname}
      </p>
      <div className="flex flex-wrap gap-3 justify-center">
        <Link to="/discover" className={`${btnPrimary} ${focusRingWhite}`}>
          {t('notFound.actionDiscover')}
        </Link>
        <Link to="/" className={`${btnSecondary} ${focusRingWhite}`}>
          {t('notFound.actionHome')}
        </Link>
      </div>
    </div>
  );
}

export default NotFoundPage;
