import { useEffect, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { RecommendationCard } from '../components/discover/RecommendationCard';
import { StretchPickCard } from '../components/discover/StretchPickCard';
import { useGameRecommendations } from '../hooks/useGameRecommendations';
import { btnPrimary, btnSecondary, focusRingWhite } from '../styles/buttonStyles';
import { isFullyAnswered, parseSearchParams, type UserMoodInput } from '../utils/urlMoodCodec';

/** Convert UserMoodInput to a UserMood (length-2 readonly tuples). */
function toUserMood(input: UserMoodInput) {
  return {
    mood: [input.mood[0], input.mood[1]] as readonly [number | null, number | null],
    skill: [input.skill[0], input.skill[1]] as readonly [number | null, number | null],
    social: [input.social[0], input.social[1]] as readonly [number | null, number | null],
    theme: [input.theme[0], input.theme[1]] as readonly [number | null, number | null],
  };
}

/**
 * Result page. Parses the URL query into a mood, runs recommendations,
 * and renders the editor's-pick hero plus TOP3 / Stretch / Also rows.
 * If the URL is malformed or the user skipped every question, falls
 * back to a warm "もう少しヒントください" hero with alphabetical TOP3.
 */
export function DiscoverResultPage() {
  const { t } = useTranslation('discover');
  const [params] = useSearchParams();
  const navigate = useNavigate();

  const parsed = useMemo(() => parseSearchParams(params), [params]);

  useEffect(() => {
    if (parsed === null) navigate('/discover', { replace: true });
  }, [parsed, navigate]);

  const moodInput = parsed ?? {
    mood: [null, null] as const,
    skill: [null, null] as const,
    social: [null, null] as const,
    theme: [null, null] as const,
  };
  const fullySkipped = !isFullyAnswered(moodInput);
  const recs = useGameRecommendations(toUserMood(moodInput));

  if (parsed === null) return null;

  return (
    <div className="flex-1 min-h-0 flex flex-col items-center px-4 py-8 gap-6">
      <div className="w-full max-w-[600px] flex flex-col gap-8">
        {fullySkipped ? (
          <section className="flex flex-col gap-4 px-5 py-6 border-2 border-dashed border-ds-border rounded-md">
            <p className="text-xs uppercase tracking-[0.18em] text-ds-accent">{t('fallback.eyebrow')}</p>
            <h1 className="font-serif text-[28px] leading-[1.15] text-ds-text-primary">{t('fallback.title')}</h1>
            <p className="text-sm text-ds-text-muted leading-relaxed">{t('fallback.body')}</p>
            <div className="flex gap-3">
              <Link to="/discover" className={btnPrimary}>
                {t('fallback.action.retry')}
              </Link>
              <a href="#top3" className={`${btnSecondary} ${focusRingWhite}`}>
                {t('fallback.action.browse')}
              </a>
            </div>
          </section>
        ) : (
          recs.top3[0] && <RecommendationCard game={recs.top3[0].game} variant="hero" topAxis={recs.top3[0].topAxis} />
        )}

        {recs.top3.slice(1).length > 0 && (
          <section id="top3" className="flex flex-col gap-3">
            <h2 className="text-xs uppercase tracking-[0.15em] text-ds-text-muted">{t('section.top3')}</h2>
            <ul className="flex flex-col gap-2">
              {recs.top3.slice(1).map((s) => (
                <li key={s.game.path}>
                  <RecommendationCard game={s.game} variant="row" topAxis={s.topAxis} />
                </li>
              ))}
            </ul>
          </section>
        )}

        {recs.stretch && !fullySkipped && (
          <section>
            <StretchPickCard game={recs.stretch.game} />
          </section>
        )}

        {recs.also.length > 0 && (
          <section className="flex flex-col gap-3">
            <h2 className="text-xs uppercase tracking-[0.15em] text-ds-text-muted">{t('section.also')}</h2>
            <ul className="flex flex-col gap-2">
              {recs.also.map((s) => (
                <li key={s.game.path}>
                  <RecommendationCard game={s.game} variant="row" topAxis={s.topAxis} />
                </li>
              ))}
            </ul>
          </section>
        )}

        <div className="flex gap-3 mt-2">
          <Link to="/discover" className={btnSecondary}>
            {t('action.retake')}
          </Link>
        </div>
      </div>
    </div>
  );
}

export default DiscoverResultPage;
