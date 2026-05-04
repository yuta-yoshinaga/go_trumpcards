import { useCallback, useEffect, useMemo } from 'react';
import { trashApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetButton } from '../components/GameResetButton';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { isGameRoundActive, useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { useMountReset } from '../hooks/useMountReset';
import { useSound } from '../providers/SoundProvider';
import { btnOutline, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TrashResponse, TrashSlot } from '../types/card';
import { TrashPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
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
  const { playSound } = useSound();
  const gameApi = useGameApi<TrashResponse, ApiArgs>((...args) => trashRunner.exec(...args));
  const { state, loading, error, retry } = gameApi;
  const apiCall = gameApi.exec;
  const { hint, hintEnabled, setHintEnabled } = useGameHint('trash', state);

  useMountReset(apiCall);

  // Drive the CPU turn automatically. Whenever it becomes the CPU's turn
  // (current === 1) and the game has not ended, fire one step after a short
  // delay so the UI can render the intermediate state.
  useEffect(() => {
    if (!state) return;
    if (state.phase === TrashPhase.GAME_OVER) return;
    if (state.current !== 1) return;
    const timer = setTimeout(() => {
      void apiCall('cpu');
    }, 500);
    return () => clearTimeout(timer);
  }, [state, apiCall]);

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
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiCall, trashCliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();

  const handleResetAction = useCallback(() => {
    void apiCall('reset');
    playSound('shuffle');
  }, [apiCall, playSound]);

  const handleDraw = useCallback(() => {
    void apiCall('draw');
    playSound('cardPlace');
  }, [apiCall, playSound]);

  const handleSlotClick = useCallback(
    (slotIdx: number) => {
      if (!state) return;
      if (state.phase !== TrashPhase.AWAIT_WILD) return;
      if (state.current !== 0) return;
      const slot = state.players[0].slots[slotIdx];
      if (slot.faceUp) return;
      void apiCall('place', slotIdx + 1);
      playSound('cardPlace');
    },
    [apiCall, state, playSound],
  );
  useGameRoundGuard(isGameRoundActive(state));

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

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.trash.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.trash')} />
      <PhaseIndicator phaseName={phaseName}>
        <span>
          {tc('moveCount', { defaultValue: 'Moves' })}: {state.moveCount}
        </span>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/trash" />
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner', { defaultValue: '' })} />

          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8 space-y-4">
            <PlayerRow
              label={t('label.cpu')}
              slots={state.players[1].slots}
              cardWidth={cardWidth}
              highlightFaceDown={false}
              onSlotClick={undefined}
              dataTutorial="tr-opponent"
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

            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            <GameFooter>
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
                isGameEnd={isGameOver}
                onReset={handleResetAction}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="tr-reset"
              />

              {isGameOver && (
                <button type="button" className={btnOutline} onClick={() => showActionLog()} disabled={loading}>
                  {tc('showActionLog', { defaultValue: 'Show action log' })}
                </button>
              )}
            </GameFooter>
          </div>
        </>
      )}

      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
      {isGameOver && state.winner === 0 && <WinCelebration show={true} />}
    </div>
  );
}

function PlayerRow({
  label,
  slots,
  cardWidth,
  highlightFaceDown,
  onSlotClick,
  dataTutorial,
}: {
  label: string;
  slots: TrashSlot[];
  cardWidth: number;
  highlightFaceDown: boolean;
  onSlotClick: ((idx: number) => void) | undefined;
  dataTutorial: string;
}) {
  return (
    <div className="flex flex-col items-center" data-tutorial={dataTutorial}>
      <span className="text-sm text-ds-secondary mb-1">{label}</span>
      <div className="grid grid-cols-5 gap-1 sm:gap-2">
        {slots.map((slot, idx) => {
          const key = `${idx}-${slot.faceUp ? `${slot.card?.design}-${slot.card?.value}` : 'face-down'}`;
          const interactive = !slot.faceUp && onSlotClick !== undefined;
          return (
            <button
              key={key}
              type="button"
              className={`relative ${focusRingWhite} rounded-lg transition-transform ${
                highlightFaceDown && !slot.faceUp ? 'ring-2 ring-ds-info hover:-translate-y-0.5' : ''
              } ${interactive ? 'cursor-pointer' : 'cursor-default'}`}
              onClick={() => onSlotClick?.(idx)}
              disabled={!interactive}
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
