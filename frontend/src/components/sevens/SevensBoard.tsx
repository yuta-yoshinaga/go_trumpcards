import { useTranslation } from 'react-i18next';
import { suitName, valueName } from '../../utils/cardUtils';
import { isPositionPlaced, isPositionPlayable, SUITS } from '../../utils/sevensUtils';

interface BoardProps {
  tablePlaced: number[];
  tunnelEnabled: boolean;
  endStopEnabled: boolean;
  jokerSelecting: boolean;
  onJokerPlace?: (suit: number, value: number) => void;
}

function Board({ tablePlaced, tunnelEnabled, endStopEnabled, jokerSelecting, onJokerPlace }: BoardProps) {
  const { t } = useTranslation('sevens');
  return (
    <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2">
      <div className="text-white font-bold mb-2">
        {t('board')}
        {tunnelEnabled && <span className="text-yellow-400 text-xs ml-2">{t('tunnelTag')}</span>}
        {jokerSelecting && <span className="text-green-400 text-xs ml-2">{t('jokerSelectHint')}</span>}
      </div>
      <div className="grid grid-cols-2 gap-2">
        {SUITS.map(({ idx, name, label, color }) => (
          <div key={name} className="bg-white/[0.08] rounded-lg py-1.5 px-2.5 flex items-center gap-2">
            <span style={{ color, fontWeight: 'bold', fontSize: '1.1em', minWidth: 18 }}>{label}</span>
            <div className="flex flex-wrap gap-[3px] items-center">
              {Array.from({ length: 13 }, (_, i) => i + 1).map((v) => {
                const placed = isPositionPlaced(tablePlaced, idx, v);
                const isCenter = v === 7;
                const canPlace =
                  jokerSelecting && isPositionPlayable(tablePlaced, idx, v, tunnelEnabled, endStopEnabled);
                const tunnelHighlight =
                  tunnelEnabled &&
                  !placed &&
                  ((v === 1 && isPositionPlaced(tablePlaced, idx, 13)) ||
                    (v === 13 && isPositionPlaced(tablePlaced, idx, 1)));
                const cellStyle: React.CSSProperties = {
                  display: 'inline-block',
                  width: 22,
                  height: 22,
                  lineHeight: '22px',
                  textAlign: 'center',
                  borderRadius: 4,
                  fontSize: '0.7em',
                  fontWeight: isCenter ? 'bold' : 'normal',
                  background: canPlace
                    ? '#3b82f6'
                    : placed
                      ? isCenter
                        ? '#f0ad4e'
                        : '#5cb85c'
                      : 'rgba(255,255,255,0.1)',
                  color: canPlace ? '#fff' : placed ? '#000' : '#555',
                  border: tunnelHighlight ? '1px solid #f59e0b' : undefined,
                  boxSizing: 'border-box',
                };
                if (canPlace) {
                  return (
                    <button
                      key={v}
                      type="button"
                      onClick={() => onJokerPlace?.(idx, v)}
                      aria-label={t('placeAriaLabel', { suit: suitName(idx), value: valueName(v) })}
                      style={{ ...cellStyle, border: '1px solid #60a5fa', cursor: 'pointer', padding: 0 }}
                    >
                      {valueName(v)}
                    </button>
                  );
                }
                return (
                  <span key={v} style={cellStyle}>
                    {valueName(v)}
                  </span>
                );
              })}
              {tunnelEnabled && (
                <span role="img" className="text-yellow-400 text-[0.65em] ml-0.5" aria-label={t('tunnelConnection')}>
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
