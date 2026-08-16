import { useMemo } from 'react';
import type { toepenApi } from '../api/gameApi';
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
import { useToepenGame } from '../hooks/useToepenGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ToepenResponse } from '../types/card';
import { ToepenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseToepenCommand, TOEPEN_HELP } from '../utils/cli/commands/toepenCommands';
import { formatToepenState } from '../utils/cli/formatters/toepenFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const TOEPEN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="toepen-ranking"]',
    messageKey: 'tutorial.ranking',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="toepen-hand"]', messageKey: 'tutorial.follow', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="toepen-lives"]',
    messageKey: 'tutorial.lastTrick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="toepen-controls"]', messageKey: 'tutorial.toep', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="toepen-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Toepen page: 32 cards, inverted ranking, only the last trick pays. */
export const ToepenPage = withTutorial(ToepenPageContent, 'toepen', TOEPEN_TUTORIAL_STEPS);

function ToepenPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('toepen');
  const game = useToepenGame();
  const { state, loading, error, retry } = game;

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('toepen');
  const cliConfig: CliGameConfig<ToepenResponse, Parameters<typeof toepenApi.exec>> = useMemo(
    () => ({
      gameName: 'toepen',
      parseCommand: parseToepenCommand,
      formatResponse: formatToepenState,
      helpText: TOEPEN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('toepen', state);

  if (!state) {
    return <GameSkeleton gameKey="toepen" layout={{ kind: 'tableau', topRow: 4, tableau: 4 }} />;
  }

  const ended = state.phase === ToepenPhase.GAME_END;
  const handOver = state.phase === ToepenPhase.HAND_END;
  const answering = state.phase === ToepenPhase.RESPOND && state.pendingRespondent === 0;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && !handOver && state.phase === ToepenPhase.PLAY && state.currentPlayerIdx === 0;

  // Legality comes from the server, which already applies the follow-suit
  // obligation. Deriving it here would be a second implementation of one rule.
  const playable = new Set(state.validPlayIndices);
  const canToep = !ended && !handOver && state.phase === ToepenPhase.PLAY && !human?.folded;

  return (
    <GamePageShell
      title={tc('nav.toepen')}
      gameThemeBg={gameTheme.toepen.bg}
      phaseName={ended ? t('phase.end') : handOver ? t('phase.handEnd') : t('phase.play')}
      gamePath="/toepen"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('hand', { n: state.handNumber })} / {t('stake')}: {state.stake}
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
            {/* Shown permanently: this ranking is inverted from every other
                game here and is the single easiest thing to misplay. */}
            <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="toepen-ranking">
              {t('ranking')}
            </div>

            <div className="flex justify-center gap-4 mb-3 text-sm flex-wrap" data-tutorial="toepen-lives">
              {state.players.map((p) => (
                <span key={`lv-${p.id.toString()}`} className={p.eliminated ? 'opacity-50 line-through' : ''}>
                  {p.isHuman ? t('you') : t('cpu', { n: p.id })}:{' '}
                  <span className="font-bold">
                    {p.lives}/{state.maxLives}
                  </span>
                  {p.folded && !p.eliminated && <span className="text-ds-text-muted ml-1">{t('folded')}</span>}
                </span>
              ))}
            </div>

            <div className="flex justify-center gap-4 mb-3">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('cpuHand', { n: o.id, count: o.cardCount })}
                  </div>
                  <div
                    className="flex gap-1 justify-center"
                    role="img"
                    aria-label={t('cpuHandAriaLabel', { n: o.id, count: o.cardCount })}
                  >
                    {Array.from({ length: o.cardCount }, (_, i) => (
                      <CardBack key={`opp-${o.id.toString()}-c${i.toString()}`} width={cardWidth} />
                    ))}
                  </div>
                </div>
              ))}
            </div>

            <div className="text-center mb-4 min-h-24">
              <div className="text-game-text-muted text-xs mb-1">
                {state.currentTrick.length > 0 ? t('trick', { n: state.trickNumber + 1 }) : t('noTrick')}
              </div>
              <div className="flex gap-1 justify-center">
                {state.currentTrick.map((tc2, i) => (
                  <AnimatedCard key={`trick-${i.toString()}`} card={tc2.card} width={cardWidth} draggable={false} />
                ))}
              </div>
            </div>

            <div className="text-center" data-tutorial="toepen-hand">
              <div className="text-game-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => {
                  const canPlay = isHumanTurn && playable.has(i);
                  // Blocked only by the follow-suit duty; off-turn cards are not
                  // "wrong", they are just not yours to play yet.
                  const blockedBySuit = isHumanTurn && !playable.has(i);
                  return (
                    <button
                      key={`hand-${i.toString()}`}
                      type="button"
                      data-hint-action="play"
                      data-testid={`toepen-hand-${i.toString()}`}
                      // Kept focusable while it cannot act so the reason is
                      // announced rather than the control leaving the tab order.
                      aria-disabled={!canPlay}
                      title={blockedBySuit ? t('followSuitTooltip') : undefined}
                      aria-label={blockedBySuit ? `${cardAlt(card)} (${t('followSuitAria')})` : cardAlt(card)}
                      onClick={() => canPlay && game.handlePlay(i)}
                      className={[
                        'rounded transition-transform',
                        canPlay ? 'hover:-translate-y-2' : 'opacity-50',
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

          <GameFooter className={`${gameTheme.toepen.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap" data-tutorial="toepen-controls">
              {answering && (
                <>
                  {/* **誰の判断に応答しているのか。** knockerIdx はサーバから届いて
                      いるのに読んでおらず、賭け点しか出していなかった。相手が複数
                      いると誰に応答するのか分からない (#5570)。CUI の respondLine は
                      最初から名前を出している。

                      応答者が toep 宣言者本人になることはない -- domain の
                      nextRespondent が knockerIdx に当たった時点で -1 を返す。
                      つまりこのブロックが出ているとき knockerIdx は必ず CPU 席で、
                      「あなた」の分岐は書いても到達しない。 */}
                  <span className="text-sm text-ds-text-muted" data-testid="toepen-toeped-by">
                    {state.knockerIdx >= 0
                      ? t('toepedBy', { name: t('cpu', { n: state.knockerIdx }), stake: state.stake })
                      : t('toepedAt', { stake: state.stake })}
                  </span>
                  <button
                    type="button"
                    className={btnPrimary}
                    data-hint-action="stay"
                    onClick={() => game.handleRespond(true)}
                    disabled={loading}
                  >
                    {t('stay')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    data-hint-action="fold"
                    onClick={() => game.handleRespond(false)}
                    disabled={loading}
                    // Folding costs the stake BEFORE the raise, so the number
                    // shown here is deliberately one less than `stake`.
                    title={t('foldCost', { cost: state.stake - 1 })}
                  >
                    {t('fold', { cost: state.stake - 1 })}
                  </button>
                </>
              )}

              {state.canRedeal && (
                <button type="button" className={btnSecondary} onClick={game.handleRedeal} disabled={loading}>
                  {t('redeal')}
                </button>
              )}

              {canToep && (
                <button type="button" className={btnSecondary} onClick={game.handleToep} disabled={loading}>
                  {t('toep')}
                </button>
              )}

              {handOver && !ended && (
                <button type="button" className={btnPrimary} onClick={game.handleNextHand} disabled={loading}>
                  {t('nextHand')}
                </button>
              )}

              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="toepen-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
