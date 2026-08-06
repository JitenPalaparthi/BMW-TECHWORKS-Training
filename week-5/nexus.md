docker run -d -p 8081:8081 -p 8082:8082 --name nexus -v nexus-data:/nexus-data sonatype/nexus3

curl -u 'admin:admin@123' \
http://localhost:8081/service/rest/v1/system/eula

DISCLAIMER=$(curl -s -u 'admin:admin@123' \
http://localhost:8081/service/rest/v1/system/eula | jq -r .disclaimer)

curl -X POST \
-u 'admin:admin@123' \
-H "Content-Type: application/json" \
http://localhost:8081/service/rest/v1/system/eula \
-d "{\"accepted\":true,\"disclaimer\":\"$DISCLAIMER\"}"

## For windows 

curl.exe -u "admin:admin@123" `
http://localhost:8081/service/rest/v1/system/eula

$response = Invoke-RestMethod `
    -Uri "http://localhost:8081/service/rest/v1/system/eula" `
    -Authentication Basic `
    -Credential (New-Object System.Management.Automation.PSCredential(
        "admin",
        (ConvertTo-SecureString "admin@123" -AsPlainText -Force)
    ))

$DISCLAIMER = $response.disclaimer

$body = @{
    accepted  = $true
    disclaimer = $DISCLAIMER
} | ConvertTo-Json

Invoke-RestMethod `
    -Method POST `
    -Uri "http://localhost:8081/service/rest/v1/system/eula" `
    -Authentication Basic `
    -Credential (New-Object System.Management.Automation.PSCredential(
        "admin",
        (ConvertTo-SecureString "admin@123" -AsPlainText -Force)
    )) `
    -ContentType "application/json" `
    -Body $body