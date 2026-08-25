import { useEffect, useMemo } from 'react';
import type { sutdaApi } from '../api/gameApi';
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
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import {
  SUTDA_CHIP_OPTIONS,
  SUTDA_CPU_DIFFICULTY_OPTIONS,
  SUTDA_SEAT_OPTIONS,
  useSutdaGame,
} from '../hooks/useSutdaGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnDanger, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SutdaResponse } from '../types/card';
import { SutdaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSutdaCommand, SUTDA_HELP } from '../utils/cli/commands/sutdaCommands';
import { formatSutdaState } from '../utils/cli/formatters/sutdaFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Sutda tutorial step definitions. */
const SUTDA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sutda-table"]',
    messageKey: 'tutorial.table',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sutda-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sutda-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sutda-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

// **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしているので
// ここは直接引く。
const SUTDA_PHASE_KEYS: Readonly<Record<string, string>> = {
  [SutdaPhase.BET]: 'bet',
  [SutdaPhase.SHOWDOWN]: 'showdown',
  [SutdaPhase.GAME_END]: 'gameEnd',
};

/**
 * Renders the Sutda page: the Korean two-card betting game on a 20-card
 * hanafuda pack.
 */
export const SutdaPage = withTutorial(SutdaPageContent, 'sutda', SUTDA_TUTORIAL_STEPS);

/** Inner content of the page, wrapped by TutorialProvider. */
function SutdaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sutda');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    sutdaConfig,
    handleConfigChange,
    reset,
    handleCall,
    handleRaise,
    handleFold,
    handleNextHand,
  } = useSutdaGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sutda');
  const cliConfig: CliGameConfig<SutdaResponse, Parameters<typeof sutdaApi.exec>> = useMemo(
    () => ({
      gameName: 'sutda',
      parseCommand: parseSutdaCommand,
      formatResponse: formatSutdaState,
      helpText: SUTDA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('sutda', state);
  const { cardWidth } = useCardDimensions();

  if (!state)
    return (
      <GameSkeleton
        gameKey="sutda"
        layout={{ kind: 'casino-table', sections: [2, 2, 2], footerStyle: 'hand', footerHandSize: 2 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBetPhase = state.phase === SutdaPhase.BET;
  const isShowdown = state.phase === SutdaPhase.SHOWDOWN;
  const isGameEnd = state.phase === SutdaPhase.GAME_END || state.gameEndFlag;
  const canAct = isBetPhase && state.isHumanTurn;

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.sutda')}
      gameThemeBg={gameTheme.sutda.bg}
      phaseName={t(`phase.${SUTDA_PHASE_KEYS[state.phase] ?? 'bet'}`)}
      isHumanTurn={state.isHumanTurn}
      gamePath="/sutda"
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
                    value: sutdaConfig.cpuDifficulty,
                    options: SUTDA_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'seats',
                    label: t('settings.seats'),
                    value: sutdaConfig.seats,
                    options: SUTDA_SEAT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('seats', v),
                  },
                  {
                    type: 'select',
                    id: 'startChips',
                    label: t('settings.startChips'),
                    value: sutdaConfig.startChips,
                    options: SUTDA_CHIP_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('startChips', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('hand', { n: state.handNumber })}</span>
              <span data-testid="sutda-pot">{t('pot', { pot: state.pot, bet: state.currentBet })}</span>
            </div>

            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="sutda-table" data-testid="sutda-table">
              {state.players.map((p) => (
                <div key={p.id} className="py-1 border-b border-white/5 last:border-0">
                  <div className="text-ds-text-muted text-sm flex items-center gap-2">
                    <span className={p.folded ? 'opacity-50 line-through' : ''}>
                      {playerName(p.id, p.isHuman)}: {t('chips', { n: p.chips })} / {t('bet', { n: p.bet })}
                    </span>
                    {p.isDealer && (
                      <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('dealerBadge')}</span>
                    )}
                    {p.folded && (
                      <span className="text-xs text-ds-text-muted" data-testid={`sutda-folded-${p.id}`}>
                        {t('folded')}
                      </span>
                    )}
                  </div>
                  {/* **伏せているうちは自分のぶんだけ。** 相手の役が見えると
                      賭ける意味が無くなる。 */}
                  {p.cards.length > 0 && (
                    <div className="flex items-center gap-1 mt-1" data-testid={`sutda-cards-${p.id}`}>
                      {p.cards.map((c, i) => (
                        <CardImage key={`${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                      ))}
                      <span className="ml-2 text-ds-text-primary text-sm">{t(`handName.${p.handName}`)}</span>
                    </div>
                  )}
                </div>
              ))}
            </div>

            {state.lastResult && (isShowdown || isGameEnd) && (
              <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="sutda-result">
                {t('showdownResult', {
                  names: state.lastResult.winners.map((w) => playerName(w, w === 0)).join(', '),
                  pot: state.lastResult.pot,
                })}
              </div>
            )}

            {isGameEnd && (
              <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="sutda-winner">
                {t('winner', { name: playerName(state.winnerIdx, state.winnerIdx === 0) })}
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

          <GameFooter className={`${gameTheme.sutda.footer} px-4 py-2.5`}>
            {/* **自分の役は常に見える。** 伏せているのは相手の札であって、自分の
                2 枚が何の役かは配られた時点で分かっている。 */}
            {humanPlayer && (
              <div className="mb-2 flex items-center gap-2" data-tutorial="sutda-hand" data-testid="sutda-hand">
                {humanPlayer.cards.map((c, i) => (
                  <CardImage key={`${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                ))}
                <span className="text-ds-text-primary">{t(`handName.${state.humanHandName}`)}</span>
              </div>
            )}

            {canAct && (
              <div className="mb-1 text-sm text-ds-text-muted" data-testid="sutda-to-call">
                {state.callAmount > 0 ? t('toCall', { n: state.callAmount }) : t('canCheck')}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* ライブ領域は常設 (#5955)。 */}
            <div data-testid="sutda-hint-live" role="status" aria-live="polite">
              {isRequestedHint(state) && state.hintAction && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {t(`action.${state.hintAction}`)} ({t(`hint.${state.hintReason}`)})
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="sutda-action-buttons">
              {canAct && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleCall}
                  disabled={loading}
                  data-testid="sutda-call"
                >
                  {state.callAmount > 0 ? t('callButton', { n: state.callAmount }) : t('checkButton')}
                </button>
              )}
              {/* **上限とチップの両方を見た結果が canRaise。** 押せない条件を
                  画面側で組み直すと、サーバの判断と食い違う。 */}
              {canAct && state.canRaise && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleRaise}
                  disabled={loading}
                  data-testid="sutda-raise"
                >
                  {t('raiseButton', { n: state.betUnit })}
                </button>
              )}
              {canAct && (
                <button
                  type="button"
                  className={btnDanger}
                  onClick={handleFold}
                  disabled={loading}
                  data-testid="sutda-fold"
                >
                  {t('foldButton')}
                </button>
              )}
              {isShowdown && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNextHand}
                  disabled={loading}
                  data-testid="sutda-next-hand"
                >
                  {t('nextHand')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sutda-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
