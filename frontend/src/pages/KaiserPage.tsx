import { useEffect, useMemo, useState } from 'react';
import { kaiserApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { KaiserResponse } from '../types/card';
import { KaiserPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { KAISER_HELP, parseKaiserCommand } from '../utils/cli/commands/kaiserCommands';
import { formatKaiserState } from '../utils/cli/formatters/kaiserFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Suit glyphs by design value (1=Spade … 4=Diamond). */
const SUIT_GLYPHS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Contract options in bidding order: trump < no trump < low no trump. */
const CONTRACTS = [
  { value: 0, labelKey: 'withTrump' },
  { value: 1, labelKey: 'noTrump' },
  { value: 2, labelKey: 'lowNoTrump' },
];

/** How many cards the declarer must discard (sync: `KaiserKittySize`). */
const KAISER_DISCARD_COUNT = 2;

/** Kaiser tutorial step definitions. */
const KAISER_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="kaiser-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kaiser-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kaiser-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kaiser-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kaiser-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KAISER_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KaiserPhase.BID]: 'bid',
  [KaiserPhase.DISCARD]: 'discard',
  [KaiserPhase.PLAY]: 'play',
  [KaiserPhase.HAND_END]: 'handEnd',
  [KaiserPhase.GAME_END]: 'gameEnd',
};

/** Renders the Kaiser game page: the Saskatchewan partnership bidding game. */
export const KaiserPage = withTutorial(KaiserPageContent, 'kaiser', KAISER_TUTORIAL_STEPS);

/** Inner content of the Kaiser page, wrapped by TutorialProvider. */
function KaiserPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('kaiser');
  const { state, loading, error, exec, retry } = useGameApi(kaiserApi.exec);

  const [contract, setContract] = useState(0);
  const [selected, setSelected] = useState<number[]>([]);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('kaiser');
  const cliConfig: CliGameConfig<KaiserResponse, Parameters<typeof kaiserApi.exec>> = useMemo(
    () => ({
      gameName: 'kaiser',
      parseCommand: parseKaiserCommand,
      formatResponse: formatKaiserState,
      helpText: KAISER_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('kaiser', KAISER_PHASE_KEYS);

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('kaiser', state);

  if (!state)
    return <GameSkeleton gameKey="kaiser" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;

  const human = state.players.find((p) => p.isHuman);
  const isBid = state.phase === KaiserPhase.BID;
  const isDiscard = state.phase === KaiserPhase.DISCARD;
  const isPlay = state.phase === KaiserPhase.PLAY;
  const isHandEnd = state.phase === KaiserPhase.HAND_END;
  const isGameEnd = state.phase === KaiserPhase.GAME_END || state.gameEndFlag;
  // 人間は席 0 = チーム 0。勝敗はチームで判定する。
  const humanWon = isGameEnd && state.winnerTeam === 0;
  const isHumanBid = isBid && state.bidPlayerIdx === 0 && !isGameEnd;
  const isHumanDeclarer = isDiscard && state.declarerIdx === 0 && !isGameEnd;
  // **切札を先に決めてからでないと捨てられない。**捨てる札の判断が切札に依る。
  const needsTrump = isHumanDeclarer && state.contract === 0 && state.trumpSuit === 0;
  // **サーバーが弾く選択肢は出さない。**設定でノートランプを切っていると
  // Bid が error を返すので、選べてしまうと押した瞬間に必ず失敗する。
  const availableContracts = state.config.allowNoTrump ? CONTRACTS : CONTRACTS.filter((c) => c.value === 0);
  // 選択中の契約が落ちたら、残っている先頭へ寄せる。
  const selectedContract = availableContracts.some((c) => c.value === contract) ? contract : 0;
  const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
  const suitGlyph = (suit: number): string => SUIT_GLYPHS[suit] ?? t('trumpNone');
  const contractLabel = (c: number): string => t(CONTRACTS[c]?.labelKey ?? 'withTrump');

  const canPlay = (i: number) => state.validPlays.includes(i);
  // **♥5 と ♠3 は捨てられない。**サーバーも弾くが、押せないほうが判りやすい。
  const isScoringCard = (design: string, value: number) =>
    (design === 'HEART' && value === 5) || (design === 'SPADE' && value === 3);

  const toggleSelect = (i: number) => {
    setSelected((prev) => {
      if (prev.includes(i)) return prev.filter((x) => x !== i);
      if (isDiscard) return prev.length >= KAISER_DISCARD_COUNT ? prev : [...prev, i];
      return [i];
    });
  };

  const handleDiscard = () => {
    if (selected.length !== KAISER_DISCARD_COUNT) return;
    exec('discard', { indices: selected });
    setSelected([]);
  };

  const handlePlay = () => {
    if (selected.length !== 1) return;
    exec('play', { cardIndex: selected[0] });
    setSelected([]);
  };

  const handleManualReset = () => {
    hideActionLog();
    setSelected([]);
    exec('reset');
  };

  const bidValues: number[] = [];
  for (let v = state.minBid; v <= state.maxBid; v++) bidValues.push(v);

  return (
    <GamePageShell
      title={tc('nav.kaiser')}
      gameThemeBg={gameTheme.kaiser.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBid || isHumanDeclarer || isHumanPlay}
      gamePath="/kaiser"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="kaiser-info">
              <span className="mr-4">{t('hand', { n: state.handNumber })}</span>
              <span className="mr-4">{t('target', { n: state.targetScore })}</span>
              {state.highBid && (
                <span className="mr-4" data-testid="kaiser-contract">
                  {t('contract')}: {state.highBid.value} {contractLabel(state.contract)}{' '}
                  {state.contract === 0 && suitGlyph(state.trumpSuit)}
                </span>
              )}
              {state.kittySize > 0 && <span data-testid="kaiser-kitty">{t('kitty', { n: state.kittySize })}</span>}
            </div>

            {/* The two cards that carry as much weight as all eight tricks. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-xs" data-testid="kaiser-specials">
              <div className="mb-1 text-ds-text-primary">{t('specialTitle')}</div>
              <div className="flex flex-wrap gap-x-3">
                <span className="text-ds-success font-semibold">{t('heartFive')}</span>
                <span className="text-ds-danger font-semibold">{t('spadeThree')}</span>
              </div>
              <div className="mt-1 text-ds-text-muted">{t('specialNote')}</div>
              {state.heartFiveBy >= 0 && (
                <div className="text-ds-text-muted" data-testid="kaiser-heart-five-taken">
                  {t('heartFiveTaken', { name: playerLabel(state.heartFiveBy, state.heartFiveBy === 0) })}
                </div>
              )}
              {state.spadeThreeBy >= 0 && (
                <div className="text-ds-text-muted" data-testid="kaiser-spade-three-taken">
                  {t('spadeThreeTaken', { name: playerLabel(state.spadeThreeBy, state.spadeThreeBy === 0) })}
                </div>
              )}
            </div>

            {/* Scores */}
            <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="kaiser-scores">
              <div className="mb-1 text-ds-text-primary">{t('scoreTitle')}</div>
              <div>{t('gameScores', { t0: state.teamScores[0], t1: state.teamScores[1] })}</div>
              <div>{t('handPoints', { t0: state.teamHandPoints[0], t1: state.teamHandPoints[1] })}</div>
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="kaiser-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="kaiser-player"
                  className={`text-sm py-0.5 flex items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  <span>({t('team', { n: p.team })})</span>
                  {p.isDealer && <span className="text-ds-accent">[{t('dealer')}]</span>}
                  {p.isDeclarer && <span className="text-ds-success">[{t('declarer')}]</span>}
                  {!p.isHuman && p.cards.length === 0 && <span>{t('hiddenHand', { count: p.cardCount })}</span>}
                </div>
              ))}
            </div>

            {/* Trick */}
            {state.trick.length > 0 && (
              <div className="mb-2 flex items-center gap-2" data-testid="kaiser-trick">
                <span className="text-ds-text-muted text-sm">{t('trick')}</span>
                {state.trick.map((c, i) => (
                  <CardImage key={`trick-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                ))}
              </div>
            )}

            {/* Settlement */}
            {(isHandEnd || isGameEnd) && (
              <div
                className={`mb-2 p-2 rounded text-sm ${badgeWarningColors}`}
                data-testid="kaiser-settlement"
                role="status"
                aria-live="polite"
              >
                <div className="mb-1">{t('settlementTitle')}</div>
                <div>{state.bidMade ? t('madeLine') : t('setLine')}</div>
                <div className="text-xs">{t('mustBidLine')}</div>
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.kaiser.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="kaiser-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-1">
                {(human?.cards ?? []).map((c, i) => {
                  const blockedByDiscard = isDiscard && isScoringCard(c.design, c.value);
                  const blockedByPlay = isPlay && !canPlay(i);
                  return (
                    <button
                      key={`hand-${c.design}-${c.value}-${i}`}
                      type="button"
                      data-hint-action="play"
                      onClick={() => toggleSelect(i)}
                      disabled={loading || blockedByDiscard || blockedByPlay}
                      className={`rounded ${selected.includes(i) ? 'ring-2 ring-ds-accent' : ''} ${
                        blockedByDiscard || blockedByPlay ? 'opacity-40' : ''
                      }`}
                    >
                      <CardImage card={c} width={cardWidth} />
                    </button>
                  );
                })}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {isHumanBid && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="kaiser-bid-notice">
                {t('bidNotice')}
              </div>
            )}
            {needsTrump && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="kaiser-trump-notice">
                {t('trumpNotice')}
              </div>
            )}
            {isHumanDeclarer && !needsTrump && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="kaiser-discard-notice">
                {t('discardNotice')}
              </div>
            )}
            {isHumanPlay && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="kaiser-play-notice">
                {t('playNotice')}
              </div>
            )}

            <label className="flex items-center gap-1 text-ds-text-primary text-xs w-full justify-center cursor-pointer min-h-[44px]">
              <input
                type="checkbox"
                checked={frontendHintEnabled}
                onChange={(e) => setFrontendHintEnabled(e.target.checked)}
              />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="kaiser-actions">
              {isHumanBid && (
                <>
                  <label
                    className="text-ds-text-muted text-xs flex items-center gap-1"
                    htmlFor="kaiser-contract-select"
                  >
                    {t('contractLabel')}
                    <select
                      id="kaiser-contract-select"
                      className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                      value={selectedContract}
                      onChange={(e) => setContract(Number(e.target.value))}
                    >
                      {availableContracts.map((c) => (
                        <option key={c.value} value={c.value}>
                          {t(c.labelKey)}
                        </option>
                      ))}
                    </select>
                  </label>
                  {bidValues.map((v) => (
                    <button
                      key={`bid-${v}`}
                      type="button"
                      className={btnPrimary}
                      onClick={() => exec('bid', { bid: v, contract: selectedContract })}
                      disabled={loading}
                    >
                      {t('bidButton', { n: v })}
                    </button>
                  ))}
                  <button type="button" className={btnWarning} onClick={() => exec('pass')} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}

              {needsTrump &&
                [1, 2, 3, 4].map((s) => (
                  <button
                    key={`trump-${s}`}
                    type="button"
                    className={btnPrimary}
                    onClick={() => exec('trump', { suit: s })}
                    disabled={loading}
                  >
                    {t('trumpButton', { suit: SUIT_GLYPHS[s] })}
                  </button>
                ))}

              {isHumanDeclarer && !needsTrump && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handleDiscard}
                  disabled={loading || selected.length !== KAISER_DISCARD_COUNT}
                >
                  {t('discardButton')}
                </button>
              )}

              {isHumanPlay && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handlePlay}
                  disabled={loading || selected.length !== 1}
                >
                  {t('playButton')}
                </button>
              )}

              {isHandEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={() => exec('next')} disabled={loading}>
                  {t('nextHand')}
                </button>
              )}

              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon ? t('win') : t('lose')}
                </span>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="kaiser-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
