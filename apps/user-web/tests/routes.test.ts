import { readdirSync, statSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const PAGES = join(dirname(fileURLToPath(import.meta.url)), '..', 'pages')

/**
 * Regression test: a `pages/X.vue` file shadows the `pages/X/` directory in
 * Nuxt file routing — nested pages under it become unreachable child routes
 * without an outlet. The blog home + article pair must stay as siblings:
 * `pages/b/[handle]/index.vue` and `pages/b/[handle]/[slug].vue`.
 */
function shadowedPairs(dir: string, out: string[] = []): string[] {
  const entries = readdirSync(dir)
  for (const entry of entries) {
    const full = join(dir, entry)
    if (!statSync(full).isDirectory()) continue
    if (entries.includes(`${entry}.vue`)) {
      out.push(`${full} is shadowed by ${entry}.vue`)
    }
    shadowedPairs(full, out)
  }
  return out
}

describe('pages routing layout', () => {
  it('has no file-shadowed page directories', () => {
    expect(shadowedPairs(PAGES)).toEqual([])
  })

  it('keeps blog home and article as sibling routes', () => {
    const files = readdirSync(join(PAGES, 'b', '[handle]'))
    expect(files).toContain('index.vue')
    expect(files).toContain('[slug].vue')
  })
})
