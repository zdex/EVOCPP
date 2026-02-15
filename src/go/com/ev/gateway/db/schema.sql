create table if not exists chargers (
  id bigserial primary key,
  charge_point_id text unique not null,
  shared_secret_hash text not null,
  vendor text,
  model text,
  ocpp_version text default '1.6J',
  is_active boolean default true,
  created_at timestamptz default now(),
  updated_at timestamptz default now()
);

create table if not exists gateway_events (
  id bigserial primary key,
  charge_point_id text not null,
  event_type text not null,
  ts timestamptz not null,
  payload jsonb not null,
  received_at timestamptz default now()
);
create index if not exists idx_gateway_events_cp_ts on gateway_events(charge_point_id, ts);

create table if not exists gateway_commands (
  id bigserial primary key,
  command_id uuid unique not null,
  charge_point_id text not null,
  command_type text not null,
  idempotency_key text not null,
  payload jsonb not null,
  status text not null default 'Queued',
  last_error text,
  created_at timestamptz default now(),
  updated_at timestamptz default now()
);
create unique index if not exists uq_gateway_commands_idem
  on gateway_commands(charge_point_id, idempotency_key);
