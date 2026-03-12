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

export function DaifugoRulesBadges({ state }: { state: DaifugoResponse }) {
  const { t } = useTranslation('daifugo');
  const badges: { label: string; bg: string; color: string }[] = [];
  if (state.revolutionActive) {
    badges.push({ label: t('badge.revolution'), bg: '#d9534f', color: '#fff' });
  }
  if (state.elevenBackActive) {
    badges.push({
      label: t('badge.elevenBack'),
      bg: 'var(--color-game-status-waiting)',
      color: 'var(--color-game-text-strong)',
    });
  }
  if (state.suitLocked) {
    const modeSuffix = state.config.suitLockMode === 1 ? ` (${t('badge.suitLockPartial')})` : '';
    badges.push({
      label: t('badge.suitLock', { suit: state.lockedSuit }) + modeSuffix,
      bg: '#5bc0de',
      color: 'var(--color-game-text-strong)',
    });
  }
  if (state.tableIsSequence) {
    badges.push({ label: t('badge.sequence'), bg: '#9b59b6', color: '#fff' });
  }
  if (state.reverseDirection) {
    badges.push({ label: t('badge.nineReverse'), bg: '#e67e22', color: '#fff' });
  }
  if (state.numberLocked) {
    badges.push({ label: t('badge.numberLock'), bg: '#1abc9c', color: '#fff' });
  }
  if (badges.length === 0) return null;
  return (
    <div className="my-1 px-1">
      {badges.map((b) => (
        <span key={b.label} style={{ ...badgeStyle, background: b.bg, color: b.color }}>
          {b.label}
        </span>
      ))}
    </div>
  );
}
