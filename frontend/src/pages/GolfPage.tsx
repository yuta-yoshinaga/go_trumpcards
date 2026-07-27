import { useCallback, useEffect, useMemo, useRef } from 'react';
import type { golfApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useChainCombo } from '../hooks/useChainCombo';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useGolfGame } from '../hooks/useGolfGame';
import {
  countGolfRemaining,
  GOLF_TOTAL_HOLES,
  golfCurrentHole,
  golfNineHoleComplete,
  golfNineHoleTotal,
  useGolfNineHole,
} from '../hooks/useGolfNineHole';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GolfResponse } from '../types/card';
import { GolfPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { GOLF_HELP, parseGolfCommand } from '../utils/cli/commands/golfCommands';
import { formatGolfState } from '../utils/cli/formatters/golfFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isGolfAdjacent } from '../utils/hints/golfHint';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Golf Solitaire tutorial step definitions. */
const GOLF_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="golf-columns"]',
    messageKey: 'tutorial.columns',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="golf-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="golf-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="golf-hint-display"]',
    messageKey: 'tutorial.hintDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="golf-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Golf Solitaire game page with 7 columns, stock/waste, and controls. */
export const GolfPage = withTutorial(GolfPageContent, 'golf', GOLF_TUTORIAL_STEPS);
/** Inner content of the Golf page, wrapped by TutorialProvider. */
function GolfPageContent() {
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
  } = useGamePageSetup('golf');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    hintError,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleUndo,
    handleUndoEscape,
    handleSelectCard,
  } = useGolfGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('golf', state);
  const { cardHeight, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('golf');
  const cliConfig: CliGameConfig<GolfResponse, Parameters<typeof golfApi.exec>> = useMemo(
    () => ({
      gameName: 'golf',
      parseCommand: parseGolfCommand,
      formatResponse: formatGolfState,
      helpText: GOLF_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === GolfPhase.PLAYING;

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw, label: 'draw' },
      { key: 'h', action: handleHint, label: 'hint' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: handleUndo, label: 'undo' },
    ],
    [handleDraw, handleHint, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  const combo = useChainCombo(state?.moveCount, state?.stockCount);

  // 9-hole mode: accumulate each finished deal's remaining-card score across 9 deals (issue #3114).
  const { nineHole, setEnabled: setNineHoleEnabled, recordHole, resetCard } = useGolfNineHole();
  const nineHoleEnabled = nineHole.enabled;
  const endPhase = state?.phase;
  const endRemaining = state?.layout ? countGolfRemaining(state.layout) : 0;
  // Guard so each finished deal is recorded exactly once (phase stays ended across re-renders).
  const recordedRef = useRef(false);
  useEffect(() => {
    const ended = endPhase === GolfPhase.GAME_CLEAR || endPhase === GolfPhase.GAME_OVER;
    if (!ended) {
      recordedRef.current = false;
      return;
    }
    if (recordedRef.current) return;
    recordedRef.current = true;
    if (nineHoleEnabled) recordHole(endRemaining);
  }, [endPhase, endRemaining, nineHoleEnabled, recordHole]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="golf"
        layout={{ kind: 'tiered-rows', rows: [5, 5, 5, 5, 5, 5, 5], stockWaste: true, columns: true }}
      />
    );

  const isPlaying = state.phase === GolfPhase.PLAYING;
  const isGameClear = state.phase === GolfPhase.GAME_CLEAR;
  // Waste-top value drives which exposed tableau cards are playable (±1, K-A wrap).
  const wasteTopValue = isPlaying ? state.waste.at(-1)?.value : undefined;
  const isGameOver = state.phase === GolfPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const COL_COUNT = 7;
  const ROW_COUNT = 5;
  const cardGap = 4;
  const ROW_OVERLAP_RATIO = isMobile ? 0.55 : 0.5;
  const rowOverlap = cardHeight * ROW_OVERLAP_RATIO;
  // px-4 on the scrollable container = 16px * 2 = 32px total horizontal padding
  const CONTAINER_PADDING = 32;
  const effectiveCardWidth = isMobile
    ? Math.floor((windowWidth - CONTAINER_PADDING - cardGap * (COL_COUNT - 1)) / COL_COUNT)
    : cardWidth;

  return (
    <GamePageShell
      title={tc('nav.golf')}
      gameThemeBg={gameTheme.golf.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/golf"
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
          {combo >= 2 && (
            <span
              data-testid="combo-badge"
              className={`px-2 py-0.5 rounded-full text-xs font-bold ${
                combo >= 5
                  ? 'bg-ds-error text-ds-text-on-accent'
                  : combo >= 3
                    ? 'bg-ds-warning text-ds-text-on-accent'
                    : 'bg-ds-info text-ds-text-on-accent'
              }`}
            >
              {t('combo', { count: combo })}
            </span>
          )}
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* 9-hole scorecard (issue #3114) */}
            {nineHoleEnabled && (
              <div data-testid="golf-scorecard" className="mb-3 max-w-2xl mx-auto">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-game-text-muted text-xs">
                    {golfNineHoleComplete(nineHole)
                      ? t('nineHole.complete', { total: golfNineHoleTotal(nineHole) })
                      : t('nineHole.progress', { current: golfCurrentHole(nineHole), total: GOLF_TOTAL_HOLES })}
                  </span>
                  {golfNineHoleComplete(nineHole) && (
                    <button
                      type="button"
                      className={`${btnPrimary} text-xs`}
                      onClick={resetCard}
                      data-testid="golf-scorecard-restart"
                    >
                      {t('nineHole.restart')}
                    </button>
                  )}
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-center text-xs border-collapse">
                    <thead>
                      <tr className="text-game-text-muted">
                        <th className="px-1 py-0.5 font-medium text-left">{t('nineHole.hole')}</th>
                        {Array.from({ length: GOLF_TOTAL_HOLES }, (_, i) => (
                          <th key={`h-${(i + 1).toString()}`} className="px-1 py-0.5 font-medium">
                            {i + 1}
                          </th>
                        ))}
                        <th className="px-1 py-0.5 font-medium text-ds-warning">{t('nineHole.total')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr>
                        <td className="px-1 py-0.5 text-left text-game-text-muted">{t('nineHole.score')}</td>
                        {Array.from({ length: GOLF_TOTAL_HOLES }, (_, i) => {
                          const played = i < nineHole.scores.length;
                          const isCurrent = i === nineHole.scores.length && !golfNineHoleComplete(nineHole);
                          return (
                            <td
                              key={`s-${(i + 1).toString()}`}
                              data-testid={`golf-hole-${(i + 1).toString()}`}
                              className={`px-1 py-0.5 tabular-nums ${
                                isCurrent ? 'text-ds-info font-bold' : played ? 'text-ds-text' : 'text-game-text-muted'
                              }`}
                            >
                              {played ? nineHole.scores[i] : t('nineHole.pending')}
                            </td>
                          );
                        })}
                        <td
                          data-testid="golf-scorecard-total"
                          className="px-1 py-0.5 font-bold text-ds-warning tabular-nums"
                        >
                          {golfNineHoleTotal(nineHole)}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {/* Tableau: 7 columns */}
            <div data-tutorial="golf-columns" className="flex justify-center mb-3">
              {Array.from({ length: COL_COUNT }, (_, colIdx) => (
                <div
                  key={`col-${colIdx.toString()}`}
                  className="relative"
                  style={{
                    width: effectiveCardWidth + cardGap,
                    height: cardHeight + (ROW_COUNT - 1) * (cardHeight - rowOverlap),
                  }}
                >
                  {Array.from({ length: ROW_COUNT }, (_, rowIdx) => {
                    const gc = state.layout[colIdx]?.[rowIdx];
                    const top = rowIdx * (cardHeight - rowOverlap);
                    if (!gc || gc.removed) {
                      return (
                        <div
                          key={`gc-${colIdx.toString()}-${rowIdx.toString()}`}
                          className="absolute"
                          style={{ top, width: effectiveCardWidth, height: cardHeight }}
                        />
                      );
                    }
                    if (!gc.card) return null;
                    const exposed = gc.exposed;
                    const isHinted = hint?.type === 'remove' && hint.col === colIdx;
                    const isPlayable =
                      exposed && wasteTopValue !== undefined && isGolfAdjacent(gc.card.value, wasteTopValue);
                    return (
                      <div key={`gc-${colIdx.toString()}-${rowIdx.toString()}`} className="absolute" style={{ top }}>
                        <button
                          type="button"
                          onClick={() => handleSelectCard(colIdx)}
                          disabled={!isPlaying || loading || !exposed}
                          aria-label={cardAlt(gc.card)}
                          data-testid={isPlayable ? 'golf-playable' : undefined}
                          className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                            isHinted && exposed
                              ? 'ring-2 ring-ds-warning motion-safe:animate-pulse'
                              : isPlayable
                                ? 'ring-2 ring-ds-success'
                                : ''
                          } ${!exposed ? 'opacity-60' : ''}`}
                        >
                          <AnimatedCard card={gc.card} width={effectiveCardWidth} />
                        </button>
                      </div>
                    );
                  })}
                </div>
              ))}
            </div>

            {/* Stock + Waste */}
            <div className="flex gap-4 justify-center mb-3" data-tutorial="golf-stock-waste">
              <div className="text-center">
                <div className="text-game-text-muted text-xs mb-1">
                  {t('stock')} ({state.stockCount})
                </div>
                {state.stockCount > 0 ? (
                  <div
                    data-testid="golf-stock"
                    className={
                      hint?.type === 'draw'
                        ? 'inline-block rounded ring-2 ring-ds-warning motion-safe:animate-pulse'
                        : 'inline-block'
                    }
                  >
                    <AnimatedCardBack
                      width={effectiveCardWidth}
                      onClick={isPlaying ? handleDraw : undefined}
                      ariaLabel={t('draw')}
                    />
                  </div>
                ) : (
                  <div
                    style={{ width: effectiveCardWidth, height: cardHeight }}
                    className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
                  >
                    {t('empty')}
                  </div>
                )}
              </div>

              <div className="text-center">
                <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
                {state.waste.length > 0 ? (
                  <AnimatedCard card={state.waste[state.waste.length - 1]} width={effectiveCardWidth} />
                ) : (
                  <div
                    style={{ width: effectiveCardWidth, height: cardHeight }}
                    className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
                  >
                    {t('empty')}
                  </div>
                )}
              </div>
            </div>

            {/* Hint display */}
            <div data-tutorial="golf-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 text-center">
                  {t('hintAvailable')}: {t(`hintType.${hint.type}`)}
                </div>
              )}
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

          {/* Settings */}
          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                  {
                    type: 'checkbox' as const,
                    id: 'golfNineHole',
                    label: t('nineHole.label'),
                    checked: nineHoleEnabled,
                    onToggle: setNineHoleEnabled,
                    testId: 'golf-ninehole-toggle',
                  },
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.golf.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="golf-controls">
                  <button type="button" className={btnPrimary} onClick={handleDraw} disabled={loading}>
                    {t('draw')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleUndo}
                    disabled={loading || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={handleUndoEscape}
                      disabled={loading}
                    />
                  )}
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
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="golf-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="golf-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
