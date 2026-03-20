import { useMemo, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardBack, CardImage } from '../components/CardImage';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { KlondikeSkeleton } from '../components/skeleton/KlondikeSkeleton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useKlondikeGame } from '../hooks/useKlondikeGame';
import { useKlondikeTimer } from '../hooks/useKlondikeTimer';
import { btnDanger, btnPrimary, btnSuccess, btnWarning, focusRingWhite } from '../styles/buttonStyles';
import { KlondikePhase, KlondikeScoringMode } from '../types/phases';
import { cardAlt } from '../utils/cardAlt';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Renders the Klondike solitaire game page with tableau, stock/waste, and foundation. */
export function KlondikePage() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('klondike');
  const {
    state,
    loading,
    error,
    hintError,
    selectedSource,
    hint,
    handleDraw,
    handleReset,
    handleResetWithConfig,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleSelectSource,
    handleSelectTarget,
  } = useKlondikeGame();
  const { cardHeight, cardOverlap, cardWidth } = useCardDimensions();

  const isPlayingForKbd = state?.phase === KlondikePhase.PLAYING;
  const { elapsedSeconds, resetTimer, timeBonus } = useKlondikeTimer(isPlayingForKbd);

  const [drawCountSetting, setDrawCountSetting] = useState(1);
  const [scoringModeSetting, setScoringModeSetting] = useState(0);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw },
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: handleGiveUp },
      { key: 'z', action: handleUndo },
    ],
    [handleDraw, handleHint, handleAutoComplete, handleGiveUp, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <KlondikeSkeleton />;

  const isPlaying = state.phase === KlondikePhase.PLAYING;
  const isGameClear = state.phase === KlondikePhase.GAME_CLEAR;
  const isGameOver = state.phase === KlondikePhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const isVegas = state.scoringMode === KlondikeScoringMode.VEGAS;

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  const formatTime = (seconds: number) => {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  // Waste display: in 3-card mode show up to 3 fanned cards, only top clickable
  const wasteDisplay = state.drawCount === 3 ? state.waste.slice(-3) : state.waste.slice(-1);

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-casino" aria-busy={loading}>
      {/* Phase indicator */}
      <PhaseIndicator
        phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      >
        <span>
          {t('moveCount')}: {state.moveCount}
        </span>
        {isVegas && (
          <span className="ml-3">
            {t('score')}: {state.score}
          </span>
        )}
        <span className="ml-3">
          {t('timer')}: {formatTime(elapsedSeconds)}
        </span>
        {isGameClear && (
          <span className="ml-3">
            {t('timeBonus')}: {timeBonus(elapsedSeconds)}
          </span>
        )}
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
            {wasteDisplay.length > 0 ? (
              <div className="relative" style={{ width: cardWidth + (wasteDisplay.length - 1) * 15 }}>
                {wasteDisplay.map((card, idx) => {
                  const isTop = idx === wasteDisplay.length - 1;
                  return (
                    <div key={`waste-${idx.toString()}`} className="absolute top-0" style={{ left: idx * 15 }}>
                      {isTop ? (
                        <button
                          type="button"
                          onClick={() => {
                            if (selectedSource) return;
                            handleSelectSource({ zone: 'waste' });
                          }}
                          disabled={!isPlaying || loading}
                          aria-label={cardAlt(card)}
                          aria-pressed={isSourceSelected('waste')}
                          className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste') ? 'ring-2 ring-yellow-400' : ''}`}
                        >
                          <CardImage card={card} width={cardWidth} />
                        </button>
                      ) : (
                        <CardImage card={card} width={cardWidth} />
                      )}
                    </div>
                  );
                })}
                <div style={{ height: cardHeight, width: cardWidth + (wasteDisplay.length - 1) * 15 }} />
              </div>
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

        {/* Score display on game clear */}
        {isGameClear && isVegas && (
          <div className="text-yellow-300 text-lg mb-2">
            {t('totalScore')}: {state.score + timeBonus(elapsedSeconds)}
          </div>
        )}

        {/* Message */}
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

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
        <ErrorAlert message={error ?? hintError} />
        <div className="flex gap-2 items-center flex-wrap">
          {isPlaying && (
            <>
              <button type="button" className={btnPrimary} onClick={handleDraw} disabled={loading}>
                {t('draw')}
              </button>
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
          {/* Draw mode toggle */}
          <select
            value={drawCountSetting}
            onChange={(e) => {
              const n = Number(e.target.value);
              setDrawCountSetting(n);
              handleResetWithConfig({ drawCount: n, scoringMode: scoringModeSetting });
              resetTimer();
            }}
            className="bg-gray-700 text-white text-sm rounded px-2 py-1"
            aria-label={t('drawMode')}
          >
            <option value={1}>{t('drawMode1')}</option>
            <option value={3}>{t('drawMode3')}</option>
          </select>
          {/* Scoring mode toggle */}
          <select
            value={scoringModeSetting}
            onChange={(e) => {
              const n = Number(e.target.value);
              setScoringModeSetting(n);
              handleResetWithConfig({ drawCount: drawCountSetting, scoringMode: n });
              resetTimer();
            }}
            className="bg-gray-700 text-white text-sm rounded px-2 py-1"
            aria-label={t('scoringMode')}
          >
            <option value={0}>{t('scoringNone')}</option>
            <option value={1}>{t('scoringVegas')}</option>
          </select>
          <button
            type="button"
            className={btnWarning}
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                resetTimer();
                setDrawCountSetting(1);
                setScoringModeSetting(0);
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
