import { useCallback, useEffect, useMemo } from 'react';
import { sixcardgolfApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
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
import { focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SixCardGolfResponse, SixCardGolfSlot } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';

const runner = sixcardgolfApi;

const SCG_PHASE_SETUP = 0;
const SCG_PHASE_PLAYER_TURN = 1;
const SCG_PHASE_DRAW_PENDING = 2;
const SCG_PHASE_ROUND_OVER = 3;
const SCG_PHASE_GAME_OVER = 4;

const TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="scg-grid"]',
    messageKey: 'tutorial.intro',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="scg-stock"]',
    messageKey: 'tutorial.draw',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="scg-grid"]',
    messageKey: 'tutorial.column',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="scg-score"]',
    messageKey: 'tutorial.score',
    placement: 'top',
    advanceOn: 'next',
  },
];

type ApiArgs = Parameters<typeof runner.exec>;

function parseSCGCommand(input: string): CliParseResult<ApiArgs> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase();
  const posArg = parts[1] ? Number.parseInt(parts[1], 10) : undefined;
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: [{ command: 'reset' }] };
    case 'fi':
    case 'flipinitial':
      if (posArg === undefined || Number.isNaN(posArg)) return { error: 'Usage: fi <pos>' };
      return { args: [{ command: 'flipinitial', position: posArg }] };
    case 'ds':
    case 'drawstock':
      return { args: [{ command: 'drawstock' }] };
    case 'dd':
    case 'drawdiscard':
      return { args: [{ command: 'drawdiscard' }] };
    case 'sw':
    case 'swap':
      if (posArg === undefined || Number.isNaN(posArg)) return { error: 'Usage: sw <pos>' };
      return { args: [{ command: 'swap', position: posArg }] };
    case 'di':
    case 'discard':
      return { args: [{ command: 'discard' }] };
    case 'fl':
    case 'flip':
      if (posArg === undefined || Number.isNaN(posArg)) return { error: 'Usage: fl <pos>' };
      return { args: [{ command: 'flip', position: posArg }] };
    case 'sf':
    case 'skipflip':
      return { args: [{ command: 'skipflip' }] };
    case 'nr':
    case 'nextround':
      return { args: [{ command: 'nextround' }] };
    case 'l':
    case 'log':
      return { args: [{ command: 'log' }] };
    default:
      return { error: `Unknown: ${cmd}` };
  }
}

const cliConfig: CliGameConfig<ApiArgs, SixCardGolfResponse> = {
  parseCommand: parseSCGCommand,
  formatState: (s) => `Round ${s.roundNumber}/${s.totalRounds} | Phase ${s.phase}`,
};

/** Six Card Golf game page content. */
function SixCardGolfPageContent() {
  const { t, tc, actionLog, showActionLog, confirmOpen, requestConfirm } = useGamePageSetup('sixcardgolf');
  const { state, loading, error, exec, retry } = useGameApi<ApiArgs, SixCardGolfResponse>(runner.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('sixcardgolf', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, clearLog } = useCliMode('sixcardgolf');
  const { cw, ch } = useCardDimensions();
  const { playWin } = useSound();

  useMountReset(useCallback(() => exec({ command: 'reset' }), [exec]));
  useCliGame(runner.exec, cliConfig, state, { logEntries, addInput, addOutput, clearLog }, cliEnabled);

  const phase = state?.phase ?? -1;
  const isGameEnd = state?.gameEndFlag ?? false;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const isHumanTurn = state ? state.currentPlayerIdx === humanIdx : false;
  const humanWon = isGameEnd && state?.winnerIdx === humanIdx;

  useEffect(() => {
    if (humanWon) playWin();
  }, [humanWon, playWin]);

  const handleFlipInitial = useCallback((pos: number) => exec({ command: 'flipinitial', position: pos }), [exec]);
  const handleDrawStock = useCallback(() => exec({ command: 'drawstock' }), [exec]);
  const handleDrawDiscard = useCallback(() => exec({ command: 'drawdiscard' }), [exec]);
  const handleSwap = useCallback((pos: number) => exec({ command: 'swap', position: pos }), [exec]);
  const handleDiscard = useCallback(() => exec({ command: 'discard' }), [exec]);
  const handleFlip = useCallback((pos: number) => exec({ command: 'flip', position: pos }), [exec]);
  const handleSkipFlip = useCallback(() => exec({ command: 'skipflip' }), [exec]);
  const handleNextRound = useCallback(() => exec({ command: 'nextround' }), [exec]);
  const handleReset = useCallback(() => exec({ command: 'reset' }), [exec]);

  const phaseLabel = useMemo(() => {
    if (!state) return '';
    if (isGameEnd) return t('phase.gameOver');
    switch (phase) {
      case SCG_PHASE_SETUP:
        return t('phase.setup');
      case SCG_PHASE_PLAYER_TURN:
        return state.canFlip ? t('phase.canFlip') : t('phase.playerTurn');
      case SCG_PHASE_DRAW_PENDING:
        return t('phase.drawPending');
      case SCG_PHASE_ROUND_OVER:
        return t('phase.roundOver');
      default:
        return '';
    }
  }, [state, phase, isGameEnd, t]);

  if (!state) return <GameSkeleton />;

  return (
    <GamePageShell
      title={tc('nav.sixcardgolf')}
      gameThemeBg={gameTheme.sixcardgolf.bg}
      gameThemeFooter={gameTheme.sixcardgolf.footer}
      phaseLabel={phaseLabel}
      isHumanTurn={isHumanTurn && !isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      onConfirmReset={handleReset}
      requestConfirm={requestConfirm}
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      headerExtra={
        <div className="flex items-center gap-2">
          <span className="text-xs opacity-75">
            {t('label.round')}: {state.roundNumber}/{state.totalRounds}
          </span>
          <CliToggle enabled={cliEnabled} onToggle={toggleCli} />
        </div>
      }
    >
      {cliEnabled ? (
        <CliTerminal entries={logEntries} />
      ) : (
        <div className="flex flex-col gap-3 p-3 overflow-y-auto">
          <LandscapeBanner />
          {error && <ErrorAlert message={error} onRetry={retry} />}
          <GameMessageBox
            messageCode={state.messageCode}
            messageParams={state.messageParams}
            message={state.message}
            ns="sixcardgolf"
          />
          {hint && hintEnabled && <HintTooltip hint={hint} ns="sixcardgolf" />}

          {/* Score Table */}
          <div className="flex gap-2 flex-wrap" data-tutorial="scg-score">
            {state.players.map((p, i) => (
              <div
                key={p.id}
                className={`text-xs px-2 py-1 rounded ${i === state.currentPlayerIdx ? 'bg-white/20 font-bold' : 'bg-white/10'}`}
              >
                {p.isHuman ? t('label.human') : t('label.cpu', { id: String(i) })}: {t('label.cumulative')}=
                {p.cumulativeScore} R={p.roundScore}
              </div>
            ))}
          </div>

          {/* Player Grids */}
          {state.players.map((player, pIdx) => (
            <div
              key={player.id}
              className={`rounded-lg p-2 ${pIdx === state.currentPlayerIdx ? 'ring-2 ring-ds-warning' : ''}`}
            >
              <div className="text-sm font-bold mb-1">
                {player.isHuman ? t('label.human') : t('label.cpu', { id: String(pIdx) })}
                {player.allFaceUp && <span className="ml-1 text-ds-warning">★</span>}
              </div>
              <div className="grid grid-cols-3 gap-1" data-tutorial={pIdx === humanIdx ? 'scg-grid' : undefined}>
                {player.grid.map((slot: SixCardGolfSlot, sIdx: number) => (
                  <GridSlotButton
                    key={`${player.id}-${sIdx}`}
                    slot={slot}
                    pos={sIdx}
                    cw={cw}
                    ch={ch}
                    isHumanGrid={player.isHuman}
                    phase={phase}
                    canFlip={state.canFlip}
                    isHumanTurn={isHumanTurn}
                    onFlipInitial={handleFlipInitial}
                    onSwap={handleSwap}
                    onFlip={handleFlip}
                  />
                ))}
              </div>
            </div>
          ))}

          {/* Stock / Discard / Drawn Card */}
          <div className="flex items-center gap-3 justify-center" data-tutorial="scg-stock">
            <div className="text-center">
              <div className="text-xs mb-1">
                {t('label.stock')} ({state.drawPileCount})
              </div>
              {phase === SCG_PHASE_PLAYER_TURN && isHumanTurn && !state.canFlip && (
                <button
                  type="button"
                  className={`px-3 py-1 rounded bg-ds-accent hover:bg-ds-accent-hover text-ds-text-on-accent text-sm ${focusRingWhite}`}
                  onClick={handleDrawStock}
                >
                  {t('button.drawStock')}
                </button>
              )}
            </div>
            {state.discardTop && (
              <div className="text-center">
                <div className="text-xs mb-1">{t('label.discard')}</div>
                <AnimatedCard card={state.discardTop} width={cw} height={ch} faceUp />
                {phase === SCG_PHASE_PLAYER_TURN && isHumanTurn && !state.canFlip && (
                  <button
                    type="button"
                    className={`mt-1 px-3 py-1 rounded bg-ds-success hover:bg-ds-success-hover text-white text-sm ${focusRingWhite}`}
                    onClick={handleDrawDiscard}
                  >
                    {t('button.drawDiscard')}
                  </button>
                )}
              </div>
            )}
            {state.drawnCard && phase === SCG_PHASE_DRAW_PENDING && (
              <div className="text-center">
                <div className="text-xs mb-1">{t('label.drawnCard')}</div>
                <AnimatedCard card={state.drawnCard} width={cw} height={ch} faceUp />
                <button
                  type="button"
                  className={`mt-1 px-3 py-1 rounded bg-ds-error hover:bg-ds-error-hover text-white text-sm ${focusRingWhite}`}
                  onClick={handleDiscard}
                >
                  {t('button.discard')}
                </button>
              </div>
            )}
          </div>

          {/* Flip / Skip buttons */}
          {state.canFlip && isHumanTurn && (
            <div className="flex justify-center">
              <button
                type="button"
                className={`px-4 py-2 rounded bg-ds-warning hover:bg-ds-warning-hover text-ds-text-on-accent text-sm ${focusRingWhite}`}
                onClick={handleSkipFlip}
              >
                {t('button.skipFlip')}
              </button>
            </div>
          )}

          {/* Next Round */}
          {phase === SCG_PHASE_ROUND_OVER && (
            <div className="flex justify-center">
              <button
                type="button"
                className={`px-4 py-2 rounded bg-ds-success hover:bg-ds-success-hover text-white ${focusRingWhite}`}
                onClick={handleNextRound}
              >
                {t('button.nextRound')}
              </button>
            </div>
          )}

          <ActionLogSection state={state} showActionLog={showActionLog} game="sixcardgolf" />
        </div>
      )}
      <GameFooter className={gameTheme.sixcardgolf.footer}>
        <GameResetButton onReset={requestConfirm} loading={loading} data-tutorial="scg-reset" />
        <HintTooltip.Toggle enabled={hintEnabled} onToggle={setHintEnabled} />
      </GameFooter>
    </GamePageShell>
  );
}

/** Grid slot button for a single card position. */
function GridSlotButton({
  slot,
  pos,
  cw,
  ch,
  isHumanGrid,
  phase,
  canFlip,
  isHumanTurn,
  onFlipInitial,
  onSwap,
  onFlip,
}: {
  slot: SixCardGolfSlot;
  pos: number;
  cw: number;
  ch: number;
  isHumanGrid: boolean;
  phase: number;
  canFlip: boolean;
  isHumanTurn: boolean;
  onFlipInitial: (pos: number) => void;
  onSwap: (pos: number) => void;
  onFlip: (pos: number) => void;
}) {
  const clickable =
    isHumanGrid &&
    isHumanTurn &&
    ((phase === SCG_PHASE_SETUP && !slot.faceUp) || phase === SCG_PHASE_DRAW_PENDING || (canFlip && !slot.faceUp));

  const handleClick = useCallback(() => {
    if (!clickable) return;
    if (phase === SCG_PHASE_SETUP) {
      onFlipInitial(pos);
    } else if (phase === SCG_PHASE_DRAW_PENDING) {
      onSwap(pos);
    } else if (canFlip) {
      onFlip(pos);
    }
  }, [clickable, phase, canFlip, pos, onFlipInitial, onSwap, onFlip]);

  return (
    <button
      type="button"
      className={`relative flex items-center justify-center rounded transition-all ${
        clickable
          ? `cursor-pointer ring-2 ring-ds-warning hover:ring-ds-warning-hover ${focusRingWhite}`
          : 'cursor-default'
      }`}
      style={{ width: cw + 8, height: ch + 8 }}
      onClick={handleClick}
      disabled={!clickable}
      aria-label={slot.faceUp && slot.card ? cardAlt(slot.card) : `Position ${pos} (face down)`}
    >
      <AnimatedCard card={slot.card} width={cw} height={ch} faceUp={slot.faceUp} />
      <span className="absolute bottom-0 right-0.5 text-[10px] opacity-50">{pos}</span>
    </button>
  );
}

/** Six Card Golf page wrapped with tutorial provider. */
export const SixCardGolfPage = withTutorial(SixCardGolfPageContent, 'sixcardgolf', TUTORIAL_STEPS);
