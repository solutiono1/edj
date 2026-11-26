-- edj Supabase schema.
-- Run in the Supabase SQL editor. The service role key is used only by the backend.

create table if not exists public.links (
  id text primary key,
  address text not null,
  amount bigint not null,
  token text,
  note text,
  created_at timestamptz not null,
  status text not null,
  derivation_index bigint not null default 0
);

create table if not exists public.payroll_runs (
  id text primary key,
  rows_json jsonb not null,
  total bigint not null,
  plan_json jsonb not null,
  status text not null,
  created_at timestamptz not null,
  results_json jsonb not null default '[]'::jsonb
);

create table if not exists public.proofs (
  id text primary key,
  subject text not null,
  threshold bigint not null,
  expires_at timestamptz not null,
  signature text not null,
  public_key text not null,
  claim text not null
);

create table if not exists public.users (
  sub text primary key,
  email text not null,
  name text not null,
  password_hash text
);

create index if not exists users_email_idx on public.users (email);
create unique index if not exists users_email_password_unique on public.users (lower(email)) where password_hash is not null;

create table if not exists public.sessions (
  token text primary key,
  sub text not null
);

create table if not exists public.notifications (
  id text primary key,
  title text not null,
  body text not null,
  created_at timestamptz not null,
  read boolean not null default false
);

create table if not exists public.meta (
  key text primary key,
  value text not null
);

alter table public.links enable row level security;
alter table public.payroll_runs enable row level security;
alter table public.proofs enable row level security;
alter table public.users enable row level security;
alter table public.sessions enable row level security;
alter table public.notifications enable row level security;
alter table public.meta enable row level security;
-- No policies are created. Only the service role key can read or write.
