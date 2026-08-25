import { useEffect, useMemo } from 'react';
import type { cometApi } from '../api/gameApi';
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
  COMET_CPU_DIFFICULTY_OPTIONS,
  COMET_PLAYER_OPTIONS,
  COMET_TARGET_OPTIONS,
  useCometGame,
} from '../hooks/useCometGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CometResponse } from '../types/card';
import { CometPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { COMET_HELP, parseCometCommand } from '../utils/cli/commands/cometCommands';
import { formatCometState } from '../utils/cli/formatters/cometFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Cards of the current sequence to show; older ones scroll off. */
const PILE_TAIL = 8;

/** Comet tutorial step definitions. */
const COMET_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="comet-pile"]',
    messageKey: 'tutorial.pile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="comet-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="comet-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="comet-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="comet-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

// **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしている。
const COMET_PHASE_KEYS: Readonly<Record<string, string>> = {
  [CometPhase.PLAY]: 'play',
  [CometPhase.ROUND_END]: 'roundEnd',
  [CometPhase.GAME_END]: 'gameEnd',
};

/**
 * Renders the Comet page: the ancestor of the stops family, where sequences
 * climb by rank ignoring suit and the 9 of diamonds is a wild Comet.
 */
export const CometPage = withTutorial(CometPageContent, 'comet', COMET_TUTORIAL_STEPS);

/** Inner content of the page, wrapped by TutorialProvider. */
function CometPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('comet');
  const { state, loading, error, exec, retry, cometConfig, handleConfigChange, play, pass, handleNextRound, reset } =
    useCometGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('comet');
  const cliConfig: CliGameConfig<CometResponse, Parameters<typeof cometApi.exec>> = useMemo(
    () => ({
      gameName: 'comet',
      parseCommand: parseCometCommand,
      formatResponse: formatCometState,
      helpText: COMET_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('comet', state);
  const { cardWidth, isMobile } = useCardDimensions();

  if (!state)
    return (
      <GameSkeleton
        gameKey="comet"
        layout={{ kind: 'casino-table', sections: [1], footerStyle: 'hand', footerHandSize: 8 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === CometPhase.PLAY;
  const isRoundEnd = state.phase === CometPhase.ROUND_END;
  const isGameEnd = state.phase === CometPhase.GAME_END || state.gameEndFlag;
  const canPlay = isPlayPhase && state.isHumanTurn;

  // **出せる札はサーバが数えたものだけ。** コメットがどのランクの代わりにも
  // なるので、画面側で組み直すと必ずずれる。
  const playable = canPlay ? state.playableIdxs : [];
  const mustPass = canPlay && playable.length === 0;
  const shownPile = state.pile.slice(-PILE_TAIL);

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.comet')}
      gameThemeBg={gameTheme.comet.bg}
      phaseName={t(`phase.${COMET_PHASE_KEYS[state.phase] ?? 'play'}`)}
      isHumanTurn={state.isHumanTurn}
      gamePath="/comet"
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
                    value: cometConfig.cpuDifficulty,
                    options: COMET_CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'players',
                    label: t('settings.players'),
                    value: cometConfig.players,
                    options: COMET_PLAYER_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('players', v),
                  },
                  {
                    type: 'select',
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: cometConfig.targetScore,
                    options: COMET_TARGET_OPTIONS.map((v) => ({ value: v, label: String(v) })),
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
              {/* **死に手の枚数は見せる。** ここに眠った札で連なりが止まる。 */}
              <span data-testid="comet-dead">{t('dead', { n: state.deadCount })}</span>
            </div>

            <div className={lgTwoColGrid}>
              <div data-tutorial="comet-pile">
                <div className="mb-1 text-ds-text-muted text-sm">{t('pileLabel')}</div>
                <div
                  className="mb-2 p-2 rounded bg-black/30 flex flex-wrap gap-1 items-center"
                  data-testid="comet-pile"
                >
                  {shownPile.length === 0 ? (
                    <span className="text-ds-text-muted text-sm">{t('pileEmpty')}</span>
                  ) : (
                    shownPile.map((c, i) => (
                      <CardImage key={`${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                    ))
                  )}
                </div>
                {/* **スートは問わない。** 数字だけで昇ることを毎回書く。 */}
                <div className="text-ds-text-primary text-sm" data-testid="comet-need">
                  {state.need > 0 ? t('need', { rank: state.need }) : t('needAny')}
                </div>
              </div>

              <div data-tutorial="comet-scores">
                <div className="mb-2 p-2 rounded bg-black/30" data-testid="comet-scores">
                  {state.players.map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                      <span className="flex items-center gap-2">
                        <span>
                          {playerName(p.id, p.isHuman)}: {t('cards', { n: p.cardCount })} / {t('score', { n: p.score })}
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

                {state.lastResult && (isRoundEnd || isGameEnd) && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="comet-round-result"
                  >
                    <div className="mb-1 text-ds-text-primary">
                      {t('goOut', {
                        name: playerName(state.lastResult.winnerIdx, state.lastResult.winnerIdx === 0),
                        points: state.lastResult.gained[state.lastResult.winnerIdx],
                      })}
                    </div>
                    <div>{t('unplayedKings', { n: state.lastResult.unplayedKings })}</div>
                    {state.lastResult.heldWildIdx >= 0 && (
                      <div className="text-ds-danger" data-testid="comet-held-wild">
                        {t('heldWild', {
                          name: playerName(state.lastResult.heldWildIdx, state.lastResult.heldWildIdx === 0),
                        })}
                      </div>
                    )}
                  </div>
                )}

                {isGameEnd && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="comet-winner">
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

          <GameFooter className={`${gameTheme.comet.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={[]}
                toggleCard={play}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="comet"
                validIndices={canPlay ? playable : undefined}
                legalIndices={canPlay ? playable : undefined}
                restrictedTooltip={t('restrictedTooltip')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* ライブ領域は常設 (#5955)。 */}
            <div data-testid="comet-hint-live" role="status" aria-live="polite">
              {isRequestedHint(state) && state.hintHandIdx >= 0 && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {t(`hint.${state.hintReason}`)} ([{state.hintHandIdx}])
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="comet-action-buttons">
              {/* **パスは出せる札が無いときだけ。** 常に出しておくと、出せるのに
                  押して弾かれるだけのボタンになる。 */}
              {mustPass && (
                <button
                  type="button"
                  className={btnSecondary}
                  onClick={pass}
                  disabled={loading}
                  data-testid="comet-pass"
                >
                  {t('passButton')}
                </button>
              )}
              {isRoundEnd && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleNextRound}
                  disabled={loading}
                  data-testid="comet-next-round"
                >
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="comet-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
