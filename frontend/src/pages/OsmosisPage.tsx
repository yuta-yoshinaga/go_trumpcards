import { useCallback, useMemo, useState } from 'react';
import { type OsmosisMoveZone, osmosisApi } from '../api/gameApi';
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
import type { Card, OsmosisResponse } from '../types/card';
import { OsmosisPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { OSMOSIS_HELP, parseOsmosisCommand } from '../utils/cli/commands/osmosisCommands';
import { formatOsmosisState } from '../utils/cli/formatters/osmosisFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { osmosisAllowedRanks, osmosisCanPlace } from '../utils/osmosisRules';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Card rank value (1–13) → short display label. */
const RANK_LABELS = ['', 'A', '2', '3', '4', '5', '6', '7', '8', '9', '10', 'J', 'Q', 'K'];

/** Tutorial steps for the Osmosis solitaire game. */
const OS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="os-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="os-reserve"]', messageKey: 'tutorial.reserve', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="os-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="os-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="os-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Osmosis solitaire game page. */
export const OsmosisPage = withTutorial(OsmosisPageContent, 'osmosis', OS_TUTORIAL_STEPS);

/** Inner content of the Osmosis page. */
function OsmosisPageContent() {
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
  } = useGamePageSetup('osmosis');
  const { state, loading, error, exec: execApi, retry } = useGameApi(osmosisApi.exec);
  const { cardWidth, cardHeight } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('osmosis', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('osmosis');
  const cliConfig: CliGameConfig<OsmosisResponse, Parameters<typeof osmosisApi.exec>> = useMemo(
    () => ({
      gameName: 'osmosis',
      parseCommand: parseOsmosisCommand,
      formatResponse: formatOsmosisState,
      helpText: OSMOSIS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  // Selected source card (waste or a reserve column top), or null.
  const [selected, setSelected] = useState<OsmosisMoveZone | null>(null);

  const handleReset = useCallback(() => {
    setSelected(null);
    execApi('reset');
  }, [execApi]);
  const handleDraw = useCallback(() => execApi('draw'), [execApi]);
  const handleGiveUp = useCallback(() => execApi('giveup'), [execApi]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);
  const handleHint = useCallback(() => execApi('hint'), [execApi]);
  const handleAutoComplete = useCallback(() => execApi('autocomplete'), [execApi]);
  const handleUndo = useCallback(() => execApi('undo'), [execApi]);

  const handleSelectSource = useCallback((zone: OsmosisMoveZone) => {
    setSelected((prev) => (prev && prev.zone === zone.zone && prev.col === zone.col ? null : zone));
  }, []);

  const handleFoundationClick = useCallback(
    (fIdx: number) => {
      if (!selected) return;
      execApi('move', selected, { zone: 'foundation', col: fIdx });
      setSelected(null);
    },
    [execApi, selected],
  );

  const theme = useMemo(() => gameTheme.osmosis, []);

  const phase = state?.phase ?? OsmosisPhase.PLAYING;
  const isPlaying = phase === OsmosisPhase.PLAYING;
  const isGameClear = phase === OsmosisPhase.GAME_CLEAR;
  const isEnded = phase === OsmosisPhase.GAME_CLEAR || phase === OsmosisPhase.GAME_OVER;

  // Drag-and-drop: dragging a waste/reserve top onto a foundation row issues the
  // same `move` command as the click-to-select flow, so both interactions coexist.
  const dispatchMove = useCallback(
    (source: OsmosisMoveZone, target: OsmosisMoveZone) => {
      execApi('move', source, target);
    },
    [execApi],
  );
  const dnd = useSolitaireDragDrop<OsmosisMoveZone>({
    onMove: dispatchMove,
    isPlaying,
    disabled: loading,
  });

  // 手詰まりはフェーズ表示ではなく GameMessageBox の messageCode
  // (osmosis.stalemate) で知らせる。姉妹のソリティアが全てそうしているので、
  // ここだけフェーズバッジを差し替えると同じことを二度言うことになる。
  const phaseName = isGameClear
    ? t('phase.gameClear')
    : phase === OsmosisPhase.GAME_OVER
      ? t('phase.gameOver')
      : t('phase.playing');

  // Bind letter-key shortcuts to the play-phase actions. Memoize so the effect
  // doesn't re-subscribe every render, and call the hook before any early return
  // to keep hook order stable. Enter/Space are intentionally NOT bound — they
  // natively activate a focused button and would double-fire.
  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw },
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'z', action: handleUndo },
      { key: 'g', action: confirmGiveUpAction },
    ],
    [handleDraw, handleHint, handleAutoComplete, handleUndo, confirmGiveUpAction],
  );
  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: state != null && isPlaying && !loading,
  });

  if (error) return <ErrorAlert message={error} onRetry={retry} />;
  if (!state) return null;

  const topWaste = state.waste.length > 0 ? state.waste[state.waste.length - 1] : null;
  const isSelected = (zone: OsmosisMoveZone) => !!selected && selected.zone === zone.zone && selected.col === zone.col;

  // Resolve a source zone (waste top or a reserve-column top) to its actual card,
  // used to flag foundation rows the card cannot be placed on.
  const topCardOf = (zone: OsmosisMoveZone | null): Card | null => {
    if (zone?.zone === 'waste') return topWaste;
    if (zone?.zone === 'reserve' && zone.col != null) {
      const col = state.reserve[zone.col] ?? [];
      return col.length > 0 ? col[col.length - 1] : null;
    }
    return null;
  };
  // The card currently held via click selection or an in-flight drag; both drive
  // the "cannot place here" foundation-row highlight.
  const selectedCard = topCardOf(selected);
  const draggedCard = topCardOf(dnd.dragSource);

  return (
    <GamePageShell
      title={tc('nav.osmosis')}
      gameThemeBg={theme.bg}
      phaseName={phaseName}
      gamePath="/osmosis"
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
            {/* Foundation rows */}
            <div className="mb-3 flex flex-col gap-2" data-tutorial="os-foundation">
              <span className="text-xs text-ds-text-muted">{t('foundation')}</span>
              {state.foundation.map((pile, i) => {
                const fZone: OsmosisMoveZone = { zone: 'foundation', col: i };
                const allowed = osmosisAllowedRanks(state.foundation, state.baseRank, i);
                // Click selection flags every invalid row; a drag only warns on the
                // row currently hovered (the drop target).
                const clickBlocked =
                  selectedCard != null && !osmosisCanPlace(state.foundation, state.baseRank, i, selectedCard);
                const isDropHover = dnd.isDropTarget(fZone);
                const dragBlocked =
                  isDropHover &&
                  draggedCard != null &&
                  !osmosisCanPlace(state.foundation, state.baseRank, i, draggedCard);
                const blocked = clickBlocked || dragBlocked;
                return (
                  <DropZone
                    key={`f-${i}`}
                    isDropTarget={isDropHover && !dragBlocked}
                    onDragOver={dnd.handleDragOver(fZone)}
                    onDrop={dnd.handleDrop(fZone)}
                    onDragLeave={dnd.handleDragLeave}
                  >
                    <button
                      type="button"
                      onClick={() => handleFoundationClick(i)}
                      disabled={!isPlaying || !selected || loading}
                      aria-label={`${t('foundation')} ${i}`}
                      // **理由は読み上げに乗せる。**`title` は支援技術が読み上げる
                      // 保証が無く、赤枠も見えない人には届かない (#5625)。名前
                      // 自体は状態で変えない ── 同じ段が選択のたびに別名で呼ばれると
                      // どれを触っているのか分からなくなる。
                      aria-describedby={blocked ? `os-foundation-blocked-${i.toString()}` : undefined}
                      title={blocked ? t('cannotPlaceHere') : undefined}
                      className={
                        blocked
                          ? `flex w-full items-center gap-2 rounded border p-1 text-left ${focusRingWhite} border-ds-error`
                          : selected || dnd.isDragging
                            ? `flex w-full items-center gap-2 rounded border p-1 text-left ${focusRingWhite} border-ds-info`
                            : `flex w-full items-center gap-2 rounded border p-1 text-left ${focusRingWhite} border-white/30`
                      }
                    >
                      <span className="w-5 text-xs text-ds-text-muted">#{i}</span>
                      <div className="relative" style={{ width: cardWidth, height: cardHeight }}>
                        {pile.length > 0 ? (
                          <AnimatedCard card={pile[pile.length - 1]} width={cardWidth} draggable={false} />
                        ) : (
                          <span className="absolute inset-0 flex items-center justify-center text-xs text-ds-text-muted/80">
                            {t('foundation')}
                          </span>
                        )}
                      </div>
                      <span className="text-xs text-ds-text-muted">({pile.length})</span>
                      <span className="text-xs text-ds-text-muted" data-testid={`os-allowed-${i}`}>
                        {i === 0 && <span className="text-ds-warning">★ </span>}
                        {allowed.length === 0
                          ? '—'
                          : i === 0 && pile.length > 0
                            ? t('anyRank')
                            : allowed.map((r) => RANK_LABELS[r]).join(' ')}
                      </span>
                    </button>
                    {blocked && (
                      <span id={`os-foundation-blocked-${i.toString()}`} className="sr-only">
                        {t('cannotPlaceHere')}
                      </span>
                    )}
                  </DropZone>
                );
              })}
            </div>

            {/* Reserve columns */}
            <div className="mb-3 flex gap-2" data-tutorial="os-reserve">
              {state.reserve.map((pile, i) => {
                const top = pile.length > 0 ? pile[pile.length - 1] : null;
                const zone: OsmosisMoveZone = { zone: 'reserve', col: i };
                return (
                  <div key={`r-${i}`} className="flex flex-col items-center gap-1">
                    <span className="text-xs text-ds-text-muted">#{i}</span>
                    {top ? (
                      <button
                        type="button"
                        draggable={isPlaying && !loading}
                        onDragStart={dnd.handleDragStart(zone)}
                        onDragEnd={dnd.handleDragEnd}
                        onClick={() => handleSelectSource(zone)}
                        disabled={!isPlaying || loading}
                        aria-label={`${t('reserve')} ${i}`}
                        aria-pressed={isSelected(zone)}
                        className={
                          isSelected(zone)
                            ? `p-0 border-2 bg-transparent cursor-pointer rounded ${focusRingWhite} border-ds-info ${dnd.isDragSource(zone) ? 'opacity-50' : ''}`
                            : `p-0 border-2 bg-transparent cursor-pointer rounded ${focusRingWhite} border-transparent ${dnd.isDragSource(zone) ? 'opacity-50' : ''}`
                        }
                      >
                        <AnimatedCard card={top} width={cardWidth} draggable={false} />
                      </button>
                    ) : (
                      <div
                        className="rounded border border-dashed border-white/30"
                        style={{ width: cardWidth, height: cardHeight }}
                      />
                    )}
                    <span className="text-xs text-ds-text-muted">({pile.length})</span>
                  </div>
                );
              })}
            </div>

            {/* Stock / Waste */}
            <div className="mb-3 flex gap-3" data-tutorial="os-stock-waste">
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
                      onClick={() => handleSelectSource({ zone: 'waste' })}
                      disabled={!isPlaying || loading}
                      aria-label={t('waste')}
                      aria-pressed={isSelected({ zone: 'waste' })}
                      className={
                        isSelected({ zone: 'waste' })
                          ? `p-0 border-2 bg-transparent cursor-pointer rounded ${focusRingWhite} border-ds-info ${dnd.isDragSource({ zone: 'waste' }) ? 'opacity-50' : ''}`
                          : `p-0 border-2 bg-transparent cursor-pointer rounded ${focusRingWhite} border-transparent ${dnd.isDragSource({ zone: 'waste' }) ? 'opacity-50' : ''}`
                      }
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
            </div>

            <GameMessageBox
              message={selected ? t('selectFoundation') : state.message}
              messageCode={selected ? undefined : state.messageCode}
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
            <div className="flex flex-wrap items-center gap-2" data-tutorial="os-action-buttons">
              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={`${btnPrimary} ${focusRingWhite}`}
                    onClick={handleDraw}
                    disabled={loading}
                    aria-keyshortcuts="d"
                  >
                    {t('draw')}
                    <KbdBadge label={t('kbd.draw')} />
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
                    className={`${btnSuccess} ${focusRingWhite}`}
                    onClick={handleAutoComplete}
                    disabled={loading}
                    aria-keyshortcuts="a"
                  >
                    {t('autoComplete')}
                    <KbdBadge label={t('kbd.autoComplete')} />
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
                    <KbdBadge label={t('kbd.giveUp')} />
                  </button>
                </>
              )}
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="os-reset-button"
                className={focusRingWhite}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
