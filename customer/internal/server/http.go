package server

import (
	"context"
	"customer/api/customer"
	v1 "customer/api/helloworld/v1"
	"customer/internal/biz"
	"customer/internal/conf"
	"customer/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server,
	greeter *service.GreeterService,
	customerService *service.CustomerService,
	logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			//自己的设置中间件
			selector.Server(jwt.Server(func(token *jwtv4.Token) (interface{}, error) {
				return []byte(biz.CustomerSecret), nil
			}),
				customerJWT(customerService)).Match(func(ctx context.Context, operation string) bool {
				//根据自己需要完成是否启用中间件的校验
				//log.Info(operation)
				noJWT := map[string]struct{}{"/api.customer.Customer/Login": {},
					"/api.customer.Customer/GetVerifiyCode": {}}
				if _, exists := noJWT[operation]; exists {
					return false
				}
				return true
			}).Build(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	//注册customer的http服务
	customer.RegisterCustomerHTTPServer(srv, customerService)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	return srv
}
