# OPENSEARCH_SCHEMA.md — OpenSearch 다국어 인덱스 매핑 및 분석기 설계 (제10조)

## 1. `treeblog_posts_v1` 인덱스 설정 및 형태소 분석기 매핑
```json
{
  "settings": {
    "index": {
      "number_of_shards": 3,
      "number_of_replicas": 1,
      "refresh_interval": "1s"
    },
    "analysis": {
      "tokenizer": {
        "korean_nori_tokenizer": {
          "type": "nori_tokenizer",
          "decompound_mode": "mixed"
        }
      },
      "analyzer": {
        "ko_analyzer": {
          "type": "custom",
          "tokenizer": "korean_nori_tokenizer",
          "filter": ["lowercase", "nori_part_of_speech", "nori_readingform"]
        },
        "multilingual_icu": {
          "type": "custom",
          "tokenizer": "icu_tokenizer",
          "filter": ["lowercase", "icu_folding"]
        },
        "edge_ngram_suggest": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase"]
        }
      }
    }
  },
  "mappings": {
    "dynamic": "strict",
    "properties": {
      "post_id": { "type": "keyword" },
      "blog_id": { "type": "keyword" },
      "blog_handle": { "type": "keyword" },
      "blog_title": { "type": "text", "analyzer": "ko_analyzer", "fields": { "keyword": { "type": "keyword" } } },
      "author_id": { "type": "keyword" },
      "author_display_name": { "type": "text", "analyzer": "ko_analyzer", "fields": { "keyword": { "type": "keyword" } } },
      "title": {
        "type": "text",
        "analyzer": "ko_analyzer",
        "fields": {
          "icu": { "type": "text", "analyzer": "multilingual_icu" },
          "exact": { "type": "keyword" }
        }
      },
      "subtitle": { "type": "text", "analyzer": "ko_analyzer" },
      "summary": { "type": "text", "analyzer": "ko_analyzer" },
      "content_plain": {
        "type": "text",
        "analyzer": "ko_analyzer",
        "fields": { "icu": { "type": "text", "analyzer": "multilingual_icu" } }
      },
      "tags": { "type": "keyword" },
      "category_name": { "type": "text", "analyzer": "ko_analyzer", "fields": { "keyword": { "type": "keyword" } } },
      "topic_slug": { "type": "keyword" },
      "language_code": { "type": "keyword" },
      "quality_score": { "type": "float" },
      "popularity_score": { "type": "float" },
      "view_count": { "type": "long" },
      "reaction_count": { "type": "integer" },
      "comment_count": { "type": "integer" },
      "published_at": { "type": "date" },
      "updated_at": { "type": "date" }
    }
  }
}
```
