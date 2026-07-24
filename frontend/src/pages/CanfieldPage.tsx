import { type DragEvent, useCallback, useMemo } from 'react';
import { type CanfieldMoveZone, canfieldApi } from '../api/gameApi';
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
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useLocalStorageToggle } from '../hooks/useLocalStorageToggle';
import { useMountReset } from '../hooks/useMountReset';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnOutline, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CanfieldResponse } from '../types/card';
import { CanfieldPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CANFIELD_HELP, parseCanfieldCommand } from '../utils/cli/commands/canfieldCommands';
import { formatCanfieldState } from '../utils/cli/formatters/canfieldFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Ring styling applied to the hinted source / target cards (mirrors Yukon / Forty Thieves). */
const HINT_RING = 'ring-2 ring-ds-info motion-safe:animate-pulse';

/**
 * Formats a Canfield hint zone into human-readable text, mirroring
 * `CanfieldCuiPresenter.HintOutput`. Tableau destinations include the 0-based
 * column number so source and target are unambiguous.
 */
function formatCanfieldHintZone(
  t: (key: string, opts?: Record<string, unknown>) => string,
  zone: string,
  col: number,
): string {
  if (zone === 'reserve') return t('reserve');
  if (zone === 'waste') return t('waste');
  if (zone === 'foundation') return t('foundation');
  return `${t('tableau')} ${col}`;
}

/** Tutorial steps for the Canfield solitaire game. */
const CF_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cf-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cf-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="cf-reserve"]', messageKey: 'tutorial.reserve', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="cf-tableau"]', messageKey: 'tutorial.tableau', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="cf-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cf-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Canfield solitaire game page. */
export const CanfieldPage = withTutorial(CanfieldPageContent, 'canfield', CF_TUTORIAL_STEPS);
/** Inner content of the Canfield page. */
function CanfieldPageContent() {
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
  } = useGamePageSetup('canfield');
  const { state, loading, error, exec: execApi, retry } = useGameApi(canfieldApi.exec);
  const { cardWidth, cardHeight, isMobile } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('canfield', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('canfield');
  // Desktop declutter: let players collapse the dense per-column action buttons
  // behind a <details> disclosure, matching the mobile treatment. Persisted so
  // the preference survives reloads. Drag-and-drop is unaffected either way.
  const [collapseColActions, setCollapseColActions] = useLocalStorageToggle('canfield-collapse-col-actions', false);
  const cliConfig: CliGameConfig<CanfieldResponse, Parameters<typeof canfieldApi.exec>> = useMemo(
    () => ({
      gameName: 'canfield',
      parseCommand: parseCanfieldCommand,
      formatResponse: formatCanfieldState,
      helpText: CANFIELD_HELP,
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
  const handleDraw = useCallback(() => execApi('draw'), [execApi]);
  const handleGiveUp = useCallback(() => execApi('giveup'), [execApi]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);
  const handleHint = useCallback(() => execApi('hint'), [execApi]);
  const handleAutoComplete = useCallback(() => execApi('autocomplete'), [execApi]);
  const handleUndo = useCallback(() => execApi('undo'), [execApi]);

  const handleMoveReserveToFoundation = useCallback(
    () => execApi('move', { zone: 'reserve' }, { zone: 'foundation' }),
    [execApi],
  );
  const handleMoveWasteToFoundation = useCallback(
    () => execApi('move', { zone: 'waste' }, { zone: 'foundation' }),
    [execApi],
  );
  const handleMoveReserveToTableau = useCallback(
    (col: number) => execApi('move', { zone: 'reserve' }, { zone: 'tableau', col }),
    [execApi],
  );
  const handleMoveWasteToTableau = useCallback(
    (col: number) => execApi('move', { zone: 'waste' }, { zone: 'tableau', col }),
    [execApi],
  );
  const handleMoveTableauToFoundation = useCallback(
    (col: number) => execApi('move', { zone: 'tableau', col }, { zone: 'foundation' }),
    [execApi],
  );
  const handleMoveTableauToTableau = useCallback(
    (fromCol: number, cardIndex: number, toCol: number) =>
      execApi('move', { zone: 'tableau', col: fromCol, cardIndex }, { zone: 'tableau', col: toCol }),
    [execApi],
  );

  const theme = useMemo(() => gameTheme.canfield, []);

  const phase = state?.phase ?? CanfieldPhase.PLAYING;
  const isPlaying = phase === CanfieldPhase.PLAYING;

  // Drag-and-drop: dispatches the same move command as button-based interaction.
  const dispatchMove = useCallback(
    (source: CanfieldMoveZone, target: CanfieldMoveZone) => {
      void execApi('move', source, target);
    },
    [execApi],
  );
  const dnd = useSolitaireDragDrop<CanfieldMoveZone>({
    onMove: dispatchMove,
    isPlaying,
    disabled: loading,
  });
  const isGameClear = phase === CanfieldPhase.GAME_CLEAR;
  const isEnded = phase === CanfieldPhase.GAME_CLEAR || phase === CanfieldPhase.GAME_OVER;

  const phaseName = isGameClear
    ? t('phase.gameClear')
    : phase === CanfieldPhase.GAME_OVER
      ? t('phase.gameOver')
      : t('phase.playing');

  if (error) return <ErrorAlert message={error} onRetry={retry} />;
  if (!state) return null;

  const topWaste = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const topReserve = state.reserve.length > 0 ? state.reserve[state.reserve.length - 1] : null;

  // Server hint (populated by the "hint" button; cleared on the next move since
  // regular responses omit the field). Resolve the hinted card + destination so
  // the move can be highlighted with a ring and announced to screen readers,
  // matching Yukon / Forty Thieves. The click handler itself needs no change.
  const hint = state.hint ?? null;
  const hintCard = hint
    ? hint.fromZone === 'reserve'
      ? topReserve
      : hint.fromZone === 'waste'
        ? topWaste
        : (state.tableau[hint.fromCol]?.[hint.cardIndex]?.card ?? null)
    : null;
  const hintCardName = hintCard ? cardAlt(hintCard) : '';
  const hintDest = hint ? formatCanfieldHintZone(t, hint.toZone, hint.toCol) : '';
  const isHintFromReserve = hint?.fromZone === 'reserve';
  const isHintFromWaste = hint?.fromZone === 'waste';
  const isHintFromTableau = (col: number, idx: number) =>
    hint != null && hint.fromZone === 'tableau' && hint.fromCol === col && hint.cardIndex === idx;
  const isHintToFoundation = (col: number) => hint?.toZone === 'foundation' && hint.toCol === col;
  const isHintToTableau = (col: number) => hint?.toZone === 'tableau' && hint.toCol === col;

  return (
    <GamePageShell
      title={tc('nav.canfield')}
      gameThemeBg={theme.bg}
      phaseName={phaseName}
      gamePath="/canfield"
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
                  {
                    type: 'checkbox' as const,
                    id: 'collapseColActions',
                    label: t('collapseColumnActions'),
                    checked: collapseColActions,
                    onToggle: setCollapseColActions,
                  },
                ],
              },
            ]}
          />

          <LandscapeBanner message={phaseName} />

          <div className="flex-1 overflow-y-auto px-4 pt-3 lg:px-8">
            {/* Foundation */}
            <div className="mb-3 flex gap-2" data-tutorial="cf-foundation">
              {state.foundation.map((pile, i) => {
                const fZone: CanfieldMoveZone = { zone: 'foundation', col: i };
                return (
                  <DropZone
                    key={`f-${i}`}
                    isDropTarget={dnd.isDropTarget(fZone)}
                    onDragOver={dnd.handleDragOver(fZone)}
                    onDrop={dnd.handleDrop(fZone)}
                    onDragLeave={dnd.handleDragLeave}
                  >
                    <div
                      className={`relative rounded border border-white/30 ${isHintToFoundation(i) ? HINT_RING : ''}`}
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

            {/* Stock / Waste / Reserve */}
            <div className="mb-3 flex gap-3" data-tutorial="cf-stock-waste">
              <div className="flex flex-col items-center">
                <button
                  type="button"
                  onClick={handleDraw}
                  disabled={!isPlaying || loading}
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

              <div className="flex flex-col items-center">
                <div style={{ width: cardWidth, height: cardHeight }}>
                  {topWaste ? (
                    <button
                      type="button"
                      draggable={isPlaying && !loading}
                      onDragStart={dnd.handleDragStart({ zone: 'waste' })}
                      onDragEnd={dnd.handleDragEnd}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isHintFromWaste ? HINT_RING : ''} ${dnd.isDragSource({ zone: 'waste' }) ? 'opacity-50' : ''}`}
                    >
                      <AnimatedCard card={topWaste} width={cardWidth} draggable={false} />
                    </button>
                  ) : (
                    <div
                      className="rounded border border-dashed border-white/30"
                      style={{ width: cardWidth, height: cardHeight }}
                    />
                  )}
                </div>
                <span className="mt-1 text-xs text-ds-text-muted">{t('waste')}</span>
              </div>

              <div className="flex flex-col items-center" data-tutorial="cf-reserve">
                <ReserveStack
                  count={state.reserve.length}
                  topCard={topReserve}
                  cardWidth={cardWidth}
                  cardHeight={cardHeight}
                  isPlaying={isPlaying}
                  loading={loading}
                  handleDragStart={dnd.handleDragStart({ zone: 'reserve' })}
                  handleDragEnd={dnd.handleDragEnd}
                  isDragSource={dnd.isDragSource({ zone: 'reserve' })}
                  isHintSource={!!isHintFromReserve}
                />
                <span className="mt-1 text-xs text-ds-text-muted">
                  {t('reserve')}: {state.reserve.length}
                </span>
              </div>
            </div>

            {/* Tableau */}
            <div className="mb-3 flex gap-2" data-tutorial="cf-tableau">
              {state.tableau.map((col, i) => {
                const tZone: CanfieldMoveZone = { zone: 'tableau', col: i };
                return (
                  <div key={`t-${i}`} className="flex flex-col gap-1">
                    <span className="text-xs text-ds-text-muted">#{i}</span>
                    <DropZone
                      isDropTarget={dnd.isDropTarget(tZone)}
                      onDragOver={dnd.handleDragOver(tZone)}
                      onDrop={dnd.handleDrop(tZone)}
                      onDragLeave={dnd.handleDragLeave}
                    >
                      <div
                        className={`relative rounded ${isHintToTableau(i) ? HINT_RING : ''}`}
                        style={{ width: cardWidth, minHeight: cardHeight }}
                      >
                        {col.length === 0 ? (
                          <div
                            className="rounded border border-dashed border-white/30"
                            style={{ width: cardWidth, height: cardHeight }}
                          />
                        ) : (
                          col.map((tc, j) => {
                            const cardZone: CanfieldMoveZone = { zone: 'tableau', col: i, cardIndex: j };
                            return (
                              <div key={`t-${i}-${j}`} className="absolute" style={{ top: j * 24, left: 0 }}>
                                <button
                                  type="button"
                                  draggable={isPlaying && !loading}
                                  onDragStart={dnd.handleDragStart(cardZone)}
                                  onDragEnd={dnd.handleDragEnd}
                                  className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isHintFromTableau(i, j) ? HINT_RING : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                                >
                                  <AnimatedCard card={tc.card} width={cardWidth} draggable={false} />
                                </button>
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
                              onClick={() => handleMoveWasteToTableau(i)}
                              disabled={!topWaste || loading}
                            >
                              {t('moveWasteToCol', { col: i })}
                            </button>
                            <button
                              type="button"
                              className={`${btnOutline} ${focusRingWhite} text-xs min-h-[44px]`}
                              onClick={() => handleMoveReserveToTableau(i)}
                              disabled={!topReserve || loading}
                            >
                              {t('moveReserveToCol', { col: i })}
                            </button>
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
                                  onClick={() => handleMoveTableauToTableau(i, col.length - 1, j)}
                                  disabled={col.length === 0 || loading}
                                >
                                  {t('moveToCol', { col: j })}
                                </button>
                              ),
                            )}
                          </div>
                        );
                        // Collapse the dense per-column action buttons behind a details
                        // disclosure when space is tight (mobile) or when the player opts in
                        // via the settings toggle on desktop — so they don't crowd below the
                        // 44px tap-target min. Drag-and-drop stays available regardless.
                        return isMobile || collapseColActions ? (
                          <details className="mt-1 w-full" data-testid={`cf-col-actions-${i}`}>
                            {/* Include the column number so a screen reader can tell the
                                per-column action panels apart. 0-based to match the rest
                                of the UI (the column header "#{i}" and the "→T0" buttons). */}
                            <summary className="text-xs text-ds-text-muted cursor-pointer min-h-[44px] flex items-center justify-center">
                              {t('columnActionsFor', { n: i })}
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
            {/* Server hint display: a visible source → target line plus a screen-reader
                announcement, mirroring CanfieldCuiPresenter.HintOutput. The hinted cards
                are also ring-highlighted above. Clears automatically once the move is
                played (the next response omits the hint field). */}
            <div data-testid="cf-hint-display">
              {hint && (
                <div className="text-sm text-ds-text-muted">
                  {t('hintAvailable')}: {formatCanfieldHintZone(t, hint.fromZone, hint.fromCol)} →{' '}
                  {formatCanfieldHintZone(t, hint.toZone, hint.toCol)}
                </div>
              )}
              <div className="sr-only" role="status" aria-live="polite" data-testid="cf-hint-announcement">
                {hint ? t('hintAnnouncement', { card: hintCardName, dest: hintDest }) : ''}
              </div>
            </div>
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
            <div className="flex flex-wrap items-center gap-2" data-tutorial="cf-action-buttons">
              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={`${btnPrimary} ${focusRingWhite}`}
                    onClick={handleMoveReserveToFoundation}
                    disabled={!topReserve || loading}
                  >
                    {t('moveReserveToFoundation')}
                  </button>
                  <button
                    type="button"
                    className={`${btnPrimary} ${focusRingWhite}`}
                    onClick={handleMoveWasteToFoundation}
                    disabled={!topWaste || loading}
                  >
                    {t('moveWasteToFoundation')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess} ${focusRingWhite}`}
                    onClick={handleHint}
                    disabled={loading}
                  >
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess} ${focusRingWhite}`}
                    onClick={handleAutoComplete}
                    disabled={loading}
                  >
                    {t('autoComplete')}
                  </button>
                  <button
                    type="button"
                    className={`${btnOutline} ${focusRingWhite}`}
                    onClick={handleUndo}
                    disabled={!state.canUndo || loading}
                  >
                    {t('undo')}
                  </button>
                  <button
                    type="button"
                    className={`${btnDanger} ${focusRingWhite}`}
                    onClick={confirmGiveUpAction}
                    disabled={loading}
                  >
                    {t('giveup')}
                  </button>
                </>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="cf-reset-button"
                className={focusRingWhite}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

/** Maximum visible stack-edge layers behind the top reserve card. */
const RESERVE_STACK_MAX_LAYERS = 5;
/** Per-layer offset in px (mirrors physical card thickness). */
const RESERVE_STACK_OFFSET_PX = 2;

interface ReserveStackProps {
  count: number;
  topCard: NonNullable<CanfieldResponse['reserve']>[number] | null;
  cardWidth: number;
  cardHeight: number;
  isPlaying: boolean;
  loading: boolean;
  handleDragStart: (e: DragEvent<HTMLElement>) => void;
  handleDragEnd: () => void;
  isDragSource: boolean;
  isHintSource: boolean;
}

/** Reserve pile with visible stack-edge layers behind the top card.
 * The number of edges reflects the remaining count (capped at
 * `RESERVE_STACK_MAX_LAYERS`) so players can sense progress at a glance. */
function ReserveStack({
  count,
  topCard,
  cardWidth,
  cardHeight,
  isPlaying,
  loading,
  handleDragStart,
  handleDragEnd,
  isDragSource,
  isHintSource,
}: ReserveStackProps) {
  const layers = Math.max(0, Math.min(count - 1, RESERVE_STACK_MAX_LAYERS));
  // Container is sized for the maximum stack so the top card sits at a
  // fixed offset regardless of `count` — prevents items-center from
  // jittering the pile as cards are consumed.
  const maxPad = RESERVE_STACK_MAX_LAYERS * RESERVE_STACK_OFFSET_PX;
  return (
    <div
      data-testid="canfield-reserve-stack"
      data-reserve-layers={layers}
      className="relative"
      style={{ width: cardWidth + maxPad, height: cardHeight + maxPad }}
    >
      {Array.from({ length: layers }, (_, i) => (
        <div
          key={`layer-${i}`}
          aria-hidden="true"
          className="absolute rounded bg-ds-info/70 border border-white/20"
          style={{
            top: maxPad - (i + 1) * RESERVE_STACK_OFFSET_PX,
            left: maxPad - (i + 1) * RESERVE_STACK_OFFSET_PX,
            width: cardWidth,
            height: cardHeight,
          }}
        />
      ))}
      <div className="absolute" style={{ top: maxPad, left: maxPad, width: cardWidth, height: cardHeight }}>
        {topCard ? (
          <button
            type="button"
            draggable={isPlaying && !loading}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isHintSource ? HINT_RING : ''} ${isDragSource ? 'opacity-50' : ''}`}
          >
            <AnimatedCard card={topCard} width={cardWidth} draggable={false} />
          </button>
        ) : (
          <div
            className="rounded border border-dashed border-white/30"
            style={{ width: cardWidth, height: cardHeight }}
          />
        )}
      </div>
    </div>
  );
}
