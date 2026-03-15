import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardBack, CardImage } from '../components/CardImage';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useActionLog } from '../hooks/useActionLog';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useConfirmDialog } from '../hooks/useConfirmDialog';
import { useKlondikeGame } from '../hooks/useKlondikeGame';
import { btnDanger, btnPrimary, btnSuccess, btnWarning, focusRingWhite } from '../styles/buttonStyles';
import { KlondikePhase } from '../types/phases';
import { cardAlt } from '../utils/cardAlt';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

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
  const { cardHeight, cardOverlap, cardWidth } = useCardDimensions();
  const { actionLog, showActionLog, hideActionLog } = useActionLog('klondike');
  const { isOpen: confirmOpen, requestConfirm, confirm: confirmReset, cancel: cancelReset } = useConfirmDialog();

  const isPlayingForKbd = state?.phase === KlondikePhase.PLAYING;

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw },
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: handleGiveUp },
    ],
    [handleDraw, handleHint, handleAutoComplete, handleGiveUp],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return null;

  const isPlaying = state.phase === KlondikePhase.PLAYING;
  const isGameClear = state.phase === KlondikePhase.GAME_CLEAR;
  const isGameOver = state.phase === KlondikePhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-casino" aria-busy={loading}>
      <LoadingSpinner loading={loading} />

      {/* Phase indicator */}
      <PhaseIndicator
        phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      >
        <span>
          {t('moveCount')}: {state.moveCount}
        </span>
      </PhaseIndicator>

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
              <CardBack width={cardWidth} onClick={isPlaying ? handleDraw : undefined} ariaLabel={t('draw')} />
            ) : (
              <button
                type="button"
                onClick={handleDraw}
                disabled={!isPlaying || loading}
                style={{ width: cardWidth, height: cardHeight }}
                className={`rounded border-2 border-dashed border-white/30 text-white/40 text-xs flex items-center justify-center ${focusRingWhite}`}
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
                className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste') ? 'ring-2 ring-yellow-400' : ''}`}
              >
                <CardImage card={state.waste[state.waste.length - 1]} width={cardWidth} />
              </button>
            ) : (
              <div
                style={{ width: cardWidth, height: cardHeight }}
                className="rounded border border-white/20 flex items-center justify-center text-white/30 text-xs"
              >
                {t('empty')}
              </div>
            )}
          </div>

          <div className="w-4" />

          {/* Foundation piles */}
          {state.foundation.map((pile, idx) => (
            <div key={`f-${idx.toString()}`} className="text-center">
              <div className="text-white/60 text-xs mb-1">{FOUNDATION_SUITS[idx]}</div>
              {pile.length > 0 ? (
                <button
                  type="button"
                  onClick={() => handleSelectTarget({ zone: 'foundation', col: idx })}
                  disabled={!isPlaying || loading || !selectedSource}
                  aria-label={t('foundationAriaLabel', { suit: FOUNDATION_SUITS[idx], count: pile.length })}
                  className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                >
                  <CardImage card={pile[pile.length - 1]} width={cardWidth} />
                </button>
              ) : (
                <button
                  type="button"
                  onClick={() => handleSelectTarget({ zone: 'foundation', col: idx })}
                  disabled={!isPlaying || loading || !selectedSource}
                  aria-label={t('emptyFoundationAriaLabel', { suit: FOUNDATION_SUITS[idx] })}
                  style={{ width: cardWidth, height: cardHeight }}
                  className={`rounded border-2 border-dashed border-white/30 text-white/30 text-xs flex items-center justify-center ${focusRingWhite}`}
                >
                  A
                </button>
              )}
            </div>
          ))}
        </div>

        {/* Tableau */}
        <div className="flex gap-2 mb-3">
          {state.tableau.map((col, colIdx) => (
            <div key={`col-${colIdx.toString()}`} className="flex-1 min-w-0">
              <div className="text-white/40 text-xs text-center mb-1">{colIdx}</div>
              <div className="relative" style={{ minHeight: cardHeight }}>
                {col.length === 0 ? (
                  <button
                    type="button"
                    onClick={() => handleSelectTarget({ zone: 'tableau', col: colIdx })}
                    disabled={!isPlaying || loading || !selectedSource}
                    style={{ height: cardHeight }}
                    className={`w-full rounded border-2 border-dashed border-white/20 text-white/20 text-xs flex items-center justify-center ${focusRingWhite}`}
                  >
                    K
                  </button>
                ) : (
                  col.map((tc, cardIdx) => (
                    <div
                      key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                      className="absolute left-0 right-0"
                      style={{ top: cardIdx * cardOverlap }}
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
                          className={`p-0 border-0 bg-transparent cursor-pointer w-full rounded ${focusRingWhite} ${isSourceSelected('tableau', colIdx, cardIdx) ? 'ring-2 ring-yellow-400' : ''}`}
                        >
                          <CardImage card={tc.card} width={cardWidth} style={{ width: '100%' }} />
                        </button>
                      ) : (
                        <CardBack width={cardWidth} className="w-full" />
                      )}
                    </div>
                  ))
                )}
                {col.length > 0 && <div style={{ height: (col.length - 1) * cardOverlap + cardHeight }} />}
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
        <ActionLogSection
          isEndPhase={isEnded}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      {/* Footer */}
      <GameFooter className="bg-game-bg-casino-dark border-white/20 px-4 py-2.5">
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
          <button
            type="button"
            className={btnWarning}
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                return handleReset();
              })
            }
            disabled={loading}
          >
            {tc('button.reset')}
          </button>
        </div>
      </GameFooter>
      <ConfirmDialog
        open={confirmOpen}
        title={tc('button.confirmReset')}
        message={tc('button.confirmResetMessage')}
        confirmLabel={tc('button.confirm')}
        cancelLabel={tc('button.cancel')}
        onConfirm={confirmReset}
        onCancel={cancelReset}
      />
    </div>
  );
}
