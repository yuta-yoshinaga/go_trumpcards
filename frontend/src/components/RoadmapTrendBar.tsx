import { useTranslation } from 'react-i18next';

/** Props for the RoadmapTrendBar component. */
export interface RoadmapTrendBarProps {
  /**
   * Outcome history. Each element must be a numeric side code. Any value not
   * matching `leftCode` or `rightCode` is treated as neutral (e.g. Tie) and
   * does not break the streak.
   */
  history: readonly number[];
  /** Numeric code for the "left" side (e.g. Player / Dragon). */
  leftCode: number;
  /** Numeric code for the "right" side (e.g. Banker / Tiger). */
  rightCode: number;
  /** Localized labels for the two sides. */
  leftLabel: string;
  rightLabel: string;
  /** How many recent outcomes to include in the win-rate split. Default 12. */
  lookback?: number;
  /** Test id for the root element. */
  testId?: string;
}

/**
 * Renders a horizontal split bar showing each side's win share over the last
 * `lookback` non-neutral outcomes, plus a streak badge (with a 🔥 indicator
 * when the streak reaches 3+). Returns null when the history is empty.
 */
export function RoadmapTrendBar({
  history,
  leftCode,
  rightCode,
  leftLabel,
  rightLabel,
  lookback = 12,
  testId,
}: RoadmapTrendBarProps) {
  const { t } = useTranslation('common');
  if (history.length === 0) return null;

  // Streak: trailing run of equal sides (ignoring neutral outcomes).
  let streakSide: number | null = null;
  let streakCount = 0;
  for (let i = history.length - 1; i >= 0; i -= 1) {
    const r = history[i];
    if (r !== leftCode && r !== rightCode) continue;
    if (streakSide === null) {
      streakSide = r;
      streakCount = 1;
    } else if (r === streakSide) {
      streakCount += 1;
    } else {
      break;
    }
  }

  const recent: number[] = [];
  for (let i = history.length - 1; i >= 0 && recent.length < lookback; i -= 1) {
    const r = history[i];
    if (r === leftCode || r === rightCode) recent.push(r);
  }
  const sideOnly = recent.length;
  const leftWins = recent.filter((r) => r === leftCode).length;
  const leftPct = sideOnly === 0 ? 50 : Math.round((leftWins / sideOnly) * 100);
  const rightPct = 100 - leftPct;
  const streakLabel = streakSide === leftCode ? leftLabel : streakSide === rightCode ? rightLabel : null;
  const showFire = streakCount >= 3 && streakLabel !== null;

  return (
    <div className="mb-2" data-testid={testId ?? 'roadmap-trend-bar'}>
      <div className="mb-1 flex items-center justify-between text-[10px] text-ds-text-muted">
        <span>
          {leftLabel} {leftPct}%
        </span>
        {showFire && streakLabel ? (
          <span className="font-bold text-ds-warning" data-testid="roadmap-trend-streak">
            🔥 {streakLabel} {t('roadmapTrend.streak', { count: streakCount, defaultValue: '{{count}} in a row' })}
          </span>
        ) : (
          <span>
            {t('roadmapTrend.lookback', {
              count: sideOnly,
              defaultValue: 'Last {{count}}',
            })}
          </span>
        )}
        <span>
          {rightPct}% {rightLabel}
        </span>
      </div>
      <div className="flex h-1.5 overflow-hidden rounded bg-black/30">
        <div className="bg-ds-info" style={{ width: `${leftPct}%` }} aria-hidden="true" />
        <div className="bg-ds-error" style={{ width: `${rightPct}%` }} aria-hidden="true" />
      </div>
    </div>
  );
}
