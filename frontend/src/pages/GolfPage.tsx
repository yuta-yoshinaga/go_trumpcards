import { useMemo } from 'react';
import type { golfApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
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
import { GolfSkeleton } from '../components/skeleton/GolfSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGolfGame } from '../hooks/useGolfGame';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnOutline, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GolfResponse } from '../types/card';
import { GolfPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { GOLF_HELP, parseGolfCommand } from '../utils/cli/commands/golfCommands';
import { formatGolfState } from '../utils/cli/formatters/golfFormatter';
import type { CliGameConfig } from '../utils/cli/types';

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
export function GolfPage() {
  return (
    <TutorialWrapper gameName="golf" steps={GOLF_TUTORIAL_STEPS}>
      <GolfPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Golf page, wrapped by TutorialProvider. */
function GolfPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('golf');
  const { playSound } = useSound();
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
    handleSelectCard,
  } = useGolfGame();
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

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw },
      { key: 'h', action: handleHint },
      { key: 'g', action: handleGiveUp },
      { key: 'z', action: handleUndo },
    ],
    [handleDraw, handleHint, handleGiveUp, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <GolfSkeleton />;

  const isPlaying = state.phase === GolfPhase.PLAYING;
  const isGameClear = state.phase === GolfPhase.GAME_CLEAR;
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
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.golf.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.golf')} />
      <PhaseIndicator
        phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      >
        <span>
          {t('moveCount')}: {state.moveCount}
        </span>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/golf" />
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
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
                    return (
                      <div key={`gc-${colIdx.toString()}-${rowIdx.toString()}`} className="absolute" style={{ top }}>
                        <button
                          type="button"
                          onClick={() => handleSelectCard(colIdx)}
                          disabled={!isPlaying || loading || !exposed}
                          aria-label={cardAlt(gc.card)}
                          className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                            isHinted && exposed ? 'ring-2 ring-yellow-400' : ''
                          } ${!exposed ? 'opacity-60' : ''}`}
                        >
                          <AnimatedCard
                            card={gc.card}
                            width={effectiveCardWidth}
                            onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                          />
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
                  <AnimatedCardBack
                    width={effectiveCardWidth}
                    onClick={isPlaying ? handleDraw : undefined}
                    ariaLabel={t('draw')}
                    onFlipComplete={() => playSound('cardFlip')}
                  />
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
                  <AnimatedCard
                    card={state.waste[state.waste.length - 1]}
                    width={effectiveCardWidth}
                    onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                  />
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
                <div className="text-yellow-300 text-sm mb-2 text-center">
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

          {/* Settings */}
          <SettingsPanel title={tc('settings.title')} groups={[]} />

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
                  <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading}>
                    {t('hint')}
                  </button>
                  <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                    {t('giveup')}
                  </button>
                </div>
              )}
              <div data-tutorial="golf-reset-button">
                <button
                  type="button"
                  className={btnOutline}
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
            </div>
          </GameFooter>
        </>
      )}
      <WinCelebration show={state.phase === GolfPhase.GAME_CLEAR} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
