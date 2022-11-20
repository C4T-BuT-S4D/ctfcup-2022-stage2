#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	create database xlebzavod;
  \c xlebzavod

  create extension pgcrypto;

  create table users (
    username text primary key,
    password text not null
  );

  create table orders (
    id uuid primary key default gen_random_uuid(),
    username text not null references users,
    bread text not null,
    created_at timestamp not null default now()
  );

  create user zavod with encrypted password '$ZAVOD_PASSWORD';
  grant select, insert, delete on users, orders to zavod;

  create user magaz with encrypted password '$MAGAZ_PASSWORD';
  grant select on orders to magaz;
EOSQL
