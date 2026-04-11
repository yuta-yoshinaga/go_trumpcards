import { useCallback, useEffect, useMemo } from 'react';
import { canfieldApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnDanger, btnOutline, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import { CanfieldPhase } from '../types/phases';

/** Renders the Canfield solitaire game page. */
export function CanfieldPage() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('canfield');
  const { state, loading, error, exec: execApi, retry } = useGameApi(canfieldApi.exec);
  const { cardWidth, cardHeight } = useCardDimensions();

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  const handleReset = useCallback(() => execApi('reset'), [execApi]);
  const handleDraw = useCallback(() => execApi('draw'), [execApi]);
  const handleGiveUp = useCallback(() => execApi('giveup'), [execApi]);
  const handleHint = useCallback(() => execApi('hint'), [execApi]);
  const handleAutoComplete = useCallback(() => execApi('autocomplete'), [execApi]);
  const handleUndo = useCallback(() => execApi('undo'), [execApi]);

  const handleMoveReserveToFoundation = useCallback(
    () => execApi('move', { zone: 'reserve' }, { zone: 'foundation' }),
    [execApi],
  );
  const handleMoveWasteToFoundation = useCallback(
    () => execApi('move', { zone: 'waste' }, { zone: 'foundation' }),
    [execApi],
  );
  const handleMoveReserveToTableau = useCallback(
    (col: number) => execApi('move', { zone: 'reserve' }, { zone: 'tableau', col }),
    [execApi],
  );
  const handleMoveWasteToTableau = useCallback(
    (col: number) => execApi('move', { zone: 'waste' }, { zone: 'tableau', col }),
    [execApi],
  );
  const handleMoveTableauToFoundation = useCallback(
    (col: number) => execApi('move', { zone: 'tableau', col }, { zone: 'foundation' }),
    [execApi],
  );
  const handleMoveTableauToTableau = useCallback(
    (fromCol: number, cardIndex: number, toCol: number) =>
      execApi('move', { zone: 'tableau', col: fromCol, cardIndex }, { zone: 'tableau', col: toCol }),
    [execApi],
  );

  const theme = useMemo(() => gameTheme.canfield, []);

  const phase = state?.phase ?? CanfieldPhase.PLAYING;
  const isPlaying = phase === CanfieldPhase.PLAYING;
  const isGameClear = phase === CanfieldPhase.GAME_CLEAR;
  const isEnded = phase === CanfieldPhase.GAME_CLEAR || phase === CanfieldPhase.GAME_OVER;

  const phaseName = isGameClear
    ? t('phase.gameClear')
    : phase === CanfieldPhase.GAME_OVER
      ? t('phase.gameOver')
      : t('phase.playing');

  if (error) return <ErrorAlert message={error} onRetry={retry} />;
  if (!state) return null;

  const topWaste = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const topReserve = state.reserve.length > 0 ? state.reserve[state.reserve.length - 1] : null;

  return (
    <div className={`flex min-h-screen flex-1 flex-col ${theme.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.canfield')} />
      <PhaseIndicator phaseName={phaseName}>
        <span className="text-sm text-white/70">
          {t('baseRank')}: {state.baseRank || '?'}
        </span>
        <span className="text-sm text-white/70">
          {t('moveCount')}: {state.moveCount}
        </span>
        <ManualButton gamePath="/canfield" />
      </PhaseIndicator>

      <LandscapeBanner message={phaseName} />

      <div className="flex-1 overflow-y-auto px-4 pt-3 lg:px-8">
        {/* Foundation */}
        <div className="mb-3 flex gap-2">
          {state.foundation.map((pile, i) => (
            <div
              key={`f-${i}`}
              className="relative rounded border border-white/30"
              style={{ width: cardWidth, height: cardHeight }}
            >
              {pile.length > 0 ? (
                <AnimatedCard card={pile[pile.length - 1]} width={cardWidth} />
              ) : (
                <span className="absolute inset-0 flex items-center justify-center text-xs text-white/40">
                  {t('foundation')}
                </span>
              )}
            </div>
          ))}
        </div>

        {/* Stock / Waste / Reserve */}
        <div className="mb-3 flex gap-3">
          <div className="flex flex-col items-center">
            <button
              type="button"
              onClick={handleDraw}
              disabled={!isPlaying || loading}
              className="rounded border border-white/30"
              aria-label={t('stock')}
              style={{ width: cardWidth, height: cardHeight }}
            >
              {state.stockCount > 0 ? (
                <AnimatedCardBack width={cardWidth} />
              ) : (
                <span className="text-xs text-white/40">{t('empty')}</span>
              )}
            </button>
            <span className="mt-1 text-xs text-white/70">
              {t('stock')}: {state.stockCount}
            </span>
          </div>

          <div className="flex flex-col items-center">
            <div style={{ width: cardWidth, height: cardHeight }}>
              {topWaste ? (
                <AnimatedCard card={topWaste} width={cardWidth} />
              ) : (
                <div
                  className="rounded border border-dashed border-white/30"
                  style={{ width: cardWidth, height: cardHeight }}
                />
              )}
            </div>
            <span className="mt-1 text-xs text-white/70">{t('waste')}</span>
          </div>

          <div className="flex flex-col items-center">
            <div style={{ width: cardWidth, height: cardHeight }}>
              {topReserve ? (
                <AnimatedCard card={topReserve} width={cardWidth} />
              ) : (
                <div
                  className="rounded border border-dashed border-white/30"
                  style={{ width: cardWidth, height: cardHeight }}
                />
              )}
            </div>
            <span className="mt-1 text-xs text-white/70">
              {t('reserve')}: {state.reserve.length}
            </span>
          </div>
        </div>

        {/* Tableau */}
        <div className="mb-3 flex gap-2">
          {state.tableau.map((col, i) => (
            <div key={`t-${i}`} className="flex flex-col gap-1">
              <span className="text-xs text-white/70">#{i}</span>
              <div className="relative" style={{ width: cardWidth, minHeight: cardHeight }}>
                {col.length === 0 ? (
                  <div
                    className="rounded border border-dashed border-white/30"
                    style={{ width: cardWidth, height: cardHeight }}
                  />
                ) : (
                  col.map((tc, j) => (
                    <div key={`t-${i}-${j}`} className="absolute" style={{ top: j * 24, left: 0 }}>
                      <AnimatedCard card={tc.card} width={cardWidth} />
                    </div>
                  ))
                )}
              </div>
              {isPlaying && (
                <div className="flex flex-col gap-1">
                  <button
                    type="button"
                    className={`${btnOutline} ${focusRingWhite} text-xs`}
                    onClick={() => handleMoveWasteToTableau(i)}
                    disabled={!topWaste || loading}
                  >
                    W→{i}
                  </button>
                  <button
                    type="button"
                    className={`${btnOutline} ${focusRingWhite} text-xs`}
                    onClick={() => handleMoveReserveToTableau(i)}
                    disabled={!topReserve || loading}
                  >
                    R→{i}
                  </button>
                  <button
                    type="button"
                    className={`${btnOutline} ${focusRingWhite} text-xs`}
                    onClick={() => handleMoveTableauToFoundation(i)}
                    disabled={col.length === 0 || loading}
                  >
                    →F
                  </button>
                  {state.tableau.map((_, j) =>
                    j === i ? null : (
                      <button
                        key={`t-${i}-to-${j}`}
                        type="button"
                        className={`${btnOutline} ${focusRingWhite} text-xs`}
                        onClick={() => handleMoveTableauToTableau(i, col.length - 1, j)}
                        disabled={col.length === 0 || loading}
                      >
                        →T{j}
                      </button>
                    ),
                  )}
                </div>
              )}
            </div>
          ))}
        </div>

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        <ActionLogSection
          isEndPhase={isEnded}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${theme.footer} px-4 py-2.5`}>
        <div className="flex flex-wrap items-center gap-2">
          {isPlaying && (
            <>
              <button
                type="button"
                className={`${btnPrimary} ${focusRingWhite}`}
                onClick={handleMoveReserveToFoundation}
                disabled={!topReserve || loading}
              >
                {t('moveReserveToFoundation')}
              </button>
              <button
                type="button"
                className={`${btnPrimary} ${focusRingWhite}`}
                onClick={handleMoveWasteToFoundation}
                disabled={!topWaste || loading}
              >
                {t('moveWasteToFoundation')}
              </button>
              <button
                type="button"
                className={`${btnSuccess} ${focusRingWhite}`}
                onClick={handleHint}
                disabled={loading}
              >
                {t('hint')}
              </button>
              <button
                type="button"
                className={`${btnSuccess} ${focusRingWhite}`}
                onClick={handleAutoComplete}
                disabled={loading}
              >
                {t('autoComplete')}
              </button>
              <button
                type="button"
                className={`${btnOutline} ${focusRingWhite}`}
                onClick={handleUndo}
                disabled={!state.canUndo || loading}
              >
                {t('undo')}
              </button>
              <button
                type="button"
                className={`${btnDanger} ${focusRingWhite}`}
                onClick={handleGiveUp}
                disabled={loading}
              >
                {t('giveup')}
              </button>
            </>
          )}
          <button
            type="button"
            className={`${btnDanger} ${focusRingWhite}`}
            onClick={() => requestConfirm(handleReset)}
          >
            {tc('reset')}
          </button>
        </div>
      </GameFooter>

      <WinCelebration show={isGameClear} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
