// Redirects legacy /@handle URLs (spec-era public address) to the canonical
// /b/handle routes with a permanent redirect, preserving any sub-path.
export default defineEventHandler((event) => {
  const pathname = getRequestURL(event).pathname
  const match = pathname.match(/^\/@([^/]+)(\/.*)?$/)
  if (!match) {
    return
  }
  const rest = match[2] ?? ''
  return sendRedirect(event, `/b/${match[1]}${rest}`, 301)
})
