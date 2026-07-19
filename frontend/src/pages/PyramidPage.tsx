import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { pyramidApi } from '../api/gameApi';
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
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePyramidGame } from '../hooks/usePyramidGame';
import { usePyramidStats } from '../hooks/usePyramidStats';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PyramidResponse } from '../types/card';
import { PyramidPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PYRAMID_HELP, parsePyramidCommand } from '../utils/cli/commands/pyramidCommands';
import { formatPyramidState } from '../utils/cli/formatters/pyramidFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Pyramid tutorial step definitions. */
const PY_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="py-pyramid"]',
    messageKey: 'tutorial.pyramid',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-hint-display"]',
    messageKey: 'tutorial.hintDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="py-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Pyramid Solitaire game page with pyramid, stock/waste, and controls. */
export const PyramidPage = withTutorial(PyramidPageContent, 'pyramid', PY_TUTORIAL_STEPS);
/** Inner content of the Pyramid page, wrapped by TutorialProvider. */
function PyramidPageContent() {
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
  } = useGamePageSetup('pyramid');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec,
    retry,
    hintError,
    selectedCard,
    hint,
    handleDraw,
    handleReset,
    handleGiveUp,
    handleHint,
    handleUndo,
    handleUndoEscape,
    handleSelectCard,
  } = usePyramidGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('pyramid', state);
  const { cardHeight, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pyramid');
  const cliConfig: CliGameConfig<PyramidResponse, Parameters<typeof pyramidApi.exec>> = useMemo(
    () => ({
      gameName: 'pyramid',
      parseCommand: parsePyramidCommand,
      formatResponse: formatPyramidState,
      helpText: PYRAMID_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === PyramidPhase.PLAYING;

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useCallback(
    () => requestGiveUpConfirm(handleGiveUp),
    [requestGiveUpConfirm, handleGiveUp],
  );

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw },
      { key: 'h', action: handleHint },
      { key: 'g', action: confirmGiveUpAction },
      { key: 'z', action: handleUndo },
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

  // Best-record persistence in localStorage (issue #3083).
  const { stats, recordResult } = usePyramidStats();
  const [newBestMoves, setNewBestMoves] = useState(false);
  // Guard so each finished game is recorded exactly once (phase stays ended across re-renders).
  const recordedRef = useRef(false);
  const endPhase = state?.phase;
  const endMoveCount = state?.moveCount ?? 0;
  useEffect(() => {
    const ended = endPhase === PyramidPhase.GAME_CLEAR || endPhase === PyramidPhase.GAME_OVER;
    if (!ended) {
      recordedRef.current = false;
      return;
    }
    if (recordedRef.current) return;
    recordedRef.current = true;
    setNewBestMoves(recordResult({ won: endPhase === PyramidPhase.GAME_CLEAR, moves: endMoveCount }));
  }, [endPhase, endMoveCount, recordResult]);

  if (!state)
    return <GameSkeleton gameKey="pyramid" layout={{ kind: 'tiered-rows', rows: [1, 2, 3, 4], stockWaste: true }} />;

  const isPlaying = state.phase === PyramidPhase.PLAYING;
  const isGameClear = state.phase === PyramidPhase.GAME_CLEAR;
  const isGameOver = state.phase === PyramidPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSelected = (zone: 'pyramid' | 'waste', row?: number, col?: number) =>
    selectedCard !== null && selectedCard.zone === zone && selectedCard.row === row && selectedCard.col === col;

  // When the player has selected the first card, compute the partner value (13 - value).
  // Kings (13) need no partner; ace (1) wants Q (12), etc.
  const selectedValue = (() => {
    if (!selectedCard) return null;
    if (selectedCard.zone === 'waste') {
      return state.waste.length > 0 ? state.waste[state.waste.length - 1].value : null;
    }
    if (selectedCard.row === undefined || selectedCard.col === undefined) return null;
    const pc = state.pyramid[selectedCard.row]?.[selectedCard.col];
    return pc?.card?.value ?? null;
  })();
  const partnerValue = selectedValue !== null && selectedValue < 13 ? 13 - selectedValue : null;
  const wasteTopCard = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const isWastePairCandidate =
    partnerValue !== null && !isSelected('waste') && wasteTopCard !== null && wasteTopCard.value === partnerValue;
  // The waste top is always exposed; a King there is removable alone (issue #3082).
  const isWasteExposedKing = wasteTopCard !== null && wasteTopCard.value === 13 && !isSelected('waste');

  // Calculate pyramid layout dimensions
  const maxCols = 7; // bottom row has 7 cards
  const cardGap = 4;
  /** Fraction of card height used for vertical overlap between rows (less on mobile for bigger tap targets) */
  const ROW_OVERLAP_RATIO = isMobile ? 0.3 : 0.35;
  const rowOverlap = cardHeight * ROW_OVERLAP_RATIO;
  // px-4 on the scrollable container = 16px * 2 = 32px total horizontal padding
  const CONTAINER_PADDING = 32;
  const effectiveCardWidth = isMobile
    ? Math.floor((windowWidth - CONTAINER_PADDING - cardGap * (maxCols - 1)) / maxCols)
    : cardWidth;
  const pyramidWidth = maxCols * (effectiveCardWidth + cardGap) - cardGap;

  return (
    <GamePageShell
      title={tc('nav.pyramid')}
      gameThemeBg={gameTheme.pyramid.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/pyramid"
      gameEndFlag={isEnded}
      winShow={isGameClear}
      onCelebrate={() => playSound('winFanfare')}
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
          {stats.fewestMoves !== null && (
            <span data-testid="py-best-moves">
              {t('bestMoves')}: {t('bestMovesValue', { moves: stats.fewestMoves })}
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

          {/* Scrollable area */}
          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* Pyramid */}
            <div data-tutorial="py-pyramid" className="flex flex-col items-center mb-3">
              {state.pyramid.map((row, rowIdx) => {
                const cols = row.length;
                const rowWidth = cols * (effectiveCardWidth + cardGap) - cardGap;
                const offsetX = (pyramidWidth - rowWidth) / 2;
                return (
                  <div
                    key={`row-${rowIdx.toString()}`}
                    className="relative"
                    style={{
                      height: rowIdx < state.pyramid.length - 1 ? cardHeight - rowOverlap : cardHeight,
                      width: pyramidWidth,
                    }}
                  >
                    {row.map((pc, colIdx) => {
                      const left = offsetX + colIdx * (effectiveCardWidth + cardGap);
                      if (pc.removed) {
                        return (
                          <div
                            key={`pc-${rowIdx.toString()}-${colIdx.toString()}`}
                            className="absolute"
                            style={{ left, width: effectiveCardWidth, height: cardHeight }}
                          />
                        );
                      }
                      if (!pc.card) return null;
                      const exposed = pc.exposed;
                      const isPairCandidate =
                        partnerValue !== null &&
                        exposed &&
                        pc.card.value === partnerValue &&
                        !isSelected('pyramid', rowIdx, colIdx);
                      // A King (13) is removable alone once exposed — surface that
                      // affordance so players don't hunt for a nonexistent partner
                      // (issue #3082). Selected/hint states take visual precedence.
                      const isExposedKing = exposed && pc.card.value === 13 && !isSelected('pyramid', rowIdx, colIdx);
                      // Server hint targets (-1 sentinels for king/waste never match a cell).
                      const isHintTarget =
                        !!hint &&
                        ((hint.row1 === rowIdx && hint.col1 === colIdx) ||
                          (hint.row2 === rowIdx && hint.col2 === colIdx));
                      // Convey the card's actionability (blocked / selected / pair
                      // candidate) to assistive tech, not just via disabled/color.
                      const cellSelected = isSelected('pyramid', rowIdx, colIdx);
                      const statusSuffix = !exposed
                        ? ` ${t('a11y.blocked')}`
                        : cellSelected
                          ? ` ${t('a11y.selected')}`
                          : isPairCandidate
                            ? ` ${t('a11y.pairCandidate')}`
                            : isExposedKing
                              ? ` ${t('a11y.kingRemovable')}`
                              : '';
                      return (
                        <div key={`pc-${rowIdx.toString()}-${colIdx.toString()}`} className="absolute" style={{ left }}>
                          <button
                            type="button"
                            onClick={() => {
                              if (!exposed || !pc.card) return;
                              handleSelectCard({ zone: 'pyramid', row: rowIdx, col: colIdx }, pc.card.value);
                            }}
                            disabled={!isPlaying || loading || !exposed}
                            aria-label={`${cardAlt(pc.card)}${statusSuffix}`}
                            aria-pressed={cellSelected}
                            data-pair-candidate={isPairCandidate ? 'true' : undefined}
                            data-king-removable={isExposedKing ? 'true' : undefined}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                              isSelected('pyramid', rowIdx, colIdx)
                                ? 'ring-2 ring-ds-warning'
                                : isHintTarget
                                  ? 'ring-2 ring-ds-warning animate-pulse'
                                  : isPairCandidate
                                    ? 'ring-2 ring-ds-success animate-pulse'
                                    : isExposedKing
                                      ? 'ring-2 ring-ds-success'
                                      : ''
                            } ${!exposed ? 'opacity-60' : ''}`}
                          >
                            <AnimatedCard card={pc.card} width={effectiveCardWidth} />
                          </button>
                        </div>
                      );
                    })}
                  </div>
                );
              })}
            </div>

            {/* Stock + Waste */}
            <div className="flex gap-4 justify-center mb-3" data-tutorial="py-stock-waste">
              {/* Stock */}
              <div className="text-center">
                <div className="text-game-text-muted text-xs mb-1">
                  {t('stock')} ({state.stockCount})
                </div>
                {state.stockCount > 0 ? (
                  <AnimatedCardBack
                    width={effectiveCardWidth}
                    onClick={isPlaying ? handleDraw : undefined}
                    ariaLabel={t('draw')}
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

              {/* Waste */}
              <div className="text-center">
                <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
                {wasteTopCard ? (
                  <button
                    type="button"
                    onClick={() => handleSelectCard({ zone: 'waste' }, wasteTopCard.value)}
                    disabled={!isPlaying || loading}
                    aria-label={`${cardAlt(wasteTopCard)}${isWasteExposedKing ? ` ${t('a11y.kingRemovable')}` : ''}`}
                    aria-pressed={isSelected('waste')}
                    data-pair-candidate={isWastePairCandidate ? 'true' : undefined}
                    data-king-removable={isWasteExposedKing ? 'true' : undefined}
                    className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                      isSelected('waste') ? 'ring-2 ring-ds-warning' : ''
                    } ${isWastePairCandidate ? 'ring-2 ring-ds-success animate-pulse' : isWasteExposedKing ? 'ring-2 ring-ds-success' : ''}`}
                  >
                    <AnimatedCard card={wasteTopCard} width={effectiveCardWidth} />
                  </button>
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
            <div data-tutorial="py-hint-display">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 text-center">
                  {t('hintAvailable')}: {t(`hintType.${hint.type}`)}
                </div>
              )}
            </div>
            {frontendHintEnabled && frontendHint && (
              <div className="flex justify-center">
                <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
              </div>
            )}

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* New fewest-moves record badge on the clear screen (#3083). */}
            {isGameClear && newBestMoves && (
              <div
                data-testid="py-best-badge"
                role="status"
                className="text-center text-ds-success font-semibold text-sm mb-2"
              >
                {t('newBestMoves', { moves: state.moveCount })}
              </div>
            )}

            {/* Best-record panel: fewest moves + clear tally (#3083). */}
            <div data-testid="py-stats-panel" className="text-game-text-muted text-xs text-center mb-2">
              {t('bestMoves')}: {stats.fewestMoves !== null ? t('bestMovesValue', { moves: stats.fewestMoves }) : '—'}
              {stats.plays > 0 && (
                <>
                  {' · '}
                  {t('clears', { wins: stats.wins, plays: stats.plays })}
                </>
              )}
            </div>

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

          {/* Footer */}
          <GameFooter className={`${gameTheme.pyramid.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="py-controls">
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
                dataTutorial="py-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
