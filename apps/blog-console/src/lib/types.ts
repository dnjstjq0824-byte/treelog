/** Console DTOs mirroring the Go backend JSON contracts. */

export interface SessionUser {
  id: string
  email: string
  username: string
  display_name: string
  bio: string
  avatar_url: string | null
  preferred_locale: string
  timezone: string
  mfa_enabled: boolean
  roles: string[]
  created_at: string
}

export interface Blog {
  id: string
  owner_user_id: string
  handle: string
  blog_type: string
  title: string
  subtitle: string
  description: string
  default_language: string
  timezone: string
  primary_url_type: string
  status: string
  subscriber_count: number
  neighbor_count: number
  post_count: number
  created_at: string
  updated_at: string
}

export interface PostSummary {
  id: string
  blog_id: string
  author_user_id: string
  author_name: string
  slug: string
  title: string
  subtitle: string
  excerpt: string
  cover_image_url?: string | null
  category_name?: string | null
  status: string
  visibility: string
  word_count: number
  reading_time_sec: number
  view_count: number
  reaction_count: number
  comment_count: number
  tags: string[]
  published_at?: string | null
  created_at: string
}

export interface PostDetail extends PostSummary {
  author_handle: string
  author_avatar_url?: string | null
  category_id?: string | null
  series_id?: string | null
  series_title?: string | null
  series_order?: number | null
  content_json: unknown
  content_html: string
  content_plain: string
  cover_media_id?: string | null
  allow_comments: string
  allow_reactions: boolean
  is_notice: boolean
  is_pinned: boolean
  seo_title?: string | null
  seo_description?: string | null
  canonical_url?: string | null
  license_type: string
  ai_disclosure_level: string
  paywall_price_minor?: number | null
  bookmark_count: number
  revision_no: number
  is_paywalled: boolean
  updated_at: string
}

export interface DraftCreated {
  id: string
  blog_id: string
  title: string
  subtitle: string
  status: string
  revision_no: number
  slug: string
}

export interface AutosaveResult {
  saved: boolean
  post_id: string
  server_saved_at: string
}

export interface AutosaveConflict {
  server_revision_no: number
  server_title: string
  server_subtitle: string
  server_content: unknown
}

export interface RevisionSummary {
  id: string
  post_id: string
  revision_no: number
  editor_user_id: string
  title: string
  subtitle: string
  change_summary: string
  is_published_snapshot: boolean
  created_at: string
}

export interface RevisionDetail extends RevisionSummary {
  content_json: unknown
  content_html: string
}

export interface PublishInput {
  title: string
  subtitle?: string
  slug?: string
  category_id?: string | null
  series_id?: string | null
  series_order?: number | null
  content_json: unknown
  content_html: string
  content_plain: string
  excerpt?: string
  cover_media_id?: string | null
  tags?: string[]
  visibility: string
  allow_comments: string
  allow_reactions?: boolean
  is_notice: boolean
  is_pinned: boolean
  seo_title?: string
  seo_description?: string
  canonical_url?: string
  license_type?: string
  ai_disclosure_level?: string
  paywall_price_minor?: number | null
  scheduled_for?: string | null
}

export interface MediaVariant {
  variant_name: string
  format: string
  width: number
  height: number
  size_bytes: number
  public_cdn_url: string
}

export interface MediaAsset {
  id: string
  blog_id: string
  uploader_user_id: string
  upload_session_id?: string | null
  media_type: string
  mime_type: string
  original_size_bytes: number
  sha256_hash: string
  width?: number | null
  height?: number | null
  blurhash?: string | null
  alt_text: string
  caption: string
  embed_provider?: string | null
  embed_external_id?: string | null
  embed_html_sanitized?: string | null
  exif_stripped: boolean
  status: string
  variants?: MediaVariant[]
  created_at: string
}

export interface UploadSession {
  session_id: string
  presigned_url: string
  storage_bucket: string
  storage_object_key: string
  expires_at: string
  deduplicated: boolean
  existing_asset?: MediaAsset | null
}

export interface BlogCategory {
  id: string
  blog_id: string
  parent_id?: string | null
  name: string
  slug: string
  description: string
  topic_code?: string | null
  list_style: string
  visibility: string
  sort_order: number
  post_count: number
  children?: BlogCategory[]
  created_at: string
  updated_at: string
}

export interface Series {
  id: string
  blog_id: string
  title: string
  slug: string
  description: string
  cover_media_id?: string | null
  is_completed: boolean
  post_count: number
  created_at: string
  updated_at: string
}

export interface GuestbookEntry {
  id: string
  blog_id: string
  author_user_id: string
  parent_entry_id?: string | null
  content_text: string
  is_secret: boolean
  status: string
  created_at: string
  author_username?: string
  author_name?: string
}

export interface BlogMember {
  id: string
  blog_id: string
  user_id: string
  role: string
  invite_status: string
  joined_at: string
  username?: string
  display_name?: string
}

export interface BlogSettings {
  blog_id: string
  default_post_visibility: string
  default_allow_comments: string
  allow_guestbook: boolean
  allow_mutual_requests: boolean
  allow_rss_full_content: boolean
  allow_search_indexing: boolean
  allow_treescan_sync: boolean
  anti_copy_protection: boolean
  require_approval_for_authors: boolean
  posts_per_page: number
  show_visitor_counter: boolean
  adult_content_warning: boolean
  updated_at: string
}

export interface ThemeCatalogItem {
  id: string
  slug: string
  name: string
  description: string
  layout_type: string
  default_tokens_json: unknown
  preview_image_url?: string | null
  version: string
  is_active: boolean
  created_at: string
}

export interface BlogThemeConfig {
  blog_id: string
  theme_id: string
  theme_slug: string
  theme_name: string
  active_layout: string
  color_palette_tokens: unknown
  typography_tokens: unknown
  custom_css: string
  dark_mode_preference: string
  updated_at: string
}

export interface Domain {
  id: string
  blog_id: string
  domain_name: string
  domain_type: string
  is_primary: boolean
  verification_token: string
  verification_status: string
  tls_status: string
  tls_expires_at?: string | null
  redirect_to_primary: boolean
  last_verified_at?: string | null
  failure_reason?: string | null
  created_at: string
  updated_at: string
}

export interface MembershipTier {
  id: string
  blog_id: string
  name: string
  description: string
  monthly_price_minor: number
  annual_price_minor?: number | null
  benefits_json: string
  is_active: boolean
  created_at: string
}

export interface Payout {
  id: string
  creator_user_id: string
  blog_id: string
  period_start: string
  period_end: string
  gross_amount_minor: number
  withholding_tax_minor: number
  net_amount_minor: number
  bank_code: string
  account_holder: string
  status: string
  requested_at: string
}

export const VISIBILITIES = [
  { id: 'PUBLIC', label: '전체 공개' },
  { id: 'UNLISTED', label: '링크 공개' },
  { id: 'NEIGHBOR_ONLY', label: '이웃 공개' },
  { id: 'MUTUAL_ONLY', label: '맞이웃 공개' },
  { id: 'PAID_MEMBERS_ONLY', label: '유료 멤버십' },
  { id: 'PRIVATE', label: '비공개' }
] as const

export const COMMENT_POLICIES = [
  { id: 'ALL_ALLOWED', label: '모두 허용' },
  { id: 'NEIGHBORS_ONLY', label: '이웃만' },
  { id: 'MUTUAL_ONLY', label: '맞이웃만' },
  { id: 'REQUIRE_APPROVAL', label: '승인 후 공개' },
  { id: 'DISABLED', label: '비활성화' }
] as const

export const LICENSES = [
  { id: 'ALL_RIGHTS_RESERVED', label: '저작권 보유' },
  { id: 'CC_BY', label: 'CC BY' },
  { id: 'CC_BY_NC', label: 'CC BY-NC' },
  { id: 'CC_BY_ND', label: 'CC BY-ND' },
  { id: 'CC_BY_SA', label: 'CC BY-SA' },
  { id: 'CC_BY_NC_SA', label: 'CC BY-NC-SA' },
  { id: 'CC_BY_NC_ND', label: 'CC BY-NC-ND' }
] as const

export const AI_LEVELS = [
  { id: 'NONE', label: 'AI 미사용' },
  { id: 'AI_ASSISTED', label: 'AI 보조' },
  { id: 'AI_GENERATED', label: 'AI 생성' }
] as const

export const POST_STATUSES = [
  { id: '', label: '전체' },
  { id: 'DRAFT', label: '초안' },
  { id: 'IN_REVIEW', label: '검토 중' },
  { id: 'SCHEDULED', label: '예약됨' },
  { id: 'PUBLISHED', label: '발행됨' },
  { id: 'UNPUBLISHED', label: '발행 취소' }
] as const
