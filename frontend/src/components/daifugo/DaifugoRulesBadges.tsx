import { useTranslation } from 'react-i18next';
import type { DaifugoResponse } from '../../types/card';

const badgeStyle: React.CSSProperties = {
  display: 'inline-block',
  borderRadius: 6,
  padding: '2px 10px',
  marginRight: 6,
  marginBottom: 4,
  fontSize: '0.8em',
  fontWeight: 'bold',
};

/** Renders active rule badges (revolution, suit lock, eleven back, etc.) for Daifugo. */
export function DaifugoRulesBadges({ state }: { state: DaifugoResponse }) {
  const { t } = useTranslation('daifugo');
  const badges: { label: string; description: string; bg: string; color: string }[] = [];
  if (state.revolutionActive) {
    badges.push({
      label: t('badge.revolution'),
      description: t('badge.revolutionDesc'),
      bg: 'var(--color-daifugo-revolution)',
      color: 'white',
    });
  }
  if (state.elevenBackActive) {
    badges.push({
      label: t('badge.elevenBack'),
      description: t('badge.elevenBackDesc'),
      bg: 'var(--color-game-status-waiting)',
      color: 'var(--color-game-text-strong)',
    });
  }
  if (state.suitLocked) {
    const modeSuffix = state.config.suitLockMode === 1 ? ` (${t('badge.suitLockPartial')})` : '';
    badges.push({
      label: t('badge.suitLock', { suit: state.lockedSuit }) + modeSuffix,
      description: t('badge.suitLockDesc'),
      bg: 'var(--color-daifugo-suit-lock)',
      color: 'var(--color-game-text-strong)',
    });
  }
  if (state.tableIsSequence) {
    badges.push({
      label: t('badge.sequence'),
      description: t('badge.sequenceDesc'),
      bg: 'var(--color-daifugo-sequence)',
      color: 'white',
    });
  }
  if (state.reverseDirection) {
    badges.push({
      label: t('badge.nineReverse'),
      description: t('badge.nineReverseDesc'),
      bg: 'var(--color-daifugo-nine-reverse)',
      color: 'white',
    });
  }
  if (state.numberLocked) {
    badges.push({
      label: t('badge.numberLock'),
      description: t('badge.numberLockDesc'),
      bg: 'var(--color-daifugo-number-lock)',
      color: 'white',
    });
  }
  if (state.sequenceLocked) {
    badges.push({
      label: t('badge.sequenceLock'),
      description: t('badge.sequenceLockDesc'),
      bg: 'var(--color-daifugo-sequence)',
      color: 'white',
    });
  }
  if (badges.length === 0) return null;
  return (
    <div className="my-1 px-1">
      {badges.map((b) => (
        <span
          key={b.label}
          className="relative group/badge inline-block cursor-help"
          style={{ ...badgeStyle, background: b.bg, color: b.color }}
          role="img"
          aria-label={`${b.label}: ${b.description}`}
        >
          {b.label}
          <span
            role="tooltip"
            className="hidden group-hover/badge:block group-focus/badge:block absolute bottom-full left-1/2 -translate-x-1/2 mb-1 bg-gray-900 text-white text-xs rounded px-2 py-1 whitespace-nowrap z-10 font-normal"
          >
            {b.description}
          </span>
        </span>
      ))}
    </div>
  );
}
