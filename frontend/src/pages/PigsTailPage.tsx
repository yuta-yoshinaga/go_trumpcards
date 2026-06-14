import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { pigtailApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CircularDeck } from '../components/CircularDeck';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeErrorColors } from '../styles/badgeStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PigsTailResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';

const SUIT_SYMBOLS: Record<string, string> = {
  SPADE: '♠',
  HEART: '♥',
  DIAMOND: '♦',
  CLOVER: '♣',
};

const PT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="pt-circle-area"]',
    messageKey: 'tutorial.circleArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pt-center-area"]',
    messageKey: 'tutorial.centerArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pt-draw-button"]',
    messageKey: 'tutorial.drawButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pt-player-area"]',
    messageKey: 'tutorial.playerArea',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PIGTAIL_PHASE_PLAY = 0;
const PIGTAIL_PHASE_END = 1;

const PIGTAIL_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PIGTAIL_PHASE_PLAY]: 'play',
  [PIGTAIL_PHASE_END]: 'end',
};

/** Renders the Pig's Tail game page. */
export const PigsTailPage = withTutorial(PigsTailPageContent, 'pigtail', PT_TUTORIAL_STEPS);
/** Inner content of the Pig's Tail page. */
function PigsTailPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pigtail');
  const { state, loading, exec: execApi } = useGameApi(pigtailApi.exec);
  const handleDraw = useCallback(() => execApi('draw'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('pigtail', state);

  useMountReset(execApi);

  const phaseNames = usePhaseNames('pigtail', PIGTAIL_PHASE_KEYS);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pigtail');
  type PtArgs = Parameters<typeof pigtailApi.exec>;
  const cliConfig: CliGameConfig<PigsTailResponse, PtArgs> = useMemo(
    () => ({
      gameName: 'pigtail',
      parseCommand: (input: string): CliParseResult<PtArgs> => {
        const cmd = input.trim().toLowerCase();
        if (cmd === 'reset' || cmd === 'r') return { args: ['reset'] };
        if (cmd === 'draw' || cmd === 'd') return { args: ['draw'] };
        return { error: `Unknown command: ${cmd}` };
      },
      formatResponse: (s: PigsTailResponse) => {
        const lines: string[] = [];
        const phase = s.gameEndFlag ? 'End' : 'Play';
        lines.push(`Phase: ${phase} | Circle: ${s.circleCount} | Center: ${s.centerCount}`);
        for (const p of s.players) {
          const tag = p.isHuman ? 'You' : `CPU ${p.id}`;
          lines.push(`${tag}: ${p.cardCount} cards`);
        }
        if (s.lastDrawCard) {
          const card = `${s.lastDrawCard.design[0]}${s.lastDrawCard.value}`;
          lines.push(`Last: ${card} ${s.lastPenalty ? '(PENALTY)' : '(safe)'}`);
        }
        if (s.message) lines.push(s.message);
        return lines.join('\n');
      },
      helpText: ['d/draw  - Draw a card', 'r/reset - Reset game'],
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  // Penalty screen-flash: fire whenever lastPenalty transitions to true on a
  // new lastDrawCard. We key on centerCount+circleCount to detect fresh draws,
  // not just spurious re-renders.
  const [penaltyFlash, setPenaltyFlash] = useState(0);
  const prevDrawSigRef = useRef<string>('');
  useEffect(() => {
    if (!state) return;
    const sig = `${state.circleCount}-${state.centerCount}-${state.lastDrawCard ? `${state.lastDrawCard.design}${state.lastDrawCard.value}` : 'none'}`;
    if (sig !== prevDrawSigRef.current && state.lastPenalty) {
      setPenaltyFlash(Date.now());
    }
    prevDrawSigRef.current = sig;
  }, [state]);
  useEffect(() => {
    if (penaltyFlash === 0) return;
    const id = window.setTimeout(() => setPenaltyFlash(0), 600);
    return () => window.clearTimeout(id);
  }, [penaltyFlash]);

  if (!state)
    return <GameSkeleton gameKey="pigtail" layout={{ kind: 'centered', rows: [2], shape: 'circle', bars: 4 }} />;

  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = !isGameEnd && state.players[state.currentTurn]?.isHuman === true;
  const currentPhaseName = isGameEnd ? phaseNames[PIGTAIL_PHASE_END] : phaseNames[PIGTAIL_PHASE_PLAY];
  const loserIsHuman = isGameEnd && state.loserIdx >= 0 && state.players[state.loserIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.pigtail')}
      gameThemeBg={gameTheme.pigtail.bg}
      phaseName={currentPhaseName ?? ''}
      gamePath="/pigtail"
      gameEndFlag={!!isGameEnd}
      winShow={isGameEnd && !loserIsHuman}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {penaltyFlash > 0 && (
            <div
              aria-hidden="true"
              data-testid="pigtail-penalty-flash"
              className="pointer-events-none fixed inset-0 z-40 bg-ds-error/40 motion-safe:animate-[pulse-once_0.6s_ease-out]"
            />
          )}
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            {/* Circle & Center area */}
            <div className="flex flex-col items-center gap-3" data-tutorial="pt-circle-area">
              <div className="text-xs text-ds-text-muted">
                {t('label.circle')}: {state.circleCount}
              </div>
              <CircularDeck
                count={state.circleCount}
                cardWidth={32}
                diameter={160}
                onDrawCard={handleDraw}
                disabled={loading || isGameEnd || !isHumanTurn}
                drawAriaLabel={t('button.draw')}
              />
              <div className="text-center" data-tutorial="pt-center-area">
                <div className="text-xs text-ds-text-muted mb-1">
                  {t('label.center')} ({state.centerCount})
                </div>
                <div className="w-16 h-16 rounded-lg bg-ds-warning/60 border-2 border-ds-warning/40 flex items-center justify-center text-xl font-bold text-white">
                  {state.centerTop ? (SUIT_SYMBOLS[state.centerTop.design] ?? '?') : '-'}
                </div>
              </div>
            </div>

            {/* Last action indicator */}
            {state.lastDrawCard && (
              <div
                className={`text-center text-sm font-medium ${state.lastPenalty ? 'text-ds-error' : 'text-ds-success'}`}
              >
                {state.lastPenalty ? t('label.penalty') : t('label.safe')}
              </div>
            )}

            {/* CPU actions */}
            {state.cpuActions.length > 0 && (
              <div className="space-y-1">
                {state.cpuActions.map((action, i) => (
                  <div
                    key={i}
                    className={`text-xs px-2 py-1 rounded ${action.penaltyFlag ? badgeErrorColors : 'bg-black/30 text-ds-text-muted'}`}
                  >
                    CPU {action.drawPlayerIdx}:{' '}
                    {action.drawnCard ? (SUIT_SYMBOLS[action.drawnCard.design] ?? '?') + action.drawnCard.value : '?'}
                    {action.penaltyFlag
                      ? ` — ${t('label.penalty')} (+${action.penaltyCount})`
                      : ` — ${t('label.safe')}`}
                  </div>
                ))}
              </div>
            )}

            {/* Players */}
            <div className="space-y-2" data-tutorial="pt-player-area">
              {state.players.map((player, idx) => (
                <div
                  key={player.id}
                  className={`flex items-center justify-between px-3 py-2 rounded ${
                    !isGameEnd && state.currentTurn === idx
                      ? 'bg-ds-warning/30 border border-ds-warning/50'
                      : 'bg-black/30'
                  } ${isGameEnd && state.loserIdx === idx ? 'bg-ds-error/40 border border-ds-error/50' : ''}`}
                >
                  <span className="text-ds-text-primary text-sm font-medium">
                    {player.isHuman ? tc('player.you') : `CPU ${player.id}`}
                  </span>
                  <span className="text-ds-text-primary text-sm">
                    {player.cardCount} {t('label.cards')}
                  </span>
                </div>
              ))}
            </div>

            {/* Message */}
            {state.message && (
              <GameMessageBox
                message={state.message}
                messageCode={state.messageCode}
                messageParams={state.messageParams}
              />
            )}

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

          <GameFooter className={`${gameTheme.pigtail.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center items-center flex-wrap">
              <label className="flex items-center gap-1 text-ds-text-primary text-xs">
                <input
                  type="checkbox"
                  checked={hintEnabled}
                  onChange={(e) => setHintEnabled(e.target.checked)}
                  aria-label={tc('hint.toggle', { ns: 'tutorial' })}
                />
                {tc('hint.toggle', { ns: 'tutorial' })}
              </label>
              <button
                type="button"
                className="px-6 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                onClick={handleDraw}
                disabled={loading || isGameEnd || !isHumanTurn}
                data-tutorial="pt-draw-button"
              >
                {t('button.draw')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
              />
              <button
                type="button"
                className="px-4 py-2 rounded-lg bg-ds-surface-elevated hover:bg-ds-surface-elevated text-ds-text-primary text-sm transition-colors"
                onClick={showActionLog}
              >
                {tc('actionLog.view')}
              </button>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
