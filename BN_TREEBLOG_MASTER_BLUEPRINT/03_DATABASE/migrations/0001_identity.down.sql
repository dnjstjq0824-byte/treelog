-- Migration: 0001_identity.down.sql
DROP TABLE IF EXISTS user_consents;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS identities;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;
