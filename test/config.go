package main

//修改mysql对应的用户名、密码和数据库，格式如下：
//dsn = "user:password@tcp(127.0.0.1:3306)/database?charset=utf8mb4&parseTime=True&loc=Local"
const mysqlInfo = "catchyou:123456@tcp(49.234.79.216:3306)/catchYou?charset=utf8mb4&parseTime=True&loc=Local" // mysql信息
const redisUser, redisPsd, redisIP = "default", "", "222.20.104.39:6379"                                      // redis信息
const redisExpire = 180				                                                                          // redis记录每次延期时间(秒)
const chanBuf = 100                                                                                           // chan的缓冲区大小
const vcodeValid = 600                                                                                        // 验证码有效时间(秒)
const rLimiter, bLimiter = 1, 5                                                                               // 令牌桶的参数r和b
const Base64Table = "IJjkKLMNO567PQX12RVW3YZaDEFGbcdefghiABCHlSTUmnopqrxyz04stuvw89+/"                        // 加密原表
