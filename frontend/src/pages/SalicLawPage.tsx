import { useCallback, useMemo } from 'react';
import type { salicLawApi } from '../api/gameApi';
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
import { useSalicLawGame } from '../hooks/useSalicLawGame';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SalicLawMoveZone, SalicLawResponse } from '../types/card';
import { SalicLawPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSalicLawCommand, SALICLAW_HELP } from '../utils/cli/commands/saliclawCommands';
import { formatSalicLawState } from '../utils/cli/formatters/saliclawFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

// **組札はスートで決まらない。**A から J まで、どのスートでも積める。列と
// 1 対 1 で対応するので、見出しは列番号でしか意味を持たない。
const TABLEAU_PILES = 8;
/** 組札に積める枚数。A から J までの 11 枚 × 8 山 = 88。K は土台に残り、Q は場に出ない。 */
const TOTAL_CARDS = 88;

const SL_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sl-queens"]', messageKey: 'tutorial.queens', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="sl-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sl-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sl-tableau"]', messageKey: 'tutorial.bareKing', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sl-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Salic Law page: eight king-based columns, eight foundations, and the retired queens. */
export const SalicLawPage = withTutorial(SalicLawPageContent, 'saliclaw', SL_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'foundation') return t('frontendHint.foundation', { idx });
  if (zone === 'stock') return t('frontendHint.stock');
  return t('frontendHint.tableau', { pile: idx });
}

function SalicLawPageContent() {
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
  } = useGamePageSetup('saliclaw');
  const game = useSalicLawGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('saliclaw', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('saliclaw');
  const cliConfig: CliGameConfig<SalicLawResponse, Parameters<typeof salicLawApi.exec>> = useMemo(
    () => ({
      gameName: 'saliclaw',
      parseCommand: parseSalicLawCommand,
      formatResponse: formatSalicLawState,
      helpText: SALICLAW_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardHeight, cardOverlap, cardWidth, isMobile } = useCardDimensions();
  const windowWidth = useWindowWidth();

  const dims = useMemo(() => {
    if (!isMobile) return { cw: cardWidth, ch: cardHeight, co: cardOverlap };
    const padX = 16;
    const gapPx = 4;
    const colW = Math.floor((windowWidth - padX - (TABLEAU_PILES - 1) * gapPx) / TABLEAU_PILES);
    const cw = Math.min(Math.max(colW, 30), cardWidth);
    const ch = Math.round(cw * 1.5);
    const co = Math.round(cw * 0.32);
    return { cw, ch, co };
  }, [isMobile, windowWidth, cardWidth, cardHeight, cardOverlap]);

  const isPlayingForKbd = state?.phase === SalicLawPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: SalicLawMoveZone, target: SalicLawMoveZone) => {
      // **置けるのは「K だけの列」だけ。**クリック経路はボタンを無効化して
      // 防いでいるが、**ドラッグ経路はここを通る**ので同じ規則をここでも見る
      // (#4906)。まだ開いていない列 (長さ 0) も置き先ではない。
      if (target.zone === 'tableau') {
        if (target.col === undefined) return;
        if ((state?.tableau[target.col]?.length ?? 0) !== 1) return;
      }
      // 土台の K は動かせないので、1 枚しかない列は移動元にならない。
      if (source.zone === 'tableau' && source.col !== undefined && (state?.tableau[source.col]?.length ?? 0) <= 1) {
        return;
      }
      void game.exec('move', source, target);
    },
    [game, state],
  );
  const dnd = useSolitaireDragDrop<SalicLawMoveZone>({
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
      { key: 'd', action: game.handleDraw, label: 'draw' },
      { key: 'h', action: game.handleHint, label: 'hint' },
      { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
      { key: 'g', action: confirmGiveUpAction, label: 'giveUp' },
      { key: 'z', action: game.handleUndo, label: 'undo' },
    ],
    [game, confirmGiveUpAction],
  );

  useActionKeyboardNav({ bindings: actionBindings, enabled: !!isPlayingForKbd && !loading });

  if (!state) {
    return <GameSkeleton gameKey="saliclaw" layout={{ kind: 'tableau', topRow: 8, tableau: TABLEAU_PILES }} />;
  }

  const isPlaying = state.phase === SalicLawPhase.PLAYING;
  const isGameClear = state.phase === SalicLawPhase.GAME_CLEAR;
  const isGameOver = state.phase === SalicLawPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const foundationCount = isGameOver ? state.foundation.reduce((sum, pile) => sum + pile.length, 0) : 0;
  const autoCompleteReady = state.foundation.some((pile) => pile.length > 0);

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  const renderPile = (pileIdx: number) => {
    const cards = state.tableau[pileIdx] ?? [];
    const pileZone: SalicLawMoveZone = { zone: 'tableau', col: pileIdx };
    // **K だけの列がこのゲーム唯一の置き場所。**まだ K が出ていない列
    // (長さ 0) とは別物なので、見た目でも操作でも区別する。
    const isBareKing = cards.length === 1;
    return (
      <div key={`pile-${pileIdx.toString()}`} className="flex-1 min-w-0">
        <div className="text-center text-xs text-ds-text-muted mb-0.5" aria-hidden="true">
          #{pileIdx}
        </div>
        <DropZone
          isDropTarget={dnd.isDropTarget(pileZone)}
          onDragOver={dnd.handleDragOver(pileZone)}
          onDrop={dnd.handleDrop(pileZone)}
          onDragLeave={dnd.handleDragLeave}
          className="relative block"
        >
          <div className="relative" style={{ minHeight: dims.ch }}>
            {cards.length === 0 ? (
              <button
                type="button"
                onClick={() => game.handleSelectTarget(pileZone)}
                // **まだ K が出ていない列は置き先ではない。**配りが進んで K が
                // 出るまで存在しないのと同じ。押せてしまうとサーバに弾かれる
                // まで気づけない (#4906)。
                disabled
                aria-label={t('emptyPileAriaLabel', { pile: pileIdx })}
                data-testid={`sl-unopened-${pileIdx.toString()}`}
                style={{ height: dims.ch }}
                className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center bg-transparent ${focusRingWhite}`}
              >
                {t('empty')}
              </button>
            ) : (
              cards.map((card, cardIdx) => {
                // Only the top card is playable; the ones under it are context.
                // 土台の K (列に 1 枚しかないとき) は動かせないので、移動元には
                // ならない ── ただし置き先としては唯一有効なので押せる。
                const isTop = cardIdx === cards.length - 1 && !isBareKing;
                return (
                  <div
                    key={`c-${pileIdx.toString()}-${cardIdx.toString()}`}
                    className="absolute left-0 right-0"
                    style={{ top: cardIdx * dims.co }}
                  >
                    <button
                      type="button"
                      onClick={() => {
                        if (selectedSource) {
                          game.handleSelectTarget(pileZone);
                        } else if (isTop) {
                          game.handleSelectSource(pileZone);
                        }
                      }}
                      // 置き先になれるのは K だけの列に限る。それ以外の列に
                      // 選択中の札を落とせると、拒まれるまで分からない。
                      disabled={!isPlaying || loading || (!isTop && !(selectedSource && isBareKing))}
                      aria-label={isBareKing ? t('bareKingPileAriaLabel', { pile: pileIdx }) : cardAlt(card)}
                      aria-pressed={isTop ? isSourceSelected('tableau', pileIdx) : undefined}
                      data-testid={isBareKing ? `sl-bare-king-${pileIdx.toString()}` : undefined}
                      draggable={isTop && isPlaying && !loading}
                      onDragStart={dnd.handleDragStart(pileZone)}
                      onDragEnd={dnd.handleDragEnd}
                      className={`p-0 border-0 bg-transparent w-full rounded cursor-pointer ${focusRingWhite} ${isTop && isSourceSelected('tableau', pileIdx) ? 'ring-2 ring-ds-warning' : ''}`}
                    >
                      <AnimatedCard
                        card={card}
                        width={dims.cw}
                        draggable={false}
                        style={{ width: '100%' }}
                        wrapperClassName="block w-full"
                      />
                    </button>
                  </div>
                );
              })
            )}
            {cards.length > 0 && <div style={{ height: (cards.length - 1) * dims.co + dims.ch }} />}
          </div>
        </DropZone>
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.saliclaw')}
      gameThemeBg={gameTheme.saliclaw.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/saliclaw"
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
            <div className="flex flex-wrap justify-center items-start gap-3 sm:gap-6 mb-3">
              <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="sl-foundation">
                {state.foundation.map((pile, idx) => {
                  const foundationZone: SalicLawMoveZone = { zone: 'foundation', col: idx };
                  return (
                    <div key={`f-${idx.toString()}`} className="text-center">
                      <div className="text-game-text-muted text-xs mb-1">F{idx}</div>
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
                            aria-label={t('foundationAriaLabel', { idx, count: pile.length })}
                            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                          >
                            <AnimatedCard
                              card={pile[pile.length - 1]}
                              width={dims.cw}
                              draggable={false}
                              dealDelay={isAutoCompleting ? idx * 0.1 : 0}
                            />
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => game.handleSelectTarget(foundationZone)}
                            disabled={!isPlaying || loading || !selectedSource}
                            aria-label={t('emptyFoundationAriaLabel', { idx })}
                            style={{ width: dims.cw, height: dims.ch }}
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

              <div className="flex gap-2 items-start" data-tutorial="sl-stock">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
                  {/* **捨て札は無い。**押すと今の列に 1 枚置かれる（K なら次の
                      列が開く）ので、これは移動元ではなく純粋な配りボタン。 */}
                  <button
                    type="button"
                    onClick={() => game.handleDraw()}
                    disabled={!isPlaying || loading || isAutoCompleting || state.stockCount === 0}
                    aria-label={
                      state.stockCount === 0
                        ? t('emptyStockAriaLabel')
                        : t('stockAriaLabel', { count: state.stockCount })
                    }
                    data-testid="sl-deal-button"
                    style={{ width: dims.cw, height: dims.ch }}
                    className={`rounded border-2 border-white/30 bg-white/10 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                  >
                    {state.stockCount}
                  </button>
                </div>
              </div>

              {/* **退場したクイーン。**8 枚が場から消えている理由は盤からは
                  読めないので、ゲーム名の由来ごと見せる。飾りなので操作しない。 */}
              <div className="text-center" data-tutorial="sl-queens">
                <div className="text-game-text-muted text-xs mb-1">{t('queens')}</div>
                <div
                  className="flex flex-wrap justify-center gap-0.5 opacity-60"
                  role="img"
                  aria-label={t('queensAriaLabel', { count: state.queens.length })}
                  data-testid="sl-queens"
                >
                  {state.queens.map((q, idx) => (
                    <AnimatedCard
                      key={`q-${idx.toString()}`}
                      card={q}
                      width={Math.round(dims.cw * 0.55)}
                      draggable={false}
                    />
                  ))}
                </div>
              </div>
            </div>

            <div className="flex gap-1 sm:gap-2 items-start" data-tutorial="sl-tableau">
              {Array.from({ length: TABLEAU_PILES }, (_, i) => i).map(renderPile)}
            </div>

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-tutorial="sl-hint-display" data-testid="sl-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2 mt-3">
                  {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromIdx)} →{' '}
                  {formatHintZone(t, hint.toZone, hint.toIdx)}
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
              <p data-testid="sl-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
                {t('gameOverSummary', {
                  count: foundationCount,
                  percent: Math.round((foundationCount / TOTAL_CARDS) * 100),
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
            groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
          />

          <GameFooter className={`${gameTheme.saliclaw.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="sl-controls">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={game.handleDraw}
                    disabled={loading || isAutoCompleting || state.stockCount === 0}
                  >
                    {t('draw')}
                  </button>
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
                dataTutorial="sl-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="saliclaw-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
