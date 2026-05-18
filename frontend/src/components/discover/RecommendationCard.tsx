import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import type { AxisKey } from '../../constants/discoverAxes';
import type { GameRoute } from '../../constants/gameRoutes';
import { btnPrimary, focusRingWhite } from '../../styles/buttonStyles';

/** Visual variant for the card — hero (TOP1) or compact row (TOP2/3, also-rans). */
export type RecommendationVariant = 'hero' | 'row';

/** Props for one recommended game card. */
export interface RecommendationCardProps {
  readonly game: GameRoute;
  readonly variant: RecommendationVariant;
  /** Axis the matcher considered most dominant — used as a small label chip. */
  readonly topAxis: AxisKey;
}

/**
 * Renders a single recommended game with an editor's-pick aesthetic.
 *
 * Hero variant: 32px Fraunces title + DM Sans blurb + accent CTA.
 * Row variant: compact icon + title + 1-line desc.
 *
 * The blurb is pulled from the lazy-loaded `discover` namespace under
 * `discover.blurb.<page-kebab>` (falls back to the i18n key string if
 * the lazy bundle has not loaded yet).
 */
export function RecommendationCard({ game, variant, topAxis }: RecommendationCardProps) {
  const { t } = useTranslation('discover');
  const labelT = useTranslation('common').t;
  const slug = game.page.toLowerCase();
  const blurbKey = `blurb.${slug}`;
  const gameName = labelT(game.labelKey);
  const axisLabel = t(`axis.${topAxis}.label`);

  if (variant === 'hero') {
    return (
      <article className="flex flex-col gap-4">
        <p className="text-xs uppercase tracking-[0.18em] text-ds-accent">
          {t('hero.eyebrow')} — {axisLabel}
        </p>
        <h1 className="font-serif text-[32px] leading-[1.15] tracking-[-0.01em] text-ds-text-primary">
          <span aria-hidden="true" className="mr-2">
            {game.icon}
          </span>
          {gameName}
        </h1>
        <p className="text-sm text-ds-text-muted leading-relaxed">{t(blurbKey)}</p>
        <Link to={game.path} className={`${btnPrimary} self-start`}>
          {t('action.play')}
        </Link>
      </article>
    );
  }

  return (
    <Link
      to={game.path}
      className={`flex items-center gap-3 px-3 py-3 rounded-md border border-ds-border bg-ds-surface-elevated hover:bg-ds-surface-elevated-hover ${focusRingWhite}`}
    >
      <span aria-hidden="true" className="text-2xl shrink-0">
        {game.icon}
      </span>
      <div className="flex flex-col min-w-0">
        <span className="text-sm font-medium text-ds-text-primary truncate">{gameName}</span>
        <span className="text-xs text-ds-text-muted truncate">{t(blurbKey)}</span>
      </div>
    </Link>
  );
}
