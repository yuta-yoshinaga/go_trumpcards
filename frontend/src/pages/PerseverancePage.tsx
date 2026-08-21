import { useCallback, useEffect, useMemo, useRef } from 'react';
import type { PerseveranceMoveZone, perseveranceApi } from '../api/gameApi';
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
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useDestinationPreview } from '../hooks/useDestinationPreview';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { usePerseveranceGame } from '../hooks/usePerseveranceGame';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, btnWarning, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PerseveranceResponse } from '../types/card';
import { PerseverancePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PERSEVERANCE_HELP, parsePerseveranceCommand } from '../utils/cli/commands/perseveranceCommands';
import { formatPerseveranceState } from '../utils/cli/formatters/perseveranceFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { perseveranceLegalTargets, perseveranceStartsRun } from '../utils/perseveranceLegalTargets';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Perseverance tutorial step definitions. */
const BD_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bd-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bd-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bd-tableau"]',
    messageKey: 'tutorial.moves',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bd-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Perseverance solitaire game page with 13 tableau columns and 4 foundations. */
export const PerseverancePage = withTutorial(PerseverancePageContent, 'perseverance', BD_TUTORIAL_STEPS);
/** Format a frontend hint zone for display. */
function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, col: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation');
  return t('frontendHint.tableau', { col });
}

/** Inner content of the Perseverance page, wrapped by TutorialProvider. */
function PerseverancePageContent() {
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
  } = useGamePageSetup('perseverance');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    hintError,
    selectedSource,
    hint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleRedeal,
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    isAutoCompleting,
  } = usePerseveranceGame();

  // Card-move SFX: play `cardPlace` whenever the server confirms a successful

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('perseverance', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('perseverance');
  const cliConfig: CliGameConfig<PerseveranceResponse, Parameters<typeof perseveranceApi.exec>> = useMemo(
    () => ({
      gameName: 'perseverance',
      parseCommand: parsePerseveranceCommand,
      formatResponse: formatPerseveranceState,
      helpText: PERSEVERANCE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  // Responsive card dimensions for 13-column layout
  const bd = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    const cols = 13;
    const colW = Math.floor((windowWidth - padX - (cols - 1) * gapPx) / cols);
    // Floor at 32px so suits/ranks stay legible; the tableau scrolls horizontally
    // when 13 columns no longer fit (e.g. a 375px portrait phone).
    const cw = Math.min(Math.max(colW, 32), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.48);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  // Keep the selected tableau column centred in the horizontally-scrolling mobile layout.
  // Declared before any early return so hook order stays stable.
  const selectedColRef = useRef<HTMLDivElement | null>(null);
  const selectedTableauCol = selectedSource?.zone === 'tableau' ? selectedSource.col : undefined;
  useEffect(() => {
    // Only relevant on the horizontally-scrolling mobile layout; a no-op on desktop.
    if (selectedTableauCol == null || !isMobile) return;
    // Optional-call: scrollIntoView is unavailable in jsdom.
    selectedColRef.current?.scrollIntoView?.({ inline: 'center', block: 'nearest', behavior: 'smooth' });
  }, [selectedTableauCol, isMobile]);

  const isPlayingForKbd = state?.phase === PerseverancePhase.PLAYING;

  const dispatchMove = useCallback(
    (source: PerseveranceMoveZone, target: PerseveranceMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );
  const dnd = useSolitaireDragDrop<PerseveranceMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

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

  // **フックは早期 return より前。** 下に置くと state が来た描画でだけ
  // フック数が増え、React が "Rendered more hooks" で落ちる。
  const preview = useDestinationPreview<PerseveranceMoveZone>(selectedSource);

  if (!state) return <GameSkeleton gameKey="perseverance" layout={{ kind: 'tableau', topRow: 4, tableau: 13 }} />;

  const isPlaying = state.phase === PerseverancePhase.PLAYING;
  const isGameClear = state.phase === PerseverancePhase.GAME_CLEAR;
  const isGameOver = state.phase === PerseverancePhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  // **選択後は押すまで正誤が分からず、クリック→サーバーエラーのループになって
  // いた (#4795)。**13列 + 4組札で移動先候補が多い。姉妹の Wasp / Accordion は
  // 選択時に合法な移動先をリング表示している。
  //
  // **枠を出すだけで、押せなくはしない。**押せなくすると E2E の
  // 「最初の列をクリック」が別の列を掴んでしまう。
  // **選ぶ前に行き先が見える (#4454)。** hover / フォーカス中の札にも、選択後と
  // まったく同じ計算を当てる ── 判定を二重に持たないので食い違わない。
  const previewSource = preview.source;
  const previewedCard =
    previewSource?.zone === 'tableau' && previewSource.col !== undefined && previewSource.cardIndex !== undefined
      ? state.tableau[previewSource.col]?.[previewSource.cardIndex]?.card
      : undefined;
  const legalTargets = perseveranceLegalTargets(state.tableau, state.foundation, previewedCard);
  /** Ring for a legal destination: softer while it is only a hover preview. */
  const targetRing = preview.isPreview ? ' rounded ring-2 ring-ds-success/70' : ' rounded ring-2 ring-ds-success';

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  return (
    <GamePageShell
      title={tc('nav.perseverance')}
      gameThemeBg={gameTheme.perseverance.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/perseverance"
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
            {t('moveCount')}: {state.moveCount}
          </span>
          <span className="text-sm text-ds-text-muted" data-testid="redeals-left">
            {t('redealsLeft', { count: state.redealsLeft })}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Scrollable area */}
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Foundation row */}
            <div className="flex gap-1 sm:gap-2 mb-3 items-start flex-wrap" data-tutorial="bd-foundation">
              {state.foundation.map((pile, idx) => {
                const foundationZone: PerseveranceMoveZone = { zone: 'foundation', col: idx };
                return (
                  <div
                    key={`f-${idx.toString()}`}
                    className={`text-center${legalTargets.foundation.has(idx) ? targetRing : ''}`}
                    data-legal-target={legalTargets.foundation.has(idx) ? 'true' : undefined}
                    data-preview-target={legalTargets.foundation.has(idx) && preview.isPreview ? 'true' : undefined}
                  >
                    <div className="text-game-text-muted text-xs mb-1">{FOUNDATION_SUITS[idx]}</div>
                    <DropZone
                      isDropTarget={dnd.isDropTarget(foundationZone)}
                      onDragOver={dnd.handleDragOver(foundationZone)}
                      onDrop={dnd.handleDrop(foundationZone)}
                      onDragLeave={dnd.handleDragLeave}
                    >
                      {pile.length > 0 ? (
                        <button
                          type="button"
                          onClick={() => handleSelectTarget(foundationZone)}
                          disabled={!isPlaying || loading || isAutoCompleting || !selectedSource}
                          aria-label={t('foundationAriaLabel', {
                            suit: FOUNDATION_SUITS[idx],
                            count: pile.length,
                          })}
                          className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                        >
                          <AnimatedCard
                            card={pile[pile.length - 1]}
                            width={bd.cw}
                            draggable={false}
                            dealDelay={isAutoCompleting ? idx * 0.15 : 0}
                          />
                        </button>
                      ) : (
                        <button
                          type="button"
                          onClick={() => handleSelectTarget(foundationZone)}
                          disabled={!isPlaying || loading || !selectedSource}
                          aria-label={t('emptyFoundationAriaLabel', { suit: FOUNDATION_SUITS[idx] })}
                          style={{ width: bd.cw, height: bd.ch }}
                          className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                        >
                          A
                        </button>
                      )}
                    </DropZone>
                  </div>
                );
              })}
            </div>

            {/* Tableau */}
            <div className="flex gap-1 sm:gap-2 mb-3 overflow-x-auto" data-tutorial="bd-tableau">
              {state.tableau.map((col, colIdx) => {
                const tableauColZone: PerseveranceMoveZone = { zone: 'tableau', col: colIdx };
                const isColumnSelected = selectedTableauCol === colIdx;
                // A column with a single card is one move away from being emptied — and in
                // Perseverance an empty column can never be refilled. Mark it persistently
                // (before any selection) so the player can plan around it.
                const isOneCardCol = col.length === 1;
                return (
                  <div
                    key={`col-${colIdx.toString()}`}
                    ref={isColumnSelected ? selectedColRef : null}
                    // On mobile keep each column at its computed width and let the row scroll;
                    // on desktop the columns flex to fill the available width as before.
                    className={`${isMobile ? 'shrink-0' : 'flex-1 min-w-0'}${
                      isOneCardCol ? ' rounded ring-1 ring-ds-warning/40 ring-dashed' : ''
                    }${legalTargets.tableau.has(colIdx) ? targetRing : ''}`}
                    data-legal-target={legalTargets.tableau.has(colIdx) ? 'true' : undefined}
                    data-preview-target={legalTargets.tableau.has(colIdx) && preview.isPreview ? 'true' : undefined}
                    style={isMobile ? { width: bd.cw } : undefined}
                    data-testid={isOneCardCol ? `bd-onecard-col-${colIdx}` : undefined}
                  >
                    <DropZone
                      isDropTarget={dnd.isDropTarget(tableauColZone)}
                      onDragOver={dnd.handleDragOver(tableauColZone)}
                      onDrop={dnd.handleDrop(tableauColZone)}
                      onDragLeave={dnd.handleDragLeave}
                      className="relative block"
                    >
                      <div className="relative" style={{ minHeight: bd.ch }}>
                        {col.length === 0 ? (
                          <div
                            style={{ height: bd.ch }}
                            className="w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
                          >
                            {t('empty')}
                          </div>
                        ) : (
                          col.map((tc, cardIdx) => {
                            const cardZone: PerseveranceMoveZone = {
                              zone: 'tableau',
                              col: colIdx,
                              cardIndex: cardIdx,
                            };
                            const isTop = cardIdx === col.length - 1;
                            // **掴めるのは上札だけではない。**同スート降順の並びは
                            // 一括で動かせるので、その開始位置ならどれでも掴める。
                            // クローン元 (Baker's Dozen) は 1 枚ずつなので isTop だけで
                            // 足りたが、それを残すとこの game の看板ルールが UI から
                            // 一切使えなくなる。
                            const canGrab = perseveranceStartsRun(col, cardIdx);
                            const isLastInCol = isTop && col.length === 1;
                            const isSelected = isSourceSelected('tableau', colIdx, cardIdx);
                            const showEmptyColumnWarning = isLastInCol && isSelected;
                            // Selection ring wins over the dashed empty-column ring when both
                            // would apply — avoids stacking two rings that look like a bug.
                            const ringClass = isSelected
                              ? 'ring-2 ring-ds-warning'
                              : isLastInCol
                                ? 'ring-2 ring-ds-warning/60 ring-dashed'
                                : '';
                            return (
                              <div
                                key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                                className="absolute left-0 right-0"
                                style={{ top: cardIdx * bd.co }}
                              >
                                {tc.card ? (
                                  <button
                                    type="button"
                                    // **動かせる札だけがプレビューを持つ。** 掴めない札に
                                    // 出しても、その札は選べないので嘘になる。
                                    {...(canGrab ? preview.previewProps(cardZone) : {})}
                                    data-testid={isLastInCol ? `bd-last-card-${colIdx}` : undefined}
                                    onClick={() => {
                                      if (selectedSource) {
                                        handleSelectTarget(tableauColZone);
                                      } else if (canGrab) {
                                        handleSelectSource(cardZone);
                                      }
                                    }}
                                    disabled={!isPlaying || loading || (!canGrab && !selectedSource)}
                                    aria-label={cardAlt(tc.card)}
                                    aria-pressed={isSelected}
                                    draggable={isPlaying && !loading && canGrab}
                                    onDragStart={dnd.handleDragStart(cardZone)}
                                    onDragEnd={dnd.handleDragEnd}
                                    className={`p-0 border-0 bg-transparent w-full rounded ${focusRingWhite} ${canGrab ? 'cursor-pointer' : 'cursor-default'} ${ringClass} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                                    title={isLastInCol ? t('lastCardWarning') : undefined}
                                  >
                                    <AnimatedCard
                                      card={tc.card}
                                      width={bd.cw}
                                      draggable={false}
                                      style={{ width: '100%' }}
                                      wrapperClassName="block w-full"
                                    />
                                  </button>
                                ) : null}
                                {showEmptyColumnWarning && (
                                  <div
                                    data-testid={`bd-empty-column-warn-${colIdx}`}
                                    role="alert"
                                    className="absolute inset-0 flex items-center justify-center pointer-events-none text-xs text-ds-error font-bold text-center px-1 motion-safe:animate-pulse"
                                  >
                                    🚫 {t('lastCardWarning')}
                                  </div>
                                )}
                              </div>
                            );
                          })
                        )}
                        {col.length > 0 && <div style={{ height: (col.length - 1) * bd.co + bd.ch }} />}
                      </div>
                    </DropZone>
                  </div>
                );
              })}
            </div>

            {/* Hint display */}
            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-tutorial="bd-hint-display" data-testid="bd-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {formatHintZone(t, 'tableau', hint.fromCol)} →{' '}
                  {formatHintZone(t, hint.toZone, hint.toCol)}
                </div>
              )}
            </div>
            <div className="flex justify-center">
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

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
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          {/* Footer */}
          <GameFooter className={`${gameTheme.perseverance.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="bd-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleUndo}
                    disabled={loading || isAutoCompleting || !state.canUndo}
                    aria-keyshortcuts="z"
                  >
                    {t('undo')}
                    <KbdBadge label={t('kbd.undo')} />
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={handleUndoEscape}
                      disabled={loading || isAutoCompleting}
                    />
                  )}
                  {/* **手詰まりのもう一つの出口。**Undo で戻るのではなく、残りを集めて
                      配り直す。残り 0 で押せなくなる ── 押せるのに何も起きない
                      ボタンにはしない。手詰まりのときは目を引かせる。 */}
                  <button
                    type="button"
                    className={`${btnWarning}${state.isStalemate && state.redealsLeft > 0 ? ' motion-safe:animate-pulse' : ''}`}
                    onClick={handleRedeal}
                    disabled={loading || isAutoCompleting || state.redealsLeft <= 0}
                    data-testid="redeal-button"
                  >
                    {t('redeal')}
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleHint}
                    disabled={loading || isAutoCompleting}
                    aria-keyshortcuts="h"
                  >
                    {t('hint')}
                    <KbdBadge label={t('kbd.hint')} />
                  </button>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleAutoComplete}
                    disabled={loading || isAutoCompleting}
                    data-testid="autocomplete-button"
                    aria-keyshortcuts="a"
                  >
                    {t('autoComplete')}
                    <KbdBadge label={t('kbd.autoComplete')} />
                  </button>
                  <button
                    type="button"
                    className={btnDanger}
                    onClick={confirmGiveUpAction}
                    disabled={loading || isAutoCompleting}
                    aria-keyshortcuts="g"
                  >
                    {t('giveup')}
                    <KbdBadge label={t('kbd.giveUp')} />
                  </button>
                </div>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="bd-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
