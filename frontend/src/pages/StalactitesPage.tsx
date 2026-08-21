import { useCallback, useMemo, useState } from 'react';
import type { StalactitesMoveZone, stalactitesApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { AutoCompleteReadyBadge } from '../components/AutoCompleteReadyBadge';
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
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGiveUpConfirm } from '../hooks/useGiveUpConfirm';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useStalactitesGame } from '../hooks/useStalactitesGame';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, StalactitesResponse } from '../types/card';
import { StalactitesPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseStalactitesCommand, STALACTITES_HELP } from '../utils/cli/commands/stalactitesCommands';
import { formatStalactitesState } from '../utils/cli/formatters/stalactitesFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import { stalactitesAutoCompleteReady } from '../utils/stalactitesAutoComplete';
import { stalactitesFoundationTarget } from '../utils/stalactitesFoundationTarget';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** Stalactites tutorial step definitions. */
const FC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="fc-free-cells"]',
    messageKey: 'tutorial.cells',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Stalactites solitaire game page with tableau, free cells, and foundation. */
export const StalactitesPage = withTutorial(StalactitesPageContent, 'stalactites', FC_TUTORIAL_STEPS);
/** Inner content of the Stalactites page, wrapped by TutorialProvider. */
/**
 * Render one side of a hint as "<zone> <col>", or just the zone when the hint
 * carries no column.
 *
 * **ゾーン識別子を i18n キーに使わない。** ドメインが返す "tableau" などをそのまま
 * `t()` に渡していたので、ゾーン名を変えたり足したりすると翻訳キーが静かに壊れ、
 * コンパイルでも型でも検出できなかった (#5494)。姉妹の Solitaire 系ページ
 * (FlowerGarden ほか) と同じ frontendHint.* 名前空間に寄せる。
 *
 * 未知のゾーンはキーではなく識別子そのものを返す -- 生の "tableau" が出るほうが、
 * 翻訳キー文字列が画面に出るより読める。
 */
function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, col: number): string {
  const label =
    zone === 'tableau' || zone === 'stalactites' || zone === 'foundation' ? t(`frontendHint.zone.${zone}`) : zone;
  return col >= 0 ? `${label} ${col}` : label;
}

function StalactitesPageContent() {
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
  } = useGamePageSetup('stalactites');
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
    handleUndo,
    handleUndoEscape,
    handleSelectSource,
    handleSelectTarget,
    handleAutoFoundation,
    isAutoCompleting,
  } = useStalactitesGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('stalactites', state);
  const { cardHeight, cardOverlap, cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('stalactites');
  const cliConfig: CliGameConfig<StalactitesResponse, Parameters<typeof stalactitesApi.exec>> = useMemo(
    () => ({
      gameName: 'stalactites',
      parseCommand: parseStalactitesCommand,
      formatResponse: formatStalactitesState,
      helpText: STALACTITES_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayingForKbd = state?.phase === StalactitesPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: StalactitesMoveZone, target: StalactitesMoveZone) => {
      void exec('move', source, target);
    },
    [exec],
  );

  // Double-click / double-tap shortcut: auto-send an exposed card (a tableau
  // top card or a free-cell card) to its foundation when a legal target
  // exists; otherwise do nothing (no error, selection untouched).
  const handleFoundationShortcut = useCallback(
    (source: StalactitesMoveZone, card: Card) => {
      if (!state) return;
      const target = stalactitesFoundationTarget(card, state.foundation, state.baseRank);
      if (!target) return;
      handleAutoFoundation(source, target);
    },
    [state, handleAutoFoundation],
  );
  const dnd = useSolitaireDragDrop<StalactitesMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    handleReset();
  }, [handleReset, hideActionLog]);

  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: handleHint, label: 'hint' },
      { key: 'a', action: handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: handleUndo, label: 'undo' },
    ],
    [handleHint, handleAutoComplete, confirmGiveUpAction, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  const [hoveredStack, setHoveredStack] = useState<{ col: number; cardIdx: number } | null>(null);

  if (!state) return <GameSkeleton gameKey="stalactites" layout={{ kind: 'tableau', topRow: 8, tableau: 8 }} />;

  const isPlaying = state.phase === StalactitesPhase.PLAYING;
  const isGameClear = state.phase === StalactitesPhase.GAME_CLEAR;
  const isGameOver = state.phase === StalactitesPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  // Supermove limit: a tableau stack of N cards can only move when the domain
  // says so. **上限はドメインが決める。**ここで一般式 ((1 + cells) *
  // 2^emptyCols) を持つと、空き列自身を経由地に使えないぶんの差
  // (maxMovableCardsToEmptyColumn) が抜け、空き列宛ての束を「動かせる」と
  // 見せてサーバーに弾かれる (#5975)。
  const supermoveLimit = state.maxMovableCards;
  const emptyColLimit = state.maxMovableCardsToEmptyColumn;

  // 選択中の束の枚数。空き列が受け取れるかはこれと emptyColLimit で決まる。
  const selectedStackSize =
    selectedSource?.zone === 'tableau' && selectedSource.col !== undefined && selectedSource.cardIndex !== undefined
      ? (state.tableau[selectedSource.col]?.length ?? 0) - selectedSource.cardIndex
      : 0;
  // Auto-complete will deterministically win once every column is descending.
  const autoCompleteReady = stalactitesAutoCompleteReady(state.tableau, state.baseRank);

  const isSourceSelected = (zone: string, col?: number, cell?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cell === cell &&
    selectedSource.cardIndex === cardIndex;

  return (
    <GamePageShell
      title={tc('nav.stalactites')}
      gameThemeBg={gameTheme.stalactites.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/stalactites"
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
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* Free cells + Foundation row */}
            <div className="flex gap-2 mb-3 items-start flex-wrap">
              {/* Free cells */}
              <div className="flex gap-2" data-tutorial="fc-free-cells">
                {state.cells.map((card: Card | null, idx: number) => {
                  const stalactitesZone: StalactitesMoveZone = { zone: 'stalactites', cell: idx };
                  return (
                    <div key={`fc-${idx.toString()}`} className="text-center">
                      <div className="text-game-text-muted text-xs mb-1">
                        <span className="hidden sm:inline">
                          {t('stalactites')} {idx}
                        </span>
                        <span className="sm:hidden">
                          {t('stalactitesShort')}
                          {idx}
                        </span>
                      </div>
                      <DropZone
                        isDropTarget={dnd.isDropTarget(stalactitesZone)}
                        onDragOver={dnd.handleDragOver(stalactitesZone)}
                        onDrop={dnd.handleDrop(stalactitesZone)}
                        onDragLeave={dnd.handleDragLeave}
                      >
                        {card ? (
                          <button
                            type="button"
                            onClick={(e) => {
                              // The second click of a double-click also fires
                              // onClick (detail === 2); ignore it so onDoubleClick
                              // owns the foundation shortcut and selection stays put.
                              if (e.detail >= 2) return;
                              handleSelectSource(stalactitesZone);
                            }}
                            onDoubleClick={() => handleFoundationShortcut(stalactitesZone, card)}
                            disabled={!isPlaying || loading}
                            aria-label={cardAlt(card)}
                            aria-pressed={isSourceSelected('stalactites', undefined, idx)}
                            draggable={isPlaying && !loading}
                            onDragStart={dnd.handleDragStart(stalactitesZone)}
                            onDragEnd={dnd.handleDragEnd}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('stalactites', undefined, idx) ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(stalactitesZone) ? 'opacity-50' : ''}`}
                          >
                            <AnimatedCard card={card} width={cardWidth} draggable={false} />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => handleSelectTarget(stalactitesZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={t('emptyCellAriaLabel', { idx: String(idx) })}
                            style={{ width: cardWidth, height: cardHeight }}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                          >
                            {t('empty')}
                          </button>
                        )}
                      </DropZone>
                    </div>
                  );
                })}
              </div>

              <div className="w-4" />

              {/* Foundation piles */}
              <div className="flex gap-2" data-tutorial="fc-foundation">
                {state.foundation.map((pile: Card[], idx: number) => {
                  const foundationZone: StalactitesMoveZone = { zone: 'foundation', col: idx };
                  return (
                    <div key={`f-${idx.toString()}`} className="text-center">
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
                              cardCount: String(pile.length),
                            })}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                          >
                            <AnimatedCard
                              card={pile[pile.length - 1]}
                              width={cardWidth}
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
                            style={{ width: cardWidth, height: cardHeight }}
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
            </div>

            {/* Max bulk-move (supermove) limit, derived from empty free cells/columns */}
            <div className="text-game-text-muted text-xs mb-2" data-testid="fc-supermove-limit">
              {t('supermoveLimitLabel', { limit: supermoveLimit })}
              {emptyColLimit > 0 && <> {t('supermoveToEmpty', { limit: emptyColLimit })}</>}
            </div>

            {/* Tableau */}
            <div className="relative">
              <div className="flex gap-0.5 sm:gap-2 mb-3" data-tutorial="fc-tableau">
                {state.tableau.map((col: (Card | null)[], colIdx: number) => {
                  const tableauColZone: StalactitesMoveZone = { zone: 'tableau', col: colIdx };
                  return (
                    <div key={`col-${colIdx.toString()}`} className="flex-1 min-w-0">
                      <DropZone
                        isDropTarget={dnd.isDropTarget(tableauColZone)}
                        onDragOver={dnd.handleDragOver(tableauColZone)}
                        onDrop={dnd.handleDrop(tableauColZone)}
                        onDragLeave={dnd.handleDragLeave}
                        className="relative block"
                      >
                        <div className="relative" style={{ minHeight: cardHeight }}>
                          {col.length === 0 ? (
                            <button
                              type="button"
                              onClick={() => handleSelectTarget(tableauColZone)}
                              disabled={!isPlaying || loading || !selectedSource}
                              style={{ height: cardHeight }}
                              data-testid={`fc-empty-col-${colIdx.toString()}`}
                              // 空き列だけ上限が低い。選んだ束が超えているなら、
                              // クリックする前に分かるようにする (#5975)。
                              data-empty-col-blocked={selectedStackSize > emptyColLimit ? 'true' : undefined}
                              title={
                                selectedStackSize > emptyColLimit
                                  ? t('emptyColLimitTooltip', { limit: emptyColLimit, size: selectedStackSize })
                                  : undefined
                              }
                              className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite} ${
                                selectedStackSize > emptyColLimit ? 'opacity-50' : ''
                              }`}
                            >
                              K
                            </button>
                          ) : (
                            col.map((card: Card | null, cardIdx: number) => {
                              const cardZone: StalactitesMoveZone = {
                                zone: 'tableau',
                                col: colIdx,
                                cardIndex: cardIdx,
                              };
                              const stackSize = col.length - cardIdx;
                              const exceedsSupermove = stackSize > supermoveLimit;
                              const isInHoveredBlock =
                                hoveredStack !== null &&
                                hoveredStack.col === colIdx &&
                                cardIdx >= hoveredStack.cardIdx &&
                                col.length - hoveredStack.cardIdx <= supermoveLimit;
                              return (
                                <div
                                  key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                                  className="absolute left-0 right-0"
                                  style={{ top: cardIdx * cardOverlap }}
                                >
                                  {card ? (
                                    <button
                                      type="button"
                                      onClick={(e) => {
                                        // The second click of a double-click also
                                        // fires onClick (detail === 2); ignore it so
                                        // onDoubleClick owns the foundation shortcut
                                        // without issuing a stray self-target move.
                                        if (e.detail >= 2) return;
                                        if (selectedSource) {
                                          handleSelectTarget(tableauColZone);
                                        } else {
                                          handleSelectSource(cardZone);
                                        }
                                      }}
                                      onDoubleClick={
                                        // Only the exposed top card of a column can
                                        // move to a foundation.
                                        cardIdx === col.length - 1
                                          ? () => handleFoundationShortcut(cardZone, card)
                                          : undefined
                                      }
                                      disabled={!isPlaying || loading}
                                      // 上限超過は title とリングだけで示していたので、
                                      // ホバーできる人にしか届かない。draggable も落として
                                      // いるのに、動かせない理由が読み上げに出ない (#5820)。
                                      aria-label={
                                        exceedsSupermove
                                          ? `${cardAlt(card)} — ${t('supermoveLimitTooltip', { limit: supermoveLimit })}`
                                          : cardAlt(card)
                                      }
                                      aria-pressed={isSourceSelected('tableau', colIdx, undefined, cardIdx)}
                                      draggable={isPlaying && !loading && !exceedsSupermove}
                                      onDragStart={dnd.handleDragStart(cardZone)}
                                      onDragEnd={dnd.handleDragEnd}
                                      onMouseEnter={() => setHoveredStack({ col: colIdx, cardIdx })}
                                      onMouseLeave={() => setHoveredStack(null)}
                                      onFocus={() => setHoveredStack({ col: colIdx, cardIdx })}
                                      onBlur={() => setHoveredStack(null)}
                                      title={
                                        exceedsSupermove
                                          ? t('supermoveLimitTooltip', { limit: supermoveLimit })
                                          : undefined
                                      }
                                      data-supermove-blocked={exceedsSupermove ? 'true' : undefined}
                                      data-supermove-block={isInHoveredBlock ? 'true' : undefined}
                                      className={[
                                        'p-0 border-0 bg-transparent cursor-pointer w-full rounded',
                                        focusRingWhite,
                                        isSourceSelected('tableau', colIdx, undefined, cardIdx) &&
                                          'ring-2 ring-ds-warning',
                                        dnd.isDragSource(cardZone) && 'opacity-50',
                                        exceedsSupermove && 'opacity-60 ring-1 ring-ds-error',
                                        isInHoveredBlock && 'ring-2 ring-ds-success',
                                      ]
                                        .filter(Boolean)
                                        .join(' ')}
                                    >
                                      <AnimatedCard
                                        card={card}
                                        width={cardWidth}
                                        draggable={false}
                                        style={{ width: '100%' }}
                                      />
                                    </button>
                                  ) : (
                                    <div style={{ width: cardWidth, height: cardHeight }} />
                                  )}
                                </div>
                              );
                            })
                          )}
                          {col.length > 0 && <div style={{ height: (col.length - 1) * cardOverlap + cardHeight }} />}
                        </div>
                      </DropZone>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Hint display */}
            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-testid="stalactites-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2" data-testid="fc-hint-line">
                  {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromCol)} →{' '}
                  {formatHintZone(t, hint.toZone, hint.toCol)}
                </div>
              )}
            </div>
            <div className="flex justify-center">
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

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

          <GameFooter className={`${gameTheme.stalactites.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="fc-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleUndo}
                    disabled={loading || isAutoCompleting || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={handleUndoEscape}
                      disabled={loading || isAutoCompleting}
                    />
                  )}
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleHint}
                    disabled={loading || isAutoCompleting}
                  >
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess}${
                      autoCompleteReady && !loading && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''
                    }`}
                    onClick={handleAutoComplete}
                    disabled={loading || isAutoCompleting || !autoCompleteReady}
                    data-testid="autocomplete-button"
                    title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
                  >
                    {t('autoComplete')}
                  </button>
                  <AutoCompleteReadyBadge ready={autoCompleteReady} testId="stalactites-autocomplete-ready-badge" />
                  <button
                    type="button"
                    className={btnDanger}
                    onClick={confirmGiveUpAction}
                    disabled={loading || isAutoCompleting}
                  >
                    {t('giveup')}
                  </button>
                </div>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="fc-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="free-cell-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
