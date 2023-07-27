#!/usr/bin/env bash

psql -U postgres postgres <<EOF
-- init illa_builder


create database illa_builder;

\c illa_builder;

create user illa_builder with encrypted password 'illa2022';

grant all privileges on database illa_builder to illa_builder;

CREATE EXTENSION pg_trgm;

CREATE EXTENSION btree_gin;

-- apps
create table if not exists apps (
    id                      bigserial                       not null primary key,
    uid                     uuid default gen_random_uuid()  not null,
    team_id                 bigserial                       not null,
    name                    varchar(200)                    not null,
    release_version         bigint                          not null,
    mainline_version        bigint                          not null,
    config                  jsonb,
    created_at              timestamp                       not null,
    created_by              bigint                          not null,
    updated_at              timestamp                       not null,
    updated_by              bigint                          not null,
    edited_by               jsonb

);

alter table apps owner to illa_builder;

-- app_snapshots
create table if not exists app_snapshots (
    id                      bigserial                       not null primary key,
    uid                     uuid default gen_random_uuid()  not null,
    team_id                 bigserial                       not null,
    app_ref_id              bigserial                       not null,
    target_version          bigint                          not null,
    trigger_mode            smallint                        not null,
    modify_history          jsonb,
    created_at              timestamp                       not null
);

alter table app_snapshots owner to illa_builder;

-- resource
create table if not exists resources (
    id                      bigserial                       not null primary key,
    uid                     uuid default gen_random_uuid()  not null,
    team_id                 bigserial                       not null,
    name                    varchar(200)                    not null,
    type                    smallint                        not null,
    options                 jsonb,
    created_at              timestamp                       not null,
    created_by              bigint                          not null,
    updated_at              timestamp                       not null,
    updated_by              bigint                          not null
);

alter table resources owner to illa_builder;

-- actions
create table if not exists actions (
    id                      bigserial                       not null primary key,
    uid                     uuid default gen_random_uuid()  not null,
    team_id                 bigserial                       not null,
    version                 bigint                          not null,
    resource_ref_id         bigint                          not null,
    app_ref_id              bigint                          not null,
    name                    varchar(255)                    not null,
    type                    smallint                        not null,
    transformer             jsonb                           not null,
    trigger_mode            varchar(16)                     not null,
    template                jsonb,
    config                  jsonb,
    created_at              timestamp                       not null,
    created_by              bigint                          not null,
    updated_at              timestamp                       not null,
    updated_by              bigint                          not null
);

create index if not exists actions_at_apprefid_and_version on actions (app_ref_id, version);
alter table actions owner to illa_builder;


ALTER TABLE actions DROP CONSTRAINT IF EXISTS actions_displayname_constrainte,
ADD CONSTRAINT actions_displayname_constrainte UNIQUE (version, app_ref_id, name);

-- tree_states, component tree_states
create table if not exists tree_states (
    id                      bigserial                       not null primary key,
    uid                     uuid default gen_random_uuid()  not null,
    team_id                 bigserial                       not null,
    state_type              smallint                        not null,
    parent_node_ref_id      bigint                          not null,
    children_node_ref_ids   jsonb,
    app_ref_id              bigint                          not null,
    version                 bigint                          not null,
    name                    text                            not null,
    content                 jsonb                           not null,
    created_at              timestamp                       not null,
    created_by              bigint                          not null,
    updated_at              timestamp                       not null,
    updated_by              bigint                          not null
);

CREATE INDEX tree_states_at_apprefid_and_version_and_statetype ON tree_states (app_ref_id, version, state_type);
CREATE INDEX tree_states_at_parentnoderefid ON tree_states (parent_node_ref_id);
CREATE INDEX tree_states_at_childrennoderefids ON tree_states (children_node_ref_ids);
CREATE INDEX tree_states_with_gin_at_childrennoderefids ON tree_states USING gin (children_node_ref_ids);
CREATE INDEX tree_states_with_gin_at_name ON tree_states USING gin (name);
CREATE INDEX tree_states_with_fulltextgin_at_name ON tree_states USING gin (to_tsvector('english', name));

ALTER TABLE tree_states DROP CONSTRAINT IF EXISTS tree_states_displayname_constrainte,
ADD CONSTRAINT tree_states_displayname_constrainte UNIQUE (version, app_ref_id, name);

alter table tree_states owner to illa_builder;

-- kv_states, component kv_states
create table if not exists kv_states (
    id                      bigserial                       not null primary key,
    uid                     uuid default gen_random_uuid()  not null,
    team_id                 bigserial                       not null,
    state_type              smallint                        not null,
    app_ref_id              bigint                          not null,
    version                 bigint                          not null,
    key                     text                            not null,
    value                   jsonb                           not null,
    created_at              timestamp                       not null,
    created_by              bigint                          not null,
    updated_at              timestamp                       not null,
    updated_by              bigint                          not null
);

CREATE INDEX kv_states_at_apprefid_and_version_and_statetype ON kv_states (app_ref_id, version, state_type);
CREATE INDEX kv_states_with_gin_at_key ON kv_states USING gin (key);
CREATE INDEX kv_states_with_fulltextgin_at_key ON kv_states USING gin (to_tsvector('english', key));
ALTER TABLE kv_states DROP CONSTRAINT IF EXISTS kv_states_displayname_constrainte,
ADD CONSTRAINT kv_states_displayname_constrainte UNIQUE (version, app_ref_id, key);

alter table kv_states owner to illa_builder;

-- set_states, component set_states
create table if not exists set_states (
    id                      bigserial                       not null primary key,
    uid                     uuid default gen_random_uuid()  not null,
    team_id                 bigserial                       not null,
    state_type              smallint                        not null,
    app_ref_id              bigint                          not null,
    version                 bigint                          not null,
    value                   text                            not null,
    created_at              timestamp                       not null,
    created_by              bigint                          not null,
    updated_at              timestamp                       not null,
    updated_by              bigint                          not null
);

CREATE INDEX set_states_at_apprefid_and_version_and_statetype ON set_states (app_ref_id, version, state_type);
CREATE INDEX set_states_with_gin_at_value ON set_states USING gin (value);
CREATE INDEX set_states_with_fulltextgin_at_value ON set_states USING gin (to_tsvector('english', value));

ALTER TABLE set_states DROP CONSTRAINT IF EXISTS set_states_displayname_constrainte,
ADD CONSTRAINT set_states_displayname_constrainte UNIQUE (version, app_ref_id, value);

alter table set_states owner to illa_builder;

EOF
