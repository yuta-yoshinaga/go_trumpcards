import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { useDocumentTitle } from '../hooks/useDocumentTitle';
import { btnSecondary, focusRingWhite } from '../styles/buttonStyles';

/** Repository root — the authoritative home of TRADEMARKS.md and the asset READMEs. */
const REPO_URL = 'https://github.com/yuta-yoshinaga/go_trumpcards';

/** Link style shared by the outbound references in the body copy. */
const inlineLink = `text-ds-accent underline underline-offset-2 hover:text-ds-accent-hover ${focusRingWhite}`;

/** Props for {@link LegalSection}. */
interface LegalSectionProps {
  title: string;
  children: React.ReactNode;
}

/** One titled block of the notice. */
function LegalSection({ title, children }: LegalSectionProps) {
  return (
    <section className="flex flex-col gap-2">
      <h2 className="font-serif text-lg text-ds-text-primary">{title}</h2>
      <div className="text-sm text-ds-text-muted flex flex-col gap-2">{children}</div>
    </section>
  );
}

/**
 * Trademark notice and asset credits, reachable from the nav.
 *
 * This page exists because the statement has to be where players are, not only
 * in the repository: the thing a rights holder objects to is a user believing
 * the game is licensed from them, and a notice in `TRADEMARKS.md` never reaches
 * that user. It deliberately does **not** restate the mark-by-mark inventory —
 * that lives in `TRADEMARKS.md`, and a second copy here would drift out of sync
 * with it. Only the parts that must be seen are inlined.
 */
export function LegalPage() {
  const { t } = useTranslation('common');
  useDocumentTitle(t('legal.title'));
  return (
    <div className="flex-1 overflow-y-auto bg-ds-surface text-ds-text-primary">
      <div className="mx-auto max-w-2xl px-6 py-10 flex flex-col gap-8">
        <header className="flex flex-col gap-2">
          <p className="text-xs uppercase tracking-[0.18em] text-ds-accent">{t('legal.eyebrow')}</p>
          <h1 className="font-serif text-[32px] leading-[1.15] tracking-[-0.01em]">{t('legal.title')}</h1>
        </header>

        <LegalSection title={t('legal.trademarks.title')}>
          <p>{t('legal.trademarks.disclaimer')}</p>
          <p>{t('legal.trademarks.rules')}</p>
          <p>
            <a
              href={`${REPO_URL}/blob/develop/TRADEMARKS.md`}
              target="_blank"
              rel="noopener noreferrer"
              className={inlineLink}
            >
              {t('legal.trademarks.inventoryLink')}
            </a>
          </p>
        </LegalSection>

        <LegalSection title={t('legal.assets.title')}>
          <p>{t('legal.assets.cards')}</p>
          <p>{t('legal.assets.sounds')}</p>
          <p>{t('legal.assets.fonts')}</p>
        </LegalSection>

        <LegalSection title={t('legal.license.title')}>
          <p>{t('legal.license.body')}</p>
          <p>
            <a
              href={`${REPO_URL}/blob/develop/LICENSE`}
              target="_blank"
              rel="noopener noreferrer"
              className={inlineLink}
            >
              {t('legal.license.link')}
            </a>
          </p>
        </LegalSection>

        <LegalSection title={t('legal.contact.title')}>
          <p>{t('legal.contact.body')}</p>
          <p>
            <a href={`${REPO_URL}/issues`} target="_blank" rel="noopener noreferrer" className={inlineLink}>
              {t('legal.contact.link')}
            </a>
          </p>
        </LegalSection>

        <div>
          <Link to="/" className={`${btnSecondary} ${focusRingWhite}`}>
            {t('legal.backHome')}
          </Link>
        </div>
      </div>
    </div>
  );
}

export default LegalPage;
