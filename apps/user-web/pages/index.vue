<template>
  <div class="tb-app">
    <!-- 1. Global Header -->
    <header class="tb-header">
      <div class="tb-container tb-header-inner">
        <div class="tb-logo-group">
          <a href="/" class="tb-logo-link" style="display: flex; align-items: center; gap: 10px;">
            <div class="tb-logo-mark">🌿</div>
            <span class="tb-logo-text">TreeBlog</span>
          </a>
          <span class="tb-logo-tagline">스토리가 뿌리내리는 곳</span>
        </div>

        <nav class="tb-nav" style="display: none; @media(min-width: 768px) { display: flex; }">
          <a href="#" :class="['tb-nav-link', activeTab === 'trending' ? 'active' : '']" @click.prevent="activeTab = 'trending'">인기글</a>
          <a href="#" :class="['tb-nav-link', activeTab === 'recommended' ? 'active' : '']" @click.prevent="activeTab = 'recommended'">에디터 추천</a>
          <a href="#" :class="['tb-nav-link', activeTab === 'latest' ? 'active' : '']" @click.prevent="activeTab = 'latest'">최신 피드</a>
          <a href="#" :class="['tb-nav-link', activeTab === 'following' ? 'active' : '']" @click.prevent="activeTab = 'following'">이웃 새글</a>
        </nav>

        <div class="tb-actions-group">
          <button class="tb-search-trigger" @click="showSearchModal = true">
            <span>🔍</span>
            <span style="display: none; @media(min-width: 640px) { display: inline; }">검색...</span>
          </button>
          
          <button class="tb-btn tb-btn-outline" @click="openAuth('login')">로그인</button>
          <button class="tb-btn tb-btn-primary" @click="openWrite">
            <span>✏️ 글쓰기</span>
          </button>
        </div>
      </div>
    </header>

    <main class="tb-container">
      <!-- 2. Topics Bar -->
      <section class="tb-topics-bar" role="navigation" aria-label="주제별 카테고리">
        <button 
          v-for="cat in categories" 
          :key="cat.slug"
          :class="['tb-topic-pill', selectedCategory === cat.slug ? 'active' : '']"
          @click="selectCategory(cat.slug)"
        >
          {{ cat.emoji }} {{ cat.name }}
        </button>
      </section>

      <!-- 3. Hero Editorial Banner -->
      <section class="tb-hero">
        <div class="tb-hero-content">
          <div class="tb-hero-eyebrow">✨ 오늘의 주목할 스토리</div>
          <h1 class="tb-hero-title">단단한 기록이 모여<br>하나의 거대한 숲이 됩니다.</h1>
          <p class="tb-hero-desc">
            TreeBlog는 깊이 있는 사유와 지속 가능한 기록을 위한 차세대 블로그 플랫폼입니다.
            Tistory의 자유로운 테마 커스텀과 Google Blogger의 심플함, 네이버 이웃의 온기를 함께 경험해 보세요.
          </p>
          <div class="tb-hero-stats">
            <div class="tb-hero-stat-item">
              <span class="tb-hero-stat-num">12,480+</span>
              <span class="tb-hero-stat-label">활성 블로그</span>
            </div>
            <div class="tb-hero-stat-item">
              <span class="tb-hero-stat-num">89,200+</span>
              <span class="tb-hero-stat-label">발행된 아티클</span>
            </div>
            <div class="tb-hero-stat-item">
              <span class="tb-hero-stat-num">99.98%</span>
              <span class="tb-hero-stat-label">글로벌 고가용성</span>
            </div>
          </div>
        </div>
        <div class="tb-hero-card" style="display: flex; align-items: center; justify-content: center;">
          <div style="background: white; border-radius: 14px; padding: 24px; box-shadow: var(--tb-shadow-md); border: 1px solid var(--tb-border); width: 100%;">
            <div style="font-size: 0.8rem; font-weight: 700; color: var(--tb-brand-primary); margin-bottom: 8px;">🌱 블로그 시작하기</div>
            <h3 style="font-size: 1.15rem; margin-bottom: 10px; font-weight: 700;">3분 만에 나만의 독립 블로그 개설</h3>
            <p style="font-size: 0.875rem; color: var(--tb-text-secondary); margin-bottom: 16px;">개인 도메인 연결, 마크다운 & 위지윅 Tiptap 에디터, 애드센스 및 자체 광고 수익화 완벽 지원.</p>
            <button class="tb-btn tb-btn-primary" style="width: 100%;" @click="openAuth('register')">내 블로그 무료로 만들기 →</button>
          </div>
        </div>
      </section>

      <!-- 4. Main Two-column Layout -->
      <div class="tb-main-grid">
        <!-- Left: Feed Content Area -->
        <section class="tb-content-area">
          <!-- Feed Tab Switcher -->
          <div class="tb-feed-nav" role="tablist">
            <button 
              :class="['tb-feed-tab', activeTab === 'trending' ? 'active' : '']"
              @click="activeTab = 'trending'"
            >
              🔥 트렌딩 인기글
            </button>
            <button 
              :class="['tb-feed-tab', activeTab === 'recommended' ? 'active' : '']"
              @click="activeTab = 'recommended'"
            >
              ✨ 에디터 추천
            </button>
            <button 
              :class="['tb-feed-tab', activeTab === 'latest' ? 'active' : '']"
              @click="activeTab = 'latest'"
            >
              ⏱ 실시간 최신글
            </button>
            <button 
              :class="['tb-feed-tab', activeTab === 'following' ? 'active' : '']"
              @click="activeTab = 'following'"
            >
              👥 이웃 새글
            </button>
          </div>

          <!-- Post Cards Stream -->
          <div class="tb-post-list">
            <article 
              v-for="post in filteredPosts" 
              :key="post.id" 
              class="tb-post-card"
            >
              <div class="tb-card-content">
                <div>
                  <div class="tb-card-author">
                    <div class="tb-avatar">{{ post.author[0] }}</div>
                    <span class="tb-author-name">{{ post.author }}</span>
                    <span class="tb-author-dot">·</span>
                    <span class="tb-card-date">{{ post.publishedAt }}</span>
                    <span class="tb-card-badge" style="margin-left: auto;">{{ post.category }}</span>
                  </div>

                  <h2 class="tb-card-title">
                    <a href="#" @click.prevent="openPost(post)">{{ post.title }}</a>
                  </h2>

                  <p class="tb-card-excerpt">
                    {{ post.excerpt }}
                  </p>
                </div>

                <div class="tb-card-footer">
                  <div class="tb-card-meta">
                    <span>📖 {{ post.readingTime }}</span>
                    <span>👀 {{ post.viewCount.toLocaleString() }}</span>
                    <span>❤️ {{ post.reactionCount }}</span>
                    <span>💬 {{ post.commentCount }}</span>
                  </div>

                  <button 
                    class="tb-bookmark-btn" 
                    :style="{ color: post.bookmarked ? 'var(--tb-brand-primary)' : 'var(--tb-text-muted)' }"
                    @click="toggleBookmark(post)"
                    title="북마크"
                  >
                    {{ post.bookmarked ? '★ 저장됨' : '☆ 저장' }}
                  </button>
                </div>
              </div>

              <div class="tb-card-thumb" :style="{ background: post.thumbGradient }">
                <span>{{ post.categoryEmoji }}</span>
              </div>
            </article>
          </div>
        </section>

        <!-- Right: Sticky Sidebar -->
        <aside class="tb-sidebar">
          <!-- Notice / Promotion Card -->
          <div class="tb-side-card" style="background: linear-gradient(135deg, #15803D 0%, #166534 100%); color: white;">
            <div style="font-size: 0.8rem; font-weight: 700; opacity: 0.9; margin-bottom: 6px;">📢 PLATFORM RELEASE</div>
            <h4 style="font-size: 1.1rem; font-weight: 700; margin-bottom: 8px;">TreeBlog 1.0 정식 오픈</h4>
            <p style="font-size: 0.85rem; opacity: 0.88; line-height: 1.5; margin-bottom: 14px;">
              고품질 한국어/글로벌 블로거를 위한 차세대 아키텍처 오픈. 검색엔진 최적화와 독립 도메인 연결을 즉시 사용하세요.
            </p>
            <a href="#" style="font-size: 0.85rem; font-weight: 700; color: #DCFCE7; text-decoration: underline;" @click.prevent="openAuth('register')">크리에이터 계정 만들기 →</a>
          </div>

          <!-- Top Creators Card -->
          <div class="tb-side-card">
            <h3 class="tb-side-title">
              <span>🌟 추천 크리에이터</span>
              <a href="#" style="font-size: 0.8rem; color: var(--tb-brand-primary); font-weight: 600;">전체보기</a>
            </h3>
            <div class="tb-creator-list">
              <div v-for="creator in creators" :key="creator.handle" class="tb-creator-item">
                <div class="tb-creator-info">
                  <div class="tb-creator-avatar">{{ creator.emoji }}</div>
                  <div>
                    <div class="tb-creator-name">{{ creator.name }}</div>
                    <div class="tb-creator-handle">@{{ creator.handle }} · 이웃 {{ creator.followers }}</div>
                  </div>
                </div>
                <button 
                  class="tb-follow-btn" 
                  :style="creator.isFollowing ? 'background: #F1F5F9; color: #475569;' : ''"
                  @click="toggleFollow(creator)"
                >
                  {{ creator.isFollowing ? '이웃중' : '+ 이웃추가' }}
                </button>
              </div>
            </div>
          </div>

          <!-- Popular Tags Card -->
          <div class="tb-side-card">
            <h3 class="tb-side-title">
              <span>🏷️ 실시간 인기 태그</span>
            </h3>
            <div style="display: flex; flex-wrap: wrap; gap: 8px;">
              <a 
                v-for="tag in popularTags" 
                :key="tag.name" 
                href="#" 
                class="tb-topic-pill"
                style="font-size: 0.8rem; padding: 5px 12px;"
                @click.prevent="selectTag(tag.name)"
              >
                #{{ tag.name }} <span style="color: var(--tb-text-muted); margin-left: 3px;">{{ tag.count }}</span>
              </a>
            </div>
          </div>
        </aside>
      </div>
    </main>

    <!-- 5. Global Footer -->
    <footer class="tb-footer">
      <div class="tb-container">
        <div class="tb-footer-inner">
          <div class="tb-footer-brand">
            <h4>🌿 TreeBlog</h4>
            <p style="max-width: 300px; line-height: 1.6; margin-bottom: 12px;">
              모든 이야기와 아이디어가 깊은 뿌리를 내리고 자라나는 곳. 차세대 글로벌 멀티도메인 블로그 플랫폼.
            </p>
            <p style="font-size: 0.8rem; color: var(--tb-text-muted);">
              Powered by Go 1.24 Monolith · PostgreSQL 16 · Nuxt 3 SSR
            </p>
          </div>

          <div class="tb-footer-col">
            <h5>둘러보기</h5>
            <ul class="tb-footer-links">
              <li><a href="#">인기 게시글</a></li>
              <li><a href="#">추천 블로그</a></li>
              <li><a href="#">카테고리 탐색</a></li>
              <li><a href="#">태그 클라우드</a></li>
            </ul>
          </div>

          <div class="tb-footer-col">
            <h5>블로그 운영</h5>
            <ul class="tb-footer-links">
              <li><a href="#" @click.prevent="openAuth('register')">블로그 개설하기</a></li>
              <li><a href="#">테마 스튜디오</a></li>
              <li><a href="#">개인 도메인 연결</a></li>
              <li><a href="#">RSS / Atom 피드</a></li>
            </ul>
          </div>

          <div class="tb-footer-col">
            <h5>지원 및 정책</h5>
            <ul class="tb-footer-links">
              <li><a href="#">이용약관</a></li>
              <li><a href="#">개인정보처리방침</a></li>
              <li><a href="#">신고 및 이의제기</a></li>
              <li><a href="#">고객센터</a></li>
            </ul>
          </div>
        </div>

        <div class="tb-footer-bottom">
          <span>© 2026 TreeBlog Platform Co., Ltd. All rights reserved.</span>
          <div style="display: flex; gap: 18px;">
            <a href="#">개인정보처리방침</a>
            <a href="#">서비스 이용약관</a>
            <a href="#">시스템 상태</a>
          </div>
        </div>
      </div>
    </footer>

    <!-- Search Modal -->
    <div v-if="showSearchModal" class="tb-modal-backdrop" @click.self="showSearchModal = false">
      <div class="tb-modal-card">
        <div class="tb-modal-header">
          <h3 class="tb-modal-title">🔍 TreeBlog 통합 검색</h3>
          <button style="font-size: 1.5rem;" @click="showSearchModal = false">✕</button>
        </div>
        <div style="margin-bottom: 16px;">
          <input 
            v-model="searchQuery" 
            type="text" 
            class="tb-input" 
            placeholder="키워드, 태그(#), 작성자명을 입력하세요..." 
            autofocus
            @keydown.enter="performSearch"
          />
        </div>
        <div style="font-size: 0.85rem; color: var(--tb-text-muted);">
          추천 검색어: 
          <span v-for="s in ['Go 아키텍처', 'Nuxt3 가이드', '제주도 카페', '생산성 도구']" :key="s" style="margin-right: 8px; color: var(--tb-brand-primary); cursor: pointer;" @click="searchQuery = s; performSearch()">
            #{{ s }}
          </span>
        </div>
      </div>
    </div>

    <!-- Auth Modal -->
    <div v-if="showAuthModal" class="tb-modal-backdrop" @click.self="showAuthModal = false">
      <div class="tb-modal-card">
        <div class="tb-modal-header">
          <h3 class="tb-modal-title">{{ authMode === 'login' ? '로그인' : '회원가입' }}</h3>
          <button style="font-size: 1.5rem;" @click="showAuthModal = false">✕</button>
        </div>
        
        <form @submit.prevent="handleAuthSubmit" style="display: flex; flex-direction: column; gap: 14px;">
          <div>
            <label style="font-size: 0.85rem; font-weight: 600; display: block; margin-bottom: 6px;">이메일</label>
            <input v-model="authEmail" type="email" class="tb-input" placeholder="name@example.com" required />
          </div>
          <div>
            <label style="font-size: 0.85rem; font-weight: 600; display: block; margin-bottom: 6px;">비밀번호</label>
            <input v-model="authPassword" type="password" class="tb-input" placeholder="최소 8자 이상" required />
          </div>
          <div v-if="authMode === 'register'">
            <label style="font-size: 0.85rem; font-weight: 600; display: block; margin-bottom: 6px;">닉네임</label>
            <input v-model="authNickname" type="text" class="tb-input" placeholder="활동하실 닉네임" required />
          </div>

          <button type="submit" class="tb-btn tb-btn-primary" style="margin-top: 10px; padding: 12px;">
            {{ authMode === 'login' ? 'TreeBlog 로그인' : '무료 회원가입 완료' }}
          </button>

          <div style="text-align: center; font-size: 0.875rem; color: var(--tb-text-secondary); margin-top: 10px;">
            <span v-if="authMode === 'login'">
              계정이 없으신가요? <a href="#" style="color: var(--tb-brand-primary); font-weight: 700;" @click.prevent="authMode = 'register'">회원가입</a>
            </span>
            <span v-else>
              이미 계정이 있으신가요? <a href="#" style="color: var(--tb-brand-primary); font-weight: 700;" @click.prevent="authMode = 'login'">로그인</a>
            </span>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const activeTab = ref('trending')
const selectedCategory = ref('all')
const showSearchModal = ref(false)
const showAuthModal = ref(false)
const authMode = ref('login')
const searchQuery = ref('')
const authEmail = ref('')
const authPassword = ref('')
const authNickname = ref('')

const categories = [
  { slug: 'all', name: '전체', emoji: '🌟' },
  { slug: 'tech', name: '테크·개발', emoji: '💻' },
  { slug: 'travel', name: '여행·아웃도어', emoji: '✈️' },
  { slug: 'food', name: '맛집·푸드', emoji: '🍲' },
  { slug: 'economy', name: '재테크·경제', emoji: '📈' },
  { slug: 'culture', name: '문화·예술', emoji: '🎨' },
  { slug: 'life', name: '일상·생각', emoji: '💡' },
]

const posts = ref([
  {
    id: 'post-1',
    title: 'Go 1.24 모듈형 모놀리스와 PostgreSQL 파티셔닝으로 일 1,000만 뷰 견디기',
    excerpt: '마이크로서비스의 복잡성을 줄이고 단일 바이너리로 극상의 생산성과 확장성을 동시에 잡는 엔지니어링 실전 아키텍처 패턴을 공유합니다.',
    author: '김아키',
    category: '테크·개발',
    categorySlug: 'tech',
    categoryEmoji: '💻',
    publishedAt: '2시간 전',
    readingTime: '7분 읽기',
    viewCount: 4210,
    reactionCount: 182,
    commentCount: 24,
    bookmarked: false,
    thumbGradient: 'linear-gradient(135deg, #10B981, #047857)'
  },
  {
    id: 'post-2',
    title: '교토 뒷골목의 80년 된 핸드드립 다방과 아날로그 필름 산책',
    excerpt: '관광객의 발길이 닿지 않는 골목 끝, 사이폰 커피 향이 가득한 작은 노포에서 마주한 교토의 진짜 가을 이야기.',
    author: '이지은',
    category: '여행·아웃도어',
    categorySlug: 'travel',
    categoryEmoji: '✈️',
    publishedAt: '4시간 전',
    readingTime: '5분 읽기',
    viewCount: 3840,
    reactionCount: 245,
    commentCount: 38,
    bookmarked: true,
    thumbGradient: 'linear-gradient(135deg, #F59E0B, #B45309)'
  },
  {
    id: 'post-3',
    title: 'Tistory와 Blogger에서 독립 도메인 커스텀 블로그로 이사하며 겪은 SEO 실험',
    excerpt: '기존 URL의 301 리다이렉트 처리, OpenGraph 및 JSON-LD 구조화 데이터 적용 후 구글 검색 유입이 180% 증가한 전수 데이터 분석 결과.',
    author: '박마케터',
    category: '재테크·경제',
    categorySlug: 'economy',
    categoryEmoji: '📈',
    publishedAt: '6시간 전',
    readingTime: '10분 읽기',
    viewCount: 5120,
    reactionCount: 310,
    commentCount: 42,
    bookmarked: false,
    thumbGradient: 'linear-gradient(135deg, #3B82F6, #1D4ED8)'
  },
  {
    id: 'post-4',
    title: '프랑스 정통 사워도우를 집에서 굽기 위한 르방 배양과 수분율의 과학',
    excerpt: '밀가루와 물, 공기 중의 효모만으로 완성하는 크리스피한 크러스트와 촉촉한 크럼. 7일간의 발효 일지와 실패 없는 베이킹 가이드.',
    author: '최베이커',
    category: '맛집·푸드',
    categorySlug: 'food',
    categoryEmoji: '🍲',
    publishedAt: '8시간 전',
    readingTime: '6분 읽기',
    viewCount: 2980,
    reactionCount: 142,
    commentCount: 19,
    bookmarked: false,
    thumbGradient: 'linear-gradient(135deg, #EC4899, #BE185D)'
  },
  {
    id: 'post-5',
    title: '아침 30분, 일상을 바꾸는 깊은 글쓰기와 모닝 저널링의 기록들',
    excerpt: '타인의 피드를 스크롤하는 대신 내면의 소리를 기록하며 발견한 몰입의 힘. 산만해진 현대인을 위한 생각 정돈 루틴.',
    author: '정에세이',
    category: '일상·생각',
    categorySlug: 'life',
    categoryEmoji: '💡',
    publishedAt: '12시간 전',
    readingTime: '4분 읽기',
    viewCount: 4620,
    reactionCount: 420,
    commentCount: 56,
    bookmarked: false,
    thumbGradient: 'linear-gradient(135deg, #8B5CF6, #6D28D9)'
  }
])

const creators = ref([
  { name: '테크 아키텍트 김아키', handle: 'archi_kim', emoji: '⚡', followers: '3.4k', isFollowing: false },
  { name: '글로벌 여행자 지은', handle: 'travel_jieun', emoji: '🌏', followers: '5.1k', isFollowing: true },
  { name: '미식 기록가 최베이커', handle: 'baker_choi', emoji: '🥖', followers: '2.8k', isFollowing: false },
  { name: '인사이트 경제노트', handle: 'economy_note', emoji: '📊', followers: '8.9k', isFollowing: false },
])

const popularTags = [
  { name: '아키텍처', count: 482 },
  { name: '도쿄여행', count: 329 },
  { name: '개발자성장', count: 294 },
  { name: '홈베이킹', count: 210 },
  { name: '에세이', count: 188 },
  { name: '독서기록', count: 164 },
]

const filteredPosts = computed(() => {
  if (selectedCategory.value === 'all') {
    return posts.value
  }
  return posts.value.filter(p => p.categorySlug === selectedCategory.value)
})

function selectCategory(slug) {
  selectedCategory.value = slug
}

function selectTag(tagName) {
  searchQuery.value = tagName
  showSearchModal.value = true
}

function toggleBookmark(post) {
  post.bookmarked = !post.bookmarked
}

function toggleFollow(creator) {
  creator.isFollowing = !creator.isFollowing
}

function openAuth(mode) {
  authMode.value = mode
  showAuthModal.value = true
}

function openWrite() {
  // If not logged in, trigger auth modal or direct to writer
  alert('TreeBlog 글쓰기 스튜디오로 연결됩니다! (게시글 초안 생성 및 에디터 활성화)')
}

function openPost(post) {
  alert(`"${post.title}" 상세 읽기 페이지로 이동합니다.`)
}

function performSearch() {
  if (!searchQuery.value) return
  alert(`"${searchQuery.value}" 검색 결과를 조회합니다.`)
  showSearchModal.value = false
}

function handleAuthSubmit() {
  alert(`${authEmail.value} 계정으로 ${authMode.value === 'login' ? '로그인' : '회원가입'} 처리가 완료되었습니다!`)
  showAuthModal.value = false
}
</script>
