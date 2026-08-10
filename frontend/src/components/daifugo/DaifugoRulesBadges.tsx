import { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useIsMobile } from '../../hooks/useCardDimensions';
import type { DaifugoResponse } from '../../types/card';
import { Modal } from '../common/Modal';

const badgeClass = 'inline-block rounded-[6px] px-2.5 py-0.5 mr-1.5 mb-1 text-xs font-bold';

interface BadgeData {
  label: string;
  description: string;
  bg: string;
  color: string;
}

/** Builds badge data from the current Daifugo state. */
function buildBadges(state: DaifugoResponse, t: (key: string, opts?: Record<string, unknown>) => string): BadgeData[] {
  const badges: BadgeData[] = [];
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
  return badges;
}

/** Modal that displays active rule badges with their descriptions. */
function RulesBadgeModal({
  badges,
  onClose,
  summaryLabel,
  closeLabel,
}: {
  badges: BadgeData[];
  onClose: () => void;
  summaryLabel: string;
  closeLabel: string;
}) {
  return (
    <Modal
      open
      onClose={onClose}
      ariaLabelledBy="rules-modal-title"
      panelClassName="glass-panel rounded-t-lg sm:rounded-lg shadow-xl p-4 w-full sm:max-w-sm sm:mx-4 max-h-[70vh] overflow-y-auto"
      backdropClassName="items-end sm:items-center justify-center bg-black/50"
    >
      <div className="flex justify-between items-center mb-3">
        <h2 id="rules-modal-title" className="text-sm font-bold text-ds-text-primary">
          {summaryLabel}
        </h2>
        <button
          type="button"
          onClick={onClose}
          className="text-ds-text-primary/60 hover:text-ds-text-primary text-lg leading-none"
          aria-label={closeLabel}
        >
          ✕
        </button>
      </div>
      <ul className="space-y-2">
        {badges.map((b) => (
          <li key={b.label} className="flex items-start gap-2">
            <span
              className="shrink-0 rounded px-2 py-0.5 text-xs font-bold"
              style={{ background: b.bg, color: b.color }}
            >
              {b.label}
            </span>
            <span className="text-xs text-ds-text-primary/80">{b.description}</span>
          </li>
        ))}
      </ul>
    </Modal>
  );
}

/** Renders active rule badges (revolution, suit lock, eleven back, etc.) for Daifugo. */
export function DaifugoRulesBadges({ state }: { state: DaifugoResponse }) {
  const { t } = useTranslation('daifugo');
  const { t: tc } = useTranslation('common');
  const isMobile = useIsMobile();
  const [modalOpen, setModalOpen] = useState(false);
  const badges = buildBadges(state, t);

  const openModal = useCallback(() => setModalOpen(true), []);
  const closeModal = useCallback(() => setModalOpen(false), []);

  if (badges.length === 0) return null;

  const summaryLabel = t('badge.activeRulesSummary', { count: badges.length });

  // Mobile: show collapsed summary button
  if (isMobile) {
    return (
      <div className="my-1 px-1">
        <button
          type="button"
          className="inline-flex items-center gap-1 rounded-md px-3 py-1 text-xs font-bold bg-ds-warning text-white"
          onClick={openModal}
          data-testid="rules-summary-button"
        >
          <span aria-hidden="true">⚠️</span>
          {summaryLabel}
        </button>
        {modalOpen && (
          <RulesBadgeModal
            badges={badges}
            onClose={closeModal}
            summaryLabel={summaryLabel}
            closeLabel={tc('manual.close')}
          />
        )}
      </div>
    );
  }

  // Desktop: show individual badges with tooltips
  return (
    <div className="my-1 px-1">
      {badges.map((b) => (
        <button
          key={b.label}
          type="button"
          className={`${badgeClass} relative group/badge cursor-help border-none`}
          style={{ background: b.bg, color: b.color }}
          aria-label={`${b.label}: ${b.description}`}
        >
          {b.label}
          <span
            role="tooltip"
            className="hidden group-hover/badge:block group-focus/badge:block absolute bottom-full left-1/2 -translate-x-1/2 mb-1 bg-ds-surface-elevated text-ds-text-primary text-xs rounded px-2 py-1 whitespace-nowrap z-10 font-normal"
          >
            {b.description}
          </span>
        </button>
      ))}
    </div>
  );
}
