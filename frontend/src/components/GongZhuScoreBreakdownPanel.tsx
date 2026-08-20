import { useTranslation } from 'react-i18next';
import type { GongZhuScoreBreakdown } from '../types/games/gongzhu';

/** Props for {@link GongZhuScoreBreakdownPanel}. */
export interface GongZhuScoreBreakdownPanelProps {
  /** One entry per player, in seat order. */
  breakdowns: GongZhuScoreBreakdown[];
  /** Renders a player's display name for a seat. */
  playerName: (idx: number) => string;
}

/**
 * Shows how each player's round score was built up.
 *
 * **起きていない項目は書かない。**取っていない猪の行が出ると、何が起きたのか
 * 読み取れなくなる。値はドメインの採点関数そのものが返したものなので、ここの
 * 説明と実際に付いた点が食い違うことはない (#5630)。
 */
export function GongZhuScoreBreakdownPanel({ breakdowns, playerName }: GongZhuScoreBreakdownPanelProps) {
  const { t } = useTranslation('gongzhu');

  return (
    <div className="mb-2 p-2 rounded bg-black/30 text-sm" data-testid="gz-score-breakdown">
      <div className="text-ds-text-primary font-bold mb-1">{t('breakdownTitle')}</div>
      {breakdowns.map((b, idx) => (
        <div key={playerName(idx)} className="mb-1">
          <div className="text-ds-text-primary">{playerName(idx)}</div>
          <ul className="text-ds-text-muted ml-3">
            {b.heartCount > 0 && <li>{t('breakdownHearts', { count: b.heartCount, sum: b.heartsSum })}</li>}
            {b.allHearts && <li>{t('breakdownAllHearts')}</li>}
            {b.aceExposed && b.heartCount > 0 && <li>{t('breakdownAceExposed')}</li>}
            {b.hasPig && <li>{t(b.pigExposed ? 'breakdownPigExposed' : 'breakdownPig')}</li>}
            {b.hasSheep && <li>{t(b.sheepExposed ? 'breakdownSheepExposed' : 'breakdownSheep')}</li>}
            {b.doublerMultiplier > 0 && (
              <>
                <li>{t('breakdownSubtotal', { subtotal: b.subtotal })}</li>
                <li>{t('breakdownDoubler', { mult: b.doublerMultiplier })}</li>
              </>
            )}
            {b.doublerStandalone !== 0 && <li>{t('breakdownDoublerStandalone', { points: b.doublerStandalone })}</li>}
            <li className="text-ds-text-primary">{t('breakdownTotal', { total: b.total })}</li>
          </ul>
        </div>
      ))}
    </div>
  );
}
