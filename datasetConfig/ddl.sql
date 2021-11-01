create table nns.dataset_config
(
    id bigint auto_increment
        primary key,
    project_id bigint not null,
    dataset_id bigint not null,
    name varchar(200) not null,
    shuffle tinyint(1) not null,
    label varchar(1024) not null,
    normalization_method varchar(1024) null,
    status varchar(10) not null,
    create_time datetime default CURRENT_TIMESTAMP not null,
    update_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
);

create index dataset_config__index_project_id
    on nns.dataset_config (project_id);

