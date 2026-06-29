#!/usr/bin/env python3
"""Рекурсивно собрать решения в один MD от корня-вершины.

Идёт по specue-импортам (specue.io/specue/*) из стартового узла,
для каждого достижимого узла кладёт его README (тело) и spec.cue (контракт).

Использование:
    python3 scripts/collect-decisions.py [start] [out.md]
    start  — узел-вершина (по умолчанию: binary)
    out.md — файл вывода (по умолчанию: DECISIONS.md в корне)
"""
import re, os, sys

ROOT = os.path.join(os.path.dirname(__file__), "..", "specue")
ROOT = os.path.normpath(ROOT)

def specue_imports(path):
    txt = open(path, encoding="utf-8").read()
    return re.findall(r'"specue\.io/specue/([^@"]+)@', txt)

def build_index():
    idx = {}
    for cur, _, files in os.walk(ROOT):
        if "spec.cue" in files and f"{os.sep}pkg{os.sep}" not in cur + os.sep:
            rel = os.path.relpath(cur, ROOT)
            idx[rel] = cur
    return idx

def reachable(start, idx):
    seen, stack = set(), [start]
    while stack:
        cur = stack.pop()
        if cur in seen or cur not in idx:
            continue
        seen.add(cur)
        for imp in specue_imports(os.path.join(idx[cur], "spec.cue")):
            if imp not in seen:
                stack.append(imp)
    return seen

def read(path):
    return open(path, encoding="utf-8").read().strip() if os.path.exists(path) else ""

def main():
    start = sys.argv[1] if len(sys.argv) > 1 else "binary"
    out = sys.argv[2] if len(sys.argv) > 2 else os.path.join(ROOT, "..", "DECISIONS.md")
    idx = build_index()
    nodes = sorted(reachable(start, idx))
    parts = [f"# Решения, достижимые из `{start}` ({len(nodes)} узлов)\n"]
    for n in nodes:
        d = idx[n]
        body = read(os.path.join(d, "README.md"))
        spec = read(os.path.join(d, "spec.cue"))
        parts.append(f"\n---\n\n## {n}\n")
        if body:
            parts.append(f"\n### Тело (README)\n\n{body}\n")
        parts.append(f"\n### Контракт (spec.cue)\n\n```cue\n{spec}\n```\n")
    open(out, "w", encoding="utf-8").write("\n".join(parts))
    print(f"Собрано {len(nodes)} узлов из '{start}' → {os.path.normpath(out)}")

if __name__ == "__main__":
    main()
