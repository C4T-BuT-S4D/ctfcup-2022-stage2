#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	create database xlebzavod;
  \c xlebzavod

  create extension pgcrypto;

  create table users (
    username text primary key,
    password text not null,
    created_at timestamp not null default now()
  );

  create table orders (
    id bigint primary key default ('x'||right(gen_random_bytes(4)::text, 8))::bit(32)::bigint,
    username text not null references users,
    bread text not null,
    recipient text not null,
    created_at timestamp not null default now()
  );

  create function clean_old() returns trigger as \$\$
    begin
      delete from users where created_at < now() - interval '30 mins';
      delete from orders where created_at < now() - interval '30 mins';
      return null;
    end;
  \$\$ language plpgsql;

  create trigger tg_clean_old
  after insert on users
  for each statement
  execute function clean_old();

  create user "$ZAVOD_USER" with encrypted password '$ZAVOD_PASSWORD';
  grant select, insert, delete on users, orders to "$ZAVOD_USER";

  create user "$MAGAZ_USER" with encrypted password '$MAGAZ_PASSWORD';
  grant select on orders to "$MAGAZ_USER";
EOSQL
