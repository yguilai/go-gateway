# go-gateway

 #### http服务
 1. 服务启动时从DB加载配置信息
 2. 反向代理以中间件形式, 注册到中间件最后一级
 3. ws、请求超时控制、限流、白名单限制、前缀匹配、https、strip_uri、重写url、head头转换均以中间件形式注册到服务
 
 #### tcp服务、grpc服务
 限流、白名单限制
 
 #### 多租户
 1. 独立接口获取jwt token
 2. jwt中间件校验token
 
 