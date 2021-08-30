create table image
(
    id bigint auto_increment
        primary key,
    user_id bigint not null,
    url varchar(1024) not null,
    create_time datetime default current_timestamp() not null,
    update_time datetime default current_timestamp() not null on update current_timestamp()
);

create table project
(
    id bigint auto_increment
        primary key,
    user_id bigint not null,
    project_no int not null,
    name varchar(45) not null,
    description varchar(2000) not null,
    config longtext collate utf8mb4_bin null,
    content longtext collate utf8mb4_bin null,
    status varchar(10) default 'EXIST' not null,
    create_time datetime default current_timestamp() not null,
    update_time datetime default current_timestamp() not null on update current_timestamp(),
    share_key varchar(100) null,
    constraint project_uk_user_id_project_no
        unique (project_no, user_id),
    constraint config
        check (json_valid(`config`)),
    constraint content
        check (json_valid(`content`))
);

create index project__index_user_id
    on project (user_id);



create index project__index_user_id
    on project (user_id);

create table user
(
    id bigint auto_increment
        primary key,
    name varchar(45) default '익명' not null,
    profile_image bigint null,
    description varchar(200) null,
    email varchar(320) null,
    web_site varchar(1024) null,
    login_id varchar(50) null,
    login_pw binary(60) null,
    status varchar(10) default 'EXIST' not null,
    create_time datetime default current_timestamp() not null,
    update_time datetime default current_timestamp() not null on update current_timestamp()
);

create index user_login_id_index
    on user (login_id);

