import { useEffect, useMemo } from 'react';
import type { quodlibetApi } from '../api/gameApi';
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
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { QUODLIBET_CPU_DIFFICULTY_OPTIONS, useQuodlibetGame } from '../hooks/useQuodlibetGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { QuodlibetResponse } from '../types/card';
import { QuodlibetPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseQuodlibetCommand, QUODLIBET_HELP } from '../utils/cli/commands/quodlibetCommands';
import { formatQuodlibetState } from '../utils/cli/formatters/quodlibetFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Quodlibet tutorial step definitions. */
const QUODLIBET_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="quodlibet-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="quodlibet-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="quodlibet-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="quodlibet-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="quodlibet-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

// **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしているので、
// ここは直接引く ── 数値化すると全フェーズが NaN に潰れて名前が消える。
const QUODLIBET_PHASE_KEYS: Readonly<Record<string, string>> = {
  [QuodlibetPhase.SELECT_CONTRACT]: 'selectContract',
  [QuodlibetPhase.PLAY]: 'play',
  [QuodlibetPhase.DEAL_END]: 'dealEnd',
  [QuodlibetPhase.GAME_END]: 'gameEnd',
};

/**
 * Renders the Quodlibet page: the Austrian compendium game of twelve deals in
 * three wheels of four contracts, where the lowest penalty total wins.
 */
export const QuodlibetPage = withTutorial(QuodlibetPageContent, 'quodlibet', QUODLIBET_TUTORIAL_STEPS);

/** Inner content of the page, wrapped by TutorialProvider. */
function QuodlibetPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('quodlibet');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    quodlibetConfig,
    handleConfigChange,
    handleToggle,
    selectedCardIndices,
    toggleCard,
    reset,
    handleSelectContract,
    handlePlay,
    handlePass,
    handleNextDeal,
  } = useQuodlibetGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('quodlibet');
  const cliConfig: CliGameConfig<QuodlibetResponse, Parameters<typeof quodlibetApi.exec>> = useMemo(
    () => ({
      gameName: 'quodlibet',
      parseCommand: parseQuodlibetCommand,
      formatResponse: formatQuodlibetState,
      helpText: QUODLIBET_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('quodlibet', state);
  const { cardWidth, isMobile } = useCardDimensions();

  if (!state)
    return <GameSkeleton gameKey="quodlibet" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isContractPhase = state.phase === QuodlibetPhase.SELECT_CONTRACT;
  const isPlayPhase = state.phase === QuodlibetPhase.PLAY;
  const isDealEnd = state.phase === QuodlibetPhase.DEAL_END;
  const isGameEnd = state.phase === QuodlibetPhase.GAME_END || state.gameEndFlag;
  const canPlay = isPlayPhase && isHumanTurn;
  const canChoose = isContractPhase && isHumanTurn;

  const handValidIndices = canPlay ? state.playableIndices : undefined;
  const humanWon = isGameEnd && state.winners.length === 1 && state.winners[0] === 0;

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.quodlibet')}
      gameThemeBg={gameTheme.quodlibet.bg}
      phaseName={t(`phase.${QUODLIBET_PHASE_KEYS[state.phase] ?? 'play'}`)}
      isHumanTurn={isHumanTurn}
      gamePath="/quodlibet"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
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
                    value: quodlibetConfig.cpuDifficulty,
                    options: QUODLIBET_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'autoSelectContract',
                    label: t('settings.autoSelectContract'),
                    checked: quodlibetConfig.autoSelectContract,
                    onToggle: (v: boolean) => handleToggle('autoSelectContract', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="quodlibet-info">
              <span className="mr-4">{t('deal', { n: state.dealNumber + 1, total: state.totalDeals })}</span>
              <span className="mr-4">{t('wheel', { n: state.roundNumber })}</span>
              {/* **種目でルールが丸ごと変わる。** 出さないと、いま何を避ける
                  ゲームなのか画面のどこにも無い。 */}
              {state.currentContract >= 0 && (
                <span data-testid="quodlibet-contract">
                  {t('contract', { name: t(`contractName.${state.currentContractName}`) })}
                </span>
              )}
              {!state.isShedding && state.currentContract >= 0 && (
                <span className="ml-4">{t('trick', { n: state.trickNumber, total: state.trickCount })}</span>
              )}
            </div>

            <div className={lgTwoColGrid}>
              <div>
                {state.isShedding ? (
                  <div className="mb-2 p-2 rounded bg-black/30" data-testid="quodlibet-shed-area">
                    <div className="text-ds-text-muted text-sm mb-1">{t('shedArea')}</div>
                    {state.stack.length > 0 ? (
                      <div className="flex flex-wrap gap-1">
                        {state.stack.map((c, i) => (
                          <CardImage key={`${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                        ))}
                      </div>
                    ) : (
                      <div className="text-ds-text-muted text-sm">{t('shedEmpty')}</div>
                    )}
                  </div>
                ) : (
                  <TrickDisplay
                    currentTrick={state.currentTrick}
                    players={state.players}
                    cardWidth={cardWidth}
                    label={t('currentTrick')}
                    dataTutorial="quodlibet-trick-display"
                  />
                )}
              </div>

              <div data-tutorial="quodlibet-scores">
                {/* **点は罰点。少ないほうが良い。** 順位の向きを書かないと、
                    多い人が勝っているように読める。 */}
                <div className="mb-1 text-ds-text-muted text-sm">{t('penaltyLegend')}</div>
                <div className="mb-2 p-2 rounded bg-black/30" data-testid="quodlibet-scores">
                  {state.players.map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5 flex items-center gap-2">
                      <span className={state.winners.includes(p.id) ? 'text-ds-success' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('penalty', { n: p.penalty })}
                        {p.dealPoints !== 0 && ` (${t('dealPoints', { n: p.dealPoints })})`}
                      </span>
                      {p.isDealer && (
                        <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                          {t('dealerBadge')}
                        </span>
                      )}
                    </div>
                  ))}
                </div>

                {/* **選べるのはこの輪の残りだけ。** 全 12 種目を並べると、
                    押せない選択肢を勧めることになる。 */}
                {canChoose && (
                  <div className="mb-2 p-2 rounded bg-black/30" data-testid="quodlibet-contract-choices">
                    <div className="text-ds-text-primary text-sm mb-1">{t('chooseContract')}</div>
                    <div className="flex flex-wrap gap-2">
                      {state.availableContracts.map((c, i) => (
                        <button
                          key={c}
                          type="button"
                          className={btnPrimary}
                          onClick={() => handleSelectContract(c)}
                          disabled={loading}
                          data-testid={`quodlibet-contract-${c}`}
                        >
                          {t(`contractName.${state.availableContractNames[i]}`)}
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                {state.lastDeal && (isDealEnd || isGameEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="quodlibet-deal-result"
                  >
                    <div className="mb-1 text-ds-text-primary">
                      {t('dealResult', { name: t(`contractName.${state.lastDeal.contractName}`) })}
                    </div>
                    {state.lastDeal.points.map((pts, i) => (
                      <div key={state.players[i]?.id ?? i}>
                        {playerName(i, i === 0)}: {t('dealPoints', { n: pts })}
                      </div>
                    ))}
                  </div>
                )}

                {isGameEnd && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="quodlibet-winner">
                    {t('winner', { names: state.winners.map((w) => playerName(w, w === 0)).join(', ') })}
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

          <GameFooter className={`${gameTheme.quodlibet.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="quodlibet"
                validIndices={handValidIndices}
                legalIndices={handValidIndices}
                restrictedTooltip={t('restrictedTooltip')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* ライブ領域は常設 (#5955)。 */}
            <div data-testid="quodlibet-hint-live" role="status" aria-live="polite">
              {isRequestedHint(state) && (state.hint || state.hintContract >= 0) && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}
                  {state.hint && `: ${t(`hint.${state.hint.reason}`)}`}
                  {state.hint?.cardIndices &&
                    state.hint.cardIndices.length > 0 &&
                    ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
                  {state.hintContract >= 0 && ` [${t(`contractName.${state.availableContractNames[0] ?? ''}`)}]`}
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="quodlibet-action-buttons">
              {canPlay && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices.length !== 1}
                  data-testid="quodlibet-play"
                >
                  {t('playButton')}
                </button>
              )}
              {/* **パスできるのは出せる札が 1 枚も無いときだけ。** 常設すると
                  「出せるのに降りる」ができてしまい、シェディングが壊れる。 */}
              {canPlay && state.canPass && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handlePass}
                  disabled={loading}
                  data-testid="quodlibet-pass"
                >
                  {t('passButton')}
                </button>
              )}
              {isDealEnd && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNextDeal}
                  disabled={loading}
                  data-testid="quodlibet-next-deal"
                >
                  {t('nextDeal')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="quodlibet-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
