-- +goose Up
CREATE INDEX "clicks_link_id_clicked_at_idx" ON "clicks" ("link_id", "clicked_at");

-- +goose Down
DROP INDEX "clicks_link_id_clicked_at_idx";
