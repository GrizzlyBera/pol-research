# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
npx tsc                      # Compile TypeScript
npx ts-node src/index.ts     # Run TypeScript directly
node src/index.js            # Run compiled JavaScript
```

## TypeScript Configuration Notes

Strict settings are enabled that require attention:
- `noUncheckedIndexedAccess`: Array/object index access may be undefined
- `exactOptionalPropertyTypes`: `prop?: T` differs from `prop: T | undefined`
- `verbatimModuleSyntax`: Use explicit `import type` for type-only imports
