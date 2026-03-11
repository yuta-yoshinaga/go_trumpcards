import { useTranslation } from 'react-i18next';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CardBack, CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { useActionLog } from '../hooks/useActionLog';
import { useKlondikeGame } from '../hooks/useKlondikeGame';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { KLONDIKE_PHASE } from '../types/card';
import { cardAlt } from '../utils/cardAlt';

export function KlondikePage() {
  const { t } = useTranslation('klondike');
  const { t: tc } = useTranslation('common');
  const {
    state,
    loading,
    error,
    hintError,
    selectedSource,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleSelectSource,
    handleSelectTarget,
  } = useKlondikeGame();
  const { actionLog, showActionLog, hideActionLog } = useActionLog('klondike');

  if (!state) return null;

  const isPlaying = state.phase === KLONDIKE_PHASE.PLAYING;
  const isGameClear = state.phase === KLONDIKE_PHASE.GAME_CLEAR;
  const isGameOver = state.phase === KLONDIKE_PHASE.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#0d5016]" aria-busy={loading}>
      {loading && <span className="sr-only">{tc('status.loading')}</span>}

      {/* Scrollable area */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* Foundation + Stock/Waste row */}
        <div className="flex gap-2 mb-3 items-start flex-wrap">
          {/* Stock */}
          <div className="text-center">
            <div className="text-white/60 text-xs mb-1">
              {t('stock')} ({state.stockCount})
            </div>
            {state.stockCount > 0 ? (
              <CardBack width={60} onClick={isPlaying ? handleDraw : undefined} ariaLabel={t('draw')} />
            ) : (
              <button
                type="button"
                onClick={handleDraw}
                disabled={!isPlaying || loading}
                className="w-[60px] h-[84px] rounded border-2 border-dashed border-white/30 text-white/40 text-xs flex items-center justify-center"
              >
                {t('draw')}
              </button>
            )}
          </div>

          {/* Waste */}
          <div className="text-center">
            <div className="text-white/60 text-xs mb-1">{t('waste')}</div>
            {state.waste.length > 0 ? (
              <button
                type="button"
                onClick={() => {
                  if (selectedSource) {
                    // clicking waste when source selected does nothing for target
                    return;
                  }
                  handleSelectSource({ zone: 'waste' });
                }}
                disabled={!isPlaying || loading}
                aria-label={cardAlt(state.waste[state.waste.length - 1])}
                aria-pressed={isSourceSelected('waste')}
                className={`p-0 border-0 bg-transparent cursor-pointer ${isSourceSelected('waste') ? 'ring-2 ring-yellow-400 rounded' : ''}`}
              >
                <CardImage card={state.waste[state.waste.length - 1]} width={60} />
              </button>
            ) : (
              <div className="w-[60px] h-[84px] rounded border border-white/20 flex items-center justify-center text-white/30 text-xs">
                {t('empty')}
              </div>
            )}
          </div>

          <div className="w-4" />

          {/* Foundation piles */}
          {state.foundation.map((pile, idx) => (
            <div key={`f-${idx.toString()}`} className="text-center">
              <div className="text-white/60 text-xs mb-1">{['♠', '♣', '♥', '♦'][idx]}</div>
              {pile.length > 0 ? (
                <button
                  type="button"
                  onClick={() => handleSelectTarget({ zone: 'foundation', col: idx })}
                  disabled={!isPlaying || loading || !selectedSource}
                  aria-label={t('foundationAriaLabel', { suit: ['♠', '♣', '♥', '♦'][idx], count: pile.length })}
                  className="p-0 border-0 bg-transparent cursor-pointer"
                >
                  <CardImage card={pile[pile.length - 1]} width={60} />
                </button>
              ) : (
                <button
                  type="button"
                  onClick={() => handleSelectTarget({ zone: 'foundation', col: idx })}
                  disabled={!isPlaying || loading || !selectedSource}
                  aria-label={t('foundationAriaLabel', { suit: ['♠', '♣', '♥', '♦'][idx], count: 0 })}
                  className="w-[60px] h-[84px] rounded border-2 border-dashed border-white/30 text-white/30 text-xs flex items-center justify-center"
                >
                  A
                </button>
              )}
            </div>
          ))}
        </div>

        {/* Move count */}
        <div className="text-white/60 text-sm mb-2">
          {t('moveCount')}: {state.moveCount}
        </div>

        {/* Tableau */}
        <div className="flex gap-2 mb-3">
          {state.tableau.map((col, colIdx) => (
            <div key={`col-${colIdx.toString()}`} className="flex-1 min-w-0">
              <div className="text-white/40 text-xs text-center mb-1">{colIdx}</div>
              <div className="relative" style={{ minHeight: 84 }}>
                {col.length === 0 ? (
                  <button
                    type="button"
                    onClick={() => handleSelectTarget({ zone: 'tableau', col: colIdx })}
                    disabled={!isPlaying || loading || !selectedSource}
                    className="w-full h-[84px] rounded border-2 border-dashed border-white/20 text-white/20 text-xs flex items-center justify-center"
                  >
                    K
                  </button>
                ) : (
                  col.map((tc, cardIdx) => (
                    <div
                      key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                      className="absolute left-0 right-0"
                      style={{ top: cardIdx * 22 }}
                    >
                      {tc.faceUp && tc.card ? (
                        <button
                          type="button"
                          onClick={() => {
                            if (selectedSource) {
                              handleSelectTarget({ zone: 'tableau', col: colIdx });
                            } else {
                              handleSelectSource({ zone: 'tableau', col: colIdx, cardIndex: cardIdx });
                            }
                          }}
                          disabled={!isPlaying || loading}
                          aria-label={cardAlt(tc.card)}
                          aria-pressed={isSourceSelected('tableau', colIdx, cardIdx)}
                          className={`p-0 border-0 bg-transparent cursor-pointer w-full ${isSourceSelected('tableau', colIdx, cardIdx) ? 'ring-2 ring-yellow-400 rounded' : ''}`}
                        >
                          <CardImage card={tc.card} width={60} style={{ width: '100%' }} />
                        </button>
                      ) : (
                        <CardBack width={60} className="w-full" />
                      )}
                    </div>
                  ))
                )}
                {col.length > 0 && <div style={{ height: (col.length - 1) * 22 + 84 }} />}
              </div>
            </div>
          ))}
        </div>

        {/* Hint display */}
        {hint && (
          <div className="text-yellow-300 text-sm mb-2">
            {t('hintAvailable')}: {hint.fromZone}
            {hint.fromCol >= 0 ? ` ${t('tableau')} ${hint.fromCol}` : ` ${t('waste')}`} → {hint.toZone}
            {hint.toCol >= 0 ? ` ${hint.toCol}` : ''}
          </div>
        )}

        {/* Message */}
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* Error */}
        <ErrorAlert message={error ?? hintError} />

        {/* Action log */}
        {isEnded && actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
        {isEnded && !actionLog && (
          <div className="text-center my-2">
            <button type="button" className={btnSecondary} onClick={showActionLog}>
              {tc('actionLog.view')}
            </button>
          </div>
        )}
      </div>

      {/* Footer */}
      <GameFooter className="bg-[#0a3a10] border-white/20 px-4 py-2.5">
        <div className="flex gap-2 items-center flex-wrap">
          {isPlaying && (
            <>
              <button type="button" className={btnPrimary} onClick={handleDraw} disabled={loading}>
                {t('draw')}
              </button>
              <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading}>
                {t('hint')}
              </button>
              <button type="button" className={btnSuccess} onClick={handleAutoComplete} disabled={loading}>
                {t('autoComplete')}
              </button>
              <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                {t('giveup')}
              </button>
            </>
          )}
          <button type="button" className={btnWarning} onClick={handleReset} disabled={loading}>
            {tc('button.reset')}
          </button>
        </div>
      </GameFooter>
    </div>
  );
}
