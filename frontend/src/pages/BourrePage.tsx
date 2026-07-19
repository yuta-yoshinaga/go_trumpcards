import { useCallback, useEffect, useMemo, useState } from 'react';
import { bourreApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSecondary, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BourreResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { isRedSuitDesign, isSuitDesign, suitSymbol } from '../utils/cardAlt';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

type ApiArgs = {
  command: string;
  decide?: boolean;
  indices?: number[];
  cardIndex?: number;
  config?: { cpuDifficulty?: number };
};

const BOURRE_PENALTY_WARN_THRESHOLD = 10;

/** Plain-text trump-suit glyph for the CLI formatter: a suit symbol, or '-' when unset. */
function trumpSuitText(design: string): string {
  return isSuitDesign(design) ? suitSymbol(design) : '-';
}

const TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bourre-hand"]', messageKey: 'tutorial.intro', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="bourre-controls"]', messageKey: 'tutorial.decide', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="bourre-table"]', messageKey: 'tutorial.play', placement: 'bottom', advanceOn: 'next' },
];

function parseBourreCommand(input: string): CliParseResult<[ApiArgs]> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase() ?? '';
  const args = parts.slice(1);
  switch (cmd) {
    case 'd':
    case 'decide': {
      const v = args[0] ? Number.parseInt(args[0], 10) : 1;
      return { args: [{ command: 'decide', decide: v !== 0 }] };
    }
    case 'dr':
    case 'draw': {
      const indices = args.map(Number).filter((n) => !Number.isNaN(n));
      return { args: [{ command: 'draw', indices }] };
    }
    case 'p':
    case 'play': {
      const idx = args[0] ? Number.parseInt(args[0], 10) : 0;
      return { args: [{ command: 'p', cardIndex: Number.isNaN(idx) ? 0 : idx }] };
    }
    case 'n':
    case 'next':
      return { args: [{ command: 'next' }] };
    case 'r':
    case 'reset':
      return { args: [{ command: 'reset' }] };
    case 'sd': {
      const v = args[0] ? Number.parseInt(args[0], 10) : 0;
      return { args: [{ command: 'reset', config: { cpuDifficulty: Number.isNaN(v) ? 0 : v } }] };
    }
    case 'log':
    case 'l':
      return { args: [{ command: 'l' }] };
    default:
      return { error: `Unknown command: ${cmd}. Try: d, dr, p, n, r, sd, log` };
  }
}

function formatBourreState(state: BourreResponse): string {
  const lines: string[] = [`Phase: ${state.phase}`, `Pot: ${state.pot}  Trump: ${trumpSuitText(state.trumpSuit)}`];
  for (const p of state.players) {
    const name = p.isHuman ? 'You' : `CPU ${p.id}`;
    const status = p.isFinished ? 'out' : p.folded ? 'folded' : `${p.tricks} tricks`;
    lines.push(`${name}: ${p.chips} chips (${status})`);
  }
  if (state.message) lines.push(state.message);
  return lines.join('\n');
}

const cliConfig: CliGameConfig<BourreResponse, [ApiArgs]> = {
  gameName: 'bourre',
  parseCommand: parseBourreCommand,
  formatResponse: formatBourreState,
  helpText: [
    'Commands:',
    '  d [0-1]      Decide (1=play, 0=fold)',
    '  dr [idx...]  Discard & draw',
    '  p [idx]      Play a card',
    '  n            Next hand',
    '  r            Reset',
    '  sd [0-2]     Set difficulty',
    '  log          Action log',
  ],
};

/** Bourré game page content. */
function BourrePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('bourre');
  const {
    state,
    loading,
    error,
    exec: apiCall,
    retry,
  } = useGameApi<BourreResponse, [ApiArgs]>((...args) => bourreApi.exec(...args));
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bourre');
  const { handleCommand } = useCliGame<BourreResponse, [ApiArgs]>(apiCall, cliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  const [selectedCards, setSelectedCards] = useState<Set<number>>(new Set());

  // biome-ignore lint/correctness/useExhaustiveDependencies: mount-only reset
  useEffect(() => {
    void apiCall({ command: 'reset' });
  }, []);

  const phase = state?.phase ?? '';
  const isGameEnd = state?.gameEndFlag ?? false;
  const humanIdx = state?.players?.findIndex((p) => p.isHuman) ?? 0;
  const humanPlayer = state?.players?.[humanIdx];
  const isHumanTurn = state ? state.currentPlayerIdx === humanIdx : false;
  const humanWon = isGameEnd && !!state && state.winnerIdx === humanIdx;
  const validPlays = useMemo(() => new Set(state?.validPlays ?? []), [state]);

  useEffect(() => {
    if (humanWon) playSound('winFanfare');
  }, [humanWon, playSound]);

  const handleDecide = useCallback(
    (play: boolean) => {
      void apiCall({ command: 'decide', decide: play });
    },
    [apiCall],
  );

  const handleDraw = useCallback(() => {
    const indices = Array.from(selectedCards).sort((a, b) => a - b);
    void apiCall({ command: 'draw', indices });
    setSelectedCards(new Set());
  }, [apiCall, selectedCards]);

  const handleKeepAll = useCallback(() => {
    void apiCall({ command: 'draw', indices: [] });
    setSelectedCards(new Set());
  }, [apiCall]);

  const handlePlayCard = useCallback(
    (idx: number) => {
      void apiCall({ command: 'p', cardIndex: idx });
    },
    [apiCall],
  );

  const handleNext = useCallback(() => {
    void apiCall({ command: 'next' });
  }, [apiCall]);
  const handleReset = useCallback(() => {
    void apiCall({ command: 'reset' });
  }, [apiCall]);

  const toggleCard = useCallback((idx: number) => {
    setSelectedCards((prev) => {
      const next = new Set(prev);
      if (next.has(idx)) next.delete(idx);
      else next.add(idx);
      return next;
    });
  }, []);

  const currentPlayerIdx = state?.currentPlayerIdx;
  // biome-ignore lint/correctness/useExhaustiveDependencies: clear selection on turn/phase change
  useEffect(() => {
    setSelectedCards(new Set());
  }, [currentPlayerIdx, phase]);

  const phaseName = useMemo(() => {
    if (!state) return '';
    return t(`phase.${phase}`);
  }, [state, phase, t]);

  const playerStatus = (p: BourreResponse['players'][number]): string => {
    if (p.isFinished) return t('label.out');
    if (p.folded) return t('label.folded');
    if (p.bourreed) return t('label.bourreed');
    return `${p.tricks} ${t('label.tricks')}`;
  };

  if (!state) return <GameSkeleton gameKey="bourre" layout={{ kind: 'card-grid', count: 5, cols: 'grid-cols-5' }} />;

  const trick = state.currentTrick.length > 0 ? state.currentTrick : state.lastTrick;

  return (
    <GamePageShell
      title={tc('nav.bourre')}
      gameThemeBg={gameTheme.bourre.bg}
      phaseName={phaseName}
      gamePath="/bourre"
      isHumanTurn={isHumanTurn && !isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      onCelebrate={() => playSound('winFanfare')}
      headerExtra={
        <div className="flex items-center gap-2">
          <span className="text-xs opacity-75">
            {t('label.pot')}: {state.pot}
            {state.carryPot > 0 ? ` (+${state.carryPot})` : ''}
          </span>
          <span className="text-xs opacity-75" data-testid="bourre-trump">
            {t('label.trump')}:{' '}
            {isSuitDesign(state.trumpSuit) ? (
              <span className={isRedSuitDesign(state.trumpSuit) ? 'text-ds-error' : undefined}>
                {suitSymbol(state.trumpSuit)}
              </span>
            ) : (
              '-'
            )}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </div>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <div className="flex flex-col gap-3 p-3 overflow-y-auto">
          <LandscapeBanner message={t('landscapeBanner')} />
          {error && <ErrorAlert message={error} onRetry={retry} />}
          <GameMessageBox messageCode={state.messageCode} messageParams={state.messageParams} message={state.message} />

          {/* CPU areas */}
          <div className="flex flex-wrap justify-around gap-2" data-tutorial="bourre-cpus">
            {state.players
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div key={p.id} className="text-center text-ds-text-primary text-sm">
                  <div className="font-bold">
                    {playerName(p.id, p.isHuman)}
                    {p.id === state.dealerIdx ? ` (${t('label.dealer')})` : ''}
                  </div>
                  <div>
                    {p.chips} {t('label.chips')}
                  </div>
                  <div className="text-xs opacity-75">{playerStatus(p)}</div>
                </div>
              ))}
          </div>

          {/* Trick / table */}
          <div className="flex flex-col items-center gap-1 min-h-[80px]" data-tutorial="bourre-table">
            <div className="flex justify-center items-center gap-1">
              {trick.length > 0 ? (
                trick.map((tcd) =>
                  tcd.card ? (
                    <div key={`trick-${tcd.playerIdx}-${tcd.card.design}-${tcd.card.value}`} className="text-center">
                      <AnimatedCard card={tcd.card} width={cardWidth * 0.8} />
                      <div className="text-xs text-ds-text-secondary">
                        {playerName(tcd.playerIdx, tcd.playerIdx === humanIdx)}
                      </div>
                    </div>
                  ) : null,
                )
              ) : (
                <span className="text-ds-text-secondary text-sm">{t('label.table')}: ---</span>
              )}
            </div>
          </div>

          {/* Phase-specific controls */}
          <div className="flex justify-center gap-2" data-tutorial="bourre-controls">
            {phase === 'decide' && isHumanTurn && (
              <>
                <button type="button" className={btnPrimary} onClick={() => handleDecide(true)}>
                  {t('button.play')}
                </button>
                <button type="button" className={btnSecondary} onClick={() => handleDecide(false)}>
                  {t('button.fold')}
                </button>
              </>
            )}
            {phase === 'draw' && isHumanTurn && (
              <>
                <button type="button" className={btnPrimary} onClick={handleDraw} disabled={selectedCards.size === 0}>
                  {t('button.discard')} ({selectedCards.size})
                </button>
                <button type="button" className={btnSecondary} onClick={handleKeepAll}>
                  {t('button.keepAll')}
                </button>
              </>
            )}
            {phase === 'roundEnd' && (
              <button type="button" className={btnWarning} onClick={handleNext}>
                {t('button.next')}
              </button>
            )}
          </div>

          {phase === 'decide' && isHumanTurn && (
            <p
              data-testid="bourre-decide-summary"
              title={t('decideSummaryHelp', {
                penalty: state.pot + state.carryPot,
              })}
              className={`mt-1 text-center text-xs ${
                state.pot + state.carryPot >= BOURRE_PENALTY_WARN_THRESHOLD
                  ? 'text-ds-warning font-medium'
                  : 'text-ds-text-muted'
              }`}
            >
              {t('decideSummary', {
                pot: state.pot,
                penalty: state.pot + state.carryPot,
              })}
            </p>
          )}

          {/* Human hand */}
          {humanPlayer && !humanPlayer.isFinished && humanPlayer.cards.length > 0 && (
            <div data-tutorial="bourre-hand">
              <div className="flex flex-wrap justify-center gap-1">
                {humanPlayer.cards.map((c, i) => {
                  const selectable =
                    (phase === 'draw' && isHumanTurn) || (phase === 'play' && isHumanTurn && validPlays.has(i));
                  const isSelected = selectedCards.has(i);
                  return (
                    <button
                      key={`hand-${c.design}-${c.value}-${i}`}
                      type="button"
                      className={
                        isSelected
                          ? 'transition-transform -translate-y-2 ring-2 ring-ds-warning rounded'
                          : 'transition-transform'
                      }
                      onClick={() => {
                        if (phase === 'draw' && isHumanTurn) toggleCard(i);
                        else if (phase === 'play' && isHumanTurn && validPlays.has(i)) handlePlayCard(i);
                      }}
                      disabled={!selectable}
                    >
                      <AnimatedCard card={c} width={cardWidth * 0.9} />
                    </button>
                  );
                })}
              </div>
            </div>
          )}

          {/* Hand / game result */}
          {(phase === 'roundEnd' || isGameEnd) && state.results.length > 0 && (
            <div className="text-center text-ds-text-primary space-y-1">
              {isGameEnd && (
                <div className="text-lg font-bold">
                  {humanWon
                    ? t('result.youWin')
                    : t('result.youLose', {
                        name: playerName(state.winnerIdx, state.winnerIdx === humanIdx),
                      })}
                </div>
              )}
              {state.results.map((r) => (
                <div key={`result-${r.playerIdx}`} className="text-sm">
                  {playerName(r.playerIdx, r.playerIdx === humanIdx)}: {r.tricks} {t('label.tricks')}
                  {r.folded ? ` (${t('label.folded')})` : ''}
                  {r.bourreed ? ` (${t('label.bourreed')})` : ''}
                  {r.wonAmount > 0 ? ` +${r.wonAmount}` : ''}
                </div>
              ))}
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
      <GameFooter className={gameTheme.bourre.footer}>
        <GameResetButton
          isGameEnd={isGameEnd}
          onReset={handleReset}
          requestConfirm={requestConfirm}
          loading={loading}
          dataTutorial="bourre-reset"
        />
      </GameFooter>
    </GamePageShell>
  );
}

/** Bourré page with tutorial wrapper. */
export const BourrePage = withTutorial(BourrePageContent, 'bourre', TUTORIAL_STEPS);
