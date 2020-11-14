create table if not exists pay_notice
(
  gateway_order_id varchar(255) not null
    primary key,
  create_date      varchar(255) null,
  fail_times       int unsigned null,
  notice_time      varchar(255) null,
  status           varchar(255) null,
  error_message    varchar(255) null,
  next_notify_time varchar(255) null
);

create index idx_pay_notice_create_date
  on pay_notice (create_date);

create index idx_pay_notice_next_notify_time
  on pay_notice (next_notify_time);

create index idx_pay_notice_notice_time
  on pay_notice (notice_time);

create index idx_pay_notice_status
  on pay_notice (status);

create table if not exists pay_notice_ok
(
  gateway_order_id varchar(255) not null
    primary key,
  create_date      varchar(255) null,
  fail_times       int unsigned null,
  notice_time      varchar(255) null
);

create index idx_pay_notice_ok_create_date
  on pay_notice_ok (create_date);

create index idx_pay_notice_ok_notice_time
  on pay_notice_ok (notice_time);

create table if not exists pay_order
(
  out_trade_no          varchar(255) null,
  channel_account       varchar(255) null,
  channel_order_id      varchar(255) null,
  gateway_order_id      varchar(255) not null
    primary key,
  pay_amount            int unsigned null,
  currency              varchar(255) null,
  notify_url            varchar(255) null,
  return_url            varchar(255) null,
  app_id                varchar(255) null,
  sign_type             varchar(255) null,
  order_time            varchar(255) null,
  request_time          varchar(255) null,
  create_date           varchar(255) null,
  user_ip               varchar(255) null,
  user_id               varchar(255) null,
  payer_account         varchar(255) null,
  product_id            varchar(255) null,
  product_name          varchar(255) null,
  product_describe      varchar(255) null,
  callback_json         text         null,
  ext_json              text         null,
  channel_response_json longtext     null,
  error_message         text         null,
  channel_id            varchar(255) null,
  method                varchar(255) null,
  remark                text         null,
  order_status          varchar(255) null,
  constraint idx_app_id_out_trade_no
  unique (out_trade_no, app_id)
);

create index idx_pay_order_channel_id
  on pay_order (channel_id);

create index idx_pay_order_create_date
  on pay_order (create_date);

create index idx_pay_order_method
  on pay_order (method);

create index idx_pay_order_order_status
  on pay_order (order_status);

create index idx_pay_order_user_id
  on pay_order (user_id);

create table if not exists pay_order_ok
(
  out_trade_no          varchar(255) null,
  channel_account       varchar(255) null,
  channel_order_id      varchar(255) null,
  gateway_order_id      varchar(255) not null
    primary key,
  pay_amount            int unsigned null,
  currency              varchar(255) null,
  notify_url            varchar(255) null,
  return_url            varchar(255) null,
  app_id                varchar(255) null,
  sign_type             varchar(255) null,
  order_time            varchar(255) null,
  request_time          varchar(255) null,
  create_date           varchar(255) null,
  user_ip               varchar(255) null,
  user_id               varchar(255) null,
  payer_account         varchar(255) null,
  product_id            varchar(255) null,
  product_name          varchar(255) null,
  product_describe      varchar(255) null,
  callback_json         text         null,
  ext_json              text         null,
  channel_response_json longtext     null,
  error_message         text         null,
  channel_id            varchar(255) null,
  method                varchar(255) null,
  remark                text         null,
  success_time          varchar(255) null,
  balance_date          varchar(255) null,
  fare_amt              int unsigned null,
  fact_amt              int unsigned null,
  send_notice_stats     varchar(255) null,
  constraint idx_app_id_out_trade_no
  unique (out_trade_no, app_id)
);

create index idx_pay_order_ok_balance_date
  on pay_order_ok (balance_date);

create index idx_pay_order_ok_channel_id
  on pay_order_ok (channel_id);

create index idx_pay_order_ok_create_date
  on pay_order_ok (create_date);

create index idx_pay_order_ok_method
  on pay_order_ok (method);

create index idx_pay_order_ok_send_notice_stats
  on pay_order_ok (send_notice_stats);

create index idx_pay_order_ok_user_id
  on pay_order_ok (user_id);

