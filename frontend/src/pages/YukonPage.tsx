import { useCallback, useMemo, useState } from 'react';
import { type YukonMoveZone, yukonApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { DropZone } from '../components/DropZone';
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
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnOutline, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { YukonResponse } from '../types/card';
import { YukonPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseYukonCommand, YUKON_HELP } from '../utils/cli/commands/yukonCommands';
import type { CliGameConfig } from '../utils/cli/types';
import { isTableauAllFaceUp } from '../utils/solitaireUtils';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;
// Localized suit-name keys parallel to FOUNDATION_SUITS, used for aria-labels so
// screen readers announce "スペード / Spades" rather than the bare "♠" glyph (#3153).
const FOUNDATION_SUIT_KEYS = ['spade', 'club', 'heart', 'diamond'] as const;
const noop = () => {};

/** Yukon tutorial step definitions. */
const YK_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="yk-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="yk-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="yk-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="yk-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** CLI help text for Yukon. */
/** Format Yukon state for CLI display. */
function formatYukonState(state: YukonResponse): string {
  const lines: string[] = [];
  lines.push('Foundation:');
  for (let i = 0; i < state.foundation.length; i++) {
    const pile = state.foundation[i];
    const top = pile.length > 0 ? `${pile[pile.length - 1].design}-${pile[pile.length - 1].value}` : 'empty';
    lines.push(`  ${FOUNDATION_SUITS[i]}: ${top} (${pile.length})`);
  }
  lines.push('');
  lines.push('Tableau:');
  for (let col = 0; col < state.tableau.length; col++) {
    const cards = state.tableau[col]
      .map((tc, i) => (tc.faceUp && tc.card ? `[${i}]${tc.card.design}-${tc.card.value}` : `[${i}]??`))
      .join(' ');
    lines.push(`  ${col}: ${cards || '(empty)'}`);
  }
  lines.push('');
  lines.push(`Moves: ${state.moveCount}  Phase: ${state.phase}`);
  return lines.join('\n');
}

/** Renders the Yukon solitaire game page. */
export const YukonPage = withTutorial(YukonPageContent, 'yukon', YK_TUTORIAL_STEPS);
/** Inner content of the Yukon page. */
function YukonPageContent() {
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
  } = useGamePageSetup('yukon');
  const { playSound } = useSound();
  const {
    state,
    setState,
    loading,
    error,
    exec: apiExec,
    retry,
  } = useGameApi<YukonResponse, Parameters<typeof yukonApi.exec>>((...args) => yukonApi.exec(...args));

  useMountReset(apiExec);

  const [selectedSource, setSelectedSource] = useState<YukonMoveZone | null>(null);
  /**
   * Yukon allows grabbing any face-up card and dragging it together with every
   * card sitting above it (regardless of suit/rank order). On hover we glow
   * the entire moving block so the player can see what they're about to lift.
   */
  const [hoveredBlock, setHoveredBlock] = useState<{ col: number; cardIdx: number } | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('yukon', state);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('yukon');
  const yukonCliConfig: CliGameConfig<YukonResponse, Parameters<typeof yukonApi.exec>> = useMemo(
    () => ({
      gameName: 'yukon',
      parseCommand: parseYukonCommand,
      formatResponse: formatYukonState,
      helpText: YUKON_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiExec, yukonCliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  // Responsive card dimensions
  const yk = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    const cols = 7;
    const colW = Math.floor((windowWidth - padX - (cols - 1) * gapPx) / cols);
    const cw = Math.min(Math.max(colW, 28), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.48);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  // Drag-and-drop
  const dispatchMove = useCallback(
    (source: YukonMoveZone, target: YukonMoveZone) => {
      void apiExec('move', source, target);
    },
    [apiExec],
  );
  const dnd = useSolitaireDragDrop<YukonMoveZone>({
    onMove: dispatchMove,
    isPlaying: state?.phase === YukonPhase.PLAYING,
    disabled: loading,
  });

  // Action handlers
  const handleManualReset = useCallback(() => {
    void apiExec('reset');
    playSound('shuffle');
  }, [apiExec, playSound]);

  const handleGiveUp = useCallback(() => {
    void apiExec('giveup');
  }, [apiExec]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useCallback(
    () => requestGiveUpConfirm(handleGiveUp),
    [requestGiveUpConfirm, handleGiveUp],
  );

  const handleHint = useCallback(async () => {
    const res = await yukonApi.exec('hint');
    setState((prev) => (prev ? { ...prev, hint: res.hint } : prev));
  }, [setState]);

  const handleAutoComplete = useCallback(() => {
    void apiExec('autocomplete');
  }, [apiExec]);

  const handleUndo = useCallback(() => {
    void apiExec('undo');
  }, [apiExec]);

  const handleUndoEscape = useCallback(
    (n: number) => {
      void apiExec('undo_n', undefined, undefined, n);
    },
    [apiExec],
  );

  const handleSelectSource = useCallback(
    (zone: string, col: number, cardIndex: number) => {
      if (
        selectedSource &&
        selectedSource.zone === zone &&
        selectedSource.col === col &&
        selectedSource.cardIndex === cardIndex
      ) {
        setSelectedSource(null);
        return;
      }
      setSelectedSource({ zone, col, cardIndex });
    },
    [selectedSource],
  );

  const handleSelectTarget = useCallback(
    (zone: string, col: number) => {
      if (!selectedSource) return;
      void apiExec('move', selectedSource, { zone, col });
      setSelectedSource(null);
      playSound('cardPlace');
    },
    [apiExec, selectedSource, playSound],
  );

  const isPlayingForKbd = state?.phase === YukonPhase.PLAYING;

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: confirmGiveUpAction },
      { key: 'z', action: handleUndo },
    ],
    [handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state) return <GameSkeleton gameKey="yukon" layout={{ kind: 'tableau', topRow: 6, tableau: 7 }} />;

  const isPlaying = state.phase === YukonPhase.PLAYING;
  const isGameClear = state.phase === YukonPhase.GAME_CLEAR;
  const isGameOver = state.phase === YukonPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const autoCompleteReady = isTableauAllFaceUp(state.tableau);
  const hintDest = state.hint
    ? state.hint.toZone === 'foundation'
      ? t('foundation')
      : `${t('tableau')} ${state.hint.toCol}`
    : '';
  const hintCard = state.hint ? state.tableau[state.hint.fromCol]?.[state.hint.cardIndex]?.card : null;
  const hintCardName = hintCard ? cardAlt(hintCard) : '';

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  return (
    <GamePageShell
      title={tc('nav.yukon')}
      gameThemeBg={gameTheme.yukon.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/yukon"
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
          <SettingsPanel
            title={tc('settings.title', { ns: 'common' })}
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
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Foundation row */}
            <div className="flex gap-1 sm:gap-2 mb-3 items-start justify-center" data-tutorial="yk-foundation">
              {state.foundation.map((pile, i) => {
                const topCard = pile.length > 0 ? pile[pile.length - 1] : null;
                const isTarget = selectedSource !== null;
                return (
                  <DropZone
                    key={i}
                    onDrop={dnd.handleDrop({ zone: 'foundation', col: i })}
                    onDragOver={dnd.handleDragOver({ zone: 'foundation', col: i })}
                    onDragLeave={dnd.handleDragLeave}
                    isDropTarget={dnd.isDropTarget({ zone: 'foundation', col: i })}
                  >
                    <button
                      type="button"
                      className={`${focusRingWhite} rounded-lg transition-colors ${
                        isTarget ? 'hover:ring-2 hover:ring-ds-warning cursor-pointer' : ''
                      }`}
                      onClick={() => isTarget && handleSelectTarget('foundation', i)}
                      disabled={!isPlaying || !isTarget}
                      aria-label={
                        topCard
                          ? t('foundationAriaLabel', {
                              suit: t(`suitNames.${FOUNDATION_SUIT_KEYS[i]}`),
                              count: pile.length,
                            })
                          : t('emptyFoundationAriaLabel', {
                              suit: t(`suitNames.${FOUNDATION_SUIT_KEYS[i]}`),
                            })
                      }
                      style={{ width: yk.cw, height: yk.ch }}
                    >
                      {topCard ? (
                        <AnimatedCard card={topCard} width={yk.cw} />
                      ) : (
                        <div
                          className="border-2 border-dashed border-game-border rounded-lg flex items-center justify-center text-game-text-muted"
                          style={{ width: yk.cw, height: yk.ch }}
                        >
                          {FOUNDATION_SUITS[i]}
                        </div>
                      )}
                    </button>
                  </DropZone>
                );
              })}
            </div>

            {/* Tableau */}
            <div className="flex gap-1 sm:gap-2 justify-center" data-tutorial="yk-tableau">
              {state.tableau.map((col, colIdx) => (
                <div key={colIdx} className="flex flex-col items-center" style={{ width: yk.cw }}>
                  <div className="text-game-text-muted text-xs mb-1">{colIdx}</div>
                  {col.length === 0 ? (
                    <DropZone
                      onDrop={dnd.handleDrop({ zone: 'tableau', col: colIdx })}
                      onDragOver={dnd.handleDragOver({ zone: 'tableau', col: colIdx })}
                      onDragLeave={dnd.handleDragLeave}
                      isDropTarget={dnd.isDropTarget({ zone: 'tableau', col: colIdx })}
                    >
                      <button
                        type="button"
                        className={`border-2 border-dashed border-game-border rounded-lg flex items-center justify-center text-game-text-muted ${focusRingWhite} ${
                          selectedSource ? 'hover:ring-2 hover:ring-ds-warning cursor-pointer' : ''
                        }`}
                        style={{ width: yk.cw, height: yk.ch }}
                        onClick={() => selectedSource && handleSelectTarget('tableau', colIdx)}
                        disabled={!isPlaying || !selectedSource}
                        aria-label={`${t('empty')} ${t('tableau')} ${colIdx}`}
                      >
                        {t('empty')}
                      </button>
                    </DropZone>
                  ) : (
                    <div className="relative" style={{ width: yk.cw, height: yk.ch + (col.length - 1) * yk.co }}>
                      {col.map((tc, cardIdx) => {
                        const isSelected = isSourceSelected('tableau', colIdx, cardIdx);
                        const zone: YukonMoveZone = { zone: 'tableau', col: colIdx, cardIndex: cardIdx };
                        const isDragSrc = dnd.isDragSource(zone);
                        const isLast = cardIdx === col.length - 1;

                        // Hint highlight (announced via the card aria-labels, no visible text panel)
                        const hintFrom =
                          state.hint && state.hint.fromCol === colIdx && state.hint.cardIndex === cardIdx;
                        const hintTo =
                          state.hint && state.hint.toZone === 'tableau' && state.hint.toCol === colIdx && isLast;
                        const hintAria = hintFrom
                          ? ` ${t('hintFromAria', { dest: hintDest })}`
                          : hintTo
                            ? ` ${t('hintToAria')}`
                            : '';

                        return (
                          <div key={cardIdx} className="absolute" style={{ top: cardIdx * yk.co, zIndex: cardIdx }}>
                            <DropZone
                              onDrop={isLast ? dnd.handleDrop(zone) : noop}
                              onDragOver={isLast ? dnd.handleDragOver(zone) : noop}
                              onDragLeave={isLast ? dnd.handleDragLeave : undefined}
                              isDropTarget={isLast && dnd.isDropTarget({ zone: 'tableau', col: colIdx })}
                            >
                              {tc.faceUp ? (
                                (() => {
                                  const inHoverBlock = hoveredBlock?.col === colIdx && cardIdx >= hoveredBlock.cardIdx;
                                  return (
                                    <button
                                      type="button"
                                      draggable={isPlaying}
                                      onDragStart={dnd.handleDragStart(zone)}
                                      onDragEnd={dnd.handleDragEnd}
                                      onMouseEnter={() => setHoveredBlock({ col: colIdx, cardIdx })}
                                      onMouseLeave={() => setHoveredBlock(null)}
                                      onFocus={() => setHoveredBlock({ col: colIdx, cardIdx })}
                                      onBlur={() => setHoveredBlock(null)}
                                      data-block-member={inHoverBlock || undefined}
                                      className={`${focusRingWhite} rounded-lg transition-all ${
                                        isSelected ? 'ring-2 ring-ds-warning -translate-y-1' : ''
                                      } ${isDragSrc ? 'opacity-50' : ''} ${
                                        hintFrom ? 'ring-2 ring-ds-info motion-safe:animate-pulse' : ''
                                      } ${hintTo ? 'ring-2 ring-ds-success motion-safe:animate-pulse' : ''} ${
                                        inHoverBlock && !isSelected ? 'ring-2 ring-ds-accent/70' : ''
                                      }`}
                                      onClick={() => {
                                        if (selectedSource) {
                                          if (isLast) {
                                            handleSelectTarget('tableau', colIdx);
                                          } else {
                                            handleSelectSource('tableau', colIdx, cardIdx);
                                          }
                                        } else {
                                          handleSelectSource('tableau', colIdx, cardIdx);
                                        }
                                      }}
                                      disabled={!isPlaying}
                                      aria-label={`${tc.card ? cardAlt(tc.card) : ''}${hintAria}`}
                                    >
                                      {tc.card && <AnimatedCard card={tc.card} width={yk.cw} />}
                                    </button>
                                  );
                                })()
                              ) : (
                                <AnimatedCardBack width={yk.cw} />
                              )}
                            </DropZone>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* Bottom controls */}
          <div data-tutorial="yk-controls">
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {error && <ErrorAlert message={error} onRetry={retry} />}

            {/* Visually hidden so the hint costs no footer space, but still announced to AT. */}
            {state.hint && (
              <div className="sr-only" role="status" aria-live="polite">
                {t('hintAnnouncement', { card: hintCardName, dest: hintDest })}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />

            <GameFooter>
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="yk-reset-button"
              />

              {isPlaying && (
                <>
                  <button type="button" className={btnOutline} onClick={handleHint} disabled={loading}>
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess}${autoCompleteReady && !loading ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={handleAutoComplete}
                    disabled={loading || !autoCompleteReady}
                    data-testid="autocomplete-button"
                    title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
                  >
                    {t('autoComplete')}
                  </button>
                  <button
                    type="button"
                    className={btnOutline}
                    onClick={handleUndo}
                    disabled={loading || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={loading}>
                    {t('giveup')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? -1}
                      onEscape={handleUndoEscape}
                      disabled={loading}
                    />
                  )}
                </>
              )}
            </GameFooter>
          </div>
        </>
      )}
    </GamePageShell>
  );
}
