import { useCallback, useEffect, useMemo, useState } from 'react';
import { cribbagesquaresApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnDanger, btnPrimary, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CribbageSquaresResponse, CribbageSquaresScore } from '../types/card';
import { CribbageSquaresPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CRIBBAGESQUARES_HELP, parseCribbageSquaresCommand } from '../utils/cli/commands/cribbagesquaresCommands';
import { formatCribbageSquaresState } from '../utils/cli/formatters/cribbagesquaresFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

const GRID_SIZE = 4;
const TOTAL_CELLS = GRID_SIZE * GRID_SIZE;

const CRIBBAGE_SQUARES_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CribbageSquaresPhase.PLAYING]: 'playing',
  [CribbageSquaresPhase.COMPLETE]: 'complete',
};

/**
 * Lists the components of a hand that actually scored.
 *
 * Returns `[]` for a hand worth nothing, so the caller can say so once rather
 * than render "15s 0, pairs 0, runs 0" and bury the points that did land.
 */
export function cribbageBreakdownParts(
  detail: CribbageSquaresScore | undefined,
  label: (key: string, n: number) => string,
): string[] {
  if (!detail) return [];
  const parts: string[] = [];
  if (detail.fifteens > 0) parts.push(label('fifteens', detail.fifteens));
  if (detail.pairs > 0) parts.push(label('pairs', detail.pairs));
  if (detail.runs > 0) parts.push(label('runs', detail.runs));
  if (detail.flush > 0) parts.push(label('flush', detail.flush));
  if (detail.nobs > 0) parts.push(label('nobs', detail.nobs));
  return parts;
}

/** Cribbage Squares tutorial step definitions. */
const CRIBBAGESQUARES_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cs-current-card"]',
    messageKey: 'tutorial.currentCard',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="cs-starter"]', messageKey: 'tutorial.starter', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="cs-board"]', messageKey: 'tutorial.board', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="cs-scores"]', messageKey: 'tutorial.scores', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="cs-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Cribbage Squares page: a 4x4 grid scored as eight cribbage hands. */
export const CribbageSquaresPage = withTutorial(
  CribbageSquaresPageContent,
  'cribbagesquares',
  CRIBBAGESQUARES_TUTORIAL_STEPS,
);
/** Inner content of the Cribbage Squares page, wrapped by TutorialProvider. */
function CribbageSquaresPageContent() {
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
  } = useGamePageSetup('cribbagesquares');
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(cribbagesquaresApi.exec);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('cribbagesquares', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cribbagesquares');
  const cliConfig: CliGameConfig<CribbageSquaresResponse, Parameters<typeof cribbagesquaresApi.exec>> = useMemo(
    () => ({
      gameName: 'cribbagesquares',
      parseCommand: parseCribbageSquaresCommand,
      formatResponse: formatCribbageSquaresState,
      helpText: CRIBBAGESQUARES_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phaseNames = usePhaseNames('cribbagesquares', CRIBBAGE_SQUARES_PHASE_KEYS);

  const isComplete = state?.phase === CribbageSquaresPhase.COMPLETE;
  const isPlaying = state?.phase === CribbageSquaresPhase.PLAYING;

  // Highlight the row and column a hovered/focused empty cell belongs to, so
  // the player can see which two hands a placement would join.
  const [crossHover, setCrossHover] = useState<{ row: number; col: number } | null>(null);
  const clearCrossHover = useCallback(() => setCrossHover(null), []);

  // Disabled buttons do not reliably fire pointerleave/blur, so the hover state
  // can stick after a click fills the cell or while a call is in flight.
  useEffect(() => {
    if (loading) {
      setCrossHover(null);
      return;
    }
    if (crossHover && state?.board[crossHover.row]?.[crossHover.col]?.card) {
      setCrossHover(null);
    }
  }, [loading, state?.board, crossHover]);

  const breakdownLabel = useCallback((key: string, n: number) => t(`part.${key}`, { n }), [t]);

  const handlePlace = (row: number, col: number) => {
    execApi('place', row, col);
  };
  const handleUndo = () => execApi('undo');
  const handleHint = () => execApi('hint');
  const handleGiveUp = useCallback(() => execApi('giveup'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);
  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  return (
    <GamePageShell
      title={tc('nav.cribbagesquares')}
      gameThemeBg={gameTheme.cribbagesquares.bg}
      phaseName={state ? phaseNames[state.phase] : ''}
      gamePath="/cribbagesquares"
      gameEndFlag={!state || isComplete}
      // Only a board that reached the target is a win; finishing at 40 points
      // is a completed game, not a celebration.
      winShow={!!state?.isWin}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      giveUpConfirmOpen={giveUpConfirmOpen}
      confirmGiveUp={confirmGiveUp}
      cancelGiveUp={cancelGiveUp}
      headerExtra={
        <>
          {state && (
            <span>
              {t('label.totalScore')}: {state.totalScore} / {state.winScore}
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
          <SettingsPanel
            title={t('settings.title')}
            groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
          />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {state && (
              <>
                <div className="flex justify-center gap-6 mb-3">
                  <div className="flex flex-col items-center" data-tutorial="cs-current-card">
                    <div className="text-game-text-muted text-xs mb-1">
                      {t('label.currentCard')} ({state.placedCount}/{TOTAL_CELLS})
                    </div>
                    {state.currentCard ? (
                      <AnimatedCard card={state.currentCard} width={cardWidth} />
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

                  <div className="flex flex-col items-center" data-tutorial="cs-starter">
                    <div className="text-game-text-muted text-xs mb-1">{t('label.starter')}</div>
                    {state.starter ? (
                      <AnimatedCard card={state.starter} width={cardWidth} />
                    ) : (
                      // Face down rather than absent: the starter existing but
                      // being unknown is the rule the whole game turns on.
                      <div data-testid="cs-starter-facedown" title={t('label.starterHidden')}>
                        <AnimatedCardBack width={cardWidth} />
                      </div>
                    )}
                  </div>
                </div>

                <div className="flex justify-center mb-3" data-tutorial="cs-board">
                  <div className="inline-flex flex-col items-start">
                    <div className="flex">
                      <div
                        className="grid gap-1"
                        style={{ gridTemplateColumns: `repeat(${GRID_SIZE}, minmax(0, 1fr))` }}
                        data-testid="cs-board"
                      >
                        {state.board.map((row, rowIdx) =>
                          row.map((cell, colIdx) => {
                            const filled = !!cell.card;
                            const cellAction = `cell-${rowIdx}-${colIdx}`;
                            const isHintTarget =
                              frontendHintEnabled && frontendHint?.targetAction === cellAction && !filled;
                            const inCross =
                              crossHover !== null && (crossHover.row === rowIdx || crossHover.col === colIdx);
                            return (
                              <button
                                type="button"
                                key={`cell-${rowIdx}-${colIdx}`}
                                data-testid={`cell-${rowIdx}-${colIdx}`}
                                data-hint-action={cellAction}
                                data-cross-hover={inCross ? 'true' : undefined}
                                aria-label={
                                  cell.card ? cardAlt(cell.card) : `${t('label.empty')} ${rowIdx + 1}-${colIdx + 1}`
                                }
                                onClick={() => handlePlace(rowIdx, colIdx)}
                                onPointerEnter={() => !filled && setCrossHover({ row: rowIdx, col: colIdx })}
                                onPointerLeave={clearCrossHover}
                                onFocus={() => !filled && setCrossHover({ row: rowIdx, col: colIdx })}
                                onBlur={clearCrossHover}
                                disabled={!isPlaying || loading || filled || !state.currentCard}
                                className={`p-0 border-0 bg-transparent rounded ${focusRingWhite} ${
                                  filled ? '' : 'cursor-pointer'
                                } ${isHintTarget ? 'ring-2 ring-ds-warning' : ''}${inCross ? ' bg-white/10' : ''}`}
                              >
                                {cell.card ? (
                                  <AnimatedCard card={cell.card} width={cardWidth} />
                                ) : (
                                  <div
                                    style={{ width: cardWidth, height: Math.round(cardWidth * 1.4) }}
                                    className={`rounded border-2 border-dashed flex items-center justify-center text-game-text-muted text-xs ${
                                      inCross ? 'border-ds-accent' : 'border-white/30'
                                    }`}
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
                        style={{ gridTemplateRows: `repeat(${GRID_SIZE}, minmax(0, 1fr))` }}
                        data-testid="cs-row-scores"
                        data-tutorial="cs-scores"
                      >
                        {state.rowScores.map((s, i) => {
                          const parts = cribbageBreakdownParts(state.rowDetails?.[i], breakdownLabel);
                          return (
                            <div
                              key={`row-score-${i}`}
                              data-testid={`row-score-${i}`}
                              data-cross-hover={crossHover?.row === i ? 'true' : undefined}
                              title={parts.join(' ')}
                              style={{ height: Math.round(cardWidth * 1.4) }}
                              className={`flex flex-col items-center justify-center text-sm font-mono px-2 rounded min-w-[3rem] ${
                                crossHover?.row === i
                                  ? 'text-ds-accent bg-ds-accent/15 ring-1 ring-ds-accent'
                                  : 'text-ds-text-primary bg-black/30'
                              }`}
                            >
                              <span>{s}</span>
                              {parts.length > 0 && (
                                <div
                                  data-testid={`row-breakdown-${i}`}
                                  className="text-[10px] text-ds-text-muted leading-none mt-0.5 text-center"
                                >
                                  {parts.join(' ')}
                                </div>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    </div>
                    <div
                      className="grid gap-1 mt-1"
                      style={{ gridTemplateColumns: `repeat(${GRID_SIZE}, minmax(0, 1fr))` }}
                      data-testid="cs-col-scores"
                    >
                      {state.colScores.map((s, i) => {
                        const parts = cribbageBreakdownParts(state.colDetails?.[i], breakdownLabel);
                        return (
                          <div
                            key={`col-score-${i}`}
                            data-testid={`col-score-${i}`}
                            data-cross-hover={crossHover?.col === i ? 'true' : undefined}
                            title={parts.join(' ')}
                            style={{ width: cardWidth }}
                            className={`text-center text-sm font-mono py-1 rounded ${
                              crossHover?.col === i
                                ? 'text-ds-accent bg-ds-accent/15 ring-1 ring-ds-accent'
                                : 'text-ds-text-primary bg-black/30'
                            }`}
                          >
                            {s}
                            {parts.length > 0 && (
                              <div
                                data-testid={`col-breakdown-${i}`}
                                className="text-[10px] text-ds-text-muted leading-none mt-0.5"
                              >
                                {parts.join(' ')}
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  </div>
                </div>

                <div className="text-center text-ds-text-primary text-lg font-bold mb-2" data-testid="total-score">
                  {t('label.totalScore')}: {state.totalScore} / {state.winScore}
                  {isComplete && (
                    <span
                      data-testid="cs-verdict"
                      className={`ml-2 text-sm ${state.isWin ? 'text-ds-success' : 'text-ds-text-muted'}`}
                    >
                      {state.isWin ? t('label.win') : t('label.lose')}
                    </span>
                  )}
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

          <GameFooter className={`${gameTheme.cribbagesquares.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            {/* サーバ側のヒント。ボタンを押したときだけ出す。 */}
            {state?.hint && isRequestedHint(state) && (
              <p className="text-center text-sm text-ds-accent mt-1" data-testid="cs-server-hint">
                {t(state.hint.synergy ? 'hint.synergy' : 'hint.plain', {
                  row: state.hint.row,
                  col: state.hint.col,
                  score: state.hint.score,
                })}
              </p>
            )}
            <div className="flex gap-2 items-center flex-wrap" data-tutorial="cs-controls">
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
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleHint}
                    disabled={loading}
                    data-testid="cs-hint-button"
                  >
                    {t('button.hint')}
                  </button>
                  <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={loading}>
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
    </GamePageShell>
  );
}
