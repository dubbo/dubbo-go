module github.com/apache/dubbo-go/test/integrate/dubbo/go-server

go 1.13

require (
	dubbo.apache.org/dubbogo/v3 v3.0.0-00010101000000-000000000000
	github.com/apache/dubbo-go-hessian2 v1.9.1
)

replace dubbo.apache.org/dubbogo/v3 => ../../../../../dubbo-go
