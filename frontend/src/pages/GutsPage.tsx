import { useEffect, useMemo } from 'react';
import type { gutsApi } from '../api/gameApi';
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
import {
  ANTE_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  STARTING_CHIPS_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  useGutsGame,
} from '../hooks/useGutsGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GutsResponse } from '../types/card';
import { GutsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { GUTS_HELP, parseGutsCommand } from '../utils/cli/commands/gutsCommands';
import { formatGutsState } from '../utils/cli/formatters/gutsFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { evaluateGutsGuide } from '../utils/gutsGuideUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Guts tutorial step definitions. */
const GUTS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="guts-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="guts-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="guts-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="guts-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const GUTS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GutsPhase.DECLARE]: 'declare',
  [GutsPhase.RESULT]: 'result',
};

/** Renders the Guts game page: a fast multi-player pot-vying gambling game. */
export const GutsPage = withTutorial(GutsPageContent, 'guts', GUTS_TUTORIAL_STEPS);

/** Inner content of the Guts page, wrapped by TutorialProvider. */
function GutsPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('guts');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    gutsConfig,
    handleConfigChange,
    reset,
    handleIn,
    handleOut,
    handleNextRound,
  } = useGutsGame();

  // Fetch a fresh game on mount (applies the current config).
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('guts');
  const cliConfig: CliGameConfig<GutsResponse, Parameters<typeof gutsApi.exec>> = useMemo(
    () => ({
      gameName: 'guts',
      parseCommand: parseGutsCommand,
      formatResponse: formatGutsState,
      helpText: GUTS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('guts', state);
  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('guts', GUTS_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="guts" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 2 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isDeclarePhase = state.phase === GutsPhase.DECLARE;
  const isResultPhase = state.phase === GutsPhase.RESULT;
  const isGameEnd = state.gameEndFlag;

  // Rough hand-name + win-chance readout shown while the human must call In/Out.
  const declareGuide =
    isDeclarePhase && !isGameEnd && humanPlayer && humanPlayer.cards.length > 0
      ? evaluateGutsGuide(humanPlayer.cards)
      : null;
  const humanWonMatch = state.matchWinnerIdx >= 0 && (state.players[state.matchWinnerIdx]?.isHuman ?? false);

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  const handName = (key?: string): string => (key ? t(`hand.${key.toLowerCase()}`, { defaultValue: key }) : '');

  const playerBadge = (p: GutsResponse['players'][number]): string =>
    p.out
      ? t('badge.out')
      : p.isWinner
        ? t('badge.winner')
        : p.isMatcher
          ? t('badge.matcher')
          : p.in
            ? t('badge.in')
            : t('badge.waiting');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.guts')}
      gameThemeBg={gameTheme.guts.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isDeclarePhase && !isGameEnd}
      gamePath="/guts"
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
                    value: gutsConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'ante',
                    label: t('settings.ante'),
                    value: gutsConfig.ante,
                    options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('ante', v),
                  },
                  {
                    type: 'select',
                    id: 'startingChips',
                    label: t('settings.startingChips'),
                    value: gutsConfig.startingChips,
                    options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('startingChips', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: gutsConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="guts-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('pot', { amount: state.pot })}</span>
              <span>{t('ante', { amount: state.ante })}</span>
            </div>

            {isDeclarePhase && (
              <div className="text-ds-text-muted text-center mb-2 text-sm font-semibold">{t('declareNotice')}</div>
            )}

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="guts-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className={`text-sm py-0.5 ${p.isWinner ? 'text-ds-success' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  {playerLabel(p.id, p.isHuman)} — {t('chips', { amount: p.chips })} ·{' '}
                  {t('roundBet', { amount: p.roundBet })} · [{playerBadge(p)}]
                  {p.handName ? ` · ${handName(p.handName)}` : ''}
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
                        {p.handName ? ` — ${handName(p.handName)}` : ''}
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

            {/* Round result */}
            {isResultPhase && state.winnerIdx >= 0 && (
              <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                <div>
                  {t('roundResult.winner', {
                    name: playerLabel(state.winnerIdx, state.winnerIdx === humanIdx),
                    pot: state.pot,
                  })}
                </div>
                {state.matchers.map((idx) => {
                  const matcher = state.players[idx];
                  if (!matcher) return null;
                  const amount = Math.max(0, matcher.roundBet - state.ante);
                  return (
                    <div
                      key={`matcher-${idx}`}
                      className="text-ds-error font-semibold"
                      data-testid="guts-matcher-payment"
                    >
                      {t('roundResult.matchPayment', {
                        name: playerLabel(idx, idx === humanIdx),
                        amount,
                      })}
                    </div>
                  );
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
          <GameFooter className={`${gameTheme.guts.footer} px-4 py-2.5`}>
            {humanPlayer && humanPlayer.cards.length > 0 ? (
              <div className="mb-2" data-tutorial="guts-hand">
                <div className="text-ds-text-muted text-xs mb-0.5">
                  {t('handLabel')}
                  {humanPlayer.handName ? ` — ${handName(humanPlayer.handName)}` : ''}
                </div>
                <div className="flex gap-1">
                  {humanPlayer.cards.map((c, i) => (
                    <CardImage key={`human-${i}`} card={c} width={cardWidth} />
                  ))}
                </div>
              </div>
            ) : (
              <div className="text-ds-text-muted text-sm mb-2" data-tutorial="guts-hand">
                {t('handLabel')}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {declareGuide && (
              <div className="mb-2 p-2 rounded bg-black/30 text-sm" data-testid="guts-declare-guide">
                <div className="text-ds-text-primary font-semibold">
                  {t('guide.handLabel')}: <span className="text-base">{handName(declareGuide.handKey)}</span>
                </div>
                <div className="text-ds-text-muted">
                  {t('guide.tierLabel')}:{' '}
                  <span data-testid="guts-guide-tier">{t(`guide.tier.${declareGuide.tier}`)}</span>
                </div>
                <div className="text-ds-error text-xs" data-testid="guts-guide-risk">
                  {t('guide.matchRisk', { pot: state.pot })}
                </div>
              </div>
            )}

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="guts-action-buttons">
              {isDeclarePhase && !isGameEnd && (
                <>
                  <button type="button" className={btnPrimary} onClick={handleIn} disabled={loading}>
                    {t('inButton')}
                  </button>
                  <button type="button" className={btnDanger} onClick={handleOut} disabled={loading}>
                    {t('outButton')}
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
                dataTutorial="guts-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
