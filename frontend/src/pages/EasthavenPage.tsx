import { useCallback, useMemo, useState } from 'react';
import { type EasthavenMoveZone, easthavenApi } from '../api/gameApi';
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
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
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
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useMountReset } from '../hooks/useMountReset';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import i18n from '../i18n';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnOutline, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, EasthavenResponse } from '../types/card';
import { EasthavenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { easthavenHelp, parseEasthavenCommand } from '../utils/cli/commands/easthavenCommands';
import { formatEasthavenState } from '../utils/cli/formatters/easthavenFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { easthavenFoundationTarget } from '../utils/easthavenFoundationTarget';
import { hintCheckboxItem } from '../utils/settingsItems';
import { isTableauAllFaceUp } from '../utils/solitaireUtils';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;
const noop = () => {};

/** Easthaven tutorial step definitions. */
const EH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="eh-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eh-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eh-stock"]',
    messageKey: 'tutorial.stock',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eh-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eh-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Easthaven game page. */
export const EasthavenPage = withTutorial(EasthavenPageContent, 'easthaven', EH_TUTORIAL_STEPS);
/** Inner content of the Easthaven page. */
function EasthavenPageContent() {
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
  } = useGamePageSetup('easthaven');
  const { playSound } = useSound();
  const {
    state,
    setState,
    loading,
    error,
    exec: apiExec,
    retry,
  } = useGameApi<EasthavenResponse, Parameters<typeof easthavenApi.exec>>((...args) => easthavenApi.exec(...args));

  useMountReset(apiExec);

  const [selectedSource, setSelectedSource] = useState<EasthavenMoveZone | null>(null);
  /**
   * Tableau coordinate the mouse is over. Easthaven lets a player grab a
   * face-up ordered run, so on hover we glow every card from `cardIdx` to the
   * column tail to make the moving block visually unambiguous.
   */
  const [hoveredBlock, setHoveredBlock] = useState<{ col: number; cardIdx: number } | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('easthaven', state);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('easthaven');
  // easthavenHelp() reads i18n internally, so depend on i18n.language to
  // re-localize the CLI help after a runtime language switch.
  // biome-ignore lint/correctness/useExhaustiveDependencies: i18n.language drives help re-localization
  const easthavenCliConfig: CliGameConfig<EasthavenResponse, Parameters<typeof easthavenApi.exec>> = useMemo(
    () => ({
      gameName: 'easthaven',
      parseCommand: parseEasthavenCommand,
      formatResponse: formatEasthavenState,
      helpText: easthavenHelp(),
    }),
    [i18n.language],
  );
  const { handleCommand } = useCliGame(apiExec, easthavenCliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });

  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const eh = useMemo(() => {
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
    (source: EasthavenMoveZone, target: EasthavenMoveZone) => {
      void apiExec('move', source, target);
    },
    [apiExec],
  );
  const dnd = useSolitaireDragDrop<EasthavenMoveZone>({
    onMove: dispatchMove,
    isPlaying: state?.phase === EasthavenPhase.PLAYING,
    disabled: loading,
  });

  // Action handlers
  const handleManualReset = useCallback(() => {
    void apiExec('reset');
    playSound('shuffle');
  }, [apiExec, playSound]);

  const handleDeal = useCallback(() => {
    void apiExec('deal');
    playSound('cardPlace');
  }, [apiExec, playSound]);

  // Empty-column deal guard: surfaces a shake animation + tooltip instead of failing silently.
  const [emptyDealAttemptKey, setEmptyDealAttemptKey] = useState(0);
  const hasEmptyColumn = useMemo(() => state?.tableau.some((col) => col.length === 0) ?? false, [state?.tableau]);
  const dealBlockedByEmpty = hasEmptyColumn && (state?.stockCount ?? 0) > 0;
  const handleDealGuarded = useCallback(() => {
    if (dealBlockedByEmpty) {
      setEmptyDealAttemptKey((k) => k + 1);
      return;
    }
    setEmptyDealAttemptKey(0);
    handleDeal();
  }, [dealBlockedByEmpty, handleDeal]);

  const handleGiveUp = useCallback(() => {
    void apiExec('giveup');
  }, [apiExec]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const handleHint = useCallback(async () => {
    const res = await easthavenApi.exec('hint');
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

  // Double-click / double-tap shortcut: auto-send a column's exposed top card
  // to its foundation when a legal target exists; otherwise do nothing (no
  // error, selection untouched). Mirrors the FreeCell foundation shortcut.
  const handleFoundationShortcut = useCallback(
    (col: number, card: Card) => {
      if (!state) return;
      const target = easthavenFoundationTarget(card, state.foundation);
      if (!target) return;
      void apiExec('move', { zone: 'tableau', col }, target);
      setSelectedSource(null);
      playSound('cardPlace');
    },
    [state, apiExec, playSound],
  );

  const isPlayingForKbd = state?.phase === EasthavenPhase.PLAYING;

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: confirmGiveUpAction },
      { key: 'z', action: handleUndo },
      { key: 'd', action: handleDealGuarded },
    ],
    [handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo, handleDealGuarded],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state) return <GameSkeleton gameKey="easthaven" layout={{ kind: 'tableau', topRow: 6, tableau: 7 }} />;

  const isPlaying = state.phase === EasthavenPhase.PLAYING;
  const isGameClear = state.phase === EasthavenPhase.GAME_CLEAR;
  const isGameOver = state.phase === EasthavenPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const autoCompleteReady = isTableauAllFaceUp(state.tableau) && state.stockCount === 0;

  // Compose the hint into a full "move <card> in column N → <dest>" sentence so
  // the aria-live announcement and visible box read completely, even when the
  // ring-flash source card is scrolled out of view (issue #3388).
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
      title={tc('nav.easthaven')}
      gameThemeBg={gameTheme.easthaven.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/easthaven"
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
          <span
            data-tutorial="eh-stock"
            data-testid="eh-stock"
            className={state.stockCount === 0 ? 'font-bold text-ds-warning' : undefined}
          >
            {t('stock')}: {state.stockCount}
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
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Foundation row */}
            <div className="flex gap-1 sm:gap-2 mb-3 items-start justify-center" data-tutorial="eh-foundation">
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
                      className={
                        isTarget
                          ? `${focusRingWhite} rounded-lg transition-colors hover:ring-2 hover:ring-ds-warning cursor-pointer`
                          : `${focusRingWhite} rounded-lg transition-colors`
                      }
                      onClick={() => isTarget && handleSelectTarget('foundation', i)}
                      disabled={!isPlaying || !isTarget}
                      aria-label={
                        topCard
                          ? t('foundationAriaLabel', { suit: FOUNDATION_SUITS[i], count: pile.length })
                          : t('emptyFoundationAriaLabel', { suit: FOUNDATION_SUITS[i] })
                      }
                      style={{ width: eh.cw, height: eh.ch }}
                    >
                      {topCard ? (
                        <AnimatedCard card={topCard} width={eh.cw} />
                      ) : (
                        <div
                          className="border-2 border-dashed border-game-border rounded-lg flex items-center justify-center text-game-text-muted"
                          style={{ width: eh.cw, height: eh.ch }}
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
            <div className="flex gap-1 sm:gap-2 justify-center" data-tutorial="eh-tableau">
              {state.tableau.map((col, colIdx) => (
                <div key={colIdx} className="flex flex-col items-center" style={{ width: eh.cw }}>
                  {/* Once the stock is empty, flag columns still hiding a face-down card. */}
                  <div
                    data-testid={`eh-col-header-${colIdx}`}
                    className={`mb-1 rounded px-1 text-game-text-muted text-xs ${
                      state.stockCount === 0 && !autoCompleteReady && col.some((c) => !c.faceUp)
                        ? 'bg-ds-warning/20'
                        : ''
                    }`}
                  >
                    {colIdx}
                  </div>
                  {col.length === 0 ? (
                    <DropZone
                      onDrop={dnd.handleDrop({ zone: 'tableau', col: colIdx })}
                      onDragOver={dnd.handleDragOver({ zone: 'tableau', col: colIdx })}
                      onDragLeave={dnd.handleDragLeave}
                      isDropTarget={dnd.isDropTarget({ zone: 'tableau', col: colIdx })}
                    >
                      <button
                        key={`empty-${colIdx.toString()}-${emptyDealAttemptKey.toString()}`}
                        type="button"
                        className={`border-2 border-dashed border-game-border rounded-lg flex items-center justify-center text-game-text-muted ${focusRingWhite} ${
                          selectedSource ? 'hover:ring-2 hover:ring-ds-warning cursor-pointer' : ''
                        }${emptyDealAttemptKey > 0 ? ' animate-shake border-ds-warning text-ds-warning' : ''}`}
                        style={{ width: eh.cw, height: eh.ch }}
                        onClick={() => selectedSource && handleSelectTarget('tableau', colIdx)}
                        disabled={!isPlaying || !selectedSource}
                        aria-label={
                          selectedSource
                            ? t('moveHereAriaLabel', { col: colIdx })
                            : `${t('empty')} ${t('tableau')} ${colIdx}`
                        }
                        data-testid={`eh-empty-col-${colIdx.toString()}`}
                      >
                        {t('empty')}
                      </button>
                    </DropZone>
                  ) : (
                    <div className="relative" style={{ width: eh.cw, height: eh.ch + (col.length - 1) * eh.co }}>
                      {col.map((tcard, cardIdx) => {
                        const isSelected = isSourceSelected('tableau', colIdx, cardIdx);
                        const zone: EasthavenMoveZone = { zone: 'tableau', col: colIdx, cardIndex: cardIdx };
                        const isDragSrc = dnd.isDragSource(zone);
                        const isLast = cardIdx === col.length - 1;

                        const hintFrom =
                          state.hint && state.hint.fromCol === colIdx && state.hint.cardIndex === cardIdx;
                        const hintTo =
                          state.hint && state.hint.toZone === 'tableau' && state.hint.toCol === colIdx && isLast;

                        return (
                          <div key={cardIdx} className="absolute" style={{ top: cardIdx * eh.co, zIndex: cardIdx }}>
                            <DropZone
                              onDrop={isLast ? dnd.handleDrop(zone) : noop}
                              onDragOver={isLast ? dnd.handleDragOver(zone) : noop}
                              onDragLeave={isLast ? dnd.handleDragLeave : undefined}
                              isDropTarget={isLast && dnd.isDropTarget({ zone: 'tableau', col: colIdx })}
                            >
                              {tcard.faceUp ? (
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
                                      onClick={(e) => {
                                        // The second click of a double-click also
                                        // fires onClick (detail === 2); ignore it so
                                        // onDoubleClick owns the foundation shortcut
                                        // without issuing a stray select/target move.
                                        if (e.detail >= 2) return;
                                        if (selectedSource) {
                                          if (isSelected) {
                                            setSelectedSource(null);
                                          } else if (isLast) {
                                            handleSelectTarget('tableau', colIdx);
                                          } else {
                                            handleSelectSource('tableau', colIdx, cardIdx);
                                          }
                                        } else {
                                          handleSelectSource('tableau', colIdx, cardIdx);
                                        }
                                      }}
                                      onDoubleClick={
                                        // Only a column's exposed top card can move
                                        // straight to a foundation.
                                        isLast && tcard.card
                                          ? () => handleFoundationShortcut(colIdx, tcard.card as Card)
                                          : undefined
                                      }
                                      disabled={!isPlaying}
                                      aria-label={tcard.card ? cardAlt(tcard.card) : ''}
                                      data-testid={isLast ? `eh-tableau-top-${colIdx.toString()}` : undefined}
                                    >
                                      {tcard.card && <AnimatedCard card={tcard.card} width={eh.cw} />}
                                    </button>
                                  );
                                })()
                              ) : (
                                <AnimatedCardBack width={eh.cw} />
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
          <div data-tutorial="eh-controls">
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {error && <ErrorAlert message={error} onRetry={retry} />}

            {state.hint && (
              <div
                className="text-sm text-ds-accent bg-ds-surface/90 border border-ds-accent rounded px-3 py-1.5 mt-1"
                role="status"
                aria-live="polite"
                data-testid="eh-hint"
              >
                {t('hintSentence', { fromCol: state.hint.fromCol, card: hintCardName, dest: hintDest })}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

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
                dataTutorial="eh-reset-button"
              />

              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleDealGuarded}
                    disabled={loading || state.stockCount === 0}
                    title={dealBlockedByEmpty ? t('cannotDealEmptyColExists') : undefined}
                  >
                    {t('deal')}
                  </button>
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
