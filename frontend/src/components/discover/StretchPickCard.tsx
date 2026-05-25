import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import type { GameRoute } from '../../constants/gameRoutes';
import { btnSecondary, focusRingWhite } from '../../styles/buttonStyles';

/** Props for the "ちょっと挑戦してみる？" Stretch Pick card. */
export interface StretchPickCardProps {
  readonly game: GameRoute;
}

/**
 * "Ever-so-slightly off-piste" recommendation. Dashed gold border + warm
 * gradient background, drawing the eye without competing with the hero.
 *
 * The blurb prefers a stretch-specific copy (`discover.stretch.<slug>`) if
 * the locale provides one, and falls back to the regular game blurb
 * (`discover.blurb.<slug>`) otherwise. Every game has a blurb entry so the
 * fallback always resolves; per-slug stretch copy can be added later
 * without changing this component.
 */
export function StretchPickCard({ game }: StretchPickCardProps) {
  const { t } = useTranslation('discover');
  const labelT = useTranslation('common').t;
  const slug = game.page.toLowerCase();
  const blurbKey = `stretch.${slug}`;
  const blurbFallback = t(`blurb.${slug}`);
  const gameName = labelT(game.labelKey);

  return (
    <article className="flex flex-col gap-3 px-4 py-4 rounded-md border-2 border-dashed border-ds-accent/40 bg-gradient-to-r from-[rgba(212,168,83,0.08)] to-[rgba(212,168,83,0.02)]">
      <p className="text-xs uppercase tracking-[0.15em] text-ds-accent">{t('stretch.eyebrow')}</p>
      <h2 className="font-serif text-lg text-ds-text-primary">
        <span aria-hidden="true" className="mr-2">
          {game.icon}
        </span>
        {gameName}
      </h2>
      <p className="text-sm text-ds-text-muted leading-relaxed">{t(blurbKey, { defaultValue: blurbFallback })}</p>
      <Link to={game.path} className={`${btnSecondary} self-start ${focusRingWhite}`}>
        {t('stretch.action')}
      </Link>
    </article>
  );
}
