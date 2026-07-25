import { useCallback, useEffect, useMemo, useState } from 'react';
import { spiteAndMaliceApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, SpiteAndMaliceResponse } from '../types/card';
import { SpiteAndMalicePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import { isGoalTopPlayableToFoundation } from '../utils/spiteAndMaliceUtils';

const samRunner = spiteAndMaliceApi;

/** CPU cadence presets — a shorter delay makes the CPU take its turn sooner. */
type SamCpuSpeed = 'slow' | 'normal' | 'fast';

/**
 * CPU turn wait (ms) per speed preset. A smaller delay advances the CPU turn
 * sooner; `normal` (500ms) preserves the historical default for backward
 * compatibility.
 */
const SAM_CPU_DELAY_MS: Record<SamCpuSpeed, number> = {
  slow: 900,
  normal: 500,
  fast: 200,
};

const SAM_CPU_SPEED_STORAGE_KEY = 'spiteandmalice:cpuSpeed';

/** Read the persisted CPU speed, falling back to `normal` when unset/invalid. */
function loadSamCpuSpeed(): SamCpuSpeed {
  try {
    const v = localStorage.getItem(SAM_CPU_SPEED_STORAGE_KEY);
    if (v === 'slow' || v === 'normal' || v === 'fast') return v;
  } catch {
    // localStorage may be unavailable (private mode / SSR); fall through to default.
  }
  return 'normal';
}

const SAM_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sam-opponent"]', messageKey: 'tutorial.opponent', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="sam-foundations"]',
    messageKey: 'tutorial.foundations',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sam-goal"]', messageKey: 'tutorial.goal', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sam-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sam-sides"]', messageKey: 'tutorial.sides', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sam-reset"]', messageKey: 'tutorial.resetButton', placement: 'top', advanceOn: 'next' },
];

type ApiArgs = Parameters<typeof samRunner.exec>;

function parseSamCommand(input: string): CliParseResult<ApiArgs> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase();
  const intArg = (s: string | undefined): number | null => {
    if (s === undefined) return null;
    const n = Number.parseInt(s, 10);
    return Number.isFinite(n) ? n : null;
  };
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'cpu':
      return { args: ['cpu'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'l':
    case 'log':
      return { args: ['log'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'ph': {
      const h = intArg(parts[1]);
      const f = intArg(parts[2]);
      if (h === null || f === null) return { error: 'Usage: ph <handIdx> <foundationIdx>' };
      return { args: ['move', { zone: 'hand', idx: h }, { zone: 'foundation', idx: f }] };
    }
    case 'pg': {
      const f = intArg(parts[1]);
      if (f === null) return { error: 'Usage: pg <foundationIdx>' };
      return { args: ['move', { zone: 'goal' }, { zone: 'foundation', idx: f }] };
    }
    case 'ps': {
      const s = intArg(parts[1]);
      const f = intArg(parts[2]);
      if (s === null || f === null) return { error: 'Usage: ps <sideIdx> <foundationIdx>' };
      return { args: ['move', { zone: 'side', idx: s }, { zone: 'foundation', idx: f }] };
    }
    case 'd':
    case 'discard': {
      const h = intArg(parts[1]);
      const s = intArg(parts[2]);
      if (h === null || s === null) return { error: 'Usage: d <handIdx> <sideIdx>' };
      return { args: ['discard', { zone: 'hand', idx: h }, { zone: 'side', idx: s }] };
    }
    default:
      return { error: `Unknown command: ${cmd}` };
  }
}

function formatSamState(state: SpiteAndMaliceResponse): string {
  const lines: string[] = [];
  lines.push(`Phase: ${state.phase}  Turn: ${state.current === 0 ? 'YOU' : 'CPU'}  Moves: ${state.moveCount}`);
  lines.push(`Foundations tops: ${state.foundationTops.map((v, i) => `F${i}=${v === 0 ? '-' : v}`).join('  ')}`);
  lines.push(`Stock: ${state.stockSize}  Completed: ${state.completedSize}`);
  for (let i = 0; i < state.players.length; i++) {
    const p = state.players[i];
    const goalTop = p.goalTop ? `${p.goalTop.design[0]}${p.goalTop.value}` : '-';
    lines.push(`${p.isCpu ? 'CPU' : 'YOU'}: goal[${goalTop} x${p.goalSize}]  hand=${p.hand.length}`);
  }
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}

/** Spite & Malice (Cat and Mouse) page — a 2-player race game vs CPU. */
export const SpiteAndMalicePage = withTutorial(SpiteAndMalicePageContent, 'spiteandmalice', SAM_TUTORIAL_STEPS);
type Selection = { kind: 'hand'; idx: number } | { kind: 'goal' } | { kind: 'side'; idx: number } | null;

function SpiteAndMalicePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('spiteandmalice');
  const { playSound } = useSound();
  const runApi = useCallback((...args: ApiArgs) => samRunner.exec(...args), []);
  const gameApi = useGameApi<SpiteAndMaliceResponse, ApiArgs>(runApi);
  const { state, loading, error, retry } = gameApi;
  const apiCall = gameApi.exec;

  const [selection, setSelection] = useState<Selection>(null);
  const [cpuSpeed, setCpuSpeed] = useState<SamCpuSpeed>(loadSamCpuSpeed);

  const handleSelectCpuSpeed = useCallback((v: string) => {
    const speed: SamCpuSpeed = v === 'slow' || v === 'fast' ? v : 'normal';
    setCpuSpeed(speed);
    try {
      localStorage.setItem(SAM_CPU_SPEED_STORAGE_KEY, speed);
    } catch {
      // Persistence is best-effort; ignore storage failures.
    }
  }, []);

  useMountReset(apiCall);

  // CPU turn driver: after a short delay, advance the CPU's turn. The delay is
  // scaled by the chosen speed preset; changing the speed restarts the timer so
  // the new cadence takes effect from the next CPU turn.
  useEffect(() => {
    if (!state) return;
    if (state.phase === SpiteAndMalicePhase.GAME_OVER) return;
    if (state.current !== 1) return;
    const timer = setTimeout(() => {
      void apiCall('cpu');
    }, SAM_CPU_DELAY_MS[cpuSpeed]);
    return () => clearTimeout(timer);
  }, [state, apiCall, cpuSpeed]);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('spiteandmalice');
  const samCliConfig: CliGameConfig<SpiteAndMaliceResponse, ApiArgs> = useMemo(
    () => ({
      gameName: 'spiteandmalice',
      parseCommand: parseSamCommand,
      formatResponse: formatSamState,
      helpText: [
        'ph <h> <f>     Play hand[h] to foundation f',
        'pg <f>         Play goal pile top to foundation f',
        'ps <s> <f>     Play side[s] top to foundation f',
        'd <h> <s>      Discard hand[h] to side[s] (ends turn)',
        'cpu            Advance CPU turn',
        'ac             Auto-complete (play every legal foundation move)',
        'h              Hint',
        'r              Reset',
        'l              Action log',
      ],
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiCall, samCliConfig, state, { addInput, addOutput, addError, clearLog });

  const { hint: currentHint, hintEnabled, setHintEnabled } = useGameHint('spiteandmalice', state);
  const { cardWidth } = useCardDimensions();

  const handleResetAction = useCallback(() => {
    setSelection(null);
    void apiCall('reset');
    playSound('shuffle');
  }, [apiCall, playSound]);

  const handleAutoComplete = useCallback(() => {
    setSelection(null);
    void apiCall('autocomplete');
  }, [apiCall]);

  const isHumanTurn = state ? state.current === 0 : false;
  const isGameOver = state ? state.phase === SpiteAndMalicePhase.GAME_OVER : false;

  const handleSelectHand = useCallback(
    (idx: number) => {
      if (!isHumanTurn || isGameOver) return;
      setSelection((prev) => (prev?.kind === 'hand' && prev.idx === idx ? null : { kind: 'hand', idx }));
    },
    [isHumanTurn, isGameOver],
  );

  const handleSelectGoal = useCallback(() => {
    if (!isHumanTurn || isGameOver) return;
    setSelection((prev) => (prev?.kind === 'goal' ? null : { kind: 'goal' }));
  }, [isHumanTurn, isGameOver]);

  const handleSelectSide = useCallback(
    (idx: number) => {
      if (!isHumanTurn || isGameOver) return;
      setSelection((prev) => (prev?.kind === 'side' && prev.idx === idx ? null : { kind: 'side', idx }));
    },
    [isHumanTurn, isGameOver],
  );

  const handleFoundationClick = useCallback(
    (fIdx: number) => {
      if (!isHumanTurn || isGameOver || !selection) return;
      if (selection.kind === 'hand') {
        void apiCall('move', { zone: 'hand', idx: selection.idx }, { zone: 'foundation', idx: fIdx });
      } else if (selection.kind === 'goal') {
        void apiCall('move', { zone: 'goal' }, { zone: 'foundation', idx: fIdx });
      } else {
        void apiCall('move', { zone: 'side', idx: selection.idx }, { zone: 'foundation', idx: fIdx });
      }
      setSelection(null);
      playSound('cardPlace');
    },
    [apiCall, selection, isHumanTurn, isGameOver, playSound],
  );

  const handleDiscardSide = useCallback(
    (sideIdx: number) => {
      if (!isHumanTurn || isGameOver) return;
      if (selection?.kind !== 'hand') return;
      void apiCall('discard', { zone: 'hand', idx: selection.idx }, { zone: 'side', idx: sideIdx });
      setSelection(null);
      playSound('cardPlace');
    },
    [apiCall, selection, isHumanTurn, isGameOver, playSound],
  );
  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state) return <GameSkeleton gameKey="spiteandmalice" layout={{ kind: 'tableau', topRow: 6, tableau: 7 }} />;

  const phaseName = isGameOver ? t('phase.gameOver') : state.current === 1 ? t('phase.cpuTurn') : t('phase.playerTurn');

  const opponent = state.players[1];
  const human = state.players[0];

  const isHintTarget = (action: string) => hintEnabled && currentHint?.targetAction === action;
  const selectionIsHand = selection?.kind === 'hand';

  return (
    <GamePageShell
      title={tc('nav.spiteandmalice')}
      gameThemeBg={gameTheme.spiteandmalice.bg}
      phaseName={phaseName}
      gamePath="/spiteandmalice"
      gameEndFlag={isGameOver}
      winShow={isGameOver && state.winner === 0}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
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
          <SettingsPanel
            title={tc('settings.title', { ns: 'common' })}
            groups={[
              {
                items: [
                  hintCheckboxItem(tc, hintEnabled, setHintEnabled),
                  {
                    type: 'select' as const,
                    id: 'samCpuSpeed',
                    testId: 'sam-cpu-speed-select',
                    label: t('settings.cpuSpeed'),
                    tooltip: t('settings.cpuSpeedHelp'),
                    value: cpuSpeed,
                    options: [
                      { value: 'slow', label: t('settings.speedSlow') },
                      { value: 'normal', label: t('settings.speedNormal') },
                      { value: 'fast', label: t('settings.speedFast') },
                    ],
                    onSelect: handleSelectCpuSpeed,
                  },
                ],
              },
            ]}
          />
          {hintEnabled && currentHint ? (
            <HintTooltip reason={t(currentHint.reason)} confidence={currentHint.confidence} />
          ) : null}
          <LandscapeBanner message={t('landscapeBanner', { defaultValue: '' })} />

          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8 space-y-4">
            <PlayerSummary
              label={t('label.cpu')}
              sidesLabel={t('label.cpuSides')}
              player={opponent}
              cardWidth={cardWidth}
              dataTutorial="sam-opponent"
              handCountLabel={(count) => t('label.handCount', { count })}
              goalLabel={(count) => t('label.cpuGoal', { count })}
            />

            <div className="flex items-center justify-center gap-2 sm:gap-4" data-tutorial="sam-foundations">
              {state.foundations.map((pile, idx) => (
                <FoundationPile
                  key={`f-${idx}`}
                  idx={idx}
                  pile={pile}
                  topValue={state.foundationTops[idx]}
                  cardWidth={cardWidth}
                  highlight={isHintTarget(`hand${selection?.kind === 'hand' ? selection.idx : ''}-to-f${idx}`)}
                  selected={selection !== null}
                  onClick={() => handleFoundationClick(idx)}
                  label={t('label.foundation')}
                />
              ))}
            </div>

            <div className="text-center text-xs text-ds-secondary">
              {t('label.stock')}: {state.stockSize} / {t('label.completed')}: {state.completedSize}
            </div>

            <div className="flex items-end justify-center gap-3" data-tutorial="sam-goal">
              <GoalPile
                top={human.goalTop}
                size={human.goalSize}
                cardWidth={cardWidth}
                selected={selection?.kind === 'goal'}
                onClick={handleSelectGoal}
                label={t('label.goal')}
                playable={
                  isHumanTurn && !isGameOver && isGoalTopPlayableToFoundation(human.goalTop, state.foundationTops)
                }
              />
            </div>

            <HandRow
              hand={human.hand}
              cardWidth={cardWidth}
              selectedIdx={selectionIsHand ? selection.idx : null}
              onSelect={handleSelectHand}
              dataTutorial="sam-hand"
              hintEnabled={hintEnabled}
              currentHint={currentHint}
              label={t('label.hand')}
              emptyLabel={t('empty')}
              hiddenLabel={t('label.hidden')}
            />

            <SideRow
              sides={human.sides}
              cardWidth={cardWidth}
              cpuLabel={false}
              onSelect={handleSelectSide}
              onDiscard={handleDiscardSide}
              selection={selection}
              dataTutorial="sam-sides"
              label={t('label.side')}
              discardLabel={t('discard')}
              discardEnabled={selectionIsHand}
              discardHint={selectionIsHand ? t('discardReady') : t('discardNeedHand')}
            />
          </div>

          <div>
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameOver}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />

            <GameFooter>
              <GameResetButton
                isGameEnd={isGameOver}
                onReset={handleResetAction}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sam-reset"
              />

              {!isGameOver && isHumanTurn && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleAutoComplete}
                  disabled={loading || !state.canAutoComplete}
                  data-testid="sam-autocomplete-btn"
                >
                  {t('autoComplete')}
                </button>
              )}
            </GameFooter>
          </div>
        </>
      )}
    </GamePageShell>
  );
}

function PlayerSummary({
  label,
  sidesLabel,
  player,
  cardWidth,
  dataTutorial,
  handCountLabel,
  goalLabel,
}: {
  label: string;
  sidesLabel: string;
  player: SpiteAndMaliceResponse['players'][number];
  cardWidth: number;
  dataTutorial: string;
  handCountLabel: (count: number) => string;
  goalLabel: (count: number) => string;
}) {
  // Side piles render at half scale to keep the opponent strip compact on mobile.
  const sideWidth = Math.round(cardWidth * 0.5);
  return (
    <div className="flex flex-col items-center" data-tutorial={dataTutorial}>
      <span className="text-sm text-ds-secondary mb-1">
        {label} ({handCountLabel(player.hand.length)})
      </span>
      <div className="flex gap-2 items-end">
        {player.goalTop ? (
          <div className="flex flex-col items-center">
            <span className="text-xs text-ds-secondary">{goalLabel(player.goalSize)}</span>
            <AnimatedCard card={player.goalTop} width={cardWidth} />
          </div>
        ) : (
          <span className="text-xs text-ds-secondary">{goalLabel(0)}</span>
        )}
        <div className="flex flex-col items-center" data-testid="sam-cpu-sides">
          <span className="text-xs text-ds-secondary">{sidesLabel}</span>
          <div className="flex gap-1">
            {player.sides.map((pile, i) => {
              const top = pile.length > 0 ? pile[pile.length - 1] : undefined;
              return (
                <div
                  // Fixed-length 4-pile array; index is a stable key.
                  key={`cpu-side-${i}`}
                  data-testid={`sam-cpu-side-${i}`}
                  className="relative flex items-center justify-center"
                >
                  {top ? (
                    <AnimatedCard card={top} width={sideWidth} />
                  ) : (
                    <div
                      style={{ width: sideWidth, height: Math.round(sideWidth * 1.4) }}
                      className="rounded border border-dashed border-game-border"
                    />
                  )}
                  {pile.length > 0 && (
                    <span className="absolute -top-1 -right-1 rounded bg-ds-surface px-1 text-[9px] text-ds-text-primary leading-tight">
                      {pile.length}
                    </span>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}

function FoundationPile({
  idx,
  pile,
  topValue,
  cardWidth,
  highlight,
  selected,
  onClick,
  label,
}: {
  idx: number;
  pile: Card[];
  topValue: number;
  cardWidth: number;
  highlight: boolean;
  selected: boolean;
  onClick: () => void;
  label: string;
}) {
  const top = pile.length > 0 ? pile[pile.length - 1] : undefined;
  const ariaLabel = top ? `${label} ${idx + 1}: ${cardAlt(top)} (top=${topValue})` : `${label} ${idx + 1}: empty`;
  const baseRing = highlight ? 'ring-2 ring-ds-info' : '';
  const interactive = selected ? 'cursor-pointer hover:-translate-y-0.5' : 'cursor-default';
  return (
    <button
      type="button"
      data-hint-action={`hand-to-f${idx}`}
      className={`relative ${focusRingWhite} rounded-lg transition-transform ${baseRing} ${interactive}`}
      onClick={onClick}
      disabled={!selected}
      aria-label={ariaLabel}
    >
      {top ? <AnimatedCard card={top} width={cardWidth} /> : <FaceDownSlot label={`F${idx + 1}`} width={cardWidth} />}
      <span className="absolute -top-2 -right-1 text-[10px] bg-ds-surface px-1 rounded text-ds-secondary">
        {topValue === 0 ? '-' : topValue}
      </span>
    </button>
  );
}

function FaceDownSlot({ label, width }: { label: string; width: number }) {
  const height = Math.round(width * 1.4);
  return (
    <div
      className="flex items-center justify-center bg-ds-surface/70 border border-dashed border-ds-secondary rounded-md text-ds-secondary text-sm"
      style={{ width, height }}
    >
      {label}
    </div>
  );
}

function GoalPile({
  top,
  size,
  cardWidth,
  selected,
  onClick,
  label,
  playable,
}: {
  top?: Card;
  size: number;
  cardWidth: number;
  selected: boolean;
  onClick: () => void;
  label: string;
  playable: boolean;
}) {
  // Selection ring wins (the user explicitly chose this pile); otherwise the
  // playable affordance pulses a warning-colored glow so the strategically
  // most important pile attracts the eye first (#1886).
  const accentClass = selected
    ? 'ring-2 ring-ds-accent -translate-y-0.5'
    : playable
      ? 'ring-2 ring-ds-warning shadow-[0_0_15px_var(--color-ds-warning)] motion-safe:animate-pulse'
      : '';
  return (
    <button
      type="button"
      data-hint-action="goal-to-f0"
      data-goal-playable={playable ? 'true' : 'false'}
      className={`${focusRingWhite} flex flex-col items-center rounded-lg transition-transform ${accentClass}`}
      onClick={onClick}
      disabled={size === 0}
      aria-label={top ? `${label} top: ${cardAlt(top)} (${size} left)` : `${label}: empty`}
    >
      <span className="text-xs text-ds-secondary mb-1">
        {label} ({size})
      </span>
      {top ? <AnimatedCard card={top} width={cardWidth} /> : <FaceDownSlot label={label} width={cardWidth} />}
    </button>
  );
}

function HandRow({
  hand,
  cardWidth,
  selectedIdx,
  onSelect,
  dataTutorial,
  hintEnabled,
  currentHint,
  label,
  emptyLabel,
  hiddenLabel,
}: {
  hand: (Card | null)[];
  cardWidth: number;
  selectedIdx: number | null;
  onSelect: (idx: number) => void;
  dataTutorial: string;
  hintEnabled: boolean;
  currentHint: { targetAction: string } | null;
  label: string;
  emptyLabel: string;
  hiddenLabel: string;
}) {
  return (
    <div className="flex flex-col items-center" data-tutorial={dataTutorial}>
      <span className="text-sm text-ds-secondary mb-1">{label}</span>
      <div className="flex gap-2 flex-wrap justify-center">
        {hand.length === 0 ? (
          <span className="text-xs text-ds-secondary">{emptyLabel}</span>
        ) : (
          hand.map((card, idx) => {
            const selected = selectedIdx === idx;
            const hint = hintEnabled && currentHint?.targetAction.startsWith(`hand${idx}-`);
            const ring = selected ? 'ring-2 ring-ds-accent -translate-y-0.5' : hint ? 'ring-2 ring-ds-info' : '';
            return (
              <button
                key={`hand-${idx}`}
                type="button"
                data-hint-action={`hand${idx}`}
                className={`${focusRingWhite} relative rounded-lg transition-transform ${ring}`}
                onClick={() => onSelect(idx)}
                disabled={card === null}
                aria-label={`${label} ${(idx + 1).toString()}: ${card ? cardAlt(card) : hiddenLabel}`}
              >
                {card ? <AnimatedCard card={card} width={cardWidth} /> : <FaceDownSlot label="?" width={cardWidth} />}
              </button>
            );
          })
        )}
      </div>
    </div>
  );
}

function SideRow({
  sides,
  cardWidth,
  cpuLabel,
  onSelect,
  onDiscard,
  selection,
  dataTutorial,
  label,
  discardLabel,
  discardEnabled,
  discardHint,
}: {
  sides: [Card[], Card[], Card[], Card[]];
  cardWidth: number;
  cpuLabel: boolean;
  onSelect: (idx: number) => void;
  onDiscard: (idx: number) => void;
  selection: Selection;
  dataTutorial: string;
  label: string;
  discardLabel: string;
  discardEnabled: boolean;
  discardHint: string;
}) {
  return (
    <div className="flex flex-col items-center gap-2" data-tutorial={dataTutorial}>
      <span className="text-sm text-ds-secondary">{cpuLabel ? `CPU ${label}` : label}</span>
      {/* Shared reason for the discard buttons' disabled state (they all gate on
          a selected hand card), announced via aria-describedby below. */}
      <span id="sam-discard-hint" className="sr-only" data-testid="sam-discard-hint">
        {discardHint}
      </span>
      <div className="grid grid-cols-4 gap-2">
        {sides.map((pile, idx) => {
          const top = pile.length > 0 ? pile[pile.length - 1] : undefined;
          const selected = selection?.kind === 'side' && selection.idx === idx;
          const ring = selected ? 'ring-2 ring-ds-accent -translate-y-0.5' : '';
          return (
            <div key={`side-${idx}`} className="flex flex-col items-center gap-1">
              <button
                type="button"
                data-hint-action={`side${idx}`}
                className={`${focusRingWhite} rounded-lg transition-transform ${ring}`}
                onClick={() => onSelect(idx)}
                disabled={top === undefined}
                aria-label={
                  top ? `${label} ${idx + 1} top: ${cardAlt(top)} (${pile.length})` : `${label} ${idx + 1}: empty`
                }
              >
                {top ? (
                  <AnimatedCard card={top} width={cardWidth} />
                ) : (
                  <FaceDownSlot label={`S${idx + 1}`} width={cardWidth} />
                )}
              </button>
              <span className="text-[10px] text-ds-secondary">
                {label} {idx + 1} ({pile.length})
              </span>
              <button
                type="button"
                data-hint-action={`discard-${idx}`}
                className={`${btnPrimary} text-xs px-2 py-1`}
                onClick={() => onDiscard(idx)}
                disabled={!discardEnabled}
                aria-describedby="sam-discard-hint"
              >
                {discardLabel}
              </button>
            </div>
          );
        })}
      </div>
    </div>
  );
}
