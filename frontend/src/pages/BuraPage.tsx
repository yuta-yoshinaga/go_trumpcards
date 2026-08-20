import { useMemo } from 'react';
import type { buraApi } from '../api/gameApi';
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
import { useBuraGame } from '../hooks/useBuraGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BuraResponse } from '../types/card';
import { BuraPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { BURA_HELP, parseBuraCommand } from '../utils/cli/commands/buraCommands';
import { formatBuraState } from '../utils/cli/formatters/buraFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const BURA_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bura-trump"]', messageKey: 'tutorial.trump', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bura-hand"]', messageKey: 'tutorial.multiCard', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="bura-table"]', messageKey: 'tutorial.beating', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bura-points"]', messageKey: 'tutorial.claim', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bura-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Bura page: a 36-card trick game won by claiming 31 points. */
export const BuraPage = withTutorial(BuraPageContent, 'bura', BURA_TUTORIAL_STEPS);

function BuraPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('bura');
  const game = useBuraGame();
  const { state, loading, error, retry, selected, toggleCard } = game;

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bura');
  const cliConfig: CliGameConfig<BuraResponse, Parameters<typeof buraApi.exec>> = useMemo(
    () => ({
      gameName: 'bura',
      parseCommand: parseBuraCommand,
      formatResponse: formatBuraState,
      helpText: BURA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('bura', state);

  if (!state) {
    return <GameSkeleton gameKey="bura" layout={{ kind: 'tableau', topRow: 3, tableau: 3 }} />;
  }

  const ended = state.phase === BuraPhase.GAME_END;
  // 役の一覧はサーバが判定順に送ってくる。訳の無いキーが来ても、キー名を
  // そのまま出すより黙って落とす方がまし ── 役の説明として読めないので。
  const declareTitle = [
    t('declareHint'),
    ...state.winningCombinations.map((key) => t(`combo.${key}`, { defaultValue: '' })).filter(Boolean),
  ].join('\n');
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;
  const leadCount = state.currentLead.length;
  // The server also ships the exact indices it recommends; the tooltip only
  // carries the reason, so read them off the state response directly.
  const hintedIndices = new Set(state.hint?.cardIndices ?? []);

  // A lead may be 1-3 cards of ONE suit; a response has to match the lead's
  // count exactly. Checking here keeps the player from posting a move the
  // server will only reject.
  const selectedCards = selected.map((i) => human?.cards[i]).filter((c) => c != null);
  const sameSuit = selectedCards.every((c) => c.design === selectedCards[0]?.design);
  const canPlay =
    isHumanTurn &&
    selected.length > 0 &&
    (leadCount > 0 ? selected.length === leadCount : sameSuit && selected.length <= 3);

  return (
    <GamePageShell
      title={tc('nav.bura')}
      gameThemeBg={gameTheme.bura.bg}
      phaseName={ended ? t('phase.end') : t('phase.play')}
      gamePath="/bura"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('trick')}: {state.trickNumber} / {t('stock')}: {state.stockRemaining}
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
            <div className="text-center mb-3" data-tutorial="bura-trump">
              <div className="text-game-text-muted text-xs mb-1">{t('trump')}</div>
              {state.trumpCard ? (
                <div className="flex justify-center">
                  <AnimatedCard card={state.trumpCard} width={cardWidth} draggable={false} />
                </div>
              ) : (
                <div className="text-game-text-muted text-sm">
                  {t('trumpDrawn', { suit: t(`suit.${state.trumpSuit}`) })}
                </div>
              )}
            </div>

            <div className="flex justify-center gap-6 mb-3 text-sm" data-tutorial="bura-points">
              <span>
                {t('you')}: <span className="font-bold text-ds-warning">{human?.points ?? 0}</span>
              </span>
              <span className="text-game-text-muted">
                {t('target')}: {state.winThreshold}
              </span>
              {opponents.map((o) => (
                <span key={`pts-${o.id.toString()}`}>
                  {t('opponent')}: <span className="font-bold">{o.points}</span>
                </span>
              ))}
            </div>

            {/* Opponent hands: backs only. The server withholds the cards, so
                there is nothing here to reveal even by mistake. */}
            <div className="flex justify-center gap-4 mb-4">
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

            <div className="text-center mb-4 min-h-24" data-tutorial="bura-table">
              <div className="text-game-text-muted text-xs mb-1">
                {leadCount > 0 ? t('ledCards', { n: leadCount }) : t('noLead')}
              </div>
              <div className="flex gap-1 justify-center">
                {state.currentLead.map((card, i) => (
                  <AnimatedCard key={`lead-${i.toString()}`} card={card} width={cardWidth} draggable={false} />
                ))}
              </div>
            </div>

            <div className="text-center" data-tutorial="bura-hand">
              <div className="text-game-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => {
                  const isSelected = selected.includes(i);
                  return (
                    <button
                      key={`hand-${i.toString()}`}
                      type="button"
                      // Kept focusable while it cannot act, so the reason is
                      // announced instead of the control vanishing from the
                      // tab order.
                      aria-disabled={!isHumanTurn}
                      aria-pressed={isSelected}
                      onClick={() => isHumanTurn && toggleCard(i)}
                      className={[
                        'rounded transition-transform',
                        isSelected ? '-translate-y-3 ring-2 ring-ds-accent' : '',
                        frontendHintEnabled && hintedIndices.has(i) ? 'ring-2 ring-ds-warning' : '',
                        isHumanTurn ? '' : 'opacity-60',
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

          <GameFooter className={`${gameTheme.bura.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap" data-tutorial="bura-controls">
              <button
                type="button"
                className={btnPrimary}
                data-hint-action="play"
                onClick={game.handlePlay}
                disabled={loading || !canPlay}
              >
                {leadCount > 0 ? t('respond', { n: leadCount }) : t('play')}
              </button>
              <button
                type="button"
                className={btnSecondary}
                data-hint-action="claim"
                onClick={game.handleClaim}
                disabled={loading || ended}
                title={t('claimWarning')}
              >
                {t('claim', { target: state.winThreshold })}
              </button>
              {/* **何が「役」かはどこにも書かれていなかった。**claim には
                  失敗時のリスクが title で出ているのに、declare は押しても
                  何も起きない理由が分からないままだった (#5568)。役の一覧は
                  サーバから来る ── 画面側で数え直すと、役を足したとき案内だけが
                  古くなる。 */}
              <button
                type="button"
                className={btnSecondary}
                data-hint-action="declare"
                onClick={game.handleDeclare}
                disabled={loading || ended}
                title={declareTitle}
              >
                {t('declare')}
              </button>

              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="bura-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
