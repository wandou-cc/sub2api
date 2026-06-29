# 本地仓库开发启动说明

这套开发环境用于“改当前仓库代码，马上在本地 Docker 里看效果”。

## 当前入口

- 前端开发页面：http://127.0.0.1:3000
- 后端开发接口：http://127.0.0.1:8081
- 原发布版容器仍在：http://127.0.0.1:8080

## 管理员账号

- 邮箱：`admin@sub2api.local`
- 密码：`5c1ce5d2acefd7684fa7fd39`

## 启动

在仓库根目录执行：

```bash
cd deploy
docker compose --env-file /dev/null -f docker-compose.repo-dev.yml up -d
```

首次启动会安装前端依赖、下载 Go 依赖并编译后端，时间会比较长。后续会使用 Docker volume 缓存。

## 查看状态

```bash
docker ps --format 'table {{.Names}}\t{{.Ports}}\t{{.Status}}'
curl http://127.0.0.1:8081/health
```

## 查看日志

```bash
docker logs -f sub2api-dev
docker logs -f sub2api-frontend-dev
```

## 开发方式

- 修改 `frontend/`：Vite 会自动热更新，浏览器访问 `http://127.0.0.1:3000`。
- 修改 `backend/`：`air` 会自动重新编译并重启后端，前端请求会代理到 `sub2api-dev:8080`。
- 数据库、Redis 和 `deploy/data` 继续使用现有本地 Docker 部署的数据。

## 停止

```bash
cd deploy
docker compose --env-file /dev/null -f docker-compose.repo-dev.yml down
```

这只停止开发容器，不会停止 `sub2api`、`sub2api-postgres`、`sub2api-redis` 这组发布版容器。
