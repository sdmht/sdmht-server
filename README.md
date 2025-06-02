# sdmht-server

神代梦华谭后端

## 安装依赖

```sh
go mod tidy
```

## 启动服务

```sh
air
```

## 测试

```sh
go test ./... -v
```

## 更新依赖

```sh
go get -u -v all
```

## 生成 graphql 代码

```sh
go run github.com/99designs/gqlgen@latest generate
```
