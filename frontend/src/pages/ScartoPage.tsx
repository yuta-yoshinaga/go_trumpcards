import { useEffect, useMemo } from 'react';
import type { scartoApi } from '../api/gameApi';
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
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  CPU_DIFFICULTY_OPTIONS,
  SCARTO_DISCARD_COUNT,
  TARGET_DEALS_OPTIONS,
  useScartoGame,
} from '../hooks/useScartoGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ScartoResponse } from '../types/card';
import { ScartoPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseScartoCommand, SCARTO_HELP } from '../utils/cli/commands/scartoCommands';
import { formatScartoState } from '../utils/cli/formatters/scartoFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { scartoUndiscardableReason } from '../utils/scartoDiscard';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Scarto tutorial step definitions. */
const SCARTO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="scarto-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="scarto-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="scarto-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="scarto-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="scarto-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SCARTO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ScartoPhase.SCARTO]: 'scarto',
  [ScartoPhase.PLAY]: 'play',
  [ScartoPhase.TRICK_END]: 'trickEnd',
  [ScartoPhase.ROUND_END]: 'roundEnd',
  [ScartoPhase.GAME_END]: 'gameEnd',
};

/** Outcome i18n keys indexed by outcome value (0=none/average, 1=above average, 2=below average). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeWin', 'outcomeLoss'] as const;

/** Formats a card-point number with up to one decimal place, dropping a trailing `.0`. */
function formatPoints(n: number): string {
  return Number.isInteger(n) ? String(n) : n.toFixed(1);
}

/** Formats a signed difference, prefixing a leading `+` for positive values. */
function formatSigned(n: number): string {
  const s = formatPoints(n);
  return n > 0 ? `+${s}` : s;
}

/** Renders the Scarto (スカルト) game page: a 3-player 78-card Italian tarocchi trick-taker with a dealer scarto (discard) and trump-priority tricks — no bidding, chien, or partnership. */
export const ScartoPage = withTutorial(ScartoPageContent, 'scarto', SCARTO_TUTORIAL_STEPS);

/** Inner content of the Scarto page, wrapped by TutorialProvider. */
function ScartoPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('scarto');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    scartoConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleScarto,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useScartoGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('scarto');
  const scartoCliConfig: CliGameConfig<ScartoResponse, Parameters<typeof scartoApi.exec>> = useMemo(
    () => ({
      gameName: 'scarto',
      parseCommand: parseScartoCommand,
      formatResponse: formatScartoState,
      helpText: SCARTO_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, scartoCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('scarto', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('scarto', SCARTO_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="scarto" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 25 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isScartoPhase = state.phase === ScartoPhase.SCARTO;
  const isPlayPhase = state.phase === ScartoPhase.PLAY;
  const isTrickEnd = state.phase === ScartoPhase.TRICK_END;
  const isRoundEnd = state.phase === ScartoPhase.ROUND_END;
  const isGameEnd = state.phase === ScartoPhase.GAME_END || state.gameEndFlag;

  const canScarto = isScartoPhase && state.isHumanScarto;
  const canPlay = isPlayPhase && isHumanTurn;

  // Scarto settles each deal against the mean: dealScore_i = N × (cardPoints_i − average),
  // so surfacing the table average and each seat's captured card-points explains the ± delta.
  const totalCardPoints = state.players.reduce((sum, p) => sum + p.cardPoints, 0);
  const averageCardPoints = state.players.length > 0 ? totalCardPoints / state.players.length : 0;

  // **捨てられる札はサーバが決める。** 色と値からここで組み立てると、
  // 捨てられるピップが足りないときに解禁される非オヌール切り札が落ちる。
  // その手を引いた親は、画面から枚数を揃えられなかった (#6236)。
  const discardableIndices = canScarto ? state.discardableIndices : undefined;

  const handValidIndices = canPlay ? state.playableIndices : canScarto ? discardableIndices : undefined;

  // During the scarto, explain per-card why an un-buriable card cannot be
  // discarded (trump / Excuse / bout / counting card) via the card tooltip.
  // Purely additive — it never blocks selection; the backend still rejects
  // illegal discards.
  const scartoTitleFor = (idx: number): string | undefined => {
    const card = humanPlayer?.cards[idx];
    if (!card) return undefined;
    // いま実際に捨てられる札には理由を出さない。ピップが足りないときの
    // 切り札は「捨てられない」ではなく捨てられるので、一覧と食い違う (#6236)。
    if (discardableIndices?.includes(idx)) return undefined;
    const reason = scartoUndiscardableReason(card);
    return reason ? t(`scartoUndiscardable.${reason}`) : undefined;
  };

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.scarto')}
      gameThemeBg={gameTheme.scarto.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/scarto"
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
                    value: scartoConfig.cpuDifficulty,
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
                    value: scartoConfig.targetDeals,
                    options: TARGET_DEALS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetDeals', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span>{t('trick', { n: state.trickNumber })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="scarto-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="scarto-info">
                {/* Per-player scores with a dealer badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.isDealer ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                      </span>
                      {p.isDealer && (
                        <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                          {t('dealerBadge')}
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

                {/* Round result: the deal settlement (signed delta from the average) */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="scarto-result">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.outcome', { outcome: t(OUTCOME_KEYS[state.outcome] ?? 'outcomeNone') })}</div>
                    {state.players.map((p, i) => {
                      const delta = state.dealScores[i] ?? 0;
                      return (
                        <div key={p.id}>
                          {t('roundResult.playerLine', {
                            name: playerName(p.id, p.isHuman),
                            delta: delta > 0 ? `+${delta}` : String(delta),
                            score: p.score,
                          })}
                        </div>
                      );
                    })}
                    {/* Average-difference breakdown: each seat's captured points vs. the table mean. */}
                    <div className="mt-1 pt-1 border-t border-white/10" data-testid="scarto-breakdown">
                      <div>{t('roundResult.average', { avg: formatPoints(averageCardPoints) })}</div>
                      {/* **平均差と実際の変動は N 倍ちがう。**式を書かないと、同じ
                          箱の中で「+15」と「平均差 +5」が並んで計算が合わないように
                          見える (#4930)。 */}
                      <div data-testid="scarto-formula">{t('roundResult.formula', { n: state.players.length })}</div>
                      {state.players.map((p) => (
                        <div key={p.id}>
                          {t('roundResult.earnedLine', {
                            name: playerName(p.id, p.isHuman),
                            points: p.cardPoints,
                            diff: formatSigned(p.cardPoints - averageCardPoints),
                            scaled: formatSigned((p.cardPoints - averageCardPoints) * state.players.length),
                          })}
                        </div>
                      ))}
                    </div>
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
          <GameFooter className={`${gameTheme.scarto.footer} px-4 py-2.5`}>
            {canScarto && (
              <div
                className="mb-1 text-center text-sm text-ds-accent font-semibold"
                data-testid="scarto-discard-prompt"
              >
                {t('scartoPhase', { count: selectedCardIndices.length, total: SCARTO_DISCARD_COUNT })}
              </div>
            )}
            {isScartoPhase && !state.isHumanScarto && (
              <div className="mb-1 text-center text-sm text-ds-text-muted" data-testid="scarto-waiting">
                {t('scartoWaiting')}
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="scarto"
                validIndices={handValidIndices}
                restrictedTooltip={canScarto ? t('scartoRestricted') : t('playButton')}
                cardTitleFor={canScarto ? scartoTitleFor : undefined}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-testid="scarto-hint-live" role="status" aria-live="polite">
              {state.hint && isRequestedHint(state) && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                  {state.hint.cardIndices &&
                    state.hint.cardIndices.length > 0 &&
                    ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="scarto-action-buttons">
              {canScarto && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleScarto}
                  disabled={loading || selectedCardIndices.length !== SCARTO_DISCARD_COUNT}
                >
                  {t('discardButton', { count: selectedCardIndices.length, total: SCARTO_DISCARD_COUNT })}
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
                dataTutorial="scarto-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
