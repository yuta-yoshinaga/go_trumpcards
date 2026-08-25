import { useEffect, useMemo } from 'react';
import type { dehlaPakadApi } from '../api/gameApi';
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
import {
  DEHLA_PAKAD_CPU_DIFFICULTY_OPTIONS,
  DEHLA_PAKAD_KOT_OPTIONS,
  DEHLA_PAKAD_SUITS,
  useDehlaPakadGame,
} from '../hooks/useDehlaPakadGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DehlaPakadResponse } from '../types/card';
import { DehlaPakadPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { DEHLA_PAKAD_HELP, parseDehlaPakadCommand } from '../utils/cli/commands/dehlaPakadCommands';
import { formatDehlaPakadState } from '../utils/cli/formatters/dehlaPakadFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Dehla Pakad tutorial step definitions. */
const DEHLA_PAKAD_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="dehlapakad-centre"]',
    messageKey: 'tutorial.centre',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dehlapakad-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dehlapakad-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dehlapakad-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dehlapakad-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

// **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしているので
// ここは直接引く ── 数値化すると全フェーズが NaN に潰れて名前が消える。
const DEHLA_PAKAD_PHASE_KEYS: Readonly<Record<string, string>> = {
  [DehlaPakadPhase.SELECT_TRUMP]: 'selectTrump',
  [DehlaPakadPhase.PLAY]: 'play',
  [DehlaPakadPhase.HAND_END]: 'handEnd',
  [DehlaPakadPhase.GAME_END]: 'gameEnd',
};

/**
 * Renders the Dehla Pakad page: the Indian "catch the tens" partnership
 * trick-taker where cards are gathered only on two consecutive tricks.
 */
export const DehlaPakadPage = withTutorial(DehlaPakadPageContent, 'dehlapakad', DEHLA_PAKAD_TUTORIAL_STEPS);

/** Inner content of the page, wrapped by TutorialProvider. */
function DehlaPakadPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('dehlapakad');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    dehlaPakadConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleSelectTrump,
    handlePlay,
    handleNextHand,
  } = useDehlaPakadGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('dehlapakad');
  const cliConfig: CliGameConfig<DehlaPakadResponse, Parameters<typeof dehlaPakadApi.exec>> = useMemo(
    () => ({
      gameName: 'dehlapakad',
      parseCommand: parseDehlaPakadCommand,
      formatResponse: formatDehlaPakadState,
      helpText: DEHLA_PAKAD_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('dehlapakad', state);
  const { cardWidth, isMobile } = useCardDimensions();

  if (!state)
    return <GameSkeleton gameKey="dehlapakad" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isTrumpPhase = state.phase === DehlaPakadPhase.SELECT_TRUMP;
  const isPlayPhase = state.phase === DehlaPakadPhase.PLAY;
  const isHandEnd = state.phase === DehlaPakadPhase.HAND_END;
  const isGameEnd = state.phase === DehlaPakadPhase.GAME_END || state.gameEndFlag;
  const canPlay = isPlayPhase && isHumanTurn;
  const canCallTrump = isTrumpPhase && isHumanTurn;

  const handValidIndices = canPlay ? state.playableIndices : undefined;
  const tens = state.teamTens ?? [];
  const kots = state.teamKots ?? [];
  const you = state.humanTeam;

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.dehlapakad')}
      gameThemeBg={gameTheme.dehlapakad.bg}
      phaseName={t(`phase.${DEHLA_PAKAD_PHASE_KEYS[state.phase] ?? 'play'}`)}
      isHumanTurn={isHumanTurn}
      gamePath="/dehlapakad"
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
                    value: dehlaPakadConfig.cpuDifficulty,
                    options: DEHLA_PAKAD_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetKots',
                    label: t('settings.targetKots'),
                    value: dehlaPakadConfig.targetKots,
                    options: DEHLA_PAKAD_KOT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetKots', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('hand', { n: state.handNumber, target: state.config.targetKots })}</span>
              {state.trumpSuit > 0 && (
                <>
                  <span className="mr-4" data-testid="dehlapakad-trump">
                    {t('trump', { suit: t(`suit.${state.trumpSuitName}`) })}
                  </span>
                  <span>{t('trick', { n: state.trickNumber, total: state.trickCount })}</span>
                </>
              )}
            </div>

            <div className={lgTwoColGrid}>
              <div data-tutorial="dehlapakad-centre">
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="dehlapakad-trick-display"
                />
                {/* **これがこのゲームの心臓部。** 取っただけでは札は手に入らず、
                    同じ席が 2 トリック続けて取ってはじめて山ごと引き取れる。
                    山に 10 が何枚乗っているかが見えないと、いま何を賭けて
                    いるのかが読めない。 */}
                {state.centrePileCount > 0 && (
                  <div
                    className="mt-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="dehlapakad-centre-pile"
                  >
                    <div className="text-ds-text-primary">
                      {t('centrePile', { n: state.centrePileCount, tens: state.centrePileTens })}
                    </div>
                    {state.prevTrickWinner >= 0 && (
                      <div data-testid="dehlapakad-pile-goes-to">
                        {t('pileGoesTo', {
                          name: playerName(state.prevTrickWinner, state.prevTrickWinner === 0),
                        })}
                      </div>
                    )}
                  </div>
                )}
              </div>

              <div data-tutorial="dehlapakad-scores">
                <div
                  className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                  data-testid="dehlapakad-scores"
                >
                  <div>{t('teamTens', { a: tens[you] ?? 0, b: tens[1 - you] ?? 0 })}</div>
                  <div>{t('teamKots', { a: kots[you] ?? 0, b: kots[1 - you] ?? 0 })}</div>
                  {/* **7 連勝もコートになる。** 出さないと、なぜ同じ組が勝ち
                      続けているのかが数字にならない。 */}
                  {state.streakCount > 1 && state.streakTeam >= 0 && (
                    <div className="text-ds-accent" data-testid="dehlapakad-streak">
                      {t('streak', { team: state.streakTeam, n: state.streakCount })}
                    </div>
                  )}
                </div>

                <div className="mb-2 p-2 rounded bg-black/30">
                  {state.players.map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5 flex items-center gap-2">
                      <span className={p.team === state.humanTeam ? 'text-ds-success' : ''}>
                        {playerName(p.id, p.isHuman)} ({t('team', { n: p.team })}): {t('cards', { count: p.cardCount })}
                      </span>
                      {p.isDealer && (
                        <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                          {t('dealerBadge')}
                        </span>
                      )}
                    </div>
                  ))}
                </div>

                {/* **決めるのは親の右隣で、見えているのは最初の 5 枚だけ。** */}
                {canCallTrump && (
                  <div className="mb-2 p-2 rounded bg-black/30" data-testid="dehlapakad-trump-choices">
                    <div className="text-ds-text-primary text-sm mb-1">{t('callTrump')}</div>
                    <div className="flex flex-wrap gap-2">
                      {DEHLA_PAKAD_SUITS.map((s) => (
                        <button
                          key={s.value}
                          type="button"
                          className={btnPrimary}
                          onClick={() => handleSelectTrump(s.value)}
                          disabled={loading}
                          data-testid={`dehlapakad-trump-${s.value}`}
                        >
                          {t(`suit.${s.key}`)}
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                {state.lastHand && (isHandEnd || isGameEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="dehlapakad-hand-result"
                  >
                    <div className="text-ds-text-primary">
                      {t('handResult', {
                        team: state.lastHand.winnerTeam,
                        a: state.lastHand.teamTens[0],
                        b: state.lastHand.teamTens[1],
                      })}
                    </div>
                    {state.lastHand.kot && (
                      <div className="text-ds-warning" data-testid="dehlapakad-kot">
                        {t(`kot.${state.lastHand.kotReason}`)}
                      </div>
                    )}
                  </div>
                )}

                {isGameEnd && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="dehlapakad-winner">
                    {t('winner', { team: state.winnerTeam })}
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

          <GameFooter className={`${gameTheme.dehlapakad.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="dehlapakad"
                validIndices={handValidIndices}
                legalIndices={handValidIndices}
                restrictedTooltip={t('restrictedTooltip')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* ライブ領域は常設 (#5955)。 */}
            <div data-testid="dehlapakad-hint-live" role="status" aria-live="polite">
              {isRequestedHint(state) && (state.hint || state.hintTrumpSuit > 0) && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}
                  {state.hint && `: ${t(`hint.${state.hint.reason}`)}`}
                  {state.hint?.cardIndices &&
                    state.hint.cardIndices.length > 0 &&
                    ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="dehlapakad-action-buttons">
              {canPlay && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices.length !== 1}
                  data-testid="dehlapakad-play"
                >
                  {t('playButton')}
                </button>
              )}
              {isHandEnd && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNextHand}
                  disabled={loading}
                  data-testid="dehlapakad-next-hand"
                >
                  {t('nextHand')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="dehlapakad-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
