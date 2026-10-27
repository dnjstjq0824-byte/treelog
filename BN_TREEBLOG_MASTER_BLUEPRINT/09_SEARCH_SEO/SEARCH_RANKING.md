# SEARCH_RANKING.md — 검색 관련도 랭킹 모델 및 악성 SEO 조작 방지 설계 (제9조, 제10조)

## 1. 하이브리드 검색 스코어링 수식 (`function_score`)
키워드 검색 시 단순 텍스트 매칭(BM25)만 사용하면 키워드를 수백 번 반복한 스팸 글이 상단에 노출될 수 있다. 이를 방지하기 위해 TreeBlog 검색 엔진은 다음 결합 수식을 사용한다.

$$\text{FinalSearchScore}(q, d) = \text{BM25}(q, d) \times \left(0.5 + \frac{\text{quality\_score}(d)}{100}\right) \times \text{GaussianDecay}(\text{published\_at}) \times \log_{10}(10 + \text{popularity\_score}(d))$$

- **필드별 BM25 가중치**: `title^4.0`, `tags^3.5`, `subtitle^2.0`, `summary^1.5`, `content_plain^1.0`, `blog_title^1.2`, `author_display_name^1.2`.
- **구문 일치 부스트(`match_phrase`)**: 검색어 전체 구문이 제목에 연속으로 일치할 경우 `+5.0` 가산점 부여.

## 2. 검색 중립성 및 스팸 조작 방어 (제9조 준수)
- **TreeScan 연동 시 공정성 원칙**: TreeBlog 게시글을 TreeScan에 연동할 때 외부 웹 문서보다 무조건 상위에 랭크시키는 인위적 가중치 조작 로직을 절대 넣지 않는다.
- **Keyword Stuffing 및 은닉 텍스트 차단**: 에디터 정화 단계에서 글자색과 배경색이 동일한 텍스트나 `font-size: 0` 텍스트를 제거하고, 동일 단어 비정상 반복 시 `content_quality_signals.spam_probability`를 높여 검색 랭킹에서 자동 강등한다.
