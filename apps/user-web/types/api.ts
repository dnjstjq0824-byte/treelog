/** API types mirroring the Go backend JSON contracts exactly. */

export interface ApiEnvelope<T> {
  data: T
  request_id: string
}

export interface ApiFaultBody {
  code: string
  message: string
  request_id: string
}

export interface CursorEnvelope<T> extends ApiEnvelope<T> {
  next_cursor: string
}

export interface Tokens {
  access_token: string
  expires_in: number
}

export interface RegisterResponse {
  user_id: string
  status: string
  email: string
}

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

export interface SessionInfo {
  id: string
  created_at: string
  last_active_at: string
  expires_at: string
  is_current: boolean
}

export interface MfaSetup {
  secret: string
  backup_codes: string[]
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

export interface CheckHandle {
  handle: string
  available: boolean
  message: string
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
  /** Client-side enrichment: known from the surrounding page, not the API. */
  blog_handle?: string
  blog_title?: string
}

export interface PostDetail {
  id: string
  blog_id: string
  author_user_id: string
  author_name: string
  author_handle: string
  author_avatar_url?: string | null
  category_id?: string | null
  category_name?: string | null
  series_id?: string | null
  series_title?: string | null
  series_order?: number | null
  slug: string
  title: string
  subtitle: string
  content_json: unknown
  content_html: string
  content_plain: string
  excerpt: string
  cover_media_id?: string | null
  cover_image_url?: string | null
  status: string
  visibility: string
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
  word_count: number
  reading_time_sec: number
  view_count: number
  reaction_count: number
  comment_count: number
  bookmark_count: number
  revision_no: number
  tags: string[]
  is_paywalled: boolean
  published_at?: string | null
  created_at: string
  updated_at: string
}

export interface SearchPostItem {
  id: string
  blog_id: string
  blog_title: string
  blog_handle: string
  author_name: string
  slug: string
  title: string
  subtitle: string
  excerpt: string
  cover_image_url?: string | null
  category_name?: string | null
  tags: string[]
  word_count: number
  reading_time_sec: number
  view_count: number
  reaction_count: number
  comment_count: number
  published_at?: string | null
  score?: number
}

export interface SearchBlogItem {
  id: string
  handle: string
  title: string
  description: string
  post_count: number
  avatar_url?: string
}

export interface SearchTagItem {
  id: string
  name: string
  slug: string
  post_count: number
}

export interface SearchPostsData {
  query: string
  items: SearchPostItem[]
  total: number
  limit: number
  offset: number
}

export interface CommentItem {
  id: string
  post_id: string
  author_user_id: string
  author_name: string
  parent_comment_id?: string | null
  depth: number
  content_text: string
  is_secret: boolean
  is_pinned: boolean
  like_count: number
  created_at: string
}

export interface ReactionSummary {
  counts: Record<string, number>
  total: number
  mine?: string | null
}

export interface NotificationItem {
  id: string
  category: string
  title: string
  body: string
  action_url: string
  is_read: boolean
  created_at: string
}

export interface BookmarkItem {
  post_id: string
  title: string
  blog_handle: string
  slug: string
  folder_name: string
  read_progress_pct: number
  created_at: string
}

export interface FollowedBlog {
  blog_id: string
  handle: string
  title: string
  relation_type: string
}

export interface SubscriptionItem {
  blog_id: string
  handle: string
  title: string
  tier: string
  notify_in_app: boolean
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

export interface Order {
  id: string
  buyer_user_id: string
  order_type: string
  recipient_blog_id?: string | null
  target_ref_id?: string | null
  gross_amount_minor: number
  platform_fee_minor: number
  net_amount_minor: number
  currency: string
  status: string
  access_expires_at?: string | null
  created_at: string
}

export interface CreateOrderInput {
  order_type: 'MEMBERSHIP_SUB' | 'SINGLE_POST_UNLOCK' | 'READER_DONATION' | 'AD_PREPAID_TOPUP'
  recipient_blog_id?: string
  target_ref_id?: string
  gross_amount_minor: number
  currency: 'KRW'
  idempotency_key: string
  donation_message?: string
}

export interface BillingConfig {
  toss_client_key: string
  payments_enabled: boolean
}
