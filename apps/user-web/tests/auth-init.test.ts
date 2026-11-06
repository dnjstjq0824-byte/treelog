import { readdirSync, readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..')

/**
 * Regression test: the session must resolve before route middleware runs.
 * `auth`/`guest` middleware read `isLoggedIn` synchronously, while `useAuth`
 * starts with `user = null`. If `initialize()` only ran from `onMounted`
 * (after middleware), every fresh load of an authed page — deep link,
 * refresh, SSR — would bounce to /login despite a valid cookie. A Nuxt
 * plugin runs before middleware on both server and client, so the
 * bootstrap call must live in a plugin, not in app.vue's mount hook.
 */
describe('auth bootstrap ordering', () => {
  it('initializes the session from a Nuxt plugin', () => {
    const plugins = readdirSync(join(ROOT, 'plugins')).filter((f) => f.endsWith('.ts'))
    const bootstrap = plugins.filter((f) => {
      const body = readFileSync(join(ROOT, 'plugins', f), 'utf8')
      return body.includes('defineNuxtPlugin') && body.includes('.initialize()')
    })
    expect(bootstrap).not.toEqual([])
  })

  it('does not defer session bootstrap to app mount', () => {
    const body = readFileSync(join(ROOT, 'app.vue'), 'utf8')
    expect(body).not.toMatch(/onMounted\(\(\) => \{\s*initialize\(\)/)
    expect(body).not.toContain('initialize()')
  })
})
