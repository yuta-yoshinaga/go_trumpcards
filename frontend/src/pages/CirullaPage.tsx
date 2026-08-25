import { useEffect, useMemo } from 'react';
import type { cirullaApi } from '../api/gameApi';
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
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CIRULLA_CPU_DIFFICULTY_OPTIONS, CIRULLA_TARGET_OPTIONS, useCirullaGame } from '../hooks/useCirullaGame';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CirullaResponse } from '../types/card';
import { CirullaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CIRULLA_HELP, parseCirullaCommand } from '../utils/cli/commands/cirullaCommands';
import { formatCirullaState } from '../utils/cli/formatters/cirullaFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Cirulla tutorial step definitions. */
const CIRULLA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cirulla-table"]',
    messageKey: 'tutorial.table',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cirulla-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cirulla-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cirulla-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cirulla-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

// **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしている。
const CIRULLA_PHASE_KEYS: Readonly<Record<string, string>> = {
  [CirullaPhase.PLAY]: 'play',
  [CirullaPhase.ROUND_END]: 'roundEnd',
  [CirullaPhase.GAME_END]: 'gameEnd',
};

/**
 * Renders the Cirulla page: the Ligurian fishing game where sum-to-fifteen is
 * added to Scopa's captures and an ace can sweep the table.
 */
export const CirullaPage = withTutorial(CirullaPageContent, 'cirulla', CIRULLA_TUTORIAL_STEPS);

/** Inner content of the page, wrapped by TutorialProvider. */
function CirullaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('cirulla');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    cirullaConfig,
    handleConfigChange,
    selectedHandIdx,
    selectHand,
    play,
    handleNextRound,
    reset,
  } = useCirullaGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cirulla');
  const cliConfig: CliGameConfig<CirullaResponse, Parameters<typeof cirullaApi.exec>> = useMemo(
    () => ({
      gameName: 'cirulla',
      parseCommand: parseCirullaCommand,
      formatResponse: formatCirullaState,
      helpText: CIRULLA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('cirulla', state);
  const { cardWidth, isMobile } = useCardDimensions();

  if (!state)
    return (
      <GameSkeleton
        gameKey="cirulla"
        layout={{ kind: 'casino-table', sections: [4], footerStyle: 'hand', footerHandSize: 3 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === CirullaPhase.PLAY;
  const isRoundEnd = state.phase === CirullaPhase.ROUND_END;
  const isGameEnd = state.phase === CirullaPhase.GAME_END || state.gameEndFlag;
  const canPlay = isPlayPhase && state.isHumanTurn;

  // **取れる組はサーバが数えたものだけ。** 3 つの規則が絡むので、画面側で
  // 組み直すと必ずずれる。
  const options = selectedHandIdx !== null ? (state.captureOptions[selectedHandIdx] ?? []) : [];
  const handValidIndices = canPlay ? humanPlayer?.cards.map((_, i) => i) : undefined;

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.cirulla')}
      gameThemeBg={gameTheme.cirulla.bg}
      phaseName={t(`phase.${CIRULLA_PHASE_KEYS[state.phase] ?? 'play'}`)}
      isHumanTurn={state.isHumanTurn}
      gamePath="/cirulla"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerIdx === 0}
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
                    value: cirullaConfig.cpuDifficulty,
                    options: CIRULLA_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: cirullaConfig.targetScore,
                    options: CIRULLA_TARGET_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetScore', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber, target: state.config.targetScore })}</span>
              <span data-testid="cirulla-deck">{t('deck', { n: state.deckRemaining })}</span>
            </div>

            <div className={lgTwoColGrid}>
              <div data-tutorial="cirulla-table">
                <div className="mb-1 text-ds-text-muted text-sm">{t('tableLabel')}</div>
                <div className="mb-2 p-2 rounded bg-black/30 flex flex-wrap gap-1" data-testid="cirulla-table">
                  {state.table.length === 0 ? (
                    <span className="text-ds-text-muted text-sm">{t('tableEmpty')}</span>
                  ) : (
                    state.table.map((c, i) => (
                      <div key={`${c.design}-${c.value}-${i}`} className="flex flex-col items-center">
                        <CardImage card={c} width={cardWidth} />
                        <span className="text-xs text-ds-text-muted">{i}</span>
                      </div>
                    ))
                  )}
                </div>

                {/* **取れる組はサーバが数えたものだけ。** 同値・合計一致・合計 15・
                    アッソの総取りが混ざるので、画面側で組み直すと必ずずれる。 */}
                {canPlay && selectedHandIdx !== null && (
                  <div className="mb-2 p-2 rounded bg-black/30" data-testid="cirulla-capture-options">
                    <div className="text-ds-text-primary text-sm mb-1">
                      {options.length > 0 ? t('chooseCapture') : t('noCapture')}
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {options.map((group) => (
                        <button
                          key={group.join('-')}
                          type="button"
                          className={btnPrimary}
                          onClick={() => play(group)}
                          disabled={loading}
                          data-testid={`cirulla-take-${group.join('-')}`}
                        >
                          {t('takeGroup', { cards: group.join(', ') })}
                        </button>
                      ))}
                      {options.length === 0 && (
                        <button
                          type="button"
                          className={btnSuccess}
                          onClick={() => play([])}
                          disabled={loading}
                          data-testid="cirulla-lay-off"
                        >
                          {t('layOffButton')}
                        </button>
                      )}
                    </div>
                  </div>
                )}
              </div>

              <div data-tutorial="cirulla-scores">
                <div className="mb-2 p-2 rounded bg-black/30" data-testid="cirulla-scores">
                  {state.players.map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                      <span className="flex items-center gap-2">
                        <span>
                          {playerName(p.id, p.isHuman)}: {t('score', { n: p.score })} /{' '}
                          {t('taken', { n: p.capturedCount, denari: p.denariCount })} / {t('scope', { n: p.scope })}
                        </span>
                        {p.isDealer && (
                          <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                            {t('dealerBadge')}
                          </span>
                        )}
                      </span>
                      {/* **配札ボーナスは出た瞬間に見せる。** 集計まで伏せると
                          なぜ点が動いたのか分からない。 */}
                      {p.lastBonus && (
                        <span className="text-ds-warning text-xs" data-testid={`cirulla-bonus-${p.id}`}>
                          {t(`bonus.${p.lastBonus}`)}
                        </span>
                      )}
                    </div>
                  ))}
                </div>

                {state.lastResult && (isRoundEnd || isGameEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="cirulla-round-result"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('roundResult')}</div>
                    {state.lastResult.lines
                      .filter((l) => l.points[0] !== 0 || l.points[1] !== 0)
                      .map((l) => (
                        <div key={l.key}>
                          {t(`score.${l.key}`)}: {l.points[0]} - {l.points[1]}
                        </div>
                      ))}
                    <div className="text-ds-text-primary">
                      {t('roundTotal', { a: state.lastResult.totals[0], b: state.lastResult.totals[1] })}
                    </div>
                    {state.lastResult.sweptDenari >= 0 && (
                      <div className="text-ds-warning" data-testid="cirulla-swept-denari">
                        {t('sweptDenari', {
                          name: playerName(state.lastResult.sweptDenari, state.lastResult.sweptDenari === 0),
                        })}
                      </div>
                    )}
                  </div>
                )}

                {isGameEnd && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="cirulla-winner">
                    {t('winner', { name: playerName(state.winnerIdx, state.winnerIdx === 0) })}
                  </div>
                )}
              </div>
            </div>

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

          <GameFooter className={`${gameTheme.cirulla.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedHandIdx === null ? [] : [selectedHandIdx]}
                toggleCard={selectHand}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="cirulla"
                validIndices={handValidIndices}
                legalIndices={handValidIndices}
                restrictedTooltip={t('restrictedTooltip')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* ライブ領域は常設 (#5955)。 */}
            <div data-testid="cirulla-hint-live" role="status" aria-live="polite">
              {isRequestedHint(state) && state.hintHandIdx >= 0 && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {t(`hint.${state.hintReason}`)} ([{state.hintHandIdx}]
                  {state.hintCaptureIdxs.length > 0 && ` → ${state.hintCaptureIdxs.join(', ')}`})
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="cirulla-action-buttons">
              {isRoundEnd && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNextRound}
                  disabled={loading}
                  data-testid="cirulla-next-round"
                >
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="cirulla-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
