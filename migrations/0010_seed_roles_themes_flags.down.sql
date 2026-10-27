-- Migration: 0010_seed_roles_themes_flags.down.sql
DELETE FROM feature_flags WHERE key IN (
    'FF_CUSTOM_DOMAINS_ACME',
    'FF_MONETIZATION_ADS',
    'FF_MONETIZATION_MEMBERSHIP',
    'FF_NATIVE_VIDEO_TRANSCODE',
    'FF_TREESCAN_INTEGRATION_ENABLED'
);
DELETE FROM themes WHERE slug IN (
    'forest-editorial', 'classic-neighbor', 'tech-monolith', 'visual-magazine', 'photo-atelier', 'paper-minimal'
);
DELETE FROM roles WHERE code IN (
    'USER', 'ADVERTISER', 'SUPPORT_AGENT', 'TRUST_SAFETY_MOD', 'AD_REVIEWER', 'FINANCE_ADMIN', 'SUPER_ADMIN'
);
