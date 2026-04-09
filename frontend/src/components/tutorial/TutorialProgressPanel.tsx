import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { useTutorialProgress } from '../../hooks/useTutorialProgress';

/** Collapsible panel showing tutorial completion progress across all games. */
export function TutorialProgressPanel() {
  const { t } = useTranslation('tutorial');
  const { t: tc } = useTranslation('common');
  const { games, completedCount, totalCount } = useTutorialProgress();
  const percentage = totalCount > 0 ? Math.round((completedCount / totalCount) * 100) : 0;

  return (
    <details className="glass-panel rounded-lg px-3 py-2 mx-2 mb-2">
      <summary className="cursor-pointer text-white font-bold text-sm select-none">
        {t('progress.title', { defaultValue: 'Tutorial Progress' })} — {completedCount}/{totalCount}
      </summary>

      <div className="mt-2">
        {/* Progress bar */}
        <div
          className="w-full bg-ds-surface rounded-full h-2 mb-2"
          role="progressbar"
          aria-label={t('progress.title', { defaultValue: 'Tutorial Progress' })}
          aria-valuenow={percentage}
          aria-valuemin={0}
          aria-valuemax={100}
        >
          <div className="bg-green-500 h-2 rounded-full transition-all" style={{ width: `${percentage}%` }} />
        </div>

        {/* Game grid */}
        <div className="grid grid-cols-5 gap-1 text-xs">
          {games.map((game) => (
            <Link
              key={game.gameName}
              to={game.path}
              className={`flex items-center justify-center px-1.5 py-1 rounded hover:bg-ds-surface-elevated-hover ${game.completed ? 'text-ds-success' : 'text-ds-text-muted'}`}
              title={tc(game.labelKey)}
            >
              {game.completed ? '✓' : '○'}
            </Link>
          ))}
        </div>
      </div>
    </details>
  );
}
