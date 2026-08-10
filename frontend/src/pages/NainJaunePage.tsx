import { useMemo, useState } from 'react';
import type { nainjauneApi } from '../api/gameApi';
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
import { useNainJauneGame } from '../hooks/useNainJauneGame';
import { btnPrimary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { NainJauneResponse } from '../types/card';
import { NainJaunePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { NAINJAUNE_HELP, parseNainJauneCommand } from '../utils/cli/commands/nainjauneCommands';
import { formatNainJauneState } from '../utils/cli/formatters/nainjauneFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const NJ_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="nj-rule"]', messageKey: 'tutorial.noSuit', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="nj-board"]', messageKey: 'tutorial.exactCard', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="nj-board"]', messageKey: 'tutorial.carryOver', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="nj-hand"]', messageKey: 'tutorial.points', placement: 'top', advanceOn: 'next' },
];

/** Renders the Le Nain Jaune page: five boxes, a suitless run, points at the end. */
export const NainJaunePage = withTutorial(NainJaunePageContent, 'nainjaune', NJ_TUTORIAL_STEPS);

function NainJaunePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('nainjaune');
  const game = useNainJauneGame();
  const { state, loading, error, retry } = game;

  const [handIdx, setHandIdx] = useState<number | null>(null);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('nainjaune');
  const cliConfig: CliGameConfig<NainJauneResponse, Parameters<typeof nainjauneApi.exec>> = useMemo(
    () => ({
      gameName: 'nainjaune',
      parseCommand: parseNainJauneCommand,
      formatResponse: formatNainJauneState,
      helpText: NAINJAUNE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('nainjaune', state);

  if (!state) {
    return <GameSkeleton gameKey="nainjaune" layout={{ kind: 'tableau', topRow: 3, tableau: 4 }} />;
  }

  const ended = state.phase === NainJaunePhase.GAME_END;
  const dealOver = state.phase === NainJaunePhase.DEAL_END;
  const playing = state.phase === NainJaunePhase.PLAY;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && playing && state.currentPlayerIdx === 0;

  const phaseName = ended ? t('phase.end') : dealOver ? t('phase.dealEnd') : t('phase.play');

  return (
    <GamePageShell
      title={tc('nav.nainjaune')}
      gameThemeBg={gameTheme.nainjaune.bg}
      phaseName={phaseName}
      gamePath="/nainjaune"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('deal', { n: state.dealNo + 1, target: state.targetDeals })}
            {' · '}
            {t('talon', { n: state.talonCount })}
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
            {/* Permanent, not tutorial-only: the run ignoring suit and paying in
                points are what a player coming from Pope Joan gets wrong. */}
            <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="nj-rule">
              {t('ruleLine')}
            </div>

            {/* **5 区画は毎回すべて出す。**取る札もスートまで見せる。 */}
            <div className="mb-4" data-tutorial="nj-board">
              <div className="text-game-text-muted text-xs mb-1 text-center">{t('board')}</div>
              <div className="flex gap-2 justify-center flex-wrap">
                {state.boxes.map((box) => (
                  <div
                    key={`box-${box.name}`}
                    data-testid="nainjaune-box"
                    className="rounded px-2 py-1 text-xs ring-1 ring-ds-border text-center"
                  >
                    <div className="text-ds-text-muted">{t(`box.${box.name}`)}</div>
                    {box.card && (
                      <div className="flex justify-center my-1">
                        <AnimatedCard card={box.card} width={Math.round(cardWidth * 0.55)} draggable={false} />
                      </div>
                    )}
                    <div className="font-medium">{box.chips}</div>
                  </div>
                ))}
              </div>
            </div>

            {state.awards.length > 0 && (
              <div className="text-center text-xs mb-3" data-testid="nainjaune-awards">
                {state.awards.map((a) => (
                  <div key={`award-${a.box}-${a.player.toString()}`}>
                    {t('awardLine', { player: a.player, box: t(`box.${a.box}`), chips: a.chips })}
                  </div>
                ))}
              </div>
            )}

            <div className="flex justify-center gap-4 mb-3 flex-wrap">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {/* 点も出す。支払いは枚数ではなく点数。 */}
                    {t('seat', { name: `CPU${o.id.toString()}`, chips: o.chips, n: o.cardCount, pts: o.points })}
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
                <div className="text-game-text-muted text-xs mb-1">{t('played')}</div>
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
              <div className="text-center text-sm mb-3" data-testid="nainjaune-deal-result">
                {t('dealResult', { n: state.dealWinner })}
              </div>
            )}

            <div className="text-center" data-tutorial="nj-hand">
              <div className="text-game-text-muted text-xs mb-1">
                {t('yourHand')}
                {' · '}
                {t('yourChips', { n: human?.chips ?? 0 })}
                {' · '}
                {t('yourPoints', { n: human?.points ?? 0 })}
              </div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => {
                  // **並びに従う義務がある。**出せない札を押せてしまうと、
                  // サーバに弾かれて初めて分かる (#4935)。空リストは「情報が
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
                  {/* 止まっているかどうかで出せる札がまるで違う。 */}
                  {state.runRank === 0 ? t('selectLead') : t('selectFollow', { rank: state.runRank + 1 })}
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

          <GameFooter className={`${gameTheme.nainjaune.footer} px-4 py-2.5`}>
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
                dataTutorial="nj-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
