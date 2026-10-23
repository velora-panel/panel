-- +goose Up
ALTER TABLE stations RENAME TO nodes;
ALTER TABLE nodes RENAME CONSTRAINT stations_pkey TO nodes_pkey;
ALTER TABLE nodes RENAME CONSTRAINT stations_name_key TO nodes_name_key;
ALTER TABLE nodes RENAME CONSTRAINT stations_name_check TO nodes_name_check;
ALTER TABLE nodes RENAME CONSTRAINT stations_description_check TO nodes_description_check;
ALTER TABLE nodes RENAME CONSTRAINT stations_status_check TO nodes_status_check;
ALTER TABLE instances RENAME COLUMN station_id TO node_id;
ALTER TABLE instances RENAME CONSTRAINT instances_station_id_fkey TO instances_node_id_fkey;
ALTER INDEX instances_station_id_idx RENAME TO instances_node_id_idx;
-- +goose Down
ALTER INDEX instances_node_id_idx RENAME TO instances_station_id_idx;
ALTER TABLE instances RENAME CONSTRAINT instances_node_id_fkey TO instances_station_id_fkey;
ALTER TABLE instances RENAME COLUMN node_id TO station_id;
ALTER TABLE nodes RENAME CONSTRAINT nodes_status_check TO stations_status_check;
ALTER TABLE nodes RENAME CONSTRAINT nodes_description_check TO stations_description_check;
ALTER TABLE nodes RENAME CONSTRAINT nodes_name_check TO stations_name_check;
ALTER TABLE nodes RENAME CONSTRAINT nodes_name_key TO stations_name_key;
ALTER TABLE nodes RENAME CONSTRAINT nodes_pkey TO stations_pkey;
ALTER TABLE nodes RENAME TO stations;
