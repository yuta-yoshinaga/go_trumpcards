import { useEffect, useMemo } from 'react';
import type { sevenTwentySevenApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  ANTE_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  STARTING_CHIPS_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  useSevenTwentySevenGame,
} from '../hooks/useSevenTwentySevenGame';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SevenTwentySevenResponse } from '../types/card';
import { SevenTwentySevenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSevenTwentySevenCommand, SEVENTWENTYSEVEN_HELP } from '../utils/cli/commands/seventwentysevenCommands';
import { formatSevenTwentySevenState } from '../utils/cli/formatters/seventwentysevenFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** SevenTwentySeven tutorial step definitions. */
const SEVENTWENTYSEVEN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="s27-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="s27-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="s27-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="s27-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SEVENTWENTYSEVEN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SevenTwentySevenPhase.DRAW]: 'draw',
  [SevenTwentySevenPhase.RESULT]: 'result',
};

/** Renders the SevenTwentySeven game page: a fast multi-player pot-vying gambling game. */
export const SevenTwentySevenPage = withTutorial(
  SevenTwentySevenPageContent,
  'seventwentyseven',
  SEVENTWENTYSEVEN_TUTORIAL_STEPS,
);

/** Inner content of the SevenTwentySeven page, wrapped by TutorialProvider. */
function SevenTwentySevenPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('seventwentyseven');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    sevenTwentySevenConfig,
    handleConfigChange,
    reset,
    handleTakeCard,
    handleStand,
    handleNextRound,
  } = useSevenTwentySevenGame();

  // Fetch a fresh game on mount (applies the current config).
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('seventwentyseven');
  const cliConfig: CliGameConfig<SevenTwentySevenResponse, Parameters<typeof sevenTwentySevenApi.exec>> = useMemo(
    () => ({
      gameName: 'seventwentyseven',
      parseCommand: parseSevenTwentySevenCommand,
      formatResponse: formatSevenTwentySevenState,
      helpText: SEVENTWENTYSEVEN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('seventwentyseven', state);
  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('seventwentyseven', SEVENTWENTYSEVEN_PHASE_KEYS);

  if (!state)
    return (
      <GameSkeleton gameKey="seventwentyseven" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 2 }} />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isDrawPhase = state.phase === SevenTwentySevenPhase.DRAW;
  const isResultPhase = state.phase === SevenTwentySevenPhase.RESULT;
  const isGameEnd = state.gameEndFlag;
  const humanWonMatch = state.matchWinnerIdx >= 0 && (state.players[state.matchWinnerIdx]?.isHuman ?? false);
  // 止まったあとは打つ手が無い（サーバがラウンドを回し切る）。
  const canAct = isDrawPhase && !isGameEnd && humanPlayer !== undefined && !humanPlayer.standing;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  /** 「6.5 / 21」の形。超過した側はサーバが "-" を返す。 */
  const scoreLabel = (p: SevenTwentySevenResponse['players'][number]): string =>
    p.lowScore || p.highScore ? `${p.lowScore || '?'} / ${p.highScore || '?'}` : '';

  const playerBadge = (p: SevenTwentySevenResponse['players'][number]): string =>
    p.out
      ? t('badge.out')
      : p.wonLow && p.wonHigh
        ? t('badge.scoop')
        : p.wonLow
          ? t('badge.wonLow')
          : p.wonHigh
            ? t('badge.wonHigh')
            : p.standing
              ? t('badge.standing')
              : t('badge.drawing');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.seventwentyseven')}
      gameThemeBg={gameTheme.seventwentyseven.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={canAct}
      gamePath="/seventwentyseven"
      gameEndFlag={isGameEnd}
      winShow={isResultPhase && (humanWonMatch || state.result > 0)}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span>{t('chips', { amount: state.chips })}</span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'playerCount',
                    label: t('settings.playerCount'),
                    value: sevenTwentySevenConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'ante',
                    label: t('settings.ante'),
                    value: sevenTwentySevenConfig.ante,
                    options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('ante', v),
                  },
                  {
                    type: 'select',
                    id: 'startingChips',
                    label: t('settings.startingChips'),
                    value: sevenTwentySevenConfig.startingChips,
                    options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('startingChips', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: sevenTwentySevenConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="s27-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('pot', { amount: state.pot })}</span>
              <span>{t('ante', { amount: state.ante })}</span>
            </div>

            {/* **2 つの目標を常に書く。** 7 と 27 のどちらに寄せるかがこのゲーム
                そのもので、書いていなければ何を選んでいるのか読めない。 */}
            <div className="text-ds-text-muted text-center mb-2 text-sm" data-testid="s27-targets-note">
              {t('targetsNote')}
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="s27-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className={`text-sm py-0.5 ${p.wonLow || p.wonHigh ? 'text-ds-success' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''}`}
                  data-testid={`s27-player-${p.id}`}
                >
                  {playerLabel(p.id, p.isHuman)} — {t('chips', { amount: p.chips })} ·{' '}
                  {t('roundBet', { amount: p.roundBet })} · [{playerBadge(p)}]
                  {scoreLabel(p) ? ` · ${scoreLabel(p)}` : ''}
                </div>
              ))}
            </div>

            {/* Revealed hands at result */}
            {isResultPhase && (
              <div className="mb-2 p-2 rounded bg-black/30">
                {state.players
                  .filter((p) => !p.isHuman && p.cards.length > 0)
                  .map((p) => (
                    <div key={p.id} className="mb-1">
                      <div className="text-ds-text-muted text-xs mb-0.5">
                        {playerLabel(p.id, p.isHuman)}
                        {scoreLabel(p) ? ` — ${scoreLabel(p)}` : ''}
                      </div>
                      <div className="flex gap-1">
                        {p.cards.map((c, i) => (
                          <CardImage key={`${p.id}-${i}`} card={c} width={cardWidth} />
                        ))}
                      </div>
                    </div>
                  ))}
              </div>
            )}

            {/* **両側とも全滅ならポットは持ち越す。** ここに何も描かないと、
                チップが動いていないのに理由が画面から消える (#4847 と同じ形)。 */}
            {isResultPhase && state.lowWinner < 0 && state.highWinner < 0 && (
              <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="s27-carry-result">
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                <div className="text-ds-warning font-semibold">
                  {t('roundResult.carry', { pot: state.carryPot, count: state.carryCount })}
                </div>
              </div>
            )}

            {/* **両側の勝者を名指しする。** どちらを取ったのかが分からないと、
                なぜ半分なのか / なぜ総取りなのかが読めない。 */}
            {isResultPhase && (state.lowWinner >= 0 || state.highWinner >= 0) && (
              <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                {state.lowWinner >= 0 && state.lowWinner === state.highWinner ? (
                  <div className="text-ds-success font-semibold" data-testid="s27-scoop-result">
                    {t('roundResult.scoop', {
                      name: playerLabel(state.lowWinner, state.lowWinner === humanIdx),
                      pot: state.pot,
                    })}
                  </div>
                ) : (
                  <>
                    <div data-testid="s27-low-result">
                      {state.lowWinner >= 0
                        ? t('roundResult.lowWinner', {
                            name: playerLabel(state.lowWinner, state.lowWinner === humanIdx),
                            score: state.players[state.lowWinner]?.lowScore ?? '',
                          })
                        : t('roundResult.lowEmpty')}
                    </div>
                    <div data-testid="s27-high-result">
                      {state.highWinner >= 0
                        ? t('roundResult.highWinner', {
                            name: playerLabel(state.highWinner, state.highWinner === humanIdx),
                            score: state.players[state.highWinner]?.highScore ?? '',
                          })
                        : t('roundResult.highEmpty')}
                    </div>
                  </>
                )}
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
          <GameFooter className={`${gameTheme.seventwentyseven.footer} px-4 py-2.5`}>
            {humanPlayer && humanPlayer.cards.length > 0 ? (
              <div className="mb-2" data-tutorial="s27-hand">
                <div className="text-ds-text-muted text-xs mb-0.5" data-testid="s27-your-score">
                  {t('handLabel')}
                  {scoreLabel(humanPlayer) ? ` — ${t('scoreLabel')}: ${scoreLabel(humanPlayer)}` : ''}
                </div>
                <div className="flex gap-1">
                  {humanPlayer.cards.map((c, i) => (
                    <CardImage key={`human-${i}`} card={c} width={cardWidth} />
                  ))}
                </div>
              </div>
            ) : (
              <div className="text-ds-text-muted text-sm mb-2" data-tutorial="s27-hand">
                {t('handLabel')}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="s27-action-buttons">
              {canAct && (
                <>
                  <button type="button" className={btnPrimary} onClick={handleTakeCard} disabled={loading}>
                    {t('cardButton')}
                  </button>
                  <button type="button" className={btnDanger} onClick={handleStand} disabled={loading}>
                    {t('standButton')}
                  </button>
                </>
              )}

              {isResultPhase && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sevenTwentySeven-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
