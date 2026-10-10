import { useCallback, useMemo, useState } from 'react';
import { crazyfourpokerApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ChipBetInput } from '../components/common/ChipBetInput';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnPrimary, btnSecondary, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, CrazyFourPokerResponse } from '../types/card';
import { CRAZY_FOUR_POKER_ANTE_UNIT, CRAZY_FOUR_POKER_RESULT } from '../types/games/crazyfourpoker';
import { CrazyFourPokerPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CRAZYFOURPOKER_CLI_HELP, parseCrazyFourPokerCommand } from '../utils/cli/commands/crazyfourpokerCommands';
import { formatCrazyFourPokerState } from '../utils/cli/formatters/crazyfourpokerFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** `FourCardHandPair` (four_card_hand_eval.go:11). */
const FOUR_CARD_HAND_PAIR = 2;
/** `CrazyFourPokerSuperBonusMinPair` (CrazyFourPokerConfig.go:164) — aces only. */
const SUPER_BONUS_MIN_PAIR = 1;
/** `CrazyFourPokerQueensUpMinPair` (CrazyFourPokerConfig.go:144) — queens or better. */
const QUEENS_UP_MIN_PAIR = 12;

/** Match best-hand cards to their positions while preserving duplicate cards. */
function bestCardIndices(hand: readonly Card[], best: readonly Card[] | undefined): Set<number> {
  if (!best) return new Set<number>();
  const remaining = new Map<string, number>();
  for (const card of best) {
    const key = `${card.design}:${card.value}`;
    remaining.set(key, (remaining.get(key) ?? 0) + 1);
  }
  const indices = new Set<number>();
  hand.forEach((card, index) => {
    const key = `${card.design}:${card.value}`;
    const count = remaining.get(key) ?? 0;
    if (count > 0) {
      indices.add(index);
      remaining.set(key, count - 1);
    }
  });
  return indices;
}

/**
 * Whether the player's hand clears the pair minimum a side bet requires.
 *
 * domain_rule_quoted:
 * 	func crazyFourPokerPairAtLeast(best []*Card, minPair int) bool {
 * 		if len(best) == 0 { return false }
 * 		rank := evalFourCardHand(best)
 * 		if rank > FourCardHandPair { return true }
 * 		if rank != FourCardHandPair { return false }
 * 		pv := fourCardPairSortedValues(best)
 * 		...
 * 	}
 *
 * edge_cases:
 * - Anything above a pair clears every minimum outright.
 * - Aces count high, so an ace pair is 14 rather than 1 on both sides of the
 *   comparison — which is why the Super Bonus minimum of 1 means "aces only"
 *   and not "any pair".
 */
function crazyFourPokerPairAtLeast(best: CrazyFourPokerResponse['playerBest'], handRank: number, minPair: number) {
  if (handRank > FOUR_CARD_HAND_PAIR) return true;
  if (handRank !== FOUR_CARD_HAND_PAIR) return false;
  const counts = new Map<number, number>();
  for (const card of best) counts.set(card.value, (counts.get(card.value) ?? 0) + 1);
  const pair = [...counts.entries()].find(([, count]) => count >= 2)?.[0];
  if (pair === undefined) return false;
  return (pair === 1 ? 14 : pair) >= (minPair === 1 ? 14 : minPair);
}

const C4P_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="c4p-bet"]', messageKey: 'tutorial.bet', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="c4p-hand"]', messageKey: 'tutorial.hand', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="c4p-actions"]', messageKey: 'tutorial.multiplier', placement: 'top', advanceOn: 'next' },
];

/** Renders the Crazy 4 Poker game page (#5260). */
export const CrazyFourPokerPage = withTutorial(CrazyFourPokerPageContent, 'crazyfourpoker', C4P_TUTORIAL_STEPS);

function CrazyFourPokerPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('crazyfourpoker');

  const [ante, setAnte] = useState(50);
  const [queensUp, setQueensUp] = useState(0);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(crazyfourpokerApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('crazyfourpoker');
  const cliConfig: CliGameConfig<CrazyFourPokerResponse, Parameters<typeof crazyfourpokerApi.exec>> = useMemo(
    () => ({
      gameName: 'crazyfourpoker',
      parseCommand: parseCrazyFourPokerCommand,
      formatResponse: formatCrazyFourPokerState,
      helpText: CRAZYFOURPOKER_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const maxAnte = state ? Math.floor(state.chips / (2 * CRAZY_FOUR_POKER_ANTE_UNIT)) * CRAZY_FOUR_POKER_ANTE_UNIT : 0;
  const effectiveAnte = Math.min(ante, maxAnte);
  const maxQueensUp = state && maxAnte > 0 ? Math.max(0, state.chips - effectiveAnte * 2) : 0;
  const effectiveQueensUp = Math.min(queensUp, maxQueensUp);
  const isBetPhase = phase === CrazyFourPokerPhase.BET;
  const isDecidePhase = phase === CrazyFourPokerPhase.DECIDE;
  const isResultPhase = phase === CrazyFourPokerPhase.RESULT;
  const gameOver = !!state?.gameEndFlag;

  const handleDeal = useCallback(
    () => execApi('bet', { ante: effectiveAnte, queensUp: effectiveQueensUp }),
    [execApi, effectiveAnte, effectiveQueensUp],
  );

  // **置ける倍率はサーバが決める。** 手役から計算し直すと、このゲームの本体である
  // 「3 倍はエースのペア以上だけ」という規則が 2 か所に増えてずれる。
  const multipliers = useMemo(() => {
    const max = state?.maxMultiplier ?? 1;
    return Array.from({ length: max }, (_, i) => i + 1);
  }, [state?.maxMultiplier]);

  const actionBindings = useMemo(
    () => [
      { key: 'f', action: () => execApi('fold'), enabled: isDecidePhase },
      { key: 'n', action: () => execApi('next'), enabled: isResultPhase && !gameOver },
    ],
    [execApi, isDecidePhase, isResultPhase, gameOver],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('crazyfourpoker', state);

  if (!state) return <GameSkeleton gameKey="crazyfourpoker" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [CrazyFourPokerPhase.BET]: t('phase.bet'),
      [CrazyFourPokerPhase.DECIDE]: t('phase.decide'),
      [CrazyFourPokerPhase.RESULT]: t('phase.result'),
    }[state.phase] ?? '';

  const resultKey = Object.entries(CRAZY_FOUR_POKER_RESULT).find(([, v]) => v === state.result)?.[0] ?? 'none';
  const staked = state.anteBet + state.superBet + state.queensUpBet + state.playBet;
  const net = state.payout - staked;
  const won = state.result === CRAZY_FOUR_POKER_RESULT.win;
  const mainReturn =
    state.result === CRAZY_FOUR_POKER_RESULT.win
      ? (state.anteBet + state.playBet) * 2
      : state.result === CRAZY_FOUR_POKER_RESULT.push
        ? state.anteBet + state.playBet
        : state.result === CRAZY_FOUR_POKER_RESULT.dealerNotQualified
          ? state.anteBet * 2 + state.playBet
          : 0;
  const fourAces = state.playerBest.length === 4 && state.playerBest.every((card) => card.value === 1);
  const superBonus = state.superBonusPayouts?.find(
    (row) =>
      row.hand === state.playerHandRank &&
      (row.hand !== FOUR_CARD_HAND_PAIR ||
        crazyFourPokerPairAtLeast(state.playerBest, state.playerHandRank, SUPER_BONUS_MIN_PAIR)) &&
      (fourAces
        ? row.name.includes('エース') || row.name.includes('Ace')
        : !row.name.includes('エース') && !row.name.includes('Ace')),
  );
  const superReturn =
    state.superBet === 0
      ? 0
      : superBonus
        ? state.superBet + Math.round((state.superBet * Number(superBonus.odds)) / 1)
        : state.result === CRAZY_FOUR_POKER_RESULT.win ||
            state.result === CRAZY_FOUR_POKER_RESULT.push ||
            state.result === CRAZY_FOUR_POKER_RESULT.dealerNotQualified
          ? state.superBet
          : 0;
  const queensUpPayout = state.queensUpPayouts?.find(
    (row) =>
      row.hand === state.playerHandRank &&
      (row.hand !== FOUR_CARD_HAND_PAIR ||
        crazyFourPokerPairAtLeast(state.playerBest, state.playerHandRank, QUEENS_UP_MIN_PAIR)),
  );
  const queensUpReturn = queensUpPayout ? state.queensUpBet * (queensUpPayout.multiplier + 1) : 0;

  const playerBestIdx =
    isDecidePhase || isResultPhase ? bestCardIndices(state.playerHand, state.playerBest) : new Set<number>();
  const dealerBestIdx = isResultPhase ? bestCardIndices(state.dealerHand, state.dealerBest) : new Set<number>();

  const handRow = (
    label: string,
    cards: CrazyFourPokerResponse['playerHand'],
    testId: string,
    bestIndices: Set<number>,
  ) => (
    <div className="mb-2">
      <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{label}</div>
      <div className="flex justify-center gap-1 flex-wrap" data-testid={testId}>
        {cards.map((card, i) => {
          const inBest = bestIndices.has(i);
          const showBest = bestIndices.size > 0;
          return (
            <div
              key={`${testId}-${card.design}-${card.value}-${i}`}
              className={`text-center transition-all ${showBest && inBest ? '-translate-y-1 ring-2 ring-ds-success rounded' : ''} ${showBest && !inBest ? 'opacity-50' : ''}`}
              data-c4p-best={showBest ? (inBest ? 'included' : 'excluded') : undefined}
            >
              <AnimatedCard card={card} width={cardWidth} />
              {showBest && (
                <span className="block text-ds-text-primary text-xs mt-1">
                  {t(inBest ? 'label.bestCardIncluded' : 'label.bestCardExcluded')}
                </span>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.crazyfourpoker')}
      gameThemeBg={gameTheme.crazyfourpoker.bg}
      phaseName={phaseName}
      gamePath="/crazyfourpoker"
      gameEndFlag={gameOver}
      winShow={isResultPhase && won}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="c4p-chips">
            {t('label.chips')}: {state.chips}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <span className="sr-only" role="status" aria-live="polite" aria-atomic="true" data-testid="c4p-rank-announcement">
        {state.playerHand.length > 0 ? `${t('label.yourBest')}: ${t(`rank.${state.playerHandRank}`)}` : ''}
      </span>
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div data-testid="card-area" className={`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="c4p-bet-line">
              {t('label.round')}: {state.roundNumber}
              {state.anteBet > 0 && (
                <>
                  {' · '}
                  {t('label.ante')}: {state.anteBet} · {t('label.superBonus')}: {state.superBet} · {t('label.queensUp')}
                  : {state.queensUpBet}
                </>
              )}
            </div>

            {state.playerHand.length > 0 && (
              <div data-tutorial="c4p-hand">
                {handRow(t('label.yourHand'), state.playerHand, 'c4p-player-hand', playerBestIdx)}
                <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="c4p-player-rank">
                  {t('label.yourBest')}: {t(`rank.${state.playerHandRank}`)}
                </div>
              </div>
            )}

            {/* **決着まではディーラーの手を出さない。** サーバも送っていない。 */}
            {isResultPhase && state.dealerHand.length > 0 ? (
              <>
                {handRow(t('label.dealerHand'), state.dealerHand, 'c4p-dealer-hand', dealerBestIdx)}
                <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="c4p-dealer-rank">
                  {t('label.dealerBest')}: {t(`rank.${state.dealerHandRank}`)}
                  {!state.dealerQualifies && ` · ${t('result.dealerNotQualified')}`}
                </div>
              </>
            ) : (
              state.playerHand.length > 0 && (
                <p className="text-ds-text-muted text-center text-xs mb-2" data-testid="c4p-dealer-hidden">
                  {t('label.hidden')}
                </p>
              )
            )}

            {isResultPhase && (
              <div className="text-center mb-2" data-testid="c4p-result">
                <div className="text-ds-text-primary text-base font-bold">{t(`result.${resultKey}`)}</div>
                <div className="text-sm">
                  {t('result.main', { amount: mainReturn - state.anteBet - state.playBet })}
                </div>
                {state.superBet > 0 && (
                  <div className="text-sm">
                    {t('result.superBonus', {
                      hand: t(`rank.${state.playerHandRank}`),
                      amount: superReturn - state.superBet,
                    })}
                  </div>
                )}
                {state.queensUpBet > 0 && (
                  <div className="text-sm">{t('result.queensUp', { amount: queensUpReturn - state.queensUpBet })}</div>
                )}
                <div className={`text-sm font-medium ${net >= 0 ? 'text-ds-success' : 'text-ds-error-text'}`}>
                  {t('label.net')}: {net}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter
            className={`${gameTheme.crazyfourpoker.footer} px-4 pt-3`}
            actions={
              isBetPhase && !gameOver ? (
                <div className="flex justify-center">
                  <button type="button" className={btnPrimary} onClick={handleDeal} disabled={loading}>
                    {t('button.deal')}
                  </button>
                </div>
              ) : isDecidePhase ? (
                <div className="flex gap-2 flex-wrap justify-center" data-tutorial="c4p-actions">
                  {multipliers.map((m) => (
                    <div key={`mult-${m}`} className="text-center">
                      <button
                        type="button"
                        className={btnPrimary}
                        data-hint-action={m === state.maxMultiplier && state.hasAcesOrBetter ? 'raise' : 'play'}
                        data-testid={`c4p-play-${m}`}
                        onClick={() => execApi('play', { multiplier: m })}
                        disabled={loading}
                      >
                        {t('button.play', { multiplier: m })}
                      </button>
                      <p className="text-ds-text-muted text-xs" data-testid={`c4p-play-wager-${m}`}>
                        {t('wagerSummary.playChoice', {
                          playBet: state.anteBet * m,
                          total: state.anteBet + state.superBet + state.queensUpBet + state.anteBet * m,
                        })}
                      </p>
                    </div>
                  ))}
                  <button type="button" className={btnWarning} onClick={() => execApi('fold')} disabled={loading}>
                    {t('button.fold')}
                  </button>
                </div>
              ) : isResultPhase && !gameOver ? (
                <div className="flex justify-center">
                  <button type="button" className={btnPrimary} onClick={() => execApi('next')} disabled={loading}>
                    {t('button.next')}
                  </button>
                </div>
              ) : null
            }
          >
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2">
              {isBetPhase && !gameOver && (
                <div className="flex flex-col items-center gap-2" data-tutorial="c4p-bet">
                  <p className="text-ds-text-muted text-sm">{t('betGuide')}</p>
                  <ChipBetInput
                    id="crazyfourpoker-ante"
                    label={t('label.ante')}
                    value={effectiveAnte}
                    onChange={setAnte}
                    max={maxAnte}
                    min={0}
                    step={CRAZY_FOUR_POKER_ANTE_UNIT}
                  />
                  {/* **アンティと Super Bonus を引いた残りしか置けない。** 上限を
                      チップ全額にすると、合計が持ち金を超える組み合わせを選べてしまう。 */}
                  <ChipBetInput
                    id="crazyfourpoker-queensup"
                    label={t('label.queensUp')}
                    value={effectiveQueensUp}
                    onChange={setQueensUp}
                    max={maxQueensUp}
                    min={0}
                    step={CRAZY_FOUR_POKER_ANTE_UNIT}
                  />
                  <section className="text-ds-text-primary text-sm text-center" data-testid="c4p-wager-summary">
                    <h2 className="font-bold">{t('wagerSummary.title')}</h2>
                    <p>{t('wagerSummary.ante', { amount: effectiveAnte })}</p>
                    <p>{t('wagerSummary.superBonus', { amount: effectiveAnte })}</p>
                    <p>{t('wagerSummary.required', { amount: effectiveAnte * 2 })}</p>
                    <p>{t('wagerSummary.optional', { amount: effectiveQueensUp })}</p>
                    <p className="font-bold">
                      {t('wagerSummary.total', { amount: effectiveAnte * 2 + effectiveQueensUp })}
                    </p>
                  </section>
                  {/* **賭ける前に見えなければ意味がない** (#5775)。何が当たれば
                      何倍かを知って額を決めるもの。倍率はサーバの配当表そのまま。 */}
                  {/* **Super Bonus はアンティに必ず付く。** 任意の Queens Up は
                      表が見えて、必須の側だけ見えないのは逆。 */}
                  <table className="text-xs text-ds-text-muted mx-auto" data-testid="c4p-superbonus-payouts">
                    <caption className="sr-only">{t('label.superBonusPayouts')}</caption>
                    <tbody>
                      {state.superBonusPayouts?.map((row) => (
                        <tr key={`sb-${row.hand}-${row.name}`}>
                          <td className="pr-3 text-left">{row.name}</td>
                          <td className="text-right">{t('label.payoutOdds', { n: row.odds })}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  <table className="text-xs text-ds-text-muted mx-auto" data-testid="c4p-queensup-payouts">
                    <caption className="sr-only">{t('label.queensUpPayouts')}</caption>
                    <tbody>
                      {state.queensUpPayouts.map((row) => (
                        <tr key={row.hand}>
                          <td className="pr-3 text-left">{row.name}</td>
                          <td className="text-right">{t('label.payoutOdds', { n: row.multiplier })}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {isDecidePhase && (
                <>
                  <p className="text-ds-text-muted text-sm">{t('decideGuide')}</p>
                  <p
                    className={`text-sm font-bold ${state.hasAcesOrBetter ? 'text-ds-success' : 'text-ds-text-muted'}`}
                    data-testid="c4p-multiplier-notice"
                  >
                    {state.hasAcesOrBetter ? t('acesNotice') : t('normalNotice')}
                  </p>
                </>
              )}

              <div className="flex gap-2">
                <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                  {tc('button.actionLog')}
                </button>
                <GameResetButton
                  isGameEnd={gameOver}
                  onReset={() => execApi('reset')}
                  requestConfirm={requestConfirm}
                  loading={loading}
                />
              </div>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
