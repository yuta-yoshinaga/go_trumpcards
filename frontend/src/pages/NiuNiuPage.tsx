import { useMemo } from 'react';
import type { niuniuApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardBack } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
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
import { useNiuNiuGame } from '../hooks/useNiuNiuGame';
import { btnPrimary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { NiuNiuHand, NiuNiuResponse } from '../types/card';
import { NiuNiuPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { NIUNIU_HELP, parseNiuNiuCommand } from '../utils/cli/commands/niuniuCommands';
import { formatNiuNiuState } from '../utils/cli/formatters/niuniuFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { niuniuRankText } from '../utils/niuniuRankText';

const BET_OPTIONS = [10, 50, 100, 500];

const NN_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="nn-seats"]', messageKey: 'tutorial.combo', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="nn-seats"]', messageKey: 'tutorial.rank', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="nn-banker"]', messageKey: 'tutorial.multiplier', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="nn-controls"]', messageKey: 'tutorial.noChoice', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="nn-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Renders the Niu Niu page: five cards a seat, three of them forming the bull. */
export const NiuNiuPage = withTutorial(NiuNiuPageContent, 'niuniu', NN_TUTORIAL_STEPS);

function NiuNiuPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('niuniu');
  const game = useNiuNiuGame();
  const { state, loading, error, retry } = game;

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('niuniu');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('niuniu', state);
  const cliConfig: CliGameConfig<NiuNiuResponse, Parameters<typeof niuniuApi.exec>> = useMemo(
    () => ({
      gameName: 'niuniu',
      parseCommand: parseNiuNiuCommand,
      formatResponse: formatNiuNiuState,
      helpText: NIUNIU_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();

  if (!state) {
    return <GameSkeleton gameKey="niuniu" layout={{ kind: 'tableau', topRow: 5, tableau: 3 }} />;
  }

  const ended = state.phase === NiuNiuPhase.END;
  const isBetting = state.phase === NiuNiuPhase.BET;

  /**
   * Render one hand. Visibility comes from the server's `hidden` flag; the page
   * never re-derives it.
   *
   * The three cards that made the bull are ringed. Five cards and a rank name
   * with nothing connecting them does not explain itself, and the server sends
   * the positions precisely so the client need not search for them.
   */
  const renderHand = (hand: NiuNiuHand, label: string, keyPrefix: string) => {
    const inCombo = new Set(hand.comboIdx);
    // 牛を作る3枚はリングでしか示されておらず、コードのコメント自身が
    // 「5枚と役名だけでは何も説明にならない」と言っているのに、その情報が
    // 読み上げには一切乗っていなかった (#6363)。
    const comboCards = (hand.comboIdx ?? [])
      .map((i) => hand.cards[i])
      .filter((c) => !!c)
      .map((c) => cardAlt(c))
      .join(' ');
    return (
      <div className="text-center">
        <div
          className="flex gap-1 justify-center"
          role="img"
          aria-label={
            hand.hidden
              ? label
              : `${t('seatAriaLabel', { name: label, rank: niuniuRankText(hand.rankKey) })}${
                  comboCards ? ` ${t('comboAriaLabel', { cards: comboCards })}` : ''
                }`
          }
        >
          {hand.cards.map((card, i) => (
            <div
              key={`${keyPrefix}-c${i.toString()}`}
              className={inCombo.has(i) ? 'rounded ring-2 ring-ds-warning' : ''}
            >
              {hand.hidden || !card ? (
                <CardBack width={cardWidth} />
              ) : (
                <AnimatedCard card={card} width={cardWidth} draggable={false} />
              )}
            </div>
          ))}
        </div>
        {!hand.hidden && (
          <div className="text-game-text-muted text-xs mt-1">
            <span className="text-ds-warning font-bold">{niuniuRankText(hand.rankKey)}</span>
            {hand.multiplier > 1 && <span className="ml-1">{t('multiplier', { mult: hand.multiplier })}</span>}
          </div>
        )}
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.niuniu')}
      gameThemeBg={gameTheme.niuniu.bg}
      phaseName={ended ? t('phase.end') : t('phase.bet')}
      gamePath="/niuniu"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('chips')}: {state.chips}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            <div className="text-center mb-4" data-tutorial="nn-banker">
              <div className="text-game-text-muted text-xs mb-1">{t('bankerHand')}</div>
              {state.bankerHand ? (
                renderHand(
                  state.bankerHand,
                  state.bankerHand.hidden
                    ? t('hiddenBankerHandAriaLabel')
                    : t('bankerHandAriaLabel', { rank: niuniuRankText(state.bankerHand.rankKey) }),
                  'banker',
                )
              ) : (
                <div className="text-game-text-muted text-sm">—</div>
              )}
            </div>

            <div className="flex flex-wrap justify-center gap-4 sm:gap-8" data-tutorial="nn-seats">
              {state.seats.map((seat, seatIdx) => {
                if (seatIdx === state.bankerIdx || !seat.hand) return null;
                return (
                  <div key={`seat-${seatIdx.toString()}`} className="text-center">
                    <div className="text-game-text-muted text-xs mb-1">{seat.name}</div>
                    {renderHand(
                      seat.hand,
                      seat.hand.hidden ? t('hiddenHandAriaLabel', { name: seat.name }) : seat.name,
                      `s${seatIdx.toString()}`,
                    )}
                    <div className="text-game-text-muted text-xs mt-1">
                      {t('bet')}: {seat.hand.bet}
                      {ended && seat.hand.payout !== 0 && (
                        <span className={seat.hand.payout > 0 ? ' text-ds-success' : ' text-ds-danger'}>
                          {' '}
                          {seat.hand.payout > 0 ? `+${seat.hand.payout}` : seat.hand.payout}
                        </span>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>

            {ended && <p className="text-ds-text-muted text-xs text-center mt-3">{t('comboHint')}</p>}

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

          <GameFooter className={`${gameTheme.niuniu.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <label className="flex items-center gap-1 text-ds-text-primary text-xs w-full justify-center cursor-pointer min-h-[44px]">
              <input
                type="checkbox"
                checked={frontendHintEnabled}
                onChange={(e) => setFrontendHintEnabled(e.target.checked)}
              />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center flex-wrap" data-tutorial="nn-controls">
              {isBetting && (
                <>
                  <span className="text-sm text-ds-text-muted">{t('betLabel')}</span>
                  {BET_OPTIONS.map((amount) => (
                    <button
                      key={`bet-${amount.toString()}`}
                      type="button"
                      className={btnPrimary}
                      onClick={() => game.handleBet(amount)}
                      // The loss can be `maxMultiplier` times the stake, so the
                      // stack has to cover that, not just the stake itself — and
                      // greying the button out never said so (#4908).
                      disabled={loading || amount * state.maxMultiplier > state.chips}
                      title={
                        amount * state.maxMultiplier > state.chips
                          ? t('betTooHigh', {
                              multiplier: state.maxMultiplier,
                              needed: amount * state.maxMultiplier,
                            })
                          : undefined
                      }
                    >
                      {t('betAmount', { amount })}
                    </button>
                  ))}
                </>
              )}

              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="nn-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
