import { useMemo } from 'react';
import type { goFishApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { GoFishBooksDisplay } from '../components/gofish/GoFishBooksDisplay';
import { GoFishPlayerArea } from '../components/gofish/GoFishPlayerArea';
import { GoFishSettingsDialog } from '../components/gofish/GoFishSettingsDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { GoFishSkeleton } from '../components/skeleton/GoFishSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGoFishGame } from '../hooks/useGoFishGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnOutline, btnPrimary } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GoFishResponse } from '../types/card';
import { GoFishPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { valueName } from '../utils/cardUtils';
import { GOFISH_HELP, parseGofishCommand } from '../utils/cli/commands/gofishCommands';
import { formatGofishState } from '../utils/cli/formatters/gofishFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Go Fish tutorial step definitions. */
const GF_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="gf-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gf-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gf-ask-button"]',
    messageKey: 'tutorial.askButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gf-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const GOFISH_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GoFishPhase.PLAY]: 'play',
  [GoFishPhase.GAME_END]: 'end',
};

/** Renders the Go Fish game page. */
export function GoFishPage() {
  return (
    <TutorialWrapper gameName="gofish" steps={GF_TUTORIAL_STEPS}>
      <GoFishPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Go Fish page, wrapped by TutorialProvider. */
function GoFishPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('gofish');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    goFishConfig,
    handleConfigChange,
    selectedTarget,
    selectedRank,
    handleSelectTarget,
    handleSelectRank,
    handleAsk,
  } = useGoFishGame();
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('gofish', state);

  const phaseNames = usePhaseNames('gofish', GOFISH_PHASE_KEYS);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('gofish');
  const cliConfig: CliGameConfig<GoFishResponse, Parameters<typeof goFishApi.exec>> = useMemo(
    () => ({
      gameName: 'gofish',
      parseCommand: parseGofishCommand,
      formatResponse: formatGofishState,
      helpText: GOFISH_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state) return <GoFishSkeleton />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);

  const isPlayPhase = state.phase === GoFishPhase.PLAY;
  const isGameEnd = state.phase === GoFishPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentTurn]?.isHuman === true;
  const canAsk = isHumanTurn && selectedTarget !== null && selectedRank !== null && !loading;

  // Get unique ranks from human hand for rank selection
  const humanRanks = humanPlayer ? [...new Set(humanPlayer.cards.map((c) => c.value))].sort((a, b) => a - b) : [];

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.gofish.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.gofish')} />
      <PhaseIndicator phaseName={phaseNames[state.phase]} isHumanTurn={isHumanTurn}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/gofish" />
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <GoFishSettingsDialog
            cpuDifficulty={goFishConfig.cpuDifficulty}
            onCpuDifficultyChange={(v) => handleConfigChange('cpuDifficulty', v)}
          />

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Turn, deck info & hint toggle */}
            <div className="text-white text-center mb-2 flex items-center justify-center gap-4">
              <span>{t('deck', { count: state.deckRemaining })}</span>
              <label className="inline-flex items-center gap-1 text-sm cursor-pointer">
                <input
                  type="checkbox"
                  checked={frontendHintEnabled}
                  onChange={(e) => setFrontendHintEnabled(e.target.checked)}
                />
                {tc('hint.toggle', { ns: 'tutorial' })}
              </label>
            </div>

            {/* CPU player areas */}
            <div data-tutorial="gf-cpu-area">
              {cpuPlayers.map((p) => (
                <GoFishPlayerArea
                  key={p.id}
                  player={p}
                  isSelected={selectedTarget === p.id}
                  onSelect={handleSelectTarget}
                  disabled={!isHumanTurn || loading}
                />
              ))}
            </div>

            {/* Rank selector */}
            {isHumanTurn && humanRanks.length > 0 && (
              <div className="my-3">
                <div className="text-ds-text-muted text-sm mb-1">{t('selectRank')}</div>
                <div className="flex flex-wrap gap-2">
                  {humanRanks.map((rank) => (
                    <button
                      key={rank}
                      type="button"
                      onClick={() => handleSelectRank(rank)}
                      className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                        selectedRank === rank
                          ? 'bg-ds-warning text-ds-text-on-accent'
                          : 'bg-white/10 text-ds-text-primary hover:bg-white/20'
                      }`}
                      aria-pressed={selectedRank === rank}
                    >
                      {valueName(rank)}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Human books display */}
            {humanPlayer && <GoFishBooksDisplay books={humanPlayer.books} />}

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.gofish.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="gf-player-hand">
                {humanPlayer.cards.map((card) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}`}
                    onClick={() => handleSelectRank(card.value)}
                    aria-label={cardAlt(card)}
                    aria-pressed={selectedRank === card.value}
                    className={`transition-transform ${focusRingCard}`}
                    style={{
                      background: 'none',
                      padding: 0,
                      borderRadius: 8,
                      ...selectedCardStyle(selectedRank === card.value),
                      boxSizing: 'border-box',
                    }}
                  >
                    <AnimatedCard
                      card={card}
                      width={cardWidth}
                      onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                    />
                  </button>
                ))}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <div className="flex gap-2 items-center" data-tutorial="gf-ask-button">
              {isHumanTurn && (
                <button type="button" className={btnPrimary} onClick={handleAsk} disabled={!canAsk}>
                  {t('button.ask')}
                </button>
              )}
              <button
                type="button"
                className={btnOutline}
                data-tutorial="gf-reset-button"
                onClick={() =>
                  requestConfirm(() => {
                    hideActionLog();
                    return exec('reset', undefined, undefined, {
                      cpuDifficulty: goFishConfig.cpuDifficulty,
                    });
                  })
                }
                disabled={loading}
              >
                {tc('button.reset')}
              </button>
            </div>
          </GameFooter>
        </>
      )}
      <WinCelebration show={!!state?.gameEndFlag} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
