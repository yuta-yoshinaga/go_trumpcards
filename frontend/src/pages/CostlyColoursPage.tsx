import { useEffect, useMemo } from 'react';
import type { costlycoloursApi } from '../api/gameApi';
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
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import {
  COSTLYCOLOURS_CPU_DIFFICULTY_OPTIONS,
  COSTLYCOLOURS_TARGET_OPTIONS,
  useCostlyColoursGame,
} from '../hooks/useCostlyColoursGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CostlyColoursResponse } from '../types/card';
import { CostlyColoursPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { COSTLYCOLOURS_HELP, parseCostlyColoursCommand } from '../utils/cli/commands/costlycoloursCommands';
import { formatCostlyColoursState } from '../utils/cli/formatters/costlycoloursFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Costly Colours tutorial step definitions. */
const COSTLYCOLOURS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="costlycolours-count"]',
    messageKey: 'tutorial.count',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="costlycolours-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="costlycolours-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="costlycolours-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="costlycolours-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

// **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしている。
const COSTLYCOLOURS_PHASE_KEYS: Readonly<Record<string, string>> = {
  [CostlyColoursPhase.MOG]: 'mog',
  [CostlyColoursPhase.PLAY]: 'play',
  [CostlyColoursPhase.SHOW]: 'show',
  [CostlyColoursPhase.GAME_END]: 'gameEnd',
};

/**
 * Renders the Costly Colours page: the Shropshire ancestor of Cribbage, where
 * the count scores at 15, 25 and 31 and the show counts a ladder of colours.
 */
export const CostlyColoursPage = withTutorial(CostlyColoursPageContent, 'costlycolours', COSTLYCOLOURS_TUTORIAL_STEPS);

/** Inner content of the page, wrapped by TutorialProvider. */
function CostlyColoursPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('costlycolours');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    costlyColoursConfig,
    handleConfigChange,
    mog,
    play,
    handleNextDeal,
    reset,
  } = useCostlyColoursGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('costlycolours');
  const cliConfig: CliGameConfig<CostlyColoursResponse, Parameters<typeof costlycoloursApi.exec>> = useMemo(
    () => ({
      gameName: 'costlycolours',
      parseCommand: parseCostlyColoursCommand,
      formatResponse: formatCostlyColoursState,
      helpText: COSTLYCOLOURS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('costlycolours', state);
  const { cardWidth, isMobile } = useCardDimensions();

  if (!state)
    return (
      <GameSkeleton
        gameKey="costlycolours"
        layout={{ kind: 'casino-table', sections: [1], footerStyle: 'hand', footerHandSize: 3 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isMogPhase = state.phase === CostlyColoursPhase.MOG;
  const isPlayPhase = state.phase === CostlyColoursPhase.PLAY;
  const isShow = state.phase === CostlyColoursPhase.SHOW;
  const isGameEnd = state.phase === CostlyColoursPhase.GAME_END || state.gameEndFlag;
  const canPlay = isPlayPhase && state.isHumanTurn;

  // **出せる札はサーバが数えたものだけ。** 31 を超える札を並べると押しても弾かれる。
  const playable = canPlay ? state.playableIdxs : [];

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.costlycolours')}
      gameThemeBg={gameTheme.costlycolours.bg}
      phaseName={t(`phase.${COSTLYCOLOURS_PHASE_KEYS[state.phase] ?? 'play'}`)}
      isHumanTurn={state.isHumanTurn}
      gamePath="/costlycolours"
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
                    value: costlyColoursConfig.cpuDifficulty,
                    options: COSTLYCOLOURS_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: costlyColoursConfig.targetScore,
                    options: COSTLYCOLOURS_TARGET_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetScore', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              {t('deal', { n: state.dealNumber, target: state.config.targetScore })}
            </div>

            <div className={lgTwoColGrid}>
              <div data-tutorial="costlycolours-count">
                {/* **表の 1 枚は常に見せる。** ショーの色役も J / 2 の 4 点もこれ次第。 */}
                <div className="mb-1 text-ds-text-muted text-sm">{t('turnUpLabel')}</div>
                <div
                  className="mb-2 p-2 rounded bg-black/30 flex gap-1 items-center"
                  data-testid="costlycolours-turnup"
                >
                  {state.turnUp ? (
                    <CardImage card={state.turnUp} width={cardWidth} />
                  ) : (
                    <span className="text-ds-text-muted text-sm">{t('noTurnUp')}</span>
                  )}
                </div>

                <div className="mb-1 text-ds-text-muted text-sm">{t('countLabel')}</div>
                <div
                  className="mb-2 p-2 rounded bg-black/30 flex flex-wrap gap-1 items-center"
                  data-testid="costlycolours-pile"
                >
                  {state.pile.length === 0 ? (
                    <span className="text-ds-text-muted text-sm">{t('pileEmpty')}</span>
                  ) : (
                    state.pile.map((c, i) => (
                      <CardImage key={`${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                    ))
                  )}
                </div>
                <div className="text-ds-text-primary" data-testid="costlycolours-total">
                  {t('total', { n: state.total })}
                </div>
              </div>

              <div data-tutorial="costlycolours-scores">
                <div className="mb-2 p-2 rounded bg-black/30" data-testid="costlycolours-scores">
                  {state.players.map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                      <span className="flex items-center gap-2">
                        <span>
                          {playerName(p.id, p.isHuman)}: {t('score', { n: p.score })} / {t('cards', { n: p.cardCount })}
                        </span>
                        {p.isDealer && (
                          <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                            {t('dealerBadge')}
                          </span>
                        )}
                      </span>
                    </div>
                  ))}
                </div>

                {state.lastResult && (isShow || isGameEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="costlycolours-show"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('showTitle')}</div>
                    {state.lastResult.lines
                      .filter((l) => l.points[0] !== 0 || l.points[1] !== 0)
                      .map((l) => (
                        <div key={l.key}>
                          {t(`score.${l.key}`)}: {l.points[0]} - {l.points[1]}
                        </div>
                      ))}
                    {/* **どの色役が付いたのかを名指す。** 点だけだと梯子のどこか分からない。 */}
                    {state.lastResult.combos.map((combo, i) =>
                      combo ? (
                        <div
                          key={combo + String(i)}
                          className="text-ds-warning"
                          data-testid={`costlycolours-combo-${i}`}
                        >
                          {playerName(i, i === 0)}: {t(`combo.${combo}`)}
                        </div>
                      ) : null,
                    )}
                    <div className="text-ds-text-primary">
                      {t('showTotal', { a: state.lastResult.totals[0], b: state.lastResult.totals[1] })}
                    </div>
                  </div>
                )}

                {isGameEnd && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="costlycolours-winner">
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

          <GameFooter className={`${gameTheme.costlycolours.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={[]}
                toggleCard={play}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="costlycolours"
                validIndices={canPlay ? playable : undefined}
                legalIndices={canPlay ? playable : undefined}
                restrictedTooltip={t('restrictedTooltip')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* ライブ領域は常設 (#5955)。 */}
            <div data-testid="costlycolours-hint-live" role="status" aria-live="polite">
              {isRequestedHint(state) && state.hintReason && state.hintReason !== 'none' && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {t(`hint.${state.hintReason}`)}
                  {state.hintHandIdx >= 0 && ` ([${state.hintHandIdx}])`}
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="costlycolours-action-buttons">
              {/* **応じる／断るは別のボタン。** 断ると相手に 1 点入るので、
                  片方を既定にしない。 */}
              {isMogPhase && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => mog(true)}
                    disabled={loading}
                    data-testid="costlycolours-mog"
                  >
                    {t('mogAccept')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={() => mog(false)}
                    disabled={loading}
                    data-testid="costlycolours-nomog"
                  >
                    {t('mogRefuse')}
                  </button>
                </>
              )}
              {isShow && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNextDeal}
                  disabled={loading}
                  data-testid="costlycolours-next-deal"
                >
                  {t('nextDeal')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="costlycolours-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
