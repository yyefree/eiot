package logic

import (
	"testing"
	"time"

	"eiot/internal/dao"
	"eiot/internal/model"
	"eiot/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// AutoMigrate required models
	err = db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.Device{},
		&model.DeviceShadow{},
		&model.Rule{},
		&model.RuleExecution{},
		&model.DeviceGroup{},
		&model.DeviceGroupMember{},
		&model.DeviceTag{},
		&model.MqttMessage{},
		&model.DeviceTopology{},
		&model.DeviceProvisioning{},
		&model.ProvisioningConfig{},
		&model.DataFlow{},
		&model.DataFlowExecution{},
		&model.ProtocolGateway{},
		&model.ProtocolMapping{},
		&model.DeviceDiagnostic{},
		&model.DeviceMetrics{},
		&model.DeviceEventHistory{},
		&model.DeviceServiceHistory{},
		&model.DeviceDataHistory{},
		&model.AlertRule{},
		&model.Message{},
		&model.Scene{},
		&model.SceneCondition{},
		&model.SceneAction{},
		&model.Home{},
		&model.HomeMember{},
		&model.Room{},
		&model.RoomDevice{},
		&model.Project{},
		&model.OTAFirmware{},
		&model.DeviceShare{},
		&model.UserDeviceUI{},
		&model.Dashboard{},
		&model.OperationLog{},
	)
	require.NoError(t, err)

	// Replace global dao.DB
	dao.DB = db
	return db
}

func TestBizError(t *testing.T) {
	err := NewBizError(400, 10001, "test error")
	assert.Equal(t, 400, err.HTTPCode)
	assert.Equal(t, 10001, err.Code)
	assert.Equal(t, "test error", err.Message)
	assert.Equal(t, "test error", err.Error())
}

func TestNewBizErrorWithCode(t *testing.T) {
	err := NewBizErrorWithCode(400, ErrCodeInvalidParams, ErrMsgInvalidParams, "test param")
	assert.Equal(t, 400, err.HTTPCode)
	assert.Equal(t, ErrCodeInvalidParams, err.Code)
	assert.Contains(t, err.Message, "test param")
}

func TestGetErrorMessage(t *testing.T) {
	msg := GetErrorMessage(ErrCodeUserNotExist, "zh")
	assert.Equal(t, "用户不存在", msg)

	msg = GetErrorMessage(99999, "zh")
	assert.Equal(t, "未知错误", msg)
}

func TestLoginByPassword(t *testing.T) {
	db := setupTestDB(t)

	// Create test user
	hashed, _ := util.HashPassword("test123")
	user := &model.User{
		Phone:    "13800000000",
		Password: hashed,
		Role:     "user",
		Status:   1,
	}
	db.Create(user)

	// Test successful login
	token, u, err := LoginByPassword("13800000000", "test123", "test-secret")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, uint(1), u.ID)

	// Test wrong password
	_, _, err = LoginByPassword("13800000000", "wrong", "test-secret")
	assert.Error(t, err)
	assert.Equal(t, ErrCodeWrongPassword, err.(*BizError).Code)

	// Test disabled user
	db.Model(&model.User{}).Where("id = ?", 1).Update("status", 0)
	_, _, err = LoginByPassword("13800000000", "test123", "test-secret")
	assert.Error(t, err)
	assert.Equal(t, ErrCodeUserDisabled, err.(*BizError).Code)
}

func TestDeviceShadowOptimisticLock(t *testing.T) {
	db := setupTestDB(t)

	deviceSN := "TEST_SHADOW_001"

	// Test UpdateDeviceShadowDesired - first create
	err := UpdateDeviceShadowDesired(deviceSN, map[string]interface{}{"temp": 25})
	assert.NoError(t, err)

	// Update with same version (should succeed)
	err = UpdateDeviceShadowDesired(deviceSN, map[string]interface{}{"temp": 26})
	assert.NoError(t, err)

	// Verify version incremented
	var shadow model.DeviceShadow
	db.Where("device_sn = ?", deviceSN).First(&shadow)
	assert.Equal(t, 2, shadow.Version)

	// Test SyncDeviceShadowReported
	err = SyncDeviceShadowReported(deviceSN, map[string]interface{}{"humidity": 60})
	assert.NoError(t, err)

	// Verify reported updated
	db.Where("device_sn = ?", deviceSN).First(&shadow)
	assert.Contains(t, shadow.ReportedJSON, "humidity")
	assert.Equal(t, 3, shadow.Version)
}

func TestRuleEngineCache(t *testing.T) {
	db := setupTestDB(t)

	// Create rules
	r1 := &model.Rule{
		Name:        "Rule 1",
		Description: "Test rule 1",
		Type:        "auto",
		Enabled:     true,
		TriggerJSON: `{"type":"device_property","property":"temp","operator":">","threshold":"30"}`,
		ActionJSON:  `[{"type":"device_control","device_sn":"DEV001","property":"fan","value":true}]`,
		CreatedBy:   1,
	}
	r2 := &model.Rule{
		Name:        "Rule 2",
		Description: "Test rule 2",
		Type:        "auto",
		Enabled:     false, // disabled
		TriggerJSON: `{"type":"device_property","property":"hum","operator":">","threshold":"80"}`,
		ActionJSON:  `[{"type":"notify","title":"High humidity"}]`,
		CreatedBy:   1,
	}
	db.Create(r1)
	db.Create(r2)

	// Invalidate cache to force reload
	invalidateRuleCache()

	// Get enabled rules - should only return r1
	rules := getEnabledRules()
	assert.Len(t, rules, 1)
	assert.Equal(t, "Rule 1", rules[0].Name)

	// Enable r2 and invalidate
	r2.Enabled = true
	db.Save(r2)
	invalidateRuleCache()

	rules = getEnabledRules()
	assert.Len(t, rules, 2)
}

func TestDeviceProvisioning(t *testing.T) {
	db := setupTestDB(t)

	// Create product
	product := &model.Product{
		Name:        "Test Product",
		ProductKey:  "PK_TEST",
		Category:    "sensor",
		NetworkType: "wifi",
		CreatedBy:   1,
	}
	db.Create(product)

	// Start provisioning
	prov, err := StartProvisioning(product.ID, "SN001", "Test Device", "ble", "", "", "", 1)
	assert.NoError(t, err)
	assert.NotEmpty(t, prov.PinCode)
	assert.Equal(t, "pending", prov.Status)

	// Complete provisioning
	err = CompleteProvisioning("SN001", true, "")
	assert.NoError(t, err)

	// Verify completed
	var prov2 model.DeviceProvisioning
	db.Where("device_sn = ?", "SN001").First(&prov2)
	assert.Equal(t, "success", prov2.Status)
	assert.NotNil(t, prov2.CompletedAt)
}

func TestMqttMessageAsync(t *testing.T) {
	db := setupTestDB(t)

	// Test async logging
	LogMqttMessageAsync("SN001", "/sys/test/prop/post", "up", `{"temp":25}`)
	LogMqttMessageAsync("SN001", "/sys/test/prop/set", "down", `{"fan":true}`)

	// Wait for async flush
	time.Sleep(1 * time.Second)

	var count int64
	db.Model(&model.MqttMessage{}).Where("device_sn = ?", "SN001").Count(&count)
	assert.GreaterOrEqual(t, count, int64(2))
}

func TestValidatePropertyReport(t *testing.T) {
	db := setupTestDB(t)

	// Create product with property specs
	product := &model.Product{
		Name:       "Test Product",
		ProductKey: "PK_TEST2",
		Category:   "sensor",
		PropertiesJSON: `[{"identifier":"temp","name":"温度","accessMode":"r","dataType":{"type":"float","specs":{"min":-40,"max":85,"unit":"℃"}}},
{"identifier":"switch","name":"开关","accessMode":"rw","dataType":{"type":"bool","specs":{}}}]`,
		CreatedBy: 1,
	}
	db.Create(product)

	// Valid report
	warnings := ValidatePropertyReport(product.ID, map[string]interface{}{"temp": 25.5, "switch": true})
	assert.Len(t, warnings, 0)

	// Out of range
	warnings = ValidatePropertyReport(product.ID, map[string]interface{}{"temp": 100})
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "temp")

	// Wrong type
	warnings = ValidatePropertyReport(product.ID, map[string]interface{}{"switch": "on"})
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "switch")
}

func TestErrorCodes(t *testing.T) {
	// Test error message mapping
	assert.Equal(t, "用户不存在", GetErrorMessage(ErrCodeUserNotExist, "zh"))
	assert.Equal(t, "设备不存在", GetErrorMessage(ErrCodeDeviceNotExist, "zh"))
	assert.Equal(t, "影子版本冲突", GetErrorMessage(ErrCodeShadowConflict, "zh"))
	assert.Equal(t, "未知错误", GetErrorMessage(99999, "zh"))
}

func TestRuleTriggerMatch(t *testing.T) {
	// Test matchTrigger function (private, test via EvaluateRule)
	db := setupTestDB(t)

	// Create a rule
	rule := &model.Rule{
		Name:        "Temp Alert",
		Type:        "auto",
		Enabled:     true,
		TriggerJSON: `{"type":"device_property","property":"temp","operator":">","threshold":"30"}`,
		ActionJSON:  `[{"type":"notify","title":"High Temp"}]`,
		CreatedBy:   1,
	}
	db.Create(rule)
	invalidateRuleCache()

	// Trigger with temp > 30
	EvaluateRule(map[string]interface{}{
		"type":       "property_report",
		"device_sn":  "TEST_DEV",
		"product_id": uint(0),
		"properties": map[string]interface{}{"temp": 35.5},
	})

	// Check execution recorded
	var count int64
	dao.DB.Model(&model.RuleExecution{}).Where("rule_id = ?", 1).Count(&count)
	assert.Equal(t, int64(1), count)
}