/** Admin console DTOs mirroring the Go backend JSON contracts. */

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

export interface PlatformStats {
  users_total: number
  users_active: number
  users_suspended: number
  blogs_total: number
  blogs_suspended: number
  posts_published: number
  posts_today: number
  comments_total: number
  media_ready: number
  cases_open: number
  appeals_pending: number
  orders_paid_today: number
  revenue_today_minor: number
}

export interface AdminUser {
  id: string
  email: string
  status: string
  username: string
  display_name: string
  roles: string[]
  trust_tier: number
  last_login_at?: string | null
  created_at: string
}

export interface AdminBlog {
  id: string
  handle: string
  title: string
  status: string
  owner_user_id: string
  post_count: number
  created_at: string
}

export interface ModerationCase {
  id: string
  target_type: string
  target_id: string
  reason_category: string
  report_count: number
  priority_score: number
  status: string
  created_at: string
}

export interface ReportRow {
  id: string
  reporter_user_id?: string | null
  reason_category: string
  description: string
  created_at: string
}

export interface ActionRow {
  id: string
  actor_moderator_id: string
  action_type: string
  policy_clause_code: string
  public_notice_reason: string
  created_at: string
}

export interface CaseDetail {
  case: ModerationCase
  reports: ReportRow[]
  actions: ActionRow[]
}

export interface Appeal {
  id: string
  case_id: string
  appellant_user_id: string
  appeal_statement: string
  status: string
  reviewer_decision_note?: string | null
  resolved_at?: string | null
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

export interface PayoutAccount {
  bank_code: string
  bank_account: string
  account_holder: string
}

export const ACTION_TYPES = [
  { id: 'DISMISS', label: '기각' },
  { id: 'WARN_USER', label: '경고' },
  { id: 'QUARANTINE_CONTENT', label: '콘텐츠 격리' },
  { id: 'RESTORE_CONTENT', label: '콘텐츠 복원' },
  { id: 'DELETE_CONTENT', label: '콘텐츠 삭제' },
  { id: 'SUSPEND_BLOG_7D', label: '블로그 7일 정지' },
  { id: 'SUSPEND_BLOG_PERM', label: '블로그 영구 정지' },
  { id: 'SUSPEND_USER', label: '사용자 정지' }
] as const
