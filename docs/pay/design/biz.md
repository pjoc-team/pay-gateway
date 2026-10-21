# 名词
- 账户余额： 作为收款账户，包括可结算金额和带结算金额
- 不可用余额： 指的是当前收款账户中不可用的金额，如正在提现中的金额都是不可用余额
- 可结算金额： 收款账户已经结算的金额，已结算的金额可提现
- 保证金： 指商户开户费账户
- 结金额： 收款账户的冻结金额，冻结的金额不可提现
- 付余额： 作为代付账户，商户通过在线充值通道充值的金额会转入该账户(实时到账)，发起代付会扣减该账户
- 可用代付余额： 指的是当前代付账户中不可用的金额，如正在代付处理中的金额都是不可用代付余额

# 各个`系统/模块`功能
## PayCenter
支付中心，用于演示支付功能。用户访问的时候可以展示支付页，然后选择支付方式来体验支付流程

## PayGateway
支付网关，用于处理下单请求。功能有：

- 校验参数
- 验证商户、签名
- 保存订单
- 分发请求到渠道
- 返回响应

## PayDatabaseService
支付数据库服务。提供订单保存、移表操作等。

## ChannelXX
渠道微服务。每个渠道（渠道服务商）都会单独起一个微服务，用于隔离服务之间的竞争，也方便各个渠道按照负载进行不同的部署。  
例如：微信支付相关的渠道都放到`channel-wechat`，然后所有相关的`微信扫码`、`微信sdk`等支付方式都在同个服务内，不同的支付方式都通过`method`来区分。
渠道实现的功能有：

- 处理支付请求
    - 构造渠道请求
    - 请求渠道
    - 返回响应
- 校验回调请求
    - 验证签名
    - 返回结果
    
## CallbackGateway
回调网关，用于处理渠道的通知。功能有：
 
- 验证参数
- 查询订单
- 分发请求到渠道服务
- 发服务
- 返回响应报文（转换渠道微服务的响应然后再返回给）

## NotifyGateway(SettlementGateway)
通知、结算网关，用于通知商户支付结果。功能有：

- 查询订单状态
- 更新订单为成功
- 发送通知
- 轮询失败通知，重试发送

## PayManagerSystem
总体分三大系统：

- `运营管理系统(operation manage system)` https://github.com/pjoc-team/operation-manage-system
- `代理商管理系统(proxy manage system)` https://github.com/pjoc-team/proxy-management-system
- `商户管理系统(merchant manage system)` https://github.com/pjoc-team/merchant-manage-system

### 运营管理系统
具有管理代理商、商户、支付、转账、权限管理等所有功能，也是功能最复杂的系统

### 代理商管理系统
- `代理商`是指可以推广支付接口并设置下级商户的用户。
- `代理商管理系统`可以查看、新增自己的下级商户，查看流水、结算信息等

### 商户管理系统
- `商户`是真实使用支付功能的用户，一个商户可以创建自己的应用，然后可以使用该应用来调用支付接口下单。
- `商户管理系统`提供管理应用、查看流水、结算等功能

# 费率计算逻辑
假如用户支付一笔订单，则计算支付系统各方的利润方法如下：
## 定义参数
- 费率(Rate)(单位：%)
    - R<sub>c</sub>: rate of channel 渠道费率（即微信、支付宝的费率）
    - R<sub>a</sub>: rate of agent 代理商费率
    - R<sub>m</sub>: rate of merchant 商户费率
- 利润(Profit)
    - P<sub>p</sub>: profit of platform 平台利润
    - P<sub>a</sub>: profit of agent 代理商利润
- 订单金额: A

## 限制

- R<sub>c</sub> <= R<sub>a</sub> <= R<sub>m</sub> < 100

## 计算
$$
P_p=\frac{(R_a-R_c) \times A}{100}
$$

$$
P_a=\frac{(R_m-R_a) \times A}{100}
$$

多级代理计算方法（P<sub>a1</sub>是P<sub>a2</sub>的上级代理）

$$
P_{a1}=\frac{(R_{a2}-R_{a1}) \times A}{100}
$$
