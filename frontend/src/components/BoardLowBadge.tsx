import { badgeInfoColors, badgeSuccessColors } from '../styles/badgeStyles';
import type { Card } from '../types/card';
import { type BoardLowStatus, boardLowPossibility } from '../utils/omahaLowCards';

/** Design-system token classes per board-low status (live=success, possible=info, impossible=muted). */
const BOARD_LOW_STATUS_CLASS: Readonly<Record<BoardLowStatus, string>> = {
  live: badgeSuccessColors,
  possible: badgeInfoColors,
  impossible: 'border-ds-border bg-ds-surface text-ds-text-muted',
};

/** Props for {@link BoardLowBadge}. */
export interface BoardLowBadgeProps {
  /** Community cards to inspect. Only the board matters — hole cards are irrelevant. */
  communityCards: Card[];
  /** Translator bound to the calling game's namespace; it must own the `boardLow.*` keys. */
  t: (key: string, opts?: Record<string, unknown>) => string;
  /** Test hook. Defaults to the Omaha Hi-Lo id the badge was first written for. */
  testId?: string;
}

/**
 * Additive badge showing whether the community board can still make a qualifying
 * Hi-Lo low (8-or-better): 3+ distinct low ranks on board = live, still
 * reachable = possible, mathematically out = impossible. Inspects the board only.
 */
export function BoardLowBadge({ communityCards, t, testId = 'omahahilo-board-low-badge' }: BoardLowBadgeProps) {
  const { status, needed } = boardLowPossibility(communityCards);
  const aria =
    status === 'live'
      ? t('boardLow.ariaLive')
      : status === 'possible'
        ? t('boardLow.ariaPossible', { needed })
        : t('boardLow.ariaImpossible');
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] font-semibold ${BOARD_LOW_STATUS_CLASS[status]}`}
      data-testid={testId}
      data-status={status}
      title={aria}
    >
      <span aria-hidden="true">{t('boardLow.label')}:</span>
      <span aria-hidden="true">{t(`boardLow.${status}`)}</span>
      <span className="sr-only">{aria}</span>
    </span>
  );
}
