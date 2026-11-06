import { loadTossPayments } from '@tosspayments/tosspayments-sdk'
import type { BillingConfig, CreateOrderInput, MembershipTier, Order } from '~/types/api'
import { ApiError, api } from '~/utils/api'

/**
 * Toss checkout flow: create a PENDING order on our API, then open the Toss
 * payment window. Toss redirects back to /billing/success (or /billing/fail),
 * which confirms the payment server-side.
 */
export function useCheckout() {
  const toast = useToast()
  const busy = ref(false)

  async function start(input: Omit<CreateOrderInput, 'currency' | 'idempotency_key'>, orderName: string, returnPath: string): Promise<void> {
    busy.value = true
    try {
      const config = await api<BillingConfig>('/api/v1/billing/config', { auth: false })
      if (!config.payments_enabled || !config.toss_client_key) {
        toast.push('결제 기능이 준비되지 않았습니다', 'error')
        return
      }
      const { user } = useAuth()
      const order = await api<Order>('/api/v1/billing/orders', {
        method: 'POST',
        body: {
          ...input,
          currency: 'KRW',
          idempotency_key: crypto.randomUUID()
        } satisfies CreateOrderInput
      })
      const toss = await loadTossPayments(config.toss_client_key)
      const payment = toss.payment({ customerKey: user.value?.id || 'guest' })
      const origin = window.location.origin
      await payment.requestPayment({
        method: 'CARD',
        amount: { currency: 'KRW', value: order.gross_amount_minor },
        orderId: order.id,
        orderName,
        successUrl: `${origin}/billing/success?blog=${encodeURIComponent(returnPath)}`,
        failUrl: `${origin}/billing/fail?blog=${encodeURIComponent(returnPath)}`
      })
    } catch (error) {
      if (error instanceof ApiError) {
        toast.push(error.message || '주문 생성에 실패했습니다', 'error')
      }
      // Toss window errors (user cancel) are silent: failUrl handles them.
    } finally {
      busy.value = false
    }
  }

  async function subscribeTier(blogId: string, blogHandle: string, tier: MembershipTier): Promise<void> {
    await start(
      {
        order_type: 'MEMBERSHIP_SUB',
        recipient_blog_id: blogId,
        target_ref_id: tier.id,
        gross_amount_minor: tier.monthly_price_minor
      },
      `${tier.name} 멤버십 (월간)`,
      blogHandle
    )
  }

  async function unlockPost(blogId: string, blogHandle: string, postId: string, priceMinor: number, title: string): Promise<void> {
    await start(
      {
        order_type: 'SINGLE_POST_UNLOCK',
        recipient_blog_id: blogId,
        target_ref_id: postId,
        gross_amount_minor: priceMinor
      },
      `글 열람: ${title}`,
      blogHandle
    )
  }

  return { busy: readonly(busy), start, subscribeTier, unlockPost }
}
