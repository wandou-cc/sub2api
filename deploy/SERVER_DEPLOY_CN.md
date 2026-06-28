# Sub2API 服务器部署流程

本文记录从本地未提交修改部署到服务器的流程，适用于当前 Docker Compose 部署方式。

## 当前服务器约定

- 服务器：`156.245.247.107`
- 运行目录：`/opt/sub2api`
- 构建目录：`/opt/sub2api-build`
- 应用端口：宿主机 `127.0.0.1:18080` -> 容器 `8080`
- Compose 文件：`/opt/sub2api/docker-compose.yml`
- 数据目录：`/opt/sub2api/data`、`/opt/sub2api/postgres_data`、`/opt/sub2api/redis_data`

生产数据只在 `/opt/sub2api` 下。构建目录只放源码和构建产物，不要把它当作数据目录。

## 部署前检查

本地确认改动范围：

```bash
git status --short --branch
```

服务器确认当前运行状态：

```bash
ssh root@156.245.247.107 'docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"'
ssh root@156.245.247.107 'cd /opt/sub2api && grep -n "image: sub2api" docker-compose.yml'
```

## 数据库备份

有迁移文件或后端数据结构改动时，先备份 PostgreSQL。

```bash
ssh root@156.245.247.107 'set -euo pipefail
cd /opt/sub2api
backup_dir=/opt/sub2api/backups
mkdir -p "$backup_dir"
backup_file="$backup_dir/pre-deploy-$(date -u +%Y%m%dT%H%M%SZ).dump"
docker exec sub2api-postgres sh -lc 'pg_dump -U "${POSTGRES_USER:-sub2api}" -d "${POSTGRES_DB:-sub2api}" -Fc' > "$backup_file"
ls -lh "$backup_file"'
```

## 同步源码

不要同步本地依赖目录和构建缓存，否则会很慢。

```bash
rsync -az --delete \
  --exclude='.git/' \
  --exclude='frontend/node_modules/' \
  --exclude='node_modules/' \
  --exclude='.gocache/' \
  --exclude='.gomodcache/' \
  --exclude='out/' \
  --exclude='runtime-image/' \
  /Users/cc/github/sub2api/ \
  root@156.245.247.107:/opt/sub2api-build/
```

建议使用 `--delete-excluded`，保证服务器构建目录不会残留旧的 `node_modules`：

```bash
rsync -az --delete --delete-excluded \
  --exclude='.git/' \
  --exclude='frontend/node_modules/' \
  --exclude='node_modules/' \
  --exclude='.gocache/' \
  --exclude='.gomodcache/' \
  --exclude='out/' \
  --exclude='runtime-image/' \
  /Users/cc/github/sub2api/ \
  root@156.245.247.107:/opt/sub2api-build/
```

如服务器构建目录残留了旧依赖，清理后再构建：

```bash
ssh root@156.245.247.107 'rm -rf /opt/sub2api-build/frontend/node_modules /opt/sub2api-build/.gocache /opt/sub2api-build/.gomodcache /opt/sub2api-build/out /opt/sub2api-build/runtime-image'
```

## 构建镜像

用唯一标签构建，避免覆盖旧镜像。

```bash
TAG=deploy-$(date -u +%Y%m%dT%H%M%SZ)
ssh root@156.245.247.107 "set -euo pipefail
cd /opt/sub2api-build
DOCKER_BUILDKIT=1 docker build -t sub2api:${TAG} --build-arg COMMIT=${TAG} ."
```

根目录 `Dockerfile` 使用 BuildKit cache mount 缓存 pnpm store、Go module cache 和 Go build cache。只要 `frontend/pnpm-lock.yaml`、`backend/go.mod`、`backend/go.sum` 没变，后续构建会明显更快。

构建成功后确认镜像存在：

```bash
ssh root@156.245.247.107 "docker image inspect sub2api:${TAG} --format '{{.Id}} {{.Created}} {{.Size}}'"
```

## 切换应用容器

只替换应用镜像，不重建 PostgreSQL 和 Redis。

```bash
ssh root@156.245.247.107 "set -euo pipefail
cd /opt/sub2api
cp docker-compose.yml docker-compose.yml.before-${TAG}
perl -0pi -e 's/image: sub2api:[^\n]+/image: sub2api:${TAG}/' docker-compose.yml
docker compose up -d --no-deps sub2api
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}' | sed -n '1,10p'"
```

## 部署后验证

健康检查：

```bash
ssh root@156.245.247.107 'docker inspect -f "{{.Config.Image}} {{.State.Health.Status}}" sub2api'
ssh root@156.245.247.107 'curl -fsS http://127.0.0.1:18080/health'
```

最近错误日志：

```bash
ssh root@156.245.247.107 'cd /opt/sub2api && docker compose logs --since=3m sub2api | grep -Ei "error|fatal|panic|migration" | tail -40 || true'
```

确认指定迁移已应用：

```bash
ssh root@156.245.247.107 'set -euo pipefail
cd /opt/sub2api
set -a
. ./.env
set +a
docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" sub2api-postgres psql -U "${POSTGRES_USER:-sub2api}" -d "${POSTGRES_DB:-sub2api}" -Atc "select filename from schema_migrations order by filename desc limit 10;"'
```

`schema_migrations` 的字段是 `filename`、`checksum`、`applied_at`，不要用 `version` 字段查询。

## 一键部署脚本

常规部署直接从仓库根目录执行：

```bash
deploy/server-deploy.sh
```

脚本默认值：

- `SERVER=root@156.245.247.107`
- `REMOTE_BUILD_DIR=/opt/sub2api-build`
- `REMOTE_RUN_DIR=/opt/sub2api`
- `TAG=deploy-<UTC timestamp>`

需要覆盖时：

```bash
TAG=deploy-test SERVER=root@156.245.247.107 deploy/server-deploy.sh
```

## 常见坑

### 不要直接拉 `weishaw/sub2api:latest`

`deploy/docker-compose*.yml` 默认使用 Docker Hub 镜像。要部署本地未提交修改，必须把源码同步到服务器并构建本地镜像，然后把 Compose 的 `sub2api` 服务改成 `sub2api:<TAG>`。

### `pnpm install --frozen-lockfile` 失败

Dockerfile 使用 `corepack prepare pnpm@9 --activate`，构建时会执行：

```bash
pnpm install --frozen-lockfile
```

如果 `frontend/package.json` 和 `frontend/pnpm-lock.yaml` 不一致，会报：

```text
ERR_PNPM_LOCKFILE_CONFIG_MISMATCH
```

处理方式：

```bash
cd frontend
npx --yes pnpm@9 install --lockfile-only
npx --yes pnpm@9 install --frozen-lockfile
```

本地全局 pnpm 可能是 11，pnpm 11 会忽略 `package.json` 里的 `pnpm.overrides`，不要用它更新本项目锁文件。以 Dockerfile 的 pnpm 9 为准。

### rsync 很慢

先检查是否同步了 `frontend/node_modules`。该目录通常很大，必须排除。Docker 构建会在容器内安装依赖。

### Go 依赖下载长时间无输出

`go mod download` 和 `go build` 在 Docker legacy builder 下可能长时间没有输出。先检查进程，不要急着中断：

```bash
ssh root@156.245.247.107 'ps -eo pid,ppid,etime,stat,pcpu,pmem,cmd | grep -E "docker build|go mod download|go build" | grep -v grep || true'
```

### 回滚

优先回滚镜像标签：

```bash
ssh root@156.245.247.107 'set -euo pipefail
cd /opt/sub2api
cp docker-compose.yml.before-<TAG> docker-compose.yml
docker compose up -d --no-deps sub2api'
```

如果已经执行了数据库迁移，迁移是 forward-only。需要恢复数据时，用部署前的 `pg_dump -Fc` 备份恢复，恢复前必须先确认会覆盖哪些数据。
