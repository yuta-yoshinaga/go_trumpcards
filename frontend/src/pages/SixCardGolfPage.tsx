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
import { useSound } from '../providers/SoundProvider';
import { focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SixCardGolfResponse, SixCardGolfSlot } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSixCardGolfCommand } from '../utils/cli/commands/sixcardgolfCommands';
import { formatSixCardGolfState } from '../utils/cli/formatters/sixcardgolfFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { sixCardGolfColumnScores } from '../utils/sixCardGolfColumnScores';

const runner = sixcardgolfApi;

const SCG_PHASE_SETUP = 0;
const SCG_PHASE_PLAYER_TURN = 1;
const SCG_PHASE_DRAW_PENDING = 2;
const SCG_PHASE_ROUND_OVER = 3;

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

const cliConfig: CliGameConfig<SixCardGolfResponse, ApiArgs> = {
  gameName: 'sixcardgolf',
  parseCommand: parseSixCardGolfCommand,
  formatResponse: formatSixCardGolfState,
  helpText: [
    'fi <pos>       Flip initial card at position',
    'ds             Draw from stock',
    'dd             Draw from discard',
    'sw <pos>       Swap drawn card with grid position',
    'di             Discard drawn card',
    'fl <pos>       Flip face-down card at position',
    'sf             Skip flip',
    'nr             Next round',
    'r              Reset',
    'l              Action log',
  ],
};

/** Six Card Golf game page content. */
function SixCardGolfPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sixcardgolf');
  const {
    state,
    loading,
    error,
    exec: apiCall,
    retry,
  } = useGameApi<SixCardGolfResponse, ApiArgs>((...args) => runner.exec(...args));
  const { hint, hintEnabled, setHintEnabled } = useGameHint('sixcardgolf', state);
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sixcardgolf');
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  // Mount-time reset: useMountReset expects (...args: ['reset']) => unknown
  // but our API takes an object param, so use a direct useEffect instead.
  // biome-ignore lint/correctness/useExhaustiveDependencies: mount-only reset
  useEffect(() => {
    void apiCall({ command: 'reset' });
  }, []);

  const { handleCommand } = useCliGame(apiCall, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const phase = state?.phase ?? -1;
  const isGameEnd = state?.gameEndFlag ?? false;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const isHumanTurn = state ? state.currentPlayerIdx === humanIdx : false;
  const humanWon = isGameEnd && state?.winnerIdx === humanIdx;

  useEffect(() => {
    if (humanWon) playSound('winFanfare');
  }, [humanWon, playSound]);

  const handleFlipInitial = useCallback((pos: number) => apiCall({ command: 'flipinitial', position: pos }), [apiCall]);
  const handleDrawStock = useCallback(() => apiCall({ command: 'drawstock' }), [apiCall]);
  const handleDrawDiscard = useCallback(() => apiCall({ command: 'drawdiscard' }), [apiCall]);
  const handleSwap = useCallback((pos: number) => apiCall({ command: 'swap', position: pos }), [apiCall]);
  const handleDiscard = useCallback(() => apiCall({ command: 'discard' }), [apiCall]);
  const handleFlip = useCallback((pos: number) => apiCall({ command: 'flip', position: pos }), [apiCall]);
  const handleSkipFlip = useCallback(() => apiCall({ command: 'skipflip' }), [apiCall]);
  const handleNextRound = useCallback(() => apiCall({ command: 'nextround' }), [apiCall]);
  const handleReset = useCallback(() => apiCall({ command: 'reset' }), [apiCall]);

  const phaseName = useMemo(() => {
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

  if (!state)
    return <GameSkeleton gameKey="sixcardgolf" layout={{ kind: 'card-grid', count: 6, cols: 'grid-cols-3' }} />;

  return (
    <GamePageShell
      title={tc('nav.sixcardgolf')}
      gameThemeBg={gameTheme.sixcardgolf.bg}
      phaseName={phaseName}
      gamePath="/sixcardgolf"
      isHumanTurn={isHumanTurn && !isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      headerExtra={
        <div className="flex items-center gap-2">
          <span className="text-xs opacity-75">
            {t('label.round')}: {state.roundNumber}/{state.totalRounds}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </div>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <div className="flex flex-col gap-3 p-3 overflow-y-auto">
          <LandscapeBanner message={t('landscapeBanner', { defaultValue: '' })} />
          {error && <ErrorAlert message={error} onRetry={retry} />}
          <GameMessageBox messageCode={state.messageCode} messageParams={state.messageParams} message={state.message} />
          {hint && hintEnabled && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

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
                {player.allFaceUp && <span className="ml-1 text-ds-warning">&#9733;</span>}
              </div>
              <div className="grid grid-cols-3 gap-1" data-tutorial={pIdx === humanIdx ? 'scg-grid' : undefined}>
                {player.grid.map((slot: SixCardGolfSlot, sIdx: number) => (
                  <GridSlotButton
                    key={`${player.id}-${sIdx}`}
                    slot={slot}
                    pos={sIdx}
                    faceDownLabel={t('gridSlotFaceDownAria', { pos: sIdx + 1 })}
                    cardWidth={cardWidth}
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
              {player.isHuman && (phase === SCG_PHASE_ROUND_OVER || isGameEnd) && (
                <div className="mt-1 grid grid-cols-3 gap-1" data-testid="scg-column-scores">
                  {sixCardGolfColumnScores(player.grid).map((col, cIdx) => (
                    <span
                      key={cIdx}
                      data-testid={`scg-column-score-${cIdx}`}
                      className={`rounded px-1 py-0.5 text-center font-medium text-xs ${
                        col.isPair ? 'bg-ds-success text-white' : 'bg-ds-surface-elevated text-ds-text-muted'
                      }`}
                    >
                      {t('label.columnScore', { score: col.score })}
                    </span>
                  ))}
                </div>
              )}
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
                <AnimatedCard card={state.discardTop} width={cardWidth} />
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
                <AnimatedCard card={state.drawnCard} width={cardWidth} />
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

          <ActionLogSection
            isEndPhase={isGameEnd}
            actionLog={actionLog}
            showActionLog={showActionLog}
            hideActionLog={hideActionLog}
          />
        </div>
      )}
      <GameFooter className={gameTheme.sixcardgolf.footer}>
        <GameResetButton
          isGameEnd={isGameEnd}
          onReset={handleReset}
          requestConfirm={requestConfirm}
          loading={loading}
          dataTutorial="scg-reset"
        />
        <label className="flex items-center gap-1 text-ds-text-primary text-xs">
          <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
          {tc('hint')}
        </label>
      </GameFooter>
    </GamePageShell>
  );
}

/** Grid slot button for a single card position. */
function GridSlotButton({
  slot,
  pos,
  faceDownLabel,
  cardWidth,
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
  faceDownLabel: string;
  cardWidth: number;
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
      style={{ width: cardWidth + 8 }}
      onClick={handleClick}
      disabled={!clickable}
      aria-label={slot.faceUp && slot.card ? cardAlt(slot.card) : faceDownLabel}
    >
      {slot.faceUp && slot.card ? (
        <AnimatedCard card={slot.card} width={cardWidth} />
      ) : (
        <div
          className="flex items-center justify-center bg-ds-surface/70 border border-dashed border-ds-secondary rounded-md text-ds-secondary text-sm"
          style={{ width: cardWidth, height: cardWidth * 1.5 }}
        >
          ?
        </div>
      )}
      <span className="absolute bottom-0 right-0.5 text-[10px] opacity-50">{pos}</span>
    </button>
  );
}

/** Six Card Golf page wrapped with tutorial provider. */
export const SixCardGolfPage = withTutorial(SixCardGolfPageContent, 'sixcardgolf', TUTORIAL_STEPS);
