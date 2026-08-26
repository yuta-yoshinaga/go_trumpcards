import { useEffect, useMemo } from 'react';
import type { piedmonteseTarotApi } from '../api/gameApi';
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
  PIEDMONTESE_CPU_DIFFICULTY_OPTIONS,
  PIEDMONTESE_SEAT_OPTIONS,
  PIEDMONTESE_TARGET_DEALS_OPTIONS,
  usePiedmonteseTarotGame,
} from '../hooks/usePiedmonteseTarotGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PiedmonteseTarotResponse } from '../types/card';
import { PiedmonteseTarotPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { PIEDMONTESE_TAROT_HELP, parsePiedmonteseTarotCommand } from '../utils/cli/commands/piedmonteseTarotCommands';
import { formatPiedmonteseTarotState } from '../utils/cli/formatters/piedmonteseTarotFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
// **捨てられない理由の判定は共有。** スカルト (捨て札) の規則は 3 人版と同じで、
// 判定はタロット札の面 (色と値) だけを見ている。
import { scartoUndiscardableReason } from '../utils/scartoDiscard';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Tarocco Piemontese tutorial step definitions. */
const PIEDMONTESE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="piedmontesetarot-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="piedmontesetarot-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="piedmontesetarot-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="piedmontesetarot-info"]',
    messageKey: 'tutorial.info',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="piedmontesetarot-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PIEDMONTESE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PiedmonteseTarotPhase.SCARTO]: 'scarto',
  [PiedmonteseTarotPhase.PLAY]: 'play',
  [PiedmonteseTarotPhase.TRICK_END]: 'trickEnd',
  [PiedmonteseTarotPhase.ROUND_END]: 'roundEnd',
  [PiedmonteseTarotPhase.GAME_END]: 'gameEnd',
};

/** Outcome i18n keys indexed by outcome value (0=none, 1=above average, 2=below). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeWin', 'outcomeLoss'] as const;

/** Formats a signed settlement, prefixing a leading `+` for positive values. */
function formatSigned(n: number): string {
  return n > 0 ? `+${n}` : String(n);
}

/**
 * Renders the Tarocco Piemontese page: the Piedmontese 78-card tarot for three
 * or four players, with the dealer burying the talon and trump-priority tricks.
 */
export const PiedmonteseTarotPage = withTutorial(
  PiedmonteseTarotPageContent,
  'piedmontesetarot',
  PIEDMONTESE_TUTORIAL_STEPS,
);

/** Inner content of the page, wrapped by TutorialProvider. */
function PiedmonteseTarotPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('piedmontesetarot');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    piedmonteseConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleScarto,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = usePiedmonteseTarotGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('piedmontesetarot');
  const cliConfig: CliGameConfig<PiedmonteseTarotResponse, Parameters<typeof piedmonteseTarotApi.exec>> = useMemo(
    () => ({
      gameName: 'piedmontesetarot',
      parseCommand: parsePiedmonteseTarotCommand,
      formatResponse: formatPiedmonteseTarotState,
      helpText: PIEDMONTESE_TAROT_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('piedmontesetarot', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('piedmontesetarot', PIEDMONTESE_PHASE_KEYS);

  if (!state)
    return (
      <GameSkeleton gameKey="piedmontesetarot" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 19 }} />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isScartoPhase = state.phase === PiedmonteseTarotPhase.SCARTO;
  const isPlayPhase = state.phase === PiedmonteseTarotPhase.PLAY;
  const isTrickEnd = state.phase === PiedmonteseTarotPhase.TRICK_END;
  const isRoundEnd = state.phase === PiedmonteseTarotPhase.ROUND_END;
  const isGameEnd = state.phase === PiedmonteseTarotPhase.GAME_END || state.gameEndFlag;

  const canScarto = isScartoPhase && state.isHumanScarto;
  const canPlay = isPlayPhase && isHumanTurn;
  // **捨てる枚数は卓が決める。** 4 人なら 2 枚、3 人なら 3 枚。定数にすると
  // どちらかの卓で必ずボタンが押せなくなる。
  const talonSize = state.talonSize;

  // 78 点を席で分け合うゼロサム精算なので、卓の合計と各席の取り分を出して
  // 変動の出どころを見せる。
  const totalThirds = state.players.reduce((sum, p) => sum + p.cardThirds, 0);

  // **捨てられる札はサーバが決める。** 色と値からここで組み立てると、
  // 捨てられるピップが足りないときに解禁される非オヌール切り札が落ちる。
  // その手を引いた親は、画面から枚数を揃えられなかった (#6236)。
  const discardableIndices = canScarto ? state.discardableIndices : undefined;

  const handValidIndices = canPlay ? state.playableIndices : canScarto ? discardableIndices : undefined;

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
      title={tc('nav.piedmontesetarot')}
      gameThemeBg={gameTheme.piedmontesetarot.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/piedmontesetarot"
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
                    id: 'seats',
                    label: t('settings.seats'),
                    value: piedmonteseConfig.seats,
                    options: PIEDMONTESE_SEAT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('seats', v),
                  },
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: piedmonteseConfig.cpuDifficulty,
                    options: PIEDMONTESE_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetDeals',
                    label: t('settings.targetDeals'),
                    value: piedmonteseConfig.targetDeals,
                    options: PIEDMONTESE_TARGET_DEALS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
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
              <span>{t('trick', { n: state.trickNumber, total: state.trickCount })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="piedmontesetarot-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="piedmontesetarot-info">
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

                {(isRoundEnd || isGameEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="piedmontesetarot-result"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.outcome', { outcome: t(OUTCOME_KEYS[state.outcome] ?? 'outcomeNone') })}</div>
                    {state.players.map((p, i) => {
                      const delta = state.dealScores[i] ?? 0;
                      return (
                        <div key={p.id}>
                          {t('roundResult.playerLine', {
                            name: playerName(p.id, p.isHuman),
                            delta: formatSigned(delta),
                            score: p.score,
                          })}
                        </div>
                      );
                    })}
                    <div className="mt-1 pt-1 border-t border-white/10" data-testid="piedmontesetarot-breakdown">
                      <div>{t('roundResult.total', { total: totalThirds / 3 })}</div>
                      {/* **式を書く。** 「取り分」と「変動」は席数倍ちがうので、
                          並べただけでは計算が合わないように見える。 */}
                      <div data-testid="piedmontesetarot-formula">
                        {t('roundResult.formula', { n: state.players.length })}
                      </div>
                      {state.players.map((p, i) => (
                        <div key={p.id}>
                          {t('roundResult.earnedLine', {
                            name: playerName(p.id, p.isHuman),
                            points: p.cardPoints,
                            scaled: formatSigned(state.dealScores[i] ?? 0),
                          })}
                        </div>
                      ))}
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

          <GameFooter className={`${gameTheme.piedmontesetarot.footer} px-4 py-2.5`}>
            {canScarto && (
              <div
                className="mb-1 text-center text-sm text-ds-accent font-semibold"
                data-testid="piedmontesetarot-discard-prompt"
              >
                {t('scartoPhase', { count: selectedCardIndices.length, total: talonSize })}
              </div>
            )}
            {isScartoPhase && !state.isHumanScarto && (
              <div className="mb-1 text-center text-sm text-ds-text-muted" data-testid="piedmontesetarot-waiting">
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
                dataTutorialPrefix="piedmontesetarot"
                validIndices={handValidIndices}
                // **出せる札には輪を付ける。** `validIndices` は「押せない札を
                // 薄くする」だけで、`data-legal` (と成功色のリング) を付けるのは
                // `legalIndices` のほう。片方だけ渡すと、出せる札が画面上で
                // まったく強調されない。
                legalIndices={handValidIndices}
                restrictedTooltip={canScarto ? t('scartoRestricted') : t('playButton')}
                cardTitleFor={canScarto ? scartoTitleFor : undefined}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* ライブ領域は常設。hint がある間だけ現れる div に付けると読み上げられない (#5955)。 */}
            <div data-testid="piedmontesetarot-hint-live" role="status" aria-live="polite">
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="piedmontesetarot-action-buttons">
              {canScarto && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => handleScarto(talonSize)}
                  disabled={loading || selectedCardIndices.length !== talonSize}
                >
                  {t('discardButton', { count: selectedCardIndices.length, total: talonSize })}
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
                dataTutorial="piedmontesetarot-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
