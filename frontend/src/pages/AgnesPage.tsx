import { useCallback, useMemo, useRef, useState } from 'react';
import { type AgnesMoveZone, agnesApi } from '../api/gameApi';
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
import { KbdBadge } from '../components/KbdBadge';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useMountReset } from '../hooks/useMountReset';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnOutline, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { AgnesResponse } from '../types/card';
import { AgnesPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { agnesHasLegalMove, agnesNextFoundationMove } from '../utils/agnesMoves';
import { AGNES_HELP, parseAgnesCommand } from '../utils/cli/commands/agnesCommands';
import { formatAgnesState } from '../utils/cli/formatters/agnesFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import { isTableauAllFaceUp } from '../utils/solitaireUtils';

/** Upper bound on auto-complete sweep iterations (one per card) to guard against loops. */
const AGNES_AUTO_MAX_STEPS = 52;

/** Tutorial steps for the Agnes Sorel solitaire game. */
const AG_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ag-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ag-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ag-tableau"]', messageKey: 'tutorial.tableau', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="ag-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ag-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Agnes Sorel solitaire game page. */
export const AgnesPage = withTutorial(AgnesPageContent, 'agnes', AG_TUTORIAL_STEPS);
/** Inner content of the Agnes Sorel page. */
function AgnesPageContent() {
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
  } = useGamePageSetup('agnes');
  const { state, setState, loading, error, exec: execApi, retry } = useGameApi(agnesApi.exec);
  const { cardWidth, cardHeight, isMobile } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('agnes', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('agnes');
  const cliConfig: CliGameConfig<AgnesResponse, Parameters<typeof agnesApi.exec>> = useMemo(
    () => ({
      gameName: 'agnes',
      parseCommand: parseAgnesCommand,
      formatResponse: formatAgnesState,
      helpText: AGNES_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });

  useMountReset(execApi);

  const handleReset = useCallback(() => execApi('reset'), [execApi]);
  const handleDeal = useCallback(() => execApi('deal'), [execApi]);
  const handleGiveUp = useCallback(() => execApi('giveup'), [execApi]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);
  const handleHint = useCallback(() => execApi('hint'), [execApi]);
  const handleUndo = useCallback(() => execApi('undo'), [execApi]);

  // Auto-complete sweep: drives the existing `move` command to send every
  // face-up tableau end card to its foundation, re-reading the fresh board after
  // each move (a move can expose a new foundation-eligible card). Mirrors the
  // Aces Up / Spiderette sequential-driver pattern (#4193). The ref guard blocks
  // a second concurrent loop and `isAutoCompleting` gates the page controls.
  const [isAutoCompleting, setIsAutoCompleting] = useState(false);
  const autoCompletingRef = useRef(false);
  const stateRef = useRef(state);
  stateRef.current = state;
  const handleAutoComplete = useCallback(async () => {
    if (autoCompletingRef.current) return;
    autoCompletingRef.current = true;
    setIsAutoCompleting(true);
    let col = -1;
    try {
      let cur = stateRef.current;
      for (let step = 0; step < AGNES_AUTO_MAX_STEPS && cur; step++) {
        col = agnesNextFoundationMove(cur.tableau, cur.foundation, cur.baseRank);
        if (col < 0) break;
        const res = await agnesApi.exec('move', { zone: 'tableau', col }, { zone: 'foundation' });
        setState(res);
        cur = res;
      }
    } catch {
      // Re-issue the failing move through the shared exec so the network failure
      // surfaces via the standard error/retry channel.
      if (col >= 0) await execApi('move', { zone: 'tableau', col }, { zone: 'foundation' });
    } finally {
      autoCompletingRef.current = false;
      setIsAutoCompleting(false);
    }
  }, [execApi, setState]);

  const handleMoveTableauToFoundation = useCallback(
    (col: number) => execApi('move', { zone: 'tableau', col }, { zone: 'foundation' }),
    [execApi],
  );
  const handleMoveTableauToTableau = useCallback(
    (fromCol: number, cardIndex: number, toCol: number) =>
      execApi('move', { zone: 'tableau', col: fromCol, cardIndex }, { zone: 'tableau', col: toCol }),
    [execApi],
  );

  const theme = useMemo(() => gameTheme.agnes, []);

  const phase = state?.phase ?? AgnesPhase.PLAYING;
  const isPlaying = phase === AgnesPhase.PLAYING;

  // Drag-and-drop: dispatches the same move command as button-based interaction.
  const dispatchMove = useCallback(
    (source: AgnesMoveZone, target: AgnesMoveZone) => {
      void execApi('move', source, target);
    },
    [execApi],
  );
  const dnd = useSolitaireDragDrop<AgnesMoveZone>({
    onMove: dispatchMove,
    isPlaying,
    disabled: loading,
  });
  const isGameClear = phase === AgnesPhase.GAME_CLEAR;
  const isEnded = phase === AgnesPhase.GAME_CLEAR || phase === AgnesPhase.GAME_OVER;

  const phaseName = isGameClear
    ? t('phase.gameClear')
    : phase === AgnesPhase.GAME_OVER
      ? t('phase.gameOver')
      : t('phase.playing');

  // Auto-complete is offered once the stock is exhausted and every tableau card
  // is face-up — the endgame sweep, mirroring Spiderette's `autoCompleteReady`.
  const autoCompleteReady = isPlaying && (state?.stockCount ?? 1) === 0 && isTableauAllFaceUp(state?.tableau ?? []);
  // Stalemate: no deal, no foundation move, and no tableau move remain. Detected
  // on the frontend from the deterministic move rules (the backend exposes no
  // stalemate flag), so the UI can prompt an undo / give-up escape.
  const hasLegalMove = state
    ? agnesHasLegalMove(state.tableau, state.foundation, state.baseRank, state.stockCount)
    : true;
  const isStalemate = isPlaying && !loading && !isAutoCompleting && !hasLegalMove;

  // Keyboard shortcuts for the primary actions, matching other solitaire pages.
  // Give-up (g) is routed through its confirm dialog since it is irreversible.
  const canPlayForKbd = isPlaying && !loading;
  const agnesBindings = useMemo(
    () => [
      { key: 'd', action: handleDeal, enabled: canPlayForKbd && (state?.stockCount ?? 0) > 0 },
      { key: 'h', action: handleHint, enabled: canPlayForKbd },
      { key: 'a', action: handleAutoComplete, enabled: canPlayForKbd && autoCompleteReady && !isAutoCompleting },
      { key: 'z', action: handleUndo, enabled: canPlayForKbd && (state?.canUndo ?? false) },
      { key: 'g', action: confirmGiveUpAction, enabled: canPlayForKbd },
    ],
    [
      handleDeal,
      handleHint,
      handleAutoComplete,
      handleUndo,
      confirmGiveUpAction,
      canPlayForKbd,
      autoCompleteReady,
      isAutoCompleting,
      state?.stockCount,
      state?.canUndo,
    ],
  );
  useActionKeyboardNav({ bindings: agnesBindings, enabled: canPlayForKbd });

  if (!state) return null;

  return (
    <GamePageShell
      title={tc('nav.agnes')}
      gameThemeBg={theme.bg}
      phaseName={phaseName}
      gamePath="/agnes"
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
          <span className="text-sm text-ds-text-muted">
            {t('baseRank')}: {state.baseRank || '?'}
          </span>
          <span className="text-sm text-ds-text-muted">
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
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          <LandscapeBanner message={phaseName} />

          <div className="flex-1 overflow-y-auto px-4 pt-3 lg:px-8">
            {/* Foundation */}
            <div className="mb-3 flex gap-2" data-tutorial="ag-foundation">
              {state.foundation.map((pile, i) => {
                const fZone: AgnesMoveZone = { zone: 'foundation', col: i };
                return (
                  <DropZone
                    key={`f-${i}`}
                    isDropTarget={dnd.isDropTarget(fZone)}
                    onDragOver={dnd.handleDragOver(fZone)}
                    onDrop={dnd.handleDrop(fZone)}
                    onDragLeave={dnd.handleDragLeave}
                  >
                    <div
                      className="relative rounded border border-white/30"
                      style={{ width: cardWidth, height: cardHeight }}
                    >
                      {pile.length > 0 ? (
                        <AnimatedCard card={pile[pile.length - 1]} width={cardWidth} />
                      ) : (
                        <span className="absolute inset-0 flex items-center justify-center text-xs text-ds-text-muted/80">
                          {t('foundation')}
                        </span>
                      )}
                    </div>
                  </DropZone>
                );
              })}
            </div>

            {/* Stock (deal) */}
            <div className="mb-3 flex gap-3" data-tutorial="ag-stock">
              <div className="flex flex-col items-center">
                <button
                  type="button"
                  onClick={handleDeal}
                  disabled={!isPlaying || loading || state.stockCount === 0}
                  className="rounded border border-white/30"
                  aria-label={t('stock')}
                  style={{ width: cardWidth, height: cardHeight }}
                >
                  {state.stockCount > 0 ? (
                    <AnimatedCardBack width={cardWidth} />
                  ) : (
                    <span className="text-xs text-ds-text-muted/80">{t('empty')}</span>
                  )}
                </button>
                <span className="mt-1 text-xs text-ds-text-muted">
                  {t('stock')}: {state.stockCount}
                </span>
              </div>
            </div>

            {/* Tableau */}
            <div className="mb-3 flex gap-2" data-tutorial="ag-tableau">
              {state.tableau.map((col, i) => {
                const tZone: AgnesMoveZone = { zone: 'tableau', col: i };
                const endIndex = col.length - 1;
                return (
                  <div key={`t-${i}`} className="flex flex-col gap-1">
                    <span className="text-xs text-ds-text-muted">#{i}</span>
                    <DropZone
                      isDropTarget={dnd.isDropTarget(tZone)}
                      onDragOver={dnd.handleDragOver(tZone)}
                      onDrop={dnd.handleDrop(tZone)}
                      onDragLeave={dnd.handleDragLeave}
                    >
                      <div className="relative" style={{ width: cardWidth, minHeight: cardHeight }}>
                        {col.length === 0 ? (
                          <div
                            className="rounded border border-dashed border-white/30"
                            style={{ width: cardWidth, height: cardHeight }}
                          />
                        ) : (
                          col.map((tcard, j) => {
                            const isEnd = j === endIndex;
                            // Only the face-up end card is a draggable move source.
                            const draggable = isEnd && tcard.faceUp && isPlaying && !loading;
                            const cardZone: AgnesMoveZone = { zone: 'tableau', col: i, cardIndex: j };
                            return (
                              <div key={`t-${i}-${j}`} className="absolute" style={{ top: j * 24, left: 0 }}>
                                {tcard.faceUp && tcard.card ? (
                                  <button
                                    type="button"
                                    draggable={draggable}
                                    onDragStart={draggable ? dnd.handleDragStart(cardZone) : undefined}
                                    onDragEnd={dnd.handleDragEnd}
                                    className={`p-0 border-0 bg-transparent ${draggable ? 'cursor-pointer' : 'cursor-default'} rounded ${focusRingWhite} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                                  >
                                    <AnimatedCard card={tcard.card} width={cardWidth} draggable={false} />
                                  </button>
                                ) : (
                                  <AnimatedCardBack width={cardWidth} />
                                )}
                              </div>
                            );
                          })
                        )}
                      </div>
                    </DropZone>
                    {isPlaying &&
                      (() => {
                        const actionButtons = (
                          <div className="flex flex-col gap-1">
                            <button
                              type="button"
                              className={`${btnOutline} ${focusRingWhite} text-xs min-h-[44px]`}
                              onClick={() => handleMoveTableauToFoundation(i)}
                              disabled={col.length === 0 || loading}
                            >
                              {t('moveToFoundation')}
                            </button>
                            {state.tableau.map((_, j) =>
                              j === i ? null : (
                                <button
                                  key={`t-${i}-to-${j}`}
                                  type="button"
                                  className={`${btnOutline} ${focusRingWhite} text-xs min-h-[44px]`}
                                  onClick={() => handleMoveTableauToTableau(i, endIndex, j)}
                                  disabled={col.length === 0 || loading}
                                >
                                  {t('moveToCol', { col: j })}
                                </button>
                              ),
                            )}
                          </div>
                        );
                        // On mobile, collapse the dense per-column action buttons behind a
                        // details disclosure so they don't crowd below the 44px tap-target min.
                        return isMobile ? (
                          <details className="mt-1 w-full" data-testid={`ag-col-actions-${i}`}>
                            <summary className="text-xs text-ds-text-muted cursor-pointer min-h-[44px] flex items-center justify-center">
                              {t('columnActions')}
                            </summary>
                            {actionButtons}
                          </details>
                        ) : (
                          actionButtons
                        );
                      })()}
                  </div>
                );
              })}
            </div>

            {isStalemate && (
              <div
                className="mt-1 flex flex-wrap items-center gap-2 text-ds-danger text-sm font-medium"
                role="status"
                data-testid="ag-stalemate-banner"
              >
                <span>{t('stalemate')}</span>
                {state.canUndo && (
                  <button
                    type="button"
                    className={`${btnOutline} ${focusRingWhite} text-xs motion-safe:animate-pulse`}
                    onClick={handleUndo}
                    disabled={loading}
                    data-testid="ag-stalemate-undo"
                  >
                    {t('stalemateUndo')}
                  </button>
                )}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${theme.footer} px-4 py-2.5`}>
            {/* Show transient API errors inline so the board stays visible (issue #3290). */}
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex flex-wrap items-center gap-2" data-tutorial="ag-action-buttons">
              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={`${btnPrimary} ${focusRingWhite}`}
                    onClick={handleDeal}
                    disabled={loading || isAutoCompleting || state.stockCount === 0}
                    aria-keyshortcuts="d"
                  >
                    {t('deal')}
                    <KbdBadge label={t('kbd.deal')} />
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess} ${focusRingWhite}`}
                    onClick={handleHint}
                    disabled={loading || isAutoCompleting}
                    aria-keyshortcuts="h"
                  >
                    {t('hint')}
                    <KbdBadge label={t('kbd.hint')} />
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess} ${focusRingWhite}${autoCompleteReady && !loading && !isAutoCompleting ? ' motion-safe:animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={handleAutoComplete}
                    disabled={loading || isAutoCompleting || !autoCompleteReady}
                    aria-keyshortcuts="a"
                    data-testid="ag-autocomplete-button"
                    title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
                  >
                    {t('autoComplete')}
                    <KbdBadge label={t('kbd.autoComplete')} />
                  </button>
                  <button
                    type="button"
                    className={`${btnOutline} ${focusRingWhite}`}
                    onClick={handleUndo}
                    disabled={!state.canUndo || loading || isAutoCompleting}
                    aria-keyshortcuts="z"
                  >
                    {t('undo')}
                    <KbdBadge label={t('kbd.undo')} />
                  </button>
                  <button
                    type="button"
                    className={`${btnDanger} ${focusRingWhite}`}
                    onClick={confirmGiveUpAction}
                    disabled={loading || isAutoCompleting}
                    aria-keyshortcuts="g"
                  >
                    {t('giveup')}
                    <KbdBadge label={t('kbd.giveup')} />
                  </button>
                </>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="ag-reset-button"
                className={focusRingWhite}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
