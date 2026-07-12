-- 启用pgvector扩展
CREATE EXTENSION IF NOT EXISTS vector;
-- 启用 cube 和 earthdistance 扩展(基础依赖)
CREATE EXTENSION IF NOT EXISTS cube;
CREATE EXTENSION IF NOT EXISTS earthdistance;
--安装 pg_trgm 扩展(支持 jsonb 相似度操作)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 删除已存在的vector_store表(如果存在)
DROP TABLE IF EXISTS "public"."vector_store";

-- 创建新表
CREATE TABLE IF NOT EXISTS "public"."vector_store" (
    id bigserial PRIMARY KEY,
    chat_id varchar(255) NOT NULL,
    role varchar(50) NOT NULL,
    content text NOT NULL,
    embedding JSONB NOT NULL,
    source_type varchar(50) NOT NULL DEFAULT 'message',
    created_at timestamptz NOT NULL DEFAULT now()
);

-- 创建向量索引（可选）
CREATE INDEX idx_vector_store_chat_id ON vector_store (chat_id);
CREATE INDEX idx_vector_store_created_at ON vector_store (created_at DESC);

-- 创建知识库内容表
CREATE TABLE IF NOT EXISTS "public"."knowledge_base" (
    id bigserial PRIMARY KEY,
    title varchar(255) NOT NULL,
    content text NOT NULL,
    embedding JSONB NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- 创建索引
CREATE INDEX idx_knowledge_base_title ON knowledge_base(title);