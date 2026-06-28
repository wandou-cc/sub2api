#!/usr/bin/env bash
# Build and deploy the current workspace to the existing Docker Compose server.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

SERVER="${SERVER:-root@156.245.247.107}"
REMOTE_BUILD_DIR="${REMOTE_BUILD_DIR:-/opt/sub2api-build}"
REMOTE_RUN_DIR="${REMOTE_RUN_DIR:-/opt/sub2api}"
TAG="${TAG:-deploy-$(date -u +%Y%m%dT%H%M%SZ)}"
IMAGE="sub2api:${TAG}"

echo "Deploy target: ${SERVER}"
echo "Image: ${IMAGE}"

echo "Checking server containers..."
ssh "${SERVER}" "docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}' | sed -n '1,10p'"

echo "Backing up PostgreSQL..."
ssh "${SERVER}" "set -euo pipefail
cd '${REMOTE_RUN_DIR}'
backup_dir='${REMOTE_RUN_DIR}/backups'
mkdir -p \"\$backup_dir\"
backup_file=\"\$backup_dir/pre-deploy-\$(date -u +%Y%m%dT%H%M%SZ).dump\"
docker exec sub2api-postgres sh -lc 'pg_dump -U \"\${POSTGRES_USER:-sub2api}\" -d \"\${POSTGRES_DB:-sub2api}\" -Fc' > \"\$backup_file\"
ls -lh \"\$backup_file\""

echo "Syncing source..."
rsync -az --delete --delete-excluded \
  --exclude='.git/' \
  --exclude='frontend/node_modules/' \
  --exclude='node_modules/' \
  --exclude='.gocache/' \
  --exclude='.gomodcache/' \
  --exclude='out/' \
  --exclude='runtime-image/' \
  "${REPO_ROOT}/" \
  "${SERVER}:${REMOTE_BUILD_DIR}/"

echo "Building image with BuildKit..."
ssh "${SERVER}" "set -euo pipefail
cd '${REMOTE_BUILD_DIR}'
DOCKER_BUILDKIT=1 docker build -t '${IMAGE}' --build-arg COMMIT='${TAG}' ."

echo "Switching sub2api container..."
ssh "${SERVER}" "set -euo pipefail
cd '${REMOTE_RUN_DIR}'
cp docker-compose.yml docker-compose.yml.before-${TAG}
perl -0pi -e 's/image: sub2api:[^\n]+/image: ${IMAGE}/' docker-compose.yml
docker compose up -d --no-deps sub2api"

echo "Waiting for health check..."
ssh "${SERVER}" "set -euo pipefail
for i in \$(seq 1 40); do
  status=\$(docker inspect -f '{{.State.Health.Status}}' sub2api 2>/dev/null || true)
  echo \"health=\$status\"
  [ \"\$status\" = healthy ] && break
  sleep 3
done
[ \"\$(docker inspect -f '{{.State.Health.Status}}' sub2api)\" = healthy ]
curl -fsS http://127.0.0.1:18080/health
echo
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}' | sed -n '1,10p'"

echo "Checking recent error logs..."
ssh "${SERVER}" "cd '${REMOTE_RUN_DIR}' && docker compose logs --since=3m sub2api | grep -Ei 'error|fatal|panic|migration' | tail -40 || true"

echo "Deployed ${IMAGE}"
