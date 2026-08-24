import { useEffect, useMemo } from 'react';
import type { gleekApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useGleekGame } from '../hooks/useGleekGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { GleekResponse } from '../types/card';
import { GleekPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { GLEEK_HELP, parseGleekCommand } from '../utils/cli/commands/gleekCommands';
import { formatGleekState } from '../utils/cli/formatters/gleekFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Gleek tutorial step definitions. */
const GLEEK_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="gleek-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gleek-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gleek-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="gleek-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="gleek-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const GLEEK_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GleekPhase.BID]: 'bid',
  [GleekPhase.DISCARD]: 'discard',
  [GleekPhase.PLAY]: 'play',
  [GleekPhase.TRICK_END]: 'trickEnd',
  [GleekPhase.ROUND_END]: 'roundEnd',
  [GleekPhase.GAME_END]: 'gameEnd',
};

/** Trump-suit i18n keys indexed by suit code (1=♠ 2=♣ 3=♥ 4=♦); index 0 = none. */
const SUIT_KEYS = ['suitNone', 'suitSpade', 'suitClub', 'suitHeart', 'suitDiamond'] as const;

/** Meld-rank i18n keys, keyed by the card value the set is made of. */
const RANK_KEYS: Readonly<Record<number, string>> = { 1: 'rankAce', 13: 'rankKing', 12: 'rankQueen', 11: 'rankJack' };

/**
 * Renders the Gleek game page: a 3-player 44-card trick-taker that settles four
 * scoring stages inside one deal — the auction for the stock, the ruff, the
 * gleek/mournival melds, and twelve tricks scored against a par read from the
 * deal itself.
 */
export const GleekPage = withTutorial(GleekPageContent, 'gleek', GLEEK_TUTORIAL_STEPS);

/** Inner content of the Gleek page, wrapped by TutorialProvider. */
function GleekPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('gleek');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    gleekConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useGleekGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('gleek');
  const gleekCliConfig: CliGameConfig<GleekResponse, Parameters<typeof gleekApi.exec>> = useMemo(
    () => ({
      gameName: 'gleek',
      parseCommand: parseGleekCommand,
      formatResponse: formatGleekState,
      helpText: GLEEK_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, gleekCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('gleek', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('gleek', GLEEK_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="gleek" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 12 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isPlayPhase = state.phase === GleekPhase.PLAY;
  const isTrickEnd = state.phase === GleekPhase.TRICK_END;
  const isRoundEnd = state.phase === GleekPhase.ROUND_END;
  const isGameEnd = state.phase === GleekPhase.GAME_END || state.gameEndFlag;

  const canBid = state.phase === GleekPhase.BID && state.isHumanBidTurn;
  const canDiscard = state.phase === GleekPhase.DISCARD && state.isHumanDiscardTurn;
  const canPlay = isPlayPhase && isHumanTurn;
  // **上限に達したら競り上げのボタンを出さない。** サーバは 0 以外を弾くので、
  // 押せるのに必ず失敗するボタンになる。
  const canRaise = canBid && state.nextBidAmount > 0;
  const discardReady = selectedCardIndices.length === state.discardCount;

  const trumpLabel = state.trumpSuit >= 1 ? t(SUIT_KEYS[state.trumpSuit] ?? 'suitNone') : t('suitNone');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.gleek')}
      gameThemeBg={gameTheme.gleek.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/gleek"
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
                    value: gleekConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: gleekConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>{t('trump', { suit: trumpLabel })}</span>
            </div>

            {/* **段階の点は出さないと見えない。** 競りとトリックの間で累積点が
                動く理由は、ここに出さないと画面のどこにも現れない。 */}
            <div className="text-ds-text-muted text-center text-sm mb-2" data-testid="gleek-stage-line">
              <div>
                {state.buyerIdx >= 0
                  ? t('stockSold', {
                      name: playerName(state.buyerIdx, state.buyerIdx === humanIdx),
                      bid: state.winningBid,
                    })
                  : t('stockUnsold', { bid: state.highestBid })}
              </div>
              {state.ruffWinnerIdx >= 0 && (
                <div data-testid="gleek-ruff-line">
                  {t('ruffLine', {
                    name: playerName(state.ruffWinnerIdx, state.ruffWinnerIdx === humanIdx),
                    total: state.players[state.ruffWinnerIdx]?.ruff ?? 0,
                    suit: t(SUIT_KEYS[state.players[state.ruffWinnerIdx]?.ruffSuit ?? 0] ?? 'suitNone'),
                  })}
                </div>
              )}
              {state.melds.map((m) => (
                <div key={`${m.playerIdx}-${m.rank}`} data-testid="gleek-meld-line">
                  {t(m.count >= 4 ? 'meldMournival' : 'meldGleek', {
                    name: playerName(m.playerIdx, m.playerIdx === humanIdx),
                    rank: t(RANK_KEYS[m.rank] ?? 'rankAce'),
                    value: m.value,
                  })}
                </div>
              ))}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="gleek-trick-display"
                  // **取った札を光らせる。** トリックが揃ってから次へ進むまでの間に
                  // 誰が取ったのかを出さないと、棋譜を開くまで分からない。
                  winnerIdx={isTrickEnd ? state.lastTrickWinner : undefined}
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="gleek-info">
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.isBuyer ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                      </span>
                      {p.isBuyer && (
                        <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('buyerBadge')}</span>
                      )}
                    </div>
                  ))}
                </div>

                <div className="mb-2 p-2 rounded bg-black/30">
                  {state.players.map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                      {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                      {t('tricks', { count: p.trickCount })} | {t('points', { points: p.trickPoints })}
                    </div>
                  ))}
                </div>

                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {/* **基準点はそのディールから数える。** 上限を出すと、
                        名札が場外に落ちたディールで説明が合わなくなる。 */}
                    <div data-testid="gleek-par-line">
                      {t('roundResult.par', { total: state.dealPoints, par: state.par })}
                    </div>
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

          {/* Footer */}
          <GameFooter className={`${gameTheme.gleek.footer} px-4 py-2.5`}>
            {canBid && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="gleek-bid-prompt">
                {t('bidPhase', { bid: state.highestBid })}
              </div>
            )}
            {canDiscard && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="gleek-discard-prompt">
                {t('discardPhase', { count: state.discardCount, selected: selectedCardIndices.length })}
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="gleek"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-testid="gleek-hint-live" role="status" aria-live="polite">
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="gleek-action-buttons">
              {canBid && (
                <div className="flex flex-wrap gap-2 items-center" data-testid="gleek-bid-controls">
                  {canRaise && (
                    <button
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleBid(state.nextBidAmount)}
                      disabled={loading}
                    >
                      {t('raiseTo', { amount: state.nextBidAmount })}
                    </button>
                  )}
                  <button type="button" className={btnSecondary} onClick={() => handleBid(0)} disabled={loading}>
                    {t('dropOut')}
                  </button>
                </div>
              )}
              {canDiscard && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleDiscard}
                  disabled={loading || !discardReady}
                  data-testid="gleek-discard-confirm"
                >
                  {t('discardButton', { count: state.discardCount })}
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
                dataTutorial="gleek-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
