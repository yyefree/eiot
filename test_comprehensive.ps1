$ErrorActionPreference = "Continue"
$base = "http://localhost:8080/api"

# Login
$r = Invoke-WebRequest -Uri "$base/auth/login" -Method POST -ContentType 'application/json; charset=utf-8' -Body '{"phone":"13800000000","password":"admin123"}' -UseBasicParsing
$token = ($r.Content | ConvertFrom-Json).data.token
Write-Host "Token: $token"

# Test RRPC service call (device 60)
$rrpc = Invoke-WebRequest -Uri "$base/device/60/service/rrpc" -Method POST -ContentType 'application/json; charset=utf-8' -Body '{"service_id":"test_service","service_name":"Test Service","input":{},"timeout_sec":5}' -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "RRPC: $($rrpc.StatusCode) - $($rrpc.Content)"

# Test batch property report
$batchBody = '{"sn":"PK_TEMP001741635300002","properties":{"temp_01":25.5,"hum_01":60}}'
$batch = Invoke-WebRequest -Uri "$base/device/report" -Method POST -ContentType 'application/json; charset=utf-8' -Body $batchBody -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Batch Report: $($batch.StatusCode) - $($batch.Content)"

# Test history data
$hist = Invoke-WebRequest -Uri "$base/device/data/PK_TEMP001741635300002?property=temp_01&limit=5" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "History: $($hist.StatusCode) - $($hist.Content)"

# Test device events
$events = Invoke-WebRequest -Uri "$base/device/60/event?page=1&size=10" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Events: $($events.StatusCode) - $($events.Content)"

# Test device services
$services = Invoke-WebRequest -Uri "$base/device/60/service?page=1&size=10" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Services: $($services.StatusCode) - $($services.Content)"

# Test device shadow
$shadow = Invoke-WebRequest -Uri "$base/device/60/shadow" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Shadow: $($shadow.StatusCode) - $($shadow.Content)"

# Test shadow update
$shadowUpdate = Invoke-WebRequest -Uri "$base/device/60/shadow" -Method PUT -ContentType 'application/json; charset=utf-8' -Body '{"desired":{"temp":26,"mode":"auto"}}' -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Shadow Update: $($shadowUpdate.StatusCode) - $($shadowUpdate.Content)"

# Test rule engine
$rules = Invoke-WebRequest -Uri "$base/admin/rule?page=1&size=10" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Rules: $($rules.StatusCode) - $($rules.Content)"

# Test device groups
$groups = Invoke-WebRequest -Uri "$base/admin/device-group?page=1&size=10" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Device Groups: $($groups.StatusCode) - $($groups.Content)"

# Test device tags
$tags = Invoke-WebRequest -Uri "$base/admin/device/60/tag" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Device Tags: $($tags.StatusCode) - $($tags.Content)"

# Test provisioning
$prov = Invoke-WebRequest -Uri "$base/admin/provisioning/records?page=1&size=10" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Provisioning Records: $($prov.StatusCode) - $($prov.Content)"

# Test data flow
$flows = Invoke-WebRequest -Uri "$base/admin/data-flow?page=1&size=10" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Data Flows: $($flows.StatusCode) - $($flows.Content)"

# Test protocol gateways
$gateways = Invoke-WebRequest -Uri "$base/admin/protocol-gateway?page=1&size=10" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Protocol Gateways: $($gateways.StatusCode) - $($gateways.Content)"

# Test device diagnostics
$diags = Invoke-WebRequest -Uri "$base/admin/device-diagnostic?page=1&size=10" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Device Diagnostics: $($diags.StatusCode) - $($diags.Content)"

# Test device metrics
$metrics = Invoke-WebRequest -Uri "$base/admin/device/PK_TEMP001741635300002/metrics?days=7" -Headers @{Authorization="Bearer $token"} -UseBasicParsing
Write-Host "Device Metrics: $($metrics.StatusCode) - $($metrics.Content)"