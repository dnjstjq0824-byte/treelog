<template>
  <div>
    <header class="tb-site-header">
      <div class="tb-container tb-site-header-inner">
        <NuxtLink to="/" class="tb-brand" aria-label="TreeBlog 홈">
          <span class="tb-brand-mark" aria-hidden="true">
            <svg viewBox="0 0 32 32" fill="none"><path d="M16 5 L24 14 H19 V24 H13 V14 H8 Z" fill="currentColor" /></svg>
          </span>
          TreeBlog
        </NuxtLink>
      </div>
    </header>
    <main id="main-content" class="tb-auth-wrap">
      <slot />
    </main>
  </div>
</template>
