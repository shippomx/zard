# go-zero

## 0. go-zero 介绍

go-zero（收录于 CNCF 云原生技术全景图：[https://landscape.cncf.io/?selected=go-zero](https://landscape.cncf.io/?selected=go-zero)）是一个集成了各种工程实践的 web 和 rpc 框架。通过弹性设计保障了大并发服务端的稳定性，经受了充分的实战检验。

go-zero 包含极简的 API 定义和生成工具 goctl，可以根据定义的 api 文件一键生成 Go, iOS, Android, Kotlin, Dart, TypeScript, JavaScript 代码，并可直接运行。

使用 go-zero 的好处：

* 轻松获得支撑千万日活服务的稳定性
* 内建级联超时控制、限流、自适应熔断、自适应降载等微服务治理能力，无需配置和额外代码
* 微服务治理中间件可无缝集成到其它现有框架使用
* 极简的 API 描述，一键生成各端代码
* 自动校验客户端请求参数合法性
* 大量微服务治理和并发工具包

![架构图](https://raw.githubusercontent.com/zeromicro/zero-doc/main/doc/images/architecture.png)

## 1. go-zero 框架背景

18 年初，我们决定从 `Java+MongoDB` 的单体架构迁移到微服务架构，经过仔细思考和对比，我们决定：

* 基于 Go 语言
    * 高效的性能
    * 简洁的语法
    * 广泛验证的工程效率
    * 极致的部署体验
    * 极低的服务端资源成本
* 自研微服务框架
    * 有过很多微服务框架自研经验
    * 需要有更快速的问题定位能力
    * 更便捷的增加新特性

## 2. go-zero 框架设计思考

对于微服务框架的设计，我们期望保障微服务稳定性的同时，也要特别注重研发效率。所以设计之初，我们就有如下一些准则：

* 保持简单，第一原则
* 弹性设计，面向故障编程
* 工具大于约定和文档
* 高可用、高并发、易扩展
* 对业务开发友好，封装复杂度
* 约束做一件事只有一种方式


## 3. go-zero 项目实现和特点

go-zero 是一个集成了各种工程实践的包含 web 和 rpc 框架，有如下主要特点：

* 强大的工具支持，尽可能少的代码编写
* 极简的接口
* 完全兼容 net/http
* 支持中间件，方便扩展
* 高性能
* 面向故障编程，弹性设计
* 内建服务发现、负载均衡
* 内建限流、熔断、降载，且自动触发，自动恢复
* API 参数自动校验
* 超时级联控制
* 自动缓存控制
* 链路跟踪、统计报警等
* 高并发支撑，稳定保障了疫情期间每天的流量洪峰

如下图，我们从多个层面保障了整体服务的高可用：

![弹性设计](https://raw.githubusercontent.com/zeromicro/zero-doc/main/doc/images/resilience.jpg)

## 4. 我们使用 go-zero 的基本架构图

<img width="1067" alt="image" src="https://user-images.githubusercontent.com/1918356/171880582-11a86658-41c3-466c-95e7-7b1220eecc52.png">


## 5. Benchmark

![benchmark](https://raw.githubusercontent.com/zeromicro/zero-doc/main/doc/images/benchmark.png)

[测试代码见这里](https://github.com/smallnest/go-web-framework-benchmark)

## 6. 文档

* API 文档

  [https://go-zero.dev/cn/](https://go-zero.dev/cn/)

* awesome 系列（更多文章见『微服务实践』公众号）

    * [快速构建高并发微服务](https://github.com/zeromicro/zero-doc/blob/main/doc/shorturl.md)
    * [快速构建高并发微服务 - 多 RPC 版](https://github.com/zeromicro/zero-doc/blob/main/docs/zero/bookstore.md)
    * [goctl 使用帮助](https://github.com/zeromicro/zero-doc/blob/main/doc/goctl.md)
    * [Examples](https://github.com/zeromicro/zero-examples)

* 精选 `goctl` 插件

    * [goctl-swagger](https://github.com/zeromicro/goctl-swagger)。 一键生成 `api` 的 `swagger` 文档
    * [goctl-android](https://github.com/zeromicro/goctl-android)。生成 `java (android)` 端 `http client` 请求代码
    * [goctl-go-compact](https://github.com/zeromicro/goctl-go-compact)。 合并 `api` 里同一个 `group` 里的 `handler` 到一个 `go` 文件

## 7. CNCF 云原生技术全景图
go-zero 收录在 [CNCF Cloud Native 云原生技术全景图](https://landscape.cncf.io/?selected=go-zero)。