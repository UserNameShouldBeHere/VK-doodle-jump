create database vk_games;

\c vk_games;

create table if not exists users (
    id integer primary key generated always as identity,
    uuid text unique not null
);

create table if not exists rating (
    user_id integer,
    max_score integer check(max_score >= 0) default 0 not null,
    avg_score integer check(avg_score >= 0) default 0 not null,
    foreign key (user_id) references users(id) on delete cascade
);
