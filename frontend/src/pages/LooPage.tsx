import { useEffect, useMemo } from 'react';
import type { looApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { ANTE_OPTIONS, CPU_DIFFICULTY_OPTIONS, useLooGame } from '../hooks/useLooGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { LooResponse } from '../types/card';
import { LooPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { LOO_HELP, parseLooCommand } from '../utils/cli/commands/looCommands';
import { formatLooState } from '../utils/cli/formatters/looFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { computeLooPotRisk } from '../utils/looPotRisk';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; 0 = unset). */
const SUIT_SYMBOLS = ['-', '♠', '♣', '♥', '♦'] as const;

/** Loo tutorial step definitions. */
const LOO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="loo-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="loo-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="loo-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="loo-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="loo-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const LOO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [LooPhase.DECIDE]: 'decide',
  [LooPhase.PLAY]: 'play',
  [LooPhase.TRICK_END]: 'trickEnd',
  [LooPhase.ROUND_END]: 'roundEnd',
};

/** Renders the Loo (Lanterloo) game page: a 4-player 52-card pot-based gambling trick-taker. */
export const LooPage = withTutorial(LooPageContent, 'loo', LOO_TUTORIAL_STEPS);

/** Inner content of the Loo page, wrapped by TutorialProvider. */
function LooPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('loo');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    looConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleDecide,
    handlePlay,
    handleNextDeal,
  } = useLooGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('loo');
  const looCliConfig: CliGameConfig<LooResponse, Parameters<typeof looApi.exec>> = useMemo(
    () => ({
      gameName: 'loo',
      parseCommand: parseLooCommand,
      formatResponse: formatLooState,
      helpText: LOO_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, looCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('loo', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('loo', LOO_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="loo" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isDecidePhase = state.phase === LooPhase.DECIDE;
  const isPlayPhase = state.phase === LooPhase.PLAY;
  const isRoundEnd = state.phase === LooPhase.ROUND_END;

  const canDecide = isDecidePhase && state.decidePlayerIdx === humanIdx && isHumanTurn;
  const canPlay = isPlayPhase && isHumanTurn;

  const trumpSymbol = state.trumpSuit >= 1 ? (SUIT_SYMBOLS[state.trumpSuit] ?? '-') : '-';

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.loo')}
      gameThemeBg={gameTheme.loo.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/loo"
      gameEndFlag={false}
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
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: looConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'ante',
                    label: t('settings.ante'),
                    value: looConfig.ante,
                    options: ANTE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('ante', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('deal', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber, total: state.totalTricks })}</span>
              <span className="mr-4">{t('pot', { pot: state.pot })}</span>
              <span>{t('trump', { suit: trumpSymbol })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="loo-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="loo-info">
                {/* Per-player chip balances with a participation badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span>
                        {playerName(p.id, p.isHuman)}: {t('chips', { chips: p.chips })}
                      </span>
                      <span
                        className={`px-1.5 py-0.5 rounded text-xs ${
                          p.playing ? 'bg-ds-accent/30 text-ds-accent' : 'bg-black/40 text-ds-text-muted'
                        }`}
                        role="status"
                        aria-label={t('statusAria', {
                          name: playerName(p.id, p.isHuman),
                          status: p.playing ? t('statusPlay') : t('statusPass'),
                        })}
                        title={t('playRiskTooltip')}
                        data-testid={`loo-status-${p.id}`}
                      >
                        {/* Colour-independent icon (● play / ○ pass) alongside the label. */}
                        <span aria-hidden="true">{p.playing ? '● ' : '○ '}</span>
                        {p.playing ? t('statusPlay') : t('statusPass')}
                      </span>
                    </div>
                  ))}
                </div>

                {/* Players: cards / tricks */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('tricks', { count: p.trickCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('tricks', { count: p.trickCount })}
                      </div>
                    ))}
                  </div>
                )}

                {/* Deal result: per-player gained chips + looed players */}
                {isRoundEnd && state.lastDealDetail && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="loo-deal-result"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('dealResult.title')}</div>
                    {state.lastDealDetail.looed.length > 0 && (
                      <div className="text-ds-error mb-1">
                        {t('dealResult.looed', {
                          names: state.lastDealDetail.looed
                            .map((i) => playerName(i, state.players[i]?.isHuman ?? false))
                            .join(', '),
                        })}
                      </div>
                    )}
                    {state.players.map((p) => (
                      <div key={p.id}>
                        {t('dealResult.gained', {
                          name: playerName(p.id, p.isHuman),
                          chips: state.lastDealDetail?.gained[p.id] ?? 0,
                        })}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isRoundEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.loo.footer} px-4 py-2.5`}>
            {isDecidePhase && !canDecide && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="loo-decide-cpu">
                {t('decideCpu', { id: state.decidePlayerIdx })}
              </div>
            )}
            {canDecide && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="loo-decide-prompt">
                {t('decidePrompt')}
              </div>
            )}
            {canDecide &&
              (() => {
                const { pot, looPenalty, perTrick } = computeLooPotRisk(state.pot, state.potStart);
                return (
                  <div
                    className="mb-2 mx-auto max-w-md p-2 rounded bg-black/30 text-center text-sm"
                    data-testid="loo-pot-risk"
                  >
                    <div className="text-ds-text-muted mb-0.5">{t('potRisk.label')}</div>
                    <div className="text-ds-accent">{t('potRisk.win', { pot, perTrick })}</div>
                    <div className="text-ds-error">{t('potRisk.loss', { penalty: looPenalty })}</div>
                  </div>
                );
              })()}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="loo"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && isRequestedHint(state) && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices &&
                  state.hint.cardIndices.length > 0 &&
                  ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="loo-action-buttons">
              {canDecide && (
                <div className="flex flex-wrap gap-2" data-testid="loo-decide-buttons">
                  <button type="button" className={btnPrimary} onClick={() => handleDecide(true)} disabled={loading}>
                    {t('decidePlay')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={() => handleDecide(false)} disabled={loading}>
                    {t('decidePass')}
                  </button>
                </div>
              )}
              {canPlay && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('playButton')}
                </button>
              )}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextDeal} disabled={loading}>
                  {t('nextDeal')}
                </button>
              )}
              <GameResetButton
                isGameEnd={false}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="loo-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
