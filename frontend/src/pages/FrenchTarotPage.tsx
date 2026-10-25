import { useEffect } from 'react';
import type { frenchtarotApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import {
  CPU_DIFFICULTY_OPTIONS,
  FRENCH_TAROT_ECART_COUNT,
  TARGET_DEALS_OPTIONS,
  useFrenchTarotGame,
} from '../hooks/useFrenchTarotGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { Card, FrenchTarotResponse } from '../types/card';
import { FrenchTarotPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { FRENCH_TAROT_HELP, parseFrenchTarotCommand } from '../utils/cli/commands/frenchtarotCommands';
import { formatFrenchTarotState } from '../utils/cli/formatters/frenchtarotFormatter';
import { frenchTarotTarget, heldBouts } from '../utils/frenchTarotBouts';
import { frenchTarotUnburiableReason } from '../utils/frenchtarotEcart';
import { isRequestedHint } from '../utils/hintRequest';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** French Tarot tutorial step definitions. */
const FRENCH_TAROT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="frenchtarot-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="frenchtarot-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="frenchtarot-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="frenchtarot-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="frenchtarot-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const FRENCH_TAROT_PHASE_KEYS: Readonly<Record<number, string>> = {
  [FrenchTarotPhase.BID]: 'bid',
  [FrenchTarotPhase.CHIEN]: 'chien',
  [FrenchTarotPhase.PLAY]: 'play',
  [FrenchTarotPhase.TRICK_END]: 'trickEnd',
  [FrenchTarotPhase.ROUND_END]: 'roundEnd',
  [FrenchTarotPhase.GAME_END]: 'gameEnd',
};

/** Contract i18n keys indexed by contract/bid value (0=none, 1=Petite, 2=Garde, 3=Garde Sans, 4=Garde Contre). */
const CONTRACT_KEYS = [
  'contractNone',
  'contractPetite',
  'contractGarde',
  'contractGardeSans',
  'contractGardeContre',
] as const;

/** Outcome i18n keys indexed by outcome value (0=none, 1=Win/made, 2=Loss/failed). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeWin', 'outcomeLoss'] as const;

/** Selectable bid contracts with their numeric strength for the "must beat the highest bid" rule. */
const BID_CHOICES = [
  { contract: 'petite', value: 1, labelKey: 'bidPetite' },
  { contract: 'garde', value: 2, labelKey: 'bidGarde' },
  { contract: 'gardesans', value: 3, labelKey: 'bidGardeSans' },
  { contract: 'gardecontre', value: 4, labelKey: 'bidGardeContre' },
] as const;

/** True for cards that may NOT be buried in the écart: Kings (Roi, value 14) and the Excuse (gold face). */
function isUndiscardable(card: Card): boolean {
  return card.value === 14 || card.color === 'gold';
}

type FrenchTarotGame = ReturnType<typeof useFrenchTarotGame>;
type FrenchTarotView = ReturnType<typeof useFrenchTarotView>;
type FrenchTarotContext = GamePageContext<FrenchTarotResponse, FrenchTarotGame, FrenchTarotView>;

function useFrenchTarotView({ game }: { game: FrenchTarotGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('frenchtarot', FRENCH_TAROT_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

function Board({ state, t, view }: FrenchTarotContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  const isBidPhase = state.phase === FrenchTarotPhase.BID;
  const isChienPhase = state.phase === FrenchTarotPhase.CHIEN;
  const isPlayPhase = state.phase === FrenchTarotPhase.PLAY;
  const isTrickEnd = state.phase === FrenchTarotPhase.TRICK_END;
  const isRoundEnd = state.phase === FrenchTarotPhase.ROUND_END;
  const isGameEnd = state.phase === FrenchTarotPhase.GAME_END || state.gameEndFlag;
  const trickPoints = state.trickHalfPoints / 2;

  // Bouts (oudlers) the human currently holds in hand: the 21, the Petit (trump 1), and the Excuse.
  const heldBoutList = humanPlayer ? heldBouts(humanPlayer.cards) : [];

  const contractLabel = t(CONTRACT_KEYS[state.contract] ?? 'contractNone');

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span>{t('contract', { contract: contractLabel })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            winnerIdx={isTrickEnd ? state.lastTrickWinner : undefined}
            winnerLabel={
              isTrickEnd
                ? t('previousTrickWinner', { name: findPlayerName(state.players, state.lastTrickWinner) })
                : undefined
            }
            dataTutorial="frenchtarot-trick-display"
          />
          {isTrickEnd && (
            <div
              className="mt-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              data-testid="frenchtarot-trick-summary"
            >
              {t('trickSummary', {
                points: Number.isInteger(trickPoints) ? trickPoints : trickPoints.toFixed(1),
                count: state.trickBouts,
              })}
            </div>
          )}
          {/* Chien reveal for the human declarer during écart */}
          {isChienPhase && state.chienRevealed && state.chien.length > 0 && (
            <div className="mt-2 p-2 rounded bg-black/30" data-testid="frenchtarot-chien">
              <div className="text-ds-text-muted text-sm mb-1">{t('chienLabel')}</div>
              <div className="flex flex-wrap gap-1">
                {state.chien.map((c, i) => (
                  <CardImage key={`chien-${c.design}-${c.value}-${i}`} card={c} width={Math.round(cardWidth * 0.7)} />
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="frenchtarot-info">
          {/* Per-player scores with a declarer badge */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {state.players.map((p) => (
              <div key={p.id} className="py-0.5 flex items-center gap-2">
                <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
                  {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                </span>
                {p.isDeclarer && (
                  <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('declarerBadge')}</span>
                )}
              </div>
            ))}
          </div>

          {/* Bouts (oudlers) held in the human's hand — drive the contract's point target. */}
          {humanPlayer && humanPlayer.cards.length > 0 && (
            <div className="mb-2 p-2 rounded bg-black/30" data-testid="frenchtarot-bouts">
              <div className="text-ds-text-muted text-sm mb-1">{t('bouts.title', { count: heldBoutList.length })}</div>
              {heldBoutList.length > 0 ? (
                <div className="flex flex-wrap gap-1">
                  {heldBoutList.map((b) => (
                    <span
                      key={b}
                      className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}
                      data-testid={`frenchtarot-bout-${b}`}
                    >
                      {t(`bouts.${b}`)}
                    </span>
                  ))}
                </div>
              ) : (
                <div className="text-ds-text-muted text-sm">{t('bouts.none')}</div>
              )}
              {/* The target follows the bouts the declarer *wins in tricks*, not
                  the ones in hand. During the bid those coincide closely enough to
                  judge a bid on; afterwards the held list drains as the bouts are
                  played, and for a non-declarer it never meant anything (#4857). */}
              {isBidPhase ? (
                <div className="text-ds-text-muted text-xs mt-1" data-testid="frenchtarot-bouts-target">
                  {t('bouts.target', {
                    count: heldBoutList.length,
                    points: frenchTarotTarget(heldBoutList.length),
                  })}
                </div>
              ) : (
                <div className="text-ds-text-muted text-xs mt-1" data-testid="frenchtarot-bouts-note">
                  {t('bouts.boutsNote')}
                </div>
              )}
            </div>
          )}

          {/* Players: cards / tricks / captured points */}
          {isMobile ? (
            <div className="mb-2 p-2 rounded bg-black/30">
              <div className="text-ds-text-muted text-sm">
                {state.players.map((p) => (
                  <div key={p.id} className="py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('tricks', { count: p.trickCount })} |{' '}
                    {t('points', { points: p.cardPoints })} | {t('boutsWon', { count: p.bouts })}
                  </div>
                ))}
              </div>
              <details className="mt-1">
                <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('cardCounts')}</summary>
                <div className="mt-1">
                  {state.players.map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-xs py-0.5">
                      {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })}
                    </div>
                  ))}
                </div>
              </details>
            </div>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('tricks', { count: p.trickCount })} | {t('points', { points: p.cardPoints })} |{' '}
                  {t('boutsWon', { count: p.bouts })}
                </div>
              ))}
            </div>
          )}

          {/* Live target: the declarer side's current target from the server */}
          {(isPlayPhase || isTrickEnd) && state.declarerIdx >= 0 && (
            <div className="text-ds-text-muted text-sm" data-testid="frenchtarot-live-target">
              {t('roundResult.target', { points: state.target })}
            </div>
          )}
          {/* Round result: the deal outcome (contract made / failed) */}
          {(isRoundEnd || isGameEnd) && state.outcome > 0 && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="frenchtarot-result">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.outcome', { outcome: t(OUTCOME_KEYS[state.outcome] ?? 'outcomeNone') })}</div>
              {state.declarerIdx >= 0 && (
                <>
                  <div>
                    {t('roundResult.declarer', {
                      name: playerName(state.declarerIdx, state.declarerIdx === humanIdx),
                    })}
                  </div>
                  <div>{t('roundResult.contract', { contract: contractLabel })}</div>
                  <div>
                    {t('roundResult.captured', {
                      points: state.declarerCaptured,
                    })}
                  </div>
                  <div>{t('roundResult.target', { points: state.target })}</div>
                  {/* **プティ・オ・ブーが乗ると、獲得点から逆算した数字と精算が
                      合わなくなる。**ルールは実装済みで精算にも乗っているのに、
                      どちらの画面にも出ていなかった (#6509)。0 は未発生。 */}
                  {state.petitAuBoutDelta !== 0 && (
                    <div data-testid="ft-petit-au-bout">
                      {t('roundResult.petitAuBout', {
                        side: t(
                          state.petitAuBoutDelta > 0
                            ? 'roundResult.petitAuBoutDeclarer'
                            : 'roundResult.petitAuBoutDefenders',
                        ),
                        points: state.petitAuBoutDelta,
                      })}
                    </div>
                  )}
                </>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterStatus({ state, frontendHint, t }: FrenchTarotContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="frenchtarot-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndices &&
              state.hint.cardIndices.length > 0 &&
              ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterHand({ game, state, t, view }: FrenchTarotContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);

  const isBidPhase = state.phase === FrenchTarotPhase.BID;
  const isChienPhase = state.phase === FrenchTarotPhase.CHIEN;
  const isPlayPhase = state.phase === FrenchTarotPhase.PLAY;

  const canBid = isBidPhase && state.isHumanBidTurn;
  const canDiscard = isChienPhase && state.isHumanDiscard;
  const canPlay = isPlayPhase && state.isHumanTurn;

  const highestBidLabel = state.highestBid > 0 ? t(CONTRACT_KEYS[state.highestBid] ?? 'contractNone') : t('bidNone');

  // During the écart, restrict selection to buryable cards (no Kings / Excuse).
  const discardableIndices =
    canDiscard && humanPlayer
      ? humanPlayer.cards.map((c, i) => (isUndiscardable(c) ? -1 : i)).filter((i) => i >= 0)
      : undefined;

  const handValidIndices = canPlay ? state.playableIndices : canDiscard ? discardableIndices : undefined;

  // During the écart, explain per-card why an un-buriable card cannot go into the
  // chien (King / Excuse / bout / trump) via the card tooltip. Purely additive —
  // it never blocks selection; the backend still rejects illegal buries.
  const ecartTitleFor = (idx: number): string | undefined => {
    const card = humanPlayer?.cards[idx];
    if (!card) return undefined;
    const reason = frenchTarotUnburiableReason(card);
    return reason ? t(`chienUnburiable.${reason}`) : undefined;
  };

  return (
    <>
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <div data-testid="frenchtarot-prompt-live" role="status" aria-live="polite">
        {canBid && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="frenchtarot-bid-prompt">
            {t('bidPhase', { bid: highestBidLabel })}
          </div>
        )}
        {canDiscard && (
          <div
            className="mb-1 text-center text-sm text-ds-accent font-semibold"
            data-testid="frenchtarot-discard-prompt"
          >
            {t('chienPhase', { count: game.selectedCardIndices.length })}
          </div>
        )}
      </div>
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={game.selectedCardIndices}
          toggleCard={game.toggleCard}
          cardWidth={view.cardWidth}
          isMobile={view.isMobile}
          dataTutorialPrefix="frenchtarot"
          stackRows
          validIndices={handValidIndices}
          restrictedTooltip={canDiscard ? t('chienRestricted') : t('playButton')}
          cardTitleFor={canDiscard ? ecartTitleFor : undefined}
          cardStatusFor={canDiscard ? ecartTitleFor : undefined}
        />
      )}
    </>
  );
}

function primaryActions({ game, state, t, busy }: FrenchTarotContext) {
  const isBidPhase = state.phase === FrenchTarotPhase.BID;
  const isChienPhase = state.phase === FrenchTarotPhase.CHIEN;
  const isPlayPhase = state.phase === FrenchTarotPhase.PLAY;
  const isTrickEnd = state.phase === FrenchTarotPhase.TRICK_END;
  const isRoundEnd = state.phase === FrenchTarotPhase.ROUND_END;

  const canBid = isBidPhase && state.isHumanBidTurn;
  const canDiscard = isChienPhase && state.isHumanDiscard;
  const canPlay = isPlayPhase && state.isHumanTurn;

  if (!canBid && !canDiscard && !canPlay && !isTrickEnd && !isRoundEnd) return null;

  return (
    <div className="flex flex-wrap gap-2 items-center" data-tutorial="frenchtarot-action-buttons">
      {canBid && (
        <>
          <button type="button" className={btnSecondary} onClick={game.handlePass} disabled={busy}>
            {t('bidPass')}
          </button>
          {BID_CHOICES.map((c) => (
            <button
              key={c.contract}
              type="button"
              className={btnPrimary}
              onClick={() => game.handleBid(c.contract)}
              disabled={busy || c.value <= state.highestBid}
            >
              {t(c.labelKey)}
            </button>
          ))}
        </>
      )}
      {canDiscard && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handleDiscard}
          disabled={busy || game.selectedCardIndices.length !== FRENCH_TAROT_ECART_COUNT}
        >
          {t('discardButton', { count: game.selectedCardIndices.length })}
        </button>
      )}
      {canPlay && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const frenchTarotDefinition: GameDefinition<
  FrenchTarotResponse,
  FrenchTarotGame,
  Parameters<typeof frenchtarotApi.exec>,
  FrenchTarotView
> = {
  key: 'frenchtarot',
  useGame: useFrenchTarotGame,
  useView: useFrenchTarotView,
  tutorial: FRENCH_TAROT_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 18 },
  cli: () => ({
    parseCommand: parseFrenchTarotCommand,
    formatResponse: formatFrenchTarotState,
    helpText: FRENCH_TAROT_HELP,
  }),
  phase: ({ state, view }) => {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const isGameEnd = state.phase === FrenchTarotPhase.GAME_END || state.gameEndFlag;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerPlayer === humanIdx,
    };
  },
  shell: ({ state }) => ({
    isHumanTurn: state.isHumanTurn,
  }),
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.frenchtarotConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetDeals',
          label: t('settings.targetDeals'),
          value: game.frenchtarotConfig.targetDeals,
          options: TARGET_DEALS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetDeals', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  footerTall: true,
  resetTutorialId: 'frenchtarot-reset-button',
  actionsClassName: '',
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the French Tarot (フレンチタロット) game page: a 4-player 78-card trick-taker with a Petite/Garde/Garde-Sans/Garde-Contre auction, a 6-card chien écart, and bouts-based deal scoring. */
export const FrenchTarotPage = createGamePage(frenchTarotDefinition);
