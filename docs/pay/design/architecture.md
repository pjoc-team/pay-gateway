# 支付系统-架构设计


<style>

span.green {
    color: #54CC76;
}
  
span.purple {
    color: #7949B3;
}
  
span.yellow {
    color: #FFCA80;
}


</style>

## 设计

### <span class="purple">目标</span>

- <span class="green">简化逻辑</span>：边界清晰、无冗余逻辑、方便测试
- <span class="green">分布式</span>：容器化部署、无状态、可扩展、低耦合
- <span class="green">高用性</span>：机房多主、自动扩容、数据库最终一致

### <span class="purple">拆分</span>

- <span class="yellow">业务</span>：代金券、活动
- <span class="yellow">支付中心</span>：商品管理、价格管理、发货路由
- <span class="green">支付网关</span>：下单、支付、代扣、通知、后督（*订单二次确认*）、对账、渠道服务

## 实现


### <span class="purple">网关</span>体系

- 支付网关：接受下单请求、校验签名、生成订单、操作DB、调用渠道服务
- 代扣网关：处理签约请求、代扣请求、互斥逻辑
- 通知网关：处理渠道通知


### 渠道服务集合

#### 原则

- 无db操作
- 无订单操作
- 只接受请求并响应对应数据
- 不关心上层逻辑

#### 微服务化

- service:
    - wechat
    - alipay
    - unionpay
- rpc:
    - grpc://wechat:8080
    - grpc://alipay:8080
    - http://unionpay:8080

> 渠道微服务化，可以在国际化业务中带来更便捷的差异化部署

#### 面向<span class="purple">协议</span>编程

- pay(用户触发扣款或签约)
    - method：alipay/wechat/unionpay
    - from: WEB/MWEB/APP/SDK/BANK/SERVER
    - type: pay/sign_pay
- confirm_pay(确认支付，用户二次验证)
    - from
    - type
    - method
- refund(退款)
- transfer(转账)

#### <span class="purple">扩展</span>&配置

- etcd配置管理
    - 高可用
    - 实时变更配置
    - 跨地域同步
- 简化配置
    - 使用field tag生成json
    - 管理后台根据field json生成配置表单


```plantuml  format="svg_inline"

!includeurl https://raw.githubusercontent.com/blademainer/plantuml-style-c4/master/c4_component.puml


'LAYOUT_WITH_LEGEND

title System Context diagram for Pay System
Actor(customer, "Customer", "User")

Enterprise_Boundary("company", "Company"){
    System_Ext(biz_system, "Business System", "Allows customers to view information about their info and orders.")

    System(pay_center_system, "Pay center system", "The system for biz.")
    System(pay_gateway_system, "Pay gateway System", "The system for user to pay or auto_renew.")
}

System_Ext(channel_system, "Channel Pay System", "The real pay system to handle users' pay orders.")

Rel_D(customer, biz_system, "Show products.")
Rel_Neighbor(customer, pay_center_system, "Pay request")
Rel_D(pay_center_system, pay_gateway_system, "Order request")
Rel_U(pay_gateway_system, channel_system, "Order request")

Rel_Neighbor(customer, channel_system, "Pay")
channel_system .> pay_gateway_system: Notify

pay_gateway_system .> pay_center_system: Notify
pay_center_system .> biz_system: Notify

```


### 支付网关系统

```plantuml format="svg_inline"
!includeurl https://raw.githubusercontent.com/blademainer/plantuml-style-c4/master/c4_container.puml

System_Ext(pc, "Pay center", "For biz")
System_Ext(ch, "Pay Channels", "wechat/alipay/unionpay...")

System_Boundary(sb, "Pay Gateway"){
  Container(ag, "Api Gateway", "Service discovery, Canary release, Parameters convert... Implements by ingress/zuul")
  
  Container(pg, "Pay Service", "Generate order")
  Container(sg, "Sign Service", "For Monthly/Quarterly/Yearly sign and renew")
  Container(cg, "Callback Service", "Receive callback from channels.")
  Container(nf, "Notify Service", "Notify to biz.")
  Container(os, "Order Supervision", "Query order status.")
  
  
  Container(cs, "Channel Services", "Deliver pay requests between gateways and channels.")
  Container(ds, "DB Service", "Db operations.")
  
  Container(mq, "Queue", "Message queue.")
  ContainerDb(db, "Database", "mysql,middlewares")
  Container(etcd, "etcd", "Config system.")
  
  ag --> pg
  ag --> sg
  ag --> cg
  
  pg --> ds
  pg --> cs
  sg --> ds
  sg --> cs
  cg --> ds
  cg --> cs
  cg --> mq: Push message
  os --> db: Query orders
  'os --> mq: Secondary confirmation
  mq --> nf: Pull message
  
  ds --> db
  db -[hidden]r- etcd
}

pc -L-> ag
cs --> ch

```


### 支付调用时序图
```plantuml format="svg_inline"
!includeurl https://raw.githubusercontent.com/blademainer/plantuml-style-c4/master/c4_component.puml

'skinparam monochrome true

actor User
participant "Channel" as E
participant "PayCenter(Biz)" as A
box "Internal Service"
participant "PayGateway" as B
participant "CallbackGateway" as F
participant "PayChannels" as G
participant "NotifyGateway" as H
participant "PayDatabase" as C
database "MySQL" as D
control "Queue" AS I
end box

User -> A: Page
activate A

A -> B: Create Order
activate B
B -> B: Verify
B -> B: Generate Order

B -> C: Save Order
activate C
C -> D: SQL
activate D
D -> C: OK
deactivate D
C -> B: OK
deactivate C

B -> G: Order Request
activate G
G -> E: Order Request
G -> B: Response Data
deactivate G

B -> A: Response
deactivate B

A -> User: Show QrCode or redirect url
deactivate A

...

User -> E: Pay
E --> F: Notify
activate F
F -> G: Notify
activate G
G -> G: Verify
G -> F: OK
deactivate G
F -> C: Update
activate C
C -> D: SQL
activate D
D -> C: OK
deactivate D
C -> F: OK
deactivate C
F -> H: Process result
activate H
H -> I: push
H -> F: OK
deactivate H
F --> E: OK
E -> A: Redirect Url
A -> User: Show PayResult Page


loop
    activate H
    H -> I: pull
    H --> A: Notify
    alt notify success
      A --> H: OK
      H -> C: Update notify status
      C -> H: OK
    else notify failed
      loop 10 times
        H --> A: Notify
      end
    end
    deactivate H
end
```


### 签约

```plantuml format="svg_inline"
!includeurl https://raw.githubusercontent.com/blademainer/plantuml-style-c4/master/c4_container.puml

actor "User" as u
participant "Channel" as ch
participant "VirtualAssetsSystem" as vas
participant "PayCenter(Biz)" as pc
box "Internal Service"
participant "PayService" as pg
participant "SignService" as spg
participant "CallbackService" as cg
participant "SignCallbackService" as sng
participant "ChannelServices" as pcs
participant "PayDatabase" as db
end box

autonumber

u -> pc: Sign request
pc -> spg ++: Sign(app_id, uid, product_id, amount)
spg -> db ++: IsExists(uid, app_id, product_id, channel)
db --> spg --: OK
spg -> pcs ++: Sign request
pcs -> pcs: Generate sign
pcs -> ch: Sign request
pcs --> spg --: Response
spg --> pc --: Response data
pc --> u --: Show QR or direct to url

...

u -> ch: Confirm

ch -> sng ++: Notify
sng -> db ++: Query sign record
db --> sng --: OK
sng -> pcs ++: Verify
pcs --> sng --: OK
sng -> db ++: Update status
db --> sng --: OK
sng --> ch --: OK

ch --> u: Direct

vas -> vas: Check expire
vas -> pc ++: Expire events(uid, product_id)
pc -> spg ++: Renew(uid, sign_id)
spg -> db ++: Query sign record
db --> spg --: OK
spg --> pg ++: Pay
pg -> db ++: Generate order
db --> pg --: OK
pg -> pcs ++: Order request
pcs -> pcs: Generate message
pcs -> ch: Order request
pcs --> pg --: OK
pg --> spg --: OK
spg --> pc --: OK
deactivate pc

ch --> cg ++: Notify
cg -> db: Query order
cg -> pcs: Verify
cg -> pc ++: Notify
pc -> vas: Notify
pc --> cg --: ok
cg --> ch --: ok
```

### 数据库<span class="purple">高可用</span>

- 跨DC同步：基于otter进行同步，双向同步（多机房使用星型结构）
- 同DC高可用：基于mycat和mgr，实现大容量、高可用db集群
- <span class="green">mgr</span>的心跳检测：二次开发mycat，对mgr节点状态实时检测并增删故障db
- 应用层：去除自增主键，按机房、机器生成无冲突、有序的流水号，防止多机房数据冲突

```plantuml format="svg_inline"
!includeurl https://raw.githubusercontent.com/blademainer/plantuml-style-c4/master/c4_container.puml

Boundary(a, "idc A (Master)"){
  Boundary(ka, "k8s cluster"){
    System(pa, "Pay Gateway"){
      Container(paa, "Apps", "Gateways")
      Boundary(chs, "Channel Services"){
        Container(ch1, "Channel Wechat", "Channel service")
        Container(ch2, "Channel Alipay", "Channel service")
        Container(chx, "Channel ...", "Channel service")
      }
    }
    System_Ext(ma, "Mycat")
  }
  SystemDb(dba, "DB"){
    ContainerDb(dba1, "db", "Group replication")
    ContainerDb(dba2, "db", "Group replication")
    ContainerDb(dba3, "db", "Group replication")
  }
  
  System(ota, "Otter", "Sync data")
  
  
  'ch1 -[hidden]D- ch2
  'ch2 -[hidden]D- chx
  paa --> ma
  paa -U-> ch1
  paa -U-> ch2
  paa -U-> chx
  
  ma -D-> dba1: W/R
  ma -D-> dba2: W/R
  ma -D-> dba3: W/R
  
  dba1 <-> dba2: Replication
  dba2 <-> dba3: Replication
  
  dba1 --> ota: binlog
  dba2 ..> ota: binlog
  dba3 ..> ota: binlog
}

Boundary(b, "idc B"){
  Boundary(kb, "k8s cluster"){
    System(pb, "Pay Gateway"){
      Container(pab, "Apps", "Gateways")
      Boundary(chbs, "Channel Services"){
        Container(chb1, "Channel Wechat", "Channel service")
        Container(chb2, "Channel Alipay", "Channel service")
        Container(chbx, "Channel ...", "Channel service")
      }
    }
    System_Ext(mb, "Mycat")
  }
  SystemDb(dbb, "DB"){
    ContainerDb(dbb1, "db", "Group replication")
    ContainerDb(dbb2, "db", "Group replication")
    ContainerDb(dbb3, "db", "Group replication")
  }
  
  System(otb, "Otter", "Sync data")
  
  pab --> mb
  pab -U-> chb1
  pab -U-> chb2
  pab -U-> chbx
  
  mb -D-> dbb1: W/R
  mb -D-> dbb2: W/R
  mb -D-> dbb3: W/R
  
  dbb1 <-> dbb2: Replication
  dbb2 <-> dbb3: Replication
  
  dbb1 --> otb: binlog
  dbb2 ..> otb: binlog
  dbb3 ..> otb: binlog
}

ota <.> otb: sync
```


### 部署架构

```plantuml format="svg_inline"
!includeurl https://raw.githubusercontent.com/blademainer/plantuml-style-c4/master/c4_component.puml

' LAYOUT_TOP_DOWN

' define
cloud Kubernetes{
'  package Core {
    node PayGateway
    node QueryGateway
    node RefundGateway
    node TransferGateway
    node CallbackGateway
    node NotifyGateway
    node PayDatabase
'  }

  package Channels as c {
    node Channel... as channels
    node ChannelWechat
    node ChannelAlipay
  }

  node PayManagerSystem
  node OrderMonitor
  node PayCenter

}

database mysql

'cloud EtcdCloud{
'  storage etcd1
'  storage etcd2
'  storage etcd3
'}

' === relations ===
'PayCenter ..> EtcdCloud
'PayGateway ..> EtcdCloud
'QueryGateway ..> EtcdCloud
'RefundGateway ..> EtcdCloud
'TransferGateway ..> EtcdCloud
'PayDatabase ..> EtcdCloud
'ChannelWechat ..> EtcdCloud
'ChannelAlipay ..> EtcdCloud
'Channel... ..> EtcdCloud

PayCenter -left-> PayGateway

PayGateway --> PayDatabase
PayGateway .up.> channels

QueryGateway --> PayDatabase
QueryGateway .> channels

RefundGateway --> PayDatabase
RefundGateway .> channels

CallbackGateway --> PayDatabase
CallbackGateway .> channels
CallbackGateway -> NotifyGateway


TransferGateway --> PayDatabase
TransferGateway .> channels

PayManagerSystem -> PayDatabase
PayManagerSystem -> QueryGateway

OrderMonitor -> PayDatabase
OrderMonitor -> QueryGateway


PayDatabase --> mysql

```

## 交互
### 配置
核心交易系统是将配置信息存储在`etcd`容器内
#### 渠道
- 基础目录: `/foo/bar/pay/config`
- 每个渠道占用一个文件夹，每个渠道账户占用一个`文件`，例如微信存放在`/foo/bar/pay/config/wechat`目录下，appId: 2088123456 所在的配置信息存储在`/foo/bar/pay/config/wechat/2088123456`

