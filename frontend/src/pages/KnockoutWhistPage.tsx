import { useEffect } from 'react';
import type { knockoutWhistApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, useKnockoutWhistGame } from '../hooks/useKnockoutWhistGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeInfoColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { KnockoutWhistResponse } from '../types/card';
import { KnockoutWhistPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { suitName } from '../utils/cardUtils';
import { KNOCKOUT_WHIST_HELP, parseKnockoutWhistCommand } from '../utils/cli/commands/knockoutWhistCommands';
import { formatKnockoutWhistState } from '../utils/cli/formatters/knockoutWhistFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Knockout Whist tutorial step definitions. */
const KNOCKOUT_WHIST_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="knockoutwhist-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="knockoutwhist-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="knockoutwhist-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="knockoutwhist-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="knockoutwhist-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KNOCKOUT_WHIST_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KnockoutWhistPhase.PLAY]: 'play',
  [KnockoutWhistPhase.TRICK_END]: 'trickEnd',
  [KnockoutWhistPhase.ROUND_END]: 'roundEnd',
  [KnockoutWhistPhase.GAME_END]: 'gameEnd',
  [KnockoutWhistPhase.TRUMP_SELECT]: 'trumpSelect',
};

type KnockoutWhistGame = ReturnType<typeof useKnockoutWhistGame>;
type KnockoutWhistView = ReturnType<typeof useKnockoutWhistView>;
type KnockoutWhistContext = GamePageContext<KnockoutWhistResponse, KnockoutWhistGame, KnockoutWhistView>;

function useKnockoutWhistView({ game }: { game: KnockoutWhistGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('knockoutwhist', KNOCKOUT_WHIST_PHASE_KEYS);

  return { cardWidth, isMobile, phaseNames };
}

function AboveContent({ state, t }: KnockoutWhistContext) {
  const isRoundEnd = state.phase === KnockoutWhistPhase.ROUND_END;
  const isGameEnd = state.phase === KnockoutWhistPhase.GAME_END || state.gameEndFlag;
  // Knockout Whist deals one fewer card each round, bottoming out at 1 (see KnockoutWhist.startRound in the Go domain).
  const nextHandSize = Math.max(state.handSize - 1, 1);

  const roundResultAnnouncement =
    (isRoundEnd || isGameEnd) && state.roundWinnerIdx >= 0
      ? [
          t('roundResult.winner', {
            name: findPlayerName(state.players, state.roundWinnerIdx),
          }),
          ...state.roundSurvived.map((idx) => t('roundResult.survived', { name: findPlayerName(state.players, idx) })),
          ...state.roundEliminated.map((idx) =>
            t('roundResult.eliminated', { name: findPlayerName(state.players, idx) }),
          ),
          ...(isRoundEnd && !isGameEnd
            ? [
                t(nextHandSize === 1 ? 'roundResult.finalRoundPreview' : 'roundResult.nextRoundPreview', {
                  count: nextHandSize,
                  name: findPlayerName(state.players, state.roundWinnerIdx),
                }),
              ]
            : []),
        ].join(t('listSeparator'))
      : '';

  return (
    <div data-testid="kw-round-result-live" role="status" aria-live="polite" className="sr-only">
      {roundResultAnnouncement}
    </div>
  );
}

function Board({ state, t, view }: KnockoutWhistContext) {
  const { cardWidth, isMobile } = view;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanEliminated = state.players[humanIdx]?.eliminated ?? false;
  const isPlayPhase = state.phase === KnockoutWhistPhase.PLAY;
  const isTrickEnd = state.phase === KnockoutWhistPhase.TRICK_END;
  const isRoundEnd = state.phase === KnockoutWhistPhase.ROUND_END;
  const isGameEnd = state.phase === KnockoutWhistPhase.GAME_END || state.gameEndFlag;

  // Knockout Whist deals one fewer card each round, bottoming out at 1 (see KnockoutWhist.startRound in the Go domain).
  const nextHandSize = Math.max(state.handSize - 1, 1);
  // Show a spectator banner while the human is knocked out but the match continues among the CPUs.
  const showSpectatorBanner = isHumanEliminated && !isGameEnd;
  const trumpSymbol = suitSymbolAt(state.trumpSuit, '');

  /** Renders one player panel showing Dogbones, round tricks, and elimination / leader / round-winner state. */
  const renderPlayerPanel = (p: KnockoutWhistResponse['players'][number]) => {
    const isLeader = !isGameEnd && state.leadPlayerIdx === p.id;
    const isRoundWinner = (isRoundEnd || isGameEnd) && state.roundWinnerIdx === p.id;
    const isDogboneWarning = !p.eliminated && p.dogbones === 0;
    const showHandCount = !p.eliminated;
    const showNextHandSize = showHandCount && (isPlayPhase || isTrickEnd) && state.handSize > 1;
    return (
      <div
        key={p.id}
        className={`py-0.5 flex flex-wrap items-center gap-2 ${p.eliminated ? 'opacity-70 line-through' : ''}`}
        data-eliminated={p.eliminated ? 'true' : 'false'}
      >
        <span className={p.eliminated ? '' : 'text-ds-text-muted'}>
          {playerName(p.id, p.isHuman)}
          {p.eliminated
            ? ` — ${t('eliminated')}`
            : ` — ${t('roundTricks', { count: p.roundTricks })} · ${t('dogbones', { count: p.dogbones })}`}
        </span>
        {state.dealerIdx === p.id && (
          <span
            className={`px-1.5 py-0.5 rounded text-xs whitespace-nowrap ${badgeInfoColors}`}
            data-testid="kw-dealer-badge"
          >
            {t('dealerBadge')}
          </span>
        )}
        {showHandCount && (
          <span data-testid="kw-player-hand-count" className="text-ds-text-muted whitespace-nowrap">
            {t('playerHandCount.current', { count: p.cardCount })}
            {showNextHandSize && ` · ${t('playerHandCount.next', { count: nextHandSize })}`}
          </span>
        )}
        {isLeader && (
          <span className={`px-1.5 py-0.5 rounded text-xs whitespace-nowrap ${badgeInfoColors}`}>{t('leader')}</span>
        )}
        {isRoundWinner && (
          <span className={`px-1.5 py-0.5 rounded text-xs whitespace-nowrap ${badgeWarningColors}`}>
            {t('roundWinner')}
          </span>
        )}
        {isDogboneWarning && (
          <span
            className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}
            data-testid="kw-dogbone-warning-badge"
          >
            {t('dogboneWarning')}
          </span>
        )}
      </div>
    );
  };

  return (
    <>
      {showSpectatorBanner && (
        <div
          role="status"
          data-testid="kw-spectator-banner"
          className="mb-3 p-2 rounded border border-ds-warning/50 bg-ds-surface text-ds-warning text-center text-sm font-medium"
        >
          {t('spectatorBanner', { n: state.activeCount })}
        </div>
      )}

      <div
        className="text-ds-text-primary text-center mb-2 flex flex-wrap justify-center gap-x-4 gap-y-1"
        data-testid="kw-status-row"
      >
        <span className="whitespace-nowrap">{t('round', { n: state.roundNumber })}</span>
        <span className="whitespace-nowrap">{t('handSize', { n: state.handSize })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        <span className="whitespace-nowrap">{t('trump', { suit: trumpSymbol })}</span>
        <span>{t('activeCount', { n: state.activeCount })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="knockoutwhist-trick-display"
          />
        </div>

        {/* Right: info sidebar — 4 player panels */}
        <div data-tutorial="knockoutwhist-info">
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1 text-sm">{state.players.map(renderPlayerPanel)}</div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30 text-sm">{state.players.map(renderPlayerPanel)}</div>
          )}

          {/* Round result */}
          {(isRoundEnd || isGameEnd) && state.roundWinnerIdx >= 0 && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>
                {t('roundResult.winner', {
                  name: findPlayerName(state.players, state.roundWinnerIdx),
                })}
              </div>
              {/* **そのラウンドで何が起きたかを一目で。**Dogbone 消費と脱落は
                  `ScoreRound` が毎ラウンド判定する本質的な出来事なのに、結果は
                  プレイヤー一覧のグレー表示と [Dogbone 数] の増減からしか
                  読み取れなかった (#6446)。サーバに聞くしかない ── 前ラウンドで
                  脱落した席も毎ラウンド `roundTricks === 0` になるので、
                  クライアントでは「今ラウンド脱落した人」を導出できない。 */}
              {state.roundSurvived.length > 0 && (
                <div className="mt-1" data-testid="kw-round-survivors">
                  {state.roundSurvived
                    .map((idx) => t('roundResult.survived', { name: findPlayerName(state.players, idx) }))
                    .join(' / ')}
                </div>
              )}
              {state.roundEliminated.length > 0 && (
                <div className="mt-1" data-testid="kw-round-eliminated">
                  {state.roundEliminated
                    .map((idx) => t('roundResult.eliminated', { name: findPlayerName(state.players, idx) }))
                    .join(' / ')}
                </div>
              )}
              {/* Preview the upcoming round: hand size drops by 1 (min 1) and the round winner chooses trump. */}
              {isRoundEnd && !isGameEnd && (
                <div data-testid="kw-next-round-preview" className="mt-1 text-ds-text-primary">
                  {t(nextHandSize === 1 ? 'roundResult.finalRoundPreview' : 'roundResult.nextRoundPreview', {
                    count: nextHandSize,
                    name: findPlayerName(state.players, state.roundWinnerIdx),
                  })}
                </div>
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

function FooterHand({ game, state, t, view }: KnockoutWhistContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanEliminated = state.players[humanIdx]?.eliminated ?? false;
  const isPlayPhase = state.phase === KnockoutWhistPhase.PLAY;
  const canPlay = isPlayPhase && state.isHumanTurn && !isHumanEliminated;
  const trumpDesign = suitName(state.trumpSuit);
  const trumpIndices = trumpDesign
    ? humanPlayer.cards.reduce<number[]>((indices, card, index) => {
        if (card.design === trumpDesign) indices.push(index);
        return indices;
      }, [])
    : [];

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={cardWidth}
      isMobile={isMobile}
      dataTutorialPrefix="knockoutwhist"
      trumpIndices={trumpIndices}
      trumpTitle={t('trumpCardDescription')}
      trumpAccessibleLabel={t('trumpCardDescription')}
      validIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
    />
  );
}

function FooterStatus({ state, t, frontendHint }: KnockoutWhistContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="knockoutwhist-hint-live" role="status" aria-live="polite">
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

function FooterActions({ game, state, t, busy }: KnockoutWhistContext) {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanEliminated = state.players[humanIdx]?.eliminated ?? false;
  const isPlayPhase = state.phase === KnockoutWhistPhase.PLAY;
  const canPlay = isPlayPhase && state.isHumanTurn && !isHumanEliminated;
  const isTrickEnd = state.phase === KnockoutWhistPhase.TRICK_END;
  const isRoundEnd = state.phase === KnockoutWhistPhase.ROUND_END;
  const isTrumpSelect = state.phase === KnockoutWhistPhase.TRUMP_SELECT;

  return (
    <>
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
      {isTrumpSelect && (
        <div className="flex flex-wrap items-center gap-2" data-testid="knockoutwhist-trump-select">
          <span className="text-ds-text-primary text-sm">{t('trumpSelectPrompt')}</span>
          {[1, 2, 3, 4].map((suit) => (
            <button
              key={suit}
              type="button"
              className={btnPrimary}
              onClick={() => game.handleSelectTrump(suit)}
              disabled={busy}
              data-testid={`knockoutwhist-trump-${suit}`}
              aria-label={t('trumpSelectSuit', { suit: suitSymbolAt(suit, '') })}
            >
              {suitSymbolAt(suit, '')}
            </button>
          ))}
        </div>
      )}
    </>
  );
}

const knockoutWhistDefinition: GameDefinition<
  KnockoutWhistResponse,
  KnockoutWhistGame,
  Parameters<typeof knockoutWhistApi.exec>,
  KnockoutWhistView
> = {
  key: 'knockoutwhist',
  useGame: useKnockoutWhistGame,
  useView: useKnockoutWhistView,
  tutorial: KNOCKOUT_WHIST_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 7 },
  cli: () => ({
    parseCommand: parseKnockoutWhistCommand,
    formatResponse: formatKnockoutWhistState,
    helpText: KNOCKOUT_WHIST_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === KnockoutWhistPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
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
          value: game.knockoutWhistConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'knockoutwhist-action-buttons',
  resetTutorialId: 'knockoutwhist-reset-button',
  AboveContent,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Knockout Whist game page: a British play-only survival trick-taker with elimination and Dogbone tokens. */
export const KnockoutWhistPage = createGamePage(knockoutWhistDefinition);
