# AdsExposureAttribution-广告归因自动化方案


## 一、 合约曝光归因介绍

### 1.1 曝光归因需求背景
  1. 用户观看了品牌广告，可能并未发生点击，但影响到用户心智，并最终直接形成了转化，对于这类行为路径，传统的点击归因无法科学评估平台对用户价值。
  2. 曝光归因可以帮助广告主更合理的进行投放归因分析，从而提供营销效率和效果。

### 1.2 曝光归因自动化方案
  为了覆盖不同类型客户的多种目标的归因需求，如不同客户侧的用户ID标识类型差异巨大，不同客户关心的归因目标也不尽相同；AMS侧提出了一套自助化化进行曝光归因的方案：融合多种场景下的归因目标和数据转化流程，无须针对特定的归因场景单独开发归因方案，降低人工对接的成本，提升效率。方案可以自动化的对合约曝光归因系统进行数据分析以及指标输出：
     
  1. 标识用户的ID类型不同（QAID/IMEI/IDFA/wx openid/QQ openid），需要进行多轮和多种类型的ID转换。
     
  2. 用户的属性不同（4A用户/转化用户），需要区分不同的优化目标或转化类型。
     
  3. 接入方式的不同（MI系统/HTTP服务/文件），不同广告主有不同的接入需求。
     
  4. 加密方式不同（加密回传/不加密回传），需要支持双线加密，保证数据安全。

## 二、 合约曝光归因自动化方案框架

  ![AdsAttribution Architecture][1]

  整体方案分四层：

  1. 数据接入层：用于接入不同客户不同类型的转化数据，并将数据结构化。
   
  2. 数据加工层：将加密传输的转化用户ID进行解密，并进行ID映射到指定用户标识ID。

  3. 数据存储层：将结构化之后的转化数据进行存储，包括但不限于Druid，ClickHouse，方便进行后续的计算。
   
  4. 数据应用层：将转化数据和曝光数据以（广告id，用户id）为主键进行曝光归因计算，可以输出多维度（地域，性别，年龄，实验号）数据进行分析对比。

## 三、 合约曝光自动归因方案流程
  对于对数据安全有较高要求的广告主，AMS提出基于两方数据双向加密的归因匹配方案，通过AMS和客户侧的多轮数据交互，在不泄露客户侧转化明文数据的情况下归因曝光的转化效果，为后续更深层次的数据分析应用提供基础。同时，本双向加密方案属于开源项目，可以节省广告主大量开发时间，直接使用。另外，对流程不熟悉的广告主，AMS提供全程指导和支持。

### 3.1 方案简介：
  整体方案交互示意图：

  ![AdsAttribution FlowChartPSI][2]

  1. 所有交互数据均进行加密处理，有效避免数据信息泄露的安全风险：
   
    a. 双方数据在线上交互前均已做加密处理

    b. 避免客户侧全量转化数据回传，仅加密回传匹配部分的转化数据
  
  2. 项目代码开源，逻辑透明：https://github.com/PTPOpenAds/AdsAttribution
  
  3. 广告主通过回传转化数据优化后续投放：

	  a. 回传数据类型：广告主回传数据包括用户id（设备号、wx-openid等，支持加密和非加密方式）、转化类型（提交表单、试用、购买等）、购买物品/服务名称、消费门店、消费时间等。

	  b. 回传数据价值和用途：广告主回传的商品信息在归因分析和合约广告效果优化中体现：在归因分析中，会根据用户购买商品类别、购买地域进行针对性分析，细分用户兴趣维度，时间维度上，我们可以观测用户在广告曝光后对品牌的心智变化。在合约广告投放效果优化中，我们会将用户转化相关的数据加入到投放算法的机器学习模型中，帮助模型精准识别对目标商品有兴趣的人群，提升广告投放效果。

	  c. 安全性分析：AMS侧保证客户侧回传数据的安全性且不可被破解转化为明文，保证客户侧的数据安全

### 3.2 整体流程介绍：
  1. 曝光归因需要AMS和广告主双方共同完成所有步骤（1-11）。
   
  2. 最终上线方案：
   
		a. 广告主侧提供可由外网访问的网络策略（IP/域名）

		b. 广告主侧提供linux环境部署AMS提供的Go二进制即可

		c. 具体联调过程参见：https://docs.qq.com/doc/DRFJuS2NMUWp4VWJP?_t=1614843855938

### 3.3 方案详细流程：

#### 3.3.1 整体方案流程示意图：

  ![AdsAttribution ClientDeploy][3]

#### 3.3.2 具体交互流程和数据协议：[用户标识以wx_openid举例， 加粗部分为该部分输出结果]

  1. 曝光数据抽取:  [wuid]
	  
    a. 广告主给定推广计划id，以及关联排期ID
	  
    b. 依据广告排期抽取曝光日志：【wuid, pingTime, adgroupId, ...】
  
  2. wuid转wx_openId：[imp_wx_opnenid]
	  
    a. 获取广告主所有小程序id以及（1）中曝光wuid，并根据广告主所提供app_id, 转化为wx_openid
  
  3. AMS曝光用户加密：【f(imp_wx_openid)】
	  
    a. Encrypt(groupId string, data string) -> string 
		    i. groupId -> GetEncryptKey, 通过groupId 管理不同广告主对应的私钥
    b. 将(2)中wx_openid进行加密，获得加密后的f_imp_wx_openid
  
  3->4: AMS将加密后的曝光ID通过http请求传输到广告主
    
    a. 广告主请求AMS，传输campign_id以及对应app_id, AMS返回状态以及加密用的大数。


广告主请求AMS /campign_recieve
```go    
请求：
{
    "promotion_id" : 123456 //广告主一次归因活动
    "account_id": 12345 //广告主ID
    "campaign_id_list": [
        123456,
        456789,
    ]
    "trace_id":123456,  //唯一表示一次请求
    "app_id_list":[
        "wx1",
        "wx2",
        "wx3"
    ]
}
返回：
{
    "code":0,
    "message":"success",
    "big_int":"ABCDEF",
    "uniq_id": "u123456_123456"  //AMS，广告主共同唯一标识一次归因
}
```

AMS请求广告主，将计算好的曝光传输到广告主 /post_imp_data
```go
请求：
{
    "trace_id":123456,
    "uniq_id":"u123456_123456",
    "imp_data":[
        "f(imp_wx_openid1)",
        "f(imp_wx_openid2)",
        "f(imp_wx_openid3)",
    ],
}
返回：
{
    "code":0,
    "message":"success",
}
```

  4. 广告主二次加密: [g(f(imp_wx_openid))]
	  
    a. Encrypt(data *big.Int, e *big.Int, prime *big.Int )
	  
    b. 将从AMS得到的wx_openid进行二次加密
  
  4->5: 广告主将二次加密用户ID传输给AMS侧

```go
请求AMS  /post_encrypted_id
{
    "uniq_id":"u123456_123456",
    "trace_id":123456,
    "imp_data":[
        "g(f(imp_wx_openid1))",
        "g(f(imp_wx_openid2))",
        "g(f(imp_wx_openid3))",
    ],
   "conv_data":[
        "g(conv_wx_openid1)",
        "g(conv_wx_openid2)",
        "g(conv_wx_openid3)",
    ],
}
返回
{
    "code":0,
    "message":"success",
}
```

  5. AMS 将（4）中数据解密: [g(imp_wx_openid)]
	  
    a. Decrypt(groupId string, data string)
  
  6. 广告主将转化数据进行加密：[g(conv_wx_openid)]
	  
    a. Decrypt(data *big.Int, e *big.Int, prime *big.Int )
  
  6->7: 广告主将加密的转化用户ID传输给AMS侧

```go
请求AMS  /post_encrypted_id
{
    "uniq_id":"u123456_123456",
    "trace_id":123456,
    "imp_data":[
        "g(f(imp_wx_openid1))",
        "g(f(imp_wx_openid2))",
        "g(f(imp_wx_openid3))",
    ],
   "conv_data":[
        "g(conv_wx_openid1)",
        "g(conv_wx_openid2)",
        "g(conv_wx_openid3)",
    ],
}
返回
{
    "code":0,
    "message":"success",
}
```

  8. AMS将（5）和（6）中数据进行求交：[g(attribution_wx_openid)]
	  
    a. 将（5）解密的数据以及广告主回传加密转化数据进行求交。
  
  7->8: AMS将交集部分数据传输到广告主侧：

```go
将交集去重以及非去重数据传输广告主 /attribution_result
{
    "uniq_id":"u123456_123456",
    "trace_id":123456,
    "data":[
        "g(join_wx_openid1)",
        "g(join_wx_openid2)",
        "g(join_wx_openid3)",
    ],
    "data_uniq":[
        "g(join_wx_openid1)",
        "g(join_wx_openid2)",
    ]
}
返回：
{
    "code":0,
    "message":"success",
}
```

  8. 广告主解密：[attribution_wx_openid]
	  
    a. Decrypt(data *big.Int, e *big.Int, prime *big.Int )
	  
    b. 将上一步得到的交集数据进行解密。
  
  9. 广告主拼接转化特征数据
	  
    a. 转化类型（表单提交、试用、购买等）、购买物品/服务名称，消费门店，消费时间等
  
  9->10: 广告主将解密后的明文数据拼接上转化明细特征，传输给AMS侧

```go
/feature_data
{
    "uniq_id":"u123456_123456",
    "trace_id":123456,
    "data":[
        {
        "wx_openid": "xdfsdf",
        "conv_type": "BUY",
        "product" : "Macdonalds",
        "action_time": "YYYMMDDHHmmss"
        }
    ]
}
返回：
{
    "code":0,
    "message":"success",
}
```

  10.  AMS接收（9）中数据
  
  11. 应用归因结果优化后续投放

## 四、 客户侧接入成本分析
  1. AMS维护开源项目，客户侧直接使用或者参与协同
  2. 客户侧只需要传入转化数据，开发量少，转化数据加密传输，无安全风险
  3. 开源监督、广告主自部署保证数据安全性
  4. 全量转化数据不出库，仅回传匹配成功的转化数据
   
   ![AdsAttribution ClientDeploy][4]

[1]: docs/readme-images/readme-arch1-4layer.png
[2]: docs/readme-images/readme-flowchart-psi.png
[3]: docs/readme-images/readme-flowchart-main.png
[4]: docs/readme-images/readme-arch2-end.png

