#!/usr/bin/env python3
"""Map game code surfaces, including the full transitive relative-import closure under components and hooks."""
import argparse
import json
import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
ROUTES = ROOT / "frontend/src/constants/gameRoutes.ts"


def pascal(slug):
    return "".join(part[:1].upper() + part[1:] for part in re.split(r"[-_]", slug))


def resolve_import(base, spec):
    if not spec.startswith("."):
        return None
    stem = (base.parent / spec).resolve()
    for p in (stem, stem.with_suffix(".tsx"), stem.with_suffix(".ts"), stem / "index.tsx", stem / "index.ts"):
        if p.is_file():
            try:
                return p.relative_to(ROOT).as_posix()
            except ValueError:
                return None
    return None


def imports(path):
    text = (ROOT / path).read_text(errors="replace")
    return [m.group(1) for m in re.finditer(r"(?:from\s*|import\s*)[\"'](\.[^\"']+)[\"']", text)]


def surface_kind(path):
    return "/components/" in path or "/hooks/" in path


def is_test_file(path):
    return bool(re.search(r"(?:\.test|\.spec)\.(?:ts|tsx)$", Path(path).name))


def registry_games():
    source = (ROOT / "internal/infrastructure/games/registry.go").read_text()
    return re.findall(r'\{Name: "([^"]+)", Category:', source)


def page_for(slug):
    pages = list((ROOT / "frontend/src/pages").glob("*Page.tsx"))
    route_source = ROUTES.read_text(errors="replace")
    route_pages = {
        label.removeprefix("nav."): name
        for label, name in re.findall(r"labelKey:\s*'([^']+)'[\s\S]*?page:\s*'([^']+)'", route_source)
    }
    route_name = route_pages.get(slug)
    if route_name:
        exact = ROOT / "frontend/src/pages" / f"{route_name}Page.tsx"
        if exact.is_file():
            return exact.relative_to(ROOT).as_posix()
    expected = pascal(slug).lower() + "page.tsx"
    match = next((p for p in pages if p.name.lower() == expected), None)
    if match is None:
        match = next((p for p in pages if p.stem[:-4].lower() == slug.lower()), None)
    return match.relative_to(ROOT).as_posix() if match else None


def build_one(slug):
    ja_path = ROOT / "frontend/src/i18n/locales/ja/common.json"
    try:
        ja = json.loads(ja_path.read_text()).get("nav", {}).get(slug)
    except (OSError, json.JSONDecodeError):
        ja = None
    page = page_for(slug)
    components, hooks = set(), set()
    if page:
        for spec in imports(page):
            p = resolve_import(ROOT / page, spec)
            if p and "/components/" in p:
                components.add(p)
            if p and "/hooks/" in p:
                hooks.add(p)
    deep = set()
    pending = []
    if page:
        pending.extend(q for spec in imports(page) if (q := resolve_import(ROOT / page, spec)) and surface_kind(q))
    direct = components | hooks
    visited = set()
    while pending:
        current = pending.pop()
        if current in visited:
            continue
        visited.add(current)
        for spec in imports(current):
            q = resolve_import(ROOT / current, spec)
            if q and surface_kind(q) and not is_test_file(q):
                if q not in direct:
                    deep.add(q)
                pending.append(q)
    presenter = ROOT / "internal/adapter/presenter"
    page_name = Path(page).stem[:-4] if page else pascal(slug)
    game_name = page_name
    # The resolved page's Pascal name is the primary backend convention.
    interactor = ROOT / "internal/usecase" / f"{game_name}Interactor.go"
    web = presenter / f"{game_name}WebPresenter.go"
    if not web.is_file(): web = presenter / f"{game_name}Presenter.go"
    cui = presenter / f"{game_name}CuiPresenter.go"
    if not interactor.is_file(): interactor = ROOT / "internal/usecase" / f"{pascal(slug)}Interactor.go"
    if not web.is_file():
        fallback = presenter / f"{pascal(slug)}WebPresenter.go"
        web = fallback if fallback.is_file() else presenter / f"{pascal(slug)}Presenter.go"
    if not cui.is_file(): cui = presenter / f"{pascal(slug)}CuiPresenter.go"
    locales = []
    for d in ("frontend/src/i18n/locales/ja", "frontend/src/i18n/locales/en", "internal/i18n/locales/ja", "internal/i18n/locales/en"):
        p = ROOT / d / f"{slug}.json"
        if p.is_file(): locales.append(p.relative_to(ROOT).as_posix())
    return {"game": slug, "ja": ja, "page": page,
            "components": sorted(components), "hooks": sorted(hooks), "deep": sorted(deep),
            "web_presenter": web.relative_to(ROOT).as_posix() if web.is_file() else None,
            "cui_presenter": cui.relative_to(ROOT).as_posix() if cui.is_file() else None,
            "interactor": interactor.relative_to(ROOT).as_posix() if interactor.is_file() else None,
            "locales": locales,
            "sync_cpu_loop": bool(interactor.is_file() and re.search(r"runCpuTurns|runCpuTurnsLoop|advanceCpu", interactor.read_text(errors="replace")))}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--game", action="append")
    ap.add_argument("--out")
    args = ap.parse_args()
    games = args.game or registry_games()
    result = [build_one(g) for g in games]
    content = json.dumps(result, ensure_ascii=False, indent=2) + "\n"
    if args.out: Path(args.out).write_text(content)
    else: print(content, end="")
    if not args.game:
        missing = sum(x["page"] is None for x in result)
        cpu = sum(x["sync_cpu_loop"] for x in result)
        counts = sorted(len(set(x["components"] + x["hooks"] + x["locales"] + [p for p in (x["page"], x["web_presenter"], x["cui_presenter"], x["interactor"]) if p])) for x in result)
        deep_counts = sorted(len(x["deep"]) for x in result)
        median_deep = (deep_counts[(len(deep_counts)-1)//2]+deep_counts[len(deep_counts)//2])/2
        print(f"games={len(result)} without_page={missing} sync_cpu_loop={cpu} median_surface_files={(counts[(len(counts)-1)//2]+counts[len(counts)//2])/2} median_deep={median_deep} max_deep={max(deep_counts, default=0)}", file=__import__("sys").stderr)

if __name__ == "__main__": main()
