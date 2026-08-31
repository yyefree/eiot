import re

with open('main.go', 'r', encoding='utf-8') as f:
    content = f.read()

# The exact old string to replace
old = """\t\t})

\t\tlog.Println("[EMQX] subscribed Feiyan-style topics: /sys/+/+/prop/post, /sys/+/+/thing/event/property/batch_post, /sys/+/+/thing/event/property/history_post, /ota/device/inform/+/+, /sys/+/+/event/post, /sys/+/+/service/+/reply, /sys/+/+/rrpc/request/+")
\t}"""

new = """\t\t})

\t// 飞燕标准 Topic：设备动态注册 /sys/{ProductKey}/{DeviceName}/thing/register
\t_ = emqx.Subscribe("/sys/+/+/thing/register", func(topic string, payload []byte) {
\t\tsegs := strings.Split(strings.Trim(topic, "/"), "/")
\t\tif len(segs) < 4:
\t\t\treturn
\t\tproductKey := segs[1]
\t\tdeviceName := segs[2]
\t\tvar body map[string]interface{}
\t\tif err := json.Unmarshal(payload, &body); err != nil:
\t\t\treturn
\t\t
\t\tvar d model.Device
\t\tif err := dao.DB.Where("product_key = ? AND device_name = ?", productKey, deviceName).First(&d).Error; err != nil:
\t\t\t# 设备不存在，可能是预注册模式，创建新设备
\t\tvar p model.Product
\t\tif err := dao.DB.Where("product_key = ?", productKey).First(&p).Error; err != nil:
\t\t\treturn
\t\t# 生成随机 DeviceSecret
\t\tdeviceSecret := fmt.Sprintf("%s%08d", productKey[:8], rand.Intn(100000000))
\t\td = model.Device{
\t\t\tDeviceName:    deviceName,
\t\t\tDeviceSN:      fmt.Sprintf("%s_%s", productKey, deviceName),
\t\t\tDeviceSecret:  deviceSecret,
\t\t\tProductKey:    productKey,
\t\t\tProductID:     p.ID,
\t\t\tOwnerID:       1, # 系统用户
\t\t\tStatus:        1,
\t\t\tNodeType:      "device",
\t\t\tBindMode:      "product_secret",
\t\t}
\t\tif err := dao.DB.Create(&d).Error; err != nil:
\t\t\treturn
\t\t}
\t\t
\t\t# 返回设备密钥
\t\tresponse := map[string]interface{}{
\t\t\t"id":      body["id"],
\t\t\t"code":    200,
\t\t\t"message": "success",
\t\t\t"data": map[string]interface{}{
\t\t\t\t"deviceSecret": d.DeviceSecret,
\t\t\t},
\t\t}
\t\tresponseJSON, _ := json.Marshal(response)
\t\treplyTopic := fmt.Sprintf("/sys/%s/%s/thing/register_reply", productKey, deviceName)
\t\temqx.Publish(replyTopic, response)
\t\tlogic.LogMqttMessageAsync(d.DeviceSN, replyTopic, "down", string(responseJSON))
\t})

\tlog.Println("[EMQX] subscribed Feiyan-style topics: /sys/+/+/prop/post, /sys/+/+/thing/event/property/batch_post, /sys/+/+/thing/event/property/history_post, /ota/device/inform/+/+, /sys/+/+/event/post, /sys/+/+/service/+/reply, /sys/+/+/rrpc/request/+, /sys/+/+/thing/register")"""

if old in content:
    content = content.replace(old, new)
    with open('main.go', 'w', encoding='utf-8') as f:
        f.write(content)
    print('Replaced successfully')
else:
    print('Old string not found')
    # Try to find with different whitespace
    import re
    matches = re.findall(r'(\}\s*\n\s*log\.Println\("\[EMQX\] subscribed Feiyan-style topics:)', content)
    print('Found patterns:', matches)