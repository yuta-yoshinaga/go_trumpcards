import { useCallback, useEffect, useMemo, useState } from 'react';
import { pokersquaresApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, PokerSquaresResponse } from '../types/card';
import { PokerSquaresPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { POKERSQUARES_HELP, parsePokerSquaresCommand } from '../utils/cli/commands/pokersquaresCommands';
import { formatPokerSquaresState } from '../utils/cli/formatters/pokersquaresFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { evaluateFiveCardHand, pokerHandKey, pokerSquaresRankToScore } from '../utils/pokerSquaresUtils';

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
export const PokerSquaresPage = withTutorial(PokerSquaresPageContent, 'pokersquares', POKERSQUARES_TUTORIAL_STEPS);
/** Inner content of the Poker Squares page, wrapped by TutorialProvider. */
function PokerSquaresPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pokersquares');
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(pokersquaresApi.exec);
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

  useMountReset(execApi);

  const phaseNames = usePhaseNames('pokersquares', POKER_SQUARES_PHASE_KEYS);

  const isComplete = state?.phase === PokerSquaresPhase.COMPLETE;
  const isPlaying = state?.phase === PokerSquaresPhase.PLAYING;

  // Cross-highlight: while the player hovers/focuses an empty cell, highlight its row + column
  // and the corresponding row/col score badges so they can preview which lines a placement affects.
  const [crossHover, setCrossHover] = useState<{ row: number; col: number } | null>(null);
  const clearCrossHover = useCallback(() => setCrossHover(null), []);

  // Disabled buttons do not always fire pointerleave/blur, so the hover state can get stuck after
  // a click (cell becomes filled) or while an API call is in flight. Clear defensively whenever
  // the loading flag flips on or the hovered cell becomes filled by the latest server state.
  useEffect(() => {
    if (loading) {
      setCrossHover(null);
      return;
    }
    if (crossHover && state?.board[crossHover.row]?.[crossHover.col]?.card) {
      setCrossHover(null);
    }
  }, [loading, state?.board, crossHover]);

  // While the player hovers an empty cell, compute the (hypothetical) row + col scores
  // they would lock in by placing `currentCard` there. We can only score full lines, so
  // the preview surfaces only when this placement would complete the row or column.
  const preview = useMemo(() => {
    if (!state || !crossHover || !state.currentCard) return null;
    const board = state.board;
    if (board[crossHover.row]?.[crossHover.col]?.card) return null;

    const placedCard = state.currentCard;
    const rowCards: (Card | null)[] = board[crossHover.row].map((c, ci) =>
      ci === crossHover.col ? placedCard : c.card,
    );
    const colCards: (Card | null)[] = board.map((r, ri) =>
      ri === crossHover.row ? placedCard : r[crossHover.col].card,
    );

    const rowComplete: Card[] | null = rowCards.every((c): c is Card => c != null) ? (rowCards as Card[]) : null;
    const colComplete: Card[] | null = colCards.every((c): c is Card => c != null) ? (colCards as Card[]) : null;
    const rowRank = rowComplete ? evaluateFiveCardHand(rowComplete) : null;
    const colRank = colComplete ? evaluateFiveCardHand(colComplete) : null;
    return {
      row: rowRank != null ? { rank: rowRank, score: pokerSquaresRankToScore(rowRank) } : null,
      col: colRank != null ? { rank: colRank, score: pokerSquaresRankToScore(colRank) } : null,
    };
  }, [state, crossHover]);

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
    <GamePageShell
      title={tc('nav.pokersquares')}
      gameThemeBg={gameTheme.pokersquares.bg}
      phaseName={state ? phaseNames[state.phase] : ''}
      gamePath="/pokersquares"
      gameEndFlag={!state || isComplete}
      winShow={isComplete}
      onCelebrate={() => playSound('winFanfare')}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          {state && (
            <span>
              {t('label.totalScore')}: {state.totalScore}
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
                                    style={{
                                      width: cardWidth,
                                      height: Math.round(cardWidth * 1.4),
                                    }}
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
                        style={{ gridTemplateRows: `repeat(5, minmax(0, 1fr))` }}
                        data-testid="ps-row-scores"
                        data-tutorial="ps-scores"
                      >
                        {state.rowScores.map((s, i) => {
                          const highlighted = crossHover?.row === i;
                          const rowPreview = highlighted ? preview?.row : null;
                          return (
                            <div
                              key={`row-score-${i}`}
                              data-testid={`row-score-${i}`}
                              data-cross-hover={highlighted ? 'true' : undefined}
                              style={{ height: Math.round(cardWidth * 1.4) }}
                              className={`flex flex-col items-center justify-center text-sm font-mono px-2 rounded min-w-[3rem] ${
                                highlighted
                                  ? 'text-ds-accent bg-ds-accent/15 ring-1 ring-ds-accent'
                                  : 'text-ds-text-primary bg-black/30'
                              }`}
                            >
                              <span>{s}</span>
                              {rowPreview && rowPreview.score - s !== 0 && (
                                <span
                                  data-testid={`row-score-preview-${i}`}
                                  className="text-[10px] text-ds-success leading-none mt-0.5"
                                >
                                  {`+${rowPreview.score - s}`}
                                  <br />
                                  {t(`hand.${pokerHandKey(rowPreview.rank)}`)}
                                </span>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    </div>
                    <div
                      className="grid gap-1 mt-1"
                      style={{ gridTemplateColumns: `repeat(5, minmax(0, 1fr))` }}
                      data-testid="ps-col-scores"
                    >
                      {state.colScores.map((s, i) => {
                        const highlighted = crossHover?.col === i;
                        const colPreview = highlighted ? preview?.col : null;
                        return (
                          <div
                            key={`col-score-${i}`}
                            data-testid={`col-score-${i}`}
                            data-cross-hover={highlighted ? 'true' : undefined}
                            style={{ width: cardWidth }}
                            className={`text-center text-sm font-mono py-1 rounded ${
                              highlighted
                                ? 'text-ds-accent bg-ds-accent/15 ring-1 ring-ds-accent'
                                : 'text-ds-text-primary bg-black/30'
                            }`}
                          >
                            {s}
                            {colPreview && colPreview.score - s !== 0 && (
                              <div
                                data-testid={`col-score-preview-${i}`}
                                className="text-[10px] text-ds-success leading-none mt-0.5"
                              >
                                +{colPreview.score - s}
                                <div>{t(`hand.${pokerHandKey(colPreview.rank)}`)}</div>
                              </div>
                            )}
                          </div>
                        );
                      })}
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
    </GamePageShell>
  );
}
