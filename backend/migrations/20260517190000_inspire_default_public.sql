-- +goose Up
-- +goose StatementBegin
ALTER TABLE `generation_result`
  MODIFY COLUMN `is_public` TINYINT NOT NULL DEFAULT 1 COMMENT '1 公开到灵感库, 0 私有';

UPDATE `generation_result`
   SET `is_public` = 1
 WHERE `is_public` = 0
   AND `deleted_at` IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `generation_result`
  MODIFY COLUMN `is_public` TINYINT NOT NULL DEFAULT 0;
-- +goose StatementEnd
