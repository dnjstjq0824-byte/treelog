/**
 * Auth bootstrap plugin.
 *
 * Runs before route middleware on both server and client, so the session is
 * resolved from the `tb_access` cookie before `auth`/`guest` middleware read
 * `isLoggedIn`. Calling `initialize()` from `onMounted` instead would run
 * after middleware and bounce deep links / refreshes on authed pages.
 */
export default defineNuxtPlugin(async () => {
  await useAuth().initialize()
})
