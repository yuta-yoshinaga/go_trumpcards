import { useMemo, useState } from 'react';
import type { laughandliedownApi } from '../api/gameApi';
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
import { useLaughAndLieDownGame } from '../hooks/useLaughAndLieDownGame';
import { gameTheme } from '../styles/gameTheme';
import type { LaughAndLieDownResponse } from '../types/card';
import { LaughAndLieDownPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { LAUGHANDLIEDOWN_HELP, parseLaughAndLieDownCommand } from '../utils/cli/commands/laughandliedownCommands';
import { formatLaughAndLieDownState } from '../utils/cli/formatters/laughandliedownFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

const LLD_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="lld-table"]', messageKey: 'tutorial.table', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="lld-hand"]', messageKey: 'tutorial.capture', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="lld-seats"]', messageKey: 'tutorial.liedown', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="lld-rule"]', messageKey: 'tutorial.settle', placement: 'bottom', advanceOn: 'next' },
];

/** Renders the Laugh and Lie Down page: capture one or three of a rank, or lay your hand down. */
export const LaughAndLieDownPage = withTutorial(LaughAndLieDownPageContent, 'laughandliedown', LLD_TUTORIAL_STEPS);

function LaughAndLieDownPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('laughandliedown');
  const game = useLaughAndLieDownGame();
  const { state, loading, error, retry } = game;

  // Which hand card is armed for a three-card take. The take size is a real
  // choice only when the server offered it, so it lives next to the card
  // rather than as a mode toggle.
  const [threeArmed, setThreeArmed] = useState<number | null>(null);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('laughandliedown');
  const cliConfig: CliGameConfig<LaughAndLieDownResponse, Parameters<typeof laughandliedownApi.exec>> = useMemo(
    () => ({
      gameName: 'laughandliedown',
      parseCommand: parseLaughAndLieDownCommand,
      formatResponse: formatLaughAndLieDownState,
      helpText: LAUGHANDLIEDOWN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('laughandliedown', state);

  if (!state) {
    return <GameSkeleton gameKey="laughandliedown" layout={{ kind: 'tableau', topRow: 4, tableau: 4 }} />;
  }

  const ended = state.phase === LaughAndLieDownPhase.GAME_END;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;

  // Both the match rule and the three-card option come from the server, which
  // owns them. Recounting the table here would put the rule in two places.
  const playable = new Set(state.validIndices);
  const threeTakes = new Set(state.threeTakeIndices);
  // **押したときだけ出す。**`frontendHintEnabled` は設定トグルであって
  // 「ヒントを押したか」ではない (#4605)。
  const showServerHint = frontendHintEnabled && isRequestedHint(state);
  // CUI は `hintPlay` に「1枚 or 3枚」を書いているのに、Web はカードを光らせる
  // だけで takeCount を捨てていた (#4884)。
  const hintWantsThree = showServerHint && state.hint?.takeCount === 3;

  const play = (i: number) => {
    if (!isHumanTurn || !playable.has(i)) return;
    const take = threeArmed === i && threeTakes.has(i) ? 3 : 1;
    setThreeArmed(null);
    game.handlePlay(i, take);
  };

  return (
    <GamePageShell
      title={tc('nav.laughandliedown')}
      gameThemeBg={gameTheme.laughandliedown.bg}
      phaseName={ended ? t('phase.end') : t('phase.play')}
      gamePath="/laughandliedown"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('pot')}: {state.pot} / {t('dealer')}: {t('seat', { n: state.dealerIdx })}
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
            {/* Permanent, not tutorial-only: "one or three" and "cannot
                capture means your whole hand goes to the table" are the two
                things a player gets wrong. */}
            <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="lld-rule">
              {t('ruleLine')}
            </div>

            {/* Opponent hands: backs only. The server withholds the cards. */}
            <div className="flex justify-center gap-4 mb-3 flex-wrap" data-tutorial="lld-seats">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('opponentHand', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
                    {' · '}
                    {t('won', { n: o.wonCount })}
                    {o.laidDown && ` · ${t('laidDown')}`}
                    {ended && ` · ${t('score', { n: o.score })}`}
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

            {/* The table is face up in full: seeing how many of each rank are
                left is what makes the three-card take a decision. */}
            <div className="text-center mb-4" data-tutorial="lld-table">
              <div className="text-game-text-muted text-xs mb-1">{t('table')}</div>
              <div className="flex gap-1 justify-center flex-wrap">
                {state.layout.length === 0 ? (
                  <span className="text-game-text-muted text-xs">—</span>
                ) : (
                  state.layout.map((card, i) => (
                    <AnimatedCard key={`table-${i.toString()}`} card={card} width={cardWidth} draggable={false} />
                  ))
                )}
              </div>
            </div>

            <div className="text-center" data-tutorial="lld-hand">
              <div className="text-game-text-muted text-xs mb-1">
                {t('yourHand')}
                {' · '}
                {t('won', { n: human?.wonCount ?? 0 })}
                {ended && ` · ${t('score', { n: human?.score ?? 0 })}`}
              </div>
              <div className="flex gap-2 justify-center flex-wrap items-start">
                {(human?.cards ?? []).map((card, i) => {
                  const canPlay = isHumanTurn && playable.has(i);
                  return (
                    <div key={`hand-${i.toString()}`} className="text-center">
                      <button
                        type="button"
                        data-hint-action="play"
                        // Kept focusable while it cannot act so the reason is
                        // announced rather than the control leaving the tab order.
                        aria-disabled={!canPlay}
                        onClick={() => play(i)}
                        className={[
                          'rounded transition-transform',
                          canPlay ? 'hover:-translate-y-2' : 'opacity-60',
                          showServerHint && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
                        ].join(' ')}
                      >
                        <AnimatedCard card={card} width={cardWidth} draggable={false} />
                      </button>
                      {/* Only offered where the server said three are on the
                          table -- the page never recounts. */}
                      {canPlay && threeTakes.has(i) && (
                        <button
                          type="button"
                          aria-pressed={threeArmed === i}
                          onClick={() => setThreeArmed(threeArmed === i ? null : i)}
                          className={[
                            'mt-1 px-2 py-1 rounded text-[10px] min-h-11',
                            threeArmed === i ? 'bg-ds-accent text-ds-on-accent' : 'bg-ds-surface-2 text-ds-text',
                            // **ヒントが 3 枚取りを勧めているなら、そのボタンも
                            // 示す。**カードを光らせるだけでは、さらにこれを
                            // 押す必要があることが伝わらない (#4884)。
                            hintWantsThree && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
                          ].join(' ')}
                          data-hint-take-three={hintWantsThree && state.hint?.cardIndex === i ? 'true' : undefined}
                        >
                          {t('takeThree')}
                        </button>
                      )}
                    </div>
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

          <GameFooter className={`${gameTheme.laughandliedown.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="lld-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
