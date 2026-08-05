import { useMemo, useState } from 'react';
import type { popejoanApi } from '../api/gameApi';
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
import { usePopeJoanGame } from '../hooks/usePopeJoanGame';
import { btnPrimary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PopeJoanResponse } from '../types/card';
import { PopeJoanPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { POPEJOAN_HELP, parsePopeJoanCommand } from '../utils/cli/commands/popejoanCommands';
import { formatPopeJoanState } from '../utils/cli/formatters/popejoanFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const PJ_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="pj-rule"]', messageKey: 'tutorial.trumpOnly', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pj-board"]', messageKey: 'tutorial.carryOver', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pj-rule"]', messageKey: 'tutorial.eightOut', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="pj-hand"]', messageKey: 'tutorial.popeExcused', placement: 'top', advanceOn: 'next' },
];

/** Renders the Pope Joan page: eight compartments, a turn-up trump and stops. */
export const PopeJoanPage = withTutorial(PopeJoanPageContent, 'popejoan', PJ_TUTORIAL_STEPS);

function PopeJoanPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('popejoan');
  const game = usePopeJoanGame();
  const { state, loading, error, retry } = game;

  const [handIdx, setHandIdx] = useState<number | null>(null);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('popejoan');
  const cliConfig: CliGameConfig<PopeJoanResponse, Parameters<typeof popejoanApi.exec>> = useMemo(
    () => ({
      gameName: 'popejoan',
      parseCommand: parsePopeJoanCommand,
      formatResponse: formatPopeJoanState,
      helpText: POPEJOAN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('popejoan', state);

  if (!state) {
    return <GameSkeleton gameKey="popejoan" layout={{ kind: 'tableau', topRow: 3, tableau: 4 }} />;
  }

  const ended = state.phase === PopeJoanPhase.GAME_END;
  const dealOver = state.phase === PopeJoanPhase.DEAL_END;
  const playing = state.phase === PopeJoanPhase.PLAY;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && playing && state.currentPlayerIdx === 0;

  const phaseName = ended ? t('phase.end') : dealOver ? t('phase.dealEnd') : t('phase.play');

  return (
    <GamePageShell
      title={tc('nav.popejoan')}
      gameThemeBg={gameTheme.popejoan.bg}
      phaseName={phaseName}
      gamePath="/popejoan"
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
            {/* Permanent, not tutorial-only: "only trumps pay" and "the 8D is
                out, so a run always dies at the 7" are what a player gets
                wrong. */}
            <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="pj-rule">
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

            {/* **8 区画は毎回すべて出す。**膨らんでいる区画がそのディールの
                狙いどころそのものになる。 */}
            <div className="mb-4" data-tutorial="pj-board">
              <div className="text-game-text-muted text-xs mb-1 text-center">{t('board')}</div>
              <div className="flex gap-2 justify-center flex-wrap">
                {state.compartments.map((comp) => (
                  <div
                    key={`comp-${comp.name}`}
                    data-testid="popejoan-compartment"
                    className="rounded px-2 py-1 text-xs ring-1 ring-ds-border text-center min-w-14"
                  >
                    <div className="text-ds-text-muted">{t(`compartment.${comp.name}`)}</div>
                    <div className="font-medium">{comp.chips}</div>
                  </div>
                ))}
              </div>
            </div>

            {state.awards.length > 0 && (
              <div className="text-center text-xs mb-3" data-testid="popejoan-awards">
                {state.awards.map((a) => (
                  <div key={`award-${a.compartment}-${a.player.toString()}`}>
                    {a.byTurnUp
                      ? t('awardTurnUpLine', {
                          player: a.player,
                          compartment: t(`compartment.${a.compartment}`),
                          chips: a.chips,
                        })
                      : t('awardLine', {
                          player: a.player,
                          compartment: t(`compartment.${a.compartment}`),
                          chips: a.chips,
                        })}
                  </div>
                ))}
              </div>
            )}

            <div className="flex justify-center gap-4 mb-3 flex-wrap">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('seat', { name: `CPU${o.id.toString()}`, chips: o.chips, n: o.cardCount })}
                    {/* Pope 保持者は支払いを免除されるので、伏せ手でも出す。 */}
                    {o.holdsPope && ` · ${t('holdsPope')}`}
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
                  {state.runSuit < 0 && ` · ${t('runStopped')}`}
                </div>
                <div className="flex gap-1 justify-center flex-wrap">
                  {state.playedPile.slice(-10).map((card, i) => (
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
              <div className="text-center text-sm mb-3" data-testid="popejoan-deal-result">
                {t('dealResult', { n: state.dealWinner })}
              </div>
            )}

            <div className="text-center" data-tutorial="pj-hand">
              <div className="text-game-text-muted text-xs mb-1">
                {t('yourHand')}
                {' · '}
                {t('yourChips', { n: human?.chips ?? 0 })}
                {human?.holdsPope && ` · ${t('holdsPope')}`}
              </div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => {
                  // **並びに従う義務がある。**出せない札を押せてしまうと、
                  // サーバに弾かれて初めて分かる (#4934)。空リストは「情報が
                  // 無い」なので制限しない。
                  const restricting = isHumanTurn && (state.validPlays?.length ?? 0) > 0;
                  const canPlay = !restricting || state.validPlays.includes(i);
                  return (
                    <button
                      key={`hand-${i.toString()}`}
                      type="button"
                      data-hint-action="play"
                      data-unplayable={restricting && !canPlay ? 'true' : undefined}
                      aria-pressed={handIdx === i}
                      disabled={restricting && !canPlay}
                      aria-disabled={!isHumanTurn || !canPlay}
                      onClick={() => isHumanTurn && canPlay && setHandIdx(handIdx === i ? null : i)}
                      className={[
                        'rounded transition-transform',
                        restricting && !canPlay ? 'opacity-40' : '',
                        isHumanTurn ? 'hover:-translate-y-2' : 'opacity-60',
                        handIdx === i ? 'ring-2 ring-ds-accent -translate-y-2' : '',
                        frontendHintEnabled && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
                      ].join(' ')}
                    >
                      <AnimatedCard card={card} width={cardWidth} draggable={false} />
                    </button>
                  );
                })}
              </div>
              {isHumanTurn && (
                <div className="text-[10px] text-ds-text-muted mt-1">
                  {state.runSuit < 0 ? t('selectLead') : t('selectFollow')}
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

          <GameFooter className={`${gameTheme.popejoan.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {isHumanTurn && (
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
                dataTutorial="pj-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
