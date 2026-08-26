import { useCallback, useMemo } from 'react';
import type { ramschApi } from '../api/gameApi';
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
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useRamschGame } from '../hooks/useRamschGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { RamschResponse } from '../types/card';
import { RamschPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseRamschCommand, RAMSCH_HELP } from '../utils/cli/commands/ramschCommands';
import { formatRamschState } from '../utils/cli/formatters/ramschFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Ramsch tutorial step definitions. */
const RAMSCH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sk-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sk-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sk-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Phase translation key map for Ramsch. */
const RAMSCH_PHASE_KEYS: Readonly<Record<number, string>> = {
  [RamschPhase.PLAY]: 'play',
  [RamschPhase.TRICK_END]: 'trickEnd',
  [RamschPhase.ROUND_END]: 'roundEnd',
  [RamschPhase.GAME_END]: 'gameEnd',
};

/** Renders the Ramsch (German trick-taking) game page. */
export const RamschPage = withTutorial(RamschPageContent, 'ramsch', RAMSCH_TUTORIAL_STEPS);
/** Inner content of the Ramsch page. */
function RamschPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ramsch');
  const ramschGame = useRamschGame();
  const {
    state,
    loading,
    error,
    ramschConfig,
    handleConfigChange,
    reset,
    selectedCardIndices,
    toggleCard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    retry,
  } = ramschGame;
  const { cardWidth, isMobile: _isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('ramsch', RAMSCH_PHASE_KEYS);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('ramsch', state);
  const cliMode = useCliMode('ramsch');
  const cliConfig: CliGameConfig<RamschResponse, Parameters<typeof ramschApi.exec>> = useMemo(
    () => ({
      gameName: 'ramsch',
      parseCommand: parseRamschCommand,
      formatResponse: formatRamschState,
      helpText: RAMSCH_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(ramschGame.dispatch, cliConfig, state, {
    addInput: cliMode.addInput,
    addOutput: cliMode.addOutput,
    addError: cliMode.addError,
    clearLog: cliMode.clearLog,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    reset();
  }, [hideActionLog, reset]);

  if (!state) {
    return (
      <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.ramsch.bg}`}>
        <div className="flex-1 flex items-center justify-center text-ds-text-primary">
          <p>{tc('skeleton.loading')}</p>
        </div>
      </div>
    );
  }

  const isPlay = state.phase === RamschPhase.PLAY;
  const isTrickEnd = state.phase === RamschPhase.TRICK_END;
  const isRoundEnd = state.phase === RamschPhase.ROUND_END;
  const isGameEnd = state.phase === RamschPhase.GAME_END || state.gameEndFlag;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = isPlay && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.ramsch')}
      gameThemeBg={gameTheme.ramsch.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/ramsch"
      gameEndFlag={isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliMode.cliEnabled} onToggle={cliMode.toggleCli} />}
    >
      {cliMode.cliEnabled ? (
        <CliTerminal logEntries={cliMode.logEntries} onCommand={handleCommand} disabled={loading} />
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
                    value: ramschConfig.cpuDifficulty,
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
                    value: ramschConfig.targetScore,
                    options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetScore', v),
                  },
                  hintCheckboxItem(tc, hintEnabled, setHintEnabled),
                ],
              },
            ]}
          />
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-4">
            {error && <ErrorAlert message={error} onRetry={retry} />}

            {/* Round info. **切り札と「点は罰点」を常に書く。** どちらもスカート系の
                つもりで来た人が真っ先に取り違えるところで、無ければ盤が読めない。 */}
            <div className="bg-black/30 text-ds-text-primary p-3 rounded space-y-1 text-sm">
              <div>
                {t('round')}: {state.roundNumber} | {t('dealer')}: CPU {state.dealerIdx}
              </div>
              <div className="text-xs text-ds-text-muted" data-testid="ramsch-trump-note">
                {t('trumpFixed')}
              </div>
              <div className="text-xs text-ds-text-muted" data-testid="ramsch-scoring-note">
                {t('scoringNote')}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Trick display */}
            <TrickDisplay
              currentTrick={state.currentTrick.map((tc) => ({ playerIdx: tc.playerIdx, card: tc.card }))}
              players={state.players.map((p) => ({ id: p.id, isHuman: p.isHuman }))}
              cardWidth={cardWidth}
              label={t('currentTrick')}
              dataTutorial="sk-trick-display"
            />

            {/* **伏せ札はラウンドが終わってから見せる。** 最終トリックの獲得者が
                受け取る 2 枚なので、途中で中身が分かると終盤の判断が完全情報になる。
                サーバもラウンド終了まで返さない。 */}
            {state.skat && state.skat.length > 0 && (
              <div className="bg-black/30 text-ds-text-primary p-3 rounded">
                <div className="text-sm mb-1">{t('skatLabel')}:</div>
                <div className="flex gap-2" data-testid="ramsch-skat-reveal">
                  {state.skat.map((c, i) => (
                    <AnimatedCard key={`skat-${i}`} card={c} width={cardWidth} dealDelay={i * 0.15} />
                  ))}
                </div>
              </div>
            )}

            {/* Player hand */}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={_isMobile}
                dataTutorialPrefix="sk"
              />
            )}

            {/* Player scores */}
            <div className="bg-black/30 text-ds-text-primary p-3 rounded text-sm">
              <table className="w-full">
                <thead>
                  <tr>
                    <th className="text-left">{t('player')}</th>
                    <th className="text-right">{t('tricks')}</th>
                    <th className="text-right">{t('cardPoints')}</th>
                    <th className="text-right">{t('total')}</th>
                  </tr>
                </thead>
                <tbody>
                  {state.players.map((p) => (
                    <tr key={p.id}>
                      <td>
                        {p.isHuman ? t('you') : `CPU ${p.id}`}
                        {/* **誰が負けたかを表の中で言う。** 罰点なので「点が多い＝負け」で、
                            数字だけ並べると多い人が勝っているように読める。 */}
                        {isRoundEnd && state.durchmarsch && state.durchmarschIdx === p.id && (
                          <span className="ml-1 text-ds-success" data-testid={`ramsch-durchmarsch-${p.id}`}>
                            ({t('durchmarschBadge')})
                          </span>
                        )}
                        {isRoundEnd && !state.durchmarsch && state.loserIdx === p.id && (
                          <span className="ml-1 text-ds-error" data-testid={`ramsch-loser-${p.id}`}>
                            ({t('loserBadge')})
                          </span>
                        )}
                      </td>
                      <td className="text-right">{p.trickCount}</td>
                      <td className="text-right">{p.cardPoints}</td>
                      <td className="text-right">{p.cumulativeScore}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <ActionLogSection
              isEndPhase={isRoundEnd || isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.ramsch.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap gap-2 items-center">
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sk-reset-button"
              />

              {/* Play phase */}
              {isPlay && isHumanTurn && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('play')}
                </button>
              )}

              {isTrickEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              )}

              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
            </div>
          </GameFooter>
          {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
        </>
      )}
    </GamePageShell>
  );
}
