import { useMemo } from 'react';
import type { skitgubbeApi } from '../api/gameApi';
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
import { useSkitgubbeGame } from '../hooks/useSkitgubbeGame';
import { gameTheme } from '../styles/gameTheme';
import type { SkitgubbeResponse } from '../types/card';
import { SkitgubbePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSkitgubbeCommand, SKITGUBBE_HELP } from '../utils/cli/commands/skitgubbeCommands';
import { formatSkitgubbeState } from '../utils/cli/formatters/skitgubbeFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const SG_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sg-rule"]', messageKey: 'tutorial.phases', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sg-table"]', messageKey: 'tutorial.duel', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sg-table"]', messageKey: 'tutorial.stunsa', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sg-seats"]', messageKey: 'tutorial.collected', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sg-hand"]', messageKey: 'tutorial.beat', placement: 'top', advanceOn: 'next' },
];

/** Renders the Skitgubbe page: duel for cards, then shed what you collected. */
export const SkitgubbePage = withTutorial(SkitgubbePageContent, 'skitgubbe', SG_TUTORIAL_STEPS);

function SkitgubbePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('skitgubbe');
  const game = useSkitgubbeGame();
  const { state, loading, error, retry } = game;

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('skitgubbe');
  const cliConfig: CliGameConfig<SkitgubbeResponse, Parameters<typeof skitgubbeApi.exec>> = useMemo(
    () => ({
      gameName: 'skitgubbe',
      parseCommand: parseSkitgubbeCommand,
      formatResponse: formatSkitgubbeState,
      helpText: SKITGUBBE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('skitgubbe', state);

  if (!state) {
    return <GameSkeleton gameKey="skitgubbe" layout={{ kind: 'tableau', topRow: 2, tableau: 3 }} />;
  }

  const ended = state.phase === SkitgubbePhase.GAME_END;
  const collecting = state.phase === SkitgubbePhase.COLLECT;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;

  // Playability comes from the server, which owns the beat rule. Re-deriving
  // "higher of the same suit, or a trump" here would put it in two places.
  const playable = new Set(state.validIndices);
  const table = collecting ? state.duel : state.pile;

  return (
    <GamePageShell
      title={tc('nav.skitgubbe')}
      gameThemeBg={gameTheme.skitgubbe.bg}
      phaseName={ended ? t('phase.end') : t('phase.play')}
      gamePath="/skitgubbe"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('stock')}: {state.stockCount} / {t('trump')}:{' '}
            {state.trumpSuit >= 0 ? ['♠', '♣', '♥', '♦'][state.trumpSuit] : t('trumpUndecided')}
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
            {/* Permanent, not tutorial-only: the two phases are different
                games, and which one is running decides what a click means. */}
            <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="sg-rule">
              {collecting ? t('phaseCollect') : t('phaseShed')} — {t('ruleLine')}
            </div>

            {/* Opponent hands: backs only. The server withholds the cards. */}
            <div className="flex justify-center gap-4 mb-3" data-tutorial="sg-seats">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('opponentHand', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
                    {' · '}
                    {t('collected', { n: o.collectedCount })}
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

            <div className="text-center mb-4" data-tutorial="sg-table">
              <div className="text-game-text-muted text-xs mb-1">{collecting ? t('duel') : t('pile')}</div>
              <div className="flex gap-1 justify-center flex-wrap">
                {table.length === 0 ? (
                  <span className="text-game-text-muted text-xs">—</span>
                ) : (
                  table.map((card, i) => (
                    <AnimatedCard key={`table-${i.toString()}`} card={card} width={cardWidth} draggable={false} />
                  ))
                )}
              </div>
            </div>

            <div className="text-center" data-tutorial="sg-hand">
              <div className="text-game-text-muted text-xs mb-1">
                {t('yourHand')}
                {' · '}
                {t('collected', { n: human?.collectedCount ?? 0 })}
              </div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => {
                  const canPlay = isHumanTurn && playable.has(i);
                  // **「出せない」だけでは理由が分からない** (#5573)。この
                  // ゲームの beat 規則 (同スートの上位か切札) はサーバが持つので、
                  // 画面は「手番でない」のか「規則に負けている」のかだけを言う。
                  const blockedByBeat = isHumanTurn && !playable.has(i);
                  const reason = blockedByBeat ? t('cannotBeatAria') : canPlay ? '' : t('notYourTurnAria');
                  return (
                    <button
                      key={`hand-${i.toString()}`}
                      type="button"
                      data-hint-action="play"
                      // Kept focusable while it cannot act so the reason is
                      // announced rather than the control leaving the tab order.
                      aria-disabled={!canPlay}
                      title={blockedByBeat ? t('cannotBeatTooltip') : undefined}
                      aria-label={reason ? `${cardAlt(card)} (${reason})` : cardAlt(card)}
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

          <GameFooter className={`${gameTheme.skitgubbe.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {/* Enabled only when the server says nothing beats the pile.
                  Ducking is never lawful, so this is not a free choice. */}
              <button
                type="button"
                data-hint-action="pickup"
                aria-disabled={!state.canPickUp}
                onClick={() => state.canPickUp && game.handlePickUp()}
                className={[
                  'px-4 py-2 rounded font-bold min-h-11',
                  state.canPickUp ? 'bg-ds-accent text-ds-on-accent' : 'bg-ds-surface-2 text-ds-text-muted',
                ].join(' ')}
              >
                {t('pickUp')}
              </button>
              {state.canPickUp && <span className="text-xs text-ds-text-muted">{t('pickUpHint')}</span>}
              {/* Every hand card turns aria-disabled the moment this flips, and
                  until now that change was signalled only by colour (#4912). */}
              {state.canPickUp && (
                <span className="sr-only" role="status" aria-live="polite" data-testid="sk-forced-pickup-notice">
                  {t('forcedPickUpNotice')}
                </span>
              )}
              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sg-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
