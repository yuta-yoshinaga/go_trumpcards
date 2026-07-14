import { useCallback, useEffect } from 'react';
import { type GapsMoveZone, gapsApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import { GapsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { valueName } from '../utils/cardUtils';
import { computeGapsGhostHint } from '../utils/gapsGhostHint';
import { gapsLockedPrefixLengths } from '../utils/gapsUtils';

const SUIT_SYMBOLS: Record<string, string> = {
  SPADE: '♠',
  HEART: '♥',
  DIAMOND: '♦',
  CLOVER: '♣',
};
const RED_DESIGNS = new Set(['HEART', 'DIAMOND']);

const GAPS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="gaps-grid"]',
    messageKey: 'tutorial.grid',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gaps-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gaps-grid"]',
    messageKey: 'tutorial.hintDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gaps-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Gaps / Montana solitaire game page. */
export const GapsPage = withTutorial(GapsPageContent, 'gaps', GAPS_TUTORIAL_STEPS);

function GapsPageContent() {
  const {
    t,
    tc,
    actionLog,
    showActionLog,
    hideActionLog,
    confirmOpen,
    requestConfirm,
    confirmReset,
    cancelReset,
    giveUpConfirmOpen,
    requestGiveUpConfirm,
    confirmGiveUp,
    cancelGiveUp,
  } = useGamePageSetup('gaps');
  const apiRun = gapsApi.exec;
  const { state, loading, error, exec: run, retry } = useGameApi(apiRun);

  useEffect(() => {
    void run('reset');
  }, [run]);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('gaps', state);

  const { cardWidth, cardHeight } = useCardDimensions();

  const handleReset = useCallback(() => {
    hideActionLog();
    void run('reset');
  }, [run, hideActionLog]);

  const handleUndo = useCallback(() => {
    void run('undo');
  }, [run]);

  const handleRedeal = useCallback(() => {
    void run('redeal');
  }, [run]);

  const handleGiveUp = useCallback(() => {
    void run('giveup');
  }, [run]);

  const handleHint = useCallback(() => {
    void run('hint');
  }, [run]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useCallback(
    () => requestGiveUpConfirm(handleGiveUp),
    [requestGiveUpConfirm, handleGiveUp],
  );

  const dispatchMove = useCallback(
    (source: GapsMoveZone, target: GapsMoveZone) => {
      void run('move', source, target);
    },
    [run],
  );

  const isPlaying = state?.phase === GapsPhase.PLAYING;
  const dnd = useSolitaireDragDrop<GapsMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlaying,
    disabled: loading,
  });

  if (!state) return <GameSkeleton gameKey="gaps" layout={{ kind: 'tiered-rows', rows: [13, 13, 13, 13] }} />;

  const isGameClear = state.phase === GapsPhase.GAME_CLEAR;
  const isGameOver = state.phase === GapsPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const lockedPrefixLengths = gapsLockedPrefixLengths(state.grid);
  // Derive the total from the already-computed per-row lengths to avoid a second board pass.
  const lockedTotal = lockedPrefixLengths.reduce((sum, n) => sum + n, 0);

  return (
    <GamePageShell
      title={tc('nav.gaps')}
      gameThemeBg={gameTheme.gaps.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/gaps"
      gameEndFlag={isEnded}
      winShow={isGameClear}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      giveUpConfirmOpen={giveUpConfirmOpen}
      confirmGiveUp={confirmGiveUp}
      cancelGiveUp={cancelGiveUp}
      headerExtra={
        <>
          <span>
            {t('moveCount')}: {state.moveCount}
          </span>
          <span>
            {t('redealsRemaining')}: {state.redealsRemaining}
          </span>
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
        <div data-tutorial="gaps-grid" className="flex flex-col items-center gap-1 mb-3">
          {state.grid.map((row, rIdx) => {
            const lockedCount = lockedPrefixLengths[rIdx] ?? 0;
            return (
              <div key={`row-${rIdx.toString()}`} className="flex gap-1">
                {row.map((cell, cIdx) => {
                  const zone: GapsMoveZone = { zone: 'grid', row: rIdx, col: cIdx };
                  const isHintFrom = state.hint && state.hint.fromRow === rIdx && state.hint.fromCol === cIdx;
                  const isHintTo = state.hint && state.hint.toRow === rIdx && state.hint.toCol === cIdx;
                  if (cell === null) {
                    const ghost = computeGapsGhostHint(row, cIdx);
                    // Mirror the (aria-hidden) ghost hint into the cell's label so
                    // SR users learn which card each gap accepts, or that it's blocked.
                    const gapAria =
                      ghost?.kind === 'needed'
                        ? t('gapAriaNeeded', { card: cardAlt({ design: ghost.design, value: ghost.value }) })
                        : ghost?.kind === 'anySuit'
                          ? t('gapAriaAnySuit', { value: valueName(ghost.value) })
                          : ghost?.kind === 'blocked'
                            ? t('gapAriaBlocked')
                            : t('gap');
                    return (
                      <button
                        type="button"
                        key={`cell-${rIdx.toString()}-${cIdx.toString()}`}
                        onDragOver={dnd.handleDragOver(zone)}
                        onDragLeave={dnd.handleDragLeave}
                        onDrop={dnd.handleDrop(zone)}
                        aria-label={gapAria}
                        data-testid={`gaps-cell-${rIdx.toString()}-${cIdx.toString()}`}
                        className={`relative flex items-center justify-center rounded border-2 ${
                          dnd.isDropTarget(zone)
                            ? 'border-ds-warning bg-ds-warning/20'
                            : 'border-dashed border-white/30'
                        } ${isHintTo ? 'ring-2 ring-ds-warning' : ''} ${focusRingWhite}`}
                        style={{ width: cardWidth, height: cardHeight }}
                        disabled={!isPlaying || loading}
                      >
                        {ghost?.kind === 'needed' && (
                          <span
                            aria-hidden="true"
                            data-testid={`gaps-ghost-${rIdx}-${cIdx}`}
                            className={`text-base font-semibold opacity-30 ${RED_DESIGNS.has(ghost.design) ? 'text-ds-error' : 'text-ds-text-primary'}`}
                          >
                            {SUIT_SYMBOLS[ghost.design]}
                            {valueName(ghost.value)}
                          </span>
                        )}
                        {ghost?.kind === 'anySuit' && (
                          <span
                            aria-hidden="true"
                            data-testid={`gaps-ghost-${rIdx}-${cIdx}`}
                            className="text-base font-semibold text-white opacity-30"
                          >
                            {valueName(ghost.value)}
                          </span>
                        )}
                        {ghost?.kind === 'blocked' && (
                          <span
                            aria-hidden="true"
                            data-testid={`gaps-ghost-${rIdx}-${cIdx}-blocked`}
                            className="text-xl opacity-40"
                          >
                            🚫
                          </span>
                        )}
                      </button>
                    );
                  }
                  const isLocked = cIdx < lockedCount;
                  // Locked cards survive redeal — make them visually fixed by
                  // dropping the draggable affordance, so users don't try to
                  // move them and end up with a no-op drop.
                  const ringClass = isHintFrom ? 'ring-2 ring-ds-warning' : isLocked ? 'ring-2 ring-ds-success' : '';
                  return (
                    <button
                      type="button"
                      key={`cell-${rIdx.toString()}-${cIdx.toString()}`}
                      draggable={isPlaying && !loading && !isLocked}
                      onDragStart={dnd.handleDragStart(zone)}
                      onDragEnd={dnd.handleDragEnd}
                      aria-label={isLocked ? `${cardAlt(cell)} ${t('lockedAria')}` : cardAlt(cell)}
                      disabled={!isPlaying || loading}
                      data-testid={
                        isLocked
                          ? `gaps-locked-${rIdx.toString()}-${cIdx.toString()}`
                          : `gaps-cell-${rIdx.toString()}-${cIdx.toString()}`
                      }
                      className={`relative p-0 border-0 bg-transparent rounded ${focusRingWhite} ${ringClass} ${dnd.isDragSource(zone) ? 'opacity-50' : ''}`}
                    >
                      <AnimatedCard card={cell} width={cardWidth} />
                      {isLocked && (
                        <span
                          aria-hidden="true"
                          title={t('lockedTooltip')}
                          className="absolute top-0.5 right-0.5 text-[10px] leading-none bg-ds-success/80 text-ds-text-on-accent rounded-sm px-1 py-0.5 shadow"
                        >
                          🔒
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            );
          })}
        </div>

        {frontendHintEnabled && frontendHint && (
          <div className="flex justify-center">
            <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
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

      <SettingsPanel
        title={tc('settings.title')}
        groups={[
          {
            items: [
              {
                type: 'checkbox' as const,
                id: 'frontendHint',
                label: tc('hint.toggle', { ns: 'tutorial' }),
                checked: frontendHintEnabled,
                onToggle: setFrontendHintEnabled,
              },
            ],
          },
        ]}
      />

      <GameFooter className={`${gameTheme.gaps.footer} px-4 py-2.5`}>
        <ErrorAlert message={error} onRetry={retry} />
        <div className="flex gap-2 items-center flex-wrap">
          {isPlaying && (
            <div data-tutorial="gaps-controls" className="flex gap-2 flex-wrap">
              <button type="button" className={btnPrimary} onClick={handleUndo} disabled={loading || !state.canUndo}>
                {t('undo')}
              </button>
              <button
                type="button"
                className={btnPrimary}
                onClick={handleRedeal}
                disabled={loading || state.redealsRemaining <= 0}
                data-testid="gaps-redeal-button"
              >
                {t('redealKeepLabel', {
                  used: state.redealsUsed,
                  total: state.redealsUsed + state.redealsRemaining,
                  locked: lockedTotal,
                })}
              </button>
              <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading}>
                {t('hint')}
              </button>
              <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={loading}>
                {t('giveup')}
              </button>
            </div>
          )}
          <GameResetButton
            isGameEnd={isEnded}
            onReset={handleReset}
            requestConfirm={requestConfirm}
            loading={loading}
            dataTutorial="gaps-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
