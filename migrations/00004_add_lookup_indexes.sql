-- +goose Up
CREATE INDEX "links_user_id_created_at_idx" ON "links" ("user_id", "created_at" DESC);
CREATE INDEX "clicks_link_id_clicked_at_idx" ON "clicks" ("link_id", "clicked_at");

-- +goose Down
DROP INDEX "clicks_link_id_clicked_at_idx";
DROP INDEX "links_user_id_created_at_idx";
