import { useCallback, useEffect, useMemo } from 'react';
import { pokersquaresApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetButton } from '../components/GameResetButton';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PokerSquaresResponse } from '../types/card';
import { PokerSquaresPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { POKERSQUARES_HELP, parsePokerSquaresCommand } from '../utils/cli/commands/pokersquaresCommands';
import { formatPokerSquaresState } from '../utils/cli/formatters/pokersquaresFormatter';
import type { CliGameConfig } from '../utils/cli/types';

const POKER_SQUARES_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PokerSquaresPhase.PLAYING]: 'playing',
  [PokerSquaresPhase.COMPLETE]: 'complete',
};

/** Poker Squares tutorial step definitions. */
const POKERSQUARES_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ps-current-card"]',
    messageKey: 'tutorial.currentCard',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ps-board"]', messageKey: 'tutorial.board', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="ps-scores"]', messageKey: 'tutorial.scores', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="ps-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Poker Squares game page with a 5x5 placement grid and per-row/column scores. */
export function PokerSquaresPage() {
  return (
    <TutorialWrapper gameName="pokersquares" steps={POKERSQUARES_TUTORIAL_STEPS}>
      <PokerSquaresPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Poker Squares page, wrapped by TutorialProvider. */
function PokerSquaresPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pokersquares');
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(pokersquaresApi.exec);
  // Issue #1609: warn before tab close / reload while a round is in progress.
  useGameRoundGuard(!!state && !state.gameEndFlag);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('pokersquares', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pokersquares');
  const cliConfig: CliGameConfig<PokerSquaresResponse, Parameters<typeof pokersquaresApi.exec>> = useMemo(
    () => ({
      gameName: 'pokersquares',
      parseCommand: parsePokerSquaresCommand,
      formatResponse: formatPokerSquaresState,
      helpText: POKERSQUARES_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

  const phaseNames = usePhaseNames('pokersquares', POKER_SQUARES_PHASE_KEYS);

  const isComplete = state?.phase === PokerSquaresPhase.COMPLETE;
  const isPlaying = state?.phase === PokerSquaresPhase.PLAYING;

  const handlePlace = (row: number, col: number) => {
    execApi('place', row, col);
  };
  const handleUndo = () => execApi('undo');
  const handleGiveUp = () => execApi('giveup');
  const handleReset = useCallback(() => execApi('reset'), [execApi]);
  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.pokersquares.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.pokersquares')} />
      <PhaseIndicator phaseName={state ? phaseNames[state.phase] : ''}>
        {state && (
          <span>
            {t('label.totalScore')}: {state.totalScore}
          </span>
        )}
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/pokersquares" />
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'checkbox',
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {state && (
              <>
                <div className="flex flex-col items-center mb-3" data-tutorial="ps-current-card">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('label.currentCard')} ({state.placedCount}/25)
                  </div>
                  {state.currentCard ? (
                    <AnimatedCard
                      card={state.currentCard}
                      width={cardWidth}
                      onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                    />
                  ) : (
                    <div
                      style={{ width: cardWidth, height: Math.round(cardWidth * 1.4) }}
                      className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
                      data-testid="current-card-empty"
                    >
                      {t('label.empty')}
                    </div>
                  )}
                </div>

                <div className="flex justify-center mb-3" data-testid="ps-board-wrapper" data-tutorial="ps-board">
                  <div className="inline-flex flex-col items-start">
                    <div className="flex">
                      <div
                        className="grid gap-1"
                        style={{ gridTemplateColumns: `repeat(5, minmax(0, 1fr))` }}
                        data-testid="ps-board"
                      >
                        {state.board.map((row, rowIdx) =>
                          row.map((cell, colIdx) => {
                            const filled = !!cell.card;
                            const cellAction = `cell-${rowIdx}-${colIdx}`;
                            const isHintTarget =
                              frontendHintEnabled && frontendHint?.targetAction === cellAction && !filled;
                            return (
                              <button
                                type="button"
                                key={`cell-${rowIdx}-${colIdx}`}
                                data-testid={`cell-${rowIdx}-${colIdx}`}
                                data-hint-action={cellAction}
                                aria-label={
                                  cell.card ? cardAlt(cell.card) : `${t('label.empty')} ${rowIdx + 1}-${colIdx + 1}`
                                }
                                onClick={() => handlePlace(rowIdx, colIdx)}
                                disabled={!isPlaying || loading || filled || !state.currentCard}
                                className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${
                                  filled ? '' : 'cursor-pointer'
                                } ${isHintTarget ? 'ring-2 ring-ds-warning' : ''}`}
                              >
                                {cell.card ? (
                                  <AnimatedCard card={cell.card} width={cardWidth} />
                                ) : (
                                  <div
                                    style={{
                                      width: cardWidth,
                                      height: Math.round(cardWidth * 1.4),
                                    }}
                                    className="rounded border-2 border-dashed border-white/30 flex items-center justify-center text-game-text-muted text-xs"
                                  >
                                    +
                                  </div>
                                )}
                              </button>
                            );
                          }),
                        )}
                      </div>
                      <div
                        className="ml-2 grid gap-1"
                        style={{ gridTemplateRows: `repeat(5, minmax(0, 1fr))` }}
                        data-testid="ps-row-scores"
                        data-tutorial="ps-scores"
                      >
                        {state.rowScores.map((s, i) => (
                          <div
                            key={`row-score-${i}`}
                            data-testid={`row-score-${i}`}
                            style={{ height: Math.round(cardWidth * 1.4) }}
                            className="flex items-center justify-center text-ds-text-primary text-sm font-mono px-2 rounded bg-black/30 min-w-[2.5rem]"
                          >
                            {s}
                          </div>
                        ))}
                      </div>
                    </div>
                    <div
                      className="grid gap-1 mt-1"
                      style={{ gridTemplateColumns: `repeat(5, minmax(0, 1fr))` }}
                      data-testid="ps-col-scores"
                    >
                      {state.colScores.map((s, i) => (
                        <div
                          key={`col-score-${i}`}
                          data-testid={`col-score-${i}`}
                          style={{ width: cardWidth }}
                          className="text-center text-ds-text-primary text-sm font-mono py-1 rounded bg-black/30"
                        >
                          {s}
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

                <div className="text-center text-ds-text-primary text-lg font-bold mb-2" data-testid="total-score">
                  {t('label.totalScore')}: {state.totalScore}
                </div>

                <GameMessageBox
                  message={state.message}
                  messageCode={state.messageCode}
                  messageParams={state.messageParams}
                />

                <ActionLogSection
                  isEndPhase={isComplete}
                  actionLog={actionLog}
                  showActionLog={showActionLog}
                  hideActionLog={hideActionLog}
                />
              </>
            )}
          </div>

          <GameFooter className={`${gameTheme.pokersquares.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}
            <div className="flex gap-2 items-center flex-wrap" data-tutorial="ps-controls">
              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleUndo}
                    disabled={loading || !state?.canUndo}
                  >
                    {t('button.undo')}
                  </button>
                  <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                    {t('button.giveup')}
                  </button>
                </>
              )}
              <GameResetButton
                isGameEnd={isComplete}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
              />
            </div>
          </GameFooter>
        </>
      )}
      <WinCelebration show={isComplete} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
