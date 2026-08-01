import { useEffect, useMemo } from 'react';
import { guandanApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardSelection } from '../hooks/useCardSelection';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GuandanResponse } from '../types/card';
import { GuandanPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { GUANDAN_HELP, parseGuandanCommand } from '../utils/cli/commands/guandanCommands';
import { formatGuandanState } from '../utils/cli/formatters/guandanFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Combination names by wire code (sync: `GuandanComboKind`). */
const COMBO_KEYS: Readonly<Record<number, string>> = {
  1: 'comboSingle',
  2: 'comboPair',
  3: 'comboTriple',
  4: 'comboFullHouse',
  5: 'comboStraight',
  6: 'comboPlate',
  7: 'comboTube',
  8: 'comboBomb',
  9: 'comboStraightFlush',
  10: 'comboJokerBomb',
};

/** Guandan tutorial step definitions. */
const GUANDAN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="guandan-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="guandan-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="guandan-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="guandan-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="guandan-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const GUANDAN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GuandanPhase.TRIBUTE]: 'tribute',
  [GuandanPhase.PLAY]: 'play',
  [GuandanPhase.HAND_END]: 'handEnd',
  [GuandanPhase.GAME_END]: 'gameEnd',
};

/** Renders the Guandan (掼蛋) game page: a two-pack climbing game played at rising levels. */
export const GuandanPage = withTutorial(GuandanPageContent, 'guandan', GUANDAN_TUTORIAL_STEPS);

/** Inner content of the Guandan page, wrapped by TutorialProvider. */
function GuandanPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('guandan');
  const { state, loading, error, exec, retry } = useGameApi(guandanApi.exec);
  const { selected, toggle, clear } = useCardSelection();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('guandan');
  const cliConfig: CliGameConfig<GuandanResponse, Parameters<typeof guandanApi.exec>> = useMemo(
    () => ({
      gameName: 'guandan',
      parseCommand: parseGuandanCommand,
      formatResponse: formatGuandanState,
      helpText: GUANDAN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('guandan', GUANDAN_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="guandan" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 12 }} />;

  const human = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === GuandanPhase.GAME_END || state.gameEndFlag;
  const isHandEnd = !isGameEnd && state.phase === GuandanPhase.HAND_END;
  const isHumanTurn = !isGameEnd && !isHandEnd && state.currentPlayerIdx === 0;
  const humanWon = isGameEnd && state.winnerTeam === 0;
  // **還貢は「受け取った側」が返す。**貢の相手が自分でなければ操作は無い。
  const owesReturn = state.tributes.some((x) => x.to === 0 && x.returned === null);
  const isTribute = state.phase === GuandanPhase.TRIBUTE && isHumanTurn && owesReturn;
  const isPlay = state.phase === GuandanPhase.PLAY && isHumanTurn;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
  const suitLabel = (design: string): string => t(`suit.${design}`);
  const cardLabel = (c: { design: string; value: number } | null): string =>
    c ? `${suitLabel(c.design)}${c.value}` : '?';
  // **レベルは 2〜A。**数字のままでは J/Q/K/A が読めない。
  const levelLabel = (level: number): string => ({ 11: 'J', 12: 'Q', 13: 'K', 14: 'A' })[level] ?? String(level);
  const comboLabel = (kind: number): string => (COMBO_KEYS[kind] ? t(COMBO_KEYS[kind]) : '-');

  const handlePlay = () => {
    if (selected.length === 0) return;
    exec('play', { cardIndexes: [...selected].sort((a, b) => a - b) });
    clear();
  };

  const handlePass = () => {
    exec('pass');
    clear();
  };

  const handleTribute = () => {
    const idx = selected[0];
    if (idx === undefined) return;
    exec('tribute', { cardIndex: idx });
    clear();
  };

  const handleNext = () => {
    clear();
    exec('next');
  };

  const handleManualReset = () => {
    hideActionLog();
    clear();
    exec('reset');
  };

  return (
    <GamePageShell
      title={tc('nav.guandan')}
      gameThemeBg={gameTheme.guandan.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/guandan"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
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
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* The level is the whole game: those cards outrank aces, and the hearts are wild. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="guandan-info">
              <div data-tutorial="guandan-info">
                {t('scoreLine', {
                  hand: state.handNumber,
                  level: levelLabel(state.level),
                  t0: levelLabel(state.teamLevels[0]),
                  t1: levelLabel(state.teamLevels[1]),
                })}
              </div>
              <div className="text-xs text-ds-text-muted" data-testid="guandan-level-note">
                {t('levelNote', { level: levelLabel(state.level) })}
              </div>
              <div className="text-xs text-ds-text-muted" data-testid="guandan-advance-note">
                {t('advanceNote', {
                  first: state.advanceFirstSecond,
                  third: state.advanceFirstThird,
                  fourth: state.advanceFirstFourth,
                })}
              </div>
            </div>

            {/* Players — the finishing order drives both tribute and the climb. */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="guandan-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="guandan-player"
                  className={`text-sm py-0.5 flex items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{t('seat', { n: p.id })}</span>
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  <span>({t('team', { n: p.team })})</span>
                  <span>{t('cardCount', { count: p.cardCount })}</span>
                  {p.finishedRank > 0 && <span className="text-ds-accent">{t('outAt', { n: p.finishedRank })}</span>}
                  {p.isCurrentTurn && !isGameEnd && <span className="text-ds-accent">[{t('turnTag')}]</span>}
                </div>
              ))}
            </div>

            {/* The table. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="guandan-table">
              {state.lastCombo ? (
                <span>
                  {t('tableCombo', {
                    combo: comboLabel(state.lastCombo.kind),
                    size: state.lastCombo.size,
                    seat: state.lastPlayerIdx,
                  })}
                </span>
              ) : (
                <span>{t('tableEmpty')}</span>
              )}
            </div>

            {/* Tribute — the previous hand reaching into this one. */}
            {state.phase === GuandanPhase.TRIBUTE && (
              <div className="mb-2 p-2 rounded bg-black/20 text-xs" data-testid="guandan-tributes">
                <div className="mb-1 text-ds-text-primary">{t('tributeTitle')}</div>
                {state.tributeCancelled ? (
                  <div data-testid="guandan-tribute-cancelled">{t('tributeCancelled')}</div>
                ) : (
                  state.tributes.map((x) => (
                    <div key={`tribute-${x.from}-${x.to}`}>
                      {t(x.returned ? 'tributeDone' : 'tributePending', {
                        from: x.from,
                        to: x.to,
                        card: cardLabel(x.card),
                        back: cardLabel(x.returned),
                      })}
                    </div>
                  ))
                )}
              </div>
            )}

            {/* The settled hand. */}
            {isHandEnd && state.lastResult && (
              <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="guandan-hand-result">
                {t(state.lastResult.firstSecond ? 'handResultFirstSecond' : 'handResult', {
                  team: state.lastResult.winnerTeam,
                  advance: state.lastResult.advance,
                  level: levelLabel(state.teamLevels[state.lastResult.winnerTeam] ?? state.level),
                })}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.guandan.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="guandan-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-1">
                {(human?.cards ?? []).map((c, i) => (
                  <button
                    key={`hand-${c.design}-${c.value}-${i}`}
                    type="button"
                    onClick={() => (isPlay || isTribute) && toggle(i)}
                    disabled={!isPlay && !isTribute}
                    className={`rounded transition-all ${selected.includes(i) ? 'ring-2 ring-ds-info -translate-y-2' : ''} ${
                      isPlay || isTribute ? 'cursor-pointer hover:opacity-90' : 'cursor-default'
                    }`}
                    data-testid={`hand-card-${i}`}
                  >
                    <CardImage card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="guandan-actions">
              {isTribute && (
                <>
                  <span className="text-ds-text-muted text-xs" data-testid="guandan-tribute-rules">
                    {t('tributeRules')}
                  </span>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleTribute}
                    disabled={loading || selected.length !== 1}
                  >
                    {t('tributeButton')}
                  </button>
                </>
              )}

              {isPlay && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePlay}
                    disabled={loading || selected.length === 0}
                  >
                    {t('playButton')}
                  </button>
                  {/* **場が流れているときはパスできない。**リードは必ず何か出す。 */}
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={handlePass}
                    disabled={loading || state.lastCombo === null}
                  >
                    {t('passButton')}
                  </button>
                </>
              )}

              {isHandEnd && (
                <button type="button" className={btnPrimary} onClick={handleNext} disabled={loading}>
                  {t('nextButton')}
                </button>
              )}

              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon ? t('win') : t('lose')}
                </span>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="guandan-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
