import { useCallback, useMemo } from 'react';
import type { KingAlbertMoveZone, kingAlbertApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
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
import { useCardDimensions, useWindowWidth } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useKingAlbertGame } from '../hooks/useKingAlbertGame';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, KingAlbertResponse } from '../types/card';
import { KingAlbertPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { KINGALBERT_HELP, parseKingAlbertCommand } from '../utils/cli/commands/kingalbertCommands';
import { formatKingAlbertState } from '../utils/cli/formatters/kingalbertFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { kingAlbertLegalTargets } from '../utils/kingAlbertLegalTargets';
import { hintCheckboxItem } from '../utils/settingsItems';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

const KA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ka-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ka-reserve"]',
    messageKey: 'tutorial.reserve',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ka-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ka-tableau"]',
    messageKey: 'tutorial.moves',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ka-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the King Albert solitaire game page with 9 tableau columns, a 7-card reserve, and 4 foundations. */
export const KingAlbertPage = withTutorial(KingAlbertPageContent, 'kingalbert', KA_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, col: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation');
  if (zone === 'reserve') return t('frontendHint.reserve', { col });
  return t('frontendHint.tableau', { col });
}

function KingAlbertPageContent() {
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
  } = useGamePageSetup('kingalbert');
  const game = useKingAlbertGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('kingalbert', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('kingalbert');
  const cliConfig: CliGameConfig<KingAlbertResponse, Parameters<typeof kingAlbertApi.exec>> = useMemo(
    () => ({
      gameName: 'kingalbert',
      parseCommand: parseKingAlbertCommand,
      formatResponse: formatKingAlbertState,
      helpText: KINGALBERT_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const dims = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    // Layout has 9 tableau columns side by side on mobile.
    const totalCols = 9;
    const colW = Math.floor((windowWidth - padX - (totalCols - 1) * gapPx) / totalCols);
    const cw = Math.min(Math.max(colW, 30), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === KingAlbertPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: KingAlbertMoveZone, target: KingAlbertMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<KingAlbertMoveZone>({
    onMove: dispatchMove,
    isPlaying: !!isPlayingForKbd,
    disabled: loading,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    game.handleReset();
  }, [game, hideActionLog]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useCallback(
    () => requestGiveUpConfirm(game.handleGiveUp),
    [requestGiveUpConfirm, game.handleGiveUp],
  );

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: game.handleHint, label: 'hint' },
      { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: game.handleUndo, label: 'undo' },
    ],
    [game, confirmGiveUpAction],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <GameSkeleton gameKey="kingalbert" layout={{ kind: 'tableau', topRow: 4, tableau: 9 }} />;

  const isPlaying = state.phase === KingAlbertPhase.PLAYING;
  const isGameClear = state.phase === KingAlbertPhase.GAME_CLEAR;
  const isGameOver = state.phase === KingAlbertPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  // How many of the 52 cards reached a foundation — only needed on game over, so skip the
  // reduce during normal (frequently re-rendered) play.
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;
  // Auto-complete becomes useful once a foundation has built past its ace, so
  // pulse the button only then (mirrors Crescent / Spiderette).
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 1);

  const isSourceSelected = (zone: string, col?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cardIndex === cardIndex;

  // 選択中の札そのもの。タブローは最上段しか動かせないので、列の一番上を読む。
  const selectedCard = ((): Card | null => {
    if (!selectedSource || selectedSource.col === undefined) return null;
    if (selectedSource.zone === 'reserve') return state.reserve[selectedSource.col] ?? null;
    const col = state.tableau[selectedSource.col];
    return col?.[col.length - 1]?.card ?? null;
  })();
  // リングは**置ける先だけ**に付ける。「選択中なら全部光る」だと、エラーになる手を
  // 勧めることになる (#5598)。
  const legalTargets = kingAlbertLegalTargets(state.tableau, state.foundation, selectedCard);

  const renderTableauColumn = (colIdx: number) => {
    const col = state.tableau[colIdx];
    const tableauColZone: KingAlbertMoveZone = { zone: 'tableau', col: colIdx };
    return (
      <div key={`col-${colIdx.toString()}`} className="flex-1 min-w-0">
        <DropZone
          isDropTarget={dnd.isDropTarget(tableauColZone)}
          onDragOver={dnd.handleDragOver(tableauColZone)}
          onDrop={dnd.handleDrop(tableauColZone)}
          onDragLeave={dnd.handleDragLeave}
          className="relative block"
        >
          <div className="relative" style={{ minHeight: dims.ch }}>
            {col.length === 0 ? (
              <button
                type="button"
                onClick={() => game.handleSelectTarget(tableauColZone)}
                disabled={!isPlaying || loading || !selectedSource}
                style={{ height: dims.ch }}
                data-target-candidate={legalTargets.tableau.has(colIdx) || undefined}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite} ${
                  legalTargets.tableau.has(colIdx) ? 'ring-1 ring-ds-info' : ''
                }`}
              >
                {t('empty')}
              </button>
            ) : (
              col.map((tc, cardIdx) => {
                const cardZone: KingAlbertMoveZone = {
                  zone: 'tableau',
                  col: colIdx,
                  cardIndex: cardIdx,
                };
                const isTop = cardIdx === col.length - 1;
                const isTargetCandidate =
                  isTop && legalTargets.tableau.has(colIdx) && !isSourceSelected('tableau', colIdx, cardIdx);
                return (
                  <div
                    key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                    className="absolute left-0 right-0"
                    style={{ top: cardIdx * dims.co }}
                  >
                    {tc.card ? (
                      <button
                        type="button"
                        onClick={() => {
                          if (selectedSource) {
                            game.handleSelectTarget(tableauColZone);
                          } else if (isTop) {
                            game.handleSelectSource(cardZone);
                          }
                        }}
                        disabled={!isPlaying || loading || (!isTop && !selectedSource)}
                        aria-label={cardAlt(tc.card)}
                        aria-pressed={isSourceSelected('tableau', colIdx, cardIdx)}
                        draggable={isPlaying && !loading && isTop}
                        onDragStart={dnd.handleDragStart(cardZone)}
                        onDragEnd={dnd.handleDragEnd}
                        // With a card in hand every top card is a possible
                        // destination; only the source said so before (#4828).
                        data-target-candidate={isTargetCandidate || undefined}
                        className={`p-0 border-0 bg-transparent w-full rounded ${focusRingWhite} ${isTop ? 'cursor-pointer' : 'cursor-default'} ${isSourceSelected('tableau', colIdx, cardIdx) ? 'ring-2 ring-ds-warning' : ''} ${isTargetCandidate ? 'ring-1 ring-ds-info motion-safe:hover:ring-2 focus:ring-2' : ''} ${dnd.isDragSource(cardZone) ? 'opacity-50' : ''}`}
                      >
                        <AnimatedCard
                          card={tc.card}
                          width={dims.cw}
                          draggable={false}
                          style={{ width: '100%' }}
                          wrapperClassName="block w-full"
                        />
                      </button>
                    ) : null}
                  </div>
                );
              })
            )}
            {col.length > 0 && <div style={{ height: (col.length - 1) * dims.co + dims.ch }} />}
          </div>
        </DropZone>
      </div>
    );
  };

  const renderReserveCell = (cellIdx: number) => {
    const reserveCard = state.reserve[cellIdx];
    const reserveZone: KingAlbertMoveZone = { zone: 'reserve', col: cellIdx };
    // Label matches the CUI (`[r0]`..`[r6]`) and the 0-based reserve number in the hint
    // text (`t('frontendHint.reserve', { col })`), so players can map a hint to its slot.
    const slotLabel = (
      <span
        data-testid={`ka-reserve-label-${cellIdx.toString()}`}
        className="text-game-text-muted text-[10px] leading-none mt-0.5"
      >
        r{cellIdx}
      </span>
    );
    if (!reserveCard) {
      return (
        <div key={`r-${cellIdx.toString()}`} className="flex flex-col items-center">
          <div
            role="img"
            // 番号は 0 始まりで通す。可視バッジは r{cellIdx}、CUI は [r0]〜[r6]、
            // ヒントの reserve col も 0 始まり。ここだけ +1 すると、
            // 「reserve 0 へ」と言われた枠が「空のリザーブ枠 1」と読まれる (#6394)。
            aria-label={t('emptyReserveSlot', { idx: cellIdx })}
            className="rounded border border-dashed border-white/10"
            style={{ width: dims.cw, height: dims.ch }}
          />
          {slotLabel}
        </div>
      );
    }
    return (
      <div key={`r-${cellIdx.toString()}`} className="flex flex-col items-center">
        <button
          type="button"
          onClick={() => game.handleSelectSource(reserveZone)}
          disabled={!isPlaying || loading}
          aria-label={cardAlt(reserveCard)}
          aria-pressed={isSourceSelected('reserve', cellIdx)}
          draggable={isPlaying && !loading}
          onDragStart={dnd.handleDragStart(reserveZone)}
          onDragEnd={dnd.handleDragEnd}
          className={`p-0 border-0 bg-transparent rounded cursor-pointer ${focusRingWhite} ${isSourceSelected('reserve', cellIdx) ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(reserveZone) ? 'opacity-50' : ''}`}
        >
          <AnimatedCard card={reserveCard} width={dims.cw} draggable={false} />
        </button>
        {slotLabel}
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.kingalbert')}
      gameThemeBg={gameTheme.kingalbert.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/kingalbert"
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
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            <div className="flex gap-2 sm:gap-3 items-start justify-center mb-3">
              <div className="flex flex-wrap gap-1 sm:gap-2 items-start" data-tutorial="ka-reserve">
                <span className="self-center text-game-text-muted text-xs mr-1">{t('reserve')}</span>
                {state.reserve.map((_, idx) => renderReserveCell(idx))}
              </div>

              <div className="flex items-start gap-1 sm:gap-2" data-tutorial="ka-foundation">
                {state.foundation.map((pile, idx) => {
                  const foundationZone: KingAlbertMoveZone = { zone: 'foundation', col: idx };
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
                            onClick={() => game.handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || isAutoCompleting || !selectedSource}
                            aria-label={t('foundationAriaLabel', {
                              suit: FOUNDATION_SUITS[idx],
                              count: pile.length,
                            })}
                            data-target-candidate={legalTargets.foundation.has(idx) || undefined}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${
                              legalTargets.foundation.has(idx)
                                ? 'ring-1 ring-ds-info motion-safe:hover:ring-2 focus:ring-2'
                                : ''
                            }`}
                          >
                            <AnimatedCard
                              card={pile[pile.length - 1]}
                              width={dims.cw}
                              draggable={false}
                              dealDelay={isAutoCompleting ? idx * 0.15 : 0}
                            />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => game.handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={t('emptyFoundationAriaLabel', { suit: FOUNDATION_SUITS[idx] })}
                            style={{ width: dims.cw, height: dims.ch }}
                            // 空の枠は A の唯一の行き先。ここを暗いままにすると、
                            // 「置ける先には光る」が片側だけ嘘になる。
                            data-target-candidate={legalTargets.foundation.has(idx) || undefined}
                            className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite} ${
                              legalTargets.foundation.has(idx)
                                ? 'ring-1 ring-ds-info motion-safe:hover:ring-2 focus:ring-2'
                                : ''
                            }`}
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

            <div className="flex gap-1 sm:gap-2" data-tutorial="ka-tableau">
              {[0, 1, 2, 3, 4, 5, 6, 7, 8].map(renderTableauColumn)}
            </div>

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-tutorial="ka-hint-display" data-testid="ka-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 mt-3">
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

            {isGameOver && (
              <p data-testid="ka-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
                {t('gameOverSummary', {
                  count: foundationCount,
                  percent: Math.round((foundationCount / 52) * 100),
                })}
              </p>
            )}

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.kingalbert.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="ka-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={game.handleUndo}
                    disabled={loading || isAutoCompleting || !state.canUndo}
                  >
                    {t('undo')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? 0}
                      onEscape={game.handleUndoEscape}
                      disabled={loading || isAutoCompleting}
                    />
                  )}
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={game.handleHint}
                    disabled={loading || isAutoCompleting}
                  >
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess}${autoCompleteReady && !loading && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={game.handleAutoComplete}
                    disabled={loading || isAutoCompleting || !autoCompleteReady}
                    data-testid="autocomplete-button"
                    title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
                  >
                    {t('autoComplete')}
                  </button>
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
                dataTutorial="ka-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="king-albert-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
