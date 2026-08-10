import { useMemo } from 'react';
import type { mushiApi } from '../api/gameApi';
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
import { useMushiGame } from '../hooks/useMushiGame';
import { btnPrimary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { MushiCard, MushiResponse } from '../types/card';
import { MushiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { MUSHI_HELP, parseMushiCommand } from '../utils/cli/commands/mushiCommands';
import { formatMushiState } from '../utils/cli/formatters/mushiFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const MUSHI_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="mushi-field"]', messageKey: 'tutorial.field', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="mushi-hand"]', messageKey: 'tutorial.match', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="mushi-hand"]', messageKey: 'tutorial.wild', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="mushi-captured"]', messageKey: 'tutorial.captured', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="mushi-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Mushi page: 40-card hanafuda with no koi-koi stop. */
export const MushiPage = withTutorial(MushiPageContent, 'mushi', MUSHI_TUTORIAL_STEPS);

function MushiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('mushi');
  const game = useMushiGame();
  const { state, loading, error, retry } = game;

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('mushi');
  const cliConfig: CliGameConfig<MushiResponse, Parameters<typeof mushiApi.exec>> = useMemo(
    () => ({
      gameName: 'mushi',
      parseCommand: parseMushiCommand,
      formatResponse: formatMushiState,
      helpText: MUSHI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('mushi', state);

  if (!state) {
    return <GameSkeleton gameKey="mushi" layout={{ kind: 'tableau', topRow: 4, tableau: 4 }} />;
  }

  const ended = state.phase === MushiPhase.GAME_END;
  const roundOver = state.phase === MushiPhase.ROUND_END;
  const choosing = state.phase === MushiPhase.SELECT || state.phase === MushiPhase.WILD_SELECT;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && !roundOver && state.currentPlayerIdx === 0;

  // Selectability comes from the server, which already applies the wild's
  // "not another willow" rule. Re-deriving it here would be a second copy of
  // a rule that has one owner.
  const selectable = new Set(state.selectableIndices);

  /** One card. Hanafuda has no PNG art, so the category label carries the identity. */
  const renderCard = (card: MushiCard, key: string, extra?: string) => (
    <div key={key} className={`text-center ${extra ?? ''}`}>
      <AnimatedCard card={card} width={cardWidth} draggable={false} />
      <div className="text-game-text-muted text-[10px] mt-0.5">
        {t(`category.${card.category}`)}
        {card.isWild && <span className="text-ds-warning ml-0.5">★</span>}
      </div>
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.mushi')}
      gameThemeBg={gameTheme.mushi.bg}
      phaseName={ended ? t('phase.end') : roundOver ? t('phase.roundEnd') : t('phase.play')}
      gamePath="/mushi"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('round', { n: state.roundNumber, total: state.targetRounds })} / {t('stock')}: {state.stockCount}
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
            <div className="flex justify-center gap-6 mb-3 text-sm">
              <span>
                {t('you')}: <span className="font-bold text-ds-warning">{human?.score ?? 0}</span>
                <span className="text-game-text-muted ml-1">({human?.capturedPoints ?? 0} pt)</span>
              </span>
              {opponents.map((o) => (
                <span key={`sc-${o.id.toString()}`}>
                  {t('opponent')}: <span className="font-bold">{o.score}</span>
                  <span className="text-game-text-muted ml-1">({o.capturedPoints} pt)</span>
                </span>
              ))}
            </div>

            {/* Opponent hand: backs only. The server withholds the cards. */}
            <div className="flex justify-center gap-4 mb-3">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('opponentHand', { n: o.cardCount })}</div>
                  <div
                    className="flex gap-1 justify-center"
                    role="img"
                    aria-label={t('opponentHandAriaLabel', { n: o.cardCount })}
                  >
                    {Array.from({ length: o.cardCount }, (_, i) => (
                      <CardBack key={`opp-${o.id.toString()}-c${i.toString()}`} width={cardWidth} />
                    ))}
                  </div>
                </div>
              ))}
            </div>

            <div className="text-center mb-4" data-tutorial="mushi-field">
              <div className="text-game-text-muted text-xs mb-1">
                {choosing && state.pendingCard
                  ? t('choosePrompt', { card: t(`category.${state.pendingCard.category}`) })
                  : t('field')}
              </div>
              {/* The wild's one exception decides most mistakes, and lived only in
                  the first-visit tutorial. Sjavs/Loba/ChineseTen all keep their
                  trickiest rule permanently on screen. */}
              <div className="text-ds-text-muted text-[11px] mb-1 text-center" data-testid="mushi-wild-rule">
                {t('wildRule')}
              </div>
              <div className="flex gap-1 justify-center flex-wrap">
                {state.field.map((card, i) => {
                  const canTake = choosing && isHumanTurn && selectable.has(i);
                  return (
                    <button
                      key={`field-${i.toString()}`}
                      type="button"
                      data-hint-action="select"
                      // Kept focusable while it cannot act so the reason is
                      // announced rather than the control leaving the tab order.
                      aria-disabled={!canTake}
                      onClick={() => canTake && game.handleSelect(i)}
                      className={[
                        'rounded transition-transform',
                        canTake ? 'ring-2 ring-ds-accent hover:-translate-y-1' : '',
                        choosing && !canTake ? 'opacity-50' : '',
                      ].join(' ')}
                    >
                      {renderCard(card, `field-c${i.toString()}`)}
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Captured cards are PUBLIC for both seats -- reading what the
                opponent is collecting is the game. */}
            <div className="mb-4" data-tutorial="mushi-captured">
              {state.players.map((p) => (
                <div key={`cap-${p.id.toString()}`} className="mb-2">
                  <div className="text-game-text-muted text-xs mb-0.5">
                    {p.isHuman ? t('yourCaptured') : t('opponentCaptured')} ({p.capturedPoints} pt)
                  </div>
                  <div className="flex gap-0.5 justify-center flex-wrap">
                    {p.captured.length === 0 ? (
                      <span className="text-game-text-muted text-xs">—</span>
                    ) : (
                      p.captured.map((card, i) => renderCard(card, `cap-${p.id.toString()}-${i.toString()}`))
                    )}
                  </div>
                </div>
              ))}
            </div>

            <div className="text-center" data-tutorial="mushi-hand">
              <div className="text-game-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => (
                  <button
                    key={`hand-${i.toString()}`}
                    type="button"
                    data-hint-action="play"
                    aria-disabled={!isHumanTurn || choosing}
                    onClick={() => isHumanTurn && !choosing && game.handlePlay(i)}
                    className={[
                      'rounded transition-transform',
                      isHumanTurn && !choosing ? 'hover:-translate-y-2' : 'opacity-60',
                      frontendHintEnabled && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
                    ].join(' ')}
                  >
                    {renderCard(card, `hand-c${i.toString()}`)}
                  </button>
                ))}
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

          <GameFooter className={`${gameTheme.mushi.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap" data-tutorial="mushi-controls">
              {roundOver && !ended && (
                <>
                  <span className="text-sm text-ds-text-muted">
                    {t('roundResult', { delta: human?.roundResult ?? 0 })}
                  </span>
                  <button type="button" className={btnPrimary} onClick={game.handleNextRound} disabled={loading}>
                    {t('nextRound')}
                  </button>
                </>
              )}

              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="mushi-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
