import { useTranslation } from 'react-i18next';
import { useIsMobile } from '../../hooks/useCardDimensions';
import { suitName, valueName } from '../../utils/cardUtils';
import { isPositionPlaced, isPositionPlayable, SUITS } from '../../utils/sevensUtils';
import { ScrollFadeHint } from '../ScrollFadeHint';

/** Props for {@link SevensBoard}. */
export interface BoardProps {
  tablePlaced: number[];
  tunnelEnabled: boolean;
  tunnelSkipWidth: number;
  endStopEnabled: boolean;
  jokerSelecting: boolean;
  onJokerPlace?: (suit: number, value: number) => void;
}

/** Cell background and text color style based on card placement state. */
function cellColors(placed: boolean, isCenter: boolean, canPlace: boolean): React.CSSProperties {
  if (canPlace) return { background: 'var(--color-ds-info)', color: 'white' };
  if (placed)
    return isCenter
      ? { background: 'var(--color-game-status-waiting)', color: 'var(--color-game-text-strong)' }
      : { background: 'var(--color-game-status-active)', color: 'var(--color-game-text-strong)' };
  return { background: 'rgba(255,255,255,0.1)', color: 'var(--color-board-cell-empty-text)' };
}

function Board({
  tablePlaced,
  tunnelEnabled,
  tunnelSkipWidth,
  endStopEnabled,
  jokerSelecting,
  onJokerPlace,
}: BoardProps) {
  const { t } = useTranslation('sevens');
  const isMobile = useIsMobile();
  const gridTemplateColumns = tunnelEnabled ? 'auto 16px repeat(13, 1fr) 16px' : 'auto repeat(13, 1fr)';
  return (
    <div className="bg-black/30 rounded-[10px] py-2 px-2 sm:px-3.5 my-2">
      <div className="text-ds-text-primary font-bold mb-1.5 text-sm">
        {t('board')}
        {tunnelEnabled && <span className="text-ds-warning text-xs ml-2">{t('tunnelTag')}</span>}
        {jokerSelecting && <span className="text-ds-success text-xs ml-2">{t('jokerSelectHint')}</span>}
      </div>
      <div className="relative">
        <div className="overflow-x-auto">
          <div
            className="grid gap-y-1 gap-x-0.5"
            style={{ minWidth: '480px', gridTemplateColumns }}
            data-testid="sevens-grid"
          >
            {SUITS.map(({ idx, name, label, color }) => (
              <div key={name} className="contents">
                <span
                  className="flex items-center justify-center font-bold text-base sm:text-lg"
                  style={{ color }}
                  aria-hidden="true"
                >
                  {label}
                  {tunnelEnabled && (
                    <span role="img" className="text-[8px] ml-0.5" style={{ color }} aria-label={t('tunnelConnection')}>
                      ↔
                    </span>
                  )}
                </span>
                {tunnelEnabled && (
                  <span
                    role="img"
                    aria-label={t('tunnelPortalLeft')}
                    className="flex items-center justify-center text-xs motion-safe:animate-pulse"
                    data-testid="tunnel-portal-left"
                    style={{ color }}
                  >
                    ◉
                  </span>
                )}
                {Array.from({ length: 13 }, (_, i) => i + 1).map((v) => {
                  const placed = isPositionPlaced(tablePlaced, idx, v);
                  const isCenter = v === 7;
                  const canPlace =
                    jokerSelecting &&
                    isPositionPlayable(tablePlaced, idx, v, tunnelEnabled, endStopEnabled, tunnelSkipWidth);
                  const tunnelHighlight =
                    tunnelEnabled &&
                    !placed &&
                    ((v === 1 && isPositionPlaced(tablePlaced, idx, 13)) ||
                      (v === 13 && isPositionPlaced(tablePlaced, idx, 1)));
                  const colors = cellColors(placed, isCenter, canPlace);
                  const baseClass =
                    'rounded text-center text-[0.6rem] sm:text-xs lg:text-sm leading-none aspect-square flex items-center justify-center';
                  const bold = isCenter ? ' font-bold' : '';
                  const border = tunnelHighlight ? ' border border-ds-warning' : '';

                  if (canPlace) {
                    return (
                      <button
                        key={v}
                        type="button"
                        onClick={() => onJokerPlace?.(idx, v)}
                        aria-label={t('placeAriaLabel', { suit: suitName(idx), value: valueName(v) })}
                        className={`${baseClass}${bold} relative ring-2 ring-ds-info ring-offset-1 ring-offset-black/30 motion-safe:animate-pulse cursor-pointer p-0 hover:brightness-110`}
                        style={colors}
                        data-testid="board-cell"
                        data-joker-placeable="true"
                      >
                        {/* Translucent joker glyph hints that this cell will accept the held joker.
                            The rank label is rendered after this span so it stacks on top in DOM
                            order, keeping the value legible at full strength. */}
                        <span
                          aria-hidden="true"
                          className="absolute inset-0 flex items-center justify-center text-[0.85em] opacity-40"
                        >
                          🃏
                        </span>
                        <span className="relative">{valueName(v)}</span>
                      </button>
                    );
                  }
                  return (
                    <span key={v} className={`${baseClass}${bold}${border}`} style={colors} data-testid="board-cell">
                      {valueName(v)}
                    </span>
                  );
                })}
                {tunnelEnabled && (
                  <span
                    role="img"
                    aria-label={t('tunnelPortalRight')}
                    className="flex items-center justify-center text-xs motion-safe:animate-pulse"
                    data-testid="tunnel-portal-right"
                    style={{ color }}
                  >
                    ◉
                  </span>
                )}
              </div>
            ))}
          </div>
        </div>
        {isMobile && <ScrollFadeHint />}
      </div>
    </div>
  );
}

/** Renders the Sevens game board showing placed cards and playable positions. */
export { Board as SevensBoard };
