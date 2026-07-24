import { useEffect, useMemo } from 'react';
import type { cegoApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CEGO_KEEP_COUNT, CPU_DIFFICULTY_OPTIONS, TARGET_DEALS_OPTIONS, useCegoGame } from '../hooks/useCegoGame';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CegoResponse } from '../types/card';
import { CegoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cegoExchangeGuide } from '../utils/cegoExchangeGuide';
import { CEGO_HELP, parseCegoCommand } from '../utils/cli/commands/cegoCommands';
import { formatCegoState } from '../utils/cli/formatters/cegoFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Cego tutorial step definitions. */
const CEGO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cego-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cego-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cego-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="cego-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="cego-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const CEGO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CegoPhase.BID]: 'bid',
  [CegoPhase.CONTRACT]: 'contract',
  [CegoPhase.EXCHANGE]: 'exchange',
  [CegoPhase.PLAY]: 'play',
  [CegoPhase.TRICK_END]: 'trickEnd',
  [CegoPhase.ROUND_END]: 'roundEnd',
  [CegoPhase.GAME_END]: 'gameEnd',
};

/** Contract-type i18n keys indexed by contractType value (0=none, 1=Cego, 2=Handspiel). */
const CONTRACT_KEYS = ['contractNone', 'contractCego', 'contractHandspiel'] as const;

/** Outcome i18n keys indexed by outcome value (0=none, 1=Win/made, 2=Loss/failed). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeWin', 'outcomeLoss'] as const;

/** Renders the Cego (チェゴ) game page: a 4-player 54-card Baden tarock trick-taker with a single-step auction, a Cego/Handspiel contract choice, and a Cego exchange (keep 1 card, take the hidden blind). */
export const CegoPage = withTutorial(CegoPageContent, 'cego', CEGO_TUTORIAL_STEPS);

/** Inner content of the Cego page, wrapped by TutorialProvider. */
function CegoPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('cego');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    cegoConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handlePass,
    handleContract,
    handleExchange,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useCegoGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cego');
  const cegoCliConfig: CliGameConfig<CegoResponse, Parameters<typeof cegoApi.exec>> = useMemo(
    () => ({
      gameName: 'cego',
      parseCommand: parseCegoCommand,
      formatResponse: formatCegoState,
      helpText: CEGO_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cegoCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('cego', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('cego', CEGO_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="cego" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 11 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === CegoPhase.BID;
  const isContractPhase = state.phase === CegoPhase.CONTRACT;
  const isExchangePhase = state.phase === CegoPhase.EXCHANGE;
  const isPlayPhase = state.phase === CegoPhase.PLAY;
  const isTrickEnd = state.phase === CegoPhase.TRICK_END;
  const isRoundEnd = state.phase === CegoPhase.ROUND_END;
  const isGameEnd = state.phase === CegoPhase.GAME_END || state.gameEndFlag;

  const canBid = isBidPhase && state.isHumanBidTurn;
  const canContract = isContractPhase && state.isHumanContract;
  const canExchange = isExchangePhase && state.isHumanExchange;
  const canPlay = isPlayPhase && isHumanTurn;

  const contractLabel = t(CONTRACT_KEYS[state.contractType] ?? 'contractNone');

  // During play the hand is restricted to the legal indices; during the Cego
  // exchange any single card may be kept, so no restriction is applied.
  const handValidIndices = canPlay ? state.playableIndices : undefined;

  // Stepper guidance for the Cego exchange: pick 1 card to keep, then take the
  // blind. Derived purely from the current selection count — no backend state.
  const exchangeGuide = cegoExchangeGuide(selectedCardIndices.length, CEGO_KEEP_COUNT, humanPlayer?.cardCount ?? 0);

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.cego')}
      gameThemeBg={gameTheme.cego.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/cego"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerPlayer === humanIdx}
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
                    value: cegoConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetDeals',
                    label: t('settings.targetDeals'),
                    value: cegoConfig.targetDeals,
                    options: TARGET_DEALS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetDeals', v),
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
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>{t('contract', { contract: contractLabel })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="cego-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="cego-info">
                {/* Blind (Cego) — count only; the contents stay hidden. */}
                {state.blindCount > 0 && (
                  <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="cego-blind">
                    {t('blind', { count: state.blindCount })}
                  </div>
                )}

                {/* Per-player scores with declarer badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                      </span>
                      {p.isDeclarer && (
                        <span className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs">
                          {t('declarerBadge')}
                        </span>
                      )}
                    </div>
                  ))}
                </div>

                {/* Players: cards / tricks / captured points */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('tricks', { count: p.trickCount })} | {t('points', { points: p.cardPoints })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('tricks', { count: p.trickCount })} | {t('points', { points: p.cardPoints })}
                      </div>
                    ))}
                  </div>
                )}

                {/* Round result: the deal outcome (contract made / failed) */}
                {(isRoundEnd || isGameEnd) && state.outcome > 0 && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="cego-result">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.outcome', { outcome: t(OUTCOME_KEYS[state.outcome] ?? 'outcomeNone') })}</div>
                    {state.declarerIdx >= 0 && (
                      <>
                        <div>
                          {t('roundResult.declarer', {
                            name: playerName(state.declarerIdx, state.declarerIdx === humanIdx),
                          })}
                        </div>
                        <div>
                          {t('roundResult.captured', {
                            points: state.players[state.declarerIdx]?.cardPoints ?? 0,
                          })}
                        </div>
                      </>
                    )}
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
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.cego.footer} px-4 py-2.5`}>
            {canBid && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="cego-bid-prompt">
                {t('bidPhase')}
              </div>
            )}
            {canContract && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="cego-contract-prompt">
                {t('contractPhase')}
              </div>
            )}
            {canContract && (
              <div
                className="mb-2 mx-auto max-w-xl p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                data-testid="cego-contract-explainer"
              >
                <div className="mb-1 text-ds-text-primary font-semibold">{t('contractExplainTitle')}</div>
                <div className="py-0.5">{t('contractCegoDesc', { count: state.blindCount })}</div>
                <div className="py-0.5">{t('contractHandspielDesc')}</div>
              </div>
            )}
            {canExchange && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="cego-exchange-prompt">
                {t('exchangePhase', { count: selectedCardIndices.length })}
              </div>
            )}
            {canExchange && (
              <div
                className="mb-2 mx-auto max-w-xl p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                data-testid="cego-exchange-guide"
              >
                <div className="mb-1 text-ds-text-primary font-semibold">
                  {t('exchangeGuide.title', { step: exchangeGuide.currentStep, total: exchangeGuide.totalSteps })}
                </div>
                <ol className="space-y-0.5">
                  <li
                    className={exchangeGuide.currentStep === 1 ? 'text-ds-accent font-semibold' : ''}
                    data-testid="cego-exchange-step-1"
                    aria-current={exchangeGuide.currentStep === 1 ? 'step' : undefined}
                  >
                    {t('exchangeGuide.step1', { keep: CEGO_KEEP_COUNT, remaining: exchangeGuide.remaining })}
                  </li>
                  <li
                    className={exchangeGuide.currentStep === 2 ? 'text-ds-accent font-semibold' : ''}
                    data-testid="cego-exchange-step-2"
                    aria-current={exchangeGuide.currentStep === 2 ? 'step' : undefined}
                  >
                    {t('exchangeGuide.step2', { layDown: exchangeGuide.layDownCount, count: state.blindCount })}
                  </li>
                </ol>
                <div className="mt-1" data-testid="cego-exchange-status">
                  {exchangeGuide.ready
                    ? t('exchangeGuide.ready')
                    : t('exchangeGuide.remaining', { remaining: exchangeGuide.remaining })}
                </div>
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="cego"
                validIndices={handValidIndices}
                restrictedTooltip={canExchange ? t('exchangeKeep') : t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices &&
                  state.hint.cardIndices.length > 0 &&
                  ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="cego-action-buttons">
              {canBid && (
                <>
                  <button type="button" className={btnSecondary} onClick={handlePass} disabled={loading}>
                    {t('bidPass')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleBid} disabled={loading}>
                    {t('bidPlay')}
                  </button>
                </>
              )}
              {canContract && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleContract('cego')}
                    disabled={loading}
                  >
                    {t('contractCego')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={() => handleContract('handspiel')}
                    disabled={loading}
                  >
                    {t('contractHandspiel')}
                  </button>
                </>
              )}
              {canExchange && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleExchange}
                  disabled={loading || selectedCardIndices.length !== CEGO_KEEP_COUNT}
                >
                  {t('exchangeButton', { count: selectedCardIndices.length })}
                </button>
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
              {isTrickEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              )}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="cego-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
