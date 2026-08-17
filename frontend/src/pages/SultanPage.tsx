import { useCallback, useMemo } from 'react';
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
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useResponsiveTableau } from '../hooks/useResponsiveTableau';
import { useSultanGame } from '../hooks/useSultanGame';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import { SultanPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSultanCommand, SULTAN_HELP } from '../utils/cli/commands/sultanCommands';
import { formatSultanState } from '../utils/cli/formatters/sultanFormatter';
import { hintCheckboxItem } from '../utils/settingsItems';
import { sultanFoundationInfo } from '../utils/sultanFoundation';

/**
 * Maximum number of waste redeals allowed in Sultan of Turkey.
 * Mirrors `domain.SultanMaxRedeal` (total of 3 passes through the stock).
 */
const SULTAN_MAX_REDEAL = 2;

/** Sultan of Turkey tutorial step definitions. */
const SULTAN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sultan-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sultan-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sultan-divan"]',
    messageKey: 'tutorial.divan',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sultan-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Sultan of Turkey solitaire game page with foundations, divan reserve, and stock/waste. */
export const SultanPage = withTutorial(SultanPageContent, 'sultan', SULTAN_TUTORIAL_STEPS);

/** Format a frontend hint zone for display. */
function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'waste') return t('frontendHint.waste');
  return t('frontendHint.divan', { idx });
}

/** Inner content of the Sultan of Turkey page, wrapped by TutorialProvider. */
function SultanPageContent() {
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
  } = useGamePageSetup('sultan');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    hintError,
    hint,
    handleDraw,
    handleRedeal,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleUndoEscape,
    handlePlay,
    isAutoCompleting,
  } = useSultanGame();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sultan');
  const sultanCliConfig = useMemo(
    () => ({
      gameName: 'sultan' as const,
      parseCommand: parseSultanCommand,
      formatResponse: formatSultanState,
      helpText: SULTAN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, sultanCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('sultan', state);

  const sultan = useResponsiveTableau(8);

  const isPlayingForKbd = state?.phase === SultanPhase.PLAYING;

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw, label: 'draw' },
      { key: 'h', action: handleHint, label: 'hint' },
      { key: 'a', action: handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: () => requestGiveUpConfirm(handleGiveUp), label: 'giveUp' },
      { key: 'z', action: handleUndo, label: 'undo' },
    ],
    [handleDraw, handleHint, handleAutoComplete, requestGiveUpConfirm, handleGiveUp, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  if (!state) return <GameSkeleton gameKey="sultan" layout={{ kind: 'tableau', topRow: 10, tableau: 8 }} />;

  const isPlaying = state.phase === SultanPhase.PLAYING;
  const isGameClear = state.phase === SultanPhase.GAME_CLEAR;
  const isGameOver = state.phase === SultanPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  // Auto-complete is offered once the stock and waste are empty.
  const autoCompleteReady = state.stockCount === 0 && state.waste.length === 0;

  // Waste display: show top card only
  const wasteDisplay = state.waste.slice(-1);

  return (
    <GamePageShell
      title={tc('nav.sultan')}
      gameThemeBg={gameTheme.sultan.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/sultan"
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
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner')} />

          {/* Scrollable area */}
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Stock/Waste + Foundation row */}
            <div className="flex gap-1 sm:gap-2 mb-3 items-start flex-wrap">
              {/* Stock + Waste */}
              <div className="flex gap-1 sm:gap-2" data-tutorial="sultan-stock-waste">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('stock')} ({state.stockCount})
                  </div>
                  {state.stockCount > 0 ? (
                    <AnimatedCardBack
                      width={sultan.cw}
                      onClick={isPlaying ? handleDraw : undefined}
                      ariaLabel={t('draw')}
                    />
                  ) : (
                    <div
                      style={{ width: sultan.cw, height: sultan.ch }}
                      className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
                    >
                      {t('empty')}
                    </div>
                  )}
                </div>

                {/* Waste */}
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
                  {wasteDisplay.length > 0 ? (
                    <button
                      type="button"
                      onClick={() => handlePlay({ zone: 'waste' })}
                      disabled={!isPlaying || loading || isAutoCompleting}
                      aria-label={cardAlt(wasteDisplay[0])}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                    >
                      <AnimatedCard card={wasteDisplay[0]} width={sultan.cw} draggable={false} />
                    </button>
                  ) : (
                    <div
                      style={{ width: sultan.cw, height: sultan.ch }}
                      className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
                    >
                      {t('empty')}
                    </div>
                  )}
                </div>
              </div>

              <div className="w-2 sm:w-4" />

              {/* Foundation piles (8 King-based piles) */}
              <div className="flex gap-1 sm:gap-2 flex-wrap" data-tutorial="sultan-foundation">
                {state.foundation.map((pile, idx) => {
                  const info = sultanFoundationInfo(pile);
                  return (
                    <div key={`f-${idx.toString()}`} className="text-center">
                      <div
                        className="text-game-text-muted text-xs mb-1"
                        data-testid={`sultan-foundation-label-${idx.toString()}`}
                      >
                        {info.suit ? `${idx + 1} ${info.suit}` : idx + 1}
                      </div>
                      {pile.length > 0 ? (
                        <div
                          role="img"
                          aria-label={t('foundationSlot', {
                            idx: idx + 1,
                            top: cardAlt(pile[pile.length - 1]),
                          })}
                        >
                          <AnimatedCard
                            card={pile[pile.length - 1]}
                            width={sultan.cw}
                            draggable={false}
                            dealDelay={isAutoCompleting ? idx * 0.15 : 0}
                          />
                        </div>
                      ) : (
                        <div
                          role="img"
                          aria-label={t('emptyFoundationSlot', { idx: idx + 1 })}
                          style={{ width: sultan.cw, height: sultan.ch }}
                          className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex flex-col items-center justify-center leading-tight"
                        >
                          <span>{idx + 1}</span>
                          <span>K</span>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Divan reserve (8 slots; null = played/empty) */}
            <div className="flex gap-1 sm:gap-2 mb-3 flex-wrap" data-tutorial="sultan-divan">
              {state.divan.map((dcard, idx) => (
                <div key={`divan-${idx.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{idx}</div>
                  {dcard ? (
                    <button
                      type="button"
                      onClick={() => handlePlay({ zone: 'divan', divanIdx: idx })}
                      disabled={!isPlaying || loading || isAutoCompleting}
                      aria-label={cardAlt(dcard)}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                    >
                      <AnimatedCard card={dcard} width={sultan.cw} draggable={false} />
                    </button>
                  ) : (
                    <div
                      role="img"
                      aria-label={t('emptyDivanSlot', { idx })}
                      style={{ width: sultan.cw, height: sultan.ch }}
                      className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
                    >
                      {t('empty')}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {/* Hint display */}
            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで現れるので変化として扱われず、読み上げられない
              ことがある (#5602)。
            */}
            <div data-tutorial="sultan-hint-display" data-testid="sultan-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromIdx)} →{' '}
                  {t('frontendHint.foundation', { idx: hint.toFoundation })}
                </div>
              )}
            </div>
            <div className="flex justify-center">
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Settings */}
          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          {/* Footer */}
          <GameFooter className={`${gameTheme.sultan.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="sultan-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDraw}
                    disabled={loading || isAutoCompleting || state.stockCount === 0}
                  >
                    {t('draw')}
                  </button>
                  {state.canRedeal && (
                    <button
                      type="button"
                      className={btnPrimary}
                      onClick={handleRedeal}
                      disabled={loading || isAutoCompleting}
                      data-testid="sultan-redeal-count"
                    >
                      {t('redealWithCount', {
                        count: SULTAN_MAX_REDEAL - state.redealCount,
                      })}
                    </button>
                  )}
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleUndo}
                    disabled={loading || isAutoCompleting || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={handleUndoEscape}
                      disabled={loading || isAutoCompleting}
                    />
                  )}
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleHint}
                    disabled={loading || isAutoCompleting}
                  >
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess}${autoCompleteReady && !loading && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={handleAutoComplete}
                    disabled={loading || isAutoCompleting || !autoCompleteReady}
                    data-testid="autocomplete-button"
                    title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
                  >
                    {t('autoComplete')}
                  </button>
                  <button
                    type="button"
                    className={btnDanger}
                    onClick={confirmGiveUpAction}
                    disabled={loading || isAutoCompleting}
                  >
                    {t('giveup')}
                  </button>
                </div>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sultan-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="sultan-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
