#!/usr/bin/env bash
#

set -axe

docker exec -i pipelines-test-db psql -U postgres <<EOF
CREATE DATABASE raw;
CREATE DATABASE cleansed;
CREATE DATABASE reporting;
EOF

docker exec -i pipelines-test-db psql -U postgres -d raw <<EOF
CREATE TABLE IF NOT EXISTS precipitation (id serial primary key, timestamp TIMESTAMP, location_name varchar, location_latitude numeric, location_longitude numeric, sensor varchar, precipitation numeric);
EOF

docker exec -i pipelines-test-db psql -U postgres -d cleansed <<EOF
CREATE TABLE IF NOT EXISTS precipitation (id serial primary key, timestamp TIMESTAMP, location_name varchar, location_latitude numeric, location_longitude numeric, sensor varchar, precipitation numeric);
EOF

docker exec -i pipelines-test-db psql -U postgres -d reporting <<EOF
CREATE TABLE IF NOT EXISTS precipitation (sensor varchar, timestamp timestamp, value numeric);
CREATE TABLE IF NOT EXISTS  sensor (sensor varchar primary key, location varchar, latitude numeric, longitude numeric);
CREATE UNIQUE INDEX precipitation_sensor_timestamp on precipitation(sensor, timestamp);
ALTER TABLE precipitation ADD CONSTRAINT sensor_fk FOREIGN KEY(sensor) REFERENCES sensor(sensor);
EOF
