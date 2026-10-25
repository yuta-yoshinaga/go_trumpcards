import type { TFunction } from 'i18next';
import { useEffect, useMemo, useState } from 'react';
import type { kingApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, useKingGame } from '../hooks/useKingGame';
import { badgeErrorColors, badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { KingPhaseValue, KingResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { KING_HELP, parseKingCommand } from '../utils/cli/commands/kingCommands';
import { formatKingState } from '../utils/cli/formatters/kingFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Selectable trump suits for the King (Trump) contract. */
const TRUMP_SUITS = [1, 2, 3, 4] as const;

/** The "King (Trump)" contract index, which requires a trump-suit choice. */
const KING_TRUMP_CONTRACT = 6;

/** King tutorial step definitions. */
const KING_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="king-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="king-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="king-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="king-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="king-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type KingGame = ReturnType<typeof useKingGame>;
type KingView = ReturnType<typeof useKingView>;
type KingContext = GamePageContext<KingResponse, KingGame, KingView>;

function useKingView({ game, t }: { game: KingGame; t: TFunction }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  // Contract to play when the dealer picks "King (Trump)" and must still pick a suit.
  const [pendingTrumpContract, setPendingTrumpContract] = useState<number | null>(null);
  const { cardWidth, isMobile } = useCardDimensions();

  const phaseNames: Record<KingPhaseValue, string> = useMemo(
    () => ({
      selectContract: t('phase.selectContract'),
      play: t('phase.play'),
      dealEnd: t('phase.dealEnd'),
      gameEnd: t('phase.gameEnd'),
    }),
    [t],
  );

  return {
    pendingTrumpContract,
    setPendingTrumpContract,
    cardWidth,
    isMobile,
    phaseNames,
  };
}

function Board({ state, t, game, view, busy }: KingContext) {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isSelectPhase = state.phase === 'selectContract';
  const isDealEnd = state.phase === 'dealEnd';
  const isGameEnd = state.phase === 'gameEnd' || state.gameEndFlag;

  const canSelect = isSelectPhase && state.dealerIdx === humanIdx;
  const contractName = state.currentContract >= 0 ? t(`contracts.${state.currentContract}`) : '-';
  const trumpSymbol = state.trumpSuit >= 1 ? suitSymbolAt(state.trumpSuit, '-') : '-';

  const handleContractClick = (contract: number) => {
    if (contract === KING_TRUMP_CONTRACT) {
      view.setPendingTrumpContract(contract);
      return;
    }
    view.setPendingTrumpContract(null);
    game.selectContract(contract, -1);
  };

  return (
    <>
      {canSelect && view.pendingTrumpContract === null && (
        <div className="grid grid-cols-3 sm:grid-cols-4 gap-1 mb-3" data-testid="king-contract-buttons">
          {state.usedContracts.map((used, contract) => {
            const isAchieve = contract === KING_TRUMP_CONTRACT;
            return (
              <button
                key={contract}
                type="button"
                className={`${btnSecondary} flex flex-col items-center justify-center gap-0.5 px-1.5 py-1 text-xs`}
                onClick={() => handleContractClick(contract)}
                disabled={busy || used}
                title={t(`contractDesc.${contract}`)}
                aria-label={`${t(`contracts.${contract}`)} — ${t(`contractType.${isAchieve ? 'achieve' : 'avoid'}`)}${used ? ` — ${t('contractUnavailable')}` : ''}`}
                aria-describedby={`king-contract-desc-${contract}`}
                data-testid={`king-contract-${contract}`}
              >
                <span className={used ? 'line-through opacity-60' : ''}>{t(`contracts.${contract}`)}</span>
                <span
                  className="flex items-center gap-1 text-xs"
                  aria-hidden="true"
                  data-testid={`king-contract-badge-${contract}`}
                >
                  <span className={`rounded px-1 py-0.5 ${isAchieve ? badgeSuccessColors : badgeErrorColors}`}>
                    {t(`contractType.${isAchieve ? 'achieve' : 'avoid'}`)}
                  </span>
                  <span>{t(`contractIcon.${contract}`)}</span>
                </span>
              </button>
            );
          })}
        </div>
      )}
      <div className="flex flex-wrap justify-center gap-x-4 gap-y-1 text-ds-text-primary mb-2">
        <span className="whitespace-nowrap">{t('deal', { n: state.dealNumber + 1, total: state.totalDeals })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        <span className="whitespace-nowrap">{t('contract', { name: contractName })}</span>
        <span className="whitespace-nowrap">{t('trump', { suit: trumpSymbol })}</span>
        <span className="whitespace-nowrap">
          {t('dealer', { name: playerName(state.dealerIdx, state.dealerIdx === humanIdx) })}
        </span>
      </div>

      {state.dealHistory.length > 0 && (
        <section className="my-3 p-2 rounded bg-ds-surface text-ds-text-muted text-sm" data-testid="king-deal-history">
          <h2 className="mb-1 text-ds-text-primary">{t('dealResult.historyTitle')}</h2>
          {state.dealHistory.map((detail, index) => (
            <div key={`${index}-${detail.contract}`} className="mb-2" data-testid={`king-history-deal-${index + 1}`}>
              <div className="text-ds-text-primary font-semibold">
                {detail.contract === KING_TRUMP_CONTRACT && detail.trumpSuit >= 1
                  ? t('dealResult.historyTrump', {
                      n: index + 1,
                      name: t(`contracts.${detail.contract}`),
                      suit: suitSymbolAt(detail.trumpSuit, '-'),
                    })
                  : t('dealResult.historyDeal', { n: index + 1, name: t(`contracts.${detail.contract}`) })}
              </div>
              {state.players.map((p) => (
                <div key={p.id}>
                  {t('dealResult.gained', {
                    name: playerName(p.id, p.isHuman),
                    points: detail.gained[p.id],
                  })}
                </div>
              ))}
            </div>
          ))}
        </section>
      )}

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={view.cardWidth}
            label={t('currentTrick')}
            dataTutorial="king-trick-display"
          />

          {/* **King にはトリック終了で一旦止まるフェーズが無い。**
              SelectContract → Play → DealEnd としか進まないので、CPU が
              続けて打つと直前のトリックは currentTrick が置き換わった瞬間に
              消える ── 見返す先が要る (#6487、Jass の previousTrick と同じ形)。 */}
          <details className="mt-2 p-2 rounded bg-black/30" data-testid="king-previous-trick">
            <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('previousTrick')}</summary>
            <div className="mt-1">
              {state.lastTrick.length > 0 ? (
                <TrickDisplay
                  currentTrick={state.lastTrick}
                  players={state.players}
                  cardWidth={Math.round(view.cardWidth * 0.7)}
                  // **中身があるなら獲得者もいる。**`lastTrick` が空でない
                  // 時点で `lastTrickWinner` は席番号なので、`>= 0` の分岐も
                  // `?.isHuman` の右辺もどのテストからも踏めない。名前は
                  // 範囲検査を内蔵した共有ヘルパに任せる。
                  label={t('previousTrickWinner', {
                    name: findPlayerName(state.players, state.lastTrickWinner),
                  })}
                  winnerIdx={state.lastTrickWinner}
                />
              ) : (
                <div className="text-ds-text-muted text-sm" data-testid="king-previous-trick-empty">
                  {t('previousTrickEmpty')}
                </div>
              )}
            </div>
          </details>
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="king-info">
          {/* Per-player match scores with a Dealer badge */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {state.players.map((p) => (
              <div key={p.id} className="py-0.5 flex items-center gap-2">
                <span className={p.id === state.dealerIdx ? 'text-ds-warning font-semibold' : ''}>
                  {playerName(p.id, p.isHuman)}: {t('score', { score: p.totalScore })}
                </span>
                {p.id === state.dealerIdx && (
                  <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('dealerBadge')}</span>
                )}
              </div>
            ))}
          </div>

          {isGameEnd && state.roundWinners.length > 0 && (
            <div
              className="mb-2 p-2 rounded bg-ds-surface text-ds-text-primary text-sm font-semibold"
              data-testid="king-final-result"
            >
              {state.roundWinners.length > 1
                ? t('finalResult.tie', {
                    names: state.roundWinners
                      .map((winner) => findPlayerName(state.players, winner))
                      .join(t('listSeparator')),
                  })
                : t('finalResult.winner', {
                    name: findPlayerName(state.players, state.roundWinners[0]),
                  })}
            </div>
          )}

          {/* Players: cards / tricks */}
          {view.isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('tricks', { count: p.trickCount })}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('tricks', { count: p.trickCount })}
                </div>
              ))}
            </div>
          )}

          {/* Deal result: contract played, its loss basis, and per-player gained points */}
          {(isDealEnd || isGameEnd) && state.lastDealDetail && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="king-deal-result">
              <div className="mb-1 text-ds-text-primary">{t('dealResult.title')}</div>
              {/* Contract breakdown: which contract was played and why points moved */}
              <div className="mb-1.5 pb-1.5 border-b border-white/10" data-testid="king-deal-breakdown">
                <div className="text-ds-text-primary font-semibold">
                  {state.lastDealDetail.contract === KING_TRUMP_CONTRACT && state.lastDealDetail.trumpSuit >= 1
                    ? t('dealResult.contractLineTrump', {
                        name: t(`contracts.${state.lastDealDetail.contract}`),
                        suit: suitSymbolAt(state.lastDealDetail.trumpSuit, '-'),
                      })
                    : t('dealResult.contractLine', {
                        name: t(`contracts.${state.lastDealDetail.contract}`),
                      })}
                </div>
                <div className="text-xs opacity-80">
                  {t('dealResult.basis', { desc: t(`contractDesc.${state.lastDealDetail.contract}`) })}
                </div>
              </div>
              {state.players.map((p) => (
                <div key={p.id} data-testid={`king-deal-breakdown-row-${p.id}`}>
                  {t('dealResult.gained', {
                    name: playerName(p.id, p.isHuman),
                    points: state.lastDealDetail?.gained[p.id] ?? 0,
                  })}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterLead({ state, t }: KingContext) {
  const isSelectPhase = state.phase === 'selectContract';
  if (!isSelectPhase) return null;

  return (
    <div className="mb-2 text-sm" data-testid="king-contract-descriptions">
      <p className="font-semibold">{t('contractDescriptionsTitle')}</p>
      <ul className="list-disc pl-5">
        {state.usedContracts.map((used, contract) => (
          <li key={contract} id={`king-contract-desc-${contract}`}>
            <span className="font-medium">{t(`contracts.${contract}`)}:</span> {used && `${t('contractUnavailable')}. `}
            {t(`contractDesc.${contract}`)}
          </li>
        ))}
      </ul>
    </div>
  );
}

function FooterStatus({ state, t, frontendHint }: KingContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="king-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {state.hint.contract !== undefined && state.hint.contract >= 0
              ? t('hintContract', {
                  contract: t(`contracts.${state.hint.contract}`),
                  reason: t(`hint.${state.hint.reason}`),
                })
              : `${t('hintAvailable')}: ${t(`hint.${state.hint.reason}`)}`}
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

function FooterHand({ state, game, t, view }: KingContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isSelectPhase = state.phase === 'selectContract';
  const canSelect = isSelectPhase && state.dealerIdx === humanIdx;
  const canPlay = state.phase === 'play' && state.isHumanTurn;

  return (
    <>
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955)。CalabresellaPage と同じ形 (#6880)。 */}
      <div data-testid="king-prompt-live" role="status" aria-live="polite">
        {isSelectPhase && !canSelect && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="king-select-cpu">
            {t('selectContractCpu', { id: state.dealerIdx })}
          </div>
        )}
        {canSelect && view.pendingTrumpContract === null && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="king-select-prompt">
            {t('selectContractPrompt')}
          </div>
        )}
        {canSelect && view.pendingTrumpContract !== null && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="king-trump-prompt">
            {t('selectTrumpPrompt')}
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
          dataTutorialPrefix="king"
          validIndices={canPlay ? state.playableIndices : undefined}
          restrictedTooltip={t('playButton')}
        />
      )}
    </>
  );
}

function primaryActions({ state, game, t, view, busy }: KingContext) {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isSelectPhase = state.phase === 'selectContract';
  const isPlayPhase = state.phase === 'play';
  const isDealEnd = state.phase === 'dealEnd';

  const canSelect = isSelectPhase && state.dealerIdx === humanIdx;
  const canPlay = isPlayPhase && state.isHumanTurn;

  if (!((canSelect && view.pendingTrumpContract !== null) || canPlay || isDealEnd)) {
    return null;
  }

  const handleTrumpClick = (suit: number) => {
    if (view.pendingTrumpContract === null) return;
    game.selectContract(view.pendingTrumpContract, suit);
    view.setPendingTrumpContract(null);
  };

  return (
    <div className="flex flex-wrap gap-2 items-center">
      {canSelect && view.pendingTrumpContract !== null && (
        <div className="flex flex-wrap gap-2" data-testid="king-trump-buttons">
          {TRUMP_SUITS.map((suit) => (
            <button
              key={suit}
              type="button"
              className={btnPrimary}
              onClick={() => handleTrumpClick(suit)}
              disabled={busy}
              aria-label={t(`trumpSuit.${suit}`)}
            >
              {suitSymbolAt(suit, '-')}
            </button>
          ))}
        </div>
      )}
      {canPlay && (
        <div data-tutorial="king-action-buttons">
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handlePlay}
            disabled={busy || game.selectedCardIndices.length !== 1}
          >
            {t('playButton')}
          </button>
        </div>
      )}
      {isDealEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextDeal} disabled={busy}>
          {t('nextDeal')}
        </button>
      )}
    </div>
  );
}

const kingDefinition: GameDefinition<KingResponse, KingGame, Parameters<typeof kingApi.exec>, KingView> = {
  key: 'king',
  useGame: useKingGame,
  useView: useKingView,
  tutorial: KING_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 13 },
  cli: () => ({
    parseCommand: parseKingCommand,
    formatResponse: formatKingState,
    helpText: KING_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === 'gameEnd' || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const humanWon = isGameEnd && state.roundWinners.includes(humanIdx);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: humanWon,
    };
  },
  shell: ({ state }) => ({
    isHumanTurn: state.isHumanTurn,
  }),
  reset: ({ game, view }) => {
    view.setPendingTrumpContract(null);
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.kingConfig.cpuDifficulty,
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
  primaryActions,
  resetTutorialId: 'king-reset-button',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  Board,
  FooterLead,
  FooterStatus,
  FooterHand,
};

/** Renders the King game page: a 4-player 52-card compendium trick-avoidance game. */
export const KingPage = createGamePage(kingDefinition);
