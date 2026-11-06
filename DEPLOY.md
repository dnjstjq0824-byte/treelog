# TreeBlog Cloudflare 배포 런북

프론트엔드 3종은 Cloudflare Pages 다이렉트 업로드로 배포한다.
Git 연동 배포를 쓰려면 푸시 권한(아래 4절) 해결 후 Pages 프로젝트에 리포지토리를 연결한다.

## 1. 사전 준비

- `npm install -g wrangler` + `wrangler login` (OAuth, 브라우저 1회 승인)
- 프론트엔드 빌드 시점에 API 주소 주입 (빌드 결과에 고정됨):
  - user-web: `NUXT_PUBLIC_API_BASE=https://api.example.com`
  - consoles: `VITE_API_BASE=https://api.example.com`

## 2. 빌드

```powershell
# user-web (SSR, cloudflare-pages 프리셋)
$env:NITRO_PRESET='cloudflare-pages'
$env:NUXT_PUBLIC_API_BASE='https://api.example.com'
$env:NUXT_PUBLIC_CONSOLE_URL='https://studio.example.com'
$env:NUXT_PUBLIC_ADMIN_URL='https://admin.example.com'
pnpm --filter @treeblog/user-web build
# 산출물: apps/user-web/dist/  (_worker.js + 정적 애셋)

# consoles (SPA 정적)
$env:VITE_API_BASE='https://api.example.com'
pnpm --filter @treeblog/blog-console --filter @treeblog/admin-console build
# 산출물: apps/*/dist/  (public/_redirects 로 SPA 폴백 포함)
```

## 3. 배포

```powershell
cd apps/user-web
wrangler pages deploy dist --project-name treeblog-user-web --branch main
cd ../blog-console
wrangler pages deploy dist --project-name treeblog-studio --branch main
cd ../admin-console
wrangler pages deploy dist --project-name treeblog-admin --branch main
```

프로젝트가 없으면 첫 배포 시 자동 생성된다. 커스텀 도메인은
Cloudflare 대시보드 → Pages → Custom domains 에서 연결한다.

## 4. 알려진 제약

- GitHub 푸시: 현재 자격증명(`dnjstjqppk-dev`)에 `dnjstjq0824-byte/treelog` 쓰기
  권한이 없어 `git push`가 403으로 거부된다. 리포지토리 소유자가 권한을
  부여하거나, 권한 있는 계정으로 직접 푸시해야 CI·Git 연동 배포가 동작한다.
- Go API는 Cloudflare Pages 대상이 아니다. 선택지:
  1. 자체 서버/VPS에서 `docker compose up --build` 로 API+Postgres+Redis
     실행 후 프론트엔드 빌드에 해당 API 주소를 주입 (권장, 최소 변경)
  2. Cloudflare Containers(Workers 컨테이너)로 API 이전 — Docker 빌드 환경과
     외부 Postgres(Hyperdrive 경유)가 별도로 필요하다.
- MinIO·OpenSearch·Mailpit은 로컬/자체 호스팅용이다. 운영에서는 S3 호환
  스토리지(R2 포함), 관리형 Postgres, SMTP 릴레이를 `.env`에 설정한다.
