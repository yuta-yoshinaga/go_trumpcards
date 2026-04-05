import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import type { speedApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, useSpeedGame } from '../hooks/useSpeedGame';
import { useSound } from '../providers/SoundProvider';
import { TutorialProvider } from '../providers/TutorialProvider';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import type { SpeedResponse } from '../types/card';
import { SpeedPhase } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { parseSpeedCommand, SPEED_HELP } from '../utils/cli/commands/speedCommands';
import { formatSpeedState } from '../utils/cli/formatters/speedFormatter';
import type { CliGameConfig } from '../utils/cli/types';

const SPEED_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sp-center-piles"]',
    messageKey: 'tutorial.centerPiles',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sp-draw-pile"]', messageKey: 'tutorial.drawPile', placement: 'left', advanceOn: 'next' },
];

const SPEED_TUTORIAL_CONFIG: TutorialConfig = { gameName: 'speed', steps: SPEED_TUTORIAL_STEPS };

/** Renders the Speed game page. */
export function SpeedPage() {
  const { t: tSpeed } = useTranslation('speed');
  return (
    <TutorialProvider config={SPEED_TUTORIAL_CONFIG} translateMessage={tSpeed}>
      <SpeedPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Speed page. */
function SpeedPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('speed');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    speedConfig,
    selectedCardIndices,
    toggleCard,
    handlePlay,
    handleFlip,
    handleHint,
    handleConfigChange,
  } = useSpeedGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('speed', state);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('speed');
  const cliConfig: CliGameConfig<SpeedResponse, Parameters<typeof speedApi.exec>> = useMemo(
    () => ({
      gameName: 'speed',
      parseCommand: parseSpeedCommand,
      formatResponse: formatSpeedState,
      helpText: SPEED_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhase = state?.phase === SpeedPhase.PLAY;
  const isStuck = state?.phase === SpeedPhase.STUCK;
  const isGameEnd = state?.phase === SpeedPhase.GAME_END || state?.gameEndFlag;
  const humanPlayer = state?.players?.[0];
  const cpuPlayer = state?.players?.[1];
  const humanWon = state?.winnerIdx === 0;

  const phaseName = state?.gameEndFlag
    ? t('phase.gameEnd')
    : state?.phase === SpeedPhase.STUCK
      ? t('phase.stuck')
      : t('phase.play');

  return (
    <div className="flex flex-col h-full gap-2 p-2" aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.speed')} />
      {state && (
        <PhaseIndicator phaseName={phaseName} isHumanTurn={isPlayPhase}>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
          <TutorialButton />
          <ManualButton gamePath="/speed" />
        </PhaseIndicator>
      )}

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {error && <ErrorAlert message={error} onRetry={retry} />}
          {state && (
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
          )}

          {state && (
            <div className="flex-1 flex flex-col gap-3 min-h-0">
              {/* CPU area */}
              <div className="flex items-center justify-center gap-2">
                <span className="text-sm text-gray-500">
                  {t('cpuHand')}: {cpuPlayer?.cardCount ?? 0}
                </span>
                <div className="flex gap-1">
                  {Array.from({ length: cpuPlayer?.cardCount ?? 0 }).map((_, i) => (
                    <AnimatedCardBack key={i} width={cardWidth * 0.7} onFlipComplete={() => playSound('cardFlip')} />
                  ))}
                </div>
                <span className="text-sm text-gray-500">
                  {t('drawPile')}: {cpuPlayer?.drawPileSize ?? 0}
                </span>
              </div>

              {/* Center piles */}
              <div className="flex items-center justify-center gap-6" data-tutorial="sp-center-piles">
                {state.centerPiles.map((card, pi) => (
                  <button
                    type="button"
                    key={pi}
                    onClick={() => handlePlay(pi)}
                    disabled={!isPlayPhase || selectedCardIndices.length !== 1 || loading}
                    className={`transition-transform hover:scale-105 disabled:opacity-50 ${focusRingCard}`}
                    aria-label={`${t('centerPile')} ${pi}`}
                  >
                    {card && (
                      <AnimatedCard
                        card={card}
                        width={cardWidth * 1.2}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                    )}
                  </button>
                ))}
              </div>

              {/* Human hand */}
              <div className="flex flex-col items-center gap-1" data-tutorial="sp-player-hand">
                <div className="flex items-center gap-2">
                  <span className="text-sm">{t('yourHand')}</span>
                  <span className="text-sm text-gray-500" data-tutorial="sp-draw-pile">
                    {t('drawPile')}: {humanPlayer?.drawPileSize ?? 0}
                  </span>
                </div>
                <div className="flex gap-1 flex-wrap justify-center">
                  {humanPlayer?.cards.map((card, idx) => (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => toggleCard(idx)}
                      disabled={!isPlayPhase || loading}
                      aria-label={`${card.design} ${card.value}`}
                      aria-pressed={selectedCardIndices.includes(idx)}
                      className={`transition-transform ${focusRingCard}`}
                      style={selectedCardStyle(selectedCardIndices.includes(idx))}
                    >
                      <AnimatedCard
                        card={card}
                        width={cardWidth}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                    </button>
                  ))}
                </div>
              </div>

              {/* Stuck message & flip button */}
              {isStuck && (
                <div className="flex flex-col items-center gap-2">
                  <p className="text-amber-600 font-bold">{t('stuckMessage')}</p>
                  <button
                    type="button"
                    onClick={handleFlip}
                    disabled={loading}
                    className="px-4 py-2 bg-amber-500 text-white rounded hover:bg-amber-600 disabled:opacity-50"
                  >
                    {t('flipButton')}
                  </button>
                </div>
              )}

              {/* Hint */}
              {state.hint?.found && isPlayPhase && (
                <p className="text-center text-sm text-blue-600">
                  {t('hint.play', { cardIndex: state.hint.cardIndex, pileIndex: state.hint.pileIndex })}
                </p>
              )}
              {frontendHintEnabled && frontendHint && (
                <div className="flex justify-center">
                  <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
                </div>
              )}
            </div>
          )}

          {/* Settings */}
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: speedConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label}`),
                    })),
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <GameFooter>
            <button
              type="button"
              onClick={handleHint}
              disabled={loading || !isPlayPhase}
              className="btn btn-sm btn-outline"
            >
              {tc('button.hint')}
            </button>
            <ActionLogSection
              isEndPhase={!!isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
            <button
              type="button"
              onClick={() => requestConfirm(() => gameExec('reset', undefined, undefined, speedConfig))}
              className="btn btn-sm btn-warning"
            >
              {tc('button.reset')}
            </button>
          </GameFooter>
        </>
      )}
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
      <WinCelebration show={!!isGameEnd && humanWon} onCelebrate={() => playSound('winFanfare')} />
    </div>
  );
}
