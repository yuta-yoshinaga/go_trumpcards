import { useMemo } from 'react';
import type { sjavsApi } from '../api/gameApi';
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
import { useSjavsGame } from '../hooks/useSjavsGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SjavsResponse } from '../types/card';
import { SjavsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSjavsCommand, SJAVS_HELP } from '../utils/cli/commands/sjavsCommands';
import { formatSjavsState } from '../utils/cli/formatters/sjavsFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const SJ_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sj-rule"]', messageKey: 'tutorial.trumps', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sj-bid"]', messageKey: 'tutorial.bid', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sj-hand"]', messageKey: 'tutorial.follow', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sj-score"]', messageKey: 'tutorial.rubber', placement: 'bottom', advanceOn: 'next' },
];

const SUIT_GLYPHS = ['♠', '♣', '♥', '♦'];

/** Renders the Sjavs page: six permanent trumps, a length bid, and a rubber counted down from 24. */
export const SjavsPage = withTutorial(SjavsPageContent, 'sjavs', SJ_TUTORIAL_STEPS);

function SjavsPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sjavs');
  const game = useSjavsGame();
  const { state, loading, error, retry } = game;

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sjavs');
  const cliConfig: CliGameConfig<SjavsResponse, Parameters<typeof sjavsApi.exec>> = useMemo(
    () => ({
      gameName: 'sjavs',
      parseCommand: parseSjavsCommand,
      formatResponse: formatSjavsState,
      helpText: SJAVS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('sjavs', state);

  if (!state) {
    return <GameSkeleton gameKey="sjavs" layout={{ kind: 'tableau', topRow: 3, tableau: 4 }} />;
  }

  const ended = state.phase === SjavsPhase.GAME_END;
  const bidding = state.phase === SjavsPhase.BID;
  const handOver = state.phase === SjavsPhase.HAND_END;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;

  // Playability comes from the server, which owns the following rule -- and in
  // this game "trump" is not the same thing as "a card of the trump suit".
  const playable = new Set(state.validIndices);

  // Every length the human can legally bid. Below minBid you must pass, and you
  // can never bid more than you hold.
  const bidChoices: number[] = [];
  for (let n = state.minBid; n <= state.myLongest; n++) {
    bidChoices.push(n);
  }

  const phaseName = ended ? t('phase.end') : handOver ? t('phase.handEnd') : bidding ? t('phase.bid') : t('phase.play');

  return (
    <GamePageShell
      title={tc('nav.sjavs')}
      gameThemeBg={gameTheme.sjavs.bg}
      phaseName={phaseName}
      gamePath="/sjavs"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted" data-tutorial="sj-score">
            {t('trump')}: {state.trumpSuit >= 0 ? SUIT_GLYPHS[state.trumpSuit] : t('trumpUndecided')}
            {state.trumpCount > 0 && ` (${t('trumpCount', { n: state.trumpCount })})`}
            {' / '}
            {t('remaining')}: {t('us')} {state.remaining[0]} · {t('them')} {state.remaining[1]}
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
            {/* Permanent, not tutorial-only: believing that only the trump suit
                is trump is the standard mistake in this game. */}
            <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="sj-rule">
              {t('ruleLine')}
            </div>

            <div className="text-center text-xs text-ds-text-muted mb-3">
              {t('handPoints', { a: state.teamPoints[0], b: state.teamPoints[1] })}
            </div>

            <div className="flex justify-center gap-4 mb-3 flex-wrap">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('opponentHand', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
                    {' · '}
                    {t('team', { n: o.team })}
                    {o.bid > 0 && ` · ${t('bidMark', { n: o.bid })}`}
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

            {!bidding && (
              <div className="text-center mb-4">
                <div className="text-game-text-muted text-xs mb-1">{t('trick')}</div>
                <div className="flex gap-1 justify-center flex-wrap">
                  {state.trick.length === 0 ? (
                    <span className="text-game-text-muted text-xs">—</span>
                  ) : (
                    state.trick.map((tc2) => (
                      <AnimatedCard
                        key={`trick-${tc2.playerIdx.toString()}`}
                        card={tc2.card}
                        width={cardWidth}
                        draggable={false}
                      />
                    ))
                  )}
                </div>
              </div>
            )}

            {handOver && state.handResult && (
              <div className="text-center text-sm mb-3" data-testid="sjavs-hand-result">
                {state.handResult.scoringTeam < 0
                  ? t('handTie')
                  : `${t('handScore', {
                      team: state.handResult.scoringTeam,
                      amount: state.handResult.amount,
                    })}${state.handResult.vol ? t('vol') : ''}`}
              </div>
            )}

            <div className="text-center" data-tutorial="sj-hand">
              <div className="text-game-text-muted text-xs mb-1">
                {t('yourHand')}
                {' · '}
                {t('team', { n: human?.team ?? 0 })}
                {(human?.bid ?? 0) > 0 && ` · ${t('bidMark', { n: human?.bid ?? 0 })}`}
              </div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => {
                  const canPlay = isHumanTurn && !bidding && playable.has(i);
                  return (
                    <button
                      key={`hand-${i.toString()}`}
                      type="button"
                      data-hint-action="play"
                      // Kept focusable while it cannot act so the reason is
                      // announced rather than the control leaving the tab order.
                      aria-disabled={!canPlay}
                      onClick={() => canPlay && game.handlePlay(i)}
                      className={[
                        'rounded transition-transform',
                        canPlay ? 'hover:-translate-y-2' : 'opacity-60',
                        frontendHintEnabled && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
                      ].join(' ')}
                    >
                      <AnimatedCard card={card} width={cardWidth} draggable={false} />
                    </button>
                  );
                })}
              </div>
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

          <GameFooter className={`${gameTheme.sjavs.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />

            {bidding && isHumanTurn && (
              <div className="mb-2" data-tutorial="sj-bid">
                <div className="text-xs text-ds-text-muted mb-1">
                  {t('bidPrompt', { min: state.minBid, longest: state.myLongest })}
                </div>
                <div className="flex gap-2 flex-wrap items-center">
                  {bidChoices.map((n) => (
                    <button
                      key={`bid-${n.toString()}`}
                      type="button"
                      data-hint-action="bid"
                      className={`${btnPrimary} min-h-11`}
                      onClick={() => game.handleBid(n)}
                    >
                      {t('bid', { n })}
                    </button>
                  ))}
                  {/* Pass is always available -- and it is a bid of zero, not
                      an omitted parameter. */}
                  <button
                    type="button"
                    data-hint-action="bid"
                    className={`${btnSecondary} min-h-11`}
                    onClick={() => game.handleBid(0)}
                  >
                    {t('pass')}
                  </button>
                </div>
              </div>
            )}

            <div className="flex gap-2 items-center flex-wrap">
              {handOver && !ended && (
                <button type="button" className={`${btnPrimary} min-h-11`} onClick={game.handleNextHand}>
                  {t('nextHand')}
                </button>
              )}
              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sj-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
