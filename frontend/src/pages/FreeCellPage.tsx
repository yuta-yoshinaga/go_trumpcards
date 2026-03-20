import { useMemo } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { FreeCellSkeleton } from '../components/skeleton/FreeCellSkeleton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useFreeCellGame } from '../hooks/useFreeCellGame';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnDanger, btnPrimary, btnSuccess, btnWarning, focusRingWhite } from '../styles/buttonStyles';
import type { Card } from '../types/card';
import { FreeCellPhase } from '../types/phases';
import { cardAlt } from '../utils/cardAlt';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Renders the FreeCell solitaire game page with tableau, free cells, and foundation. */
export function FreeCellPage() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('freecell');
  const {
    state,
    loading,
    error,
    hintError,
    selectedSource,
    hint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleSelectSource,
    handleSelectTarget,
  } = useFreeCellGame();
  const { cardHeight, cardOverlap, cardWidth } = useCardDimensions();

  const isPlayingForKbd = state?.phase === FreeCellPhase.PLAYING;

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: handleGiveUp },
      { key: 'z', action: handleUndo },
    ],
    [handleHint, handleAutoComplete, handleGiveUp, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <FreeCellSkeleton />;

  const isPlaying = state.phase === FreeCellPhase.PLAYING;
  const isGameClear = state.phase === FreeCellPhase.GAME_CLEAR;
  const isGameOver = state.phase === FreeCellPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSourceSelected = (zone: string, col?: number, cell?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cell === cell &&
    selectedSource.cardIndex === cardIndex;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-casino" aria-busy={loading}>
      <PhaseIndicator
        phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      >
        <span>
          {t('moveCount')}: {state.moveCount}
        </span>
      </PhaseIndicator>

      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* Free cells + Foundation row */}
        <div className="flex gap-2 mb-3 items-start flex-wrap">
          {/* Free cells */}
          {state.freeCells.map((card: Card | null, idx: number) => (
            <div key={`fc-${idx.toString()}`} className="text-center">
              <div className="text-white/60 text-xs mb-1">
                {t('freecell')} {idx}
              </div>
              {card ? (
                <button
                  type="button"
                  onClick={() => handleSelectSource({ zone: 'freecell', cell: idx })}
                  disabled={!isPlaying || loading}
                  aria-label={cardAlt(card)}
                  aria-pressed={isSourceSelected('freecell', undefined, idx)}
                  className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('freecell', undefined, idx) ? 'ring-2 ring-yellow-400' : ''}`}
                >
                  <CardImage card={card} width={cardWidth} />
                </button>
              ) : (
                <button
                  type="button"
                  onClick={() => handleSelectTarget({ zone: 'freecell', cell: idx })}
                  disabled={!isPlaying || loading || !selectedSource}
                  aria-label={t('emptyFreecellAriaLabel', { idx: String(idx) })}
                  style={{ width: cardWidth, height: cardHeight }}
                  className={`rounded border-2 border-dashed border-white/30 text-white/30 text-xs flex items-center justify-center ${focusRingWhite}`}
                >
                  {t('empty')}
                </button>
              )}
            </div>
          ))}

          <div className="w-4" />

          {/* Foundation piles */}
          {state.foundation.map((pile: Card[], idx: number) => (
            <div key={`f-${idx.toString()}`} className="text-center">
              <div className="text-white/60 text-xs mb-1">{FOUNDATION_SUITS[idx]}</div>
              {pile.length > 0 ? (
                <button
                  type="button"
                  onClick={() => handleSelectTarget({ zone: 'foundation', col: idx })}
                  disabled={!isPlaying || loading || !selectedSource}
                  aria-label={t('foundationAriaLabel', { suit: FOUNDATION_SUITS[idx], cardCount: String(pile.length) })}
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
          {state.tableau.map((col: (Card | null)[], colIdx: number) => (
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
                  col.map((card: Card | null, cardIdx: number) => (
                    <div
                      key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                      className="absolute left-0 right-0"
                      style={{ top: cardIdx * cardOverlap }}
                    >
                      {card ? (
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
                          aria-label={cardAlt(card)}
                          aria-pressed={isSourceSelected('tableau', colIdx, undefined, cardIdx)}
                          className={`p-0 border-0 bg-transparent cursor-pointer w-full rounded ${focusRingWhite} ${isSourceSelected('tableau', colIdx, undefined, cardIdx) ? 'ring-2 ring-yellow-400' : ''}`}
                        >
                          <CardImage card={card} width={cardWidth} style={{ width: '100%' }} />
                        </button>
                      ) : (
                        <div style={{ width: cardWidth, height: cardHeight }} />
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
            {hint.fromCol >= 0 ? ` ${hint.fromCol}` : ''} → {hint.toZone}
            {hint.toCol >= 0 ? ` ${hint.toCol}` : ''}
          </div>
        )}

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        <ActionLogSection
          isEndPhase={isEnded}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className="bg-game-bg-casino-dark border-white/20 px-4 py-2.5">
        <ErrorAlert message={error ?? hintError} />
        <div className="flex gap-2 items-center flex-wrap">
          {isPlaying && (
            <>
              <button type="button" className={btnPrimary} onClick={handleUndo} disabled={loading || !state.canUndo}>
                {t('undo')}
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
