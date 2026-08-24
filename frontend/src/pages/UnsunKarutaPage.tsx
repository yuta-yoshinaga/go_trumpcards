import { useEffect, useMemo } from 'react';
import type { unsunKarutaApi } from '../api/gameApi';
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
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  UNSUN_CPU_DIFFICULTY_OPTIONS,
  UNSUN_TARGET_DEALS_OPTIONS,
  useUnsunKarutaGame,
} from '../hooks/useUnsunKarutaGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { UnsunKarutaResponse } from '../types/card';
import { UnsunKarutaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseUnsunKarutaCommand, UNSUN_KARUTA_HELP } from '../utils/cli/commands/unsunKarutaCommands';
import { formatUnsunKarutaState } from '../utils/cli/formatters/unsunKarutaFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Unsun Karuta tutorial step definitions. */
const UNSUN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="unsunkaruta-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="unsunkaruta-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="unsunkaruta-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="unsunkaruta-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="unsunkaruta-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const UNSUN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [UnsunKarutaPhase.PLAY]: 'play',
  [UnsunKarutaPhase.TRICK_END]: 'trickEnd',
  [UnsunKarutaPhase.ROUND_END]: 'roundEnd',
  [UnsunKarutaPhase.GAME_END]: 'gameEnd',
};

/**
 * Renders the Unsun Karuta (八人メリ) page: the 75-card five-suit Japanese
 * trick-taker for eight players in two teams.
 */
export const UnsunKarutaPage = withTutorial(UnsunKarutaPageContent, 'unsunkaruta', UNSUN_TUTORIAL_STEPS);

/** Inner content of the page, wrapped by TutorialProvider. */
function UnsunKarutaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('unsunkaruta');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    unsunConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useUnsunKarutaGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('unsunkaruta');
  const cliConfig: CliGameConfig<UnsunKarutaResponse, Parameters<typeof unsunKarutaApi.exec>> = useMemo(
    () => ({
      gameName: 'unsunkaruta',
      parseCommand: parseUnsunKarutaCommand,
      formatResponse: formatUnsunKarutaState,
      helpText: UNSUN_KARUTA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('unsunkaruta', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('unsunkaruta', UNSUN_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="unsunkaruta" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 9 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isPlayPhase = state.phase === UnsunKarutaPhase.PLAY;
  const isTrickEnd = state.phase === UnsunKarutaPhase.TRICK_END;
  const isRoundEnd = state.phase === UnsunKarutaPhase.ROUND_END;
  const isGameEnd = state.phase === UnsunKarutaPhase.GAME_END || state.gameEndFlag;
  const canPlay = isPlayPhase && isHumanTurn;

  const handValidIndices = canPlay ? state.playableIndices : undefined;
  const teamTricks = state.teamTricks ?? [];
  const teamScores = state.teamScores ?? [];

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.unsunkaruta')}
      gameThemeBg={gameTheme.unsunkaruta.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/unsunkaruta"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerTeam === state.humanTeam}
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
                    value: unsunConfig.cpuDifficulty,
                    options: UNSUN_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetDeals',
                    label: t('settings.targetDeals'),
                    value: unsunConfig.targetDeals,
                    options: UNSUN_TARGET_DEALS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
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
              <span className="mr-4">{t('trick', { n: state.trickNumber, total: state.trickCount })}</span>
              {/* **切り札は返した 1 枚で決まる。** 出さないとどのスートが強いのか
                  画面のどこにも無い。 */}
              <span className="inline-flex items-center gap-1" data-testid="unsunkaruta-trump">
                {t('trump', { suit: t(`suit.${state.trumpSuitName}`) })}
                {state.trumpCard && <CardImage card={state.trumpCard} width={Math.round(cardWidth * 0.6)} />}
              </span>
            </div>

            <div className={lgTwoColGrid}>
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="unsunkaruta-trick-display"
                />
              </div>

              <div data-tutorial="unsunkaruta-info">
                {/* チーム別の「コ」。**席番号では味方が分からない** ので、
                    組ごとにまとめて出す。 */}
                <div
                  className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                  data-testid="unsunkaruta-teams"
                >
                  <div>
                    {t('teamTricks', { a: teamTricks[state.humanTeam] ?? 0, b: teamTricks[1 - state.humanTeam] ?? 0 })}
                  </div>
                  <div>
                    {t('teamScores', { a: teamScores[state.humanTeam] ?? 0, b: teamScores[1 - state.humanTeam] ?? 0 })}
                  </div>
                </div>

                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)} ({t('team', { n: p.team })}):{' '}
                          {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5 flex items-center gap-2">
                        <span className={p.team === state.humanTeam ? 'text-ds-success' : ''}>
                          {playerName(p.id, p.isHuman)} ({t('team', { n: p.team })}):{' '}
                          {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                        </span>
                        {p.isDealer && (
                          <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                            {t('dealerBadge')}
                          </span>
                        )}
                      </div>
                    ))}
                  </div>
                )}

                {(isRoundEnd || isGameEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="unsunkaruta-result"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>
                      {t('roundResult.tricks', {
                        a: teamTricks[state.humanTeam] ?? 0,
                        b: teamTricks[1 - state.humanTeam] ?? 0,
                      })}
                    </div>
                    {isGameEnd && (
                      <div className="text-ds-text-primary" data-testid="unsunkaruta-winner">
                        {state.winnerTeam < 0
                          ? t('roundResult.draw')
                          : t('roundResult.winner', { team: state.winnerTeam })}
                      </div>
                    )}
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

          <GameFooter className={`${gameTheme.unsunkaruta.footer} px-4 py-2.5`}>
            {/* **フォロー義務は宣言で生まれる。** 出ているかどうかを書かないと、
                なぜこの札しか押せないのかが読めない。 */}
            {canPlay && state.mustFollow && (
              <div className="mb-1 text-center text-sm text-ds-accent" data-testid="unsunkaruta-must-follow">
                {t('mustFollow')}
              </div>
            )}
            {canPlay && state.canDeclare && (
              <div className="mb-1 text-center text-sm text-ds-text-muted" data-testid="unsunkaruta-can-declare">
                {t('canDeclare')}
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="unsunkaruta"
                validIndices={handValidIndices}
                legalIndices={handValidIndices}
                restrictedTooltip={t('mustFollow')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* ライブ領域は常設 (#5955)。 */}
            <div data-testid="unsunkaruta-hint-live" role="status" aria-live="polite">
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="unsunkaruta-action-buttons">
              {canPlay && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => handlePlay(false)}
                  disabled={loading || selectedCardIndices.length !== 1}
                  data-testid="unsunkaruta-play"
                >
                  {t('playButton')}
                </button>
              )}
              {/* 宣言できるのはリードのときだけ。 */}
              {canPlay && state.canDeclare && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={() => handlePlay(true)}
                  disabled={loading || selectedCardIndices.length !== 1}
                  data-testid="unsunkaruta-declare"
                >
                  {t('declareButton')}
                </button>
              )}
              {isTrickEnd && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNextTrick}
                  disabled={loading}
                  data-testid="unsunkaruta-next-trick"
                >
                  {t('nextTrick')}
                </button>
              )}
              {isRoundEnd && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNextRound}
                  disabled={loading}
                  data-testid="unsunkaruta-next-round"
                >
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="unsunkaruta-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
