import { useCallback, useMemo, useRef, useState } from 'react';
import { type AccordionMoveZone, accordionApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { KeyboardShortcutsPanel } from '../components/KeyboardShortcutsPanel';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
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
import { badgeErrorColors, badgeSuccessColors } from '../styles/badgeStyles';
import { btnDanger, btnOutline, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { AccordionResponse } from '../types/card';
import { AccordionPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { accordionLegalOffsets, accordionLegalTargets, accordionNextAutoMove } from '../utils/accordionUtils';
import { cardAlt } from '../utils/cardAlt';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Upper bound on autocomplete merges (a 52-card deck needs at most 51) — loop guard (#3192). */
const AC_MAX_AUTO_MERGES = 52;

/** Accordion tutorial step definitions. */
const AC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ac-piles"]',
    messageKey: 'tutorial.piles',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ac-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ac-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type ApiArgs = Parameters<typeof accordionApi.exec>;

/** Parses an Accordion CLI command. */
function parseAccordionCommand(input: string): CliParseResult<ApiArgs> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase();
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'l':
    case 'log':
      return { args: ['log'] };
    case 'm':
    case 'move': {
      if (parts.length !== 3) return { error: 'Usage: m <fromIdx> <toIdx>' };
      const from = Number.parseInt(parts[1], 10);
      const to = Number.parseInt(parts[2], 10);
      if (Number.isNaN(from) || Number.isNaN(to)) return { error: 'Invalid index' };
      return { args: ['move', { zone: 'pile', index: from }, { zone: 'pile', index: to }] };
    }
    default:
      return { error: `Unknown command: ${cmd}` };
  }
}

/** Formats Accordion state for CLI display. */
function formatAccordionState(state: AccordionResponse): string {
  const lines: string[] = [];
  lines.push(`Piles: ${state.pileCount}  Moves: ${state.moveCount}`);
  const tops = state.piles
    .map((p, i) => {
      const top = p.cards[0];
      const label = top ? `${top.design[0]}${top.value}` : '??';
      return `[${i}]${label}${p.size > 1 ? `(+${p.size - 1})` : ''}`;
    })
    .join(' ');
  lines.push(tops);
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}

/** Renders the Accordion solitaire game page. */
export const AccordionPage = withTutorial(AccordionPageContent, 'accordion', AC_TUTORIAL_STEPS);
/** Inner content of the Accordion page. */
function AccordionPageContent() {
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
  } = useGamePageSetup('accordion');
  const {
    state,
    setState,
    loading,
    error,
    exec: apiCall,
    retry,
  } = useGameApi<AccordionResponse, ApiArgs>((...args) => accordionApi.exec(...args));

  useMountReset(apiCall);

  const [selectedIdx, setSelectedIdx] = useState<number | null>(null);
  // Autocomplete drives the single `move` action in a loop (issue #3192), which
  // bypasses useGameApi's `loading` flag, so `isAutoCompleting` gates the UI
  // while the batch runs and `stateRef`/`autoCompletingRef` let the loop read
  // the freshest board and block a second concurrent run (mirrors AcesUp #3347).
  const [isAutoCompleting, setIsAutoCompleting] = useState(false);
  const stateRef = useRef<AccordionResponse | null>(state);
  stateRef.current = state;
  const autoCompletingRef = useRef(false);
  // Tracks the pile under cursor/focus so we can paint legal -1/-3 targets
  // (same suit OR same rank) without waiting for click. Reset by mouseleave/blur
  // and on every state change (handled implicitly because piles re-key on size).
  const [hoveredIdx, setHoveredIdx] = useState<number | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('accordion', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('accordion');
  const accordionCliConfig: CliGameConfig<AccordionResponse, ApiArgs> = useMemo(
    () => ({
      gameName: 'accordion',
      parseCommand: parseAccordionCommand,
      formatResponse: formatAccordionState,
      helpText: [
        'm <from> <to>  Merge pile `from` onto `to` (to must be from-1 or from-3)',
        'g              Give up',
        'h              Hint',
        'u              Undo',
        'l              Action log',
        'r              Reset',
      ],
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiCall, accordionCliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();

  const handleManualReset = useCallback(() => {
    void apiCall('reset');
    setSelectedIdx(null);
  }, [apiCall]);

  const handleGiveUp = useCallback(() => {
    void apiCall('giveup');
  }, [apiCall]);

  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  const confirmGiveUpAction = useGiveUpConfirm(handleGiveUp, requestGiveUpConfirm);

  const handleHint = useCallback(() => {
    void apiCall('hint');
  }, [apiCall]);

  const handleUndo = useCallback(() => {
    void apiCall('undo');
    setSelectedIdx(null);
  }, [apiCall]);

  const handleUndoEscape = useCallback(
    (n: number) => {
      void apiCall('undo_n', undefined, undefined, n);
      setSelectedIdx(null);
    },
    [apiCall],
  );

  /**
   * Auto-plays every forced/recommended merge in one click (issue #3192). It
   * drives the existing single `move` action sequentially, re-reading the fresh
   * board after each merge and applying the next {@link accordionNextAutoMove}
   * (the same offset-3-first heuristic the backend hint uses) until no legal
   * merge remains. Terminates because each merge removes a pile; the re-entry
   * guard blocks a second concurrent run.
   */
  const handleAutoComplete = useCallback(async () => {
    if (autoCompletingRef.current) return;
    autoCompletingRef.current = true;
    setIsAutoCompleting(true);
    setSelectedIdx(null);
    let move = stateRef.current ? accordionNextAutoMove(stateRef.current.piles) : null;
    try {
      // Upper bound: a 52-pile deck can never need more than 51 merges.
      for (let step = 0; step < AC_MAX_AUTO_MERGES && move; step++) {
        const res = await accordionApi.exec(
          'move',
          { zone: 'pile', index: move.fromIdx },
          { zone: 'pile', index: move.toIdx },
        );
        setState(res);
        if (res.phase !== AccordionPhase.PLAYING) break;
        move = accordionNextAutoMove(res.piles);
      }
    } catch {
      // Surface the failure through the shared exec so the standard
      // error/retry channel handles it.
      if (move) {
        await apiCall('move', { zone: 'pile', index: move.fromIdx }, { zone: 'pile', index: move.toIdx });
      }
    } finally {
      autoCompletingRef.current = false;
      setIsAutoCompleting(false);
    }
  }, [apiCall, setState]);

  const dispatchMove = useCallback(
    (fromIdx: number, toIdx: number) => {
      const from: AccordionMoveZone = { zone: 'pile', index: fromIdx };
      const to: AccordionMoveZone = { zone: 'pile', index: toIdx };
      void apiCall('move', from, to);
      setSelectedIdx(null);
    },
    [apiCall],
  );

  const handlePileClick = useCallback(
    (idx: number) => {
      if (selectedIdx === null) {
        setSelectedIdx(idx);
        return;
      }
      if (selectedIdx === idx) {
        setSelectedIdx(null);
        return;
      }
      const offset = selectedIdx - idx;
      if (offset === 1 || offset === 3) {
        dispatchMove(selectedIdx, idx);
      } else {
        setSelectedIdx(idx);
      }
    },
    [selectedIdx, dispatchMove],
  );

  // Keyboard shortcuts: arrow keys scrub the selection, `1`/`3` perform the
  // two legal merges, `u`/`h`/`g` mirror the action buttons. Hook reads from
  // the live state so it stays in sync without a separate effect.
  const pileCount = state?.pileCount ?? 0;
  const moveSelection = useCallback(
    (delta: number) => {
      setSelectedIdx((prev) => {
        if (prev === null) return delta > 0 ? 0 : Math.max(0, pileCount - 1);
        const next = prev + delta;
        if (next < 0 || next >= pileCount) return prev;
        return next;
      });
    },
    [pileCount],
  );
  const mergeFromSelection = useCallback(
    (offset: 1 | 3) => {
      if (selectedIdx === null) return;
      const target = selectedIdx - offset;
      if (target < 0) return;
      dispatchMove(selectedIdx, target);
    },
    [dispatchMove, selectedIdx],
  );
  const accordionBindings = useMemo(
    () => [
      { key: 'ArrowLeft', action: () => moveSelection(-1) },
      { key: 'ArrowRight', action: () => moveSelection(1) },
      { key: '1', action: () => mergeFromSelection(1) },
      { key: '3', action: () => mergeFromSelection(3) },
      { key: 'a', action: () => void handleAutoComplete() },
      { key: 'u', action: handleUndo },
      { key: 'h', action: handleHint },
      { key: 'g', action: confirmGiveUpAction },
      { key: 'Escape', action: () => setSelectedIdx(null) },
    ],
    [moveSelection, mergeFromSelection, handleAutoComplete, handleUndo, handleHint, confirmGiveUpAction],
  );
  // The autocomplete loop bypasses useGameApi's `loading` flag, so gate every
  // interactive control on `busy` (loading OR a batch in flight) — issue #3192.
  const busy = loading || isAutoCompleting;
  useActionKeyboardNav({
    bindings: accordionBindings,
    enabled: state?.phase === AccordionPhase.PLAYING && !busy,
  });

  if (error) return <ErrorAlert message={error} onRetry={retry} />;
  if (!state) return <GameSkeleton gameKey="accordion" layout={{ kind: 'tableau', topRow: 6, tableau: 7 }} />;

  const isPlaying = state.phase === AccordionPhase.PLAYING;
  const isGameClear = state.phase === AccordionPhase.GAME_CLEAR;
  const isGameOver = state.phase === AccordionPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;
  // Autocomplete is only useful while a legal merge remains; disable it (and skip
  // the pulse cue) otherwise, mirroring the sibling solitaires' gating.
  const hasAutoMove = isPlaying && accordionNextAutoMove(state.piles) !== null;

  const phaseName = isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing');

  return (
    <GamePageShell
      title={tc('nav.accordion')}
      gameThemeBg={gameTheme.accordion.bg}
      phaseName={phaseName}
      gamePath="/accordion"
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
          <span>
            {t('piles')}: {state.pileCount}
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
            <div className="flex flex-wrap gap-1 sm:gap-2 justify-center" data-tutorial="ac-piles">
              {(() => {
                const hoverTargets =
                  isPlaying && hoveredIdx !== null ? new Set(accordionLegalTargets(state.piles, hoveredIdx)) : null;
                // Touch devices never fire mouseenter, so also paint the selected
                // pile's legal -1/-3 targets persistently once it is picked (#3190).
                const selectedTargets =
                  isPlaying && selectedIdx !== null ? new Set(accordionLegalTargets(state.piles, selectedIdx)) : null;
                return state.piles.map((pile, idx) => {
                  const top = pile.cards[0];
                  const isSelected = selectedIdx === idx;
                  const hintFrom = state.hint?.fromIdx === idx;
                  const hintTo = state.hint?.toIdx === idx;
                  const isHoverTarget = hoverTargets?.has(idx) ?? false;
                  const isSelectedTarget = selectedTargets?.has(idx) ?? false;
                  // Highlight legal targets whether reached by hover (mouse) or by
                  // selecting a source pile (touch/keyboard) so all inputs get parity.
                  const isLegalTarget = isHoverTarget || isSelectedTarget;
                  // When a pile is selected, spell out its legal merge offsets (1/3)
                  // in the aria-label so screen-reader users learn where it can go (#2596).
                  const mergeSuffix =
                    isSelected && top
                      ? (() => {
                          const offsets = accordionLegalOffsets(state.piles, idx);
                          return offsets.length > 0
                            ? ` — ${offsets.map((o) => t('mergeOffsetAvailable', { offset: o })).join(t('listSeparator'))}`
                            : ` — ${t('noMergeAvailable')}`;
                        })()
                      : '';
                  const baseLabel = top ? `${idx}: ${cardAlt(top)}` : `${idx}: empty`;
                  // Keying by the top card's identity lets React preserve
                  // per-pile state (ring/selection class transitions, AnimatedCard
                  // instances) when piles shift left after a merge.
                  const pileKey = top ? `${top.design}-${top.value}-${pile.size}` : `empty-${idx}`;
                  return (
                    <button
                      key={pileKey}
                      type="button"
                      className={`relative ${focusRingWhite} rounded-lg transition-transform ${
                        isSelected ? 'ring-2 ring-ds-warning -translate-y-1' : ''
                      } ${hintFrom ? 'ring-2 ring-ds-info animate-pulse' : ''} ${
                        hintTo ? 'ring-2 ring-ds-success animate-pulse' : ''
                      } ${isLegalTarget && !isSelected && !hintTo && !hintFrom ? 'ring-2 ring-ds-success' : ''}`}
                      onClick={() => handlePileClick(idx)}
                      onMouseEnter={isPlaying ? () => setHoveredIdx(idx) : undefined}
                      onMouseLeave={() => setHoveredIdx((cur) => (cur === idx ? null : cur))}
                      onFocus={isPlaying ? () => setHoveredIdx(idx) : undefined}
                      onBlur={() => setHoveredIdx((cur) => (cur === idx ? null : cur))}
                      disabled={!isPlaying || busy}
                      data-hover-target={isHoverTarget ? 'true' : 'false'}
                      data-legal-target={isLegalTarget ? 'true' : 'false'}
                      aria-label={`${baseLabel}${mergeSuffix}`}
                    >
                      {top && <AnimatedCard card={top} width={cardWidth} />}
                      <span className="absolute top-0 left-0 text-[10px] bg-black/40 text-ds-text-primary rounded-br px-1">
                        {idx}
                      </span>
                      {pile.size > 1 && (
                        <span className="absolute bottom-0 right-0 text-[10px] bg-black/60 text-ds-text-primary rounded-tl px-1">
                          +{pile.size - 1}
                        </span>
                      )}
                    </button>
                  );
                });
              })()}
            </div>
          </div>

          {/* SR-only live region announcing the selected pile's available merges (#2596). */}
          <div className="sr-only" role="status" aria-live="polite" data-testid="ac-selection-status">
            {isPlaying && selectedIdx !== null
              ? (() => {
                  const count = accordionLegalOffsets(state.piles, selectedIdx).length;
                  return count > 0
                    ? t('selectionMoves', { idx: selectedIdx, count })
                    : t('selectionNoMoves', { idx: selectedIdx });
                })()
              : ''}
          </div>

          <div data-tutorial="ac-controls">
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {isEnded && (
              <div
                data-testid="result-banner"
                role="status"
                aria-live="polite"
                className={`text-sm text-center font-medium rounded px-3 py-1.5 mt-1 ${
                  isGameClear ? badgeSuccessColors : badgeErrorColors
                }`}
              >
                {isGameClear
                  ? t('result.clear', { moveCount: state.moveCount })
                  : t('result.gameOver', { pileCount: state.pileCount, moveCount: state.moveCount })}
              </div>
            )}

            {state.hint && isRequestedHint(state) && (
              <div
                className="text-sm text-ds-accent bg-ds-surface/90 border border-ds-accent rounded px-3 py-1.5 mt-1"
                role="status"
                aria-live="polite"
              >
                {t('hintMove', { from: state.hint.fromIdx, to: state.hint.toIdx })}
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
                dataTutorial="ac-reset-button"
              />

              {isPlaying && (
                <>
                  <button
                    type="button"
                    className={`${btnSuccess}${hasAutoMove && !busy ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
                    onClick={() => void handleAutoComplete()}
                    disabled={busy || !hasAutoMove}
                    aria-keyshortcuts="a"
                    data-testid="ac-autocomplete"
                  >
                    {t('autoComplete')}
                  </button>
                  <button
                    type="button"
                    className={btnOutline}
                    onClick={handleHint}
                    disabled={busy}
                    aria-keyshortcuts="h"
                  >
                    {t('hint')}
                  </button>
                  <button
                    type="button"
                    className={btnOutline}
                    onClick={handleUndo}
                    disabled={busy || !state.canUndo}
                    aria-keyshortcuts="u"
                  >
                    {t('undo')}
                  </button>
                  <button
                    type="button"
                    className={btnDanger}
                    onClick={confirmGiveUpAction}
                    disabled={busy}
                    aria-keyshortcuts="g"
                  >
                    {t('giveup')}
                  </button>
                  {state.isStalemate && (
                    <StalemateEscapeButton
                      undoToEscape={state.undoToEscape ?? -1}
                      onEscape={handleUndoEscape}
                      disabled={busy}
                    />
                  )}
                </>
              )}
            </GameFooter>
            {isPlaying && (
              <KeyboardShortcutsPanel
                title={t('kbd.title')}
                data-testid="ac-kbd-shortcuts"
                shortcuts={[
                  { keys: ['←', '→'], description: t('kbd.move') },
                  { keys: ['1', '3'], description: t('kbd.merge') },
                  { keys: ['a'], description: t('kbd.autoComplete') },
                  { keys: ['h'], description: t('kbd.hint') },
                  { keys: ['u'], description: t('kbd.undo') },
                  { keys: ['g'], description: t('kbd.giveup') },
                  { keys: ['Esc'], description: t('kbd.clear') },
                ]}
              />
            )}
          </div>
        </>
      )}
    </GamePageShell>
  );
}
