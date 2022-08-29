## 目录

* asset: 资产相关链码
* fund: 资金相关链码
* goods: 商品相关链码
* order: 订单相关链码

## 链码简介

asset 和 fund 这两个链码逻辑比较简单，提供基本的增加，修改和查询的功能

方法说明如下：

* add: 增加记录，参数为字符串数组 ["key","value"]
    * 第一个参数为 id
    * 第二个参数为 存储的内容

* getById: 根据id获取内容，参数为字符串数组 ["key"]
    * 第一个参数为 id

* update: 更新记录，将对应key的值更新为新的值，参数为字符串数组 ["key","value"]
    * 第一个参数为 id
    * 第二个参数为 存储的内容

goods 和 order 为商品和订单的链码，详细方法和参数见接口文档

## 部署链码

asset 和 fund 是单文件链码，直接在 Baas 中上传文件，然后安装部署即可

order 和 goods 涉及到 couchdb 的索引，需要将各自文件夹下的文件打包成 zip 文件后上传。
<em>⚠️，以goods链码为例，进入goods文件夹，全选所有的文件，然后打包成zip，而不是将goods文件夹打包成zip<em>

由于 goods 链码依赖 order 链码，所以 部署时需要先部署 order 链码，且链码标识为 order；然后在部署
 goods 链码。调用时，只需要调用 goods 链码，goods 链码中有 API 文档的所有方法，包括 order 相关
的方法。goods和order链码需要部署在同一通道下。