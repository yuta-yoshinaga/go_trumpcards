import { useMemo, useState } from 'react';
import type { pochApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardBack } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePochGame } from '../hooks/usePochGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PochResponse } from '../types/card';
import { PochPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { POCH_HELP, parsePochCommand } from '../utils/cli/commands/pochCommands';
import { formatPochState } from '../utils/cli/formatters/pochFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const PC_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="pc-rule"]', messageKey: 'tutorial.paySuit', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pc-board"]', messageKey: 'tutorial.carryOver', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pc-rule"]', messageKey: 'tutorial.pochen', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pc-hand"]', messageKey: 'tutorial.stops', placement: 'top', advanceOn: 'next' },
];

/** Renders the Poch page: nine pools, a pay suit, a set comparison and stops. */
export const PochPage = withTutorial(PochPageContent, 'poch', PC_TUTORIAL_STEPS);

function PochPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('poch');
  const game = usePochGame();
  const { state, loading, error, retry } = game;

  const [handIdx, setHandIdx] = useState<number | null>(null);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('poch');
  const cliConfig: CliGameConfig<PochResponse, Parameters<typeof pochApi.exec>> = useMemo(
    () => ({
      gameName: 'poch',
      parseCommand: parsePochCommand,
      formatResponse: formatPochState,
      helpText: POCH_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('poch', state);

  if (!state) {
    return <GameSkeleton gameKey="poch" layout={{ kind: 'tableau', topRow: 3, tableau: 4 }} />;
  }

  const ended = state.phase === PochPhase.GAME_END;
  const dealOver = state.phase === PochPhase.DEAL_END;
  const betting = state.phase === PochPhase.POCHEN;
  const playing = state.phase === PochPhase.STOPS;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;

  const phaseName = ended
    ? t('phase.end')
    : dealOver
      ? t('phase.dealEnd')
      : betting
        ? t('phase.pochen')
        : t('phase.stops');

  return (
    <GamePageShell
      title={tc('nav.poch')}
      gameThemeBg={gameTheme.poch.bg}
      phaseName={phaseName}
      gamePath="/poch"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('deal', { n: state.dealNo + 1, target: state.targetDeals })}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      <SettingsPanel
        title={tc('settings.title')}
        groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
      />

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Permanent, not tutorial-only: "only the pay suit pays" and
                "pochen is a comparison, not a declaration" are what a player
                gets wrong. */}
            <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="pc-rule">
              {t('ruleLine')}
            </div>

            <div className="text-center mb-3">
              <div className="text-game-text-muted text-xs mb-1">{t('turnUp')}</div>
              {state.turnUp ? (
                <div className="flex justify-center">
                  <AnimatedCard card={state.turnUp} width={cardWidth} draggable={false} />
                </div>
              ) : (
                <span className="text-game-text-muted text-xs">—</span>
              )}
            </div>

            {/* **9 区画は毎回すべて出す。**膨らんでいる区画がそのディールの
                狙いどころそのものになる。 */}
            <div className="mb-4" data-tutorial="pc-board">
              <div className="text-game-text-muted text-xs mb-1 text-center">{t('board')}</div>
              <div className="flex gap-2 justify-center flex-wrap">
                {state.pools.map((pool) => (
                  <div
                    key={`pool-${pool.name}`}
                    data-testid="poch-pool"
                    className="rounded px-2 py-1 text-xs ring-1 ring-ds-border text-center min-w-14"
                  >
                    <div className="text-ds-text-muted">{t(`pool.${pool.name}`)}</div>
                    <div className="font-medium">{pool.chips}</div>
                  </div>
                ))}
              </div>
            </div>

            {/* 第 1 段階は自動で解決するので、結果を出さないと何が起きたのか読めない。 */}
            {state.stakingAwards.length > 0 && (
              <div className="text-center text-xs mb-3" data-testid="poch-staking">
                {state.stakingAwards.map((a) => (
                  <div key={`award-${a.pool}`}>
                    {t('awardLine', { player: a.player, pool: t(`pool.${a.pool}`), chips: a.chips })}
                  </div>
                ))}
              </div>
            )}

            <div className="flex justify-center gap-4 mb-3 flex-wrap">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('seat', { name: `CPU${o.id.toString()}`, chips: o.chips, n: o.cardCount })}
                    {o.folded && ` · ${t('folded')}`}
                  </div>
                  <div
                    className="flex gap-1 justify-center flex-wrap"
                    role="img"
                    aria-label={t('opponentHandAriaLabel', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
                  >
                    {Array.from({ length: o.cardCount }, (_, i) => (
                      <CardBack key={`opp-${o.id.toString()}-c${i.toString()}`} width={cardWidth} />
                    ))}
                  </div>
                </div>
              ))}
            </div>

            {state.playedPile.length > 0 && (
              <div className="text-center mb-3">
                <div className="text-game-text-muted text-xs mb-1">
                  {t('played')}
                  {/* 止まっているかどうかで出せる札がまるで違う。 */}
                  {state.stopsSuit < 0 && ` · ${t('runStopped')}`}
                </div>
                <div className="flex gap-1 justify-center flex-wrap">
                  {state.playedPile.slice(-8).map((card, i) => (
                    <AnimatedCard
                      key={`pile-${i.toString()}`}
                      card={card}
                      width={Math.round(cardWidth * 0.7)}
                      draggable={false}
                    />
                  ))}
                </div>
              </div>
            )}

            {dealOver && state.dealWinner >= 0 && (
              <div className="text-center text-sm mb-3" data-testid="poch-deal-result">
                {t('dealResult', { n: state.dealWinner })}
              </div>
            )}
            {playing && state.pochenWinner >= 0 && (
              <div className="text-center text-sm mb-3" data-testid="poch-pochen-result">
                {t('pochenResult', { n: state.pochenWinner, pot: state.pochenPot })}
              </div>
            )}

            <div className="text-center" data-tutorial="pc-hand">
              <div className="text-game-text-muted text-xs mb-1">
                {t('yourHand')}
                {' · '}
                {t('yourChips', { n: human?.chips ?? 0 })}
                {betting && ` · ${t('yourBet', { n: human?.bet ?? 0, target: state.betTarget })}`}
              </div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => (
                  <button
                    key={`hand-${i.toString()}`}
                    type="button"
                    data-hint-action="play"
                    aria-pressed={handIdx === i}
                    aria-disabled={!isHumanTurn || !playing}
                    onClick={() => isHumanTurn && playing && setHandIdx(handIdx === i ? null : i)}
                    className={[
                      'rounded transition-transform',
                      isHumanTurn && playing ? 'hover:-translate-y-2' : 'opacity-60',
                      handIdx === i ? 'ring-2 ring-ds-accent -translate-y-2' : '',
                      frontendHintEnabled && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
                    ].join(' ')}
                  >
                    <AnimatedCard card={card} width={cardWidth} draggable={false} />
                  </button>
                ))}
              </div>
              {isHumanTurn && playing && (
                <div className="text-[10px] text-ds-text-muted mt-1">
                  {state.stopsSuit < 0 ? t('selectFreeLead') : t('selectFollow')}
                </div>
              )}
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={ended}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.poch.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isHumanTurn && betting && (
                <>
                  <button
                    type="button"
                    data-hint-action="bet"
                    className={`${btnPrimary} min-h-11`}
                    onClick={game.handleBet}
                  >
                    {t('bet')}
                  </button>
                  <button
                    type="button"
                    data-hint-action="fold"
                    className={`${btnSecondary} min-h-11`}
                    onClick={game.handleFold}
                  >
                    {t('fold')}
                  </button>
                </>
              )}
              {isHumanTurn && playing && (
                <button
                  type="button"
                  className={`${btnPrimary} min-h-11`}
                  disabled={handIdx === null}
                  onClick={() => {
                    if (handIdx !== null) {
                      game.handlePlay(handIdx);
                      setHandIdx(null);
                    }
                  }}
                >
                  {t('play')}
                </button>
              )}
              {dealOver && !ended && (
                <button type="button" className={`${btnPrimary} min-h-11`} onClick={game.handleNextDeal}>
                  {t('nextDeal')}
                </button>
              )}
              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="pc-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
