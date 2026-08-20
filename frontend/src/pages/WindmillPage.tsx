import { useCallback, useMemo } from 'react';
import type { windmillApi } from '../api/gameApi';
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
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSolitaireDragDrop } from '../hooks/useSolitaireDragDrop';
import { useWindmillGame } from '../hooks/useWindmillGame';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { WindmillMoveZone, WindmillResponse } from '../types/card';
import { WindmillPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseWindmillCommand, WINDMILL_HELP } from '../utils/cli/commands/windmillCommands';
import { formatWindmillState } from '../utils/cli/formatters/windmillFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const SAIL_CNT = 8;
const CORNER_CNT = 4;
const TOTAL_CARDS = 104;

const WM_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="wm-center"]', messageKey: 'tutorial.center', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="wm-corners"]', messageKey: 'tutorial.corners', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="wm-sails"]', messageKey: 'tutorial.sails', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="wm-stock"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="wm-corners"]', messageKey: 'tutorial.transfer', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="wm-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Windmill page: eight sails, one centre foundation, and four corner foundations. */
export const WindmillPage = withTutorial(WindmillPageContent, 'windmill', WM_TUTORIAL_STEPS);

function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'center') return t('frontendHint.center');
  if (zone === 'corner') return t('frontendHint.corner', { idx });
  if (zone === 'sail') return t('frontendHint.sail', { idx });
  if (zone === 'stock') return t('frontendHint.stock');
  return t('frontendHint.waste');
}

function WindmillPageContent() {
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
  } = useGamePageSetup('windmill');
  const game = useWindmillGame();
  const { state, loading, error, retry, hintError, selectedSource, hint, isAutoCompleting } = game;

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('windmill', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('windmill');
  const cliConfig: CliGameConfig<WindmillResponse, Parameters<typeof windmillApi.exec>> = useMemo(
    () => ({
      gameName: 'windmill',
      parseCommand: parseWindmillCommand,
      formatResponse: formatWindmillState,
      helpText: WINDMILL_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardHeight, cardWidth } = useCardDimensions();

  const isPlayingForKbd = state?.phase === WindmillPhase.PLAYING;

  const dispatchMove = useCallback(
    (source: WindmillMoveZone, target: WindmillMoveZone) => {
      void game.exec('move', source, target);
    },
    [game],
  );
  const dnd = useSolitaireDragDrop<WindmillMoveZone>({
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
    return <GameSkeleton gameKey="windmill" layout={{ kind: 'tableau', topRow: 5, tableau: SAIL_CNT }} />;
  }

  const isPlaying = state.phase === WindmillPhase.PLAYING;
  const isGameClear = state.phase === WindmillPhase.GAME_CLEAR;
  const isGameOver = state.phase === WindmillPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  const foundationCount = isGameOver
    ? state.center.length + state.corners.reduce((sum, pile) => sum + pile.length, 0)
    : 0;
  // Auto-complete only ever moves sail and waste cards, so it is pointless until
  // one of those piles holds something.
  const autoCompleteReady = state.sails.some((c) => c !== null) || state.waste.length > 0;

  const isSourceSelected = (zone: string, col?: number) =>
    selectedSource !== null && selectedSource.zone === zone && selectedSource.col === col;

  const centerZone: WindmillMoveZone = { zone: 'center' };
  const centerTop = state.center.length > 0 ? state.center[state.center.length - 1] : null;
  const wasteTop = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const wasteZone: WindmillMoveZone = { zone: 'waste' };

  const renderSail = (idx: number) => {
    const card = state.sails[idx] ?? null;
    const sailZone: WindmillMoveZone = { zone: 'sail', col: idx };
    return (
      <div key={`sail-${idx.toString()}`} className="text-center">
        <div className="text-xs text-ds-text-muted mb-0.5" aria-hidden="true">
          #{idx}
        </div>
        {card ? (
          <button
            type="button"
            onClick={() => game.handleSelectSource(sailZone)}
            disabled={!isPlaying || loading}
            aria-label={t('sailAriaLabel', { card: cardAlt(card), idx })}
            aria-pressed={isSourceSelected('sail', idx)}
            draggable={isPlaying && !loading}
            onDragStart={dnd.handleDragStart(sailZone)}
            onDragEnd={dnd.handleDragEnd}
            className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('sail', idx) ? 'ring-2 ring-ds-warning' : ''} ${dnd.isDragSource(sailZone) ? 'opacity-50' : ''}`}
          >
            <AnimatedCard card={card} width={cardWidth} draggable={false} />
          </button>
        ) : (
          <div
            role="img"
            aria-label={t('emptySailAriaLabel', { idx })}
            style={{ width: cardWidth, height: cardHeight }}
            className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
          >
            {t('empty')}
          </div>
        )}
      </div>
    );
  };

  const renderCorner = (idx: number) => {
    const pile = state.corners[idx] ?? [];
    const cornerZone: WindmillMoveZone = { zone: 'corner', col: idx };
    const top = pile.length > 0 ? pile[pile.length - 1] : null;
    // **移動先としては塞がない。**四隅は「降順に積む置き場」と「中央への引き戻し
    // 元」を兼ねていて、transferBlocked が禁じるのは後者だけ。ボタンごと無効に
    // すると、影響を受けないはずの「四隅へ置く」手まで潰れる。
    const blockedAsSource = selectedSource === null && state.transferBlocked;
    return (
      <div key={`corner-${idx.toString()}`} className="text-center">
        <div className="text-game-text-muted text-xs mb-1" aria-hidden="true">
          K{idx}
        </div>
        <DropZone
          isDropTarget={dnd.isDropTarget(cornerZone)}
          onDragOver={dnd.handleDragOver(cornerZone)}
          onDrop={dnd.handleDrop(cornerZone)}
          onDragLeave={dnd.handleDragLeave}
        >
          {top ? (
            <button
              type="button"
              onClick={() =>
                selectedSource ? game.handleSelectTarget(cornerZone) : game.handleSelectSource(cornerZone)
              }
              disabled={!isPlaying || loading || isAutoCompleting || blockedAsSource}
              title={blockedAsSource ? t('transferBlocked') : undefined}
              aria-label={t('cornerAriaLabel', { idx, count: pile.length })}
              aria-pressed={isSourceSelected('corner', idx)}
              draggable={isPlaying && !loading && !blockedAsSource}
              onDragStart={dnd.handleDragStart(cornerZone)}
              onDragEnd={dnd.handleDragEnd}
              className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('corner', idx) ? 'ring-2 ring-ds-warning' : ''}`}
            >
              <AnimatedCard
                card={top}
                width={cardWidth}
                draggable={false}
                dealDelay={isAutoCompleting ? idx * 0.1 : 0}
              />
            </button>
          ) : (
            <button
              type="button"
              onClick={() => game.handleSelectTarget(cornerZone)}
              disabled={!isPlaying || loading || !selectedSource}
              aria-label={t('emptyCornerAriaLabel', { idx })}
              style={{ width: cardWidth, height: cardHeight }}
              className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
            >
              {t('kingsOnly')}
            </button>
          )}
        </DropZone>
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.windmill')}
      gameThemeBg={gameTheme.windmill.bg}
      phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      gamePath="/windmill"
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
              <div className="text-center" data-tutorial="wm-center">
                <div className="text-game-text-muted text-xs mb-1">{t('center')}</div>
                <DropZone
                  isDropTarget={dnd.isDropTarget(centerZone)}
                  onDragOver={dnd.handleDragOver(centerZone)}
                  onDrop={dnd.handleDrop(centerZone)}
                  onDragLeave={dnd.handleDragLeave}
                >
                  {centerTop ? (
                    <button
                      type="button"
                      onClick={() => game.handleSelectTarget(centerZone)}
                      disabled={!isPlaying || loading || isAutoCompleting || !selectedSource}
                      aria-label={t('centerAriaLabel', { count: state.center.length })}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                    >
                      <AnimatedCard card={centerTop} width={cardWidth} draggable={false} />
                    </button>
                  ) : (
                    <button
                      type="button"
                      onClick={() => game.handleSelectTarget(centerZone)}
                      disabled={!isPlaying || loading || !selectedSource}
                      aria-label={t('emptyCenterAriaLabel')}
                      style={{ width: cardWidth, height: cardHeight }}
                      className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                    >
                      A
                    </button>
                  )}
                </DropZone>
                <div className="text-game-text-muted text-xs mt-1">{state.center.length}/52</div>
              </div>

              <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="wm-corners">
                {Array.from({ length: CORNER_CNT }, (_, i) => i).map(renderCorner)}
              </div>

              <div className="flex gap-2 items-start" data-tutorial="wm-stock">
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('stock')}</div>
                  <button
                    type="button"
                    onClick={game.handleDraw}
                    disabled={!isPlaying || loading || isAutoCompleting || state.stockCount === 0}
                    aria-label={t('stockAriaLabel', { count: state.stockCount })}
                    style={{ width: cardWidth, height: cardHeight }}
                    className={`rounded border-2 border-white/30 bg-white/10 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                  >
                    {state.stockCount}
                  </button>
                </div>
                <div className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
                  {wasteTop ? (
                    <button
                      type="button"
                      onClick={() => game.handleSelectSource(wasteZone)}
                      disabled={!isPlaying || loading}
                      aria-label={cardAlt(wasteTop)}
                      aria-pressed={isSourceSelected('waste', undefined)}
                      draggable={isPlaying && !loading}
                      onDragStart={dnd.handleDragStart(wasteZone)}
                      onDragEnd={dnd.handleDragEnd}
                      className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('waste', undefined) ? 'ring-2 ring-ds-warning' : ''}`}
                    >
                      <AnimatedCard card={wasteTop} width={cardWidth} draggable={false} />
                    </button>
                  ) : (
                    <div
                      role="img"
                      aria-label={t('emptyWasteAriaLabel')}
                      style={{ width: cardWidth, height: cardHeight }}
                      className="rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center"
                    >
                      {t('empty')}
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* The block is invisible in the layout, so the board has to state it. */}
            {isPlaying && state.transferBlocked && (
              <p className="text-ds-warning text-sm text-center mb-2">{t('transferBlocked')}</p>
            )}

            <div className="flex flex-wrap justify-center gap-1 sm:gap-2" data-tutorial="wm-sails">
              {Array.from({ length: SAIL_CNT }, (_, i) => i).map(renderSail)}
            </div>

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-tutorial="wm-hint-display" data-testid="wm-hint-live" role="status" aria-live="polite">
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
              <p data-testid="wm-gameover-summary" className="text-ds-text-muted text-sm text-center mt-1">
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

          <GameFooter className={`${gameTheme.windmill.footer} px-4 py-2.5`}>
            <ErrorAlert message={error ?? hintError} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isPlaying && (
                <div data-tutorial="wm-controls">
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
                dataTutorial="wm-reset-button"
              />
            </div>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="windmill-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
