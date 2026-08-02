import { useEffect, useMemo, useState } from 'react';
import { sixBidSoloApi } from '../api/gameApi';
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
import type { SixBidSoloResponse } from '../types/card';
import { SixBidSoloPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSixBidSoloCommand, SIXBIDSOLO_HELP } from '../utils/cli/commands/sixbidsoloCommands';
import { formatSixBidSoloState } from '../utils/cli/formatters/sixbidsoloFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** The six bids, in ascending order. Index 0 is a pass and is never offered. */
const BIDS = [1, 2, 3, 4, 5, 6];

/** Trump suits: 1=♠ 2=♣ 3=♥ 4=♦. */
const SUITS = [1, 2, 3, 4];

/** Ranks a call solo may name, high to low within the 36-card pack. */
const CALL_RANKS = [1, 13, 12, 11, 10, 9, 8, 7, 6];

/** The bid that requires a named card (sync: `SixBidSoloBidCall`). */
const CALL_SOLO = 6;

/** Six-Bid Solo tutorial step definitions. */
const SIXBIDSOLO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sixbidsolo-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sixbidsolo-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sixbidsolo-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sixbidsolo-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sixbidsolo-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SIXBIDSOLO_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SixBidSoloPhase.BID]: 'bid',
  [SixBidSoloPhase.DECLARE]: 'declare',
  [SixBidSoloPhase.PLAY]: 'play',
  [SixBidSoloPhase.HAND_END]: 'handEnd',
  [SixBidSoloPhase.GAME_END]: 'gameEnd',
};

/** Renders the Six-Bid Solo game page: a 3-player skat descendant with six bids. */
export const SixBidSoloPage = withTutorial(SixBidSoloPageContent, 'sixbidsolo', SIXBIDSOLO_TUTORIAL_STEPS);

/** Inner content of the Six-Bid Solo page, wrapped by TutorialProvider. */
function SixBidSoloPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sixbidsolo');
  const { state, loading, error, exec, retry } = useGameApi(sixBidSoloApi.exec);

  const [bidKind, setBidKind] = useState(1);
  const [suit, setSuit] = useState(1);
  const [calledSuit, setCalledSuit] = useState(1);
  const [calledValue, setCalledValue] = useState(1);
  const [selected, setSelected] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sixbidsolo');
  const cliConfig: CliGameConfig<SixBidSoloResponse, Parameters<typeof sixBidSoloApi.exec>> = useMemo(
    () => ({
      gameName: 'sixbidsolo',
      parseCommand: parseSixBidSoloCommand,
      formatResponse: formatSixBidSoloState,
      helpText: SIXBIDSOLO_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('sixbidsolo', SIXBIDSOLO_PHASE_KEYS);

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('sixbidsolo', state);

  if (!state)
    return <GameSkeleton gameKey="sixbidsolo" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 11 }} />;

  const human = state.players.find((p) => p.isHuman);
  const isBid = state.phase === SixBidSoloPhase.BID;
  const isDeclare = state.phase === SixBidSoloPhase.DECLARE;
  const isPlay = state.phase === SixBidSoloPhase.PLAY;
  const isHandEnd = state.phase === SixBidSoloPhase.HAND_END;
  const isGameEnd = state.phase === SixBidSoloPhase.GAME_END || state.gameEndFlag;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const isHumanBid = isBid && state.bidPlayerIdx === 0 && !isGameEnd;
  // **切札を名乗るのは落札者。**手番ではない。
  const isHumanDeclare = isDeclare && state.declarerIdx === 0 && !isGameEnd;
  const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
  const bidLabel = (kind: number): string => t(`bidName.${kind}`);
  const suitLabel = (s: number): string => (s > 0 ? t(`suitName.${s}`) : t('noTrump'));

  const canPlay = (i: number) => state.validPlays.includes(i);

  // **上回る宣言だけが通る。**通らない値は出さない。
  const bids = BIDS.filter((b) => state.highBid === null || b > state.highBid.kind);
  const selectedBid = bids.length > 0 && !bids.includes(bidKind) ? bids[0] : bidKind;

  // **指名札はコール・ソロだけに要る。**
  const needsCalledCard = state.highBid?.kind === CALL_SOLO;

  const handleBid = () => {
    exec('bid', { bid: selectedBid });
  };

  const handleDeclare = () => {
    exec('declare', needsCalledCard ? { suit, calledSuit, calledValue } : { suit });
  };

  const handlePlay = () => {
    if (selected === null) return;
    exec('play', { cardIndex: selected });
    setSelected(null);
  };

  const handleManualReset = () => {
    hideActionLog();
    setSelected(null);
    exec('reset');
  };

  const result = state.lastResult;

  return (
    <GamePageShell
      title={tc('nav.sixbidsolo')}
      gameThemeBg={gameTheme.sixbidsolo.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBid || isHumanDeclare || isHumanPlay}
      gamePath="/sixbidsolo"
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
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="sixbidsolo-info">
              <span className="mr-4">{t('hand', { n: state.handNumber, total: state.targetHands })}</span>
              {state.highBid && (
                <span className="mr-4" data-testid="sixbidsolo-contract">
                  {t('contract')}: {bidLabel(state.highBid.kind)} / {t('trump')}:{' '}
                  {state.declared ? suitLabel(state.trumpSuit) : t('trumpUndecided')} /{' '}
                  {t('target', { n: state.bidTargets[state.highBid.kind] ?? 0 })}
                </span>
              )}
            </div>

            {/* Eleven each plus a three-card widow — the deal the issue got wrong. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-xs" data-testid="sixbidsolo-widow">
              <div className="text-ds-text-primary">
                {t('widowTitle')}:{' '}
                {state.widow.length > 0 ? (
                  <span className="inline-flex gap-1 align-middle">
                    {state.widow.map((c, i) => (
                      <CardImage key={`widow-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                    ))}
                  </span>
                ) : (
                  t('widowHidden', { count: state.widowSize })
                )}
              </div>
              <div className="mt-1 text-ds-text-muted">{t('widowNote')}</div>
              {state.calledCard && (
                <div className="mt-1 text-ds-text-muted" data-testid="sixbidsolo-called">
                  {t('calledCard')}: {t('calledNote')}
                </div>
              )}
              {state.spreadOpen && (
                <div className="mt-1 text-ds-text-muted" data-testid="sixbidsolo-spread">
                  {t('spreadNote')}
                </div>
              )}
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="sixbidsolo-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="sixbidsolo-player"
                  className={`text-sm py-0.5 flex items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  {p.isDealer && <span className="text-ds-accent">[{t('dealer')}]</span>}
                  {p.isDeclarer && <span className="text-ds-success">[{t('declarer')}]</span>}
                  <span>{t('points', { n: p.points })}</span>
                  <span>{t('tricksWon', { n: p.tricksWon })}</span>
                  <span>{t('score', { n: p.score })}</span>
                  {!p.isHuman && p.cards.length === 0 && <span>{t('hiddenHand', { count: p.cardCount })}</span>}
                </div>
              ))}
            </div>

            {/* Trick */}
            {state.trick.length > 0 && (
              <div className="mb-2 flex items-center gap-2" data-testid="sixbidsolo-trick">
                <span className="text-ds-text-muted text-sm">{t('trick')}</span>
                {state.trick.map((c, i) => (
                  <CardImage key={`trick-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                ))}
              </div>
            )}

            {/* Settlement */}
            {(isHandEnd || isGameEnd) && result && (
              <div
                className={`mb-2 p-2 rounded text-sm ${badgeWarningColors}`}
                data-testid="sixbidsolo-settlement"
                role="status"
                aria-live="polite"
              >
                <div className="mb-1">{t('settlementTitle')}</div>
                <div>
                  {result.made
                    ? t('madeLine', {
                        bid: bidLabel(result.kind),
                        points: result.declarerPoints,
                        need: result.target,
                      })
                    : t('setLine', {
                        bid: bidLabel(result.kind),
                        points: result.declarerPoints,
                        need: result.target,
                      })}
                </div>
                {/* **ウィドウは宣言者に入る。ミゼール系だけは 0。** */}
                <div data-testid="sixbidsolo-widow-credit">{t('widowCredit', { n: result.widowPoints })}</div>
                <div>{t('valueLine', { n: result.value })}</div>
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
          <GameFooter className={`${gameTheme.sixbidsolo.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="sixbidsolo-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-1">
                {(human?.cards ?? []).map((c, i) => (
                  <button
                    key={`hand-${c.design}-${c.value}-${i}`}
                    type="button"
                    data-hint-action="play"
                    onClick={() => setSelected(i)}
                    disabled={loading || (isPlay && !canPlay(i))}
                    className={`rounded ${selected === i ? 'ring-2 ring-ds-accent' : ''} ${
                      isPlay && !canPlay(i) ? 'opacity-40' : ''
                    }`}
                  >
                    <CardImage card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {isHumanBid && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="sixbidsolo-bid-notice">
                <div>{t('bidNotice')}</div>
                {/* **通常ビッドは 61 点以上。**60 ちょうどでは足りない。 */}
                <div data-testid="sixbidsolo-base-note">{t('baseTargetNote', { total: state.totalPoints })}</div>
                {/* **ミゼールは 0 トリックではなく 0 点。** */}
                <div data-testid="sixbidsolo-misere-note">{t('misereNote')}</div>
              </div>
            )}
            {isHumanDeclare && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="sixbidsolo-declare-notice">
                <div>{t('declareNotice')}</div>
                {needsCalledCard && <div data-testid="sixbidsolo-call-notice">{t('callNotice')}</div>}
              </div>
            )}
            {isHumanPlay && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="sixbidsolo-play-notice">
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="sixbidsolo-actions">
              {isHumanBid && (
                <>
                  <label className="text-ds-text-muted text-xs flex items-center gap-1" htmlFor="sixbidsolo-bid-select">
                    {t('bidLabel')}
                    <select
                      id="sixbidsolo-bid-select"
                      className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                      value={selectedBid}
                      onChange={(e) => setBidKind(Number(e.target.value))}
                    >
                      {bids.map((b) => (
                        <option key={b} value={b}>
                          {bidLabel(b)} ({t('target', { n: state.bidTargets[b] ?? 0 })})
                        </option>
                      ))}
                    </select>
                  </label>
                  <button type="button" className={btnPrimary} onClick={handleBid} disabled={loading}>
                    {t('bidButton')}
                  </button>
                  <button type="button" className={btnWarning} onClick={() => exec('pass')} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}

              {isHumanDeclare && (
                <>
                  <label
                    className="text-ds-text-muted text-xs flex items-center gap-1"
                    htmlFor="sixbidsolo-suit-select"
                  >
                    {t('suitLabel')}
                    <select
                      id="sixbidsolo-suit-select"
                      className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                      value={suit}
                      onChange={(e) => setSuit(Number(e.target.value))}
                    >
                      {SUITS.map((s) => (
                        <option key={s} value={s}>
                          {suitLabel(s)}
                        </option>
                      ))}
                    </select>
                  </label>
                  {needsCalledCard && (
                    <>
                      <label
                        className="text-ds-text-muted text-xs flex items-center gap-1"
                        htmlFor="sixbidsolo-called-suit-select"
                      >
                        {t('calledSuitLabel')}
                        <select
                          id="sixbidsolo-called-suit-select"
                          className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                          value={calledSuit}
                          onChange={(e) => setCalledSuit(Number(e.target.value))}
                        >
                          {SUITS.map((s) => (
                            <option key={s} value={s}>
                              {suitLabel(s)}
                            </option>
                          ))}
                        </select>
                      </label>
                      <label
                        className="text-ds-text-muted text-xs flex items-center gap-1"
                        htmlFor="sixbidsolo-called-value-select"
                      >
                        {t('calledValueLabel')}
                        <select
                          id="sixbidsolo-called-value-select"
                          className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                          value={calledValue}
                          onChange={(e) => setCalledValue(Number(e.target.value))}
                        >
                          {CALL_RANKS.map((v) => (
                            <option key={v} value={v}>
                              {v}
                            </option>
                          ))}
                        </select>
                      </label>
                    </>
                  )}
                  <button type="button" className={btnPrimary} onClick={handleDeclare} disabled={loading}>
                    {t('declareButton')}
                  </button>
                </>
              )}

              {isHumanPlay && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handlePlay}
                  disabled={loading || selected === null}
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
                dataTutorial="sixbidsolo-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
