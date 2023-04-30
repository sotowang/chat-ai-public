```sql
create
database shiyu;

create table shiyu.users
(
    id              int auto_increment
        primary key,
    email           varchar(64)                         not null,
    password        varchar(255)                        not null,
    vip_status      tinyint   default 0 null comment 'VIP状态：0-非VIP，1-普通VIP，2-高级VIP',
    vip_expire_date int       default 0 null comment 'VIP过期时间',
    createdAt       timestamp default CURRENT_TIMESTAMP not null,
    updatedAt       timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    status          int       default 0                 not null comment '0: 可用, 1: 封禁',
    ban_time        int       default 0                 not null comment '用户封禁时间',
    constraint email
        unique (email)
);

CREATE TABLE products
(
    product_id   INT            NOT NULL AUTO_INCREMENT,
    product_name VARCHAR(255)   NOT NULL,
    description  VARCHAR(255)            DEFAULT '',
    price        DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    stock        INT            NOT NULL DEFAULT 0,
    image        VARCHAR(255)            DEFAULT '',
    created_at   DATETIME                DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME                DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (product_id)
);

create table shiyu.orders
(
    id          bigint unsigned auto_increment comment '订单ID'
        primary key,
    user_id     bigint unsigned not null,
    order_no    varchar(191)          not null,
    product_id  bigint unsigned default '0' not null,
    status      bigint      default 0 not null comment '订单状态：0，未付款 1，订单完成 2，超时 ',
    total_price double                not null,
    source      varchar(32) default '' null comment '支付来源：app',
    created_at  datetime(3) null,
    updated_at  datetime(3) null,
    deleted_at  datetime(3) null,
    pay_type    tinyint     default 0 not null comment '支付类型：0，支付宝',
    constraint order_no
        unique (order_no),
    constraint order_no_2
        unique (order_no)
) comment '订单表';

create index idx_order_no
    on shiyu.orders (order_no);

create index idx_orders_deleted_at
    on shiyu.orders (deleted_at);

create index idx_user_id
    on shiyu.orders (user_id);


CREATE TABLE verification_code
(
    id          INT PRIMARY KEY AUTO_INCREMENT,
    user_id     INT,
    code        VARCHAR(6),
    expire_time TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL 60 MINUTE),
    created_at  DATETIME  DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME  DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user (id)
);

CREATE TABLE pdf_record
(
    id         INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id    INT          NOT NULL,
    doc_id     VARCHAR(255) NOT NULL,
    doc_type   VARCHAR(255) NOT NULL,
    upload_at  DATETIME     NOT NULL,
    qa_count   INT          NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```