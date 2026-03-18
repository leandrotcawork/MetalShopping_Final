$ErrorActionPreference = "Stop"

$containerName = "metalshopping-keycloak"
$themeName = "metalshopping"

docker exec $containerName /opt/keycloak/bin/kcadm.sh config credentials `
  --server http://127.0.0.1:8080 `
  --realm master `
  --user admin `
  --password admin | Out-Null

docker exec $containerName /opt/keycloak/bin/kcadm.sh update realms/metalshopping `
  -s "loginTheme=$themeName" | Out-Null

Write-Host "Applied Keycloak login theme '$themeName' to realm 'metalshopping'." -ForegroundColor Green
