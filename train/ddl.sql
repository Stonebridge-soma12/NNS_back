create table train
(
    id bigint auto_increment comment 'id'
        primary key,
    user_id bigint not null,
    train_no bigint not null,
    project_id bigint not null,
    acc float null comment 'acc',
    loss float null comment 'loss',
    val_acc float null comment 'val_acc',
    val_loss float null comment 'val_loss',
    name varchar(45) null comment 'name',
    epochs int default 0 null,
    result_url text null,
    status varchar(10) null,
    constraint train_uk_user_id_train_no
        unique (user_id, train_no),
    constraint train_ibfk_1
        foreign key (project_id) references project (id)
            on delete cascade
);

create index train_id
	on train (project_id);


create table train_config
(
    id bigint auto_increment
        primary key,
    train_id bigint not null,
    train_dataset_url varchar(1024) not null,
    valid_dataset_url varchar(1024) null,
    dataset_shuffle tinyint(1) not null,
    dataset_label varchar(512) not null,
    dataset_normalization_usage tinyint(1) not null,
    dataset_normalization_method varchar(512) null,
    model_content json not null,
    model_config json not null,
    create_time datetime default CURRENT_TIMESTAMP not null,
    update_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
);


create table train_log
(
    id int auto_increment
        primary key,
    train_id bigint null,
    msg text null,
    status_code int null,
    create_time datetime default CURRENT_TIMESTAMP not null,
    update_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint train_log_ibfk_1
        foreign key (train_id) references train (id)
            on update cascade
);

create index train_id
	on train_log (train_id);



create table epoch
(
    id bigint auto_increment comment 'id'
        primary key,
    train_id bigint null comment 'train_id',
    epoch int null comment 'epoch',
    acc float null comment 'acc',
    loss float null comment 'loss',
    val_acc float null comment 'val_acc',
    val_loss float null comment 'val_loss',
    learning_rate float null comment 'lr',
    create_time datetime default CURRENT_TIMESTAMP not null,
    update_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
);

create index train_id
	on epoch (train_id);

