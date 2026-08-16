import { useCallback, useEffect, useMemo, useState } from 'react';
import { trashApi } from '../api/gameApi';
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
import { focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TrashResponse, TrashSlot } from '../types/card';
import { TrashPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { hintCliText, isHintCommand } from '../utils/cli/hintText';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';

const trashRunner = trashApi;

const TR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tr-opponent"]',
    messageKey: 'tutorial.welcome',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-stock"]',
    messageKey: 'tutorial.draw',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-player"]',
    messageKey: 'tutorial.wild',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tr-reset"]',
    messageKey: 'tutorial.win',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** CPU turn-pacing presets — a shorter delay advances the CPU turn faster. */
type TrashCpuSpeed = 'slow' | 'normal' | 'fast';

/**
 * Delay (ms) before the CPU turn auto-advances, per speed preset. `normal`
 * (500ms) preserves the historical default for backward compatibility.
 */
const TRASH_CPU_DELAY_MS: Record<TrashCpuSpeed, number> = {
  slow: 900,
  normal: 500,
  fast: 200,
};

const TRASH_CPU_SPEED_STORAGE_KEY = 'trash:cpuSpeed';

/** Read the persisted CPU speed, falling back to `normal` when unset/invalid. */
function loadTrashCpuSpeed(): TrashCpuSpeed {
  try {
    const v = localStorage.getItem(TRASH_CPU_SPEED_STORAGE_KEY);
    if (v === 'slow' || v === 'normal' || v === 'fast') return v;
  } catch {
    // localStorage may be unavailable (private mode / SSR); fall through to default.
  }
  return 'normal';
}

type ApiArgs = Parameters<typeof trashRunner.exec>;

function parseTrashCommand(input: string): CliParseResult<ApiArgs> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase();
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'cpu':
      return { args: ['cpu'] };
    case 'l':
    case 'log':
      return { args: ['log'] };
    case 'p':
    case 'place': {
      if (parts.length !== 2) return { error: 'Usage: p <pos>' };
      const pos = Number.parseInt(parts[1], 10);
      if (Number.isNaN(pos) || pos < 1 || pos > 10) return { error: 'Position must be 1-10' };
      return { args: ['place', pos] };
    }
    default:
      return { error: `Unknown command: ${cmd}` };
  }
}

function formatTrashState(state: TrashResponse): string {
  const lines: string[] = [];
  lines.push(`Phase: ${state.phase}  Turn: ${state.current === 0 ? 'YOU' : 'CPU'}  Moves: ${state.moveCount}`);
  const dTop = state.discardTop;
  lines.push(
    `Stock: ${state.stockSize}  Discard: ${state.discardSize}${dTop ? ` (top: ${dTop.design[0]}${dTop.value})` : ''}`,
  );
  for (let i = 0; i < state.players.length; i++) {
    const p = state.players[i];
    const row = p.slots
      .map((s, idx) => (s.faceUp && s.card ? `[${idx + 1}:${s.card.design[0]}${s.card.value}]` : `[${idx + 1}: ?]`))
      .join(' ');
    lines.push(`${p.isCpu ? 'CPU' : 'YOU'}: ${row}`);
  }
  if (state.pending) lines.push(`Pending: ${state.pending.design[0]}${state.pending.value}`);
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}

export const TrashPage = withTutorial(TrashPageContent, 'trash', TR_TUTORIAL_STEPS);
function TrashPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('trash');
  const gameApi = useGameApi<TrashResponse, ApiArgs>((...args) => trashRunner.exec(...args));
  const { state, loading, error, retry } = gameApi;
  const apiCall = gameApi.exec;
  const { hint, hintEnabled, setHintEnabled } = useGameHint('trash', state);
  const [cpuSpeed, setCpuSpeed] = useState<TrashCpuSpeed>(loadTrashCpuSpeed);

  const handleSelectCpuSpeed = useCallback((v: string) => {
    const speed: TrashCpuSpeed = v === 'slow' || v === 'fast' ? v : 'normal';
    setCpuSpeed(speed);
    try {
      localStorage.setItem(TRASH_CPU_SPEED_STORAGE_KEY, speed);
    } catch {
      // Persistence is best-effort; ignore storage failures.
    }
  }, []);

  useMountReset(apiCall);

  // Drive the CPU turn automatically. Whenever it becomes the CPU's turn
  // (current === 1) and the game has not ended, fire one step after a delay
  // (scaled by the chosen speed preset) so the UI can render the intermediate
  // state. Changing the speed reschedules the pending timer.
  useEffect(() => {
    if (!state) return;
    if (state.phase === TrashPhase.GAME_OVER) return;
    if (state.current !== 1) return;
    const timer = setTimeout(() => {
      void apiCall('cpu');
    }, TRASH_CPU_DELAY_MS[cpuSpeed]);
    return () => clearTimeout(timer);
  }, [state, apiCall, cpuSpeed]);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('trash');
  const trashCliConfig: CliGameConfig<TrashResponse, ApiArgs> = useMemo(
    () => ({
      gameName: 'trash',
      parseCommand: parseTrashCommand,
      formatResponse: formatTrashState,
      helpText: [
        'd              Draw from the stock',
        'p <pos>        Place wild at position pos (1-10)',
        'cpu            Advance CPU turn',
        'r              Reset',
        'l              Action log',
      ],
      localCommand: (input: string) => (isHintCommand(input) ? hintCliText(hint) : null),
    }),
    [hint],
  );
  const { handleCommand } = useCliGame(apiCall, trashCliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();

  const handleResetAction = useCallback(() => {
    void apiCall('reset');
  }, [apiCall]);

  const handleDraw = useCallback(() => {
    void apiCall('draw');
  }, [apiCall]);

  const handleSlotClick = useCallback(
    (slotIdx: number) => {
      if (!state) return;
      if (state.phase !== TrashPhase.AWAIT_WILD) return;
      if (state.current !== 0) return;
      const slot = state.players[0].slots[slotIdx];
      if (slot.faceUp) return;
      void apiCall('place', slotIdx + 1);
    },
    [apiCall, state],
  );

  if (error) return <ErrorAlert message={error} onRetry={retry} />;

  if (!state) return <GameSkeleton gameKey="trash" layout={{ kind: 'tableau', topRow: 6, tableau: 7 }} />;

  const isGameOver = state.phase === TrashPhase.GAME_OVER;
  const isAwaitWild = state.phase === TrashPhase.AWAIT_WILD;
  const isHumanTurn = state.current === 0;

  const phaseName = isGameOver
    ? t('phase.gameOver')
    : state.current === 1
      ? t('phase.cpuTurn')
      : isAwaitWild
        ? t('phase.awaitWild')
        : t('phase.playerTurn');

  // Slot index where the pending non-wild card should land. Wild cards (K/Joker)
  // route through AWAIT_WILD's existing highlightFaceDown path, so this stays
  // null whenever isAwaitWild is true or the value is outside 1..10.
  const pendingTargetIdx =
    isHumanTurn && !isAwaitWild && state.pending && state.pending.value >= 1 && state.pending.value <= 10
      ? state.pending.value - 1
      : null;

  // Announce a freshly drawn (pending) card and where it can go, since the pulse
  // ring, "保留中" label, and slot highlight are all purely visual. Wild cards
  // (K / Joker) let the player choose any slot; J/Q are dead and get discarded.
  const pendingIsWild = !!state.pending && (state.pending.value === 13 || state.pending.design === 'JOKER');
  const pendingAnnounce = (() => {
    if (!state.pending || !isHumanTurn) return '';
    const cardName = cardAlt(state.pending);
    if (isAwaitWild || pendingIsWild) return t('pendingAnnounce.wild', { card: cardName });
    // Announce a target slot only when that slot is still face-down (placeable);
    // if it is already filled the card is dead, mirroring the visual highlight
    // (pendingHighlight is gated on !slot.faceUp).
    const targetSlot = pendingTargetIdx !== null ? state.players[0].slots[pendingTargetIdx] : null;
    if (targetSlot && !targetSlot.faceUp) {
      return t('pendingAnnounce.slot', { card: cardName, slot: (pendingTargetIdx ?? 0) + 1 });
    }
    return t('pendingAnnounce.dead', { card: cardName });
  })();

  return (
    <GamePageShell
      title={tc('nav.trash')}
      gameThemeBg={gameTheme.trash.bg}
      phaseName={phaseName}
      gamePath="/trash"
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
          <LandscapeBanner message={t('landscapeBanner', { defaultValue: '' })} />

          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8 space-y-4">
            <PlayerRow
              label={t('label.cpu')}
              badge={t('label.cpuOpen', {
                open: state.players[1].slots.filter((s) => s.faceUp).length,
                total: state.players[1].slots.length,
              })}
              slots={state.players[1].slots}
              cardWidth={cardWidth}
              highlightFaceDown={false}
              onSlotClick={undefined}
              dataTutorial="tr-opponent"
              pendingTargetIdx={null}
            />

            <div className="flex items-center justify-center gap-6" data-tutorial="tr-stock">
              <StockPile
                size={state.stockSize}
                onClick={handleDraw}
                disabled={!isHumanTurn || isAwaitWild || isGameOver}
              />
              {state.pending && (
                <div className="flex flex-col items-center">
                  <span className="text-xs text-ds-secondary mb-1">{t('label.pending')}</span>
                  <AnimatedCard card={state.pending} width={cardWidth} />
                </div>
              )}
              {/* Always-mounted live region (only its text is conditional) so screen
                  readers reliably announce each drawn card and its placement target. */}
              <div className="sr-only" role="status" aria-live="polite" data-testid="tr-pending-announce">
                {pendingAnnounce}
              </div>
              <DiscardPile
                top={state.discardTop}
                size={state.discardSize}
                cardWidth={cardWidth}
                label={t('label.discard')}
              />
            </div>

            <PlayerRow
              label={t('label.you')}
              slots={state.players[0].slots}
              cardWidth={cardWidth}
              highlightFaceDown={isAwaitWild && isHumanTurn}
              onSlotClick={isAwaitWild && isHumanTurn ? handleSlotClick : undefined}
              dataTutorial="tr-player"
              pendingTargetIdx={pendingTargetIdx}
            />
          </div>

          <div>
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <SettingsPanel
              title={tc('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'select' as const,
                      id: 'trashCpuSpeed',
                      testId: 'trash-cpu-speed-select',
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

            <ActionLogSection
              isEndPhase={isGameOver}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />

            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            <GameFooter>
              <label className="flex items-center gap-1 text-ds-text-primary text-xs min-h-[44px]">
                <input
                  type="checkbox"
                  checked={hintEnabled}
                  onChange={(e) => setHintEnabled(e.target.checked)}
                  aria-label={tc('hint.toggle', { ns: 'tutorial' })}
                />
                {tc('hint.toggle', { ns: 'tutorial' })}
              </label>
              <GameResetButton
                isGameEnd={isGameOver}
                onReset={handleResetAction}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="tr-reset"
              />
            </GameFooter>
          </div>
        </>
      )}
    </GamePageShell>
  );
}

function PlayerRow({
  label,
  badge,
  slots,
  cardWidth,
  highlightFaceDown,
  onSlotClick,
  dataTutorial,
  pendingTargetIdx,
}: {
  label: string;
  /** Optional progress indicator rendered next to the label (e.g. "3/10 枚オープン"). */
  badge?: string;
  slots: TrashSlot[];
  cardWidth: number;
  highlightFaceDown: boolean;
  onSlotClick: ((idx: number) => void) | undefined;
  dataTutorial: string;
  /** Slot index (0..9) where a freshly drawn non-wild card should land, or null. */
  pendingTargetIdx: number | null;
}) {
  return (
    <div className="flex flex-col items-center" data-tutorial={dataTutorial}>
      <span className="text-sm text-ds-secondary mb-1">
        {label}
        {badge && (
          <span className="ml-2 px-1.5 py-0.5 rounded bg-ds-surface/70 text-xs text-ds-secondary whitespace-nowrap">
            {badge}
          </span>
        )}
      </span>
      <div className="grid grid-cols-5 gap-1 sm:gap-2">
        {slots.map((slot, idx) => {
          const key = `${idx}-${slot.faceUp ? `${slot.card?.design}-${slot.card?.value}` : 'face-down'}`;
          const interactive = !slot.faceUp && onSlotClick !== undefined;
          const wildHighlight = highlightFaceDown && !slot.faceUp;
          const pendingHighlight = pendingTargetIdx === idx && !slot.faceUp;
          // Wild highlight (info ring on every face-down slot) wins over the
          // single-slot pending highlight so the broader affordance reads first.
          const ringClass = wildHighlight
            ? 'ring-2 ring-ds-info hover:-translate-y-0.5'
            : pendingHighlight
              ? 'ring-2 ring-ds-warning motion-safe:animate-pulse'
              : '';
          return (
            <button
              key={key}
              type="button"
              className={`relative ${focusRingWhite} rounded-lg transition-transform ${ringClass} ${
                interactive ? 'cursor-pointer' : 'cursor-default'
              }`}
              onClick={() => onSlotClick?.(idx)}
              disabled={!interactive}
              data-pending-target={pendingHighlight ? 'true' : 'false'}
              aria-label={slot.faceUp && slot.card ? `${idx + 1}: ${cardAlt(slot.card)}` : `${idx + 1}: face-down`}
            >
              {slot.faceUp && slot.card ? (
                <AnimatedCard card={slot.card} width={cardWidth} />
              ) : (
                <FaceDownSlot idx={idx + 1} width={cardWidth} />
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}

function FaceDownSlot({ idx, width }: { idx: number; width: number }) {
  const height = Math.round(width * 1.4);
  return (
    <div
      className="flex items-center justify-center bg-ds-surface/70 border border-dashed border-ds-secondary rounded-md text-ds-secondary text-sm"
      style={{ width, height }}
    >
      {idx}
    </div>
  );
}

function StockPile({ size, onClick, disabled }: { size: number; onClick: () => void; disabled: boolean }) {
  return (
    <button
      type="button"
      className={`${focusRingWhite} flex flex-col items-center px-3 py-2 rounded-md border border-ds-accent bg-ds-surface/80 hover:bg-ds-surface disabled:opacity-50 disabled:cursor-not-allowed`}
      onClick={onClick}
      disabled={disabled}
    >
      <span className="text-3xl">🂠</span>
      <span className="text-xs text-ds-secondary mt-1">{size}</span>
    </button>
  );
}

function DiscardPile({
  top,
  size,
  cardWidth,
  label,
}: {
  top: TrashSlot['card'];
  size: number;
  cardWidth: number;
  label: string;
}) {
  return (
    <div className="flex flex-col items-center">
      <span className="text-xs text-ds-secondary mb-1">
        {label} ({size})
      </span>
      {top ? (
        <AnimatedCard card={top} width={cardWidth} />
      ) : (
        <div style={{ width: cardWidth, height: Math.round(cardWidth * 1.4) }} />
      )}
    </div>
  );
}
