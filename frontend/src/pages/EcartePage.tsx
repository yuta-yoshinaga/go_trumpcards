import { useEffect, useMemo } from 'react';
import type { ecarteApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useEcarteGame } from '../hooks/useEcarteGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { EcarteResponse } from '../types/card';
import { EcarteNegStep, EcartePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardLabel } from '../utils/cardUtils';
import { ECARTE_HELP, parseEcarteCommand } from '../utils/cli/commands/ecarteCommands';
import { formatEcarteState } from '../utils/cli/formatters/ecarteFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 = undeclared). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Écarté tutorial step definitions. */
const ECARTE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="ecarte-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="ecarte-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ecarte-player-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ecarte-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const ECARTE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [EcartePhase.EXCHANGE]: 'exchange',
  [EcartePhase.PLAY]: 'play',
  [EcartePhase.ROUND_END]: 'roundEnd',
  [EcartePhase.GAME_END]: 'gameEnd',
};

/**
 * Exchange-phase negotiation sub-step → i18n key, mirroring the CUI's
 * `ecarteNegPromptKey` grouping (both discard steps share one label) so the
 * Web UI names the current sub-step instead of just showing a generic banner (#2678).
 */
const ECARTE_NEG_STEP_KEYS: Readonly<Record<number, string>> = {
  [EcarteNegStep.ELDER_DECIDE]: 'negStep.elderDecide',
  [EcarteNegStep.DEALER_RESPOND]: 'negStep.dealerRespond',
  [EcarteNegStep.ELDER_DISCARD]: 'negStep.discard',
  [EcarteNegStep.DEALER_DISCARD]: 'negStep.discard',
};

/**
 * Exchange-phase sub-step → one-line "what do these options mean" helper key.
 * Surfaces Écarté's stakes (esp. the dealer's refusal-vulnerability rule) right
 * under the negStep label so first-time players understand the consequences (#3455).
 */
const ECARTE_NEG_HELP_KEYS: Readonly<Record<number, string>> = {
  [EcarteNegStep.ELDER_DECIDE]: 'negHelp.elderDecide',
  [EcarteNegStep.DEALER_RESPOND]: 'negHelp.dealerRespond',
  [EcarteNegStep.ELDER_DISCARD]: 'negHelp.discard',
  [EcarteNegStep.DEALER_DISCARD]: 'negHelp.discard',
};

/** Renders the Écarté game page: a 2-player French trick game with an Exchange phase. */
export const EcartePage = withTutorial(EcartePageContent, 'ecarte', ECARTE_TUTORIAL_STEPS);

/** Inner content of the Écarté page, wrapped by TutorialProvider. */
function EcartePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ecarte');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    ecarteConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handlePropose,
    handleStand,
    handleRespond,
    handleDiscard,
    handleNextRound,
  } = useEcarteGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('ecarte');
  const cliConfig: CliGameConfig<EcarteResponse, Parameters<typeof ecarteApi.exec>> = useMemo(
    () => ({
      gameName: 'ecarte',
      parseCommand: parseEcarteCommand,
      formatResponse: formatEcarteState,
      helpText: ECARTE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('ecarte', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('ecarte', ECARTE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="ecarte" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isExchangePhase = state.phase === EcartePhase.EXCHANGE;
  const isPlayPhase = state.phase === EcartePhase.PLAY;
  const isRoundEnd = state.phase === EcartePhase.ROUND_END;
  const isGameEnd = state.phase === EcartePhase.GAME_END || state.gameEndFlag;

  // The web contract carries no explicit turn flags, so derive them from the
  // current seat: it is the human's turn whenever currentPlayerIdx is the human.
  const isHumanCurrent = state.currentPlayerIdx === humanIdx;
  const isHumanPlayTurn = isPlayPhase && isHumanCurrent;
  const isHumanExchangeTurn = isExchangePhase && isHumanCurrent;
  const canPlay = isHumanPlayTurn;

  // Exchange-phase sub-step controls (only when it is the human's exchange turn).
  const isElderDecide = isHumanExchangeTurn && state.negStep === EcarteNegStep.ELDER_DECIDE;
  const isDealerRespond = isHumanExchangeTurn && state.negStep === EcarteNegStep.DEALER_RESPOND;
  const isDiscardStep =
    isHumanExchangeTurn &&
    (state.negStep === EcarteNegStep.ELDER_DISCARD || state.negStep === EcarteNegStep.DEALER_DISCARD);

  // Discard-step guards: block an empty selection (server-error otherwise) and a
  // selection larger than the remaining stock, surfacing the reason inline (#3454).
  const discardCount = selectedCardIndices.length;
  const discardExceedsStock = discardCount > state.stockRemaining;
  const discardDisabled = loading || discardCount === 0 || discardExceedsStock;
  const discardReasonKey =
    discardCount === 0 ? 'discardReasonEmpty' : discardExceedsStock ? 'discardReasonExceed' : null;

  const trumpSymbol = state.trumpSuit >= 1 && state.trumpSuit <= 4 ? SUIT_SYMBOLS[state.trumpSuit] : t('noTrump');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.ecarte')}
      gameThemeBg={gameTheme.ecarte.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={(isHumanPlayTurn || isHumanExchangeTurn) && !isGameEnd}
      gamePath="/ecarte"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerIdx === humanIdx}
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
                    value: ecarteConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: ecarteConfig.targetScore,
                    options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetScore', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="ecarte-info">
              <span className="mr-4">{t('deal', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span className="mr-4">{t('stock', { count: state.stockRemaining })}</span>
              <span>{t('target', { points: state.config.targetScore })}</span>
            </div>

            {isExchangePhase && (
              <div className="text-center mb-2">
                <div className="text-ds-text-muted text-sm font-semibold">{t('exchangeNotice')}</div>
                <div className="text-ds-accent text-xs mt-0.5" data-testid="ecarte-neg-step">
                  {t('negStepLabel', { step: t(ECARTE_NEG_STEP_KEYS[state.negStep]) })}
                </div>
                <div className="text-ds-text-muted text-xs mt-0.5" data-testid="ecarte-neg-help">
                  {t(ECARTE_NEG_HELP_KEYS[state.negStep])}
                </div>
              </div>
            )}

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="ecarte-trick-display"
                />
                {state.trumpCard && (
                  <div className="mt-2 text-center text-ds-text-muted text-sm">
                    {t('trumpCard', { card: cardLabel(state.trumpCard) })}
                  </div>
                )}
              </div>

              {/* Right: score sidebar */}
              <div>
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div className="mb-1 text-ds-text-primary text-xs">{t('matchScoreTitle')}</div>
                  {state.players.map((p) => (
                    <div key={p.id} className={p.isHuman ? 'text-ds-text-primary' : ''}>
                      {t('scoreLine', {
                        name: p.isHuman ? t('you') : t('cpu', { id: p.id }),
                        match: state.matchScore[p.id] ?? 0,
                        deal: state.dealPoints[p.id] ?? 0,
                        tricks: p.trickCount,
                      })}
                    </div>
                  ))}
                </div>

                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {state.players.map((p) => (
                      <div key={p.id}>
                        {t('roundResult.line', {
                          name: p.isHuman ? t('you') : t('cpu', { id: p.id }),
                          points: state.dealPoints[p.id] ?? 0,
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
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.ecarte.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="ecarte"
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && isRequestedHint(state) && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndex != null && ` ([${state.hint.cardIndex}])`}
                {/* **識別子をそのまま出さない。**`propose` のような英語が
                    日本語 UI に混ざる (#4727)。訳が無ければ識別子に落とす
                    (キー文字列は出さない)。 */}
                {state.hint.action != null &&
                  ` (${t(`action.${state.hint.action}`, { defaultValue: state.hint.action })})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="ecarte-action-buttons">
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
              {isElderDecide && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePropose}
                    disabled={loading}
                    title={t('consequence.propose')}
                    aria-describedby="ecarte-propose-desc"
                    data-testid="ecarte-propose"
                  >
                    {t('proposeButton')}
                  </button>
                  <span id="ecarte-propose-desc" className="sr-only">
                    {t('consequence.propose')}
                  </span>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleStand}
                    disabled={loading}
                    title={t('consequence.stand')}
                    aria-describedby="ecarte-stand-desc"
                    data-testid="ecarte-stand"
                  >
                    {t('standButton')}
                  </button>
                  <span id="ecarte-stand-desc" className="sr-only">
                    {t('consequence.stand')}
                  </span>
                </>
              )}
              {isDealerRespond && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleRespond(true)}
                    disabled={loading}
                    title={t('consequence.accept')}
                    aria-describedby="ecarte-accept-desc"
                    data-testid="ecarte-accept"
                  >
                    {t('acceptButton')}
                  </button>
                  <span id="ecarte-accept-desc" className="sr-only">
                    {t('consequence.accept')}
                  </span>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={() => handleRespond(false)}
                    disabled={loading}
                    title={t('consequence.refuse')}
                    aria-describedby="ecarte-refuse-desc"
                    data-testid="ecarte-refuse"
                  >
                    {t('refuseButton')}
                  </button>
                  <span id="ecarte-refuse-desc" className="sr-only">
                    {t('consequence.refuse')}
                  </span>
                </>
              )}
              {/* **損得説明が title と sr-only にしか無かった** (#5658)。タッチ端末は
                  hover が起きないので、目の見える利用者がこの情報に到達できない。
                  狭い画面でだけ本文として出す (デスクトップは従来どおりツールチップ)。 */}
              {(isElderDecide || isDealerRespond) && (
                <div
                  className="sm:hidden basis-full text-xs text-ds-text-muted mt-1 space-y-0.5"
                  data-testid="ecarte-consequences"
                >
                  {(isElderDecide ? (['propose', 'stand'] as const) : (['accept', 'refuse'] as const)).map((action) => (
                    <div key={action}>
                      {t(`${action}Button`)}: {t(`consequence.${action}`)}
                    </div>
                  ))}
                </div>
              )}
              {isDiscardStep && (
                <>
                  <span className="text-xs text-ds-text-muted self-center mr-1">{t('discardPrompt')}</span>
                  <span className="text-xs text-ds-text-primary self-center mr-1" data-testid="ecarte-discard-guide">
                    {t('discardSelectionGuide', { count: selectedCardIndices.length })}
                  </span>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDiscard}
                    disabled={discardDisabled}
                    data-testid="ecarte-discard"
                  >
                    {t('discardButton', { count: selectedCardIndices.length })}
                  </button>
                  {discardReasonKey && (
                    <span className="text-xs text-ds-warning self-center" data-testid="ecarte-discard-reason">
                      {t(discardReasonKey)}
                    </span>
                  )}
                </>
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
                dataTutorial="ecarte-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
