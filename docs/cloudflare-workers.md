# Cloudflare Workers (WASM)

Games are deployed to Cloudflare Workers as WASM binaries via TinyGo. **Six** workers split games into size buckets to keep each binary under the free-tier 1 MB gzipped limit. The fourth, **`extra`**, was added in [ADR-0032](adr/0032-fourth-worker-capacity.md) once the original three approached the limit, and **`extra2`/`extra3`** in [ADR-0036](adr/0036-fifth-sixth-worker-capacity.md). Like the others, the `Category` is purely a binary-size bucket, **not** a user-facing taxonomy. **No bucket is *the* overflow one:** a new game goes into whichever worker currently has the most gzip headroom, measured rather than assumed.

Each row lists the games' **registry keys** verbatim, so the table can be checked
mechanically -- `TestDocsMatchRegistry` in `internal/infrastructure/games` fails if
this table and `registry.go` disagree. It is generated, not curated: the previous
hand-written prose drifted repeatedly (it still named three workers after the fourth
shipped), which is why the lists are now flat and guarded.

A bucket is chosen purely by binary size. Nothing about a game's genre puts it in a
particular worker, and games move between buckets whenever one approaches the limit.

| Worker | Entry point | Games | Registry keys |
|--------|-------------|-------|---------------|
| **casino** | `cmd/workers/casino/main.go` | 55 | `baccarat`, `badugi`, `bigo`, `bigohilo`, `blackjack`, `blackjackswitch`, `bourre`, `caribbeanstud`, `casinoholdem`, `casinowar`, `chinesepoker`, `courtpiece`, `crazypineapple`, `deuceswild`, `deucetoseven`, `doppelkopf`, `dragontiger`, `ecarte`, `fivecardstud`, `fortyfives`, `fourcardpoker`, `highcardflush`, `holdem`, `indianpoker`, `irishpoker`, `jokerpoker`, `letitride`, `mississippistud`, `mus`, `napoleon`, `oasispoker`, `omaha`, `omahahilo`, `openfacechinese`, `paigow`, `pineapple`, `poker`, `razz`, `reddog`, `russianpoker`, `sevencardstud`, `sevencardstudhilo`, `shortdeck`, `spanish21`, `sueca`, `tarneeb`, `teenpatti`, `texasholdembonus`, `threecard`, `threecardbrag`, `tressette`, `tute`, `twentynine`, `ultimatetexasholdem`, `videopoker` |
| **classic** | `cmd/workers/classic/main.go` | 41 | `allfours`, `bezique`, `briscola`, `callbreak`, `cassino`, `catchten`, `crazyeights`, `daifugo`, `doudizhu`, `durak`, `egyptianratscrew`, `escoba`, `hearts`, `klaverjas`, `knockoutwhist`, `labellelucie`, `manille`, `marias`, `nap`, `ninetynine`, `ohhell`, `oldmaid`, `pageone`, `pitch`, `preference`, `president`, `prsi`, `scopa`, `scopone`, `sedma`, `sevens`, `shithead`, `simplesimon`, `slapjack`, `solowhist`, `spades`, `spoilfive`, `tonk`, `truco`, `twotenjack`, `whist` |
| **solo** | `cmd/workers/solo/main.go` | 45 | `accordion`, `acesup`, `bakersdozen`, `bakersgame`, `barbu`, `beleagueredcastle`, `bidwhist`, `blackhole`, `bristol`, `calculation`, `canfield`, `clocksolitaire`, `crescent`, `cruel`, `easthaven`, `eightoff`, `euchre`, `fivehundred`, `fortythieves`, `freecell`, `gaps`, `golf`, `gongzhu`, `klondike`, `macau`, `memory`, `montecarlo`, `osmosis`, `penguin`, `pokersquares`, `pyramid`, `russianbank`, `russiansolitaire`, `schnapsen`, `scorpion`, `seahaventowers`, `spider`, `spiderette`, `thirtyone`, `tienlen`, `tripeaks`, `wasp`, `yaniv`, `yukon`, `zheng` |
| **extra** | `cmd/workers/extra/main.go` | 34 | `agnes`, `anaconda`, `burraco`, `calabresella`, `canasta`, `carioca`, `chinchon`, `cinch`, `conquian`, `contractrummy`, `flowergarden`, `frenchtarot`, `gaigel`, `ginrummy`, `gostop`, `guts`, `hachihachi`, `handandfoot`, `indianrummy`, `kalooki`, `king`, `kingalbert`, `koenigrufen`, `machiavelli`, `oichokabu`, `pan`, `rummy500`, `samba`, `streetsandalleys`, `sultan`, `threethirteen`, `trenteetquarante`, `tysiac`, `watten` |
| **extra2** | `cmd/workers/extra2/main.go` | 39 | `americantoad`, `beggarmyneighbour`, `bigtwo`, `bisley`, `braid`, `chineseten`, `cuarenta`, `cuckoo`, `doubleklondike`, `doubt`, `duchess`, `faro`, `fiftyone`, `gofish`, `grandfathersclock`, `kemps`, `laughandliedown`, `loba`, `mighty`, `missmilligan`, `mushi`, `napoleonssquare`, `nertz`, `pigtail`, `pinochle`, `pishti`, `pontoon`, `settemezzo`, `sirtommy`, `sjavs`, `sixcardgolf`, `speed`, `spiteandmalice`, `spoons`, `tichu`, `trash`, `war`, `windmill`, `zwicker` |
| **extra3** | `cmd/workers/extra3/main.go` | 31 | `basra`, `belote`, `bouillotte`, `bridge`, `bura`, `cego`, `congress`, `cribbage`, `desmoche`, `fortyandeight`, `jass`, `koikoi`, `loo`, `mao`, `michigan`, `niuniu`, `ombre`, `piquet`, `primero`, `rook`, `scarto`, `sevenbridge`, `sheepshead`, `skat`, `skitgubbe`, `tablanet`, `terrace`, `toepen`, `trex`, `ulti`, `wizard` |

The worker entry points (`cmd/workers/{casino,classic,solo,extra,extra2,extra3}/main.go`) are thin shells that blank-import the matching `internal/infrastructure/games/<category>` sub-package and call `games.RegisterCategory(mux, games.Category…)`. The registry itself (`internal/infrastructure/games/registry.go`) stores `{Name, Category}` for each game; the human-readable descriptions live in a separate `descriptions.go` map (build-tagged `//go:build !js || !wasm` to keep them out of the WASM binaries). The Web-server factories live in `games_server.go` (excluded from WASM via build tags) and the Worker bindings live in per-category sub-packages — this split is what keeps each Cloudflare Worker binary under the 1 MB gzipped free-tier limit by letting TinyGo dead-code-eliminate the games from the other categories.

**When adding/modifying a game, always update:**
1. `internal/infrastructure/games/registry.go` — `{Name, Category}` entry (selects the worker), plus the matching CLI display title in `descriptions.go`
2. `internal/infrastructure/games/games_server.go` — `BindWebControllerFor("<name>", …)` for the HTTP server factory
3. `internal/infrastructure/games/{casino,classic,solo,extra,extra2,extra3}/<category>.go` — `games.RegisterKVGame("<name>", games.Category…, …)` for the KV-backed worker route (must match the `Category`)
4. `frontend/src/api/gameApi.ts` `workerUrl` — must match the `Category`

Build: `make build-worker-{solo,casino,classic,extra,extra2,extra3}` or `make build-workers` (requires TinyGo).

## ローカルでサイズを実測する

ADR-0032 の時点では TinyGo をローカルに持たず CI のレポート頼りだったが、いまは手元で測れる
（Go 1.25.8 + TinyGo 0.40.1、CI と同じ組み合わせ）。

```sh
export PATH="$HOME/sdk/go1.25.8/bin:$HOME/.local/opt/tinygo/bin:$PATH"
export GOTOOLCHAIN=local            # ← 必須
mkdir -p workers/<w>/build
go run github.com/syumai/workers/cmd/workers-assets-gen -mode=tinygo -o workers/<w>/build
tinygo build -tags <w> -o workers/<w>/build/app.wasm -target wasm \
  -stack-size=128KB -no-debug -opt=z ./cmd/workers/<w>
wasm-opt --enable-bulk-memory --enable-nontrapping-float-to-int --enable-sign-ext \
  -Oz workers/<w>/build/app.wasm -o workers/<w>/build/app.wasm
gzip -c workers/<w>/build/app.wasm | wc -c    # 1,048,576 と比較する
```

- **`GOTOOLCHAIN=local` を忘れないこと。** `go.mod` の `toolchain go1.26.0` により Go 1.25 でも
  1.26 に自動アップグレードされ、TinyGo 0.40.1 が
  `requires go version 1.19 through 1.25, got go1.26` で止まる。CI も同じ理由で明示している。
- **`wasm-opt` 前の値で判断しない。** extra は 1,077,248 → 1,029,817 と 47 KB 縮む。最適化前だと
  上限超過に見える。
- `make` が無い環境では上のコマンドが Makefile レシピの展開そのもの。
