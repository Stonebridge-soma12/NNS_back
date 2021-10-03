create table train
(
    id bigint auto_increment comment 'id'
        primary key,
    user_id bigint null,
    train_no bigint null,
    project_id bigint null,
    acc float null comment 'acc',
    loss float null comment 'loss',
    val_acc float null comment 'val_acc',
    val_loss float null comment 'val_loss',
    name varchar(45) null comment 'name',
    epochs int default 0 null,
    url text null,
    status varchar(10) null,
    constraint train_uk_user_id_train_no
        unique (user_id, train_no),
    constraint train_ibfk_1
        foreign key (project_id) references project (id)
            on delete cascade
);

create index train_id
    on train (project_id);

