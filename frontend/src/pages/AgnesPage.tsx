import { useCallback, useMemo } from 'react';
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
import { HintTooltip } from '../components/hint/HintTooltip';
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
import { useMountReset } from '../hooks/useMountReset';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnOutline, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { AgnesResponse } from '../types/card';
import { AgnesPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { AGNES_HELP, parseAgnesCommand } from '../utils/cli/commands/agnesCommands';
import { formatAgnesState } from '../utils/cli/formatters/agnesFormatter';
import type { CliGameConfig } from '../utils/cli/types';

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
  const { state, loading, error, exec: execApi, retry } = useGameApi(agnesApi.exec);
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
  const confirmGiveUpAction = useCallback(
    () => requestGiveUpConfirm(handleGiveUp),
    [requestGiveUpConfirm, handleGiveUp],
  );
  const handleHint = useCallback(() => execApi('hint'), [execApi]);
  const handleUndo = useCallback(() => execApi('undo'), [execApi]);

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

  // Keyboard shortcuts for the primary actions, matching other solitaire pages.
  // Give-up (g) is routed through its confirm dialog since it is irreversible.
  const canPlayForKbd = isPlaying && !loading;
  const agnesBindings = useMemo(
    () => [
      { key: 'd', action: handleDeal, enabled: canPlayForKbd && (state?.stockCount ?? 0) > 0 },
      { key: 'h', action: handleHint, enabled: canPlayForKbd },
      { key: 'z', action: handleUndo, enabled: canPlayForKbd && (state?.canUndo ?? false) },
      { key: 'g', action: confirmGiveUpAction, enabled: canPlayForKbd },
    ],
    [handleDeal, handleHint, handleUndo, confirmGiveUpAction, canPlayForKbd, state?.stockCount, state?.canUndo],
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

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

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
                    disabled={loading || state.stockCount === 0}
                    aria-keyshortcuts="d"
                  >
                    {t('deal')}
                    <KbdBadge label={t('kbd.deal')} />
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess} ${focusRingWhite}`}
                    onClick={handleHint}
                    disabled={loading}
                    aria-keyshortcuts="h"
                  >
                    {t('hint')}
                    <KbdBadge label={t('kbd.hint')} />
                  </button>
                  <button
                    type="button"
                    className={`${btnOutline} ${focusRingWhite}`}
                    onClick={handleUndo}
                    disabled={!state.canUndo || loading}
                    aria-keyshortcuts="z"
                  >
                    {t('undo')}
                    <KbdBadge label={t('kbd.undo')} />
                  </button>
                  <button
                    type="button"
                    className={`${btnDanger} ${focusRingWhite}`}
                    onClick={confirmGiveUpAction}
                    disabled={loading}
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
