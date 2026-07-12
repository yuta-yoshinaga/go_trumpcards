import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { type NertzMoveZone, nertzApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnOutline, btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, NertzPlayerData, NertzResponse, NertzTableauCard } from '../types/card';
import { NertzPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { NERTZ_HELP, parseNertzCommand } from '../utils/cli/commands/nertzCommands';
import { formatNertzState } from '../utils/cli/formatters/nertzFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** CPU tick interval in milliseconds while the round is active. */
const NERTZ_TICK_INTERVAL_MS = 700;

/** Duration to leave the collision shake/red-ring on a rejected foundation. */
const NERTZ_COLLISION_FEEDBACK_MS = 500;

/** Tutorial step definitions for the Nertz / Pounce page. */
const NERTZ_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="nertz-foundations"]',
    messageKey: 'tutorial.foundations',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="nertz-pile"]', messageKey: 'tutorial.nertzPile', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="nertz-tableau"]', messageKey: 'tutorial.tableau', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="nertz-stock"]', messageKey: 'tutorial.stock', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="nertz-reset"]', messageKey: 'tutorial.resetButton', placement: 'top', advanceOn: 'next' },
];

type Selection = { kind: 'nertz' } | { kind: 'waste' } | { kind: 'tableau'; col: number; cardIndex: number } | null;

/** Renders the Nertz / Pounce game page. */
export const NertzPage = withTutorial(NertzPageContent, 'nertz', NERTZ_TUTORIAL_STEPS);

function NertzPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('nertz');
  const runApi = useCallback((...args: Parameters<typeof nertzApi.exec>) => nertzApi.exec(...args), []);
  const gameApi = useGameApi<NertzResponse, Parameters<typeof nertzApi.exec>>(runApi);
  const { state, loading, error, retry } = gameApi;
  const apiCall = gameApi.exec;
  const { hint, hintEnabled, setHintEnabled } = useGameHint('nertz', state);
  const cliMode = useCliMode('nertz');
  const cliConfig: CliGameConfig<NertzResponse, Parameters<typeof nertzApi.exec>> = useMemo(
    () => ({
      gameName: 'nertz',
      parseCommand: parseNertzCommand,
      formatResponse: formatNertzState,
      helpText: NERTZ_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiCall, cliConfig, state, {
    addInput: cliMode.addInput,
    addOutput: cliMode.addOutput,
    addError: cliMode.addError,
    clearLog: cliMode.clearLog,
  });

  const [selection, setSelection] = useState<Selection>(null);

  useMountReset(apiCall);

  // CPU tick driver: while the round is active, periodically advance CPUs.
  useEffect(() => {
    if (!state) return;
    if (state.phase !== NertzPhase.PLAYING) return;
    const id = window.setInterval(() => {
      void apiCall('tick');
    }, NERTZ_TICK_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [state, apiCall]);

  const human = state?.players[0];
  const isHumanTurn = state?.phase === NertzPhase.PLAYING;
  const isRoundEnd = state?.phase === NertzPhase.ROUND_END;
  const isGameEnd = state?.phase === NertzPhase.GAME_END;

  const handleManualReset = useCallback(() => {
    setSelection(null);
    void apiCall('reset');
  }, [apiCall]);

  const handleNextRound = useCallback(() => {
    setSelection(null);
    void apiCall('nr');
  }, [apiCall]);

  const handleUndo = useCallback(() => {
    setSelection(null);
    void apiCall('u');
  }, [apiCall]);

  const handleDrawStock = useCallback(() => {
    setSelection(null);
    void apiCall('d', { playerIdx: 0 });
  }, [apiCall]);

  const handleSelectNertz = useCallback(() => {
    if (!isHumanTurn) return;
    setSelection((prev) => (prev?.kind === 'nertz' ? null : { kind: 'nertz' }));
  }, [isHumanTurn]);

  const handleSelectWaste = useCallback(() => {
    if (!isHumanTurn) return;
    setSelection((prev) => (prev?.kind === 'waste' ? null : { kind: 'waste' }));
  }, [isHumanTurn]);

  const handleSelectTableau = useCallback(
    (col: number, cardIndex: number) => {
      if (!isHumanTurn) return;
      setSelection((prev) =>
        prev?.kind === 'tableau' && prev.col === col && prev.cardIndex === cardIndex
          ? null
          : { kind: 'tableau', col, cardIndex },
      );
    },
    [isHumanTurn],
  );

  const [collidedFoundationIdx, setCollidedFoundationIdx] = useState<number | null>(null);
  const [foundationAnnounce, setFoundationAnnounce] = useState('');
  const [collisionTick, setCollisionTick] = useState(0);
  /**
   * Map of foundation index → active placement flash. Multiple CPUs (or a CPU
   * + the human) can all place on different foundations within a single tick;
   * tracking flashes per-foundation lets every placement light up rather than
   * just the first one we detect.
   */
  const [placedFlashes, setPlacedFlashes] = useState<Map<number, { placedBy: 'human' | 'cpu'; key: number }>>(
    () => new Map(),
  );
  const prevFoundationSizesRef = useRef<number[]>([]);
  const flashKeyRef = useRef(0);
  // `isCollisionError` flags that the current `error` from useGameApi was
  // attributed to a foundation collision (already conveyed via the shake
  // animation). The global ErrorAlert is suppressed for the lifetime of that
  // error so it does not pop in once the shake animation expires.
  const [isCollisionError, setIsCollisionError] = useState(false);
  // Tracks the target foundation of the most recent player-initiated move so we
  // can attribute the next error from `useGameApi` to a specific cell.
  const pendingFoundationRef = useRef<number | null>(null);
  const prevErrorRef = useRef<string | null>(null);

  // Successful moves call `setState(res)` on the useGameApi side; whenever a new
  // state arrives we know the in-flight move resolved without error, so the
  // stale foundation pointer must be cleared. Without this, a later unrelated
  // error (a tick that fails) would attribute itself to the previous foundation
  // and flash the wrong cell.
  useEffect(() => {
    if (!state) return;
    // Detect foundation growth → flash success ring on *every* foundation that
    // grew. Nertz ticks can drop multiple cards across multiple foundations in
    // a single state update; the previous single-flash early-break only lit
    // the first one we saw.
    const prev = prevFoundationSizesRef.current;
    const grown: number[] = [];
    for (let idx = 0; idx < state.foundations.length; idx += 1) {
      const newSize = state.foundations[idx].size;
      const oldSize = prev[idx] ?? 0;
      if (newSize > oldSize) grown.push(idx);
    }
    if (grown.length > 0) {
      const humanIdx = pendingFoundationRef.current;
      setPlacedFlashes((current) => {
        const next = new Map(current);
        for (const idx of grown) {
          flashKeyRef.current += 1;
          next.set(idx, { placedBy: humanIdx === idx ? 'human' : 'cpu', key: flashKeyRef.current });
        }
        return next;
      });
      const lastIdx = grown[grown.length - 1];
      setFoundationAnnounce(
        t(humanIdx === lastIdx ? 'foundationAnnounce.human' : 'foundationAnnounce.cpu', { foundation: lastIdx + 1 }),
      );
      // Schedule a removal for each idx independently so the visible flash
      // duration is constant regardless of when sibling flashes start.
      for (const idx of grown) {
        window.setTimeout(() => {
          setPlacedFlashes((current) => {
            if (!current.has(idx)) return current;
            const next = new Map(current);
            next.delete(idx);
            return next;
          });
        }, NERTZ_COLLISION_FEEDBACK_MS);
      }
    }
    prevFoundationSizesRef.current = state.foundations.map((f) => f.size);
    pendingFoundationRef.current = null;
  }, [state, t]);

  useEffect(() => {
    if (error && error !== prevErrorRef.current) {
      if (pendingFoundationRef.current !== null) {
        const collidedIdx = pendingFoundationRef.current;
        setCollidedFoundationIdx(collidedIdx);
        setCollisionTick((n) => n + 1);
        setIsCollisionError(true);
        setFoundationAnnounce(t('foundationAnnounce.collision', { foundation: collidedIdx + 1 }));
        pendingFoundationRef.current = null;
      } else {
        setIsCollisionError(false);
      }
    } else if (!error) {
      setIsCollisionError(false);
    }
    prevErrorRef.current = error;
  }, [error, t]);

  useEffect(() => {
    // `collisionTick` ensures repeated collisions on the same foundation reset the timer.
    if (collidedFoundationIdx === null) return;
    void collisionTick;
    const id = window.setTimeout(() => setCollidedFoundationIdx(null), NERTZ_COLLISION_FEEDBACK_MS);
    return () => window.clearTimeout(id);
  }, [collidedFoundationIdx, collisionTick]);

  const dispatchMove = useCallback(
    (to: NertzMoveZone) => {
      if (!selection) return;
      const from: NertzMoveZone =
        selection.kind === 'tableau'
          ? { zone: 'tableau', col: selection.col, cardIndex: selection.cardIndex }
          : { zone: selection.kind };
      if (to.zone === 'foundation' && to.idx !== undefined) {
        pendingFoundationRef.current = to.idx;
      }
      void apiCall('m', { playerIdx: 0, from, to });
      setSelection(null);
    },
    [apiCall, selection],
  );

  const handleFoundationClick = useCallback(
    (idx: number) => {
      if (!isHumanTurn || !selection) return;
      dispatchMove({ zone: 'foundation', idx });
    },
    [dispatchMove, isHumanTurn, selection],
  );

  const handleTableauTargetClick = useCallback(
    (col: number) => {
      if (!isHumanTurn || !selection) return;
      if (selection.kind === 'tableau' && selection.col === col) {
        setSelection(null);
        return;
      }
      dispatchMove({ zone: 'tableau', col });
    },
    [dispatchMove, isHumanTurn, selection],
  );

  // Keyboard shortcuts for the realtime competitive flow — matches the issue
  // spec (`d` to draw stock, `n`/`w` to pick the Nertz/waste pile, `1-9` to
  // route a held card to a foundation index, `u` to undo).
  const handleFoundationKey = useCallback(
    (idx: number) => {
      if (!state || idx >= state.foundations.length) return;
      if (selection) {
        dispatchMove({ zone: 'foundation', idx });
      }
    },
    [dispatchMove, selection, state],
  );
  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDrawStock },
      { key: 'n', action: handleSelectNertz },
      { key: 'w', action: handleSelectWaste },
      { key: 'u', action: handleUndo },
      ...Array.from({ length: 9 }, (_, i) => ({
        key: String(i + 1),
        action: () => handleFoundationKey(i),
      })),
      { key: 'Escape', action: () => setSelection(null) },
    ],
    [handleDrawStock, handleSelectNertz, handleSelectWaste, handleUndo, handleFoundationKey],
  );
  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: state?.phase === NertzPhase.PLAYING && !loading,
  });

  const phaseName = useMemo(() => {
    if (isGameEnd) return t('phase.gameEnd');
    if (isRoundEnd) return t('phase.roundEnd');
    return t('phase.playing');
  }, [isGameEnd, isRoundEnd, t]);

  if (!state || !human) {
    return (
      <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.nertz.bg}`}>
        <div className="flex-1 flex items-center justify-center text-ds-text-primary">
          <p>{tc('skeleton.loading')}</p>
        </div>
      </div>
    );
  }

  return (
    <GamePageShell
      title={tc('nav.nertz')}
      gameThemeBg={gameTheme.nertz.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/nertz"
      gameEndFlag={isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliMode.cliEnabled} onToggle={cliMode.toggleCli} />}
    >
      {cliMode.cliEnabled ? (
        <CliTerminal logEntries={cliMode.logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto px-3 py-2 space-y-4">
            <div className="sr-only" role="status" aria-live="polite" aria-atomic="true" data-testid="nertz-announce">
              {foundationAnnounce}
            </div>
            <div className="bg-black/30 text-ds-text-primary p-3 rounded text-sm">
              <div className="flex flex-wrap gap-x-4 gap-y-1">
                <span>
                  {t('labels.round')}: {state.roundNumber}
                </span>
                <span>
                  {t('labels.moveCount')}: {state.moveCount}
                </span>
              </div>
              <div className="mt-2 space-y-1">
                {state.players.map((p, i) => {
                  const label = p.isHuman ? t('labels.you') : `${t('labels.cpu')}${i}`;
                  const pct =
                    state.targetScore > 0 ? Math.max(0, Math.min(100, (p.score / state.targetScore) * 100)) : 0;
                  return (
                    <div key={`scorebar-${i}`} className="flex items-center gap-2 text-xs">
                      <span className="w-14 shrink-0">{label}</span>
                      <div
                        className="relative h-3 flex-1 overflow-hidden rounded bg-black/40"
                        role="progressbar"
                        aria-valuemin={0}
                        aria-valuemax={state.targetScore}
                        // WAI-ARIA requires aria-valuenow within [min, max]; clamp it
                        // (the raw score is still shown in the numeric label).
                        aria-valuenow={Math.max(0, Math.min(p.score, state.targetScore))}
                        aria-label={t('labels.scoreBarAria', {
                          player: label,
                          score: p.score,
                          target: state.targetScore,
                        })}
                      >
                        <div
                          data-testid={`nertz-scorebar-${i.toString()}`}
                          className={`h-full ${p.isHuman ? 'bg-ds-success' : 'bg-ds-warning'} motion-safe:transition-[width] motion-safe:duration-300`}
                          style={{ width: `${pct.toString()}%` }}
                        />
                      </div>
                      <span className="w-12 shrink-0 text-right tabular-nums">
                        {p.score} ({p.nertzSize})
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {error && !isCollisionError && <ErrorAlert message={error} onRetry={retry} />}

            <div data-tutorial="nertz-foundations" className="bg-black/30 text-ds-text-primary p-3 rounded">
              <div className="text-xs uppercase tracking-wide text-ds-text-muted mb-2">{t('labels.foundation')}</div>
              <div className="flex flex-wrap gap-2">
                {state.foundations.map((f, idx) => {
                  const flash = placedFlashes.get(idx);
                  return (
                    <FoundationCell
                      key={`f-${idx}`}
                      idx={idx}
                      top={f.top}
                      size={f.size}
                      onClick={() => handleFoundationClick(idx)}
                      disabled={!isHumanTurn || !selection}
                      ariaLabel={t('labels.foundationN', { n: idx, defaultValue: `Foundation ${idx}` })}
                      collided={collidedFoundationIdx === idx}
                      placedBy={flash?.placedBy ?? null}
                      placedFlashKey={flash?.key ?? 0}
                    />
                  );
                })}
              </div>
            </div>

            {state.players.length > 1 && (
              <div className="bg-black/30 text-ds-text-primary p-3 rounded text-sm space-y-1">
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={`cpu-${p.deckIdx}`} className="flex justify-between">
                      <span>
                        {t('labels.cpu')}
                        {p.deckIdx} — {p.name}
                      </span>
                      <span>
                        {t('labels.nertz')}: {p.nertzSize} / {t('labels.score')}: {p.score}
                      </span>
                    </div>
                  ))}
              </div>
            )}

            <div className="bg-black/30 text-ds-text-primary p-3 rounded space-y-2">
              <div className="flex flex-wrap gap-3 items-start">
                <div data-tutorial="nertz-pile" className="space-y-1">
                  <div className="text-xs uppercase tracking-wide text-ds-text-muted">{t('labels.nertz')}</div>
                  <CardButton
                    card={human.nertzTop ?? null}
                    label={`${human.nertzSize}`}
                    selected={selection?.kind === 'nertz'}
                    disabled={!isHumanTurn || !human.nertzTop}
                    onClick={handleSelectNertz}
                  />
                </div>

                <div className="space-y-1">
                  <div className="text-xs uppercase tracking-wide text-ds-text-muted">{t('labels.waste')}</div>
                  <CardButton
                    card={human.wasteTop ?? null}
                    label={`${human.wasteSize}`}
                    selected={selection?.kind === 'waste'}
                    disabled={!isHumanTurn || !human.wasteTop}
                    onClick={handleSelectWaste}
                  />
                </div>

                <div data-tutorial="nertz-stock" className="space-y-1">
                  <div className="text-xs uppercase tracking-wide text-ds-text-muted">{t('labels.stock')}</div>
                  <button
                    type="button"
                    onClick={handleDrawStock}
                    disabled={!isHumanTurn || loading}
                    className={`${btnSecondary} min-w-[3rem]`}
                  >
                    {human.stockSize}
                  </button>
                </div>
              </div>

              <div data-tutorial="nertz-tableau" className="space-y-1">
                <div className="text-xs uppercase tracking-wide text-ds-text-muted">{t('labels.tableau')}</div>
                <div className="grid grid-cols-4 gap-2">
                  {human.tableau.map((col, colIdx) => (
                    <TableauColumn
                      key={`tab-${colIdx}`}
                      col={col}
                      colIdx={colIdx}
                      selection={selection}
                      onSelectCard={handleSelectTableau}
                      onTarget={() => handleTableauTargetClick(colIdx)}
                      disabled={!isHumanTurn}
                    />
                  ))}
                </div>
              </div>
            </div>

            <ActionLogSection
              isEndPhase={isGameEnd || isRoundEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

          <GameFooter className={`${gameTheme.nertz.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap gap-2 items-center">
              <label className="flex items-center gap-1 text-ds-text-primary text-xs">
                <input
                  type="checkbox"
                  checked={hintEnabled}
                  onChange={(e) => setHintEnabled(e.target.checked)}
                  aria-label={tc('hint.toggle', { ns: 'tutorial' })}
                />
                {tc('hint.toggle', { ns: 'tutorial' })}
              </label>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="nertz-reset"
              />
              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnPrimary} onClick={handleNextRound} disabled={loading}>
                  {t('actions.nextRound')}
                </button>
              )}
              {isHumanTurn && state.canUndo && (
                <button type="button" className={btnOutline} onClick={handleUndo} disabled={loading}>
                  {t('actions.undo')}
                </button>
              )}
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

interface FoundationCellProps {
  idx: number;
  top?: Card;
  size: number;
  onClick: () => void;
  disabled: boolean;
  ariaLabel: string;
  /** Apply collision feedback (shake + red ring) when a move to this foundation was just rejected. */
  collided?: boolean;
  /** Which side just placed a card on this foundation, for a brief success flash. */
  placedBy?: 'human' | 'cpu' | null;
  /** Re-applies the placed flash even when `placedBy` stays the same (e.g., two human placements in a row). */
  placedFlashKey?: number;
}

function FoundationCell({
  idx,
  top,
  size,
  onClick,
  disabled,
  ariaLabel,
  collided = false,
  placedBy = null,
  placedFlashKey = 0,
}: FoundationCellProps) {
  const { cardWidth } = useCardDimensions();
  const w = Math.max(44, Math.round(cardWidth * 0.6));
  const collisionCls = collided ? 'animate-shake ring-2 ring-ds-error' : '';
  // The placement flash is a sibling overlay so we can remount *it* (via key)
  // to re-fire animate-pulse-once on back-to-back placements without
  // unmounting the underlying <button> and breaking keyboard focus.
  const flashOverlayClass =
    placedBy === 'human'
      ? 'pointer-events-none absolute inset-0 rounded ring-2 ring-ds-success motion-safe:animate-pulse-once'
      : placedBy === 'cpu'
        ? 'pointer-events-none absolute inset-0 rounded ring-2 ring-ds-info motion-safe:animate-pulse-once'
        : null;
  return (
    <div className="relative inline-block">
      <button
        type="button"
        onClick={onClick}
        disabled={disabled}
        className={`flex flex-col items-center rounded p-0.5 text-xs text-ds-text-muted disabled:opacity-50 ${collisionCls}`}
        aria-label={ariaLabel}
        data-testid={`nertz-foundation-${idx}`}
        data-collided={collided || undefined}
        data-placed-by={placedBy ?? undefined}
      >
        <span className="block leading-none">F{idx}</span>
        {top ? (
          <AnimatedCard card={top} width={w} />
        ) : (
          <span
            className="block rounded border border-dashed border-white/30"
            style={{ width: w, height: Math.round(w * 1.4) }}
          />
        )}
        <span className="block">({size})</span>
      </button>
      {flashOverlayClass && (
        // Remount the overlay (via key) on each new placement so the
        // animate-pulse-once keyframe re-fires for back-to-back placements.
        <span
          aria-hidden="true"
          key={`f-flash-${idx}-${placedFlashKey}`}
          className={flashOverlayClass}
          data-testid={`nertz-foundation-flash-${idx}`}
        />
      )}
    </div>
  );
}

interface CardButtonProps {
  card: Card | null;
  label: string;
  selected: boolean;
  disabled: boolean;
  onClick: () => void;
}

function CardButton({ card, label, selected, disabled, onClick }: CardButtonProps) {
  const { cardWidth } = useCardDimensions();
  const w = Math.max(44, Math.round(cardWidth * 0.6));
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className="flex flex-col items-center rounded p-0.5 text-xs text-ds-text-muted disabled:opacity-50"
    >
      {card ? (
        <AnimatedCard card={card} width={w} isSelected={selected} />
      ) : (
        <span
          className="block rounded border border-dashed border-white/30"
          style={{ width: w, height: Math.round(w * 1.4) }}
        />
      )}
      <span className="block">{label}</span>
    </button>
  );
}

interface TableauColumnProps {
  col: NertzTableauCard[];
  colIdx: number;
  selection: Selection;
  onSelectCard: (col: number, cardIndex: number) => void;
  onTarget: () => void;
  disabled: boolean;
}

function TableauColumn({ col, colIdx, selection, onSelectCard, onTarget, disabled }: TableauColumnProps) {
  const { cardWidth } = useCardDimensions();
  const w = Math.max(44, Math.round(cardWidth * 0.6));
  if (col.length === 0) {
    return (
      <button
        type="button"
        onClick={onTarget}
        disabled={disabled}
        className="min-h-[4rem] rounded border border-dashed border-white/30 text-ds-text-muted text-xs"
        style={{ width: w }}
      >
        —
      </button>
    );
  }
  return (
    <div className="flex flex-col gap-1">
      {col.map((tc, i) => {
        const isSelected = selection?.kind === 'tableau' && selection.col === colIdx && selection.cardIndex === i;
        const isLast = i === col.length - 1;
        // Dual-purpose click: the bottom card acts as a drop target when a
        // source is already selected (saves the user a second tap on an
        // empty drop zone), otherwise tapping any card selects it as the
        // move source. PR #1528 review noted this is non-obvious.
        return (
          <button
            key={`t-${colIdx}-${i}`}
            type="button"
            onClick={() => (isLast && selection && !isSelected ? onTarget() : onSelectCard(colIdx, i))}
            disabled={disabled || !tc.card}
            className="rounded disabled:opacity-50"
          >
            {tc.card ? (
              <AnimatedCard card={tc.card} width={w} isSelected={isSelected} />
            ) : (
              <span
                className="block rounded border border-dashed border-white/30"
                style={{ width: w, height: Math.round(w * 1.4) }}
              />
            )}
          </button>
        );
      })}
    </div>
  );
}

export type _NertzPagePlayerSnapshot = NertzPlayerData;
