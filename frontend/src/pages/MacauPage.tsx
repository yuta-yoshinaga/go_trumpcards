import { useCallback, useMemo } from 'react';
import type { macauApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useMacauGame } from '../hooks/useMacauGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MacauResponse } from '../types/card';
import { CrazyEightsSuit, MacauPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MACAU_HELP, parseMacauCommand } from '../utils/cli/commands/macauCommands';
import { formatMacauState } from '../utils/cli/formatters/macauFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

const MACAU_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MacauPhase.PLAY]: 'play',
  [MacauPhase.CHOOSE_SUIT]: 'chooseSuit',
  [MacauPhase.MUST_DECLARE]: 'mustDeclare',
  [MacauPhase.ROUND_END]: 'roundEnd',
  [MacauPhase.GAME_END]: 'gameEnd',
};

const SUIT_BUTTONS = [
  { suit: CrazyEightsSuit.SPADE, key: 'suitSpade' },
  { suit: CrazyEightsSuit.CLOVER, key: 'suitClover' },
  { suit: CrazyEightsSuit.HEART, key: 'suitHeart' },
  { suit: CrazyEightsSuit.DIAMOND, key: 'suitDiamond' },
] as const;

const SUIT_SYMBOLS: Record<number, string> = {
  [CrazyEightsSuit.SPADE]: '♠',
  [CrazyEightsSuit.CLOVER]: '♣',
  [CrazyEightsSuit.HEART]: '♥',
  [CrazyEightsSuit.DIAMOND]: '♦',
};

/** Macau tutorial step definitions. */
const MACAU_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="macau-discard-pile"]',
    messageKey: 'tutorial.discardPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="macau-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="macau-play-draw"]', messageKey: 'tutorial.playDraw', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="macau-magic"]',
    messageKey: 'tutorial.magicCards',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="macau-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Macau game page with magic-card play, suit selection, and declarations. */
export const MacauPage = withTutorial(MacauPageContent, 'macau', MACAU_TUTORIAL_STEPS);
/** Inner content of the Macau page, wrapped by TutorialProvider. */
function MacauPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('macau');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    macauConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleDraw,
    handleChooseSuit,
    handleDeclare,
    handleSkipDeclare,
    handleNextRound,
  } = useMacauGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('macau', state);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('macau');
  const cliConfig: CliGameConfig<MacauResponse, Parameters<typeof macauApi.exec>> = useMemo(
    () => ({
      gameName: 'macau',
      parseCommand: parseMacauCommand,
      formatResponse: formatMacauState,
      helpText: MACAU_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === MacauPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    handlePlay();
  }, [handlePlay]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, undefined, {
      cpuDifficulty: macauConfig.cpuDifficulty,
      pointLimit: macauConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, macauConfig.cpuDifficulty, macauConfig.pointLimit]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('macau', MACAU_PHASE_KEYS);

  if (!state)
    return (
      <GameSkeleton
        gameKey="macau"
        layout={{ kind: 'trick-taking', centerCard: true, trickArea: true, footerHandSize: 5 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === MacauPhase.PLAY;
  const isChooseSuit = state.phase === MacauPhase.CHOOSE_SUIT;
  const isMustDeclare = state.phase === MacauPhase.MUST_DECLARE;
  const isRoundEnd = state.phase === MacauPhase.ROUND_END;
  const isGameEnd = state.phase === MacauPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const hasPenalty = state.penaltyDrawCount > 0;
  const directionLabel = state.direction < 0 ? '←' : '→';

  return (
    <GamePageShell
      title={tc('nav.macau')}
      gameThemeBg={gameTheme.macau.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn || isChooseSuit || (isMustDeclare && state.players[state.currentPlayerIdx]?.isHuman)}
      gamePath="/macau"
      gameEndFlag={!!state.gameEndFlag}
      onCelebrate={() => playSound('winFanfare')}
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
                    value: macauConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'pointLimit',
                    label: t('settings.pointLimit'),
                    value: macauConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
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
              <span className="mr-4">{t('drawPile', { count: state.drawPileCount })}</span>
              <span>{t('direction', { dir: directionLabel })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Discard pile top */}
                {state.discardTop && (
                  <div
                    className="my-3 p-3 rounded bg-black/40 flex items-center gap-3 relative overflow-hidden"
                    data-tutorial="macau-discard-pile"
                  >
                    {state.chosenSuit > 0 && (
                      <span
                        aria-hidden="true"
                        data-testid="chosen-suit-watermark"
                        className="pointer-events-none absolute inset-0 flex items-center justify-end pr-4 text-[6rem] leading-none opacity-15 text-ds-warning motion-safe:animate-suit-watermark"
                      >
                        {SUIT_SYMBOLS[state.chosenSuit] ?? '?'}
                      </span>
                    )}
                    <div className="relative">
                      <AnimatedCard card={state.discardTop} width={cardWidth} />
                    </div>
                    <div className="text-ds-text-muted text-sm relative">
                      <div>{t('discardTop')}</div>
                      {state.chosenSuit > 0 && (
                        <div className="text-ds-warning">
                          {t('chosenSuit')}: {SUIT_SYMBOLS[state.chosenSuit] ?? '?'}
                        </div>
                      )}
                    </div>
                  </div>
                )}

                {hasPenalty && (
                  <div
                    className="my-2 p-2 rounded bg-ds-warning/20 text-ds-warning text-sm font-semibold"
                    role="status"
                  >
                    {t('penaltyBanner', { count: state.penaltyDrawCount })}
                  </div>
                )}

                {isMustDeclare && state.players[state.currentPlayerIdx]?.isHuman && (
                  <div
                    className="my-2 p-2 rounded bg-ds-info/20 text-ds-info text-sm font-semibold"
                    role="status"
                    data-testid="macau-must-declare-banner"
                  >
                    {t('mustDeclareBanner')}
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

              {/* Right: info sidebar */}
              <div>
                {/* CPU players */}
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                        {t('roundScore', { score: p.roundScore })}
                      </div>
                    </div>
                  ))}

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <table className="w-full text-sm text-ds-text-muted">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {t('scoresPlayer')}
                        </th>
                        <th scope="col">{t('scoresRound')}</th>
                        <th scope="col">{t('scoresTotal')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state.players.map((p) => (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>{playerName(p.id, p.isHuman)}</td>
                          <td className="text-center">{p.roundScore}</td>
                          <td className="text-center">{p.cumulativeScore}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>

          <GameFooter className={`${gameTheme.macau.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="macau-player-hand">
                {humanPlayer.cards.map((card, idx) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}-${idx}`}
                    onClick={() => toggleCard(idx)}
                    aria-label={cardAlt(card)}
                    aria-pressed={selectedCardIndices.includes(idx)}
                    className={`transition-transform ${focusRingCard}`}
                    style={{
                      background: 'none',
                      padding: 0,
                      borderRadius: 8,
                      ...selectedCardStyle(selectedCardIndices.includes(idx)),
                      boxSizing: 'border-box',
                    }}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                ))}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex gap-2 items-center flex-wrap" data-tutorial="macau-magic">
              {isHumanTurn && (
                <div className="flex gap-2" data-tutorial="macau-play-draw">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePlay}
                    disabled={loading || selectedCardIndices.length !== 1}
                  >
                    {t('playButton')}
                  </button>
                  <button
                    type="button"
                    className={`${hasPenalty ? btnDanger : btnPrimary} relative`}
                    onClick={handleDraw}
                    disabled={loading}
                  >
                    {hasPenalty ? t('takePenaltyButton', { count: state.penaltyDrawCount }) : t('drawButton')}
                    {hasPenalty && (
                      <span
                        data-testid="penalty-badge"
                        aria-hidden="true"
                        className="pointer-events-none absolute -top-2 -right-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-ds-error px-1 font-bold text-white text-xs"
                      >
                        {state.penaltyDrawCount}
                      </span>
                    )}
                  </button>
                </div>
              )}
              {isChooseSuit && state.players[state.currentPlayerIdx]?.isHuman && (
                <div className="flex gap-1" data-tutorial="macau-suit-choice">
                  {SUIT_BUTTONS.map(({ suit, key }) => (
                    <button
                      key={suit}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleChooseSuit(suit)}
                      disabled={loading}
                    >
                      {t(key)}
                    </button>
                  ))}
                </div>
              )}
              {isMustDeclare && state.players[state.currentPlayerIdx]?.isHuman && (
                <div className="flex gap-2">
                  <button type="button" className={btnSuccess} onClick={handleDeclare} disabled={loading}>
                    {t('declareButton')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleSkipDeclare} disabled={loading}>
                    {t('skipDeclareButton')}
                  </button>
                </div>
              )}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="macau-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
