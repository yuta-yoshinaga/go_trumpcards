import { useCallback, useEffect, useMemo } from 'react';
import { pigtailApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { ManualButton } from '../components/ManualButton';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { PigsTailSkeleton } from '../components/skeleton/PigsTailSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
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
export function PigsTailPage() {
  return (
    <TutorialWrapper gameName="pigtail" steps={PT_TUTORIAL_STEPS}>
      <PigsTailPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Pig's Tail page. */
function PigsTailPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pigtail');
  const { state, loading, exec: execApi } = useGameApi(pigtailApi.exec);
  const handleDraw = useCallback(() => execApi('draw'), [execApi]);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);

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

  if (!state) return <PigsTailSkeleton />;

  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = !isGameEnd && state.players[state.currentTurn]?.isHuman === true;
  const currentPhaseName = isGameEnd ? phaseNames[PIGTAIL_PHASE_END] : phaseNames[PIGTAIL_PHASE_PLAY];
  const loserIsHuman = isGameEnd && state.loserIdx >= 0 && state.players[state.loserIdx]?.isHuman === true;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-green" aria-busy={loading}>
      <GamePageHeading title={tc('nav.pigtail')} />
      <PhaseIndicator phaseName={currentPhaseName ?? ''}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/pigtail" />
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            {/* Circle & Center area */}
            <div className="flex justify-center gap-8 items-center" data-tutorial="pt-circle-area">
              <div className="text-center">
                <div className="text-xs text-white/60 mb-1">{t('label.circle')}</div>
                <div className="w-16 h-16 rounded-full bg-blue-900/60 border-2 border-blue-400/40 flex items-center justify-center text-2xl font-bold text-white">
                  {state.circleCount}
                </div>
              </div>
              <div className="text-center" data-tutorial="pt-center-area">
                <div className="text-xs text-white/60 mb-1">
                  {t('label.center')} ({state.centerCount})
                </div>
                <div className="w-16 h-16 rounded-lg bg-amber-900/60 border-2 border-amber-400/40 flex items-center justify-center text-xl font-bold text-white">
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
                    className={`text-xs px-2 py-1 rounded ${action.penaltyFlag ? 'bg-red-900/40 text-red-200' : 'bg-black/30 text-white/70'}`}
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
                      ? 'bg-yellow-600/30 border border-yellow-400/50'
                      : 'bg-black/30'
                  } ${isGameEnd && state.loserIdx === idx ? 'bg-red-900/40 border border-red-500/50' : ''}`}
                >
                  <span className="text-white text-sm font-medium">
                    {player.isHuman ? tc('player.you') : `CPU ${player.id}`}
                  </span>
                  <span className="text-white/80 text-sm">
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
          </div>

          <GameFooter className="bg-game-bg-green-dark border-white/20 px-4 py-2.5">
            <div className="flex gap-2 justify-center">
              <button
                type="button"
                className="px-6 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                onClick={handleDraw}
                disabled={loading || isGameEnd || !isHumanTurn}
                data-tutorial="pt-draw-button"
              >
                {t('button.draw')}
              </button>
              <button
                type="button"
                className="px-4 py-2 rounded-lg bg-gray-600 hover:bg-gray-500 text-white text-sm transition-colors"
                onClick={() => requestConfirm(handleReset)}
              >
                {tc('button.reset')}
              </button>
              <button
                type="button"
                className="px-4 py-2 rounded-lg bg-gray-600 hover:bg-gray-500 text-white text-sm transition-colors"
                onClick={showActionLog}
              >
                {tc('actionLog.view')}
              </button>
            </div>
          </GameFooter>

          <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
          <ActionLogSection
            isEndPhase={isGameEnd}
            actionLog={actionLog}
            showActionLog={showActionLog}
            hideActionLog={hideActionLog}
          />
          <WinCelebration show={isGameEnd && !loserIsHuman} />
        </>
      )}
    </div>
  );
}
