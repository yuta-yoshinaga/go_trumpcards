import { useEffect, useMemo } from 'react';
import type { bouillotteApi } from '../api/gameApi';
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
import { HintTooltip } from '../components/hint/HintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import {
  ANTE_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  STARTING_CHIPS_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  useBouillotteGame,
} from '../hooks/useBouillotteGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BouillotteResponse } from '../types/card';
import { BouillottePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { computeBouillottePotOdds } from '../utils/bouillottePotOdds';
import { analyzeRetourneMatch } from '../utils/bouillotteRetourne';
import { BOUILLOTTE_HELP, parseBouillotteCommand } from '../utils/cli/commands/bouillotteCommands';
import { formatBouillotteState } from '../utils/cli/formatters/bouillotteFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Bouillotte tutorial step definitions. */
const BOUILLOTTE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bouillotte-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bouillotte-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bouillotte-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bouillotte-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BOUILLOTTE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BouillottePhase.BETTING]: 'betting',
  [BouillottePhase.RESULT]: 'result',
};

/** Renders the Bouillotte game page: an 18th-century French 3-card pot-vying game. */
export const BouillottePage = withTutorial(BouillottePageContent, 'bouillotte', BOUILLOTTE_TUTORIAL_STEPS);

/** Inner content of the Bouillotte page, wrapped by TutorialProvider. */
function BouillottePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('bouillotte');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    bouillotteConfig,
    handleConfigChange,
    reset,
    handleCall,
    handleRaise,
    handleFold,
    handleNextRound,
  } = useBouillotteGame();

  // Fetch a fresh game on mount (applies the current config).
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bouillotte');
  const cliConfig: CliGameConfig<BouillotteResponse, Parameters<typeof bouillotteApi.exec>> = useMemo(
    () => ({
      gameName: 'bouillotte',
      parseCommand: parseBouillotteCommand,
      formatResponse: formatBouillotteState,
      helpText: BOUILLOTTE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('bouillotte', state);
  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('bouillotte', BOUILLOTTE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="bouillotte" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 3 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  // Pot odds and chip costs facing the human at the Call/Raise/Fold decision.
  const humanRoundBet = humanPlayer?.roundBet ?? 0;
  const potOdds = computeBouillottePotOdds(state.pot, state.currentBet, humanRoundBet);
  const raiseCost = Math.max(0, state.currentBet + state.ante - humanRoundBet);

  // Which of the human's cards share the retourne's rank, and any combo it completes.
  const retourneMatch = analyzeRetourneMatch(humanPlayer?.cards ?? [], state.retourne);
  const matchingSet = new Set(retourneMatch.matchingIndices);

  const isBettingPhase = state.phase === BouillottePhase.BETTING;
  const isResultPhase = state.phase === BouillottePhase.RESULT;
  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;
  const humanWonMatch = state.matchWinnerIdx >= 0 && (state.players[state.matchWinnerIdx]?.isHuman ?? false);

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  const handName = (key?: string): string => (key ? t(`hand.${key.toLowerCase()}`, { defaultValue: key }) : '');

  const playerBadge = (p: BouillotteResponse['players'][number]): string =>
    p.out ? t('badge.out') : p.isWinner ? t('badge.winner') : p.folded ? t('badge.folded') : t('badge.active');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.bouillotte')}
      gameThemeBg={gameTheme.bouillotte.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isBettingPhase && isHumanTurn && !isGameEnd}
      gamePath="/bouillotte"
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
                    value: bouillotteConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'ante',
                    label: t('settings.ante'),
                    value: bouillotteConfig.ante,
                    options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('ante', v),
                  },
                  {
                    type: 'select',
                    id: 'startingChips',
                    label: t('settings.startingChips'),
                    value: bouillotteConfig.startingChips,
                    options: STARTING_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('startingChips', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: bouillotteConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="bouillotte-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('pot', { amount: state.pot })}</span>
              <span className="mr-4">{t('ante', { amount: state.ante })}</span>
              <span>{t('currentBet', { amount: state.currentBet })}</span>
            </div>

            {/* Retourne (shared turned-up card) */}
            {state.retourne && (
              <div className="mb-2 flex flex-col items-center">
                <div className="text-ds-text-muted text-xs mb-0.5">{t('retourneLabel')}</div>
                <span
                  data-testid="retourne-card"
                  className={matchingSet.size > 0 ? 'inline-block rounded-md ring-2 ring-ds-accent' : 'inline-block'}
                >
                  <CardImage card={state.retourne} width={cardWidth} />
                </span>
              </div>
            )}

            {isBettingPhase && isHumanTurn && (
              <div className="text-ds-text-muted text-center mb-2 text-sm font-semibold">{t('bettingNotice')}</div>
            )}

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="bouillotte-players">
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
          <GameFooter className={`${gameTheme.bouillotte.footer} px-4 py-2.5`}>
            {humanPlayer && humanPlayer.cards.length > 0 ? (
              <div className="mb-2" data-tutorial="bouillotte-hand">
                <div className="text-ds-text-muted text-xs mb-0.5">
                  {t('handLabel')}
                  {humanPlayer.handName ? ` — ${handName(humanPlayer.handName)}` : ''}
                </div>
                <div className="flex gap-1">
                  {humanPlayer.cards.map((c, i) => (
                    <span
                      key={`human-${i}`}
                      data-testid={`hand-card-${i}`}
                      className={matchingSet.has(i) ? 'inline-block rounded-md ring-2 ring-ds-accent' : 'inline-block'}
                    >
                      <CardImage card={c} width={cardWidth} />
                    </span>
                  ))}
                </div>
                {retourneMatch.noteKey && (
                  <div className="mt-1 text-xs font-semibold text-ds-accent" data-testid="retourne-note">
                    {t(`retourneNote.${retourneMatch.noteKey}`)}
                  </div>
                )}
              </div>
            ) : (
              <div className="text-ds-text-muted text-sm mb-2" data-tutorial="bouillotte-hand">
                {t('handLabel')}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            {isBettingPhase && isHumanTurn && !isGameEnd && (
              <div className="mb-2 text-ds-text-muted text-sm" data-testid="bouillotte-pot-odds" aria-live="polite">
                {potOdds.isFree
                  ? t('potOdds.free')
                  : t('potOdds.value', {
                      call: potOdds.callAmount,
                      pot: state.pot,
                      percentage: potOdds.percentage,
                      ratioPot: potOdds.ratioPot,
                      ratioCall: potOdds.ratioCall,
                    })}
              </div>
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="bouillotte-action-buttons">
              {isBettingPhase && isHumanTurn && !isGameEnd && (
                <>
                  <button type="button" className={btnPrimary} onClick={handleCall} disabled={loading}>
                    {potOdds.isFree ? t('callButton') : t('callButtonAmount', { amount: potOdds.callAmount })}
                  </button>
                  {state.canRaise && (
                    <button type="button" className={btnSuccess} onClick={handleRaise} disabled={loading}>
                      {t('raiseButtonAmount', { amount: raiseCost })}
                    </button>
                  )}
                  <button type="button" className={btnDanger} onClick={handleFold} disabled={loading}>
                    {t('foldButton')}
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
                dataTutorial="bouillotte-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
