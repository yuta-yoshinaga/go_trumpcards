import { useTranslation } from 'react-i18next';
import { useCardDimensions } from '../../hooks/useCardDimensions';
import { suitName, valueName } from '../../utils/cardUtils';
import { isPositionPlaced, isPositionPlayable, SUITS } from '../../utils/sevensUtils';

interface BoardProps {
  tablePlaced: number[];
  tunnelEnabled: boolean;
  tunnelSkipWidth: number;
  endStopEnabled: boolean;
  jokerSelecting: boolean;
  onJokerPlace?: (suit: number, value: number) => void;
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
  const { sevensCellSize, sevensFontSize } = useCardDimensions();
  return (
    <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2">
      <div className="text-white font-bold mb-2">
        {t('board')}
        {tunnelEnabled && <span className="text-yellow-400 text-xs ml-2">{t('tunnelTag')}</span>}
        {jokerSelecting && <span className="text-green-400 text-xs ml-2">{t('jokerSelectHint')}</span>}
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        {SUITS.map(({ idx, name, label, color }) => (
          <div key={name} className="bg-white/[0.08] rounded-lg py-1.5 px-2.5 flex items-center gap-2">
            <span className="min-w-[18px]" style={{ color, fontWeight: 'bold', fontSize: '1.1em' }}>
              {label}
            </span>
            <div className="flex flex-wrap gap-[3px] items-center">
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
                const cellStyle: React.CSSProperties = {
                  display: 'inline-block',
                  width: sevensCellSize,
                  height: sevensCellSize,
                  lineHeight: `${sevensCellSize}px`,
                  textAlign: 'center',
                  borderRadius: 4,
                  fontSize: sevensFontSize,
                  fontWeight: isCenter ? 'bold' : 'normal',
                  background: canPlace
                    ? 'var(--color-blue-500)'
                    : placed
                      ? isCenter
                        ? 'var(--color-game-status-waiting)'
                        : 'var(--color-game-status-active)'
                      : 'rgba(255,255,255,0.1)',
                  color: canPlace ? 'white' : placed ? 'black' : 'var(--color-board-cell-empty-text)',
                  boxSizing: 'border-box',
                };
                if (canPlace) {
                  return (
                    <button
                      key={v}
                      type="button"
                      onClick={() => onJokerPlace?.(idx, v)}
                      aria-label={t('placeAriaLabel', { suit: suitName(idx), value: valueName(v) })}
                      className="border border-blue-400"
                      style={{ ...cellStyle, cursor: 'pointer', padding: 0 }}
                    >
                      {valueName(v)}
                    </button>
                  );
                }
                return (
                  <span key={v} className={tunnelHighlight ? 'border border-amber-400' : ''} style={cellStyle}>
                    {valueName(v)}
                  </span>
                );
              })}
              {tunnelEnabled && (
                <span role="img" className="text-yellow-400 text-xs ml-0.5" aria-label={t('tunnelConnection')}>
                  ↔
                </span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export { Board as SevensBoard };
