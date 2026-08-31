$ErrorActionPreference = "Continue"
$base = "http://localhost:8080/api"

# Login
$r = Invoke-WebRequest -Uri "$base/auth/login" -Method POST -ContentType 'application/json; charset=utf-8' -Body '{"phone":"13800000000","password":"admin123"}' -UseBasicParsing
$token = ($r.Content | ConvertFrom-Json).data.token
Write-Host "Token: $token"

# Test dynamic registration
$regBody = '{"product_key":"PK_TEST","device_name":"test_dev_001"}'
$reg = Invoke-WebRequest -Uri "http://localhost:8080/api/sys/PK_TEST/test_dev_001/thing/register" -Method POST -ContentType 'application/json; charset=utf-8' -Body $regBody -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Register: $($reg.StatusCode) - $($reg.Content)"

# Test RRPC service call
$rrpcBody = '{"service_id":"test_service","service_name":"Test Service","input":{},"timeout_sec":5}'
$rrpc = Invoke-WebRequest -Uri "http://localhost:8080/api/device/60/service/rrpc" -Method POST -ContentType 'application/json; charset=utf-8' -Body '{"service_id":"test_service","service_name":"Test Service","input":{},"timeout_sec":5}' -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "RRPC: $($rrpc.StatusCode) - $($rrpc.Content)"

# Test batch property report
$batchBody = '{"sn":"PK_TEMP001741635300002","properties":{"temp_01":25.5,"hum_01":60}}'
$batch = Invoke-WebRequest -Uri "http://localhost:8080/api/device/report" -Method POST -ContentType 'application/json; charset=utf-8' -Body $batchBody -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Batch Report: $($batch.StatusCode) - $($batch.Content)"

# Test history data
$hist = Invoke-WebRequest -Uri "http://localhost:8080/api/device/data/PK_TEMP001741635300002?property=temp_01&limit=5" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "History: $($hist.StatusCode) - $($hist.Content)"