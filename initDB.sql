create table delivery (
                         id	bigserial not null primary key,
                         Name  varchar(128),
                         Phone  varchar(128),
                         Zip  varchar(128),
                         City  varchar(128),
                         Address varchar(128),
                         Region varchar(128),
                         Email  varchar(128)
);

create table payment (
                         id	bigserial not null primary key,
                         Transaction  varchar(128),
                         Request_Id varchar(128),
                         Currency     varchar(128),
                         Provider     varchar(128),
                         Amount       int    ,
                         Payment_Dt    int  ,
                         Bank         varchar(128),
                         Delivery_Cost int   ,
                         Goods_Total   int,
                         Custom_Fee int
);

create table items (
                       id	bigserial not null primary key,
                       Chrt_Id     int,
                       Track_Number varchar(128),
                       Price      int,
                       Rid        varchar(256),
                       Name       varchar(128),
                       Sale       int,
                       Size       varchar(128),
                       Total_Price int,
                       Nm_Id       int,
                       Brand      varchar(128),
                       Status int
);

create table orders (
                          id	bigserial not null primary key,
                          Order_Uid          varchar(128),
                          Track_Number       varchar(128),
                          Entry             varchar(128),
                          Locale            varchar(128),
                          Internal_Signature varchar(128),
                          Customer_Id        varchar(128),
                          Delivery_Service   varchar(128),
                          Shardkey          varchar(128),
                          Sm_Id              int,
                          Date_created varchar(128),
                          Oof_shard varchar(128),
                          delivery_fkey bigserial references delivery(id),
                          payment_fkey bigserial references payment(id)
);

create table "order_items" (
                               id	bigserial not null primary key,
                               order_id_fk        bigserial references orders(id),
                               item_id_fk         bigserial  references items(id)
);

create table "cache" (
                         id	bigserial not null primary key,
                         order_id	int8,
                         app_key        varchar(128)
);
