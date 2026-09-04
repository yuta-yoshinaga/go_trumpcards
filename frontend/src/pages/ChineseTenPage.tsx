import { useMemo } from 'react';
import type { chinesetenApi } from '../api/gameApi';
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
import { useChineseTenGame } from '../hooks/useChineseTenGame';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { gameTheme } from '../styles/gameTheme';
import type { ChineseTenCard, ChineseTenResponse } from '../types/card';
import { ChineseTenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import type { ChineseTenBlockedReason } from '../utils/chineseTenDisabledReason';
import { chineseTenHandBlockedReason, chineseTenLayoutBlockedReason } from '../utils/chineseTenDisabledReason';
import { CHINESETEN_HELP, parseChineseTenCommand } from '../utils/cli/commands/chinesetenCommands';
import { formatChineseTenState } from '../utils/cli/formatters/chinesetenFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

const CT_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="ct-rule"]', messageKey: 'tutorial.capture', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ct-layout"]', messageKey: 'tutorial.layout', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ct-hand"]', messageKey: 'tutorial.flip', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="ct-captured"]', messageKey: 'tutorial.red', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="ct-scores"]', messageKey: 'tutorial.tie', placement: 'bottom', advanceOn: 'next' },
];

/** Renders the Chinese Ten page: capture to ten or by rank, only red cards score. */
export const ChineseTenPage = withTutorial(ChineseTenPageContent, 'chineseten', CT_TUTORIAL_STEPS);

function ChineseTenPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('chineseten');
  const game = useChineseTenGame();
  const { state, loading, error, retry } = game;

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('chineseten');
  const cliConfig: CliGameConfig<ChineseTenResponse, Parameters<typeof chinesetenApi.exec>> = useMemo(
    () => ({
      gameName: 'chineseten',
      parseCommand: parseChineseTenCommand,
      formatResponse: formatChineseTenState,
      helpText: CHINESETEN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('chineseten', state);
  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回ヒントを
  // 載せるので、`state.hint` を直接読むと常時ハイライトになる (#4605)。
  const showServerHint = frontendHintEnabled && state !== null && isRequestedHint(state);

  if (!state) {
    return <GameSkeleton gameKey="chineseten" layout={{ kind: 'tableau', topRow: 4, tableau: 4 }} />;
  }

  const ended = state.phase === ChineseTenPhase.GAME_END;
  const choosing = state.phase === ChineseTenPhase.SELECT;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;

  // Selectability comes from the server, which applies BOTH capture rules.
  // Re-deriving them here would put a pair of non-overlapping rules in two
  // places.
  const selectable = new Set(state.selectableIndices);
  // **aria-disabled は「押しても何も起きない」までしか言わない。**待つのか、
  // 別の札を選ぶのか、先に手を出すのかは理由が分からないと決められない (#5571)。
  // 判定は既存の isHumanTurn / choosing / selectable をそのまま読む。
  const turnState = { ended, isHumanTurn, choosing };
  const blockedText = (reason: ChineseTenBlockedReason | null) => (reason ? t(`blocked.${reason}`) : undefined);
  // 手札の理由は札によらないので一度だけ。場札は札ごとに変わるので描画時に引く。
  const handBlocked = blockedText(chineseTenHandBlockedReason(turnState));

  /** One card, annotating the ones that actually score. */
  const renderCard = (card: ChineseTenCard, key: string) => (
    <div key={key} className="text-center">
      <AnimatedCard card={card} width={cardWidth} draggable={false} />
      {card.points > 0 && <div className="text-ds-warning text-[10px] mt-0.5 font-bold">{card.points}</div>}
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.chineseten')}
      gameThemeBg={gameTheme.chineseten.bg}
      phaseName={ended ? t('phase.end') : t('phase.play')}
      gamePath="/chineseten"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('stock')}: {state.stockCount} / {t('tie')}: {state.tieScore}
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
            {/* Permanent, not tutorial-only: A-9 and 10-K capture by different
                rules, and that is what a player gets wrong. */}
            <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="ct-rule">
              {t('captureRule')}
            </div>

            <div className="flex justify-center gap-6 mb-3 text-sm" data-tutorial="ct-scores">
              <span>
                {t('you')}: <span className="font-bold text-ds-warning">{human?.score ?? 0}</span>
              </span>
              {opponents.map((o) => (
                <span key={`sc-${o.id.toString()}`}>
                  {t('opponent')}: <span className="font-bold">{o.score}</span>
                </span>
              ))}
            </div>

            {/* Opponent hand: backs only. The server withholds the cards. */}
            <div className="flex justify-center gap-4 mb-3">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">{t('opponentHand', { n: o.cardCount })}</div>
                  <div
                    className="flex gap-1 justify-center flex-wrap"
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

            {/* 出した札そのものを見せる。場札は複数あるので、これが無いと
                「何に対して」10 や同ランクを選んでいるのか記憶に頼ることになる
                (CUI は chineseten.pendingLine で最初から出していた) (#6367)。
                札が無いときは枠ごと出さない。読み上げは AnimatedCard -> CardImage の
                alt={cardAlt(card)} が担うので、aria-label を重ねない。 */}
            {choosing && state.pendingCard && (
              <div className="text-center mb-4">
                <div className="text-game-text-muted text-xs mb-1">{t('pendingLine')}</div>
                <div className="flex justify-center">{renderCard(state.pendingCard, 'pending-card')}</div>
              </div>
            )}

            <div className="text-center mb-4" data-tutorial="ct-layout">
              <div className="text-game-text-muted text-xs mb-1">
                {choosing && state.pendingCard ? t('choosePrompt') : t('layout')}
              </div>
              <div className="flex gap-1 justify-center flex-wrap">
                {state.layout.map((card, i) => {
                  const canTake = choosing && isHumanTurn && selectable.has(i);
                  const layoutBlocked = blockedText(chineseTenLayoutBlockedReason(turnState, selectable.has(i)));
                  const isHintedLayout = showServerHint && state.hint?.layoutIndex === i;
                  return (
                    <button
                      key={`layout-${i.toString()}`}
                      type="button"
                      data-hint-action="select"
                      // Kept focusable while it cannot act so the reason is
                      // announced rather than the control leaving the tab order.
                      aria-disabled={!canTake}
                      title={layoutBlocked}
                      aria-label={[cardAlt(card), layoutBlocked].filter(Boolean).join(' — ')}
                      data-hinted-layout={isHintedLayout || undefined}
                      onClick={() => canTake && game.handleSelect(i)}
                      className={[
                        'rounded transition-transform',
                        // The hint carries layoutIndex — which table card to take —
                        // and only the hand card was ever ringed (#4881). Two ring-*
                        // utilities on one element would fight, so the hint wins
                        // outright when it names this card.
                        isHintedLayout
                          ? 'ring-2 ring-ds-warning'
                          : canTake
                            ? 'ring-2 ring-ds-accent hover:-translate-y-1'
                            : '',
                        choosing && !canTake ? 'opacity-50' : '',
                      ].join(' ')}
                    >
                      {renderCard(card, `layout-c${i.toString()}`)}
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Captures are PUBLIC for both seats -- reading what has gone is
                how the remaining cards are worked out. */}
            <div className="mb-4" data-tutorial="ct-captured">
              {state.players.map((p) => (
                <div key={`cap-${p.id.toString()}`} className="mb-2">
                  <div className="text-game-text-muted text-xs mb-0.5">
                    {p.isHuman ? t('yourCaptured') : t('opponentCaptured')} ({p.score})
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

            <div className="text-center" data-tutorial="ct-hand">
              <div className="text-game-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => (
                  <button
                    key={`hand-${i.toString()}`}
                    type="button"
                    data-hint-action="play"
                    aria-disabled={!isHumanTurn || choosing}
                    title={handBlocked}
                    aria-label={[cardAlt(card), handBlocked].filter(Boolean).join(' — ')}
                    onClick={() => isHumanTurn && !choosing && game.handlePlay(i)}
                    className={[
                      'rounded transition-transform',
                      isHumanTurn && !choosing ? 'hover:-translate-y-2' : 'opacity-60',
                      showServerHint && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
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

          <GameFooter className={`${gameTheme.chineseten.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="ct-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
